# Telegram → 微信

用一个 Telegram 用户账号，把选中的群组和频道的新消息转发到微信（PushPlus）。Go 后端 + React 控制台，前端资源嵌入同一个可执行文件。

## 准备

1. 在 [my.telegram.org/apps](https://my.telegram.org/apps) 创建应用，拿到 `api_id` 和 `api_hash`。
2. PushPlus 完成实名认证，拿到用户 Token 和 SecretKey。
3. 一个 Chevereto 兼容图床和 API Key（转发图片用）。

## 运行

```bash
cp .env.example .env    # 填 APP_SECRET、TELEGRAM_API_ID、TELEGRAM_API_HASH
docker compose pull                      # 或 docker compose build 从源码构建
docker compose run --rm forwarder hash   # 生成 ADMIN_PASSWORD_HASH，见下一节
docker compose up -d
```

镜像发布在 GitHub Packages：`ghcr.io/henryxiaoyang/tg-wechat-forwarder`，含 amd64 和 arm64。主分支是 `main`（compose 默认拉这个），打 `v*` 标签才会发布版本号和 `latest`。

### 生成密码哈希

`.env` 只存密码的 bcrypt 哈希，不存明文。任选一种方式：

```bash
docker compose run --rm forwarder hash   # 用镜像
go run ./cmd/forwarder hash              # 本地装了 Go
./forwarder hash                         # 用 release 里的二进制
```

按提示输入密码后会打印一行哈希。想让密码不回显、也不进 shell 历史：

```bash
read -rs PW && printf '%s\n' "$PW" | docker compose run --rm -T forwarder hash; unset PW
```

把整行填进 `.env`，**必须用单引号包住**，否则 `$` 会被当成变量展开：

```
ADMIN_PASSWORD_HASH='$2a$10$CPd3G4hyL78j.wc4frIeFew0rdZ51lFTot9kVP6dynNZkqUsbCGTW'
```

打开 <http://127.0.0.1:8080> 登录，扫码接入 Telegram，在设置里填 PushPlus 和图床，然后勾选要转发的会话。

- 默认只监听回环。要从别的机器访问，请在前面用 Caddy/Nginx 提供 HTTPS，或把 `.env` 的 `BIND_ADDR` 改成 `0.0.0.0`。
- 管理员密码只以 bcrypt 哈希形式存在 `.env` 里，改密码重新跑一次 `forwarder hash` 即可（会顺带让已有会话全部失效）。
- 所有配置在 `./data/data.db`。Telegram 会话、第三方密钥和消息内容用 `APP_SECRET` 加密，**请备份且不要更改**，否则只能重新登录和配置。

## 转发约定

- 只转发勾选之后收到的新消息，不回放历史、不处理编辑和删除。
- 图片和静态贴纸经图床后用 HTML 模板推送；视频、语音、文件、GIF 不转发。
- 每个来源可单独开启广告过滤，规则支持关键词和 `re:` 前缀的正则。
- PushPlus 免费账号限 5 条/分钟，程序逐条排队（约 13 秒一条），等待超过 10 分钟丢弃。

## 开发

需要 Go 1.25 和 Node.js 24：

```bash
cd web && npm ci && npm run build && cd ..
go test ./...
CGO_ENABLED=0 go build -trimpath -o forwarder ./cmd/forwarder
```

E2E（会启动真实后端，不访问第三方账号）：

```bash
cd web && npx playwright install chromium && npm run test:e2e
```

`v*` 标签会由 GitHub Actions 发布各平台二进制。
