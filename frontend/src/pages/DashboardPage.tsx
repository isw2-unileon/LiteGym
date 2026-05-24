import type { CSSProperties, ReactNode } from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
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
  performed_at: string;
  duration_minutes: number;
  exercise_count: number;
};

type DashboardCalendarWorkout = {
  id: string;
  name?: string | null;
  routine_id?: string | null;
  routine_name?: string | null;
  performed_at?: string | null;
  planned_at?: string | null;
  duration_minutes: number;
  exercise_count: number;
};

type CalendarDialog =
  | { type: "plan"; dateKey: string }
  | { type: "planned"; dateKey: string; workouts: DashboardCalendarWorkout[] }
  | { type: "performed"; dateKey: string; workouts: DashboardCalendarWorkout[] };

type DashboardProgressMetric = {
  current?: number | null;
  previous?: number | null;
  delta?: number | null;
};

type DashboardOverview = {
  calendar: {
    month: string;
    trained_days: string[];
    planned_days: string[];
    calendar_workouts: DashboardCalendarWorkout[];
    sessions_count: number;
    current_streak: number;
    weekly_goal: number;
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
const recentWorkoutDateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "2-digit",
  month: "short",
});
const workoutTimeFormatter = new Intl.DateTimeFormat("es-ES", {
  hour: "2-digit",
  minute: "2-digit",
});
const muscleLabelMap: Record<string, string> = {
  chest: "Pecho",
  shoulders: "Hombros",
  shoulder: "Hombro",
  back: "Espalda",
  legs: "Pierna",
  leg: "Pierna",
  core: "Core",
  cardio: "Cardio",
  arms: "Brazos",
  arm: "Brazo",
  biceps: "Biceps",
  triceps: "Triceps",
  glutes: "Gluteos",
  glute: "Gluteo",
};

function buildMonthDays(referenceDate: Date, todayDate: Date) {
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
      isToday: isCurrentMonth && date.toDateString() === todayDate.toDateString(),
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

function toMonthParam(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  return `${year}-${month}`;
}

function getMonthOffset(from: Date, to: Date) {
  return (to.getFullYear() - from.getFullYear()) * 12 + to.getMonth() - from.getMonth();
}

function toDateKey(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function workoutTitle(workout: DashboardCalendarWorkout) {
  return workout.name || workout.routine_name || "Entreno";
}

function normalizeMuscleLabel(label: string) {
  const normalizedKey = label.trim().toLowerCase();
  return muscleLabelMap[normalizedKey] ?? label;
}

function formatRecentWorkoutBadgeDate(value: string) {
  const formatted = recentWorkoutDateFormatter.format(new Date(value));
  const [day = "", month = ""] = formatted.replace(".", "").split(" ");
  return `${day}\n${month.toUpperCase()}`;
}

function formatRecentWorkoutMeta(value: string) {
  return workoutDateFormatter.format(new Date(value));
}

function toGoogleCalendarDateTime(date: Date) {
  return date.toISOString().replaceAll("-", "").replaceAll(":", "").replace(/\.\d{3}Z$/, "Z");
}

function buildGoogleCalendarLink(workout: DashboardCalendarWorkout) {
  if (!workout.planned_at) {
    return null;
  }

  const startDate = new Date(workout.planned_at);
  const durationMinutes = workout.duration_minutes > 0 ? workout.duration_minutes : 60;
  const endDate = new Date(startDate.getTime() + durationMinutes * 60 * 1000);
  const title = workoutTitle(workout);
  const detailParts = [
    `Rutina: ${workout.routine_name || "Entreno libre"}`,
    `${workout.exercise_count} ejercicios`,
    `Duracion estimada: ${durationMinutes} min`,
  ];

  const params = new URLSearchParams({
    action: "TEMPLATE",
    text: title,
    details: detailParts.join("\n"),
    dates: `${toGoogleCalendarDateTime(startDate)}/${toGoogleCalendarDateTime(endDate)}`,
  });

  return `https://calendar.google.com/calendar/render?${params.toString()}`;
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
  const [visibleMonth, setVisibleMonth] = useState(() => {
    const today = new Date();
    return new Date(today.getFullYear(), today.getMonth(), 1);
  });
  const [routines, setRoutines] = useState<DashboardRoutineSummary[]>([]);
  const [calendarDialog, setCalendarDialog] = useState<CalendarDialog | null>(null);
  const [selectedRoutineID, setSelectedRoutineID] = useState("");
  const [selectedPlanTime, setSelectedPlanTime] = useState("18:00");
  const [planStatus, setPlanStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [calendarActionStatus, setCalendarActionStatus] = useState<"idle" | "loading" | "error">("idle");
  const [statsRange, setStatsRange] = useState<"year" | "month">("year");
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const selectedPlanDate = calendarDialog?.type === "plan" ? calendarDialog.dateKey : null;

  const fetchDashboard = useCallback(async () => {
    setStatus("loading");

    try {
      const response = await fetch(apiUrl(`/api/dashboard?month=${toMonthParam(visibleMonth)}`), {
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
  }, [visibleMonth]);

  const fetchRoutines = useCallback(async () => {
    try {
      const response = await fetch(apiUrl("/api/routines"), {
        credentials: "include",
      });

      if (!response.ok) {
        setRoutines([]);
        return;
      }

      const payload = (await response.json()) as DashboardRoutineSummary[];
      setRoutines(payload);
    } catch {
      setRoutines([]);
    }
  }, []);

  useEffect(() => {
    void fetchDashboard();
  }, [fetchDashboard]);

  const todayDate = useMemo(() => new Date(), []);
  const todayMonth = useMemo(() => new Date(todayDate.getFullYear(), todayDate.getMonth(), 1), [todayDate]);
  const visibleMonthOffset = getMonthOffset(todayMonth, visibleMonth);
  const canShowPreviousMonth = visibleMonthOffset > -1;
  const canShowNextMonth = visibleMonthOffset < 1;
  const currentMonth = useMemo(
    () => buildMonthDate(dashboard?.calendar.month ?? toMonthParam(visibleMonth)),
    [dashboard?.calendar.month, visibleMonth],
  );
  const monthDays = useMemo(() => buildMonthDays(currentMonth, todayDate), [currentMonth, todayDate]);
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
  const plannedDays = useMemo(
    () => new Set(dashboard?.calendar.planned_days ?? []),
    [dashboard?.calendar.planned_days],
  );
  const workoutsByDate = useMemo(() => {
    const grouped = new Map<string, DashboardCalendarWorkout[]>();

    for (const workout of dashboard?.calendar.calendar_workouts ?? []) {
      const dateValue = workout.performed_at ?? workout.planned_at;
      if (!dateValue) {
        continue;
      }

      const dateKey = toDateKey(new Date(dateValue));
      grouped.set(dateKey, [...(grouped.get(dateKey) ?? []), workout]);
    }

    return grouped;
  }, [dashboard?.calendar.calendar_workouts]);
  const todayKey = toDateKey(todayDate);

  const handleChangeVisibleMonth = (offset: number) => {
    const nextOffset = visibleMonthOffset + offset;
    if (nextOffset < -1 || nextOffset > 1) {
      return;
    }

    setCalendarDialog(null);
    setVisibleMonth((current) => new Date(current.getFullYear(), current.getMonth() + offset, 1));
  };

  const handleOpenPlanModal = (dateKey: string) => {
    setCalendarDialog({ type: "plan", dateKey });
    resetPlanForm();
    void fetchRoutines();
  };

  const handleCloseCalendarDialog = () => {
    setCalendarDialog(null);
    resetPlanForm();
  };

  const resetPlanForm = () => {
    setSelectedRoutineID("");
    setSelectedPlanTime("18:00");
    setPlanStatus("idle");
    setCalendarActionStatus("idle");
  };

  const handleCalendarDayClick = (
    dateKey: string,
    performedWorkouts: DashboardCalendarWorkout[],
    plannedWorkouts: DashboardCalendarWorkout[],
  ) => {
    if (plannedWorkouts.length > 0 && dateKey >= todayKey) {
      setCalendarDialog({ type: "planned", dateKey, workouts: plannedWorkouts });
      setCalendarActionStatus("idle");
      return;
    }

    if (performedWorkouts.length > 0) {
      setCalendarDialog({ type: "performed", dateKey, workouts: performedWorkouts });
      setCalendarActionStatus("idle");
      return;
    }

    if (dateKey >= todayKey) {
      handleOpenPlanModal(dateKey);
    }
  };

  const handlePlanWorkout = async () => {
    if (!selectedPlanDate || !selectedRoutineID || !selectedPlanTime) {
      setPlanStatus("error");
      return;
    }

    const routine = routines.find((item) => item.id === selectedRoutineID);
    const [year, month, day] = selectedPlanDate.split("-").map(Number);
    if (!year || !month || !day) {
      setPlanStatus("error");
      return;
    }

    const timeParts = selectedPlanTime.split(":").map(Number);
    const hour = timeParts[0];
    const minute = timeParts[1];
    if (typeof hour !== "number" || typeof minute !== "number") {
      setPlanStatus("error");
      return;
    }
    if (!Number.isInteger(hour) || !Number.isInteger(minute) || hour < 0 || hour > 23 || minute < 0 || minute > 59) {
      setPlanStatus("error");
      return;
    }

    const plannedAt = new Date(year, month - 1, day, hour, minute, 0).toISOString();

    setPlanStatus("loading");
    try {
      const response = await fetch(apiUrl("/api/workouts/planned"), {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          routine_id: selectedRoutineID,
          name: routine?.name ?? "Entreno planificado",
          planned_at: plannedAt,
        }),
      });

      if (!response.ok) {
        setPlanStatus("error");
        return;
      }

      setPlanStatus("success");
      setCalendarDialog(null);
      await fetchDashboard();
    } catch {
      setPlanStatus("error");
    }
  };

  const handleCancelPlannedWorkout = async (workoutID: string) => {
    setCalendarActionStatus("loading");

    try {
      const response = await fetch(apiUrl(`/api/workout/${workoutID}`), {
        method: "DELETE",
        credentials: "include",
      });

      if (!response.ok) {
        setCalendarActionStatus("error");
        return;
      }

      setCalendarDialog(null);
      setCalendarActionStatus("idle");
      await fetchDashboard();
    } catch {
      setCalendarActionStatus("error");
    }
  };

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

    const axes = (topGroups.length > 0 ? topGroups.map((group) => normalizeMuscleLabel(group.name)) : fallbackAxes).slice(0, 6);
    let fallbackIndex = 0;
    while (axes.length < 6) {
      const candidate = fallbackAxes[fallbackIndex] ?? `Grupo ${axes.length + 1}`;
      fallbackIndex += 1;

      if (!axes.includes(candidate)) {
        axes.push(candidate);
      }
    }

    const values = axes.map((axis) => {
      const match = topGroups.find((group) => normalizeMuscleLabel(group.name) === axis);
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
    <div className="relative isolate overflow-x-hidden text-[#1f1b16] [font-family:'Inter',system-ui,sans-serif] antialiased">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]"
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10"
      />

      <div className="px-6 pb-20 pt-8 sm:px-8">
        <section className="mx-auto mb-6 grid max-w-[1280px] grid-cols-1 items-end gap-6 md:grid-cols-[1fr_auto]">
          <div>
            <div className="mb-2.5 flex items-center gap-2.5 text-[12px] font-extrabold uppercase tracking-[0.30em] text-[#265c52]">
              <span className="inline-block h-0.5 w-6 bg-[#265c52]" />
              Panel principal
            </div>
            <h1 className="m-0 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[44px] font-black leading-[0.92] tracking-[-0.055em] text-[#1f1b16] sm:text-[64px]">
              Hola,{" "}
              <span className="px-1 text-[#ea7130] [background:linear-gradient(180deg,transparent_60%,rgba(234,113,48,0.18)_60%)]">
                {user?.username ?? "atleta"}
              </span>
            </h1>
            <p className="mt-3.5 max-w-[640px] text-[15px] leading-[1.55] text-[#3a332c]">
              Aqui ves tu actividad reciente y los dias del mes en los que ya has entrenado. Planifica un dia tocando una celda futura.
            </p>
            {status === "error" && (
              <p className="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-3 py-2 text-[12px] font-bold text-[#9f2f22]">
                No se ha podido cargar toda la informacion del dashboard.
              </p>
            )}
          </div>

          <div className="flex flex-wrap gap-2.5">
            <Stat n={String(dashboard?.calendar.sessions_count ?? 0)} l="Sesiones del mes" />
            <Stat n={String(dashboard?.calendar.weekly_goal ?? 2)} l="Objetivo semanal" />
            <Stat n={`${dashboard?.calendar.current_streak ?? 0}`} l="Racha semanal" accent />
          </div>
        </section>

        <section className="mx-auto grid max-w-[1280px] grid-cols-1 gap-[18px] xl:grid-cols-[minmax(0,1.55fr)_minmax(20rem,0.95fr)]">
          <div className="flex flex-col gap-[18px]">
            <Card accent="#265c52">
              <CardHeader
                kicker="Calendario"
                title={<span className="capitalize">{calendarFormatter.format(currentMonth)}</span>}
                right={(
                  <div className="flex items-center gap-1.5">
                    <NavBtn dir="prev" disabled={!canShowPreviousMonth} onClick={() => handleChangeVisibleMonth(-1)} />
                    <NavBtn dir="next" disabled={!canShowNextMonth} onClick={() => handleChangeVisibleMonth(1)} />
                  </div>
                )}
                rightChip={dashboard?.calendar.next_objective ?? "Sin proximo objetivo"}
              />

              <div className="relative z-[2] mt-4 grid grid-cols-7 gap-1.5">
                {weekdayLabels.map((label) => (
                  <p
                    className="text-center [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.18em] text-[#3a332c]"
                    key={label}
                  >
                    {label}
                  </p>
                ))}

                {monthDays.map((day) => {
                  const dateKey = toDateKey(day.date);
                  const dateWorkouts = workoutsByDate.get(dateKey) ?? [];
                  const performedWorkouts = dateWorkouts.filter((workout) => Boolean(workout.performed_at));
                  const plannedWorkouts = dateWorkouts.filter((workout) => !workout.performed_at && Boolean(workout.planned_at));
                  const isTrained = day.isCurrentMonth && trainingDays.has(dateKey);
                  const isPlanned = day.isCurrentMonth && !isTrained && dateKey >= todayKey && plannedDays.has(dateKey);
                  const isPast = dateKey < todayKey;
                  const isInteractive = day.isCurrentMonth && (!isPast || performedWorkouts.length > 0);

                  return (
                    <button
                      key={day.key}
                      type="button"
                      disabled={!isInteractive}
                      onClick={() => handleCalendarDayClick(dateKey, performedWorkouts, plannedWorkouts)}
                      className={[
                        "relative grid aspect-square place-items-center rounded-[10px] border text-[13px] font-bold transition",
                        !day.isCurrentMonth && "border-transparent bg-transparent text-[#1f1b16]/30",
                        day.isCurrentMonth && isTrained && "border-transparent bg-[#265c52] text-[#fffaf0] shadow-[0_6px_14px_rgba(38,92,82,0.30)]",
                        day.isCurrentMonth && !isTrained && isPlanned && "border-[#ea7130]/45 bg-[#ea7130]/12 text-[#1f1b16]",
                        day.isCurrentMonth && !isTrained && !isPlanned && day.isToday && "border-[#1f1b16] bg-[#fffaf0] text-[#1f1b16] shadow-[inset_0_0_0_2px_#ea7130]",
                        day.isCurrentMonth && !isTrained && !isPlanned && !day.isToday && isPast && "border-transparent bg-[#fffaf0]/50 text-[#1f1b16]/45",
                        day.isCurrentMonth && !isTrained && !isPlanned && !day.isToday && !isPast && "border-[#1f1b16]/10 bg-[#fffaf0]/85 text-[#1f1b16]",
                        isInteractive ? "cursor-pointer hover:-translate-y-px hover:border-[#ea7130] hover:shadow-[0_6px_14px_rgba(234,113,48,0.18)]" : "cursor-not-allowed",
                      ].filter(Boolean).join(" ")}
                    >
                      {day.dayNumber}
                      {isPlanned && (
                        <span aria-hidden="true" className="absolute bottom-1 h-1 w-1 rounded-full bg-[#ea7130]" />
                      )}
                      {day.isToday && !isTrained && (
                        <span aria-hidden="true" className="absolute right-1 top-1 h-1 w-1 rounded-full bg-[#265c52]" />
                      )}
                    </button>
                  );
                })}
              </div>

              <div className="relative z-[2] mt-4 flex flex-wrap items-center gap-3 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.14em] text-[#3a332c]">
                <span className="inline-flex items-center gap-1.5">
                  <span className="h-2.5 w-2.5 rounded-[3px] bg-[#265c52]" /> Entrenado
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <span className="h-2.5 w-2.5 rounded-[3px] border border-[#ea7130]/45 bg-[#ea7130]/15" /> Planificado
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <span className="h-2.5 w-2.5 rounded-[3px] border-2 border-[#ea7130] bg-[#fffaf0]" /> Hoy
                </span>
              </div>
            </Card>

            <Card accent="#ea7130">
              <CardHeader kicker="Entrenos" title="Ultimos entrenos" />

              {dashboard?.recent_workouts.length ? (
                <ul className="relative z-[2] mt-4 flex flex-col gap-2.5">
                  {dashboard.recent_workouts.map((workout) => (
                    <li
                      key={workout.id}
                      className="grid grid-cols-[44px_1fr_auto_auto] items-center gap-3.5 rounded-[16px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-3 transition hover:border-[#ea7130]/40 hover:bg-[#fffaf0]"
                    >
                      <div className="grid h-9 w-9 place-items-center rounded-[10px] bg-[#1f1b16] [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold text-[#f1a45b]">
                        <span className="whitespace-pre text-center leading-[1.05]">
                          {formatRecentWorkoutBadgeDate(workout.performed_at)}
                        </span>
                      </div>
                      <div className="min-w-0">
                        <p className="m-0 truncate text-[14px] font-extrabold tracking-tight text-[#1f1b16]">
                          {workout.name || workout.routine_name || "Sesion"}
                        </p>
                        <p className="mt-0.5 truncate text-[12px] font-semibold text-[#3a332c]/75">
                          {workout.routine_name || "Entreno libre"}
                        </p>
                      </div>
                      <span className="justify-self-end inline-flex items-center rounded-md border border-[#265c52]/18 bg-[#265c52]/10 px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-wider text-[#265c52]">
                        {workout.exercise_count} ejc · {workout.duration_minutes}min
                      </span>
                      <span className="justify-self-end [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.14em] text-[#3a332c]">
                        {formatRecentWorkoutMeta(workout.performed_at)}
                      </span>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="relative z-[2] mt-4 rounded-[16px] border border-dashed border-[#1f1b16]/18 bg-transparent p-4 text-[13px] font-semibold leading-[1.5] text-[#3a332c]">
                  Cuando registres sesiones, aqui apareceran tus entrenos mas recientes.
                </p>
              )}
            </Card>

            <Card accent="#6b4ea0">
              <CardHeader
                kicker="Rutinas"
                title="Tus rutinas"
                right={(
                  <Link
                    to="/routines"
                    className="inline-flex cursor-pointer items-center gap-1.5 rounded-[10px] border-0 bg-[#1f1b16] px-3.5 py-2 text-[12px] font-extrabold leading-none tracking-[0.03em] text-[#fffaf0] no-underline shadow-[0_10px_22px_rgba(31,27,22,0.18)] transition hover:-translate-y-px hover:bg-[#2c261f]"
                  >
                    Ver rutinas
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.6} strokeLinecap="round" strokeLinejoin="round" className="h-3 w-3">
                      <path d="M5 12h14M13 6l6 6-6 6" />
                    </svg>
                  </Link>
                )}
              />

              {dashboard?.recent_routines.length ? (
                <ul className="relative z-[2] mt-4 grid gap-2.5 sm:grid-cols-2">
                  {dashboard.recent_routines.map((routine) => (
                    <li
                      key={routine.id}
                      className="rounded-[16px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-4 transition hover:-translate-y-px hover:border-[#6b4ea0]/40 hover:shadow-[0_10px_22px_rgba(31,27,22,0.08)]"
                    >
                      <div className="flex items-center justify-between gap-2">
                        <span className="inline-flex items-center rounded-md bg-[#6b4ea0] px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.16em] text-[#fffaf0]">
                          Rutina
                        </span>
                        <span className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.14em] text-[#3a332c]">
                          {routine.exercise_count} ejercicios
                        </span>
                      </div>
                      <h4 className="m-0 mt-2.5 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[18px] font-black leading-tight tracking-[-0.03em] text-[#1f1b16]">
                        {routine.name}
                      </h4>
                      <p className="mt-1.5 text-[12px] font-semibold leading-[1.5] text-[#3a332c]/80 line-clamp-2">
                        {routine.description || "Rutina lista para retomar cuando quieras."}
                      </p>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="relative z-[2] mt-4 rounded-[16px] border border-dashed border-[#1f1b16]/18 bg-transparent p-4 text-[13px] font-semibold leading-[1.5] text-[#3a332c]">
                  Cuando tengas rutinas guardadas, aqui apareceran las mas recientes.
                </p>
              )}
            </Card>

            <Card accent="#3b8678">
              <CardHeader kicker="Estadisticas" title="Resumen de progreso" />
              <div className="relative z-[2] mt-4 grid gap-2.5 sm:grid-cols-3">
                <MetricTile
                  label="Peso"
                  value={formatMetricValue(dashboard?.progress.weight_kg.current, " kg")}
                  delta={formatMetricDelta(dashboard?.progress.weight_kg.delta, " kg")}
                  trend={dashboard?.progress.weight_kg.delta}
                />
                <MetricTile
                  label="Grasa corporal"
                  value={formatMetricValue(dashboard?.progress.body_fat_percentage.current, "%")}
                  delta={formatMetricDelta(dashboard?.progress.body_fat_percentage.delta, "%")}
                  trend={dashboard?.progress.body_fat_percentage.delta}
                  invertTrend
                />
                <MetricTile
                  label="Masa muscular"
                  value={formatMetricValue(dashboard?.progress.muscle_mass_kg.current, " kg")}
                  delta={formatMetricDelta(dashboard?.progress.muscle_mass_kg.delta, " kg")}
                  trend={dashboard?.progress.muscle_mass_kg.delta}
                />
              </div>
            </Card>
          </div>

          <div className="flex flex-col gap-[18px]">
            <Card accent="#265c52" dark>
              <CardHeader
                kicker="Estadisticas"
                title="Distribucion muscular"
                onDark
                right={(
                  <div
                    role="group"
                    aria-label="Rango de estadisticas"
                    className="inline-flex gap-0.5 rounded-[10px] border border-[#fffaf0]/20 bg-[#fffaf0]/8 p-1"
                  >
                    {(["year", "month"] as const).map((range) => (
                      <button
                        key={range}
                        type="button"
                        onClick={() => setStatsRange(range)}
                        aria-pressed={statsRange === range}
                        className={[
                          "rounded-[7px] px-2.5 py-1.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase leading-none tracking-[0.14em] transition",
                          statsRange === range
                            ? "bg-[#ea7130] text-[#1f1b16]"
                            : "bg-transparent text-[#fffaf0]/70 hover:text-[#fffaf0]",
                        ].join(" ")}
                      >
                        {range === "year" ? "Año" : "Mes"}
                      </button>
                    ))}
                  </div>
                )}
              />

              <div className="relative z-[2] mx-auto mt-4 w-full max-w-[18rem]">
                <svg
                  className="h-auto w-full"
                  viewBox="0 0 216 216"
                  role="img"
                  aria-label="Grafico hexagonal de distribucion muscular"
                >
                  {hexagonLevels.map((level) => {
                    const points = hexagonStats
                      .map((_, index) => polarToCartesian(chartCenter, chartCenter, chartRadius * level, (360 / hexagonStats.length) * index))
                      .map((point) => `${point.x},${point.y}`)
                      .join(" ");

                    return <polygon fill="none" key={level} points={points} stroke="rgba(255,250,240,0.16)" strokeWidth="1.4" />;
                  })}

                  {hexagonStats.map((item, index) => {
                    const outerPoint = polarToCartesian(chartCenter, chartCenter, chartRadius, (360 / hexagonStats.length) * index);
                    const labelPoint = polarToCartesian(chartCenter, chartCenter, chartRadius + 20, (360 / hexagonStats.length) * index);

                    return (
                      <g key={item.axis}>
                        <line stroke="rgba(255,250,240,0.14)" strokeWidth="1.4" x1={chartCenter} x2={outerPoint.x} y1={chartCenter} y2={outerPoint.y} />
                        <text
                          fill="rgba(255,250,240,0.85)"
                          fontFamily="'JetBrains Mono', ui-monospace, monospace"
                          fontSize="10"
                          fontWeight={700}
                          textAnchor="middle"
                          x={labelPoint.x}
                          y={labelPoint.y}
                          style={{ letterSpacing: "0.06em", textTransform: "uppercase" }}
                        >
                          {normalizeMuscleLabel(item.axis)}
                        </text>
                      </g>
                    );
                  })}

                  <polygon fill="rgba(234,113,48,0.28)" points={chartPolygon} stroke="#ea7130" strokeWidth="2.4" strokeLinejoin="round" />
                  {chartPoints.map((point, index) => (
                    <circle key={index} cx={point.x} cy={point.y} r="3.5" fill="#ea7130" stroke="#fffaf0" strokeWidth="1.4" />
                  ))}
                </svg>
              </div>

              <div className="relative z-[2] mt-3 grid grid-cols-2 gap-2">
                <div className="rounded-[14px] border border-[#fffaf0]/10 bg-[#fffaf0]/6 p-3.5">
                  <div className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[28px] font-black leading-none tracking-[-0.04em] text-[#fffaf0]">
                    {selectedMuscleGroupCount}
                  </div>
                  <div className="mt-1 text-sm font-semibold text-[#fffaf0]">
                    Grupos musculares presentes
                  </div>
                </div>
                <div className="rounded-[14px] border border-[#fffaf0]/10 bg-[#fffaf0]/6 p-3.5">
                  <div className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[28px] font-black leading-none tracking-[-0.04em] text-[#fffaf0]">
                    {selectedExerciseCount}
                  </div>
                  <div className="mt-1 text-sm font-semibold text-[#fffaf0]">
                    {statsRange === "year" ? "Ultimo año" : "Ultimo mes"} · {selectedExerciseCount} ejercicios considerados
                  </div>
                </div>
              </div>
            </Card>

            <Card accent="#265c52">
              <CardHeader kicker="Top grupos" title="Reparto" />
              {selectedBreakdown.length > 0 ? (
                <ul className="relative z-[2] mt-4 flex flex-col gap-2.5">
                  {selectedBreakdown.slice(0, 5).map((group) => (
                    <li
                      key={group.name}
                      className="rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-3.5"
                    >
                      <div className="flex items-center justify-between gap-3">
                        <div>
                          <p className="m-0 text-[13px] font-extrabold tracking-tight text-[#1f1b16]">{normalizeMuscleLabel(group.name)}</p>
                          <p className="mt-0.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.14em] text-[#3a332c]">
                            {group.count} ejercicios
                          </p>
                        </div>
                        <span className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[20px] font-black leading-none tracking-[-0.03em] text-[#265c52]">
                          {group.percentage}%
                        </span>
                      </div>
                      <div className="mt-2.5 h-1.5 overflow-hidden rounded-full bg-[#1f1b16]/8">
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
                <p className="relative z-[2] mt-4 rounded-[14px] border border-dashed border-[#1f1b16]/18 p-4 text-[13px] font-semibold leading-[1.5] text-[#3a332c]">
                  Cuando haya historial de entrenos, aqui veras que grupos musculares predominan.
                </p>
              )}
            </Card>
          </div>
        </section>
      </div>

      {selectedPlanDate && (
        <DialogShell
          kicker="Planificar entreno"
          kickerColor="#ea7130"
          title={selectedPlanDate}
          onClose={handleCloseCalendarDialog}
        >
          <label
            htmlFor="routine-select"
            className="mt-2 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-[0.16em] text-[#3a332c]"
          >
            Rutina
          </label>
          <select
            id="routine-select"
            value={selectedRoutineID}
            onChange={(event) => setSelectedRoutineID(event.target.value)}
            className="mt-2 w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/85 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15"
          >
            <option value="">Selecciona una rutina</option>
            {routines.map((routine) => (
              <option key={routine.id} value={routine.id}>
                {routine.name}
              </option>
            ))}
          </select>

          <label
            htmlFor="planned-time"
            className="mt-4 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-[0.16em] text-[#3a332c]"
          >
            Hora estimada
          </label>
          <input
            id="planned-time"
            type="time"
            value={selectedPlanTime}
            onChange={(event) => setSelectedPlanTime(event.target.value)}
            className="mt-2 w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/85 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15"
          />

          {planStatus === "error" && (
            <p className="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-3 py-2 text-[12px] font-bold text-[#9f2f22]">
              No se ha podido planificar el entreno.
            </p>
          )}

          <div className="mt-5 flex justify-end gap-2.5">
            <button
              type="button"
              onClick={handleCloseCalendarDialog}
              className="cursor-pointer rounded-[14px] border border-[#1f1b16]/18 bg-transparent px-4 py-3 text-[13px] font-extrabold tracking-[0.04em] text-[#1f1b16] transition hover:bg-[#1f1b16]/5"
            >
              Cancelar
            </button>
            <button
              type="button"
              onClick={handlePlanWorkout}
              disabled={planStatus === "loading" || !selectedRoutineID || !selectedPlanTime}
              className="group relative cursor-pointer overflow-hidden rounded-[14px] bg-[#ea7130] px-5 py-3 text-[13px] font-extrabold tracking-[0.04em] text-[#1f1b16] shadow-[0_18px_35px_rgba(234,113,48,0.30)] transition hover:-translate-y-px hover:bg-[#ff8b47] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:translate-y-0"
            >
              <span className="relative z-10">
                {planStatus === "loading" ? "Guardando..." : "Guardar entreno"}
              </span>
              <span className="absolute inset-y-0 -left-1/3 w-1/3 -skew-x-[20deg] bg-white/30 transition-[left] duration-500 group-hover:left-[120%]" />
            </button>
          </div>
        </DialogShell>
      )}

      {calendarDialog?.type === "planned" && (
        <DialogShell
          kicker="Entreno planificado"
          kickerColor="#ea7130"
          title={calendarDialog.dateKey}
          onClose={handleCloseCalendarDialog}
        >
          <ul className="mt-2 flex flex-col gap-2.5">
            {calendarDialog.workouts.map((workout) => {
              const googleCalendarLink = buildGoogleCalendarLink(workout);

              return (
              <li
                key={workout.id}
                className="rounded-[16px] border border-[#1f1b16]/12 bg-[#fffaf0]/85 p-4"
              >
                <p className="m-0 text-[14px] font-extrabold tracking-tight text-[#1f1b16]">
                  {workoutTitle(workout)}
                </p>
                <p className="mt-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.16em] text-[#3a332c]">
                  {workout.routine_name || "Entreno libre"}
                </p>
                {workout.planned_at && (
                  <p className="mt-2 text-[13px] font-semibold text-[#3a332c]">
                    Hora estimada:{" "}
                    <strong className="text-[#1f1b16]">
                      {workoutTimeFormatter.format(new Date(workout.planned_at))}
                    </strong>
                  </p>
                )}
                <div className="mt-3 flex flex-wrap gap-2">
                  {googleCalendarLink && (
                    <a
                      href={googleCalendarLink}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex cursor-pointer rounded-[12px] border border-[#265c52]/25 bg-[#265c52]/10 px-3.5 py-2 text-[12px] font-extrabold tracking-[0.04em] text-[#265c52] transition hover:bg-[#265c52]/15"
                    >
                      Añadir a Calendar
                    </a>
                  )}
                  <button
                    type="button"
                    onClick={() => handleCancelPlannedWorkout(workout.id)}
                    disabled={calendarActionStatus === "loading"}
                    className="cursor-pointer rounded-[12px] border border-[#9f2f22]/25 bg-[#9f2f22]/10 px-3.5 py-2 text-[12px] font-extrabold tracking-[0.04em] text-[#9f2f22] transition hover:bg-[#9f2f22]/15 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    Cancelar entreno
                  </button>
                </div>
              </li>
              );
            })}
          </ul>
          {calendarActionStatus === "error" && (
            <p className="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-3 py-2 text-[12px] font-bold text-[#9f2f22]">
              No se ha podido cancelar el entreno.
            </p>
          )}
        </DialogShell>
      )}

      {calendarDialog?.type === "performed" && (
        <DialogShell
          kicker="Entreno realizado"
          kickerColor="#265c52"
          title={calendarDialog.dateKey}
          onClose={handleCloseCalendarDialog}
        >
          <ul className="mt-2 flex flex-col gap-2.5">
            {calendarDialog.workouts.map((workout) => (
              <li
                key={workout.id}
                className="rounded-[16px] border border-[#1f1b16]/12 bg-[#fffaf0]/85 p-4"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="m-0 text-[14px] font-extrabold tracking-tight text-[#1f1b16]">
                      {workoutTitle(workout)}
                    </p>
                    <p className="mt-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.16em] text-[#3a332c]">
                      {workout.routine_name || "Entreno libre"}
                    </p>
                  </div>
                  {workout.performed_at && (
                    <span className="rounded-md border border-[#265c52]/20 bg-[#265c52]/10 px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.14em] text-[#265c52]">
                      {workoutTimeFormatter.format(new Date(workout.performed_at))}
                    </span>
                  )}
                </div>
                <p className="mt-3 text-[13px] font-semibold text-[#3a332c]">
                  <strong className="text-[#1f1b16]">{workout.exercise_count}</strong> ejercicios ·{" "}
                  <strong className="text-[#1f1b16]">{workout.duration_minutes}</strong> min
                </p>
              </li>
            ))}
          </ul>
        </DialogShell>
      )}
    </div>
  );
}

function Stat({ n, l, accent }: { n: string; l: string; accent?: boolean }) {
  return (
    <div
      className={[
        "min-w-[120px] rounded-[18px] border px-4 py-3 backdrop-blur",
        accent
          ? "border-[#1f1b16] bg-[#1f1b16] text-[#fffaf0]"
          : "border-[#1f1b16]/12 bg-[#fffaf0]/75",
      ].join(" ")}
    >
      <div className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[32px] font-black leading-none tracking-[-0.04em]">
        {n}
      </div>
      <div
        className={[
          "mt-1.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-[0.16em]",
          accent ? "text-[#f1a45b]" : "text-[#3a332c]",
        ].join(" ")}
      >
        {l}
      </div>
    </div>
  );
}

function Card({
  children,
  accent = "#ea7130",
  dark,
}: {
  children: ReactNode;
  accent?: string;
  dark?: boolean;
}) {
  return (
    <article
      style={{ ["--accent" as never]: accent } as CSSProperties}
      className={[
        "group relative isolate overflow-hidden rounded-[24px] border p-5 backdrop-blur sm:p-6",
        dark
          ? "border-[#1f1b16] bg-[#1f1b16] text-[#fffaf0]"
          : "border-[#1f1b16]/12 bg-[#fffaf0]/85",
      ].join(" ")}
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

function CardHeader({
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
    <header className="relative z-[2] flex flex-wrap items-end justify-between gap-3">
      <div>
        <div
          className={[
            "mb-1.5 flex items-center gap-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.22em]",
            onDark ? "text-[#f1a45b]" : "text-[#265c52]",
          ].join(" ")}
        >
          <span
            className="inline-block h-0.5 w-4"
            style={{ background: onDark ? "#f1a45b" : "#265c52" }}
          />
          {kicker}
        </div>
        <h3
          className={[
            "m-0 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[22px] font-black leading-[1.05] tracking-[-0.035em]",
            onDark ? "text-[#fffaf0]" : "text-[#1f1b16]",
          ].join(" ")}
        >
          {title}
        </h3>
        {rightChip && (
          <div
            className={[
              "mt-1.5 inline-flex max-w-full items-center gap-1.5 rounded-md border px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.14em]",
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

function NavBtn({
  dir,
  onClick,
  disabled,
}: {
  dir: "prev" | "next";
  onClick: () => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      aria-label={dir === "prev" ? "Mes anterior" : "Mes siguiente"}
      disabled={disabled}
      onClick={onClick}
      className="grid h-9 w-9 cursor-pointer place-items-center rounded-[10px] border border-[#1f1b16]/12 bg-[#fffaf0]/80 text-[#1f1b16] transition hover:border-[#ea7130] hover:text-[#ea7130] disabled:cursor-not-allowed disabled:opacity-40"
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2.4}
        strokeLinecap="round"
        strokeLinejoin="round"
        className="h-4 w-4"
      >
        {dir === "prev" ? <path d="M15 6l-6 6 6 6" /> : <path d="M9 6l6 6-6 6" />}
      </svg>
    </button>
  );
}

function MetricTile({
  label,
  value,
  delta,
  trend,
  invertTrend,
}: {
  label: string;
  value: string;
  delta: string;
  trend?: number | null;
  invertTrend?: boolean;
}) {
  const direction =
    typeof trend === "number" && trend !== 0
      ? (trend > 0) !== !!invertTrend
        ? "up"
        : "down"
      : "flat";

  const color =
    direction === "up" ? "#265c52" : direction === "down" ? "#9f2f22" : "#3a332c";

  return (
    <div className="rounded-[16px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-4">
      <p className="m-0 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.16em] text-[#3a332c]">
        {label}
      </p>
      <p className="mt-1.5 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[26px] font-black leading-none tracking-[-0.04em] text-[#1f1b16]">
        {value}
      </p>
      <p
        className="mt-2 inline-flex items-center gap-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold tracking-wider"
        style={{ color }}
      >
        {direction === "up" && <span>▲</span>}
        {direction === "down" && <span>▼</span>}
        {direction === "flat" && <span>-</span>}
        {delta}
      </p>
    </div>
  );
}

function DialogShell({
  kicker,
  kickerColor = "#265c52",
  title,
  onClose,
  children,
}: {
  kicker: string;
  kickerColor?: string;
  title: string;
  onClose: () => void;
  children: ReactNode;
}) {
  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-[#1f1b16]/45 px-4 backdrop-blur-sm">
      <section className="w-full max-w-md overflow-hidden rounded-[24px] border border-[#1f1b16]/12 bg-[#fffaf0] shadow-[0_30px_80px_rgba(31,27,22,0.30),0_8px_22px_rgba(31,27,22,0.12)]">
        <header className="grid grid-cols-[1fr_auto] items-center gap-4 bg-[#1f1b16] px-6 py-5 text-[#fffaf0]">
          <div>
            <div
              className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-extrabold uppercase tracking-[0.30em]"
              style={{ color: kickerColor === "#265c52" ? "#3b8678" : "#f1a45b" }}
            >
              {kicker}
            </div>
            <h3 className="mt-1 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[26px] font-black leading-none tracking-[-0.04em]">
              {title}
            </h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Cerrar"
            className="grid h-9 w-9 cursor-pointer place-items-center rounded-[10px] border border-[#fffaf0]/20 bg-transparent text-[#fffaf0] transition hover:rotate-90 hover:bg-[#fffaf0]/10"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" className="h-4 w-4">
              <path d="M6 6l12 12M18 6L6 18" />
            </svg>
          </button>
        </header>
        <div className="px-6 py-5">{children}</div>
      </section>
    </div>
  );
}
