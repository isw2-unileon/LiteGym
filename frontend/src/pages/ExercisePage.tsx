import * as React from "react";

type Exercise = {
    id: number;
    name: string;
    description: string | null;
    muscle_group: string;
    exercise_type: string | null;
};

type ExerciseStatus = "idle" | "loading" | "success" | "error";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

export default function ExercisePage() {
    const [exercises, setExercises] = React.useState<Exercise[]>([]);
    const [status, setStatus] = React.useState<ExerciseStatus>("idle");
    const [message, setMessage] = React.useState("");

    React.useEffect(() => {
        const fetchExercise = async () => {
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

        fetchExercise();
    }, []);

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
                    <div className="mb-8 max-w-2xl" data-block="page-header">
                        <p
                            className="mb-5 inline-flex rounded-full border border-[#1f1b16]/15 bg-white/35 px-4 py-2 text-sm font-semibold uppercase tracking-[0.28em] text-[#265c52] backdrop-blur"
                            data-block="brand-badge"
                        >
                            Grupo 16 Fitness
                        </p>

                        <h1
                            className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-5xl font-black leading-[0.95] tracking-[-0.06em] text-[#1f1b16] sm:text-7xl"
                            data-block="page-title"
                        >
                            Explora tus ejercicios.
                        </h1>

                        <p
                            className="mt-4 max-w-xl text-base text-[#5d5348] sm:text-lg"
                            data-block="page-description"
                        >
                            Aqui puedes ver los ejercicios disponibles para
                            usarlos al crear tus rutinas.
                        </p>
                    </div>

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
                                <p
                                    className="text-sm text-[#6b5d4d]"
                                    data-block="empty-state"
                                >
                                    No hay ejercicios disponibles.
                                </p>
                            )}

                            {status === "success" && exercises.length > 0 && (
                                <ul
                                    className="grid gap-5 sm:grid-cols-2 xl:grid-cols-3"
                                    data-block="exercise-list"
                                >
                                    {exercises.map((exercise) => (
                                        <li
                                            key={exercise.id}
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
                                                    Tipo:{" "}
                                                    {exercise.exercise_type}
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
                                    ))}
                                </ul>
                            )}
                        </div>
                    </div>
                </div>
            </section>
        </main>
    );
}
