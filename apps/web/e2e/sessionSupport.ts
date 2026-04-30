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
} from "./harnessState";
import { cookieValueFromStorageState, sessionCookieName } from "./helpers";
import type { StorageState } from "./playwrightTypes";

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

export type TrackedSessionSnapshot = {
  createdBy: string;
  email: string;
  purpose: string;
  sessionCookie: string;
  storageState: StorageState;
  userId: string;
};

export type SessionTrackerDependencies = {
  label: string;
  revokeAllSessions: (userId: string, reason: string) => Promise<void>;
  verifyRevokedSession: (snapshot: TrackedSessionSnapshot) => Promise<void>;
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
    throw new Error(
      `missing worker admin manifest at ${workerAdminManifestFilePath}`,
    );
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

export class OwnedSessionTracker {
  private readonly sessionsByUser = new Map<string, TrackedSessionSnapshot[]>();

  constructor(private readonly dependencies: SessionTrackerDependencies) {}

  registerSession(snapshot: Omit<TrackedSessionSnapshot, "sessionCookie">) {
    const sessionCookie = cookieValueFromStorageState(
      snapshot.storageState,
      sessionCookieName,
    );
    if (!sessionCookie) {
      throw new Error(
        `cannot track session without ${sessionCookieName} cookie for ${snapshot.userId}`,
      );
    }

    const nextSnapshot = {
      ...snapshot,
      sessionCookie,
    } satisfies TrackedSessionSnapshot;
    const existing = this.sessionsByUser.get(snapshot.userId) ?? [];
    if (
      existing.some(
        (candidate) => candidate.sessionCookie === nextSnapshot.sessionCookie,
      )
    ) {
      return;
    }
    existing.push(nextSnapshot);
    this.sessionsByUser.set(snapshot.userId, existing);
  }

  sessionCount() {
    return [...this.sessionsByUser.values()].reduce(
      (total, snapshots) => total + snapshots.length,
      0,
    );
  }

  userCount() {
    return this.sessionsByUser.size;
  }

  async cleanup() {
    const failures: string[] = [];
    for (const [userId, snapshots] of this.sessionsByUser) {
      const email = snapshots[0]?.email ?? "unknown";
      const reason = `playwright cleanup: ${this.dependencies.label}`;
      try {
        await this.dependencies.revokeAllSessions(userId, reason);
      } catch (error) {
        failures.push(
          `revoke-all failed for ${userId} (${email}) in ${this.dependencies.label}: ${formatError(error)}`,
        );
        continue;
      }

      for (const snapshot of snapshots) {
        try {
          await this.dependencies.verifyRevokedSession(snapshot);
        } catch (error) {
          failures.push(
            `revocation verification failed for ${snapshot.userId} (${snapshot.email}) session ${snapshot.sessionCookie} [${snapshot.purpose}] via ${snapshot.createdBy}: ${formatError(error)}`,
          );
        }
      }
    }

    if (failures.length > 0) {
      throw new Error(failures.join("\n"));
    }
  }

  async newTrackedContext<TContext>(
    browser: {
      newContext: (options: {
        storageState: StorageState;
      }) => Promise<TContext>;
    },
    storageState: StorageState,
  ) {
    return browser.newContext({ storageState });
  }
}

function formatError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}
