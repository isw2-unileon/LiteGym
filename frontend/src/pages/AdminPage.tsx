import { Navigate, useOutletContext } from "react-router-dom";
import type { LayoutUser } from "../components/AppLayout";

type OutletContext = {
  user?: LayoutUser | null;
};

export default function AdminPage() {
  const { user } = useOutletContext<OutletContext>();

  if (user?.role !== "admin") {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <section className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/85 p-6 shadow-[0_24px_60px_rgba(47,39,27,0.14)]">
      <h1 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.05em]">
        Panel admin
      </h1>
      <p className="mt-4 text-sm font-semibold leading-6 text-[#5d5348]">Por hacer</p>
    </section>
  );
}
