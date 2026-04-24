import * as React from "react";
import { Link } from "react-router-dom"; 
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
          setMessage("Failed to load exercises.");
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
      const matchesType =
        typeFilter === "" || exercise.exercise_type === typeFilter;

      const matchesMuscle =
        muscleFilter === "" || exercise.muscle_group === muscleFilter;

      const matchesSearch =
        normalizedSearch === "" ||
        [
          exercise.name,
          exercise.exercise_type ?? "",
          exercise.muscle_group ?? "",
        ].some((value) => value.toLowerCase().includes(normalizedSearch));

      return matchesType && matchesMuscle && matchesSearch;
    });
  }, [exercises, typeFilter, muscleFilter, search]);

  return (
    <main
      className="min-h-screen overflow-hidden bg-[#f4efe2] text-[#1f1b16]"
      data-section="exercise-page"
    >
      <section
        className="relative isolate min-h-screen px-6 py-8 sm:px-10 lg:px-16"
        data-block="page-layout"
      >
        {/* Background Gradients and Shapes */}
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
          
          {/* Top Navigation Bar for Profile Link */}
          <div className="mb-6 flex justify-end">
            <Link
              to="/profile"
              className="inline-flex items-center gap-2 rounded-full border border-[#1f1b16]/15 bg-white/35 px-5 py-2 text-sm font-bold text-[#265c52] backdrop-blur transition hover:bg-white/50 hover:text-[#1f1b16] shadow-sm"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" />
              </svg>
              Mi perfil
            </Link>
          </div>

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