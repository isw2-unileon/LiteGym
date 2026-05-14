import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { apiUrl } from "../lib/api";
import {
  Radar,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from "recharts";

// --- INTERFACES ---
interface UserProfile {
  id: string;
  username: string;
  email: string;
  role: string;
  created_at?: string;
}

interface BodyMetric {
  id: string;
  weight_kg: number;
  body_fat_percentage?: number;
  recorded_at: string;
}

interface UserGoal {
  short_term: string;
  long_term: string;
  target_days: number;
}

interface ExerciseStat {
  name: string;
  sets: number;
}

interface MuscleRadarStat {
  muscle: string;
  value: number;
}

interface ProfileStats {
  total_workouts: number;
  total_duration_minutes: number;
  total_volume_kg: number;
  total_sets: number;
  streak_days: string[];
  top_exercises: ExerciseStat[];
  muscle_radar: MuscleRadarStat[];
  weight_history: BodyMetric[];
  goals: UserGoal | null;
}

export default function Profile() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [stats, setStats] = useState<ProfileStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timeFilter, setTimeFilter] = useState("month");
  
  const [goals, setGoals] = useState({ shortTerm: "", longTerm: "", targetDays: 3 });
  const [isSavingGoals, setIsSavingGoals] = useState(false);
  const [aiInsight, setAiInsight] = useState<string | null>(null);
  const [isAnalyzing, setIsAnalyzing] = useState(false);

  useEffect(() => {
    const fetchProfileData = async () => {
      try {
        const [userRes, statsRes] = await Promise.all([
          fetch(apiUrl("/api/auth/me"), { credentials: "include" }),
          fetch(apiUrl("/api/profile/dashboard"), { credentials: "include" })
        ]);

        if (!userRes.ok) throw new Error("No se pudo cargar el perfil del usuario.");
        
        const userData = await userRes.json();
        setUser(userData.user);

        if (statsRes.ok) {
          const statsData = await statsRes.json() as ProfileStats;
          setStats(statsData);
          
          if (statsData.goals) {
            setGoals({
              shortTerm: statsData.goals.short_term,
              longTerm: statsData.goals.long_term,
              targetDays: statsData.goals.target_days
            });
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Error inesperado");
      } finally {
        setIsLoading(false);
      }
    };

    void fetchProfileData();
  }, []);

  const handleSaveGoals = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSavingGoals(true);
    try {
      const response = await fetch(apiUrl("/api/profile/goals"), {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          short_term: goals.shortTerm,
          long_term: goals.longTerm,
          target_days: goals.targetDays
        }),
        credentials: "include",
      });

      if (!response.ok) throw new Error("Error al guardar metas");
      alert("¡Metas guardadas correctamente!");
    } catch (err) {
      alert("Hubo un problema al guardar tus metas.");
    } finally {
      setIsSavingGoals(false);
    }
  };

  // --- EXTRACCIÓN SEGURA DE DATOS PARA TYPESCRIPT ---
  // Al hacer "|| []", le garantizamos a TypeScript que esto siempre será un Array válido y nunca 'undefined'.
  
  const radarData = stats?.muscle_radar || [];
  const topExercises = stats?.top_exercises || [];
  const weightHistory = stats?.weight_history || [];
  const streakDays = stats?.streak_days || [];

  const handleAIAnalysis = async () => {
    setIsAnalyzing(true);
    try {
      await new Promise((resolve) => setTimeout(resolve, 1500));
      
      const topMuscle = radarData.length > 0 
        ? [...radarData].sort((a, b) => b.value - a.value)[0].muscle 
        : "Pecho";

      setAiInsight(
        `¡Vas por muy buen camino! He analizado tu historial y noto que dominas el trabajo de ${topMuscle}. Para optimizar tu ${goals.shortTerm || "progreso general"}, te sugiero enfocarte un poco más en los músculos secundarios.`
      );
    } finally {
      setIsAnalyzing(false);
    }
  };

  // 1. Calendario de Racha
  const last28Days = Array.from({ length: 28 }).map((_, i) => {
    const d = new Date();
    d.setDate(d.getDate() - (27 - i));
    
    // Le decimos a TS que confíe en que el split devuelve un string as string
    const dateString = (d.toISOString().split('T')[0]) as string; 
    const isActive = streakDays.includes(dateString);
    
    return { day: i, active: isActive, date: dateString };
  });

  const activeDaysCount = last28Days.filter(d => d.active).length;

  // 2. Gráfico de Peso e IMC
  const weightData = weightHistory.map(w => ({
    date: new Date(w.recorded_at).toLocaleDateString('es-ES', { day: 'numeric', month: 'short' }),
    weight: w.weight_kg
  }));

  const currentWeight = weightData.length > 0 ? weightData[weightData.length - 1].weight : 0;
  const currentFat = weightHistory.length > 0 ? weightHistory[weightHistory.length - 1].body_fat_percentage : null;

  // --- RENDERIZADO CONDICIONAL ---
  if (isLoading) {
    return (
      <div className="mx-auto flex h-64 w-full items-center justify-center rounded-[2rem] border-2 border-dashed border-[#1f1b16]/20 bg-[#fffaf0]/50 backdrop-blur-sm">
        <p className="animate-pulse font-semibold text-[#5d5348]">Cargando perfil y estadísticas...</p>
      </div>
    );
  }

  if (error) return <div className="mx-auto w-full rounded-[1.5rem] border border-[#c94b32]/20 bg-[#c94b32]/10 p-6 font-bold text-[#9f2f22]">Error: {error}</div>;
  if (!user) return null;

  return (
    <div className="flex flex-col gap-8 pb-12">
      {/* 1. SECCIÓN SUPERIOR */}
      <div className="grid gap-8 lg:grid-cols-3">
        <div className="mx-auto w-full animate-[rise_700ms_ease-out_both] rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-[0_30px_80px_rgba(47,39,27,0.20)] backdrop-blur-md sm:p-8 lg:col-span-2 flex flex-col sm:flex-row items-center gap-8">
          <div className="flex h-32 w-32 shrink-0 items-center justify-center rounded-full border-4 border-[#fffaf0] bg-[#265c52] text-6xl font-black uppercase text-[#fffaf0] shadow-lg">
            {user.username.charAt(0)}
          </div>
          <div className="flex-1 text-center sm:text-left">
            <h2 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.04em] text-[#1f1b16]">
              {user.username}
            </h2>
            <p className="mt-1 font-medium text-[#5d5348]">{user.email}</p>
            <div className="mt-3 inline-block rounded bg-gray-200 px-3 py-1 text-xs font-bold uppercase tracking-widest text-gray-600">
              Rol actual: {user.role}
            </div>
            
            <div className="mt-6 flex flex-col gap-3 sm:flex-row sm:items-center">
              {user.created_at && (
                <div className="rounded-2xl border border-[#1f1b16]/10 bg-[#1f1b16] px-5 py-3 text-left shadow-inner">
                  <p className="text-[10px] font-semibold uppercase tracking-[0.2em] text-[#f1a45b]">Miembro desde</p>
                  <p className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-lg font-bold text-[#fffaf0]">
                    {new Date(user.created_at).toLocaleDateString()}
                  </p>
                </div>
              )}
              {user.role === "admin" && (
                <Link to="/admin" className="rounded-2xl bg-[#ea7130] px-6 py-4 text-sm font-bold text-white shadow-lg transition-transform hover:scale-[1.02]">
                  Panel de Administración
                </Link>
              )}
            </div>
          </div>
        </div>

        <div className="flex flex-col justify-center rounded-[2rem] border border-[#265c52]/20 bg-[#265c52]/5 p-6 backdrop-blur-md lg:col-span-1">
          <h3 className="mb-4 text-lg font-black text-[#265c52]">Asistente de Progreso</h3>
          <button
            onClick={handleAIAnalysis}
            disabled={isAnalyzing}
            className="w-full rounded-2xl bg-gradient-to-r from-[#265c52] to-[#1a4039] px-6 py-4 font-bold text-[#fffaf0] shadow-lg transition hover:scale-[1.02] hover:shadow-xl disabled:opacity-70"
          >
            {isAnalyzing ? "✨ Analizando datos..." : "✨ Análisis IA de mi progreso"}
          </button>
          {aiInsight && (
            <div className="mt-4 rounded-xl bg-white/60 p-4 text-sm font-medium text-[#1f1b16] shadow-sm animate-[rise_300ms_ease-out_both]">
              <span className="mb-1 block text-xl">🤖</span>
              {aiInsight}
            </div>
          )}
        </div>
      </div>

      {/* 2. FILTROS Y STATS RÁPIDAS */}
      <div className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-xl backdrop-blur-md">
        <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <h2 className="text-xl font-black text-[#1f1b16]">Resumen Estadístico</h2>
          <select
            className="rounded-xl border-none bg-white/50 p-2.5 text-sm font-bold shadow-sm outline-none focus:ring-2 focus:ring-[#f1a45b] cursor-pointer"
            value={timeFilter}
            onChange={(e) => setTimeFilter(e.target.value)}
          >
            <option value="month">Este mes</option>
            <option value="all">Histórico total</option>
          </select>
        </div>
        <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
          <div className="rounded-xl bg-white/60 p-4 shadow-sm">
            <p className="text-[10px] font-bold uppercase tracking-wider text-[#5d5348]">Entrenamientos</p>
            <p className="mt-1 text-3xl font-black text-[#ea7130]">{stats?.total_workouts || 0}</p>
          </div>
          <div className="rounded-xl bg-white/60 p-4 shadow-sm">
            <p className="text-[10px] font-bold uppercase tracking-wider text-[#5d5348]">Tiempo Total</p>
            <p className="mt-1 text-3xl font-black text-[#ea7130]">
              {Math.floor((stats?.total_duration_minutes || 0) / 60)}h {(stats?.total_duration_minutes || 0) % 60}m
            </p>
          </div>
          <div className="rounded-xl bg-white/60 p-4 shadow-sm">
            <p className="text-[10px] font-bold uppercase tracking-wider text-[#5d5348]">Volumen Movido</p>
            <p className="mt-1 text-3xl font-black text-[#ea7130]">
              {((stats?.total_volume_kg || 0) / 1000).toFixed(1)}k <span className="text-sm">kg</span>
            </p>
          </div>
          <div className="rounded-xl bg-white/60 p-4 shadow-sm">
            <p className="text-[10px] font-bold uppercase tracking-wider text-[#5d5348]">Series Totales</p>
            <p className="mt-1 text-3xl font-black text-[#ea7130]">{stats?.total_sets || 0}</p>
          </div>
        </div>
      </div>

      {/* 3. COLUMNAS DE GRÁFICOS Y METAS */}
      <div className="grid gap-8 lg:grid-cols-2">
        
        {/* COLUMNA IZQUIERDA */}
        <div className="flex flex-col gap-8">
          
          <div className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-xl backdrop-blur-md">
            <h3 className="mb-4 text-lg font-black text-[#1f1b16]">Actividad (Últimos 28 días)</h3>
            <div className="grid grid-cols-7 gap-2 sm:gap-3">
              {last28Days.map((day, i) => (
                <div
                  key={i}
                  className={`aspect-square w-full rounded-lg sm:rounded-xl transition-all ${
                    day.active ? "bg-[#ea7130] shadow-sm hover:scale-110" : "bg-[#1f1b16]/5"
                  }`}
                  title={day.date}
                />
              ))}
            </div>
            <p className="mt-4 text-center text-sm font-bold text-[#5d5348]">
              🔥 Días activos: <span className="text-[#ea7130]">{activeDaysCount} días</span>
            </p>
          </div>

          <div className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-xl backdrop-blur-md">
            <h3 className="mb-4 text-lg font-black text-[#1f1b16]">Top Ejercicios (Frecuencia)</h3>
            {topExercises.length > 0 ? (
              <ul className="space-y-3">
                {topExercises.map((ex, i) => (
                  <li key={i} className="flex items-center justify-between rounded-xl bg-white/50 p-3 shadow-sm">
                    <span className="font-bold text-[#1f1b16]">#{i + 1} {ex.name}</span>
                    <span className="rounded-lg bg-[#265c52]/10 px-2 py-1 text-xs font-black text-[#265c52]">
                      {ex.sets} series
                    </span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm font-semibold text-[#5d5348] text-center p-4 bg-white/40 rounded-xl">Registra entrenamientos para ver tu top.</p>
            )}
          </div>

          <div className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-xl backdrop-blur-md">
            <h3 className="mb-4 text-lg font-black text-[#1f1b16]">Mis Metas (Gym Goals)</h3>
            <form onSubmit={handleSaveGoals} className="flex flex-col gap-4">
              <div>
                <label className="mb-1 block text-xs font-bold uppercase text-[#5d5348]">Meta a Corto Plazo</label>
                <input
                  type="text"
                  placeholder="Ej: Levantar 100kg en Press Banca"
                  className="w-full rounded-xl border-none bg-white p-3 text-sm shadow-inner"
                  value={goals.shortTerm}
                  onChange={(e) => setGoals({ ...goals, shortTerm: e.target.value })}
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-bold uppercase text-[#5d5348]">Meta a Largo Plazo</label>
                <input
                  type="text"
                  placeholder="Ej: Bajar al 15% de grasa corporal"
                  className="w-full rounded-xl border-none bg-white p-3 text-sm shadow-inner"
                  value={goals.longTerm}
                  onChange={(e) => setGoals({ ...goals, longTerm: e.target.value })}
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-bold uppercase text-[#5d5348]">Días objetivo por semana</label>
                <input
                  type="number"
                  min="1"
                  max="7"
                  className="w-full rounded-xl border-none bg-white p-3 text-sm shadow-inner"
                  value={goals.targetDays}
                  onChange={(e) => setGoals({ ...goals, targetDays: parseInt(e.target.value) || 0 })}
                />
              </div>
              <button
                type="submit"
                disabled={isSavingGoals}
                className="mt-2 rounded-xl bg-[#1f1b16] py-3 text-sm font-bold text-[#fffaf0] transition hover:bg-[#ea7130]"
              >
                {isSavingGoals ? "Guardando..." : "Guardar Metas"}
              </button>
            </form>
          </div>
        </div>

        {/* COLUMNA DERECHA */}
        <div className="flex flex-col gap-8">
          
          <div className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-xl backdrop-blur-md">
            <h3 className="mb-4 text-lg font-black text-[#1f1b16]">Distribución Muscular <span className="text-xs font-bold text-[#5d5348] font-normal">(Heatmap)</span></h3>
            <div className="h-[250px] w-full">
              {radarData.length > 2 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <RadarChart cx="50%" cy="50%" outerRadius="75%" data={radarData}>
                    <PolarGrid stroke="#1f1b16" strokeOpacity={0.1} />
                    <PolarAngleAxis dataKey="muscle" tick={{ fill: '#5d5348', fontSize: 12, fontWeight: 'bold' }} />
                    <Radar
                      name="Series Realizadas"
                      dataKey="value"
                      stroke="#265c52"
                      strokeWidth={2}
                      fill="#265c52"
                      fillOpacity={0.4}
                    />
                  </RadarChart>
                </ResponsiveContainer>
              ) : (
                 <div className="flex h-full items-center justify-center bg-white/40 rounded-xl">
                   <p className="text-sm font-semibold text-[#5d5348] text-center px-4">Registra entrenamientos en varios grupos musculares para ver el radar.</p>
                 </div>
              )}
            </div>
          </div>

          <div className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-xl backdrop-blur-md">
            <h3 className="mb-4 text-lg font-black text-[#1f1b16]">Medidas Corporales</h3>
            <div className="mb-6 flex gap-3">
              <div className="flex-1 rounded-xl bg-white/50 p-3 text-center shadow-sm">
                <p className="text-[10px] font-bold uppercase text-[#5d5348]">Peso</p>
                <p className="text-xl font-black text-[#1f1b16]">{currentWeight > 0 ? currentWeight : "--"}<span className="text-xs">kg</span></p>
              </div>
              <div className="flex-1 rounded-xl bg-white/50 p-3 text-center shadow-sm">
                <p className="text-[10px] font-bold uppercase text-[#5d5348]">% Grasa</p>
                <p className="text-xl font-black text-[#1f1b16]">{currentFat ? currentFat + "%" : "--"}</p>
              </div>
            </div>
            
            <div className="h-[180px] w-full">
              {weightData.length > 1 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={weightData} margin={{ top: 5, right: 5, bottom: 5, left: -20 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#1f1b16" strokeOpacity={0.1} vertical={false} />
                    <XAxis dataKey="date" tick={{ fill: '#5d5348', fontSize: 11 }} axisLine={false} tickLine={false} />
                    <YAxis domain={['dataMin - 1', 'dataMax + 1']} tick={{ fill: '#5d5348', fontSize: 11 }} axisLine={false} tickLine={false} />
                    <Tooltip 
                      contentStyle={{ borderRadius: '12px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)', fontSize: '12px', fontWeight: 'bold' }}
                      itemStyle={{ color: '#ea7130' }}
                    />
                    <Line type="monotone" dataKey="weight" stroke="#ea7130" strokeWidth={3} dot={{ r: 4, fill: '#ea7130' }} activeDot={{ r: 6 }} />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div className="flex h-full items-center justify-center bg-white/40 rounded-xl">
                   <p className="text-sm font-semibold text-[#5d5348] text-center px-4">Añade más pesajes para ver tu gráfica de evolución.</p>
                </div>
              )}
            </div>
          </div>

        </div>
      </div>
    </div>
  );
}