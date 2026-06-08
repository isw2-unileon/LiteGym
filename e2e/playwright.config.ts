import { defineConfig, devices } from "@playwright/test";

const browserProject = process.env.PLAYWRIGHT_BROWSER ?? "chromium";
const browserDevices: Record<string, (typeof devices)[keyof typeof devices]> = {
  chromium: devices["Desktop Chrome"],
  firefox: devices["Desktop Firefox"],
  webkit: devices["Desktop Safari"],
};

if (!browserDevices[browserProject]) {
  throw new Error(
    `Unsupported PLAYWRIGHT_BROWSER "${browserProject}". Use chromium, firefox, or webkit.`,
  );
}

export default defineConfig({
  testDir: "./tests",
  globalSetup: "./global.setup.ts",
  timeout: 30_000,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: "html",

  use: {
    baseURL: "http://localhost:5173",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },

  webServer: [
    {
      command: "go run ./backend/cmd/server",
      cwd: "..",
      port: 8080,
      reuseExistingServer: true,
    },
    {
      command: "cd frontend && npm run dev",
      cwd: "..",
      port: 5173,
      reuseExistingServer: true,
    },
  ],

  projects: [{ name: browserProject, use: { ...browserDevices[browserProject] } }],
});
