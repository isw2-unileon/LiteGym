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
        <div className="fixed inset-0 z-50 flex items-end bg-[#1f1b16]/45 px-0 pb-0 pt-10 backdrop-blur-sm sm:grid sm:place-items-center sm:px-4 sm:py-4">
            <section className="flex max-h-[min(88dvh,42rem)] w-full max-w-none flex-col overflow-hidden rounded-t-[24px] border-2 border-[#fffaf0]/20 shadow-[0_30px_80px_rgba(31,27,22,0.30),0_8px_22px_rgba(31,27,22,0.12)] sm:max-h-[calc(100dvh-2rem)] sm:max-w-md sm:rounded-[24px]">
                <header className="grid grid-cols-[1fr_auto] items-center gap-4 bg-[#1f1b16] px-4 py-4 text-[#fffaf0] sm:rounded-t-[24px] sm:px-6 sm:py-5">
                    <div>
                        <div
                            className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-extrabold uppercase tracking-[0.16em] sm:text-[16px] sm:tracking-[0.30em]"
                            style={{ color: "#f1a45b" }}
                        >
                            {kicker}
                        </div>
                        <h3 className="mt-1 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[20px] font-black leading-tight sm:text-[22px]">
                            {title}
                        </h3>
                    </div>
                    <button
                        type="button"
                        onClick={onClose}
                        aria-label="Cerrar"
                        className="grid h-11 w-11 cursor-pointer place-items-center rounded-[14px] border border-[#fffaf0]/20 bg-[#fffaf0]/8 text-[#fffaf0] transition hover:rotate-90 hover:bg-[#fffaf0]/10 sm:h-9 sm:w-9 sm:rounded-[10px] sm:bg-transparent"
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" className="h-4 w-4">
                            <path d="M6 6l12 12M18 6L6 18" />
                        </svg>
                    </button>
                </header>
                <div className="min-h-0 overflow-y-auto bg-[#fffaf0] px-4 pb-[calc(1.25rem+env(safe-area-inset-bottom))] pt-4 sm:px-6 sm:py-5">{children}</div>
            </section>
        </div>
    );
}
