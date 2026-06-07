import { expect, test } from "@playwright/test";

test.use({ storageState: ".auth/laura.json" });

test("support page opens while authenticated", async ({ page }) => {
  await page.goto("/support");
  await expect(page).toHaveURL(/\/support$/);
});
