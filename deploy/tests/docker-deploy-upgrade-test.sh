#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
test_root=$(mktemp -d)

cleanup() {
  case "$test_root" in
    "${TMPDIR:-/tmp}"/*|/tmp/*) rm -rf -- "$test_root" ;;
    *) printf 'refusing to remove unexpected test path: %s\n' "$test_root" >&2 ;;
  esac
}
trap cleanup EXIT HUP INT TERM

fail() {
  printf 'docker deploy upgrade test failed: %s\n' "$1" >&2
  exit 1
}

mkdir -p "$test_root/assets" "$test_root/bin"
cp "$repo_root/deploy/docker-compose.local.yml" "$test_root/assets/docker-compose.local.yml"
cp "$repo_root/deploy/.env.example" "$test_root/assets/.env.example"

cat >"$test_root/bin/curl" <<'EOF'
#!/bin/sh
set -eu
url=
destination=
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o)
      shift
      destination=$1
      ;;
    http://*|https://*) url=$1 ;;
  esac
  shift
done
[ -n "$destination" ] && [ -n "$url" ]
case "$url" in
  */docker-compose.local.yml) source_file="$DEPLOY_TEST_ASSETS/docker-compose.local.yml" ;;
  */.env.example) source_file="$DEPLOY_TEST_ASSETS/.env.example" ;;
  *) exit 64 ;;
esac
cp "$source_file" "$destination"
EOF

cat >"$test_root/bin/openssl" <<'EOF'
#!/bin/sh
set -eu
if [ "${DEPLOY_TEST_OPENSSL_MUST_NOT_RUN:-0}" = 1 ]; then
  exit 70
fi
printf '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\n'
EOF
chmod +x "$test_root/bin/curl" "$test_root/bin/openssl"

run_deploy() {
  target=$1
  must_not_generate=${2:-0}
  (
    cd "$target"
    PATH="$test_root/bin:$PATH" \
      DEPLOY_TEST_ASSETS="$test_root/assets" \
      DEPLOY_TEST_OPENSSL_MUST_NOT_RUN="$must_not_generate" \
      DEPLOY_OVERWRITE=1 IMAGE_TAG=9.8.7 \
      bash "$repo_root/deploy/docker-deploy.sh" >/dev/null
  )
}

assert_preserved_upgrade() {
  target=$1
  expected_env=$2
  grep -Fqx 'POSTGRES_PASSWORD=postgres-keep' "$target/.env" || fail 'PostgreSQL secret changed during upgrade'
  grep -Fqx 'JWT_SECRET=jwt-keep' "$target/.env" || fail 'JWT secret changed during upgrade'
  grep -Fqx 'TOTP_ENCRYPTION_KEY=totp-keep' "$target/.env" || fail 'TOTP key changed during upgrade'
  grep -Fqx 'ADMIN_PASSWORD=admin-keep' "$target/.env" || fail 'administrator secret changed during upgrade'
  grep -Fqx 'IMAGE_TAG=9.8.7' "$target/.env" || fail 'requested image tag was not persisted'

  set -- "$target"/.isacapi-deploy-backup-*
  [ "$#" -eq 1 ] && [ -d "$1" ] || fail 'upgrade did not create exactly one backup directory'
  cmp "$expected_env" "$1/.env" >/dev/null || fail 'backup does not contain the original .env'
  [ "$(find "$1" -maxdepth 1 -type f | wc -l | tr -d ' ')" -ge 1 ] || fail 'backup is empty'
}

make_existing_env() {
  destination=$1
  cat >"$destination" <<'EOF'
IMAGE_TAG=1.2.3
POSTGRES_PASSWORD=postgres-keep
JWT_SECRET=jwt-keep
TOTP_ENCRYPTION_KEY=totp-keep
ADMIN_PASSWORD=admin-keep
EOF
  chmod 600 "$destination"
}

# Normal upgrade: both managed files exist. Secrets must be preserved and both
# original files must be recoverable from a private backup directory.
mkdir "$test_root/both"
make_existing_env "$test_root/both/.env"
cp "$test_root/both/.env" "$test_root/both/original.env"
printf 'old compose\n' >"$test_root/both/docker-compose.yml"
run_deploy "$test_root/both" 1
assert_preserved_upgrade "$test_root/both" "$test_root/both/original.env"
set -- "$test_root/both"/.isacapi-deploy-backup-*
grep -Fqx 'old compose' "$1/docker-compose.yml" || fail 'original Compose file was not backed up'

# Recoverable partial state: .env exists but Compose is absent. Preserve the
# environment and install the selected Compose asset.
mkdir "$test_root/env-only"
make_existing_env "$test_root/env-only/.env"
cp "$test_root/env-only/.env" "$test_root/env-only/original.env"
run_deploy "$test_root/env-only" 1
assert_preserved_upgrade "$test_root/env-only" "$test_root/env-only/original.env"
[ -f "$test_root/env-only/docker-compose.yml" ] || fail 'Compose was not restored for env-only state'

# Unsafe partial state: Compose exists but its secret-bearing .env is gone.
# Fail before download or secret generation and leave the file untouched.
mkdir "$test_root/compose-only"
printf 'orphan compose\n' >"$test_root/compose-only/docker-compose.yml"
if run_deploy "$test_root/compose-only" 1 2>"$test_root/compose-only/error.log"; then
  fail 'compose-only state was accepted'
fi
grep -Fq 'Restore the original .env' "$test_root/compose-only/error.log" || fail 'compose-only failure lacks recovery guidance'
grep -Fqx 'orphan compose' "$test_root/compose-only/docker-compose.yml" || fail 'compose-only failure changed the existing file'
[ ! -e "$test_root/compose-only/.env" ] || fail 'compose-only failure generated replacement secrets'

# Fresh install still generates all first-run credentials, including the admin
# password required by AUTO_SETUP.
mkdir "$test_root/fresh"
run_deploy "$test_root/fresh" 0
for key in POSTGRES_PASSWORD JWT_SECRET TOTP_ENCRYPTION_KEY ADMIN_PASSWORD; do
  grep -Eq "^${key}=.+$" "$test_root/fresh/.env" || fail "fresh install did not generate $key"
done
grep -Fqx 'IMAGE_TAG=9.8.7' "$test_root/fresh/.env" || fail 'fresh install did not pin IMAGE_TAG'
[ "$(stat -c '%a' "$test_root/fresh/.env")" = 600 ] || fail 'fresh .env is not mode 0600'

printf 'docker deploy upgrade test passed\n'
