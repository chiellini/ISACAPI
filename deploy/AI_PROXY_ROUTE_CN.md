# 机器1通过机器2代理访问 AI 上游

这个脚本用于在**机器1**上配置代理出口，让 Claude / Anthropic、Grok / xAI、Gemini / Google、OpenAI 等客户端请求通过**机器2**出去。机器2需要已经能自然访问这些上游域名，并且已经运行 HTTP 或 SOCKS 代理服务。

## 机器2要求

机器2需要开放一个代理端口，例如：

- CCProxy HTTP 代理：常见端口 `808`
- Clash / mihomo mixed-port：常见端口 `7890`
- sing-box mixed/http/socks 入站：按你的配置填写
- Squid / 3proxy：按你的配置填写

确认机器1能连到机器2的代理端口：

```bash
nc -vz 机器2IP 代理端口
```

## 一键配置机器2

在机器2上进入 ISACAPI 目录运行：

```bash
chmod +x deploy/setup-machine2-ai-proxy-server.sh
sudo ./deploy/setup-machine2-ai-proxy-server.sh --allow-client 机器1IP --proxy-port 7890
```

这个脚本会安装并配置 `tinyproxy`，只允许 `--allow-client` 指定的机器1 IP/CIDR 访问代理，避免机器2变成公网开放代理。

如果机器2有防火墙，并且你希望脚本自动开放端口：

```bash
sudo ./deploy/setup-machine2-ai-proxy-server.sh --allow-client 机器1IP --proxy-port 7890 --open-firewall
```

如果想安装后顺便从机器2本机测试代理能否连到 Anthropic：

```bash
sudo ./deploy/setup-machine2-ai-proxy-server.sh --allow-client 机器1IP --proxy-port 7890 --check
```

## 一键配置机器1

在机器1上进入 ISACAPI 目录运行：

```bash
chmod +x deploy/configure-machine1-ai-proxy-route.sh
./deploy/configure-machine1-ai-proxy-route.sh --machine2-ip 机器2IP --proxy-port 7890 --proxy-type http
```

如果机器2是 CCProxy HTTP 代理，通常类似：

```bash
./deploy/configure-machine1-ai-proxy-route.sh --machine2-ip 机器2IP --proxy-port 808 --proxy-type http
```

如果机器2是 SOCKS5：

```bash
./deploy/configure-machine1-ai-proxy-route.sh --machine2-ip 机器2IP --proxy-port 1080 --proxy-type socks5h
```

## Claude Code 注意

脚本会把代理变量写入 `~/.claude/settings.json`，这样 Claude Code 会通过机器2代理出站。

如果你当前 `~/.claude/settings.json` 里还有：

```json
"ANTHROPIC_BASE_URL": "https://..."
```

Claude 访问的就不是官方 Anthropic 域名，而是这个自定义地址。若你想让 Claude Code 请求官方 Anthropic，再通过机器2出去，运行时加：

```bash
./deploy/configure-machine1-ai-proxy-route.sh --machine2-ip 机器2IP --proxy-port 7890 --proxy-type http --official-upstream
```

如果原来的 `ANTHROPIC_AUTH_TOKEN` / `ANTHROPIC_API_KEY` 是旧网关的 key，也可以同时清掉：

```bash
./deploy/configure-machine1-ai-proxy-route.sh --machine2-ip 机器2IP --proxy-port 7890 --proxy-type http --official-upstream --clear-claude-auth
```

## 当前 Shell 立即生效

脚本会写入：

```text
~/.config/isacapi/ai-proxy-env.sh
```

当前终端要立即生效：

```bash
source ~/.config/isacapi/ai-proxy-env.sh
```

新开的 Bash / Zsh 会自动加载。

## PAC 分流文件

脚本也会生成：

```text
~/.config/isacapi/ai-proxy.pac
```

支持 PAC 的应用可以使用这个文件做域名分流：匹配 AI 相关域名走机器2，其它域名直连。

命令行工具通常不读取 PAC，而是读取 `HTTP_PROXY` / `HTTPS_PROXY` / `ALL_PROXY` 环境变量。
