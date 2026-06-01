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
    <div className="grid gap-[24px] lg:grid-cols-[1fr_1.8fr] items-start">
      
      {/* Toast Alert Status Message */}
      {statusMessage && (
        <div className={`col-span-full rounded-2xl px-5 py-4 border text-sm font-semibold tracking-tight text-center transition ${
          statusMessage.type === "success" 
            ? "bg-[#265c52]/10 text-[#265c52] border-[#265c52]/20" 
            : "bg-[#c94b32]/10 text-[#c94b32] border-[#c94b32]/20"
        }`}>
          {statusMessage.text}
        </div>
      )}

      {/* Card Column: Create New User */}
      <article className="flex flex-col group relative isolate overflow-hidden rounded-[24px] border border-[#1f1b16]/12 bg-[#fffaf0]/85 p-5 backdrop-blur sm:p-6 lg:col-span-1">
        <span aria-hidden="true" className="absolute inset-x-0 top-0 z-[1] h-1.5 bg-[#ea7130]" />
        
        <header className="relative z-[2] mb-5">
          <div className="mb-1.5 flex items-center gap-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.22em] text-[#ea7130]">
            <span className="inline-block h-0.5 w-4 bg-[#ea7130]" /> Nuevo Registro
          </div>
          <h3 className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[22px] font-black tracking-tight text-[#1f1b16]">
            Nuevo Usuario
          </h3>
        </header>

        <form onSubmit={handleCreateUser} className="relative z-[2] flex flex-col gap-4">
          <div className="flex flex-col">
            <label className="mb-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-wider text-[#3a332c]">
              Nombre de usuario
            </label>
            <input 
              type="text" 
              placeholder="Nombre de usuario" 
              required 
              className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/70 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/12" 
              value={newUser.username} 
              onChange={(e) => setNewUser({ ...newUser, username: e.target.value })} 
            />
          </div>

          <div className="flex flex-col">
            <label className="mb-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-wider text-[#3a332c]">
              Correo electrónico
            </label>
            <input 
              type="email" 
              placeholder="Correo electrónico" 
              required 
              className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/70 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/12" 
              value={newUser.email} 
              onChange={(e) => setNewUser({ ...newUser, email: e.target.value })} 
            />
          </div>

          <div className="flex flex-col">
            <label className="mb-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-wider text-[#3a332c]">
              Contraseña
            </label>
            <input 
              type="password" 
              placeholder="Contraseña" 
              required 
              className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/70 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none transition focus:border-[#ea7130] focus:ring-4 focus:ring-[#ea7130]/12" 
              value={newUser.password} 
              onChange={(e) => setNewUser({ ...newUser, password: e.target.value })} 
            />
          </div>

          <div className="flex flex-col">
            <label className="mb-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-wider text-[#3a332c]">
              Rol Asignado
            </label>
            <select 
              className="w-full rounded-[14px] border border-[#1f1b16]/12 bg-white/90 px-4 py-3 text-[14px] font-semibold text-[#1f1b16] outline-none cursor-pointer transition focus:border-[#ea7130]"
              value={newUser.role} 
              onChange={(e) => setNewUser({ ...newUser, role: e.target.value })}
            >
              <option value="user">Usuario normal</option>
              <option value="admin">Administrador</option>
            </select>
          </div>

          <button 
            type="submit" 
            className="group relative mt-2 w-full overflow-hidden rounded-[14px] bg-[#ea7130] px-5 py-3.5 text-[13px] font-black text-[#1f1b16] shadow-[0_12px_24px_rgba(234,113,48,0.22)] transition hover:-translate-y-px hover:bg-[#ff8b47]"
          >
            Crear Usuario
          </button>
        </form>
      </article>

      {/* Card Column: Registered Users list */}
      <article className="flex flex-col group relative isolate overflow-hidden rounded-[24px] border border-[#1f1b16]/12 bg-[#fffaf0]/85 p-5 backdrop-blur sm:p-6 lg:col-span-1 lg:h-full">
        <span aria-hidden="true" className="absolute inset-x-0 top-0 z-[1] h-1.5 bg-[#265c52]" />
        
        <header className="relative z-[2] mb-5">
          <div className="mb-1.5 flex items-center gap-2 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[10px] font-bold uppercase tracking-[0.22em] text-[#265c52]">
            <span className="inline-block h-0.5 w-4 bg-[#265c52]" /> Directorio Activo
          </div>
          <h3 className="[font-family:'Bricolage_Grotesque','Aptos_Display',sans-serif] text-[22px] font-black tracking-tight text-[#1f1b16]">
            Usuarios Registrados ({users.length})
          </h3>
        </header>

        <div className="relative z-[2] max-h-[460px] overflow-y-auto pr-1 flex flex-col gap-3">
          {users.map((item) => (
            <div 
              key={item.id} 
              className="flex items-center justify-between rounded-[16px] border border-[#1f1b16]/10 bg-white/70 p-4 transition hover:shadow-sm hover:border-[#1f1b16]/18"
            >
              <div>
                <div className="flex items-center gap-2">
                  <p className="m-0 text-[14px] font-extrabold tracking-tight text-[#1f1b16]">
                    {item.username}
                  </p>
                  <span className={`rounded px-2 py-0.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[9px] font-extrabold uppercase tracking-wider ${
                    item.role === 'admin' 
                      ? 'bg-[#ea7130]/15 text-[#ea7130]' 
                      : 'bg-[#265c52]/12 text-[#265c52]'
                  }`}>
                    {item.role}
                  </span>
                </div>
                <p className="mt-1 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[11px] font-bold text-[#3a332c]/60">
                  {item.email}
                </p>
              </div>
              <button 
                onClick={() => handleDeleteUser(item.id)} 
                className="rounded-[10px] border border-[#c94b32]/20 bg-[#c94b32]/8 px-3.5 py-2 text-[12px] font-bold text-[#c94b32] transition hover:bg-[#c94b32] hover:text-white"
              >
                Eliminar
              </button>
            </div>
          ))}
        </div>
      </article>

    </div>
  );
}