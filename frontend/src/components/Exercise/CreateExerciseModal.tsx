import { useId, useState } from "react";

export type CreateExercisePayload = {
  name: string;
  description: string;
  muscle_group: string;
  secondary_muscle_group: string;
  exercise_type: string;
  is_official: boolean;
};

type CreateExerciseModalProps = {
  isOpen: boolean;
  isSubmitting: boolean;
  errorMessage: string;
  onClose: () => void;
  onSubmit: (payload: CreateExercisePayload) => Promise<void>;
};

const initialForm = {
  name: "",
  description: "",
  muscle_group: "",
  secondary_muscle_group: "",
  exercise_type: "",
  is_official: false,
} satisfies CreateExercisePayload;

export default function CreateExerciseModal({
  isOpen,
  isSubmitting,
  errorMessage,
  onClose,
  onSubmit,
}: CreateExerciseModalProps) {
  const [form, setForm] = useState<CreateExercisePayload>(initialForm);
  const [validationMessage, setValidationMessage] = useState("");

  const nameId = useId();
  const descriptionId = useId();
  const muscleGroupId = useId();
  const secondaryMuscleGroupId = useId();
  const exerciseTypeId = useId();

  if (!isOpen) {
    return null;
  }

  const handleChange =
    (field: keyof CreateExercisePayload) =>
    (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const value =
        event.target instanceof HTMLInputElement &&
        event.target.type === "checkbox"
          ? event.target.checked
          : event.target.value;

      setForm((current) => ({
        ...current,
        [field]: value,
      }));
      setValidationMessage("");
    };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (form.name.trim() === "" || form.muscle_group.trim() === "") {
      setValidationMessage(
        "Nombre y grupo muscular son obligatorios para crear el ejercicio.",
      );
      return;
    }

    await onSubmit({
      ...form,
      name: form.name.trim(),
      description: form.description.trim(),
      muscle_group: form.muscle_group.trim(),
      secondary_muscle_group: form.secondary_muscle_group.trim(),
      exercise_type: form.exercise_type.trim(),
    });
  };

  const handleClose = () => {
    if (isSubmitting) {
      return;
    }

    setForm(initialForm);
    setValidationMessage("");
    onClose();
  };

  const feedbackMessage = validationMessage || errorMessage;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-[#1f1b16]/55 px-4 py-6 backdrop-blur-sm"
      data-block="create-exercise-modal-overlay"
    >
      <div
        className="relative w-full max-w-4xl overflow-hidden rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0] shadow-[0_40px_120px_rgba(31,27,22,0.32)]"
        data-block="create-exercise-modal"
      >
        <div className="grid gap-0 lg:grid-cols-[0.95fr_1.45fr]">
          <div className="bg-[#1f1b16] px-6 py-8 text-[#fffaf0] sm:px-8">
            <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">
              Nuevo ejercicio
            </p>

            <h3 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.05em]">
              Disena un movimiento que encaje con tu rutina.
            </h3>

            <p className="mt-4 max-w-sm text-sm leading-6 text-[#efe4d2]">
              Anade nombre, grupo muscular, descripcion y tipo. Si quieres que
              sea un ejercicio propio, dejalo como no oficial.
            </p>

            <button
              className="mt-8 rounded-2xl border border-[#fffaf0]/20 px-4 py-2 text-sm font-bold text-[#fffaf0] transition hover:bg-[#fffaf0] hover:text-[#1f1b16] disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              onClick={handleClose}
              disabled={isSubmitting}
            >
              Cerrar
            </button>
          </div>

          <form
            className="grid gap-5 px-6 py-8 sm:px-8"
            onSubmit={handleSubmit}
            data-block="create-exercise-form"
          >
            <div className="grid gap-5 sm:grid-cols-2">
              <label className="grid gap-2 text-sm font-semibold text-[#3b3128]">
                <span>Nombre</span>
                <input
                  id={nameId}
                  type="text"
                  value={form.name}
                  onChange={handleChange("name")}
                  className="rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none transition focus:border-[#ea7130]"
                  placeholder="Ej. Bench Press"
                />
              </label>

              <label className="grid gap-2 text-sm font-semibold text-[#3b3128]">
                <span>Grupo muscular</span>
                <input
                  id={muscleGroupId}
                  type="text"
                  value={form.muscle_group}
                  onChange={handleChange("muscle_group")}
                  className="rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none transition focus:border-[#ea7130]"
                  placeholder="Ej. chest"
                />
              </label>
            </div>

            <label className="grid gap-2 text-sm font-semibold text-[#3b3128]">
              <span>Descripcion</span>
              <textarea
                id={descriptionId}
                value={form.description}
                onChange={handleChange("description")}
                className="min-h-32 rounded-[1.5rem] border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none transition focus:border-[#ea7130]"
                placeholder="Describe como se ejecuta o para que sirve."
              />
            </label>

            <div className="grid gap-5 sm:grid-cols-2">
              <label className="grid gap-2 text-sm font-semibold text-[#3b3128]">
                <span>Musculo secundario</span>
                <input
                  id={secondaryMuscleGroupId}
                  type="text"
                  value={form.secondary_muscle_group}
                  onChange={handleChange("secondary_muscle_group")}
                  className="rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none transition focus:border-[#ea7130]"
                  placeholder="Ej. triceps"
                />
              </label>

              <label className="grid gap-2 text-sm font-semibold text-[#3b3128]">
                <span>Tipo de ejercicio</span>
                <input
                  id={exerciseTypeId}
                  type="text"
                  value={form.exercise_type}
                  onChange={handleChange("exercise_type")}
                  className="rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none transition focus:border-[#ea7130]"
                  placeholder="Ej. strength"
                />
              </label>
            </div>

            <label className="flex items-center gap-3 rounded-2xl border border-[#1f1b16]/10 bg-[#f6efdf] px-4 py-3 text-sm font-semibold text-[#3b3128]">
              <input
                type="checkbox"
                checked={form.is_official}
                onChange={handleChange("is_official")}
                className="h-4 w-4 accent-[#ea7130]"
              />
              Marcar como ejercicio oficial
            </label>

            {feedbackMessage && (
              <p className="rounded-2xl border border-[#9f2f22]/15 bg-[#fff0ec] px-4 py-3 text-sm font-semibold text-[#9f2f22]">
                {feedbackMessage}
              </p>
            )}

            <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <button
                className="rounded-2xl border border-[#1f1b16]/15 px-5 py-3 text-sm font-bold text-[#1f1b16] transition hover:bg-[#1f1b16] hover:text-[#fffaf0] disabled:cursor-not-allowed disabled:opacity-60"
                type="button"
                onClick={handleClose}
                disabled={isSubmitting}
              >
                Cancelar
              </button>

              <button
                className="rounded-2xl bg-[#ea7130] px-5 py-3 text-sm font-black text-[#1f1b16] transition hover:bg-[#1f1b16] hover:text-[#fffaf0] disabled:cursor-not-allowed disabled:opacity-60"
                type="submit"
                disabled={isSubmitting}
              >
                {isSubmitting ? "Creando..." : "Crear ejercicio"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
