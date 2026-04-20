import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/healthz": {
        target: "http://127.0.0.1:8080",
      },
      "/readyz": {
        target: "http://127.0.0.1:8080",
      },
      "/api": {
        target: "http://127.0.0.1:8080",
      },
      "/ws": {
        target: "http://127.0.0.1:8080",
        ws: true,
      },
    },
  },
  test: {
    environment: "jsdom",
    include: [
      "src/**/*.test.ts",
      "src/**/*.test.tsx",
      "src/**/*.spec.ts",
      "src/**/*.spec.tsx",
    ],
  },
});
