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
    return <div className="flex min-h-screen items-center justify-center bg-[#f4efe2] text-[#1f1b16]">Cargando panel...</div>;
  }

  if (isAuthorized === false) {
    return <Navigate to="/exercises" replace />;
  }

  return (
    <main className="min-h-screen overflow-hidden bg-[#f4efe2] text-[#1f1b16]">
      <section className="relative isolate min-h-screen px-6 py-8 sm:px-10 lg:px-16">
        <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_left,_rgba(234,113,48,0.30),_transparent_34%),radial-gradient(circle_at_bottom_right,_rgba(38,92,82,0.35),_transparent_32%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_44%,_#d8e1d0_100%)]" />
        <div className="absolute left-8 top-10 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/25 blur-sm" />

        <div className="mx-auto max-w-5xl">
          <div className="mb-8 flex items-center justify-between">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">Administración</p>
              <h1 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.04em]">Centro de Mando</h1>
            </div>
            <Link to="/dashboard" className="rounded-full bg-white/50 px-5 py-2 text-sm font-bold backdrop-blur transition hover:bg-white/80">
              Volver a la App
            </Link>
          </div>

          <div className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-xl backdrop-blur-md sm:p-8">
            <div className="mb-8 flex border-b border-[#1f1b16]/10 overflow-x-auto">
              {/* Tab buttons */}
              <button
                className={`px-6 py-3 font-bold transition-colors ${activeTab === "tickets" ? "border-b-2 border-[#ea7130] text-[#ea7130]" : "text-[#5d5348] hover:text-[#1f1b16]"}`}
                onClick={() => setActiveTab("tickets")}
              >
                Soporte y Tickets
              </button>
              <button
                className={`px-6 py-3 font-bold transition-colors ${activeTab === "users" ? "border-b-2 border-[#ea7130] text-[#ea7130]" : "text-[#5d5348] hover:text-[#1f1b16]"}`}
                onClick={() => setActiveTab("users")}
              >
                Gestión de Usuarios
              </button>
            </div>

            {/* Render the active component */}
            {activeTab === "tickets" && <AdminTickets />}
            {activeTab === "users" && <AdminUsers />}

          </div>
        </div>
      </section>
    </main>
  );
}