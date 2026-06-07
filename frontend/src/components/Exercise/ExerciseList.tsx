import type { Exercise, ExerciseStatus } from "../../types/exercise";
import ExerciseCard from "./ExerciseCard";

type ExerciseListProps = {
    status: ExerciseStatus;
    message: string;
    exercises: Exercise[];
    selectedExerciseId?: string;
    onSelectExercise?: (exercise: Exercise) => void;
};

export default function ExerciseList({
    status,
    message,
    exercises,
    selectedExerciseId,
    onSelectExercise,
}: ExerciseListProps) {
    const hasCustomExercises = exercises.some(
        (exercise) => exercise.is_official === false,
    );

    return (
        <div
            className="mt-4 rounded-3xl border border-dashed border-[#1f1b16]/20 bg-white/45 p-2 md:h-full md:min-h-[24rem]"
            data-block="exercise-list-container"
        >
          <div className="max-h-[32rem] overflow-y-auto p-2 [scrollbar-gutter:stable] sm:p-3 md:h-full md:max-h-none">
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
                <>
                    {hasCustomExercises && (
                        <div className="sticky top-0 z-10 mb-4 flex items-center gap-2 rounded-xl bg-[#fffaf0]/95 px-3 py-2 text-xs font-semibold text-[#5d5348] backdrop-blur-sm">
                            <span
                                className="h-2.5 w-2.5 rounded-full bg-[#ea7130] ring-4 ring-[#265c52]/10"
                                aria-hidden="true"
                            />
                            Ejercicio propio
                        </div>
                    )}

                    <ul
                        className="grid gap-3 sm:gap-5 md:grid-cols-2 xl:grid-cols-2"
                        data-block="exercise-list"
                    >
                        {exercises.map((exercise) => (
                            <ExerciseCard
                                key={exercise.id}
                                exercise={exercise}
                                isSelected={exercise.id === selectedExerciseId}
                                onSelect={onSelectExercise}
                            />
                        ))}
                    </ul>
                </>
            )}
          </div>
        </div>
    );
}
