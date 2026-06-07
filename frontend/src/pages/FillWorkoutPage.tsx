import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams, useSearchParams } from "react-router-dom";
import { apiUrl } from "../lib/api";

type WorkoutSetDetail = {
  id: string;
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

export default function FillWorkoutPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
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

  const handleSave = async () => {
    if (!id || workoutName.trim() === "") {
      setSaveStatus("error");
      return;
    }

    setSaveStatus("saving");

    try {
      for (const exercise of exercises) {
        for (const set of exercise.sets) {
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
          duration_minutes: parseInteger(workoutDuration),
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
    <main className="mx-auto max-w-6xl px-6 pb-16 pt-8 text-[#1f1b16]">
      <div className="mb-6 flex flex-wrap items-center gap-3">
        <Link
          to="/dashboard"
          className="inline-flex items-center rounded-[12px] border border-[#1f1b16]/15 px-3 py-2 text-xs font-extrabold uppercase tracking-[0.14em] text-[#3a332c] no-underline transition hover:bg-[#1f1b16]/5"
        >
          Volver al dashboard
        </Link>
        <span className="rounded-[12px] bg-[#ea7130]/12 px-3 py-2 text-xs font-extrabold uppercase tracking-[0.14em] text-[#ea7130]">
          Fill workout
        </span>
      </div>

      <section className="rounded-[28px] border border-[#1f1b16]/10 bg-[#fffaf0]/90 p-6 shadow-[0_24px_60px_rgba(31,27,22,0.08)]">
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
          <div>
            <p className="m-0 text-[12px] font-extrabold uppercase tracking-[0.30em] text-[#265c52]">
              Entrenamiento en curso
            </p>
            <input
              value={workoutName}
              onChange={(event) => setWorkoutName(event.target.value)}
              className="mt-3 w-full border-0 bg-transparent p-0 font-['Bricolage_Grotesque','Aptos_Display',sans-serif] text-[40px] font-black leading-[0.95] tracking-[-0.05em] text-[#1f1b16] outline-none"
            />
            <p className="mt-3 text-sm font-semibold text-[#3a332c]">
              Fecha asociada: <strong className="text-[#1f1b16]">{formatWorkoutDate(workout.performed_at ?? workout.planned_at)}</strong>
            </p>
            <textarea
              value={workoutNotes}
              onChange={(event) => setWorkoutNotes(event.target.value)}
              placeholder="Notas del entrenamiento"
              className="mt-4 min-h-28 w-full rounded-[20px] border border-[#1f1b16]/10 bg-white/70 px-4 py-3 text-sm font-medium text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15"
            />
          </div>

          <aside className="rounded-[22px] border border-[#1f1b16]/10 bg-[#1f1b16] p-5 text-[#fffaf0]">
            <p className="m-0 text-xs font-extrabold uppercase tracking-[0.22em] text-[#f1a45b]">Resumen</p>
            <div className="mt-4 grid gap-3">
              <SummaryStat label="Completadas" value={String(summary.completed)} />
              <SummaryStat label="Saltadas" value={String(summary.skipped)} />
              <SummaryStat label="Pendientes" value={String(summary.pending)} />
            </div>
            <label className="mt-5 block text-xs font-extrabold uppercase tracking-[0.18em] text-[#fffaf0]/75" htmlFor="workout-duration">
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
              className="mt-2 w-full rounded-[14px] border border-white/15 bg-white/10 px-3 py-2 text-sm font-semibold text-[#fffaf0] outline-none transition focus:border-[#f1a45b] focus:ring-4 focus:ring-[#f1a45b]/15"
            />
          </aside>
        </div>
      </section>

      <section className="mt-6 grid gap-4">
        {exercises.map((exercise) => (
          <article
            key={exercise.id}
            className="rounded-[24px] border border-[#1f1b16]/10 bg-[#fffaf0]/88 p-5 shadow-[0_18px_38px_rgba(31,27,22,0.06)]"
          >
            <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 className="m-0 font-['Bricolage_Grotesque','Aptos_Display',sans-serif] text-[24px] font-black tracking-[-0.04em] text-[#1f1b16]">
                  {exercise.name}
                </h2>
                <p className="mt-1 text-xs font-extrabold uppercase tracking-[0.18em] text-[#265c52]">
                  {exercise.muscleGroup || "Ejercicio"}
                </p>
              </div>
              {exercise.notes && (
                <span className="rounded-[12px] border border-[#1f1b16]/10 bg-white/80 px-3 py-2 text-xs font-semibold text-[#3a332c]">
                  {exercise.notes}
                </span>
              )}
            </div>

            <div className="grid gap-3">
              {exercise.sets.map((set) => {
                const isCompleted = set.status === "completed";
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

                      <div className="flex flex-wrap gap-2">
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
                        className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white px-3 py-2 text-sm font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15 disabled:cursor-not-allowed disabled:bg-[#1f1b16]/5 disabled:text-[#1f1b16]/35"
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
                        className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white px-3 py-2 text-sm font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/15 disabled:cursor-not-allowed disabled:bg-[#1f1b16]/5 disabled:text-[#1f1b16]/35"
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </article>
        ))}
      </section>

      {saveStatus === "error" && (
        <p className="mt-4 rounded-[16px] border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-4 py-3 text-sm font-bold text-[#9f2f22]">
          No se ha podido guardar el entrenamiento.
        </p>
      )}

      {cancelStatus === "error" && (
        <p className="mt-4 rounded-[16px] border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-4 py-3 text-sm font-bold text-[#9f2f22]">
          No se ha podido cancelar el entrenamiento.
        </p>
      )}

      <div className="mt-8 flex flex-wrap justify-end gap-3">
        <button
          type="button"
          onClick={handleCancel}
          disabled={cancelStatus === "loading" || saveStatus === "saving"}
          className="rounded-[16px] border border-[#1f1b16]/15 bg-transparent px-5 py-3 text-sm font-extrabold tracking-[0.04em] text-[#1f1b16] transition hover:bg-[#1f1b16]/5 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {cancelStatus === "loading" ? "Cancelando..." : "Cancelar entrenamiento"}
        </button>
        <button
          type="button"
          onClick={handleSave}
          disabled={saveStatus === "saving"}
          className="rounded-[16px] bg-[#ea7130] px-5 py-3 text-sm font-extrabold tracking-[0.04em] text-[#1f1b16] shadow-[0_18px_35px_rgba(234,113,48,0.28)] transition hover:-translate-y-px hover:bg-[#ff8b47] disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:translate-y-0"
        >
          {saveStatus === "saving" ? "Guardando..." : "Guardar entrenamiento"}
        </button>
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
