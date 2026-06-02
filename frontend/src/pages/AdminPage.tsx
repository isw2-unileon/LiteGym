import * as React from "react";
import { Link, Navigate, useOutletContext } from "react-router-dom";
import type { LayoutUser } from "../components/AppLayout";

// Import sub-components
import AdminUsers from "../components/admin/AdminUsers";
import AdminTickets from "../components/admin/AdminTickets";

type OutletContext = { user?: LayoutUser | null };

export default function AdminPage() {
  const { user } = useOutletContext<OutletContext>();
  const [isAuthorized, setIsAuthorized] = React.useState<boolean | null>(null);
  const [activeTab, setActiveTab] = React.useState<"users" | "exercises" | "tickets">("tickets");

  React.useEffect(() => {
    // Admin verification logic
    if (user?.role !== "admin") {
      setIsAuthorized(false);
    } else {
      setIsAuthorized(true);
    }
  }, [user]);

  if (isAuthorized === null) {
    return (
      <div className="flex min-h-screen items-center justify-center font-bold text-[#5d5348] animate-pulse">
        Cargando panel...
      </div>
    );
  }

  if (isAuthorized === false) {
    return <Navigate to="/exercises" replace />;
  }

  return (
    <div className="relative isolate text-[#1f1b16] [font-family:'Inter',system-ui,sans-serif] antialiased pb-20">
      {/* Background decoration blobs */}
      <div aria-hidden="true" className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]" />
      <div aria-hidden="true" className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10" />

      {/* Header section */}
      <header className="mb-8 flex flex-col sm:flex-row sm:items-end justify-between gap-6">
        <div>
          <div className="mb-2.5 flex items-center gap-2.5 text-[12px] font-extrabold uppercase tracking-[0.30em] text-[#265c52]">
            <span className="inline-block h-0.5 w-6 bg-[#265c52]" /> Panel de Administración
          </div>
          <h1 className="m-0 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[40px] font-black leading-[0.95] tracking-[-0.05em] text-[#1f1b16] sm:text-[54px]">
            Centro de <span className="px-1 text-[#ea7130] [background:linear-gradient(180deg,transparent_60%,rgba(234,113,48,0.18)_60%)]">Mando</span>
          </h1>
        </div>
        <Link 
          to="/dashboard" 
          className="rounded-[14px] border border-[#1f1b16]/15 bg-[#fffaf0]/80 px-5 py-3 text-[13px] font-bold backdrop-blur transition hover:bg-white hover:border-[#ea7130] hover:text-[#ea7130] shadow-sm text-center"
        >
          Volver a la App
        </Link>
      </header>

      {/* Main glassmorphism card container */}
      <div className="rounded-[2.2rem] border border-[#1f1b16]/12 bg-[#fffaf0]/85 p-5 shadow-[0_30px_80px_rgba(47,39,27,0.15)] backdrop-blur-md sm:p-8">
        
        {/* Navigation Tabs Pillbox */}
        <div className="mb-8 inline-flex rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0]/60 p-1.5 backdrop-blur-sm overflow-x-auto max-w-full">
          <button
            className={`rounded-[10px] px-5 py-2.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-[0.14em] transition whitespace-nowrap ${
              activeTab === "tickets" 
                ? "bg-[#ea7130] text-[#1f1b16] shadow-sm" 
                : "bg-transparent text-[#3a332c] hover:bg-[#1f1b16]/5"
            }`}
            onClick={() => setActiveTab("tickets")}
          >
            Soporte y Tickets
          </button>
          <button
            className={`rounded-[10px] px-5 py-2.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-[0.14em] transition whitespace-nowrap ${
              activeTab === "users" 
                ? "bg-[#ea7130] text-[#1f1b16] shadow-sm" 
                : "bg-transparent text-[#3a332c] hover:bg-[#1f1b16]/5"
            }`}
            onClick={() => setActiveTab("users")}
          >
            Gestión de Usuarios
          </button>
        </div>

        {/* Content Render Area */}
        <div className="relative z-10">
          {activeTab === "tickets" && <AdminTickets />}
          {activeTab === "users" && <AdminUsers />}
        </div>

      </div>
    </div>
  );
}