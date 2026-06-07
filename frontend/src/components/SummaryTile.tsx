import type { ReactNode } from "react";

export function SummaryTile({
  value,
  label,
  dark,
}: {
  value: ReactNode;
  label: ReactNode;
  dark?: boolean;
}) {
  return (
    <div
      className={[
        "rounded-[14px] border p-3.5",
        dark
          ? "border-[#fffaf0]/10 bg-[#fffaf0]/6"
          : "border-[#1f1b16]/10 bg-[#fffaf0]/75",
      ].join(" ")}
    >
      <div
        className={[
          "[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[28px] font-black leading-none tracking-[-0.04em]",
          dark ? "text-[#fffaf0]" : "text-[#1f1b16]",
        ].join(" ")}
      >
        {value}
      </div>
      <div className={["mt-1 text-sm font-semibold", dark ? "text-[#fffaf0]" : "text-[#3a332c]"].join(" ")}>
        {label}
      </div>
    </div>
  );
}
