import { useEffect, useMemo, useState } from "react";
import { Link, useOutletContext } from "react-router-dom";
import type { LayoutUser } from "../components/AppLayout";
import { apiUrl } from "../lib/api";
import type { Exercise } from "../types/exercise";

type OutletContext = {
  user?: LayoutUser | null;
};

type ExerciseLoadStatus = "idle" | "loading" | "success" | "error";

export default function DashboardPage() {
  const { user } = useOutletContext<OutletContext>();
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [exerciseStatus, setExerciseStatus] = useState<ExerciseLoadStatus>("idle");

  useEffect(() => {
    const fetchExercises = async () => {
      setExerciseStatus("loading");

      try {
        const response = await fetch(apiUrl("/api/exercises"), {
          credentials: "include",
        });

        if (!response.ok) {
          setExerciseStatus("error");
          return;
        }

        const data = (await response.json()) as Exercise[];
        setExercises(data);
        setExerciseStatus("success");
      } catch {
        setExerciseStatus("error");
      }
    };

    void fetchExercises();
  }, []);

  const latestExercises = useMemo(() => {
    return [...exercises]
      .sort((first, second) => {
        if (first.created_at && second.created_at) {
          return new Date(second.created_at).getTime() - new Date(first.created_at).getTime();
        }

        return second.id - first.id;
      })
      .slice(0, 5);
  }, [exercises]);

  return (
    <>
      <header className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/85 p-6 shadow-[0_24px_60px_rgba(47,39,27,0.14)] backdrop-blur-md">
        <p className="text-sm font-black uppercase tracking-[0.18em] text-[#265c52]">Panel principal</p>
        <h2 className="mt-4 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.05em] sm:text-5xl">
          {user?.username ? `Hola, ${user.username}` : "Hola"}
        </h2>
      </header>

      <section className="mt-8 grid gap-6 lg:grid-cols-[1fr_0.9fr]">
        <article className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/85 p-6 shadow-[0_24px_60px_rgba(47,39,27,0.14)]">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-sm font-black uppercase tracking-[0.18em] text-[#265c52]">Ejercicios</p>
              <h3 className="mt-2 text-2xl font-black tracking-[-0.03em]">Ultimos añadidos</h3>
            </div>
            <Link className="text-sm font-black text-[#b94f22] hover:text-[#1f1b16]" to="/exercises">
              Ver todos
            </Link>
          </div>

          <div className="mt-5">
            {exerciseStatus === "loading" && <p className="text-sm font-semibold text-[#5d5348]">Cargando ejercicios...</p>}
            {exerciseStatus === "error" && <p className="text-sm font-bold text-[#9f2f22]">No se pudieron cargar los ejercicios.</p>}
            {exerciseStatus === "success" && latestExercises.length === 0 && (
              <p className="text-sm font-semibold text-[#5d5348]">Todavia no hay ejercicios para mostrar.</p>
            )}
            {exerciseStatus === "success" && latestExercises.length > 0 && (
              <ul className="space-y-3">
                {latestExercises.map((exercise) => (
                  <li className="rounded-2xl border border-[#1f1b16]/10 bg-white/65 px-4 py-3" key={exercise.id}>
                    <p className="text-base font-black text-[#1f1b16]">{exercise.name}</p>
                    <p className="mt-1 text-sm font-semibold text-[#5d5348]">
                      {exercise.muscle_group}
                      {exercise.exercise_type ? ` · ${exercise.exercise_type}` : ""}
                    </p>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </article>

        <article className="rounded-[2rem] border border-[#1f1b16]/10 bg-white/60 p-6 shadow-[0_24px_60px_rgba(47,39,27,0.10)]">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-sm font-black uppercase tracking-[0.18em] text-[#265c52]">Rutinas</p>
              <h3 className="mt-2 text-2xl font-black tracking-[-0.03em]">Tus rutinas</h3>
            </div>
            <Link className="text-sm font-black text-[#b94f22] hover:text-[#1f1b16]" to="/routines">
              Ver rutinas
            </Link>
          </div>

          <div className="mt-5 rounded-2xl border border-dashed border-[#1f1b16]/20 bg-[#fffaf0]/60 p-4">
            <p className="text-sm font-semibold leading-6 text-[#5d5348]">Todavia no hay rutinas para mostrar.</p>
          </div>
        </article>
      </section>
    </>
  );
}
