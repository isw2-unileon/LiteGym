import * as React from "react";
import { Link } from "react-router-dom";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

export default function SupportPage() {
  const [ticket, setTicket] = React.useState({ category: "General", title: "", description: "" });
  const [statusMessage, setStatusMessage] = React.useState<{ text: string; type: "success" | "error" } | null>(null);

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
    <div className="relative isolate text-[#1f1b16] [font-family:'Inter',system-ui,sans-serif] antialiased pb-20">
      {/* Background decoration blobs */}
      <div aria-hidden="true" className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]" />
      <div aria-hidden="true" className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10" />

      {/* Header section */}
      <header className="mb-8 flex flex-col sm:flex-row sm:items-end justify-between gap-6 max-w-2xl mx-auto">
        <div>
          <div className="mb-2.5 flex items-center gap-2.5 text-[12px] font-extrabold uppercase tracking-[0.30em] text-[#ea7130]">
            <span className="inline-block h-0.5 w-6 bg-[#ea7130]" /> Centro de Ayuda
          </div>
          <h1 className="m-0 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[40px] font-black leading-[0.95] tracking-[-0.05em] text-[#1f1b16] sm:text-[54px]">
            Soporte <span className="px-1 text-[#ea7130] [background:linear-gradient(180deg,transparent_60%,rgba(234,113,48,0.18)_60%)]">Técnico</span>
          </h1>
        </div>
        <Link 
          to="/profile" 
          className="rounded-[14px] border border-[#1f1b16]/15 bg-[#fffaf0]/80 px-5 py-3 text-[13px] font-bold backdrop-blur transition hover:bg-white hover:border-[#ea7130] hover:text-[#ea7130] shadow-sm text-center"
        >
          Volver
        </Link>
      </header>

      {/* Main glassmorphic card container */}
      <div className="max-w-2xl mx-auto rounded-[2.2rem] border border-[#1f1b16]/12 bg-[#fffaf0]/85 p-5 shadow-[0_30px_80px_rgba(47,39,27,0.15)] backdrop-blur-md sm:p-8">
        
        {/* Status notification toast */}
        {statusMessage && (
          <div className={`mb-6 rounded-2xl px-5 py-4 border text-sm font-semibold tracking-tight text-center transition ${
            statusMessage.type === "success" 
              ? "bg-[#265c52]/10 text-[#265c52] border-[#265c52]/20" 
              : "bg-[#c94b32]/10 text-[#c94b32] border-[#c94b32]/20"
          }`}>
            {statusMessage.text}
          </div>
        )}

        {/* Support Ticket Submission Form */}
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          <div className="flex flex-col">
            <label className="mb-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-wider text-[#3a332c]/85">
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
            <label className="mb-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-wider text-[#3a332c]/85">
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
            <label className="mb-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-wider text-[#3a332c]/85">
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
            className="group relative mt-2 w-full overflow-hidden rounded-[14px] bg-[#ea7130] px-5 py-4 text-[14.5px] font-black text-[#1f1b16] shadow-[0_12px_24px_rgba(234,113,48,0.22)] transition hover:-translate-y-px hover:bg-[#ff8b47]"
          >
            Enviar Ticket
          </button>
        </form>

      </div>
    </div>
  );
}