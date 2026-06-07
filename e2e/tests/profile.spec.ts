import { expect, test } from "@playwright/test";

const seededUser = {
  email: "sergio@example.com",
  password: "1234",
};

test("can update profile goals and keep them after reload", async ({ page }) => {
  const shortTerm = `Bajar grasa ${Date.now()}`;
  const longTerm = `Ganar masa ${Date.now()}`;
  const targetDays = "5";

  await page.goto("/");
  await page.getByRole("textbox", { name: /email/i }).fill(seededUser.email);
  await page.getByLabel(/contrasena/i).fill(seededUser.password);
  await page.getByRole("button", { name: /iniciar sesion/i }).click();

  await expect(page).toHaveURL(/\/dashboard$/);
  await page.getByRole("link", { name: /perfil/i }).click();

  await expect(page).toHaveURL(/\/profile$/);
  await expect(page.getByText(/perfil de atleta/i)).toBeVisible();
  await expect(page.getByRole("heading", { name: /hola,\s*sergio/i })).toBeVisible();
  await expect(page.getByText(/email: sergio@example\.com/i)).toBeVisible();

  const profileTextInputs = page.locator('input[type="text"]');
  await profileTextInputs.nth(0).fill(shortTerm);
  await page.locator('input[type="number"]').fill(targetDays);
  await profileTextInputs.nth(1).fill(longTerm);

  const dialogPromise = page.waitForEvent("dialog");
  await page.getByRole("button", { name: /actualizar metas/i }).click();
  const dialog = await dialogPromise;
  expect(dialog.message()).toMatch(/metas guardadas correctamente/i);
  await dialog.accept();

  await page.reload();

  await expect(profileTextInputs.nth(0)).toHaveValue(shortTerm);
  await expect(profileTextInputs.nth(1)).toHaveValue(longTerm);
  await expect(page.locator('input[type="number"]')).toHaveValue(targetDays);
});
