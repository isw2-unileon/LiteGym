import * as React from "react";
import { Navigate, useOutletContext } from "react-router-dom";
import type { LayoutUser } from "../components/AppLayout";

// Import sub-components
import AdminUsers from "../components/admin/AdminUsers";
import AdminTickets from "../components/admin/AdminTickets";
import {HelloHeader} from "@/components/HelloHeader.tsx";
import {Card, CardHeader} from "@/components/Card.tsx";
import {useIsMobile} from "@/lib/useIsMobile.ts";

type OutletContext = { user?: LayoutUser | null };

export default function AdminPage() {
  const { user } = useOutletContext<OutletContext>();
  const [isAuthorized, setIsAuthorized] = React.useState<boolean | null>(null);
  const [activeTab, setActiveTab] = React.useState<"users" | "exercises" | "tickets">("tickets");
  const isMobile = useIsMobile();

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

  if (!isAuthorized) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <main className="relative isolate text-[#1f1b16] [font-family:'Inter',system-ui,sans-serif] antialiased pb-20">
      {/* Background decoration blobs */}
      <div aria-hidden="true" className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]" />
      <div aria-hidden="true" className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10" />
      <div className="px-4 pt-5 sm:px-6 sm:pt-8 md:px-8">
        <section className="mx-auto mb-6 max-w-[1280px] items-start gap-6 md:grid-cols-[1fr_auto]">
          <HelloHeader page={"PANEL DE ADMINISTRACIÓN"} user={user?.username ?? "Atleta"} />
        </section>
        <section className="mx-auto max-w-[1280px] gap-[18px] xl:grid-cols-[minmax(0,1.55fr)_minmax(20rem,0.95fr)]">
          {isMobile ? (
            <MobileAdminPanel activeTab={activeTab} onTabChange={setActiveTab} />
          ) : (
          <Card accent="#ea7130">
            <CardHeader kicker={"ADMINISTRACIÓN"} title={"CENTRO DE MANDO"}/>
            <div className="mb-8 mt-4 inline-flex rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0]/60 p-1.5 backdrop-blur-sm overflow-x-auto max-w-full">
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
            <div className="relative z-10 overflow-y-auto">
              {activeTab === "tickets" && <AdminTickets />}
              {activeTab === "users" && <AdminUsers />}
            </div>
          </Card>
          )}
        </section>
      </div>
    </main>
  );
}

function MobileAdminPanel({
  activeTab,
  onTabChange,
}: {
  activeTab: "users" | "exercises" | "tickets";
  onTabChange: React.Dispatch<React.SetStateAction<"users" | "exercises" | "tickets">>;
}) {
  return (
    <section className="grid gap-4">
      <div className="overflow-hidden rounded-[30px] border border-[#1f1b16]/10 bg-[#fffaf0]/88 shadow-[0_14px_34px_rgba(31,27,22,0.1)]">
        <div className="bg-[#ea7130] px-5 py-4">
          <p className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.18em] text-[#1f1b16]/70">
            Admin
          </p>
          <h2 className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[31px] font-black leading-none tracking-[-0.055em] text-[#1f1b16]">
            Centro de mando
          </h2>
        </div>
        <div className="grid grid-cols-2 gap-2 p-3">
          <button
            type="button"
            onClick={() => onTabChange("tickets")}
            className={[
              "rounded-[18px] px-4 py-4 text-left text-sm font-black",
              activeTab === "tickets" ? "bg-[#1f1b16] text-[#f1a45b]" : "bg-white/64 text-[#1f1b16]",
            ].join(" ")}
          >
            Tickets
          </button>
          <button
            type="button"
            onClick={() => onTabChange("users")}
            className={[
              "rounded-[18px] px-4 py-4 text-left text-sm font-black",
              activeTab === "users" ? "bg-[#1f1b16] text-[#f1a45b]" : "bg-white/64 text-[#1f1b16]",
            ].join(" ")}
          >
            Usuarios
          </button>
        </div>
      </div>
      <div className="min-w-0 overflow-hidden rounded-[26px] border border-[#1f1b16]/10 bg-[#fffaf0]/84 p-3 shadow-[0_12px_28px_rgba(31,27,22,0.08)]">
        {activeTab === "tickets" && <AdminTickets />}
        {activeTab === "users" && <AdminUsers />}
      </div>
    </section>
  );
}
