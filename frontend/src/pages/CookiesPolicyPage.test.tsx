import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";
import CookiesPolicyPage from "./CookiesPolicyPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <CookiesPolicyPage />
    </MemoryRouter>,
  );
}

describe("CookiesPolicyPage", () => {
  afterEach(cleanup);

  it("renders the cookies policy content", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "Política de cookies" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "1. Cookies utilizadas" })).toBeInTheDocument();
    expect(screen.getByText("Mayo de 2026")).toBeInTheDocument();
  });

  it("links back to the home page", () => {
    renderPage();

    expect(screen.getByRole("link", { name: "Volver al inicio" })).toHaveAttribute("href", "/");
  });
});
