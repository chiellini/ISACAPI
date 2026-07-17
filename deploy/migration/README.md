# ISACAPI 跨服务器 / 跨 AWS 账号迁移手册

将 Docker Compose 部署的 ISACAPI（应用 + PostgreSQL + Redis）整体迁移到另一台服务器。
适用场景：换机器、换 AWS 账号、换云厂商。新机反向代理与 TLS 准备好后，
DNS 在 Cloudflare 等第三方时通常只需切换 A/AAAA 记录。

## 推荐：一键 rsync 迁移

`rsync-migrate.sh` 会自动完成：本地/远端预检、传输当前精确 Docker 镜像、
停止写入、Redis `SAVE`、干净关闭 PostgreSQL、最终 rsync、逐文件复核、
新机启动和健康检查。执行失败时，它会先停止新机，再尝试恢复旧机，避免双写。

脚本默认只预检，不会停服或写远端：

首次连接应先手动 SSH，核对并记录新机 host key 指纹：

```bash
ssh -i /home/ec2-user/<新机私钥>.pem ec2-user@<新机IP>
```

确认指纹后退出 SSH，再运行：

```bash
cd /home/ec2-user/sub2api
bash deploy/migration/rsync-migrate.sh \
  --host <新机IP> \
  --user ec2-user \
  --identity /home/ec2-user/<新机私钥>.pem
```

确认预检结果后，增加 `--execute`：

```bash
bash deploy/migration/rsync-migrate.sh \
  --host <新机IP> \
  --user ec2-user \
  --identity /home/ec2-user/<新机私钥>.pem \
  --execute
```

新机必须预先安装 `bash`、`rsync`、`gzip`、Docker 与 Docker Compose plugin；
SSH 用户需为 root 或拥有无需交互密码的 sudo。默认目标目录是远端
`$HOME/sub2api`，可用 `--dest /绝对路径` 修改。无人值守执行还需加 `--yes`。
目标目录必须不存在，目标父目录必须已经创建。
一键物理 PGDATA 迁移仅支持 rootful Docker；检测到 rootless 或
`userns-remap` 时脚本会拒绝执行，以免 UID/GID 映射不同导致数据库不可用。
若确实要在首次连接时采用 TOFU，可显式增加 `--accept-new-host-key`；默认会
拒绝 unknown host key，以免把数据库和密钥传给未经核对的主机。

从执行迁移到脚本报告成功之前，不得把 DNS、反向代理、安全组转发或客户端
流量指向新机。否则目标验活失败并回启旧机时，新机期间的写入不会自动反向迁回。

脚本迁移整个项目，但排除 `.cache`、`.git` 和 `node_modules`；这些都不是
运行数据。`.env`、`data`、`postgres_data`、`redis_data` 会完整迁移。

> 当前部署使用 `docker-compose.local.yml` + `docker-compose.build.yml` 和本地
> `sub2api:local` 镜像。新脚本会把实际运行的镜像一并传走，避免新机误用
> registry 中不同版本的 `latest`。
>
> 一键脚本专门支持当前的 local bind-mount 三目录部署；若 Docker 实际挂载
> 不是 `deploy/{data,postgres_data,redis_data}`，预检会拒绝执行。命名卷或
> standalone 外部数据库部署请使用对应的备份/恢复流程。

## 备用：tar + scp 两段式迁移

下面的 `backup-old-server.sh` / `restore-new-server.sh` 是旧的两段式方案。
它适合标准 registry 镜像部署；若当前使用本地构建镜像，优先使用上面的
rsync 脚本。

## 目录内容

| 文件 | 在哪跑 | 作用 |
|---|---|---|
| `rsync-migrate.sh` | 旧服务器 | **推荐**：预检、停服、rsync、远端启动与失败回退一体化 |
| `backup-old-server.sh` | 旧服务器 | 停服 → 自动识别数据模式 → 打包部署目录 + 全部数据为单个 tar.gz |
| `restore-new-server.sh` | 新服务器 | 解包 → 恢复数据 → 启动 → 轮询健康检查 |

下面两个旧脚本支持两种数据模式，自动识别、无需选择：

- **本地目录模式**（`docker-compose.local.yml`）：数据在 `deploy/` 下的 `./data`、`./postgres_data`、`./redis_data`
- **命名卷模式**（`docker-compose.yml`）：数据在 Docker 卷里（`/var/lib/docker/volumes/...`），脚本会逐卷导出/回灌

## 迁移步骤

### 1. 准备新机
- 规格 ≥ 旧机，磁盘 ≥ 数据量 × 2；安全组放行 80/443（SSH 22 限自己 IP）
- 在切 DNS 前安装并配置 Nginx/Caddy/Tunnel/LB 和 TLS 证书；应用本身只暴露 8080
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
先确认新机的 Nginx/Caddy/Tunnel/LB 能通过正式域名和 HTTPS 访问，再把
Cloudflare 的 A/AAAA 记录改为新机公网 IP，并保持所需的代理/TLS 模式。

### 6. 验证清单
1. `docker compose ps` 三个容器全部 healthy
2. 用**旧密码**登录管理后台——能登录说明 `.env` 里的 `JWT_SECRET` 带对了
3. 「账号管理」确认上游账号状态正常
4. 用现有 API key 发一条测试请求，在请求日志里看到成功记录
5. 确认新机出站能连上游（Anthropic 等）：个别 IP 段可能被上游风控，若全部 403/超时需换 EIP 或配代理

### 7. 收尾
旧机保留观察 1–2 天。新机尚未产生写入时，可先停新机再启动旧机；新机
已经产生写入后，回滚前必须把最新数据反向迁回，不能只改 DNS，否则会丢数据。

## 关键注意事项

- **必须 sudo**：`postgres_data` 归容器内用户所有且权限为 700，非 root 读不了。脚本已内置检查。
- **禁止在容器运行中拷贝 `postgres_data`**：数据文件与 WAL 不一致会导致库损坏。备份脚本第一步 `compose down` 就是为此。
- **`.env` 必须原样带走**（脚本已包含）：`JWT_SECRET` 影响现有登录会话；`TOTP_ENCRYPTION_KEY` 还用于 2FA、监控/API 密钥、备份和支付相关密文；对话归档加密密钥丢失后，历史加密内容无法解密。
- **Postgres 镜像大版本必须一致**（当前 `postgres:18-alpine`）：大版本不同会拒绝加载旧 PGDATA。
- **HTTPS 终结**：compose 内应用只监听 8080。当前机器的 Nginx 配置和 Let's Encrypt 证书位于项目目录之外，不在迁移包中；必须在新机另行迁移或重建。
- **外部守护进程**：若另行启用了 `datamanagementd`，还需单独迁移其 `/var/lib/sub2api/datamanagement/`；当前机器未启用该组件。
- **Elastic IP 不能跨账号转移**，新 IP 不可避免，靠 DNS 切换。
- Redis 数据（缓存、sticky session、调度状态等）已包含在迁移中，不应主动丢弃。
