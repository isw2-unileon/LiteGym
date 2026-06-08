export function HelloHeader({ page, user }: {
    page: string;
    user: string;
}
) {
  return (
      <div>
          <div className="mb-2.5 flex items-center gap-2.5 text-[14px] font-extrabold uppercase tracking-[0.30em] text-[#265c52]">
              <span className="inline-block h-0.5 w-6 bg-[#265c52]" />
              {page}
          </div>
            <h1 className="m-0 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[44px] font-black leading-[0.92] tracking-[-0.055em] text-[#1f1b16] sm:text-[64px]">
                Hola,{" "}
                <span className="px-1 text-[#ea7130] capitalize [background:linear-gradient(180deg,transparent_60%,rgba(234,113,48,0.18)_60%)]">
                    {user}
                </span>
            </h1>
      </div>
  );
}