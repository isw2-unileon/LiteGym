import { cleanup, render, screen } from "@testing-library/react"; 
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import Profile from "./ProfilePage"; 

// --- Helpers ---
function setupFetchMock(mockData: unknown, isOk: boolean = true) {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(JSON.stringify(mockData), {
      status: isOk ? 200 : 401,
      headers: { "Content-Type": "application/json" },
    })
  );
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function renderProfilePage() {
  return render(
    <MemoryRouter>
      <Profile />
    </MemoryRouter>
  );
}

// --- Test Suite ---
describe("ProfilePage", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders the loading state initially", () => {
    // Stub fetch with a pending promise to trigger the loading state
    vi.stubGlobal("fetch", () => new Promise(() => {}));
    
    renderProfilePage();
    
    expect(screen.getByText("Loading profile...")).toBeInTheDocument();
  });

  it("renders an error message if the fetch fails", async () => {
    // Simulate an unauthorized access (401)
    setupFetchMock({ error: "unauthorized" }, false);
    
    renderProfilePage();

    expect(await screen.findByText("Error: Profile not found or unauthorized")).toBeInTheDocument();
  });

  it("renders user data correctly for a normal user (no admin button)", async () => {
    const mockUser = {
      id: "uuid-1234",
      username: "atleta_pro",
      email: "atleta@test.com",
      role: "user",
      created_at: "2026-04-24T10:00:00Z"
    };
    setupFetchMock(mockUser, true);
    
    renderProfilePage();

    expect(await screen.findByRole("heading", { name: "atleta_pro" })).toBeInTheDocument();
    expect(screen.getByText("atleta@test.com")).toBeInTheDocument();
    expect(screen.getByText("Rol actual: user")).toBeInTheDocument();
    
    // Check avatar initials
    expect(screen.getByText("a")).toBeInTheDocument();

    // Ensure Admin link is NOT rendered for normal users
    expect(screen.queryByRole("link", { name: "Panel de Administración" })).not.toBeInTheDocument();
  });

  it("renders the Admin Panel button if the user is an admin", async () => {
    const mockAdmin = {
      id: "uuid-9999",
      username: "super_admin",
      email: "admin@test.com",
      role: "admin",
      created_at: "2026-01-01T10:00:00Z"
    };
    setupFetchMock(mockAdmin, true);
    
    renderProfilePage();

    expect(await screen.findByRole("heading", { name: "super_admin" })).toBeInTheDocument();
    expect(screen.getByText("Rol actual: admin")).toBeInTheDocument();

    // Verify the Admin link is available and points to the correct route
    const adminLink = screen.getByRole("link", { name: "Panel de Administración" });
    expect(adminLink).toBeInTheDocument();
    expect(adminLink.getAttribute("href")).toBe("/admin");
  });
});