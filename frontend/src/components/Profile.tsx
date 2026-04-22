import { useState, useEffect } from "react";
import { apiUrl } from "../lib/api";


interface UserProfile {
  id: string;
  username?: string;
  email?: string;
  role?: string;
}

export default function Profile() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const response = await fetch(apiUrl("/api/auth/me"), {
          credentials: "include",
        });
        
        if (!response.ok) {
          throw new Error("Profile not found");
        }

        const data = await response.json();
        setUser(data.user);
      } catch (err) {
        const message = err instanceof Error 
          ? err.message 
          : "Unexpected error";

        setError(message);
      } finally {
        setIsLoading(false);
      }
    };

    fetchProfile();
  }, []);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64 border-2 border-dashed border-gray-300 rounded-lg">
        <p className="text-gray-500 font-medium animate-pulse">Loading profile...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border-l-4 border-red-500 p-4 rounded-md">
        <p className="text-red-700 font-medium">Error: {error}</p>
      </div>
    );
  }

  if (!user) return null;

  return (
    <div className="max-w-md w-full bg-white shadow-md rounded-xl overflow-hidden border border-gray-100">
      <div className="px-6 py-8 text-center">
        {/* Circular avatar with the first letter of the username */}
        <div className="inline-flex items-center justify-center w-20 h-20 bg-blue-100 text-blue-600 rounded-full text-4xl font-bold mb-4 uppercase">
          {(user.username ?? "U").charAt(0)}
        </div>
        
        <h2 className="text-2xl font-bold text-gray-800 mb-1">{user.username ?? "Usuario"}</h2>
        <p className="text-gray-500 mb-6">{user.email}</p>
        
        <div className="border-t border-gray-100 pt-6 mt-2">
          <div className="bg-gray-50 rounded-lg p-4 flex justify-between items-center text-sm">
            <span className="text-gray-500 font-medium">Rol</span>
            <span className="text-gray-900 font-semibold">{user.role ?? "user"}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
