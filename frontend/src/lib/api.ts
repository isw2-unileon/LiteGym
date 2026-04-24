const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

export function apiUrl(path: string) {
  return `${API_BASE_URL}${path}`;
}
