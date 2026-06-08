import * as React from "react";
import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { apiUrl } from "../lib/api";

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const navigate = useNavigate();

  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [message, setMessage] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!token) {
      setStatus("error");
      setMessage("El enlace de recuperación es inválido o falta el token.");
      return;
    }

    if (password !== confirmPassword) {
      setStatus("error");
      setMessage("Las contraseñas no coinciden.");
      return;
    }

    setStatus("loading");
    setMessage("");

    try {
      const response = await fetch(apiUrl("/api/auth/reset-password"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token, new_password: password }),
      });

      if (response.ok) {
        setStatus("success");
        setMessage("¡Tu contraseña ha sido restablecida correctamente! Redirigiendo al inicio de sesión...");
        setTimeout(() => {
          navigate("/?verified=true"); // We can reuse the verified banner or we can use another param
        }, 2500);
      } else {
        const payload = await response.json().catch(() => null);
        setStatus("error");
        if (payload?.error === "invalid or expired token") {
          setMessage("El enlace ha expirado o ya ha sido utilizado.");
        } else {
          setMessage(payload?.error || "Ocurrió un error al restablecer la contraseña.");
        }
      }
    } catch {
      setStatus("error");
      setMessage("Error de conexión. Revisa tu internet e inténtalo de nuevo.");
    }
  };

  if (!token) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#f4efe2] px-4 font-['Inter',sans-serif]">
        <div className="w-full max-w-md rounded-[1.5rem] bg-[#1f1b16] p-8 text-center text-[#fffaf0] shadow-2xl">
          <h2 className="mb-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-2xl font-black text-[#ea7130]">
            Enlace inválido
          </h2>
          <p className="mb-8 text-[#fffaf0]/80">
            Falta el token de seguridad. Por favor, asegúrate de haber copiado el enlace completo de tu correo.
          </p>
          <Link
            to="/forgot-password"
            className="inline-block rounded-2xl bg-[#ea7130] w-full px-5 py-4 text-base font-black text-[#fffaf0] transition hover:-translate-y-0.5 hover:bg-[#ff8b47]"
          >
            Solicitar nuevo enlace
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen bg-[#f4efe2] font-['Inter',sans-serif]">
      {/* Left panel */}
      <div className="hidden w-1/2 flex-col justify-between bg-[#1f1b16] p-12 text-[#fffaf0] lg:flex">
        <div>
          <h1 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-5xl font-black tracking-[-0.04em] text-[#ea7130]">LiteGym</h1>
        </div>
        <div className="mb-20">
          <p className="text-xl font-medium tracking-tight text-[#f1a45b] opacity-80 uppercase">Seguridad</p>
          <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-6xl font-black leading-[1.05] tracking-[-0.05em]">
            Crea una nueva<br />contraseña.
          </h2>
        </div>
      </div>

      {/* Right panel */}
      <div className="flex w-full flex-col justify-center px-8 lg:w-1/2 lg:px-24">
        <div className="mx-auto w-full max-w-md">
          <div className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-6 text-[#fffaf0] shadow-inner">
            <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">Paso final</p>
            <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]">
              Restablecer
            </h2>
          </div>

          <form className="mt-7 space-y-5" onSubmit={handleSubmit}>
            {status === "error" && (
              <div className="rounded-2xl bg-[#ea7130]/10 px-4 py-3 text-sm font-semibold text-[#ea7130]">
                {message}
              </div>
            )}
            
            {status === "success" && (
              <div className="rounded-2xl bg-[#265c52]/10 px-4 py-3 text-sm font-semibold text-[#265c52]">
                {message}
              </div>
            )}

            {status !== "success" && (
              <>
                <label className="block">
                  <span className="text-sm font-bold text-[#3a332c]">Nueva contraseña</span>
                  <input
                    className="mt-2 w-full rounded-2xl border border-[#1f1b16]/15 bg-white/75 px-4 py-3 text-base outline-none ring-[#ea7130]/25 transition focus:border-[#ea7130] focus:ring-4"
                    type="password"
                    placeholder="Escribe tu nueva contraseña"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                  />
                </label>

                <label className="block">
                  <span className="text-sm font-bold text-[#3a332c]">Confirmar nueva contraseña</span>
                  <input
                    className="mt-2 w-full rounded-2xl border border-[#1f1b16]/15 bg-white/75 px-4 py-3 text-base outline-none ring-[#ea7130]/25 transition focus:border-[#ea7130] focus:ring-4"
                    type="password"
                    placeholder="Repite la contraseña"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    required
                  />
                </label>

                <button
                  type="submit"
                  disabled={status === "loading"}
                  className="mt-6 w-full rounded-2xl bg-[#ea7130] px-5 py-4 text-base font-black text-[#fffaf0] shadow-[0_18px_35px_rgba(234,113,48,0.25)] transition hover:-translate-y-0.5 hover:bg-[#ff8b47] disabled:opacity-50"
                >
                  {status === "loading" ? "Actualizando..." : "Guardar contraseña"}
                </button>
              </>
            )}
          </form>
        </div>
      </div>
    </div>
  );
}
