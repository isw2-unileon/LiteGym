
import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { apiUrl } from "../lib/api";

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const navigate = useNavigate();
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [message, setMessage] = useState("");

  const handleVerify = async () => {
    if (!token) {
      setStatus("error");
      setMessage("No se ha proporcionado un token válido en la URL.");
      return;
    }

    setStatus("loading");
    setMessage("Verificando tu cuenta...");

    try {
      const response = await fetch(apiUrl(`/api/auth/verify-email?token=${token}`), {
        method: "GET",
      });

      if (response.ok) {
        setStatus("success");
        setMessage("¡Tu cuenta ha sido verificada correctamente! Redirigiendo al inicio de sesión...");
        setTimeout(() => {
          navigate("/?verified=true");
        }, 2500);
      } else {
        const payload = await response.json().catch(() => null);
        setStatus("error");
        setMessage(
          payload?.error === "invalid or expired token"
            ? "El token es inválido o ha expirado. Si acabas de pedir uno nuevo, usa el correo más reciente."
            : "Ha ocurrido un error al verificar tu cuenta."
        );
      }
    } catch {
      setStatus("error");
      setMessage("No se pudo conectar con el servidor.");
    }
  };

  return (
    <main className="min-h-screen overflow-hidden bg-[#f4efe2] text-[#1f1b16] flex items-center justify-center">
      <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,_rgba(234,113,48,0.24),_transparent_30%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_52%,_#d8e1d0_100%)]" />
      <div className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]" />
      <div className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10" />

      <div className="max-w-md w-full p-8 rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 shadow-[0_30px_80px_rgba(47,39,27,0.20)] backdrop-blur-md text-center animate-[rise_700ms_ease-out_both]">
        <h2 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em] mb-4">
          Verificación de Email
        </h2>
        
        {status === "idle" && (
          <div className="mt-6">
            <p className="mb-6 text-[#5b5347]">Haz clic en el botón para confirmar tu cuenta y empezar a usar LiteGym.</p>
            <button
              onClick={handleVerify}
              className="inline-block rounded-2xl bg-[#ea7130] w-full px-5 py-4 text-base font-black text-[##fffaf0] shadow-[0_18px_35px_rgba(234,113,48,0.35)] transition hover:-translate-y-0.5 hover:bg-[#ff8b47]"
            >
              Confirmar mi correo electrónico
            </button>
          </div>
        )}

        {status === "loading" && (
          <div className="animate-pulse flex space-x-4 justify-center items-center h-12 mt-6">
            <div className="rounded-full bg-[#ea7130] h-4 w-4"></div>
            <div className="rounded-full bg-[#ea7130] h-4 w-4"></div>
            <div className="rounded-full bg-[#ea7130] h-4 w-4"></div>
          </div>
        )}

        {status !== "idle" && status !== "loading" && (
          <p className={`mt-5 rounded-2xl px-4 py-3 text-sm font-semibold ${
              status === "success" ? "bg-[#265c52]/10 text-[#265c52]" : 
              "bg-[#c94b32]/10 text-[#9f2f22]"
          }`}>
            {message}
          </p>
        )}

        {status !== "loading" && (
          <div className="mt-8">
            <Link 
              to="/" 
              className="inline-block rounded-2xl bg-[#ea7130] px-5 py-4 text-base font-black text-[##fffaf0] shadow-[0_18px_35px_rgba(38,92,82,0.30)] transition hover:-translate-y-0.5 hover:bg-[#ff8b47]"
            >
              Ir a Iniciar Sesión
            </Link>
          </div>
        )}
      </div>
    </main>
  );
}
