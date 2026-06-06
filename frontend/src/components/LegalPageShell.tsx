import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import LegalLinks from "./LegalLinks";

type LegalPageShellProps = {
  title: string;
  subtitle: string;
  updatedAt: string;
  children: ReactNode;
};

export default function LegalPageShell({ title, subtitle, updatedAt, children }: LegalPageShellProps) {
  return (
    <main className="min-h-screen bg-[#f4efe2] px-6 py-10 text-[#1f1b16] sm:px-10 lg:px-16">
      <section className="mx-auto max-w-5xl rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/85 p-6 shadow-[0_30px_80px_rgba(47,39,27,0.18)] backdrop-blur-md sm:p-10">
        <div className="flex flex-col gap-6 border-b border-[#1f1b16]/10 pb-6 sm:flex-row sm:items-start sm:justify-between">
          <div className="max-w-3xl">
            <p className="text-xs font-semibold uppercase tracking-[0.28em] text-[#265c52]">
              LiteGym legal
            </p>
            <h1 className="mt-3 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.05em] sm:text-5xl">
              {title}
            </h1>
            <p className="mt-4 max-w-3xl text-base leading-7 text-[#4f453c]">
              {subtitle}
            </p>
          </div>

          <div className="rounded-2xl border border-[#1f1b16]/10 bg-[#1f1b16] px-4 py-3 text-sm font-semibold text-[#fffaf0] shadow-inner">
            <p className="text-xs uppercase tracking-[0.24em] text-[#f1a45b]">Actualizado</p>
            <p className="mt-1">{updatedAt}</p>
          </div>
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
      </section>
    </main>
  );
}
