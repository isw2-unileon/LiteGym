import type { Exercise } from "../../types/exercise";

type ExerciseCardProps = {
    exercise: Exercise;
};

export default function ExerciseCard({ exercise }: ExerciseCardProps) {
    return (
        <li
            className="rounded-2xl border border-[#1f1b16]/10 bg-[#fffaf0] p-4"
            data-block="exercise-card"
        >
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

            {exercise.exercise_type && (
                <p
                    className="mt-1 text-sm text-[#5d5348]"
                    data-block="exercise-type"
                >
                    Tipo: {exercise.exercise_type}
                </p>
            )}

            {exercise.description && (
                <p
                    className="mt-2 text-sm text-[#5d5348]"
                    data-block="exercise-description"
                >
                    {exercise.description}
                </p>
            )}
        </li>
    );
}
