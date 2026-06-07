import * as React from "react";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { apiUrl } from "../lib/api";

type RegisterStatus = "idle" | "loading" | "success" | "error";

function normalizeEmail(value: string) {
  return value.trim().toLowerCase();
}

function isValidEmail(value: string) {
  const email = normalizeEmail(value);
  const parts = email.split("@");

  if (parts.length !== 2) {
    return false;
  }

  const [localPart, domain] = parts;
  if (!localPart || !domain) {
    return false;
  }

  if (!domain.includes(".") || domain.startsWith(".") || domain.endsWith(".") || domain.includes("..")) {
    return false;
  }

  return true;
}

function translateRegisterError(message?: string | null) {
  switch (message) {
    case "demasiados intentos de registro, prueba de nuevo en unos segundos":
      return "Has intentado registrarte demasiadas veces. Espera unos segundos y vuelve a intentarlo.";
    case "username or email already registered":
    case "email already registered":
      return "Ya existe un usuario con ese correo electronico.";
    case "invalid email address":
      return "Introduce un email valido con usuario y dominio.";
    case "registration rate limit exceeded":
      return "Has intentado registrarte demasiadas veces. Espera unos segundos y vuelve a intentarlo.";
    default:
      return message ?? "No se pudo crear la cuenta.";
  }
}

export default function RegisterPage() {
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [registerStatus, setRegisterStatus] = useState<RegisterStatus>("idle");
  const [registerMessage, setRegisterMessage] = useState("");
  const emailValidationMessage =
    email.trim() !== "" && !isValidEmail(email) ? "Introduce un email valido con usuario y dominio." : "";

  const handleRegister = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setRegisterMessage("");

    if (password !== confirmPassword) {
      setRegisterStatus("error");
      setRegisterMessage("Las contrasenas no coinciden.");
      return;
    }

    if (!isValidEmail(email)) {
      setRegisterStatus("error");
      setRegisterMessage("Introduce un email valido con usuario y dominio.");
      return;
    }

    setRegisterStatus("loading");

    const normalizedEmail = normalizeEmail(email);

    try {
      const response = await fetch(apiUrl("/api/auth/register"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          username,
          email: normalizedEmail,
          password,
        }),
      });

      if (!response.ok) {
        const payload = await response.json().catch(() => null);
        setRegisterStatus("error");
        setRegisterMessage(translateRegisterError(payload?.error));
        return;
      }

      setRegisterStatus("success");
      setRegisterMessage("Registro exitoso. Por favor, revisa tu correo electrónico para verificar tu cuenta antes de iniciar sesión.");
    } catch {
      setRegisterStatus("error");
      setRegisterMessage("No se pudo conectar con el backend.");
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
              Crea tu cuenta y empieza a entrenar con orden.
            </h1>
          </div>

          <div className="animate-[rise_700ms_ease-out_120ms_both] rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-5 shadow-[0_30px_80px_rgba(47,39,27,0.20)] backdrop-blur-md sm:p-8">
            <div className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-6 text-[#fffaf0] shadow-inner">
              <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f7d08a]">Registro</p>
              <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]">
                Abre tu cuenta
              </h2>
            </div>

            <form className="mt-7 space-y-5" onSubmit={handleRegister}>
              <label className="block">
                <span className="text-sm font-bold text-[#3a332c]">Nombre de usuario</span>
                <input
                  className="mt-2 w-full rounded-2xl border border-[#1f1b16]/15 bg-white/75 px-4 py-3 text-base outline-none ring-[#265c52]/25 transition focus:border-[#265c52] focus:ring-4"
                  type="text"
                  value={username}
                  autoComplete="username"
                  onChange={(event) => setUsername(event.target.value)}
                  placeholder={"Nombre de usuario"}
                  required
                />
              </label>

              <label className="block">
                <span className="text-sm font-bold text-[#3a332c]">Email</span>
                <input
                  className="mt-2 w-full rounded-2xl border border-[#1f1b16]/15 bg-white/75 px-4 py-3 text-base outline-none ring-[#265c52]/25 transition focus:border-[#265c52] focus:ring-4"
                  type="email"
                  value={email}
                  autoComplete="email"
                  onChange={(event) => {
                    setEmail(event.target.value);
                    if (registerStatus === "error" && registerMessage === "Introduce un email valido con usuario y dominio.") {
                      setRegisterMessage("");
                      setRegisterStatus("idle");
                    }
                  }}
                  placeholder={"example@example.com"}
                  required
                />
                {emailValidationMessage && (
                  <span className="mt-2 block text-sm font-semibold text-[#9f2f22]">{emailValidationMessage}</span>
                )}
              </label>

              <label className="block">
                <span className="text-sm font-bold text-[#3a332c]">Contrasena</span>
                <input
                  className="mt-2 w-full rounded-2xl border border-[#1f1b16]/15 bg-white/75 px-4 py-3 text-base outline-none ring-[#265c52]/25 transition focus:border-[#265c52] focus:ring-4"
                  type="password"
                  value={password}
                  autoComplete="new-password"
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder={"••••••••"}
                  required
                />
              </label>

              <label className="block">
                <span className="text-sm font-bold text-[#3a332c]">Repite la contrasena</span>
                <input
                  className="mt-2 w-full rounded-2xl border border-[#1f1b16]/15 bg-white/75 px-4 py-3 text-base outline-none ring-[#265c52]/25 transition focus:border-[#265c52] focus:ring-4"
                  type="password"
                  value={confirmPassword}
                  autoComplete="new-password"
                  onChange={(event) => setConfirmPassword(event.target.value)}
                  placeholder={"••••••••"}
                  required
                />
              </label>

              <button
                className="group relative w-full overflow-hidden rounded-2xl bg-[#ea7130] px-5 py-4 text-base font-black text-[##fffaf0] shadow-[0_18px_35px_rgba(38,92,82,0.30)] transition hover:-translate-y-0.5 hover:bg-[#ff8b47] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
                type="submit"
                disabled={registerStatus === "loading"}
              >
                <span className="relative z-10">{registerStatus === "loading" ? "Creando cuenta..." : "Crear cuenta"}</span>
              </button>
            </form>

            {registerMessage && (
              <p
                className={`mt-5 rounded-2xl px-4 py-3 text-sm font-semibold ${
                  registerStatus === "success" ? "bg-[#265c52]/10 text-[#265c52]" : "bg-[#c94b32]/10 text-[#9f2f22]"
                }`}
              >
                {registerMessage}
              </p>
            )}

            <p className="mt-6 text-sm font-semibold text-[#5b5347]">
              ¿Ya tienes cuenta?{" "}
              <Link className="text-[#ea7130] underline decoration-[#ea7130]/35 underline-offset-4" to="/">
                Inicia sesion
              </Link>
            </p>
          </div>
        </div>
      </section>
    </main>
  );
}
