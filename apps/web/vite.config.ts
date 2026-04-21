import path from "node:path";

import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  server: {
    fs: {
      allow: [path.resolve(__dirname, "..", "..")],
    },
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
      "../../packages/**/*.test.ts",
      "../../packages/**/*.test.tsx",
      "e2e/**/*.test.ts",
    ],
  },
});
