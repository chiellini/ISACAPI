#!/bin/bash
# =============================================================================
# ISACAPI Docker Deployment Preparation Script
# =============================================================================
# This script prepares deployment files for ISACAPI:
#   - Downloads docker-compose.local.yml and .env.example
#   - Generates secure secrets for a new deployment
#   - Preserves and backs up an existing .env during an upgrade
#   - Creates necessary data directories
#
# After running this script, you can start services with:
#   docker compose up -d
# =============================================================================

set -euo pipefail
umask 077

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Download only this fork's deployment assets. DEPLOY_REF may be an immutable
# release tag or commit. By default it is derived from IMAGE_TAG so deployment
# assets and the application image cannot silently come from different releases.
GITHUB_REPO="chiellini/ISACAPI"
IMAGE_TAG_OVERRIDE="${IMAGE_TAG:-}"

if [ -z "$IMAGE_TAG_OVERRIDE" ]; then
    echo "IMAGE_TAG is required; set it to a reviewed published release (for example 0.1.174)." >&2
    exit 1
fi
if [[ ! "$IMAGE_TAG_OVERRIDE" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][A-Za-z0-9][A-Za-z0-9._-]*)?$ ]]; then
    echo "IMAGE_TAG must be an exact published version such as 0.1.174; floating tags such as latest are not allowed: $IMAGE_TAG_OVERRIDE" >&2
    exit 1
fi

DEPLOY_REF="${DEPLOY_REF:-v${IMAGE_TAG_OVERRIDE}}"
if [[ ! "$DEPLOY_REF" =~ ^[A-Za-z0-9][A-Za-z0-9._/-]*$ ]] || [[ "$DEPLOY_REF" == *".."* ]]; then
    echo "Invalid DEPLOY_REF: $DEPLOY_REF" >&2
    exit 1
fi

GITHUB_RAW_URL="https://raw.githubusercontent.com/${GITHUB_REPO}/${DEPLOY_REF}/deploy"

# Print colored message
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Generate random secret
generate_secret() {
    openssl rand -hex 32
}

# Replace one key without passing its value to sed/awk child-process arguments.
# The input file is copied atomically and retains mode 0600 under this script's
# umask. If the key is absent (for example in an older template), append it.
set_env_value() {
    local key="$1"
    local value="$2"
    local file="$3"
    local edit_file="${STAGING_DIR}/env.edit"
    local found=0
    local line

    : >"${edit_file}"
    while IFS= read -r line || [ -n "${line}" ]; do
        case "${line}" in
            "${key}="*)
                printf '%s=%s\n' "${key}" "${value}" >>"${edit_file}"
                found=1
                ;;
            *)
                printf '%s\n' "${line}" >>"${edit_file}"
                ;;
        esac
    done <"${file}"
    if [ "${found}" -eq 0 ]; then
        printf '%s=%s\n' "${key}" "${value}" >>"${edit_file}"
    fi
    mv "${edit_file}" "${file}"
    chmod 600 "${file}"
}

download_file() {
    local source_url="$1"
    local destination="$2"

    if command_exists curl; then
        curl -fsSL "${source_url}" -o "${destination}"
    elif command_exists wget; then
        wget -q "${source_url}" -O "${destination}"
    else
        print_error "Neither curl nor wget is installed. Please install one of them."
        exit 1
    fi
}

cleanup_staging() {
    case "${STAGING_DIR:-}" in
        .docker-deploy.tmp.*)
            if [ -d "${STAGING_DIR}" ]; then
                rm -rf -- "${STAGING_DIR}"
            fi
            ;;
    esac
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Main installation function
main() {
    echo ""
    echo "=========================================="
    echo "  ISACAPI Deployment Preparation"
    echo "=========================================="
    echo ""

    local is_upgrade=false
    local backup_dir=""

    if [ -L ".env" ] || [ -L "docker-compose.yml" ]; then
        print_error "Refusing to replace a symbolic-link deployment file."
        exit 1
    fi

    # A Compose file without its original environment file is ambiguous: the
    # database volumes may already exist, and generating replacement secrets
    # would break database access, sessions, and encrypted TOTP data.
    if [ -f "docker-compose.yml" ] && [ ! -f ".env" ]; then
        print_error "docker-compose.yml exists but .env is missing."
        print_error "Restore the original .env from backup before upgrading; no files were changed."
        exit 1
    fi

    if [ -f ".env" ] || [ -f "docker-compose.yml" ]; then
        is_upgrade=true
        print_warning "Existing deployment files were detected."
        if [ "${DEPLOY_OVERWRITE:-}" != "1" ]; then
            if ! read -p "Back up the current files and install release ${IMAGE_TAG_OVERRIDE}? (y/N): " -r; then
                echo
                print_error "Confirmation input is unavailable. Re-run with DEPLOY_OVERWRITE=1 after reviewing the upgrade."
                exit 1
            fi
            echo
            if [[ ! ${REPLY} =~ ^[Yy]$ ]]; then
                print_info "Cancelled."
                exit 0
            fi
        fi
    elif ! command_exists openssl; then
        print_error "openssl is not installed. Please install openssl first."
        exit 1
    fi

    STAGING_DIR=$(mktemp -d ".docker-deploy.tmp.XXXXXX")
    trap cleanup_staging EXIT
    trap 'exit 130' HUP INT TERM

    # Stage downloads first so a network failure cannot partially overwrite a
    # working deployment.
    print_info "Downloading docker-compose.yml..."
    download_file "${GITHUB_RAW_URL}/docker-compose.local.yml" "${STAGING_DIR}/docker-compose.yml"
    print_success "Downloaded docker-compose.yml"

    print_info "Downloading .env.example..."
    download_file "${GITHUB_RAW_URL}/.env.example" "${STAGING_DIR}/.env.example"
    print_success "Downloaded .env.example"

    if [ "${is_upgrade}" = true ]; then
        backup_dir=".isacapi-deploy-backup-$(date -u +%Y%m%dT%H%M%SZ)-$$"
        mkdir -m 700 "${backup_dir}"
        cp -p ".env" "${backup_dir}/.env"
        if [ -f "docker-compose.yml" ]; then
            cp -p "docker-compose.yml" "${backup_dir}/docker-compose.yml"
        fi
        if [ -f ".env.example" ]; then
            cp -p ".env.example" "${backup_dir}/.env.example"
        fi

        # Preserve the complete existing environment, including all database,
        # JWT, TOTP, and administrator credentials. Only the explicitly chosen
        # immutable image tag is changed.
        set_env_value "IMAGE_TAG" "${IMAGE_TAG_OVERRIDE}" ".env"
        mv "${STAGING_DIR}/docker-compose.yml" "docker-compose.yml"
        mv "${STAGING_DIR}/.env.example" ".env.example"
        print_success "Existing .env was preserved; backup saved in ${backup_dir}/"
    else
        print_info "Generating secure secrets for a new deployment..."
        JWT_SECRET=$(generate_secret)
        TOTP_ENCRYPTION_KEY=$(generate_secret)
        POSTGRES_PASSWORD=$(generate_secret)
        ADMIN_PASSWORD=$(generate_secret)

        cp "${STAGING_DIR}/.env.example" "${STAGING_DIR}/.env"
        set_env_value "JWT_SECRET" "${JWT_SECRET}" "${STAGING_DIR}/.env"
        set_env_value "TOTP_ENCRYPTION_KEY" "${TOTP_ENCRYPTION_KEY}" "${STAGING_DIR}/.env"
        set_env_value "POSTGRES_PASSWORD" "${POSTGRES_PASSWORD}" "${STAGING_DIR}/.env"
        set_env_value "ADMIN_PASSWORD" "${ADMIN_PASSWORD}" "${STAGING_DIR}/.env"
        set_env_value "IMAGE_TAG" "${IMAGE_TAG_OVERRIDE}" "${STAGING_DIR}/.env"

        mv "${STAGING_DIR}/docker-compose.yml" "docker-compose.yml"
        mv "${STAGING_DIR}/.env.example" ".env.example"
        mv "${STAGING_DIR}/.env" ".env"
    fi

    # Create data directories
    print_info "Creating data directories..."
    mkdir -p data postgres_data redis_data
    print_success "Created data directories"

    # Set secure permissions for .env file (readable/writable only by owner)
    chmod 600 .env
    echo ""

    # Display completion message
    echo "=========================================="
    echo "  Preparation Complete!"
    echo "=========================================="
    echo ""
    if [ "${is_upgrade}" = true ]; then
        print_success "Deployment files were upgraded without rotating credentials."
    else
        print_success "Generated credentials were written to .env (mode 600)."
    fi
    print_warning "Secret values are intentionally not printed; back up .env securely."
    echo ""
    echo "Directory structure:"
    echo "  docker-compose.yml        - Docker Compose configuration"
    echo "  .env                      - Preserved or generated environment variables"
    echo "  .env.example              - Example template (for reference)"
    echo "  data/                     - Application data (will be created on first run)"
    echo "  postgres_data/            - PostgreSQL data"
    echo "  redis_data/               - Redis data"
    echo ""
    echo "Next steps:"
    echo "  1. (Optional) Edit .env to customize configuration"
    echo "  2. Start services:"
    echo "     docker compose up -d"
    echo ""
    echo "  3. View logs:"
    echo "     docker compose logs -f sub2api"
    echo ""
    echo "  4. Access Web UI:"
    echo "     http://localhost:8080"
    echo ""
    if [ "${is_upgrade}" = false ]; then
        print_info "The generated admin password is stored in .env and is not printed."
        print_info "Read it locally when needed, then rotate it after first login."
    else
        print_info "Existing credentials remain in .env; the original file is in ${backup_dir}/.env."
    fi
    echo ""
}

# Run main function
main "$@"
