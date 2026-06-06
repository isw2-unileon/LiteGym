import type {CSSProperties, ReactNode} from "react";

export function Card({children, accent = "#ea7130", dark, className}: {
    children: ReactNode;
    accent?: string;
    dark?: boolean;
    className?: string;
}) {
    return (
        <article
            style={{ ["--accent" as never]: accent } as CSSProperties}
            className={[
                "group relative isolate overflow-hidden rounded-[24px] border p-5 backdrop-blur sm:p-6",
                dark
                    ? "border-[#1f1b16] bg-[#1f1b16] text-[#fffaf0]"
                    : "border-[#1f1b16]/12 bg-[#fffaf0]/85",
                className,
            ].filter(Boolean).join(" ")}
        >
            {!dark && (
                <span
                    aria-hidden="true"
                    className="absolute inset-x-0 top-0 z-[1] h-1.5"
                    style={{ background: "var(--accent)" }}
                />
            )}
            <span
                aria-hidden="true"
                className="pointer-events-none absolute -right-10 -top-10 z-0 h-40 w-40 rounded-full opacity-10 blur-[30px]"
                style={{ background: "var(--accent)" }}
            />
            {children}
        </article>
    );
}

export function CardHeader({
                        kicker,
                        title,
                        right,
                        rightChip,
                        onDark,
                    }: {
    kicker: string;
    title: ReactNode;
    right?: ReactNode;
    rightChip?: string;
    onDark?: boolean;
}) {
    return (
        <header
            className={[
                "relative z-[2] flex flex-wrap justify-between gap-3",
                title ? "items-end" : "items-center",
            ].join(" ")}
        >
            <div>
                <div
                    className={[
                        "flex items-center gap-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.22em]",
                        title ? "mb-1.5" : "",
                        onDark ? "text-[#f1a45b]" : "text-[#265c52]",
                    ].join(" ")}
                >
          <span
              className="inline-block h-0.5 w-4"
              style={{ background: onDark ? "#f1a45b" : "#265c52" }}
          />
                    {kicker}
                </div>
                {title && (
                    <h3
                        className={[
                            "m-0 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[24px] font-black leading-[1.05] tracking-[-0.035em]",
                            onDark ? "text-[#fffaf0]" : "text-[#1f1b16]",
                        ].join(" ")}
                    >
                        {title}
                    </h3>
                )}
                {rightChip && (
                    <div
                        className={[
                            "mt-1.5 inline-flex max-w-full items-center gap-1.5 rounded-md border px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.14em]",
                            onDark
                                ? "border-[#fffaf0]/15 bg-[#fffaf0]/8 text-[#fffaf0]/80"
                                : "border-[#1f1b16]/10 bg-[#1f1b16]/5 text-[#3a332c]",
                        ].join(" ")}
                    >
                        <span className="h-1.5 w-1.5 rounded-full" style={{ background: "var(--accent)" }} />
                        <span className="truncate">{rightChip}</span>
                    </div>
                )}
            </div>
            {right}
        </header>
    );
}