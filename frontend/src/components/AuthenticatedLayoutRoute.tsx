import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import AppLayout, { type LayoutUser } from "./AppLayout";
import { apiUrl } from "../lib/api";

type AuthStatus = "checking" | "allowed" | "blocked";

type SessionResponse = {
  user?: LayoutUser;
};

export default function AuthenticatedLayoutRoute() {
  const [authStatus, setAuthStatus] = useState<AuthStatus>("checking");
  const [user, setUser] = useState<LayoutUser | null>(null);

  useEffect(() => {
    const checkSession = async () => {
      try {
        const response = await fetch(apiUrl("/api/auth/me"), {
          credentials: "include",
        });

        if (!response.ok) {
          setAuthStatus("blocked");
          return;
        }

        const payload = (await response.json().catch(() => ({}))) as SessionResponse;
        setUser(payload.user ?? null);
        setAuthStatus("allowed");
      } catch {
        setAuthStatus("blocked");
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

  if (authStatus === "blocked") {
    return <Navigate to="/" replace />;
  }

  return <AppLayout user={user} />;
}
