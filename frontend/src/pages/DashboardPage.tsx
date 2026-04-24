import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiUrl } from "../lib/api";

type AccessStatus = "checking" | "allowed" | "blocked" | "error";

export default function DashboardPage() {
  const navigate = useNavigate();
  const [accessStatus, setAccessStatus] = useState<AccessStatus>("checking");

  const checkApiAccess = async () => {
    setAccessStatus("checking");

    try {
      const response = await fetch(apiUrl("/api/exercises"), {
        credentials: "include",
      });

      if (response.status === 401) {
        setAccessStatus("blocked");
        navigate("/", { replace: true });
        return;
      }

      if (!response.ok) {
        setAccessStatus("error");
        return;
      }

      setAccessStatus("allowed");
    } catch {
      setAccessStatus("error");
    }
  };

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

  useEffect(() => {
    const checkSession = async () => {
      setAccessStatus("checking");

      try {
        const response = await fetch(apiUrl("/api/auth/me"), {
          credentials: "include",
        });

        if (response.status === 401) {
          setAccessStatus("blocked");
          navigate("/", { replace: true });
          return;
        }

        if (!response.ok) {
          setAccessStatus("error");
          return;
        }

        setAccessStatus("allowed");
      } catch {
        setAccessStatus("error");
      }
    };

    void checkSession();
  }, [navigate]);

  return (
    <main className="min-h-screen bg-[#f4efe2] px-6 py-8 text-[#1f1b16] sm:px-10 lg:px-16">
      <section className="mx-auto max-w-6xl">
        <header className="flex flex-col gap-6 rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-6 shadow-[0_20px_60px_rgba(47,39,27,0.12)] backdrop-blur md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#265c52]">Grupo 16 Fitness</p>
            <h1 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.05em] sm:text-5xl">
              Dashboard
            </h1>
          </div>

          <button
            className="rounded-2xl bg-[#1f1b16] px-5 py-3 text-sm font-bold text-[#fffaf0] transition hover:bg-[#ea7130] hover:text-[#1f1b16]"
            type="button"
            onClick={handleLogout}
          >
            Cerrar sesion
          </button>
        </header>

        <section className="mt-8 rounded-[2rem] border border-dashed border-[#1f1b16]/20 bg-white/45 p-6">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-sm font-black uppercase tracking-[0.18em] text-[#265c52]">Comprobar acceso</p>
              <p className="mt-1 text-sm text-[#5d5348]">Prueba el acceso a la api.</p>
            </div>

            <button
              className="rounded-2xl border border-[#1f1b16]/15 px-4 py-3 text-sm font-black text-[#1f1b16] transition hover:bg-[#1f1b16] hover:text-[#fffaf0] disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              onClick={checkApiAccess}
              disabled={accessStatus === "checking"}
            >
              {accessStatus === "checking" ? "Probando..." : "Probar"}
            </button>
          </div>

          <AccessStatusMessage status={accessStatus} />
        </section>
      </section>
    </main>
  );
}

function AccessStatusMessage({ status }: { status: AccessStatus }) {
  if (status === "checking") {
    return <p className="mt-4 text-sm text-[#6b5d4d]">Un momento, estamos probando el acceso.</p>;
  }

  if (status === "allowed") {
    return <p className="mt-4 text-sm font-bold text-[#265c52]">Todo listo, ya tienes acceso.</p>;
  }

  if (status === "blocked") {
    return <p className="mt-4 text-sm font-bold text-[#9f2f22]">Todavia no puedes entrar. Inicia sesion primero.</p>;
  }

  return <p className="mt-4 text-sm font-bold text-[#9f2f22]">No he podido comprobarlo. Revisa que el backend este activo.</p>;
}
