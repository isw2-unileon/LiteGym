import type { ReactNode } from "react";
import { useIsMobile } from "../lib/useIsMobile";

type MobileOnlyDisclosureProps = {
  kicker?: string;
  title: string;
  children: ReactNode;
  defaultOpen?: boolean;
};

export function MobileOnlyDisclosure({ kicker, title, children, defaultOpen = false }: MobileOnlyDisclosureProps) {
  const isMobile = useIsMobile();

  return (
    <details
      open={!isMobile || defaultOpen}
      className="group w-full max-w-full overflow-hidden rounded-[22px] border border-[#1f1b16]/12 bg-[#fffaf0]/82 shadow-[0_12px_28px_rgba(31,27,22,0.08)] backdrop-blur md:overflow-visible md:rounded-none md:border-0 md:bg-transparent md:shadow-none md:backdrop-blur-0"
    >
      <summary className="flex cursor-pointer list-none items-center justify-between gap-3 px-4 py-4 md:hidden [&::-webkit-details-marker]:hidden">
        <span className="min-w-0 flex-1">
          {kicker && (
            <span className="block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.16em] text-[#265c52]">
              {kicker}
            </span>
          )}
          <span className="mt-1 block truncate [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[21px] font-black leading-none tracking-[-0.04em] text-[#1f1b16]">
            {title}
          </span>
        </span>
        <span className="grid h-9 w-9 shrink-0 place-items-center rounded-[13px] bg-[#1f1b16] text-[#f1a45b]">
          <svg
            viewBox="0 0 24 24"
            aria-hidden="true"
            className="h-4 w-4 transition group-open:rotate-45"
          >
            <path d="M12 5v14M5 12h14" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" />
          </svg>
        </span>
      </summary>
      <div className="min-w-0 max-w-full overflow-hidden border-t border-[#1f1b16]/10 px-3 pb-3 pt-3 md:overflow-visible md:border-0 md:p-0">{children}</div>
    </details>
  );
}
