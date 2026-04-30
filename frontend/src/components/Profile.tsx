import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { apiUrl } from "../lib/api";

interface UserProfile {
  id: string;
  username: string;
  email: string;
  role: string;
  created_at?: string;
}

export default function Profile() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const response = await fetch(apiUrl("/api/auth/me"), {
          credentials: "include",
        });

        if (!response.ok) {
          throw new Error("Profile not found or unauthorized");
        }

        const payload = (await response.json()) as { user: UserProfile };
        setUser(payload.user);
      } catch (err) {
        const message = err instanceof Error ? err.message : "Unexpected error";
        setError(message);
      } finally {
        setIsLoading(false);
      }
    };

    void fetchProfile();
  }, []);

  if (isLoading) {
    return (
      <div className="mx-auto flex h-64 w-full items-center justify-center rounded-[2rem] border-2 border-dashed border-[#1f1b16]/20 bg-[#fffaf0]/50 backdrop-blur-sm">
        <p className="animate-pulse font-semibold text-[#5d5348]">Loading profile...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="mx-auto w-full rounded-[1.5rem] border border-[#c94b32]/20 bg-[#c94b32]/10 p-6">
        <p className="font-bold text-[#9f2f22]">Error: {error}</p>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return (
    <div className="mx-auto w-full animate-[rise_700ms_ease-out_both] rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-[0_30px_80px_rgba(47,39,27,0.20)] backdrop-blur-md sm:p-8">
      <div className="text-center">
        <div className="mx-auto mb-5 flex h-24 w-24 items-center justify-center rounded-full border-4 border-[#fffaf0] bg-[#265c52] text-5xl font-black uppercase text-[#fffaf0] shadow-lg">
          {user.username.charAt(0)}
        </div>

        <h2 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em] text-[#1f1b16]">
          {user.username}
        </h2>
        <p className="mt-1 font-medium text-[#5d5348]">{user.email}</p>

        <div className="mt-2 inline-block rounded bg-gray-200 px-3 py-1 text-xs font-bold uppercase tracking-widest text-gray-600">
          Rol actual: {user.role}
        </div>

        {user.created_at && (
          <div className="mt-8 rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-5 text-left shadow-inner">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[#f1a45b]">
              Miembro desde
            </p>
            <p className="mt-1 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-xl font-bold text-[#fffaf0]">
              {new Date(user.created_at).toLocaleDateString()}
            </p>
          </div>
        )}

        {user.role === "admin" && (
          <div className="mt-4">
            <Link to="/admin" className="block w-full rounded-xl bg-[#ea7130] py-3 font-bold text-white shadow-lg transition-transform hover:scale-[1.02]">
              Panel de Administración
            </Link>
          </div>
        )}
      </div>
    </div>
  );
}
