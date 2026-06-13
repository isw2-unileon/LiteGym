import  { ReactNode } from "react";
import { Link } from "react-router-dom";
import LegalLinks from "./LegalLinks";
import {Card, CardHeader} from "@/components/Card.tsx";
import {Stat} from "@/components/Stat.tsx";
import {useIsMobile} from "@/lib/useIsMobile.ts";

type LegalPageShellProps = {
  title: string;
  subtitle: string;
  updatedAt: string;
  children: ReactNode;
};

export default function LegalPageShell({ title, subtitle, updatedAt, children }: LegalPageShellProps) {
  const isMobile = useIsMobile();

  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top_right,_rgba(234,113,48,0.24),_transparent_30%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_52%,_#d8e1d0_100%)] px-6 py-10 text-[#1f1b16] sm:px-10 lg:px-16">
      {isMobile ? (
        <section className="mx-auto max-w-xl">
          <Link
            to="/"
            className="mb-5 inline-flex rounded-[14px] bg-[#1f1b16] px-4 py-3 text-xs font-black uppercase tracking-[0.08em] text-[#f1a45b]"
          >
            Volver
          </Link>
          <article className="overflow-hidden rounded-[30px] border border-[#1f1b16]/10 bg-[#fffaf0]/88 shadow-[0_14px_34px_rgba(31,27,22,0.1)]">
            <header className="bg-[#ea7130] px-5 py-5">
              <p className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.18em] text-[#1f1b16]/70">
                LiteGym legal
              </p>
              <h1 className="mt-2 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[34px] font-black leading-none tracking-[-0.06em] text-[#1f1b16]">
                {title}
              </h1>
              <p className="mt-4 text-sm font-bold leading-relaxed text-[#1f1b16]/72">{subtitle}</p>
              <span className="mt-4 inline-flex rounded-[14px] bg-[#1f1b16] px-3 py-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.12em] text-[#f1a45b]">
                {updatedAt}
              </span>
            </header>
            <div className="space-y-4 px-5 py-5 text-[0.95rem] font-medium leading-7 text-[#3d342c] [&_h2]:mt-8 [&_h2]:font-['Bricolage_Grotesque','Aptos_Display',sans-serif] [&_h2]:text-[24px] [&_h2]:font-black [&_h2]:leading-none [&_h2]:tracking-[-0.045em] [&_h2]:text-[#1f1b16] [&_ul]:list-disc [&_ul]:space-y-2 [&_ul]:pl-5 [&_a]:font-black [&_a]:text-[#ea7130]">
              {children}
            </div>
            <div className="border-t border-[#1f1b16]/10 px-5 py-5">
              <LegalLinks compact />
            </div>
          </article>
        </section>
      ) : (
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
      )}
    </main>
  );
}
