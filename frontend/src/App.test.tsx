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

    expect(await screen.findByRole("heading", { name: "Hola" })).toBeInTheDocument();

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

  it("redirects unknown routes to dashboard when the session is valid", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ user: { id: "user-id" } }, { status: 200 })));

    render(
      <MemoryRouter initialEntries={["/no-existe"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("Redirigiendo...")).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "Hola" })).toBeInTheDocument();
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
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ user: { id: "user-id", role: "user" } }, { status: 200 })));

    render(
      <MemoryRouter initialEntries={["/admin"]}>
        <App />
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "Hola" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Panel admin" })).not.toBeInTheDocument();
  });
});
