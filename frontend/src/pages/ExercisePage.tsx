import * as React from "react";
import CreateExerciseModal, {
  type CreateExercisePayload,
} from "../components/Exercise/CreateExerciseModal";
import { Link } from "react-router-dom";
import ExerciseFilters from "../components/Exercise/ExerciseFilters";
import ExerciseHeader from "../components/Exercise/ExerciseHeader";
import ExerciseInsightModal from "../components/Exercise/ExerciseInsightModal";
import ExerciseList from "../components/Exercise/ExerciseList";
import { apiUrl } from "../lib/api";
import type {
  Exercise,
  ExerciseInsights,
  ExerciseMetadataResponse,
  ExerciseStatus,
  ExerciseWorkoutSessionSummary,
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

const workoutSessionDateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "numeric",
  month: "short",
  year: "numeric",
});

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
  const [exerciseWorkoutSessions, setExerciseWorkoutSessions] = React.useState<
    ExerciseWorkoutSessionSummary[]
  >([]);
  const [exerciseWorkoutSessionsStatus, setExerciseWorkoutSessionsStatus] =
    React.useState<ExerciseStatus>("idle");
  const [exerciseWorkoutSessionsMessage, setExerciseWorkoutSessionsMessage] =
    React.useState("");
  const [exerciseInsights, setExerciseInsights] =
    React.useState<ExerciseInsights | null>(null);
  const [exerciseInsightsStatus, setExerciseInsightsStatus] =
    React.useState<ExerciseStatus>("idle");
  const [exerciseInsightsMessage, setExerciseInsightsMessage] =
    React.useState("");
  const [isInsightsModalOpen, setIsInsightsModalOpen] = React.useState(false);

  const [search, setSearch] = React.useState("");
  const [typeFilter, setTypeFilter] = React.useState("");
  const [muscleFilter, setMuscleFilter] = React.useState("");
  const [exerciseTypeOptions, setExerciseTypeOptions] = React.useState<
    SelectOption[]
  >([]);
  const [muscleGroupOptions, setMuscleGroupOptions] = React.useState<
    SelectOption[]
  >([]);
  const [metadataMessage, setMetadataMessage] = React.useState("");

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
      setMetadataMessage("");

      try {
        const response = await fetch(apiUrl("/api/exercises/metadata"), {
          credentials: "include",
        });

        if (!response.ok) {
          setMetadataMessage(
            "No se pudieron cargar las opciones de ejercicios.",
          );
          return;
        }

        const data = (await response.json()) as ExerciseMetadataResponse;
        setExerciseTypeOptions(data.exercise_types);
        setMuscleGroupOptions(data.muscle_groups);
      } catch (error) {
        console.error(error);
        setMetadataMessage("No se pudieron cargar las opciones de ejercicios.");
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
        params.set("muscle_group", muscleFilter);
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
          secondaryMuscleGroups.length > 0 ? secondaryMuscleGroups : [""],
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

  React.useEffect(() => {
    if (!selectedExercise?.id) {
      setExerciseWorkoutSessions([]);
      setExerciseWorkoutSessionsStatus("idle");
      setExerciseWorkoutSessionsMessage("");
      setExerciseInsights(null);
      setExerciseInsightsStatus("idle");
      setExerciseInsightsMessage("");
      setIsInsightsModalOpen(false);
      return;
    }

    let ignore = false;

    const fetchExerciseWorkoutSessions = async () => {
      setExerciseWorkoutSessions([]);
      setExerciseWorkoutSessionsStatus("loading");
      setExerciseWorkoutSessionsMessage("");

      try {
        const response = await fetch(
          apiUrl(`/api/exercises/${selectedExercise.id}/workout-sessions`),
          {
            credentials: "include",
          },
        );

        if (ignore) {
          return;
        }

        if (response.status === 401) {
          setExerciseWorkoutSessions([]);
          setExerciseWorkoutSessionsStatus("error");
          setExerciseWorkoutSessionsMessage(
            "No autorizado. Por favor, inicia sesión.",
          );
          return;
        }

        if (!response.ok) {
          setExerciseWorkoutSessions([]);
          setExerciseWorkoutSessionsStatus("error");
          setExerciseWorkoutSessionsMessage(
            "No se pudieron cargar los entrenos de este ejercicio.",
          );
          return;
        }

        const data =
          (await response.json()) as ExerciseWorkoutSessionSummary[];
        setExerciseWorkoutSessions(data);
        setExerciseWorkoutSessionsStatus("success");
      } catch (error) {
        if (ignore) {
          return;
        }

        console.error(error);
        setExerciseWorkoutSessions([]);
        setExerciseWorkoutSessionsStatus("error");
        setExerciseWorkoutSessionsMessage(
          "No se pudieron cargar los entrenos de este ejercicio.",
        );
      }
    };

    void fetchExerciseWorkoutSessions();

    return () => {
      ignore = true;
    };
  }, [selectedExercise?.id]);

  React.useEffect(() => {
    if (!selectedExercise?.id) {
      return;
    }

    let ignore = false;

    const fetchExerciseInsights = async () => {
      setExerciseInsights(null);
      setExerciseInsightsStatus("loading");
      setExerciseInsightsMessage("");

      try {
        const response = await fetch(
          apiUrl(`/api/exercises/${selectedExercise.id}/insights`),
          {
            credentials: "include",
          },
        );

        if (ignore) {
          return;
        }

        if (response.status === 401) {
          setExerciseInsights(null);
          setExerciseInsightsStatus("error");
          setExerciseInsightsMessage(
            "No autorizado. Por favor, inicia sesión.",
          );
          return;
        }

        if (!response.ok) {
          setExerciseInsights(null);
          setExerciseInsightsStatus("error");
          setExerciseInsightsMessage(
            "No se pudieron cargar las estadisticas del ejercicio.",
          );
          return;
        }

        const data = (await response.json()) as ExerciseInsights;
        setExerciseInsights(data);
        setExerciseInsightsStatus("success");
      } catch (error) {
        if (ignore) {
          return;
        }

        console.error(error);
        setExerciseInsights(null);
        setExerciseInsightsStatus("error");
        setExerciseInsightsMessage(
          "No se pudieron cargar las estadisticas del ejercicio.",
        );
      }
    };

    void fetchExerciseInsights();

    return () => {
      ignore = true;
    };
  }, [selectedExercise?.id]);

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
        setCreateErrorMessage("Tu sesion ha expirado. Inicia sesion otra vez.");
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
        setEditErrorMessage("Tu sesion ha expirado. Inicia sesion otra vez.");
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
          exercise.id === updatedExercise.id ? updatedExercise : exercise,
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
    (selectedExercise.is_official === false || currentUser?.role === "admin");
  const editInitialPayload = selectedExercise
    ? buildExercisePayload(selectedExercise)
    : undefined;

  return (
    <main
      data-ui="exercise-page-main"
      className="min-h-screen w-full overflow-hidden bg-transparent text-[#1f1b16]"
      data-section="exercise-page"
    >
      <header
        data-ui="exercise-page-header"
        className="w-full pl-3 pr-3 pt-4 pb-4 sm:pl-4 sm:pr-4 lg:pl-6 lg:pr-6"
        data-block="page-header"
      >
        <div
          data-ui="exercise-page-header-grid"
          className="grid w-full grid-cols-[minmax(0,1fr)_auto] items-start gap-4"
        >
          <div
            data-ui="exercise-header-wrapper"
            className="min-w-0 pl-4 sm:pl-6 lg:pl-8"
          >
            <ExerciseHeader />
          </div>

          <Link
            data-ui="profile-link-button"
            to="/profile"
            className="inline-flex shrink-0 items-center gap-2 rounded-full border border-[#1f1b16]/15 bg-white/35 pl-5 pr-5 py-2 text-sm font-bold text-[#265c52] shadow-sm backdrop-blur transition hover:bg-white/50 hover:text-[#1f1b16]"
          >
            <svg
              data-ui="profile-link-icon"
              xmlns="http://www.w3.org/2000/svg"
              className="h-4 w-4"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                data-ui="profile-link-icon-path"
                fillRule="evenodd"
                d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                clipRule="evenodd"
              />
            </svg>
            Mi perfil
          </Link>
        </div>
      </header>

      <section
        data-ui="exercise-page-layout-section"
        className="w-full pl-3 pr-3 pb-4 sm:pl-4 sm:pr-4 sm:pb-6 lg:pl-6 lg:pr-6"
        data-block="page-layout"
      >
        <div
          data-ui="exercise-page-container"
          className="mx-auto w-[96%] max-w-none"
          data-block="page-container"
        >
          <div
            data-ui="exercise-three-column-grid"
            className="grid w-full gap-5 xl:grid-cols-[minmax(0,1.55fr)_minmax(20rem,0.75fr)]"
            data-block="exercise-shell"
          >
            <aside
              data-ui="exercise-left-sidebar"
              className="min-w-0 rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-5 shadow-[0_2rem_5rem_rgba(47,39,27,0.14)] backdrop-blur-md xl:col-span-2"
              data-block="exercise-sidebar"
            >
              <div
                data-ui="exercise-left-sidebar-content"
                className="grid gap-4 xl:grid-cols-[minmax(16rem,0.8fr)_minmax(0,1.4fr)_minmax(16rem,0.8fr)] xl:items-stretch"
              >
                <div
                  data-ui="exercise-left-sidebar-header-card"
                  className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-6 text-[#fffaf0] shadow-inner"
                >
                  <p
                    data-ui="exercise-left-sidebar-label"
                    className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]"
                  >
                    Control
                  </p>

                  <h2
                    data-ui="exercise-left-sidebar-title"
                    className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]"
                  >
                    Explora o crea
                  </h2>

                  <p
                    data-ui="exercise-left-sidebar-description"
                    className="mt-3 text-sm leading-6 text-[#efe4d2]"
                  >
                    Filtra tu biblioteca y abre nuevos movimientos desde un
                    solo panel.
                  </p>

                  <button
                    data-ui="create-exercise-button"
                    type="button"
                    className="mt-5 w-full rounded-2xl bg-[#ea7130] pl-5 pr-5 py-4 text-sm font-black text-[#1f1b16] transition hover:bg-[#fffaf0] hover:text-[#1f1b16]"
                    onClick={() => {
                      setCreateErrorMessage("");
                      setIsCreateModalOpen(true);
                    }}
                  >
                    Crear nuevo ejercicio
                  </button>
                </div>

                {status === "success" && (
                  <div
                  data-ui="exercise-filters-card"
                    className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/65 p-4"
                  >
                    <p
                      data-ui="exercise-filters-title"
                      className="mb-4 text-xs font-black uppercase tracking-[0.2em] text-[#265c52]"
                    >
                      Buscar y filtrar
                    </p>

                    <ExerciseFilters
                      search={search}
                      typeFilter={typeFilter}
                      muscleFilter={muscleFilter}
                      exerciseTypes={exerciseTypeOptions}
                      muscleGroups={muscleGroupOptions}
                      onSearchChange={handleSearchChange}
                      onTypeFilterChange={handleTypeFilterChange}
                      onMuscleFilterChange={handleMuscleFilterChange}
                    />

                    {metadataMessage && (
                      <p
                        data-ui="exercise-metadata-message"
                        className="mt-3 rounded-2xl border border-[#9f2f22]/15 bg-[#fff0ec] pl-3 pr-3 py-2 text-xs font-semibold text-[#9f2f22]"
                      >
                        {metadataMessage}
                      </p>
                    )}
                  </div>
                )}

                <div
                  data-ui="exercise-library-stats-card"
                  className="grid gap-3 rounded-[1.5rem] border border-dashed border-[#1f1b16]/20 bg-white/60 p-4 text-center"
                >
                  <p
                    data-ui="exercise-library-stats-title"
                    className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]"
                  >
                    Biblioteca
                  </p>

                  <div
                    data-ui="exercise-library-stats-grid"
                    className="grid gap-3 sm:grid-cols-2"
                  >
                    <StatTile label="Resultados" value={String(total)} />
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
              data-ui="exercise-center-panel"
              className="flex min-w-0 flex-col rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-5 shadow-[0_2rem_5rem_rgba(47,39,27,0.20)] backdrop-blur-md sm:p-8"
              data-block="exercises-panel"
            >
              <div
                data-ui="exercise-center-panel-header-card"
                className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-6 text-[#fffaf0] shadow-inner"
                data-block="panel-header"
              >
                <div
                  data-ui="exercise-center-panel-header-content"
                  className="flex flex-col gap-5 md:flex-row md:items-end md:justify-between"
                >
                  <div data-ui="exercise-center-panel-title-group">
                    <p
                      data-ui="exercise-center-panel-label"
                      className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]"
                      data-block="panel-label"
                    >
                      Ejercicios
                    </p>

                    <h2
                      data-ui="exercise-center-panel-title"
                      className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]"
                      data-block="panel-title"
                    >
                      Lista disponible
                    </h2>
                  </div>

                  <p
                    data-ui="exercise-center-panel-description"
                    className="max-w-sm text-sm leading-6 text-[#efe4d2]"
                  >
                    Selecciona una tarjeta para ver su ficha completa sin salir
                    de la página.
                  </p>
                </div>
              </div>

              <div
                data-ui="exercise-results-summary-card"
                className="mt-6 flex items-center justify-between gap-4 rounded-[1.5rem] border border-dashed border-[#1f1b16]/15 bg-white/55 pl-4 pr-4 py-3"
              >
                <div data-ui="exercise-results-summary-text-group">
                  <p
                    data-ui="exercise-results-summary-label"
                    className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]"
                  >
                    Resultado actual
                  </p>

                  <p
                    data-ui="exercise-results-summary-text"
                    className="mt-1 text-sm text-[#5d5348]"
                  >
                    {total} ejercicios encontrados con los filtros aplicados.
                  </p>
                </div>

                {selectedExercise && (
                  <span
                    data-ui="selected-exercise-badge"
                    className="rounded-full bg-[#1f1b16] pl-3 pr-3 py-2 text-xs font-bold uppercase tracking-[0.18em] text-[#fffaf0]"
                  >
                    Seleccionado: {selectedExercise.name}
                  </span>
                )}
              </div>

              <div
                data-ui="exercise-list-wrapper"
                className="mt-6 min-h-0 flex-1"
              >
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
                <div
                  data-ui="exercise-pagination-card"
                  className="mt-6 flex items-center justify-between rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/60 pl-4 pr-4 py-3"
                >
                  <button
                    data-ui="exercise-pagination-previous-button"
                    type="button"
                    disabled={page <= 1}
                    onClick={() =>
                      setPage((current) => Math.max(1, current - 1))
                    }
                    className="rounded-xl border border-[#1f1b16]/15 pl-4 pr-4 py-2 text-sm font-bold disabled:cursor-not-allowed disabled:opacity-40"
                  >
                    Anterior
                  </button>

                  <p
                    data-ui="exercise-pagination-text"
                    className="text-sm font-bold text-[#5d5348]"
                  >
                    Página {page} de {totalPages}
                  </p>

                  <button
                    data-ui="exercise-pagination-next-button"
                    type="button"
                    disabled={page >= totalPages}
                    onClick={() =>
                      setPage((current) => Math.min(totalPages, current + 1))
                    }
                    className="rounded-xl border border-[#1f1b16]/15 pl-4 pr-4 py-2 text-sm font-bold disabled:cursor-not-allowed disabled:opacity-40"
                  >
                    Siguiente
                  </button>
                </div>
              )}
            </section>

            <aside
              data-ui="exercise-right-detail-panel"
              className="min-w-0 rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-5 shadow-[0_2rem_5rem_rgba(47,39,27,0.14)] backdrop-blur-md"
              data-block="exercise-detail-panel"
            >
              <div
                data-ui="exercise-right-detail-header-card"
                className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-[linear-gradient(180deg,#265c52_0%,#173e37_100%)] p-6 text-[#fffaf0] shadow-inner"
              >
                <p
                  data-ui="exercise-right-detail-label"
                  className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f6c98d]"
                >
                  Panel Detalles
                </p>

                <h2
                  data-ui="exercise-right-detail-title"
                  className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em]"
                >
                  Ficha rápida
                </h2>
              </div>

              {selectedExercise ? (
                <div
                  data-ui="selected-exercise-detail-content"
                  className="mt-6 grid gap-4"
                >
                  <div
                    data-ui="selected-exercise-name-card"
                    className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/70 p-5"
                  >
                    <div
                      data-ui="selected-exercise-name-card-content"
                      className="flex flex-col items-start gap-4"
                    >
                      <div
                        data-ui="selected-exercise-name-text-group"
                        className="min-w-0"
                      >
                        <p
                          data-ui="selected-exercise-name-label"
                          className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]"
                        >
                          Nombre
                        </p>

                        <h3
                          data-ui="selected-exercise-name"
                          className="mt-3 break-words text-2xl font-black tracking-[-0.04em] text-[#1f1b16]"
                        >
                          {selectedExercise.name}
                        </h3>

                        <p
                          data-ui="selected-exercise-muscle-group"
                          className="mt-2 break-words text-sm font-semibold text-[#5d5348]"
                        >
                          {selectedExercise.muscle_group}
                        </p>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        <button
                          data-ui="open-exercise-insights-button"
                          type="button"
                          className="rounded-2xl bg-[#265c52] pl-4 pr-4 py-3 text-xs font-black uppercase tracking-[0.12em] text-[#fffaf0] transition hover:bg-[#1f1b16]"
                          onClick={() => setIsInsightsModalOpen(true)}
                        >
                          Ver rendimiento
                        </button>

                        {canEditSelectedExercise && (
                          <button
                            data-ui="edit-exercise-button"
                            type="button"
                            className="rounded-2xl border border-[#1f1b16]/15 pl-4 pr-4 py-3 text-xs font-black uppercase tracking-[0.12em] text-[#1f1b16] transition hover:bg-[#1f1b16] hover:text-[#fffaf0]"
                            onClick={() => {
                              setEditErrorMessage("");
                              setIsEditModalOpen(true);
                            }}
                          >
                            Editar
                          </button>
                        )}
                      </div>
                    </div>
                  </div>

                  <div
                    data-ui="selected-exercise-info-grid"
                    className="grid gap-3 sm:grid-cols-2 xl:grid-cols-1"
                  >
                    <InfoTile
                      label="Tipo"
                      value={
                        selectedExercise.exercise_type ?? "Sin especificar"
                      }
                    />
                    <InfoTile
                      label="Musculo secundario"
                      value={
                        selectedExercise.secondary_muscle_group || "No definido"
                      }
                    />
                    <InfoTile
                      label="Estado"
                      value={
                        selectedExercise.is_official === false
                          ? "Ejercicio propio"
                          : "Ejercicio oficial"
                      }
                    />
                  </div>

                  <div
                    data-ui="selected-exercise-description-card"
                    className="rounded-[1.5rem] border border-dashed border-[#1f1b16]/15 bg-[#fff8ea] p-5"
                  >
                    <p
                      data-ui="selected-exercise-description-label"
                      className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]"
                    >
                      Descripcion
                    </p>

                    <p
                      data-ui="selected-exercise-description-text"
                      className="mt-3 break-words text-sm leading-7 text-[#5d5348]"
                    >
                      {selectedExercise.description ||
                        "Todavia no hay una descripcion para este ejercicio. Puedes completarla cuando edites o crees uno nuevo."}
                    </p>
                  </div>

                  <div
                    data-ui="selected-exercise-workout-sessions-card"
                    className="rounded-[1.5rem] border border-[#1f1b16]/10 bg-white/70 p-5"
                  >
                    <div className="flex flex-col items-start gap-3 sm:flex-row sm:justify-between">
                      <div className="min-w-0">
                        <p
                          data-ui="selected-exercise-workout-sessions-label"
                          className="text-xs font-black uppercase tracking-[0.2em] text-[#265c52]"
                        >
                          Entrenos
                        </p>

                        <h3
                          data-ui="selected-exercise-workout-sessions-title"
                          className="mt-2 text-xl font-black tracking-[-0.04em] text-[#1f1b16]"
                        >
                          Sesiones donde aparece
                        </h3>
                      </div>

                      <span className="rounded-full bg-[#265c52]/10 pl-3 pr-3 py-1 text-xs font-black text-[#265c52]">
                        {exerciseWorkoutSessions.length}
                      </span>
                    </div>

                    {exerciseWorkoutSessionsStatus === "loading" && (
                      <p className="mt-4 rounded-2xl border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/80 p-4 text-sm font-semibold text-[#5d5348]">
                        Cargando sesiones...
                      </p>
                    )}

                    {exerciseWorkoutSessionsStatus === "error" && (
                      <p className="mt-4 rounded-2xl border border-[#9f2f22]/15 bg-[#fff0ec] p-4 text-sm font-semibold text-[#9f2f22]">
                        {exerciseWorkoutSessionsMessage}
                      </p>
                    )}

                    {exerciseWorkoutSessionsStatus === "success" &&
                      exerciseWorkoutSessions.length === 0 && (
                        <p className="mt-4 rounded-2xl border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/80 p-4 text-sm font-semibold leading-6 text-[#5d5348]">
                          Este ejercicio todavia no aparece en ninguna sesion
                          registrada.
                        </p>
                      )}

                    {exerciseWorkoutSessions.length > 0 && (
                      <ul className="mt-4 grid max-h-[24rem] gap-3 overflow-y-auto pr-1">
                        {exerciseWorkoutSessions.map((session) => (
                          <li
                            data-ui="selected-exercise-workout-session-item"
                            className="rounded-2xl border border-[#1f1b16]/10 bg-[#fffaf0] p-4"
                            key={session.id}
                          >
                            <div className="flex items-start justify-between gap-3">
                              <div className="min-w-0">
                                <p className="break-words text-sm font-black text-[#1f1b16]">
                                  {session.name ||
                                    session.routine_name ||
                                    "Sesion"}
                                </p>

                                <p className="mt-1 break-words text-xs font-semibold uppercase tracking-[0.14em] text-[#5d5348]">
                                  {session.routine_name || "Entreno libre"}
                                </p>
                              </div>

                              <p className="shrink-0 text-right text-xs font-black text-[#265c52]">
                                {workoutSessionDateFormatter.format(
                                  new Date(session.started_at),
                                )}
                              </p>
                            </div>

                            <p className="mt-3 text-sm font-semibold text-[#5d5348]">
                              Orden {session.exercise_order} ·{" "}
                              {session.set_count} series ·{" "}
                              {session.duration_minutes} min
                            </p>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                </div>
              ) : (
                <div
                  data-ui="no-selected-exercise-message-card"
                  className="mt-6 rounded-[1.5rem] border border-dashed border-[#1f1b16]/20 bg-white/60 p-5"
                >
                  <p
                    data-ui="no-selected-exercise-message-text"
                    className="text-sm font-bold text-[#5d5348]"
                  >
                    Selecciona un ejercicio de la lista para ver su detalle
                    completo aquí.
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

      <ExerciseInsightModal
        isOpen={isInsightsModalOpen}
        exercise={selectedExercise}
        insights={exerciseInsights}
        status={exerciseInsightsStatus}
        message={exerciseInsightsMessage}
        onClose={() => setIsInsightsModalOpen(false)}
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
