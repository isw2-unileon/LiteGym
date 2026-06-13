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

const numberFormatter = new Intl.NumberFormat("es-ES", {
  maximumFractionDigits: 1,
});

const dateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "2-digit",
  month: "short",
  year: "numeric",
});

export default function ExerciseInsightPanel({
  insights,
  status,
  message,
  variant = "card",
}: ExerciseInsightPanelProps) {
  const rootClassName =
    variant === "embedded"
      ? "rounded-[1.5rem] bg-transparent"
      : "rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/70 p-5 backdrop-blur"

  return (
    <div className={rootClassName}>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="text-xs font-black uppercase tracking-[0.22em] text-[#265c52]">
            Rendimiento
          </p>
          <h3 className="mt-2 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-2xl font-black tracking-[-0.04em] text-[#1f1b16]">
            Vista avanzada
          </h3>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-[#5d5348]">
            La lectura de arriba hacia abajo prioriza primero la foto general,
            luego la progresión y al final el detalle de cada sesión.
          </p>
        </div>

        {insights && (
          <div className="rounded-full border border-[#265c52]/15 bg-[#265c52]/10 px-3 py-2 text-xs font-black uppercase tracking-[0.16em] text-[#265c52]">
            {trendLabels[insights.summary.trend]}
          </div>
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
          <section className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-4 sm:p-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#265c52]">
                  Resumen rapido
                </p>
                <p className="mt-2 text-sm font-semibold leading-6 text-[#5d5348]">
                  Contexto resumido del ejercicio con los indicadores mas utiles
                  para comparar la progresion de un vistazo.
                </p>
              </div>
              <span className="rounded-full bg-[#1f1b16] px-3 py-1 text-xs font-black uppercase tracking-[0.16em] text-[#f1a45b]">
                {insights.summary.session_count} sesiones
              </span>
            </div>

            <div className="mt-4">
              <ExerciseStatsCards insights={insights} />
            </div>

            {insights.personal_records.length > 0 && (
              <div className="mt-4 rounded-[1.25rem] border border-[#1f1b16]/10 bg-white p-4">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="text-[11px] font-black uppercase tracking-[0.2em] text-[#265c52]">
                      Marcas personales
                    </p>
                    <p className="mt-2 text-sm font-semibold leading-6 text-[#5d5348]">
                      Mejores registros detectados en el historial del ejercicio.
                    </p>
                  </div>
                  <span className="rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black uppercase tracking-[0.14em] text-[#265c52]">
                    {insights.personal_records.length} registros
                  </span>
                </div>

                <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                  {insights.personal_records.slice(0, 6).map((record) => (
                    <article
                      key={`${record.type}-${record.performed_at}`}
                      className="rounded-[1rem] border border-[#1f1b16]/10 bg-[#fffaf0] p-3"
                    >
                      <p className="text-[11px] font-black uppercase tracking-[0.18em] text-[#265c52]">
                        {record.label}
                      </p>
                      <p className="mt-2 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-2xl font-black tracking-[-0.04em] text-[#1f1b16]">
                        {numberFormatter.format(record.value)} {record.unit}
                      </p>
                      <p className="mt-2 text-xs font-semibold uppercase tracking-[0.14em] text-[#5d5348]">
                        {dateFormatter.format(new Date(record.performed_at))}
                      </p>
                    </article>
                  ))}
                </div>
              </div>
            )}
          </section>

          <section className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-4 sm:p-5">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#265c52]">
                  Progresion
                </p>
                <p className="mt-2 text-sm font-semibold leading-6 text-[#5d5348]">
                  Curvas con lectura de peso, repeticiones y volumen para ver
                  el ritmo de mejora y detectar estancamientos.
                </p>
              </div>
              <span className="rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-1 text-xs font-black uppercase tracking-[0.14em] text-[#3a332c]">
                {insights.progression.length} registros
              </span>
            </div>

            <div className="mt-4 grid gap-3">
              <ExerciseProgressChart
                points={insights.progression}
                metric="max_weight_kg"
              />
              <ExerciseProgressChart
                points={insights.progression}
                metric="max_reps"
              />
              <ExerciseProgressChart
                points={insights.progression}
                metric="volume_kg"
              />
            </div>
          </section>

          <section className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-4 sm:p-5">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#265c52]">
                  Historico
                </p>
                <p className="mt-2 text-sm font-semibold leading-6 text-[#5d5348]">
                  El detalle de cada sesion con sus series, cargas y volumen
                  total para comparar la ejecucion real.
                </p>
              </div>
              <span className="rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-1 text-xs font-black uppercase tracking-[0.14em] text-[#3a332c]">
                {insights.history.length} sesiones
              </span>
            </div>

            <div className="mt-4">
              <ExerciseSessionHistory history={insights.history} />
            </div>
          </section>
        </div>
      )}
    </div>
  );
}
