import type { Exercise, ExerciseInsights, ExerciseStatus } from "../../types/exercise";
import ExerciseInsightPanel from "./ExerciseInsightPanel";

type ExerciseInsightModalProps = {
  isOpen: boolean;
  exercise: Exercise | null;
  insights: ExerciseInsights | null;
  status: ExerciseStatus;
  message: string;
  onClose: () => void;
};

export default function ExerciseInsightModal({
  isOpen,
  exercise,
  insights,
  status,
  message,
  onClose,
}: ExerciseInsightModalProps) {
  if (!isOpen || !exercise) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#1f1b16]/55 px-3 py-5 backdrop-blur-sm sm:px-6">
      <div className="relative flex max-h-[92vh] w-full max-w-6xl flex-col overflow-hidden rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0] shadow-[0_40px_120px_rgba(31,27,22,0.32)]">
        <div className="grid gap-0 lg:grid-cols-[0.72fr_1.6fr]">
          <aside className="bg-[#1f1b16] px-6 py-7 text-[#fffaf0] sm:px-8">
            <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">
              Rendimiento
            </p>

            <h3 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.05em]">
              {exercise.name}
            </h3>

            <p className="mt-3 text-sm font-semibold text-[#efe4d2]">
              {exercise.muscle_group}
            </p>

            <p className="mt-5 max-w-sm text-sm leading-6 text-[#efe4d2]">
              Estadisticas calculadas desde las sesiones registradas donde
              aparece este ejercicio.
            </p>

            <button
              className="mt-8 rounded-2xl border border-[#fffaf0]/20 px-4 py-2 text-sm font-bold text-[#fffaf0] transition hover:bg-[#fffaf0] hover:text-[#1f1b16]"
              type="button"
              onClick={onClose}
            >
              Cerrar
            </button>
          </aside>

          <div className="max-h-[92vh] overflow-y-auto px-5 py-6 sm:px-7">
            <ExerciseInsightPanel
              insights={insights}
              status={status}
              message={message}
              variant="embedded"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
