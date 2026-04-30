import { useEffect, useMemo, useState } from "react";
import { Link, useOutletContext } from "react-router-dom";
import type { LayoutUser } from "../components/AppLayout";
import { apiUrl } from "../lib/api";

type OutletContext = {
  user?: LayoutUser | null;
};

type DashboardRoutineSummary = {
  id: string;
  name: string;
  description?: string | null;
  exercise_count: number;
  updated_at: string;
};

type DashboardWorkoutSummary = {
  id: string;
  name?: string | null;
  routine_name?: string | null;
  started_at: string;
  duration_minutes: number;
  exercise_count: number;
};

type DashboardProgressMetric = {
  current?: number | null;
  previous?: number | null;
  delta?: number | null;
};

type DashboardOverview = {
  calendar: {
    month: string;
    trained_days: string[];
    sessions_count: number;
    current_streak: number;
    next_objective: string;
  };
  recent_routines: DashboardRoutineSummary[];
  recent_workouts: DashboardWorkoutSummary[];
  progress: {
    last_recorded_at?: string | null;
    weight_kg: DashboardProgressMetric;
    body_fat_percentage: DashboardProgressMetric;
    muscle_mass_kg: DashboardProgressMetric;
  };
  muscle_distribution: {
    year: Array<{ name: string; count: number; percentage: number }>;
    month: Array<{ name: string; count: number; percentage: number }>;
    year_exercise_count: number;
    month_exercise_count: number;
  };
};

const calendarFormatter = new Intl.DateTimeFormat("es-ES", { month: "long", year: "numeric" });
const weekdayFormatter = new Intl.DateTimeFormat("es-ES", { weekday: "short" });
const workoutDateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "numeric",
  month: "short",
});

function buildMonthDays(referenceDate: Date) {
  const year = referenceDate.getFullYear();
  const month = referenceDate.getMonth();
  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0);
  const offset = (firstDay.getDay() + 6) % 7;
  const totalCells = Math.ceil((offset + lastDay.getDate()) / 7) * 7;

  return Array.from({ length: totalCells }, (_, index) => {
    const dayNumber = index - offset + 1;
    const date = new Date(year, month, dayNumber);
    const isCurrentMonth = date.getMonth() === month;

    return {
      key: `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`,
      date,
      dayNumber: date.getDate(),
      isCurrentMonth,
      isToday: isCurrentMonth && date.toDateString() === referenceDate.toDateString(),
    };
  });
}

function polarToCartesian(cx: number, cy: number, radius: number, angleInDegrees: number) {
  const angleInRadians = ((angleInDegrees - 90) * Math.PI) / 180;

  return {
    x: cx + radius * Math.cos(angleInRadians),
    y: cy + radius * Math.sin(angleInRadians),
  };
}

function buildMonthDate(monthValue?: string) {
  if (!monthValue) {
    return new Date();
  }

  const [year, month] = monthValue.split("-").map(Number);
  if (!year || !month) {
    return new Date();
  }

  return new Date(year, month - 1, 1);
}

function toDateKey(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function formatMetricValue(value?: number | null, suffix = "") {
  if (typeof value !== "number") {
    return "Sin datos";
  }

  return `${value.toFixed(1)}${suffix}`;
}

function formatMetricDelta(value?: number | null, suffix = "") {
  if (typeof value !== "number") {
    return "Aun no hay comparativa";
  }

  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(1)}${suffix}`;
}

export default function DashboardPage() {
  const { user } = useOutletContext<OutletContext>();
  const [dashboard, setDashboard] = useState<DashboardOverview | null>(null);
  const [statsRange, setStatsRange] = useState<"year" | "month">("year");
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");

  useEffect(() => {
    const fetchDashboard = async () => {
      setStatus("loading");

      try {
        const response = await fetch(apiUrl("/api/dashboard"), {
          credentials: "include",
        });

        if (!response.ok) {
          setDashboard(null);
          setStatus("error");
          return;
        }

        const payload = (await response.json()) as DashboardOverview;
        setDashboard(payload);
        setStatus("success");
      } catch {
        setDashboard(null);
        setStatus("error");
      }
    };

    void fetchDashboard();
  }, []);

  const currentMonth = useMemo(() => buildMonthDate(dashboard?.calendar.month), [dashboard?.calendar.month]);
  const monthDays = useMemo(() => buildMonthDays(currentMonth), [currentMonth]);
  const weekdayLabels = useMemo(() => {
    const baseMonday = new Date(2026, 3, 6);
    return Array.from({ length: 7 }, (_, index) => {
      const date = new Date(baseMonday);
      date.setDate(baseMonday.getDate() + index);
      return weekdayFormatter.format(date).replace(".", "");
    });
  }, []);

  const trainingDays = useMemo(
    () => new Set(dashboard?.calendar.trained_days ?? []),
    [dashboard?.calendar.trained_days],
  );

  const selectedBreakdown = useMemo(
    () => (
      statsRange === "year"
        ? (dashboard?.muscle_distribution.year ?? [])
        : (dashboard?.muscle_distribution.month ?? [])
    ),
    [dashboard, statsRange],
  );
  const selectedExerciseCount = statsRange === "year"
    ? (dashboard?.muscle_distribution.year_exercise_count ?? 0)
    : (dashboard?.muscle_distribution.month_exercise_count ?? 0);
  const selectedMuscleGroupCount = selectedBreakdown.length;

  const hexagonStats = useMemo(() => {
    const fallbackAxes = ["Pecho", "Pierna", "Espalda", "Hombro", "Core", "Cardio"];
    const topGroups = selectedBreakdown.slice(0, 6);
    const maxCount = topGroups[0]?.count ?? 1;

    const axes = (topGroups.length > 0 ? topGroups.map((group) => group.name) : fallbackAxes).slice(0, 6);
    let fallbackIndex = 0;
    while (axes.length < 6) {
      const candidate = fallbackAxes[fallbackIndex] ?? `Grupo ${axes.length + 1}`;
      fallbackIndex += 1;

      if (!axes.includes(candidate)) {
        axes.push(candidate);
      }
    }

    const values = axes.map((axis) => {
      const match = topGroups.find((group) => group.name === axis);
      return match ? Math.max(0.22, match.count / maxCount) : 0.22;
    });

    return axes.map((axis, index) => ({
      axis,
      value: values[index] ?? 0.22,
    }));
  }, [selectedBreakdown]);

  const hexagonLevels = [0.25, 0.5, 0.75, 1];
  const chartCenter = 108;
  const chartRadius = 78;
  const chartPoints = hexagonStats.map((item, index) =>
    polarToCartesian(chartCenter, chartCenter, chartRadius * item.value, (360 / hexagonStats.length) * index),
  );
  const chartPolygon = chartPoints.map((point) => `${point.x},${point.y}`).join(" ");

  return (
    <>
      <section className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/85 p-5 shadow-[0_24px_60px_rgba(47,39,27,0.14)] backdrop-blur-md sm:p-6">
        <div className="flex flex-col gap-6 xl:flex-row xl:items-start xl:justify-between">
          <div className="max-w-xl">
            <p className="text-sm font-black uppercase tracking-[0.18em] text-[#265c52]">Calendario</p>
            <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.05em] sm:text-5xl">
              {user?.username ? `Hola, ${user.username}` : "Hola"}
            </h2>
            <p className="mt-3 text-sm font-semibold leading-6 text-[#5d5348]">
              Aqui se ve tu actividad reciente y los dias del mes en los que ya has entrenado.
            </p>

            <div className="mt-6 flex flex-wrap gap-3">
              <div className="rounded-2xl border border-[#1f1b16]/10 bg-white/70 px-4 py-3">
                <p className="text-xs font-black uppercase tracking-[0.14em] text-[#265c52]">Sesiones del mes</p>
                <p className="mt-2 text-2xl font-black text-[#1f1b16]">{dashboard?.calendar.sessions_count ?? 0}</p>
              </div>
              <div className="rounded-2xl border border-[#1f1b16]/10 bg-white/70 px-4 py-3">
                <p className="text-xs font-black uppercase tracking-[0.14em] text-[#265c52]">Racha actual</p>
                <p className="mt-2 text-2xl font-black text-[#1f1b16]">{dashboard?.calendar.current_streak ?? 0} dias</p>
              </div>
              <div className="rounded-2xl border border-[#1f1b16]/10 bg-white/70 px-4 py-3">
                <p className="text-xs font-black uppercase tracking-[0.14em] text-[#265c52]">Proximo objetivo</p>
                <p className="mt-2 text-sm font-black text-[#1f1b16]">{dashboard?.calendar.next_objective ?? "Sin datos"}</p>
              </div>
            </div>

            {status === "error" && (
              <p className="mt-4 text-sm font-semibold text-[#7a4d2a]">
                No se ha podido cargar toda la informacion del dashboard.
              </p>
            )}
          </div>

          <div className="w-full rounded-[1.8rem] border border-[#1f1b16]/10 bg-white/70 p-4 sm:p-5 xl:max-w-[29rem]">
            <div className="flex items-center justify-between gap-4">
              <div>
                <p className="text-xs font-black uppercase tracking-[0.14em] text-[#265c52]">Vista mensual</p>
                <h3 className="mt-2 text-2xl font-black capitalize tracking-[-0.03em] text-[#1f1b16]">
                  {calendarFormatter.format(currentMonth)}
                </h3>
              </div>
              <span className="rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black uppercase tracking-[0.14em] text-[#265c52]">
                Datos reales
              </span>
            </div>

            <div className="mt-5 grid grid-cols-7 gap-1.5 sm:gap-2">
              {weekdayLabels.map((label) => (
                <p className="text-center text-xs font-black uppercase tracking-[0.14em] text-[#5d5348]" key={label}>
                  {label}
                </p>
              ))}

              {monthDays.map((day) => {
                const isTrainedDay = day.isCurrentMonth && trainingDays.has(toDateKey(day.date));

                return (
                  <div
                    className={`grid aspect-square place-items-center rounded-xl text-xs font-black transition sm:rounded-2xl sm:text-sm ${
                      day.isCurrentMonth
                        ? isTrainedDay
                          ? "bg-[#265c52] text-[#fffaf0]"
                          : day.isToday
                            ? "border border-[#265c52]/30 bg-[#265c52]/10 text-[#1f1b16]"
                            : "bg-[#fffaf0] text-[#1f1b16]"
                        : "bg-transparent text-[#b3a89b]"
                    }`}
                    key={day.key}
                  >
                    {day.dayNumber}
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </section>

      <section className="mt-8 grid gap-6 xl:grid-cols-[minmax(0,1.55fr)_minmax(19rem,0.95fr)]">
        <div className="space-y-6">
          <article className="rounded-[2rem] border border-[#1f1b16]/10 bg-white/60 p-6 shadow-[0_18px_40px_rgba(47,39,27,0.08)]">
            <p className="text-sm font-black uppercase tracking-[0.18em] text-[#265c52]">Entrenos</p>
            <h3 className="mt-2 text-2xl font-black tracking-[-0.03em] text-[#1f1b16]">Ultimos entrenos</h3>

            {dashboard?.recent_workouts.length ? (
              <ul className="mt-5 space-y-3">
                {dashboard.recent_workouts.map((workout) => (
                  <li className="rounded-[1.4rem] border border-[#1f1b16]/10 bg-[#fffaf0]/70 p-4" key={workout.id}>
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <p className="text-sm font-black text-[#1f1b16]">{workout.name || workout.routine_name || "Sesion"}</p>
                        <p className="mt-1 text-xs font-semibold uppercase tracking-[0.14em] text-[#5d5348]">
                          {workout.routine_name || "Entreno libre"}
                        </p>
                      </div>
                      <p className="text-xs font-black uppercase tracking-[0.14em] text-[#265c52]">
                        {workoutDateFormatter.format(new Date(workout.started_at))}
                      </p>
                    </div>
                    <p className="mt-3 text-sm font-semibold text-[#5d5348]">
                      {workout.exercise_count} ejercicios · {workout.duration_minutes} min
                    </p>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="mt-4 rounded-[1.4rem] border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/70 p-4 text-sm font-semibold leading-6 text-[#5d5348]">
                Cuando registres sesiones, aqui apareceran tus entrenos mas recientes.
              </p>
            )}
          </article>

          <article className="rounded-[2rem] border border-[#1f1b16]/10 bg-white/60 p-6 shadow-[0_18px_40px_rgba(47,39,27,0.08)]">
            <div className="flex items-center justify-between gap-4">
              <div>
                <p className="text-sm font-black uppercase tracking-[0.18em] text-[#265c52]">Rutinas</p>
                <h3 className="mt-2 text-2xl font-black tracking-[-0.03em] text-[#1f1b16]">Tus rutinas</h3>
              </div>
              <Link
                className="inline-flex rounded-2xl border border-[#1f1b16]/10 bg-[#fffaf0]/80 px-4 py-3 text-sm font-black text-[#1f1b16] transition hover:bg-white"
                to="/routines"
              >
                Ver rutinas
              </Link>
            </div>

            {dashboard?.recent_routines.length ? (
              <ul className="mt-5 space-y-3">
                {dashboard.recent_routines.map((routine) => (
                  <li className="rounded-[1.4rem] border border-[#1f1b16]/10 bg-[#fffaf0]/70 p-4" key={routine.id}>
                    <p className="text-sm font-black text-[#1f1b16]">{routine.name}</p>
                    <p className="mt-1 text-xs font-semibold uppercase tracking-[0.14em] text-[#5d5348]">
                      {routine.exercise_count} ejercicios
                    </p>
                    <p className="mt-3 text-sm font-semibold leading-6 text-[#5d5348]">
                      {routine.description || "Rutina lista para retomar cuando quieras."}
                    </p>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="mt-4 rounded-[1.4rem] border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/70 p-4 text-sm font-semibold leading-6 text-[#5d5348]">
                Cuando tengas rutinas guardadas, aqui apareceran las mas recientes.
              </p>
            )}
          </article>

          <article className="rounded-[2rem] border border-[#1f1b16]/10 bg-white/60 p-6 shadow-[0_18px_40px_rgba(47,39,27,0.08)]">
            <p className="text-sm font-black uppercase tracking-[0.18em] text-[#265c52]">Estadisticas</p>
            <h3 className="mt-2 text-2xl font-black tracking-[-0.03em] text-[#1f1b16]">Resumen de progreso</h3>

            <div className="mt-5 grid gap-3 sm:grid-cols-3">
              <div className="rounded-[1.4rem] border border-[#1f1b16]/10 bg-[#fffaf0]/70 p-4">
                <p className="text-xs font-black uppercase tracking-[0.14em] text-[#5d5348]">Peso</p>
                <p className="mt-2 text-2xl font-black text-[#1f1b16]">
                  {formatMetricValue(dashboard?.progress.weight_kg.current, " kg")}
                </p>
                <p className="mt-2 text-sm font-semibold text-[#5d5348]">
                  {formatMetricDelta(dashboard?.progress.weight_kg.delta, " kg")}
                </p>
              </div>

              <div className="rounded-[1.4rem] border border-[#1f1b16]/10 bg-[#fffaf0]/70 p-4">
                <p className="text-xs font-black uppercase tracking-[0.14em] text-[#5d5348]">Grasa corporal</p>
                <p className="mt-2 text-2xl font-black text-[#1f1b16]">
                  {formatMetricValue(dashboard?.progress.body_fat_percentage.current, "%")}
                </p>
                <p className="mt-2 text-sm font-semibold text-[#5d5348]">
                  {formatMetricDelta(dashboard?.progress.body_fat_percentage.delta, "%")}
                </p>
              </div>

              <div className="rounded-[1.4rem] border border-[#1f1b16]/10 bg-[#fffaf0]/70 p-4">
                <p className="text-xs font-black uppercase tracking-[0.14em] text-[#5d5348]">Masa muscular</p>
                <p className="mt-2 text-2xl font-black text-[#1f1b16]">
                  {formatMetricValue(dashboard?.progress.muscle_mass_kg.current, " kg")}
                </p>
                <p className="mt-2 text-sm font-semibold text-[#5d5348]">
                  {formatMetricDelta(dashboard?.progress.muscle_mass_kg.delta, " kg")}
                </p>
              </div>
            </div>
          </article>
        </div>

        <div className="space-y-6">
          <article className="rounded-[2rem] border border-[#1f1b16]/10 bg-white/65 p-6 shadow-[0_20px_50px_rgba(47,39,27,0.10)]">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-sm font-black uppercase tracking-[0.18em] text-[#265c52]">Estadisticas</p>
                <h3 className="mt-2 text-2xl font-black tracking-[-0.03em]">Distribucion muscular</h3>
              </div>

              <div
                aria-label="Rango de estadisticas"
                className="inline-flex rounded-2xl border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-1"
                role="group"
              >
                <button
                  aria-pressed={statsRange === "year"}
                  className={`rounded-xl px-3 py-1.5 text-sm font-black transition ${
                    statsRange === "year" ? "bg-[#1f1b16] text-[#fffaf0]" : "text-[#5d5348] hover:bg-white"
                  }`}
                  type="button"
                  onClick={() => setStatsRange("year")}
                >
                  Año
                </button>
                <button
                  aria-pressed={statsRange === "month"}
                  className={`rounded-xl px-3 py-1.5 text-sm font-black transition ${
                    statsRange === "month" ? "bg-[#1f1b16] text-[#fffaf0]" : "text-[#5d5348] hover:bg-white"
                  }`}
                  type="button"
                  onClick={() => setStatsRange("month")}
                >
                  Mes
                </button>
              </div>
            </div>

            <div className="mt-5 grid gap-6">
              <div className="mx-auto w-full max-w-[18rem]">
                <svg className="h-auto w-full" viewBox="0 0 216 216" role="img" aria-label="Grafico hexagonal de distribucion muscular">
                  {hexagonLevels.map((level) => {
                    const points = hexagonStats
                      .map((_, index) => polarToCartesian(chartCenter, chartCenter, chartRadius * level, (360 / hexagonStats.length) * index))
                      .map((point) => `${point.x},${point.y}`)
                      .join(" ");

                    return <polygon fill="none" key={level} points={points} stroke="rgba(31,27,22,0.14)" strokeWidth="1.5" />;
                  })}

                  {hexagonStats.map((item, index) => {
                    const outerPoint = polarToCartesian(chartCenter, chartCenter, chartRadius, (360 / hexagonStats.length) * index);
                    const labelPoint = polarToCartesian(chartCenter, chartCenter, chartRadius + 20, (360 / hexagonStats.length) * index);

                    return (
                      <g key={item.axis}>
                        <line stroke="rgba(31,27,22,0.12)" strokeWidth="1.5" x1={chartCenter} x2={outerPoint.x} y1={chartCenter} y2={outerPoint.y} />
                        <text
                          fill="#5d5348"
                          fontSize="10"
                          fontWeight="700"
                          textAnchor="middle"
                          x={labelPoint.x}
                          y={labelPoint.y}
                        >
                          {item.axis}
                        </text>
                      </g>
                    );
                  })}

                  <polygon fill="rgba(38,92,82,0.16)" points={chartPolygon} stroke="#265c52" strokeWidth="3" />
                </svg>
              </div>

              <div className="rounded-[1.4rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-4">
                <p className="text-3xl font-black text-[#1f1b16]">{selectedMuscleGroupCount}</p>
                <p className="mt-1 text-sm font-semibold text-[#5d5348]">
                  Grupos musculares presentes
                </p>
                <p className="mt-2 text-xs font-semibold uppercase tracking-[0.14em] text-[#5d5348]">
                  {statsRange === "year" ? "Ultimo año" : "Ultimo mes"} · {selectedExerciseCount} ejercicios considerados
                </p>
              </div>

              {selectedBreakdown.length > 0 ? (
                <ul className="space-y-3">
                  {selectedBreakdown.slice(0, 4).map((group) => (
                    <li className="rounded-[1.4rem] border border-[#1f1b16]/10 bg-[#fffaf0]/70 p-4" key={group.name}>
                      <div className="flex items-center justify-between gap-4">
                        <div>
                          <p className="text-sm font-black text-[#1f1b16]">{group.name}</p>
                          <p className="mt-1 text-xs font-semibold uppercase tracking-[0.14em] text-[#5d5348]">
                            {group.count} ejercicios
                          </p>
                        </div>
                        <p className="text-lg font-black text-[#265c52]">{group.percentage}%</p>
                      </div>
                      <div className="mt-3 h-2.5 overflow-hidden rounded-full bg-[#1f1b16]/8">
                        <div
                          aria-hidden="true"
                          className="h-full rounded-full bg-[#265c52]"
                          style={{ width: `${group.percentage}%` }}
                        />
                      </div>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="rounded-[1.4rem] border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/70 p-4 text-sm font-semibold leading-6 text-[#5d5348]">
                  Cuando haya historial de entrenos, aqui veras que grupos musculares predominan.
                </p>
              )}
            </div>
          </article>
        </div>
      </section>
    </>
  );
}
