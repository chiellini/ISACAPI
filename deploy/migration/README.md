# ISACAPI 跨服务器 / 跨 AWS 账号迁移手册

将 Docker Compose 部署的 ISACAPI（应用 + PostgreSQL + Redis）整体迁移到另一台服务器。
适用场景：换机器、换 AWS 账号、换云厂商。DNS 在 Cloudflare 等第三方时最简单——迁完只需改一条 A 记录。

## 目录内容

| 文件 | 在哪跑 | 作用 |
|---|---|---|
| `backup-old-server.sh` | 旧服务器 | 停服 → 自动识别数据模式 → 打包部署目录 + 全部数据为单个 tar.gz |
| `restore-new-server.sh` | 新服务器 | 解包 → 恢复数据 → 启动 → 轮询健康检查 |

两个脚本都支持两种数据模式，自动识别、无需选择：

- **本地目录模式**（`docker-compose.local.yml`）：数据在 `deploy/` 下的 `./data`、`./postgres_data`、`./redis_data`
- **命名卷模式**（`docker-compose.yml`）：数据在 Docker 卷里（`/var/lib/docker/volumes/...`），脚本会逐卷导出/回灌

## 迁移步骤

### 1. 准备新机
- 规格 ≥ 旧机，磁盘 ≥ 数据量 × 2；安全组放行 80/443（SSH 22 限自己 IP）
- 安装 Docker + Compose plugin：
  ```bash
  # Amazon Linux 2023
  sudo dnf install -y docker && sudo systemctl enable --now docker
  sudo curl -SL https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64 \
    -o /usr/local/lib/docker/cli-plugins/docker-compose --create-dirs
  sudo chmod +x /usr/local/lib/docker/cli-plugins/docker-compose
  ```
- 把 `restore-new-server.sh` 传到新机

### 2. 旧机打包（⚠️ 执行即停服，停机窗口开始）
```bash
cd <部署目录>/deploy        # 例如 ~/sub2api/deploy
sudo bash migration/backup-old-server.sh
```
产物：`~/sub2api-migration-<时间戳>.tar.gz`。
脚本同时把当前镜像 digest 记入包内 `image-digests.txt`（`weishaw/sub2api:latest` 在新机可能拉到更新版本，出问题可按 digest 回退）。

若要推迟迁移/回滚，旧机直接重启即可：`docker compose -f <compose文件> up -d`

### 3. 传输
```bash
scp ~/sub2api-migration-*.tar.gz user@新机IP:~/
```
数据量大（几十 GB 以上）建议走 S3 中转：旧机上传 → 生成预签名 URL → 新机 curl 下载（跨账号无需配桶策略）。

### 4. 新机恢复
```bash
sudo bash restore-new-server.sh ~/sub2api-migration-<时间戳>.tar.gz /home/ec2-user/sub2api
```
第二个参数是目标父目录（默认 `/home/isacai/ISACAPI`，按需指定）。
脚本启动后轮询 `http://localhost:8080/health`，最多等 150 秒，就绪时打印 `[OK]`。

### 5. 切 DNS
Cloudflare 把域名的 A 记录改为新机公网 IP，保持原有代理（橙云）/ TLS 模式不变。开着 Cloudflare 代理时切换基本即时生效。

### 6. 验证清单
1. `docker compose ps` 三个容器全部 healthy
2. 用**旧密码**登录管理后台——能登录说明 `.env` 里的 `JWT_SECRET` 带对了
3. 「账号管理」确认上游账号状态正常
4. 用现有 API key 发一条测试请求，在请求日志里看到成功记录
5. 确认新机出站能连上游（Anthropic 等）：个别 IP 段可能被上游风控，若全部 403/超时需换 EIP 或配代理

### 7. 收尾
旧机保留观察 1–2 天（回滚方案 = Cloudflare 改回旧 IP + 旧机 `compose up -d`），确认无异常后关停释放。

## 关键注意事项

- **必须 sudo**：`postgres_data` 归容器内用户所有（uid 999、权限 700），非 root 读不了。脚本已内置检查。
- **禁止在容器运行中拷贝 `postgres_data`**：数据文件与 WAL 不一致会导致库损坏。备份脚本第一步 `compose down` 就是为此。
- **`.env` 必须原样带走**（脚本已包含）：`JWT_SECRET` 变了所有登录会话失效；`TOTP_ENCRYPTION_KEY` 变了所有用户的 2FA 永久解不开。
- **Postgres 镜像大版本必须一致**（当前 `postgres:18-alpine`）：大版本不同会拒绝加载旧 PGDATA。
- **HTTPS 终结**：compose 内应用只监听 8080。若旧机有 compose 之外的 nginx/caddy 做 443，其配置和证书不在备份包里，需在新机另行安装；Cloudflare 直接回源则无此步。
- **Elastic IP 不能跨账号转移**，新 IP 不可避免，靠 DNS 切换。
- Redis 数据（缓存、sticky session）已包含在备份里；即使丢失也只影响首日性能，不影响功能。
