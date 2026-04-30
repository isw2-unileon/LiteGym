import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import AppLayout from "../components/AppLayout";
import DashboardPage from "./DashboardPage";

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: {
      "Content-Type": "application/json",
    },
    ...init,
  });
}

function buildDashboardResponse() {
  return {
    calendar: {
      month: "2026-04",
      trained_days: ["2026-04-21", "2026-04-25", "2026-04-28"],
      sessions_count: 3,
      current_streak: 1,
      next_objective: "Llegar a 8 sesiones este mes",
    },
    recent_routines: [
      {
        id: "routine-1",
        name: "Push Pull Legs",
        description: "Rutina dividida en 3 dias",
        exercise_count: 3,
        updated_at: "2026-04-28T10:00:00Z",
      },
    ],
    recent_workouts: [
      {
        id: "workout-1",
        name: "Push Day 1",
        routine_name: "Push Pull Legs",
        started_at: "2026-04-28T10:00:00Z",
        duration_minutes: 70,
        exercise_count: 3,
      },
    ],
    progress: {
      last_recorded_at: "2026-04-26T10:00:00Z",
      weight_kg: {
        current: 78.5,
        previous: 79.2,
        delta: -0.7,
      },
      body_fat_percentage: {
        current: 17.9,
        previous: 18.5,
        delta: -0.6,
      },
      muscle_mass_kg: {
        current: 36.4,
        previous: 36.0,
        delta: 0.4,
      },
    },
    muscle_distribution: {
      year: [
        { name: "Espalda", count: 5, percentage: 50 },
        { name: "Pecho", count: 3, percentage: 30 },
        { name: "Pierna", count: 2, percentage: 20 },
      ],
      month: [
        { name: "Espalda", count: 2, percentage: 67 },
        { name: "Pecho", count: 1, percentage: 33 },
      ],
      year_exercise_count: 10,
      month_exercise_count: 3,
    },
  };
}

function renderDashboardPage(role = "user", fetchImplementation?: ReturnType<typeof vi.fn>) {
  const defaultFetch =
    fetchImplementation ??
    vi.fn().mockResolvedValue(
      jsonResponse(buildDashboardResponse()),
    );

  vi.stubGlobal("fetch", defaultFetch);

  return render(
    <MemoryRouter initialEntries={["/dashboard"]}>
      <Routes>
        <Route element={<AppLayout user={{ id: "1", username: "raul", email: "raul@example.com", role }} />}>
          <Route path="/dashboard" element={<DashboardPage />} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe("DashboardPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("shows the dashboard inside the shared sidebar layout", async () => {
    const user = userEvent.setup();
    renderDashboardPage();

    expect(screen.getByRole("heading", { name: "Hola, raul" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Mostrar menu" }));

    const navigation = screen.getByRole("navigation", { name: "Navegacion principal" });
    expect(navigation).toBeInTheDocument();
    expect(within(navigation).getByRole("link", { name: "Dashboard" })).toHaveAttribute("href", "/dashboard");
    expect(within(navigation).getByRole("link", { name: "Perfil" })).toHaveAttribute("href", "/profile");
    expect(within(navigation).getByRole("link", { name: "Crear rutina" })).toHaveAttribute("href", "/routines/new");
    expect(within(navigation).getByRole("link", { name: "Crear ejercicio" })).toHaveAttribute("href", "/exercises/new");
    expect(within(navigation).getByRole("link", { name: "Mis rutinas" })).toHaveAttribute("href", "/routines");
    expect(within(navigation).getByRole("link", { name: "Mis ejercicios" })).toHaveAttribute("href", "/exercises");
  });

  it("shows the routines, progress, and training sections with backend data", async () => {
    renderDashboardPage();

    expect(await screen.findByRole("heading", { name: "Tus rutinas" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Resumen de progreso" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Ultimos entrenos" })).toBeInTheDocument();
    expect(screen.getAllByText("Push Pull Legs").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("Push Day 1")).toBeInTheDocument();
    expect(screen.getByText("78.5 kg")).toBeInTheDocument();
  });

  it("shows the calendar and hexagon statistics for the backend response", async () => {
    renderDashboardPage();

    expect(await screen.findByRole("img", { name: "Grafico hexagonal de distribucion muscular" })).toBeInTheDocument();
    expect(screen.getByText("Grupos musculares presentes")).toBeInTheDocument();
    expect(screen.getByText("Sesiones del mes")).toBeInTheDocument();
    expect(screen.getByText("Racha actual")).toBeInTheDocument();
    expect(screen.getAllByText("Espalda").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText(/10 ejercicios considerados/i)).toBeInTheDocument();
  });

  it("shows fallback statistics when no workout history is available", async () => {
    renderDashboardPage(
      "user",
      vi.fn().mockResolvedValue(
        jsonResponse({
          ...buildDashboardResponse(),
          muscle_distribution: {
            year: [],
            month: [],
            year_exercise_count: 0,
            month_exercise_count: 0,
          },
        }),
      ),
    );

    expect(await screen.findByText("Grupos musculares presentes")).toBeInTheDocument();
    expect(screen.getByText("Cuando haya historial de entrenos, aqui veras que grupos musculares predominan.")).toBeInTheDocument();
  });

  it("lets the user switch the statistics card between year and month", async () => {
    const user = userEvent.setup();
    renderDashboardPage();

    const yearButton = await screen.findByRole("button", { name: "Año" });
    const monthButton = screen.getByRole("button", { name: "Mes" });

    expect(yearButton).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByText(/Ultimo año/i)).toBeInTheDocument();

    await user.click(monthButton);

    expect(monthButton).toHaveAttribute("aria-pressed", "true");
    expect(yearButton).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByText(/Ultimo mes/i)).toBeInTheDocument();
    expect(screen.getByText(/3 ejercicios considerados/i)).toBeInTheDocument();
  });

  it("shows the admin shortcut only for admin users", async () => {
    const user = userEvent.setup();
    renderDashboardPage("admin");

    await user.click(screen.getByRole("button", { name: "Mostrar menu" }));

    expect(screen.getByRole("link", { name: "Panel admin" })).toHaveAttribute("href", "/admin");
  });

  it("does not show the admin shortcut for regular users", () => {
    renderDashboardPage();

    expect(screen.queryByRole("link", { name: "Panel admin" })).not.toBeInTheDocument();
  });

  it("starts with the sidebar hidden and lets the user open and close it", async () => {
    const user = userEvent.setup();
    renderDashboardPage();

    const sidebar = screen.getAllByRole("complementary", { hidden: true })[0];

    expect(sidebar).toHaveClass("-translate-x-full");
    expect(screen.getByRole("button", { name: "Mostrar menu" })).toHaveClass("translate-x-0");

    await user.click(screen.getByRole("button", { name: "Mostrar menu" }));

    expect(sidebar).toHaveClass("translate-x-0");

    await user.click(screen.getByRole("button", { name: "Ocultar menu" }));

    expect(sidebar).toHaveClass("-translate-x-full");
  });
});
