import { type FormEvent, type ReactNode, useCallback, useEffect, useState } from "react";
import { apiUrl } from "../lib/api";
import AIRoutinePreviewModal, {
  type AIRoutinePreview,
  type AIRoutinePreviewSet,
} from "../components/Routine/AIRoutinePreviewModal";
import type { Exercise } from "../types/exercise";

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

type AIRoutineUpgradeDiffSet = {
  change_type: string;
  set_number: number;
  before?: AIRoutinePreviewSet;
  after?: AIRoutinePreviewSet;
};

type AIRoutineUpgradeDiffExercise = {
  change_type: string;
  exercise_id: string;
  before_name?: string;
  after_name?: string;
  before_order?: number;
  after_order?: number;
  before_muscle_group?: string;
  after_muscle_group?: string;
  before_exercise_type?: string;
  after_exercise_type?: string;
  is_mandatory_changed?: boolean;
  sets?: AIRoutineUpgradeDiffSet[];
};

type AIRoutineUpgradeDiff = {
  summary: {
    added_exercises: number;
    removed_exercises: number;
    modified_exercises: number;
    unchanged_exercises: number;
  };
  exercises: AIRoutineUpgradeDiffExercise[];
};

type AIRoutineUpgradeResponse = {
  routine_json: AIRoutinePreview;
  diff: AIRoutineUpgradeDiff;
  rate_limit: {
    limit?: number;
    remaining: number;
    reset_at: string;
  };
};

type AIRoutineUpgradeStatus = "idle" | "loading" | "success" | "error";
type AIRoutineUpgradePersistAction = "idle" | "save_as_new" | "overwrite";

type RoutineStatus = "idle" | "loading" | "success" | "error";

type ExerciseListResponse = {
  items: Exercise[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
};

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
  const [aiDuration, setAIDuration] = useState("60");
  const [aiMuscleGroups, setAIMuscleGroups] = useState("");
  const [aiNotes, setAINotes] = useState("");
  const [exerciseSearch, setExerciseSearch] = useState("");
  const [selectedMandatoryExercises, setSelectedMandatoryExercises] = useState<
    string[]
  >([]);
  const [availableExercises, setAvailableExercises] = useState<Exercise[]>([]);
  const [availableExercisesStatus, setAvailableExercisesStatus] =
    useState<RoutineStatus>("idle");
  const [availableExercisesMessage, setAvailableExercisesMessage] = useState("");
  const [aiStatus, setAIStatus] = useState<RoutineStatus>("idle");
  const [aiMessage, setAIMessage] = useState("");
  const [aiPreviewRoutine, setAIPreviewRoutine] =
    useState<AIRoutinePreview | null>(null);
  const [isAIPreviewOpen, setIsAIPreviewOpen] = useState(false);
  const [isAIPreviewSaving, setIsAIPreviewSaving] = useState(false);
  const [aiPreviewError, setAIPreviewError] = useState("");
  const [isAIUpgradeOpen, setIsAIUpgradeOpen] = useState(false);
  const [aiUpgradeMessage, setAIUpgradeMessage] = useState("");
  const [aiUpgradeStatus, setAIUpgradeStatus] =
    useState<AIRoutineUpgradeStatus>("idle");
  const [aiUpgradeNotice, setAIUpgradeNotice] = useState("");
  const [aiUpgradeFeedback, setAIUpgradeFeedback] = useState("");
  const [aiUpgradeResult, setAIUpgradeResult] =
    useState<AIRoutineUpgradeResponse | null>(null);
  const [isAIUpgradeResultOpen, setIsAIUpgradeResultOpen] = useState(false);
  const [aiUpgradePersistAction, setAIUpgradePersistAction] =
    useState<AIRoutineUpgradePersistAction>("idle");

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

  const fetchAvailableExercises = useCallback(async () => {
    setAvailableExercisesStatus("loading");
    setAvailableExercisesMessage("");

    try {
      const params = new URLSearchParams({
        page: "1",
        limit: "100",
      });

      const response = await fetch(
        apiUrl(`/api/exercises?${params.toString()}`),
        {
          credentials: "include",
        },
      );

      if (!response.ok) {
        setAvailableExercises([]);
        setAvailableExercisesStatus("error");
        setAvailableExercisesMessage(
          "No se pudieron cargar los ejercicios disponibles.",
        );
        return;
      }

      const payload = (await response.json()) as ExerciseListResponse;
      setAvailableExercises(payload.items);
      setAvailableExercisesStatus("success");
    } catch {
      setAvailableExercises([]);
      setAvailableExercisesStatus("error");
      setAvailableExercisesMessage(
        "No se pudieron cargar los ejercicios disponibles.",
      );
    }
  }, []);

  useEffect(() => {
    void fetchRoutines();
  }, [fetchRoutines]);

  useEffect(() => {
    if (!isAIFormOpen) {
      return;
    }

    if (availableExercisesStatus === "loading" || availableExercisesStatus === "success") {
      return;
    }

    void fetchAvailableExercises();
  }, [availableExercisesStatus, fetchAvailableExercises, isAIFormOpen]);

  const toggleMandatoryExercise = (exerciseName: string) => {
    const normalizedName = exerciseName.trim();
    if (normalizedName === "") {
      return;
    }

    setSelectedMandatoryExercises((current) =>
      current.some(
        (currentName) => currentName.toLowerCase() === normalizedName.toLowerCase(),
      )
        ? current.filter(
            (currentName) =>
              currentName.toLowerCase() !== normalizedName.toLowerCase(),
          )
        : [...current, normalizedName],
    );
  };

  const handleGenerateAIRoutine = async (
    event: FormEvent<HTMLFormElement>,
  ) => {
    event.preventDefault();

    const durationMinutes = Number.parseInt(aiDuration.trim(), 10);
    if (!Number.isFinite(durationMinutes) || durationMinutes <= 0) {
      setAIStatus("error");
      setAIMessage("Introduce unos minutos válidos.");
      return;
    }

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
            duration_minutes: durationMinutes,
            target_muscle_groups: splitCommaList(aiMuscleGroups),
            mandatory_exercises: selectedMandatoryExercises,
            notes: aiNotes.trim(),
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

  const handleOpenAIUpgrade = () => {
    if (selectedRoutine == null) {
      return;
    }

    setAIUpgradeStatus("idle");
    setAIUpgradeNotice("");
    setAIUpgradeFeedback("");
    setAIUpgradeMessage(
      `Mantén ${selectedRoutine.name} y mejora el equilibrio general de la sesión.`,
    );
    setIsAIUpgradeOpen(true);
  };

  const handleCloseAIUpgrade = () => {
    if (aiUpgradeStatus === "loading") {
      return;
    }

    setIsAIUpgradeOpen(false);
    setAIUpgradeStatus("idle");
    setAIUpgradeNotice("");
  };

  const handleSubmitAIUpgrade = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (selectedRoutine == null) {
      setAIUpgradeStatus("error");
      setAIUpgradeNotice("Selecciona una rutina antes de continuar.");
      return;
    }

    if (aiUpgradeMessage.trim() === "") {
      setAIUpgradeStatus("error");
      setAIUpgradeNotice("Escribe al menos una instrucción para la IA.");
      return;
    }

    setAIUpgradeStatus("loading");
    setAIUpgradeNotice("");

    try {
      const response = await fetch(
        apiUrl(`/api/routines/${selectedRoutine.id}/ai/upgrade`),
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            message: aiUpgradeMessage.trim(),
          }),
        },
      );

      const payload = (await response.json()) as AIRoutineUpgradeResponse & {
        error?: string;
        detail?: string;
      };

      if (!response.ok) {
        setAIUpgradeStatus("error");
        setAIUpgradeNotice(
          payload.detail ||
            payload.error ||
            "No se pudo preparar la mejora de la rutina.",
        );
        return;
      }

      if (!payload.routine_json || !payload.diff) {
        setAIUpgradeStatus("error");
        setAIUpgradeNotice("No se recibio una propuesta valida.");
        return;
      }

      setAIUpgradeResult(normalizeAIRoutineUpgradeResponse(payload));
      setAIUpgradeStatus("success");
      setAIUpgradeFeedback("");
      setIsAIUpgradeOpen(false);
      setIsAIUpgradeResultOpen(true);
    } catch {
      setAIUpgradeStatus("error");
      setAIUpgradeNotice("No se pudo conectar con el servicio.");
    }
  };

  const handleCloseAIUpgradeResult = () => {
    if (aiUpgradePersistAction !== "idle") {
      return;
    }

    setIsAIUpgradeResultOpen(false);
    setAIUpgradeResult(null);
    setAIUpgradeFeedback("");
  };

  const handleRegenerateAIUpgrade = async () => {
    if (selectedRoutine == null || aiUpgradeMessage.trim() === "") {
      setAIUpgradeStatus("error");
      setAIUpgradeNotice("Faltan datos para regenerar la propuesta.");
      return;
    }

    if (aiUpgradeFeedback.trim() === "") {
      setAIUpgradeStatus("error");
      setAIUpgradeNotice("Escribe una indicación para ajustar la propuesta.");
      return;
    }

    setAIUpgradeStatus("loading");
    setAIUpgradeNotice("");

    try {
      const response = await fetch(
        apiUrl(`/api/routines/${selectedRoutine.id}/ai/upgrade`),
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            message: aiUpgradeMessage.trim(),
            feedback_message: aiUpgradeFeedback.trim(),
          }),
        },
      );

      const payload = (await response.json()) as AIRoutineUpgradeResponse & {
        error?: string;
        detail?: string;
      };

      if (!response.ok) {
        setAIUpgradeStatus("error");
        setAIUpgradeNotice(
          payload.detail ||
            payload.error ||
            "No se pudo regenerar la mejora de la rutina.",
        );
        return;
      }

      if (!payload.routine_json || !payload.diff) {
        setAIUpgradeStatus("error");
        setAIUpgradeNotice("No se recibio una propuesta valida.");
        return;
      }

      setAIUpgradeResult(normalizeAIRoutineUpgradeResponse(payload));
      setAIUpgradeStatus("success");
      setAIUpgradeNotice("Propuesta actualizada.");
    } catch {
      setAIUpgradeStatus("error");
      setAIUpgradeNotice("No se pudo conectar con el servicio.");
    }
  };

  const handleSaveUpgradedRoutineAsNew = async () => {
    if (selectedRoutine == null || aiUpgradeResult == null) {
      return;
    }

    setAIUpgradePersistAction("save_as_new");
    setAIUpgradeNotice("");

    try {
      const response = await fetch(
        apiUrl(`/api/routines/${selectedRoutine.id}/ai/save-as-new`),
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            routine_json: aiUpgradeResult.routine_json,
          }),
        },
      );

      const payload = (await response.json()) as AIRoutineSaveResponse & {
        error?: string;
        detail?: string;
      };

      if (!response.ok) {
        setAIUpgradeNotice(
          payload.detail ||
            payload.error ||
            "No se pudo guardar la propuesta como nueva rutina.",
        );
        return;
      }

      setAIUpgradeNotice("");
      setIsAIUpgradeResultOpen(false);
      setAIUpgradeResult(null);
      setAIUpgradeFeedback("");
      setAIMessage("Rutina guardada como nueva.");
      await fetchRoutines();
      if (payload.routine_id) {
        await fetchRoutineDetail(payload.routine_id);
      }
    } catch {
      setAIUpgradeNotice("No se pudo conectar con el servicio.");
    } finally {
      setAIUpgradePersistAction("idle");
    }
  };

  const handleOverwriteRoutineWithUpgrade = async () => {
    if (selectedRoutine == null || aiUpgradeResult == null) {
      return;
    }

    setAIUpgradePersistAction("overwrite");
    setAIUpgradeNotice("");

    try {
      const response = await fetch(
        apiUrl(`/api/routines/${selectedRoutine.id}/ai/overwrite`),
        {
          method: "PUT",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            routine_json: aiUpgradeResult.routine_json,
          }),
        },
      );

      const payload = (await response.json()) as AIRoutineSaveResponse & {
        error?: string;
        detail?: string;
      };

      if (!response.ok) {
        setAIUpgradeNotice(
          payload.detail ||
            payload.error ||
            "No se pudo sobrescribir la rutina actual.",
        );
        return;
      }

      setAIUpgradeNotice("");
      setIsAIUpgradeResultOpen(false);
      setAIUpgradeResult(null);
      setAIUpgradeFeedback("");
      setAIMessage("Rutina actualizada.");
      await fetchRoutines();
      await fetchRoutineDetail(selectedRoutine.id);
    } catch {
      setAIUpgradeNotice("No se pudo conectar con el servicio.");
    } finally {
      setAIUpgradePersistAction("idle");
    }
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
          className="grid gap-4 rounded-[2rem] border border-[#265c52]/20 bg-[#ecf5ef] p-5 shadow-[0_18px_45px_rgba(47,39,27,0.10)] lg:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]"
          onSubmit={handleGenerateAIRoutine}
        >
          <div className="grid gap-4 lg:col-span-2 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
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
                inputMode="numeric"
                pattern="[0-9]*"
                type="text"
                value={aiDuration}
                onChange={(event) =>
                  setAIDuration(event.target.value.replace(/[^\d]/g, ""))
                }
              />
            </label>
          </div>

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
              Notas para la IA
            </span>
            <textarea
              className="mt-2 min-h-28 w-full resize-y rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-semibold leading-6 outline-none transition placeholder:font-medium focus:border-[#265c52]"
              placeholder="Ej. prioriza press con barra, evita sentadillas traseras y mantén descansos largos."
              value={aiNotes}
              onChange={(event) => setAINotes(event.target.value)}
            />
          </label>

          <section className="lg:col-span-2 rounded-[1.75rem] border border-[#1f1b16]/10 bg-white/70 p-4 shadow-[0_12px_28px_rgba(47,39,27,0.06)]">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                  Ejercicios obligatorios
                </span>
                <p className="mt-2 text-sm font-semibold leading-6 text-[#5d5348]">
                  Selecciona ejercicios por nombre. La IA recibira exactamente
                  esos nombres como referencia.
                </p>
              </div>
              <span className="rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black text-[#265c52]">
                {selectedMandatoryExercises.length} seleccionados
              </span>
            </div>

            <label className="mt-4 block">
              <span className="sr-only">Buscar ejercicios</span>
              <input
                className="w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition placeholder:font-medium focus:border-[#265c52]"
                placeholder="Buscar ejercicios por nombre o musculo"
                value={exerciseSearch}
                onChange={(event) => setExerciseSearch(event.target.value)}
              />
            </label>

            {selectedMandatoryExercises.length > 0 && (
              <div className="mt-4 flex flex-wrap gap-2">
                {selectedMandatoryExercises.map((exerciseName) => (
                  <button
                    key={exerciseName}
                    className="rounded-full border border-[#265c52]/20 bg-[#ecf5ef] px-3 py-1 text-xs font-black text-[#265c52] transition hover:bg-[#dcefe4]"
                    type="button"
                    onClick={() => toggleMandatoryExercise(exerciseName)}
                  >
                    {exerciseName} ×
                  </button>
                ))}
              </div>
            )}

            <div className="mt-4 max-h-72 overflow-y-auto rounded-2xl border border-[#1f1b16]/10 bg-[#fffaf0] p-3">
              {availableExercisesStatus === "loading" && (
                <p className="rounded-2xl bg-white/70 p-4 text-sm font-semibold text-[#5d5348]">
                  Cargando ejercicios disponibles...
                </p>
              )}

              {availableExercisesStatus === "error" && (
                <p className="rounded-2xl border border-[#9b2d20]/20 bg-[#fff0ed] p-4 text-sm font-semibold text-[#9b2d20]">
                  {availableExercisesMessage}
                </p>
              )}

              {availableExercisesStatus === "success" && (
                <>
                  {availableExercises
                    .filter((exercise) => {
                      const search = exerciseSearch.trim().toLowerCase();
                      if (search === "") {
                        return true;
                      }

                      return (
                        exercise.name.toLowerCase().includes(search) ||
                        exercise.muscle_group.toLowerCase().includes(search) ||
                        (exercise.exercise_type ?? "")
                          .toLowerCase()
                          .includes(search)
                      );
                    })
                    .length === 0 ? (
                    <p className="rounded-2xl bg-white/70 p-4 text-sm font-semibold text-[#5d5348]">
                      No hay ejercicios que coincidan con el filtro.
                    </p>
                  ) : (
                    <div className="space-y-2">
                      {availableExercises
                        .filter((exercise) => {
                          const search = exerciseSearch.trim().toLowerCase();
                          if (search === "") {
                            return true;
                          }

                          return (
                            exercise.name.toLowerCase().includes(search) ||
                            exercise.muscle_group.toLowerCase().includes(search) ||
                            (exercise.exercise_type ?? "")
                              .toLowerCase()
                              .includes(search)
                          );
                        })
                        .map((exercise) => {
                          const isSelected = selectedMandatoryExercises.some(
                            (name) =>
                              name.toLowerCase() === exercise.name.toLowerCase(),
                          );

                          return (
                            <button
                              key={exercise.id}
                              className={`flex w-full items-start justify-between gap-4 rounded-2xl border px-4 py-3 text-left transition ${
                                isSelected
                                  ? "border-[#265c52] bg-[#ecf5ef]"
                                  : "border-[#1f1b16]/10 bg-white/90 hover:border-[#265c52]/30 hover:bg-white"
                              }`}
                              type="button"
                              onClick={() => toggleMandatoryExercise(exercise.name)}
                            >
                              <div className="min-w-0">
                                <p className="break-words text-sm font-black text-[#1f1b16]">
                                  {exercise.name}
                                </p>
                                <p className="mt-1 text-[11px] font-bold uppercase tracking-[0.14em] text-[#7a6b5c]">
                                  {exercise.muscle_group}
                                  {exercise.exercise_type
                                    ? ` · ${exercise.exercise_type}`
                                    : ""}
                                </p>
                              </div>
                              <span
                                className={`shrink-0 rounded-full px-3 py-1 text-xs font-black ${
                                  isSelected
                                    ? "bg-[#265c52] text-white"
                                    : "bg-[#265c52]/10 text-[#265c52]"
                                }`}
                              >
                                {isSelected ? "Quitar" : "Añadir"}
                              </span>
                            </button>
                          );
                        })}
                    </div>
                  )}
                </>
              )}
            </div>
          </section>

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
                <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                  <div>
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

                  <button
                    className="inline-flex shrink-0 items-center justify-center rounded-2xl bg-[#ea7130] px-4 py-3 text-sm font-black text-white shadow-[0_12px_26px_rgba(234,113,48,0.26)] transition hover:bg-[#1f1b16]"
                    type="button"
                    onClick={handleOpenAIUpgrade}
                  >
                    Mejorar con IA
                  </button>
                </div>
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

      {isAIUpgradeOpen && selectedRoutine && (
        <RoutineDialogShell
          kicker="AI Upgrade"
          kickerColor="#ea7130"
          title={`Mejorar ${selectedRoutine.name}`}
          maxWidthClass="max-w-3xl"
          onClose={handleCloseAIUpgrade}
        >
          <form className="space-y-4" onSubmit={handleSubmitAIUpgrade}>
            <div>
              <p className="text-sm font-semibold leading-6 text-[#5d5348]">
                Describe qué te gustaría ajustar en esta rutina para generar
                una nueva propuesta sin tocar todavía la versión guardada.
              </p>
            </div>

            <div className="grid gap-3 sm:grid-cols-[minmax(0,1.15fr)_minmax(220px,0.85fr)]">
              <div className="rounded-2xl border border-[#1f1b16]/10 bg-[#fff4ea] p-4">
                <p className="text-xs font-black uppercase tracking-[0.18em] text-[#ea7130]">
                  Rutina seleccionada
                </p>
                <p className="mt-2 text-lg font-black text-[#1f1b16]">
                  {selectedRoutine.name}
                </p>
                <p className="mt-2 text-sm font-semibold text-[#5d5348]">
                  {selectedRoutine.exercises.length} ejercicios cargados para
                  preparar la mejora.
                </p>
              </div>

              <div className="rounded-2xl border border-[#265c52]/12 bg-[#ecf5ef] p-4">
                <p className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                  Qué puedes pedir
                </p>
                <p className="mt-2 text-sm font-semibold leading-6 text-[#355249]">
                  Más intensidad, menos fatiga, mejor equilibrio muscular,
                  cambios de orden o nuevas variantes.
                </p>
              </div>
            </div>

            <label className="block">
              <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                Instrucciones para la IA
              </span>
              <textarea
                className="mt-2 min-h-36 w-full resize-y rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-semibold leading-6 outline-none transition placeholder:font-medium focus:border-[#ea7130]"
                placeholder="Ej. mantén Bench Press, reduce fatiga en hombro y añade más trabajo unilateral."
                value={aiUpgradeMessage}
                onChange={(event) => setAIUpgradeMessage(event.target.value)}
              />
            </label>

            {aiUpgradeNotice && (
              <p
                className={`text-sm font-black ${
                  aiUpgradeStatus === "error"
                    ? "text-[#9b2d20]"
                    : "text-[#265c52]"
                }`}
              >
                {aiUpgradeNotice}
              </p>
            )}

            <div className="flex flex-col gap-3 sm:flex-row sm:justify-end">
              <button
                className="rounded-2xl border border-[#1f1b16]/10 px-4 py-3 text-sm font-black text-[#1f1b16] transition hover:border-[#1f1b16] hover:bg-[#1f1b16] hover:text-white"
                type="button"
                onClick={handleCloseAIUpgrade}
              >
                Cerrar
              </button>
              <button
                className="rounded-2xl bg-[#ea7130] px-5 py-3 text-sm font-black text-white transition hover:bg-[#1f1b16] disabled:cursor-not-allowed disabled:opacity-60"
                disabled={aiUpgradeStatus === "loading"}
                type="submit"
              >
                {aiUpgradeStatus === "loading"
                  ? "Preparando..."
                  : "Preparar mejora"}
              </button>
            </div>
          </form>
        </RoutineDialogShell>
      )}

      {isAIUpgradeResultOpen && aiUpgradeResult && (
        <RoutineDialogShell
          kicker="Upgrade Preview"
          kickerColor="#ea7130"
          title={aiUpgradeResult.routine_json.name || "Propuesta de mejora"}
          maxWidthClass="max-w-5xl"
          onClose={handleCloseAIUpgradeResult}
        >
          <div className="space-y-5">
            <div className="grid gap-3 sm:grid-cols-4">
              <DiffStatCard
                label="Añadidos"
                tone="added"
                value={aiUpgradeResult.diff.summary.added_exercises}
              />
              <DiffStatCard
                label="Eliminados"
                tone="removed"
                value={aiUpgradeResult.diff.summary.removed_exercises}
              />
              <DiffStatCard
                label="Modificados"
                tone="modified"
                value={aiUpgradeResult.diff.summary.modified_exercises}
              />
              <DiffStatCard
                label="Sin cambios"
                tone="unchanged"
                value={aiUpgradeResult.diff.summary.unchanged_exercises}
              />
            </div>

            <div className="grid gap-4 lg:grid-cols-2">
              <RoutineUpgradeSnapshotCard
                badge="Original"
                badgeTone="neutral"
                title={selectedRoutine?.name || "Rutina actual"}
                description={
                  selectedRoutine?.description?.trim() ||
                  "Rutina base seleccionada para preparar la mejora."
                }
                exerciseCount={selectedRoutine?.exercises.length ?? 0}
              />
              <RoutineUpgradeSnapshotCard
                badge="Propuesta"
                badgeTone="accent"
                title={aiUpgradeResult.routine_json.name}
                description={
                  aiUpgradeResult.routine_json.objective || "Sin objetivo"
                }
                exerciseCount={aiUpgradeResult.routine_json.exercises?.length ?? 0}
              />
            </div>

            <div className="rounded-2xl border border-[#1f1b16]/10 bg-[#fffdf8] p-4">
              <p className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                Cambios detectados
              </p>

              <div className="mt-4 space-y-3">
                {(aiUpgradeResult.diff.exercises ?? []).map((exercise) => (
                  <article
                    key={`${exercise.exercise_id}-${exercise.change_type}-${exercise.before_order ?? exercise.after_order ?? 0}`}
                    className={`rounded-2xl border px-4 py-3 ${diffEntryStyles(exercise.change_type)}`}
                  >
                    <div className="grid gap-4 lg:grid-cols-[minmax(0,0.7fr)_minmax(0,1.3fr)]">
                      <div>
                        <p className="text-xs font-black uppercase tracking-[0.16em]">
                          {diffEntryLabel(exercise.change_type)}
                        </p>
                        <h5 className="mt-1 text-base font-black">
                          {diffExerciseTitle(exercise)}
                        </h5>
                        <p className="mt-1 text-sm font-semibold opacity-80">
                          {(exercise.after_muscle_group ||
                            exercise.before_muscle_group ||
                            "").trim()}
                          {(exercise.after_exercise_type || exercise.before_exercise_type)
                            ? ` · ${exercise.after_exercise_type || exercise.before_exercise_type}`
                            : ""}
                        </p>
                        <div className="mt-3 flex flex-wrap gap-2">
                          <span className="rounded-full bg-white/70 px-3 py-1 text-xs font-black text-[#1f1b16]">
                            {formatDiffOrder(
                              exercise.before_order,
                              exercise.after_order,
                            )}
                          </span>
                          {exercise.is_mandatory_changed && (
                            <span className="rounded-full bg-white/70 px-3 py-1 text-xs font-black text-[#1f1b16]">
                              Obligatorio actualizado
                            </span>
                          )}
                        </div>
                      </div>

                      <div className="rounded-xl bg-white/70 p-3 text-sm font-semibold text-[#1f1b16]">
                        <div className="grid gap-3 sm:grid-cols-2">
                          <div className="rounded-xl border border-[#1f1b16]/8 bg-white/80 p-3">
                            <p className="text-[11px] font-black uppercase tracking-[0.16em] text-[#7a6b5c]">
                              Antes
                            </p>
                            <p className="mt-2">
                              {formatExerciseSideLabel(
                                exercise.before_name,
                                exercise.change_type,
                                "before",
                              )}
                            </p>
                            {exercise.sets &&
                              exercise.sets.some(
                                (set) => set.change_type !== "added",
                              ) && (
                                <div className="mt-3 space-y-2">
                                  {exercise.sets
                                    .filter((set) => set.change_type !== "added")
                                    .map((set) => (
                                      <p key={`${exercise.exercise_id}-before-${set.set_number}`}>
                                        <span className="font-black">
                                          Serie {set.set_number}:
                                        </span>{" "}
                                        {formatPreviewSetDiffSide(set.before)}
                                      </p>
                                    ))}
                                </div>
                              )}
                          </div>

                          <div className="rounded-xl border border-[#1f1b16]/8 bg-white/80 p-3">
                            <p className="text-[11px] font-black uppercase tracking-[0.16em] text-[#7a6b5c]">
                              Ahora
                            </p>
                            <p className="mt-2">
                              {formatExerciseSideLabel(
                                exercise.after_name,
                                exercise.change_type,
                                "after",
                              )}
                            </p>
                            {exercise.sets &&
                              exercise.sets.some(
                                (set) => set.change_type !== "removed",
                              ) && (
                                <div className="mt-3 space-y-2">
                                  {exercise.sets
                                    .filter((set) => set.change_type !== "removed")
                                    .map((set) => (
                                      <p key={`${exercise.exercise_id}-after-${set.set_number}`}>
                                        <span className="font-black">
                                          Serie {set.set_number}:
                                        </span>{" "}
                                        {formatPreviewSetDiffSide(set.after)}
                                      </p>
                                    ))}
                                </div>
                              )}
                          </div>
                        </div>
                      </div>
                    </div>
                  </article>
                ))}
              </div>
            </div>

            <div className="rounded-2xl border border-[#1f1b16]/10 bg-white/80 p-4">
              <label className="block">
                <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                  Ajustar la propuesta
                </span>
                <textarea
                  className="mt-2 min-h-28 w-full resize-y rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-semibold leading-6 outline-none transition placeholder:font-medium focus:border-[#ea7130]"
                  placeholder="Ej. mantén el trabajo de hombro, pero recorta volumen total y conserva más tirón horizontal."
                  value={aiUpgradeFeedback}
                  onChange={(event) => setAIUpgradeFeedback(event.target.value)}
                />
              </label>

              {aiUpgradeNotice && (
                <p
                  className={`mt-3 text-sm font-black ${
                    aiUpgradeStatus === "error"
                      ? "text-[#9b2d20]"
                      : "text-[#265c52]"
                  }`}
                >
                  {aiUpgradeNotice}
                </p>
              )}

              {aiUpgradeStatus !== "error" && (
                <p className="mt-3 text-sm font-semibold text-[#5d5348]">
                  {formatUpgradeRateLimit(aiUpgradeResult.rate_limit)}
                </p>
              )}
            </div>

            <div className="flex flex-col gap-3 sm:flex-row sm:justify-end">
              <button
                className="rounded-2xl border border-[#ea7130]/20 bg-[#fff4ea] px-5 py-3 text-sm font-black text-[#8a4519] transition hover:border-[#ea7130] hover:bg-[#ffe7d7] disabled:cursor-not-allowed disabled:opacity-60"
                disabled={aiUpgradeStatus === "loading" || aiUpgradePersistAction !== "idle"}
                type="button"
                onClick={() => void handleRegenerateAIUpgrade()}
              >
                {aiUpgradeStatus === "loading"
                  ? "Regenerando..."
                  : "Regenerar propuesta"}
              </button>
              <button
                className="rounded-2xl border border-[#265c52]/20 bg-[#ecf5ef] px-5 py-3 text-sm font-black text-[#265c52] transition hover:border-[#265c52] hover:bg-[#dcefe4] disabled:cursor-not-allowed disabled:opacity-60"
                disabled={aiUpgradeStatus === "loading" || aiUpgradePersistAction !== "idle"}
                type="button"
                onClick={() => void handleSaveUpgradedRoutineAsNew()}
              >
                {aiUpgradePersistAction === "save_as_new"
                  ? "Guardando..."
                  : "Guardar como nueva"}
              </button>
              <button
                className="rounded-2xl bg-[#ea7130] px-5 py-3 text-sm font-black text-white transition hover:bg-[#c95d24] disabled:cursor-not-allowed disabled:opacity-60"
                disabled={aiUpgradeStatus === "loading" || aiUpgradePersistAction !== "idle"}
                type="button"
                onClick={() => void handleOverwriteRoutineWithUpgrade()}
              >
                {aiUpgradePersistAction === "overwrite"
                  ? "Sobrescribiendo..."
                  : "Sobrescribir rutina"}
              </button>
              <button
                className="rounded-2xl bg-[#1f1b16] px-5 py-3 text-sm font-black text-white transition hover:bg-[#265c52]"
                disabled={aiUpgradePersistAction !== "idle"}
                type="button"
                onClick={handleCloseAIUpgradeResult}
              >
                Cerrar vista previa
              </button>
            </div>
          </div>
        </RoutineDialogShell>
      )}
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

function normalizeAIRoutineUpgradeResponse(
  payload: AIRoutineUpgradeResponse,
): AIRoutineUpgradeResponse {
  return {
    ...payload,
    routine_json: {
      ...payload.routine_json,
      name: payload.routine_json?.name ?? "Propuesta de mejora",
      objective: payload.routine_json?.objective ?? "",
      exercises: Array.isArray(payload.routine_json?.exercises)
        ? payload.routine_json.exercises.map((exercise) => ({
            ...exercise,
            sets: Array.isArray(exercise.sets) ? exercise.sets : [],
          }))
        : [],
    },
    diff: {
      summary: {
        added_exercises: payload.diff?.summary?.added_exercises ?? 0,
        removed_exercises: payload.diff?.summary?.removed_exercises ?? 0,
        modified_exercises: payload.diff?.summary?.modified_exercises ?? 0,
        unchanged_exercises: payload.diff?.summary?.unchanged_exercises ?? 0,
      },
      exercises: Array.isArray(payload.diff?.exercises)
        ? payload.diff.exercises.map((exercise) => ({
            ...exercise,
            sets: Array.isArray(exercise.sets) ? exercise.sets : [],
          }))
        : [],
    },
    rate_limit: {
      limit: payload.rate_limit?.limit ?? 2,
      remaining: payload.rate_limit?.remaining ?? 0,
      reset_at: payload.rate_limit?.reset_at ?? "",
    },
  };
}

function formatPreviewSetReps(set: AIRoutinePreviewSet) {
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

function formatPreviewSetDiffSide(set?: AIRoutinePreviewSet) {
  if (!set) {
    return "Sin referencia";
  }

  const parts = [formatPreviewSetReps(set)];

  if (set.target_weight_kg != null) {
    parts.push(`${set.target_weight_kg} kg`);
  }
  if (set.target_rir != null) {
    parts.push(`RIR ${set.target_rir}`);
  }
  if (set.rest_seconds != null) {
    parts.push(`${set.rest_seconds}s descanso`);
  }

  return parts.join(" · ");
}

function formatUpgradeRateLimit(rateLimit: AIRoutineUpgradeResponse["rate_limit"]) {
  if (rateLimit.remaining <= 0) {
    return `Límite alcanzado. Podrás volver a intentarlo tras ${formatRoutineDate(rateLimit.reset_at)}.`;
  }

  return `Quedan ${rateLimit.remaining} intentos en esta hora. Reinicio estimado: ${formatRoutineDate(rateLimit.reset_at)}.`;
}

function diffExerciseTitle(exercise: AIRoutineUpgradeDiffExercise) {
  const beforeName = (exercise.before_name || "").trim();
  const afterName = (exercise.after_name || "").trim();

  if (beforeName !== "" && beforeName === afterName) {
    return beforeName;
  }
  if (beforeName !== "" && afterName !== "") {
    return `${beforeName} → ${afterName}`;
  }

  return afterName || beforeName || "Ejercicio";
}

function formatExerciseSideLabel(
  name: string | undefined,
  changeType: string,
  side: "before" | "after",
) {
  const normalizedName = (name || "").trim();
  if (normalizedName !== "") {
    return normalizedName;
  }
  if (changeType === "added" && side === "before") {
    return "No estaba en la rutina";
  }
  if (changeType === "removed" && side === "after") {
    return "Se elimina de la rutina";
  }

  return "Sin referencia";
}

function diffEntryLabel(changeType: string) {
  switch (changeType) {
    case "added":
      return "Añadido";
    case "removed":
      return "Eliminado";
    case "modified":
      return "Modificado";
    default:
      return "Sin cambios";
  }
}

function diffEntryStyles(changeType: string) {
  switch (changeType) {
    case "added":
      return "border-[#2f7d5f]/20 bg-[#edf8f1] text-[#215541]";
    case "removed":
      return "border-[#9b2d20]/20 bg-[#fff0ed] text-[#7b2419]";
    case "modified":
      return "border-[#ea7130]/20 bg-[#fff4ea] text-[#8a4519]";
    default:
      return "border-[#1f1b16]/10 bg-white/70 text-[#5d5348]";
  }
}

function formatDiffOrder(beforeOrder?: number, afterOrder?: number) {
  if (beforeOrder != null && afterOrder != null && beforeOrder !== afterOrder) {
    return `#${beforeOrder} → #${afterOrder}`;
  }
  if (afterOrder != null) {
    return `#${afterOrder}`;
  }
  if (beforeOrder != null) {
    return `#${beforeOrder}`;
  }
  return "Sin orden";
}

function DiffStatCard({
  label,
  tone,
  value,
}: {
  label: string;
  tone: "added" | "removed" | "modified" | "unchanged";
  value: number;
}) {
  const tones = {
    added: "border-[#2f7d5f]/20 bg-[#edf8f1] text-[#215541]",
    removed: "border-[#9b2d20]/20 bg-[#fff0ed] text-[#7b2419]",
    modified: "border-[#ea7130]/20 bg-[#fff4ea] text-[#8a4519]",
    unchanged: "border-[#1f1b16]/10 bg-white/75 text-[#5d5348]",
  } as const;

  return (
    <div className={`rounded-2xl border p-4 ${tones[tone]}`}>
      <p className="text-[11px] font-black uppercase tracking-[0.18em]">
        {label}
      </p>
      <p className="mt-2 text-2xl font-black">{value}</p>
    </div>
  );
}

function RoutineUpgradeSnapshotCard({
  badge,
  badgeTone,
  title,
  description,
  exerciseCount,
}: {
  badge: string;
  badgeTone: "neutral" | "accent";
  title: string;
  description: string;
  exerciseCount: number;
}) {
  const badgeClass =
    badgeTone === "accent"
      ? "bg-[#ea7130]/10 text-[#ea7130]"
      : "bg-[#1f1b16]/5 text-[#1f1b16]";

  return (
    <div className="rounded-2xl border border-[#1f1b16]/10 bg-white/80 p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-xs font-black uppercase tracking-[0.18em] text-[#7a6b5c]">
            {badge}
          </p>
          <h4 className="mt-2 text-xl font-black text-[#1f1b16]">{title}</h4>
          <p className="mt-2 text-sm font-semibold leading-6 text-[#5d5348]">
            {description}
          </p>
        </div>
        <span className={`rounded-full px-3 py-1 text-xs font-black ${badgeClass}`}>
          {exerciseCount} ejercicios
        </span>
      </div>
    </div>
  );
}

function RoutineDialogShell({
  kicker,
  kickerColor = "#265c52",
  title,
  maxWidthClass = "max-w-2xl",
  onClose,
  children,
}: {
  kicker: string;
  kickerColor?: string;
  title: string;
  maxWidthClass?: string;
  onClose: () => void;
  children: ReactNode;
}) {
  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-[#1f1b16]/45 px-4 backdrop-blur-sm">
      <section
        className={`w-full ${maxWidthClass} max-h-[calc(100vh-2rem)] overflow-hidden rounded-[28px] border border-[#1f1b16]/12 bg-[#fffaf0] shadow-[0_30px_80px_rgba(31,27,22,0.30),0_8px_22px_rgba(31,27,22,0.12)]`}
      >
        <header className="grid grid-cols-[1fr_auto] items-center gap-4 bg-[#1f1b16] px-6 py-5 text-[#fffaf0]">
          <div>
            <div
              className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-extrabold uppercase tracking-[0.30em]"
              style={{ color: kickerColor }}
            >
              {kicker}
            </div>
            <h3 className="mt-1 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-[28px] font-black leading-none tracking-[-0.04em]">
              {title}
            </h3>
          </div>
          <button
            aria-label="Cerrar modal de mejora con IA"
            className="grid h-9 w-9 place-items-center rounded-[10px] border border-[#fffaf0]/20 bg-transparent text-[#fffaf0] transition hover:rotate-90 hover:bg-[#fffaf0]/10"
            type="button"
            onClick={onClose}
          >
            <svg
              className="h-4 w-4"
              fill="none"
              stroke="currentColor"
              strokeLinecap="round"
              strokeWidth={2.2}
              viewBox="0 0 24 24"
            >
              <path d="M6 6l12 12M18 6L6 18" />
            </svg>
          </button>
        </header>
        <div className="max-h-[calc(100vh-10rem)] overflow-y-auto px-6 py-5">
          {children}
        </div>
      </section>
    </div>
  );
}
