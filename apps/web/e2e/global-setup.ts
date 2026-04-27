import { mkdirSync, rmSync, writeFileSync } from "node:fs";
import type { FullConfig } from "@playwright/test";
import { prepareWorkerAdminSuite } from "./authRuntime";
import {
  isExternalServerHarnessMode,
  resolvePlaywrightStateDirectory,
  usesSharedPlaywrightState,
} from "./harnessState";
import { clearSuiteAdminTotpSecret, prepareSuiteAdminState } from "./helpers";
import { clearWorkerAdminSuiteState } from "./sessionSupport";

export default async function globalSetup(config: FullConfig) {
  if (!isExternalServerHarnessMode()) {
    clearWorkerAdminSuiteState();
    clearSuiteAdminTotpSecret();
  }
  const configuredWorkerCount = positiveIntegerEnv(
    "CARTULARY_PLAYWRIGHT_WORKER_COUNT",
  );
  const workerCount = Math.max(
    typeof config.workers === "number" ? config.workers : 1,
    configuredWorkerCount ?? 1,
  );
  await withSharedGlobalSetupLock(async () => {
    await prepareSuiteAdminState();
    await prepareWorkerAdminSuite(workerCount);
  });
}

function positiveIntegerEnv(name: string) {
  const value = process.env[name];
  if (value === undefined || value.trim() === "") {
    return null;
  }
  const parsed = Number.parseInt(value, 10);
  if (!Number.isInteger(parsed) || parsed < 1 || String(parsed) !== value) {
    throw new Error(`${name} must be a positive integer`);
  }
  return parsed;
}

async function withSharedGlobalSetupLock<T>(callback: () => Promise<T>) {
  if (!usesSharedPlaywrightState()) {
    return callback();
  }

  const lockDirectory = resolvePlaywrightStateDirectory(
    "cartulary-playwright-global-setup.lock",
  );
  const deadline = Date.now() + 120_000;
  while (true) {
    try {
      mkdirSync(lockDirectory);
      writeFileSync(`${lockDirectory}/owner`, `${process.pid}\n`, "utf8");
      break;
    } catch (error) {
      if ((error as NodeJS.ErrnoException).code !== "EEXIST") {
        throw error;
      }
      if (Date.now() >= deadline) {
        throw new Error(
          `timed out waiting for Playwright global setup lock at ${lockDirectory}`,
        );
      }
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
  }

  try {
    return await callback();
  } finally {
    rmSync(lockDirectory, { force: true, recursive: true });
  }
}
