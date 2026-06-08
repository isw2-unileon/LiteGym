import * as React from "react";
import { useState } from "react";
import { Link } from "react-router-dom";
import { apiUrl } from "../lib/api";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [message, setMessage] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setStatus("loading");
    setMessage("");

    try {
      const response = await fetch(apiUrl("/api/auth/forgot-password"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });

      if (response.ok) {
        setStatus("success");
        setMessage("Si el correo está registrado en nuestra base de datos, recibirás un enlace de recuperación en los próximos minutos.");
      } else {
        setStatus("error");
        setMessage("No se pudo enviar la solicitud. Por favor, inténtalo más tarde.");
      }
    } catch {
      setStatus("error");
      setMessage("Error de conexión. Revisa tu internet e inténtalo de nuevo.");
    }
  };

  return (
    <div className="flex min-h-screen bg-[#f4efe2] font-['Inter',sans-serif]">
      {/* Left panel */}
      <div className="hidden w-1/2 flex-col justify-between bg-[#1f1b16] p-12 text-[#fffaf0] lg:flex">
        <div>
          <h1 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-5xl font-black tracking-[-0.04em] text-[#ea7130]">LiteGym</h1>
        </div>
        <div className="mb-20">
          <p className="text-xl font-medium tracking-tight text-[#f1a45b] opacity-80 uppercase">Recuperación</p>
          <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-6xl font-black leading-[1.05] tracking-[-0.05em]">
            No pierdas<br />tu progreso.
          </h2>
        </div>
      </div>

      {/* Right panel */}
      <div className="flex w-full flex-col justify-center px-8 lg:w-1/2 lg:px-24">
        <div className="mx-auto w-full max-w-md">
          <Link to="/" className="mb-8 inline-block text-sm font-semibold text-[#1f1b16]/60 hover:text-[#ea7130] transition">
            ← Volver a inicio de sesión
          </Link>
          
          <div className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-6 text-[#fffaf0] shadow-inner">
            <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">Contraseña</p>
            <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]">
              Recuperar acceso
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
                <p className="text-sm font-medium text-[#3a332c]/70 leading-relaxed mb-4">
                  Introduce el correo electrónico asociado a tu cuenta y te enviaremos un enlace para que puedas establecer una nueva contraseña.
                </p>

                <label className="block">
                  <span className="text-sm font-bold text-[#3a332c]">Email</span>
                  <input
                    className="mt-2 w-full rounded-2xl border border-[#1f1b16]/15 bg-white/75 px-4 py-3 text-base outline-none ring-[#ea7130]/25 transition focus:border-[#ea7130] focus:ring-4"
                    type="email"
                    placeholder="tu@email.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                  />
                </label>

                <button
                  type="submit"
                  disabled={status === "loading"}
                  className="mt-6 w-full rounded-2xl bg-[#ea7130] px-5 py-4 text-base font-black text-[#fffaf0] shadow-[0_18px_35px_rgba(234,113,48,0.25)] transition hover:-translate-y-0.5 hover:bg-[#ff8b47] disabled:opacity-50"
                >
                  {status === "loading" ? "Enviando..." : "Enviar enlace de recuperación"}
                </button>
              </>
            )}
          </form>
        </div>
      </div>
    </div>
  );
}
