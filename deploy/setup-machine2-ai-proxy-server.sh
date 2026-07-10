#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
PROXY_PORT="7890"
LISTEN_HOST="0.0.0.0"
CONFIG_PATH="/etc/tinyproxy/tinyproxy.conf"
SERVICE_NAME="tinyproxy"
INSTALL_PACKAGE=1
OPEN_FIREWALL=0
CHECK_PROXY=0
DRY_RUN=0
ALLOW_CLIENTS=()

usage() {
  cat <<USAGE
Usage:
  sudo ${SCRIPT_NAME} --allow-client <machine1-ip-or-cidr> [options]

This configures machine 2 as an HTTP/HTTPS CONNECT proxy for machine 1.
Machine 1 can then route Anthropic/Grok/Gemini/Google/OpenAI traffic through
machine 2 by using deploy/configure-machine1-ai-proxy-route.sh.

Required:
  --allow-client <IP/CIDR>   Machine 1 IP or CIDR allowed to use this proxy.
                             Can be repeated or comma-separated.

Options:
  --proxy-port <port>        Proxy port to listen on. Default: 7890
  --listen <ip>              Listen address. Default: 0.0.0.0
  --config <path>            Tinyproxy config path. Default: /etc/tinyproxy/tinyproxy.conf
  --no-install               Do not install tinyproxy package
  --open-firewall            Add allow rules for ufw or firewalld if present
  --check                    Test proxy locally with curl after restart
  --dry-run                  Print actions without writing files or restarting services
  -h, --help                 Show this help

Examples:
  sudo ${SCRIPT_NAME} --allow-client 10.0.0.11
  sudo ${SCRIPT_NAME} --allow-client 10.0.0.0/24 --proxy-port 808
  sudo ${SCRIPT_NAME} --allow-client 203.0.113.10 --open-firewall --check
USAGE
}

die() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

info() {
  printf '[ISACAPI] %s\n' "$*"
}

warn() {
  printf '[ISACAPI] WARNING: %s\n' "$*" >&2
}

require_value() {
  local opt="${1:-}"
  local value="${2:-}"
  if [[ -z "${value}" || "${value}" == --* ]]; then
    die "${opt} requires a value"
  fi
}

trim_space() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

add_allow_clients() {
  local raw="$1"
  local item
  local old_ifs="$IFS"
  IFS=","
  for item in $raw; do
    item="$(trim_space "$item")"
    [[ -n "$item" ]] && ALLOW_CLIENTS+=("$item")
  done
  IFS="$old_ifs"
}

is_open_proxy_allow() {
  local value="$1"
  case "$value" in
    "*"|"all"|"ALL"|"any"|"ANY"|"0.0.0.0"|"0.0.0.0/0"|"::"|"::/0")
      return 0
      ;;
  esac
  return 1
}

validate() {
  [[ "$PROXY_PORT" =~ ^[0-9]+$ ]] || die "--proxy-port must be a number"
  (( PROXY_PORT >= 1 && PROXY_PORT <= 65535 )) || die "--proxy-port must be between 1 and 65535"
  [[ -n "$LISTEN_HOST" ]] || die "--listen cannot be empty"
  [[ -n "$CONFIG_PATH" ]] || die "--config cannot be empty"
  (( ${#ALLOW_CLIENTS[@]} > 0 )) || die "--allow-client is required. Use machine 1 IP or CIDR"

  local client
  for client in "${ALLOW_CLIENTS[@]}"; do
    is_open_proxy_allow "$client" && die "refusing open proxy allow rule: ${client}. Use machine 1 IP/CIDR instead"
  done

  return 0
}

require_root_if_needed() {
  if [[ "$DRY_RUN" -eq 0 && "${EUID:-$(id -u)}" -ne 0 ]]; then
    die "run this script with sudo on machine 2"
  fi
}

has_yum_repo() {
  local repo="$1"
  grep -Rqs "^\[${repo}\]" /etc/yum.repos.d 2>/dev/null
}

rpm_install_package() {
  local pm="$1"
  shift
  local opts=()

  if has_yum_repo "docker-ce-stable"; then
    opts+=(--disablerepo=docker-ce-stable)
  fi

  "$pm" "${opts[@]}" install -y "$@"
}

enable_epel_if_possible() {
  local pm="$1"

  if command -v rpm >/dev/null 2>&1 && rpm -q epel-release >/dev/null 2>&1; then
    return 0
  fi

  if command -v amazon-linux-extras >/dev/null 2>&1; then
    if amazon-linux-extras install -y epel; then
      return 0
    fi
  fi

  rpm_install_package "$pm" epel-release
}

install_tinyproxy_rpm() {
  local pm="$1"

  if rpm_install_package "$pm" tinyproxy; then
    return 0
  fi

  warn "tinyproxy was not found in enabled ${pm} repositories; trying EPEL"
  if enable_epel_if_possible "$pm"; then
    if rpm_install_package "$pm" tinyproxy; then
      return 0
    fi
  else
    warn "could not enable/install epel-release automatically"
  fi

  die "tinyproxy is not available from current ${pm} repositories. Enable EPEL or install tinyproxy manually, then rerun with --no-install. If docker-ce-stable is broken, the script already tries to ignore it temporarily."
}

install_tinyproxy() {
  [[ "$INSTALL_PACKAGE" -eq 1 ]] || return 0

  if command -v tinyproxy >/dev/null 2>&1; then
    info "tinyproxy is already installed"
    return 0
  fi

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Would install tinyproxy"
    return 0
  fi

  if command -v apt-get >/dev/null 2>&1; then
    apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y tinyproxy
  elif command -v dnf >/dev/null 2>&1; then
    install_tinyproxy_rpm dnf
  elif command -v yum >/dev/null 2>&1; then
    install_tinyproxy_rpm yum
  elif command -v zypper >/dev/null 2>&1; then
    zypper --non-interactive install tinyproxy
  elif command -v pacman >/dev/null 2>&1; then
    pacman -Sy --noconfirm tinyproxy
  else
    die "could not find a supported package manager. Install tinyproxy manually, then rerun with --no-install"
  fi
}

detect_proxy_user() {
  if id -u tinyproxy >/dev/null 2>&1; then
    printf 'tinyproxy'
  else
    printf 'nobody'
  fi
}

detect_proxy_group() {
  if getent group tinyproxy >/dev/null 2>&1; then
    printf 'tinyproxy'
  elif getent group nogroup >/dev/null 2>&1; then
    printf 'nogroup'
  else
    printf 'nobody'
  fi
}

write_tinyproxy_config() {
  local proxy_user="$1"
  local proxy_group="$2"

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Would write tinyproxy config: ${CONFIG_PATH}"
    info "Would allow clients: ${ALLOW_CLIENTS[*]}"
    return 0
  fi

  mkdir -p "$(dirname "$CONFIG_PATH")"
  if [[ -f "$CONFIG_PATH" ]]; then
    cp "$CONFIG_PATH" "${CONFIG_PATH}.bak.$(date +%Y%m%d%H%M%S)"
  fi

  {
    printf '# Generated by ISACAPI %s. Re-run the script to update.\n' "$SCRIPT_NAME"
    printf '# Machine 2 proxy server for machine 1 AI upstream traffic.\n\n'
    printf 'User %s\n' "$proxy_user"
    printf 'Group %s\n' "$proxy_group"
    printf 'Port %s\n' "$PROXY_PORT"
    printf 'Listen %s\n' "$LISTEN_HOST"
    printf 'Timeout 600\n'
    printf 'MaxClients 200\n'
    printf 'StartServers 10\n'
    printf 'MinSpareServers 5\n'
    printf 'MaxSpareServers 20\n'
    printf 'LogLevel Info\n'
    printf 'PidFile "/run/tinyproxy/tinyproxy.pid"\n'
    printf 'ViaProxyName "isacapi-machine2"\n'
    printf 'DisableViaHeader Yes\n'
    printf 'ConnectPort 443\n'
    printf 'ConnectPort 563\n'
    printf 'ConnectPort 8443\n'
    printf '\n'
    local client
    for client in "${ALLOW_CLIENTS[@]}"; do
      printf 'Allow %s\n' "$client"
    done
  } >"$CONFIG_PATH"

  info "Wrote tinyproxy config: ${CONFIG_PATH}"
}

restart_service() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Would enable and restart ${SERVICE_NAME}"
    return 0
  fi

  if command -v systemctl >/dev/null 2>&1; then
    systemctl enable "$SERVICE_NAME"
    systemctl restart "$SERVICE_NAME"
    systemctl --no-pager --full status "$SERVICE_NAME" || true
  elif command -v service >/dev/null 2>&1; then
    service "$SERVICE_NAME" restart
  else
    warn "Could not find systemctl or service. Start tinyproxy manually."
  fi
}

firewall_family() {
  local value="$1"
  if [[ "$value" == *:* ]]; then
    printf 'ipv6'
  else
    printf 'ipv4'
  fi
}

open_firewall() {
  [[ "$OPEN_FIREWALL" -eq 1 ]] || return 0

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Would open firewall for ${ALLOW_CLIENTS[*]} -> tcp/${PROXY_PORT}"
    return 0
  fi

  local client
  if command -v ufw >/dev/null 2>&1; then
    for client in "${ALLOW_CLIENTS[@]}"; do
      ufw allow proto tcp from "$client" to any port "$PROXY_PORT"
    done
    info "Updated ufw rules"
  elif command -v firewall-cmd >/dev/null 2>&1; then
    for client in "${ALLOW_CLIENTS[@]}"; do
      local family
      family="$(firewall_family "$client")"
      firewall-cmd --permanent --add-rich-rule="rule family=\"${family}\" source address=\"${client}\" port port=\"${PROXY_PORT}\" protocol=\"tcp\" accept"
    done
    firewall-cmd --reload
    info "Updated firewalld rules"
  else
    warn "No ufw/firewalld found. Open tcp/${PROXY_PORT} from machine 1 in your firewall/security group."
  fi
}

check_proxy() {
  [[ "$CHECK_PROXY" -eq 1 ]] || return 0

  if ! command -v curl >/dev/null 2>&1; then
    warn "curl not found; skipped proxy check"
    return 0
  fi

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Would test proxy through http://127.0.0.1:${PROXY_PORT}"
    return 0
  fi

  info "Testing local proxy with https://api.anthropic.com/"
  if curl -sS --max-time 15 --proxy "http://127.0.0.1:${PROXY_PORT}" https://api.anthropic.com/ >/dev/null; then
    info "Local proxy test passed"
  else
    warn "Local proxy test failed. Check tinyproxy logs and machine 2 outbound network."
  fi
}

print_machine1_command() {
  local public_hint
  public_hint="$(hostname -I 2>/dev/null | awk '{print $1}' || true)"
  [[ -n "$public_hint" ]] || public_hint="机器2IP"

  info "Machine 1 command example:"
  printf '  ./deploy/configure-machine1-ai-proxy-route.sh --machine2-ip %s --proxy-port %s --proxy-type http --official-upstream\n' "$public_hint" "$PROXY_PORT"

  return 0
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --allow-client)
      require_value "$1" "${2:-}"
      add_allow_clients "$2"
      shift 2
      ;;
    --proxy-port|--port)
      require_value "$1" "${2:-}"
      PROXY_PORT="$2"
      shift 2
      ;;
    --listen)
      require_value "$1" "${2:-}"
      LISTEN_HOST="$2"
      shift 2
      ;;
    --config)
      require_value "$1" "${2:-}"
      CONFIG_PATH="$2"
      shift 2
      ;;
    --no-install)
      INSTALL_PACKAGE=0
      shift
      ;;
    --open-firewall)
      OPEN_FIREWALL=1
      shift
      ;;
    --check)
      CHECK_PROXY=1
      shift
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --*)
      die "unknown option: $1"
      ;;
    *)
      die "unexpected argument: $1"
      ;;
  esac
done

validate
require_root_if_needed

info "Proxy listen: ${LISTEN_HOST}:${PROXY_PORT}"
info "Allowed machine 1 clients: ${ALLOW_CLIENTS[*]}"

install_tinyproxy
PROXY_USER="$(detect_proxy_user)"
PROXY_GROUP="$(detect_proxy_group)"
write_tinyproxy_config "$PROXY_USER" "$PROXY_GROUP"
open_firewall
restart_service
check_proxy
print_machine1_command
