import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { apiUrl } from "../lib/api";

export default function NotFoundRedirect() {
  const navigate = useNavigate();

  useEffect(() => {
    const redirectBySession = async () => {
      try {
        const response = await fetch(apiUrl("/api/auth/me"), {
          credentials: "include",
        });

        navigate(response.ok ? "/dashboard" : "/", { replace: true });
      } catch {
        navigate("/", { replace: true });
      }
    };

    void redirectBySession();
  }, [navigate]);

  return (
    <main className="grid min-h-screen place-items-center bg-[#f4efe2] px-6 text-[#1f1b16]">
      <p className="rounded-2xl bg-white/60 px-5 py-4 text-sm font-bold shadow-[0_20px_50px_rgba(47,39,27,0.12)]">
        Redirigiendo...
      </p>
    </main>
  );
}
