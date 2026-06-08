import { exerciseTypeLabel, muscleGroupLabel } from "../../lib/exerciseLabels";

export type AIRoutinePreviewSet = {
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

export type AIRoutinePreviewExercise = {
  exercise_id: string;
  name: string;
  muscle_group: string;
  exercise_type?: string;
  is_mandatory: boolean;
  sets?: AIRoutinePreviewSet[];
};

export type AIRoutinePreview = {
  name: string;
  objective: string;
  duration_minutes: number;
  target_muscles: string[];
  mandatory_count: number;
  generated_at: string;
  generation_source: string;
  exercises: AIRoutinePreviewExercise[];
};

type AIRoutinePreviewModalProps = {
  isOpen: boolean;
  isSaving: boolean;
  errorMessage: string;
  routine: AIRoutinePreview | null;
  onClose: () => void;
  onConfirm: () => void;
};

const generatedAtFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

function formatSetReps(set: AIRoutinePreviewSet) {
  if (set.target_reps_text) {
    return set.target_reps_text;
  }
  if (set.target_reps_min != null && set.target_reps_max != null) {
    if (set.target_reps_min === set.target_reps_max) {
      return `${set.target_reps_min}`;
    }
    return `${set.target_reps_min}-${set.target_reps_max}`;
  }
  if (set.target_reps_min != null) {
    return `${set.target_reps_min}+`;
  }
  if (set.target_reps_max != null) {
    return `Hasta ${set.target_reps_max}`;
  }
  return "-";
}

function formatSetLoad(set: AIRoutinePreviewSet) {
  if (set.target_weight_kg != null) {
    return `${set.target_weight_kg} kg`;
  }
  if (set.target_distance_km != null) {
    return `${set.target_distance_km} km`;
  }
  if (set.target_duration_seconds != null) {
    return `${set.target_duration_seconds} s`;
  }
  return "-";
}

function formatSetRest(set: AIRoutinePreviewSet) {
  return set.rest_seconds != null ? `${set.rest_seconds} s` : "-";
}

function formatGeneratedAt(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return generatedAtFormatter.format(date);
}

export default function AIRoutinePreviewModal({
  isOpen,
  isSaving,
  errorMessage,
  routine,
  onClose,
  onConfirm,
}: AIRoutinePreviewModalProps) {
  if (!isOpen || routine == null) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#1f1b16]/45 px-3 py-4 backdrop-blur-sm sm:px-4 sm:py-6">
      <div className="relative flex h-[calc(100vh-12rem)] w-full max-w-7xl flex-col overflow-hidden rounded-[1.75rem] border-2 border-[#fffaf0]/20 bg-[#fffaf0] bg-clip-padding shadow-[0_30px_80px_rgba(31,27,22,0.30),0_8px_22px_rgba(31,27,22,0.12)] sm:rounded-[2rem]">
        <div className="grid min-h-0 flex-1 gap-0 lg:grid-cols-[minmax(280px,0.82fr)_minmax(0,1.18fr)]">
          <aside className="flex flex-col justify-between bg-[#1f1b16] px-5 py-6 text-[#fffaf0] sm:px-6 sm:py-8 lg:sticky lg:top-0">
            <div>
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-xs font-extrabold uppercase tracking-[0.30em] text-[#f1a45b]">
                    Vista previa
                  </p>
                  <h3 className="mt-3 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-2xl font-black tracking-[-0.04em] sm:text-3xl">
                    Revisa la rutina antes de guardarla.
                  </h3>
                </div>
                <button
                  className="rounded-full border border-[#fffaf0]/15 px-3 py-2 text-xs font-black uppercase tracking-[0.18em] text-[#fffaf0] transition hover:bg-[#fffaf0] hover:text-[#1f1b16] disabled:cursor-not-allowed disabled:opacity-60"
                  type="button"
                  onClick={onClose}
                  disabled={isSaving}
                  aria-label="Cerrar vista previa"
                >
                  Cerrar
                </button>
              </div>

              <p className="mt-4 max-w-sm text-sm leading-6 text-[#efe4d2]">
                Si te encaja, la guardamos y el backend crea automaticamente
                cualquier ejercicio que falte en la base de datos.
              </p>

              <div className="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
                <div className="rounded-2xl border border-[#fffaf0]/15 bg-[#fffaf0]/10 p-4">
                  <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#f1a45b]">
                    Objetivo
                  </p>
                  <p className="mt-2 text-sm font-bold leading-6">
                    {routine.objective || "Sin objetivo"}
                  </p>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div className="rounded-2xl border border-[#fffaf0]/15 bg-[#fffaf0]/10 p-4">
                    <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#f1a45b]">
                      Duracion
                    </p>
                    <p className="mt-2 text-2xl font-black">
                      {routine.duration_minutes} min
                    </p>
                  </div>
                  <div className="rounded-2xl border border-[#fffaf0]/15 bg-[#fffaf0]/10 p-4">
                    <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#f1a45b]">
                      Ejercicios
                    </p>
                    <p className="mt-2 text-2xl font-black">
                      {routine.exercises.length}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex flex-wrap gap-3 pt-6">
              <button
                className="rounded-2xl border border-[#fffaf0]/20 px-4 py-3 text-sm font-black text-[#fffaf0] transition hover:bg-[#fffaf0] hover:text-[#1f1b16] disabled:cursor-not-allowed disabled:opacity-60"
                type="button"
                onClick={onClose}
                disabled={isSaving}
              >
                Cancelar
              </button>
              <button
                className="rounded-2xl bg-[#ea7130] px-4 py-3 text-sm font-black text-white transition hover:bg-[#f1a45b] disabled:cursor-not-allowed disabled:opacity-60"
                type="button"
                onClick={onConfirm}
                disabled={isSaving}
              >
                {isSaving ? "Guardando..." : "Guardar rutina"}
              </button>
            </div>
          </aside>

          <section className="flex min-h-0 flex-col bg-[#fffaf0] px-4 py-4 sm:px-6 sm:py-6 lg:px-8 lg:py-8">
            <div className="flex flex-wrap items-start justify-between gap-4 border-b border-[#1f1b16]/10 pb-4">
              <div className="min-w-0">
                <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#265c52]">
                  {routine.generation_source}
                </p>
                <h4 className="mt-2 break-words text-2xl font-black tracking-[-0.04em] text-[#1f1b16] sm:text-3xl">
                  {routine.name || "Rutina generada por IA"}
                </h4>
              </div>
              <div className="shrink-0 rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black text-[#265c52]">
                {routine.mandatory_count} obligatorios
              </div>
            </div>

            <div className="mt-4 grid gap-3 rounded-[1.35rem] border border-[#1f1b16]/10 bg-white/70 p-4 text-sm font-semibold text-[#5d5348] sm:grid-cols-2">
              <div>
                <span className="block text-[11px] font-black uppercase tracking-[0.18em] text-[#265c52]">
                  Musculos objetivo
                </span>
                <div className="mt-2 flex flex-wrap gap-2">
                  {routine.target_muscles && routine.target_muscles.length > 0 ? (
                    routine.target_muscles.map((muscle) => (
                      <span
                        className="rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-1 text-xs font-black uppercase tracking-[0.12em] text-[#1f1b16]"
                        key={muscle}
                      >
                        {muscleGroupLabel(muscle)}
                      </span>
                    ))
                  ) : (
                    <span className="leading-6">Sin especificar</span>
                  )}
                </div>
              </div>
              <div>
                <span className="block text-[11px] font-black uppercase tracking-[0.18em] text-[#265c52]">
                  Generada
                </span>
                <p className="mt-2 leading-6">{formatGeneratedAt(routine.generated_at)}</p>
              </div>
            </div>

            <div className="mt-4 min-h-0 flex-1 overflow-y-auto pr-1">
              {routine.exercises.length === 0 ? (
                <p className="rounded-2xl border border-dashed border-[#1f1b16]/15 bg-[#fff8ea] p-5 text-sm font-bold text-[#5d5348]">
                  La IA no devolvio ejercicios validos para esta rutina.
                </p>
              ) : (
                <div className="space-y-3">
                  {routine.exercises.map((exercise) => (
                    <article
                      className="rounded-[1.35rem] border border-[#1f1b16]/10 bg-white p-4 shadow-[0_10px_25px_rgba(47,39,27,0.07)]"
                      key={`${exercise.exercise_id}-${exercise.name}`}
                    >
                      <div className="flex flex-wrap items-start justify-between gap-3">
                        <div className="min-w-0">
                          <h5 className="break-words text-lg font-black text-[#1f1b16]">
                            {exercise.name}
                          </h5>
                          <p className="mt-1 text-[11px] font-bold uppercase tracking-[0.16em] text-[#7a6b5c]">
                            {muscleGroupLabel(exercise.muscle_group)}
                            {exercise.exercise_type
                              ? ` · ${exerciseTypeLabel(exercise.exercise_type)}`
                              : ""}
                          </p>
                        </div>
                        {exercise.is_mandatory && (
                          <span className="shrink-0 rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black text-[#265c52]">
                            Obligatorio
                          </span>
                        )}
                      </div>

                      {exercise.sets && exercise.sets.length > 0 ? (
                        <div className="mt-4 overflow-x-auto">
                          <table className="w-full min-w-[520px] text-left text-sm">
                            <thead className="text-[11px] uppercase tracking-[0.14em] text-[#7a6b5c]">
                              <tr>
                                <th className="py-2 pr-3">Serie</th>
                                <th className="py-2 pr-3">Reps</th>
                                <th className="py-2 pr-3">Carga</th>
                                <th className="py-2 pr-3">RIR</th>
                                <th className="py-2 pr-3">Descanso</th>
                              </tr>
                            </thead>
                            <tbody className="font-bold text-[#1f1b16]">
                              {exercise.sets.map((set) => (
                                <tr
                                  className="border-t border-[#1f1b16]/10"
                                  key={`${exercise.exercise_id}-${set.set_number}`}
                                >
                                  <td className="py-2 pr-3">{set.set_number}</td>
                                  <td className="py-2 pr-3">{formatSetReps(set)}</td>
                                  <td className="py-2 pr-3">{formatSetLoad(set)}</td>
                                  <td className="py-2 pr-3">
                                    {set.target_rir != null ? set.target_rir : "-"}
                                  </td>
                                  <td className="py-2 pr-3">{formatSetRest(set)}</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      ) : (
                        <p className="mt-4 text-sm font-semibold text-[#5d5348]">
                          La rutina no incluye series planificadas para este ejercicio.
                        </p>
                      )}
                    </article>
                  ))}
                </div>
              )}
            </div>

            {errorMessage && (
              <p className="mt-4 rounded-2xl border border-[#9b2d20]/20 bg-[#fff0ed] p-4 text-sm font-bold text-[#9b2d20]">
                {errorMessage}
              </p>
            )}
          </section>
        </div>
      </div>
    </div>
  );
}
