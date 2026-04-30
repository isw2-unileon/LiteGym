import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import SupportPage from "./SupportPage"; 

describe("SupportPage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renderiza el formulario de soporte correctamente", () => {
    render(
      <MemoryRouter>
        <SupportPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "Soporte Técnico" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Volver" })).toBeInTheDocument();

    expect(screen.getByRole("combobox")).toBeInTheDocument(); 
    expect(screen.getByPlaceholderText("Ej: No me carga una rutina")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Explica detalladamente tu problema...")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Enviar Ticket" })).toBeInTheDocument();
  });

  it("envía el ticket correctamente, muestra mensaje de éxito y resetea el formulario", async () => {
    const user = userEvent.setup();
    
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ id: "123", message: "Created" }), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    render(
      <MemoryRouter>
        <SupportPage />
      </MemoryRouter>
    );

    const categorySelect = screen.getByRole("combobox");
    const titleInput = screen.getByPlaceholderText("Ej: No me carga una rutina");
    const descriptionInput = screen.getByPlaceholderText("Explica detalladamente tu problema...");
    const submitButton = screen.getByRole("button", { name: "Enviar Ticket" });

    
    await user.selectOptions(categorySelect, "IA");
    await user.type(titleInput, "La IA me habla en binario");
    await user.type(descriptionInput, "No entiendo nada de lo que dice el asistente.");

    await user.click(submitButton);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/tickets"),
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            category: "IA",
            title: "La IA me habla en binario",
            description: "No entiendo nada de lo que dice el asistente.",
          }),
        })
      );
    });

    expect(await screen.findByText("Ticket enviado correctamente.")).toBeInTheDocument();

    expect(titleInput).toHaveValue("");
    expect(descriptionInput).toHaveValue("");
    expect(categorySelect).toHaveValue("General"); 
  });

  it("muestra un mensaje de error si el servidor falla al enviar el ticket", async () => {
    const user = userEvent.setup();
    
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: "Server error" }), {
        status: 500,
        headers: { "Content-Type": "application/json" },
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    render(
      <MemoryRouter>
        <SupportPage />
      </MemoryRouter>
    );

    await user.type(screen.getByPlaceholderText("Ej: No me carga una rutina"), "Error fatal");
    await user.type(screen.getByPlaceholderText("Explica detalladamente tu problema..."), "La app explota al abrir");
    await user.click(screen.getByRole("button", { name: "Enviar Ticket" }));

    expect(await screen.findByText("Error al enviar el ticket.")).toBeInTheDocument();
  });
});