import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "./App";

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: {
      "Content-Type": "application/json",
    },
    ...init,
  });
}

function dashboardResponse() {
  return {
    calendar: {
      month: "2026-04",
      trained_days: [],
      planned_days: [],
      calendar_workouts: [],
      sessions_count: 0,
      current_streak: 0,
      weekly_goal: 2,
      next_objective: "",
    },
    recent_routines: [],
    recent_workouts: [],
    progress: {
      last_recorded_at: null,
      weight_kg: { current: null, previous: null, delta: null },
      body_fat_percentage: { current: null, previous: null, delta: null },
      muscle_mass_kg: { current: null, previous: null, delta: null },
    },
    muscle_distribution: {
      year: [],
      month: [],
      year_exercise_count: 0,
      month_exercise_count: 0,
    },
  };
}

describe("App", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("redirects to the dashboard after login", async () => {
    const user = userEvent.setup();
    let isLoggedIn = false;
    const fetchMock = vi.fn().mockImplementation((url: string) => {
      if (url.includes("/api/auth/login")) {
        isLoggedIn = true;
        return Promise.resolve(jsonResponse({ user: { id: 1, email: "raul@example.com" } }, { status: 200 }));
      }

      if (url.includes("/api/auth/me")) {
        return Promise.resolve(
          isLoggedIn
            ? jsonResponse({ user: { id: 1, email: "raul@example.com" } }, { status: 200 })
            : jsonResponse({ error: "authentication required" }, { status: 401 }),
        );
      }

      if (url.includes("/api/dashboard")) {
        return Promise.resolve(jsonResponse(dashboardResponse(), { status: 200 }));
      }

      return Promise.resolve(jsonResponse([], { status: 200 }));
    });
    vi.stubGlobal("fetch", fetchMock);

    render(
      <MemoryRouter initialEntries={["/"]}>
        <App />
      </MemoryRouter>,
    );

    await user.type(await screen.findByLabelText("Email"), "raul@example.com");
    await user.type(screen.getByLabelText("Contrasena"), "secret123");
    await user.click(screen.getByRole("button", { name: "Iniciar sesion" }));

    expect(await screen.findByRole("heading", { name: /Hola/i })).toBeInTheDocument();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/auth/me"),
        expect.objectContaining({
          credentials: "include",
        }),
      );
    });
  });

  it("does not keep the dashboard visible when the session is invalid", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ error: "authentication required" }, { status: 401 })));

    render(
      <MemoryRouter initialEntries={["/dashboard"]}>
        <App />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Iniciar sesion" })).toBeInTheDocument();
    });
  });

  it("renders the register page route", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ error: "authentication required" }, { status: 401 })));

    render(
      <MemoryRouter initialEntries={["/register"]}>
        <App />
      </MemoryRouter>,
    );

    expect(
      await screen.findByRole("heading", { name: "Crea tu cuenta y empieza a entrenar con orden." }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Crear cuenta" })).toBeInTheDocument();
  });

  it("redirects unknown routes to dashboard when the session is valid", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ user: { id: "user-id" } }, { status: 200 })));

    render(
      <MemoryRouter initialEntries={["/no-existe"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("Redirigiendo...")).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: /Hola/i })).toBeInTheDocument();
  });

  it("redirects unknown routes to login when the session is missing", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ error: "authentication required" }, { status: 401 })));

    render(
      <MemoryRouter initialEntries={["/no-existe"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("Redirigiendo...")).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: "Iniciar sesion" })).toBeInTheDocument();
  });

  it("does not allow regular users to open the admin route directly", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((url: string) => {
        if (url.includes("/api/auth/me")) {
          return Promise.resolve(jsonResponse({ user: { id: "user-id", role: "user" } }, { status: 200 }));
        }

        if (url.includes("/api/dashboard")) {
          return Promise.resolve(jsonResponse(dashboardResponse(), { status: 200 }));
        }

        return Promise.resolve(jsonResponse({}, { status: 200 }));
      }),
    );

    render(
      <MemoryRouter initialEntries={["/admin"]}>
        <App />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: /Hola/i })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "CENTRO DE MANDO" })).not.toBeInTheDocument();
  });
});
