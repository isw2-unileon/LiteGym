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
        .mockResolvedValueOnce(jsonResponse([], { status: 200 })),
    );
    const user = userEvent.setup();

    renderExercisePage();

    await screen.findByRole("button", { name: "Crear nuevo ejercicio" });
    await user.click(screen.getByRole("button", { name: "Crear nuevo ejercicio" }));

    expect(
      screen.getByRole("heading", {
        name: "Disena un movimiento que encaje con tu rutina.",
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
      .mockResolvedValueOnce(jsonResponse([], { status: 200 }))
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
      );
    vi.stubGlobal("fetch", fetchMock);

    renderExercisePage();

    await user.click(await screen.findByRole("button", { name: "Crear nuevo ejercicio" }));
    await user.type(screen.getByLabelText("Nombre"), "Bench Press");
    await user.type(screen.getByLabelText("Grupo muscular"), "chest");
    await user.type(screen.getByLabelText("Descripcion"), "Flat bench press");
    const secondaryMuscleInputs = screen.getAllByPlaceholderText("Ej. triceps");
    expect(secondaryMuscleInputs).toHaveLength(1);
    await user.type(secondaryMuscleInputs[0]!, "triceps");
    await user.click(screen.getByRole("button", { name: "Anadir musculo secundario" }));
    const expandedSecondaryMuscleInputs = screen.getAllByPlaceholderText("Ej. triceps");
    expect(expandedSecondaryMuscleInputs).toHaveLength(2);
    await user.type(expandedSecondaryMuscleInputs[1]!, "shoulders");
    await user.type(screen.getByLabelText("Tipo de ejercicio"), "strength");
    await user.click(screen.getByRole("button", { name: "Crear ejercicio" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenLastCalledWith(
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
        .mockResolvedValueOnce(jsonResponse([], { status: 200 })),
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
        jsonResponse(
          [
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
          ],
          { status: 200 },
        ),
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

    const secondaryMuscleInputs = screen.getAllByPlaceholderText("Ej. triceps");
    await user.clear(secondaryMuscleInputs[0]!);
    await user.type(secondaryMuscleInputs[0]!, "front delts");
    await user.clear(secondaryMuscleInputs[1]!);
    await user.type(secondaryMuscleInputs[1]!, "triceps");

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
            secondary_muscle_groups: ["front delts", "triceps"],
            exercise_type: "strength",
            is_official: false,
          }),
        }),
      );
    });

    expect(await screen.findAllByText("Bench Press Incline")).toHaveLength(2);
  });
});
