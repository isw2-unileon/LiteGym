import { type FormEvent, type ReactNode, useCallback, useEffect, useRef, useState } from "react";
import { apiUrl } from "../lib/api";
import { exerciseTypeLabel, muscleGroupLabel } from "../lib/exerciseLabels";
import AIRoutinePreviewModal, {
  type AIRoutinePreview,
  type AIRoutinePreviewSet,
} from "../components/Routine/AIRoutinePreviewModal";
import type { Exercise } from "../types/exercise";
import {HelloHeader} from "@/components/HelloHeader.tsx";
import {useOutletContext} from "react-router-dom";
import type {LayoutUser} from "@/components/AppLayout.tsx";
import {Stat} from "@/components/Stat.tsx";
import {Card, CardHeader} from "@/components/Card.tsx";
import {DialogPopup} from "@/components/DialogPopup.tsx";

type RoutineFilter = {
  key: string;
  label: string;
  count: number;
};

type RoutineSummary = {
  is_predefined: boolean;
  source: string;
  id: string;
  name: string;
  description?: string | null;
  exercise_count: number;
  updated_at?: string;
  routine_type?: string;
};

type RoutineAction =
    | "details"
    | "edit"
    | "improve"
    | "plan"
    | "duplicate"
    | "delete";


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
  description: string | null;
  muscle_group: string;
  secondary_muscle_group?: string;
  exercise_type?: string;
  exercise_order: number;
  notes: string | null;
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

type OutletContext = {
  user?: LayoutUser | null;
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

// ENUM from database to filter
type RoutineFilterKey = "Fuerza" | "Movilidad" | "Resistencia" | "Sin clasificar";

const ROUTINE_TYPE_FILTERS: { key: RoutineFilterKey; label: string }[] = [
  { key: "Fuerza", label: "Fuerza" },
  { key: "Movilidad", label: "Movilidad" },
  { key: "Resistencia", label: "Resistencia" },
  { key: "Sin clasificar", label: "Sin clasificar" },
];

function routineFilterKey(routine: RoutineSummary): RoutineFilterKey {
  switch (routine.routine_type?.trim()) {
    case "Fuerza":
      return "Fuerza";
    case "Movilidad":
      return "Movilidad";
    case "Resistencia":
      return "Resistencia";
    default:
      return "Sin clasificar";
  }
}

const MAX_ROUTINES_PER_USER = 50;

export default function UserRoutinesPage() {
  const [routines, setRoutines] = useState<RoutineSummary[]>([]);
  const [status, setStatus] = useState<RoutineStatus>("idle");
  const [selectedRoutineID, setSelectedRoutineID] = useState("");
  const [selectedRoutine, setSelectedRoutine] = useState<RoutineDetail | null>(
    null,
  );
  const [detailStatus, setDetailStatus] = useState<RoutineStatus>("idle");
  const [routineActionStatus, setRoutineActionStatus] =
    useState<RoutineStatus>("idle");
  const [routineActionMessage, setRoutineActionMessage] = useState("");
  const [activeFilter, setActiveFilter] = useState("all");
  const [routineSearch, setRoutineSearch] = useState("");
  const [debouncedRoutineSearch, setDebouncedRoutineSearch] = useState("");
  const [isAIFormOpen, setIsAIFormOpen] = useState(false);
  const [isManualFormOpen, setIsManualFormOpen] = useState(false);
  const [editingRoutineId, setEditingRoutineId] = useState("");
  const [manualName, setManualName] = useState("");
  const [manualObjective, setManualObjective] = useState("");
  const [manualNotes, setManualNotes] = useState("");
  const [manualType, setManualType] = useState<RoutineFilterKey>("Sin clasificar");
  const [manualExercises, setManualExercises] = useState<ManualRoutineExerciseState[]>([]);
  const [manualExerciseSearch, setManualExerciseSearch] = useState("");
  const [manualStatus, setManualStatus] = useState<RoutineStatus>("idle");
  const [manualMessage, setManualMessage] = useState("");
  const [isExercisesCollapsed, setIsExercisesCollapsed] = useState(false);
  const [isSetsCollapsed, setIsSetsCollapsed] = useState(false);
  const [collapsedExerciseIds, setCollapsedExerciseIds] = useState<string[]>([]);

  type ManualRoutineExerciseState = {
    exercise_id: string;
    name: string;
    muscle_group: string;
    exercise_type?: string | null;
    sets: RoutineSetDetail[];
  };
  const [isPlanOpen, setIsPlanOpen] = useState(false);
  const [planDate, setPlanDate] = useState("");
  const [planTime, setPlanTime] = useState("");
  const [planStatus, setPlanStatus] = useState<RoutineStatus>("idle");
  const [planMessage, setPlanMessage] = useState("");
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
    setRoutineActionStatus("idle");
    setRoutineActionMessage("");

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

  const handleRoutineAction = (action: RoutineAction) => {
    if (action === "improve") {
      handleOpenAIUpgrade();
      return;
    }
    if (action === "duplicate") {
      void handleDuplicateRoutine();
      return;
    }
    if (action === "delete") {
      void handleDeleteRoutine();
      return;
    }
    if (action === "edit") {
      handleOpenEditRoutine();
      return;
    }
    if (action === "plan") {
      handleOpenPlanWorkout();
      return;
    }
  };

  const fetchRoutines = useCallback(async () => {
    setStatus("loading");

    try {
      const response = await fetch(
        apiUrl("/api/routines"),
        {
          credentials: "include",
        },
      );

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
    if (routineSearch === debouncedRoutineSearch) {
      return;
    }

    const timeoutId = window.setTimeout(() => {
      setDebouncedRoutineSearch(routineSearch);
    }, 300);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [routineSearch, debouncedRoutineSearch]);

  useEffect(() => {
    if (!isAIFormOpen && !isManualFormOpen) {
      return;
    }

    if (availableExercisesStatus === "loading" || availableExercisesStatus === "success") {
      return;
    }

    void fetchAvailableExercises();
  }, [availableExercisesStatus, fetchAvailableExercises, isAIFormOpen, isManualFormOpen]);

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

  const toggleExerciseCollapse = (exerciseId: string) => {
    setCollapsedExerciseIds((current) =>
      current.includes(exerciseId)
        ? current.filter((id) => id !== exerciseId)
        : [...current, exerciseId],
    );
  };

  const toggleManualExercise = (exercise: Exercise) => {
    setManualExercises((current) => {
      const exists = current.some((ex) => ex.exercise_id === exercise.id);
      if (exists) {
        setCollapsedExerciseIds((ids) => ids.filter((id) => id !== exercise.id));
        return current.filter((ex) => ex.exercise_id !== exercise.id);
      }

      const newState: ManualRoutineExerciseState = {
        exercise_id: exercise.id,
        name: exercise.name,
        muscle_group: exercise.muscle_group,
        exercise_type: exercise.exercise_type,
        sets: [
          {
            id: Math.random().toString(36).substring(2, 9),
            set_number: 1,
          },
        ],
      };
      return [...current, newState];
    });
  };

  const addManualSet = (exerciseId: string) => {
    setManualExercises((current) =>
      current.map((ex) => {
        if (ex.exercise_id !== exerciseId) return ex;
        const newSet: RoutineSetDetail = {
          id: Math.random().toString(36).substring(2, 9),
          set_number: ex.sets.length + 1,
        };
        return { ...ex, sets: [...ex.sets, newSet] };
      }),
    );
  };

  const removeManualSet = (exerciseId: string, setIndex: number) => {
    setManualExercises((current) =>
      current.map((ex) => {
        if (ex.exercise_id !== exerciseId) return ex;
        if (ex.sets.length <= 1) return ex;
        const newSets = ex.sets
          .filter((_, i) => i !== setIndex)
          .map((set, i) => ({ ...set, set_number: i + 1 }));
        return { ...ex, sets: newSets };
      }),
    );
  };

  const updateManualSet = (
    exerciseId: string,
    setIndex: number,
    updates: Partial<RoutineSetDetail>,
  ) => {
    setManualExercises((current) =>
      current.map((ex) => {
        if (ex.exercise_id !== exerciseId) return ex;
        const newSets = [...ex.sets];
        const currentSet = newSets[setIndex];
        if (!currentSet) return ex;

        const updatedSet: RoutineSetDetail = { ...currentSet, ...updates };
        newSets[setIndex] = updatedSet;
        return { ...ex, sets: newSets };
      }),
    );
  };

  const manualMuscleGroupKeys = Array.from(
    new Set(
      manualExercises
        .map((exercise) => exercise.muscle_group)
        .filter((group) => group.trim() !== ""),
    ),
  );

  const resetManualForm = () => {
    setEditingRoutineId("");
    setManualName("");
    setManualObjective("");
    setManualNotes("");
    setManualType("Sin clasificar");
    setManualExercises([]);
    setManualExerciseSearch("");
    setManualStatus("idle");
    setManualMessage("");
  };

  const handleOpenCreateRoutine = () => {
    resetManualForm();
    setIsManualFormOpen(true);
  };

  const handleOpenEditRoutine = () => {
    if (selectedRoutine == null) {
      return;
    }

    resetManualForm();
    setEditingRoutineId(selectedRoutine.id);
    setManualName(selectedRoutine.name);
    setManualExercises(
      selectedRoutine.exercises.map((ex) => ({
        exercise_id: ex.exercise_id,
        name: ex.name,
        muscle_group: ex.muscle_group,
        exercise_type: ex.exercise_type,
        sets: ex.sets.map((s) => ({
          id: s.id,
          set_number: s.set_number,
          target_reps_min: s.target_reps_min ?? undefined,
          target_reps_max: s.target_reps_max ?? undefined,
          target_reps_text: s.target_reps_text,
          target_weight_kg: s.target_weight_kg ?? undefined,
          target_duration_seconds: s.target_duration_seconds ?? undefined,
          target_distance_km: s.target_distance_km ?? undefined,
          target_rir: s.target_rir ?? undefined,
          rest_seconds: s.rest_seconds ?? undefined,
          notes: s.notes,
        })),
      })),
    );
    setManualType(routineFilterKey(selectedRoutine));

    const lines = (selectedRoutine.description || "").split("\n");
    let objective = "";
    const cleanNotesLines: string[] = [];

    for (const line of lines) {
      if (line.startsWith("[Objetivo] ")) {
        objective = line.replace("[Objetivo] ", "").trim();
      } else if (line.startsWith("[Grupos musculares] ")) {
        // We skip this line since muscle groups are now stored in a structured way and shown separately in the form
      } else {
        cleanNotesLines.push(line);
      }
    }

    setManualObjective(objective);
    setManualNotes(cleanNotesLines.join("\n").trim());

    setIsManualFormOpen(true);
  };

  const handleCloseManualForm = () => {
    if (manualStatus === "loading") {
      return;
    }

    setIsManualFormOpen(false);
  };

  const handleCreateManualRoutine = async (
    event: FormEvent<HTMLFormElement>,
  ) => {
    event.preventDefault();

    const trimmedName = manualName.trim();
    const finalName =
      trimmedName !== ""
        ? trimmedName
        : manualExercises.map((ex) => ex.name).join(", ");

    if (finalName === "") {
      setManualStatus("error");
      setManualMessage("Ponle un nombre o selecciona algún ejercicio.");
      return;
    }

    for (const ex of manualExercises) {
      for (const s of ex.sets) {
        if (
          s.target_reps_min != null &&
          s.target_reps_max != null &&
          s.target_reps_min > s.target_reps_max
        ) {
          setManualStatus("error");
          setManualMessage(
            `En "${ex.name}", el mínimo de reps (${s.target_reps_min}) no puede ser mayor que el máximo (${s.target_reps_max}).`,
          );
          return;
        }

        if (s.target_rir != null && (s.target_rir < 0 || s.target_rir > 10)) {
          setManualStatus("error");
          setManualMessage(
            `En "${ex.name}", el RIR (${s.target_rir}) debe estar entre 0 y 10.`,
          );
          return;
        }
      }
    }

    setManualStatus("loading");
    setManualMessage("");

    const noteSections: string[] = [];
    if (manualObjective.trim() !== "") {
      noteSections.push(`[Objetivo] ${manualObjective.trim()}`);
    }
    if (manualMuscleGroupKeys.length > 0) {
      noteSections.push(
        `[Grupos musculares] ${manualMuscleGroupKeys
          .map((group) => muscleGroupLabel(group))
          .join(", ")}`,
      );
    }
    if (manualNotes.trim() !== "") {
      noteSections.push(manualNotes.trim());
    }
    const composedNotes = noteSections.join("\n");

    const isEditing = editingRoutineId !== "";

    try {
      const response = await fetch(
        isEditing
          ? apiUrl(`/api/routines/${editingRoutineId}`)
          : apiUrl("/api/routines"),
        {
          method: isEditing ? "PUT" : "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            name: finalName,
            objective: manualObjective.trim(),
            routine_type: manualType,
            target_muscle_groups: manualMuscleGroupKeys,
            notes: composedNotes,
            exercises: manualExercises.map((ex) => ({
              exercise_id: ex.exercise_id,
              sets: ex.sets.map((s) => ({
                set_number: s.set_number,
                target_reps_min: s.target_reps_min && s.target_reps_min > 0 ? s.target_reps_min : null,
                target_reps_max: s.target_reps_max && s.target_reps_max > 0 ? s.target_reps_max : null,
                target_reps_text: s.target_reps_text,
                target_weight_kg: s.target_weight_kg != null ? s.target_weight_kg : null,
                target_duration_seconds: s.target_duration_seconds && s.target_duration_seconds > 0 ? s.target_duration_seconds : null,
                target_distance_km: s.target_distance_km && s.target_distance_km > 0 ? s.target_distance_km : null,
                target_rir: s.target_rir != null ? s.target_rir : null,
                rest_seconds: s.rest_seconds != null ? s.rest_seconds : null,
                notes: s.notes,
              })),
            })),
          }),
        },
      );

      const payload = (await response.json().catch(() => null)) as
        | (AIRoutineSaveResponse & { error?: string; detail?: string })
        | null;

      if (!response.ok) {
        setManualStatus("error");
        setManualMessage(
          payload?.detail ||
            payload?.error ||
            (isEditing
              ? "No se pudo guardar la rutina."
              : "No se pudo crear la rutina."),
        );
        return;
      }

      const routineToShow = isEditing ? editingRoutineId : payload?.routine_id;

      setManualStatus("success");
      setIsManualFormOpen(false);
      resetManualForm();
      await fetchRoutines();

      if (routineToShow) {
        await fetchRoutineDetail(routineToShow);
      }
    } catch {
      setManualStatus("error");
      setManualMessage(
        isEditing
          ? "No se pudo conectar para guardar la rutina."
          : "No se pudo conectar para crear la rutina.",
      );
    }
  };

  const handleDuplicateRoutine = async () => {
    if (selectedRoutine == null) {
      return;
    }

    setRoutineActionStatus("loading");
    setRoutineActionMessage("");

    try {
      const response = await fetch(
        apiUrl(`/api/routines/${selectedRoutine.id}/duplicate`),
        {
          method: "POST",
          credentials: "include",
        },
      );

      const payload = (await response.json().catch(() => null)) as
        | (AIRoutineSaveResponse & { error?: string; detail?: string })
        | null;

      if (!response.ok) {
        setRoutineActionStatus("error");
        setRoutineActionMessage(
          payload?.detail || payload?.error || "No se pudo duplicar la rutina.",
        );
        return;
      }

      setRoutineActionStatus("idle");
      setRoutineActionMessage("");
      await fetchRoutines();

      if (payload?.routine_id) {
        await fetchRoutineDetail(payload.routine_id);
      }
    } catch {
      setRoutineActionStatus("error");
      setRoutineActionMessage("No se pudo conectar para duplicar la rutina.");
    }
  };

  const handleDeleteRoutine = async () => {
    if (selectedRoutine == null) {
      return;
    }

    if (!window.confirm(`¿Eliminar la rutina "${selectedRoutine.name}"?`)) {
      return;
    }

    setRoutineActionStatus("loading");
    setRoutineActionMessage("");

    try {
      const response = await fetch(
        apiUrl(`/api/routines/${selectedRoutine.id}`),
        {
          method: "DELETE",
          credentials: "include",
        },
      );

      if (!response.ok) {
        const payload = (await response.json().catch(() => null)) as
          | { error?: string; detail?: string }
          | null;
        setRoutineActionStatus("error");
        setRoutineActionMessage(
          payload?.detail || payload?.error || "No se pudo eliminar la rutina.",
        );
        return;
      }

      setRoutineActionStatus("idle");
      setRoutineActionMessage("");
      setSelectedRoutineID("");
      setSelectedRoutine(null);
      setDetailStatus("idle");
      await fetchRoutines();
    } catch {
      setRoutineActionStatus("error");
      setRoutineActionMessage("No se pudo conectar para eliminar la rutina.");
    }
  };

  const handleOpenPlanWorkout = () => {
    if (selectedRoutine == null) {
      return;
    }

    const today = new Date();
    const year = String(today.getFullYear());
    const month = String(today.getMonth() + 1).padStart(2, "0");
    const day = String(today.getDate()).padStart(2, "0");

    setPlanDate(`${year}-${month}-${day}`);
    setPlanTime("");
    setPlanStatus("idle");
    setPlanMessage("");
    setIsPlanOpen(true);
  };

  const handleClosePlanWorkout = () => {
    if (planStatus === "loading") {
      return;
    }

    setIsPlanOpen(false);
  };

  const handlePlanWorkout = async () => {
    if (selectedRoutine == null) {
      return;
    }

    const [yearStr, monthStr, dayStr] = planDate.split("-");
    const year = Number(yearStr);
    const month = Number(monthStr);
    const day = Number(dayStr);
    if (!Number.isInteger(year) || !Number.isInteger(month) || !Number.isInteger(day)) {
      setPlanStatus("error");
      setPlanMessage("Introduce una fecha válida.");
      return;
    }

    const [hourStr, minuteStr] = planTime.split(":");
    const hour = Number(hourStr);
    const minute = Number(minuteStr);
    if (
      !Number.isInteger(hour) ||
      !Number.isInteger(minute) ||
      hour < 0 ||
      hour > 23 ||
      minute < 0 ||
      minute > 59
    ) {
      setPlanStatus("error");
      setPlanMessage("Introduce una hora válida.");
      return;
    }

    const plannedAt = new Date(
      year,
      month - 1,
      day,
      hour,
      minute,
      0,
    ).toISOString();

    setPlanStatus("loading");
    setPlanMessage("");

    try {
      const response = await fetch(apiUrl("/api/workouts/planned"), {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          routine_id: selectedRoutine.id,
          name: selectedRoutine.name,
          planned_at: plannedAt,
        }),
      });

      if (!response.ok) {
        const payload = (await response.json().catch(() => null)) as
          | { error?: string; detail?: string }
          | null;
        setPlanStatus("error");
        setPlanMessage(
          payload?.detail ||
            payload?.error ||
            "No se pudo planificar el entreno.",
        );
        return;
      }

      setPlanStatus("success");
      setIsPlanOpen(false);
    } catch {
      setPlanStatus("error");
      setPlanMessage("No se pudo conectar para planificar el entreno.");
    }
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

  const { user } = useOutletContext<OutletContext>();

  const routineLimitReached = routines.length >= MAX_ROUTINES_PER_USER;

  const routineFilters: RoutineFilter[] = [
    { key: "all", label: "Todas", count: routines.length },
    ...ROUTINE_TYPE_FILTERS.map((filter) => ({
      ...filter,
      count: routines.filter((routine) => routineFilterKey(routine) === filter.key)
        .length,
    })),
  ];

  const visibleRoutines = routines.filter((routine) => {
    const search = debouncedRoutineSearch.trim().toLowerCase();
    const matchesSearch =
      search === "" ||
      routine.name.toLowerCase().includes(search) ||
      (routine.description || "").toLowerCase().includes(search);

    const matchesFilter =
      activeFilter === "all" || routineFilterKey(routine) === activeFilter;

    return matchesSearch && matchesFilter;
  });


  useEffect(() => {
    const stillVisible =
      selectedRoutineID !== "" &&
      visibleRoutines.some((routine) => routine.id === selectedRoutineID);

    if (stillVisible) {
      return;
    }

    const firstRoutine = visibleRoutines[0];
    if (firstRoutine) {
      void fetchRoutineDetail(firstRoutine.id);
    } else if (selectedRoutineID !== "") {
      setSelectedRoutineID("");
      setSelectedRoutine(null);
      setDetailStatus("idle");
    }
  }, [visibleRoutines, selectedRoutineID, fetchRoutineDetail]);

  function countAIRoutines() {
    return String(routines.filter((r) => r.source === "ai").length);
  }

  function countDefaultRoutines() {
    return String(routines.filter((r) => r.is_predefined).length);
  }

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
        <div className="px-6 pt-8 sm:px-8">
          <section className="mx-auto mb-6 grid max-w-[1280px] grid-cols-1 items-start gap-6 md:grid-cols-[1fr_auto]">
            <div>
              <HelloHeader page={"LISTADO DE RUTINAS"} user={user?.username ?? "Atleta"} />
              <p className="mt-3.5 max-w-[640px] text-[15px] leading-[1.55] text-[#3a332c]">
                Organiza, edita y planifica cualquier sesión con un solo click.
              </p>
              {status === "error" && (
                  <p className="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-3 py-2 text-[12px] font-bold text-[#9f2f22]">
                    No se ha podido cargar toda la información de las rutinas.
                  </p>
              )}
            </div>

            <div className="flex flex-wrap gap-2.5">
              <Stat n={String(routines?.length)} l="Rutinas totales" />
              <Stat n={`${routines.length - parseInt(countDefaultRoutines())} / ${MAX_ROUTINES_PER_USER}`} l="Límite de rutinas propias" />
              <Stat n={countDefaultRoutines()} l="Rutinas por defecto" />
              <Stat n={countAIRoutines()} l="Rutinas con IA" />
            </div>
          </section>
          <section className="mx-auto max-w-[1280px]">
            <Toolbar
                filters={routineFilters}
                activeFilter={activeFilter}
                onFilterChange={setActiveFilter}
                search={routineSearch}
                onSearchChange={setRoutineSearch}
                onCreate={handleOpenCreateRoutine}
                onCreateAI={() => {
                  setAIMessage("");
                  setAIStatus("idle");
                  setIsAIFormOpen(true);
                }}
                createDisabled={routineLimitReached}
            />
          </section>
          <section className="mx-auto grid max-w-[1280px] grid-cols-1 gap-[18px] xl:grid-cols-[3fr_7fr]">
            <div className="flex flex-col gap-[18px] mt-4">
              <Card accent="#ea7130" className={"flex flex-col h-[70vh]"}>
                <CardHeader
                    kicker={"Rutinas"}
                    title={"Elige una rutina"}
                />
                <ul className="relative z-[2] mt-6 flex min-h-0 flex-1 flex-col gap-2.5 overflow-y-auto pr-1">
                  {visibleRoutines.length === 0 ? (
                      <li className="rounded-[16px] border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/60 p-6 text-center text-[14px] font-semibold text-[#3a332c]/70">
                        {status === "loading"
                          ? "Cargando rutinas..."
                          : routineSearch.trim() !== ""
                            ? "No hay rutinas que coincidan con la búsqueda."
                            : routines.length === 0
                              ? "Todavía no tienes rutinas guardadas."
                              : "No hay rutinas de este tipo."}
                      </li>
                  ) : (
                      visibleRoutines.map((routine) => (
                      <li
                          key={routine.id}
                          onClick={() => void fetchRoutineDetail(routine.id)}
                          className={[
                            " cursor-pointer rounded-[16px] border bg-[#fffaf0]/75 p-4 transition hover:border-[#ea7130]/35 hover:bg-[#fff7ea] hover:shadow-[0_10px_22px_rgba(31,27,22,0.08)] has-[[aria-expanded=true]]:relative has-[[aria-expanded=true]]:z-20",
                            selectedRoutineID === routine.id
                                ? "border-[#ea7130] shadow-[0_10px_22px_rgba(31,27,22,0.08)]"
                                : "border-[#1f1b16]/10",
                          ].join(" ")}
                      >
                        <div className="flex items-center justify-between gap-2">
                          <span className="inline-flex items-center rounded-md bg-[#1f1b16] px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.16em] text-[#f1a45b]">
                            {routine.is_predefined ? "Rutina por defecto" : "Rutina"}
                          </span>
                        </div>
                        <h4 className="m-0 mt-2.5 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[18px] font-black leading-tight tracking-[-0.03em] text-[#1f1b16]">
                          {routine.name}
                        </h4>
                        <p className="mt-1.5 text-[14px] font-semibold leading-[1.5] text-[#3a332c]/80 line-clamp-2">
                          {cleanDescription(routine.description) || "Esta rutina no tiene descripción."}
                        </p>
                        <span className="mt-2 inline-flex items-center rounded-md border border-[#265c52]/18 bg-[#265c52]/10 px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-wider text-[#265c52]">
                              {routine.exercise_count} {routine.exercise_count === 1 ? "ejercicio": "ejercicios"}
                            </span>
                      </li>
                      ))
                  )}

                </ul>
              </Card >
            </div>
            <div className="flex flex-col gap-[18px] mt-4">
              <Card accent="#ea7130" className={"flex flex-col h-[70vh]"}>
                <CardHeader
                    kicker={"Detalle"}
                    title={selectedRoutine ? selectedRoutine.name : "Detalle de la rutina"}
                    right={
                      <RoutineActionsMenu
                          routineName={selectedRoutine ? selectedRoutine.name : "Detalle de la rutina"}
                          onSelect={(action) => handleRoutineAction(action)}
                      />}
                />
                {detailStatus === "loading" ? (
                    <p className="relative z-[2] mt-4 rounded-[16px] border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/60 p-6 text-center text-[14px] font-semibold text-[#3a332c]/70">
                      Cargando detalle...
                    </p>
                ) : detailStatus === "error" ? (
                    <p className="relative z-[2] mt-4 rounded-[16px] border border-[#9f2f22]/20 bg-[#9f2f22]/8 p-6 text-center text-[14px] font-semibold text-[#9f2f22]">
                      No se pudo cargar el detalle de la rutina.
                    </p>
                ) : selectedRoutine ? (
                    <div className="relative mt-4 flex min-h-0 flex-1 flex-col gap-4">
                      <div className="flex flex-wrap gap-2.5">
                        <Stat n={formatRoutineDate(selectedRoutine.updated_at)} l="Última modificación" />
                        <Stat n={selectedRoutine.source === "ai" ? selectedRoutine.source.toUpperCase() : selectedRoutine.source} l="Origen" />
                        <Stat n={String(selectedRoutine.exercises.length)} l="Ejercicios" />
                        <Stat n={String(selectedRoutine.exercises.reduce((total, exercise) => total + exercise.sets.length, 0))} l="Series" />
                      </div>
                      <div className="flex flex-wrap items-center gap-2">
                        <div className="flex flex-wrap gap-2">
                          {ROUTINE_DETAIL_ACTIONS.map((action) => (
                              <button
                                  key={action.key}
                                  type="button"
                                  onClick={() => handleRoutineAction(action.key)}
                                  className="inline-flex items-center gap-2 rounded-[10px] border border-[#1f1b16]/15 bg-[#fffaf0]/75 px-3.5 py-2 text-[14px] font-bold text-[#1f1b16] transition hover:-translate-y-px hover:border-[#ea7130]/40 hover:bg-[#f1a45b]/10"
                              >
                                {action.icon}
                                {action.label}
                              </button>
                          ))}
                        </div>
                        <span className="ml-auto inline-flex items-center rounded-md bg-[#1f1b16] px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.16em] text-[#f1a45b]">
                        {selectedRoutine.is_predefined ? "Rutina por defecto" : "Rutina"}
                      </span>
                        <span className="inline-flex items-center rounded-md border border-[#265c52]/18 bg-[#265c52]/10 px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-wider text-[#265c52]">
                        {selectedRoutine.exercise_count} {selectedRoutine.exercise_count === 1 ? "ejercicio" : "ejercicios"}
                      </span>
                      </div>
                      <p className="text-[15px] font-semibold leading-[0.75] text-[#3a332c]">
                        {cleanDescription(selectedRoutine.description) || "Esta rutina no tiene descripción."}
                      </p>
                      {routineActionMessage && (
                        <p
                          className={`text-[13px] font-bold ${
                            routineActionStatus === "error"
                              ? "text-[#9f2f22]"
                              : "text-[#265c52]"
                          }`}
                        >
                          {routineActionMessage}
                        </p>
                      )}
                      {selectedRoutine.exercises.length === 0 ? (
                          <p className="rounded-[16px] border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/60 p-6 text-center text-[14px] font-semibold text-[#3a332c]/70">
                            Esta rutina no tiene ejercicios guardados.
                          </p>
                      ) : (
                          <ul className="flex min-h-0 flex-1 flex-col gap-2.5 overflow-y-auto pr-1">
                            {selectedRoutine.exercises.map((exercise) => {
                              const firstSet = exercise.sets[0];
                              return (
                                  <li
                                      key={exercise.id}
                                      className="flex items-center justify-between gap-2 rounded-[16px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-4 text-[15px] font-bold text-[#1f1b16]"
                                  >
                                    <div className="flex flex-col gap-1">
                                      <div className="flex items-center gap-2">
                                        <span className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[20px] font-black text-[#265c52] opacity-80">#{exercise.exercise_order}</span>
                                        <p className="text-[16px]">{exercise.name}</p>
                                      </div>
                                      <p className="text-[16px]">
                                        {muscleGroupLabel(exercise.muscle_group)}
                                        {exercise.secondary_muscle_group ? ` · ${muscleGroupLabel(exercise.secondary_muscle_group)}` : ""}
                                      </p>
                                      <p className="text-[16px]">{exercise.exercise_type ? exerciseTypeLabel(exercise.exercise_type) : "Sin tipo"}</p>
                                    </div>
                                    <span className="inline-flex items-center rounded-md border border-[#265c52]/18 bg-[#265c52]/10 px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[14px] font-bold uppercase tracking-wider text-[#265c52]">
                                      {[
                                        `${exercise.sets.length} ${exercise.sets.length === 1 ? "serie" : "series"}`,
                                        firstSet ? formatReps(firstSet) : "",
                                        firstSet?.target_rir != null ? `RIR ${firstSet.target_rir}` : "",
                                      ]
                                        .filter(Boolean)
                                        .join(" · ")}
                                    </span>
                                  </li>
                              );
                            })}
                          </ul>
                      )}
                    </div>
                ) : (
                    <p className="relative z-[2] mt-4 rounded-[16px] border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/60 p-6 text-center text-[14px] font-semibold text-[#3a332c]/70">
                      Selecciona una rutina del listado para ver sus datos aquí.
                    </p>
                )}

              </Card>
            </div>

          </section>

        </div>
    <section className="space-y-6">
      {isManualFormOpen && (
        <RoutineDialogShell
          kicker={editingRoutineId !== "" ? "Editar rutina" : "Nueva rutina"}
          kickerColor="#265c52"
          title={editingRoutineId !== "" ? "Editar rutina" : "Crear rutina manual"}
          maxWidthClass="max-w-3xl"
          onClose={handleCloseManualForm}
        >
          <form className="grid gap-4" onSubmit={handleCreateManualRoutine}>
            <div className="grid gap-4 lg:grid-cols-2">
              <label className="block">
                <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                  Nombre
                </span>
                <input
                  className="mt-2 w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition focus:border-[#265c52]"
                  placeholder="Opcional: se genera con los ejercicios"
                  value={manualName}
                  onChange={(event) => setManualName(event.target.value)}
                />
              </label>

              <label className="block">
                <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                  Objetivo
                </span>
                <input
                  className="mt-2 w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition focus:border-[#265c52]"
                  placeholder="Ej. ganar fuerza"
                  value={manualObjective}
                  onChange={(event) => setManualObjective(event.target.value)}
                />
              </label>
            </div>

            <div className="grid gap-4 lg:grid-cols-2">
              <label className="block">
                <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                  Tipo de rutina
                </span>
                <select
                  className="mt-2 w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition focus:border-[#265c52]"
                  value={manualType}
                  onChange={(event) =>
                    setManualType(event.target.value as RoutineFilterKey)
                  }
                >
                  {ROUTINE_TYPE_FILTERS.map((filter) => (
                    <option key={filter.key} value={filter.key}>
                      {filter.label}
                    </option>
                  ))}
                </select>
              </label>

              <div className="block">
                <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                  Grupos musculares
                </span>
                <div className="mt-2 flex min-h-[3.25rem] w-full flex-wrap items-center gap-2 rounded-2xl border border-[#1f1b16]/10 bg-white/60 px-4 py-3">
                  {manualMuscleGroupKeys.length === 0 ? (
                    <span className="text-sm font-medium text-[#3a332c]/50">
                      Se completan al elegir ejercicios
                    </span>
                  ) : (
                    <span className="text-sm font-bold capitalize text-[#1f1b16]">
                      {manualMuscleGroupKeys
                        .map((group) => muscleGroupLabel(group))
                        .join(", ")}
                    </span>
                  )}
                </div>
              </div>
            </div>

            <label className="block">
              <span className="text-xs font-black uppercase tracking-[0.18em] text-[#265c52]">
                Notas
              </span>
              <textarea
                className="mt-2 min-h-28 w-full resize-y rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-semibold leading-6 outline-none transition placeholder:font-medium focus:border-[#265c52]"
                placeholder="Notas o indicaciones para esta rutina."
                value={manualNotes}
                onChange={(event) => setManualNotes(event.target.value)}
              />
            </label>

            <Card accent="#ea7130">
              <div
                onClick={() => setIsExercisesCollapsed(!isExercisesCollapsed)}
                className="cursor-pointer"
              >
                <CardHeader
                  kicker="Selección de ejercicios"
                  title="Ejercicios"
                  right={
                    <div className="flex items-center gap-3">
                      <span className="rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black text-[#265c52]">
                        {manualExercises.length} seleccionados
                      </span>
                      <svg
                        className={`h-5 w-5 transition-transform duration-300 ${
                          isExercisesCollapsed ? "" : "rotate-180"
                        }`}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2.5}
                          d="M19 9l-7 7-7-7"
                        />
                      </svg>
                    </div>
                  }
                />
              </div>

              <div
                className={`grid transition-[grid-template-rows] duration-300 ${
                  isExercisesCollapsed ? "grid-rows-[0fr]" : "grid-rows-[1fr]"
                }`}
              >
                <div className="min-h-0 overflow-hidden">
                  <label className="relative z-[2] mt-4 block">
                    <span className="sr-only">Buscar ejercicios</span>
                    <input
                      className="w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition placeholder:font-medium focus:border-[#265c52]"
                      placeholder="Buscar ejercicios por nombre o musculo"
                      value={manualExerciseSearch}
                      onChange={(event) =>
                        setManualExerciseSearch(event.target.value)
                      }
                    />
                  </label>

                  <div className="relative z-[2] mt-4 max-h-72 overflow-y-auto pr-1">
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

                    {availableExercisesStatus === "success" &&
                      (() => {
                        const search = manualExerciseSearch.trim().toLowerCase();
                        const filtered = availableExercises.filter(
                          (exercise) => {
                            if (search === "") {
                              return true;
                            }
                            return (
                              exercise.name.toLowerCase().includes(search) ||
                              exercise.muscle_group
                                .toLowerCase()
                                .includes(search) ||
                              (exercise.exercise_type ?? "")
                                .toLowerCase()
                                .includes(search)
                            );
                          },
                        );

                        if (filtered.length === 0) {
                          return (
                            <p className="rounded-2xl bg-white/70 p-4 text-sm font-semibold text-[#5d5348]">
                              No hay ejercicios que coincidan con el filtro.
                            </p>
                          );
                        }

                        return (
                          <div className="space-y-2">
                            {filtered.map((exercise) => {
                              const isSelected = manualExercises.some(
                                (ex) => ex.exercise_id === exercise.id,
                              );

                              return (
                                <button
                                  key={exercise.id}
                                  className={`flex w-full items-start gap-4 rounded-2xl border px-4 py-3 text-left transition ${
                                    isSelected
                                      ? "border-[#ea7130] bg-[#fff4ea]"
                                      : "border-[#1f1b16]/10 bg-white/90 hover:border-[#ea7130]/35 hover:bg-[#fff7ea]"
                                  }`}
                                  type="button"
                                  onClick={() => toggleManualExercise(exercise)}
                                >
                                  <div className="min-w-0">
                                    <p className="break-words text-sm font-black text-[#1f1b16]">
                                      {exercise.name}
                                    </p>
                                    <p className="mt-1 text-[11px] font-bold uppercase tracking-[0.14em] text-[#7a6b5c]">
                                      {muscleGroupLabel(exercise.muscle_group)}
                                      {exercise.exercise_type
                                        ? ` · ${exerciseTypeLabel(
                                            exercise.exercise_type,
                                          )}`
                                        : ""}
                                    </p>
                                  </div>
                                </button>
                              );
                            })}
                          </div>
                        );
                      })()}
                  </div>
                </div>
              </div>
            </Card>

            <Card accent="#ea7130">
              <div
                onClick={() => setIsSetsCollapsed(!isSetsCollapsed)}
                className="cursor-pointer"
              >
                <CardHeader
                  kicker="Configuración"
                  title="Series y Repeticiones"
                  right={
                    <svg
                      className={`h-5 w-5 transition-transform duration-300 ${
                        isSetsCollapsed ? "" : "rotate-180"
                      }`}
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2.5}
                        d="M19 9l-7 7-7-7"
                      />
                    </svg>
                  }
                />
              </div>

              <div
                className={`grid transition-[grid-template-rows] duration-300 ${
                  isSetsCollapsed ? "grid-rows-[0fr]" : "grid-rows-[1fr]"
                }`}
              >
                <div className="min-h-0 overflow-hidden">
                  <div className="relative z-[2] mt-4 space-y-6">
                    {manualExercises.length === 0 ? (
                      <p className="rounded-2xl border border-dashed border-[#1f1b16]/15 bg-white/50 p-6 text-center text-[14px] font-semibold text-[#3a332c]/70">
                        Selecciona ejercicios en la tarjeta anterior para configurar sus series.
                      </p>
                    ) : (
                      manualExercises.map((ex) => {
                        const isCollapsed = collapsedExerciseIds.includes(
                          ex.exercise_id,
                        );
                        return (
                          <div
                            key={ex.exercise_id}
                            className="rounded-2xl border border-[#1f1b16]/10 bg-white/50 p-4"
                          >
                            <div
                              onClick={() =>
                                toggleExerciseCollapse(ex.exercise_id)
                              }
                              className="flex cursor-pointer items-center justify-between gap-2"
                            >
                              <div className="flex items-center gap-2.5">
                                <h5 className="text-sm font-black text-[#1f1b16]">
                                  {ex.name}
                                </h5>
                                {isCollapsed && (
                                  <span className="rounded-full bg-[#265c52]/10 px-2 py-0.5 text-[10px] font-black text-[#265c52]">
                                    {ex.sets.length}{" "}
                                    {ex.sets.length === 1 ? "serie" : "series"}
                                  </span>
                                )}
                              </div>
                              <svg
                                className={`h-4 w-4 transition-transform duration-300 ${
                                  isCollapsed ? "" : "rotate-180"
                                }`}
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2.5}
                                  d="M19 9l-7 7-7-7"
                                />
                              </svg>
                            </div>

                            <div
                              className={`grid transition-[grid-template-rows] duration-300 ${
                                isCollapsed ? "grid-rows-[0fr]" : "grid-rows-[1fr]"
                              }`}
                            >
                              <div className="min-h-0 overflow-hidden">
                                <div className="mt-4 space-y-3">
                                  {ex.sets.map((set, idx) => {
                                    const isRangeInvalid =
                                      set.target_reps_min != null &&
                                      set.target_reps_max != null &&
                                      set.target_reps_min > set.target_reps_max;

                                    const isRirInvalid =
                                      set.target_rir != null &&
                                      (set.target_rir < 0 || set.target_rir > 10);

                                    return (
                                      <div key={set.id} className="space-y-1">
                                        {/* Etiquetas superiores */}
                                        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 pr-[92px]">
                                          <span className="text-[10px] font-black uppercase tracking-wider text-[#7a6b5c]">
                                            Min Reps
                                          </span>
                                          <span className="text-[10px] font-black uppercase tracking-wider text-[#7a6b5c]">
                                            Max Reps
                                          </span>
                                          <span className="text-[10px] font-black uppercase tracking-wider text-[#7a6b5c]">
                                            RIR
                                          </span>
                                        </div>

                                        {/* Inputs y Botón centrado */}
                                        <div className="flex items-center gap-3">
                                          <div className="flex-1 grid grid-cols-2 gap-2 sm:grid-cols-3">
                                            <input
                                              type="number"
                                              className={`w-full rounded-lg border bg-white px-2 py-1.5 text-xs font-bold transition ${
                                                isRangeInvalid
                                                  ? "border-[#9b2d20] ring-1 ring-[#9b2d20]/20"
                                                  : "border-[#1f1b16]/10"
                                              }`}
                                              value={set.target_reps_min ?? ""}
                                              onChange={(e) =>
                                                updateManualSet(ex.exercise_id, idx, {
                                                  target_reps_min: e.target.value
                                                    ? parseInt(e.target.value)
                                                    : undefined,
                                                })
                                              }
                                            />
                                            <input
                                              type="number"
                                              className={`w-full rounded-lg border bg-white px-2 py-1.5 text-xs font-bold transition ${
                                                isRangeInvalid
                                                  ? "border-[#9b2d20] ring-1 ring-[#9b2d20]/20"
                                                  : "border-[#1f1b16]/10"
                                              }`}
                                              value={set.target_reps_max ?? ""}
                                              onChange={(e) =>
                                                updateManualSet(ex.exercise_id, idx, {
                                                  target_reps_max: e.target.value
                                                    ? parseInt(e.target.value)
                                                    : undefined,
                                                })
                                              }
                                            />
                                            <input
                                              type="number"
                                              min={0}
                                              max={10}
                                              className={`w-full rounded-lg border bg-white px-2 py-1.5 text-xs font-bold transition ${
                                                isRirInvalid
                                                  ? "border-[#9b2d20] ring-1 ring-[#9b2d20]/20"
                                                  : "border-[#1f1b16]/10"
                                              }`}
                                              value={set.target_rir ?? ""}
                                              onChange={(e) => {
                                                const val = e.target.value
                                                  ? parseInt(e.target.value)
                                                  : undefined;
                                                updateManualSet(ex.exercise_id, idx, {
                                                  target_rir: val,
                                                });
                                              }}
                                            />
                                          </div>
                                          <button
                                            type="button"
                                            onClick={() =>
                                              removeManualSet(ex.exercise_id, idx)
                                            }
                                            disabled={ex.sets.length <= 1}
                                            className="px-3 py-1.5 text-[11px] font-black uppercase tracking-wider rounded-lg bg-[#9b2d20]/10 text-[#9b2d20] transition hover:bg-[#9b2d20] hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
                                          >
                                            Eliminar
                                          </button>
                                        </div>

                                        {(isRangeInvalid || isRirInvalid) && (
                                          <div className="mt-2 space-y-1.5">
                                            {isRangeInvalid && (
                                              <div className="flex items-start gap-2 rounded-lg border border-[#9b2d20]/20 bg-[#9b2d20]/5 p-2">
                                                <svg className="mt-0.5 w-3.5 h-3.5 text-[#9b2d20] shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                                </svg>
                                                <p className="text-[10px] font-black text-[#9b2d20] leading-[1.3] uppercase tracking-wide">
                                                  Error de repeticiones: El mínimo ({set.target_reps_min}) no puede superar al máximo ({set.target_reps_max})
                                                </p>
                                              </div>
                                            )}
                                            {isRirInvalid && (
                                              <div className="flex items-start gap-2 rounded-lg border border-[#9b2d20]/20 bg-[#9b2d20]/5 p-2">
                                                <svg className="mt-0.5 w-3.5 h-3.5 text-[#9b2d20] shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                                </svg>
                                                <p className="text-[10px] font-black text-[#9b2d20] leading-[1.3] uppercase tracking-wide">
                                                  Error de RIR: El valor ({set.target_rir}) debe estar entre 0 y 10
                                                </p>
                                              </div>
                                            )}
                                          </div>
                                        )}
                                      </div>
                                    );
                                  })}
                                  <button
                                    type="button"
                                    onClick={() => addManualSet(ex.exercise_id)}
                                    className="mt-2 inline-flex items-center gap-1.5 text-[11px] font-black uppercase tracking-wider text-[#265c52] hover:text-[#1f1b16]"
                                  >
                                    <svg
                                      viewBox="0 0 24 24"
                                      fill="none"
                                      stroke="currentColor"
                                      strokeWidth={3}
                                      className="h-3 w-3"
                                    >
                                      <path d="M12 5v14M5 12h14" />
                                    </svg>
                                    Añadir serie
                                  </button>
                                </div>
                              </div>
                            </div>
                          </div>
                        );
                      })
                    )}
                  </div>
                </div>
              </div>
            </Card>

            {manualMessage && (
              <p
                className={`text-sm font-black ${
                  manualStatus === "error" ? "text-[#9b2d20]" : "text-[#265c52]"
                }`}
              >
                {manualMessage}
              </p>
            )}

            <div className="flex flex-col gap-3 sm:flex-row sm:justify-end">
              <button
                className="rounded-2xl border border-[#1f1b16]/10 px-4 py-3 text-sm font-black text-[#1f1b16] transition hover:border-[#1f1b16] hover:bg-[#1f1b16] hover:text-white"
                type="button"
                onClick={handleCloseManualForm}
              >
                Cancelar
              </button>
              <button
                className="rounded-2xl bg-[#ea7130] px-5 py-3 text-sm font-black text-white transition hover:bg-[#1f1b16] disabled:cursor-not-allowed disabled:opacity-60"
                disabled={manualStatus === "loading"}
                type="submit"
              >
                {editingRoutineId !== ""
                  ? manualStatus === "loading"
                    ? "Guardando..."
                    : "Guardar cambios"
                  : manualStatus === "loading"
                    ? "Creando..."
                    : "Crear rutina"}
              </button>
            </div>
          </form>
        </RoutineDialogShell>
      )}

      {isPlanOpen && selectedRoutine && (
        <DialogPopup
          kicker="Planificar entreno"
          kickerColor="#ea7130"
          title={selectedRoutine.name}
          onClose={handleClosePlanWorkout}
        >
          <label
            htmlFor="routine-plan-date"
            className="block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[14px] font-bold uppercase tracking-[0.16em] text-[#3a332c]"
          >
            Día
          </label>
          <input
            id="routine-plan-date"
            type="date"
            value={planDate}
            onChange={(event) => setPlanDate(event.target.value)}
            className="mt-2 w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/85 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15"
          />

          <label
            htmlFor="routine-plan-time"
            className="mt-4 block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[14px] font-bold uppercase tracking-[0.16em] text-[#3a332c]"
          >
            Hora estimada
          </label>
          <input
            id="routine-plan-time"
            type="time"
            value={planTime}
            onChange={(event) => setPlanTime(event.target.value)}
            className="mt-2 w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/85 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15"
          />

          {planMessage && (
            <p className="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-3 py-2 text-[12px] font-bold text-[#9f2f22]">
              {planMessage}
            </p>
          )}

          <div className="mt-5 flex justify-end gap-2.5">
            <button
              type="button"
              onClick={handleClosePlanWorkout}
              className="cursor-pointer rounded-[14px] border border-[#1f1b16]/18 bg-transparent px-4 py-3 text-[13px] font-extrabold tracking-[0.04em] text-[#9f2f22] transition hover:bg-[#1f1b16]/5"
            >
              Cancelar
            </button>
            <button
              type="button"
              onClick={() => void handlePlanWorkout()}
              disabled={planStatus === "loading" || !planDate || !planTime}
              className="cursor-pointer rounded-[14px] bg-[#ea7130] px-5 py-3 text-[13px] font-extrabold tracking-[0.04em] text-[#1f1b16] shadow-[0_18px_35px_rgba(234,113,48,0.30)] transition hover:-translate-y-px hover:bg-[#ff8b47] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:translate-y-0"
            >
              {planStatus === "loading" ? "Guardando..." : "Guardar entreno"}
            </button>
          </div>
        </DialogPopup>
      )}

      {isAIFormOpen && (
        <RoutineDialogShell
          kicker="AI Routine"
          kickerColor="#265c52"
          title="Crear rutina con IA"
          maxWidthClass="max-w-3xl"
          onClose={() => setIsAIFormOpen(false)}
        >
        <form
          className="grid gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]"
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
              placeholder="Pecho, Espalda, Piernas ..."
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

          <Card accent="#ea7130" className="lg:col-span-2">
            <CardHeader
              kicker="Selección de ejercicios"
              title="Ejercicios obligatorios"
              right={
                <span className="rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black text-[#265c52]">
                  {selectedMandatoryExercises.length} seleccionados
                </span>
              }
            />

            <p className="relative z-[2] mt-2 text-sm font-semibold leading-6 text-[#5d5348]">
              Selecciona ejercicios por nombre. La IA recibira exactamente esos
              nombres como referencia.
            </p>

            <label className="relative z-[2] mt-4 block">
              <span className="sr-only">Buscar ejercicios</span>
              <input
                className="w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm font-bold outline-none transition placeholder:font-medium focus:border-[#265c52]"
                placeholder="Buscar ejercicios por nombre o musculo"
                value={exerciseSearch}
                onChange={(event) => setExerciseSearch(event.target.value)}
              />
            </label>

            {selectedMandatoryExercises.length > 0 && (
              <div className="relative z-[2] mt-4 flex flex-wrap gap-2">
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

            <div className="relative z-[2] mt-4 max-h-72 overflow-y-auto pr-1">
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
                              className={`flex w-full items-start gap-4 rounded-2xl border px-4 py-3 text-left transition ${
                                isSelected
                                  ? "border-[#ea7130] bg-[#fff4ea]"
                                  : "border-[#1f1b16]/10 bg-white/90 hover:border-[#ea7130]/35 hover:bg-[#fff7ea]"
                              }`}
                              type="button"
                              onClick={() => toggleMandatoryExercise(exercise.name)}
                            >
                              <div className="min-w-0">
                                <p className="break-words text-sm font-black text-[#1f1b16]">
                                  {exercise.name}
                                </p>
                                <p className="mt-1 text-[11px] font-bold uppercase tracking-[0.14em] text-[#7a6b5c]">
                                  {muscleGroupLabel(exercise.muscle_group)}
                                  {exercise.exercise_type
                                    ? ` · ${exerciseTypeLabel(exercise.exercise_type)}`
                                    : ""}
                                </p>
                              </div>
                            </button>
                          );
                        })}
                    </div>
                  )}
                </>
              )}
            </div>
          </Card>

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
        </RoutineDialogShell>
      )}

      <AIRoutinePreviewModal
        isOpen={isAIPreviewOpen}
        isSaving={isAIPreviewSaving}
        errorMessage={aiPreviewError}
        routine={aiPreviewRoutine}
        onClose={handleCloseAIRoutinePreview}
        onConfirm={() => void handleConfirmAIRoutine()}
      />

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
                  cleanDescription(selectedRoutine?.description) ||
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
      </main>
  );
}

function formatReps(set: RoutineSetDetail) {
  if (set.target_distance_km != null) {
    return `${set.target_distance_km} km`;
  }
  if (set.target_duration_seconds != null) {
    return `${set.target_duration_seconds} segundos`;
  }
  if (set.target_reps_text?.trim()) {
    return `${set.target_reps_text} repeticiones`;
  }
  if (set.target_reps_min != null && set.target_reps_max != null) {
    return `${set.target_reps_min}-${set.target_reps_max} repeticiones`;
  }
  if (set.target_reps_min != null) {
    return `${set.target_reps_min}+ repeticiones`;
  }
  return "";
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
        className={`w-full ${maxWidthClass} max-h-[calc(100vh-10rem)] overflow-hidden rounded-[24px] border-2 border-[#fffaf0]/20 shadow-[0_30px_80px_rgba(31,27,22,0.30),0_8px_22px_rgba(31,27,22,0.12)]`}
      >
        <header className="grid grid-cols-[1fr_auto] items-center gap-4 rounded-t-[24px] bg-[#1f1b16] px-6 py-5 text-[#fffaf0]">
          <div>
            <div
              className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[16px] font-extrabold uppercase tracking-[0.30em]"
              style={{ color: "#f1a45b" }}
            >
              {kicker}
            </div>
            <h3 className="mt-1 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[22px] font-black leading-none tracking-[-0.04em]">
              {title}
            </h3>
          </div>
          <button
            aria-label="Cerrar modal de mejora con IA"
            className="grid h-9 w-9 cursor-pointer place-items-center rounded-[10px] border border-[#fffaf0]/20 bg-transparent text-[#fffaf0] transition hover:rotate-90 hover:bg-[#fffaf0]/10"
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
        <div className="max-h-[calc(100vh-18rem)] overflow-y-auto bg-[#fffaf0] px-6 py-5">
          {children}
        </div>
      </section>
    </div>
  );
}

function Toolbar({
                   filters,
                   activeFilter,
                   onFilterChange,
                   search,
                   onSearchChange,
                   onCreate,
                   onCreateAI,
                   createDisabled = false,
                 }: {
  filters: RoutineFilter[];
  activeFilter: string;
  onFilterChange: (key: string) => void;
  search: string;
  onSearchChange: (value: string) => void;
  onCreate: () => void;
  onCreateAI: () => void;
  createDisabled?: boolean;
}) {
  return (
      <div className="flex flex-wrap items-center gap-2.5 rounded-[20px] border border-[#1f1b16]/12 bg-[#fffaf0]/75 p-2.5 backdrop-blur">
        <div className="flex flex-wrap items-center gap-1.5">
          {filters.map((filter) => {
            const isActive = filter.key === activeFilter;
            return (
                <button
                    key={filter.key}
                    type="button"
                    onClick={() => onFilterChange(filter.key)}
                    className={[
                      "inline-flex items-center gap-2 rounded-full px-4 py-2 text-[14px] font-extrabold tracking-[-0.01em] transition",
                      isActive
                          ? "bg-[#1f1b16] text-[#fffaf0]"
                          : "text-[#3a332c] hover:bg-[#1f1b16]/5",
                    ].join(" ")}
                >
                  {filter.label}
                  <span
                      className={[
                        "inline-flex min-w-[22px] items-center justify-center rounded-full px-1.5 py-0.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-black",
                        isActive
                            ? "bg-[#fffaf0]/20 text-[#fffaf0]"
                            : "bg-[#1f1b16]/8 text-[#3a332c]",
                      ].join(" ")}
                  >
                {filter.count}
              </span>
                </button>
            );
          })}
        </div>

        <label className="relative flex min-w-[220px] flex-1 items-center">
          <span className="sr-only">Buscar rutina</span>
          <svg
              aria-hidden="true"
              viewBox="0 0 24 24"
              className="pointer-events-none absolute left-3.5 h-4 w-4 text-[#3a332c]/50"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="7" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
              type="search"
              value={search}
              onChange={(event) => onSearchChange(event.target.value)}
              placeholder="Buscar por nombre, grupo..."
              className="w-full rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] py-2.5 pl-10 pr-4 text-[14px] font-bold text-[#1f1b16] outline-none transition placeholder:font-semibold placeholder:text-[#3a332c]/50 focus:border-[#265c52]"
          />
        </label>

        <button
            type="button"
            onClick={onCreate}
            disabled={createDisabled}
            title={createDisabled ? "Has alcanzado el límite de rutinas propias" : undefined}
            className="inline-flex cursor-pointer items-center gap-1.5 rounded-[10px] border-0 bg-[#1f1b16] px-3.5 py-2.5 text-[14px] font-extrabold leading-none tracking-[0.03em] text-[#f1a45b] no-underline shadow-[0_10px_22px_rgba(31,27,22,0.18)] transition hover:-translate-y-px hover:bg-[#2c261f] disabled:cursor-not-allowed disabled:opacity-50"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={6} strokeLinecap="round" strokeLinejoin="round" className="h-3 w-3">
            <path d="M12 5v14M5 12h14" />
          </svg>
          Crear rutina
        </button>

        <button
            type="button"
            onClick={onCreateAI}
            disabled={createDisabled}
            title={createDisabled ? "Has alcanzado el límite de rutinas" : undefined}
            className="inline-flex cursor-pointer items-center gap-1.5 rounded-[10px] border-0 bg-[#1f1b16] px-3.5 py-2.5 text-[14px] font-extrabold leading-none tracking-[0.03em] text-[#f1a45b] no-underline shadow-[0_10px_22px_rgba(31,27,22,0.18)] transition hover:-translate-y-px hover:bg-[#2c261f] disabled:cursor-not-allowed disabled:opacity-50"
        >
          <svg viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" strokeWidth={0.75} strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4">
            <path d="M12 2l1.9 5.6a3 3 0 0 0 1.9 1.9L21.5 11.5l-5.6 1.9a3 3 0 0 0-1.9 1.9L12 21l-1.9-5.6a3 3 0 0 0-1.9-1.9L2.5 11.5l5.6-1.9a3 3 0 0 0 1.9-1.9L12 2z" />
          </svg>
          Generar rutina con IA
        </button>
      </div>
  );
}

function RoutineActionsMenu({
                              routineName,
                              onSelect,
                            }: {
  routineName: string;
  onSelect: (action: RoutineAction) => void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen]);

  const handleSelect = (action: RoutineAction) => {
    setIsOpen(false);
    onSelect(action);
  };

  return (
      <div ref={menuRef} className={isOpen ? "relative z-50" : ""}>
        <button
            type="button"
            aria-haspopup="menu"
            aria-expanded={isOpen}
            aria-label={`Acciones de ${routineName}`}
            onClick={() => setIsOpen((open) => !open)}
            className="grid h-[30px] w-[30px] place-items-center rounded-[10px] border border-[#1f1b16]/15 bg-[#fffaf0]/75 text-[#1f1b16] transition hover:-translate-y-px hover:bg-[#f1a45b]/10"
        >
          <svg viewBox="0 0 24 24" fill="currentColor" className="h-5 w-5">
            <circle cx="5" cy="12" r="1.6" />
            <circle cx="12" cy="12" r="1.6" />
            <circle cx="19" cy="12" r="1.6" />
          </svg>
        </button>

        {isOpen && (
            <div className="absolute right-0 top-[calc(100%+0.5rem)] z-40 w-48 overflow-hidden rounded-[14px] border border-[#1f1b16]/10 bg-[#fffaf0]/95 backdrop-blur-md shadow-[0_10px_30px_rgba(31,27,22,0.10)]">
              {ROUTINE_MENU_ACTIONS.map((action) => (
                  <button
                      key={action.key}
                      type="button"
                      onClick={() => handleSelect(action.key)}
                      className={[
                        "flex w-full items-center gap-2.5 px-4 py-3 text-left text-sm font-bold transition",
                        action.destructive
                            ? "border-t border-[#1f1b16]/10 text-[#9f2f22] hover:bg-[#9f2f22]/10"
                            : "text-[#1f1b16]/80 hover:bg-[#f1a45b]/10 hover:text-[#1f1b16]",
                      ].join(" ")}
                  >
                    {action.icon}
                    {action.label}
                  </button>
              ))}
            </div>
        )}
      </div>
  );
}

const ROUTINE_DETAIL_ACTIONS: {
  key: RoutineAction;
  label: string;
  icon: ReactNode;
}[] = [
  {
    key: "edit",
    label: "Editar",
    icon: (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" className="h-[18px] w-[18px] shrink-0">
          <path d="M12 20h9" />
          <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4Z" />
        </svg>
    ),
  },
  {
    key: "plan",
    label: "Planificar",
    icon: (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" className="h-[18px] w-[18px] shrink-0">
          <rect x="3" y="4" width="18" height="18" rx="2" />
          <path d="M16 2v4" />
          <path d="M8 2v4" />
          <path d="M3 10h18" />
        </svg>
    ),
  },
  {
    key: "improve",
    label: "Mejorar con IA",
    icon: (
        <svg viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" strokeWidth={0.75} strokeLinecap="round" strokeLinejoin="round" className="h-[18px] w-[18px] shrink-0">
          <path d="M12 2l1.9 5.6a3 3 0 0 0 1.9 1.9L21.5 11.5l-5.6 1.9a3 3 0 0 0-1.9 1.9L12 21l-1.9-5.6a3 3 0 0 0-1.9-1.9L2.5 11.5l5.6-1.9a3 3 0 0 0 1.9-1.9L12 2z" />
        </svg>
    ),
  },
];

const ROUTINE_MENU_ACTIONS: {
  key: RoutineAction;
  label: string;
  icon: ReactNode;
  destructive?: boolean;
}[] = [
  {
    key: "duplicate",
    label: "Duplicar",
    icon: (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" className="h-[18px] w-[18px] shrink-0">
          <rect x="9" y="9" width="13" height="13" rx="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
    ),
  },
  {
    key: "delete",
    label: "Eliminar",
    destructive: true,
    icon: (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" className="h-[18px] w-[18px] shrink-0">
          <path d="M3 6h18" />
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" />
          <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
          <path d="M10 11v6" />
          <path d="M14 11v6" />
        </svg>
    ),
  },
];