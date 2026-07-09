#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_NAME="$(basename "$0")"
MACHINE2_IP=""
PROXY_PORT="7890"
PROXY_TYPE="http"
ENV_FILE="${HOME}/.config/isacapi/ai-proxy-env.sh"
PAC_FILE="${HOME}/.config/isacapi/ai-proxy.pac"
CLAUDE_SETTINGS="${HOME}/.claude/settings.json"
WRITE_SHELL_PROFILE=1
WRITE_CLAUDE_SETTINGS=1
WRITE_PAC=1
OFFICIAL_UPSTREAM=0
CLEAR_CLAUDE_AUTH=0
CHECK_PROXY=0
DRY_RUN=0

BEGIN_MARK="# >>> ISACAPI AI proxy route >>>"
END_MARK="# <<< ISACAPI AI proxy route <<<"

AI_PROXY_DOMAINS=(
  "anthropic.com"
  "claude.ai"
  "claudeusercontent.com"
  "claude.sh"
  "x.ai"
  "api.x.ai"
  "grok.com"
  "generativelanguage.googleapis.com"
  "cloudcode-pa.googleapis.com"
  "aiplatform.googleapis.com"
  "oauth2.googleapis.com"
  "accounts.google.com"
  "www.googleapis.com"
  "googleapis.com"
  "google.com"
  "gstatic.com"
  "googleusercontent.com"
  "openai.com"
  "api.openai.com"
  "chatgpt.com"
  "oaistatic.com"
  "oaiusercontent.com"
)

usage() {
  cat <<USAGE
Usage:
  ${SCRIPT_NAME} --machine2-ip <IP> [options]
  ${SCRIPT_NAME} <IP> [options]

This configures machine 1 to use machine 2 as the proxy exit for AI traffic.
Machine 2 must already be running an HTTP or SOCKS proxy, for example CCProxy,
Clash/mihomo, sing-box, Squid, or 3proxy.

Required:
  --machine2-ip <IP>       Machine 2 IP address. Do not include http://.

Options:
  --proxy-port <port>      Proxy port on machine 2. Default: 7890
  --proxy-type <type>      http, socks5, or socks5h. Default: http
  --env-file <path>        Env file to write. Default: ~/.config/isacapi/ai-proxy-env.sh
  --pac-file <path>        PAC file to write. Default: ~/.config/isacapi/ai-proxy.pac
  --claude-settings <path> Claude Code settings path. Default: ~/.claude/settings.json
  --official-upstream      Remove ANTHROPIC_BASE_URL from Claude settings so Claude
                           uses official Anthropic endpoints through the proxy.
  --clear-claude-auth      Also remove ANTHROPIC_AUTH_TOKEN/ANTHROPIC_API_KEY from
                           Claude settings. Use only if they are old gateway keys.
  --no-shell-profile       Do not add source block to ~/.bashrc or ~/.zshrc
  --no-claude              Do not update Claude Code settings.json
  --no-pac                 Do not write PAC file
  --check                  Test the proxy with curl after writing config
  --dry-run                Print what would be written without changing files
  -h, --help               Show this help

Examples:
  ${SCRIPT_NAME} --machine2-ip 10.0.0.2
  ${SCRIPT_NAME} 10.0.0.2 --proxy-port 808 --proxy-type http
  ${SCRIPT_NAME} 10.0.0.2 --proxy-port 1080 --proxy-type socks5h --official-upstream
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

expand_path() {
  local path="$1"
  case "$path" in
    "~") printf '%s\n' "$HOME" ;;
    "~/"*) printf '%s/%s\n' "$HOME" "${path#~/}" ;;
    *) printf '%s\n' "$path" ;;
  esac
}

quote_sh() {
  local value="$1"
  printf "'"
  printf '%s' "$value" | sed "s/'/'\\\\''/g"
  printf "'"
}

emit_export() {
  local name="$1"
  local value="$2"
  printf 'export %s=' "$name"
  quote_sh "$value"
  printf '\n'
}

url_host() {
  local host="$1"
  if [[ "$host" == *:* && "$host" != \[*\] ]]; then
    printf '[%s]' "$host"
  else
    printf '%s' "$host"
  fi
}

proxy_url() {
  local host="$1"
  case "$PROXY_TYPE" in
    http) printf 'http://%s:%s' "$host" "$PROXY_PORT" ;;
    socks5) printf 'socks5://%s:%s' "$host" "$PROXY_PORT" ;;
    socks5h) printf 'socks5h://%s:%s' "$host" "$PROXY_PORT" ;;
    *) die "--proxy-type must be http, socks5, or socks5h" ;;
  esac
}

pac_proxy_spec() {
  case "$PROXY_TYPE" in
    http) printf 'PROXY %s:%s' "$MACHINE2_IP" "$PROXY_PORT" ;;
    socks5|socks5h) printf 'SOCKS5 %s:%s' "$MACHINE2_IP" "$PROXY_PORT" ;;
    *) die "--proxy-type must be http, socks5, or socks5h" ;;
  esac
}

validate() {
  [[ -n "$MACHINE2_IP" ]] || die "machine 2 IP is required. Use --machine2-ip <IP>"
  [[ "$MACHINE2_IP" != *"://"* ]] || die "--machine2-ip must be only an IP/host, without http:// or https://"
  [[ "$MACHINE2_IP" != */* ]] || die "--machine2-ip must not contain a path"
  [[ "$PROXY_PORT" =~ ^[0-9]+$ ]] || die "--proxy-port must be a number"
  (( PROXY_PORT >= 1 && PROXY_PORT <= 65535 )) || die "--proxy-port must be between 1 and 65535"
  case "$PROXY_TYPE" in
    http|socks5|socks5h) ;;
    *) die "--proxy-type must be http, socks5, or socks5h" ;;
  esac
}

write_env_file() {
  local proxy="$1"
  local no_proxy="127.0.0.1,localhost,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Would write env file: ${ENV_FILE}"
    info "Proxy URL: ${proxy}"
    return 0
  fi

  mkdir -p "$(dirname "$ENV_FILE")"
  local tmp="${ENV_FILE}.tmp.$$"

  {
    printf '# Generated by ISACAPI %s. Re-run the script to update.\n' "$SCRIPT_NAME"
    printf '# Source this file before starting AI clients on machine 1.\n'
    printf '# Note: proxy environment variables affect any proxy-aware process in that shell.\n\n'

    emit_export ISACAPI_AI_PROXY_MACHINE2_IP "$MACHINE2_IP"
    emit_export ISACAPI_AI_PROXY_URL "$proxy"
    emit_export ISACAPI_AI_PROXY_PAC "$PAC_FILE"
    printf '\n'

    emit_export HTTP_PROXY "$proxy"
    emit_export HTTPS_PROXY "$proxy"
    emit_export ALL_PROXY "$proxy"
    emit_export http_proxy "$proxy"
    emit_export https_proxy "$proxy"
    emit_export all_proxy "$proxy"
    emit_export NO_PROXY "$no_proxy"
    emit_export no_proxy "$no_proxy"
    printf '\n'

    emit_export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC "1"
    emit_export CLAUDE_CODE_ATTRIBUTION_HEADER "0"
  } >"$tmp"

  mv "$tmp" "$ENV_FILE"
  chmod 600 "$ENV_FILE"
  info "Wrote env file: ${ENV_FILE}"
}

write_pac_file() {
  [[ "$WRITE_PAC" -eq 1 ]] || return 0

  local spec
  spec="$(pac_proxy_spec)"

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Would write PAC file: ${PAC_FILE}"
    info "PAC proxy spec: ${spec}"
    return 0
  fi

  mkdir -p "$(dirname "$PAC_FILE")"
  local tmp="${PAC_FILE}.tmp.$$"

  {
    printf 'function FindProxyForURL(url, host) {\n'
    printf '  host = host.toLowerCase();\n'
    printf '  var proxy = "%s";\n' "$spec"
    printf '  var domains = [\n'
    local i
    local comma
    for i in "${!AI_PROXY_DOMAINS[@]}"; do
      comma=","
      if (( i == ${#AI_PROXY_DOMAINS[@]} - 1 )); then
        comma=""
      fi
      printf '    "%s"%s\n' "${AI_PROXY_DOMAINS[$i]}" "$comma"
    done
    printf '  ];\n'
    printf '  for (var i = 0; i < domains.length; i++) {\n'
    printf '    var d = domains[i];\n'
    printf '    if (host === d || dnsDomainIs(host, "." + d)) return proxy;\n'
    printf '  }\n'
    printf '  return "DIRECT";\n'
    printf '}\n'
  } >"$tmp"

  mv "$tmp" "$PAC_FILE"
  chmod 600 "$PAC_FILE"
  info "Wrote PAC file: ${PAC_FILE}"
}

append_profile_block() {
  local profile="$1"
  [[ "$WRITE_SHELL_PROFILE" -eq 1 ]] || return 0
  [[ -n "$profile" ]] || return 0

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Would ensure ${profile} sources ${ENV_FILE}"
    return 0
  fi

  if [[ ! -e "$profile" ]]; then
    : >"$profile"
  fi

  if grep -Fq "$BEGIN_MARK" "$profile"; then
    info "Shell profile already has ISACAPI source block: ${profile}"
    return 0
  fi

  {
    printf '\n%s\n' "$BEGIN_MARK"
    printf 'if [ -f %s ]; then\n' "$(quote_sh "$ENV_FILE")"
    printf '  . %s\n' "$(quote_sh "$ENV_FILE")"
    printf 'fi\n'
    printf '%s\n' "$END_MARK"
  } >>"$profile"

  info "Updated shell profile: ${profile}"
}

update_claude_settings() {
  local proxy="$1"
  local no_proxy="127.0.0.1,localhost,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"

  [[ "$WRITE_CLAUDE_SETTINGS" -eq 1 ]] || return 0

  if ! command -v python3 >/dev/null 2>&1; then
    warn "python3 not found; skipped Claude settings update"
    return 0
  fi

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Would update Claude settings: ${CLAUDE_SETTINGS}"
    return 0
  fi

  mkdir -p "$(dirname "$CLAUDE_SETTINGS")"
  if [[ -f "$CLAUDE_SETTINGS" ]]; then
    cp "$CLAUDE_SETTINGS" "${CLAUDE_SETTINGS}.bak.$(date +%Y%m%d%H%M%S)"
  fi

  ISACAPI_CLAUDE_SETTINGS="$CLAUDE_SETTINGS" \
  ISACAPI_PROXY_URL="$proxy" \
  ISACAPI_NO_PROXY="$no_proxy" \
  ISACAPI_OFFICIAL_UPSTREAM="$OFFICIAL_UPSTREAM" \
  ISACAPI_CLEAR_CLAUDE_AUTH="$CLEAR_CLAUDE_AUTH" \
  python3 - <<'PY'
import json
import os
from pathlib import Path

path = Path(os.environ["ISACAPI_CLAUDE_SETTINGS"])
proxy = os.environ["ISACAPI_PROXY_URL"]
no_proxy = os.environ["ISACAPI_NO_PROXY"]
official = os.environ["ISACAPI_OFFICIAL_UPSTREAM"] == "1"
clear_auth = os.environ["ISACAPI_CLEAR_CLAUDE_AUTH"] == "1"

if path.exists() and path.read_text(encoding="utf-8").strip():
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise SystemExit(f"invalid Claude settings JSON: {exc}") from exc
else:
    data = {}

if not isinstance(data, dict):
    data = {}

env = data.get("env")
if not isinstance(env, dict):
    env = {}

env["HTTP_PROXY"] = proxy
env["HTTPS_PROXY"] = proxy
env["ALL_PROXY"] = proxy
env["http_proxy"] = proxy
env["https_proxy"] = proxy
env["all_proxy"] = proxy
env["NO_PROXY"] = no_proxy
env["no_proxy"] = no_proxy
env["CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"] = "1"
env["CLAUDE_CODE_ATTRIBUTION_HEADER"] = "0"

if official:
    for key in ("ANTHROPIC_BASE_URL", "ANTHROPIC_API_URL", "CLAUDE_API_BASE_URL"):
        env.pop(key, None)

if clear_auth:
    for key in ("ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_API_KEY"):
        env.pop(key, None)

data["env"] = env
path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
PY

  chmod 600 "$CLAUDE_SETTINGS"
  info "Updated Claude settings: ${CLAUDE_SETTINGS}"

  if [[ "$OFFICIAL_UPSTREAM" -eq 0 ]]; then
    warn "If Claude settings already contain ANTHROPIC_BASE_URL, Claude will request that host through machine 2."
    warn "Run again with --official-upstream if you want Claude to use official Anthropic endpoints through machine 2."
  fi
}

check_proxy() {
  local proxy="$1"
  [[ "$CHECK_PROXY" -eq 1 ]] || return 0

  if ! command -v curl >/dev/null 2>&1; then
    warn "curl not found; skipped proxy check"
    return 0
  fi

  info "Checking proxy with https://api.anthropic.com/"
  if curl -fsS --max-time 10 --proxy "$proxy" https://api.anthropic.com/ >/dev/null; then
    info "Proxy check passed"
  else
    warn "Proxy check failed. Verify machine 2 proxy service, port, firewall, and proxy type."
  fi
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --machine2-ip)
      require_value "$1" "${2:-}"
      MACHINE2_IP="$2"
      shift 2
      ;;
    --proxy-port|--port)
      require_value "$1" "${2:-}"
      PROXY_PORT="$2"
      shift 2
      ;;
    --proxy-type|--type)
      require_value "$1" "${2:-}"
      PROXY_TYPE="$2"
      shift 2
      ;;
    --env-file)
      require_value "$1" "${2:-}"
      ENV_FILE="$(expand_path "$2")"
      shift 2
      ;;
    --pac-file)
      require_value "$1" "${2:-}"
      PAC_FILE="$(expand_path "$2")"
      shift 2
      ;;
    --claude-settings)
      require_value "$1" "${2:-}"
      CLAUDE_SETTINGS="$(expand_path "$2")"
      shift 2
      ;;
    --official-upstream)
      OFFICIAL_UPSTREAM=1
      shift
      ;;
    --clear-claude-auth)
      CLEAR_CLAUDE_AUTH=1
      shift
      ;;
    --no-shell-profile)
      WRITE_SHELL_PROFILE=0
      shift
      ;;
    --no-claude)
      WRITE_CLAUDE_SETTINGS=0
      shift
      ;;
    --no-pac)
      WRITE_PAC=0
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
      if [[ -z "$MACHINE2_IP" ]]; then
        MACHINE2_IP="$1"
        shift
      else
        die "unexpected argument: $1"
      fi
      ;;
  esac
done

validate

HOST_FOR_URL="$(url_host "$MACHINE2_IP")"
PROXY_URL="$(proxy_url "$HOST_FOR_URL")"

info "Machine 2 IP: ${MACHINE2_IP}"
info "Proxy URL: ${PROXY_URL}"
info "Proxy type: ${PROXY_TYPE}"
info "Proxy port: ${PROXY_PORT}"

write_env_file "$PROXY_URL"
write_pac_file
update_claude_settings "$PROXY_URL"
append_profile_block "${HOME}/.bashrc"
if [[ -f "${HOME}/.zshrc" ]]; then
  append_profile_block "${HOME}/.zshrc"
fi
check_proxy "$PROXY_URL"

info "Done. For the current shell, run:"
printf '  source %q\n' "$ENV_FILE"
if [[ "$WRITE_PAC" -eq 1 ]]; then
  info "PAC file for domain-based proxy-aware apps: ${PAC_FILE}"
fi
