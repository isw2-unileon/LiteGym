import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import ProfilePage from "./ProfilePage";

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: {
      "Content-Type": "application/json",
    },
    ...init,
  });
}

type ProfileFetchSetup = {
  me?: unknown;
  stats?: unknown;
  meStatus?: number;
  statsStatus?: number;
  aiStatus?: number;
  aiResponse?: unknown;
};

function setupProfileFetchMock({
  me,
  stats,
  meStatus = 200,
  statsStatus = 200,
  aiStatus = 200,
  aiResponse,
}: ProfileFetchSetup) {
  const fetchMock = vi.fn().mockImplementation((input: RequestInfo | URL) => {
    const url = typeof input === "string" ? input : input.toString();

    if (url.includes("/api/auth/me")) {
      return Promise.resolve(jsonResponse(me ?? { user: null }, { status: meStatus }));
    }

    if (url.includes("/api/profile/dashboard")) {
      return Promise.resolve(jsonResponse(stats ?? {}, { status: statsStatus }));
    }

    if (url.includes("/api/profile/ai-analysis")) {
      return Promise.resolve(jsonResponse(aiResponse ?? { analysis: "Analysis OK" }, { status: aiStatus }));
    }

    return Promise.resolve(jsonResponse({}, { status: 200 }));
  });

  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function renderProfilePage() {
  return render(
    <MemoryRouter>
      <ProfilePage />
    </MemoryRouter>,
  );
}

describe("ProfilePage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders the loading state initially", () => {
    vi.stubGlobal("fetch", () => new Promise(() => { }));

    renderProfilePage();

    expect(screen.getByText("Cargando perfil...")).toBeInTheDocument();
  });

  it("renders an error message if the fetch fails", async () => {
    setupProfileFetchMock({ me: { error: "unauthorized" }, meStatus: 401 });

    renderProfilePage();

    expect(await screen.findByText("No se pudo cargar el perfil del usuario.")).toBeInTheDocument();
  });

  it("renders user data correctly for a normal user", async () => {
    const mockUser = {
      user: {
        id: "uuid-1234",
        username: "atleta_pro",
        email: "atleta@test.com",
        role: "user",
        created_at: "2026-04-24T10:00:00Z",
      },
    };
    const mockStats = {
      total_workouts: 0,
      total_duration_minutes: 0,
      total_volume_kg: 0,
      total_sets: 0,
      streak_days: [],
      streak_activities: [],
      top_exercises: [],
      muscle_radar: [],
      weight_history: [],
      goals: null,
    };
    setupProfileFetchMock({ me: mockUser, stats: mockStats });

    renderProfilePage();

    expect(await screen.findByRole("heading", { name: /Hola,\s*atleta_pro/i })).toBeInTheDocument();
    expect(screen.getByText("atleta@test.com")).toBeInTheDocument();
    expect(screen.getByText("user")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Mes anterior" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Mes siguiente" })).toBeInTheDocument();
  });

  it("renders user data correctly for an admin", async () => {
    const mockAdmin = {
      user: {
        id: "uuid-9999",
        username: "super_admin",
        email: "admin@test.com",
        role: "admin",
        created_at: "2026-01-01T10:00:00Z",
      },
    };
    const mockStats = {
      total_workouts: 0,
      total_duration_minutes: 0,
      total_volume_kg: 0,
      total_sets: 0,
      streak_days: [],
      streak_activities: [],
      top_exercises: [],
      muscle_radar: [],
      weight_history: [],
      goals: null,
    };
    setupProfileFetchMock({ me: mockAdmin, stats: mockStats });

    renderProfilePage();

    expect(await screen.findByRole("heading", { name: /Hola,\s*super_admin/i })).toBeInTheDocument();
    expect(screen.getByText("admin")).toBeInTheDocument();
  });

  it("shows rate limit message when AI analysis returns 429", async () => {
    const mockUser = { user: { id: "uuid-1234", username: "user1", email: "u@u.com", role: "user" } };
    setupProfileFetchMock({ me: mockUser, stats: {}, aiStatus: 429 });

    renderProfilePage();

    const analyzeBtn = await screen.findByRole("button", { name: /Generar Análisis/i });
    analyzeBtn.click();

    expect(await screen.findByText("¡Uf! He estado analizando muchos perfiles hoy y necesito un breve descanso para recuperar energía. Vuelve a intentarlo en un ratito, ¡y seguiremos dándolo todo!")).toBeInTheDocument();
  });

  it("shows generic error message when AI analysis connection fails", async () => {
    const mockUser = { user: { id: "uuid-1234", username: "user1", email: "u@u.com", role: "user" } };
    setupProfileFetchMock({ me: mockUser, stats: {}, aiStatus: 500 });

    renderProfilePage();

    const analyzeBtn = await screen.findByRole("button", { name: /Generar Análisis/i });
    analyzeBtn.click();

    expect(await screen.findByText("Parece que hubo un pequeño problema de conexión con tu entrenador. Inténtalo de nuevo más tarde.")).toBeInTheDocument();
  });
});
