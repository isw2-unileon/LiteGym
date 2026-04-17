import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import LoginPage from "./LoginPage";

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

describe("LoginPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders the login form", () => {
    render(<LoginPage />);

    expect(screen.getByText("Grupo 16 Fitness")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Entra, entrena y controla tu progreso." })).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Contrasena")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Iniciar sesion" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Probar" })).toBeInTheDocument();
  });

  it("logs in and sends credentials with cookies enabled", async () => {
    const user = userEvent.setup();
    const fetchMock = mockFetch(jsonResponse({ user: { id: 1, email: "raul@example.com" } }, { status: 200 }));

    render(<LoginPage />);

    await user.click(screen.getByRole("button", { name: "Iniciar sesion" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/auth/login",
        expect.objectContaining({
          method: "POST",
          credentials: "include",
          body: JSON.stringify({
            email: "raul@example.com",
            password: "123456",
          }),
        }),
      );
    });

    expect(await screen.findByText("Sesion iniciada. La cookie HttpOnly ya esta guardada por el navegador.")).toBeInTheDocument();
  });

  it("shows the backend error when login fails", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ error: "invalid credentials" }, { status: 401 }));

    render(<LoginPage />);

    await user.click(screen.getByRole("button", { name: "Iniciar sesion" }));

    expect(await screen.findByText("invalid credentials")).toBeInTheDocument();
  });

  it("checks the protected exercises route with cookies enabled", async () => {
    const user = userEvent.setup();
    const fetchMock = mockFetch(jsonResponse([], { status: 200 }));

    render(<LoginPage />);

    await user.click(screen.getByRole("button", { name: "Probar" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/exercises",
        expect.objectContaining({
          credentials: "include",
        }),
      );
    });

    expect(await screen.findByText("Todo listo, ya tienes acceso.")).toBeInTheDocument();
  });

  it("shows a blocked message when the protected route returns unauthorized", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ error: "authentication required" }, { status: 401 }));

    render(<LoginPage />);

    await user.click(screen.getByRole("button", { name: "Probar" }));

    expect(await screen.findByText("Todavia no puedes entrar. Inicia sesion primero.")).toBeInTheDocument();
  });
});
