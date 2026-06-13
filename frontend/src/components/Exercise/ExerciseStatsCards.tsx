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
    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
      <InsightStatCard
        label="Peso maximo"
        value={formatOptionalNumber(summary.max_weight_kg, " kg máx.")}
        helper="Mejor repeticion con carga"
      />
      <InsightStatCard
        label="Mejor serie"
        value={formatBestSet(insights)}
        helper="Combinacion mas fuerte registrada"
      />
      <InsightStatCard
        label="Volumen total"
        value={formatTotalVolume(summary.total_volume_kg)}
        helper="Suma de todas las series"
      />
      <InsightStatCard
        label="Sesiones"
        value={String(summary.session_count)}
        helper="Entradas historicas disponibles"
      />
      <InsightStatCard
        label="Ultima vez"
        value={
          summary.last_performed_at
            ? dateFormatter.format(new Date(summary.last_performed_at))
            : "Sin registros"
        }
        helper="Momento del ultimo registro"
      />
      <InsightStatCard
        label="Frecuencia"
        value={
          typeof summary.average_days_between === "number"
            ? `Cada ${numberFormatter.format(summary.average_days_between)} dias`
            : "Sin comparativa"
        }
        helper="Distancia media entre sesiones"
      />
    </div>
  );
}

function InsightStatCard({
  label,
  value,
  helper,
}: {
  label: string;
  value: string;
  helper: string;
}) {
  return (
    <div className="relative overflow-hidden rounded-[1.25rem] border border-[#1f1b16]/10 bg-white p-4 shadow-[0_8px_18px_rgba(31,27,22,0.04)]">
      <span aria-hidden="true" className="absolute inset-x-0 top-0 h-1.5 bg-[#ea7130]" />
      <p className="text-[11px] font-black uppercase tracking-[0.2em] text-[#265c52]">
        {label}
      </p>
      <p className="mt-2 break-words font-['Aptos_Display','Trebuchet_MS',sans-serif] text-[1.4rem] font-black leading-[1.05] tracking-[-0.04em] text-[#1f1b16]">
        {value}
      </p>
      <p className="mt-2 text-sm leading-5 text-[#5d5348]">{helper}</p>
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

function formatTotalVolume(value: number) {
  return `${numberFormatter.format(value)} kg`;
}
