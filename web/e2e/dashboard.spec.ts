import { expect, test } from "@playwright/test"
import { dashboardFixture, login, mockDashboard } from "./helpers"

let dashboardRequests: () => number

test.beforeEach(async ({ page }) => {
  await login(page)
  dashboardRequests = await mockDashboard(page)
  await page.reload()
  await expect(page.getByText("科技前沿频道", { exact: true })).toBeVisible()
})

test("shows every dialog but only groups and channels are selectable", async ({ page }) => {
  await page.route("**/api/dialogs/**", (route) => route.fulfill({ status: 204 }))

  await expect(page.getByText("Alice Chen")).toBeVisible()
  await expect(page.getByText("Release Bot")).toBeVisible()
  await expect(page.getByRole("checkbox", { name: /^转发 / })).toHaveCount(2)
  await expect(page.getByRole("checkbox", { name: "转发 Alice Chen" })).toHaveCount(0)

  await expect(page.getByRole("checkbox", { name: "转发 产品公告归档" })).toHaveCount(0)
  await page.getByRole("button", { name: /已归档的聊天/ }).click()
  await expect(page.getByRole("checkbox", { name: "转发 产品公告归档" })).toBeChecked()

  const group = page.getByRole("checkbox", { name: "转发 Go 开发者社区" })
  await expect(group).not.toBeChecked()
  await group.check()
  await expect(group).toBeChecked()

  await page.getByPlaceholder("搜索对话").fill("Alice")
  await expect(page.getByText("Alice Chen")).toBeVisible()
  await expect(page.getByText("Go 开发者社区")).toBeHidden()
  await expect(page.getByRole("checkbox", { name: /^转发 / })).toHaveCount(0)
})

test("does not reset unsaved settings when dashboard polling refreshes", async ({ page }) => {
  await page.getByRole("button", { name: "设置" }).first().click()

  const topic = page.getByLabel("Topic")
  await topic.fill("typed-but-not-saved")
  const before = dashboardRequests()
  await expect.poll(dashboardRequests, { timeout: 8_000 }).toBeGreaterThan(before)
  await expect(topic).toHaveValue("typed-but-not-saved")
})

test("supports mobile tabs and theme switching", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(page.getByRole("region", { name: "Telegram 会话" })).toBeVisible()
  await expect(page.getByRole("region", { name: "微信推送" })).toBeHidden()

  await page.getByRole("button", { name: "微信推送" }).click()
  await expect(page.getByRole("region", { name: "微信推送" })).toBeVisible()
  await expect(page.getByText("新版本已经发布")).toBeVisible()

  const wasDark = await page.locator("html").evaluate((element) => element.classList.contains("dark"))
  await page.getByRole("button", { name: /切换到(亮色|暗色)/ }).click()
  await expect.poll(() => page.locator("html").evaluate((element) => element.classList.contains("dark"))).toBe(!wasDark)
})

test("shows an actionable Telegram connection timeout", async ({ page }) => {
  await page.unroute("**/api/dashboard")
  await page.route("**/api/dashboard", (route) => route.fulfill({
    status: 200,
    contentType: "application/json",
    body: JSON.stringify({
      ...dashboardFixture,
      telegram: { status: "error", error: "连接 Telegram 超时，请检查 TELEGRAM_API_ID、TELEGRAM_API_HASH 和服务器网络；后台仍会自动重试" },
      dialogs: [],
    }),
  }))
  await page.reload()
  await expect(page.getByRole("heading", { name: "连接异常" })).toBeVisible()
  await expect(page.getByText(/请检查 TELEGRAM_API_ID/)).toBeVisible()
})
