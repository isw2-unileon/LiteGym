import type { Exercise } from "../../types/exercise";
import { exerciseTypeLabel, muscleGroupLabel } from "../../lib/exerciseLabels";

type ExerciseCardProps = {
    exercise: Exercise;
    isSelected?: boolean;
    onSelect?: (exercise: Exercise) => void;
};

export default function ExerciseCard({
    exercise,
    isSelected = false,
    onSelect,
}: ExerciseCardProps) {
    return (
        <li
            className={`relative rounded-2xl border p-4 transition ${
                isSelected
                    ? "border-[#ea7130] bg-[#fff3df] shadow-[0_18px_45px_rgba(234,113,48,0.18)]"
                    : "border-[#1f1b16]/10 bg-[#fffaf0] hover:border-[#ea7130]/35 hover:bg-[#fff7ea]"
            }`}
            data-block="exercise-card"
        >
            {exercise.is_official === false && (
                <span
                    aria-label="Ejercicio propio"
                    className="absolute right-3 top-3 h-2.5 w-2.5 rounded-full bg-[#ea7130] ring-4 ring-[#265c52]/10"
                    role="img"
                    title="Ejercicio propio"
                />
            )}

            <button
                type="button"
                className="w-full text-left"
                onClick={() => onSelect?.(exercise)}
            >
                <div>
                    <div className="min-w-0">
                        <h3
                            className="break-words pr-5 text-lg font-black text-[#1f1b16]"
                            data-block="exercise-name"
                        >
                            {exercise.name}
                        </h3>

                        <p
                            className="mt-1 text-sm font-semibold text-[#265c52]"
                            data-block="exercise-muscle-group"
                        >
                            {muscleGroupLabel(exercise.muscle_group)}
                        </p>
                    </div>
                </div>

                {exercise.exercise_type && (
                    <p
                        className="mt-3 text-sm text-[#5d5348]"
                        data-block="exercise-type"
                    >
                        Tipo: {exerciseTypeLabel(exercise.exercise_type)}
                    </p>
                )}

                {exercise.description && (
                    <p
                        className="mt-2 line-clamp-3 text-sm text-[#5d5348]"
                        data-block="exercise-description"
                    >
                        {exercise.description}
                    </p>
                )}
            </button>
        </li>
    );
}
