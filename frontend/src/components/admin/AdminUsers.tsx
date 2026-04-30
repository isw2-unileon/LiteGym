import * as React from "react";
import { apiUrl } from "../../lib/api";

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  is_active: boolean;
}

export default function AdminUsers() {
  const [users, setUsers] = React.useState<User[]>([]);
  const [newUser, setNewUser] = React.useState({ username: "", email: "", password: "", role: "user" });
  const [statusMessage, setStatusMessage] = React.useState<{ text: string; type: "success" | "error" } | null>(null);

  React.useEffect(() => {
    const fetchUsers = async () => {
      try {
        const response = await fetch(apiUrl("/api/users"), { credentials: "include" });
        if (response.ok) {
          setUsers(await response.json());
        }
      } catch (error) {
        console.error("Error fetching users:", error);
      }
    };
    void fetchUsers();
  }, []);

  const handleCreateUser = async (event: React.FormEvent) => {
    event.preventDefault();
    try {
      const response = await fetch(apiUrl("/api/users"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newUser),
        credentials: "include",
      });

      if (!response.ok) throw new Error("Error creating user");

      setStatusMessage({ text: "Usuario creado correctamente.", type: "success" });
      setNewUser({ username: "", email: "", password: "", role: "user" });

      // Reload list
      const usersResponse = await fetch(apiUrl("/api/users"), { credentials: "include" });
      if (usersResponse.ok) setUsers(await usersResponse.json());
      
    } catch {
      setStatusMessage({ text: "Error al crear el usuario.", type: "error" });
    }
  };

  const handleDeleteUser = async (id: string) => {
    if (!window.confirm("¿Estás seguro de que deseas eliminar este usuario?")) return;

    try {
      const response = await fetch(apiUrl(`/api/users/${id}`), {
        method: "DELETE",
        credentials: "include",
      });

      if (!response.ok) throw new Error("Error deleting user");

      setStatusMessage({ text: "Usuario eliminado.", type: "success" });
      setUsers(users.filter((item) => item.id !== id));
    } catch {
      setStatusMessage({ text: "Error al eliminar el usuario.", type: "error" });
    }
  };

  return (
    <div className="grid gap-8 lg:grid-cols-3">
      {statusMessage && (
        <div className={`col-span-full rounded-xl p-4 text-center font-bold ${statusMessage.type === "success" ? "bg-[#265c52]/20 text-[#265c52]" : "bg-[#c94b32]/20 text-[#c94b32]"}`}>
          {statusMessage.text}
        </div>
      )}

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

      <div className="rounded-2xl bg-white/50 p-6 lg:col-span-2">
        <h3 className="mb-4 text-xl font-black">Usuarios Registrados ({users.length})</h3>
        <div className="max-h-[400px] overflow-y-auto pr-2">
          {users.map((item) => (
            <div key={item.id} className="mb-3 flex items-center justify-between rounded-xl bg-white p-4 shadow-sm">
              <div>
                <p className="font-bold">{item.username} <span className="ml-2 rounded bg-gray-200 px-2 py-0.5 text-xs text-gray-600">{item.role}</span></p>
                <p className="text-sm text-gray-500">{item.email}</p>
              </div>
              <button onClick={() => handleDeleteUser(item.id)} className="rounded-lg bg-red-100 px-4 py-2 text-sm font-bold text-red-600 transition hover:bg-red-200">
                Eliminar
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}