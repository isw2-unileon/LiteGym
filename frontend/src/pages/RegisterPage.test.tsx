import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import RegisterPage from "./RegisterPage";

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

function renderRegisterPage() {
  return render(
    <MemoryRouter>
      <RegisterPage />
    </MemoryRouter>,
  );
}

describe("RegisterPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders the register form", () => {
    renderRegisterPage();

    expect(screen.getByText("Grupo 16 Fitness")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Crea tu cuenta y empieza a entrenar con orden." })).toBeInTheDocument();
    expect(screen.getByLabelText("Nombre de usuario")).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Contrasena")).toBeInTheDocument();
    expect(screen.getByLabelText("Repite la contrasena")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Crear cuenta" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Inicia sesion" })).toHaveAttribute("href", "/");
  });

  it("registers, sends credentials with cookies enabled and notifies success", async () => {
    const user = userEvent.setup();
    const fetchMock = mockFetch(jsonResponse({ user: { id: 1, email: "new@example.com" } }, { status: 201 }));

    renderRegisterPage();

    await user.type(screen.getByLabelText("Nombre de usuario"), "nuevo");
    await user.type(screen.getByLabelText("Email"), "new@example.com");
    await user.type(screen.getByLabelText("Contrasena"), "password123");
    await user.type(screen.getByLabelText("Repite la contrasena"), "password123");
    await user.click(screen.getByRole("button", { name: "Crear cuenta" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/auth/register"),
        expect.objectContaining({
          method: "POST",
          credentials: "include",
          body: JSON.stringify({
            username: "nuevo",
            email: "new@example.com",
            password: "password123",
          }),
        }),
      );
    });

    expect(await screen.findByText("Cuenta creada. Entrando al panel...")).toBeInTheDocument();
  });

  it("shows a client-side error when passwords do not match", async () => {
    const user = userEvent.setup();
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    renderRegisterPage();

    await user.type(screen.getByLabelText("Nombre de usuario"), "nuevo");
    await user.type(screen.getByLabelText("Email"), "new@example.com");
    await user.type(screen.getByLabelText("Contrasena"), "password123");
    await user.type(screen.getByLabelText("Repite la contrasena"), "password456");
    await user.click(screen.getByRole("button", { name: "Crear cuenta" }));

    expect(await screen.findByText("Las contrasenas no coinciden.")).toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("shows the backend error when register fails", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ error: "username or email already registered" }, { status: 409 }));

    renderRegisterPage();

    await user.type(screen.getByLabelText("Nombre de usuario"), "nuevo");
    await user.type(screen.getByLabelText("Email"), "new@example.com");
    await user.type(screen.getByLabelText("Contrasena"), "password123");
    await user.type(screen.getByLabelText("Repite la contrasena"), "password123");
    await user.click(screen.getByRole("button", { name: "Crear cuenta" }));

    expect(await screen.findByText("username or email already registered")).toBeInTheDocument();
  });
});
