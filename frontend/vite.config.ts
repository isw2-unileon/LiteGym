import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig({
    plugins: [react(), tailwindcss()],
    envDir: path.resolve(__dirname, "./"),
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    server: {
        host: "0.0.0.0",
        port: 5173,
        proxy: {
            "/api": {
                target: "http://127.0.0.1:8080",
                changeOrigin: true,
                timeout: 120000,
                proxyTimeout: 120000,
            },
            "/health": {
                target: "http://127.0.0.1:8080",
                changeOrigin: true,
                timeout: 120000,
                proxyTimeout: 120000,
            },
        },
    },
});
