import type { ExerciseInsightSessionHistory } from "../../types/exercise";

type ExerciseSessionHistoryProps = {
  history: ExerciseInsightSessionHistory[];
};

const dateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "numeric",
  month: "short",
  year: "numeric",
});

const numberFormatter = new Intl.NumberFormat("es-ES", {
  maximumFractionDigits: 1,
});

export default function ExerciseSessionHistory({
  history,
}: ExerciseSessionHistoryProps) {
  if (history.length === 0) {
    return (
      <p className="rounded-[1.25rem] border border-dashed border-[#1f1b16]/15 bg-white/75 p-4 text-sm font-semibold leading-6 text-[#5d5348]">
        No hay series registradas para este ejercicio.
      </p>
    );
  }

  return (
    <div className="grid gap-3">
      {history.slice(0, 6).map((session) => {
        const setCount = session.sets.length;
        const volume = numberFormatter.format(session.volume_kg);
        return (
          <details
            className="group rounded-[1.25rem] border border-[#1f1b16]/10 bg-white/80 p-4 shadow-[0_8px_18px_rgba(31,27,22,0.04)]"
            key={session.session_id}
          >
            <summary className="cursor-pointer list-none">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="break-words text-sm font-black tracking-[-0.02em] text-[#1f1b16]">
                    {session.session_name || session.routine_name || "Sesion"}
                  </p>
                  <p className="mt-1 text-xs font-semibold uppercase tracking-[0.14em] text-[#5d5348]">
                    {dateFormatter.format(new Date(session.performed_at))}
                  </p>
                </div>

                <div className="flex shrink-0 flex-col items-end gap-2">
                  <span className="rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black uppercase tracking-[0.14em] text-[#265c52]">
                    {volume} kg
                  </span>
                  <span className="rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-1 text-xs font-black uppercase tracking-[0.14em] text-[#3a332c]">
                    {setCount} {setCount === 1 ? "serie" : "series"}
                  </span>
                </div>
              </div>

              <div className="mt-3 flex flex-wrap gap-2">
                <InlineStat label="Duracion" value={`${session.duration_minutes} min`} />
                <InlineStat label="Volumen" value={`${volume} kg`} />
                <InlineStat label="Series" value={String(setCount)} />
              </div>
            </summary>

            <div className="mt-4 grid gap-2">
              {session.sets.map((set) => (
                <div
                  className="grid grid-cols-[auto_1fr_auto] items-center gap-3 rounded-[1rem] border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-3 text-sm"
                  key={`${session.session_id}-${set.set_number}`}
                >
                  <span className="inline-flex h-8 w-8 items-center justify-center rounded-[0.8rem] bg-[#1f1b16] text-xs font-black text-[#f1a45b]">
                    {set.set_number}
                  </span>
                  <span className="font-semibold text-[#5d5348]">
                    {formatSet(set.weight_kg, set.reps, set.volume_kg)}
                  </span>
                  <span className="rounded-full border border-[#1f1b16]/10 px-2.5 py-1 text-xs font-black uppercase tracking-[0.12em] text-[#3a332c]">
                    {typeof set.rir === "number" ? `RIR ${set.rir}` : "Sin RIR"}
                  </span>
                </div>
              ))}
            </div>
          </details>
        );
      })}
    </div>
  );
}

function InlineStat({ label, value }: { label: string; value: string }) {
  return (
    <span className="inline-flex items-center gap-2 rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-1 text-xs font-black uppercase tracking-[0.12em] text-[#3a332c]">
      <span className="text-[#265c52]">{label}</span>
      <span className="normal-case tracking-[-0.02em]">{value}</span>
    </span>
  );
}

function formatSet(
  weightKg: number | null | undefined,
  reps: number | null | undefined,
  volumeKg: number,
) {
  const weight = typeof weightKg === "number" ? `${numberFormatter.format(weightKg)} kg` : "sin peso";
  const repetitions = typeof reps === "number" ? `${reps} reps` : "sin reps";
  return `${weight} x ${repetitions} · ${numberFormatter.format(volumeKg)} kg`;
}
