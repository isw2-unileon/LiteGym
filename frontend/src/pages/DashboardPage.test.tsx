import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import DashboardPage from "./DashboardPage";

function mockFetch(response: Response) {
  const fetchMock = vi.fn().mockResolvedValue(response);
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: {
      "Content-Type": "application/json",
    },
    ...init,
  });
}

function renderDashboardPage() {
  return render(
    <MemoryRouter>
      <DashboardPage />
    </MemoryRouter>,
  );
}

describe("DashboardPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("checks api access with cookies when it loads", async () => {
    const fetchMock = mockFetch(jsonResponse([], { status: 200 }));

    renderDashboardPage();

    expect(screen.getByRole("heading", { name: "Dashboard" })).toBeInTheDocument();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/auth/me"),
        expect.objectContaining({
          credentials: "include",
        }),
      );
    });

    expect(await screen.findByText("Todo listo, ya tienes acceso.")).toBeInTheDocument();
  });

  it("redirects when the backend says the user is not authenticated", async () => {
    mockFetch(jsonResponse({ error: "authentication required" }, { status: 401 }));

    renderDashboardPage();

    expect(await screen.findByText("Todavia no puedes entrar. Inicia sesion primero.")).toBeInTheDocument();
  });

  it("lets the user re-check api access from the dashboard", async () => {
    const user = userEvent.setup();
    const fetchMock = mockFetch(jsonResponse([], { status: 200 }));

    renderDashboardPage();

    await screen.findByText("Todo listo, ya tienes acceso.");
    await user.click(screen.getByRole("button", { name: "Probar" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    expect(fetchMock).toHaveBeenLastCalledWith(
      expect.stringContaining("/api/exercises"),
      expect.objectContaining({
        credentials: "include",
      }),
    );
  });

  it("logs out and returns to login", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({ user: { id: 1, email: "raul@example.com" } }, { status: 200 }))
      .mockResolvedValueOnce(jsonResponse({ message: "session closed" }, { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    renderDashboardPage();

    await screen.findByText("Todo listo, ya tienes acceso.");
    await user.click(screen.getByRole("button", { name: "Cerrar sesion" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenLastCalledWith(
        expect.stringContaining("/api/auth/logout"),
        expect.objectContaining({
          method: "POST",
          credentials: "include",
        }),
      );
    });
  });
});
