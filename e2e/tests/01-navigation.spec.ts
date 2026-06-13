import { expect, test } from "@playwright/test";

test.use({ storageState: ".auth/laura.json" });

test("the shared navigation moves between the main views", async ({ page }) => {
  await page.goto("/dashboard");
  await page.getByRole("link", { name: "Mis rutinas" }).click();
  await expect(page).toHaveURL(/\/routines$/);

  await page.getByRole("link", { name: "Mis ejercicios" }).click();
  await expect(page).toHaveURL(/\/exercises$/);

  await page.getByRole("link", { name: /soporte técnico/i }).click();
  await expect(page).toHaveURL(/\/support$/);

  await page.getByRole("link", { name: "Perfil" }).click();
  await expect(page).toHaveURL(/\/profile$/);
});
