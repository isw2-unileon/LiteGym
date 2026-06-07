import * as React from "react";
import CreateExerciseModal, {
  type CreateExercisePayload,
} from "../components/Exercise/CreateExerciseModal";
import ExerciseFilters from "../components/Exercise/ExerciseFilters";
import ExerciseHeader from "../components/Exercise/ExerciseHeader";
import ExerciseInsightModal from "../components/Exercise/ExerciseInsightModal";
import ExerciseList from "../components/Exercise/ExerciseList";
import { apiUrl } from "../lib/api";
import { useIsMobile } from "../lib/useIsMobile";
import type {
  Exercise,
  ExerciseInsights,
  ExerciseMetadataResponse,
  ExerciseStatus,
  ExerciseWorkoutSessionSummary,
  SelectOption,
} from "../types/exercise";
import {HelloHeader} from "@/components/HelloHeader.tsx";
import {Stat} from "@/components/Stat.tsx";
import {Card, CardHeader} from "@/components/Card.tsx";
import { exerciseTypeLabel, muscleGroupLabel } from "../lib/exerciseLabels";
import { DialogPopup } from "@/components/DialogPopup.tsx";

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
  const isMobile = useIsMobile();
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
  const [isDeletingExercise, setIsDeletingExercise] = React.useState(false);
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
  const [isMobileExerciseDetailOpen, setIsMobileExerciseDetailOpen] = React.useState(false);

  const [search, setSearch] = React.useState("");
  const [debouncedSearch, setDebouncedSearch] = React.useState("");
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
  const [limit] = React.useState(200);

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

  React.useEffect(() => {
    if (search === debouncedSearch) {
      return;
    }

    const timeoutId = window.setTimeout(() => {
      setDebouncedSearch(search);
    }, 300);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [search, debouncedSearch]);

  const isSearchDebouncing = search !== debouncedSearch;

  const fetchExercises = React.useCallback(async () => {
    setStatus("loading");
    setMessage("");

    try {
      const params = new URLSearchParams();

      if (debouncedSearch.trim() !== "") {
        params.set("search", debouncedSearch.trim());
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
  }, [debouncedSearch, typeFilter, muscleFilter, page, limit]);

  React.useEffect(() => {
    if (isSearchDebouncing) {
      return;
    }

    void fetchExercises();
  }, [fetchExercises, isSearchDebouncing]);

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

  const filteredExercises = React.useMemo(() => {
    const isCustom = (exercise: Exercise) => exercise.is_official === false;
    return [...exercises].sort(
      (a, b) => Number(isCustom(b)) - Number(isCustom(a)),
    );
  }, [exercises]);

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

  const handleDeleteExercise = async () => {
    if (!selectedExercise) {
      return;
    }

    const confirmed = window.confirm(
      `¿Seguro que quieres eliminar "${selectedExercise.name}"? Esta acción no se puede deshacer.`,
    );
    if (!confirmed) {
      return;
    }

    setIsDeletingExercise(true);

    try {
      const response = await fetch(
        apiUrl(`/api/exercises/${selectedExercise.id}`),
        {
          method: "DELETE",
          credentials: "include",
        },
      );

      if (!response.ok) {
        const errorBody = (await response.json().catch(() => null)) as {
          error?: string;
        } | null;

        window.alert(errorBody?.error ?? "No se pudo eliminar el ejercicio.");
        return;
      }

      const deletedId = selectedExercise.id;
      setExercises((current) =>
        current.filter((exercise) => exercise.id !== deletedId),
      );
      setSelectedExerciseId("");
    } catch (error) {
      console.error(error);
      window.alert(
        "No se pudo eliminar el ejercicio. Revisa la conexion con el backend.",
      );
    } finally {
      setIsDeletingExercise(false);
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
  const secondaryMuscleGroupValues = selectedExercise
    ? selectedExercise.secondary_muscle_groups?.filter(Boolean) ??
      (selectedExercise.secondary_muscle_group
        ? selectedExercise.secondary_muscle_group
            .split(",")
            .map((value) => value.trim())
            .filter(Boolean)
        : [])
    : [];

  return (
      <main className="relative isolate overflow-x-hidden text-[#1f1b16] [font-family:'Inter',system-ui,sans-serif] antialiased pb-20">
        <div
            aria-hidden="true"
            className="pointer-events-none absolute left-8 top-12 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/20 blur-[1px]"
        />
        <div
            aria-hidden="true"
            className="pointer-events-none absolute bottom-16 right-12 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10"
        />
        <div className="px-4 pt-5 sm:px-6 sm:pt-8 md:px-8">
          <section className="mx-auto mb-6 grid max-w-[1280px] grid-cols-1 items-start gap-6 md:grid-cols-[1fr_auto]">
            <div>
              <HelloHeader page={"PANEL PRINCIPAL"} user={currentUser?.username ?? "Atleta"} />
              <ExerciseHeader />
            </div>
            <div className="grid w-full grid-cols-3 gap-2.5 md:flex md:w-auto md:flex-wrap">
              <Stat n={String(exercises.length)} l="Ejercicios" />
              <Stat n={String(officialCount)} l="Ejercicios oficiales" />
              <Stat n={String(customCount)} l="Ejercicios no oficiales" />
            </div>
          </section>

          <section className="mx-auto max-w-[1280px]">
            <div className="grid grid-cols-1 justify-center gap-[18px] rounded-[20px] border border-[#1f1b16]/12 bg-[#fffaf0]/75 p-2.5 backdrop-blur xl:[grid-template-columns:3fr_7fr]">
              <Card accent="#ea7130" className="hidden flex-col md:flex" dark={true}>
                <CardHeader kicker={"Control"} title={"Crea un ejercicio"} onDark={true} />
                <button
                    data-ui="create-exercise-button"
                    type="button"
                    className="mt-4 shrink-0 group relative cursor-pointer overflow-hidden rounded-[14px] bg-[#ea7130] px-6 py-4 text-[13px] font-extrabold tracking-[0.04em] text-[#1f1b16] transition hover:-translate-y-px hover:bg-[#ff8b47] disabled:opacity-50"
                    onClick={() => {
                      setCreateErrorMessage("");
                      setIsCreateModalOpen(true);
                    }}
                >
                  Crear nuevo ejercicio
                </button>
              </Card>
              <Card accent="#ea7130" className="flex flex-col">
                <CardHeader
                  kicker={"Filtro y Búsqueda"}
                  title={""}
                  right={(
                    <button
                      data-ui="create-exercise-button-mobile"
                      type="button"
                      className="rounded-[12px] bg-[#1f1b16] px-3 py-2 text-xs font-black text-[#f1a45b] md:hidden"
                      onClick={() => {
                        setCreateErrorMessage("");
                        setIsCreateModalOpen(true);
                      }}
                    >
                      Crear
                    </button>
                  )}
                />
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
              </Card>
            </div>
          </section>

          {isMobile && (
          <section className="mx-auto mt-4 grid w-full max-w-[1280px] min-w-0 gap-4">
            {selectedExercise && (
              <button
                type="button"
                onClick={() => setIsMobileExerciseDetailOpen(true)}
                className="w-full min-w-0 rounded-[22px] border border-[#1f1b16]/12 bg-[#1f1b16] p-4 text-left text-[#fffaf0] shadow-[0_16px_36px_rgba(31,27,22,0.18)]"
              >
                <span className="block [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.18em] text-[#f1a45b]">
                  Ejercicio seleccionado
                </span>
                <span className="mt-2 block [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[24px] font-black leading-none tracking-[-0.04em]">
                  {selectedExercise.name}
                </span>
                <span className="mt-3 block text-sm font-bold text-[#fffaf0]/75">
                  Toca para ver músculos, rendimiento y edición.
                </span>
              </button>
            )}

            <MobileExerciseDeck
              status={status}
              message={message}
              exercises={filteredExercises}
              selectedExerciseId={selectedExercise?.id}
              onOpenExercise={(exercise) => {
                setSelectedExerciseId(exercise.id);
                setIsMobileExerciseDetailOpen(true);
              }}
            />
          </section>
          )}

          {!isMobile && (
          <section className="mx-auto grid max-w-[1280px] grid-cols-1 gap-[18px] xl:grid-cols-[7fr_3fr]">
            <div className="flex flex-col gap-[18px] mt-4">
              <Card accent="#ea7130" className={"flex flex-col xl:h-[70vh]"}>
                <CardHeader kicker={"EJERCICIOS"} title={"Elige un ejercicio"} />
                <ExerciseList
                    status={status}
                    message={message}
                    exercises={filteredExercises}
                    selectedExerciseId={selectedExercise?.id}
                    onSelectExercise={(exercise) =>
                        setSelectedExerciseId(exercise.id)
                    }
                />
              </Card>
            </div>
            <div className="flex flex-col gap-[18px] mt-4">
              <Card accent="#ea7130" className={"flex flex-col xl:h-[70vh]"}>
                <CardHeader
                    kicker={"DETALLE"}
                    title={selectedExercise?.name ?? "Detalle del ejercicio"}
                    right={
                      <button
                          type="button"
                          onClick={handleDeleteExercise}
                          disabled={!canEditSelectedExercise || isDeletingExercise}
                          className="inline-flex shrink-0 items-center justify-center gap-2 rounded-[10px] border border-[#9f2f22]/25 bg-[#fff0ec] px-3.5 py-2 text-[14px] font-bold text-[#9f2f22] transition hover:-translate-y-px hover:border-[#9f2f22]/45 hover:bg-[#9f2f22]/10 disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {isDeletingExercise ? "Eliminando..." : "Eliminar"}
                      </button>
                    }
                />
                <div className="relative z-[2] mt-4 grid grid-cols-1 gap-2.5 sm:grid-cols-2">
                  <button
                      type="button"
                      onClick={() => setIsInsightsModalOpen(true)}
                      disabled={!selectedExercise}
                      className="inline-flex w-full items-center justify-center gap-2 rounded-[10px] border border-[#1f1b16]/15 bg-[#fffaf0]/75 px-3.5 py-2 text-[14px] font-bold text-[#1f1b16] transition hover:-translate-y-px hover:border-[#ea7130]/40 hover:bg-[#f1a45b]/10 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    Ver rendimiento
                  </button>
                  <button
                      type="button"
                      onClick={() => {
                        setEditErrorMessage("");
                        setIsEditModalOpen(true);
                      }}
                      disabled={!canEditSelectedExercise}
                      className="inline-flex w-full items-center justify-center gap-2 rounded-[10px] border border-[#1f1b16]/15 bg-[#fffaf0]/75 px-3.5 py-2 text-[14px] font-bold text-[#1f1b16] transition hover:-translate-y-px hover:border-[#ea7130]/40 hover:bg-[#f1a45b]/10 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    Editar
                  </button>
                </div>
                {selectedExercise ? (
                  <div className="relative z-[2] mt-4 flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto pr-1">
                    <InfoTile
                        label="Tipo"
                        value={selectedExercise.exercise_type ? exerciseTypeLabel(selectedExercise.exercise_type) : "—"}
                    />
                    <InfoTile
                        label="Grupos musculares"
                        value={
                          [selectedExercise.muscle_group, ...secondaryMuscleGroupValues]
                            .filter(Boolean)
                            .map(muscleGroupLabel)
                            .join(", ") || "—"
                        }
                    />
                    <InfoTile
                        label="Estado"
                        value={selectedExercise.is_official === false ? "No oficial" : "Oficial"}
                    />
                    <InfoTile
                        label="Descripción"
                        value={selectedExercise.description?.trim() || "Sin descripción."}
                    />

                    {currentUser?.role === "admin" && selectedExercise.owner_user_id && (
                      <InfoTile
                          label="Propietario"
                          value={`${selectedExercise.owner_user_id} · ${selectedExercise.owner_email ?? "—"}`}
                      />
                    )}

                    <Card accent="#ea7130">
                      <CardHeader kicker={"SESIONES"} title={""} right={selectedExercise ? (
                          <span className="rounded-full bg-[#265c52]/10 px-3 py-1 text-xs font-black text-[#265c52]">
                          {exerciseWorkoutSessions.length}
                        </span>
                      ) : undefined}/>

                      <div className="relative z-[2] mt-4 max-h-64 overflow-y-auto pr-1">
                        {exerciseWorkoutSessionsStatus === "loading" && (
                          <p className="text-[14px] font-semibold text-[#3a332c]/70">
                            Cargando sesiones...
                          </p>
                        )}

                        {exerciseWorkoutSessionsStatus === "error" && (
                          <p className="text-[14px] font-bold text-[#9f2f22]">
                            {exerciseWorkoutSessionsMessage}
                          </p>
                        )}

                        {exerciseWorkoutSessionsStatus === "success" &&
                          exerciseWorkoutSessions.length === 0 && (
                            <p className="rounded-[16px] border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/60 p-6 text-center text-[14px] font-semibold text-[#3a332c]/70">
                              Este ejercicio todavía no aparece en ninguna sesión registrada.
                            </p>
                          )}

                        {exerciseWorkoutSessionsStatus === "success" &&
                          exerciseWorkoutSessions.length > 0 && (
                            <ul className="grid gap-2.5">
                              {exerciseWorkoutSessions.map((session) => (
                                <li
                                    key={session.id}
                                    className="rounded-[16px] border border-[#1f1b16]/10 bg-[#fffaf0]/75 p-4"
                                >
                                  <div className="flex items-center justify-between gap-2">
                                    <span className="inline-flex items-center rounded-md border border-[#265c52]/18 bg-[#265c52]/10 px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-wider text-[#265c52]">
                                      {session.routine_name || "Sesión libre"}
                                    </span>
                                    <span className="inline-flex items-center rounded-md bg-[#1f1b16] px-2 py-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase tracking-[0.16em] text-[#f1a45b]">
                                      {session.set_count} {session.set_count === 1 ? "set" : "sets"}
                                    </span>
                                  </div>
                                  <h4 className="m-0 mt-2.5 [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[18px] font-black leading-tight tracking-[-0.03em] text-[#1f1b16]">
                                    {session.name || "Entreno"}
                                  </h4>
                                  <p className="mt-1.5 text-[14px] font-semibold text-[#3a332c]/80">
                                    {workoutSessionDateFormatter.format(new Date(session.started_at))} · {session.duration_minutes} min
                                  </p>
                                </li>
                              ))}
                            </ul>
                          )}
                      </div>
                    </Card>
                  </div>
                ) : (
                  <p className="relative z-[2] mt-4 rounded-[16px] border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/60 p-6 text-center text-[14px] font-semibold text-[#3a332c]/70">
                    Selecciona un ejercicio del listado para ver su detalle aquí.
                  </p>
                )}
              </Card>
            </div>
          </section>
          )}


        </div>
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
        {isMobileExerciseDetailOpen && selectedExercise && (
          <DialogPopup
            kicker="Ejercicio"
            title={selectedExercise.name}
            onClose={() => setIsMobileExerciseDetailOpen(false)}
          >
            <div className="grid gap-3">
              <div className="rounded-[16px] border border-[#1f1b16]/10 bg-white/70 p-3">
                <p className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.16em] text-[#265c52]">Tipo</p>
                <p className="mt-1 text-sm font-bold text-[#1f1b16]">
                  {selectedExercise.exercise_type ? exerciseTypeLabel(selectedExercise.exercise_type) : "Sin tipo"}
                </p>
              </div>
              <div className="rounded-[16px] border border-[#1f1b16]/10 bg-white/70 p-3">
                <p className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.16em] text-[#265c52]">Músculos</p>
                <p className="mt-1 text-sm font-bold text-[#1f1b16]">
                  {[selectedExercise.muscle_group, ...secondaryMuscleGroupValues].filter(Boolean).map(muscleGroupLabel).join(", ") || "Sin grupo"}
                </p>
              </div>
            </div>
            <div className="mt-4 grid grid-cols-2 gap-2">
              <button type="button" onClick={() => { setIsMobileExerciseDetailOpen(false); setIsInsightsModalOpen(true); }} className="rounded-[14px] bg-[#ea7130] px-3 py-3 text-sm font-black text-[#1f1b16]">
                Rendimiento
              </button>
              <button type="button" onClick={() => { setIsMobileExerciseDetailOpen(false); setEditErrorMessage(""); setIsEditModalOpen(true); }} disabled={!canEditSelectedExercise} className="rounded-[14px] border border-[#1f1b16]/15 px-3 py-3 text-sm font-black text-[#1f1b16] disabled:opacity-50">
                Editar
              </button>
            </div>
          </DialogPopup>
        )}
      </main>
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

function MobileExerciseDeck({
  status,
  message,
  exercises,
  selectedExerciseId,
  onOpenExercise,
}: {
  status: ExerciseStatus;
  message: string;
  exercises: Exercise[];
  selectedExerciseId?: string;
  onOpenExercise: (exercise: Exercise) => void;
}) {
  return (
    <section className="min-w-0 max-w-full overflow-hidden rounded-[28px] border border-[#1f1b16]/10 bg-[#fffaf0]/86 shadow-[0_14px_34px_rgba(31,27,22,0.1)]">
      <div className="flex items-center justify-between gap-3 bg-[#ea7130] px-5 py-4">
        <div className="min-w-0">
          <p className="[font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.18em] text-[#1f1b16]/70">
            Ejercicios
          </p>
          <h2 className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[31px] font-black leading-none tracking-[-0.055em] text-[#1f1b16]">
            {exercises.length} resultados
          </h2>
        </div>
      </div>

      <div className="grid gap-2 p-3">
        {status === "loading" && (
          <p className="rounded-[18px] bg-white/60 px-4 py-5 text-sm font-bold text-[#3a332c]">
            Cargando ejercicios...
          </p>
        )}

        {status === "error" && (
          <p className="rounded-[18px] border border-[#9f2f22]/20 bg-[#9f2f22]/8 px-4 py-5 text-sm font-black text-[#9f2f22]">
            {message}
          </p>
        )}

        {status === "success" && exercises.length === 0 && (
          <p className="rounded-[18px] border border-dashed border-[#1f1b16]/14 bg-white/60 px-4 py-5 text-sm font-bold text-[#3a332c]">
            No hay ejercicios con esos filtros.
          </p>
        )}

        {status === "success" &&
          exercises.map((exercise) => {
            const isSelected = exercise.id === selectedExerciseId;
            return (
              <button
                key={exercise.id}
                type="button"
                onClick={() => onOpenExercise(exercise)}
                className={[
                  "w-full min-w-0 max-w-full overflow-hidden rounded-[20px] border px-4 py-3 text-left transition",
                  isSelected
                    ? "border-[#1f1b16] bg-[#1f1b16] text-[#fffaf0] shadow-[0_12px_26px_rgba(31,27,22,0.16)]"
                    : "border-[#1f1b16]/10 bg-white/64 text-[#1f1b16]",
                ].join(" ")}
              >
                <div className="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-start gap-2.5">
                  <div className="min-w-0 max-w-full overflow-hidden">
                    <p className="max-w-full truncate [font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[21px] font-black leading-none tracking-[-0.04em]">
                      {exercise.name}
                    </p>
                    <p className={["mt-2 max-w-full truncate text-[11px] font-black uppercase tracking-[0.08em]", isSelected ? "text-[#fffaf0]/66" : "text-[#3a332c]/72"].join(" ")}>
                      {exercise.exercise_type ? exerciseTypeLabel(exercise.exercise_type) : "Sin tipo"}
                    </p>
                  </div>
                  <span
                    className={[
                      "max-w-[5.4rem] shrink-0 truncate rounded-[13px] px-2.5 py-1 text-center [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-black uppercase tracking-[0.08em]",
                      exercise.is_official === false
                        ? "bg-[#ea7130] text-[#1f1b16]"
                        : isSelected
                          ? "bg-[#fffaf0]/10 text-[#f1a45b]"
                          : "bg-[#265c52]/10 text-[#265c52]",
                    ].join(" ")}
                  >
                    {exercise.is_official === false ? "Propio" : "Oficial"}
                  </span>
                </div>
                <div className="mt-3 flex min-w-0 max-w-full flex-wrap gap-1.5 overflow-hidden">
                  {[exercise.muscle_group, ...(exercise.secondary_muscle_groups ?? [])].filter(Boolean).slice(0, 3).map((muscle) => (
                    <span
                      key={muscle}
                      title={muscleGroupLabel(muscle)}
                      className={[
                        "max-w-full truncate rounded-full px-2.5 py-1 text-[11px] font-black sm:max-w-[10rem]",
                        isSelected ? "bg-[#fffaf0]/10 text-[#fffaf0]/78" : "bg-[#1f1b16]/6 text-[#3a332c]",
                      ].join(" ")}
                    >
                      {muscleGroupLabel(muscle)}
                    </span>
                  ))}
                </div>
              </button>
            );
          })}
      </div>
    </section>
  );
}
