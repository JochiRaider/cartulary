import { readFileSync } from "node:fs";
import { Agent } from "node:https";
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
  process.env.CARTULARY_WEB_E2E_API_ORIGIN ??
  process.env.CARTULARY_DEV_API_ORIGIN ??
  "https://127.0.0.1:8080";
const tlsRoot =
  process.env.CARTULARY_WEB_E2E_TLS_ROOT_CERTIFICATE ??
  process.env.CARTULARY_DEV_TLS_ROOT_CERTIFICATE;
const tlsCertificate =
  process.env.CARTULARY_WEB_E2E_TLS_CERTIFICATE ??
  process.env.CARTULARY_DEV_TLS_CERTIFICATE;
const tlsPrivateKey =
  process.env.CARTULARY_WEB_E2E_TLS_PRIVATE_KEY ??
  process.env.CARTULARY_DEV_TLS_PRIVATE_KEY;
const fixtureTLS =
  tlsCertificate && tlsPrivateKey
    ? {
        cert: readFileSync(tlsCertificate),
        key: readFileSync(tlsPrivateKey),
        minVersion: "TLSv1.3" as const,
        maxVersion: "TLSv1.3" as const,
      }
    : undefined;
const backendAgent = tlsRoot
  ? new Agent({
      ca: readFileSync(tlsRoot),
      minVersion: "TLSv1.3",
      maxVersion: "TLSv1.3",
    })
  : undefined;
const e2eBackendProxy = {
  "/healthz": {
    target: e2eAPIOrigin,
    ...(backendAgent ? { agent: backendAgent } : {}),
  },
  "/readyz": {
    target: e2eAPIOrigin,
    ...(backendAgent ? { agent: backendAgent } : {}),
  },
  "/api": {
    target: e2eAPIOrigin,
    ...(backendAgent ? { agent: backendAgent } : {}),
  },
  "/ws": {
    target: e2eAPIOrigin,
    ...(backendAgent ? { agent: backendAgent } : {}),
    ws: true,
  },
};

export default defineConfig(({ mode }) => ({
  plugins: [
    react(),
    {
      name: "production-fixture-boundary",
      generateBundle() {
        if (mode === "measurement") return;
        for (const id of this.getModuleIds()) {
          if (
            id.includes("/src/measurement/") ||
            id.includes("NetworkFlowGridLoadFixture")
          ) {
            this.error(
              "Production module graph contains a measurement fixture",
            );
          }
        }
      },
    },
  ],
  build: {
    rollupOptions: {
      input:
        mode === "measurement"
          ? {
              app: path.resolve(__dirname, "index.html"),
              measurement: path.resolve(__dirname, "measurement.html"),
              presence: path.resolve(__dirname, "presence.html"),
            }
          : path.resolve(__dirname, "index.html"),
    },
  },
  server: {
    ...(fixtureTLS ? { https: fixtureTLS } : {}),
    fs: {
      allow: [path.resolve(__dirname, "..", "..")],
    },
    proxy: e2eBackendProxy,
  },
  preview: {
    ...(fixtureTLS ? { https: fixtureTLS } : {}),
    proxy: e2eBackendProxy,
  },
  test: {
    projects: [
      defineProject({
        test: {
          name: "browser-unit",
          environment: "jsdom",
          include: browserUnitIncludes,
          testTimeout: 15_000,
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
}));
