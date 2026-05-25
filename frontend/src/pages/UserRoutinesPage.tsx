import { type FormEvent, useCallback, useEffect, useState } from "react";
import { apiUrl } from "../lib/api";
import AIRoutinePreviewModal, {
  type AIRoutinePreview,
} from "../components/Routine/AIRoutinePreviewModal";

type RoutineSummary = {
  id: string;
  name: string;
  description?: string;
  exercise_count: number;
  updated_at?: string;
};

type RoutineSetDetail = {
  id: string;
  set_number: number;
  target_reps_min?: number;
  target_reps_max?: number;
  target_reps_text?: string;
  target_weight_kg?: number;
  target_duration_seconds?: number;
  target_distance_km?: number;
  target_rir?: number;
  rest_seconds?: number;
  notes?: string;
};

type RoutineExerciseDetail = {
  id: string;
  exercise_id: string;
  name: string;
  description?: string;
  muscle_group: string;
  secondary_muscle_group?: string;
  exercise_type?: string;
  exercise_order: number;
  notes?: string;
  sets: RoutineSetDetail[];
};

type RoutineDetail = RoutineSummary & {
  source: string;
  created_at: string;
  exercises: RoutineExerciseDetail[];
};

type AIRoutineGenerateResponse = {
  routine_json: AIRoutinePreview;
  routine_id?: string;
  rate_limit?: {
    remaining: number;
    reset_at: string;
  };
};

type AIRoutineSaveResponse = {
  routine_json: AIRoutinePreview;
  routine_id: string;
};

type RoutineStatus = "idle" | "loading" | "success" | "error";

const dateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "numeric",
  month: "short",
  year: "numeric",
});

function formatRoutineDate(value?: string) {
  if (!value) {
    return "Sin fecha";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Sin fecha";
  }

  return dateFormatter.format(date);
}

export default function UserRoutinesPage() {
  const [routines, setRoutines] = useState<RoutineSummary[]>([]);
  const [status, setStatus] = useState<RoutineStatus>("idle");
  const [selectedRoutineID, setSelectedRoutineID] = useState("");
  const [selectedRoutine, setSelectedRoutine] = useState<RoutineDetail | null>(
    null,
  );
  const [detailStatus, setDetailStatus] = useState<RoutineStatus>("idle");
  const [isAIFormOpen, setIsAIFormOpen] = useState(false);
  const [aiObjective, setAIObjective] = useState("Ganar fuerza");
  const [aiDuration, setAIDuration] = useState(60);
  const [aiMuscleGroups, setAIMuscleGroups] = useState("");
  const [aiMandatoryExerciseIDs, setAIMandatoryExerciseIDs] = useState("");
  const [aiStatus, setAIStatus] = useState<RoutineStatus>("idle");
  const [aiMessage, setAIMessage] = useState("");
  const [aiPreviewRoutine, setAIPreviewRoutine] =
    useState<AIRoutinePreview | null>(null);
  const [isAIPreviewOpen, setIsAIPreviewOpen] = useState(false);
  const [isAIPreviewSaving, setIsAIPreviewSaving] = useState(false);
  const [aiPreviewError, setAIPreviewError] = useState("");

  const fetchRoutineDetail = useCallback(async (routineID: string) => {
    setSelectedRoutineID(routineID);
    setDetailStatus("loading");

    try {
      const response = await fetch(apiUrl(`/api/routines/${routineID}`), {
        credentials: "include",
      });

      if (!response.ok) {
        setSelectedRoutine(null);
        setDetailStatus("error");
        return;
      }

      const payload = (await response.json()) as RoutineDetail;
      setSelectedRoutine(payload);
      setDetailStatus("success");
    } catch {
      setSelectedRoutine(null);
      setDetailStatus("error");
    }
  }, []);

  const fetchRoutines = useCallback(async () => {
    setStatus("loading");

    try {
      const response = await fetch(apiUrl("/api/routines"), {
        credentials: "include",
      });

      if (!response.ok) {
        setRoutines([]);
        setStatus("error");
        return;
      }

      const payload = (await response.json()) as RoutineSummary[];
      setRoutines(payload);
      setStatus("success");
    } catch {
      setRoutines([]);
      setStatus("error");
    }
  }, []);

  useEffect(() => {
    void fetchRoutines();
  }, [fetchRoutines]);

  const handleGenerateAIRoutine = async (
    event: FormEvent<HTMLFormElement>,
  ) => {
    event.preventDefault();
    setAIStatus("loading");
    setAIMessage("");

    try {
      const response = await fetch(apiUrl("/api/routines/ai/generate"), {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          objective: aiObjective.trim(),
          duration_minutes: aiDuration,
          target_muscle_groups: splitCommaList(aiMuscleGroups),
          mandatory_exercise_ids: splitCommaList(aiMandatoryExerciseIDs),
        }),
      });

      const payload = (await response.json()) as AIRoutineGenerateResponse & {
        error?: string;
        detail?: string;
      };

      if (!response.ok) {
        setAIStatus("error");
        setAIMessage(payload.detail || payload.error || "No se pudo generar la rutina.");
        return;
      }

      if (!payload.routine_json) {
        setAIStatus("error");
        setAIMessage("La IA no devolvio una vista previa valida.");
        return;
      }

      setAIStatus("success");
      setAIPreviewRoutine(payload.routine_json);
      setAIPreviewError("");
      setIsAIPreviewOpen(true);
      setIsAIFormOpen(false);
    } catch {
      setAIStatus("error");
      setAIMessage("No se pudo conectar con la inteligencia artificial.");
    }
  };

  const handleConfirmAIRoutine = async () => {
    if (aiPreviewRoutine == null) {
      return;
    }

    setIsAIPreviewSaving(true);
    setAIPreviewError("");
    setAIMessage("");

    try {
      const response = await fetch(apiUrl("/api/routines/ai/save"), {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          routine_json: aiPreviewRoutine,
        }),
      });

      const payload = (await response.json()) as AIRoutineSaveResponse & {
        error?: string;
        detail?: string;
      };

      if (!response.ok) {
        setAIPreviewError(
          payload.detail || payload.error || "No se pudo guardar la rutina.",
        );
        return;
      }

      setIsAIPreviewOpen(false);
      setAIPreviewRoutine(null);
      setAIStatus("success");
      setAIMessage("Rutina guardada.");
      await fetchRoutines();

      if (payload.routine_id) {
        await fetchRoutineDetail(payload.routine_id);
      }
    } catch {
      setAIPreviewError("No se pudo conectar para guardar la rutina.");
    } finally {
      setIsAIPreviewSaving(false);
    }
  };

  const handleCloseAIRoutinePreview = () => {
    if (isAIPreviewSaving) {
      return;
    }

    setIsAIPreviewOpen(false);
    setAIPreviewRoutine(null);
    setAIPreviewError("");
  };

  const totalExercises = routines.reduce(
    (total, routine) => total + routine.exercise_count,
    0,
  );

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/88 p-6 shadow-[0_24px_60px_rgba(47,39,27,0.14)] lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p className="text-xs font-black uppercase tracking-[0.24em] text-[#265c52]">
            Biblioteca
          </p>
          <h1 className="mt-3 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.05em]">
            Mis rutinas
          </h1>
          <p className="mt-3 max-w-2xl text-sm font-semibold leading-6 text-[#5d5348]">
            Rutinas guardadas para planificar entrenamientos y reutilizar
            ejercicios.
          </p>
        </div>

        <div className="space-y-3 lg:min-w-80">
          <div className="grid grid-cols-2 gap-3">
            <div className="rounded-2xl border border-[#1f1b16]/10 bg-white/55 p-4">
              <p className="text-xs font-black uppercase tracking-[0.18em] text-[#7a6b5c]">
                Rutinas
              </p>
              <p className="mt-2 text-3xl font-black">{routines.length}</p>
            </div>
            <div className="rounded-2xl border border-[#1f1b16]/10 bg-white/55 p-4">
              <p className="text-xs font-black uppercase tracking-[0.18em] text-[#7a6b5c]">
                Ejercicios
              </p>
              <p className="mt-2 text-3xl font-black">{totalExercises}</p>
            </div>
          </div>

          <button
            className="w-full rounded-2xl bg-[#265c52] px-4 py-3 text-sm font-black text-white shadow-[0_12px_25px_rgba(38,92,82,0.22)] transition hover:bg-[#1f1b16] disabled:cursor-not-allowed disabled:opacity-60"
            type="button"
            onClick={() => setIsAIFormOpen((current) => !current)}
          >
            Crear rutina con IA
          </button>
        </div>
      </div>

      {isAIFormOpen && (
        <form
          className="grid gap-4 rounded-[2rem] border border-[#265c52]/20 bg-[#ecf5ef] p-5 shadow-[0_18px_45px_rgba(47,39,27,0.10)] lg:grid-cols-[minmax(0,1.4fr)_160px]"
          onSubmit={handleGenerateAIRoutine}
        >
          <label className="block">
            <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
              Objetivo
            </span>
            <input
              className="mt-2 w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition focus:border-[#265c52]"
              required
              value={aiObjective}
              onChange={(event) => setAIObjective(event.target.value)}
            />
          </label>

          <label className="block">
            <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
              Minutos
            </span>
            <input
              className="mt-2 w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition focus:border-[#265c52]"
              min={15}
              required
              type="number"
              value={aiDuration}
              onChange={(event) => setAIDuration(Number(event.target.value))}
            />
          </label>

          <label className="block lg:col-span-2">
            <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
              Musculos objetivo
            </span>
            <input
              className="mt-2 w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition focus:border-[#265c52]"
              placeholder="chest, back, legs"
              value={aiMuscleGroups}
              onChange={(event) => setAIMuscleGroups(event.target.value)}
            />
          </label>

          <label className="block lg:col-span-2">
            <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
              Ejercicios obligatorios
            </span>
            <input
              className="mt-2 w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition focus:border-[#265c52]"
              placeholder="IDs separados por coma"
              value={aiMandatoryExerciseIDs}
              onChange={(event) =>
                setAIMandatoryExerciseIDs(event.target.value)
              }
            />
          </label>

          <div className="flex flex-col gap-3 lg:col-span-2 sm:flex-row sm:items-center">
            <button
              className="rounded-2xl bg-[#1f1b16] px-5 py-3 text-sm font-black text-white transition hover:bg-[#265c52] disabled:cursor-not-allowed disabled:opacity-60"
              disabled={aiStatus === "loading"}
              type="submit"
            >
              {aiStatus === "loading" ? "Generando..." : "Generar vista previa"}
            </button>
            {aiMessage && (
              <p
                className={`text-sm font-black ${
                  aiStatus === "error" ? "text-[#9b2d20]" : "text-[#265c52]"
                }`}
              >
                {aiMessage}
              </p>
            )}
          </div>
        </form>
      )}

      <AIRoutinePreviewModal
        isOpen={isAIPreviewOpen}
        isSaving={isAIPreviewSaving}
        errorMessage={aiPreviewError}
        routine={aiPreviewRoutine}
        onClose={handleCloseAIRoutinePreview}
        onConfirm={() => void handleConfirmAIRoutine()}
      />

      <div className="grid gap-6 xl:grid-cols-[minmax(0,0.95fr)_minmax(420px,1.05fr)]">
        <div className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/88 p-5 shadow-[0_18px_45px_rgba(47,39,27,0.10)]">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-2xl font-black">
            Rutinas guardadas
          </h2>
          <button
            className="rounded-2xl border border-[#1f1b16]/10 px-4 py-2 text-sm font-black text-[#265c52] transition hover:border-[#265c52] hover:bg-[#265c52] hover:text-white"
            type="button"
            onClick={() => void fetchRoutines()}
          >
            Actualizar
          </button>
        </div>

        {status === "loading" && (
          <p className="rounded-2xl bg-white/55 p-5 text-sm font-bold text-[#5d5348]">
            Cargando rutinas...
          </p>
        )}

        {status === "error" && (
          <p className="rounded-2xl border border-[#9b2d20]/20 bg-[#fff0ed] p-5 text-sm font-bold text-[#9b2d20]">
            No se pudieron cargar las rutinas.
          </p>
        )}

        {status === "success" && routines.length === 0 && (
          <p className="rounded-2xl bg-white/55 p-5 text-sm font-bold text-[#5d5348]">
            Todavia no tienes rutinas guardadas.
          </p>
        )}

        {status === "success" && routines.length > 0 && (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-1">
            {routines.map((routine) => (
              <button
                className={`rounded-2xl border p-5 text-left shadow-[0_12px_28px_rgba(47,39,27,0.08)] transition hover:border-[#265c52] hover:bg-white/85 ${
                  selectedRoutineID === routine.id
                    ? "border-[#265c52] bg-white/90"
                    : "border-[#1f1b16]/10 bg-white/60"
                }`}
                key={routine.id}
                type="button"
                onClick={() => void fetchRoutineDetail(routine.id)}
              >
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <h3 className="text-lg font-black text-[#1f1b16]">
                      {routine.name}
                    </h3>
                    <p className="mt-1 text-xs font-bold uppercase tracking-[0.16em] text-[#7a6b5c]">
                      Actualizada {formatRoutineDate(routine.updated_at)}
                    </p>
                  </div>
                  <span className="rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black text-[#265c52]">
                    {routine.exercise_count} ejercicios
                  </span>
                </div>

                <p className="mt-4 text-sm font-semibold leading-6 text-[#5d5348]">
                  {routine.description?.trim() ||
                    "Rutina sin descripcion guardada."}
                </p>
              </button>
            ))}
          </div>
        )}
        </div>

        <aside className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/88 p-5 shadow-[0_18px_45px_rgba(47,39,27,0.10)]">
          <h2 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-2xl font-black">
            Detalle
          </h2>

          {detailStatus === "idle" && (
            <p className="mt-4 rounded-2xl bg-white/55 p-5 text-sm font-bold text-[#5d5348]">
              Selecciona una rutina para ver sus ejercicios.
            </p>
          )}

          {detailStatus === "loading" && (
            <p className="mt-4 rounded-2xl bg-white/55 p-5 text-sm font-bold text-[#5d5348]">
              Cargando detalle...
            </p>
          )}

          {detailStatus === "error" && (
            <p className="mt-4 rounded-2xl border border-[#9b2d20]/20 bg-[#fff0ed] p-5 text-sm font-bold text-[#9b2d20]">
              No se pudo cargar el detalle de la rutina.
            </p>
          )}

          {detailStatus === "success" && selectedRoutine && (
            <div className="mt-4 space-y-4">
              <div className="rounded-2xl bg-white/60 p-5">
                <p className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                  {selectedRoutine.source}
                </p>
                <h3 className="mt-2 text-2xl font-black">
                  {selectedRoutine.name}
                </h3>
                <p className="mt-2 text-sm font-semibold leading-6 text-[#5d5348]">
                  {selectedRoutine.description?.trim() ||
                    "Rutina sin descripcion guardada."}
                </p>
              </div>

              {selectedRoutine.exercises.length === 0 ? (
                <p className="rounded-2xl bg-white/55 p-5 text-sm font-bold text-[#5d5348]">
                  Esta rutina no tiene ejercicios guardados.
                </p>
              ) : (
                <div className="space-y-3">
                  {selectedRoutine.exercises.map((exercise) => (
                    <article
                      className="rounded-2xl border border-[#1f1b16]/10 bg-white/65 p-4"
                      key={exercise.id}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div>
                          <h4 className="font-black">{exercise.name}</h4>
                          <p className="mt-1 text-xs font-bold uppercase tracking-[0.14em] text-[#7a6b5c]">
                            {exercise.muscle_group}
                            {exercise.exercise_type
                              ? ` · ${exercise.exercise_type}`
                              : ""}
                          </p>
                        </div>
                        <span className="rounded-full bg-[#1f1b16]/5 px-3 py-1 text-xs font-black">
                          #{exercise.exercise_order}
                        </span>
                      </div>

                      {exercise.notes && (
                        <p className="mt-3 text-sm font-semibold text-[#5d5348]">
                          {exercise.notes}
                        </p>
                      )}

                      {exercise.sets.length > 0 && (
                        <div className="mt-4 overflow-x-auto">
                          <table className="w-full min-w-[520px] text-left text-sm">
                            <thead className="text-xs uppercase tracking-[0.14em] text-[#7a6b5c]">
                              <tr>
                                <th className="py-2 pr-3">Serie</th>
                                <th className="py-2 pr-3">Reps</th>
                                <th className="py-2 pr-3">Peso</th>
                                <th className="py-2 pr-3">RIR</th>
                                <th className="py-2 pr-3">Descanso</th>
                              </tr>
                            </thead>
                            <tbody className="font-bold text-[#1f1b16]">
                              {exercise.sets.map((set) => (
                                <tr
                                  className="border-t border-[#1f1b16]/10"
                                  key={set.id}
                                >
                                  <td className="py-2 pr-3">
                                    {set.set_number}
                                  </td>
                                  <td className="py-2 pr-3">
                                    {formatReps(set)}
                                  </td>
                                  <td className="py-2 pr-3">
                                    {set.target_weight_kg != null
                                      ? `${set.target_weight_kg} kg`
                                      : "-"}
                                  </td>
                                  <td className="py-2 pr-3">
                                    {set.target_rir ?? "-"}
                                  </td>
                                  <td className="py-2 pr-3">
                                    {set.rest_seconds != null
                                      ? `${set.rest_seconds}s`
                                      : "-"}
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      )}
                    </article>
                  ))}
                </div>
              )}
            </div>
          )}
        </aside>
      </div>
    </section>
  );
}

function formatReps(set: RoutineSetDetail) {
  if (set.target_reps_text?.trim()) {
    return set.target_reps_text;
  }
  if (set.target_reps_min != null && set.target_reps_max != null) {
    return `${set.target_reps_min}-${set.target_reps_max}`;
  }
  if (set.target_reps_min != null) {
    return `${set.target_reps_min}+`;
  }
  if (set.target_duration_seconds != null) {
    return `${set.target_duration_seconds}s`;
  }
  if (set.target_distance_km != null) {
    return `${set.target_distance_km} km`;
  }
  return "-";
}

function splitCommaList(value: string) {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}
