import * as React from "react";
import { apiUrl } from "../../lib/api";

export default function AdminExercises() {
  const [newExercise, setNewExercise] = React.useState({ name: "", muscle_group: "", exercise_type: "", description: "" });
  const [statusMessage, setStatusMessage] = React.useState<{ text: string; type: "success" | "error" } | null>(null);

  const handleCreateExercise = async (event: React.FormEvent) => {
    event.preventDefault();
    try {
      const payload = { ...newExercise, is_official: true };

      const response = await fetch(apiUrl("/api/exercises"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
        credentials: "include",
      });

      if (!response.ok) throw new Error("Error creating exercise");

      setStatusMessage({ text: "Ejercicio global creado correctamente.", type: "success" });
      setNewExercise({ name: "", muscle_group: "", exercise_type: "", description: "" });
    } catch {
      setStatusMessage({ text: "Error al crear el ejercicio.", type: "error" });
    }
  };

  return (
    <div className="mx-auto max-w-md rounded-2xl bg-white/50 p-6">
      {statusMessage && (
        <div className={`mb-6 rounded-xl p-4 text-center font-bold ${statusMessage.type === "success" ? "bg-[#265c52]/20 text-[#265c52]" : "bg-[#c94b32]/20 text-[#c94b32]"}`}>
          {statusMessage.text}
        </div>
      )}

      <h3 className="mb-4 text-xl font-black">Crear Ejercicio Global</h3>
      <form onSubmit={handleCreateExercise} className="flex flex-col gap-4">
        <input type="text" placeholder="Nombre del ejercicio (ej: Press Banca)" required className="rounded-xl border-none bg-white p-3 shadow-inner" value={newExercise.name} onChange={(e) => setNewExercise({ ...newExercise, name: e.target.value })} />
        <input type="text" placeholder="Grupo muscular (ej: Pecho)" required className="rounded-xl border-none bg-white p-3 shadow-inner" value={newExercise.muscle_group} onChange={(e) => setNewExercise({ ...newExercise, muscle_group: e.target.value })} />
        <input type="text" placeholder="Tipo (ej: Fuerza, Cardio)" className="rounded-xl border-none bg-white p-3 shadow-inner" value={newExercise.exercise_type} onChange={(e) => setNewExercise({ ...newExercise, exercise_type: e.target.value })} />
        <input type="text" placeholder="Detalle (ej: Tracción horizontal)" className="rounded-xl border-none bg-white p-3 shadow-inner" value={newExercise.description} onChange={(e) => setNewExercise({ ...newExercise, description: e.target.value })} />
        <button type="submit" className="mt-2 rounded-xl bg-[#265c52] py-3 font-bold text-[#fffaf0] transition hover:bg-[#1a4039]">Publicar Ejercicio</button>
      </form>
    </div>
  );
}