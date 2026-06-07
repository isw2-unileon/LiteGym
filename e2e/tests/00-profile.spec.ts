import { expect, test } from "@playwright/test";
import path from "node:path";

import { uniqueSuffix } from "./helpers";

test("profile goals can be updated and persist after reload", async ({ browser }) => {
  const authStatePath = path.resolve(__dirname, "..", ".auth", "laura.json");
  const context = await browser.newContext({ storageState: authStatePath });
  const page = await context.newPage();

  try {
    await page.goto("/profile");
    await expect(page).toHaveURL(/\/profile$/);
    await expect(page.getByRole("heading", { name: /hola,\s*laura/i })).toBeVisible({
      timeout: 10000,
    });
    await expect(page.getByText(/email: laura@example\.com/i)).toBeVisible();
    await expect(page.getByRole("button", { name: /actualizar metas/i })).toBeVisible();

    const shortTerm = page.locator('input[type="text"]').first();
    const targetDays = page.locator('input[type="number"]').first();
    const longTerm = page.locator('input[type="text"]').last();

    const updatedShortTerm = `Bajar grasa ${uniqueSuffix("perfil")}`;
    const updatedLongTerm = `Ganar masa ${uniqueSuffix("perfil")}`;

    await shortTerm.fill(updatedShortTerm);
    await targetDays.fill("5");
    await longTerm.fill(updatedLongTerm);

    page.once("dialog", async (dialog) => {
      expect(dialog.message()).toMatch(/metas guardadas correctamente/i);
      await dialog.accept();
    });
    await page.getByRole("button", { name: /actualizar metas/i }).click();

    await page.reload();
    await expect(shortTerm).toHaveValue(updatedShortTerm, { timeout: 10000 });
    await expect(targetDays).toHaveValue("5", { timeout: 10000 });
    await expect(longTerm).toHaveValue(updatedLongTerm, { timeout: 10000 });
  } finally {
    await context.close();
  }
});
