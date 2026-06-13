import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import AppLayout from "../components/AppLayout";
import AdminPage from "./AdminPage";

function setupFetchMock() {
  const fetchMock = vi.fn().mockImplementation((url: string, options?: RequestInit) => {
    const method = options?.method || "GET";

    // --- USER MOCKS ---
    if (url.endsWith("/api/users") && method === "GET") {
      return Promise.resolve(jsonResponse([
        { id: "1", username: "admin_user", email: "admin@test.com", role: "admin", is_active: true },
        { id: "2", username: "normal_user", email: "user@test.com", role: "user", is_active: true },
      ], { status: 200 }));
    }

    if (url.endsWith("/api/users") && method === "POST") {
      return Promise.resolve(jsonResponse({ message: "Created" }, { status: 201 }));
    }

    if (url.includes("/api/users/") && method === "DELETE") {
      return Promise.resolve(jsonResponse({ message: "Deleted" }, { status: 200 }));
    }

    // --- TICKET MOCKS ---
    if (url.endsWith("/api/tickets") && method === "GET") {
      return Promise.resolve(jsonResponse([
        { id: "t1", user_id: "u1", title: "[General] Problema app", description: "No carga", status: "open", created_at: "2026-04-29T10:00:00Z" },
      ], { status: 200 }));
    }

    if (url.includes("/api/tickets/") && url.endsWith("/close") && method === "PATCH") {
      return Promise.resolve(jsonResponse({ message: "Closed" }, { status: 200 }));
    }

    return Promise.resolve(jsonResponse({}, { status: 200 }));
  });

  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
}

function renderAdminPage(role = "admin") {
  return render(
    <MemoryRouter initialEntries={["/admin"]}>
      <Routes>
        <Route element={<AppLayout user={{ id: "1", username: "admin", email: "admin@test.com", role }} />}>
          <Route path="/admin" element={<AdminPage />} />
          <Route path="/exercises" element={<div>Exercises Page Mock</div>} />
          <Route path="/dashboard" element={<div>Dashboard Page Mock</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe("AdminPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("redirects to dashboard if user is not an admin", async () => {
    setupFetchMock();
    renderAdminPage("user");

    expect(await screen.findByText("Dashboard Page Mock")).toBeInTheDocument();
  });

  // --- TICKET TESTS ---
  it("renders admin panel, shows tickets tab by default, and allows closing a ticket", async () => {
    const user = userEvent.setup();
    const fetchMock = setupFetchMock();
    renderAdminPage();

    expect(await screen.findByRole("heading", { name: "CENTRO DE MANDO" })).toBeInTheDocument();
    
    // Verify that the tickets tab is active and shows the mocked ticket
    expect(await screen.findByText("[General] Problema app")).toBeInTheDocument();
    
    const closeButton = screen.getByRole("button", { name: "Resolver ticket" });
    await user.click(closeButton);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/tickets/t1/close"),
        expect.objectContaining({ method: "PATCH" })
      );
    });
  });

  // --- USER TESTS ---
  it("switches to users tab and fetches users", async () => {
    const user = userEvent.setup();
    setupFetchMock();
    renderAdminPage();

    const usersTab = await screen.findByRole("button", { name: "Gestión de Usuarios" });
    await user.click(usersTab);

    expect(await screen.findByText("Usuarios Registrados (2)")).toBeInTheDocument();
    expect(screen.getByText("admin_user")).toBeInTheDocument();
    expect(screen.getByText("normal_user")).toBeInTheDocument();
  });

  it("submits the new user form successfully", async () => {
    const user = userEvent.setup();
    const fetchMock = setupFetchMock();
    renderAdminPage();

    const usersTab = await screen.findByRole("button", { name: "Gestión de Usuarios" });
    await user.click(usersTab);

    await user.type(screen.getByPlaceholderText("Nombre de usuario"), "newcomer");
    await user.type(screen.getByPlaceholderText("Correo electrónico"), "new@test.com");
    await user.type(screen.getByPlaceholderText("••••••••"), "pass123");
    await user.click(screen.getByRole("button", { name: "Crear Usuario" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/users"),
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            username: "newcomer",
            email: "new@test.com",
            password: "pass123",
            role: "user",
          }),
        }),
      );
    });

    expect(await screen.findByText("Usuario creado correctamente.")).toBeInTheDocument();
  });

  it("deletes a user after confirmation", async () => {
    const user = userEvent.setup();
    const fetchMock = setupFetchMock();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);

    renderAdminPage();

    const usersTab = await screen.findByRole("button", { name: "Gestión de Usuarios" });
    await user.click(usersTab);

    const deleteButtons = await screen.findAllByRole("button", { name: "Eliminar" });
    await user.click(deleteButtons[0]!);

    expect(confirmSpy).toHaveBeenCalledWith("¿Estás seguro de que deseas eliminar este usuario?");

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/users/1"),
        expect.objectContaining({ method: "DELETE" }),
      );
    });

    expect(await screen.findByText("Usuario eliminado.")).toBeInTheDocument();
  });
});