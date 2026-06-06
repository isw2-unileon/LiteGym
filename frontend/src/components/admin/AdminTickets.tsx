import * as React from "react";
import { apiUrl } from "../../lib/api";
import {Card, CardHeader} from "@/components/Card.tsx";

type Ticket = {
  id: string;
  user_id: string;
  title: string;
  description: string;
  status: string;
  created_at: string;
};

type TicketListProps = {
  items: Ticket[];
  hasAnyTickets: boolean;
  expandedTickets: Record<string, boolean>;
  onToggle: (id: string) => void;
  onClose: (id: string) => void;
};

function TicketList({ items, hasAnyTickets, expandedTickets, onToggle, onClose }: TicketListProps) {
  return (
    <div className="flex flex-col gap-4 flex-1 min-h-0 mt-4 overflow-y-auto [scrollbar-gutter:stable] pr-1.5 scroll-smooth">
      {!hasAnyTickets ? (
        <p className="text-sm font-semibold text-[#5d5348] text-center py-10 bg-white/40 rounded-2xl border border-dashed border-[#1f1b16]/15">
          No hay tickets en el sistema.
        </p>
      ) : items.length === 0 ? (
        <p className="text-sm font-semibold text-[#5d5348] text-center py-10 bg-white/40 rounded-2xl border border-dashed border-[#1f1b16]/15">
          No hay tickets que coincidan con este estado.
        </p>
      ) : (
        items.map((ticket) => {
          const isClosed = ticket.status === "closed";
          const isExpanded = !!expandedTickets[ticket.id];

          // Extract category for the badge
          let category = "General";
          if (ticket.title.startsWith("[")) {
            const closingBracket = ticket.title.indexOf("]");
            if (closingBracket !== -1) {
              category = ticket.title.substring(1, closingBracket);
            }
          }

          // Decide styling based on category
          let categoryStyle = "bg-[#265c52]/10 text-[#265c52] border-[#265c52]/20";
          if (category === "Ejercicios") {
            categoryStyle = "bg-[#6b4ea0]/12 text-[#6b4ea0] border-[#6b4ea0]/20";
          } else if (category === "IA" || category === "Asistente IA") {
            categoryStyle = "bg-[#ea7130]/12 text-[#ea7130] border-[#ea7130]/20";
          }

          return (
            <article
              key={ticket.id}
              onClick={() => onToggle(ticket.id)}
              className={`shrink-0 relative overflow-hidden rounded-[20px] border p-4.5 transition cursor-pointer select-none ${
                isClosed
                  ? "border-[#265c52]/18 bg-[#265c52]/3 opacity-75 shadow-none"
                  : isExpanded
                    ? "border-[#ea7130]/30 bg-white/90 shadow-[0_12px_28px_rgba(234,113,48,0.06)]"
                    : "border-[#1f1b16]/12 bg-white/80 shadow-[0_8px_20px_rgba(31,27,22,0.02)] hover:shadow-[0_12px_28px_rgba(31,27,22,0.05)] hover:border-[#1f1b16]/18"
              }`}
            >
              {/* Visual side accent bar */}
              {isClosed ? (
                <span aria-hidden="true" className="absolute inset-y-0 left-0 w-1 bg-[#265c52]/60" />
              ) : (
                <span aria-hidden="true" className="absolute inset-y-0 left-0 w-1 bg-[#ea7130]" />
              )}

              <div className="flex flex-col gap-2 h-[150px]">
                {/* Badges row */}
                <div className="flex flex-wrap items-center gap-2">
                  <span className={`rounded-md border px-2 py-0.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-extrabold uppercase tracking-wider ${
                    isClosed ? "bg-[#265c52]/18 text-[#265c52] border-[#265c52]/30" : "bg-[#ea7130]/18 text-[#ea7130] border-[#ea7130]/30"
                  }`}>
                    {ticket.status}
                  </span>
                  <span className={`rounded-md border px-2 py-0.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-extrabold uppercase tracking-wider ${categoryStyle}`}>
                    {category}
                  </span>
                  <svg
                      className={`ml-auto h-5 w-5 text-[#3a332c]/50 transition-transform duration-300 shrink-0 ${
                          isExpanded ? 'rotate-180' : ''
                      }`}
                      viewBox="0 0 20 20"
                      fill="currentColor"
                  >
                    <path fillRule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clipRule="evenodd" />
                  </svg>
                </div>

                {/* Date below badges */}
                <span className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold text-[#3a332c]/50">
                  {new Date(ticket.created_at).toLocaleDateString()}
                </span>

                {/* Full title */}
                <h3 className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[17px] font-black tracking-tight text-[#1f1b16] break-words line-clamp-2">
                  {ticket.title}
                </h3>

                {/* Actions pinned to bottom */}
                <div className="mt-auto flex items-center gap-3">
                  {!isClosed && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation(); // Avoid expanding card when clicking close button
                        onClose(ticket.id);
                      }}
                      className="ml-auto inline-flex cursor-pointer items-center gap-1.5 rounded-[10px] border-0 bg-[#1f1b16] px-3.5 py-2 text-[14px] font-extrabold leading-none tracking-[0.03em] text-[#f1a45b] no-underline shadow-[0_10px_22px_rgba(31,27,22,0.18)] transition hover:-translate-y-px hover:bg-[#2c261f]"
                    >
                      Resolver ticket
                    </button>
                  )}

                  {/* Collapsible status chevron icon */}

                </div>
              </div>

              {/* Collapsible Detailed Content with standard max-height slide */}
              <div
                className={`transition-all duration-300 ease-in-out overflow-hidden ${
                  isExpanded
                    ? "max-h-[500px] opacity-100"
                    : "max-h-0 opacity-0"
                }`}
              >
                <div className="mt-4 border-t border-[#1f1b16]/8 pt-4">
                  <div className="rounded-[14px] bg-black/4 p-4 border border-[#1f1b16]/5 max-h-36 overflow-y-auto">
                    <p className="text-[13px] font-semibold text-[#3b352f] leading-relaxed whitespace-pre-wrap">
                      {ticket.description}
                    </p>
                  </div>
                  <footer className="mt-3.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold text-[#3a332c]/40 uppercase flex flex-col gap-1">
                    <span>ID Usuario: <span className="font-mono block break-all normal-case">{ticket.user_id}</span></span>
                    <span>Creado: {new Date(ticket.created_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</span>
                  </footer>
                </div>
              </div>
            </article>
          );
        })
      )}
    </div>
  );
}

export default function AdminTickets() {
  const [tickets, setTickets] = React.useState<Ticket[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  const [statusFilter, setStatusFilter] = React.useState<"all" | "open" | "closed">("open"); 
  const [expandedTickets, setExpandedTickets] = React.useState<Record<string, boolean>>({});

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

  const toggleTicketExpanded = (id: string) => {
    setExpandedTickets(prev => ({
      ...prev,
      [id]: !prev[id]
    }));
  };

  const filteredGeneralTickets = tickets.filter(ticket => {
    const matchesStatus = statusFilter === "all" || ticket.status === statusFilter;
    const matchesCategory = ticket.title.includes(`[General]`);
    return matchesStatus && matchesCategory
  });

  const filteredIATickets = tickets.filter(ticket => {
    const matchesStatus = statusFilter === "all" || ticket.status === statusFilter;
    const matchesCategory = ticket.title.includes(`[IA]`) || ticket.title.includes(`[Asistente IA]`);
    return matchesStatus && matchesCategory
  });

  const filteredExercisesTickets = tickets.filter(ticket => {
    const matchesStatus = statusFilter === "all" || ticket.status === statusFilter;
    const matchesCategory = ticket.title.includes(`[Ejercicios]`);
    return matchesStatus && matchesCategory
  });


  if (loading) {
    return (
      <div className="p-12 text-center font-bold text-[#5d5348] animate-pulse">
        Cargando tickets...
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-12 text-center font-bold text-[#c94b32]">
        {error}
      </div>
    );
  }

  return (
    <div className="flex-col gap-6 h-[46vh]">
      <div className="flex flex-col min-w-[160px] overflow-y-auto [scrollbar-gutter:stable]">
        <label className="mb-1.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[14px] font-bold uppercase tracking-wider text-[#3a332c]/85">
          Estado
        </label>
        <select
            className="rounded-[12px] border border-[#1f1b16]/12 bg-white/90 px-3.5 py-2 text-sm font-semibold outline-none focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/12 cursor-pointer transition"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as "all" | "open" | "closed")}
        >
          <option value="open">Pendientes (Abiertos)</option>
          <option value="closed">Resueltos (Cerrados)</option>
          <option value="all">Mostrar Todos</option>
        </select>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-1 lg:grid-cols-3 gap-6 mt-4 overflow-y-auto [scrollbar-gutter:stable]">
      <Card accent="#ea7130" className="relative z-[2] mb-5 h-[38vh] flex flex-col">
        <CardHeader kicker={"GENERAL"} title={""} />
        <TicketList
          items={filteredGeneralTickets}
          hasAnyTickets={tickets.length > 0}
          expandedTickets={expandedTickets}
          onToggle={toggleTicketExpanded}
          onClose={handleCloseTicket}
        />
      </Card>
      <Card accent="#ea7130" className="relative z-[2] mb-5 h-[38vh] flex flex-col">
        <CardHeader kicker={"Asistente IA"} title={""} />
        <TicketList
            items={filteredIATickets}
            hasAnyTickets={tickets.length > 0}
            expandedTickets={expandedTickets}
            onToggle={toggleTicketExpanded}
            onClose={handleCloseTicket}
        />
      </Card>
      <Card accent="#ea7130" className="relative z-[2] mb-5 h-[38vh] flex flex-col">
        <CardHeader kicker={"Ejercicios"} title={""} />
        <TicketList
            items={filteredExercisesTickets}
            hasAnyTickets={tickets.length > 0}
            expandedTickets={expandedTickets}
            onToggle={toggleTicketExpanded}
            onClose={handleCloseTicket}
        />
      </Card>
      </div>
    </div>
  );
}

