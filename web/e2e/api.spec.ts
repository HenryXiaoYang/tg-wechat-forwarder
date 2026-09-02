import { expect, test } from "@playwright/test"
import { credentials } from "./helpers"

test("enforces auth and returns stable empty collection contracts", async ({ request }) => {
  const health = await request.get("/api/health")
  expect(health.ok()).toBeTruthy()
  expect(health.headers()["cache-control"]).toBe("no-store")
  expect(await health.json()).toMatchObject({ status: "ok" })

  const unauthorized = await request.get("/api/dashboard")
  expect(unauthorized.status()).toBe(401)

  const login = await request.post("/api/login", { data: credentials })
  expect(login.ok()).toBeTruthy()

  const dashboard = await request.get("/api/dashboard")
  expect(dashboard.ok()).toBeTruthy()
  const body = await dashboard.json()
  expect(body.dialogs).toEqual([])
  expect(body.filterRules).toEqual([])
  expect(body.deliveries).toEqual([])
  expect(dashboard.headers()["cache-control"]).toBe("no-store")
})

test("validates integration settings at the API boundary", async ({ request }) => {
  expect((await request.post("/api/login", { data: credentials })).ok()).toBeTruthy()

  const badRegex = await request.put("/api/settings/filters", { data: { rules: ["re:["] } })
  expect(badRegex.status()).toBe(400)
  expect((await badRegex.json()).error).toContain("规则")

  const badImageHost = await request.put("/api/settings/chevereto", { data: { baseUrl: "not-a-url", apiKey: "key" } })
  expect(badImageHost.status()).toBe(400)
  expect((await badImageHost.json()).error).toContain("http(s)")

  expect((await request.put("/api/settings/filters", { data: { rules: ["推广", "re:(?i)promo"] } })).status()).toBe(204)
  expect((await request.put("/api/settings/chevereto", { data: { baseUrl: "https://img.example.com", apiKey: "e2e-image-key" } })).status()).toBe(204)
  const saved = await (await request.get("/api/dashboard")).json()
  expect(saved.filterRules).toEqual(["推广", "re:(?i)promo"])
  expect(saved.chevereto).toEqual({ baseUrl: "https://img.example.com", keyConfigured: true })

  const emptyMessage = await request.post("/api/push/message", { data: { content: "  " } })
  expect(emptyMessage.status()).toBe(400)
  const unconfiguredPush = await request.post("/api/push/message", { data: { content: "来自控制台" } })
  expect(unconfiguredPush.status()).toBe(400)
  expect((await unconfiguredPush.json()).error).toContain("PushPlus")

  expect((await request.put("/api/settings/filters", { data: { rules: [] } })).status()).toBe(204)
  expect((await request.put("/api/settings/chevereto", { data: { baseUrl: "", clearKey: true } })).status()).toBe(204)
})
