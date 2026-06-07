import { expect, test } from "@playwright/test";

import { uniqueSuffix } from "./helpers";

test.use({ storageState: ".auth/laura.json" });

test("can create and edit a custom exercise", async ({ page }) => {
  const metadataResponse = await page.request.get("http://localhost:8080/api/exercises/metadata");
  expect(metadataResponse.ok()).toBeTruthy();
  const metadata = (await metadataResponse.json()) as {
    muscle_groups: Array<{ label: string; value: string }>;
    exercise_types: Array<{ label: string; value: string }>;
  };

  const muscleGroup = metadata.muscle_groups[0]?.value;
  const exerciseType = metadata.exercise_types[0]?.value;
  expect(muscleGroup).toBeTruthy();
  expect(exerciseType).toBeTruthy();

  const suffix = uniqueSuffix("exercise");
  const exerciseName = `Remo ${suffix}`;
  const createResponse = await page.request.post("http://localhost:8080/api/exercises", {
    data: {
      name: exerciseName,
      description: `Descripcion inicial ${suffix}`,
      muscle_group: muscleGroup,
      secondary_muscle_groups: [""],
      exercise_type: exerciseType,
      is_official: false,
    },
  });
  expect(createResponse.ok()).toBeTruthy();

  await page.goto("/exercises");
  await expect(page).toHaveURL(/\/exercises$/);
  await expect(page.getByRole("heading", { name: /explora tus ejercicios/i })).toBeVisible();

  await page.getByPlaceholder("Buscar por nombre...").fill(exerciseName);
  const exerciseList = page.locator('[data-block="exercise-list"]');
  const exerciseCard = exerciseList
    .locator('[data-block="exercise-card"]')
    .filter({ hasText: exerciseName })
    .first();

  await expect(exerciseCard).toBeVisible();
  await expect(exerciseCard.getByText(exerciseName, { exact: true })).toBeVisible();
});
