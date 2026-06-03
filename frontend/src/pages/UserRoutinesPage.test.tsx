import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import UserRoutinesPage from "./UserRoutinesPage";

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: {
      "Content-Type": "application/json",
    },
    ...init,
  });
}

describe("UserRoutinesPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("shows saved routines from the API", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse([
          {
            id: "routine-1",
            name: "Upper Strength",
            description: "Trabajo principal de torso",
            exercise_count: 5,
            updated_at: "2026-05-24T10:00:00Z",
          },
        ]),
      ),
    );

    render(<UserRoutinesPage />);

    expect(await screen.findByRole("heading", { name: "Upper Strength" })).toBeInTheDocument();
    expect(screen.getByText("Trabajo principal de torso")).toBeInTheDocument();
    expect(screen.getByText("5 ejercicios")).toBeInTheDocument();
  });

  it("loads routine details when a routine is selected", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse([
          {
            id: "routine-1",
            name: "Upper Strength",
            description: "Trabajo principal de torso",
            exercise_count: 1,
            updated_at: "2026-05-24T10:00:00Z",
          },
        ]),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          id: "routine-1",
          name: "Upper Strength",
          description: "Trabajo principal de torso",
          exercise_count: 1,
          source: "ai",
          created_at: "2026-05-24T10:00:00Z",
          updated_at: "2026-05-24T10:00:00Z",
          exercises: [
            {
              id: "routine-exercise-1",
              exercise_id: "exercise-1",
              name: "Bench Press",
              muscle_group: "chest",
              exercise_order: 1,
              sets: [
                {
                  id: "set-1",
                  set_number: 1,
                  target_reps_min: 8,
                  target_reps_max: 10,
                  target_weight_kg: 72.5,
                  target_rir: 2,
                  rest_seconds: 120,
                },
              ],
            },
          ],
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<UserRoutinesPage />);

    await user.click(await screen.findByRole("button", { name: /Upper Strength/i }));

    expect(await screen.findByRole("heading", { name: "Bench Press" })).toBeInTheDocument();
    expect(screen.getByText("8-10")).toBeInTheDocument();
    expect(screen.getByText("72.5 kg")).toBeInTheDocument();
  });

  it("shows the AI upgrade action and opens its modal for a selected routine", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse([
          {
            id: "routine-1",
            name: "Upper Strength",
            description: "Trabajo principal de torso",
            exercise_count: 2,
            updated_at: "2026-05-24T10:00:00Z",
          },
        ]),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          id: "routine-1",
          name: "Upper Strength",
          description: "Trabajo principal de torso",
          exercise_count: 2,
          source: "manual",
          created_at: "2026-05-24T10:00:00Z",
          updated_at: "2026-05-24T10:00:00Z",
          exercises: [
            {
              id: "routine-exercise-1",
              exercise_id: "exercise-1",
              name: "Bench Press",
              muscle_group: "chest",
              exercise_order: 1,
              sets: [],
            },
            {
              id: "routine-exercise-2",
              exercise_id: "exercise-2",
              name: "Row",
              muscle_group: "back",
              exercise_order: 2,
              sets: [],
            },
          ],
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<UserRoutinesPage />);

    await user.click(await screen.findByRole("button", { name: /Upper Strength/i }));
    await user.click(await screen.findByRole("button", { name: "Mejorar con IA" }));

    expect(await screen.findByRole("heading", { name: "Mejorar Upper Strength" })).toBeInTheDocument();
    expect(screen.getByText("2 ejercicios cargados para preparar la mejora.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Preparar mejora" })).toBeInTheDocument();
  });

  it("requests an AI routine upgrade and shows the preview diff", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse([
          {
            id: "routine-1",
            name: "Upper Strength",
            description: "Trabajo principal de torso",
            exercise_count: 2,
            updated_at: "2026-05-24T10:00:00Z",
          },
        ]),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          id: "routine-1",
          name: "Upper Strength",
          description: "Trabajo principal de torso",
          exercise_count: 2,
          source: "manual",
          created_at: "2026-05-24T10:00:00Z",
          updated_at: "2026-05-24T10:00:00Z",
          exercises: [
            {
              id: "routine-exercise-1",
              exercise_id: "exercise-1",
              name: "Bench Press",
              muscle_group: "chest",
              exercise_order: 1,
              sets: [],
            },
            {
              id: "routine-exercise-2",
              exercise_id: "exercise-2",
              name: "Row",
              muscle_group: "back",
              exercise_order: 2,
              sets: [],
            },
          ],
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          routine_json: {
            name: "Upper Strength Plus",
            objective: "Equilibrar la sesion y subir intensidad",
            duration_minutes: 50,
            target_muscles: ["chest", "back", "shoulders"],
            mandatory_count: 1,
            generated_at: "2026-05-24T10:00:00Z",
            generation_source: "gemini",
            exercises: [
              {
                exercise_id: "exercise-1",
                name: "Bench Press",
                muscle_group: "chest",
                exercise_type: "strength",
                is_mandatory: true,
                sets: [
                  {
                    set_number: 1,
                    target_reps_min: 5,
                    target_reps_max: 7,
                    target_weight_kg: 77.5,
                  },
                ],
              },
              {
                exercise_id: "exercise-3",
                name: "Shoulder Press",
                muscle_group: "shoulders",
                exercise_type: "strength",
                is_mandatory: false,
                sets: [],
              },
            ],
          },
          diff: {
            summary: {
              added_exercises: 1,
              removed_exercises: 1,
              modified_exercises: 1,
              unchanged_exercises: 0,
            },
            exercises: [
              {
                change_type: "modified",
                exercise_id: "exercise-1",
                before_name: "Bench Press",
                after_name: "Bench Press",
                before_order: 1,
                after_order: 1,
                before_muscle_group: "chest",
                after_muscle_group: "chest",
                sets: [
                  {
                    change_type: "modified",
                    set_number: 1,
                    before: {
                      set_number: 1,
                      target_reps_min: 6,
                      target_reps_max: 8,
                    },
                    after: {
                      set_number: 1,
                      target_reps_min: 5,
                      target_reps_max: 7,
                      target_weight_kg: 77.5,
                    },
                  },
                ],
              },
              {
                change_type: "removed",
                exercise_id: "exercise-2",
                before_name: "Row",
                before_order: 2,
                before_muscle_group: "back",
                sets: [],
              },
              {
                change_type: "added",
                exercise_id: "exercise-3",
                after_name: "Shoulder Press",
                after_order: 2,
                after_muscle_group: "shoulders",
                sets: [],
              },
            ],
          },
          rate_limit: {
            remaining: 1,
            reset_at: "2026-05-24T11:00:00Z",
          },
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<UserRoutinesPage />);

    await user.click(await screen.findByRole("button", { name: /Upper Strength/i }));
    await user.click(await screen.findByRole("button", { name: "Mejorar con IA" }));
    await user.clear(screen.getByLabelText("Instrucciones para la IA"));
    await user.type(
      screen.getByLabelText("Instrucciones para la IA"),
      "Sube intensidad y añade hombro.",
    );
    await user.click(screen.getByRole("button", { name: "Preparar mejora" }));

    expect(await screen.findAllByRole("heading", { name: "Upper Strength Plus" })).toHaveLength(2);
    expect(screen.getAllByText("Cambios detectados")).toHaveLength(1);
    expect(screen.getByText("Añadidos")).toBeInTheDocument();
    expect(screen.getAllByText("Shoulder Press")).toHaveLength(2);
    expect(screen.getByText(/Quedan 1 intentos en esta hora/i)).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/api/routines/routine-1/ai/upgrade"),
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          message: "Sube intensidad y añade hombro.",
        }),
      }),
    );
  });

  it("regenerates the AI upgrade proposal with feedback", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse([
          {
            id: "routine-1",
            name: "Upper Strength",
            description: "Trabajo principal de torso",
            exercise_count: 2,
            updated_at: "2026-05-24T10:00:00Z",
          },
        ]),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          id: "routine-1",
          name: "Upper Strength",
          description: "Trabajo principal de torso",
          exercise_count: 2,
          source: "manual",
          created_at: "2026-05-24T10:00:00Z",
          updated_at: "2026-05-24T10:00:00Z",
          exercises: [
            {
              id: "routine-exercise-1",
              exercise_id: "exercise-1",
              name: "Bench Press",
              muscle_group: "chest",
              exercise_order: 1,
              sets: [],
            },
            {
              id: "routine-exercise-2",
              exercise_id: "exercise-2",
              name: "Row",
              muscle_group: "back",
              exercise_order: 2,
              sets: [],
            },
          ],
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          routine_json: {
            name: "Upper Strength Plus",
            objective: "Equilibrar la sesion y subir intensidad",
            duration_minutes: 50,
            target_muscles: ["chest", "back", "shoulders"],
            mandatory_count: 1,
            generated_at: "2026-05-24T10:00:00Z",
            generation_source: "gemini",
            exercises: [],
          },
          diff: {
            summary: {
              added_exercises: 1,
              removed_exercises: 0,
              modified_exercises: 1,
              unchanged_exercises: 0,
            },
            exercises: [],
          },
          rate_limit: {
            remaining: 1,
            reset_at: "2026-05-24T11:00:00Z",
          },
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          routine_json: {
            name: "Upper Strength Refined",
            objective: "Mantener intensidad con menos volumen",
            duration_minutes: 48,
            target_muscles: ["chest", "back", "shoulders"],
            mandatory_count: 1,
            generated_at: "2026-05-24T10:15:00Z",
            generation_source: "gemini",
            exercises: [],
          },
          diff: {
            summary: {
              added_exercises: 0,
              removed_exercises: 0,
              modified_exercises: 2,
              unchanged_exercises: 0,
            },
            exercises: [],
          },
          rate_limit: {
            remaining: 0,
            reset_at: "2026-05-24T11:00:00Z",
          },
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<UserRoutinesPage />);

    await user.click(await screen.findByRole("button", { name: /Upper Strength/i }));
    await user.click(await screen.findByRole("button", { name: "Mejorar con IA" }));
    await user.clear(screen.getByLabelText("Instrucciones para la IA"));
    await user.type(
      screen.getByLabelText("Instrucciones para la IA"),
      "Sube intensidad y añade hombro.",
    );
    await user.click(screen.getByRole("button", { name: "Preparar mejora" }));

    await user.type(
      screen.getByLabelText("Ajustar la propuesta"),
      "Recorta una serie y mantén el press principal.",
    );
    await user.click(screen.getByRole("button", { name: "Regenerar propuesta" }));

    expect(await screen.findAllByRole("heading", { name: "Upper Strength Refined" })).toHaveLength(2);
    expect(screen.getByText("Propuesta actualizada.")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenLastCalledWith(
      expect.stringContaining("/api/routines/routine-1/ai/upgrade"),
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          message: "Sube intensidad y añade hombro.",
          feedback_message: "Recorta una serie y mantén el press principal.",
        }),
      }),
    );
  });

  it("generates an AI routine preview and saves it after confirmation", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse([]))
      .mockResolvedValueOnce(
        jsonResponse({
          items: [
            {
              id: "exercise-1",
              name: "Bench Press",
              description: "Press de pecho con barra",
              muscle_group: "chest",
              exercise_type: "strength",
            },
            {
              id: "exercise-2",
              name: "Squat",
              description: "Sentadilla libre",
              muscle_group: "legs",
              exercise_type: "strength",
            },
          ],
          page: 1,
          limit: 100,
          total: 2,
          total_pages: 1,
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          routine_json: {
            name: "AI Strength",
            objective: "Ganar masa muscular",
            duration_minutes: 45,
            target_muscles: ["legs", "back"],
            mandatory_count: 1,
            generated_at: "2026-05-24T10:00:00Z",
            generation_source: "gemini",
            exercises: [
              {
                exercise_id: "exercise-1",
                name: "Squat",
                muscle_group: "legs",
                exercise_type: "strength",
                is_mandatory: false,
                sets: [],
              },
            ],
          },
          rate_limit: {
            remaining: 1,
            reset_at: "2026-05-24T11:00:00Z",
          },
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          routine_id: "routine-ai-1",
          routine_json: {
            name: "AI Strength",
            objective: "Ganar masa muscular",
            duration_minutes: 45,
            target_muscles: ["legs", "back"],
            mandatory_count: 1,
            generated_at: "2026-05-24T10:00:00Z",
            generation_source: "gemini",
            exercises: [
              {
                exercise_id: "exercise-1",
                name: "Squat",
                muscle_group: "legs",
                exercise_type: "strength",
                is_mandatory: false,
                sets: [],
              },
            ],
          },
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse([
          {
            id: "routine-ai-1",
            name: "AI Strength",
            description: "Generated routine",
            exercise_count: 1,
            updated_at: "2026-05-24T10:00:00Z",
          },
        ]),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          id: "routine-ai-1",
          name: "AI Strength",
          description: "Generated routine",
          exercise_count: 1,
          source: "ai",
          created_at: "2026-05-24T10:00:00Z",
          updated_at: "2026-05-24T10:00:00Z",
          exercises: [
            {
              id: "routine-exercise-ai-1",
              exercise_id: "exercise-1",
              name: "Squat",
              muscle_group: "legs",
              exercise_order: 1,
              sets: [],
            },
          ],
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<UserRoutinesPage />);

    await user.click(screen.getByRole("button", { name: "Crear rutina con IA" }));
    await user.clear(screen.getByLabelText("Objetivo"));
    await user.type(screen.getByLabelText("Objetivo"), "Ganar masa muscular");
    await user.clear(screen.getByLabelText("Minutos"));
    await user.type(screen.getByLabelText("Minutos"), "45");
    await user.type(screen.getByLabelText("Musculos objetivo"), "legs, back");
    await user.type(
      screen.getByPlaceholderText(
        "Ej. prioriza press con barra, evita sentadillas traseras y mantén descansos largos.",
      ),
      "Prioriza press con barra y agrega trabajo de espalda.",
    );
    await user.click(
      await screen.findByRole("button", { name: /Bench Press/i }),
    );
    await user.click(screen.getByRole("button", { name: "Generar vista previa" }));

    expect(await screen.findByRole("heading", { name: "AI Strength" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Guardar rutina" }));

    expect(await screen.findByRole("heading", { name: "Squat" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/api/routines/ai/generate"),
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({
          objective: "Ganar masa muscular",
          duration_minutes: 45,
          target_muscle_groups: ["legs", "back"],
          mandatory_exercises: ["Bench Press"],
          notes: "Prioriza press con barra y agrega trabajo de espalda.",
        }),
      }),
    );
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/api/routines/ai/save"),
      expect.objectContaining({
        method: "POST",
      }),
    );
  });

  it("shows an empty state when the user has no routines", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse([])));

    render(<UserRoutinesPage />);

    expect(await screen.findByText("Todavia no tienes rutinas guardadas.")).toBeInTheDocument();
  });

  it("shows an error state when the API fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(jsonResponse({ error: "failed" }, { status: 500 })),
    );

    render(<UserRoutinesPage />);

    expect(await screen.findByText("No se pudieron cargar las rutinas.")).toBeInTheDocument();
  });
});
