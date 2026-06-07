export function Stat({ n, l, accent }: {
    n: string;
    l: string;
    accent?: boolean
}) {
    return (
        <div className={["min-w-0 rounded-[18px] border px-3 py-2.5 backdrop-blur sm:min-w-[120px] sm:px-4 sm:py-3", accent ? "border-[#1f1b16] bg-[#1f1b16] text-[#fffaf0]" : "border-[#1f1b16]/12 bg-[#fffaf0]/75"].join(" ")}>
            <div className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] capitalize text-[26px] font-black leading-none tracking-[-0.04em] sm:text-[32px]">{n}</div>
            <div className={["mt-1.5 break-words [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[9px] font-bold uppercase leading-[1.25] tracking-[0.08em] sm:text-[12px] sm:tracking-[0.16em]", accent ? "text-[#f1a45b]" : "text-[#3a332c]"].join(" ")}>{l}</div>
        </div>
    );
}
