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
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse([], { status: 200 })));
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
  });

  it("submits a new exercise to the backend", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse([], { status: 200 }))
      .mockResolvedValueOnce(
        jsonResponse(
          {
            id: "550e8400-e29b-41d4-a716-446655440999",
            name: "Bench Press",
            description: "Flat bench press",
            muscle_group: "chest",
            secondary_muscle_group: "triceps",
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
    await user.type(screen.getByLabelText("Musculo secundario"), "triceps");
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
            secondary_muscle_group: "triceps",
            exercise_type: "strength",
            is_official: false,
          }),
        }),
      );
    });

    expect(await screen.findByText("Bench Press")).toBeInTheDocument();
  });
});
