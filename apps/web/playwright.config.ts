import path from "node:path";
import { fileURLToPath } from "node:url";

import { defineConfig } from "@playwright/test";

const currentDirectory = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  workers: 1,
  timeout: 60_000,
  use: {
    baseURL: "http://127.0.0.1:4173",
    trace: "retain-on-failure",
  },
  webServer: {
    command: "bash ./scripts/start-web-e2e.sh",
    cwd: path.resolve(currentDirectory, "..", ".."),
    url: "http://127.0.0.1:4173",
    reuseExistingServer: false,
    timeout: 180_000,
  },
});
