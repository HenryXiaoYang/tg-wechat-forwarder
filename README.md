# Telegram → 微信 PushPlus 转发器

使用一个 Telegram 用户账号的 MTProto 会话，选择性地把群组和频道新消息推送到微信。后端为 Go，控制面板为 React + Tailwind CSS + shadcn/ui，前端资源嵌入同一个可执行文件。

## 功能

- Telegram App 扫码登录，支持两步验证；加密保存单账号会话。
- 显示账号的云端对话；仅群组、超级群组、论坛和频道可勾选转发。
- 文本使用 PushPlus `txt` 模板；图片和静态贴纸经 Chevereto 兼容 API 上传后使用 `html` 模板。
- 视频、语音、文件、GIF 和动画贴纸不转发；不回放离线或重启期间的消息。
- 每个来源可单独开启广告过滤，规则支持关键词及 `re:` 前缀的 Go 正则表达式。
- PushPlus 一对一或 Topic 推送、测试消息、投递状态及远端历史。
- 单管理员用户名密码登录，亮色/暗色与移动端布局。
- 所有设置位于 `./data/data.db`；Telegram 会话、PushPlus 和图床密钥使用 AES-GCM 加密。

## 准备

1. 在 [Telegram API development tools](https://my.telegram.org/apps) 创建应用，取得 `api_id` 和 `api_hash`。二维码只负责授权用户账号，不能替代这两个应用凭据。
2. 在 PushPlus 完成实名认证，取得用户 Token 和 SecretKey。普通实名账号受每分钟 5 次、每天 200 次限制；程序逐条排队、不合并普通消息，等待超过 10 分钟便丢弃。
3. 准备一个 Chevereto V1 兼容上传接口和 API Key。Base URL 可填站点根地址、`/api/1` 或完整 `/api/1/upload`。

本服务只主动连接 Telegram、PushPlus 和 Chevereto，不需要公网地址。若从公网访问控制面板，请在前面使用 Caddy/Nginx 提供 HTTPS；否则管理员密码会在 HTTP 链路上明文传输。

## Docker Compose

```bash
cp .env.example .env
# 编辑 .env，至少替换管理员密码、APP_SECRET、TELEGRAM_API_ID 和 TELEGRAM_API_HASH
mkdir -p data
docker compose up -d --build
```

打开 `http://localhost:8080`，用 `.env` 中的管理员账号登录，然后：

1. 在左侧扫描 Telegram 二维码。
2. 在设置中填写 PushPlus、Chevereto 和广告规则并分别测试。
3. 勾选要转发的群组或频道；漏斗图标控制该来源是否过滤广告。

Linux 上若 `./data` 无法写入，把 `.env` 的 `PUID`/`PGID` 改为 `id -u`/`id -g` 的结果。

请备份并保持 `APP_SECRET` 不变；它用于解密 Telegram 会话与第三方密钥，丢失后只能重新登录和配置。

## 目录结构

```text
cmd/forwarder/  可执行程序入口
internal/app/   后端业务、协议接入、存储与测试
web/            React 源码、Playwright E2E 与嵌入式 dist
```

## 本地构建

需要 Go 1.25、Node.js 24：

```bash
cd web
npm ci
npm run build
cd ..
go test ./...
CGO_ENABLED=0 go build -trimpath -o forwarder ./cmd/forwarder
./forwarder
```

直接运行时会自动读取当前目录的 `.env`。GitHub Actions 会在 `v*` 标签上发布 Linux、macOS 和 Windows 的五种独立二进制，运行时不需要外部 static 目录。

## End-to-end 测试

Playwright 会启动真实 Go 后端，用虚构凭据验证认证和 API，再用浏览器 fixture 覆盖对话选择、设置轮询、移动标签与主题切换，不访问真实第三方账号：

```bash
cd web
npx playwright install chromium ffmpeg
npm run test:e2e
```

失败产物位于 `web/test-results/`，HTML 报告位于 `web/playwright-report/`。

## 转发语义

- 只处理勾选之后实时收到的新消息；忽略历史、编辑、删除、反应、服务消息及账号自己发出的消息。
- Telegram 内容保护开启时不提供勾选框，也不会复制其内容。
- 相册等待约 1.2 秒后合为一条 PushPlus 请求；其他消息不合并。
- 网络错误最多重试三次。PushPlus 鉴权、额度或内容错误不重试；其返回 `code=900` 后，官方 SDK 会在当天停止继续请求。
- SQLite 只临时保存加密的待发送队列，成功、永久失败或过期后立即删除；不保存本地消息、过滤或投递历史。
- 图片在 Chevereto 中永久保留，程序不保留本地图片副本。
