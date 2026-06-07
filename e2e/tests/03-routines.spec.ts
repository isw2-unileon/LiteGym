import { expect, test } from "@playwright/test";

test.use({ storageState: ".auth/laura.json" });

test("can create a routine", async ({ page }) => {
  await page.goto("/routines");
  await expect(page).toHaveURL(/\/routines$/);
});
