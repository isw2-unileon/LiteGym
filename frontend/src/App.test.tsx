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

function exerciseMetadataResponse() {
  return {
    exercise_types: [{ value: "strength", label: "Fuerza" }],
    muscle_groups: [{ value: "chest", label: "Pecho" }],
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

describe("App", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("redirects to the dashboard after login", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({ user: { id: 1, email: "raul@example.com" } }, { status: 200 }))
      .mockResolvedValueOnce(jsonResponse([], { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(
      <MemoryRouter initialEntries={["/"]}>
        <App />
      </MemoryRouter>,
    );

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
    render(
      <MemoryRouter initialEntries={["/register"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Crea tu cuenta y empieza a entrenar con orden." })).toBeInTheDocument();
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

        if (url.includes("/api/exercises/metadata")) {
          return Promise.resolve(jsonResponse(exerciseMetadataResponse(), { status: 200 }));
        }

        if (url.includes("/api/exercises")) {
          return Promise.resolve(jsonResponse(exerciseListResponse([]), { status: 200 }));
        }

        return Promise.resolve(jsonResponse({}, { status: 200 }));
      }),
    );

    render(
      <MemoryRouter initialEntries={["/admin"]}>
        <App />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Crear nuevo ejercicio" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Panel admin" })).not.toBeInTheDocument();
  });
});
