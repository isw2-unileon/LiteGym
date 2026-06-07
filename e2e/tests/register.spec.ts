import crypto from "node:crypto";

import { expect, test } from "@playwright/test";

test("can register through the UI", async ({ page }) => {
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
  const registerResponsePromise = page.waitForResponse(
    (response) =>
      response.url().includes("/api/auth/register") &&
      response.request().method() === "POST",
  );
  await page.getByRole("button", { name: /crear cuenta/i }).click();

  const registerResponse = await registerResponsePromise;
  expect(registerResponse.ok()).toBeTruthy();
});

test("shows an error when registering with an existing email", async ({ page }) => {
  await page.goto("/register");

  await page.getByRole("textbox", { name: /nombre de usuario/i }).fill(`e2e-register-${Date.now()}`);
  await page.getByRole("textbox", { name: /email/i }).fill("marta@example.com");
  await page.getByLabel(/^contrasena$/i).first().fill("Password123!");
  await page.getByLabel(/repite la contrasena/i).fill("Password123!");
  await page.getByRole("button", { name: /crear cuenta/i }).click();

  await expect(page).toHaveURL(/\/register$/);
  await expect(
    page.getByText(/ya existe un usuario con ese correo electr[oó]nico/i),
  ).toBeVisible();
});
