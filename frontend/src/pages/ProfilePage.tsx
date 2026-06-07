import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiUrl } from "../lib/api";
import {
ResponsiveContainer,
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip,
} from "recharts";
import {Card, CardHeader} from "../components/Card";
import Model from "react-body-highlighter";
import {HelloHeader} from "@/components/HelloHeader.tsx";
import { Stat } from "../components/Stat";
import {MuscleDistribution, type MuscleDistributionData} from "@/components/MuscleDistribution.tsx";
import {DialogPopup} from "@/components/DialogPopup.tsx";


// --- INTERFACES ---
interface UserProfile { id: string; username: string; email: string; role: string; created_at?: string; }
interface BodyMetric { id: string; weight_kg: number; height_cm?: number; body_fat_percentage?: number; recorded_at: string; }
interface UserGoal { short_term: string; long_term: string; target_days: number; }
interface ExerciseStat { name: string; sets: number; }
interface MuscleRadarStat { muscle: string; value: number; }
interface CalendarActivity { id: string; date: string; workout_name: string; duration: number; exercise_count: number; is_planned: boolean; }

interface ProfileStats {
  total_workouts: number; total_duration_minutes: number; total_volume_kg: number; total_sets: number;
  weekly_streak: number; weekly_goal: number;
  streak_days: string[]; streak_activities: CalendarActivity[]; top_exercises: ExerciseStat[];
  muscle_radar: MuscleRadarStat[]; weight_history: BodyMetric[]; goals: UserGoal | null;
}

const calendarFormatter = new Intl.DateTimeFormat("es-ES", { month: "long", year: "numeric" });
const weekdayFormatter = new Intl.DateTimeFormat("es-ES", { weekday: "short" });

function toDateKey(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

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

export default function Profile() {
  const navigate = useNavigate();
  const [user, setUser] = useState<UserProfile | null>(null);
  const [stats, setStats] = useState<ProfileStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statsRange, setStatsRange] = useState<"month" | "all">("month");

  const [visibleMonth, setVisibleMonth] = useState(() => {
    const today = new Date();
    return new Date(today.getFullYear(), today.getMonth(), 1);
  });

  const [goals, setGoals] = useState({ shortTerm: "", longTerm: "", targetDays: 3 });
  const [isSavingGoals, setIsSavingGoals] = useState(false);
  const [aiInsight, setAiInsight] = useState<string | null>(null);
  const [isAnalyzing, setIsAnalyzing] = useState(false);

  const [selectedDayInfo, setSelectedDayInfo] = useState<{ date: string, activities: CalendarActivity[], isFuture: boolean } | null>(null);
  const [isCancelling, setIsCancelling] = useState(false);

  const [isAddingMetric, setIsAddingMetric] = useState(false);
  const [newMetric, setNewMetric] = useState({ weightKg: "", bodyFat: "" });
  const [isSavingMetric, setIsSavingMetric] = useState(false);
  const [metricError, setMetricError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const userRes = await fetch(apiUrl("/api/auth/me"), { credentials: "include" });
        if (!userRes.ok) throw new Error("No se pudo cargar el perfil del usuario.");
        const userData = await userRes.json();
        setUser(userData.user);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Error inesperado");
        setIsLoading(false);
      }
    };
    void fetchUser();
  }, []);

  useEffect(() => {
    if (!user) return;
    const fetchStats = async () => {
      setIsLoading(true);
      try {
        let url = `/api/profile/dashboard?range=${statsRange}`;
        if (statsRange === "month") {
          url += `&year=${visibleMonth.getFullYear()}&month=${visibleMonth.getMonth() + 1}`;
        }
        const statsRes = await fetch(apiUrl(url), { credentials: "include" });
        if (statsRes.ok) {
          const statsData = (await statsRes.json()) as ProfileStats;
          setStats(statsData);
          if (statsData.goals) {
            setGoals({ shortTerm: statsData.goals.short_term, longTerm: statsData.goals.long_term, targetDays: statsData.goals.target_days });
          }
        }
      } catch (err) { console.error(err); } finally { setIsLoading(false); }
    };
    void fetchStats();
  }, [user, statsRange, visibleMonth]);

  const handleSaveGoals = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSavingGoals(true);
    try {
      await fetch(apiUrl("/api/profile/goals"), {
        method: "PUT", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ short_term: goals.shortTerm, long_term: goals.longTerm, target_days: goals.targetDays }),
        credentials: "include",
      });
      alert("¡Metas guardadas correctamente!");
    } catch { alert("Error al guardar metas."); } finally { setIsSavingGoals(false); }
  };

  const handleAIAnalysis = async () => {
    setIsAnalyzing(true);
    try {
      const res = await fetch(apiUrl("/api/profile/ai-analysis"), {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) {
        if (res.status === 429) {
          setAiInsight("¡Uf! He estado analizando muchos perfiles hoy y necesito un breve descanso para recuperar energía. Vuelve a intentarlo en un ratito, ¡y seguiremos dándolo todo!");
          return;
        }
        throw new Error("Error al obtener el análisis del entrenador IA");
      }
      const data = await res.json();
      setAiInsight(data.analysis);
    } catch {
      setAiInsight("Parece que hubo un pequeño problema de conexión con tu entrenador. Inténtalo de nuevo más tarde.");
    } finally {
      setIsAnalyzing(false);
    }
  };

  const handleSaveMetric = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSavingMetric(true);
    setMetricError(null);

    const weight = parseFloat(newMetric.weightKg);
    if (isNaN(weight) || weight <= 0 || weight > 500) {
      setMetricError("Por favor, introduce un peso válido mayor que 0.");
      setIsSavingMetric(false);
      return;
    }

    const payload: Record<string, number> = { weight_kg: weight };

    if (newMetric.bodyFat) {
      const fat = parseFloat(newMetric.bodyFat);
      if (isNaN(fat) || fat < 0 || fat > 100) {
        setMetricError("El porcentaje de grasa debe estar entre 0 y 100.");
        setIsSavingMetric(false);
        return;
      }
      payload.body_fat_percentage = fat;
    }

    try {

      const res = await fetch(apiUrl("/api/profile/metrics"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
        credentials: "include",
      });

      if (!res.ok) throw new Error("Error al guardar pesaje");

      // Reload stats
      const statsRes = await fetch(apiUrl(`/api/profile/dashboard?range=${statsRange}&year=${visibleMonth.getFullYear()}&month=${visibleMonth.getMonth() + 1}`), { credentials: "include" });
      if (statsRes.ok) {
        const statsData = (await statsRes.json()) as ProfileStats;
        setStats(statsData);
      }
      setIsAddingMetric(false);
      setNewMetric({ weightKg: "", bodyFat: "" });
    } catch (err) {
      setMetricError(err instanceof Error ? err.message : "Error al guardar el pesaje.");
    } finally {
      setIsSavingMetric(false);
    }
  };

  const handleCancelPlannedWorkout = async (workoutID: string) => {
    setIsCancelling(true);
    try {
      const res = await fetch(apiUrl(`/api/workout/${workoutID}`), {
        method: "DELETE",
        credentials: "include",
      });
      if (!res.ok) return;

      setSelectedDayInfo(null);
      let url = `/api/profile/dashboard?range=${statsRange}`;
      if (statsRange === "month") {
        url += `&year=${visibleMonth.getFullYear()}&month=${visibleMonth.getMonth() + 1}`;
      }
      const statsRes = await fetch(apiUrl(url), { credentials: "include" });
      if (statsRes.ok) {
        setStats((await statsRes.json()) as ProfileStats);
      }
    } catch {
      // no-op
    } finally {
      setIsCancelling(false);
    }
  };

  // --- EXTRACCIÓN DE DATOS ---
  const radarData = useMemo(() => stats?.muscle_radar ?? [], [stats?.muscle_radar]);
  // El backend ya filtra muscle_radar según statsRange (mes / todo el historial),
  // así que el heptágono se alimenta del dato actual sin toggle propio.
  const muscleDistributionData = useMemo<MuscleDistributionData>(() => {
    const breakdown = radarData.map((m) => ({ name: m.muscle, count: m.value, percentage: 0 }));
    return { year: breakdown, month: breakdown, year_exercise_count: 0, month_exercise_count: 0 };
  }, [radarData]);
  const topExercises = stats?.top_exercises || [];
  const weightHistory = stats?.weight_history || [];
  const activitiesByDate = useMemo(() => {
    const grouped = new Map<string, CalendarActivity[]>();
    for (const activity of stats?.streak_activities ?? []) {
      grouped.set(activity.date, [...(grouped.get(activity.date) ?? []), activity]);
    }
    return grouped;
  }, [stats?.streak_activities]);

  const weightData = weightHistory.map((w, index) => {
    let bmiValue = null;
    if (w.weight_kg && w.height_cm) {
      const h = w.height_cm / 100;
      bmiValue = parseFloat((w.weight_kg / (h * h)).toFixed(1));
    }
    return {
      id: w.id || `metric-${index}`,
      date: new Date(w.recorded_at).toLocaleDateString('es-ES', { day: 'numeric', month: 'short' }),
      weight: w.weight_kg,
      bmi: bmiValue,
      fat: w.body_fat_percentage,
      isLast: index === weightHistory.length - 1
    };
  });
  const lastMetric = weightHistory.length > 0 ? weightHistory[weightHistory.length - 1] : null;
  const currentWeight = lastMetric?.weight_kg || 0;
  const currentFat = lastMetric?.body_fat_percentage || null;
  let currentBMI: string | null = null;
  if (lastMetric?.weight_kg && lastMetric?.height_cm) {
    const h = lastMetric.height_cm / 100;
    currentBMI = (lastMetric.weight_kg / (h * h)).toFixed(1);
  }

  // --- CALENDARIO MENSUAL (navegable) ---
  const todayDate = useMemo(() => new Date(), []);
  const todayKey = useMemo(() => toDateKey(todayDate), [todayDate]);
  const canShowNextMonth = true;

  const weekdayLabels = useMemo(() => {
    const baseMonday = new Date(2026, 5, 1); // Un Lunes conocido
    return Array.from({ length: 7 }, (_, index) => {
      const date = new Date(baseMonday);
      date.setDate(baseMonday.getDate() + index);
      return weekdayFormatter.format(date).replace(".", "");
    });
  }, []);

  const calendarDays = useMemo(() => buildMonthDays(visibleMonth, todayDate), [todayDate, visibleMonth]);

  const handleChangeVisibleMonth = useCallback((offset: number) => {
    setSelectedDayInfo(null);
    setVisibleMonth((current) => new Date(current.getFullYear(), current.getMonth() + offset, 1));
  }, []);

  if (isLoading && !stats) return <div className="p-20 text-center font-bold text-[#5d5348] animate-pulse">Cargando perfil...</div>;
  if (!user) {
    if (error) {
      return <div className="p-20 text-center font-bold text-[#9f2f22]">{error}</div>;
    }
    return null;
  }

  return (
    <div className="relative isolate overflow-x-hidden text-[#1f1b16] [font-family:'Inter',system-ui,sans-serif] antialiased pb-20">
      {/* Background blobs decorativos del dashboard */}
      <div aria-hidden="true" className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]" />
      <div aria-hidden="true" className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10" />

      <div className="px-6 pt-8 sm:px-8">

        {/* HERO Y STATS SUPERIORES */}
        <section className="mx-auto mb-6 grid max-w-[1280px] grid-cols-1 items-start gap-6 md:grid-cols-[1fr_auto]">
          <div>
            <HelloHeader page={"PERFIL DE ATLETA"} user={user.username}/>
            <p className="mt-5.5 max-w-[640px] text-[15px] leading-[1.55] text-[#3a332c] flex items-center gap-2">
              <span className="bg-[#1f1b16]/10 px-2 py-1 rounded-md [font-family:'JetBrains_Mono',ui-monospace,monospace] text-xs font-bold">Email: {user.email}</span>
            </p>
            {user.role === "admin" && (
            <p className="mt-3.5 max-w-[640px] text-[15px] leading-[1.55] text-[#3a332c] flex items-center gap-2">
              <span className="bg-[#ea7130]/10 text-[#ea7130] px-2 py-1 rounded-md [font-family:'JetBrains_Mono',ui-monospace,monospace] text-xs font-bold uppercase">Rol: {user.role}</span>
            </p>
            )}
            {error && (
              <p className="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-3 py-2 text-[12px] font-bold text-[#9f2f22]">
                {error}
              </p>
            )}
          </div>

          <div className="flex flex-col items-end gap-3">
            <div className="inline-flex gap-0.5 rounded-[12px] border border-[#1f1b16]/10 bg-[#fffaf0]/60 p-1 backdrop-blur-sm">
              {(["month", "all"] as const).map((r) => (
                <button key={r} onClick={() => setStatsRange(r)} className={`rounded-[8px] px-3.5 py-2.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase leading-none tracking-[0.14em] transition ${statsRange === r ? "bg-[#ea7130] text-[#1f1b16] shadow-sm" : "bg-transparent text-[#3a332c] hover:bg-[#1f1b16]/5"}`}>
                  {r === "month" ? "Mes Actual" : "Todo el Historial"}
                </button>
              ))}
            </div>
            <div className="flex flex-wrap justify-end gap-2.5">
              <Stat n={String(stats?.weekly_streak ?? 0)} l="Racha semanal" accent />
              <Stat n={String(stats?.total_workouts ?? 0)} l="Entrenos" />
              <Stat n={`${stats?.total_duration_minutes ?? 0} m`} l="Duración" />
              <Stat n={`${((stats?.total_volume_kg || 0) / 1000).toFixed(1)} KG`} l="Vol. (kg)" />
              <Stat n={String(stats?.total_sets ?? 0)} l="Series" />
            </div>
          </div>
        </section>

        {/* IA ASSISTANT */}
        <section className="mx-auto mb-8 max-w-[1280px]">
          <Card accent="#ea7130" dark>
            <div className="flex flex-col sm:flex-row gap-6 items-center justify-between">
              <div>
                <h3 className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[22px] font-black text-[#fffaf0]">Entrenador IA</h3>
                <p className="mt-1 text-sm text-[#fffaf0]/80 max-w-2xl">{aiInsight || "Haz clic para generar un análisis personalizado de tu rendimiento basado en tus últimas métricas."}</p>
              </div>
              <button onClick={handleAIAnalysis} disabled={isAnalyzing} className="shrink-0 group relative cursor-pointer overflow-hidden rounded-[14px] bg-[#ea7130] px-6 py-4 text-[13px] font-extrabold tracking-[0.04em] text-[#1f1b16] shadow-[0_18px_35px_rgba(234,113,48,0.30)] transition hover:-translate-y-px hover:bg-[#ff8b47] disabled:opacity-50">
                <span className="relative z-10">{isAnalyzing ? "Analizando datos..." : "Generar Análisis"}</span>
              </button>
            </div>
          </Card>
        </section>

        {/* CONTENIDO PRINCIPAL */}
        <section className="mx-auto grid max-w-[1280px] grid-cols-1 gap-[18px] xl:grid-cols-[minmax(0,1.55fr)_minmax(20rem,0.95fr)]">

          {/* COLUMNA IZQUIERDA */}
          <div className="flex flex-col gap-[18px] h-full">
            <Card accent="#ea7130" className="flex-1">
              <CardHeader
                  kicker="Progreso"
                  title="Evolución de Peso"
                  right={
                    <button
                        onClick={() => setIsAddingMetric(true)}
                        className="mt-2 w-max rounded-[14px] border border-[#1f1b16]/18 bg-[#1f1b16] px-5 py-3 text-[14px] font-extrabold tracking-[0.04em] text-[#f1a45b] transition hover:-translate-y-px hover:bg-[#2c261f]"
                    >
                      Añadir Pesaje
                    </button>
                  }
              />
              <div className="relative z-[2] mt-4 flex gap-2.5">
                <MetricTile label="Peso Actual" value={`${currentWeight || "--"} kg`} />
                <MetricTile label="IMC" value={`${currentBMI || "--"}`} />
                <MetricTile label="% Grasa" value={`${currentFat ? currentFat + "%" : "--"}`} />
              </div>
              <div className="relative z-[2] mt-6 h-[240px] w-full">
                {weightData.length > 1 ? (
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart data={weightData} margin={{ left: -20, bottom: 0 }}>
                        <CartesianGrid strokeDasharray="3 3" vertical={false} strokeOpacity={0.1} stroke="#1f1b16" />
                        <XAxis dataKey="id" tickFormatter={(id) => weightData.find(d => d.id === id)?.date || ''} tick={{ fontSize: 10, fill: '#3a332c', fontFamily: 'JetBrains Mono' }} axisLine={false} tickLine={false} />
                        <YAxis yAxisId="left" domain={['dataMin - 1', 'dataMax + 1']} tick={{ fontSize: 10, fill: '#3a332c', fontFamily: 'JetBrains Mono' }} axisLine={false} tickLine={false} />
                        <YAxis yAxisId="right" orientation="right" domain={['auto', 'auto']} tick={{ fontSize: 10, fill: '#6b4ea0', fontFamily: 'JetBrains Mono' }} axisLine={false} tickLine={false} />
                        <Tooltip
                            content={({ active, payload }) => {
                              if (active && payload && payload.length > 0) {
                                const data = payload[0]?.payload;
                                if (!data) return null;
                                return (
                                    <div className="rounded-[14px] bg-[#fffaf0] p-3 shadow-[0_10px_22px_rgba(31,27,22,0.1)] border border-[#1f1b16]/10">
                                      <p className="mb-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase text-[#3a332c]">{data.date}</p>
                                      <p className="text-[13px] font-bold text-[#ea7130]">Peso: {data.weight} kg</p>
                                      {data.fat && <p className="text-[13px] font-bold text-[#3b8678]">% Grasa: {data.fat}%</p>}
                                      {data.bmi && <p className="text-[13px] font-bold text-[#6b4ea0]">IMC: {data.bmi}</p>}
                                    </div>
                                );
                              }
                              return null;
                            }}
                        />

                        <Line yAxisId="left" type="monotone" dataKey="weight" name="Peso (kg)" stroke="#ea7130" strokeWidth={3}
                              dot={(props: { cx?: number; cy?: number; payload?: { isLast?: boolean; date?: string } }) => {
                                const { cx, cy, payload } = props;
                                if (!cx || !cy || !payload) return null;
                                if (payload.isLast) {
                                  return <circle cx={cx} cy={cy} r={6} fill="#ea7130" stroke="#fffaf0" strokeWidth={2} key={`dot-${payload.date}`} />;
                                }
                                return <circle cx={cx} cy={cy} r={4} fill="#ea7130" key={`dot-${payload.date}`} />;
                              }}
                              activeDot={{ r: 6 }}
                        />
                        <Line yAxisId="right" type="monotone" dataKey="bmi" name="IMC" stroke="#6b4ea0" strokeWidth={2} strokeDasharray="5 5" connectNulls={true}
                              dot={(props: { cx?: number; cy?: number; payload?: { isLast?: boolean; date?: string; bmi?: number } }) => {
                                const { cx, cy, payload } = props;
                                if (!cx || !cy || !payload || !payload.bmi) return null;
                                if (payload.isLast) {
                                  return <circle cx={cx} cy={cy} r={5} fill="#6b4ea0" stroke="#fffaf0" strokeWidth={2} key={`dot-bmi-${payload.date}`} />;
                                }
                                return <circle cx={cx} cy={cy} r={3} fill="#6b4ea0" key={`dot-bmi-${payload.date}`} />;
                              }}
                        />
                        <Line yAxisId="right" type="monotone" dataKey="fat" name="% Grasa" stroke="#3b8678" strokeWidth={2} strokeDasharray="5 5" connectNulls={true}
                              dot={(props: { cx?: number; cy?: number; payload?: { isLast?: boolean; date?: string; fat?: number } }) => {
                                const { cx, cy, payload } = props;
                                if (!cx || !cy || !payload || !payload.fat) return null;
                                if (payload.isLast) {
                                  return <circle cx={cx} cy={cy} r={5} fill="#3b8678" stroke="#fffaf0" strokeWidth={2} key={`dot-fat-${payload.date}`} />;
                                }
                                return <circle cx={cx} cy={cy} r={3} fill="#3b8678" key={`dot-fat-${payload.date}`} />;
                              }}
                        />
                      </LineChart>
                    </ResponsiveContainer>
                ) : <div className="h-full flex items-center justify-center [font-family:'JetBrains_Mono',ui-monospace,monospace] text-xs text-[#3a332c]">Añade pesajes para ver tu gráfica.</div>}
              </div>
            </Card>
            <Card accent="#ea7130">
              <CardHeader
                  kicker="Calendario"
                  title={<span className="capitalize">{calendarFormatter.format(visibleMonth)}</span>}
                  right={(
                      <div className="flex items-center gap-1.5">
                        <button
                            onClick={() => {
                              const today = new Date();
                              setVisibleMonth(new Date(today.getFullYear(), today.getMonth(), 1));
                              setSelectedDayInfo(null);
                            }}
                            className="h-9 rounded-[20px] border border-[#1f1b16]/12 bg-[#fffaf0]/80 px-3 py-1.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold capitalize tracking-[0.14em] text-[#1f1b16] transition hover:border-[#ea7130] hover:text-[#ea7130]"
                        >
                          Hoy
                        </button>
                        <NavBtn dir="prev" onClick={() => handleChangeVisibleMonth(-1)} />
                        <NavBtn dir="next" disabled={!canShowNextMonth} onClick={() => handleChangeVisibleMonth(1)} />
                      </div>
                  )}
              />
              <div className="relative z-[2] mt-4 grid grid-cols-7 gap-1.5">
                {weekdayLabels.map((label) => (
                    <p key={label} className="text-center [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.18em] text-[#3a332c]">{label}</p>
                ))}
                {calendarDays.map((day) => {
                  const dateKey = toDateKey(day.date);
                  const isPast = dateKey < todayKey;
                  const isFuture = dateKey > todayKey;
                  const activities = activitiesByDate.get(dateKey) ?? [];
                  const isTrained = day.isCurrentMonth && activities.some(a => !a.is_planned);
                  const isPlanned = day.isCurrentMonth && !isTrained && activities.some(a => a.is_planned);
                  const isDisabled = !day.isCurrentMonth;

                  return (
                      <button
                          key={day.key}
                          type="button"
                          disabled={isDisabled}
                          onClick={() => setSelectedDayInfo({ date: dateKey, activities, isFuture })}
                          className={[
                            "relative grid aspect-square place-items-center rounded-[10px] border text-[16px] font-bold transition",
                            !isDisabled && "cursor-pointer hover:-translate-y-px hover:border-[#ea7130] hover:shadow-[0_6px_14px_rgba(234,113,48,0.18)]",
                            isDisabled && "cursor-not-allowed border-transparent bg-transparent text-[#1f1b16]/30",
                            isTrained && "border-transparent bg-[#265c52] text-[#fffaf0] shadow-[0_6px_14px_rgba(38,92,82,0.30)]",
                            isPlanned && "border-[#ea7130]/45 bg-[#ea7130]/12 text-[#1f1b16]",
                            !isTrained && !isPlanned && day.isToday && "border-[#1f1b16] bg-[#fffaf0] text-[#1f1b16]",
                            !isTrained && !isPlanned && !day.isToday && isPast && "border-transparent bg-[#fffaf0]/50 text-[#1f1b16]/45",
                            !isTrained && !isPlanned && !day.isToday && !isPast && isFuture && "border-transparent bg-transparent text-[#1f1b16]/30 hover:border-transparent hover:shadow-none hover:translate-y-0",
                            !isTrained && !isPlanned && !day.isToday && !isPast && !isFuture && "border-transparent bg-[#fffaf0]/50 text-[#1f1b16]/45",
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
              <CardHeader kicker="Metas" title="Tus Objetivos" />
              <form onSubmit={handleSaveGoals} className="relative z-[2] mt-4 flex flex-col gap-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="mb-1 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-wider text-[#3a332c]">A Corto Plazo</label>
                    <input type="text" className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/85 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130]" value={goals.shortTerm} onChange={(e) => setGoals({ ...goals, shortTerm: e.target.value })} />
                  </div>
                  <div>
                    <label className="mb-1 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-wider text-[#3a332c]">Días Objetivo</label>
                    <input type="number" min="1" max="7" className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/85 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130]" value={goals.targetDays} onChange={(e) => setGoals({ ...goals, targetDays: parseInt(e.target.value) || 0 })} />
                  </div>
                </div>
                <div>
                  <label className="mb-1 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-wider text-[#3a332c]">A Largo Plazo</label>
                  <input type="text" className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/85 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130]" value={goals.longTerm} onChange={(e) => setGoals({ ...goals, longTerm: e.target.value })} />
                </div>
                <button type="submit" disabled={isSavingGoals} className="mt-2 w-max rounded-[14px] border border-[#1f1b16]/18 bg-[#1f1b16] px-5 py-3 text-[14px] font-extrabold tracking-[0.04em] text-[#f1a45b] transition hover:-translate-y-px hover:bg-[#2c261f]">
                  {isSavingGoals ? "Guardando..." : "Actualizar Metas"}
                </button>
              </form>
            </Card>



          </div>

          {/* COLUMNA DERECHA */}
          <div className="flex flex-col gap-[18px] h-full">

            <MuscleDistribution
              data={muscleDistributionData}
              range={statsRange === "all" ? "year" : "month"}
              summary={false}
            />

            <Card accent="#ea7130" dark>
              <CardHeader kicker="Heatmap" title="Mapa Muscular" onDark />
              <div className="relative z-[2] mt-6 h-[260px] w-full">
                <BodyHeatmap data={radarData} />
              </div>
            </Card>

            <Card accent="#ea7130" className={"flex-1"}>
              <CardHeader kicker="Top grupos" title="Ejercicios Favoritos" />
              {topExercises.length > 0 ? (
                <ul className="relative z-[2] mt-4 flex flex-col gap-2.5">
                  {topExercises.map((ex, i) => (
                    <li key={i} className="rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-3.5 flex items-center justify-between">
                      <div>
                        <p className="m-0 text-[16px] font-extrabold tracking-tight text-[#1f1b16] uppercase">{ex.name}</p>
                        <p className="mt-0.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.14em] text-[#3a332c]">{ex.sets} {ex.sets === 1 ? "serie completada":"series completadas"}</p>
                      </div>
                      <span className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[20px] font-black text-[#265c52] opacity-80">#{i + 1}</span>
                    </li>
                  ))}
                </ul>
              ) : <p className="relative z-[2] mt-4 rounded-[14px] border border-dashed border-[#1f1b16]/18 p-4 text-[13px] font-semibold text-[#3a332c]">Sin datos de ejercicios.</p>}
            </Card>

          </div>
        </section>
      </div>

      {/* DIALOG SHELL CALENDARIO (ESTILO DASHBOARD) */}
      {selectedDayInfo && (
        <DialogPopup kicker="Detalle del día" kickerColor="#ea7130" title={selectedDayInfo.date} onClose={() => setSelectedDayInfo(null)}>
          {selectedDayInfo.activities.length > 0 ? (
            <ul className="mt-2 flex flex-col gap-2.5">
              {selectedDayInfo.activities.map((act, i) => (
                <li key={i} className="rounded-[16px] border border-[#1f1b16]/12 bg-[#fffaf0]/85 p-4">
                  <p className="m-0 text-[16px] uppercase font-extrabold tracking-tight text-[#1f1b16]">{act.workout_name}</p>
                  {!act.is_planned && (
                    <p className="mt-1 text-[13px] font-semibold text-[#3a332c]">
                      <strong className="text-[#1f1b16]">{act.exercise_count}</strong> {act.exercise_count === 1 ? "ejercicio" : "ejercicios"} ·{" "}
                      <strong className="text-[#1f1b16]">{act.duration}</strong> minutos
                    </p>
                  )}
                  <div className="mt-3 flex flex-wrap gap-2">
                    {act.is_planned ? (
                      <>
                        {!selectedDayInfo.isFuture && (
                          <button type="button" onClick={() => navigate(`/workout/${act.id}`)} className="inline-flex cursor-pointer rounded-[12px] border border-[#ea7130]/25 bg-[#ea7130]/10 px-3.5 py-2 text-[12px] font-extrabold tracking-[0.04em] text-[#ea7130] transition hover:bg-[#ea7130]/15">
                            Rellenar
                          </button>
                        )}
                        <button type="button" onClick={() => handleCancelPlannedWorkout(act.id)} disabled={isCancelling} className="cursor-pointer rounded-[12px] border border-[#9f2f22]/25 bg-[#9f2f22]/10 px-3.5 py-2 text-[12px] font-extrabold tracking-[0.04em] text-[#9f2f22] transition hover:bg-[#9f2f22]/15 disabled:cursor-not-allowed disabled:opacity-50">
                          Cancelar
                        </button>
                      </>
                    ) : (
                      <button type="button" onClick={() => navigate(`/workout/${act.id}/details`)} className="inline-flex cursor-pointer rounded-[12px] border border-[#265c52]/25 bg-[#265c52]/10 px-3.5 py-2 text-[12px] font-extrabold tracking-[0.04em] text-[#265c52] transition hover:bg-[#265c52]/15">
                        Ver detalles
                      </button>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          ) : (
            <div className="mt-2 text-center">
              <p className="text-[13px] font-semibold text-[#3a332c] mb-4">No hay entrenamientos registrados.</p>
              {selectedDayInfo.isFuture && (
                <button onClick={() => navigate("/dashboard")} className="w-full cursor-pointer rounded-[14px] bg-[#ea7130] px-5 py-3 text-[13px] font-extrabold tracking-[0.04em] text-[#1f1b16] shadow-[0_18px_35px_rgba(234,113,48,0.30)] transition hover:-translate-y-px hover:bg-[#ff8b47]">
                  Ir a planificar al Dashboard
                </button>
              )}
            </div>
          )}
        </DialogPopup>
      )}

      {/* DIALOG SHELL REGISTRAR PESAJE */}
      {isAddingMetric && (
        <DialogPopup kicker="Progreso" kickerColor="#3b8678" title="Registrar Pesaje" onClose={() => { setIsAddingMetric(false); setMetricError(null); }}>
          <form onSubmit={handleSaveMetric} className="mt-2 flex flex-col gap-4">
            {metricError && (
              <div className="rounded-lg border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-3 py-2 text-[12px] font-bold text-[#9f2f22]">
                {metricError}
              </div>
            )}
            <div>
              <label className="mb-1 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-wider text-[#3a332c]">Peso (kg) *</label>
              <input type="number" step="0.1" required className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/85 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130]" value={newMetric.weightKg} onChange={(e) => setNewMetric({ ...newMetric, weightKg: e.target.value })} />
            </div>
            <div className="grid grid-cols-1 gap-4">
              <div>
                <label className="mb-1 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold uppercase tracking-wider text-[#3a332c]">% Grasa (opc)</label>
                <input type="number" step="0.1" className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/85 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130]" value={newMetric.bodyFat} onChange={(e) => setNewMetric({ ...newMetric, bodyFat: e.target.value })} />
              </div>
            </div>
            <button type="submit" disabled={isSavingMetric} className="mt-2 w-full rounded-[14px] bg-[#3b8678] px-5 py-3 text-[13px] font-extrabold text-[#fffaf0] shadow-md transition hover:bg-[#2c655a] disabled:opacity-50">
              {isSavingMetric ? "Guardando..." : "Guardar Registro"}
            </button>
          </form>
        </DialogPopup>
      )}

    </div>
  );
}

// ============================================================================
// UI COMPONENTS LOCALES (Calcados de DashboardPage.tsx para mantener el diseño)
// ============================================================================

function MetricTile({ label, value }: { label: string; value: string; }) {
  return (
    <div className="flex-1 rounded-[16px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-4">
      <p className="m-0 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.16em] text-[#3a332c]">{label}</p>
      <p className="mt-1.5 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[26px] font-black leading-none tracking-[-0.04em] text-[#1f1b16]">{value}</p>
    </div>
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

const muscleLabelsEs: Record<string, string> = {
  chest: "Pecho", pecho: "Pecho",
  back: "Espalda", espalda: "Espalda",
  shoulders: "Hombros", hombros: "Hombros",
  biceps: "Bíceps",
  triceps: "Tríceps",
  arms: "Brazos", brazos: "Brazos",
  legs: "Piernas", piernas: "Piernas",
  glutes: "Glúteos", gluteos: "Glúteos",
  core: "Core",
};

function BodyHeatmap({ data }: { data: MuscleRadarStat[] }) {
  const [selected, setSelected] = useState<{ name: string; sets: number } | null>(null);

  const muscleMapping: Record<string, string[]> = {
    chest: ["chest"],
    back: ["upper-back", "lower-back", "trapezius"],
    shoulders: ["front-deltoids", "back-deltoids"],
    biceps: ["biceps"],
    triceps: ["triceps"],
    arms: ["biceps", "triceps", "forearm"],
    legs: ["quadriceps", "hamstring", "calves", "gluteal", "adductor", "abductors"],
    glutes: ["gluteal"],
    core: ["abs", "obliques"],
    pecho: ["chest"],
    espalda: ["upper-back", "lower-back", "trapezius"],
    hombros: ["front-deltoids", "back-deltoids"],
    brazos: ["biceps", "triceps", "forearm"],
    piernas: ["quadriceps", "hamstring", "calves", "gluteal", "adductor", "abductors"],
  };

  const maxVal = Math.max(...data.map((d) => d.value), 1);

  const modelData = data.map((d) => {
    const key = d.muscle.toLowerCase();
    return {
      name: d.muscle,
      muscles: muscleMapping[key] || [key],
      frequency: Math.ceil((d.value / maxVal) * 5),
    };
  }) as React.ComponentProps<typeof Model>["data"];

  const colors = [
    "rgba(234, 113, 48, 0.2)",
    "rgba(234, 113, 48, 0.4)",
    "rgba(234, 113, 48, 0.6)",
    "rgba(234, 113, 48, 0.8)",
    "rgba(234, 113, 48, 1.0)",
  ];

  const handleClick = useCallback(({ data: clickedData }: { data: { exercises?: string[] } | null }) => {
    if (!clickedData || !clickedData.exercises || clickedData.exercises.length === 0) {
      setSelected(null);
      return;
    }
    const dbMuscleName = clickedData.exercises[0];
    const originalStat = data.find((d) => d.muscle === dbMuscleName);
    if (originalStat) {
      const key = originalStat.muscle.toLowerCase();
      setSelected({ name: muscleLabelsEs[key] ?? originalStat.muscle, sets: originalStat.value });
    }
  }, [data]);

  return (
    <div className="relative h-full w-full flex flex-col items-center justify-center">
      {selected && (
        <div className="absolute top-0 left-1/2 -translate-x-1/2 z-10 rounded-[10px] bg-[#1f1b16] px-3 py-1.5 text-center text-[11px] font-bold text-[#fffaf0] shadow-md border border-[#fffaf0]/10">
          <span className="text-[#ea7130] capitalize">{selected.name}</span>: {selected.sets} {selected.sets === 1 ? "serie" : "series"}
        </div>
      )}
      <div className="relative h-full w-full flex items-center justify-center gap-8 mt-0">
        {/* FRONT */}
        <div className="flex h-full flex-col items-center justify-center cursor-pointer">
          <div className="h-[200px] w-auto drop-shadow-md">
            <Model
              data={modelData}
              style={{ width: "10rem" }}
              highlightedColors={colors}
              bodyColor="rgba(255, 250, 240, 0.08)"
              onClick={handleClick}
            />
          </div>
          <span className="relative z-10 mt-3 rounded-[10px] border border-[#fffaf0]/10 bg-[#1f1b16] px-3 py-1.5 text-[12px] font-bold text-[#fffaf0] shadow-md">FRONTAL</span>
        </div>

        {/* BACK */}
        <div className="flex h-full flex-col items-center justify-center cursor-pointer">
          <div className="h-[200px] w-auto drop-shadow-md">
            <Model
              type="posterior"
              data={modelData}
              style={{ width: "10rem" }}
              highlightedColors={colors}
              bodyColor="rgba(255, 250, 240, 0.08)"
              onClick={handleClick}
            />
          </div>
          <span className="relative z-10 mt-3 rounded-[10px] border border-[#fffaf0]/10 bg-[#1f1b16] px-3 py-1.5 text-[12px] font-bold text-[#fffaf0] shadow-md">TRASERA</span>
        </div>
      </div>
    </div>
  );
}