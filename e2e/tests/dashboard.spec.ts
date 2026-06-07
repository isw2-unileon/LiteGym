import { expect, test } from "@playwright/test";

test.use({ storageState: ".auth/laura.json" });

test("dashboard opens", async ({ page }) => {
  await page.goto("/dashboard");
  await expect(page).toHaveURL(/\/dashboard$/);
});
