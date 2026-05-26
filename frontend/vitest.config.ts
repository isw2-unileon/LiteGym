import { defineConfig } from "vitest/config";
import path from "path";

export default defineConfig({
    resolve: {
        alias: {
            "@": path.resolve(__dirname, ".."),
        },
    },
    test: {
        environment: "happy-dom",
        setupFiles: ["./src/test/setup.ts"],
        coverage: {
            provider: "v8",
            thresholds: {
                statements: 50,
                branches: 50,
                functions: 50,
                lines: 50,
            },
        },
    },
});
