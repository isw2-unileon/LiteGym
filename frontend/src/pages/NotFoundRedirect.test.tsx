import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import NotFoundRedirect from "./NotFoundRedirect";

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
}

function renderUnknownRoute() {
  return render(
    <MemoryRouter initialEntries={["/ruta-inexistente"]}>
      <Routes>
        <Route path="/" element={<div>Login Mock</div>} />
        <Route path="/dashboard" element={<div>Dashboard Mock</div>} />
        <Route path="*" element={<NotFoundRedirect />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("NotFoundRedirect", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("shows the redirecting message while the session is checked", () => {
    vi.stubGlobal("fetch", vi.fn(() => new Promise(() => {})));

    renderUnknownRoute();

    expect(screen.getByText("Redirigiendo...")).toBeInTheDocument();
  });

  it("redirects to the dashboard when the session is valid", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ user: { id: "1" } }, { status: 200 })));

    renderUnknownRoute();

    expect(await screen.findByText("Dashboard Mock")).toBeInTheDocument();
  });

  it("redirects to the login page when there is no valid session", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ error: "unauthorized" }, { status: 401 })));

    renderUnknownRoute();

    expect(await screen.findByText("Login Mock")).toBeInTheDocument();
  });

  it("redirects to the login page when the session check fails", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("network error")));

    renderUnknownRoute();

    expect(await screen.findByText("Login Mock")).toBeInTheDocument();
  });
});
