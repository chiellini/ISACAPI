# ISACAPI 跨服务器 / 跨 AWS 账号迁移手册

将 Docker Compose 部署的 ISACAPI（应用 + PostgreSQL + Redis）整体迁移到另一台服务器。
适用场景：换机器、换 AWS 账号、换云厂商。新机反向代理准备好后，按本文完成
公网 IP、DNS 和 TLS 切换。

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

一键脚本报告迁移成功后，直接继续执行本文第 5～8 步，完成公网入口、
Nginx、证书、DNS 和业务验证。

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
- 在切 DNS 前准备 Nginx/Caddy/Tunnel/LB；应用本身只监听 8080，不要把 8080 暴露到公网
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

### 5. 配置新机公网入口、Nginx 和 SSL

以下步骤适用于当前的 Amazon Linux 2023、域名 `isacai.space` / `www.isacai.space`
和 Nginx + Let's Encrypt。`rsync-migrate.sh` 只迁移应用和数据，不会修改安全组、
DNS，也不会迁移 `/etc/nginx` 或 `/etc/letsencrypt`。

#### 5.1 绑定稳定公网 IP 并放行端口

不要长期使用 EC2 自动分配、停机后可能变化的公网 IPv4。如果采用新 IP，现在给
新实例分配并关联 Elastic IP；如果准备复用旧账号的 EIP，先保留新机临时公网 IP，
等 Nginx 和证书准备好后再切换。随后在安全组设置入站规则：

| 类型 | 来源 | 说明 |
|---|---|---|
| TCP 22 | 管理员固定 IP | 仅用于 SSH，不要向全网开放 |
| TCP 80 | `0.0.0.0/0` | HTTP 和 Let's Encrypt HTTP-01 验证 |
| TCP 443 | `0.0.0.0/0` | 正式 HTTPS 流量 |
| TCP 80/443 | `::/0` | 仅在实例确实配置 IPv6 时添加 |

应用端口 8080 只供本机 Nginx 使用，不应加入公网入站规则。先在新机确认应用正常：

```bash
curl -fsS http://127.0.0.1:8080/health
```

公网 IP 不需要写进 Nginx 的 `server_name`；Nginx 绑定正式域名，公网 IP 由 EC2
关联 EIP，域名再通过 DNS A 记录指向该 EIP。

跨 AWS 账号时，同一区域内的 Elastic IP 可以转给新账号。先在旧账号通过 EC2
控制台的 **Elastic IP addresses → Actions → Enable transfer** 授权转移；等新机
Nginx 和证书准备好后，再从旧实例解绑 EIP，由新账号在 7 天内接受并关联到新实例。
接受转移时 EIP 不能仍关联旧实例，因此解绑到重新关联期间会有短暂中断。不同区域
不能转移 EIP，应在新账号重新分配 EIP，随后修改 DNS。完整限制见
[AWS 官方 EIP 跨账号转移文档](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/transfer-EIPs-intro-ec2.html)。

#### 5.2 安装 Nginx 和 Certbot

在新机执行：

```bash
sudo dnf install -y nginx python3 python3-devel augeas-devel gcc bind-utils cronie
sudo systemctl enable --now nginx crond

sudo python3 -m venv /opt/certbot/
sudo /opt/certbot/bin/pip install --upgrade pip
sudo /opt/certbot/bin/pip install certbot certbot-nginx
sudo ln -sf /opt/certbot/bin/certbot /usr/local/bin/certbot

nginx -v
certbot --version
```

这里采用 Certbot 官方的 [Python 虚拟环境安装方式](https://certbot.eff.org/instructions?os=pip&ws=nginx)。

#### 5.3 先配置 HTTP 反向代理

证书尚未生成时不能直接启用仓库中的 HTTPS 模板，否则 Nginx 会因证书文件不存在
而启动失败。先创建只监听 80 的临时配置：

```bash
sudo tee /etc/nginx/conf.d/isacai.space.conf >/dev/null <<'NGINX'
server {
    listen 80;
    listen [::]:80;
    server_name isacai.space www.isacai.space;

    underscores_in_headers on;
    client_max_body_size 256m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;

        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
NGINX

sudo nginx -t
sudo systemctl reload nginx
curl -fsS -H 'Host: isacai.space' http://127.0.0.1/health
```

`underscores_in_headers on` 不能省略；Codex 客户端可能发送 `session_id` 等含下划线
的请求头，Nginx 默认会丢弃它们。

#### 5.4 把域名解析到新机

采用新 EIP 时，在域名 DNS 控制台修改记录。若已把旧 EIP 跨账号转移并关联到新机，
IP 没有变化，DNS 记录无需修改，但仍需检查记录是否正确：

| 主机记录 | 类型 | 值 |
|---|---|---|
| `@` | A | 新机 Elastic IP |
| `www` | CNAME | `isacai.space`（也可以使用指向同一新 IP 的 A 记录） |

删除仍指向旧机的 A 记录。如果新机没有配置 IPv6，也要删除旧的 AAAA 记录，否则
部分客户端和 Let's Encrypt 可能访问错误的 IPv6 地址。Cloudflare 用户申请证书时
可先设为 **DNS only**；证书生效后再启用代理，并使用 **Full (strict)** TLS 模式。

等待公共 DNS 返回新 IP：

```bash
dig +short A isacai.space @1.1.1.1
dig +short A www.isacai.space @1.1.1.1
dig +short AAAA isacai.space @1.1.1.1
```

前两个命令必须最终指向新机；未配置 IPv6 时，第三个命令应无输出。Certbot 的
HTTP-01 验证会从公网访问域名的 80 端口，因此 DNS 尚未指向新机时不要申请证书。

#### 5.5 申请证书并启用正式 HTTPS 配置

DNS 生效且公网 80 端口可访问后，在新机执行：

```bash
sudo certbot certonly --nginx -d isacai.space -d www.isacai.space

sudo install -m 0644 \
  /home/ec2-user/sub2api/deploy/nginx-isacai.example.conf \
  /etc/nginx/conf.d/isacai.space.conf

sudo nginx -t
sudo systemctl reload nginx
curl -fsS https://isacai.space/health
```

正式模板会把 HTTP 以 308 重定向到 HTTPS，并把流量反向代理到
`127.0.0.1:8080`。如果一键迁移时使用了非默认 `--dest`，请把模板的源路径改为
实际项目路径。

配置自动续期并做一次模拟续期：

```bash
echo "0 0,12 * * * root /opt/certbot/bin/python -c 'import random; import time; time.sleep(random.random() * 3600)' && /usr/local/bin/certbot renew -q --deploy-hook '/usr/bin/systemctl reload nginx'" \
  | sudo tee /etc/cron.d/certbot-renew >/dev/null
sudo chmod 0644 /etc/cron.d/certbot-renew
sudo certbot renew --dry-run
```

以后应定期更新 Certbot：

```bash
sudo /opt/certbot/bin/pip install --upgrade certbot certbot-nginx
```

#### 5.6 可选：复用旧机证书，减少 HTTPS 切换空窗

现有 Let's Encrypt 证书可以迁到新机。需要在旧机安全打包完整的 Nginx 和
Let's Encrypt 目录，再传到新机；压缩包包含私钥，严禁公开、长期保存或提交 Git。

旧机执行：

```bash
sudo tar -C / -czf /tmp/isacapi-nginx-tls.tar.gz etc/nginx etc/letsencrypt
sudo chown "$USER:$USER" /tmp/isacapi-nginx-tls.tar.gz
chmod 0600 /tmp/isacapi-nginx-tls.tar.gz
scp -i /home/ec2-user/<新机私钥>.pem \
  /tmp/isacapi-nginx-tls.tar.gz ec2-user@<新机IP>:~/
```

新机已安装 Nginx/Certbot 后执行：

```bash
sudo tar -C / -xzf ~/isacapi-nginx-tls.tar.gz
sudo nginx -t
sudo systemctl enable --now nginx

curl --resolve isacai.space:443:<新机IP> \
  https://isacai.space/health
```

通过 `--resolve` 验证成功后再切 DNS。确认新机续期正常后，删除两台机器上的临时
压缩包。不要只复制 `/etc/letsencrypt/live`：其中是符号链接，必须连同
`archive`、`renewal` 和 `accounts` 一起迁移。域名切换后再运行
`sudo certbot renew --dry-run`，否则 HTTP-01 仍会验证旧机。

生产入口仍应使用域名。Let's Encrypt 虽已支持裸 IP 证书，但它是约 6 天有效期的
短期证书，Certbot 的 Nginx 插件也不能自动安装该类证书，不属于本文部署流程；
具体限制见 [Let's Encrypt 官方说明](https://letsencrypt.org/2026/03/11/shorter-certs-certbot/)。

### 6. 完成 DNS/HTTPS 切换

确认 `curl --resolve`（复用证书）或正式域名（重新签发证书）访问成功后，把所有
A/AAAA 记录切到新 EIP；复用旧 EIP 时无需改记录，只需确认 EIP 已关联新实例。
等待 DNS 生效期间保留旧机，不要启动旧机的应用容器，避免两套数据库同时接收
写入。Cloudflare 等代理服务应检查回源地址和 TLS 模式。

### 7. 验证清单
1. `docker compose ps` 三个容器全部 healthy
2. `sudo nginx -t` 通过，`curl -I http://isacai.space` 返回 HTTPS 重定向
3. `curl -fsS https://isacai.space/health` 成功，浏览器证书包含两个正式域名
4. `sudo certbot renew --dry-run` 通过，`crond` 已启用
5. 用**旧密码**登录管理后台——能登录说明 `.env` 里的 `JWT_SECRET` 带对了
6. 「账号管理」确认上游账号状态正常
7. 用现有 API key 发一条测试请求，在请求日志里看到成功记录
8. 确认新机出站能连上游（Anthropic 等）：个别 IP 段可能被上游风控，若全部 403/超时需换 EIP 或配代理

### 8. 收尾
旧机保留观察 1–2 天。新机尚未产生写入时，可先停新机再启动旧机；新机
已经产生写入后，回滚前必须把最新数据反向迁回，不能只改 DNS，否则会丢数据。

## 关键注意事项

- **必须 sudo**：`postgres_data` 归容器内用户所有且权限为 700，非 root 读不了。脚本已内置检查。
- **禁止在容器运行中拷贝 `postgres_data`**：数据文件与 WAL 不一致会导致库损坏。备份脚本第一步 `compose down` 就是为此。
- **`.env` 必须原样带走**（脚本已包含）：`JWT_SECRET` 影响现有登录会话；`TOTP_ENCRYPTION_KEY` 还用于 2FA、监控/API 密钥、备份和支付相关密文；对话归档加密密钥丢失后，历史加密内容无法解密。
- **Postgres 镜像大版本必须一致**（当前 `postgres:18-alpine`）：大版本不同会拒绝加载旧 PGDATA。
- **HTTPS 终结**：compose 内应用只监听 8080。当前机器的 Nginx 配置和 Let's Encrypt 证书位于项目目录之外，不在迁移包中；必须在新机另行迁移或重建。
- **外部守护进程**：若另行启用了 `datamanagementd`，还需单独迁移其 `/var/lib/sub2api/datamanagement/`；当前机器未启用该组件。
- **Elastic IP**：同一 AWS 区域支持跨账号转移；跨区域不能转移，改用新 EIP 和 DNS 切换。
- Redis 数据（缓存、sticky session、调度状态等）已包含在迁移中，不应主动丢弃。
