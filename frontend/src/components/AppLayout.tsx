import { useState } from "react";
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import { apiUrl } from "../lib/api";
import { clearAuthToken } from "../lib/authSession";

export type LayoutUser = {
  id: string;
  username?: string;
  email?: string;
  role?: string;
};

type AppLayoutProps = {
  user?: LayoutUser | null;
};

const navigationItems = [
  { label: "Dashboard", to: "/dashboard" },
  { label: "Perfil", to: "/profile" },
  { label: "Mis rutinas", to: "/routines" },
  { label: "Mis ejercicios", to: "/exercises" },
];

const pageBackground =
  "absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,_rgba(234,113,48,0.24),_transparent_30%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_52%,_#d8e1d0_100%)]";

export default function AppLayout({ user }: AppLayoutProps) {
  const navigate = useNavigate();
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const isAdmin = user?.role === "admin";

  const handleLogout = async () => {
    try {
      await fetch(apiUrl("/api/auth/logout"), {
        method: "POST",
        credentials: "include",
      });
    } finally {
      clearAuthToken();
      navigate("/", { replace: true });
    }
  };

  return (
    <main className="min-h-screen bg-[#f4efe2] text-[#1f1b16]">
      <div className="min-h-screen">
        <aside
          aria-hidden={!isSidebarOpen}
          className={`fixed inset-y-0 left-0 z-30 w-72 border-r border-[#1f1b16]/10 bg-[#1f1b16] px-5 py-6 text-[#fffaf0] shadow-[20px_0_60px_rgba(31,27,22,0.18)] transition-transform duration-300 ${
            isSidebarOpen ? "translate-x-0" : "-translate-x-full"
          }`}
        >
          <div className="flex items-center justify-between gap-4">
            <Link to="/dashboard">
              <p className="text-xs font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">Grupo 16</p>
              <h1 className="mt-3 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.05em]">
                Fitness
              </h1>
            </Link>

            <button
              aria-expanded={isSidebarOpen}
              aria-label="Ocultar menu"
              className="grid h-10 w-10 place-items-center rounded-r-none rounded-l-2xl border border-white/15 text-lg font-black text-[#fffaf0] transition hover:border-[#f1a45b] hover:text-[#f1a45b]"
              type="button"
              onClick={() => setIsSidebarOpen(false)}
            >
              &lt;
            </button>
          </div>

          <nav className="mt-8 space-y-2" aria-label="Navegacion principal">
            {navigationItems.map((item) => (
              <NavLink
                className={({ isActive }) =>
                  `block rounded-2xl px-4 py-3 text-sm font-black transition ${
                    isActive ? "bg-white/15 text-white" : "text-[#fffaf0]/80 hover:bg-white/10 hover:text-white"
                  }`
                }
                key={item.to}
                to={item.to}
              >
                {item.label}
              </NavLink>
            ))}

            {isAdmin && (
              <NavLink
                className={({ isActive }) =>
                  `block rounded-2xl px-4 py-3 text-sm font-black transition ${
                    isActive ? "bg-[#ffbc76] text-[#1f1b16]" : "bg-[#f1a45b] text-[#1f1b16] hover:bg-[#ffbc76]"
                  }`
                }
                to="/admin"
              >
                Panel admin
              </NavLink>
            )}
          </nav>

          <div className="mt-8 border-t border-white/10 pt-5">
            <Link
              to="/support"
              className="mb-6 block text-sm font-bold text-[#fffaf0]/80 transition hover:text-[#f1a45b]"
            >
              Soporte Técnico
            </Link>

            <p className="mb-4 text-xs font-semibold text-[#fffaf0]/65">
              {user?.username ?? user?.email ?? "Usuario"}
            </p>
            <button
              className="w-full rounded-2xl border border-white/15 px-4 py-3 text-sm font-bold text-[#fffaf0] transition hover:border-[#f1a45b] hover:text-[#f1a45b]"
              type="button"
              onClick={handleLogout}
            >
              Cerrar sesion
            </button>
          </div>
        </aside>

        <section className="relative isolate min-h-screen px-6 py-8 sm:px-10 lg:px-12 xl:px-16">
          <div className={pageBackground} />

          <button
            aria-expanded={isSidebarOpen}
            aria-label="Mostrar menu"
            className={`fixed left-0 top-8 z-40 grid h-14 w-9 place-items-center rounded-r-2xl bg-[#1f1b16] text-lg font-black text-[#fffaf0] shadow-[10px_10px_35px_rgba(31,27,22,0.18)] transition hover:bg-[#265c52] ${
              isSidebarOpen ? "pointer-events-none -translate-x-full opacity-0" : "translate-x-0 opacity-100"
            }`}
            type="button"
            onClick={() => setIsSidebarOpen(true)}
          >
            &gt;
          </button>

          <div
            className={`mx-auto pt-3 transition-[max-width,padding] duration-300 ${
              isSidebarOpen
                ? "max-w-6xl lg:max-w-[min(1100px,calc(100vw-9rem))] xl:max-w-[min(1180px,calc(100vw-10rem))]"
                : "max-w-7xl lg:max-w-[min(1320px,calc(100vw-4rem))] xl:max-w-[min(1440px,calc(100vw-5rem))]"
            }`}
          >
            <Outlet context={{ user }} />
          </div>
        </section>
      </div>
    </main>
  );
}
