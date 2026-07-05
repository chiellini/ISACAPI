#!/usr/bin/env bash
# =============================================================================
# ISACAPI 迁移脚本 1/2：旧服务器打包
# 用法：把本脚本放到旧服务器的 deploy 目录里执行：
#   cd /home/isacai/ISACAPI/deploy && bash migration/backup-old-server.sh
# 产物：$HOME/sub2api-migration-<日期>.tar.gz（含部署目录 + 全部数据）
# 注意：执行即开始停机窗口（docker compose down）。
# =============================================================================
set -euo pipefail

# postgres_data 归容器内用户所有（uid 999，权限 700），非 root 无法读取
if [ "$(id -u)" -ne 0 ]; then
  echo "[ERR] 请用 sudo 执行（需要读取 postgres_data 等容器数据目录）"; exit 1
fi

DEPLOY_DIR="$(pwd)"
STAMP="$(date +%Y%m%d-%H%M%S)"
WORK_DIR="$HOME/sub2api-migration-$STAMP"
OUT_TAR="$HOME/sub2api-migration-$STAMP.tar.gz"

[ -f "$DEPLOY_DIR/.env" ] || { echo "[ERR] 当前目录没有 .env，请 cd 到 deploy 目录再执行"; exit 1; }

# 判定模式：存在 ./postgres_data 目录 = 本地目录模式，否则 = 命名卷模式
if [ -d "$DEPLOY_DIR/postgres_data" ]; then
  MODE="local"
  COMPOSE_FILE="docker-compose.local.yml"
else
  MODE="volume"
  COMPOSE_FILE="docker-compose.yml"
fi
echo "[INFO] 检测到数据模式: $MODE (compose 文件: $COMPOSE_FILE)"

mkdir -p "$WORK_DIR"

# 记录当前镜像的精确 digest —— weishaw/sub2api:latest 在新机可能拉到更新版本，
# 出问题时可用这里记录的 digest 回退到与旧机完全相同的镜像。
docker images --digests --format '{{.Repository}}:{{.Tag}} {{.Digest}}' \
  | grep -E 'sub2api|postgres|redis' > "$WORK_DIR/image-digests.txt" || true
echo "[INFO] 镜像 digest 已记录到 image-digests.txt"

echo "[WARN] 即将停止服务（停机窗口开始）……"
docker compose -f "$COMPOSE_FILE" down

if [ "$MODE" = "local" ]; then
  # 本地目录模式：数据就在 deploy 目录里，整目录打包即可
  tar czf "$OUT_TAR" -C "$(dirname "$DEPLOY_DIR")" "$(basename "$DEPLOY_DIR")" -C "$WORK_DIR" image-digests.txt
else
  # 命名卷模式：卷名带 compose 项目前缀（默认=目录名），用项目 label 精确筛选
  PROJECT="$(basename "$DEPLOY_DIR" | tr '[:upper:]' '[:lower:]' | tr -cd 'a-z0-9_-')"
  VOLUMES="$(docker volume ls -q --filter "label=com.docker.compose.project=$PROJECT")"
  if [ -z "$VOLUMES" ]; then
    echo "[ERR] 未找到项目 $PROJECT 的卷。手动确认卷名: docker volume ls"; exit 1
  fi
  echo "[INFO] 待导出卷: $VOLUMES"
  for v in $VOLUMES; do
    echo "[INFO] 导出卷 $v ..."
    docker run --rm -v "$v":/from:ro -v "$WORK_DIR":/to busybox \
      tar czf "/to/volume-$v.tar.gz" -C /from .
  done
  cp -a "$DEPLOY_DIR" "$WORK_DIR/deploy"
  tar czf "$OUT_TAR" -C "$WORK_DIR" .
fi

rm -rf "$WORK_DIR"
echo ""
echo "[OK] 打包完成: $OUT_TAR ($(du -h "$OUT_TAR" | cut -f1))"
echo "[NEXT] 传输到新机: scp $OUT_TAR user@新机IP:~/"
echo "[NEXT] 若要回滚/推迟迁移，旧机重新启动: docker compose -f $COMPOSE_FILE up -d"
