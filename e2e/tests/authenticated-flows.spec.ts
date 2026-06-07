import { expect, test, type Page } from "@playwright/test";

type TestUser = {
  email: string;
  password: string;
};

test.describe.configure({ mode: "serial" });

const seededUser: TestUser = {
  email: "laura@example.com",
  password: "1234",
};

async function loginWithUser(page: Page, user: TestUser) {
  await page.goto("/");

  await page.getByRole("textbox", { name: /email/i }).fill(user.email);
  await page.getByLabel(/contrasena/i).fill(user.password);
  await page.getByRole("button", { name: /iniciar sesion/i }).click();

  await expect(page).toHaveURL(/\/dashboard$/);
  await expect(page.getByRole("link", { name: /^Panel$/ })).toBeVisible();
  await expect(page.getByRole("link", { name: /^Panel$/ })).toHaveAttribute("href", "/dashboard");
}

async function submitLoginForm(page: Page, user: TestUser) {
  await page.goto("/");
  await page.getByRole("textbox", { name: /email/i }).fill(user.email);
  await page.getByLabel(/contrasena/i).fill(user.password);
  await page.getByRole("button", { name: /iniciar sesion/i }).click();
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

test("shows an error when login credentials are invalid", async ({ page }) => {
  const invalidLoginUser = {
    email: "diego@example.com",
    password: "wrong-password",
  };

  await submitLoginForm(page, invalidLoginUser);

  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByText(/el correo o la contrase[nñ]a no son correctos/i)).toBeVisible();
});

test("shows a rate limit error after repeated invalid login attempts", async ({ page }) => {
  const invalidLoginUser = {
    email: "diego@example.com",
    password: "wrong-password",
  };
  const invalidCredentialsMessage = page.getByText(
    /el correo o la contrase[nñ]a no son correctos/i,
  );
  const rateLimitMessage = page.getByText(
    /too many login attempts|iniciar sesion demasiadas veces|rate limit exceeded/i,
  );
  let rateLimited = false;

  for (let attempt = 0; attempt < 5; attempt += 1) {
    await submitLoginForm(page, invalidLoginUser);
    await expect(page).toHaveURL(/\/$/);

    const outcome = await Promise.race([
      rateLimitMessage
        .waitFor({ state: "visible", timeout: 5000 })
        .then(() => "rate"),
      invalidCredentialsMessage
        .waitFor({ state: "visible", timeout: 5000 })
        .then(() => "invalid"),
    ]);

    if (outcome === "rate") {
      rateLimited = true;
      break;
    }
  }

  expect(rateLimited).toBeTruthy();
});
