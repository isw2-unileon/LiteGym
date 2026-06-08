import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import TermsPage from "./TermsPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <TermsPage />
    </MemoryRouter>,
  );
}

describe("TermsPage", () => {
  afterEach(cleanup);

  it("renders the terms and conditions content", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "Términos y condiciones" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "1. Uso aceptable" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "3. Uso de IA" })).toBeInTheDocument();
    expect(screen.getByText("Mayo de 2026")).toBeInTheDocument();
  });

  it("links back to the home page", () => {
    renderPage();

    expect(screen.getByRole("link", { name: "Volver al inicio" })).toHaveAttribute("href", "/");
  });
});
