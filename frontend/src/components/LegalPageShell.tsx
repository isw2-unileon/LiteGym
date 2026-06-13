import  { ReactNode } from "react";
import { Link } from "react-router-dom";
import LegalLinks from "./LegalLinks";
import {Card, CardHeader} from "@/components/Card.tsx";
import {Stat} from "@/components/Stat.tsx";

type LegalPageShellProps = {
  title: string;
  subtitle: string;
  updatedAt: string;
  children: ReactNode;
};

export default function LegalPageShell({ title, subtitle, updatedAt, children }: LegalPageShellProps) {
  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top_right,_rgba(234,113,48,0.24),_transparent_30%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_52%,_#d8e1d0_100%)] px-6 py-10 text-[#1f1b16] sm:px-10 lg:px-16">
      <div className="px-6  pt-8 sm:px-8">
        <section className="mx-auto max-w-[1280px] gap-[18px] xl:grid-cols-[minmax(0,1.55fr)_minmax(20rem,0.95fr)]">
          <Card accent="#ea7130">
            <CardHeader kicker={"LITEGYM LEGAL"} title={title} />
            <div className="flex flex-col gap-6 border-b border-[#1f1b16]/10 pb-6 sm:flex-row sm:items-start sm:justify-between">
              <div className="max-w-3xl">
                <p className="mt-4 max-w-3xl text-base leading-7 text-[#4f453c]">
                  {subtitle}
                </p>
              </div>
              <Stat n={"Actualizado"} l={updatedAt} accent={true}/>
            </div>
            <article className="mt-8 max-w-none space-y-4 text-[0.98rem] leading-7 text-[#3d342c] [&_h2]:mt-10 [&_h2]:font-['Aptos_Display','Trebuchet_MS',sans-serif] [&_h2]:text-2xl [&_h2]:font-black [&_h2]:tracking-[-0.04em] [&_h2]:text-[#1f1b16] [&_h3]:mt-8 [&_h3]:font-['Aptos_Display','Trebuchet_MS',sans-serif] [&_h3]:text-xl [&_h3]:font-black [&_h3]:tracking-[-0.03em] [&_h3]:text-[#1f1b16] [&_ul]:list-disc [&_ul]:space-y-2 [&_ul]:pl-6 [&_a]:font-bold [&_a]:text-[#ea7130] [&_a]:no-underline hover:[&_a]:underline">
              {children}
            </article>
            <div className="mt-10 flex flex-col gap-4 border-t border-[#1f1b16]/10 pt-6 sm:flex-row sm:items-center sm:justify-between">
              <Link
                  to="/"
                  className="inline-flex items-center justify-center rounded-2xl border border-[#1f1b16]/15 px-4 py-3 text-sm font-bold text-[#1f1b16] transition hover:border-[#ea7130] hover:text-[#ea7130]"
              >
                Volver al inicio
              </Link>

              <LegalLinks compact />
            </div>
          </Card>
        </section>
      </div>
    </main>
  );
}
