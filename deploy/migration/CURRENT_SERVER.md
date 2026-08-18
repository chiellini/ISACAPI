# 当前生产服务器配置基线

本文记录 `isacai.space` 当前服务器在 **2026-07-27** 的实际运行配置，供迁移、
恢复和故障排查使用。它不是数据库 migration；`rsync-migrate.sh` 会迁移项目、
镜像和数据，但不会修改新机的 `/etc/nginx`、`/etc/letsencrypt`、systemd、
DNS、安全组或 Elastic IP。

## 主机与运行方式

| 项目 | 当前值 |
|---|---|
| 系统 | Amazon Linux 2023 (`2023.12.20260724`) |
| SELinux | `Permissive` |
| 项目目录 | `/home/ec2-user/sub2api` |
| Compose project | `deploy` |
| Compose 文件 | `deploy/docker-compose.local.yml` + `deploy/docker-compose.build.yml` |
| Docker / Compose | Engine `25.0.16`（API `1.44`）/ Compose `v5.3.1` |
| 容器镜像 | `sub2api:local`、`postgres:18-alpine`、`redis:8-alpine` |
| Nginx | `1.30.3`，systemd 已启用并运行 |
| Certbot | `4.2.0`，`/usr/bin/certbot` 指向 `/opt/certbot/bin/certbot` |

当前更新应用使用：

```bash
cd /home/ec2-user/sub2api/deploy
sudo docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.build.yml \
  up -d --build --force-recreate sub2api
```

应用映射 `0.0.0.0:8080 -> 8080/tcp`，但 AWS 安全组不得向公网开放 8080；
公网只开放 80/443，由本机 Nginx 转发。PostgreSQL 5432 和 Redis 6379
只在 Compose 网络内开放。

持久化数据使用 bind mount：

- `deploy/data` → `/app/data`
- `deploy/postgres_data` → `/var/lib/postgresql/data`
- `deploy/redis_data` → `/data`

PostgreSQL 镜像另外创建了挂在 `/var/lib/postgresql` 的 Docker 匿名卷；实际
`PGDATA=/var/lib/postgresql/data` 位于上面的 bind mount，迁移脚本以该目录为准。

`.env` 和 `deploy/data/config.yaml` 权限均为 `0600`，迁移时必须原样保留，
但绝不能提交 Git。后者包含数据库、Redis、JWT、默认额度、限流和模型别名配置；
不要根据本文手工重建。当前 `.env` 中需要复现的非敏感运行值如下：

```dotenv
BIND_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=release
RUN_MODE=standard
TZ=Asia/Shanghai
SERVER_MAX_REQUEST_BODY_SIZE=268435456
GATEWAY_MAX_BODY_SIZE=268435456
SERVER_H2C_ENABLED=true
SERVER_H2C_MAX_CONCURRENT_STREAMS=50
SERVER_H2C_IDLE_TIMEOUT=75
SERVER_H2C_MAX_READ_FRAME_SIZE=1048576
SERVER_H2C_MAX_UPLOAD_BUFFER_PER_CONNECTION=2097152
SERVER_H2C_MAX_UPLOAD_BUFFER_PER_STREAM=524288
GATEWAY_STREAM_DATA_INTERVAL_TIMEOUT=900
GATEWAY_STREAM_KEEPALIVE_INTERVAL=10
GATEWAY_OPENAI_WS_READ_TIMEOUT_SECONDS=1800
GATEWAY_IMAGE_STREAM_DATA_INTERVAL_TIMEOUT=900
GATEWAY_IMAGE_STREAM_KEEPALIVE_INTERVAL=10
```

未在现场 `.env` 显式设置的请求头和纯文本请求限制继续使用程序默认值：
`server.max_header_bytes=65536`、`server.read_header_timeout=10`，
以及 `gateway.text_max_body_size=33554432`。

## 当前生效的 Nginx

站点配置位于 `/etc/nginx/conf.d/isacai.conf`。`nginx.conf` 保持 Amazon Linux
包默认值，其中 `worker_processes auto`、`worker_connections 1024`、
`keepalive_timeout 65`、`types_hash_max_size 4096`。当前站点配置为：

```nginx
server {
    listen 80;
    server_name isacai.space www.isacai.space;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name isacai.space www.isacai.space;

    ssl_certificate /etc/letsencrypt/live/isacai.space/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/isacai.space/privkey.pem;

    client_max_body_size 50M;
    proxy_read_timeout 900s;
    proxy_connect_timeout 75s;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

因此当前公网有效请求体上限是 **50 MiB**，不是应用层配置的 256 MiB；请求先被
Nginx 限制。当前上游读取超时为 **900 秒**，建连超时为 **75 秒**。

仓库的 `deploy/nginx-isacai.example.conf` 是迁移后的推荐目标，它有意使用
`308`、HTTP/2、`proxy_http_version 1.1`、下划线请求头支持，以及
`256m / 3600s / 3600s`。如果必须完全复制旧机行为，使用上面的现场配置；
如果需要完整 256 MiB 上传和更长的流式请求，使用仓库模板。切换前必须明确选择，
不要误把两套数值混用。

## TLS 与自动续期

当前证书目录为 `/etc/letsencrypt/live/isacai.space/`，证书类型 ECDSA。
现场检查发现证书只包含 `isacai.space`，但 Nginx 还声明了
`www.isacai.space`。新机必须二选一：

```bash
# www DNS 确实指向本机时，同时签发两个名称
sudo certbot --nginx -d isacai.space -d www.isacai.space

# 不使用 www 时，只签发主域名，并从 Nginx server_name 删除 www
sudo certbot --nginx -d isacai.space
```

现场还没有 Certbot systemd timer 或 cron，不能把这一缺口复制到新机。证书签发
成功后创建以下续期任务：

```bash
sudo tee /etc/systemd/system/certbot-renew.service >/dev/null <<'EOF'
[Unit]
Description=Renew Let's Encrypt certificates
After=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/bin/certbot renew --quiet --deploy-hook "/usr/bin/systemctl reload nginx"
EOF

sudo tee /etc/systemd/system/certbot-renew.timer >/dev/null <<'EOF'
[Unit]
Description=Run Certbot renewal twice daily

[Timer]
OnCalendar=*-*-* 00,12:00:00
RandomizedDelaySec=3600
Persistent=true

[Install]
WantedBy=timers.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now certbot-renew.timer
sudo certbot renew --dry-run
```

迁移证书时必须整体传输 `/etc/letsencrypt`，因为 `live/` 内是符号链接；临时包
包含私钥，只能用 `0600` 权限传输，验证后立即删除，永远不要提交 Git。也可以在
DNS 指向新机且 80 端口可访问后重新签发。

## 新机验收

```bash
cd /home/ec2-user/sub2api/deploy
sudo docker compose \
  -f docker-compose.local.yml \
  -f docker-compose.build.yml ps
curl -fsS http://127.0.0.1:8080/health
sudo nginx -t
curl -I http://isacai.space
curl -fsS https://isacai.space/health
sudo systemctl status certbot-renew.timer --no-pager
sudo certbot renew --dry-run
```

三容器必须 healthy，HTTP 必须跳转 HTTPS，正式域名健康检查和证书续期模拟必须
成功。新机产生写入后若需回滚，先停新机并把最新数据反向迁回，不能只切 DNS。
