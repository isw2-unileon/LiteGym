export function Stat({ n, l, accent }: {
    n: string;
    l: string;
    accent?: boolean
}) {
    return (
        <div className={["min-w-[120px] rounded-[18px] border px-4 py-3 backdrop-blur", accent ? "border-[#1f1b16] bg-[#1f1b16] text-[#fffaf0]" : "border-[#1f1b16]/12 bg-[#fffaf0]/75"].join(" ")}>
            <div className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[32px] font-black leading-none tracking-[-0.04em]">{n}</div>
            <div className={["mt-1.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.16em]", accent ? "text-[#f1a45b]" : "text-[#3a332c]"].join(" ")}>{l}</div>
        </div>
    );
}
