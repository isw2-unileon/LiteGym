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

  it("generates an AI routine preview and saves it after confirmation", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse([]))
      .mockResolvedValueOnce(
        jsonResponse({
          routine_json: {
            name: "AI Strength",
            objective: "Ganar masa muscular",
            duration_minutes: 45,
            target_muscles: ["legs", "back"],
            mandatory_count: 0,
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
            mandatory_count: 0,
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
          mandatory_exercise_ids: [],
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
