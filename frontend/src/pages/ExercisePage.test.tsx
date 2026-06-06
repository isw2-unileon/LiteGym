import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import ExercisePage from "./ExercisePage";

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: {
      "Content-Type": "application/json",
    },
    ...init,
  });
}

function exerciseMetadataResponse() {
  return {
    exercise_types: [{ value: "strength", label: "Fuerza" }],
    muscle_groups: [
      { value: "chest", label: "Pecho" },
      { value: "triceps", label: "Tríceps" },
      { value: "shoulders", label: "Hombros" },
      { value: "front_delts", label: "Deltoides anteriores" },
    ],
  };
}

function exerciseListResponse(items: unknown[]) {
  return {
    items,
    page: 1,
    limit: 20,
    total: items.length,
    total_pages: items.length > 0 ? 1 : 0,
  };
}

function emptyExerciseInsightsResponse() {
  return {
    summary: {
      session_count: 0,
      set_count: 0,
      total_volume_kg: 0,
      trend: "empty",
    },
    best_set: null,
    personal_records: [],
    progression: [],
    history: [],
  };
}

function renderExercisePage() {
  return render(
    <MemoryRouter>
      <ExercisePage />
    </MemoryRouter>,
  );
}

describe("ExercisePage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("opens the modal to create a new exercise", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(
          jsonResponse(
            { user: { id: "1", email: "admin@example.com", username: "admin", role: "admin" } },
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(jsonResponse(exerciseMetadataResponse(), { status: 200 }))
        .mockResolvedValueOnce(jsonResponse(exerciseListResponse([]), { status: 200 })),
    );
    const user = userEvent.setup();

    renderExercisePage();

    await screen.findByRole("button", { name: "Crear nuevo ejercicio" });
    await user.click(screen.getByRole("button", { name: "Crear nuevo ejercicio" }));

    expect(
      screen.getByRole("heading", {
        name: "Diseña un movimiento que encaje con tu rutina.",
      }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Nombre")).toBeInTheDocument();
    expect(screen.getByLabelText("Grupo muscular")).toBeInTheDocument();
    expect(screen.getByText("Anadir musculo secundario")).toBeInTheDocument();
    expect(screen.getByText("Marcar como ejercicio oficial")).toBeInTheDocument();
  });

  it("submits a new exercise to the backend", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse(
          { user: { id: "1", email: "admin@example.com", username: "admin", role: "admin" } },
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(jsonResponse(exerciseMetadataResponse(), { status: 200 }))
      .mockResolvedValueOnce(jsonResponse(exerciseListResponse([]), { status: 200 }))
      .mockResolvedValueOnce(
        jsonResponse(
          {
            id: "550e8400-e29b-41d4-a716-446655440999",
            name: "Bench Press",
            description: "Flat bench press",
            muscle_group: "chest",
            secondary_muscle_group: "triceps, shoulders",
            exercise_type: "strength",
            is_official: false,
            created_at: "2026-04-24T10:00:00Z",
          },
          { status: 201 },
        ),
      )
      .mockResolvedValueOnce(jsonResponse([], { status: 200 }))
      .mockResolvedValueOnce(
        jsonResponse(emptyExerciseInsightsResponse(), { status: 200 }),
      );
    vi.stubGlobal("fetch", fetchMock);

    renderExercisePage();

    await user.click(await screen.findByRole("button", { name: "Crear nuevo ejercicio" }));
    await user.type(screen.getByLabelText("Nombre"), "Bench Press");
    await user.selectOptions(screen.getByLabelText("Grupo muscular"), "chest");
    await user.type(screen.getByLabelText("Descripcion"), "Flat bench press");
    const secondaryMuscleInputs = screen.getAllByLabelText("Musculos secundarios");
    expect(secondaryMuscleInputs).toHaveLength(1);
    await user.selectOptions(secondaryMuscleInputs[0]!, "triceps");
    await user.click(screen.getByRole("button", { name: "Anadir musculo secundario" }));
    const expandedSecondaryMuscleInputs = screen.getAllByLabelText("Musculos secundarios");
    expect(expandedSecondaryMuscleInputs).toHaveLength(2);
    await user.selectOptions(expandedSecondaryMuscleInputs[1]!, "shoulders");
    await user.selectOptions(screen.getByLabelText("Tipo de ejercicio"), "strength");
    await user.click(screen.getByRole("button", { name: "Crear ejercicio" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/exercises"),
        expect.objectContaining({
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            name: "Bench Press",
            description: "Flat bench press",
            muscle_group: "chest",
            secondary_muscle_groups: ["triceps", "shoulders"],
            exercise_type: "strength",
            is_official: false,
          }),
        }),
      );
    });

    expect(await screen.findAllByText("Bench Press")).toHaveLength(2);
  });

  it("hides the official toggle for non admin users", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(
          jsonResponse(
            { user: { id: "2", email: "user@example.com", username: "user", role: "user" } },
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(jsonResponse(exerciseMetadataResponse(), { status: 200 }))
        .mockResolvedValueOnce(jsonResponse(exerciseListResponse([]), { status: 200 })),
    );
    const user = userEvent.setup();

    renderExercisePage();

    await user.click(await screen.findByRole("button", { name: "Crear nuevo ejercicio" }));

    expect(screen.queryByText("Marcar como ejercicio oficial")).not.toBeInTheDocument();
  });

  it("edits a custom exercise from the detail panel", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse(
          { user: { id: "2", email: "user@example.com", username: "user", role: "user" } },
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        jsonResponse(exerciseMetadataResponse(), { status: 200 }),
      )
      .mockResolvedValueOnce(
        jsonResponse(
          exerciseListResponse([
            {
              id: "550e8400-e29b-41d4-a716-446655440111",
              name: "Bench Press",
              description: "Flat bench press",
              muscle_group: "chest",
              secondary_muscle_group: "triceps, shoulders",
              exercise_type: "strength",
              is_official: false,
              created_at: "2026-04-24T10:00:00Z",
            },
          ]),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(jsonResponse([], { status: 200 }))
      .mockResolvedValueOnce(
        jsonResponse(emptyExerciseInsightsResponse(), { status: 200 }),
      )
      .mockResolvedValueOnce(
        jsonResponse(
          {
            id: "550e8400-e29b-41d4-a716-446655440111",
            name: "Bench Press Incline",
            description: "Incline bench press",
            muscle_group: "chest",
            secondary_muscle_group: "front delts, triceps",
            exercise_type: "strength",
            is_official: false,
            created_at: "2026-04-24T10:00:00Z",
          },
          { status: 200 },
        ),
      );
    vi.stubGlobal("fetch", fetchMock);

    renderExercisePage();

    await screen.findByRole("button", { name: "Editar" });
    await user.click(screen.getByRole("button", { name: "Editar" }));

    const nameInput = screen.getByLabelText("Nombre");
    await user.clear(nameInput);
    await user.type(nameInput, "Bench Press Incline");

    const descriptionInput = screen.getByLabelText("Descripcion");
    await user.clear(descriptionInput);
    await user.type(descriptionInput, "Incline bench press");

    const secondaryMuscleInputs = screen.getAllByLabelText("Musculos secundarios");
    await user.selectOptions(secondaryMuscleInputs[0]!, "front_delts");
    await user.selectOptions(secondaryMuscleInputs[1]!, "triceps");

    await user.click(screen.getByRole("button", { name: "Guardar cambios" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenLastCalledWith(
        expect.stringContaining("/api/exercises/550e8400-e29b-41d4-a716-446655440111"),
        expect.objectContaining({
          method: "PUT",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            name: "Bench Press Incline",
            description: "Incline bench press",
            muscle_group: "chest",
            secondary_muscle_groups: ["front_delts", "triceps"],
            exercise_type: "strength",
            is_official: false,
          }),
        }),
      );
    });

    expect(await screen.findAllByText("Bench Press Incline")).toHaveLength(2);
  });

  it("shows workout sessions where the selected exercise appears", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(
          jsonResponse(
            { user: { id: "2", email: "user@example.com", username: "user", role: "user" } },
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(jsonResponse(exerciseMetadataResponse(), { status: 200 }))
        .mockResolvedValueOnce(
          jsonResponse(
            exerciseListResponse([
              {
                id: "550e8400-e29b-41d4-a716-446655440111",
                name: "Bench Press",
                description: "Flat bench press",
                muscle_group: "chest",
                secondary_muscle_group: "triceps, shoulders",
                exercise_type: "strength",
                is_official: false,
                created_at: "2026-04-24T10:00:00Z",
              },
            ]),
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(
          jsonResponse(
            [
              {
                id: "660e8400-e29b-41d4-a716-446655440111",
                name: "Push Day",
                routine_name: "Push Pull Legs",
                started_at: "2026-04-25T10:00:00Z",
                duration_minutes: 45,
                exercise_order: 2,
                set_count: 3,
              },
            ],
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(
          jsonResponse(emptyExerciseInsightsResponse(), { status: 200 }),
        ),
    );

    renderExercisePage();

    expect(await screen.findByText("Push Day")).toBeInTheDocument();
    expect(screen.getByText("Push Pull Legs")).toBeInTheDocument();
    expect(screen.getByText("3 sets")).toBeInTheDocument();
  });

  it("shows advanced exercise insights for the selected exercise", async () => {
    const user = userEvent.setup();

    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(
          jsonResponse(
            { user: { id: "2", email: "user@example.com", username: "user", role: "user" } },
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(jsonResponse(exerciseMetadataResponse(), { status: 200 }))
        .mockResolvedValueOnce(
          jsonResponse(
            exerciseListResponse([
              {
                id: "550e8400-e29b-41d4-a716-446655440111",
                name: "Bench Press",
                description: "Flat bench press",
                muscle_group: "chest",
                secondary_muscle_group: "triceps, shoulders",
                exercise_type: "strength",
                is_official: false,
                created_at: "2026-04-24T10:00:00Z",
              },
            ]),
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(jsonResponse([], { status: 200 }))
        .mockResolvedValueOnce(
          jsonResponse(
            {
              summary: {
                session_count: 2,
                set_count: 6,
                total_volume_kg: 3120,
                max_weight_kg: 90,
                max_reps: 12,
                last_performed_at: "2026-04-25T10:00:00Z",
                first_performed_at: "2026-04-20T10:00:00Z",
                average_days_between: 5,
                trend: "up",
              },
              best_set: {
                session_id: "session-2",
                session_name: "Push Day",
                performed_at: "2026-04-25T10:00:00Z",
                set_number: 1,
                reps: 10,
                weight_kg: 90,
                volume_kg: 900,
                completed: true,
              },
              personal_records: [],
              progression: [
                {
                  session_id: "session-1",
                  date: "2026-04-20T10:00:00Z",
                  max_weight_kg: 80,
                  max_reps: 10,
                  volume_kg: 1200,
                  set_count: 3,
                },
                {
                  session_id: "session-2",
                  date: "2026-04-25T10:00:00Z",
                  max_weight_kg: 90,
                  max_reps: 12,
                  volume_kg: 1920,
                  set_count: 3,
                },
              ],
              history: [],
            },
            { status: 200 },
          ),
        ),
    );

    renderExercisePage();

    await user.click(
      await screen.findByRole("button", { name: "Ver rendimiento" }),
    );

    expect(await screen.findByText("Vista avanzada")).toBeInTheDocument();
    expect(await screen.findByText("Subiendo")).toBeInTheDocument();
    expect(screen.getByText("90 kg")).toBeInTheDocument();
    expect(
      screen.getByText((content) => {
        return content.includes("kg") && content.replace(/\D/g, "") === "3120";
      }),
    ).toBeInTheDocument();
  });

  it("shows edit action for admins on official exercises", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(
          jsonResponse(
            { user: { id: "1", email: "admin@example.com", username: "admin", role: "admin" } },
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(
          jsonResponse(
            exerciseMetadataResponse(),
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(
          jsonResponse(
            exerciseListResponse([
              {
                id: "550e8400-e29b-41d4-a716-446655440222",
                name: "Deadlift",
                description: "Conventional deadlift",
                muscle_group: "back",
                secondary_muscle_group: "glutes, hamstrings",
                exercise_type: "strength",
                is_official: true,
                created_at: "2026-04-24T10:00:00Z",
              },
            ]),
            { status: 200 },
          ),
        )
        .mockResolvedValueOnce(jsonResponse([], { status: 200 }))
        .mockResolvedValueOnce(
          jsonResponse(emptyExerciseInsightsResponse(), { status: 200 }),
        ),
    );

    renderExercisePage();

    expect(await screen.findByRole("button", { name: "Editar" })).toBeInTheDocument();
  });
});
