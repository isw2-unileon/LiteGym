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
      <p className="rounded-2xl border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/80 p-4 text-sm font-semibold text-[#5d5348]">
        No hay series registradas para este ejercicio.
      </p>
    );
  }

  return (
    <div className="grid gap-3">
      {history.slice(0, 5).map((session) => (
        <details
          className="rounded-2xl border border-[#1f1b16]/10 bg-white/70 p-4"
          key={session.session_id}
        >
          <summary className="cursor-pointer list-none">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="break-words text-sm font-black text-[#1f1b16]">
                  {session.session_name || session.routine_name || "Sesion"}
                </p>
                <p className="mt-1 text-xs font-semibold uppercase tracking-[0.14em] text-[#5d5348]">
                  {dateFormatter.format(new Date(session.performed_at))}
                </p>
              </div>
              <p className="shrink-0 text-xs font-black text-[#265c52]">
                {numberFormatter.format(session.volume_kg)} kg
              </p>
            </div>
          </summary>

          <div className="mt-3 grid gap-2">
            {session.sets.map((set) => (
              <div
                className="grid grid-cols-[auto_1fr] gap-3 rounded-xl bg-[#fffaf0] px-3 py-2 text-sm"
                key={`${session.session_id}-${set.set_number}`}
              >
                <span className="font-black text-[#265c52]">
                  S{set.set_number}
                </span>
                <span className="font-semibold text-[#5d5348]">
                  {formatSet(set.weight_kg, set.reps, set.volume_kg)}
                </span>
              </div>
            ))}
          </div>
        </details>
      ))}
    </div>
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
