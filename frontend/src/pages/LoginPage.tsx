import * as React from "react";
import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import LegalLinks from "../components/LegalLinks";
import { apiUrl } from "../lib/api";
import { setAuthToken } from "../lib/authSession";

type LoginStatus = "idle" | "loading" | "success" | "error";

function normalizeEmail(value: string) {
  return value.trim().toLowerCase();
}

function translateLoginError(message?: string | null) {
  switch (message) {
    case "invalid credentials":
      return "El correo o la contraseña no son correctos.";
    case "unverified email":
      return "Tu cuenta no está verificada. Por favor revisa tu correo electrónico.";
    case "invalid email address":
      return "Introduce un email valido.";
    case "login rate limit exceeded":
      return "Has intentado iniciar sesion demasiadas veces. Espera unos segundos y vuelve a intentarlo.";
    default:
      return message ?? "No se pudo iniciar sesion.";
  }
}

export default function LoginPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const isVerified = searchParams.get("verified") === "true";
  const isRegistered = searchParams.get("registered") === "true";
  const isPasswordReset = searchParams.get("reset") === "true";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loginStatus, setLoginStatus] = useState<LoginStatus>("idle");
  const [loginMessage, setLoginMessage] = useState("");
  const [resendStatus, setResendStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [resendMessage, setResendMessage] = useState("");

  const handleResendEmail = async () => {
    setResendStatus("loading");
    setResendMessage("");

    try {
      const response = await fetch(apiUrl("/api/auth/resend-verification"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          email: normalizeEmail(email),
        }),
      });

      if (!response.ok) {
        const payload = await response.json().catch(() => null);
        setResendStatus("error");
        setResendMessage(
          payload?.error === "user is already verified"
            ? "Tu cuenta ya está verificada. Intenta iniciar sesión."
            : "Error al reenviar el correo.",
        );
        return;
      }

      setResendStatus("success");
      setResendMessage("Se ha reenviado el correo de verificación.");
    } catch {
      setResendStatus("error");
      setResendMessage("No se pudo conectar con el servidor.");
    }
  };

  const handleLogin = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoginStatus("loading");
    setLoginMessage("");

    const normalizedEmail = normalizeEmail(email);

    try {
      const response = await fetch(apiUrl("/api/auth/login"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          email: normalizedEmail,
          password,
        }),
      });

      if (!response.ok) {
        const payload = await response.json().catch(() => null);
        setLoginStatus("error");
        setLoginMessage(translateLoginError(payload?.error));
        return;
      }

      const payload = (await response.json().catch(() => null)) as { token?: string } | null;
      if (payload?.token) {
        setAuthToken(payload.token);
      }

      setLoginStatus("success");
      setLoginMessage("Sesion iniciada. Entrando al panel...");
      navigate("/dashboard", { replace: true });
    } catch {
      setLoginStatus("error");
      setLoginMessage("No se pudo conectar con el backend.");
    }
  };

  return (
    <main className="min-h-screen overflow-hidden bg-[#f4efe2] text-[#1f1b16]">
      <section className="relative isolate min-h-screen px-6 py-8 sm:px-10 lg:px-16">
        <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,_rgba(234,113,48,0.24),_transparent_30%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_52%,_#d8e1d0_100%)]" />
        <div className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]" />
        <div className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10" />

        <div className="mx-auto grid min-h-[calc(100vh-4rem)] max-w-6xl items-center gap-10 lg:grid-cols-[1.05fr_0.95fr]">
          <div className="max-w-2xl animate-[rise_700ms_ease-out_both]">
            <p className="mb-5 inline-flex rounded-full border border-[#1f1b16]/15 bg-white/35 px-4 py-2 text-sm font-semibold uppercase tracking-[0.28em] text-[#265c52] backdrop-blur">
              LiteGym
            </p>
            <h1 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-5xl font-black leading-[0.95] tracking-[-0.06em] text-[#1f1b16] sm:text-7xl">
              Entra, entrena y controla tu progreso.
            </h1>
          </div>

          <div className="animate-[rise_700ms_ease-out_120ms_both] rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-5 shadow-[0_30px_80px_rgba(47,39,27,0.20)] backdrop-blur-md sm:p-8">
            <div className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-6 text-[#fffaf0] shadow-inner">
              <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">Inicio de sesion</p>
              <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]">
                Accede a tu cuenta
              </h2>
            </div>

            {isVerified && (
              <div className="mt-5 rounded-2xl bg-[#265c52]/10 px-4 py-3 text-sm font-semibold text-[#265c52]">
                ¡Tu cuenta ha sido verificada con éxito! Ya puedes iniciar sesión.
              </div>
            )}

            {isRegistered && (
              <div className="mt-5 rounded-2xl bg-[#265c52]/10 px-4 py-3 text-sm font-semibold text-[#265c52]">
                Cuenta creada. Revisa tu correo para verificar tu cuenta.
              </div>
            )}

            {isPasswordReset && (
              <div className="mt-5 rounded-2xl bg-[#265c52]/10 px-4 py-3 text-sm font-semibold text-[#265c52]">
                Tu contraseña ha sido actualizada con éxito. Ya puedes iniciar sesión.
              </div>
            )}

            <form className="mt-7 space-y-5" onSubmit={handleLogin}>
              <label className="block">
                <span className="text-sm font-bold text-[#3a332c]">Email</span>
                <input
                  className="mt-2 w-full rounded-2xl border border-[#1f1b16]/15 bg-white/75 px-4 py-3 text-base outline-none ring-[#ea7130]/25 transition focus:border-[#ea7130] focus:ring-4"
                  type="email"
                  value={email}
                  autoComplete="email"
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder={"example@example.com"}
                  required
                />
              </label>

              <label className="block">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-bold text-[#3a332c]">Contraseña</span>
                  <Link to="/forgot-password" className="text-xs font-semibold text-[#ea7130] transition hover:text-[#ff8b47]">
                    ¿Olvidaste tu contraseña?
                  </Link>
                </div>
                <input
                  className="mt-2 w-full rounded-2xl border border-[#1f1b16]/15 bg-white/75 px-4 py-3 text-base outline-none ring-[#ea7130]/25 transition focus:border-[#ea7130] focus:ring-4"
                  type="password"
                  value={password}
                  autoComplete="current-password"
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder={"••••••••"}
                  required
                />
              </label>

              <button
                className="group relative w-full overflow-hidden rounded-2xl bg-[#ea7130] px-5 py-4 text-base font-black text-[#1f1b16] shadow-[0_18px_35px_rgba(234,113,48,0.35)] transition hover:-translate-y-0.5 hover:bg-[#ff8b47] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
                type="submit"
                disabled={loginStatus === "loading"}
              >
                <span className="relative z-10">{loginStatus === "loading" ? "Entrando..." : "Iniciar sesion"}</span>
              </button>
            </form>

            {loginMessage && (
              <div
                className={`mt-5 rounded-2xl px-4 py-3 text-sm font-semibold ${
                  loginStatus === "success" ? "bg-[#265c52]/10 text-[#265c52]" : "bg-[#c94b32]/10 text-[#9f2f22]"
                }`}
              >
                <p>{loginMessage}</p>
                {loginMessage === "Tu cuenta no está verificada. Por favor revisa tu correo electrónico." && (
                  <div className="mt-3">
                    <button
                      onClick={handleResendEmail}
                      disabled={resendStatus === "loading" || resendStatus === "success"}
                      className="text-[#ea7130] underline underline-offset-2 transition hover:text-[#ff8b47] disabled:opacity-50"
                      type="button"
                    >
                      {resendStatus === "loading" ? "Reenviando..." : "¿No recibiste el correo? Reenviar"}
                    </button>
                    {resendMessage && (
                      <p className={`mt-2 ${resendStatus === "success" ? "text-[#265c52]" : "text-[#9f2f22]"}`}>
                        {resendMessage}
                      </p>
                    )}
                  </div>
                )}
              </div>
            )}

            <p className="mt-6 text-sm font-semibold text-[#5b5347]">
              ¿No tienes cuenta?{" "}
              <Link className="text-[#ea7130] underline decoration-[#ea7130]/35 underline-offset-4" to="/register">
                Registrate
              </Link>
            </p>

            <div className="mt-6 border-t border-[#1f1b16]/10 pt-6">
              <p className="mb-3 text-xs font-semibold uppercase tracking-[0.28em] text-[#5b5347]">
                Información legal
              </p>
              <LegalLinks />
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
