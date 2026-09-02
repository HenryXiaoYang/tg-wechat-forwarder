import { expect, test } from "@playwright/test"
import { credentials } from "./helpers"

test("rejects invalid credentials with a useful error", async ({ page }) => {
  await page.goto("/")
  await page.getByLabel("用户名").fill(credentials.username)
  await page.getByLabel("密码").fill("definitely-wrong")
  const loginResponse = page.waitForResponse((response) => response.url().endsWith("/api/login"))
  await page.getByRole("button", { name: "登录" }).click()
  expect((await loginResponse).status()).toBe(401)
  await expect(page.getByText("用户名或密码错误")).toBeVisible()
  await expect(page.getByRole("heading", { name: "Telegram Forwarder" })).toBeVisible()
})

test("logs in through the real Go backend", async ({ page }) => {
  await page.goto("/")
  await page.getByLabel("用户名").fill(credentials.username)
  await page.getByLabel("密码").fill(credentials.password)
  await page.getByRole("button", { name: "登录" }).click()
  await expect(page.getByText("Telegram → 微信")).toBeVisible()
  await expect(page.getByRole("region", { name: "Telegram 会话" })).toBeVisible()
  await expect(page.getByRole("region", { name: "微信推送" })).toBeVisible()
})
