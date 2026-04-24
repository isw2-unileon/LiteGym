import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { apiUrl } from "../lib/api";

// Define the data contract. Must match the JSON returned by the Go backend.
interface UserProfile {
  id: string;
  username: string;
  email: string;
  role: string;
  created_at: string;
}

export default function Profile() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const response = await fetch(apiUrl("/api/users/me"), {
          credentials: "include",
        });

        if (!response.ok) {
          throw new Error("Profile not found or unauthorized");
        }

        const data = await response.json();
        setUser(data);
      } catch (err) {
        const message = err instanceof Error ? err.message : "Unexpected error";
        setError(message);
      } finally {
        setIsLoading(false);
      }
    };

    fetchProfile();
  }, []); 

  // Render content based on request state.
  const renderContent = () => {
    // LOADING STATE: Aligned with dashed borders and warm colors
    if (isLoading) {
      return (
        <div className="mx-auto flex h-64 w-full items-center justify-center rounded-[2rem] border-2 border-dashed border-[#1f1b16]/20 bg-[#fffaf0]/50 backdrop-blur-sm">
          <p className="animate-pulse font-semibold text-[#5d5348]">Loading profile...</p>
        </div>
      );
    }

    // ERROR STATE: Using the exact red tones from the LoginPage
    if (error) {
      return (
        <div className="mx-auto w-full rounded-[1.5rem] border border-[#c94b32]/20 bg-[#c94b32]/10 p-6">
          <p className="font-bold text-[#9f2f22]">Error: {error}</p>
        </div>
      );
    }

    // SAFETY CHECK
    if (!user) return null;
    return (
      <div className="mx-auto w-full animate-[rise_700ms_ease-out_both] rounded-[2rem] border border-[#1f1b16]/10 bg-[#fffaf0]/80 p-6 shadow-[0_30px_80px_rgba(47,39,27,0.20)] backdrop-blur-md sm:p-8">
        <div className="text-center">
          
          {/* Avatar: Using the dark green accent for contrast */}
          <div className="mx-auto mb-5 flex h-24 w-24 items-center justify-center rounded-full border-4 border-[#fffaf0] bg-[#265c52] text-5xl font-black uppercase text-[#fffaf0] shadow-lg">
            {user.username.charAt(0)}
          </div>
          
          {/* Title: Using the system's font and dark color */}
          <h2 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-3xl font-black tracking-[-0.04em] text-[#1f1b16]">
            {user.username}
          </h2>
          <p className="mt-1 font-medium text-[#5d5348]">{user.email}</p>
          
          {/* DEBUG ROLE: Remove this once we confirm it works */}
          <div className="mt-2 inline-block rounded bg-gray-200 px-3 py-1 text-xs font-bold uppercase tracking-widest text-gray-600">
            Rol actual: {user.role || "DESCONOCIDO / VACÍO"}
          </div>
          
          {/* Dark inner card for secondary data */}
          <div className="mt-8 rounded-[1.5rem] border border-[#1f1b16]/10 bg-[#1f1b16] p-5 text-left shadow-inner">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[#f1a45b]">
              Miembro desde
            </p>
            <p className="mt-1 font-['Aptos_Display','Trebuchet_MS',sans-serif] text-xl font-bold text-[#fffaf0]">
              {new Date(user.created_at).toLocaleDateString()}
            </p>
          </div>

          {/* Admin Panel Button (Only visible if role is admin) */}
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
  };

  // MAIN RENDER: Incorporating the shared background layout
  return (
    <main className="min-h-screen overflow-hidden bg-[#f4efe2] text-[#1f1b16]">
      <section className="relative isolate flex min-h-screen items-center justify-center px-6 py-8 sm:px-10 lg:px-16">
        
        {/* Background Gradients and Shapes (Matching LoginPage) */}
        <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_left,_rgba(234,113,48,0.30),_transparent_34%),radial-gradient(circle_at_bottom_right,_rgba(38,92,82,0.35),_transparent_32%),linear-gradient(135deg,_#f8f0db_0%,_#efe1c3_44%,_#d8e1d0_100%)]" />
        <div className="absolute left-8 top-10 -z-10 h-32 w-32 rounded-full border border-[#1f1b16]/10 bg-white/25 blur-sm" />
        <div className="absolute bottom-8 right-10 -z-10 h-52 w-52 rotate-12 rounded-[3rem] border border-[#1f1b16]/10 bg-[#265c52]/10" />

        {/* Content Wrapper */}
        <div className="w-full max-w-md">
          {renderContent()}
        </div>

      </section>
    </main>
  );
}