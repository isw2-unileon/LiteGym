import type { ExerciseInsights, ExerciseInsightTrend } from "../../types/exercise";
import ExerciseProgressChart from "./ExerciseProgressChart";
import ExerciseSessionHistory from "./ExerciseSessionHistory";
import ExerciseStatsCards from "./ExerciseStatsCards";

type ExerciseInsightPanelProps = {
  insights: ExerciseInsights | null;
  status: "idle" | "loading" | "success" | "error";
  message: string;
  variant?: "card" | "embedded";
};

const trendLabels: Record<ExerciseInsightTrend, string> = {
  up: "Subiendo",
  stable: "Estable",
  down: "Bajando",
  empty: "Sin datos",
};

export default function ExerciseInsightPanel({
  insights,
  status,
  message,
  variant = "card",
}: ExerciseInsightPanelProps) {
  const rootClassName =
    variant === "embedded"
      ? "rounded-[1.5rem] bg-transparent"
      : "rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/70 p-5";

  return (
    <div className={rootClassName}>
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]">
            Rendimiento
          </p>
          <h3 className="mt-2 text-xl font-black tracking-[-0.04em] text-[#1f1b16]">
            Vista avanzada
          </h3>
        </div>

        {insights && (
          <span className="rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black text-[#265c52]">
            {trendLabels[insights.summary.trend]}
          </span>
        )}
      </div>

      {status === "loading" && (
        <p className="mt-4 rounded-2xl border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/80 p-4 text-sm font-semibold text-[#5d5348]">
          Cargando estadisticas...
        </p>
      )}

      {status === "error" && (
        <p className="mt-4 rounded-2xl border border-[#9f2f22]/15 bg-[#fff0ec] p-4 text-sm font-semibold text-[#9f2f22]">
          {message}
        </p>
      )}

      {status === "success" && insights && insights.summary.session_count === 0 && (
        <p className="mt-4 rounded-2xl border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/80 p-4 text-sm font-semibold leading-6 text-[#5d5348]">
          Todavia no hay datos historicos para calcular rendimiento.
        </p>
      )}

      {status === "success" && insights && insights.summary.session_count > 0 && (
        <div className="mt-4 grid gap-4">
          <details open className="group">
            <summary className="cursor-pointer list-none text-sm font-black uppercase tracking-[0.16em] text-[#265c52]">
              Resumen
            </summary>
            <div className="mt-3">
              <ExerciseStatsCards insights={insights} />
            </div>
          </details>

          <details className="group">
            <summary className="cursor-pointer list-none text-sm font-black uppercase tracking-[0.16em] text-[#265c52]">
              Progresion
            </summary>
            <div className="mt-3 grid gap-3">
              <ExerciseProgressChart
                points={insights.progression}
                metric="max_weight_kg"
              />
              <ExerciseProgressChart
                points={insights.progression}
                metric="volume_kg"
              />
            </div>
          </details>

          <details className="group">
            <summary className="cursor-pointer list-none text-sm font-black uppercase tracking-[0.16em] text-[#265c52]">
              Historial
            </summary>
            <div className="mt-3">
              <ExerciseSessionHistory history={insights.history} />
            </div>
          </details>
        </div>
      )}
    </div>
  );
}
