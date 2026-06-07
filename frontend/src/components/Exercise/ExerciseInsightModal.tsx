import type { Exercise, ExerciseInsights, ExerciseStatus } from "../../types/exercise";
import ExerciseInsightPanel from "./ExerciseInsightPanel";
import { muscleGroupLabel } from "../../lib/exerciseLabels";

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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[#1f1b16]/55 px-3 py-4 backdrop-blur-sm sm:px-6 sm:py-5">
      <div className="relative flex h-[calc(100dvh-2rem)] w-full max-w-4xl flex-col overflow-hidden rounded-[24px] border-2 border-[#fffaf0]/20 bg-[#fffaf0] bg-clip-padding shadow-[0_40px_120px_rgba(31,27,22,0.32)] sm:h-[calc(100vh-10rem)]">
        <div className="grid min-h-0 flex-1 gap-0 overflow-y-auto lg:grid-cols-[0.72fr_1.6fr] lg:overflow-hidden">
          <aside className="bg-[#1f1b16] px-5 py-6 text-[#fffaf0] sm:px-8 sm:py-7">
            <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">
              Rendimiento
            </p>

            <h3 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.05em]">
              {exercise.name}
            </h3>

            <p className="mt-3 text-sm font-semibold text-[#efe4d2]">
              {muscleGroupLabel(exercise.muscle_group)}
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

          <div className="h-full overflow-y-auto px-5 py-6 [scrollbar-gutter:stable] sm:px-7">
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
