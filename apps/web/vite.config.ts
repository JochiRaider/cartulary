import path from "node:path";

import react from "@vitejs/plugin-react";
import { defineConfig, defineProject } from "vitest/config";

const browserUnitIncludes = [
  "src/**/*.test.ts",
  "src/**/*.test.tsx",
  "src/**/*.spec.ts",
  "src/**/*.spec.tsx",
  "../../packages/*/src/**/*.test.ts",
  "../../packages/*/src/**/*.test.tsx",
];

const harnessNodeIncludes = ["e2e/**/*.test.ts"];
const e2eAPIOrigin =
  process.env.CARTULARY_WEB_E2E_API_ORIGIN ?? "http://127.0.0.1:8080";
const e2eBackendProxy = {
  "/healthz": {
    target: e2eAPIOrigin,
  },
  "/readyz": {
    target: e2eAPIOrigin,
  },
  "/api": {
    target: e2eAPIOrigin,
  },
  "/ws": {
    target: e2eAPIOrigin,
    ws: true,
  },
};

export default defineConfig({
  plugins: [react()],
  server: {
    fs: {
      allow: [path.resolve(__dirname, "..", "..")],
    },
    proxy: e2eBackendProxy,
  },
  preview: {
    proxy: e2eBackendProxy,
  },
  test: {
    projects: [
      defineProject({
        test: {
          name: "browser-unit",
          environment: "jsdom",
          include: browserUnitIncludes,
          setupFiles: [
            "./src/testing/testSetup.ts",
            "./src/testing/testSetup.dom.ts",
          ],
        },
      }),
      defineProject({
        test: {
          name: "harness-node",
          environment: "node",
          include: harnessNodeIncludes,
          setupFiles: ["./src/testing/testSetup.ts"],
        },
      }),
    ],
  },
});
