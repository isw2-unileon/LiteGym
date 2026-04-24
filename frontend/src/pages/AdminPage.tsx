import * as React from "react";
import { Link, Navigate } from "react-router-dom";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

// --- Interfaces ---
interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  is_active: boolean;
}

export default function AdminPage() {
  // --- State: Auth & UI ---
  const [isAuthorized, setIsAuthorized] = React.useState<boolean | null>(null);
  const [activeTab, setActiveTab] = React.useState<"users" | "exercises">("users");
  const [statusMessage, setStatusMessage] = React.useState<{ text: string; type: "success" | "error" } | null>(null);

  // --- State: Data ---
  const [users, setUsers] = React.useState<User[]>([]);

  // --- State: Forms ---
  const [newUser, setNewUser] = React.useState({ username: "", email: "", password: "", role: "user" });
  
  const [newExercise, setNewExercise] = React.useState({ name: "", muscle_group: "", exercise_type: "", description: "" });

  // --- Effect: Check Admin Role & Fetch Data ---
  React.useEffect(() => {
    const verifyAdminAndFetchData = async () => {
      try {
        const meResponse = await fetch(`${API_BASE_URL}/api/users/me`, { credentials: "include" });
        if (!meResponse.ok) throw new Error("Unauthorized");
        
        const meData = await meResponse.json();
        if (meData.role !== "admin") {
          setIsAuthorized(false);
          return;
        }
        
        setIsAuthorized(true);

        const usersResponse = await fetch(`${API_BASE_URL}/api/users`, { credentials: "include" });
        if (usersResponse.ok) {
          const usersData = await usersResponse.json();
          setUsers(usersData);
        }
      } catch {
  setIsAuthorized(false);
}
    };

    verifyAdminAndFetchData();
  }, []);

  // --- Handlers: Users ---
  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await fetch(`${API_BASE_URL}/api/users`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newUser),
        credentials: "include",
      });

      if (!response.ok) throw new Error("Error creating user");
      
      setStatusMessage({ text: "Usuario creado correctamente.", type: "success" });
      setNewUser({ username: "", email: "", password: "", role: "user" }); // Reset form
      
      // Refresh user list
      const usersResponse = await fetch(`${API_BASE_URL}/api/users`, { credentials: "include" });
      if (usersResponse.ok) setUsers(await usersResponse.json());
      
    } catch {
      setStatusMessage({ text: "Error al crear el usuario.", type: "error" });
    }
  };

  const handleDeleteUser = async (id: number) => {
    if (!window.confirm("¿Estás seguro de que deseas eliminar este usuario?")) return;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/users/${id}`, {
        method: "DELETE",
        credentials: "include",
      });

      if (!response.ok) throw new Error("Error deleting user");
      
      setStatusMessage({ text: "Usuario eliminado.", type: "success" });
      setUsers(users.filter((user) => user.id !== id));
    } catch  {
      setStatusMessage({ text: "Error al eliminar el usuario.", type: "error" });
    }
  };

  // --- Handlers: Exercises ---
  const handleCreateExercise = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const payload = {
        ...newExercise,
        is_official: true, 
      };

      const response = await fetch(`${API_BASE_URL}/api/exercises`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
        credentials: "include",
      });

      if (!response.ok) throw new Error("Error creating exercise");
      
      setStatusMessage({ text: "Ejercicio global creado correctamente.", type: "success" });
      
      setNewExercise({ name: "", muscle_group: "", exercise_type: "", description: "" }); 
    } catch  {
      setStatusMessage({ text: "Error al crear el ejercicio.", type: "error" });
    }
  };

  // --- Render States ---
  if (isAuthorized === null) {
    return <div className="flex min-h-screen items-center justify-center bg-[#f4efe2] text-[#1f1b16]">Cargando panel...</div>;
  }

  if (isAuthorized === false) {
    return <Navigate to="/exercises" replace />; // Kick non-admins out to the exercises page
  }

  // --- Main Render ---
  return (
    <main className="min-h-screen overflow-hidden bg-[#f4efe2] text-[#1f1b16]">
      <section className="relative isolate min-h-screen px-6 py-8 sm:px-10 lg:px-16">
        
        {/* Background Gradients (Matching the app theme) */}
        <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_left,_rgba(234,113,48,0.30),_transparent_34%),radial-gradient(circle_at_bottom_right,_rgba(38,92,82,0.35),_transparent_32%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_44%,_#d8e1d0_100%)]" />
        <div className="absolute left-8 top-10 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/25 blur-sm" />
        
        <div className="mx-auto max-w-5xl">
          
          {/* Header & Navigation */}
          <div className="mb-8 flex items-center justify-between">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#f1a45b]">Administración</p>
              <h1 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.04em]">Centro de Mando</h1>
            </div>
            <Link to="/exercises" className="rounded-full bg-white/50 px-5 py-2 text-sm font-bold backdrop-blur transition hover:bg-white/80">
              Volver a la App
            </Link>
          </div>

          {/* Status Message Alert */}
          {statusMessage && (
            <div className={`mb-6 rounded-xl p-4 text-center font-bold ${statusMessage.type === "success" ? "bg-[#265c52]/20 text-[#265c52]" : "bg-[#c94b32]/20 text-[#c94b32]"}`}>
              {statusMessage.text}
            </div>
          )}

          {/* Main Content Glass Panel */}
          <div className="rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-xl backdrop-blur-md sm:p-8">
            
            {/* Tabs */}
            <div className="mb-8 flex border-b border-[#1f1b16]/10">
              <button
                className={`px-6 py-3 font-bold transition-colors ${activeTab === "users" ? "border-b-2 border-[#ea7130] text-[#ea7130]" : "text-[#5d5348] hover:text-[#1f1b16]"}`}
                onClick={() => setActiveTab("users")}
              >
                Gestión de Usuarios
              </button>
              <button
                className={`px-6 py-3 font-bold transition-colors ${activeTab === "exercises" ? "border-b-2 border-[#ea7130] text-[#ea7130]" : "text-[#5d5348] hover:text-[#1f1b16]"}`}
                onClick={() => setActiveTab("exercises")}
              >
                Ejercicios Globales
              </button>
            </div>

            {/* TAB CONTENT: USERS */}
            {activeTab === "users" && (
              <div className="grid gap-8 lg:grid-cols-3">
                {/* Create User Form */}
                <div className="rounded-2xl bg-white/50 p-6 lg:col-span-1">
                  <h3 className="mb-4 text-xl font-black">Nuevo Usuario</h3>
                  <form onSubmit={handleCreateUser} className="flex flex-col gap-4">
                    <input type="text" placeholder="Nombre de usuario" required className="rounded-xl border-none bg-white p-3 shadow-inner" value={newUser.username} onChange={(e) => setNewUser({ ...newUser, username: e.target.value })} />
                    <input type="email" placeholder="Correo electrónico" required className="rounded-xl border-none bg-white p-3 shadow-inner" value={newUser.email} onChange={(e) => setNewUser({ ...newUser, email: e.target.value })} />
                    <input type="password" placeholder="Contraseña" required className="rounded-xl border-none bg-white p-3 shadow-inner" value={newUser.password} onChange={(e) => setNewUser({ ...newUser, password: e.target.value })} />
                    <select className="rounded-xl border-none bg-white p-3 shadow-inner" value={newUser.role} onChange={(e) => setNewUser({ ...newUser, role: e.target.value })}>
                      <option value="user">Usuario normal</option>
                      <option value="admin">Administrador</option>
                    </select>
                    <button type="submit" className="mt-2 rounded-xl bg-[#1f1b16] py-3 font-bold text-[#fffaf0] transition hover:bg-[#ea7130]">Crear Usuario</button>
                  </form>
                </div>

                {/* Users List */}
                <div className="rounded-2xl bg-white/50 p-6 lg:col-span-2">
                  <h3 className="mb-4 text-xl font-black">Usuarios Registrados ({users.length})</h3>
                  <div className="max-h-[400px] overflow-y-auto pr-2">
                    {users.map((user) => (
                      <div key={user.id} className="mb-3 flex items-center justify-between rounded-xl bg-white p-4 shadow-sm">
                        <div>
                          <p className="font-bold">{user.username} <span className="ml-2 rounded bg-gray-200 px-2 py-0.5 text-xs text-gray-600">{user.role}</span></p>
                          <p className="text-sm text-gray-500">{user.email}</p>
                        </div>
                        <button onClick={() => handleDeleteUser(user.id)} className="rounded-lg bg-red-100 px-4 py-2 text-sm font-bold text-red-600 transition hover:bg-red-200">
                          Eliminar
                        </button>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* TAB CONTENT: EXERCISES */}
            {activeTab === "exercises" && (
              <div className="mx-auto max-w-md rounded-2xl bg-white/50 p-6">
                <h3 className="mb-4 text-xl font-black">Crear Ejercicio Global</h3>
                <form onSubmit={handleCreateExercise} className="flex flex-col gap-4">
                  <input type="text" placeholder="Nombre del ejercicio (ej: Press Banca)" required className="rounded-xl border-none bg-white p-3 shadow-inner" value={newExercise.name} onChange={(e) => setNewExercise({ ...newExercise, name: e.target.value })} />
                  <input type="text" placeholder="Grupo muscular (ej: Pecho)" required className="rounded-xl border-none bg-white p-3 shadow-inner" value={newExercise.muscle_group} onChange={(e) => setNewExercise({ ...newExercise, muscle_group: e.target.value })} />
                  <input type="text" placeholder="Tipo (ej: Fuerza, Cardio)" className="rounded-xl border-none bg-white p-3 shadow-inner" value={newExercise.exercise_type} onChange={(e) => setNewExercise({ ...newExercise, exercise_type: e.target.value })} />
                  
                  {/* CORRECTION 4: Added the description input field */}
                  <input type="text" placeholder="Detalle (ej: Tracción horizontal)" className="rounded-xl border-none bg-white p-3 shadow-inner" value={newExercise.description} onChange={(e) => setNewExercise({ ...newExercise, description: e.target.value })} />
                  
                  <button type="submit" className="mt-2 rounded-xl bg-[#265c52] py-3 font-bold text-[#fffaf0] transition hover:bg-[#1a4039]">Publicar Ejercicio</button>
                </form>
              </div>
            )}

          </div>
        </div>
      </section>
    </main>
  );
}