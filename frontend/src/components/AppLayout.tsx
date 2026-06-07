import {useEffect, useRef, useState} from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { apiUrl } from "../lib/api";
import { legalLinks } from "./legalLinksData";

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
  {label: "Panel", to: "/dashboard"},
  {label: "Mis rutinas", to: "/routines"},
  {label: "Mis ejercicios", to: "/exercises"},
  {label: "Soporte Técnico", to: "/support"},
  {label: "Perfil", to: "/profile"},
]

const mobileNavigationItems = [
  {label: "Panel", to: "/dashboard", shortLabel: "Panel", icon: "dashboard"},
  {label: "Mis rutinas", to: "/routines", shortLabel: "Rutinas", icon: "routines"},
  {label: "Mis ejercicios", to: "/exercises", shortLabel: "Ejercicios", icon: "exercises"},
  {label: "Perfil", to: "/profile", shortLabel: "Perfil", icon: "profile"},
];

const pageBackground =
  "min-h-screen text-[#1f1b16] bg-[radial-gradient(circle_at_top_right,_rgba(234,113,48,0.24),_transparent_30%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_52%,_#d8e1d0_100%)]";

const capitalize = (value: string) =>
  value.charAt(0).toUpperCase() + value.slice(1);

type MobileNavIconName = (typeof mobileNavigationItems)[number]["icon"] | "more";

function MobileNavIcon({ name }: { name: MobileNavIconName }) {
  if (name === "dashboard") {
    return (
      <svg aria-hidden="true" className="h-[19px] w-[19px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 13a8 8 0 0 1 16 0" />
        <path d="M12 13l4-4" />
        <path d="M5 19h14" />
      </svg>
    );
  }

  if (name === "routines") {
    return (
      <svg aria-hidden="true" className="h-[19px] w-[19px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M7 4h10" />
        <path d="M6 8h12" />
        <path d="M5 12h14" />
        <path d="M7 16h10" />
        <path d="M9 20h6" />
      </svg>
    );
  }

  if (name === "exercises") {
    return (
      <svg aria-hidden="true" className="h-[19px] w-[19px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M6 8v8" />
        <path d="M18 8v8" />
        <path d="M3 10v4" />
        <path d="M21 10v4" />
        <path d="M6 12h12" />
      </svg>
    );
  }

  if (name === "profile") {
    return (
      <svg aria-hidden="true" className="h-[19px] w-[19px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z" />
        <path d="M4 21a8 8 0 0 1 16 0" />
      </svg>
    );
  }

  return (
    <svg aria-hidden="true" className="h-[19px] w-[19px]" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" strokeLinejoin="round">
      <path d="M5 12h14" />
      <path d="M12 5v14" />
    </svg>
  );
}

export default function AppLayout({ user }: AppLayoutProps) {
  const isAdmin = user?.role === "admin";
  const displayName = capitalize(user?.username ?? user?.email ?? "Usuario");
  const navigate = useNavigate();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const [isLegalOpen, setIsLegalOpen] = useState(false);
  const legalRef = useRef<HTMLDivElement>(null);
  const [isMobileMoreOpen, setIsMobileMoreOpen] = useState(false);
  const mobileMoreRef = useRef<HTMLDivElement>(null);

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

  useEffect(() => {
    if (!isMobileMoreOpen) {
      return;
    }

    const handleClickOutside = (event: MouseEvent) => {
      if (mobileMoreRef.current && !mobileMoreRef.current.contains(event.target as Node)) {
        setIsMobileMoreOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isMobileMoreOpen]);

  const handleLogout = async () => {
    try {
      await fetch(apiUrl("/api/auth/logout"), {
        method: "POST",
        credentials: "include",
      });
    } finally {
      navigate("/", { replace: true });
    }
  };

  return (
      <main className={pageBackground}>
        <header className="sticky top-0 z-30 mb-4 border-b border-[#1f1b16]/10 bg-[#fffaf0]/80 py-2.5 backdrop-blur-md shadow-[0_10px_30px_rgba(31,27,22,0.10)] md:mb-7">
          <div className="grid grid-cols-[1fr_auto] items-center px-4 sm:px-6 md:grid-cols-3 md:px-8 lg:px-[10rem]">
            <div className="flex items-center gap-3 justify-self-start [font-family:'Bricolage_Grotesque','Aptos_Display','Trebuchet_MS',sans-serif] text-[22px] font-black tracking-[-0.04em]">
              <div className="grid h-[34px] w-[34px] place-items-center rounded-[10px] bg-[#1f1b16] text-[18px] font-black tracking-[-0.06em] text-[#f1a45b] shadow-[0_8px_18px_rgba(31,27,22,0.18)]">
                L
              </div>
              LiteGym
            </div>
            <nav className="hidden gap-1 justify-self-center rounded-[14px] border border-[#1f1b16]/12 p-1 md:flex">
              {navigationItems.map((item) => (
                  <NavLink
                      className={({ isActive }) =>
                          `flex items-center justify-center text-center rounded-[14px] px-3 py-2 text-sm font-bold transition ${
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
                          `flex items-center justify-center text-center rounded-[14px] px-3 py-2 text-sm font-bold transition ${
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
                    className="flex items-center justify-center text-center rounded-[14px] px-3 py-2 text-sm font-bold text-[#1f1b16]/80 transition hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
                >
                  Legal
                </button>

                {isLegalOpen && (
                    <div className="absolute left-0 top-[calc(100%+0.625rem)] z-40 w-48 overflow-hidden rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0] backdrop-blur-md shadow-[0_10px_30px_rgba(31,27,22,0.10)]">
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
              <span className="hidden text-sm font-bold sm:inline">{displayName}</span>
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
                  <div className="absolute right-0 top-[calc(100%+0.625rem)] z-40 w-48 overflow-hidden rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0]/80 backdrop-blur-md shadow-[0_10px_30px_rgba(31,27,22,0.10)]">
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

        <div className="mx-auto max-w-7xl pb-[calc(6rem+env(safe-area-inset-bottom))] pt-1 md:pb-0 md:pt-3 lg:max-w-[min(1320px,calc(100vw-4rem))] xl:max-w-[min(1440px,calc(100vw-5rem))]">
          <Outlet context={{ user }} />
        </div>

        <div
          role="toolbar"
          aria-label="Navegación móvil"
          className="fixed inset-x-3 bottom-3 z-40 rounded-[22px] border border-[#1f1b16]/12 bg-[#fffaf0]/92 px-2 py-2 shadow-[0_18px_45px_rgba(31,27,22,0.20)] backdrop-blur-md md:hidden"
        >
          <div className="grid grid-cols-5 items-center gap-1">
            {mobileNavigationItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `flex min-h-[3.45rem] flex-col items-center justify-center rounded-[16px] px-1.5 text-[11px] font-black transition ${
                    isActive
                      ? "bg-[#f1a45b] text-[#1f1b16]"
                      : "text-[#1f1b16]/72 hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
                  }`
                }
              >
                <MobileNavIcon name={item.icon} />
                <span className="mt-1.5 leading-none">{item.shortLabel}</span>
              </NavLink>
            ))}

            <div ref={mobileMoreRef} className="relative">
              <button
                type="button"
                aria-haspopup="menu"
                aria-expanded={isMobileMoreOpen}
                onClick={() => setIsMobileMoreOpen((open) => !open)}
                className="flex min-h-[3.45rem] w-full flex-col items-center justify-center rounded-[16px] px-1.5 text-[11px] font-black text-[#1f1b16]/72 transition hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
              >
                <MobileNavIcon name="more" />
                <span className="mt-1.5 leading-none">Más</span>
              </button>

              {isMobileMoreOpen && (
                <div className="absolute bottom-[calc(100%+0.75rem)] right-0 w-56 overflow-hidden rounded-[18px] border border-[#1f1b16]/10 bg-[#fffaf0] shadow-[0_16px_40px_rgba(31,27,22,0.18)]">
                  <NavLink
                    to="/support"
                    onClick={() => setIsMobileMoreOpen(false)}
                    className="block px-4 py-3 text-sm font-bold text-[#1f1b16]/80 transition hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
                  >
                    Soporte Técnico
                  </NavLink>
                  {isAdmin && (
                    <NavLink
                      to="/admin"
                      onClick={() => setIsMobileMoreOpen(false)}
                      className="block border-t border-[#1f1b16]/10 px-4 py-3 text-sm font-bold text-[#1f1b16]/80 transition hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
                    >
                      Panel administrativo
                    </NavLink>
                  )}
                  {legalLinks.map((link) => (
                    <NavLink
                      key={link.to}
                      to={link.to}
                      onClick={() => setIsMobileMoreOpen(false)}
                      className="block border-t border-[#1f1b16]/10 px-4 py-3 text-sm font-bold text-[#1f1b16]/80 transition hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]"
                    >
                      {link.label}
                    </NavLink>
                  ))}
                  <button
                    type="button"
                    onClick={() => {
                      setIsMobileMoreOpen(false);
                      void handleLogout();
                    }}
                    className="block w-full border-t border-[#1f1b16]/10 px-4 py-3 text-left text-sm font-bold text-[#9f2f22] transition hover:bg-[#9f2f22]/10"
                  >
                    Cerrar sesión
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
  );
}
