#!/usr/bin/env bash
# =============================================================================
# Sub2API / ISACAPI one-command rsync migration
#
# Run this script on the old server. By default it only performs read-only
# checks. Pass --execute to perform the migration.
# =============================================================================
set -Eeuo pipefail

umask 077

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
readonly DEPLOY_DIR="$(cd -- "$SCRIPT_DIR/.." && pwd -P)"
readonly PROJECT_ROOT="$(cd -- "$DEPLOY_DIR/.." && pwd -P)"
readonly LOCAL_COMPOSE_FILE="$DEPLOY_DIR/docker-compose.local.yml"
readonly LOCAL_BUILD_FILE="$DEPLOY_DIR/docker-compose.build.yml"

HOST=""
REMOTE_USER="ec2-user"
SSH_PORT="22"
IDENTITY_FILE=""
DEST_DIR=""
MODE="dry-run"
ASSUME_YES=false
ACCEPT_NEW_HOST_KEY=false

TMP_DIR=""
SSH_WRAPPER=""
RSYNC_RSH=""
LOCAL_IMAGE_BUNDLE=""
REMOTE_IMAGE_BUNDLE=""
REMOTE_LOCK_DIR=""
STAGE_BASE=""
STAGE_DIR=""
SSH_REMOTE=""
RSYNC_REMOTE=""
REMOTE_ROOT=""
REMOTE_RSYNC_PATH=""

SOURCE_STOPPED=false
TARGET_MAY_BE_RUNNING=false
TARGET_FINALIZED=false
REMOTE_PREPARED=false
REMOTE_BUNDLE_TOUCHED=false
REMOTE_LOCK_OWNED=false
MIGRATION_SUCCEEDED=false
SUDO_KEEPALIVE_PID=""

declare -a LOCAL_ROOT=()
declare -a COMPOSE=()
declare -a IMAGE_REFS=()
declare -a IMAGE_IDS=()
declare -a COMPOSE_NETWORKS=()
declare -a CRITICAL_PATHS=()

usage() {
  cat <<'EOF'
用法：
  bash deploy/migration/rsync-migrate.sh --host <新机IP或主机名> [选项]

必填：
  --host HOST             新服务器 IP 或主机名

SSH / 目标选项：
  --user USER             SSH 用户，默认 ec2-user
  --port PORT             SSH 端口，默认 22
  --identity FILE         SSH 私钥绝对路径；也可使用当前 ssh-agent/SSH 配置
  --dest DIR              新机上的项目绝对路径，默认 <远端HOME>/sub2api
  --accept-new-host-key   首次连接时接受并记录新机 SSH host key（默认拒绝未知 key）

执行模式：
  --dry-run               只做预检（默认，不停服、不写远端）
  --execute               执行迁移
  --yes                   跳过执行前的确认提示，适合无人值守
  -h, --help              显示帮助

示例：
  # 1. 只预检
  bash deploy/migration/rsync-migrate.sh \
    --host 203.0.113.10 --user ec2-user --identity /home/ec2-user/new.pem

  # 2. 确认预检无误后执行
  bash deploy/migration/rsync-migrate.sh \
    --host 203.0.113.10 --user ec2-user --identity /home/ec2-user/new.pem \
    --execute

前提：
  - 新机已安装 bash、rsync、gzip、Docker 和 Docker Compose plugin。
  - SSH 用户是 root，或拥有无需交互密码的 sudo 权限。
  - 源机与新机 CPU 架构相同。
  - 默认要求新机 host key 已在 known_hosts 中；请先手动 ssh 核对指纹，
    或明确传 --accept-new-host-key 使用 TOFU。
  - 迁移完成前，不要把 DNS、反向代理或其他业务流量指向新机。

脚本会迁移项目文件、PostgreSQL、Redis、data、.env，并传输当前实际
运行的三个 Docker 镜像。不会迁移 /etc/nginx、/etc/letsencrypt、DNS、
安全组、EIP 或宿主机定时任务。
EOF
}

log() {
  printf '[INFO] %s\n' "$*"
}

ok() {
  printf '[ OK ] %s\n' "$*"
}

warn() {
  printf '[WARN] %s\n' "$*" >&2
}

die() {
  printf '[ERR ] %s\n' "$*" >&2
  exit 1
}

quote() {
  printf '%q' "$1"
}

human_bytes() {
  local bytes=$1
  awk -v bytes="$bytes" 'BEGIN {
    split("B KiB MiB GiB TiB", unit, " ")
    value = bytes + 0
    unit_index = 1
    while (value >= 1024 && unit_index < 5) {
      value /= 1024
      unit_index++
    }
    printf "%.1f %s", value, unit[unit_index]
  }'
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "本机缺少命令: $1"
}

normalize_arch() {
  case "$1" in
    x86_64|amd64) printf 'amd64\n' ;;
    aarch64|arm64) printf 'arm64\n' ;;
    *) printf '%s\n' "$1" ;;
  esac
}

while (($# > 0)); do
  case "$1" in
    --host)
      (($# >= 2)) || die "--host 缺少参数"
      HOST=$2
      shift 2
      ;;
    --user)
      (($# >= 2)) || die "--user 缺少参数"
      REMOTE_USER=$2
      shift 2
      ;;
    --port)
      (($# >= 2)) || die "--port 缺少参数"
      SSH_PORT=$2
      shift 2
      ;;
    --identity)
      (($# >= 2)) || die "--identity 缺少参数"
      IDENTITY_FILE=$2
      shift 2
      ;;
    --dest)
      (($# >= 2)) || die "--dest 缺少参数"
      DEST_DIR=$2
      shift 2
      ;;
    --dry-run)
      MODE="dry-run"
      shift
      ;;
    --execute)
      MODE="execute"
      shift
      ;;
    --yes)
      ASSUME_YES=true
      shift
      ;;
    --accept-new-host-key)
      ACCEPT_NEW_HOST_KEY=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "未知参数: $1（用 --help 查看帮助）"
      ;;
  esac
done

[[ -n "$HOST" ]] || die "必须提供 --host"
[[ "$REMOTE_USER" =~ ^[A-Za-z_][A-Za-z0-9._-]*$ ]] || die "--user 格式不合法"
[[ "$SSH_PORT" =~ ^[0-9]{1,5}$ ]] || die "--port 必须是 1-5 位数字"
SSH_PORT=$((10#$SSH_PORT))
((SSH_PORT >= 1 && SSH_PORT <= 65535)) || die "--port 超出范围"
[[ "$HOST" != *[[:space:]]* ]] || die "--host 不能包含空白字符"

if [[ -n "$IDENTITY_FILE" ]]; then
  [[ "$IDENTITY_FILE" = /* ]] || die "--identity 必须使用绝对路径"
  [[ -r "$IDENTITY_FILE" ]] || die "无法读取 SSH 私钥: $IDENTITY_FILE"
fi

require_command bash
require_command docker
require_command gzip
require_command flock
require_command find
require_command readlink
require_command rsync
require_command seq
require_command ssh
require_command sudo
require_command timeout

[[ -f "$DEPLOY_DIR/.env" ]] || die "缺少 $DEPLOY_DIR/.env"
[[ -f "$LOCAL_COMPOSE_FILE" ]] || die "缺少 $LOCAL_COMPOSE_FILE"
[[ -f "$LOCAL_BUILD_FILE" ]] || die "缺少 $LOCAL_BUILD_FILE"
[[ -d "$DEPLOY_DIR/data" ]] || die "缺少 $DEPLOY_DIR/data"
[[ -d "$DEPLOY_DIR/postgres_data" ]] || die "缺少 $DEPLOY_DIR/postgres_data"
[[ -d "$DEPLOY_DIR/redis_data" ]] || die "缺少 $DEPLOY_DIR/redis_data"
CRITICAL_PATHS=(
  "$DEPLOY_DIR/.env" \
  "$LOCAL_COMPOSE_FILE" \
  "$LOCAL_BUILD_FILE" \
  "$DEPLOY_DIR/data" \
  "$DEPLOY_DIR/postgres_data" \
  "$DEPLOY_DIR/redis_data"
)
for critical_path in "${CRITICAL_PATHS[@]}"; do
  [[ ! -L "$critical_path" ]] \
    || die "关键迁移路径不能是符号链接: $critical_path"
done

exec {MIGRATION_LOCK_FD}<"$LOCAL_COMPOSE_FILE"
flock -n "$MIGRATION_LOCK_FD" \
  || die "已有另一个迁移进程正在使用此部署；拒绝并发执行"

if ((EUID == 0)); then
  LOCAL_ROOT=()
else
  log "请求本机 sudo 权限（只用于 Docker 与读取 PostgreSQL 数据目录）..."
  sudo -v
  LOCAL_ROOT=(sudo -n)
fi
COMPOSE=("${LOCAL_ROOT[@]}" docker compose -f "$LOCAL_COMPOSE_FILE" -f "$LOCAL_BUILD_FILE")

RUN_USER="${SUDO_USER:-$(id -un)}"
RUN_UID="$(id -u "$RUN_USER")"
RUN_HOME="$(getent passwd "$RUN_USER" | awk -F: '{print $6}')"
[[ -n "$RUN_HOME" ]] || RUN_HOME="/root"

TMP_DIR="$(mktemp -d "/tmp/sub2api-rsync-migrate.XXXXXX")"
SSH_WRAPPER="$TMP_DIR/ssh-wrapper"
LOCAL_IMAGE_BUNDLE="$TMP_DIR/sub2api-images.tar.gz"

create_ssh_wrapper() {
  local ssh_path sudo_path
  local -a ssh_options

  ssh_path="$(command -v ssh)"
  sudo_path="$(command -v sudo)"
  ssh_options=(
    -p "$SSH_PORT"
    -o BatchMode=yes
    -o ConnectTimeout=12
    -o ServerAliveInterval=15
    -o ServerAliveCountMax=4
  )
  if [[ "$ACCEPT_NEW_HOST_KEY" == true ]]; then
    ssh_options+=(-o StrictHostKeyChecking=accept-new)
  else
    ssh_options+=(-o StrictHostKeyChecking=yes)
  fi
  if [[ -n "$IDENTITY_FILE" ]]; then
    ssh_options+=(-o IdentitiesOnly=yes -i "$IDENTITY_FILE")
  fi

  {
    printf '#!/usr/bin/env bash\n'
    printf 'set -e\n'
    printf 'if [[ "$(id -u)" -eq %q ]]; then\n' "$RUN_UID"
    printf '  exec %q' "$ssh_path"
    printf ' %q' "${ssh_options[@]}"
    printf ' "$@"\n'
    printf 'else\n'
    printf '  exec %q -n -u %q env HOME=%q' "$sudo_path" "$RUN_USER" "$RUN_HOME"
    if [[ -n "${SSH_AUTH_SOCK:-}" ]]; then
      printf ' SSH_AUTH_SOCK=%q' "$SSH_AUTH_SOCK"
    fi
    printf ' %q' "$ssh_path"
    printf ' %q' "${ssh_options[@]}"
    printf ' "$@"\n'
    printf 'fi\n'
  } >"$SSH_WRAPPER"
  chmod 700 "$SSH_WRAPPER"
}

create_ssh_wrapper
RSYNC_RSH="bash $SSH_WRAPPER"

SSH_REMOTE="$REMOTE_USER@$HOST"
if [[ "$HOST" == *:* ]]; then
  RSYNC_REMOTE="[$REMOTE_USER@$HOST]"
else
  RSYNC_REMOTE="$SSH_REMOTE"
fi

remote_sh() {
  local escaped
  printf -v escaped '%q' "$1"
  bash "$SSH_WRAPPER" "$SSH_REMOTE" "bash -lc $escaped"
}

remote_root_command() {
  if [[ -n "$REMOTE_ROOT" ]]; then
    printf '%s %s' "$REMOTE_ROOT" "$1"
  else
    printf '%s' "$1"
  fi
}

start_sudo_keepalive() {
  local parent_pid

  if ((EUID == 0)); then
    return
  fi
  sudo -v
  parent_pid=$$
  (
    trap - EXIT INT TERM
    exec {MIGRATION_LOCK_FD}<&-
    while kill -0 "$parent_pid" >/dev/null 2>&1 && sudo -n -v >/dev/null 2>&1; do
      for _ in $(seq 1 30); do
        kill -0 "$parent_pid" >/dev/null 2>&1 || exit 0
        sleep 1
      done
    done
  ) &
  SUDO_KEEPALIVE_PID=$!
}

validate_mount() {
  local container=$1
  local destination=$2
  local expected=$3
  local actual actual_type

  "${LOCAL_ROOT[@]}" docker inspect "$container" >/dev/null 2>&1 \
    || die "找不到当前容器 $container；为避免误判数据目录，脚本拒绝继续"
  actual="$("${LOCAL_ROOT[@]}" docker inspect \
    --format "{{range .Mounts}}{{if eq .Destination \"$destination\"}}{{.Source}}{{end}}{{end}}" \
    "$container")"
  actual_type="$("${LOCAL_ROOT[@]}" docker inspect \
    --format "{{range .Mounts}}{{if eq .Destination \"$destination\"}}{{.Type}}{{end}}{{end}}" \
    "$container")"
  [[ -n "$actual" ]] || die "$container 没有挂载 $destination"
  [[ "$actual_type" == "bind" ]] || die "$container 的 $destination 不是 bind mount"
  [[ "$(readlink -m -- "$actual")" == "$(readlink -m -- "$expected")" ]] \
    || die "$container 的 $destination 实际来自 $actual，不是 $expected"
}

validate_mount_inventory() {
  local container=$1
  local type source destination unexpected mounts

  mounts="$("${LOCAL_ROOT[@]}" docker inspect --format \
    '{{range .Mounts}}{{printf "%s\t%s\t%s\n" .Type .Source .Destination}}{{end}}' "$container")" \
    || die "无法读取 $container 的完整挂载清单"
  while IFS=$'\t' read -r type source destination; do
    [[ -n "$destination" ]] || continue
    case "$container:$destination" in
      sub2api:/app/data)
        [[ "$type" == "bind" && "$(readlink -m -- "$source")" == "$DEPLOY_DIR/data" ]] \
          || die "sub2api /app/data 挂载不符合预期"
        ;;
      sub2api-postgres:/var/lib/postgresql/data)
        [[ "$type" == "bind" && "$(readlink -m -- "$source")" == "$DEPLOY_DIR/postgres_data" ]] \
          || die "sub2api-postgres PGDATA 挂载不符合预期"
        ;;
      sub2api-postgres:/var/lib/postgresql)
        [[ "$type" == "volume" ]] \
          || die "PostgreSQL 镜像父目录卷不符合预期"
        "${LOCAL_ROOT[@]}" test -d "$source" \
          || die "PostgreSQL 镜像父目录卷路径不可读: $source"
        unexpected="$("${LOCAL_ROOT[@]}" find "$source" -mindepth 1 ! -type d -print -quit)"
        [[ -z "$unexpected" ]] \
          || die "PostgreSQL 父目录匿名卷包含额外数据，需人工迁移: $unexpected"
        ;;
      sub2api-redis:/data)
        [[ "$type" == "bind" && "$(readlink -m -- "$source")" == "$DEPLOY_DIR/redis_data" ]] \
          || die "sub2api-redis /data 挂载不符合预期"
        ;;
      *)
        die "$container 存在脚本未覆盖的额外持久化挂载: $source -> $destination"
        ;;
    esac
  done <<<"$mounts"
}

validate_no_external_symlinks() {
  local root=$1
  local canonical_root link resolved raw_target link_list

  canonical_root="$(readlink -f -- "$root")"
  link_list="$TMP_DIR/symlinks.$RANDOM.$RANDOM"
  "${LOCAL_ROOT[@]}" find "$root" -type l -print0 >"$link_list" \
    || die "无法完整扫描持久化目录中的符号链接: $root"
  while IFS= read -r -d '' link; do
    raw_target="$("${LOCAL_ROOT[@]}" readlink -- "$link")" \
      || die "无法读取符号链接: $link"
    [[ "$raw_target" != /* ]] \
      || die "持久化目录中不允许绝对符号链接（目标路径改变后会失效）: $link -> $raw_target"
    resolved="$("${LOCAL_ROOT[@]}" readlink -f -- "$link")" \
      || die "持久化目录中有失效符号链接，rsync 无法安全迁移: $link"
    case "$resolved" in
      "$canonical_root"|"$canonical_root"/*) ;;
      *) die "持久化目录中的符号链接指向目录外，需单独处理后再迁移: $link -> $resolved" ;;
    esac
  done <"$link_list"
  rm -f -- "$link_list"
}

validate_compose_image_set() {
  local expected_images running_images

  expected_images="$("${COMPOSE[@]}" config --images | LC_ALL=C sort)"
  running_images="$(printf '%s\n' "${IMAGE_REFS[@]}" | LC_ALL=C sort)"
  [[ -n "$expected_images" && "$expected_images" == "$running_images" ]] \
    || die "实际运行镜像与 local+build Compose 期望镜像不一致，拒绝迁移"
}

validate_compose_config_hashes() {
  local hash_output service expected_hash actual_hash container
  declare -A expected_hashes=()

  hash_output="$("${COMPOSE[@]}" config --hash='*')" \
    || die "无法计算当前 Compose 配置哈希"
  while read -r service expected_hash; do
    [[ -n "$service" && -n "$expected_hash" ]] || continue
    expected_hashes["$service"]="$expected_hash"
  done <<<"$hash_output"

  for container in sub2api sub2api-postgres sub2api-redis; do
    service="$("${LOCAL_ROOT[@]}" docker inspect \
      --format '{{index .Config.Labels "com.docker.compose.service"}}' "$container")"
    actual_hash="$("${LOCAL_ROOT[@]}" docker inspect \
      --format '{{index .Config.Labels "com.docker.compose.config-hash"}}' "$container")"
    [[ -n "$service" && -n "$actual_hash" && "${expected_hashes[$service]:-}" == "$actual_hash" ]] \
      || die "$container 的运行配置与当前 .env/Compose 不一致；请先 recreate 并验证，再迁移"
  done
}

validate_source_credentials() {
  "${LOCAL_ROOT[@]}" docker exec sub2api-postgres sh -c \
    'PGPASSWORD="$POSTGRES_PASSWORD" psql -h 127.0.0.1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atqc "SELECT 1"' \
    | grep -qx 1 \
    || die "PostgreSQL 当前角色密码与容器配置不一致"
  "${LOCAL_ROOT[@]}" docker exec sub2api-redis sh -c \
    'if [ -n "${REDISCLI_AUTH:-}" ]; then redis-cli ping; else env -u REDISCLI_AUTH redis-cli ping; fi' \
    | grep -qx PONG \
    || die "Redis 当前密码与容器配置不一致"
}

restart_source_exact() {
  local container status
  local all_present=true
  local dependencies_healthy=false

  for container in sub2api sub2api-postgres sub2api-redis; do
    "${LOCAL_ROOT[@]}" docker inspect "$container" >/dev/null 2>&1 \
      || all_present=false
  done

  if [[ "$all_present" != true ]]; then
    "${COMPOSE[@]}" up -d --no-build --pull never
    return
  fi

  "${LOCAL_ROOT[@]}" docker start sub2api-postgres sub2api-redis >/dev/null
  for _ in $(seq 1 60); do
    status="$("${LOCAL_ROOT[@]}" docker inspect \
      --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' \
      sub2api-postgres)"
    [[ "$status" == "healthy" ]] || { sleep 1; continue; }
    status="$("${LOCAL_ROOT[@]}" docker inspect \
      --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' \
      sub2api-redis)"
    if [[ "$status" == "healthy" ]]; then
      dependencies_healthy=true
      break
    fi
    sleep 1
  done
  [[ "$dependencies_healthy" == true ]] || return 1
  "${LOCAL_ROOT[@]}" docker start sub2api >/dev/null
  for _ in $(seq 1 60); do
    status="$("${LOCAL_ROOT[@]}" docker inspect \
      --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' \
      sub2api)"
    [[ "$status" == "healthy" ]] && return 0
    sleep 1
  done
  return 1
}

cleanup() {
  local status=$?
  local remote_stopped=true
  local compose_dir_q=""

  trap - EXIT
  trap '' INT TERM
  set +e

  if [[ "$MIGRATION_SUCCEEDED" != true && "$MODE" == "execute" ]]; then
    warn "迁移未完成，开始安全回退..."

    if [[ "$TARGET_MAY_BE_RUNNING" == true && "$TARGET_FINALIZED" == true ]]; then
      compose_dir_q="$(quote "$DEST_DIR/deploy")"
      remote_sh "$(remote_root_command "sh -c 'cd $compose_dir_q && docker compose -p $COMPOSE_PROJECT -f docker-compose.local.yml -f docker-compose.build.yml down --timeout 120'")" \
        >/dev/null 2>&1 || remote_stopped=false
      if remote_sh "running=\"\$($(remote_root_command "docker ps --format '{{.Names}}'"))\" || exit 2; for n in sub2api sub2api-postgres sub2api-redis; do if printf '%s\\n' \"\$running\" | grep -Fxq \"\$n\"; then exit 1; fi; done; exit 0" \
        >/dev/null 2>&1; then
        remote_stopped=true
      else
        remote_stopped=false
      fi
    fi

    if [[ "$SOURCE_STOPPED" == true ]]; then
      if [[ "$remote_stopped" == true ]]; then
        warn "重新启动旧机服务..."
        local_sudo_ready=true
        if ((EUID != 0)) && ! sudo -n true >/dev/null 2>&1; then
          if [[ -t 0 ]]; then
            warn "本机 sudo 授权已过期，请重新验证以启动旧机..."
            sudo -v || local_sudo_ready=false
          else
            local_sudo_ready=false
          fi
        fi
        if [[ "$local_sudo_ready" == true ]]; then
          restart_source_exact >/dev/null 2>&1 \
            || warn "旧机自动重启失败，请立即手动执行 Compose up"
        else
          warn "本机 sudo 授权不可用，旧机未自动启动；请立即手动执行 Compose up"
        fi
      else
        warn "无法确认新机已停止。为防双写，未自动启动旧机！"
        warn "请先停掉新机三个容器，再手动启动旧机。"
      fi
    fi

    if [[ "$REMOTE_PREPARED" == true && "$TARGET_FINALIZED" != true ]]; then
      remote_sh "$(remote_root_command "rm -rf -- $(quote "$STAGE_BASE")")" \
        >/dev/null 2>&1 || true
    fi
    if [[ "$REMOTE_BUNDLE_TOUCHED" == true ]]; then
      remote_sh "$(remote_root_command "rm -f -- $(quote "$REMOTE_IMAGE_BUNDLE")")" \
        >/dev/null 2>&1 || true
    fi
    if [[ "$REMOTE_LOCK_OWNED" == true ]]; then
      remote_sh "$(remote_root_command "rmdir -- $(quote "$REMOTE_LOCK_DIR")")" \
        >/dev/null 2>&1 || true
    fi
  fi

  if [[ -n "$SUDO_KEEPALIVE_PID" ]]; then
    kill "$SUDO_KEEPALIVE_PID" >/dev/null 2>&1 || true
    wait "$SUDO_KEEPALIVE_PID" >/dev/null 2>&1 || true
  fi
  [[ -z "$TMP_DIR" ]] || rm -rf -- "$TMP_DIR"
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

log "检查本机 Docker Compose 配置..."
"${LOCAL_ROOT[@]}" docker info >/dev/null
"${COMPOSE[@]}" config -q

validate_no_external_symlinks "$DEPLOY_DIR/data"
validate_no_external_symlinks "$DEPLOY_DIR/postgres_data"
validate_no_external_symlinks "$DEPLOY_DIR/redis_data"

validate_mount sub2api /app/data "$DEPLOY_DIR/data"
validate_mount sub2api-postgres /var/lib/postgresql/data "$DEPLOY_DIR/postgres_data"
validate_mount sub2api-redis /data "$DEPLOY_DIR/redis_data"
validate_mount_inventory sub2api
validate_mount_inventory sub2api-postgres
validate_mount_inventory sub2api-redis
"${LOCAL_ROOT[@]}" docker exec sub2api-postgres sh -c \
  'test "$PGDATA" = /var/lib/postgresql/data' \
  || die "PostgreSQL 容器的 PGDATA 不是预期路径"

PG_MAJOR="$("${LOCAL_ROOT[@]}" head -n 1 "$DEPLOY_DIR/postgres_data/PG_VERSION")"
[[ "$PG_MAJOR" =~ ^[0-9]+$ ]] || die "无法识别 PostgreSQL PG_VERSION: $PG_MAJOR"

for container in sub2api sub2api-postgres sub2api-redis; do
  running="$("${LOCAL_ROOT[@]}" docker inspect --format '{{.State.Running}}' "$container")"
  health="$("${LOCAL_ROOT[@]}" docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container")"
  [[ "$running" == "true" ]] || die "源机容器未运行: $container"
  [[ "$health" == "none" || "$health" == "healthy" ]] \
    || die "源机容器不健康: $container ($health)"
  image_ref="$("${LOCAL_ROOT[@]}" docker inspect --format '{{.Config.Image}}' "$container")"
  running_id="$("${LOCAL_ROOT[@]}" docker inspect --format '{{.Image}}' "$container")"
  tagged_id="$("${LOCAL_ROOT[@]}" docker image inspect --format '{{.Id}}' "$image_ref")"
  [[ "$running_id" == "$tagged_id" ]] \
    || die "$container 当前运行镜像与本地标签 $image_ref 已不一致，拒绝导出错误镜像"
  IMAGE_REFS+=("$image_ref")
  IMAGE_IDS+=("$running_id")
done
validate_compose_image_set

COMPOSE_PROJECT="$("${LOCAL_ROOT[@]}" docker inspect \
  --format '{{index .Config.Labels "com.docker.compose.project"}}' sub2api)"
[[ -n "$COMPOSE_PROJECT" ]] || die "源容器缺少 Docker Compose project 标签"
[[ "$COMPOSE_PROJECT" =~ ^[a-z0-9][a-z0-9_-]*$ ]] \
  || die "源 Docker Compose project 名称不安全: $COMPOSE_PROJECT"
for container in sub2api sub2api-postgres sub2api-redis; do
  [[ "$("${LOCAL_ROOT[@]}" docker inspect --format '{{index .Config.Labels "com.docker.compose.project"}}' "$container")" == "$COMPOSE_PROJECT" ]] \
    || die "三个源容器不属于同一个 Docker Compose project"
done
COMPOSE=("${LOCAL_ROOT[@]}" docker compose -p "$COMPOSE_PROJECT" -f "$LOCAL_COMPOSE_FILE" -f "$LOCAL_BUILD_FILE")
validate_compose_config_hashes
validate_source_credentials
NETWORK_OUTPUT="$("${LOCAL_ROOT[@]}" docker network ls \
  --filter "label=com.docker.compose.project=$COMPOSE_PROJECT" --format '{{.Name}}')"
[[ -n "$NETWORK_OUTPUT" ]] || die "找不到源 Docker Compose 网络"
mapfile -t COMPOSE_NETWORKS <<<"$NETWORK_OUTPUT"

LOCAL_ARCH="$(normalize_arch "$("${LOCAL_ROOT[@]}" docker image inspect --format '{{.Architecture}}' "${IMAGE_REFS[0]}")")"

log "检查 SSH 连接与新机环境..."
remote_sh "command -v bash >/dev/null && command -v rsync >/dev/null && command -v gzip >/dev/null && command -v docker >/dev/null" \
  || die "SSH 无法连接/认证，或新机缺少 bash、rsync、gzip、docker；请先手动测试 ssh"

REMOTE_UID="$(remote_sh 'id -u')"
if [[ "$REMOTE_UID" == "0" ]]; then
  REMOTE_ROOT=""
  REMOTE_RSYNC_PATH="rsync"
else
  remote_sh "command -v sudo >/dev/null && sudo -n true" \
    || die "新机 SSH 用户需要无需交互密码的 sudo 权限"
  REMOTE_ROOT="sudo -n"
  REMOTE_RSYNC_PATH="sudo -n rsync"
fi

remote_sh "$(remote_root_command "docker info >/dev/null")"
remote_sh "$(remote_root_command "docker compose version >/dev/null")"
remote_sh "$(remote_root_command "docker compose up --help") | grep -q -- '--pull'" \
  || die "新机 Docker Compose 不支持 --pull，版本过旧"
remote_sh "mv --help 2>&1 | grep -q -- '--no-target-directory'" \
  || die "新机 mv 不支持 -T（需要 GNU coreutils）"
remote_sh "sync --help 2>&1 | grep -q -- '--file-system'" \
  || die "新机 sync 不支持 -f（需要 GNU coreutils）"
REMOTE_RSYNC_VERSION="$(remote_sh "rsync --version | head -n 1")"
if [[ "$REMOTE_RSYNC_VERSION" =~ version[[:space:]]+([0-9]+)\. ]]; then
  REMOTE_RSYNC_MAJOR="${BASH_REMATCH[1]}"
else
  die "新机 rsync 版本过旧（需要 3.x+）: $REMOTE_RSYNC_VERSION"
fi
((REMOTE_RSYNC_MAJOR >= 3)) \
  || die "新机 rsync 版本过旧（需要 3.x+）: $REMOTE_RSYNC_VERSION"

LOCAL_DOCKER_SECURITY="$("${LOCAL_ROOT[@]}" docker info --format '{{json .SecurityOptions}}')"
REMOTE_DOCKER_SECURITY="$(remote_sh "$(remote_root_command "docker info --format '{{json .SecurityOptions}}'")")"
if [[ "$LOCAL_DOCKER_SECURITY" == *"name=userns"* \
  || "$LOCAL_DOCKER_SECURITY" == *"name=rootless"* \
  || "$REMOTE_DOCKER_SECURITY" == *"name=userns"* \
  || "$REMOTE_DOCKER_SECURITY" == *"name=rootless"* ]]; then
  die "检测到 Docker userns-remap/rootless；物理 PGDATA 的 UID 映射可能不同，请改用人工迁移流程"
fi

for i in "${!IMAGE_REFS[@]}"; do
  ref_q="$(quote "${IMAGE_REFS[$i]}")"
  if remote_sh "$(remote_root_command "docker image inspect $ref_q >/dev/null 2>&1")"; then
    remote_id="$(remote_sh "$(remote_root_command "docker image inspect --format '{{.Id}}' $ref_q")")"
    [[ "$remote_id" == "${IMAGE_IDS[$i]}" ]] \
      || die "新机已有同名但内容不同的镜像，拒绝覆盖: ${IMAGE_REFS[$i]}"
  else
    remote_sh "$(remote_root_command "docker info >/dev/null")" \
      || die "检查新机镜像时 Docker 不可用"
  fi
done

for network in "${COMPOSE_NETWORKS[@]}"; do
  network_q="$(quote "$network")"
  if remote_sh "$(remote_root_command "docker network inspect $network_q >/dev/null 2>&1")"; then
    die "新机已有与 Compose 冲突的 Docker 网络: $network"
  else
    remote_sh "$(remote_root_command "docker info >/dev/null")" \
      || die "检查新机 Docker 网络时 daemon 不可用"
  fi
done

SOURCE_PUBLISHED_PORT="$("${LOCAL_ROOT[@]}" docker port sub2api 8080/tcp | head -n 1)"
SOURCE_PUBLISHED_PORT="${SOURCE_PUBLISHED_PORT##*:}"
[[ "$SOURCE_PUBLISHED_PORT" =~ ^[0-9]+$ ]] \
  || die "无法识别 sub2api 的宿主机端口"
if remote_sh "command -v ss >/dev/null"; then
  remote_sh "listeners=\"\$(ss -H -ltn)\" || exit 2; if printf '%s\\n' \"\$listeners\" | awk '{print \$4}' | grep -Eq '(^|:)$SOURCE_PUBLISHED_PORT\$'; then exit 1; fi; exit 0" \
    || die "新机端口 $SOURCE_PUBLISHED_PORT 已被占用，或无法确认监听状态"
fi

REMOTE_HOME="$(remote_sh 'printf "%s" "$HOME"')"
if [[ -z "$DEST_DIR" ]]; then
  DEST_DIR="${REMOTE_HOME%/}/sub2api"
fi
while [[ "$DEST_DIR" != "/" && "$DEST_DIR" == */ ]]; do
  DEST_DIR="${DEST_DIR%/}"
done
[[ "$DEST_DIR" = /* ]] || die "--dest 必须是新机上的绝对路径"
[[ "$DEST_DIR" =~ ^/[A-Za-z0-9._/-]+$ ]] || die "--dest 只能包含字母、数字、点、下划线、横线和斜杠"
[[ "$DEST_DIR" != "/" && "$DEST_DIR" != *"/../"* && "$DEST_DIR" != */.. ]] \
  || die "--dest 不安全"

STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
STAGE_BASE="${DEST_DIR}.incoming-${STAMP}"
STAGE_DIR="$STAGE_BASE/payload"
REMOTE_IMAGE_BUNDLE="${DEST_DIR}.images-${STAMP}.tar.gz"
REMOTE_LOCK_DIR="${DEST_DIR}.migration-lock"

DEST_Q="$(quote "$DEST_DIR")"
STAGE_Q="$(quote "$STAGE_DIR")"
STAGE_BASE_Q="$(quote "$STAGE_BASE")"
REMOTE_BUNDLE_Q="$(quote "$REMOTE_IMAGE_BUNDLE")"
REMOTE_LOCK_Q="$(quote "$REMOTE_LOCK_DIR")"
DEST_PARENT="$(dirname -- "$DEST_DIR")"
DEST_PARENT_Q="$(quote "$DEST_PARENT")"

remote_sh "$(remote_root_command "test -d $DEST_PARENT_Q")" \
  || die "新机目标父目录不存在: $DEST_PARENT"
remote_sh "$(remote_root_command "test ! -e $DEST_Q && test ! -L $DEST_Q")" \
  || die "新机目标目录已存在，为防覆盖拒绝继续: $DEST_DIR"
remote_sh "$(remote_root_command "test ! -e $STAGE_BASE_Q && test ! -L $STAGE_BASE_Q && test ! -e $REMOTE_BUNDLE_Q && test ! -L $REMOTE_BUNDLE_Q")" \
  || die "新机临时迁移路径已存在，请先人工检查"
remote_sh "$(remote_root_command "test ! -e $REMOTE_LOCK_Q && test ! -L $REMOTE_LOCK_Q")" \
  || die "新机存在迁移锁，可能有另一迁移正在运行: $REMOTE_LOCK_DIR"

CONFLICTS="$(remote_sh "for n in sub2api sub2api-postgres sub2api-redis; do if $(remote_root_command "docker inspect \"\$n\"") >/dev/null 2>&1; then printf '%s\\n' \"\$n\"; fi; done")"
[[ -z "$CONFLICTS" ]] || die "新机已有同名容器，拒绝覆盖: $CONFLICTS"

LOCAL_MACHINE_ID="$(cat /etc/machine-id 2>/dev/null || true)"
REMOTE_MACHINE_ID="$(remote_sh 'cat /etc/machine-id 2>/dev/null || true')"
if [[ -n "$LOCAL_MACHINE_ID" && "$LOCAL_MACHINE_ID" == "$REMOTE_MACHINE_ID" ]]; then
  die "源机和目标机 machine-id 相同，疑似把目标指回了本机"
fi

REMOTE_ARCH="$(normalize_arch "$(remote_sh "$(remote_root_command "docker info --format '{{.Architecture}}'")")")"
[[ "$LOCAL_ARCH" == "$REMOTE_ARCH" ]] \
  || die "CPU 架构不一致：源镜像=$LOCAL_ARCH，新机=$REMOTE_ARCH"

SOURCE_BYTES="$(cd -- "$PROJECT_ROOT" && "${LOCAL_ROOT[@]}" du -sb \
  --exclude='./.cache' --exclude='./.git' --exclude='./frontend/node_modules' \
  --exclude='./node_modules' . | awk '{print $1}')"
IMAGE_BYTES=0
for ref in "${IMAGE_REFS[@]}"; do
  size="$("${LOCAL_ROOT[@]}" docker image inspect --format '{{.Size}}' "$ref")"
  ((IMAGE_BYTES += size))
done
MARGIN_BYTES=$((SOURCE_BYTES / 5))
((MARGIN_BYTES >= 1073741824)) || MARGIN_BYTES=1073741824
REQUIRED_BYTES=$((SOURCE_BYTES + IMAGE_BYTES + MARGIN_BYTES))
REMOTE_FREE_BYTES="$(remote_sh "$(remote_root_command "df -PB1 $DEST_PARENT_Q") | awk 'NR==2 {print \$4}'")"
[[ "$REMOTE_FREE_BYTES" =~ ^[0-9]+$ ]] || die "无法读取新机可用磁盘空间"
((REMOTE_FREE_BYTES >= REQUIRED_BYTES)) \
  || die "新机空间不足：预计至少 $(human_bytes "$REQUIRED_BYTES")，可用 $(human_bytes "$REMOTE_FREE_BYTES")"
LOCAL_TMP_FREE_BYTES="$(df -PB1 "$TMP_DIR" | awk 'NR==2 {print $4}')"
[[ "$LOCAL_TMP_FREE_BYTES" =~ ^[0-9]+$ ]] || die "无法读取本机临时目录可用空间"
((LOCAL_TMP_FREE_BYTES >= IMAGE_BYTES)) \
  || die "本机临时目录空间不足：导出镜像最多需要 $(human_bytes "$IMAGE_BYTES")，可用 $(human_bytes "$LOCAL_TMP_FREE_BYTES")"

ok "预检通过"
printf '  源项目:       %s\n' "$PROJECT_ROOT"
printf '  核心 PGDATA:  %s（PostgreSQL %s）\n' "$DEPLOY_DIR/postgres_data" "$PG_MAJOR"
printf '  待传数据约:   %s（不含镜像压缩收益）\n' "$(human_bytes "$SOURCE_BYTES")"
printf '  当前镜像:     %s\n' "${IMAGE_REFS[*]}"
printf '  新机:         %s\n' "$SSH_REMOTE"
printf '  目标目录:     %s\n' "$DEST_DIR"
printf '  新机可用空间: %s\n' "$(human_bytes "$REMOTE_FREE_BYTES")"
printf '  排除项:       .cache、.git、node_modules（均非运行数据）\n'

if [[ "$MODE" == "dry-run" ]]; then
  printf '\n[DRY-RUN] 没有写入新机，也没有停止旧机。\n'
  printf '确认后在原命令末尾加 --execute；无人值守可再加 --yes。\n'
  MIGRATION_SUCCEEDED=true
  exit 0
fi

warn "执行期间必须保持 DNS/反向代理/客户端流量不指向新机；否则失败回退时可能丢失新机上的新写入。"

if [[ "$ASSUME_YES" != true ]]; then
  [[ -t 0 ]] || die "非交互执行必须加 --yes"
  printf '\n即将把 %s 迁移到 %s:%s。\n' "$PROJECT_ROOT" "$SSH_REMOTE" "$DEST_DIR"
  printf '最终同步阶段会停止旧机；成功后旧机保持停止。输入 MIGRATE 继续: '
  read -r confirmation
  [[ "$confirmation" == "MIGRATE" ]] || die "已取消"
fi

start_sudo_keepalive

log "收紧源机敏感配置权限（.env/config.yaml -> 600）..."
"${LOCAL_ROOT[@]}" chmod 600 "$DEPLOY_DIR/.env"
if [[ -f "$DEPLOY_DIR/data/config.yaml" ]]; then
  "${LOCAL_ROOT[@]}" chmod 600 "$DEPLOY_DIR/data/config.yaml"
fi

log "在新机获取独占迁移锁..."
remote_sh "$(remote_root_command "mkdir -- $REMOTE_LOCK_Q")"
REMOTE_LOCK_OWNED=true

log "在新机创建隔离的临时目录..."
remote_sh "$(remote_root_command "mkdir -- $STAGE_BASE_Q")"
REMOTE_PREPARED=true
remote_sh "$(remote_root_command "chmod 700 $STAGE_BASE_Q") && $(remote_root_command "mkdir -- $STAGE_Q")"

RSYNC_COMMON=(
  -aHAXz
  --numeric-ids
  --human-readable
  --info=progress2
  --partial
  --exclude=/.cache/
  --exclude=/.git/
  --exclude=/frontend/node_modules/
  --exclude=/node_modules/
)

log "预传项目文件（数据库仍在旧机运行，此阶段不复制 PGDATA/Redis）..."
"${LOCAL_ROOT[@]}" rsync "${RSYNC_COMMON[@]}" \
  --exclude=/deploy/postgres_data/ \
  --exclude=/deploy/redis_data/ \
  -e "$RSYNC_RSH" --rsync-path="$REMOTE_RSYNC_PATH" \
  "$PROJECT_ROOT/" "$RSYNC_REMOTE:$STAGE_DIR/"
remote_sh "$(remote_root_command "sh -c 'chmod 700 $STAGE_BASE_Q && chmod 600 $STAGE_Q/deploy/.env && if [ -f $STAGE_Q/deploy/data/config.yaml ]; then chmod 600 $STAGE_Q/deploy/data/config.yaml; fi'")"

log "导出当前实际运行的 Docker 镜像（确保新机版本完全一致）..."
"${LOCAL_ROOT[@]}" docker save "${IMAGE_REFS[@]}" | gzip -1 >"$LOCAL_IMAGE_BUNDLE"

log "用 rsync 传输镜像包..."
remote_sh "$(remote_root_command "sh -c 'set -C; : > $REMOTE_BUNDLE_Q'")"
REMOTE_BUNDLE_TOUCHED=true
"${LOCAL_ROOT[@]}" rsync -a --partial --human-readable --info=progress2 \
  -e "$RSYNC_RSH" --rsync-path="$REMOTE_RSYNC_PATH" \
  "$LOCAL_IMAGE_BUNDLE" "$RSYNC_REMOTE:$REMOTE_IMAGE_BUNDLE"

log "在新机导入镜像并核对 image ID..."
remote_sh "$(remote_root_command "gzip -dc -- $REMOTE_BUNDLE_Q") | $(remote_root_command "docker load") >/dev/null"
for i in "${!IMAGE_REFS[@]}"; do
  ref_q="$(quote "${IMAGE_REFS[$i]}")"
  remote_id="$(remote_sh "$(remote_root_command "docker image inspect --format '{{.Id}}' $ref_q")")"
  [[ "$remote_id" == "${IMAGE_IDS[$i]}" ]] \
    || die "新机镜像校验失败: ${IMAGE_REFS[$i]}"
done
remote_sh "$(remote_root_command "rm -f -- $REMOTE_BUNDLE_Q")"
REMOTE_BUNDLE_TOUCHED=false

log "在停机前验证新机 Compose 配置..."
remote_sh "$(remote_root_command "sh -c 'cd $STAGE_Q/deploy && docker compose -p $COMPOSE_PROJECT -f docker-compose.local.yml -f docker-compose.build.yml config -q'")"

log "再次确认源机挂载、容器和镜像在预传期间没有变化..."
"${COMPOSE[@]}" config -q
validate_compose_image_set
validate_compose_config_hashes
validate_source_credentials
for critical_path in "${CRITICAL_PATHS[@]}"; do
  [[ ! -L "$critical_path" ]] \
    || die "预传期间关键路径被改成符号链接: $critical_path"
done
validate_no_external_symlinks "$DEPLOY_DIR/data"
validate_no_external_symlinks "$DEPLOY_DIR/postgres_data"
validate_no_external_symlinks "$DEPLOY_DIR/redis_data"
validate_mount sub2api /app/data "$DEPLOY_DIR/data"
validate_mount sub2api-postgres /var/lib/postgresql/data "$DEPLOY_DIR/postgres_data"
validate_mount sub2api-redis /data "$DEPLOY_DIR/redis_data"
validate_mount_inventory sub2api
validate_mount_inventory sub2api-postgres
validate_mount_inventory sub2api-redis
"${LOCAL_ROOT[@]}" docker exec sub2api-postgres sh -c \
  'test "$PGDATA" = /var/lib/postgresql/data' \
  || die "预传期间 PostgreSQL PGDATA 发生变化"
containers=(sub2api sub2api-postgres sub2api-redis)
for i in "${!containers[@]}"; do
  container="${containers[$i]}"
  [[ "$("${LOCAL_ROOT[@]}" docker inspect --format '{{.State.Running}}' "$container")" == "true" ]] \
    || die "预传期间源容器停止: $container"
  [[ "$("${LOCAL_ROOT[@]}" docker inspect --format '{{.Config.Image}}' "$container")" == "${IMAGE_REFS[$i]}" ]] \
    || die "预传期间源容器镜像引用变化: $container"
  [[ "$("${LOCAL_ROOT[@]}" docker inspect --format '{{.Image}}' "$container")" == "${IMAGE_IDS[$i]}" ]] \
    || die "预传期间源容器镜像 ID 变化: $container"
  [[ "$("${LOCAL_ROOT[@]}" docker image inspect --format '{{.Id}}' "${IMAGE_REFS[$i]}")" == "${IMAGE_IDS[$i]}" ]] \
    || die "预传期间本地镜像标签变化: ${IMAGE_REFS[$i]}"
done

log "停止旧机应用写入..."
if ((EUID != 0)); then
  sudo -v
fi
SOURCE_STOPPED=true
"${COMPOSE[@]}" stop -t 60 sub2api

log "同步保存 Redis，并让 PostgreSQL 执行 CHECKPOINT..."
"${LOCAL_ROOT[@]}" docker exec sub2api-redis sh -c \
  'if [ -n "${REDISCLI_AUTH:-}" ]; then redis-cli SAVE; else env -u REDISCLI_AUTH redis-cli SAVE; fi' \
  >/dev/null
"${LOCAL_ROOT[@]}" docker exec sub2api-postgres sh -c \
  'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -c "CHECKPOINT;"' \
  >/dev/null

log "干净停止 PostgreSQL 和 Redis（保留旧容器以便快速回退）..."
if ! timeout --kill-after=10s 150s \
  "${COMPOSE[@]}" stop -t 120 postgres redis; then
  die "停止 PostgreSQL/Redis 失败或超时；拒绝复制数据"
fi
running_names="$("${LOCAL_ROOT[@]}" docker ps --format '{{.Names}}')" \
  || die "无法确认旧机 Docker 容器状态"
for container in sub2api sub2api-postgres sub2api-redis; do
  if printf '%s\n' "$running_names" | grep -Fxq "$container"; then
    die "旧机容器仍在运行: $container"
  fi
done
"${LOCAL_ROOT[@]}" test ! -e "$DEPLOY_DIR/postgres_data/postmaster.pid" \
  || die "PostgreSQL 未留下干净关机状态（postmaster.pid 仍存在）"

log "最终 rsync：复制停机后的 PostgreSQL/Redis，并逐文件校验..."
"${LOCAL_ROOT[@]}" rsync "${RSYNC_COMMON[@]}" \
  --checksum --delete-delay \
  -e "$RSYNC_RSH" --rsync-path="$REMOTE_RSYNC_PATH" \
  "$PROJECT_ROOT/" "$RSYNC_REMOTE:$STAGE_DIR/"
remote_sh "$(remote_root_command "sh -c 'chmod 700 $STAGE_BASE_Q && chmod 600 $STAGE_Q/deploy/.env && if [ -f $STAGE_Q/deploy/data/config.yaml ]; then chmod 600 $STAGE_Q/deploy/data/config.yaml; fi'")"
remote_sh "$(remote_root_command "sync -f $STAGE_Q")"

log "复核源端与新机临时目录是否零差异..."
VERIFY_STDOUT_FILE="$TMP_DIR/verify.stdout"
VERIFY_STDERR_FILE="$TMP_DIR/verify.stderr"
set +e
"${LOCAL_ROOT[@]}" rsync -aHAXcn --numeric-ids --delete-delay \
  --exclude=/.cache/ --exclude=/.git/ \
  --exclude=/frontend/node_modules/ --exclude=/node_modules/ \
  --out-format='%i %n%L' \
  -e "$RSYNC_RSH" --rsync-path="$REMOTE_RSYNC_PATH" \
  "$PROJECT_ROOT/" "$RSYNC_REMOTE:$STAGE_DIR/" \
  >"$VERIFY_STDOUT_FILE" 2>"$VERIFY_STDERR_FILE"
VERIFY_STATUS=$?
set -e
VERIFY_OUTPUT="$(<"$VERIFY_STDOUT_FILE")"
VERIFY_ERROR="$(<"$VERIFY_STDERR_FILE")"
((VERIFY_STATUS == 0)) || die "最终 rsync 复核命令失败: $VERIFY_ERROR"
[[ -z "$VERIFY_ERROR" ]] || warn "最终 rsync 复核产生非致命 SSH/rsync 提示: $VERIFY_ERROR"
[[ -z "$VERIFY_OUTPUT" ]] || die "最终 rsync 后仍有差异，拒绝启动新机: $VERIFY_OUTPUT"

log "将校验后的临时目录原子切换为正式目录..."
remote_sh "$(remote_root_command "sh -c 'test ! -e $DEST_Q && test ! -L $DEST_Q && mv -T -- $STAGE_Q $DEST_Q'")"
TARGET_FINALIZED=true
remote_sh "$(remote_root_command "rmdir -- $STAGE_BASE_Q")" >/dev/null 2>&1 || true
remote_sh "$(remote_root_command "sync -f $DEST_Q")"

remote_sh "$(remote_root_command "chmod 600 $DEST_Q/deploy/.env") && \
  if [ -f $DEST_Q/deploy/data/config.yaml ]; then $(remote_root_command "chmod 600 $DEST_Q/deploy/data/config.yaml"); fi"
remote_sh "$(remote_root_command "sh -c 'cd $DEST_Q/deploy && docker compose -p $COMPOSE_PROJECT -f docker-compose.local.yml -f docker-compose.build.yml config -q'")"

log "启动新机（禁止拉取或重建镜像）..."
TARGET_MAY_BE_RUNNING=true
remote_sh "$(remote_root_command "sh -c 'cd $DEST_Q/deploy && docker compose -p $COMPOSE_PROJECT -f docker-compose.local.yml -f docker-compose.build.yml up -d --no-build --pull never'")"

log "等待三个容器通过健康检查..."
healthy=false
for attempt in $(seq 1 36); do
  statuses="$(remote_sh "for n in sub2api-postgres sub2api-redis sub2api; do printf '%s=' \"\$n\"; $(remote_root_command "docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' \"\$n\"") 2>/dev/null || printf 'missing\\n'; done" || true)"
  if [[ "$statuses" == *"sub2api-postgres=healthy"* \
    && "$statuses" == *"sub2api-redis=healthy"* \
    && "$statuses" == *"sub2api=healthy"* ]]; then
    healthy=true
    break
  fi
  if ((attempt == 1 || attempt % 3 == 0)); then
    log "新机状态: ${statuses//$'\n'/ }"
  fi
  sleep 5
done
[[ "$healthy" == true ]] || die "新机 180 秒内未全部 healthy"

log "执行数据库、Redis 与应用冒烟检查..."
remote_sh "$(remote_root_command "docker exec sub2api-postgres sh -c 'psql -U \"\$POSTGRES_USER\" -d \"\$POSTGRES_DB\" -Atqc \"SELECT 1\"'")" \
  | grep -qx 1
remote_sh "$(remote_root_command "docker exec sub2api-redis sh -c 'if [ -n \"\${REDISCLI_AUTH:-}\" ]; then redis-cli ping; else env -u REDISCLI_AUTH redis-cli ping; fi'")" \
  | grep -qx PONG
remote_sh "$(remote_root_command "docker exec sub2api wget -q -T 5 -O /dev/null http://localhost:8080/health")"

remote_sh "$(remote_root_command "rmdir -- $REMOTE_LOCK_Q")" >/dev/null 2>&1 \
  || warn "无法删除新机迁移锁，请确认无并发迁移后手动删除: $REMOTE_LOCK_DIR"
REMOTE_LOCK_OWNED=false
MIGRATION_SUCCEEDED=true
ok "迁移完成；新机服务 healthy，旧机服务保持停止"
printf '\n'
printf '新机项目目录: %s\n' "$DEST_DIR"
printf '下一步：\n'
printf '  1. 单独配置 Nginx/Certbot；本脚本未迁移 /etc/nginx 和 /etc/letsencrypt。\n'
printf '  2. 通过新机反向代理做业务验证，再切 DNS。\n'
printf '  3. 不要同时启动旧机，避免两套数据库分叉。\n'
printf '  4. 新机产生写入后若要回滚，必须先停新机并反向迁移最新数据。\n'
