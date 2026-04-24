import type { Exercise } from "../../types/exercise";

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
            className={`rounded-2xl border p-4 transition ${
                isSelected
                    ? "border-[#ea7130] bg-[#fff3df] shadow-[0_18px_45px_rgba(234,113,48,0.18)]"
                    : "border-[#1f1b16]/10 bg-[#fffaf0] hover:border-[#ea7130]/35 hover:bg-[#fff7ea]"
            }`}
            data-block="exercise-card"
        >
            <button
                type="button"
                className="w-full text-left"
                onClick={() => onSelect?.(exercise)}
            >
                <div className="flex items-start justify-between gap-3">
                    <div>
                        <h3
                            className="text-lg font-black text-[#1f1b16]"
                            data-block="exercise-name"
                        >
                            {exercise.name}
                        </h3>

                        <p
                            className="mt-1 text-sm font-semibold text-[#265c52]"
                            data-block="exercise-muscle-group"
                        >
                            {exercise.muscle_group}
                        </p>
                    </div>

                    {exercise.is_official === false && (
                        <span className="rounded-full border border-[#1f1b16]/10 bg-[#265c52]/10 px-2.5 py-1 text-[11px] font-bold uppercase tracking-[0.18em] text-[#265c52]">
                            Propio
                        </span>
                    )}
                </div>

                {exercise.exercise_type && (
                    <p
                        className="mt-3 text-sm text-[#5d5348]"
                        data-block="exercise-type"
                    >
                        Tipo: {exercise.exercise_type}
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
