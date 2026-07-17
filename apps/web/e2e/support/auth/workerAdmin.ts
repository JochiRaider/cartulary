import {
  existsSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { join } from "node:path";

import {
  resolvePlaywrightStateDirectory,
  resolvePlaywrightStateFile,
} from "../runtime/harnessState";

export type WorkerAdminBlueprint = {
  parallelIndex: number;
  email: string;
  password: string;
  displayName: string;
};

export type WorkerAdminEntry = {
  parallel_index: number;
  user_id: string;
  email: string;
  password: string;
};

export type WorkerAdminManifest = {
  worker_admins: WorkerAdminEntry[];
};

const workerAdminManifestFilePath = resolvePlaywrightStateFile(
  "cartulary-playwright-worker-admins.json",
);

const workerAdminCleanupMarkerDirectory = resolvePlaywrightStateDirectory(
  "cartulary-playwright-worker-admin-cleanup",
);

export function buildWorkerAdminBlueprints(workerCount: number) {
  return Array.from({ length: workerCount }, (_, parallelIndex) => ({
    parallelIndex,
    email: `playwright-worker-admin-${parallelIndex}@example.test`,
    password: `PlaywrightWorker${parallelIndex}Pass!`,
    displayName: `Playwright Worker Admin ${parallelIndex}`,
  })) satisfies WorkerAdminBlueprint[];
}

export function writeWorkerAdminManifest(manifest: WorkerAdminManifest) {
  writeFileSync(
    workerAdminManifestFilePath,
    `${JSON.stringify(manifest, null, 2)}\n`,
    "utf8",
  );
}

export function loadWorkerAdminManifest() {
  if (!existsSync(workerAdminManifestFilePath)) {
    throw new Error("missing worker admin manifest");
  }
  return JSON.parse(
    readFileSync(workerAdminManifestFilePath, "utf8"),
  ) as WorkerAdminManifest;
}

export function loadWorkerAdminManifestIfPresent() {
  if (!existsSync(workerAdminManifestFilePath)) {
    return null;
  }
  return loadWorkerAdminManifest();
}

export function clearWorkerAdminSuiteState() {
  rmSync(workerAdminManifestFilePath, { force: true });
  rmSync(workerAdminCleanupMarkerDirectory, { force: true, recursive: true });
}

export function clearWorkerAdminCleanupMarkers() {
  rmSync(workerAdminCleanupMarkerDirectory, { force: true, recursive: true });
}

export function ensureWorkerAdminCleanupMarkerDirectory() {
  mkdirSync(workerAdminCleanupMarkerDirectory, { recursive: true });
}

export function workerAdminCleanupMarkerPath(parallelIndex: number) {
  return join(
    workerAdminCleanupMarkerDirectory,
    `worker-${parallelIndex}-cleaned`,
  );
}

export function markWorkerAdminCleaned(parallelIndex: number) {
  ensureWorkerAdminCleanupMarkerDirectory();
  writeFileSync(
    workerAdminCleanupMarkerPath(parallelIndex),
    "cleaned\n",
    "utf8",
  );
}

export function workerAdminNeedsJanitor(parallelIndex: number) {
  return !existsSync(workerAdminCleanupMarkerPath(parallelIndex));
}
