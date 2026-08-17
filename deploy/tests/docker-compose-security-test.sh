#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
cd "$repo_root"

check_application_security_opt() {
  file=$1
  count=$(
    awk '
      $0 == "  sub2api:" {
        in_application = 1
        next
      }
      in_application && $0 ~ /^  [A-Za-z0-9_-]+:$/ {
        in_application = 0
      }
      in_application && $0 == "    security_opt:" {
        in_security_opt = 1
        next
      }
      in_application && in_security_opt && $0 == "      - no-new-privileges:true" {
        count++
      }
      END { print count + 0 }
    ' "$file"
  )

  if [ "$count" -ne 1 ]; then
    printf '%s must enable no-new-privileges exactly once for the sub2api service\n' "$file" >&2
    exit 1
  fi
}

assert_line() {
  file=$1
  line=$2
  grep -Fqx "$line" "$file" || {
    printf '%s is missing secure production default: %s\n' "$file" "$line" >&2
    exit 1
  }
}

for compose_file in \
  deploy/docker-compose.yml \
  deploy/docker-compose.local.yml \
  deploy/docker-compose.standalone.yml \
  deploy/docker-compose.dev.yml
do
  check_application_security_opt "$compose_file"
done

for compose_file in \
  deploy/docker-compose.yml \
  deploy/docker-compose.local.yml \
  deploy/docker-compose.standalone.yml
do
  assert_line "$compose_file" '    image: "${IMAGE_REPOSITORY:-ghcr.io/chiellini/sub2api}:${IMAGE_TAG:?Set IMAGE_TAG to a published release version}"'
  assert_line "$compose_file" '      - ADMIN_PASSWORD=${ADMIN_PASSWORD:-}'
  assert_line "$compose_file" '      - JWT_SECRET=${JWT_SECRET:?JWT_SECRET is required}'
  assert_line "$compose_file" '      - TOTP_ENCRYPTION_KEY=${TOTP_ENCRYPTION_KEY:?TOTP_ENCRYPTION_KEY is required}'
  assert_line "$compose_file" '      - "${BIND_HOST:-127.0.0.1}:${SERVER_PORT:-8080}:8080"'
  assert_line "$compose_file" '      - CONVERSATION_ARCHIVE_ENABLED=${CONVERSATION_ARCHIVE_ENABLED:-false}'
  assert_line "$compose_file" '      - CONVERSATION_ARCHIVE_ENCRYPT_CONTENT=${CONVERSATION_ARCHIVE_ENCRYPT_CONTENT:-true}'
  assert_line "$compose_file" '      - SECURITY_URL_ALLOWLIST_ENABLED=${SECURITY_URL_ALLOWLIST_ENABLED:-true}'
  assert_line "$compose_file" '      - SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=${SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP:-false}'
  assert_line "$compose_file" '      - SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=${SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS:-false}'
  assert_line "$compose_file" '      - SECURITY_URL_ALLOWLIST_TRUST_UPSTREAM_PROXY=${SECURITY_URL_ALLOWLIST_TRUST_UPSTREAM_PROXY:-false}'
  assert_line "$compose_file" '      - SECURITY_TRUST_FORWARDED_IP_FOR_API_KEY_ACL=${SECURITY_TRUST_FORWARDED_IP_FOR_API_KEY_ACL:-false}'
done

assert_line deploy/docker-compose.dev.yml '      - ADMIN_PASSWORD=${ADMIN_PASSWORD:-dev-only-change-me-now}'

grep -Fq 'GITHUB_REPO="chiellini/ISACAPI"' deploy/docker-deploy.sh || {
  printf 'docker-deploy.sh must download deployment assets from the fork\n' >&2
  exit 1
}
grep -Fq 'DEPLOY_REF="${DEPLOY_REF:-v${IMAGE_TAG_OVERRIDE}}"' deploy/docker-deploy.sh || {
  printf 'docker-deploy.sh must derive its default deployment ref from IMAGE_TAG\n' >&2
  exit 1
}
if IMAGE_TAG=latest bash deploy/docker-deploy.sh >/dev/null 2>&1; then
  printf 'docker-deploy.sh must reject floating image tags\n' >&2
  exit 1
fi
grep -Fq 'SERVER_HOST="127.0.0.1"' deploy/install.sh || {
  printf 'install.sh must bind the first-run setup wizard to loopback by default\n' >&2
  exit 1
}
grep -Fq 'Environment=SETUP_HOST=${SERVER_HOST}' deploy/install.sh || {
  printf 'install.sh must pass the selected host to the setup-only listener\n' >&2
  exit 1
}
grep -Fq 'Environment=SETUP_PORT=${SERVER_PORT}' deploy/install.sh || {
  printf 'install.sh must pass the selected port to the setup-only listener\n' >&2
  exit 1
}
for compose_file in deploy/docker-compose.yml deploy/docker-compose.local.yml deploy/docker-compose.dev.yml
do
  test "$(grep -Fxc '      - REDIS_PASSWORD=${REDIS_PASSWORD:-}' "$compose_file")" -eq 2 || {
    printf '%s must pass REDIS_PASSWORD separately to the app and Redis services\n' "$compose_file" >&2
    exit 1
  }
  grep -Fq '$${REDIS_PASSWORD:+--requirepass "$$REDIS_PASSWORD"}' "$compose_file" || {
    printf '%s must defer Redis password expansion until container runtime\n' "$compose_file" >&2
    exit 1
  }
  awk '
    /REDIS_PASSWORD:\+/ && index($0, "$${REDIS_PASSWORD:+") == 0 { unsafe = 1 }
    END { exit unsafe ? 0 : 1 }
  ' "$compose_file" && {
    printf '%s interpolates REDIS_PASSWORD into a shell program\n' "$compose_file" >&2
    exit 1
  }
done
for upstream_host in \
  api.x.ai '*.api.x.ai' chatgpt.com vidgen.x.ai 'bedrock-runtime.*.amazonaws.com' \
  cli-chat-proxy.grok.com api.minimax.io api.deepseek.com \
  daily-cloudcode-pa.sandbox.googleapis.com aiplatform.googleapis.com \
  '*.aiplatform.googleapis.com' ollama.com www.ollama.com
do
  grep -Fq -- "- \"$upstream_host\"" deploy/config.example.yaml || {
    printf 'config.example.yaml is missing official upstream host: %s\n' "$upstream_host" >&2
    exit 1
  }
done
if grep -Fq 'POSTGRES_PASSWORD:     ${POSTGRES_PASSWORD}' deploy/docker-deploy.sh; then
  printf 'docker-deploy.sh must not print generated credentials\n' >&2
  exit 1
fi
for upgrade_guard in \
  'docker-compose.yml exists but .env is missing.' \
  'cp -p ".env" "${backup_dir}/.env"' \
  'Existing .env was preserved' \
  'without rotating credentials'
do
  grep -Fq "$upgrade_guard" deploy/docker-deploy.sh || {
    printf 'docker-deploy.sh is missing upgrade credential protection: %s\n' "$upgrade_guard" >&2
    exit 1
  }
done
grep -Fqx '.isacapi-deploy-backup-*/' deploy/.gitignore || {
  printf 'deployment secret backups must be ignored by git\n' >&2
  exit 1
}
if grep -Eq 'Wei-Shaw/sub2api|weishaw/sub2api|sub2api:latest' deploy/docker-deploy.sh deploy/docker-compose*.yml; then
  printf 'production deployment files must not reference upstream or latest app images\n' >&2
  exit 1
fi

printf 'docker compose security test passed\n'
