import {
  type APIError,
  extractError,
  type HTTPOperationResult,
} from "../services/browserApi";
import type {
  AuthorizationRecoveryPort,
  AuthorizationRecoveryResult,
} from "../shared/authorizationRecovery";
import {
  loadAccountPreferences,
  loadExtensions,
  loadSession,
} from "./api/authAccountClient";
import type {
  AccountPreferencesResource,
  ExtensionProfileResource,
  SessionData,
} from "./api/publicHttpTypes";

export type SessionResourceState<T> =
  | { readonly kind: "unresolved" | "loading" }
  | { readonly kind: "ready"; readonly value: T }
  | {
      readonly kind: "failed";
      readonly error: APIError;
      readonly failure: "transient" | "contract";
    };

export type AppSessionSnapshot = {
  readonly session: SessionData | null;
  readonly state: "unresolved" | "anonymous" | "authenticated" | "unavailable";
  readonly lifetime: string | null;
  readonly revision: number;
  readonly observing: boolean;
  readonly error: APIError | null;
  readonly ended: boolean;
  readonly preferences: SessionResourceState<AccountPreferencesResource>;
  readonly extensions: SessionResourceState<
    readonly ExtensionProfileResource[]
  >;
};

type SessionPorts = {
  readonly session: (signal: AbortSignal) => ReturnType<typeof loadSession>;
  readonly preferences: (
    signal: AbortSignal,
  ) => ReturnType<typeof loadAccountPreferences>;
  readonly extensions: (
    signal: AbortSignal,
  ) => ReturnType<typeof loadExtensions>;
  readonly retireLifetime: (nextLifetime: string | null) => void;
  readonly replaceAccount: () => void;
};
type Observation<T> =
  | { kind: "completed"; value: T }
  | { kind: "cancelled" | "timeout" | "transport" };
type SessionObservation =
  | { kind: "accepted"; session: SessionData }
  | Exclude<
      AuthorizationRecoveryResult,
      { kind: "authorized" | "access_lost" }
    >;
const observationTimeoutMs = 30_000;
const unavailableError: APIError = {
  code: "session_observation_unavailable",
  status: 503,
  retryable: true,
};

function identity(session: SessionData) {
  return `${session.user_id}:${session.authenticated_at}:${session.provider_type}`;
}
function failure(error: APIError | null): "transient" | "contract" {
  return error?.code === "invalid_public_contract_response"
    ? "contract"
    : "transient";
}

/** Application session acceptance authority. Aborting transport is only an optimization. */
export class AppSessionController {
  private snapshot: AppSessionSnapshot = {
    session: null,
    state: "unresolved",
    lifetime: null,
    revision: 0,
    observing: false,
    error: null,
    ended: false,
    preferences: { kind: "unresolved" },
    extensions: { kind: "unresolved" },
  };
  private readonly listeners = new Set<() => void>();
  private readonly observations = new Map<string, () => void>();
  private accountId: string | null = null;
  private sessionRead = 0;
  private readonly resourceReads = { preferences: 0, extensions: 0 };
  private recoveryScope = 0;
  private disposed = false;
  private readonly ports: SessionPorts;

  constructor(ports: Partial<SessionPorts> = {}) {
    this.ports = {
      session: (signal) => loadSession({ signal }),
      preferences: (signal) => loadAccountPreferences({ signal }),
      extensions: (signal) => loadExtensions({ signal }),
      retireLifetime: () => {},
      replaceAccount: () => {},
      ...ports,
    };
  }
  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };
  start() {
    if (this.disposed) return;
    if (this.snapshot.session === null) void this.refreshSession();
    else this.refreshResources();
  }
  stop() {
    ++this.sessionRead;
    ++this.recoveryScope;
    this.cancelObservations();
    if (this.snapshot.observing) this.publish({ observing: false });
  }
  dispose() {
    this.stop();
    this.disposed = true;
    this.listeners.clear();
  }
  navigationChanged() {
    ++this.recoveryScope;
  }
  authenticationCompleted(
    session: SessionData,
    expectedRevision: number,
  ): boolean {
    if (this.disposed || this.snapshot.revision !== expectedRevision)
      return false;
    ++this.sessionRead;
    this.cancelObservations();
    this.accept(session, true);
    return true;
  }
  async confirmAuthentication(expectedRevision: number): Promise<boolean> {
    if (this.snapshot.revision !== expectedRevision) return false;
    return (await this.observeSession(true)).kind === "accepted";
  }
  logoutConfirmed() {
    this.endSession();
  }
  credentialsRevoked() {
    this.endSession();
  }
  sessionLost() {
    this.endSession();
  }
  private endSession() {
    if (this.disposed) return;
    ++this.sessionRead;
    ++this.recoveryScope;
    this.cancelObservations();
    this.ports.retireLifetime(null);
    this.publish({
      session: null,
      state: "anonymous",
      lifetime: null,
      revision: this.snapshot.revision + 1,
      observing: false,
      error: null,
      ended: true,
      preferences: { kind: "unresolved" },
      extensions: { kind: "unresolved" },
    });
  }
  refreshSession = async () => {
    await this.observeSession(false);
  };
  preferencesChanged(
    value: AccountPreferencesResource,
    lifetime: string | null,
  ) {
    if (
      this.disposed ||
      lifetime === null ||
      lifetime !== this.snapshot.lifetime ||
      value.user_id !== this.snapshot.session?.user_id
    )
      return;
    ++this.resourceReads.preferences;
    this.observations.get("preferences")?.();
    this.publish({ preferences: { kind: "ready", value } });
  }
  refreshResources = () => {
    void this.refreshPreferences();
    void this.refreshExtensions();
  };
  refreshPreferences = () =>
    this.observeResource(
      "preferences",
      this.ports.preferences,
      (result) => result.data,
    );
  refreshExtensions = () =>
    this.observeResource(
      "extensions",
      this.ports.extensions,
      (result) => result.data.extensions,
    );

  recoveryPort(
    isCurrent: (incidentId: string) => boolean = () => true,
  ): AuthorizationRecoveryPort {
    return {
      recover: async ({ incidentId, signal }) => {
        const lifetime = this.snapshot.lifetime;
        const scope = this.recoveryScope;
        const current = () =>
          !signal.aborted &&
          scope === this.recoveryScope &&
          isCurrent(incidentId);
        if (!current()) return { kind: "cancelled" };
        if (lifetime === null) return { kind: "session_lost" };
        const result = await this.observeSession(
          false,
          signal,
          current,
          "caller",
        );
        if (result.kind !== "accepted") return result;
        if (!current() || lifetime !== this.snapshot.lifetime)
          return { kind: "cancelled" };
        const membership = result.session.memberships.find(
          (entry) => entry.incident_id === incidentId,
        );
        if (membership === undefined) return { kind: "access_lost" };
        return {
          kind: "authorized",
          role: membership.role,
          userId: result.session.user_id,
        };
      },
    };
  }

  private accept(session: SessionData, reauthenticated: boolean) {
    const replacement =
      reauthenticated ||
      this.snapshot.session === null ||
      identity(session) !== identity(this.snapshot.session);
    if (replacement) {
      this.cancelObservations();
      const revision = this.snapshot.revision + 1;
      const lifetime = `${revision}:${identity(session)}`;
      // These callbacks run before the new account or lifetime becomes observable.
      this.ports.retireLifetime(lifetime);
      if (this.accountId !== null && this.accountId !== session.user_id)
        this.ports.replaceAccount();
      this.accountId = session.user_id;
      this.publish({
        session,
        state: "authenticated",
        revision,
        lifetime,
        observing: false,
        error: null,
        ended: false,
        preferences: { kind: "unresolved" },
        extensions: { kind: "unresolved" },
      });
      this.refreshResources();
    } else {
      this.publish({
        session,
        state: "authenticated",
        observing: false,
        error: null,
        ended: false,
      });
    }
  }
  private async observeSession(
    reauthenticated: boolean,
    signal?: AbortSignal,
    canAccept: () => boolean = () => true,
    reporting: "shell" | "caller" = "shell",
  ): Promise<SessionObservation> {
    if (this.disposed || signal?.aborted || !canAccept())
      return { kind: "cancelled" };
    const read = ++this.sessionRead;
    const revision = this.snapshot.revision;
    this.publish({
      observing: reporting === "shell",
      ...(reporting === "shell" ? { error: null } : {}),
    });
    const observation = await this.observe(
      "session",
      this.ports.session,
      signal,
    );
    if (
      this.disposed ||
      read !== this.sessionRead ||
      revision !== this.snapshot.revision
    )
      return { kind: "cancelled" };
    if (observation.kind === "cancelled" || signal?.aborted || !canAccept()) {
      this.publish({ observing: false });
      return { kind: "cancelled" };
    }
    const result = observation.kind === "completed" ? observation.value : null;
    if (result?.ok) {
      this.accept(result.payload.data, reauthenticated);
      return { kind: "accepted", session: result.payload.data };
    }
    const error =
      result === null
        ? unavailableError
        : (extractError(result.payload) ?? unavailableError);
    if (
      result?.status === 401 &&
      (error.code === "session_required" || result.payload.error === undefined)
    ) {
      const wasAuthenticated = this.accountId !== null;
      this.endSession();
      if (!wasAuthenticated) this.publish({ ended: false });
      return { kind: "session_lost" };
    }
    if (reporting === "shell")
      this.publish({
        observing: false,
        error,
        state: this.snapshot.session === null ? "unavailable" : "authenticated",
      });
    return { kind: "unavailable", failure: failure(error) };
  }
  private async observeResource<K extends "preferences" | "extensions", T>(
    key: K,
    request: (signal: AbortSignal) => Promise<HTTPOperationResult<T>>,
    select: (
      payload: T,
    ) => K extends "preferences"
      ? AccountPreferencesResource
      : readonly ExtensionProfileResource[],
  ) {
    if (this.disposed || this.snapshot.lifetime === null) return;
    const lifetime = this.snapshot.lifetime;
    const read = ++this.resourceReads[key];
    this.publish({ [key]: { kind: "loading" } });
    const observation = await this.observe(key, request);
    if (
      this.disposed ||
      lifetime !== this.snapshot.lifetime ||
      read !== this.resourceReads[key] ||
      observation.kind === "cancelled"
    )
      return;
    const result = observation.kind === "completed" ? observation.value : null;
    if (result?.ok) {
      this.publish({ [key]: { kind: "ready", value: select(result.payload) } });
      return;
    }
    const error =
      result === null
        ? unavailableError
        : (extractError(result.payload) ?? unavailableError);
    if (result?.status === 401 && error.code === "session_required") {
      this.sessionLost();
      return;
    }
    this.publish({ [key]: { kind: "failed", error, failure: failure(error) } });
  }
  private observe<T>(
    key: string,
    task: (signal: AbortSignal) => Promise<T>,
    signal?: AbortSignal,
  ): Promise<Observation<T>> {
    this.observations.get(key)?.();
    return new Promise((resolve) => {
      const controller = new AbortController();
      let settled = false;
      const finish = (result: Observation<T>) => {
        if (settled) return;
        settled = true;
        clearTimeout(timer);
        signal?.removeEventListener("abort", cancel);
        if (this.observations.get(key) === cancel)
          this.observations.delete(key);
        resolve(result);
        controller.abort();
      };
      const cancel = () => finish({ kind: "cancelled" });
      const timer = setTimeout(
        () => finish({ kind: "timeout" }),
        observationTimeoutMs,
      );
      this.observations.set(key, cancel);
      signal?.addEventListener("abort", cancel, { once: true });
      if (signal?.aborted) {
        cancel();
        return;
      }
      void Promise.resolve().then(async () => {
        if (settled) return;
        try {
          finish({ kind: "completed", value: await task(controller.signal) });
        } catch {
          finish({ kind: "transport" });
        }
      });
    });
  }
  private cancelObservations() {
    for (const cancel of [...this.observations.values()]) cancel();
  }
  private publish(patch: Partial<AppSessionSnapshot>) {
    this.snapshot = { ...this.snapshot, ...patch };
    for (const listener of this.listeners) listener();
  }
}
