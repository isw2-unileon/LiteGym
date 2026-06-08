import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { MemoryRouter, Outlet, Route, Routes } from "react-router-dom";
import UserRoutinesPage from "./UserRoutinesPage";

type RoutineSummaryLike = {
  id: string;
  name: string;
  description?: string;
  source?: string;
  is_predefined?: boolean;
  routine_type?: string;
  exercise_count?: number;
  updated_at?: string;
};

type FetchCall = { url: string; method: string; body?: string };

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
}

function buildDetail(
  id: string,
  routines: RoutineSummaryLike[],
  detailById?: Record<string, unknown>,
) {
  if (detailById && detailById[id]) {
    return detailById[id];
  }
  const routine = routines.find((item) => item.id === id);
  return {
    id,
    name: routine?.name ?? "Rutina",
    description: routine?.description ?? "",
    source: routine?.source ?? "manual",
    is_predefined: routine?.is_predefined ?? false,
    routine_type: routine?.routine_type ?? "Sin clasificar",
    exercise_count: routine?.exercise_count ?? 0,
    created_at: "2026-05-24T10:00:00Z",
    updated_at: routine?.updated_at ?? "2026-05-24T10:00:00Z",
    exercises: [],
  };
}

type InstallOptions = {
  routines?: RoutineSummaryLike[];
  exercises?: unknown[];
  detailById?: Record<string, unknown>;
  failList?: boolean;
};

function installFetch(options: InstallOptions = {}) {
  const routines = options.routines ?? [];
  const exercises = options.exercises ?? [];
  const calls: FetchCall[] = [];

  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === "string" ? input : String(input);
    const method = (init?.method ?? "GET").toUpperCase();
    const body = init?.body != null ? String(init.body) : undefined;
    calls.push({ url, method, body });

    const path = new URL(url, "http://localhost").pathname;

    if (path.startsWith("/api/exercises")) {
      return jsonResponse({
        items: exercises,
        page: 1,
        limit: 100,
        total: exercises.length,
        total_pages: 1,
      });
    }

    if (path === "/api/workouts/planned") {
      return jsonResponse({});
    }

    if (path === "/api/routines/ai/generate" || path === "/api/routines/ai/save") {
      return jsonResponse({ routine_json: {}, routine_id: "ai-routine-1" });
    }

    const duplicateMatch = path.match(/^\/api\/routines\/([^/]+)\/duplicate$/);
    if (duplicateMatch && method === "POST") {
      return jsonResponse({ routine_id: "duplicated-1" });
    }

    const byIdMatch = path.match(/^\/api\/routines\/([^/]+)$/);
    if (byIdMatch) {
      const id = byIdMatch[1] ?? "";
      if (method === "GET") {
        return jsonResponse(buildDetail(id, routines, options.detailById));
      }
      if (method === "PUT") {
        return jsonResponse({ routine_id: id });
      }
      if (method === "DELETE") {
        return jsonResponse({});
      }
    }

    if (path === "/api/routines") {
      if (method === "POST") {
        return jsonResponse({ routine_id: "created-1" });
      }
      if (options.failList) {
        return jsonResponse({ error: "failed" }, { status: 500 });
      }
      return jsonResponse(routines);
    }

    return jsonResponse({}, { status: 404 });
  });

  vi.stubGlobal("fetch", fetchMock);
  return { fetchMock, calls };
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/rutinas"]}>
      <Routes>
        <Route
          element={<Outlet context={{ user: { id: "u1", username: "diego" } }} />}
        >
          <Route path="/rutinas" element={<UserRoutinesPage />} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe("UserRoutinesPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("muestra las rutinas guardadas desde la API", async () => {
    installFetch({
      routines: [
        {
          id: "routine-1",
          name: "Upper Strength",
          description: "Trabajo principal de torso",
          exercise_count: 5,
          source: "manual",
        },
      ],
    });

    renderPage();

    expect(
      await screen.findByText("Trabajo principal de torso"),
    ).toBeInTheDocument();
    expect(screen.getAllByText("Upper Strength").length).toBeGreaterThan(0);
    expect(screen.getAllByText("5 ejercicios").length).toBeGreaterThan(0);
  });

  it("muestra el estado vacío cuando no hay rutinas", async () => {
    installFetch({ routines: [] });

    renderPage();

    expect(
      await screen.findByText("Todavía no tienes rutinas guardadas."),
    ).toBeInTheDocument();
  });

  it("muestra el estado de error cuando la API falla", async () => {
    installFetch({ failList: true });

    renderPage();

    expect(
      await screen.findByText(
        "No se ha podido cargar toda la información de las rutinas.",
      ),
    ).toBeInTheDocument();
  });

  it("carga el detalle al seleccionar una rutina de la lista", async () => {
    installFetch({
      routines: [
        { id: "routine-1", name: "Upper Strength", exercise_count: 0 },
        { id: "routine-2", name: "Lower Power", exercise_count: 1 },
      ],
      detailById: {
        "routine-2": {
          id: "routine-2",
          name: "Lower Power",
          description: "",
          source: "manual",
          is_predefined: false,
          routine_type: "Fuerza",
          exercise_count: 1,
          created_at: "2026-05-24T10:00:00Z",
          updated_at: "2026-05-24T10:00:00Z",
          exercises: [
            {
              id: "re-1",
              exercise_id: "ex-1",
              name: "Sentadilla",
              muscle_group: "legs",
              exercise_order: 1,
              sets: [],
            },
          ],
        },
      },
    });

    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByText("Lower Power"));

    expect(await screen.findByText("Sentadilla")).toBeInTheDocument();
  });

  it("filtra las rutinas por tipo", async () => {
    installFetch({
      routines: [
        {
          id: "routine-1",
          name: "Upper Strength",
          routine_type: "Fuerza",
          exercise_count: 0,
        },
        {
          id: "routine-2",
          name: "Lower Mobility",
          routine_type: "Movilidad",
          exercise_count: 1,
        },
      ],
      detailById: {
        "routine-2": {
          id: "routine-2",
          name: "Lower Mobility",
          description: "",
          source: "manual",
          is_predefined: false,
          routine_type: "Movilidad",
          exercise_count: 1,
          created_at: "2026-05-24T10:00:00Z",
          updated_at: "2026-05-24T10:00:00Z",
          exercises: [
            {
              id: "re-1",
              exercise_id: "ex-1",
              name: "Movilidad de cadera",
              muscle_group: "legs",
              exercise_order: 1,
              sets: [],
            },
          ],
        },
      },
    });

    const user = userEvent.setup();
    renderPage();

    await screen.findByText("Lower Mobility");
    await user.click(screen.getByRole("button", { name: /Movilidad/ }));

    expect(await screen.findByText("Movilidad de cadera")).toBeInTheDocument();
    expect(screen.queryByText("Upper Strength")).not.toBeInTheDocument();
  });

  it("filtra la lista por el término de búsqueda", async () => {
    installFetch({
      routines: [
        { id: "routine-1", name: "Upper Strength", exercise_count: 0 },
        { id: "routine-2", name: "Lower Power", exercise_count: 0 },
      ],
    });

    const user = userEvent.setup();
    renderPage();

    await screen.findAllByText("Lower Power");
    await user.type(
      screen.getByPlaceholderText("Buscar por nombre, grupo..."),
      "upper",
    );

    await waitFor(() => {
      expect(screen.queryByText("Lower Power")).not.toBeInTheDocument();
    });
    expect(screen.getAllByText("Upper Strength").length).toBeGreaterThan(0);
  });

  it("crea una rutina manual", async () => {
    const { calls } = installFetch({ routines: [] });

    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Crear rutina" }));
    await user.type(screen.getByLabelText("Nombre"), "Mi rutina");

    const createButtons = screen.getAllByRole("button", { name: "Crear rutina" });
    const submitButton = createButtons[createButtons.length - 1];
    if (!submitButton) {
      throw new Error("submit button not found");
    }
    await user.click(submitButton);

    await waitFor(() => {
      expect(
        calls.some(
          (call) => call.method === "POST" && call.url.endsWith("/api/routines"),
        ),
      ).toBe(true);
    });
  });

  it("edita una rutina existente", async () => {
    const { calls } = installFetch({
      routines: [{ id: "routine-1", name: "Upper Strength", exercise_count: 0 }],
    });

    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Editar" }));
    expect(await screen.findByDisplayValue("Upper Strength")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Guardar cambios" }));

    await waitFor(() => {
      expect(
        calls.some(
          (call) => call.method === "PUT" && call.url.endsWith("/api/routines/routine-1"),
        ),
      ).toBe(true);
    });
  });

  it("elimina una rutina tras confirmar", async () => {
    vi.stubGlobal("confirm", vi.fn(() => true));
    const { calls } = installFetch({
      routines: [{ id: "routine-1", name: "Upper Strength", exercise_count: 0 }],
    });

    const user = userEvent.setup();
    renderPage();

    await user.click(
      await screen.findByRole("button", { name: "Acciones de Upper Strength" }),
    );
    await user.click(await screen.findByRole("button", { name: "Eliminar" }));

    await waitFor(() => {
      expect(
        calls.some(
          (call) => call.method === "DELETE" && call.url.endsWith("/api/routines/routine-1"),
        ),
      ).toBe(true);
    });
  });

  it("duplica una rutina", async () => {
    const { calls } = installFetch({
      routines: [{ id: "routine-1", name: "Upper Strength", exercise_count: 0 }],
    });

    const user = userEvent.setup();
    renderPage();

    await user.click(
      await screen.findByRole("button", { name: "Acciones de Upper Strength" }),
    );
    await user.click(await screen.findByRole("button", { name: "Duplicar" }));

    await waitFor(() => {
      expect(
        calls.some(
          (call) =>
            call.method === "POST" &&
            call.url.endsWith("/api/routines/routine-1/duplicate"),
        ),
      ).toBe(true);
    });
  });

  it("deshabilita la creación al alcanzar el límite de rutinas", async () => {
    const routines = Array.from({ length: 50 }, (_, index) => ({
      id: `routine-${index}`,
      name: `Rutina ${index}`,
      exercise_count: 0,
    }));
    installFetch({ routines });

    renderPage();

    await screen.findAllByText("Rutina 0");
    expect(screen.getByRole("button", { name: "Crear rutina" })).toBeDisabled();
  });

  it("abre el modal de creación con IA", async () => {
    installFetch({ routines: [] });

    const user = userEvent.setup();
    renderPage();

    await user.click(
      await screen.findByRole("button", { name: "Generar rutina con IA" }),
    );

    const dialogTitle = await screen.findByText("Crear rutina con IA");
    expect(dialogTitle).toBeInTheDocument();
    expect(within(document.body).getByLabelText("Objetivo")).toBeInTheDocument();
  });
});
