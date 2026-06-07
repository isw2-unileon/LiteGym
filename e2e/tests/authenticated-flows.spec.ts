import { expect, test, type Page } from "@playwright/test";

type TestUser = {
  email: string;
  password: string;
};

const seededUser: TestUser = {
  email: "diego@example.com",
  password: "1234",
};

async function loginWithUser(page: Page, user: TestUser) {
  await page.goto("/");

  await page.getByRole("textbox", { name: /email/i }).fill(user.email);
  await page.getByLabel(/contrasena/i).fill(user.password);
  await page.getByRole("button", { name: /iniciar sesion/i }).click();

  await expect(page).toHaveURL(/\/dashboard$/);
  await expect(page.getByText(/panel principal/i)).toBeVisible();
  await expect(page.getByRole("link", { name: /^Panel$/ })).toHaveAttribute("href", "/dashboard");
}

test("can log in to reach the dashboard", async ({ page }) => {
  await loginWithUser(page, seededUser);
});

test("can submit a support ticket while authenticated", async ({ page }) => {
  await loginWithUser(page, seededUser);

  await page.getByRole("link", { name: /soporte/i }).click();
  await expect(page).toHaveURL(/\/support$/);
  await expect(page.getByText(/centro de ayuda/i)).toBeVisible();
  await expect(page.getByText(/rellena un ticket/i)).toBeVisible();

  await page.locator("select").selectOption("IA");
  await page.getByPlaceholder("Ej: No me carga una rutina").fill(
    `E2E support ticket ${seededUser.email}`,
  );
  await page.getByPlaceholder("Explica detalladamente tu problema...").fill(
    "Prueba automatica e2e para verificar que el formulario de soporte funciona con un usuario autenticado.",
  );
  await page.getByRole("button", { name: /enviar ticket/i }).click();

  await expect(page.getByText(/ticket enviado correctamente/i)).toBeVisible();
});
