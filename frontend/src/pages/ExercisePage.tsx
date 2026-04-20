import * as React from "react";
import ExerciseFilters from "../components/Exercise/ExerciseFilters";
import ExerciseHeader from "../components/Exercise/ExerciseHeader";
import ExerciseList from "../components/Exercise/ExerciseList";
import type { Exercise, ExerciseStatus } from "../types/exercise";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

export default function ExercisePage() {
    const [exercises, setExercises] = React.useState<Exercise[]>([]);
    const [status, setStatus] = React.useState<ExerciseStatus>("idle");
    const [message, setMessage] = React.useState("");

    const [search, setSearch] = React.useState("");
    const [typeFilter, setTypeFilter] = React.useState("");
    const [muscleFilter, setMuscleFilter] = React.useState("");

    React.useEffect(() => {
        const fetchExercises = async () => {
            setStatus("loading");
            setMessage("");

            try {
                const response = await fetch(`${API_BASE_URL}/api/exercises`, {
                    credentials: "include",
                });

                if (response.status === 401) {
                    setMessage("No autorizado. Por favor, inicia sesión.");
                    setStatus("error");
                    return;
                }

                if (!response.ok) {
                    setStatus("error");
                    setMessage("No se pudieron cargar los ejercicios.");
                    return;
                }

                const data = await response.json();
                setExercises(data);
                setStatus("success");
            } catch (error) {
                console.error(error);
                setMessage(
                    "Error al cargar los ejercicios. Por favor, intenta de nuevo.",
                );
                setStatus("error");
            }
        };

        fetchExercises();
    }, []);

    const exerciseTypes = React.useMemo(() => {
        return Array.from(
            new Set(
                exercises
                    .map((exercise) => exercise.exercise_type)
                    .filter((type): type is string => Boolean(type)),
            ),
        ).sort();
    }, [exercises]);

    const muscleGroups = React.useMemo(() => {
        return Array.from(
            new Set(exercises.map((exercise) => exercise.muscle_group)),
        ).sort();
    }, [exercises]);

    const filteredExercises = React.useMemo(() => {
        const normalizedSearch = search.toLowerCase().trim();

        return exercises.filter((exercise) => {
            const matchesSearch =
                normalizedSearch === "" ||
                [
                    exercise.name,
                    exercise.muscle_group,
                    exercise.exercise_type ?? "",
                ].some((value) =>
                    value.toLowerCase().includes(normalizedSearch),
                );

            const matchesType =
                typeFilter === "" || exercise.exercise_type === typeFilter;

            const matchesMuscle =
                muscleFilter === "" || exercise.muscle_group === muscleFilter;

            return matchesSearch && matchesType && matchesMuscle;
        });
    }, [exercises, search, typeFilter, muscleFilter]);

    return (
        <main
            className="min-h-screen overflow-hidden bg-[#f4efe2] text-[#1f1b16]"
            data-section="exercise-page"
        >
            <section
                className="relative isolate min-h-screen px-6 py-8 sm:px-10 lg:px-16"
                data-block="page-layout"
            >
                <div
                    className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_left,_rgba(234,113,48,0.30),_transparent_34%),radial-gradient(circle_at_bottom_right,_rgba(38,92,82,0.35),_transparent_32%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_44%,_#d8e1d0_100%)]"
                    data-block="background-gradient"
                />
                <div
                    className="absolute left-8 top-10 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/25 blur-sm"
                    data-block="background-circle"
                />
                <div
                    className="absolute bottom-8 right-10 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10"
                    data-block="background-shape"
                />

                <div className="mx-auto max-w-6xl" data-block="page-container">
                    <ExerciseHeader />

                    <div
                        className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-5 shadow-[0_30px_80px_rgba(47,39,27,0.20)] backdrop-blur-md sm:p-8"
                        data-block="exercises-panel"
                    >
                        <div
                            className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-6 text-[#fffaf0] shadow-inner"
                            data-block="panel-header"
                        >
                            <p
                                className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]"
                                data-block="panel-label"
                            >
                                Ejercicios
                            </p>

                            <h2
                                className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]"
                                data-block="panel-title"
                            >
                                Lista disponible
                            </h2>
                        </div>

                        {status === "success" && (
                            <ExerciseFilters
                                search={search}
                                typeFilter={typeFilter}
                                muscleFilter={muscleFilter}
                                exerciseTypes={exerciseTypes}
                                muscleGroups={muscleGroups}
                                onSearchChange={setSearch}
                                onTypeFilterChange={setTypeFilter}
                                onMuscleFilterChange={setMuscleFilter}
                            />
                        )}

                        <ExerciseList
                            status={status}
                            message={message}
                            exercises={filteredExercises}
                        />
                    </div>
                </div>
            </section>
        </main>
    );
}
