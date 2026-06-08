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

function renderLoginPage(initialEntries?: string[]) {
  return render(
    <MemoryRouter initialEntries={initialEntries ?? ["/"]}>
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

    expect(screen.getByText("LiteGym")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Entra, entrena y controla tu progreso." })).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("••••••••")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Iniciar sesion" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Registrate" })).toHaveAttribute("href", "/register");
    expect(screen.queryByRole("button", { name: "Probar" })).not.toBeInTheDocument();
  });

  it("shows a password reset success message when redirected from reset flow", () => {
    renderLoginPage(["/?reset=true"]);

    expect(
      screen.getByText("Tu contraseña ha sido actualizada con éxito. Ya puedes iniciar sesión."),
    ).toBeInTheDocument();
  });

  it("logs in, sends credentials with cookies enabled and notifies success", async () => {
    const user = userEvent.setup();
    const fetchMock = mockFetch(jsonResponse({ user: { id: 1, email: "raul@example.com" } }, { status: 200 }));

    renderLoginPage();

    await user.clear(screen.getByLabelText("Email"));
    await user.type(screen.getByLabelText("Email"), " Raul@Example.com ");
    await user.type(screen.getByPlaceholderText("••••••••"), "secret123");
    await user.click(screen.getByRole("button", { name: "Iniciar sesion" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/auth/login"),
        expect.objectContaining({
          method: "POST",
          credentials: "include",
          body: JSON.stringify({
            email: "raul@example.com",
            password: "secret123",
          }),
        }),
      );
    });

    expect(await screen.findByText("Sesion iniciada. Entrando al panel...")).toBeInTheDocument();
  });

  it("shows the backend error when login fails", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ error: "invalid credentials" }, { status: 401 }));

    renderLoginPage();

    await user.type(screen.getByLabelText("Email"), "raul@example.com");
    await user.type(screen.getByPlaceholderText("••••••••"), "secret123");
    await user.click(screen.getByRole("button", { name: "Iniciar sesion" }));

    expect(await screen.findByText("El correo o la contraseña no son correctos.")).toBeInTheDocument();
  });

  it("shows the translated invalid email message from the backend", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ error: "invalid email address" }, { status: 400 }));

    renderLoginPage();

    await user.clear(screen.getByLabelText("Email"));
    await user.type(screen.getByLabelText("Email"), "user@domain");
    await user.type(screen.getByPlaceholderText("••••••••"), "secret123");
    await user.click(screen.getByRole("button", { name: "Iniciar sesion" }));

    expect(await screen.findByText("Introduce un email valido.")).toBeInTheDocument();
  });

  it("shows the translated rate limit message from the backend", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ error: "login rate limit exceeded" }, { status: 429 }));

    renderLoginPage();

    await user.type(screen.getByLabelText("Email"), "raul@example.com");
    await user.type(screen.getByPlaceholderText("••••••••"), "secret123");
    await user.click(screen.getByRole("button", { name: "Iniciar sesion" }));

    expect(
      await screen.findByText("Has intentado iniciar sesion demasiadas veces. Espera unos segundos y vuelve a intentarlo."),
    ).toBeInTheDocument();
  });
});
