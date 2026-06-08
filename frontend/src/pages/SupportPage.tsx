import * as React from "react";
import {HelloHeader} from "@/components/HelloHeader.tsx";
import {useOutletContext} from "react-router-dom";
import type {LayoutUser} from "@/components/AppLayout.tsx";
import {Card, CardHeader} from "@/components/Card.tsx";


type OutletContext = {
  user?: LayoutUser | null;
};

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";


export default function SupportPage() {
  const [ticket, setTicket] = React.useState({ category: "General", title: "", description: "" });
  const [statusMessage, setStatusMessage] = React.useState<{ text: string; type: "success" | "error" } | null>(null);
  const { user } = useOutletContext<OutletContext>();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await fetch(`${API_BASE_URL}/api/tickets`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(ticket),
        credentials: "include",
      });

      if (!response.ok) throw new Error("Error creating ticket");
      
      setStatusMessage({ text: "Ticket enviado correctamente.", type: "success" });
      setTicket({ category: "General", title: "", description: "" }); // Reset
    } catch {
      setStatusMessage({ text: "Error al enviar el ticket.", type: "error" });
    }
  };

  return (
    <main className="relative isolate text-[#1f1b16] [font-family:'Inter',system-ui,sans-serif] antialiased pb-20">
      {/* Background decoration blobs */}
      <div aria-hidden="true" className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]" />
      <div aria-hidden="true" className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10" />
      <div className="px-6  pt-8 sm:px-8">
        <section className="mx-auto mb-6 max-w-[1280px] items-start gap-6 md:grid-cols-[1fr_auto]">
          <HelloHeader page={"CENTRO DE AYUDA"} user={user?.username ?? "Atleta"} />
        </section>
        <section className="mx-auto max-w-[1280px] gap-[18px] xl:grid-cols-[minmax(0,1.55fr)_minmax(20rem,0.95fr)]">
          <Card accent="#ea7130">
            <CardHeader kicker={"SOPORTE TÉCNICO"} title={"Rellena un ticket"} />
            <form onSubmit={handleSubmit} className="flex flex-col gap-5 mt-4">
              <div className="flex flex-col">
                <label className="mb-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[16px] font-bold uppercase tracking-wider text-[#3a332c]/85">
                  Categoría
                </label>
                <select
                    className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/90 px-4 py-3.5 text-[14px] font-semibold text-[#1f1b16] outline-none cursor-pointer transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/12"
                    value={ticket.category}
                    onChange={(e) => setTicket({ ...ticket, category: e.target.value })}
                >
                  <option value="General">Problema General</option>
                  <option value="Ejercicios">Error en Ejercicios Oficiales</option>
                  <option value="IA">Comportamiento del Asistente IA</option>
                </select>
              </div>

              <div className="flex flex-col">
                <label className="mb-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[16px] font-bold uppercase tracking-wider text-[#3a332c]/85">
                  Asunto del ticket
                </label>
                <input
                    type="text"
                    required
                    placeholder="Ej: No me carga una rutina"
                    className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/70 px-4 py-3.5 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/12"
                    value={ticket.title}
                    onChange={(e) => setTicket({ ...ticket, title: e.target.value })}
                />
              </div>

              <div className="flex flex-col">
                <label className="mb-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[16px] font-bold uppercase tracking-wider text-[#3a332c]/85">
                  Descripción detallada
                </label>
                <textarea
                    required
                    rows={5}
                    placeholder="Explica detalladamente tu problema..."
                    className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/70 px-4 py-3.5 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/12 resize-none leading-relaxed"
                    value={ticket.description}
                    onChange={(e) => setTicket({ ...ticket, description: e.target.value })}
                />
              </div>

              <button
                  type="submit"
                  className="inline-flex cursor-pointer items-center justify-center gap-1.5 rounded-[10px] border-0 bg-[#1f1b16] px-3.5 py-4 text-[14px] font-extrabold leading-none tracking-[0.03em] text-[#f1a45b] no-underline shadow-[0_10px_22px_rgba(31,27,22,0.18)] transition hover:-translate-y-px hover:bg-[#2c261f]"
              >
                Enviar Ticket
              </button>
            </form>
            {statusMessage && (
                <div className={`mt-4 mb-6 rounded-2xl px-5 py-4 border text-sm font-semibold tracking-tight text-center transition ${
                    statusMessage.type === "success"
                        ? "bg-[#265c52]/10 text-[#265c52] border-[#265c52]/20"
                        : "bg-[#c94b32]/10 text-[#c94b32] border-[#c94b32]/20"
                }`}>
                  {statusMessage.text}
                </div>
                )}
          </Card>
        </section>
      </div>
    </main>
  );
}