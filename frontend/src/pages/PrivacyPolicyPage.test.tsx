import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import PrivacyPolicyPage from "./PrivacyPolicyPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <PrivacyPolicyPage />
    </MemoryRouter>,
  );
}

describe("PrivacyPolicyPage", () => {
  afterEach(cleanup);

  it("renders the privacy policy content", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "Política de privacidad" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "1. Responsable del tratamiento" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "5. Derechos del usuario" })).toBeInTheDocument();
    expect(screen.getByText("Mayo de 2026")).toBeInTheDocument();
  });

  it("links back to the home page", () => {
    renderPage();

    expect(screen.getByRole("link", { name: "Volver al inicio" })).toHaveAttribute("href", "/");
  });
});
