import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import AppLayout from "../components/AppLayout";
import DashboardPage from "./DashboardPage";

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: {
      "Content-Type": "application/json",
    },
    ...init,
  });
}

function renderDashboardPage(role = "user") {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      jsonResponse([
        {
          id: 1,
          name: "Sentadilla",
          muscle_group: "Pierna",
          exercise_type: "Fuerza",
          description: null,
          created_at: "2026-04-20T10:00:00Z",
        },
        {
          id: 2,
          name: "Press banca",
          muscle_group: "Pecho",
          exercise_type: "Fuerza",
          description: null,
          created_at: "2026-04-21T10:00:00Z",
        },
      ]),
    ),
  );

  return render(
    <MemoryRouter initialEntries={["/dashboard"]}>
      <Routes>
        <Route element={<AppLayout user={{ id: "1", username: "raul", email: "raul@example.com", role }} />}>
          <Route path="/dashboard" element={<DashboardPage />} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe("DashboardPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("shows the dashboard inside the shared sidebar layout", () => {
    renderDashboardPage();

    expect(screen.getByRole("heading", { name: "Hola, raul" })).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "Navegacion principal" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Dashboard" })).toHaveAttribute("href", "/dashboard");
    expect(screen.getByRole("link", { name: "Perfil" })).toHaveAttribute("href", "/profile");
    expect(screen.getByRole("link", { name: "Crear rutina" })).toHaveAttribute("href", "/routines/new");
    expect(screen.getByRole("link", { name: "Crear ejercicio" })).toHaveAttribute("href", "/exercises/new");
    expect(screen.getByRole("link", { name: "Mis rutinas" })).toHaveAttribute("href", "/routines");
    expect(screen.getByRole("link", { name: "Mis ejercicios" })).toHaveAttribute("href", "/exercises");
    expect(screen.queryByText("Usa el menu lateral para entrar en las secciones principales de la aplicacion.")).not.toBeInTheDocument();
  });

  it("shows the latest exercises and the routines panel", async () => {
    renderDashboardPage();

    expect(await screen.findByText("Press banca")).toBeInTheDocument();
    expect(screen.getByText("Sentadilla")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Tus rutinas" })).toBeInTheDocument();
    expect(screen.getByText("Todavia no hay rutinas para mostrar.")).toBeInTheDocument();
  });

  it("shows the admin shortcut only for admin users", () => {
    renderDashboardPage("admin");

    expect(screen.getByRole("link", { name: "Panel admin" })).toHaveAttribute("href", "/admin");
  });

  it("does not show the admin shortcut for regular users", () => {
    renderDashboardPage();

    expect(screen.queryByRole("link", { name: "Panel admin" })).not.toBeInTheDocument();
  });

  it("hides the sidebar completely and shows the edge handle", async () => {
    const user = userEvent.setup();
    renderDashboardPage();

    const sidebar = screen.getByRole("complementary", { hidden: true });

    expect(sidebar).toHaveClass("translate-x-0");

    await user.click(screen.getByRole("button", { name: "Ocultar menu" }));

    expect(sidebar).toHaveClass("-translate-x-full");
    expect(screen.getByRole("button", { name: "Mostrar menu" })).toHaveClass("translate-x-0");

    await user.click(screen.getByRole("button", { name: "Mostrar menu" }));

    expect(sidebar).toHaveClass("translate-x-0");
  });
});
