import type {
  APIRequestContext,
  Browser,
  BrowserContext,
  BrowserContextOptions,
  Page,
} from "@playwright/test";
import type { StorageState } from "../auth/storageState";
import {
  cookieValueFromStorageState,
  sessionCookieName,
} from "../auth/storageState";

type TrackedSessionSnapshot = {
  createdBy: string;
  email: string;
  purpose: string;
  sessionCookie: string;
  storageState: StorageState;
  userId: string;
};

export type BarrierOutcome = "released" | "cancelled";
export type OwnedBarrier = {
  readonly promise: Promise<BarrierOutcome>;
  release(): void;
  cancel(): void;
};

type CleanupOperation =
  | "context.acquire"
  | "context.drain"
  | "context.close"
  | "request.acquire"
  | "request.dispose"
  | "actor.acquire"
  | "session.acquire"
  | "feature.settle"
  | "route.acquire"
  | "route.handle"
  | "session.revoke_all"
  | "session.verify_revoked";

type Dependencies = {
  browser: Pick<Browser, "newContext">;
  initializeContext?: (context: BrowserContext) => Promise<void>;
  revokeAllSessions: (userId: string, reason: string) => Promise<void>;
  verifyRevokedSession: (snapshot: TrackedSessionSnapshot) => Promise<void>;
};

const contextScopes = new WeakMap<BrowserContext, OwnedBrowserScope>();
const requestScopes = new WeakMap<object, OwnedBrowserScope>();

export function browserScopeFor(page: Page): OwnedBrowserScope {
  const scope = contextScopes.get(page.context());
  if (!scope) throw new Error("Browser context has no resource owner");
  return scope;
}

export function requestScopeFor(
  request: object,
): OwnedBrowserScope | undefined {
  return requestScopes.get(request);
}

/** One test lifetime; the provisioner belongs to the enclosing worker. */
export class OwnedBrowserScope {
  readonly #contexts = new Set<BrowserContext>();
  readonly #requests = new Set<APIRequestContext>();
  readonly #actors = new Map<string, TrackedSessionSnapshot[]>();
  readonly #acquisitions = new Set<Promise<void>>();
  readonly #barriers = new Map<OwnedBarrier, BrowserContext | undefined>();
  readonly #settlers: {
    context?: BrowserContext;
    settle: () => Promise<boolean>;
  }[] = [];
  readonly #closing = new Map<BrowserContext, Promise<void>>();
  readonly #failures = new Map<CleanupOperation, number>();
  #admitting = true;
  #completion: Promise<void> | undefined;

  constructor(private readonly dependencies: Dependencies) {}

  registerActor(userId: string) {
    this.#assertAdmitting();
    this.#ownActor(userId);
  }

  // The admission is registered before invoking user creation. A response that
  // arrives during disposal still establishes ownership, even if its caller fails.
  acquireActor<T extends { user_id: string }>(
    create: () => Promise<T>,
  ): Promise<T> {
    return this.#acquire(create, "actor.acquire", (actor) => {
      this.#ownActor(actor.user_id);
    });
  }

  newContext(options?: BrowserContextOptions): Promise<BrowserContext> {
    // Playwright's instrumented public browser inherits effective project/test
    // options. Do not duplicate that merge or depend on private fixture fields.
    return this.#acquire(
      () => this.dependencies.browser.newContext(options),
      "context.acquire",
      async (context) => {
        this.#contexts.add(context);
        contextScopes.set(context, this);
        requestScopes.set(context.request, this);
        if (this.#admitting)
          await this.dependencies.initializeContext?.(context);
      },
    );
  }

  ownRequestContext(
    acquire: APIRequestContext | (() => Promise<APIRequestContext>),
  ): Promise<APIRequestContext> {
    return this.#acquire(
      typeof acquire === "function" ? acquire : async () => acquire,
      "request.acquire",
      (context) => {
        this.#requests.add(context);
        requestScopes.set(context, this);
      },
    );
  }

  registerSession(snapshot: Omit<TrackedSessionSnapshot, "sessionCookie">) {
    this.#assertAdmitting();
    this.#ownSession(snapshot);
  }

  acquireSession<T extends { storageState: StorageState }>(
    details: Omit<TrackedSessionSnapshot, "sessionCookie" | "storageState">,
    acquire: () => Promise<T>,
  ): Promise<T> {
    this.registerActor(details.userId);
    return this.#acquire(acquire, "session.acquire", (result) => {
      this.#ownSession({ ...details, storageState: result.storageState });
    });
  }

  #ownSession(snapshot: Omit<TrackedSessionSnapshot, "sessionCookie">) {
    const sessions = this.#ownActor(snapshot.userId);
    const sessionCookie = cookieValueFromStorageState(
      snapshot.storageState,
      sessionCookieName,
    );
    if (!sessionCookie)
      throw new Error("Cannot track a session without its cookie");
    if (
      !sessions.some((candidate) => candidate.sessionCookie === sessionCookie)
    ) {
      sessions.push({
        createdBy: snapshot.createdBy,
        email: snapshot.email,
        purpose: snapshot.purpose,
        userId: snapshot.userId,
        storageState: snapshot.storageState,
        sessionCookie,
      });
    }
  }

  barrier(context?: BrowserContext): OwnedBarrier {
    this.#assertAdmitting();
    let settle!: (outcome: BarrierOutcome) => void;
    const promise = new Promise<BarrierOutcome>((resolve) => {
      settle = resolve;
    });
    let outcome: BarrierOutcome | undefined;
    const finish = (next: BarrierOutcome) => {
      if (outcome !== undefined) return;
      outcome = next;
      this.#barriers.delete(barrier);
      settle(next);
    };
    const barrier: OwnedBarrier = {
      promise,
      release: () => finish("released"),
      cancel: () => finish("cancelled"),
    };
    this.#barriers.set(barrier, context);
    return barrier;
  }

  settleBeforeClosing(settle: () => Promise<void>, context?: BrowserContext) {
    this.#assertAdmitting();
    let completion: Promise<boolean> | undefined;
    this.#settlers.push({
      ...(context ? { context } : {}),
      settle: () => {
        completion ??= this.#attempt("feature.settle", settle);
        return completion;
      },
    });
  }

  acquireRoute(install: () => Promise<unknown>) {
    return this.#acquire(install, "route.acquire", () => {});
  }

  recordRouteFailure() {
    this.#recordFailure("route.handle");
  }

  closeContext(context: BrowserContext): Promise<void> {
    if (!this.#contexts.has(context))
      throw new Error("Cannot close an unowned context");
    const existing = this.#closing.get(context);
    if (existing) return existing;
    if (this.#completion !== undefined) return this.#completion;
    for (const [barrier, owner] of this.#barriers) {
      if (owner === context) barrier.cancel();
    }
    const completion = (async () => {
      const settled = await Promise.all(
        this.#settlers
          .filter((entry) => entry.context === context)
          .map((entry) => entry.settle()),
      );
      if (
        !(await this.#drainAndClose([context])) ||
        settled.some((value) => !value)
      )
        throw new Error("Owned context cleanup failed");
    })();
    this.#closing.set(context, completion);
    return completion;
  }

  dispose(): Promise<void> {
    if (this.#completion !== undefined) return this.#completion;
    this.#admitting = false;
    for (const barrier of this.#barriers.keys()) barrier.cancel();
    this.#completion = this.#disposeResources();
    return this.#completion;
  }

  #assertAdmitting() {
    if (!this.#admitting) throw new Error("Browser scope disposal has begun");
  }

  #ownActor(userId: string): TrackedSessionSnapshot[] {
    let sessions = this.#actors.get(userId);
    if (!sessions) {
      sessions = [];
      this.#actors.set(userId, sessions);
    }
    return sessions;
  }

  #acquire<T>(
    acquire: () => Promise<T>,
    operation: CleanupOperation,
    own: (value: T) => void | Promise<void>,
  ): Promise<T> {
    this.#assertAdmitting();
    let acquired = false;
    const result = Promise.resolve()
      .then(acquire)
      .then(async (value) => {
        await own(value);
        acquired = true;
        if (!this.#admitting)
          throw new Error("Resource acquired during browser scope disposal");
        return value;
      });
    const settlement = result.then(
      () => {},
      () => {
        if (!acquired) this.#recordFailure(operation);
      },
    );
    this.#acquisitions.add(settlement);
    void settlement.then(() => this.#acquisitions.delete(settlement));
    return result;
  }

  #recordFailure(operation: CleanupOperation) {
    this.#failures.set(operation, (this.#failures.get(operation) ?? 0) + 1);
  }

  async #attempt(operation: CleanupOperation, action: () => Promise<unknown>) {
    try {
      await action();
      return true;
    } catch {
      this.#recordFailure(operation);
      return false;
    }
  }

  async #drainAndClose(contexts: BrowserContext[]) {
    // Begin every drain before awaiting any. Fetched responses require the
    // context's request client to remain alive until all handlers have settled.
    let enumerated = true;
    const owners: (Page | BrowserContext)[] = [];
    for (const context of contexts) {
      try {
        owners.push(...context.pages());
      } catch {
        enumerated = false;
        this.#recordFailure("context.drain");
      }
      owners.push(context);
    }
    const drains = await Promise.all(
      owners.map((owner) =>
        this.#attempt("context.drain", () =>
          owner.unrouteAll({ behavior: "wait" }),
        ),
      ),
    );
    const closes = await Promise.all(
      contexts.map((context) =>
        this.#attempt("context.close", () => context.close()),
      ),
    );
    return enumerated && [...drains, ...closes].every(Boolean);
  }

  async #disposeResources() {
    await Promise.all(this.#acquisitions);
    await Promise.all(this.#settlers.map((entry) => entry.settle()));
    const remaining = [...this.#contexts].filter(
      (context) => !this.#closing.has(context),
    );
    await Promise.allSettled([
      this.#drainAndClose(remaining),
      ...this.#closing.values(),
    ]);
    for (const [userId, snapshots] of [...this.#actors].reverse()) {
      await this.#attempt("session.revoke_all", () =>
        this.dependencies.revokeAllSessions(
          userId,
          "playwright owned browser scope cleanup",
        ),
      );
      for (const snapshot of [...snapshots].reverse()) {
        await this.#attempt("session.verify_revoked", () =>
          this.dependencies.verifyRevokedSession(snapshot),
        );
      }
    }
    await Promise.all(
      [...this.#requests].map((context) =>
        this.#attempt("request.dispose", () => context.dispose()),
      ),
    );
    this.#actors.clear();
    if (this.#failures.size) {
      // Never retain dependency errors, user identifiers, URLs, or credentials.
      throw new Error(
        `Owned browser cleanup failed: ${[...this.#failures].map(([operation, count]) => `${operation}=${count}`).join(", ")}`,
      );
    }
  }
}

export function responseBarrier(page: Page): OwnedBarrier {
  return browserScopeFor(page).barrier(page.context());
}
