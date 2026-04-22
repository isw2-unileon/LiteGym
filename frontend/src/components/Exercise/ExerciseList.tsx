import type { Exercise, ExerciseStatus } from "../../types/exercise";
import ExerciseCard from "./ExerciseCard";

type ExerciseListProps = {
    status: ExerciseStatus;
    message: string;
    exercises: Exercise[];
};

export default function ExerciseList({
    status,
    message,
    exercises,
}: ExerciseListProps) {
    return (
        <div
            className="mt-7 max-h-[32rem] overflow-y-auto rounded-3xl border border-dashed border-[#1f1b16]/20 bg-white/45 p-5"
            data-block="exercise-list-container"
        >
            {status === "loading" && (
                <p
                    className="text-sm text-[#6b5d4d]"
                    data-block="loading-state"
                >
                    Cargando ejercicios...
                </p>
            )}

            {status === "error" && (
                <p
                    className="text-sm font-bold text-[#9f2f22]"
                    data-block="error-state"
                >
                    {message}
                </p>
            )}

            {status === "success" && exercises.length === 0 && (
                <p className="text-sm text-[#6b5d4d]" data-block="empty-state">
                    No se encontraron ejercicios con esos filtros.
                </p>
            )}

            {status === "success" && exercises.length > 0 && (
                <ul
                    className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3"
                    data-block="exercise-list"
                >
                    {exercises.map((exercise) => (
                        <ExerciseCard key={exercise.id} exercise={exercise} />
                    ))}
                </ul>
            )}
        </div>
    );
}
