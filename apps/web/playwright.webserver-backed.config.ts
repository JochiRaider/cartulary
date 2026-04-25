import path from "node:path";
import { fileURLToPath } from "node:url";

import { defineConfig } from "@playwright/test";

const currentDirectory = path.dirname(fileURLToPath(import.meta.url));
const configuredWorkers = Number.parseInt(
  process.env.PLAYWRIGHT_WORKERS ?? "2",
  10,
);
const publicOrigin =
  process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN ?? "http://127.0.0.1:4173";
const useExternalServer =
  process.env.CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER === "1";
const functionalGrep = process.env.CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP;

if (!functionalGrep || functionalGrep.trim() === "") {
  throw new Error("CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP is required");
}

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  globalSetup: path.resolve(currentDirectory, "e2e", "global-setup.ts"),
  globalTeardown: path.resolve(currentDirectory, "e2e", "global-teardown.ts"),
  workers:
    Number.isNaN(configuredWorkers) || configuredWorkers < 1
      ? 1
      : configuredWorkers,
  timeout: 60_000,
  use: {
    baseURL: publicOrigin,
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "functional",
      grep: new RegExp(functionalGrep),
      testMatch: [
        "phase1.spec.ts",
        "phase2.spec.ts",
        "phase3.spec.ts",
        "phase4.spec.ts",
      ],
    },
    {
      name: "support",
      testMatch: ["phase2.support.spec.ts", "phase3.support.spec.ts"],
    },
  ],
  webServer: useExternalServer
    ? undefined
    : {
        command: "bash ./scripts/start-web-e2e.sh",
        cwd: path.resolve(currentDirectory, "..", ".."),
        url: publicOrigin,
        reuseExistingServer: false,
        timeout: 180_000,
      },
});
