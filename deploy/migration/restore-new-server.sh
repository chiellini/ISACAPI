#!/usr/bin/env bash
# =============================================================================
# ISACAPI 迁移脚本 2/2：新服务器恢复
# 前置：新机已装 Docker + Compose plugin，安全组放行 80/443。
# 用法：
#   bash restore-new-server.sh ~/sub2api-migration-<日期>.tar.gz [目标目录]
# 目标目录默认 /home/isacai/ISACAPI（与旧机一致，保证命名卷前缀不变）。
# =============================================================================
set -euo pipefail

# 恢复时需要原样保留数据目录的属主（postgres uid 999 等），必须 root
if [ "$(id -u)" -ne 0 ]; then
  echo "[ERR] 请用 sudo 执行（需要保留容器数据目录的属主/权限）"; exit 1
fi

TARBALL="${1:?用法: bash restore-new-server.sh <备份包.tar.gz> [目标目录]}"
TARGET_BASE="${2:-/home/isacai/ISACAPI}"

[ -f "$TARBALL" ] || { echo "[ERR] 找不到备份包: $TARBALL"; exit 1; }
command -v docker >/dev/null || { echo "[ERR] 未安装 docker"; exit 1; }

TMP="$(mktemp -d)"
tar xzf "$TARBALL" -C "$TMP"

# 备份包两种布局：
#   本地目录模式: <deploy目录>/ + image-digests.txt
#   命名卷模式:   deploy/ + volume-*.tar.gz + image-digests.txt
if ls "$TMP"/volume-*.tar.gz >/dev/null 2>&1; then
  MODE="volume"; SRC_DEPLOY="$TMP/deploy"
else
  MODE="local"; SRC_DEPLOY="$(find "$TMP" -maxdepth 1 -mindepth 1 -type d | head -1)"
fi
echo "[INFO] 备份包模式: $MODE"

# 部署目录必须与旧机同名（compose 命名卷/容器前缀 = 目录名）
mkdir -p "$TARGET_BASE"
DEPLOY_DIR="$TARGET_BASE/$(basename "$SRC_DEPLOY")"
if [ -e "$DEPLOY_DIR" ]; then
  echo "[ERR] $DEPLOY_DIR 已存在，为防覆盖请先移走它"; exit 1
fi
cp -a "$SRC_DEPLOY" "$DEPLOY_DIR"
[ -f "$TMP/image-digests.txt" ] && cp "$TMP/image-digests.txt" "$DEPLOY_DIR/"

cd "$DEPLOY_DIR"
[ -f .env ] || { echo "[ERR] 恢复的目录里没有 .env（JWT_SECRET/TOTP 密钥都在里面），备份包不完整"; exit 1; }

if [ "$MODE" = "volume" ]; then
  COMPOSE_FILE="docker-compose.yml"
  for f in "$TMP"/volume-*.tar.gz; do
    v="$(basename "$f" .tar.gz)"; v="${v#volume-}"
    echo "[INFO] 恢复卷 $v ..."
    docker volume create "$v" >/dev/null
    docker run --rm -v "$v":/to -v "$TMP":/from busybox \
      tar xzf "/from/$(basename "$f")" -C /to
  done
else
  COMPOSE_FILE="docker-compose.local.yml"
fi

echo "[INFO] 启动服务 ..."
docker compose -f "$COMPOSE_FILE" up -d

echo "[INFO] 等待健康检查 ..."
for i in $(seq 1 30); do
  if curl -fsS -m 5 http://localhost:8080/health >/dev/null 2>&1; then
    echo "[OK] 服务已就绪: http://localhost:8080"
    echo ""
    echo "[NEXT] 1. Cloudflare 把 isacai.space 的 A 记录指向本机公网 IP"
    echo "[NEXT] 2. 浏览器登录管理后台，用旧密码（能登 = JWT_SECRET 带对了）"
    echo "[NEXT] 3. 发一条测试请求并在请求日志确认成功"
    echo "[NEXT] 4. 观察 1-2 天后再关停旧机"
    rm -rf "$TMP"; exit 0
  fi
  sleep 5
done

echo "[ERR] 120 秒内健康检查未通过，查看日志:"
echo "  docker compose -f $COMPOSE_FILE ps"
echo "  docker compose -f $COMPOSE_FILE logs sub2api | tail -50"
echo "提示：若 sub2api 镜像版本与旧机不同引发问题，可按 image-digests.txt 固定镜像后重试"
rm -rf "$TMP"; exit 1
