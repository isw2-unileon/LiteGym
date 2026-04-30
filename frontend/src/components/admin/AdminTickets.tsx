import * as React from "react";
import { apiUrl } from "../../lib/api";

type Ticket = {
  id: string;
  user_id: string;
  title: string;
  description: string;
  status: string;
  created_at: string;
};

export default function AdminTickets() {
  const [tickets, setTickets] = React.useState<Ticket[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  const [statusFilter, setStatusFilter] = React.useState<"all" | "open" | "closed">("open"); 
  const [categoryFilter, setCategoryFilter] = React.useState<string>("all");

  React.useEffect(() => {
    fetchTickets();
  }, []);

  const fetchTickets = async () => {
    try {
      const response = await fetch(apiUrl("/api/tickets"), {
        credentials: "include",
      });
      if (!response.ok) throw new Error("Error al cargar los tickets");
      
      const data = await response.json();
      setTickets(data);
    } catch {
      setError("No se pudieron cargar los tickets.");
    } finally {
      setLoading(false);
    }
  };

  const handleCloseTicket = async (id: string) => {
    try {
      const response = await fetch(apiUrl(`/api/tickets/${id}/close`), {
        method: "PATCH",
        credentials: "include",
      });
      
      if (!response.ok) throw new Error("Error al cerrar");

      setTickets(tickets.map(t => 
        t.id === id ? { ...t, status: "closed" } : t
      ));
    } catch {
      alert("Hubo un error al intentar cerrar el ticket.");
    }
  };

  // --- Filter logic ---
  const filteredTickets = tickets.filter(ticket => {
    // status filter
    const matchesStatus = statusFilter === "all" || ticket.status === statusFilter;
    
    // category filter 
    const matchesCategory = categoryFilter === "all" || ticket.title.includes(`[${categoryFilter}]`);

    return matchesStatus && matchesCategory;
  });

  if (loading) return <p className="font-bold text-[#5d5348] text-center p-8">Cargando tickets...</p>;
  if (error) return <p className="font-bold text-[#c94b32] text-center p-8">{error}</p>;

  return (
    <div className="flex flex-col gap-6">
      
      {/* --- Filter --- */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center rounded-2xl bg-white/50 p-4 shadow-sm border border-[#1f1b16]/5">
        <div className="flex flex-col">
          <label className="mb-1 text-xs font-bold text-[#5d5348] uppercase tracking-wider">Estado</label>
          <select 
            className="rounded-xl border-none bg-white p-2.5 text-sm font-semibold shadow-inner focus:ring-2 focus:ring-[#f1a45b] outline-none cursor-pointer"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as "all" | "open" | "closed")}
          >
            <option value="open"> Pendientes (Abiertos)</option>
            <option value="closed"> Resueltos (Cerrados)</option>
            <option value="all"> Mostrar Todos</option>
          </select>
        </div>

        <div className="flex flex-col">
          <label className="mb-1 text-xs font-bold text-[#5d5348] uppercase tracking-wider">Categoría</label>
          <select 
            className="rounded-xl border-none bg-white p-2.5 text-sm font-semibold shadow-inner focus:ring-2 focus:ring-[#f1a45b] outline-none cursor-pointer"
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
          >
            <option value="all">Todas las categorías</option>
            <option value="General">General</option>
            <option value="Ejercicios">Ejercicios</option>
            <option value="IA">Asistente IA</option>
          </select>
        </div>

        <div className="mt-auto ml-auto">
          <span className="text-sm font-bold text-[#5d5348] bg-white px-3 py-1.5 rounded-lg shadow-sm">
            Mostrando: {filteredTickets.length}
          </span>
        </div>
      </div>

      {/* --- ticket list --- */}
      <div className="flex flex-col gap-4 max-h-[500px] overflow-y-auto pr-2">
        {tickets.length === 0 ? (
          <p className="text-[#5d5348] text-center p-8 bg-white/50 rounded-2xl">No hay tickets en el sistema.</p>
        ) : filteredTickets.length === 0 ? (
          <p className="text-[#5d5348] text-center p-8 bg-white/50 rounded-2xl">No hay tickets que coincidan con estos filtros.</p>
        ) : (
          filteredTickets.map((ticket) => (
            <div 
              key={ticket.id} 
              className={`rounded-2xl border p-5 shadow-sm transition ${
                ticket.status === "closed" 
                  ? "border-green-900/10 bg-green-50/50 opacity-75" 
                  : "border-[#1f1b16]/10 bg-white"
              }`}
            >
              <div className="mb-3 flex flex-col sm:flex-row items-start justify-between gap-4">
                <div>
                  <div className="mb-1 flex items-center gap-2">
                    <span className={`rounded-md px-2 py-0.5 text-xs font-black uppercase tracking-wider ${
                      ticket.status === "closed" ? "bg-green-200 text-green-800" : "bg-[#f1a45b] text-[#1f1b16]"
                    }`}>
                      {ticket.status}
                    </span>
                    <span className="text-xs font-bold text-[#5d5348]">
                      {new Date(ticket.created_at).toLocaleDateString()} a las {new Date(ticket.created_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}
                    </span>
                  </div>
                  <h3 className="text-lg font-black text-[#1f1b16]">{ticket.title}</h3>
                </div>
                
                {ticket.status !== "closed" && (
                  <button
                    onClick={() => handleCloseTicket(ticket.id)}
                    className="shrink-0 rounded-xl bg-[#1f1b16] px-4 py-2 text-sm font-bold text-white transition hover:bg-[#265c52] whitespace-nowrap"
                  >
                    Marcar como resuelto
                  </button>
                )}
              </div>
              
              <div className="rounded-xl bg-black/5 p-4">
                <p className="text-sm text-[#3b352f] whitespace-pre-wrap">{ticket.description}</p>
              </div>
              <p className="mt-4 text-xs font-semibold text-[#5d5348]/60">ID Usuario: <span className="font-mono">{ticket.user_id}</span></p>
            </div>
          ))
        )}
      </div>
    </div>
  );
}