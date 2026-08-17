# ISACAPI Deployment Files

This directory contains files for deploying ISACAPI on Linux servers and Apple-silicon Macs.

## Deployment Methods

| Method | Best For | Setup Wizard |
|--------|----------|--------------|
| **Docker Compose** | Quick setup, all-in-one | Not needed (auto-setup) |
| **Apple container** | Native local stack on macOS 26 | Not needed (auto-setup) |
| **Binary Install** | Production servers, systemd | Web-based wizard |

## Files

| File | Description |
|------|-------------|
| `docker-compose.yml` | Docker Compose configuration (named volumes) |
| `docker-compose.local.yml` | Docker Compose configuration (local directories, easy migration) |
| `docker-deploy.sh` | **One-click Docker deployment script (recommended)** |
| `apple-container.sh` | Native Apple `container` lifecycle script |
| `APPLE_CONTAINER.md` | Apple `container` deployment and operations guide |
| `.env.example` | Container environment variables template |
| `DOCKER.md` | Docker Hub documentation |
| `install.sh` | One-click binary installation script |
| `install-datamanagementd.sh` | datamanagementd 一键安装脚本 |
| `sub2api.service` | Systemd service unit file |
| `sub2api-datamanagementd.service` | datamanagementd systemd service unit file |
| `DATAMANAGEMENTD_CN.md` | datamanagementd 部署与联动说明（中文） |
| `config.example.yaml` | Example configuration file |
| `EDGE_SECURITY.md` | Reverse proxy, CDN/WAF, trusted proxy, and ingress hardening guide |

---

## Apple container Deployment

Apple-silicon Macs running macOS 26 can run the complete ISACAPI, PostgreSQL, and Redis stack with Apple `container` 1.1.0 or newer:

```bash
IMAGE_TAG=X.Y.Z ./apple-container.sh init
./apple-container.sh up
./apple-container.sh status
./apple-container.sh logs app -f
```

The script uses Apple named volumes, starts dependencies in order, and performs live readiness checks. It does not provide a continuous restart supervisor; run `./apple-container.sh up` after a host reboot. Docker Compose remains the recommended production deployment path.

See [APPLE_CONTAINER.md](./APPLE_CONTAINER.md) for configuration, upgrades, persistence, networking behavior, and limitations.

---

## Docker Deployment (Recommended)

### Method 1: One-Click Deployment (Recommended)

Use the automated preparation script for the easiest setup:

```bash
# Pin both deployment assets and the container image to one release.
# Replace X.Y.Z with a reviewed published release.
ISACAPI_VERSION=X.Y.Z
curl -fsSL "https://raw.githubusercontent.com/chiellini/ISACAPI/v${ISACAPI_VERSION}/deploy/docker-deploy.sh" \
  | IMAGE_TAG="${ISACAPI_VERSION}" bash

# Or download first, then run
curl -fsSL "https://raw.githubusercontent.com/chiellini/ISACAPI/v${ISACAPI_VERSION}/deploy/docker-deploy.sh" -o docker-deploy.sh
chmod +x docker-deploy.sh
IMAGE_TAG="${ISACAPI_VERSION}" ./docker-deploy.sh
```

**What the script does:**
- Downloads `docker-compose.local.yml` and `.env.example`
- Automatically generates secure secrets (JWT_SECRET, TOTP_ENCRYPTION_KEY, POSTGRES_PASSWORD, ADMIN_PASSWORD)
- Creates `.env` file with generated secrets
- Creates necessary data directories (data/, postgres_data/, redis_data/)
- Stores generated credentials only in `.env` with mode `0600`; secret values are not printed
- On upgrade, creates a private `.isacapi-deploy-backup-*` directory and
  preserves the complete existing `.env`; only `IMAGE_TAG` is updated

**After running the script:**
```bash
# Start services (the script saves the local-directory variant as docker-compose.yml)
docker compose up -d

# View logs
docker compose logs -f sub2api

# Read the generated admin password locally, then rotate it after first login:
grep '^ADMIN_PASSWORD=' .env

# Access Web UI
# http://localhost:8080
```

### Method 2: Manual Deployment

If you prefer manual control:

```bash
# Clone repository
git clone https://github.com/chiellini/ISACAPI.git
cd ISACAPI/deploy

# Configure environment
cp .env.example .env
ISACAPI_VERSION=X.Y.Z
POSTGRES_PASSWORD=$(openssl rand -hex 32)
ADMIN_PASSWORD=$(openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 32)
TOTP_ENCRYPTION_KEY=$(openssl rand -hex 32)
sed -i "s/^IMAGE_TAG=.*/IMAGE_TAG=${ISACAPI_VERSION}/" .env
sed -i "s/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" .env
sed -i "s/^ADMIN_PASSWORD=.*/ADMIN_PASSWORD=${ADMIN_PASSWORD}/" .env
sed -i "s/^JWT_SECRET=.*/JWT_SECRET=${JWT_SECRET}/" .env
sed -i "s/^TOTP_ENCRYPTION_KEY=.*/TOTP_ENCRYPTION_KEY=${TOTP_ENCRYPTION_KEY}/" .env
chmod 600 .env

# Create data directories
mkdir -p data postgres_data redis_data

# Start all services using local directory version
docker compose -f docker-compose.local.yml up -d

# View logs
docker compose -f docker-compose.local.yml logs -f sub2api

# Access Web UI
# http://localhost:8080
```

### Deployment Version Comparison

| Version | Data Storage | Migration | Best For |
|---------|-------------|-----------|----------|
| **docker-compose.local.yml** | Local directories (./data, ./postgres_data, ./redis_data) | ✅ Easy (tar entire directory) | Production, need frequent backups/migration |
| **docker-compose.yml** | Named volumes (/var/lib/docker/volumes/) | ⚠️ Requires docker commands | Simple setup, don't need migration |

**Recommendation:** Use `docker-compose.local.yml` (deployed by `docker-deploy.sh`) for easier data management and migration.

### How Auto-Setup Works

When using Docker Compose with `AUTO_SETUP=true`:

1. On first run, the system automatically:
   - Connects to PostgreSQL and Redis
   - Applies database migrations (SQL files in `backend/migrations/*.sql`) and records them in `schema_migrations`
   - Creates the administrator account from the credentials in `.env`
   - Writes config.yaml

2. No manual Setup Wizard is needed. Before the first start, set an exact
   `IMAGE_TAG`, `POSTGRES_PASSWORD`, `ADMIN_PASSWORD`, `JWT_SECRET`, and
   `TOTP_ENCRYPTION_KEY`. The preparation scripts generate the secrets without
   printing their values.

3. Read the generated administrator password only from the local mode-`0600`
   `.env` file, then rotate it after first login.

### Safe Upgrade

Re-run the script for the reviewed target release. `DEPLOY_OVERWRITE=1` is
appropriate for unattended upgrades after review; without it, the script asks
for confirmation. Existing PostgreSQL, JWT, TOTP, and administrator credentials
are copied to a private backup and preserved unchanged.

```bash
ISACAPI_VERSION=X.Y.Z
curl -fsSL "https://raw.githubusercontent.com/chiellini/ISACAPI/v${ISACAPI_VERSION}/deploy/docker-deploy.sh" \
  | DEPLOY_OVERWRITE=1 IMAGE_TAG="${ISACAPI_VERSION}" bash
docker compose pull
docker compose up -d
```

If `docker-compose.yml` exists but `.env` is missing, the script stops without
changing files. Restore the original `.env`; generating replacements can make
an existing database inaccessible and invalidate sessions or encrypted TOTP
data. An `.env`-only partial deployment is recoverable: the script backs it up,
preserves it, and restores the Compose file.

### Database Migration Notes (PostgreSQL)

- Migrations are applied in lexicographic order (e.g. `001_...sql`, `002_...sql`).
- `schema_migrations` tracks applied migrations (filename + checksum).
- Migrations are forward-only; rollback requires a DB backup restore or a manual compensating SQL script.

**Verify `users.allowed_groups` → `user_allowed_groups` backfill**

During the incremental GORM→Ent migration, `users.allowed_groups` (legacy `BIGINT[]`) is being replaced by a normalized join table `user_allowed_groups(user_id, group_id)`.

Run this query to compare the legacy data vs the join table:

```sql
WITH old_pairs AS (
  SELECT DISTINCT u.id AS user_id, x.group_id
  FROM users u
  CROSS JOIN LATERAL unnest(u.allowed_groups) AS x(group_id)
  WHERE u.allowed_groups IS NOT NULL
)
SELECT
  (SELECT COUNT(*) FROM old_pairs)           AS old_pair_count,
  (SELECT COUNT(*) FROM user_allowed_groups) AS new_pair_count;
```

### datamanagementd（数据管理）联动

如需启用管理后台“数据管理”功能，请额外部署宿主机 `datamanagementd`：

- 主进程固定探测 `/tmp/sub2api-datamanagement.sock`
- Docker 场景下需把宿主机 Socket 挂载到容器内同路径
- 详细步骤见：`deploy/DATAMANAGEMENTD_CN.md`

### Commands

For **local directory version** (docker-compose.local.yml):

```bash
# Start services
docker compose -f docker-compose.local.yml up -d

# Stop services
docker compose -f docker-compose.local.yml down

# View logs
docker compose -f docker-compose.local.yml logs -f sub2api

# Restart ISACAPI only
docker compose -f docker-compose.local.yml restart sub2api

# Set an audited release explicitly, then pull and recreate
sed -i 's/^IMAGE_TAG=.*/IMAGE_TAG=X.Y.Z/' .env
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d

# Remove all data (caution!)
docker compose -f docker-compose.local.yml down
rm -rf data/ postgres_data/ redis_data/
```

For **named volumes version** (docker-compose.yml):

```bash
# Start services
docker compose up -d

# Stop services
docker compose down

# View logs
docker compose logs -f sub2api

# Restart ISACAPI only
docker compose restart sub2api

# Set an audited release explicitly, then pull and recreate
sed -i 's/^IMAGE_TAG=.*/IMAGE_TAG=X.Y.Z/' .env
docker compose pull
docker compose up -d

# Remove all data (caution!)
docker compose down -v
```

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `POSTGRES_PASSWORD` | **Yes** | - | PostgreSQL password |
| `IMAGE_TAG` | **Yes** | - | Exact published application version; floating tags are rejected by preparation scripts |
| `JWT_SECRET` | **Yes** | - | JWT secret (fixed for persistent sessions) |
| `TOTP_ENCRYPTION_KEY` | **Yes** | - | TOTP encryption key (fixed for persistent 2FA) |
| `SERVER_PORT` | No | `8080` | Server port |
| `ADMIN_EMAIL` | No | `admin@sub2api.local` | Admin email |
| `ADMIN_PASSWORD` | **Yes for first start** | - | Initial administrator password; rotate after first login |
| `TZ` | No | `Asia/Shanghai` | Timezone |
| `UPDATE_GITHUB_TOKEN` | No | *(empty)* | Token for `api.github.com` release checks only; asset downloads remain anonymous. |
| `GEMINI_OAUTH_CLIENT_ID` | No | *(builtin)* | Google OAuth client ID (Gemini OAuth). Leave empty to use the built-in Gemini CLI client. |
| `GEMINI_OAUTH_CLIENT_SECRET` | No | *(builtin)* | Google OAuth client secret (Gemini OAuth). Leave empty to use the built-in Gemini CLI client. |
| `GEMINI_OAUTH_SCOPES` | No | *(default)* | OAuth scopes (Gemini OAuth) |
| `GEMINI_QUOTA_POLICY` | No | *(empty)* | JSON overrides for Gemini local quota simulation (Code Assist only). |

See `.env.example` for all available options.

> **Note:** The `docker-deploy.sh` script generates `JWT_SECRET`, `TOTP_ENCRYPTION_KEY`, `POSTGRES_PASSWORD`, and `ADMIN_PASSWORD` into the mode-`0600` `.env` file without printing their values.

The URL allowlist contains only reviewed official endpoints. For a custom
upstream, explicitly append its exact hostname (or a reviewed wildcard suffix)
to `security.url_allowlist.upstream_hosts`. Private/local upstreams additionally
require `security.url_allowlist.allow_private_hosts: true` and should only be
used on a trusted network. Account-configured HTTP/SOCKS proxies remain blocked
by default. After auditing a proxy, opt in separately with
`security.url_allowlist.trust_upstream_proxy: true`; do not use
`allow_private_hosts` as a proxy-trust switch. URL host, scheme, and redirect
policy still applies, while DNS pinning beyond the trusted proxy is necessarily
delegated to that proxy.

### Easy Migration (Local Directory Version)

When using `docker-compose.local.yml`, all data is stored in local directories, making migration simple:

```bash
# On source server: Stop services and create archive
cd /path/to/deployment
docker compose -f docker-compose.local.yml down
cd ..
tar czf sub2api-complete.tar.gz deployment/

# Transfer to new server
scp sub2api-complete.tar.gz user@new-server:/path/to/destination/

# On new server: Extract and start
tar xzf sub2api-complete.tar.gz
cd deployment/
docker compose -f docker-compose.local.yml up -d
```

Your entire deployment (configuration + data) is migrated!

---

## Gemini OAuth Configuration

ISACAPI supports three methods to connect to Gemini:

### Method 1: Code Assist OAuth (Recommended for GCP Users)

**No configuration needed** - always uses the built-in Gemini CLI OAuth client (public).

1. Leave `GEMINI_OAUTH_CLIENT_ID` and `GEMINI_OAUTH_CLIENT_SECRET` empty
2. In the Admin UI, create a Gemini OAuth account and select **"Code Assist"** type
3. Complete the OAuth flow in your browser

> Note: Even if you configure `GEMINI_OAUTH_CLIENT_ID` / `GEMINI_OAUTH_CLIENT_SECRET` for AI Studio OAuth,
> Code Assist OAuth will still use the built-in Gemini CLI client.

**Requirements:**
- Google account with access to Google Cloud Platform
- A GCP project (auto-detected or manually specified)

**How to get Project ID (if auto-detection fails):**
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click the project dropdown at the top of the page
3. Copy the Project ID (not the project name) from the list
4. Common formats: `my-project-123456` or `cloud-ai-companion-xxxxx`

### Method 2: AI Studio OAuth (For Regular Google Accounts)

Requires your own OAuth client credentials.

**Step 1: Create OAuth Client in Google Cloud Console**

1. Go to [Google Cloud Console - Credentials](https://console.cloud.google.com/apis/credentials)
2. Create a new project or select an existing one
3. **Enable the Generative Language API:**
   - Go to "APIs & Services" → "Library"
   - Search for "Generative Language API"
   - Click "Enable"
4. **Configure OAuth Consent Screen** (if not done):
   - Go to "APIs & Services" → "OAuth consent screen"
   - Choose "External" user type
   - Fill in app name, user support email, developer contact
   - Add scopes: `https://www.googleapis.com/auth/generative-language.retriever` (and optionally `https://www.googleapis.com/auth/cloud-platform`)
   - Add test users (your Google account email)
5. **Create OAuth 2.0 credentials:**
   - Go to "APIs & Services" → "Credentials"
   - Click "Create Credentials" → "OAuth client ID"
   - Application type: **Web application** (or **Desktop app**)
   - Name: e.g., "ISACAPI Gemini"
   - Authorized redirect URIs: Add `http://localhost:1455/auth/callback`
6. Copy the **Client ID** and **Client Secret**
7. **⚠️ Publish to Production (IMPORTANT):**
   - Go to "APIs & Services" → "OAuth consent screen"
   - Click "PUBLISH APP" to move from Testing to Production
   - **Testing mode limitations:**
     - Only manually added test users can authenticate (max 100 users)
     - Refresh tokens expire after 7 days
     - Users must be re-added periodically
   - **Production mode:** Any Google user can authenticate, tokens don't expire
   - Note: For sensitive scopes, Google may require verification (demo video, privacy policy)

**Step 2: Configure Environment Variables**

```bash
GEMINI_OAUTH_CLIENT_ID=your-client-id.apps.googleusercontent.com
GEMINI_OAUTH_CLIENT_SECRET=GOCSPX-your-client-secret

# 可选：如需使用 Gemini CLI 内置 OAuth Client（Code Assist / Google One）
# 安全说明：本仓库不会内置该 client_secret，请在运行环境通过环境变量注入。
# GEMINI_CLI_OAUTH_CLIENT_SECRET=GOCSPX-your-built-in-secret
```

**Step 3: Create Account in Admin UI**

1. Create a Gemini OAuth account and select **"AI Studio"** type
2. Complete the OAuth flow
   - After consent, your browser will be redirected to `http://localhost:1455/auth/callback?code=...&state=...`
   - Copy the full callback URL (recommended) or just the `code` and paste it back into the Admin UI

### Method 3: API Key (Simplest)

1. Go to [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Click "Create API key"
3. In Admin UI, create a Gemini **API Key** account
4. Paste your API key (starts with `AIza...`)

### Comparison Table

| Feature | Code Assist OAuth | AI Studio OAuth | API Key |
|---------|-------------------|-----------------|---------|
| Setup Complexity | Easy (no config) | Medium (OAuth client) | Easy |
| GCP Project Required | Yes | No | No |
| Custom OAuth Client | No (built-in) | Yes (required) | N/A |
| Rate Limits | GCP quota | Standard | Standard |
| Best For | GCP developers | Regular users needing OAuth | Quick testing |

---

## Binary Installation

For production servers using systemd.

### One-Line Installation

```bash
# Replace X.Y.Z with a reviewed published release.
ISACAPI_VERSION=X.Y.Z
curl -fsSL "https://raw.githubusercontent.com/chiellini/ISACAPI/v${ISACAPI_VERSION}/deploy/install.sh" \
  | sudo bash -s -- install --version "v${ISACAPI_VERSION}"
```

The service binds to `127.0.0.1:8080` by default so an unauthenticated first-run
wizard is not exposed publicly. From your workstation, create an SSH tunnel and
open the wizard locally:

```bash
ssh -N -L 8080:127.0.0.1:8080 <user>@<server>
# Open http://127.0.0.1:8080 in your browser.

# Read the owner-only one-time token on the server and paste it into the wizard.
sudo cat /opt/sub2api/.setup-token
```

Only change `SERVER_HOST` to `0.0.0.0` after placing the origin behind a trusted
TLS reverse proxy/firewall and configuring exact trusted proxy addresses.

### Manual Installation

1. Download an explicit release from [GitHub Releases](https://github.com/chiellini/ISACAPI/releases)
2. Extract and copy the binary to `/opt/sub2api/`
3. Copy `sub2api.service` to `/etc/systemd/system/`
4. Run:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable sub2api
   sudo systemctl start sub2api
   ```
5. Read `/opt/sub2api/.setup-token` as root, open the Setup Wizard, and paste
   the one-time token to complete configuration. The token file is removed
   after a successful installation.

### Commands

```bash
# Install
sudo ./install.sh

# Upgrade
sudo ./install.sh upgrade

# Uninstall
sudo ./install.sh uninstall
```

### Service Management

```bash
# Start the service
sudo systemctl start sub2api

# Stop the service
sudo systemctl stop sub2api

# Restart the service
sudo systemctl restart sub2api

# Check status
sudo systemctl status sub2api

# View logs
sudo journalctl -u sub2api -f

# Enable auto-start on boot
sudo systemctl enable sub2api
```

### Configuration

#### Server Address and Port

During installation, you will be prompted to configure the server listen address and port. These settings are stored in the systemd service file as environment variables.

To change after installation:

1. Edit the systemd service:
   ```bash
   sudo systemctl edit sub2api
   ```

2. Add or modify:
   ```ini
   [Service]
   Environment=SERVER_HOST=127.0.0.1
   Environment=SERVER_PORT=3000
   Environment=SETUP_HOST=127.0.0.1
   Environment=SETUP_PORT=3000
   ```

3. Reload and restart:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl restart sub2api
   ```

#### Gemini OAuth Configuration

If you need to use AI Studio OAuth for Gemini accounts, add the OAuth client credentials to the systemd service file:

1. Edit the service file:
   ```bash
   sudo nano /etc/systemd/system/sub2api.service
   ```

2. Add your OAuth credentials in the `[Service]` section (after the existing `Environment=` lines):
   ```ini
   Environment=GEMINI_OAUTH_CLIENT_ID=your-client-id.apps.googleusercontent.com
   Environment=GEMINI_OAUTH_CLIENT_SECRET=GOCSPX-your-client-secret
   ```

   如需使用“内置 Gemini CLI OAuth Client”（Code Assist / Google One），还需要注入：
   ```ini
   Environment=GEMINI_CLI_OAUTH_CLIENT_SECRET=GOCSPX-your-built-in-secret
   ```

3. Reload and restart:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl restart sub2api
   ```

> **Note:** Code Assist OAuth does not require any configuration - it uses the built-in Gemini CLI client.
> See the [Gemini OAuth Configuration](#gemini-oauth-configuration) section above for detailed setup instructions.

#### Application Configuration

The main config file is at `/etc/sub2api/config.yaml` (created by Setup Wizard).

### Prerequisites

- Linux server (Ubuntu 20.04+, Debian 11+, CentOS 8+, etc.)
- PostgreSQL 14+
- Redis 6+
- systemd

### Directory Structure

```
/opt/sub2api/
├── sub2api              # Main binary
├── sub2api.backup       # Backup (after upgrade)
└── data/                # Runtime data

/etc/sub2api/
└── config.yaml          # Configuration file
```

---

## Troubleshooting

### Docker

For **local directory version**:

```bash
# Check container status
docker compose -f docker-compose.local.yml ps

# View detailed logs
docker compose -f docker-compose.local.yml logs --tail=100 sub2api

# Check database connection
docker compose -f docker-compose.local.yml exec postgres pg_isready

# Check Redis connection
docker compose -f docker-compose.local.yml exec redis redis-cli ping

# Restart all services
docker compose -f docker-compose.local.yml restart

# Check data directories
ls -la data/ postgres_data/ redis_data/
```

For **named volumes version**:

```bash
# Check container status
docker compose ps

# View detailed logs
docker compose logs --tail=100 sub2api

# Check database connection
docker compose exec postgres pg_isready

# Check Redis connection
docker compose exec redis redis-cli ping

# Restart all services
docker compose restart
```

### Binary Install

```bash
# Check service status
sudo systemctl status sub2api

# View recent logs
sudo journalctl -u sub2api -n 50

# Check config file
sudo cat /etc/sub2api/config.yaml

# Check PostgreSQL
sudo systemctl status postgresql

# Check Redis
sudo systemctl status redis
```

### Common Issues

1. **Port already in use**: Change `SERVER_PORT` in `.env` or systemd config
2. **Database connection failed**: Check PostgreSQL is running and credentials are correct
3. **Redis connection failed**: Check Redis is running and password is correct
4. **Permission denied**: Ensure proper file ownership for binary install

---

## TLS Fingerprint Configuration

ISACAPI supports TLS fingerprint simulation to make requests appear as if they come from the official Claude CLI (Node.js client).

> **💡 Tip:** Visit **[tls.sub2api.org](https://tls.sub2api.org/)** to get TLS fingerprint information for different devices and browsers.

### Default Behavior

- Built-in `claude_cli_v2` profile simulates Node.js 20.x + OpenSSL 3.x
- JA3 Hash: `1a28e69016765d92e3b381168d68922c`
- JA4: `t13d5911h1_a33745022dd6_1f22a2ca17c4`
- Profile selection: `accountID % profileCount`

### Configuration

```yaml
gateway:
  tls_fingerprint:
    enabled: true  # Global switch
    profiles:
      # Simple profile (uses default cipher suites)
      profile_1:
        name: "Profile 1"

      # Profile with custom cipher suites (use compact array format)
      profile_2:
        name: "Profile 2"
        cipher_suites: [4866, 4867, 4865, 49199, 49195, 49200, 49196]
        curves: [29, 23, 24]
        point_formats: 0

      # Another custom profile
      profile_3:
        name: "Profile 3"
        cipher_suites: [4865, 4866, 4867, 49199, 49200]
        curves: [29, 23, 24, 25]
```

### Profile Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name (required) |
| `cipher_suites` | []uint16 | Cipher suites in decimal. Empty = default |
| `curves` | []uint16 | Elliptic curves in decimal. Empty = default |
| `point_formats` | []uint8 | EC point formats. Empty = default |

### Common Values Reference

**Cipher Suites (TLS 1.3):** `4865` (AES_128_GCM), `4866` (AES_256_GCM), `4867` (CHACHA20)

**Cipher Suites (TLS 1.2):** `49195`, `49196`, `49199`, `49200` (ECDHE variants)

**Curves:** `29` (X25519), `23` (P-256), `24` (P-384), `25` (P-521)
