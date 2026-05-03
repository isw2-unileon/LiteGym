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
      
      setStatusMessage({ text: "Ticket enviado correctamente. ", type: "success" });
      setTicket({ category: "General", title: "", description: "" }); // Reset
    } catch {
      setStatusMessage({ text: "Error al enviar el ticket.", type: "error" });
    }
  };

  return (
    <main className="min-h-screen bg-[#f4efe2] text-[#1f1b16] py-12 px-6 flex items-center justify-center">
      <div className="w-full max-w-xl rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-8 shadow-xl backdrop-blur-md">
        <div className="mb-6 flex justify-between items-center">
          <h1 className="text-3xl font-black">Soporte Técnico</h1>
          <Link to="/profile" className="text-sm font-bold text-[#ea7130] hover:underline">Volver</Link>
        </div>
        
        {statusMessage && (
          <div className={`mb-6 rounded-xl p-4 text-center font-bold ${statusMessage.type === "success" ? "bg-[#265c52]/20 text-[#265c52]" : "bg-[#c94b32]/20 text-[#c94b32]"}`}>
            {statusMessage.text}
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          <div>
            <label className="block mb-2 font-bold text-sm text-[#5d5348]">Categoría</label>
            <select className="w-full rounded-xl border-none bg-white p-3 shadow-inner" value={ticket.category} onChange={(e) => setTicket({ ...ticket, category: e.target.value })}>
              <option value="General">Problema General</option>
              <option value="Ejercicios">Error en Ejercicios Oficiales</option>
              <option value="IA">Comportamiento del Asistente IA</option>
            </select>
          </div>
          <div>
            <label className="block mb-2 font-bold text-sm text-[#5d5348]">Asunto del ticket</label>
            <input type="text" required placeholder="Ej: No me carga una rutina" className="w-full rounded-xl border-none bg-white p-3 shadow-inner" value={ticket.title} onChange={(e) => setTicket({ ...ticket, title: e.target.value })} />
          </div>
          <div>
            <label className="block mb-2 font-bold text-sm text-[#5d5348]">Descripción detallada</label>
            <textarea required rows={5} placeholder="Explica detalladamente tu problema..." className="w-full rounded-xl border-none bg-white p-3 shadow-inner resize-none" value={ticket.description} onChange={(e) => setTicket({ ...ticket, description: e.target.value })} />
          </div>
          <button type="submit" className="mt-4 rounded-xl bg-[#1f1b16] py-3 font-bold text-[#fffaf0] transition hover:bg-[#ea7130]">Enviar Ticket</button>
        </form>
      </div>
    </main>
  );
}