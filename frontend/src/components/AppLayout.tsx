import { useEffect, useRef, useState } from "react";
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import { apiUrl } from "../lib/api";
import { clearAuthToken } from "../lib/authSession";
import { legalLinks } from "../lib/legalLinksData";

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
  { label: "Panel", to: "/dashboard" },
  { label: "Mis rutinas", to: "/routines" },
  { label: "Mis ejercicios", to: "/exercises" },
  { label: "Soporte Técnico", to: "/support" },
  { label: "Perfil", to: "/profile" },
];

const pageBackground =
  "min-h-screen text-[#1f1b16] bg-[radial-gradient(circle_at_top_right,_rgba(234,113,48,0.24),_transparent_30%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_52%,_#d8e1d0_100%)]";

const capitalize = (value: string) => value.charAt(0).toUpperCase() + value.slice(1);

export default function AppLayout({ user }: AppLayoutProps) {
  const isAdmin = user?.role === "admin";
  const displayName = capitalize(user?.username ?? user?.email ?? "Usuario");
  const navigate = useNavigate();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const [isLegalOpen, setIsLegalOpen] = useState(false);
  const legalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isMenuOpen) {
      return;
    }

    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsMenuOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isMenuOpen]);

  useEffect(() => {
    if (!isLegalOpen) {
      return;
    }

    const handleClickOutside = (event: MouseEvent) => {
      if (legalRef.current && !legalRef.current.contains(event.target as Node)) {
        setIsLegalOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isLegalOpen]);

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
    <main className={pageBackground}>
      <header className="sticky top-0 z-30 mb-7 border-b border-[#1f1b16]/10 bg-[#fffaf0]/80 py-2.5 shadow-[0_10px_30px_rgba(31,27,22,0.10)] backdrop-blur-md">
        <div className="grid grid-cols-3 items-center px-[10rem]">
          <Link
            to="/dashboard"
            className="flex items-center gap-3 justify-self-start [font-family:'Bricolage_Grotesque','Aptos_Display','Trebuchet_MS',sans-serif] text-[22px] font-black tracking-[-0.04em]"
          >
            <div className="grid h-[34px] w-[34px] place-items-center rounded-[10px] bg-[#1f1b16] text-[18px] font-black tracking-[-0.06em] text-[#f1a45b] shadow-[0_8px_18px_rgba(31,27,22,0.18)]">
              L
            </div>
            LiteGym
          </Link>

          <nav className="flex gap-1 justify-self-center rounded-[14px] border border-[#1f1b16]/12 p-1">
            {navigationItems.map((item) => (
              <NavLink
                className={({ isActive }) =>
                  `flex items-center justify-center rounded-[14px] px-3 py-2 text-center text-sm font-bold transition ${
                    isActive
                      ? "bg-[#f1a45b] text-[#1f1b16]"
                      : "text-[#1f1b16]/80 hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
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
                  `flex items-center justify-center rounded-[14px] px-3 py-2 text-center text-sm font-bold transition ${
                    isActive
                      ? "bg-[#f1a45b] text-[#1f1b16]"
                      : "text-[#1f1b16]/80 hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
                  }`
                }
                to="/admin"
              >
                Panel administrativo
              </NavLink>
            )}

            <div ref={legalRef} className="relative flex">
              <button
                type="button"
                aria-haspopup="menu"
                aria-expanded={isLegalOpen}
                onClick={() => setIsLegalOpen((open) => !open)}
                className="flex items-center justify-center rounded-[14px] px-3 py-2 text-center text-sm font-bold text-[#1f1b16]/80 transition hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
              >
                Legal
              </button>

              {isLegalOpen && (
                <div className="absolute left-0 top-[calc(100%+0.625rem)] z-40 w-48 overflow-hidden rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0] shadow-[0_10px_30px_rgba(31,27,22,0.10)] backdrop-blur-md">
                  {legalLinks.map((link) => (
                    <NavLink
                      key={link.to}
                      to={link.to}
                      onClick={() => setIsLegalOpen(false)}
                      className="block px-4 py-3 text-sm font-bold text-[#1f1b16]/80 transition hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
                    >
                      {link.label}
                    </NavLink>
                  ))}
                </div>
              )}
            </div>
          </nav>

          <div ref={menuRef} className="relative flex items-center gap-2.5 justify-self-end self-stretch">
            <span className="text-sm font-bold">{displayName}</span>
            <button
              type="button"
              aria-haspopup="menu"
              aria-expanded={isMenuOpen}
              aria-label="Abrir menu de perfil"
              onClick={() => setIsMenuOpen((open) => !open)}
              className="grid h-[34px] w-[34px] place-items-center rounded-[12px] border border-[#1f1b16]/15 bg-gradient-to-br from-[#ea7130] to-[#ff8b47] [font-family:'Bricolage_Grotesque','Aptos_Display','Trebuchet_MS',sans-serif] text-[18px] font-black tracking-[-0.02em] text-[#1f1b16] shadow-[0_8px_18px_rgba(234,113,48,0.30)] transition hover:-translate-y-px"
            >
              {displayName.charAt(0)}
            </button>

            {isMenuOpen && (
              <div className="absolute right-0 top-[calc(100%+0.625rem)] z-40 w-48 overflow-hidden rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0]/80 shadow-[0_10px_30px_rgba(31,27,22,0.10)] backdrop-blur-md">
                <NavLink
                  to="/profile"
                  onClick={() => setIsMenuOpen(false)}
                  className="block px-4 py-3 text-sm font-bold text-[#1f1b16]/80 transition hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
                >
                  Perfil
                </NavLink>
                <button
                  type="button"
                  onClick={handleLogout}
                  className="block w-full border-t border-[#1f1b16]/10 px-4 py-3 text-left text-sm font-bold text-[#9f2f22] transition hover:bg-[#9f2f22]/10"
                >
                  Cerrar sesión
                </button>
              </div>
            )}
          </div>
        </div>
      </header>

      <div className="mx-auto max-w-7xl pt-3 lg:max-w-[min(1320px,calc(100vw-4rem))] xl:max-w-[min(1440px,calc(100vw-5rem))]">
        <Outlet context={{ user }} />
      </div>
    </main>
  );
}
