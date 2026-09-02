import type { Page } from "@playwright/test"

export const credentials = {
  username: "e2e-admin",
  password: "e2e-password-123",
}

export const dashboardFixture = {
  version: "e2e",
  telegram: { status: "connected", account: "Lin (@lin)" },
  dialogs: [
    { peerKey: "channel:1", kind: "channel", title: "科技前沿频道", subtitle: "频道 · @future", selectable: true, selected: true, adFilter: true, archived: false },
    { peerKey: "chat:2", kind: "group", title: "Go 开发者社区", subtitle: "群组", selectable: true, selected: false, adFilter: false, archived: false },
    { peerKey: "user:3", kind: "user", title: "Alice Chen", subtitle: "私聊 · @alice", selectable: false, selected: false, adFilter: false, archived: false },
    { peerKey: "user:4", kind: "user", title: "Release Bot", subtitle: "机器人 · @releasebot", selectable: false, selected: false, adFilter: false, archived: false },
    { peerKey: "channel:5", kind: "channel", title: "产品公告归档", subtitle: "频道 · @archive", selectable: true, selected: true, adFilter: false, archived: true },
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
