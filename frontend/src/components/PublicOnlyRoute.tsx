import { useEffect, useState } from "react";
import { Navigate, Outlet } from "react-router-dom";
import { apiUrl } from "../lib/api";

type AuthStatus = "checking" | "guest" | "authenticated";

export default function PublicOnlyRoute() {
  const [authStatus, setAuthStatus] = useState<AuthStatus>("checking");

  useEffect(() => {
    const checkSession = async () => {
      try {
        const response = await fetch(apiUrl("/api/auth/me"), {
          credentials: "include",
        });
        setAuthStatus(response.ok ? "authenticated" : "guest");
      } catch {
        setAuthStatus("guest");
      }
    };

    void checkSession();
  }, []);

  if (authStatus === "checking") {
    return (
      <main className="grid min-h-screen place-items-center bg-[#f4efe2] px-6 text-[#1f1b16]">
        <p className="rounded-2xl bg-white/60 px-5 py-4 text-sm font-bold shadow-[0_20px_50px_rgba(47,39,27,0.12)]">
          Comprobando sesion...
        </p>
      </main>
    );
  }

  if (authStatus === "authenticated") {
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
}
