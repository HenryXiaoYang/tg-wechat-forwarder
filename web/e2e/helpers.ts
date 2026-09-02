import type { Page } from "@playwright/test"

export const credentials = {
  username: "e2e-admin",
  password: "e2e-password-123",
}

export const dashboardFixture = {
  version: "e2e",
  telegram: { status: "connected", account: "Lin (@lin)" },
  dialogs: [
    { peerKey: "channel:1", kind: "channel", title: "科技前沿频道", subtitle: "频道", username: "@future", lastSender: "Alex", lastText: "今晚发版", selectable: true, selected: true, adFilter: true, archived: false, pinned: true },
    { peerKey: "chat:2", kind: "group", title: "Go 开发者社区", subtitle: "超级群组", username: "@golang", lastSender: "Wei", lastText: "[图片]", selectable: true, selected: false, adFilter: false, archived: false, pinned: false },
    { peerKey: "user:3", kind: "user", title: "Alice Chen", subtitle: "私聊", username: "@alice", lastSender: "你", lastText: "晚点回你", selectable: false, selected: false, adFilter: false, archived: false, pinned: false },
    { peerKey: "user:4", kind: "user", title: "Release Bot", subtitle: "机器人", username: "@releasebot", lastText: "v1.2.0 已发布", selectable: false, selected: false, adFilter: false, archived: false, pinned: false },
    { peerKey: "channel:5", kind: "channel", title: "产品公告归档", subtitle: "频道", username: "@archive", lastText: "本周更新", selectable: true, selected: true, adFilter: false, archived: true, pinned: false },
  ],
  pushplus: { topic: "ops", tokenConfigured: true, secretConfigured: true },
  chevereto: { baseUrl: "https://img.example.com", keyConfigured: true },
  filterRules: ["推广", "re:(?i)promo"],
  deliveries: [
    { shortCode: "a1", title: "科技前沿频道 · Alex", content: "新版本已经发布", time: "09-02 00:31", status: "sent" },
    { shortCode: "a2", title: "产品公告归档 · 编辑部", content: "本周更新<br><img src=\"https://img.example.com/a.jpg\">", time: "09-02 00:12", status: "accepted" },
  ],
  pushStatus: "ready",
}

export async function login(page: Page) {
  await page.goto("/")
  await page.getByLabel("用户名").fill(credentials.username)
  await page.getByLabel("密码").fill(credentials.password)
  await page.getByRole("button", { name: "登录" }).click()
  await page.getByRole("region", { name: "Telegram 会话" }).waitFor()
}

export async function mockDashboard(page: Page) {
  let requests = 0
  await page.route("**/api/dashboard", async (route) => {
    requests++
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(dashboardFixture) })
  })
  return () => requests
}
