import crypto from "node:crypto";

import { expect, test } from "@playwright/test";

test("can register through the UI and reach the dashboard", async ({ page }) => {
  const suffix = crypto.randomUUID().slice(0, 8);
  const username = `e2e-register-${suffix}`;
  const email = `${username}@example.com`;
  const password = "Password123!";

  await page.goto("/register");

  await expect(
    page.getByRole("heading", { name: /crea tu cuenta y empieza a entrenar con orden/i }),
  ).toBeVisible();
  await expect(page.getByRole("heading", { name: /abre tu cuenta/i })).toBeVisible();

  await page.getByRole("textbox", { name: /nombre de usuario/i }).fill(username);
  await page.getByRole("textbox", { name: /email/i }).fill(email);
  await page.getByLabel(/^contrasena$/i).first().fill(password);
  await page.getByLabel(/repite la contrasena/i).fill(password);
  await page.getByRole("button", { name: /crear cuenta/i }).click();

  await expect(page).toHaveURL(/\/dashboard$/);
  await expect(page.getByText(/panel principal/i)).toBeVisible();
  await expect(page.getByText(new RegExp(`hola,\\s*${username}`, "i"))).toBeVisible();
});
