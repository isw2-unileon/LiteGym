import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import LegalNoticePage from "./LegalNoticePage";

function renderPage() {
  return render(
    <MemoryRouter>
      <LegalNoticePage />
    </MemoryRouter>,
  );
}

describe("LegalNoticePage", () => {
  afterEach(cleanup);

  it("renders the legal notice content", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "Aviso legal" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "1. Titular del sitio" })).toBeInTheDocument();
    expect(screen.getByText("Mayo de 2026")).toBeInTheDocument();
  });

  it("exposes the legal contact email as a mailto link", () => {
    renderPage();

    expect(screen.getByRole("link", { name: "legal@tu-dominio.com" })).toHaveAttribute(
      "href",
      "mailto:legal@tu-dominio.com",
    );
  });

  it("links back to the home page", () => {
    renderPage();

    expect(screen.getByRole("link", { name: "Volver al inicio" })).toHaveAttribute("href", "/");
  });
});
