import type { ExerciseInsights } from "../../types/exercise";

type ExerciseStatsCardsProps = {
  insights: ExerciseInsights;
};

const numberFormatter = new Intl.NumberFormat("es-ES", {
  maximumFractionDigits: 1,
});

const dateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "numeric",
  month: "short",
  year: "numeric",
});

export default function ExerciseStatsCards({ insights }: ExerciseStatsCardsProps) {
  const { summary } = insights;

  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      <InsightStatCard
        label="Peso máximo"
        value={formatOptionalNumber(summary.max_weight_kg, " kg")}
      />
      <InsightStatCard
        label="Mejor serie"
        value={formatBestSet(insights)}
      />
      <InsightStatCard
        label="Volumen total"
        value={`${numberFormatter.format(summary.total_volume_kg)} kg`}
      />
      <InsightStatCard
        label="Sesiones"
        value={String(summary.session_count)}
      />
      <InsightStatCard
        label="Ultima vez"
        value={
          summary.last_performed_at
            ? dateFormatter.format(new Date(summary.last_performed_at))
            : "Sin registros"
        }
      />
      <InsightStatCard
        label="Frecuencia"
        value={
          typeof summary.average_days_between === "number"
            ? `Cada ${numberFormatter.format(summary.average_days_between)} dias`
            : "Sin comparativa"
        }
      />
    </div>
  );
}

function InsightStatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-[#1f1b16]/10 bg-white/70 p-4">
      <p className="text-[11px] font-black uppercase tracking-[0.18em] text-[#265c52]">
        {label}
      </p>
      <p className="mt-2 break-words text-lg font-black tracking-[-0.03em] text-[#1f1b16]">
        {value}
      </p>
    </div>
  );
}

function formatOptionalNumber(value: number | null | undefined, suffix: string) {
  if (typeof value !== "number") {
    return "Sin datos";
  }

  return `${numberFormatter.format(value)}${suffix}`;
}

function formatBestSet(insights: ExerciseInsights) {
  const bestSet = insights.best_set;
  if (!bestSet) {
    return "Sin datos";
  }

  const reps = typeof bestSet.reps === "number" ? `${bestSet.reps} reps` : "sin reps";
  const weight =
    typeof bestSet.weight_kg === "number"
      ? `${numberFormatter.format(bestSet.weight_kg)} kg`
      : "sin peso";

  return `${weight} x ${reps}`;
}
