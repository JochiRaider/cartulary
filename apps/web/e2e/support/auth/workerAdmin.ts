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
import { atomicWritePrivateFile } from "../runtime/privateState";

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
  schema_id: typeof workerAdminManifestSchemaID;
  worker_admins: WorkerAdminEntry[];
};

export const workerAdminManifestSchemaID =
  "cartulary.playwright_worker_admin_manifest.v1" as const;

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
  atomicWritePrivateFile(
    workerAdminManifestFilePath,
    `${JSON.stringify(manifest, null, 2)}\n`,
  );
}

function requireExactKeys(
  value: Record<string, unknown>,
  expected: readonly string[],
  label: string,
) {
  const actual = Object.keys(value).sort();
  const wanted = [...expected].sort();
  if (JSON.stringify(actual) !== JSON.stringify(wanted)) {
    throw new Error(`${label} has an invalid field set`);
  }
}

function requireObject(value: unknown, label: string) {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value as Record<string, unknown>;
}

function requireNonEmptyString(value: unknown, label: string) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value;
}

export function parseWorkerAdminManifest(
  contents: string,
): WorkerAdminManifest {
  let decoded: unknown;
  try {
    decoded = JSON.parse(contents);
  } catch {
    throw new Error("worker admin manifest contains invalid JSON");
  }
  const manifest = requireObject(decoded, "worker admin manifest");
  requireExactKeys(
    manifest,
    ["schema_id", "worker_admins"],
    "worker admin manifest",
  );
  if (manifest.schema_id !== workerAdminManifestSchemaID) {
    throw new Error("worker admin manifest declares an unsupported schema");
  }
  if (!Array.isArray(manifest.worker_admins)) {
    throw new Error("worker admin manifest worker_admins must be an array");
  }
  const parallelIndexes = new Set<number>();
  const userIDs = new Set<string>();
  const emails = new Set<string>();
  const workerAdmins = manifest.worker_admins.map((rawEntry, index) => {
    const label = `worker admin manifest worker_admins[${index + 1}]`;
    const entry = requireObject(rawEntry, label);
    requireExactKeys(
      entry,
      ["parallel_index", "user_id", "email", "password"],
      label,
    );
    if (
      !Number.isInteger(entry.parallel_index) ||
      Number(entry.parallel_index) < 0
    ) {
      throw new Error(`${label}.parallel_index must be a non-negative integer`);
    }
    const parallelIndex = Number(entry.parallel_index);
    const userID = requireNonEmptyString(entry.user_id, `${label}.user_id`);
    const email = requireNonEmptyString(entry.email, `${label}.email`);
    const password = requireNonEmptyString(entry.password, `${label}.password`);
    if (parallelIndexes.has(parallelIndex)) {
      throw new Error(`${label} has a duplicate parallel_index`);
    }
    if (userIDs.has(userID)) {
      throw new Error(`${label} has a duplicate user_id`);
    }
    if (emails.has(email)) {
      throw new Error(`${label} has a duplicate email`);
    }
    parallelIndexes.add(parallelIndex);
    userIDs.add(userID);
    emails.add(email);
    return {
      parallel_index: parallelIndex,
      user_id: userID,
      email,
      password,
    };
  });
  return {
    schema_id: workerAdminManifestSchemaID,
    worker_admins: workerAdmins,
  };
}

export function loadWorkerAdminManifest() {
  if (!existsSync(workerAdminManifestFilePath)) {
    throw new Error("missing worker admin manifest");
  }
  return parseWorkerAdminManifest(
    readFileSync(workerAdminManifestFilePath, "utf8"),
  );
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

function workerAdminCleanupMarkerPath(parallelIndex: number) {
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
