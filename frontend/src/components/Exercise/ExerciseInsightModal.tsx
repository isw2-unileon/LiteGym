import type { Exercise, ExerciseInsights, ExerciseStatus } from "../../types/exercise";
import { exerciseTypeLabel, muscleGroupLabel } from "../../lib/exerciseLabels";
import ExerciseInsightPanel from "./ExerciseInsightPanel";

type ExerciseInsightModalProps = {
  isOpen: boolean;
  exercise: Exercise | null;
  insights: ExerciseInsights | null;
  status: ExerciseStatus;
  message: string;
  onClose: () => void;
};

const numberFormatter = new Intl.NumberFormat("es-ES", {
  maximumFractionDigits: 1,
});

const dateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "2-digit",
  month: "short",
  year: "numeric",
});

export default function ExerciseInsightModal({
  isOpen,
  exercise,
  insights,
  status,
  message,
  onClose,
}: ExerciseInsightModalProps) {
  if (!isOpen || !exercise) {
    return null;
  }

  const summary = insights?.summary;
  const sessionCount = summary?.session_count ?? 0;
  const totalVolume = summary?.total_volume_kg ?? 0;
  const maxWeight = summary?.max_weight_kg ?? null;
  const trend = summary?.trend ?? "empty";
  const firstPerformedAt = summary?.first_performed_at ? dateFormatter.format(new Date(summary.first_performed_at)) : "Sin datos";
  const lastPerformedAt = summary?.last_performed_at ? dateFormatter.format(new Date(summary.last_performed_at)) : "Sin datos";
  const personalRecordCount = insights?.personal_records.length ?? 0;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#1f1b16]/68 px-3 py-4 backdrop-blur-md sm:px-6"
      onClick={onClose}
      role="presentation"
    >
      <div
        className="relative flex h-[min(92dvh,980px)] w-full max-w-7xl flex-col overflow-hidden rounded-[28px] border-2 border-[#fffaf0]/20 bg-[#fffaf0] shadow-[0_40px_120px_rgba(31,27,22,0.32)]"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="grid min-h-0 flex-1 gap-0 overflow-hidden lg:grid-cols-[minmax(280px,0.92fr)_minmax(0,1.28fr)]">
          <aside className="flex min-h-0 flex-col justify-between overflow-hidden bg-[#1f1b16] px-5 py-5 text-[#fffaf0] sm:px-7 sm:py-7 lg:sticky lg:top-0">
            <div className="space-y-5">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-xs font-black uppercase tracking-[0.28em] text-[#f1a45b]">
                    Rendimiento
                  </p>
                  <h3 className="mt-3 break-words font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.05em] text-[#fffaf0]">
                    {exercise.name}
                  </h3>
                </div>

                <button
                  className="grid h-10 w-10 shrink-0 place-items-center rounded-[12px] border border-[#fffaf0]/15 text-[#fffaf0] transition hover:bg-[#fffaf0]/10 hover:rotate-90"
                  type="button"
                  onClick={onClose}
                  aria-label="Cerrar modal"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" className="h-4 w-4">
                    <path d="M6 6l12 12M18 6L6 18" />
                  </svg>
                </button>
              </div>

              <div className="flex flex-wrap gap-2">
                <Chip dark>{muscleGroupLabel(exercise.muscle_group)}</Chip>
                {exercise.exercise_type && <Chip dark>{exerciseTypeLabel(exercise.exercise_type)}</Chip>}
                <Chip dark tone={exercise.is_official === false ? "warn" : "calm"}>
                  {exercise.is_official === false ? "Ejercicio propio" : "Ejercicio oficial"}
                </Chip>
              </div>

              <p className="max-w-md text-sm leading-6 text-[#efe4d2]">
                {exercise.description?.trim() ||
                  "El panel concentra el historial, el volumen y la progresión real del ejercicio para leer la evolución sin salir del detalle."}
              </p>

              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
                <MetricCard label="Sesiones" value={String(sessionCount)} helper="Registradas con este ejercicio" />
                <MetricCard
                  label="Volumen total"
                  value={`${numberFormatter.format(totalVolume)} kg`}
                  helper={trendLabel(trend)}
                />
                <MetricCard
                  label="Peso máximo"
                  value={maxWeight == null ? "Sin datos" : `${numberFormatter.format(maxWeight)} kg PR`}
                  helper={`${personalRecordCount} marcas personales`}
                />
              </div>

              <div className="grid grid-cols-2 gap-3 rounded-[22px] border border-[#fffaf0]/10 bg-[#fffaf0]/6 p-4">
                <div>
                  <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#f1a45b]">
                    Primera vez
                  </p>
                  <p className="mt-2 text-sm font-semibold text-[#fffaf0]">{firstPerformedAt}</p>
                </div>
                <div>
                  <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#f1a45b]">
                    Ultima vez
                  </p>
                  <p className="mt-2 text-sm font-semibold text-[#fffaf0]">{lastPerformedAt}</p>
                </div>
              </div>
            </div>

          </aside>

          <div className="min-h-0 overflow-y-auto px-4 py-4 [scrollbar-gutter:stable] sm:px-6 sm:py-6 lg:px-7">
            <ExerciseInsightPanel
              insights={insights}
              status={status}
              message={message}
              variant="embedded"
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function MetricCard({
  label,
  value,
  helper,
}: {
  label: string;
  value: string;
  helper: string;
}) {
  return (
    <div className="rounded-[20px] border border-[#fffaf0]/10 bg-[#fffaf0]/6 p-4">
      <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#f1a45b]">
        {label}
      </p>
      <p className="mt-2 break-words font-['Aptos_Display','Trebuchet_MS',sans-serif] text-2xl font-black leading-none tracking-[-0.04em] text-[#fffaf0]">
        {value}
      </p>
      <p className="mt-2 text-sm leading-5 text-[#efe4d2]">{helper}</p>
    </div>
  );
}

function Chip({
  children,
  dark = false,
  tone = "default",
}: {
  children: string;
  dark?: boolean;
  tone?: "default" | "warn" | "calm";
}) {
  const toneClasses =
    tone === "warn"
      ? "border-[#ea7130]/35 bg-[#ea7130]/15 text-[#f7d08a]"
      : tone === "calm"
        ? "border-[#fffaf0]/15 bg-[#fffaf0]/10 text-[#fffaf0]"
        : "border-[#fffaf0]/15 bg-[#fffaf0]/8 text-[#fffaf0]/90";

  return (
    <span
      className={[
        "inline-flex max-w-full items-center rounded-full border px-3 py-1 text-[11px] font-black uppercase tracking-[0.14em]",
        dark ? toneClasses : "border-[#1f1b16]/10 bg-[#fffaf0] text-[#1f1b16]",
      ].join(" ")}
    >
      <span className="truncate">{children}</span>
    </span>
  );
}

function trendLabel(trend: string) {
  switch (trend) {
    case "up":
      return "Tendencia positiva";
    case "down":
      return "Tendencia a la baja";
    case "stable":
      return "Estable";
    default:
      return "Aun sin tendencia";
  }
}
