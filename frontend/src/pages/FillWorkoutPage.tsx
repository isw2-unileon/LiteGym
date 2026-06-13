import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { Card, CardHeader } from "../components/Card";
import { apiUrl } from "../lib/api";
import { MobileDisclosure } from "../components/MobileDisclosure";
import { useIsMobile } from "../lib/useIsMobile";

type WorkoutSetDetail = {
  id: string;
  routine_exercise_set_id?: string;
  set_number: number;
  target_reps_min?: number;
  target_reps_max?: number;
  target_reps_text?: string;
  target_weight_kg?: number;
  reps?: number;
  weight_kg?: number;
  completed?: boolean | null;
};

type WorkoutExerciseDetail = {
  id: string;
  name: string;
  muscle_group: string;
  exercise_order: number;
  notes?: string;
  sets: WorkoutSetDetail[];
};

type WorkoutDetail = {
  id: string;
  name: string;
  planned_at?: string | null;
  performed_at?: string | null;
  notes?: string | null;
  exercises: WorkoutExerciseDetail[];
};

type EditableSet = {
  id: string;
  setNumber: number;
  targetLabel: string;
  targetWeightKg?: number;
  reps: string;
  weightKg: string;
  status: "pending" | "completed" | "skipped";
  isPersisted: boolean;
  isTemplateSet: boolean;
};

type EditableExercise = {
  id: string;
  name: string;
  muscleGroup: string;
  notes: string;
  sets: EditableSet[];
};

const dateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "numeric",
  month: "long",
  hour: "2-digit",
  minute: "2-digit",
});

function formatWorkoutDate(value?: string | null) {
  if (!value) {
    return "Sin fecha";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Sin fecha";
  }

  return dateFormatter.format(date);
}

function buildTargetLabel(set: WorkoutSetDetail) {
  if (set.target_reps_text && set.target_reps_text.trim() !== "") {
    return set.target_reps_text;
  }
  if (typeof set.target_reps_min === "number" && typeof set.target_reps_max === "number") {
    return `${set.target_reps_min}-${set.target_reps_max} reps`;
  }
  if (typeof set.target_reps_min === "number") {
    return `${set.target_reps_min} reps`;
  }
  if (typeof set.target_reps_max === "number") {
    return `${set.target_reps_max} reps`;
  }
  return "Sin objetivo";
}

function resolveInitialSetStatus(detail: WorkoutDetail, set: WorkoutSetDetail): EditableSet["status"] {
  const isPlannedWorkout = !detail.performed_at && Boolean(detail.planned_at);
  const hasRecordedValues = typeof set.reps === "number" || typeof set.weight_kg === "number";

  if (set.completed === true) {
    return "completed";
  }

  if (set.completed === false) {
    return isPlannedWorkout && !hasRecordedValues ? "pending" : "skipped";
  }

  return "pending";
}

function buildEditableExercises(detail: WorkoutDetail): EditableExercise[] {
  return detail.exercises
    .slice()
    .sort((left, right) => left.exercise_order - right.exercise_order)
    .map((exercise) => ({
      id: exercise.id,
      name: exercise.name,
      muscleGroup: exercise.muscle_group,
      notes: exercise.notes ?? "",
      sets: exercise.sets.map((set) => ({
        id: set.id,
        setNumber: set.set_number,
        targetLabel: buildTargetLabel(set),
        targetWeightKg: set.target_weight_kg,
        reps: typeof set.reps === "number" ? String(set.reps) : "",
        weightKg: typeof set.weight_kg === "number" ? String(set.weight_kg) : "",
        status: resolveInitialSetStatus(detail, set),
        isPersisted: true,
        isTemplateSet: Boolean(set.routine_exercise_set_id),
      })),
    }));
}

function parseInteger(value: string) {
  if (value.trim() === "") {
    return null;
  }
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : null;
}

function parseDecimal(value: string) {
  const normalized = value.trim().replace(",", ".");
  if (normalized === "") {
    return null;
  }
  const parsed = Number(normalized);
  return Number.isFinite(parsed) ? parsed : null;
}

function isWholeNumber(value: string) {
  return /^\d+$/.test(value.trim());
}

function isDecimalNumber(value: string) {
  return /^\d+([.,]\d+)?$/.test(value.trim());
}

export default function FillWorkoutPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const isMobile = useIsMobile();
  const [workout, setWorkout] = useState<WorkoutDetail | null>(null);
  const [workoutName, setWorkoutName] = useState("");
  const [workoutNotes, setWorkoutNotes] = useState("");
  const [workoutDuration, setWorkoutDuration] = useState("");
  const [exercises, setExercises] = useState<EditableExercise[]>([]);
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [saveStatus, setSaveStatus] = useState<"idle" | "saving" | "error">("idle");
  const [cancelStatus, setCancelStatus] = useState<"idle" | "loading" | "error">("idle");
  const preserveOnCancel = searchParams.get("source") === "planned";

  useEffect(() => {
    if (!id) {
      setStatus("error");
      return;
    }

    let mounted = true;
    setStatus("loading");

    void fetch(apiUrl(`/api/workout/${id}/detail`), {
      credentials: "include",
    })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error("failed");
        }

        const payload = (await response.json()) as WorkoutDetail;
        if (!mounted) {
          return;
        }

        setWorkout(payload);
        setWorkoutName(payload.name);
        setWorkoutNotes(payload.notes ?? "");
        setExercises(buildEditableExercises(payload));
        setStatus("success");
      })
      .catch(() => {
        if (!mounted) {
          return;
        }
        setStatus("error");
      });

    return () => {
      mounted = false;
    };
  }, [id]);

  const summary = useMemo(() => {
    return exercises.reduce(
      (acc, exercise) => {
        for (const set of exercise.sets) {
          acc[set.status] += 1;
        }
        return acc;
      },
      { pending: 0, completed: 0, skipped: 0 },
    );
  }, [exercises]);

  const durationMinutes = useMemo(() => parseInteger(workoutDuration), [workoutDuration]);

  const setValidationErrors = useMemo(() => {
    const errors = new Map<string, { reps?: string; weightKg?: string }>();

    for (const exercise of exercises) {
      for (const set of exercise.sets) {
        if (set.status !== "completed") {
          continue;
        }

        const setErrors: { reps?: string; weightKg?: string } = {};

        if (!isWholeNumber(set.reps) || parseInteger(set.reps) === null || (parseInteger(set.reps) ?? 0) <= 0) {
          setErrors.reps = "Introduce unas repeticiones validas";
        }

        if (!isDecimalNumber(set.weightKg) || parseDecimal(set.weightKg) === null || (parseDecimal(set.weightKg) ?? 0) < 0) {
          setErrors.weightKg = "Introduce un peso valido";
        }

        if (setErrors.reps || setErrors.weightKg) {
          errors.set(set.id, setErrors);
        }
      }
    }

    return errors;
  }, [exercises]);

  const durationError = useMemo(() => {
    if (workoutDuration.trim() === "") {
      return "";
    }
    if (!isWholeNumber(workoutDuration)) {
      return "La duracion debe ser un numero entero";
    }
    return "";
  }, [workoutDuration]);

  const hasValidationErrors = setValidationErrors.size > 0 || durationError !== "";

  const updateSet = (exerciseID: string, setID: string, updater: (set: EditableSet) => EditableSet) => {
    setExercises((current) =>
      current.map((exercise) =>
        exercise.id !== exerciseID
          ? exercise
          : {
            ...exercise,
            sets: exercise.sets.map((set) => (set.id === setID ? updater(set) : set)),
          },
      ),
    );
  };

  const addSet = (exerciseID: string) => {
    setExercises((current) =>
      current.map((exercise) => {
        if (exercise.id !== exerciseID) {
          return exercise;
        }

        const nextSetNumber = exercise.sets.length > 0
          ? Math.max(...exercise.sets.map((set) => set.setNumber)) + 1
          : 1;

        return {
          ...exercise,
          sets: [
            ...exercise.sets,
            {
              id: `draft-${exerciseID}-${Date.now()}-${nextSetNumber}`,
              setNumber: nextSetNumber,
              targetLabel: "Sin objetivo",
              reps: "",
              weightKg: "",
              status: "pending",
              isPersisted: false,
              isTemplateSet: false,
            },
          ],
        };
      }),
    );
  };

  const removeSet = async (exerciseID: string, setID: string) => {
    const targetExercise = exercises.find((exercise) => exercise.id === exerciseID);
    const targetSet = targetExercise?.sets.find((set) => set.id === setID);
    if (!targetSet) {
      return;
    }

    if (!targetSet.isPersisted) {
      setExercises((current) =>
        current.map((exercise) =>
          exercise.id !== exerciseID
            ? exercise
            : {
              ...exercise,
              sets: exercise.sets.filter((set) => set.id !== setID),
            },
        ),
      );
      return;
    }

    try {
      const response = await fetch(apiUrl(`/api/workout/${id}/exercises/${exerciseID}/sets/${setID}`), {
        method: "DELETE",
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("remove failed");
      }

      setExercises((current) =>
        current.map((exercise) =>
          exercise.id !== exerciseID
            ? exercise
            : {
              ...exercise,
              sets: exercise.sets.filter((set) => set.id !== setID),
            },
        ),
      );
    } catch {
      setSaveStatus("error");
    }
  };

  const handleSave = async () => {
    if (!id || workoutName.trim() === "" || hasValidationErrors) {
      setSaveStatus("error");
      return;
    }

    setSaveStatus("saving");

    try {
        for (const exercise of exercises) {
        for (const set of exercise.sets) {
          if (!set.isPersisted) {
            const createResponse = await fetch(apiUrl(`/api/workout/${id}/exercises/${exercise.id}/set`), {
              method: "POST",
              credentials: "include",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify({
                set_number: set.setNumber,
                target_reps_text: set.targetLabel === "Sin objetivo" ? "" : set.targetLabel,
                completed: false,
              }),
            });

            if (!createResponse.ok) {
              throw new Error("set create failed");
            }

            const createdSet = (await createResponse.json()) as { id: string };
            set.id = createdSet.id;
            set.isPersisted = true;
          }

          const isCompleted = set.status === "completed";
          const completed = set.status === "completed" ? true : false;

          const response = await fetch(apiUrl(`/api/workout/${id}/exercises/${exercise.id}/sets/${set.id}`), {
            method: "POST",
            credentials: "include",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              set_number: set.setNumber,
              completed,
              reps: isCompleted ? parseInteger(set.reps) : null,
              weight_kg: isCompleted ? parseDecimal(set.weightKg) : null,
            }),
          });

          if (!response.ok) {
            throw new Error("set update failed");
          }
        }
      }

      const finishResponse = await fetch(apiUrl(`/api/workout/${id}/finish`), {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: workoutName.trim(),
          notes: workoutNotes.trim() === "" ? null : workoutNotes.trim(),
          duration_minutes: durationMinutes !== null && durationMinutes > 0 ? durationMinutes : null,
        }),
      });

      if (!finishResponse.ok) {
        throw new Error("finish failed");
      }

      navigate("/dashboard", { replace: true });
    } catch {
      setSaveStatus("error");
    }
  };

  const handleCancel = async () => {
    if (!id) {
      return;
    }

    if (preserveOnCancel) {
      navigate("/dashboard", { replace: true });
      return;
    }

    setCancelStatus("loading");

    try {
      const response = await fetch(apiUrl(`/api/workout/${id}`), {
        method: "DELETE",
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("cancel failed");
      }

      navigate("/dashboard", { replace: true });
    } catch {
      setCancelStatus("error");
    }
  };

  if (status === "loading") {
    return <div className="px-6 py-12 text-sm font-semibold text-[#3a332c]">Cargando entrenamiento...</div>;
  }

  if (status === "error" || !workout) {
    return (
      <div className="px-6 py-12">
        <p className="rounded-[18px] border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-4 py-3 text-sm font-bold text-[#9f2f22]">
          No se ha podido cargar este entrenamiento.
        </p>
      </div>
    );
  }

  return (
    <main className="relative isolate overflow-x-hidden pb-16 pt-5 text-[#1f1b16] sm:pt-8">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]"
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10"
      />

      <div className="mx-auto max-w-[1280px] px-4 sm:px-6 md:px-8">
        <section className="mb-6 grid grid-cols-1 items-start gap-6">
          <div>
            <div className="mb-2.5 flex items-center gap-2.5 text-[12px] font-extrabold uppercase tracking-[0.22em] text-[#265c52] sm:text-[14px] sm:tracking-[0.30em]">
              <span className="inline-block h-0.5 w-5 bg-[#265c52] sm:w-6" />
              ENTRENAMIENTO
            </div>
            <h1 className="m-0 break-words font-['Bricolage_Grotesque','Aptos_Display',sans-serif] text-[clamp(2.35rem,14vw,3.35rem)] font-black leading-[0.92] tracking-[-0.055em] text-[#1f1b16] sm:text-[64px]">
              <span className="px-1 text-[#ea7130] [background:linear-gradient(180deg,transparent_60%,rgba(234,113,48,0.18)_60%)]">
                {workout.name || "Sesion"}
              </span>
            </h1>
            <p className="mt-3.5 max-w-[680px] text-[15px] leading-[1.55] text-[#3a332c]">
              Ajusta lo que realmente hiciste en cada ejercicio, añade sets cuando lo necesites y guarda el entrenamiento final sin tocar la rutina original.
            </p>
          </div>
        </section>

        <section className="grid items-stretch gap-[18px] xl:grid-cols-[minmax(0,1.55fr)_minmax(20rem,0.95fr)]">
          <Card accent="#ea7130" className="h-full">
              <CardHeader
                kicker="Sesion"
                title="Resumen del entrenamiento"
                rightChip={formatWorkoutDate(workout.performed_at ?? workout.planned_at)}
              />

              <div className="relative z-[2] mt-5 grid gap-5">
                <div className="grid gap-2">
                  <span className="text-[11px] font-extrabold uppercase tracking-[0.2em] text-[#265c52]">
                    Rutina base
                  </span>
                  <p className="m-0 font-['Bricolage_Grotesque','Aptos_Display',sans-serif] text-[32px] font-black leading-[0.98] tracking-[-0.04em] text-[#1f1b16] sm:text-[38px]">
                    {workout.name || "Sesion"}
                  </p>
                </div>

                <div className="grid gap-2">
                  <label
                    htmlFor="workout-name"
                    className="text-[11px] font-extrabold uppercase tracking-[0.2em] text-[#265c52]"
                  >
                    Nombre del entrenamiento
                  </label>
                  <input
                    id="workout-name"
                    value={workoutName}
                    onChange={(event) => setWorkoutName(event.target.value)}
                    placeholder="Ponle un nombre a este entrenamiento"
                    className="w-full rounded-[20px] border border-[#1f1b16]/10 bg-white/80 px-4 py-3 font-['Bricolage_Grotesque','Aptos_Display',sans-serif] text-[26px] font-black leading-[1] tracking-[-0.04em] text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15 sm:text-[30px]"
                  />
                </div>

                <div className="grid gap-2">
                  <label
                    htmlFor="workout-notes"
                    className="text-[11px] font-extrabold uppercase tracking-[0.2em] text-[#265c52]"
                  >
                    Notas sobre el entrenamiento
                  </label>
                  <textarea
                    id="workout-notes"
                    value={workoutNotes}
                    onChange={(event) => setWorkoutNotes(event.target.value)}
                    placeholder="Notas del entrenamiento"
                    className="min-h-36 w-full resize-y rounded-[20px] border border-[#1f1b16]/10 bg-white/70 px-4 py-3 text-sm font-medium text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15"
                  />
                </div>
              </div>
          </Card>

          <Card accent="#ea7130" dark className="h-full self-stretch">
            <CardHeader kicker="Control" title="Estado del registro" onDark />
            <div className="relative z-[2] mt-4 grid gap-4">
              <SummaryStat label="Completadas" value={String(summary.completed)} />
              <SummaryStat label="Saltadas" value={String(summary.skipped)} />
              <SummaryStat label="Pendientes" value={String(summary.pending)} />
              <div className="rounded-[16px] border border-white/10 bg-white/6 px-3 py-3">
                <label className="block text-xs font-extrabold uppercase tracking-[0.18em] text-[#fffaf0]/75" htmlFor="workout-duration">
                  Duracion total
                </label>
                <input
                  id="workout-duration"
                  type="number"
                  min="0"
                  inputMode="numeric"
                  value={workoutDuration}
                  onChange={(event) => setWorkoutDuration(event.target.value)}
                  placeholder="Minutos"
                  className={[
                    "mt-2 w-full rounded-[14px] border bg-white/10 px-3 py-3 text-sm font-semibold text-[#fffaf0] outline-none transition focus:border-[#f1a45b] focus:ring-4 focus:ring-[#f1a45b]/15",
                    durationError ? "border-[#ffb38a]" : "border-white/15",
                  ].join(" ")}
                />
                <p className="mt-3 text-xs font-semibold text-[#fffaf0]/75">
                  {durationMinutes !== null && durationMinutes > 0 ? `${durationMinutes} min` : "Sin tiempo"}
                </p>
                {durationError && (
                  <p className="mt-2 text-xs font-bold text-[#ffb38a]">{durationError}</p>
                )}
              </div>
            </div>
          </Card>
        </section>

        {isMobile && (
          <section className="mt-[18px] grid gap-4">
            {exercises.map((exercise) => (
              <MobileDisclosure
                key={`mobile-${exercise.id}`}
                kicker={exercise.muscleGroup || "Ejercicio"}
                title={exercise.name}
                defaultOpen={false}
              >
                {exercise.notes && (
                  <p className="mb-3 rounded-[12px] border border-[#1f1b16]/10 bg-white/70 px-3 py-2 text-xs font-semibold text-[#3a332c]">
                    {exercise.notes}
                  </p>
                )}

                <div className="mb-3 flex justify-stretch">
                  <button
                    type="button"
                    onClick={() => addSet(exercise.id)}
                    className="w-full rounded-[12px] border border-[#ea7130]/25 bg-[#ea7130]/10 px-3.5 py-2 text-xs font-extrabold uppercase tracking-[0.14em] text-[#ea7130]"
                  >
                    Añadir set
                  </button>
                </div>

                <div className="grid gap-3">
                  {exercise.sets.length === 0 && (
                    <div className="rounded-[20px] border border-dashed border-[#1f1b16]/14 bg-white/60 px-4 py-5 text-sm font-semibold text-[#3a332c]">
                      Este ejercicio no tiene sets planificados todavia.
                    </div>
                  )}
                  {exercise.sets.map((set) => {
                    const isCompleted = set.status === "completed";
                    const errors = setValidationErrors.get(set.id);
                    const statusClass =
                      set.status === "completed"
                        ? "border-[#265c52]/25 bg-[#265c52]/8"
                        : set.status === "skipped"
                          ? "border-[#9f2f22]/20 bg-[#9f2f22]/6"
                          : "border-[#1f1b16]/10 bg-white/70";

                    return (
                      <div key={set.id} className={`rounded-[20px] border p-4 ${statusClass}`}>
                        <p className="m-0 text-sm font-extrabold text-[#1f1b16]">Set {set.setNumber}</p>
                        <p className="mt-1 text-xs font-semibold text-[#3a332c]">
                          Objetivo: {set.targetLabel}
                          {typeof set.targetWeightKg === "number" ? ` · ${set.targetWeightKg} kg` : ""}
                        </p>
                        <div className="mt-3 grid grid-cols-3 gap-2">
                          {(["pending", "completed", "skipped"] as const).map((statusValue) => (
                            <button
                              key={statusValue}
                              type="button"
                              onClick={() =>
                                updateSet(exercise.id, set.id, (current) => ({
                                  ...current,
                                  status: statusValue,
                                  reps: statusValue === "completed" ? current.reps : "",
                                  weightKg: statusValue === "completed" ? current.weightKg : "",
                                }))
                              }
                              className={[
                                "rounded-[12px] px-2 py-2 text-[11px] font-extrabold uppercase tracking-[0.08em]",
                                set.status === statusValue
                                  ? "bg-[#1f1b16] text-[#fffaf0]"
                                  : "border border-[#1f1b16]/12 bg-white/80 text-[#3a332c]",
                              ].join(" ")}
                            >
                              {statusValue === "pending" ? "Pend." : statusValue === "completed" ? "Hecho" : "Salto"}
                            </button>
                          ))}
                        </div>
                        <div className="mt-3 grid grid-cols-2 gap-2">
                          <input
                            type="number"
                            min="0"
                            inputMode="numeric"
                            disabled={!isCompleted}
                            value={set.reps}
                            onChange={(event) =>
                              updateSet(exercise.id, set.id, (current) => ({
                                ...current,
                                reps: event.target.value,
                              }))
                            }
                            placeholder="Reps"
                            className={[
                              "w-full rounded-[14px] border bg-white px-3 py-3 text-sm font-semibold text-[#1f1b16] outline-none disabled:bg-[#1f1b16]/5 disabled:text-[#1f1b16]/35",
                              errors?.reps ? "border-[#9f2f22]" : "border-[#1f1b16]/12",
                            ].join(" ")}
                          />
                          <input
                            type="number"
                            min="0"
                            step="0.5"
                            inputMode="decimal"
                            disabled={!isCompleted}
                            value={set.weightKg}
                            onChange={(event) =>
                              updateSet(exercise.id, set.id, (current) => ({
                                ...current,
                                weightKg: event.target.value,
                              }))
                            }
                            placeholder="Kg"
                            className={[
                              "w-full rounded-[14px] border bg-white px-3 py-3 text-sm font-semibold text-[#1f1b16] outline-none disabled:bg-[#1f1b16]/5 disabled:text-[#1f1b16]/35",
                              errors?.weightKg ? "border-[#9f2f22]" : "border-[#1f1b16]/12",
                            ].join(" ")}
                          />
                        </div>
                        {!set.isTemplateSet && (
                          <button
                            type="button"
                            onClick={() => void removeSet(exercise.id, set.id)}
                            className="mt-3 w-full rounded-[12px] border border-[#9f2f22]/20 bg-[#9f2f22]/6 px-3 py-2 text-xs font-extrabold uppercase tracking-[0.14em] text-[#9f2f22]"
                          >
                            Eliminar set
                          </button>
                        )}
                      </div>
                    );
                  })}
                </div>
              </MobileDisclosure>
            ))}
          </section>
        )}

        {!isMobile && (
          <section className="mt-[18px] grid gap-4">
            {exercises.map((exercise) => (
          <Card
            key={exercise.id}
            accent="#ea7130"
          >
            <CardHeader
              kicker={exercise.muscleGroup || "Ejercicio"}
              title={exercise.name}
              right={exercise.notes ? (
                <span className="rounded-[12px] border border-[#1f1b16]/10 bg-white/80 px-3 py-2 text-xs font-semibold text-[#3a332c]">
                  {exercise.notes}
                </span>
              ) : undefined}
            />

            <div className="relative z-[2] mt-5 flex justify-stretch sm:justify-end">
              <button
                type="button"
                onClick={() => addSet(exercise.id)}
                className="w-full rounded-[12px] border border-[#ea7130]/25 bg-[#ea7130]/10 px-3.5 py-2 text-xs font-extrabold uppercase tracking-[0.14em] text-[#ea7130] transition hover:bg-[#ea7130]/15 sm:w-auto"
              >
                Añadir set
              </button>
            </div>

            <div className="relative z-[2] mt-3 grid gap-3">
              {exercise.sets.length === 0 && (
                <div className="rounded-[20px] border border-dashed border-[#1f1b16]/14 bg-white/60 px-4 py-5 text-sm font-semibold text-[#3a332c]">
                  Este ejercicio no tiene sets planificados todavia. Puedes anadir el primero cuando quieras.
                </div>
              )}
              {exercise.sets.map((set) => {
                const isCompleted = set.status === "completed";
                const errors = setValidationErrors.get(set.id);
                const statusClass =
                  set.status === "completed"
                    ? "border-[#265c52]/25 bg-[#265c52]/8"
                    : set.status === "skipped"
                      ? "border-[#9f2f22]/20 bg-[#9f2f22]/6"
                      : "border-[#1f1b16]/10 bg-white/70";

                return (
                  <div key={set.id} className={`rounded-[20px] border p-4 ${statusClass}`}>
                    <div className="grid gap-4 lg:grid-cols-[auto_minmax(0,1fr)_12rem_12rem] lg:items-center">
                      <div>
                        <p className="m-0 text-sm font-extrabold text-[#1f1b16]">Set {set.setNumber}</p>
                        <p className="mt-1 text-xs font-semibold text-[#3a332c]">
                          Objetivo: {set.targetLabel}
                          {typeof set.targetWeightKg === "number" ? ` · ${set.targetWeightKg} kg` : ""}
                        </p>
                      </div>

                      <div className="grid grid-cols-1 gap-2 sm:flex sm:flex-wrap">
                        {(["pending", "completed", "skipped"] as const).map((statusValue) => (
                          <button
                            key={statusValue}
                            type="button"
                            onClick={() =>
                              updateSet(exercise.id, set.id, (current) => ({
                                ...current,
                                status: statusValue,
                                reps: statusValue === "completed" ? current.reps : "",
                                weightKg: statusValue === "completed" ? current.weightKg : "",
                              }))
                            }
                            className={[
                              "rounded-[12px] px-3 py-2 text-xs font-extrabold uppercase tracking-[0.14em] transition",
                              set.status === statusValue
                                ? "bg-[#1f1b16] text-[#fffaf0]"
                                : "border border-[#1f1b16]/12 bg-white/80 text-[#3a332c] hover:bg-[#1f1b16]/5",
                            ].join(" ")}
                          >
                            {statusValue === "pending" ? "Pendiente" : statusValue === "completed" ? "Hecho" : "Saltado"}
                          </button>
                        ))}
                        {!set.isTemplateSet && (
                          <button
                            type="button"
                            onClick={() => void removeSet(exercise.id, set.id)}
                            className="rounded-[12px] border border-[#9f2f22]/20 bg-[#9f2f22]/6 px-3 py-2 text-xs font-extrabold uppercase tracking-[0.14em] text-[#9f2f22] transition hover:bg-[#9f2f22]/12"
                          >
                            Eliminar
                          </button>
                        )}
                      </div>

                      <input
                        type="number"
                        min="0"
                        inputMode="numeric"
                        disabled={!isCompleted}
                        value={set.reps}
                        onChange={(event) =>
                          updateSet(exercise.id, set.id, (current) => ({
                            ...current,
                            reps: event.target.value,
                          }))
                        }
                        placeholder="Reps reales"
                        className={[
                          "w-full rounded-[14px] border bg-white px-3 py-2 text-sm font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15 disabled:cursor-not-allowed disabled:bg-[#1f1b16]/5 disabled:text-[#1f1b16]/35",
                          errors?.reps ? "border-[#9f2f22]" : "border-[#1f1b16]/12",
                        ].join(" ")}
                      />

                      <input
                        type="number"
                        min="0"
                        step="0.5"
                        inputMode="decimal"
                        disabled={!isCompleted}
                        value={set.weightKg}
                        onChange={(event) =>
                          updateSet(exercise.id, set.id, (current) => ({
                            ...current,
                            weightKg: event.target.value,
                          }))
                        }
                        placeholder="Peso real (kg)"
                        className={[
                          "w-full rounded-[14px] border bg-white px-3 py-2 text-sm font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15 disabled:cursor-not-allowed disabled:bg-[#1f1b16]/5 disabled:text-[#1f1b16]/35",
                          errors?.weightKg ? "border-[#9f2f22]" : "border-[#1f1b16]/12",
                        ].join(" ")}
                      />
                    </div>
                    {(errors?.reps || errors?.weightKg) && (
                      <div className="mt-3 flex flex-wrap gap-3 text-xs font-bold text-[#9f2f22]">
                        {errors.reps && <span>{errors.reps}</span>}
                        {errors.weightKg && <span>{errors.weightKg}</span>}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </Card>
            ))}
          </section>
        )}

      {saveStatus === "error" && (
        <p className="mt-4 rounded-[16px] border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-4 py-3 text-sm font-bold text-[#9f2f22]">
          {hasValidationErrors ? "Corrige los campos numericos antes de guardar." : "No se ha podido guardar el entrenamiento."}
        </p>
      )}

      {cancelStatus === "error" && (
        <p className="mt-4 rounded-[16px] border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-4 py-3 text-sm font-bold text-[#9f2f22]">
          No se ha podido cancelar el entrenamiento.
        </p>
      )}

        <div className="sticky bottom-[6.25rem] z-20 mt-8 flex flex-col-reverse gap-3 rounded-[22px] border border-[#1f1b16]/10 bg-[#fffaf0]/92 p-3 shadow-[0_18px_45px_rgba(31,27,22,0.16)] backdrop-blur-md sm:static sm:flex-row sm:flex-wrap sm:justify-end sm:border-0 sm:bg-transparent sm:p-0 sm:shadow-none sm:backdrop-blur-0">
          <button
            type="button"
            onClick={handleCancel}
            disabled={cancelStatus === "loading" || saveStatus === "saving"}
            className="w-full rounded-[16px] border border-[#1f1b16]/15 bg-transparent px-5 py-3 text-sm font-extrabold tracking-[0.04em] text-[#1f1b16] transition hover:bg-[#1f1b16]/5 disabled:cursor-not-allowed disabled:opacity-50 sm:w-auto"
          >
            {cancelStatus === "loading" ? "Cancelando..." : "Cancelar entrenamiento"}
          </button>
          <button
            type="button"
            onClick={handleSave}
            disabled={saveStatus === "saving" || hasValidationErrors}
            className="w-full rounded-[16px] bg-[#ea7130] px-5 py-3 text-sm font-extrabold tracking-[0.04em] text-[#1f1b16] shadow-[0_18px_35px_rgba(234,113,48,0.28)] transition hover:-translate-y-px hover:bg-[#ff8b47] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:translate-y-0 sm:w-auto"
          >
            {saveStatus === "saving" ? "Guardando..." : "Guardar entrenamiento"}
          </button>
        </div>
      </div>
    </main>
  );
}

function SummaryStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-[16px] border border-white/10 bg-white/6 px-3 py-3">
      <div className="font-['Bricolage_Grotesque','Aptos_Display',sans-serif] text-[30px] font-black leading-none tracking-[-0.04em] text-[#fffaf0]">
        {value}
      </div>
      <div className="mt-1 text-xs font-bold uppercase tracking-[0.16em] text-[#fffaf0]/75">{label}</div>
    </div>
  );
}
