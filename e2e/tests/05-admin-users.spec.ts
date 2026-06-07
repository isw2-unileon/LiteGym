import { expect, test } from "@playwright/test";

test.use({ storageState: ".auth/admin.json" });

test("admin area opens", async ({ page }) => {
  await page.goto("/admin");
  await expect(page).toHaveURL(/\/admin$/);
});
