import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import AdminPage from "./AdminPage";

// --- Helpers ---

// A smart mock that returns different responses based on the requested URL and method
function setupFetchMock(isAdmin: boolean = true, authSuccess: boolean = true) {
  const fetchMock = vi.fn().mockImplementation((url: string, options?: RequestInit) => {
    const method = options?.method || "GET";

    if (url.includes("/api/users/me")) {
      if (!authSuccess) return Promise.resolve(new Response(null, { status: 401 }));
      return Promise.resolve(jsonResponse({ id: "1", username: "admin", role: isAdmin ? "admin" : "user" }, { status: 200 }));
    }

    if (url.endsWith("/api/users") && method === "GET") {
      return Promise.resolve(jsonResponse([
        { id: "1", username: "admin_user", email: "admin@test.com", role: "admin" },
        { id: "2", username: "normal_user", email: "user@test.com", role: "user" }
      ], { status: 200 }));
    }

    if (url.endsWith("/api/users") && method === "POST") {
      return Promise.resolve(jsonResponse({ message: "Created" }, { status: 201 }));
    }

    if (url.includes("/api/users/") && method === "DELETE") {
      return Promise.resolve(jsonResponse({ message: "Deleted" }, { status: 200 }));
    }

    // 5. Create Exercise
    if (url.endsWith("/api/exercises") && method === "POST") {
      return Promise.resolve(jsonResponse({ message: "Created" }, { status: 201 }));
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

function renderAdminPage() {
  return render(
    // We add a dummy /exercises route to catch the Navigate redirection if not admin
    <MemoryRouter initialEntries={["/admin"]}>
      <Routes>
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/exercises" element={<div>Exercises Page Mock</div>} />
      </Routes>
    </MemoryRouter>
  );
}

// --- Test Suite ---

describe("AdminPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders loading state initially", () => {
    // Delay the fetch promise to see the loading state
    vi.stubGlobal("fetch", () => new Promise(() => {}));
    renderAdminPage();
    expect(screen.getByText("Cargando panel...")).toBeInTheDocument();
  });

  it("redirects to /exercises if user is not an admin", async () => {
    setupFetchMock(false, true); // isAdmin = false
    renderAdminPage();

    // Wait for the redirection to happen
    expect(await screen.findByText("Exercises Page Mock")).toBeInTheDocument();
  });

  it("renders admin panel and fetches users if user is an admin", async () => {
    setupFetchMock(true, true);
    renderAdminPage();

    // Check header
    expect(await screen.findByRole("heading", { name: "Centro de Mando" })).toBeInTheDocument();
    
    // Check if the mock users were rendered
    expect(await screen.findByText("Usuarios Registrados (2)")).toBeInTheDocument();
    expect(screen.getByText("admin_user")).toBeInTheDocument();
    expect(screen.getByText("normal_user")).toBeInTheDocument();
  });

  it("submits the new user form successfully", async () => {
    const user = userEvent.setup();
    const fetchMock = setupFetchMock(true, true);
    renderAdminPage();

    // Wait for page to load
    expect(await screen.findByRole("heading", { name: "Centro de Mando" })).toBeInTheDocument();

    // Fill the form
    await user.type(screen.getByPlaceholderText("Nombre de usuario"), "newcomer");
    await user.type(screen.getByPlaceholderText("Correo electrónico"), "new@test.com");
    await user.type(screen.getByPlaceholderText("Contraseña"), "pass123");
    
    // Submit
    await user.click(screen.getByRole("button", { name: "Crear Usuario" }));

    // Verify fetch was called with correct payload
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/users"),
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            username: "newcomer",
            email: "new@test.com",
            password: "pass123",
            role: "user"
          })
        })
      );
    });

    // Check success message
    expect(await screen.findByText("Usuario creado correctamente.")).toBeInTheDocument();
  });

  it("deletes a user after confirmation", async () => {
    const user = userEvent.setup();
    const fetchMock = setupFetchMock(true, true);
    
    // Mock window.confirm to always return true
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);

    renderAdminPage();

    // Wait for users to load
    const deleteButtons = await screen.findAllByRole("button", { name: "Eliminar" });
    
    // Click the delete button of the first user
    await user.click(deleteButtons[0]);

    expect(confirmSpy).toHaveBeenCalledWith("¿Estás seguro de que deseas eliminar este usuario?");

    // Verify the DELETE fetch was called with the ID of the first mock user ("1")
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/users/1"),
        expect.objectContaining({ method: "DELETE" })
      );
    });

    // Check success message
    expect(await screen.findByText("Usuario eliminado.")).toBeInTheDocument();
  });

  it("switches to exercises tab and creates an exercise", async () => {
    const user = userEvent.setup();
    const fetchMock = setupFetchMock(true, true);
    renderAdminPage();

    // Wait for load and click the Exercises tab
    const exercisesTab = await screen.findByRole("button", { name: "Ejercicios Globales" });
    await user.click(exercisesTab);

    // Fill the exercise form
    await user.type(screen.getByPlaceholderText("Nombre del ejercicio (ej: Press Banca)"), "Sentadilla");
    await user.type(screen.getByPlaceholderText("Grupo muscular (ej: Pecho)"), "Piernas");
    await user.type(screen.getByPlaceholderText("Detalle (ej: Tracción horizontal)"), "Empuje vertical");

    // Submit
    await user.click(screen.getByRole("button", { name: "Publicar Ejercicio" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/exercises"),
        expect.objectContaining({
          method: "POST",
          body: expect.stringContaining('"is_official":true') 
        })
      );
    });

    expect(await screen.findByText("Ejercicio global creado correctamente.")).toBeInTheDocument();
  });
});