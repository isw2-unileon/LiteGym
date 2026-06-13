import { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useOutletContext } from "react-router-dom";
import type { LayoutUser } from "../components/AppLayout";
import { Card, CardHeader } from "../components/Card";
import { apiUrl } from "../lib/api";
import { useIsMobile } from "../lib/useIsMobile";
import { HelloHeader } from "../components/HelloHeader";
import { Stat } from "../components/Stat";
import { MuscleDistribution } from "../components/MuscleDistribution";
import {DialogPopup} from "@/components/DialogPopup.tsx";

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
    return "Aún no hay comparativa";
  }

  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(1)}${suffix}`;
}

function cleanDescription(description?: string | null) {
  if (!description) return "";
  return description
    .split("\n")
    .filter(
      (line) =>
        !line.startsWith("[Objetivo] ") && !line.startsWith("[Grupos musculares] "),
    )
    .join("\n")
    .trim();
}

export default function DashboardPage() {
  const { user } = useOutletContext<OutletContext>();
  const navigate = useNavigate();
  const isMobile = useIsMobile();
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
  const [quickStartOpen, setQuickStartOpen] = useState(false);
  const [quickStartStatus, setQuickStartStatus] = useState<"idle" | "loading" | "error">("idle");
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
    setQuickStartStatus("idle");
  };

  const handleOpenQuickStart = () => {
    setQuickStartOpen(true);
    setSelectedRoutineID("");
    setQuickStartStatus("idle");
    void fetchRoutines();
  };

  const handleCloseQuickStart = () => {
    setQuickStartOpen(false);
    setSelectedRoutineID("");
    setQuickStartStatus("idle");
  };

  const handleCalendarDayClick = (
    dateKey: string,
    performedWorkouts: DashboardCalendarWorkout[],
    plannedWorkouts: DashboardCalendarWorkout[],
  ) => {
    if (plannedWorkouts.length > 0) {
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

  const handleStartWorkoutFromRoutine = async () => {
    if (!selectedRoutineID) {
      setQuickStartStatus("error");
      return;
    }

    const routine = routines.find((item) => item.id === selectedRoutineID);
    setQuickStartStatus("loading");

    try {
      const response = await fetch(apiUrl("/api/workout/start"), {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          routine_id: selectedRoutineID,
          name: routine?.name ?? "Entrenamiento",
          performed_at: new Date().toISOString(),
        }),
      });

      if (!response.ok) {
        setQuickStartStatus("error");
        return;
      }

      const payload = (await response.json()) as { id: string };
      setQuickStartOpen(false);
      navigate(`/workouts/${payload.id}/fill`);
    } catch {
      setQuickStartStatus("error");
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

  return (
    <main className="relative isolate overflow-x-hidden text-[#1f1b16] [font-family:'Inter',system-ui,sans-serif] antialiased pb-20">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]"
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10"
      />

      <div className="px-4 pt-5 sm:px-6 sm:pt-8 md:px-8">
        <section className="mx-auto mb-6 grid max-w-[1280px] grid-cols-1 items-start gap-6 md:grid-cols-[1fr_auto]">
          <div>
          <HelloHeader page={"PANEL PRINCIPAL"} user={user?.username ?? "Atleta"} />
            <p className="mt-3.5 max-w-[640px] text-[15px] leading-[1.55] text-[#3a332c]">
              Aquí ves tu actividad reciente y los días del mes en los que ya has entrenado. Planifica un día tocando una celda futura.
            </p>
            <div className="mt-5 grid grid-cols-2 gap-3 md:hidden">
              <button
                type="button"
                onClick={handleOpenQuickStart}
                className="rounded-[18px] bg-[#1f1b16] px-4 py-4 text-left text-sm font-black text-[#f1a45b] shadow-[0_14px_30px_rgba(31,27,22,0.18)]"
              >
                <span className="block text-[11px] uppercase tracking-[0.16em] text-[#fffaf0]/70">Ahora</span>
                Iniciar entreno
              </button>
              <Link
                to="/routines"
                className="rounded-[18px] border border-[#1f1b16]/12 bg-[#fffaf0]/85 px-4 py-4 text-left text-sm font-black text-[#1f1b16] shadow-[0_12px_28px_rgba(31,27,22,0.08)]"
              >
                <span className="block text-[11px] uppercase tracking-[0.16em] text-[#265c52]">Plan</span>
                Ver rutinas
              </Link>
            </div>
            {status === "error" && (
              <p className="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-3 py-2 text-[12px] font-bold text-[#9f2f22]">
                No se ha podido cargar toda la información del dashboard.
              </p>
            )}
          </div>

          <div className="grid w-full grid-cols-3 gap-2.5 md:flex md:w-auto md:flex-wrap">
            <Stat n={String(dashboard?.calendar.sessions_count ?? 0)} l="Sesiones del mes" />
            <Stat n={String(dashboard?.calendar.weekly_goal ?? 2)} l="Objetivo semanal" />
            <Stat n={`${dashboard?.calendar.current_streak ?? 0}`} l="Racha semanal" accent />
          </div>
        </section>

        {isMobile && (
        <section className="mx-auto grid max-w-[1280px] gap-4">
          <section className="overflow-hidden rounded-[30px] border border-[#1f1b16]/10 bg-[#fffaf0]/86 shadow-[0_14px_34px_rgba(31,27,22,0.1)]">
            <div className="flex items-end justify-between gap-3 bg-[#ea7130] px-5 py-4">
              <div className="min-w-0">
                <p className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.18em] text-[#1f1b16]/70">
                  Mes actual
                </p>
                <h2 className="truncate [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[30px] font-black capitalize leading-none tracking-[-0.055em] text-[#1f1b16]">
                  {calendarFormatter.format(currentMonth)}
                </h2>
              </div>
            </div>
            <div className="p-4">
            <div className="mb-3 flex justify-end gap-1.5">
              <NavBtn dir="prev" disabled={!canShowPreviousMonth} onClick={() => handleChangeVisibleMonth(-1)} />
              <NavBtn dir="next" disabled={!canShowNextMonth} onClick={() => handleChangeVisibleMonth(1)} />
            </div>

            <div className="relative z-[2] mt-4 grid grid-cols-7 gap-1">
              {weekdayLabels.map((label) => (
                <p
                  className="text-center [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[9px] font-black uppercase tracking-[0.04em] text-[#3a332c]/80"
                  key={label}
                >
                  {label.slice(0, 1)}
                </p>
              ))}

              {monthDays.map((day) => {
                const dateKey = toDateKey(day.date);
                const dateWorkouts = workoutsByDate.get(dateKey) ?? [];
                const performedWorkouts = dateWorkouts.filter((workout) => Boolean(workout.performed_at));
                const plannedWorkouts = dateWorkouts.filter((workout) => !workout.performed_at && Boolean(workout.planned_at));
                const isTrained = day.isCurrentMonth && trainingDays.has(dateKey);
                const isPlanned = day.isCurrentMonth && !isTrained && plannedWorkouts.length > 0;
                const isPast = dateKey < todayKey;
                const isInteractive = day.isCurrentMonth && (!isPast || performedWorkouts.length > 0 || plannedWorkouts.length > 0);

                return (
                  <button
                    key={day.key}
                    type="button"
                    disabled={!isInteractive}
                    onClick={() => handleCalendarDayClick(dateKey, performedWorkouts, plannedWorkouts)}
                    className={[
                      "relative grid aspect-square place-items-center rounded-[9px] border text-[12px] font-black transition",
                      !day.isCurrentMonth && "border-transparent bg-transparent text-[#1f1b16]/25",
                      day.isCurrentMonth && isTrained && "border-transparent bg-[#265c52] text-[#fffaf0]",
                      day.isCurrentMonth && !isTrained && isPlanned && "border-[#ea7130]/45 bg-[#ea7130]/12 text-[#1f1b16]",
                      day.isCurrentMonth && !isTrained && !isPlanned && day.isToday && "border-[#1f1b16] bg-[#fffaf0] text-[#1f1b16]",
                      day.isCurrentMonth && !isTrained && !isPlanned && !day.isToday && isPast && "border-transparent bg-[#fffaf0]/50 text-[#1f1b16]/45",
                      day.isCurrentMonth && !isTrained && !isPlanned && !day.isToday && !isPast && "border-[#1f1b16]/10 bg-[#fffaf0]/85 text-[#1f1b16]",
                    ].filter(Boolean).join(" ")}
                  >
                    {day.dayNumber}
                    {isPlanned && <span aria-hidden="true" className="absolute bottom-1 h-1 w-1 rounded-full bg-[#ea7130]" />}
                  </button>
                );
              })}
            </div>
            </div>
          </section>

          <MobileDashboardPanel kicker="Próximo" title="Actividad reciente">
            <div className="grid gap-3">
              {(dashboard?.recent_workouts ?? []).slice(0, 3).map((workout) => (
                <div key={workout.id} className="rounded-[16px] border border-[#1f1b16]/10 bg-white/65 p-3">
                  <p className="m-0 text-[15px] font-black text-[#1f1b16]">{workout.name || workout.routine_name || "Sesion"}</p>
                  <p className="mt-1 text-[12px] font-bold text-[#3a332c]/75">
                    {workout.routine_name || "Entreno libre"} · {workout.duration_minutes} min · {workout.exercise_count} ejercicios
                  </p>
                </div>
              ))}
              {(dashboard?.recent_workouts ?? []).length === 0 && (
                <p className="rounded-[16px] border border-dashed border-[#1f1b16]/15 p-4 text-sm font-bold text-[#3a332c]/70">
                  Aún no hay entrenos registrados.
                </p>
              )}
            </div>
          </MobileDashboardPanel>

          <MobileDashboardPanel kicker="Rutinas" title="Acceso rápido">
            <Link to="/routines" className="mb-3 inline-flex rounded-[12px] bg-[#1f1b16] px-3 py-2 text-xs font-black text-[#f1a45b]">
              Ver todas
            </Link>
            <div className="grid gap-3">
              {(dashboard?.recent_routines ?? []).slice(0, 3).map((routine) => (
                <div key={routine.id} className="rounded-[16px] border border-[#1f1b16]/10 bg-white/65 p-3">
                  <p className="m-0 text-[15px] font-black text-[#1f1b16]">{routine.name}</p>
                  <p className="mt-1 text-[12px] font-bold text-[#3a332c]/75">{routine.exercise_count} ejercicios</p>
                </div>
              ))}
            </div>
          </MobileDashboardPanel>
        </section>
        )}

        {!isMobile && (
        <section className="mx-auto grid max-w-[1280px] grid-cols-1 gap-[18px] xl:grid-cols-[minmax(0,1.55fr)_minmax(20rem,0.95fr)]">
          <div className="flex flex-col gap-[18px]">
            <Card accent="#ea7130">
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
                    className="text-center [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.08em] text-[#3a332c] sm:text-[12px] sm:tracking-[0.18em]"
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
                  const isPlanned = day.isCurrentMonth && !isTrained && plannedWorkouts.length > 0;
                  const isPast = dateKey < todayKey;
                  const isInteractive = day.isCurrentMonth && (!isPast || performedWorkouts.length > 0 || plannedWorkouts.length > 0);

                  return (
                    <button
                      key={day.key}
                      type="button"
                      disabled={!isInteractive}
                      onClick={() => handleCalendarDayClick(dateKey, performedWorkouts, plannedWorkouts)}
                      className={[
                        "relative grid aspect-square place-items-center rounded-[10px] border text-[13px] font-bold transition sm:text-[16px]",
                        !day.isCurrentMonth && "border-transparent bg-transparent text-[#1f1b16]/30",
                        day.isCurrentMonth && isTrained && "border-transparent bg-[#265c52] text-[#fffaf0] shadow-[0_6px_14px_rgba(38,92,82,0.30)]",
                        day.isCurrentMonth && !isTrained && isPlanned && "border-[#ea7130]/45 bg-[#ea7130]/12 text-[#1f1b16]",
                        day.isCurrentMonth && !isTrained && !isPlanned && day.isToday && "border-[#1f1b16] bg-[#fffaf0] text-[#1f1b16]",
                        day.isCurrentMonth && !isTrained && !isPlanned && !day.isToday && isPast && "border-transparent bg-[#fffaf0]/50 text-[#1f1b16]/45",
                        day.isCurrentMonth && !isTrained && !isPlanned && !day.isToday && !isPast && "border-[#1f1b16]/10 bg-[#fffaf0]/85 text-[#1f1b16]",
                        isInteractive ? "cursor-pointer hover:-translate-y-px hover:border-[#ea7130] hover:shadow-[0_6px_14px_rgba(234,113,48,0.18)]" : "cursor-not-allowed",
                      ].filter(Boolean).join(" ")}
                    >
                      {day.dayNumber}
                      {day.isToday && day.isCurrentMonth && (
                        <span aria-hidden="true" className="pointer-events-none absolute inset-0 rounded-[10px] ring-2 ring-inset ring-[#ea7130]" />
                      )}
                      {isPlanned && (
                        <span aria-hidden="true" className="absolute bottom-1 h-1.5 w-1.5 rounded-full bg-[#ea7130]" />
                      )}
                      {day.isToday && !isTrained && (
                        <span aria-hidden="true" className="absolute right-1 top-1 h-1.5 w-1.5 rounded-full bg-[#265c52]" />
                      )}
                    </button>
                  );
                })}
              </div>

              <div className="relative z-[2] mt-4 flex flex-wrap items-center gap-3 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.14em] text-[#3a332c]">
                <span className="inline-flex items-center gap-1.5"><span className="h-2.5 w-2.5 rounded-[3px] bg-[#265c52]" /> Entrenado</span>
                <span className="inline-flex items-center gap-1.5"><span className="h-2.5 w-2.5 rounded-[3px] border border-[#ea7130]/45 bg-[#ea7130]/15" /> Planificado</span>
                <span className="inline-flex items-center gap-1.5"><span className="h-2.5 w-2.5 rounded-[3px] border-2 border-[#ea7130] bg-[#fffaf0]" /> Hoy</span>
              </div>
            </Card>

            <Card accent="#ea7130">
              <CardHeader kicker="Entrenos" title="Últimos entrenos" />
              {dashboard?.recent_workouts.length && dashboard.recent_workouts.length > 0 ? (
                <ul className="relative z-[2] mt-4 flex flex-col gap-2.5">
                  {dashboard.recent_workouts.map((workout) => (
                    <li
                      key={workout.id}
                      className="grid grid-cols-[44px_minmax(0,1fr)] items-center gap-3 rounded-[16px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-3 transition hover:border-[#ea7130]/40 hover:bg-[#fffaf0] sm:grid-cols-[44px_1fr_auto_auto] sm:gap-3.5"
                    >
                      <div className="grid h-9 w-9 place-items-center rounded-[10px] bg-[#1f1b16] [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[14px] font-bold text-[#f1a45b]">
                        <span className="whitespace-pre text-center leading-[1.05]">
                          {formatRecentWorkoutBadgeDate(workout.performed_at)}
                        </span>
                      </div>
                      <div className="min-w-0">
                        <p className="m-0 truncate text-[16px] font-extrabold tracking-tight text-[#1f1b16]">
                          {workout.name || workout.routine_name || "Sesion"}
                        </p>
                        <p className="mt-0.5 truncate text-[14px] font-semibold text-[#3a332c]/75">
                          {workout.routine_name || "Entreno libre"}
                        </p>
                      </div>
                      <span className="col-span-2 inline-flex items-center justify-self-start rounded-md border border-[#265c52]/18 bg-[#265c52]/10 px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-wider text-[#265c52] sm:col-span-1 sm:justify-self-end">
                        {workout.duration_minutes} {workout.duration_minutes === 1 ? "minuto": "minutos"} · {workout.exercise_count} {workout.exercise_count === 1 ? "ejercicio": "ejercicios"}
                      </span>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="relative z-[2] mt-4 rounded-[16px] border border-dashed border-[#1f1b16]/18 bg-transparent p-4 text-[13px] font-semibold leading-[1.5] text-[#3a332c]">
                  Cuando registres sesiones, aquí apareceran tus entrenos mas recientes.
                </p>
              )}
            </Card>

            <Card accent="#ea7130">
              <CardHeader
                kicker="Rutinas"
                title="Tus rutinas"
                right={(
                  <Link
                    to="/routines"
                    className="inline-flex w-full cursor-pointer items-center justify-center gap-1.5 rounded-[10px] border-0 bg-[#1f1b16] px-3.5 py-4 text-[14px] font-extrabold leading-none tracking-[0.03em] text-[#f1a45b] no-underline shadow-[0_10px_22px_rgba(31,27,22,0.18)] transition hover:-translate-y-px hover:bg-[#2c261f] sm:w-auto"
                  >
                    Ver todas tus rutinas
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
                      className="rounded-[16px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-4 transition hover:-translate-y-px hover:border-[#ea7130]/40 hover:shadow-[0_10px_22px_rgba(31,27,22,0.08)]"
                    >
                      <div className="flex items-center justify-between gap-2">
                        <span className="inline-flex items-center rounded-md bg-[#1f1b16] px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.16em] text-[#f1a45b]">
                          Rutina
                        </span>
                        <span className="justify-self-end inline-flex items-center rounded-md border border-[#265c52]/18 bg-[#265c52]/10 px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-wider text-[#265c52]">
                        {routine.exercise_count} {routine.exercise_count === 1 ? "ejercicio": "ejercicios"}
                      </span>
                      </div>
                      <h4 className="m-0 mt-2.5 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[18px] font-black leading-tight tracking-[-0.03em] text-[#1f1b16]">
                        {routine.name}
                      </h4>
                      <p className="mt-1.5 text-[14px] font-semibold leading-[1.5] text-[#3a332c]/80 line-clamp-2">
                        {cleanDescription(routine.description) || "Esta rutina no tiene descripción."}
                      </p>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="relative z-[2] mt-4 rounded-[16px] border border-dashed border-[#1f1b16]/18 bg-transparent p-4 text-[13px] font-semibold leading-[1.5] text-[#3a332c]">
                  Cuando tengas rutinas guardadas, aquí apareceran las mas recientes.
                </p>
              )}
            </Card>
          </div>

          <div className="flex flex-col gap-[18px]">
            <MuscleDistribution
                data={dashboard?.muscle_distribution}
                range={statsRange}
                onRangeChange={setStatsRange}
                summary={true}
            />

            <Card accent="#ea7130">
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

            <Card accent="#ea7130" className={"flex-1"}>
              <CardHeader kicker="Top grupos" title="Reparto" />
              {selectedBreakdown.length > 0 ? (
                <ul className="relative z-[2] mt-4 flex flex-col gap-2.5">
                  {selectedBreakdown.slice(0, 8).map((group) => (
                    <li
                      key={group.name}
                      className="rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-3.5"
                    >
                      <div className="flex items-center justify-between gap-3">
                        <div>
                          <p className="m-0 text-[14px] font-extrabold tracking-tight text-[#1f1b16]">{normalizeMuscleLabel(group.name)}</p>

                          <span className="justify-self-end inline-flex items-center rounded-md border border-[#265c52]/18 bg-[#265c52]/10 px-1 py-0.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-wider text-[#265c52]">
                            {group.count} {group.count === 1 ? "ejercicio" : "ejercicios"}
                      </span>
                        </div>
                        <span className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[20px] font-black leading-none tracking-[-0.03em] text-[#1f1b16]">
                          {group.percentage}%
                        </span>
                      </div>
                      <div className="mt-2.5 h-1.5 overflow-hidden rounded-full bg-[#1f1b16]/8">
                        <div
                          aria-hidden="true"
                          className="h-full rounded-full bg-[#1f1b16]"
                          style={{ width: `${group.percentage}%` }}
                        />
                      </div>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="relative z-[2] mt-4 rounded-[14px] border border-dashed border-[#1f1b16]/18 p-4 text-[13px] font-semibold leading-[1.5] text-[#3a332c]">
                  Cuando haya historial de entrenos, aquí veras que grupos musculares predominan.
                </p>
              )}
            </Card>
          </div>
        </section>
        )}
      </div>

      <button
        type="button"
        onClick={handleOpenQuickStart}
        className="fixed bottom-8 right-8 z-30 hidden items-center gap-3 rounded-full bg-[#ea7130] px-5 py-4 text-sm font-extrabold tracking-[0.04em] text-[#1f1b16] shadow-[0_24px_48px_rgba(234,113,48,0.35)] transition hover:-translate-y-px hover:bg-[#ff8b47] md:inline-flex"
      >
        <span className="grid h-9 w-9 place-items-center rounded-full bg-[#1f1b16] text-[#fffaf0]">+</span>
        Iniciar entrenamiento
      </button>

      {selectedPlanDate && (
        <DialogPopup
          kicker="Planificar entreno"
          kickerColor="#ea7130"
          title={selectedPlanDate}
          onClose={handleCloseCalendarDialog}
        >
          <label
            htmlFor="routine-select"
            className="mt-2 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[14px] font-bold uppercase tracking-[0.16em] text-[#3a332c]"
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
            className="mt-4 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[14px] font-bold uppercase tracking-[0.16em] text-[#3a332c]"
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
              className="cursor-pointer rounded-[14px] border border-[#1f1b16]/18 bg-transparent px-4 py-3 text-[13px] font-extrabold tracking-[0.04em] text-[#9f2f22] transition hover:bg-[#1f1b16]/5"
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
            </button>
          </div>
        </DialogPopup>
      )}

      {quickStartOpen && (
        <DialogPopup
          kicker="Nuevo entrenamiento"
          kickerColor="#ea7130"
          title="Elegir rutina"
          onClose={handleCloseQuickStart}
        >
          <label
            htmlFor="quick-start-routine"
            className="mt-2 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[14px] font-bold uppercase tracking-[0.16em] text-[#3a332c]"
          >
            Rutina
          </label>
          <select
            id="quick-start-routine"
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

          {quickStartStatus === "error" && (
            <p className="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-3 py-2 text-[12px] font-bold text-[#9f2f22]">
              No se ha podido iniciar el entrenamiento.
            </p>
          )}

          <div className="mt-5 flex justify-end gap-2.5">
            <button
              type="button"
              onClick={handleCloseQuickStart}
              className="cursor-pointer rounded-[14px] border border-[#1f1b16]/18 bg-transparent px-4 py-3 text-[13px] font-extrabold tracking-[0.04em] text-[#9f2f22] transition hover:bg-[#1f1b16]/5"
            >
              Cancelar
            </button>
            <button
              type="button"
              onClick={handleStartWorkoutFromRoutine}
              disabled={quickStartStatus === "loading" || !selectedRoutineID}
              className="cursor-pointer rounded-[14px] bg-[#ea7130] px-5 py-3 text-[13px] font-extrabold tracking-[0.04em] text-[#1f1b16] shadow-[0_18px_35px_rgba(234,113,48,0.30)] transition hover:-translate-y-px hover:bg-[#ff8b47] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:translate-y-0"
            >
              {quickStartStatus === "loading" ? "Iniciando..." : "Empezar"}
            </button>
          </div>
        </DialogPopup>
      )}

      {calendarDialog?.type === "planned" &&  (
        <DialogPopup
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
                <p className="m-0 text-[16px] uppercase font-extrabold tracking-tight text-[#1f1b16]">
                  {workoutTitle(workout)}
                </p>
                {workout.planned_at && (
                  <p className="mt-2 text-[14px] font-semibold text-[#3a332c]">
                    Hora estimada:{" "}
                    <strong className="text-[#1f1b16]">
                      {workoutTimeFormatter.format(new Date(workout.planned_at))}
                    </strong>
                  </p>
                )}
                <div className="mt-3 flex flex-wrap gap-2">
                  {googleCalendarLink && calendarDialog.dateKey >= todayKey && (
                      <a
                          href={googleCalendarLink}
                          target="_blank"
                          rel="noreferrer"
                          className="inline-flex cursor-pointer rounded-[12px] border border-[#265c52]/25 bg-[#265c52]/10 px-3.5 py-2 text-[12px] font-extrabold tracking-[0.04em] text-[#265c52] transition hover:bg-[#265c52]/15"
                      >
                        Añadir a Calendar
                      </a>
                  )}
                  {calendarDialog.dateKey === todayKey && (
                      <button
                          type="button"
                          onClick={() => navigate(`/workouts/${workout.id}/fill?source=planned`)}
                          className="inline-flex cursor-pointer rounded-[12px] border border-[#ea7130]/25 bg-[#ea7130]/10 px-3.5 py-2 text-[12px] font-extrabold tracking-[0.04em] text-[#ea7130] transition hover:bg-[#ea7130]/15"
                      >
                        Comenzar entrenamiento
                      </button>
                  )}
                  <button
                    type="button"
                    onClick={() => handleCancelPlannedWorkout(workout.id)}
                    disabled={calendarActionStatus === "loading"}
                    className="cursor-pointer rounded-[12px] border border-[#9f2f22]/25 bg-[#9f2f22]/10 px-3.5 py-2 text-[12px] font-extrabold tracking-[0.04em] text-[#9f2f22] transition hover:bg-[#9f2f22]/15 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    Cancelar
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
        </DialogPopup>
      )}

      {calendarDialog?.type === "performed" && (
        <DialogPopup
          kicker="Entreno realizado"
          kickerColor="#ea7130"
          title={calendarDialog.dateKey}
          onClose={handleCloseCalendarDialog}
        >
          <ul className="mt-2 flex flex-col gap-2.5">
            {calendarDialog.workouts.map((workout) => (
              <li
                key={workout.id}
                className="rounded-[16px] border border-[#1f1b16]/12 bg-[#fffaf0]/85 p-4"
              >
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="m-0 text-[16px] font-extrabold tracking-tight text-[#1f1b16] uppercase">
                      {workoutTitle(workout)}
                    </p>
                  </div>
                  {workout.performed_at && (
                    <span className="rounded-md border border-[#265c52]/20 bg-[#265c52]/10 px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.14em] text-[#265c52]">
                      {workoutTimeFormatter.format(new Date(workout.performed_at))}
                    </span>
                  )}
                </div>
                <div className="mt-3 flex items-center justify-between gap-3">
                  <p className="text-[13px] font-semibold text-[#3a332c]">
                    <strong className="text-[#1f1b16]">{workout.exercise_count}</strong> {workout.exercise_count === 1 ? "ejercicio" : "ejercicios"} ·{" "}
                    <strong className="text-[#1f1b16]">{workout.duration_minutes}</strong> minutos
                  </p>
                  <button
                    type="button"
                    onClick={() => navigate(`/workouts/${workout.id}/fill`)}
                    className="inline-flex cursor-pointer rounded-[12px] border border-[#265c52]/25 bg-[#265c52]/10 px-3.5 py-2 text-[12px] font-extrabold tracking-[0.04em] text-[#265c52] transition hover:bg-[#265c52]/15"
                  >
                    Ver entrenamiento
                  </button>
                </div>
              </li>
            ))}
          </ul>
        </DialogPopup>
      )}
    </main>
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

function MobileDashboardPanel({
  kicker,
  title,
  children,
}: {
  kicker: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="overflow-hidden rounded-[26px] border border-[#1f1b16]/10 bg-[#fffaf0]/84 shadow-[0_12px_28px_rgba(31,27,22,0.08)]">
      <div className="flex items-center justify-between gap-3 px-4 py-4">
        <div className="min-w-0">
          <p className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.18em] text-[#265c52]">
            {kicker}
          </p>
          <h2 className="mt-1 truncate [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[24px] font-black leading-none tracking-[-0.045em] text-[#1f1b16]">
            {title}
          </h2>
        </div>
      </div>
      <div className="border-t border-[#1f1b16]/10 px-4 pb-4 pt-3">{children}</div>
    </section>
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
    <div className="flex h-full flex-col rounded-[16px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-4">
      <p className="m-0 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[14px] font-bold uppercase tracking-[0.16em] text-[#3a332c]">
        {label}
      </p>
      <p className="mt-auto [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[26px] font-black leading-none tracking-[-0.04em] text-[#1f1b16]">
        {value}
      </p>
      <p
        className="mt-2 inline-flex items-center gap-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold tracking-wider"
        style={{ color }}
      >
        {direction === "up" && <span>▲</span>}
        {direction === "down" && <span>▼</span>}
        {direction === "flat" && delta !== "Aún no hay comparativa" && <span>-</span>}
        {delta}
      </p>
    </div>
  );
}
