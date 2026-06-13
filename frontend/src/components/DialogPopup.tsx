import type {ReactNode} from "react";

export function DialogPopup({
    kicker,
    title,
    onClose,
    children,
    }: {
    kicker: string;
    kickerColor?: string;
    title: string;
    onClose: () => void;
    children: ReactNode;
}) {
    return (
        <div className="fixed inset-0 z-50 grid place-items-center bg-[#1f1b16]/45 px-4 backdrop-blur-sm">
            <section className="w-full max-w-md overflow-hidden rounded-[24px] border-2 border-[#fffaf0]/20 shadow-[0_30px_80px_rgba(31,27,22,0.30),0_8px_22px_rgba(31,27,22,0.12)]">
                <header className="grid grid-cols-[1fr_auto] items-center gap-4 rounded-t-[24px] bg-[#1f1b16] px-6 py-5 text-[#fffaf0]">
                    <div>
                        <div
                            className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[16px] font-extrabold uppercase tracking-[0.30em]"
                            style={{ color: "#f1a45b" }}
                        >
                            {kicker}
                        </div>
                        <h3 className="mt-1 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[22px] font-black leading-none tracking-[-0.04em]">
                            {title}
                        </h3>
                    </div>
                    <button
                        type="button"
                        onClick={onClose}
                        aria-label="Cerrar"
                        className="grid h-9 w-9 cursor-pointer place-items-center rounded-[10px] border border-[#fffaf0]/20 bg-transparent text-[#fffaf0] transition hover:rotate-90 hover:bg-[#fffaf0]/10"
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" className="h-4 w-4">
                            <path d="M6 6l12 12M18 6L6 18" />
                        </svg>
                    </button>
                </header>
                <div className="bg-[#fffaf0] px-6 py-5">{children}</div>
            </section>
        </div>
    );
}