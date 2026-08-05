import type { StorageState } from "../auth/storageState";
import {
  cookieValueFromStorageState,
  sessionCookieName,
} from "../auth/storageState";

export type TrackedSessionSnapshot = {
  createdBy: string;
  email: string;
  purpose: string;
  sessionCookie: string;
  storageState: StorageState;
  userId: string;
};

type SessionTrackerDependencies = {
  label: string;
  revokeAllSessions: (userId: string, reason: string) => Promise<void>;
  verifyRevokedSession: (snapshot: TrackedSessionSnapshot) => Promise<void>;
};

type SessionOwnership = "borrowed" | "owned";

type RegisteredUserSessions = {
  readonly snapshots: TrackedSessionSnapshot[];
  readonly userId: string;
};

export class OwnedSessionTracker {
  readonly #sessionsByUser = new Map<string, RegisteredUserSessions>();
  readonly #usersInAcquisitionOrder: RegisteredUserSessions[] = [];
  #cleanupConsumed = false;

  constructor(private readonly dependencies: SessionTrackerDependencies) {}

  registerSession(
    snapshot: Omit<TrackedSessionSnapshot, "sessionCookie">,
    ownership: SessionOwnership = "owned",
  ) {
    if (this.#cleanupConsumed) {
      throw new Error("cannot register a session after cleanup was consumed");
    }
    if (ownership === "borrowed") {
      return;
    }
    const sessionCookie = cookieValueFromStorageState(
      snapshot.storageState,
      sessionCookieName,
    );
    if (!sessionCookie) {
      throw new Error(
        `cannot track session without ${sessionCookieName} cookie`,
      );
    }

    const nextSnapshot = {
      ...snapshot,
      sessionCookie,
    } satisfies TrackedSessionSnapshot;
    let registration = this.#sessionsByUser.get(snapshot.userId);
    if (registration === undefined) {
      registration = { snapshots: [], userId: snapshot.userId };
      this.#sessionsByUser.set(snapshot.userId, registration);
      this.#usersInAcquisitionOrder.push(registration);
    }
    if (
      registration.snapshots.some(
        (candidate) => candidate.sessionCookie === sessionCookie,
      )
    ) {
      return;
    }
    registration.snapshots.push(nextSnapshot);
  }

  sessionCount() {
    return this.#usersInAcquisitionOrder.reduce(
      (total, registration) => total + registration.snapshots.length,
      0,
    );
  }

  userCount() {
    return this.#usersInAcquisitionOrder.length;
  }

  async cleanup() {
    if (this.#cleanupConsumed) {
      return;
    }
    this.#cleanupConsumed = true;
    const failedOperations: string[] = [];
    const reason = `playwright cleanup: ${this.dependencies.label}`;

    for (const registration of [...this.#usersInAcquisitionOrder].reverse()) {
      try {
        await this.dependencies.revokeAllSessions(registration.userId, reason);
      } catch {
        failedOperations.push("session.revoke_all");
        continue;
      }

      for (const snapshot of [...registration.snapshots].reverse()) {
        try {
          await this.dependencies.verifyRevokedSession(snapshot);
        } catch {
          failedOperations.push("session.verify_revoked");
        }
      }
    }

    this.#sessionsByUser.clear();
    this.#usersInAcquisitionOrder.length = 0;
    if (failedOperations.length > 0) {
      throw new Error(
        `Owned session cleanup failed (${failedOperations.length} operations): ${failedOperations.join(", ")}`,
      );
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
