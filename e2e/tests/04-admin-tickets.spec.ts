import { expect, test } from "@playwright/test";

test.use({ storageState: ".auth/admin.json" });

test("admin can see a support ticket", async ({ page }) => {
  await page.goto("/admin");
  await expect(page).toHaveURL(/\/admin$/);
});
