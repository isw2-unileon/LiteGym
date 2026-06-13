import * as React from "react";
import {HelloHeader} from "@/components/HelloHeader.tsx";
import {useOutletContext} from "react-router-dom";
import type {LayoutUser} from "@/components/AppLayout.tsx";
import {Card, CardHeader} from "@/components/Card.tsx";
import {useIsMobile} from "@/lib/useIsMobile.ts";


type OutletContext = {
  user?: LayoutUser | null;
};

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";


export default function SupportPage() {
  const [ticket, setTicket] = React.useState({ category: "General", title: "", description: "" });
  const [statusMessage, setStatusMessage] = React.useState<{ text: string; type: "success" | "error" } | null>(null);
  const { user } = useOutletContext<OutletContext>();
  const isMobile = useIsMobile();

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
      <div className="px-4 pt-5 sm:px-6 sm:pt-8 md:px-8">
        <section className="mx-auto mb-6 max-w-[1280px] items-start gap-6 md:grid-cols-[1fr_auto]">
          <HelloHeader page={"CENTRO DE AYUDA"} user={user?.username ?? "Atleta"} />
        </section>
        <section className="mx-auto max-w-[1280px] gap-[18px] xl:grid-cols-[minmax(0,1.55fr)_minmax(20rem,0.95fr)]">
          {isMobile ? (
            <MobileSupportTicket
              ticket={ticket}
              statusMessage={statusMessage}
              onTicketChange={setTicket}
              onSubmit={handleSubmit}
            />
          ) : (
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
                  className="inline-flex w-full cursor-pointer items-center justify-center gap-1.5 rounded-[10px] border-0 bg-[#1f1b16] px-3.5 py-4 text-[14px] font-extrabold leading-none tracking-[0.03em] text-[#f1a45b] no-underline shadow-[0_10px_22px_rgba(31,27,22,0.18)] transition hover:-translate-y-px hover:bg-[#2c261f] sm:w-auto"
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
          )}
        </section>
      </div>
    </main>
  );
}

function MobileSupportTicket({
  ticket,
  statusMessage,
  onTicketChange,
  onSubmit,
}: {
  ticket: { category: string; title: string; description: string };
  statusMessage: { text: string; type: "success" | "error" } | null;
  onTicketChange: React.Dispatch<React.SetStateAction<{ category: string; title: string; description: string }>>;
  onSubmit: (event: React.FormEvent) => void;
}) {
  return (
    <section className="overflow-hidden rounded-[30px] border border-[#1f1b16]/10 bg-[#fffaf0]/88 shadow-[0_14px_34px_rgba(31,27,22,0.1)]">
      <div className="bg-[#ea7130] px-5 py-4">
        <p className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.18em] text-[#1f1b16]/70">
          Soporte
        </p>
        <h2 className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[31px] font-black leading-none tracking-[-0.055em] text-[#1f1b16]">
          Cuenta qué pasa
        </h2>
      </div>
      <form onSubmit={onSubmit} className="grid gap-3 p-4">
        <select
          className="w-full rounded-[18px] border border-[#1f1b16]/12 bg-white/82 px-4 py-4 text-sm font-black text-[#1f1b16] outline-none focus:border-[#ea7130]"
          value={ticket.category}
          onChange={(event) => onTicketChange((current) => ({ ...current, category: event.target.value }))}
        >
          <option value="General">Problema general</option>
          <option value="Ejercicios">Ejercicios oficiales</option>
          <option value="IA">Asistente IA</option>
        </select>
        <input
          type="text"
          required
          placeholder="Asunto"
          className="w-full rounded-[18px] border border-[#1f1b16]/12 bg-white/82 px-4 py-4 text-sm font-black text-[#1f1b16] outline-none focus:border-[#ea7130]"
          value={ticket.title}
          onChange={(event) => onTicketChange((current) => ({ ...current, title: event.target.value }))}
        />
        <textarea
          required
          rows={5}
          placeholder="Describe el problema con detalle..."
          className="w-full resize-none rounded-[18px] border border-[#1f1b16]/12 bg-white/82 px-4 py-4 text-sm font-bold leading-relaxed text-[#1f1b16] outline-none focus:border-[#ea7130]"
          value={ticket.description}
          onChange={(event) => onTicketChange((current) => ({ ...current, description: event.target.value }))}
        />
        <button type="submit" className="rounded-[18px] bg-[#1f1b16] px-5 py-4 text-sm font-black uppercase tracking-[0.08em] text-[#f1a45b]">
          Enviar ticket
        </button>
        {statusMessage && (
          <p className={[
            "rounded-[18px] px-4 py-3 text-center text-sm font-black",
            statusMessage.type === "success" ? "bg-[#265c52]/10 text-[#265c52]" : "bg-[#9f2f22]/10 text-[#9f2f22]",
          ].join(" ")}>
            {statusMessage.text}
          </p>
        )}
      </form>
    </section>
  );
}
