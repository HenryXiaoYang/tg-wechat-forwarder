import { defineConfig, devices } from "@playwright/test"

const externalBaseURL = process.env.E2E_BASE_URL
const baseURL = externalBaseURL || "http://127.0.0.1:18085"

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: [["list"], ["html", { open: "never" }]],
  use: {
    baseURL,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  webServer: externalBaseURL ? undefined : {
    command: "go run ../cmd/forwarder",
    url: `${baseURL}/api/health`,
    timeout: 120_000,
    reuseExistingServer: !process.env.CI,
    env: {
      ...process.env,
      LISTEN_ADDR: "127.0.0.1:18085",
      ADMIN_USERNAME: "e2e-admin",
      ADMIN_PASSWORD: "e2e-password-123",
      APP_SECRET: "e2e-only-secret-with-at-least-32-characters",
      TELEGRAM_API_ID: "12345",
      TELEGRAM_API_HASH: "not-a-real-telegram-api-hash",
    },
  },
})
