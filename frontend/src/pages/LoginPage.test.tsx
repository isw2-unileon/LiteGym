import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
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

function renderLoginPage() {
  return render(
    <MemoryRouter>
      <LoginPage />
    </MemoryRouter>,
  );
}

describe("LoginPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders the login form", () => {
    renderLoginPage();

    expect(screen.getByText("Grupo 16 Fitness")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Entra, entrena y controla tu progreso." })).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Contrasena")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Iniciar sesion" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Probar" })).not.toBeInTheDocument();
  });

  it("logs in, sends credentials with cookies enabled and notifies success", async () => {
    const user = userEvent.setup();
    const fetchMock = mockFetch(
      jsonResponse({ user: { id: 1, email: "raul@example.com" }, token: "test-token" }, { status: 200 }),
    );

    renderLoginPage();

    await user.click(screen.getByRole("button", { name: "Iniciar sesion" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/auth/login"),
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

    expect(await screen.findByText("Sesion iniciada. Entrando al panel...")).toBeInTheDocument();
    expect(window.localStorage.getItem("litegym_auth_token")).toBe("test-token");
  });

  it("shows the backend error when login fails", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ error: "invalid credentials" }, { status: 401 }));

    renderLoginPage();

    await user.click(screen.getByRole("button", { name: "Iniciar sesion" }));

    expect(await screen.findByText("invalid credentials")).toBeInTheDocument();
  });
});
