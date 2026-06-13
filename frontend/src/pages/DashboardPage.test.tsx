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
      planned_days: [],
      calendar_workouts: [
        {
          id: "workout-1",
          name: "Push Day 1",
          routine_name: "Push Pull Legs",
          performed_at: "2026-04-28T10:00:00Z",
          planned_at: null,
          duration_minutes: 70,
          exercise_count: 3,
        },
      ],
      sessions_count: 3,
      current_streak: 1,
      weekly_goal: 2,
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
        performed_at: "2026-04-28T10:00:00Z",
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
    vi.useRealTimers();
  });

  it("shows the dashboard inside the shared header layout", () => {
    renderDashboardPage();

    expect(screen.getByRole("heading", { name: "Hola, raul" })).toBeInTheDocument();

    const navigation = screen.getByRole("navigation");
    expect(navigation).toBeInTheDocument();
    expect(within(navigation).getByRole("link", { name: "Panel" })).toHaveAttribute("href", "/dashboard");
    expect(within(navigation).getByRole("link", { name: "Perfil" })).toHaveAttribute("href", "/profile");
    expect(within(navigation).getByRole("link", { name: "Mis rutinas" })).toHaveAttribute("href", "/routines");
    expect(within(navigation).getByRole("link", { name: "Mis ejercicios" })).toHaveAttribute("href", "/exercises");
  });

  it("shows the routines, progress, and training sections with backend data", async () => {
    renderDashboardPage();

    expect(await screen.findByRole("heading", { name: "Tus rutinas" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Resumen de progreso" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Últimos entrenos" })).toBeInTheDocument();
    expect(screen.getAllByText("Push Pull Legs").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("Push Day 1")).toBeInTheDocument();
    expect(screen.getByText("78.5 kg")).toBeInTheDocument();
  });

  it("shows the calendar and hexagon statistics for the backend response", async () => {
    renderDashboardPage();

    expect(await screen.findByRole("img", { name: "Grafico octogonal de distribucion muscular" })).toBeInTheDocument();
    expect(screen.getByText("Grupos musculares presentes")).toBeInTheDocument();
    expect(screen.getByText("Sesiones del mes")).toBeInTheDocument();
    expect(screen.getByText("Racha semanal")).toBeInTheDocument();
    expect(screen.getByText("Objetivo semanal")).toBeInTheDocument();
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
    expect(screen.getByText("Cuando haya historial de entrenos, aquí veras que grupos musculares predominan.")).toBeInTheDocument();
  });

  it("lets the user switch the statistics card between year and month", async () => {
    const user = userEvent.setup();
    renderDashboardPage();

    const yearButton = await screen.findByRole("button", { name: "Año" });
    const monthButton = screen.getByRole("button", { name: "Mes" });

    expect(yearButton).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByText(/Último año/i)).toBeInTheDocument();

    await user.click(monthButton);

    expect(monthButton).toHaveAttribute("aria-pressed", "true");
    expect(yearButton).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByText(/Último mes/i)).toBeInTheDocument();
    expect(screen.getByText(/3 ejercicios considerados/i)).toBeInTheDocument();
  });

  it("shows the admin shortcut only for admin users", () => {
    renderDashboardPage("admin");

    expect(screen.getByRole("link", { name: "Panel administrativo" })).toHaveAttribute("href", "/admin");
  });

  it("does not show the admin shortcut for regular users", () => {
    renderDashboardPage();

    expect(screen.queryByRole("link", { name: "Panel administrativo" })).not.toBeInTheDocument();
  });

  it("starts with the profile menu closed and lets the user open and close it", async () => {
    const user = userEvent.setup();
    renderDashboardPage();

    const menuButton = screen.getByRole("button", { name: "Abrir menu de perfil" });

    expect(menuButton).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("button", { name: "Cerrar sesión" })).not.toBeInTheDocument();

    await user.click(menuButton);

    expect(menuButton).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByRole("button", { name: "Cerrar sesión" })).toBeInTheDocument();

    await user.click(screen.getByRole("heading", { name: "Hola, raul" }));

    expect(menuButton).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("button", { name: "Cerrar sesión" })).not.toBeInTheDocument();
  });

  it("shows a google calendar link for planned workouts in the calendar dialog", async () => {
    const user = userEvent.setup();
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    const year = tomorrow.getFullYear();
    const month = String(tomorrow.getMonth() + 1).padStart(2, "0");
    const day = String(tomorrow.getDate()).padStart(2, "0");
    const dateKey = `${year}-${month}-${day}`;
    const plannedAt = `${dateKey}T18:00:00Z`;

    renderDashboardPage(
      "user",
      vi.fn().mockResolvedValue(
        jsonResponse({
          ...buildDashboardResponse(),
          calendar: {
            ...buildDashboardResponse().calendar,
            month: `${year}-${month}`,
            trained_days: [],
            planned_days: [dateKey],
            calendar_workouts: [
              {
                id: "planned-1",
                name: "Push Day 2",
                routine_name: "Push Pull Legs",
                performed_at: null,
                planned_at: plannedAt,
                duration_minutes: 70,
                exercise_count: 4,
              },
            ],
          },
        }),
      ),
    );

    await screen.findByText("PANEL PRINCIPAL");
    await user.click(screen.getByRole("button", { name: tomorrow.getDate().toString() }));

    const calendarLink = await screen.findByRole("link", { name: "Añadir a Calendar" });
    expect(calendarLink).toHaveAttribute("href", expect.stringContaining("calendar.google.com/calendar/render"));
    expect(calendarLink).toHaveAttribute("href", expect.stringContaining("action=TEMPLATE"));
  });
});
