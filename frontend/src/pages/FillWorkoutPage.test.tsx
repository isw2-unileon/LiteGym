import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import FillWorkoutPage from "./FillWorkoutPage";

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
}

describe("FillWorkoutPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("loads the workout detail and shows editable sets", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse({
          id: "workout-1",
          name: "Push Day",
          planned_at: "2026-06-07T10:00:00Z",
          notes: "",
          exercises: [
            {
              id: "exercise-1",
              name: "Bench Press",
              muscle_group: "chest",
              exercise_order: 1,
              notes: "Pausa abajo",
              sets: [
                {
                  id: "set-1",
                  set_number: 1,
                  target_reps_min: 8,
                  target_reps_max: 10,
                  target_weight_kg: 60,
                  completed: null,
                },
              ],
            },
          ],
        }),
      ),
    );

    render(
      <MemoryRouter initialEntries={["/workouts/workout-1/fill"]}>
        <Routes>
          <Route path="/workouts/:id/fill" element={<FillWorkoutPage />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByDisplayValue("Push Day")).toBeInTheDocument();
    expect(screen.getByText("Bench Press")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pendiente" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Reps reales")).toBeInTheDocument();
  });

  it("lets the user mark a set as completed and enter values", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse({
          id: "workout-1",
          name: "Push Day",
          exercises: [
            {
              id: "exercise-1",
              name: "Bench Press",
              muscle_group: "chest",
              exercise_order: 1,
              sets: [
                {
                  id: "set-1",
                  set_number: 1,
                  target_reps_text: "8-10",
                  completed: null,
                },
              ],
            },
          ],
        }),
      ),
    );

    const user = userEvent.setup();
    render(
      <MemoryRouter initialEntries={["/workouts/workout-1/fill"]}>
        <Routes>
          <Route path="/workouts/:id/fill" element={<FillWorkoutPage />} />
        </Routes>
      </MemoryRouter>,
    );

    await user.click(await screen.findByRole("button", { name: "Hecho" }));
    const repsInput = screen.getByPlaceholderText("Reps reales");
    await user.type(repsInput, "9");

    expect(repsInput).toHaveValue(9);
  });
});
