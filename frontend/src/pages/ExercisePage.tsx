import * as React from "react";
import CreateExerciseModal, {
    type CreateExercisePayload,
} from "../components/Exercise/CreateExerciseModal";
import { Link } from "react-router-dom";
import ExerciseFilters from "../components/Exercise/ExerciseFilters";
import ExerciseHeader from "../components/Exercise/ExerciseHeader";
import ExerciseList from "../components/Exercise/ExerciseList";
import { apiUrl } from "../lib/api";
import type {
    Exercise,
    ExerciseMetadataResponse,
    ExerciseStatus,
    SelectOption,
} from "../types/exercise";

type CurrentUser = {
    id: string;
    email: string;
    username: string;
    role: string;
};

type ExerciseListResponse = {
    items: Exercise[];
    page: number;
    limit: number;
    total: number;
    total_pages: number;
};

export default function ExercisePage() {
    const [exercises, setExercises] = React.useState<Exercise[]>([]);
    const [currentUser, setCurrentUser] = React.useState<CurrentUser | null>(
        null,
    );
    const [status, setStatus] = React.useState<ExerciseStatus>("idle");
    const [message, setMessage] = React.useState("");
    const [isCreateModalOpen, setIsCreateModalOpen] = React.useState(false);
    const [createErrorMessage, setCreateErrorMessage] = React.useState("");
    const [isCreatingExercise, setIsCreatingExercise] = React.useState(false);
    const [isEditModalOpen, setIsEditModalOpen] = React.useState(false);
    const [editErrorMessage, setEditErrorMessage] = React.useState("");
    const [isUpdatingExercise, setIsUpdatingExercise] = React.useState(false);
    const [selectedExerciseId, setSelectedExerciseId] = React.useState("");

    const [search, setSearch] = React.useState("");
    const [typeFilter, setTypeFilter] = React.useState("");
    const [muscleFilter, setMuscleFilter] = React.useState("");
    const [exerciseTypeOptions, setExerciseTypeOptions] = React.useState<
        SelectOption[]
    >([]);
    const [muscleGroupOptions, setMuscleGroupOptions] = React.useState<
        SelectOption[]
    >([]);

    const [page, setPage] = React.useState(1);
    const [limit] = React.useState(20);
    const [total, setTotal] = React.useState(0);
    const [totalPages, setTotalPages] = React.useState(0);

    React.useEffect(() => {
        const fetchCurrentUser = async () => {
            try {
                const response = await fetch(apiUrl("/api/auth/me"), {
                    credentials: "include",
                });

                if (!response.ok) {
                    return;
                }

                const data = (await response.json()) as { user: CurrentUser };
                setCurrentUser(data.user);
            } catch (error) {
                console.error(error);
            }
        };

        void fetchCurrentUser();
    }, []);

    React.useEffect(() => {
        const fetchExerciseMetadata = async () => {
            try {
                const response = await fetch(apiUrl("/api/exercises/metadata"), {
                    credentials: "include",
                });

                if (!response.ok) {
                    return;
                }

                const data =
                    (await response.json()) as ExerciseMetadataResponse;
                setExerciseTypeOptions(data.exercise_types);
                setMuscleGroupOptions(data.muscle_groups);
            } catch (error) {
                console.error(error);
            }
        };

        void fetchExerciseMetadata();
    }, []);

    const fetchExercises = React.useCallback(async () => {
        setStatus("loading");
        setMessage("");

        try {
            const params = new URLSearchParams();

            if (search.trim() !== "") {
                params.set("search", search.trim());
            }

            if (typeFilter !== "") {
                params.set("type", typeFilter);
            }

            if (muscleFilter !== "") {
                params.set("muscle", muscleFilter);
            }

            params.set("page", String(page));
            params.set("limit", String(limit));

            const response = await fetch(
                apiUrl(`/api/exercises?${params.toString()}`),
                {
                    credentials: "include",
                },
            );

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

            const data = (await response.json()) as ExerciseListResponse;

            setExercises(data.items);
            setTotal(data.total);
            setTotalPages(data.total_pages);

            setSelectedExerciseId((current) => {
                const stillExists = data.items.some(
                    (exercise) => exercise.id === current,
                );

                if (stillExists) {
                    return current;
                }

                const firstExercise = data.items[0];

                return firstExercise ? firstExercise.id : "";
            });

            setStatus("success");
        } catch (error) {
            console.error(error);
            setMessage(
                "Error al cargar los ejercicios. Por favor, intenta de nuevo.",
            );
            setStatus("error");
        }
    }, [search, typeFilter, muscleFilter, page, limit]);

    React.useEffect(() => {
        void fetchExercises();
    }, [fetchExercises]);

    const handleSearchChange = (value: string) => {
        setSearch(value);
        setPage(1);
    };

    const handleTypeFilterChange = (value: string) => {
        setTypeFilter(value);
        setPage(1);
    };

    const handleMuscleFilterChange = (value: string) => {
        setMuscleFilter(value);
        setPage(1);
    };

    const buildExercisePayload = React.useCallback(
        (exercise: Exercise): CreateExercisePayload => {
            const secondaryMuscleGroups =
                exercise.secondary_muscle_groups?.filter(Boolean) ??
                (exercise.secondary_muscle_group
                    ? exercise.secondary_muscle_group
                          .split(",")
                          .map((value) => value.trim())
                          .filter(Boolean)
                    : []);

            return {
                name: exercise.name,
                description: exercise.description ?? "",
                muscle_group: exercise.muscle_group,
                secondary_muscle_groups:
                    secondaryMuscleGroups.length > 0
                        ? secondaryMuscleGroups
                        : [""],
                exercise_type: exercise.exercise_type ?? "",
                is_official: exercise.is_official === true,
            };
        },
        [],
    );

    const filteredExercises = exercises;

    const selectedExercise = React.useMemo(() => {
        const selected = filteredExercises.find(
            (exercise) => exercise.id === selectedExerciseId,
        );

        if (selected) {
            return selected;
        }

        const firstExercise = filteredExercises[0];

        return firstExercise ?? null;
    }, [filteredExercises, selectedExerciseId]);

    const handleCreateExercise = async (payload: CreateExercisePayload) => {
        setIsCreatingExercise(true);
        setCreateErrorMessage("");

        try {
            const response = await fetch(apiUrl("/api/exercises"), {
                method: "POST",
                credentials: "include",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(payload),
            });

            if (response.status === 401) {
                setCreateErrorMessage(
                    "Tu sesion ha expirado. Inicia sesion otra vez.",
                );
                return;
            }

            if (!response.ok) {
                const errorBody = (await response.json().catch(() => null)) as {
                    error?: string;
                } | null;

                setCreateErrorMessage(
                    errorBody?.error ?? "No se pudo crear el ejercicio.",
                );
                return;
            }

            const createdExercise = (await response.json()) as Exercise;

            setExercises((current) => [createdExercise, ...current]);
            setSelectedExerciseId(createdExercise.id);
            setTotal((current) => current + 1);
            setStatus("success");
            setMessage("");
            setIsCreateModalOpen(false);
        } catch (error) {
            console.error(error);
            setCreateErrorMessage(
                "No se pudo crear el ejercicio. Revisa la conexion con el backend.",
            );
        } finally {
            setIsCreatingExercise(false);
        }
    };

    const handleUpdateExercise = async (payload: CreateExercisePayload) => {
        if (!selectedExercise) {
            return;
        }

        setIsUpdatingExercise(true);
        setEditErrorMessage("");

        try {
            const response = await fetch(
                apiUrl(`/api/exercises/${selectedExercise.id}`),
                {
                    method: "PUT",
                    credentials: "include",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(payload),
                },
            );

            if (response.status === 401) {
                setEditErrorMessage(
                    "Tu sesion ha expirado. Inicia sesion otra vez.",
                );
                return;
            }

            if (!response.ok) {
                const errorBody = (await response.json().catch(() => null)) as {
                    error?: string;
                } | null;

                setEditErrorMessage(
                    errorBody?.error ?? "No se pudo actualizar el ejercicio.",
                );
                return;
            }

            const updatedExercise = (await response.json()) as Exercise;

            setExercises((current) =>
                current.map((exercise) =>
                    exercise.id === updatedExercise.id
                        ? updatedExercise
                        : exercise,
                ),
            );
            setSelectedExerciseId(updatedExercise.id);
            setIsEditModalOpen(false);
        } catch (error) {
            console.error(error);
            setEditErrorMessage(
                "No se pudo actualizar el ejercicio. Revisa la conexion con el backend.",
            );
        } finally {
            setIsUpdatingExercise(false);
        }
    };

    const officialCount = exercises.filter(
        (exercise) => exercise.is_official !== false,
    ).length;
    const customCount = exercises.length - officialCount;
    const canCreateOfficial = currentUser?.role === "admin";
    const canEditSelectedExercise =
        selectedExercise != null &&
        (selectedExercise.is_official === false ||
            currentUser?.role === "admin");
    const editInitialPayload = selectedExercise
        ? buildExercisePayload(selectedExercise)
        : undefined;

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

                <div
                    className="mx-auto max-w-[min(1600px,96vw)]"
                    data-block="page-container"
                >
                    <div className="mb-6 flex justify-end">
                        <Link
                            to="/profile"
                            className="inline-flex items-center gap-2 rounded-full border border-[#1f1b16]/15 bg-white/35 px-5 py-2 text-sm font-bold text-[#265c52] shadow-sm backdrop-blur transition hover:bg-white/50 hover:text-[#1f1b16]"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                className="h-4 w-4"
                                viewBox="0 0 20 20"
                                fill="currentColor"
                            >
                                <path
                                    fillRule="evenodd"
                                    d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                                    clipRule="evenodd"
                                />
                            </svg>
                            Mi perfil
                        </Link>
                    </div>

                    <ExerciseHeader />

                    <div
                        className="grid gap-6 xl:grid-cols-[300px_minmax(0,1fr)_320px]"
                        data-block="exercise-shell"
                    >
                        <aside
                            className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-5 shadow-[0_30px_80px_rgba(47,39,27,0.14)] backdrop-blur-md"
                            data-block="exercise-sidebar"
                        >
                            <div className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-6 text-[#fffaf0] shadow-inner">
                                <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">
                                    Control
                                </p>
                                <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]">
                                    Explora o crea
                                </h2>
                                <p className="mt-3 text-sm leading-6 text-[#efe4d2]">
                                    Filtra tu biblioteca y abre nuevos
                                    movimientos desde un solo panel.
                                </p>
                            </div>

                            <div className="mt-6 grid gap-4">
                                <button
                                    type="button"
                                    className="rounded-2xl bg-[#ea7130] px-5 py-4 text-sm font-black text-[#1f1b16] transition hover:bg-[#1f1b16] hover:text-[#fffaf0]"
                                    onClick={() => {
                                        setCreateErrorMessage("");
                                        setIsCreateModalOpen(true);
                                    }}
                                >
                                    Crear nuevo ejercicio
                                </button>

                                {status === "success" && (
                                    <div className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/65 p-4">
                                        <p className="mb-4 text-xs font-black uppercase tracking-[0.2em] text-[#265c52]">
                                            Buscar y filtrar
                                        </p>
                                        <ExerciseFilters
                                            search={search}
                                            typeFilter={typeFilter}
                                            muscleFilter={muscleFilter}
                                            exerciseTypes={exerciseTypeOptions}
                                            muscleGroups={muscleGroupOptions}
                                            onSearchChange={handleSearchChange}
                                            onTypeFilterChange={
                                                handleTypeFilterChange
                                            }
                                            onMuscleFilterChange={
                                                handleMuscleFilterChange
                                            }
                                        />
                                    </div>
                                )}

                                <div className="grid gap-3 rounded-[1.5rem] border border-dashed border-[#1f1b16]/20 bg-white/60 p-4 text-center">
                                    <p className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]">
                                        Biblioteca
                                    </p>
                                    <div className="grid gap-3 sm:grid-cols-3 xl:grid-cols-1">
                                        <StatTile
                                            label="Resultados"
                                            value={String(total)}
                                        />
                                        <StatTile
                                            label="En página"
                                            value={String(exercises.length)}
                                        />
                                        <StatTile
                                            label="Oficiales página"
                                            value={String(officialCount)}
                                        />
                                        <StatTile
                                            label="Propios página"
                                            value={String(customCount)}
                                        />
                                    </div>
                                </div>
                            </div>
                        </aside>

                        <section
                            className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-5 shadow-[0_30px_80px_rgba(47,39,27,0.20)] backdrop-blur-md sm:p-8"
                            data-block="exercises-panel"
                        >
                            <div
                                className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-6 text-[#fffaf0] shadow-inner"
                                data-block="panel-header"
                            >
                                <div className="flex flex-col gap-5 md:flex-row md:items-end md:justify-between">
                                    <div>
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

                                    <p className="max-w-xs text-sm leading-6 text-[#efe4d2]">
                                        Selecciona una tarjeta para ver su ficha
                                        completa sin salir de la página.
                                    </p>
                                </div>
                            </div>

                            <div className="mt-6 flex items-center justify-between gap-4 rounded-[1.5rem] border border-dashed border-[#1f1b16]/15 bg-white/55 px-4 py-3">
                                <div>
                                    <p className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]">
                                        Resultado actual
                                    </p>
                                    <p className="mt-1 text-sm text-[#5d5348]">
                                        {total} ejercicios encontrados con los
                                        filtros aplicados.
                                    </p>
                                </div>

                                {selectedExercise && (
                                    <span className="rounded-full bg-[#1f1b16] px-3 py-2 text-xs font-bold uppercase tracking-[0.18em] text-[#fffaf0]">
                                        Seleccionado: {selectedExercise.name}
                                    </span>
                                )}
                            </div>

                            <div className="mt-6">
                                <ExerciseList
                                    status={status}
                                    message={message}
                                    exercises={filteredExercises}
                                    selectedExerciseId={selectedExercise?.id}
                                    onSelectExercise={(exercise) =>
                                        setSelectedExerciseId(exercise.id)
                                    }
                                />
                            </div>

                            {status === "success" && totalPages > 1 && (
                                <div className="mt-6 flex items-center justify-between rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/60 px-4 py-3">
                                    <button
                                        type="button"
                                        disabled={page <= 1}
                                        onClick={() =>
                                            setPage((current) =>
                                                Math.max(1, current - 1),
                                            )
                                        }
                                        className="rounded-xl border border-[#1f1b16]/15 px-4 py-2 text-sm font-bold disabled:cursor-not-allowed disabled:opacity-40"
                                    >
                                        Anterior
                                    </button>

                                    <p className="text-sm font-bold text-[#5d5348]">
                                        Página {page} de {totalPages}
                                    </p>

                                    <button
                                        type="button"
                                        disabled={page >= totalPages}
                                        onClick={() =>
                                            setPage((current) =>
                                                Math.min(
                                                    totalPages,
                                                    current + 1,
                                                ),
                                            )
                                        }
                                        className="rounded-xl border border-[#1f1b16]/15 px-4 py-2 text-sm font-bold disabled:cursor-not-allowed disabled:opacity-40"
                                    >
                                        Siguiente
                                    </button>
                                </div>
                            )}
                        </section>

                        <aside
                            className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-5 shadow-[0_30px_80px_rgba(47,39,27,0.14)] backdrop-blur-md"
                            data-block="exercise-detail-panel"
                        >
                            <div className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[linear-gradient(180deg,#265c52_0%,#173e37_100%)] p-6 text-[#fffaf0] shadow-inner">
                                <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f6c98d]">
                                    Panel Detalles
                                </p>
                                <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]">
                                    Ficha rápida
                                </h2>
                            </div>

                            {selectedExercise ? (
                                <div className="mt-6 grid gap-4">
                                    <div className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/70 p-5">
                                        <div className="flex items-start justify-between gap-3">
                                            <div>
                                                <p className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]">
                                                    Nombre
                                                </p>
                                                <h3 className="mt-3 text-2xl font-black tracking-[-0.04em] text-[#1f1b16]">
                                                    {selectedExercise.name}
                                                </h3>
                                                <p className="mt-2 text-sm font-semibold text-[#5d5348]">
                                                    {
                                                        selectedExercise.muscle_group
                                                    }
                                                </p>
                                            </div>

                                            {canEditSelectedExercise && (
                                                <button
                                                    type="button"
                                                    className="rounded-2xl border border-[#1f1b16]/15 px-4 py-3 text-xs font-black uppercase tracking-[0.12em] text-[#1f1b16] transition hover:bg-[#1f1b16] hover:text-[#fffaf0]"
                                                    onClick={() => {
                                                        setEditErrorMessage("");
                                                        setIsEditModalOpen(
                                                            true,
                                                        );
                                                    }}
                                                >
                                                    Editar
                                                </button>
                                            )}
                                        </div>
                                    </div>

                                    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
                                        <InfoTile
                                            label="Tipo"
                                            value={
                                                selectedExercise.exercise_type ??
                                                "Sin especificar"
                                            }
                                        />
                                        <InfoTile
                                            label="Musculo secundario"
                                            value={
                                                selectedExercise.secondary_muscle_group ||
                                                "No definido"
                                            }
                                        />
                                        <InfoTile
                                            label="Estado"
                                            value={
                                                selectedExercise.is_official ===
                                                false
                                                    ? "Ejercicio propio"
                                                    : "Ejercicio oficial"
                                            }
                                        />
                                    </div>

                                    <div className="rounded-[1.5rem] border border-dashed border-[#1f1b16]/15 bg-[#fff8ea] p-5">
                                        <p className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]">
                                            Descripcion
                                        </p>
                                        <p className="mt-3 text-sm leading-7 text-[#5d5348]">
                                            {selectedExercise.description ||
                                                "Todavia no hay una descripcion para este ejercicio. Puedes completarla cuando edites o crees uno nuevo."}
                                        </p>
                                    </div>
                                </div>
                            ) : (
                                <div className="mt-6 rounded-[1.5rem] border border-dashed border-[#1f1b16]/20 bg-white/60 p-5">
                                    <p className="text-sm font-bold text-[#5d5348]">
                                        Selecciona un ejercicio de la lista para
                                        ver su detalle completo aquí.
                                    </p>
                                </div>
                            )}
                        </aside>
                    </div>
                </div>
            </section>

            <CreateExerciseModal
                isOpen={isCreateModalOpen}
                isSubmitting={isCreatingExercise}
                errorMessage={createErrorMessage}
                canCreateOfficial={canCreateOfficial}
                mode="create"
                exerciseTypeOptions={exerciseTypeOptions}
                muscleGroupOptions={muscleGroupOptions}
                onClose={() => {
                    setCreateErrorMessage("");
                    setIsCreateModalOpen(false);
                }}
                onSubmit={handleCreateExercise}
            />

            <CreateExerciseModal
                isOpen={isEditModalOpen}
                isSubmitting={isUpdatingExercise}
                errorMessage={editErrorMessage}
                canCreateOfficial={canCreateOfficial}
                mode="edit"
                initialPayload={editInitialPayload}
                exerciseTypeOptions={exerciseTypeOptions}
                muscleGroupOptions={muscleGroupOptions}
                onClose={() => {
                    setEditErrorMessage("");
                    setIsEditModalOpen(false);
                }}
                onSubmit={handleUpdateExercise}
            />
        </main>
    );
}

function StatTile({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-2xl border border-[#1f1b16]/10 bg-[#fffaf0] p-4 text-center">
            <p className="text-[11px] font-black uppercase tracking-[0.18em] text-[#265c52]">
                {label}
            </p>
            <p className="mt-2 text-2xl font-black tracking-[-0.05em] text-[#1f1b16]">
                {value}
            </p>
        </div>
    );
}

function InfoTile({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/70 p-4">
            <p className="text-[11px] font-black uppercase tracking-[0.18em] text-[#265c52]">
                {label}
            </p>
            <p className="mt-2 text-sm font-semibold text-[#1f1b16]">{value}</p>
        </div>
    );
}
