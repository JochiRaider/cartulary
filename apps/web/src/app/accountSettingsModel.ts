import { clientTxnID, type HTTPOperationResult } from "../services/browserApi";
import {
  isAccountPreferencesResource,
  isAccountProfileResource,
  loadAccountProfile,
  patchAccountProfile,
  putAccountPreferences,
} from "./api/authAccountClient";
import type {
  AccountPreferencesResource,
  AccountProfileResource,
  DensityMode,
} from "./api/publicHttpTypes";
import type { AppSessionController } from "./appSessionController";

export type AccountResourceKind = "profile" | "appearance";
type Resource = AccountProfileResource | AccountPreferencesResource;
type Value = string | null;
type Failure =
  | "transport"
  | "timeout"
  | "contract"
  | "authorization"
  | "unavailable";
export type AccountAttempt = Readonly<{
  actor: string;
  lifetime: string;
  kind: AccountResourceKind;
  baseVersion: number;
  desired: Value;
  revision: number;
  clientTxnId: string;
}>;
export type AccountOperation =
  | { kind: "idle" }
  | { kind: "saving"; attempt: AccountAttempt; replay: boolean }
  | { kind: "uncertain"; attempt: AccountAttempt; reason: Failure }
  | {
      kind: "rejected";
      reason: "validation" | "transaction" | "authorization" | "preparation";
    }
  | {
      kind: "conflict";
      code: "user_version_conflict" | "preferences_version_conflict";
    }
  | {
      kind: "confirmed";
      attempt: AccountAttempt;
      resource: Resource;
      publication: "pending" | "ready" | "failed";
    };
export type AccountEditState = {
  readonly saved: Resource | null;
  readonly draft: Value;
  readonly revision: number;
  readonly baseVersion: number | null;
  readonly followSaved: boolean;
  readonly read: "unresolved" | "loading" | "refreshing" | "ready" | "failed";
  readonly readFailure: Failure | null;
  readonly fieldError: string | null;
  readonly operation: AccountOperation;
};
export type AccountSettingsSnapshot = Readonly<
  Record<AccountResourceKind, AccountEditState>
>;
type Ports = {
  readProfile: typeof loadAccountProfile;
  patchProfile: typeof patchAccountProfile;
  putPreferences: typeof putAccountPreferences;
  newTransactionId: (kind: AccountResourceKind) => string;
};
type Observation<T> =
  | { kind: "completed"; value: T }
  | { kind: "cancelled" | "timeout" | "transport" };
const initial = (): AccountEditState => ({
  saved: null,
  draft: null,
  revision: 0,
  baseVersion: null,
  followSaved: true,
  read: "unresolved",
  readFailure: null,
  fieldError: null,
  operation: { kind: "idle" },
});
const initialSnapshot = (): AccountSettingsSnapshot => ({
  profile: initial(),
  appearance: initial(),
});
export const accountSavedValue = (resource: Resource | null): Value =>
  resource === null
    ? null
    : "user_version" in resource
      ? resource.display_name
      : resource.density_mode;
export const accountResourceVersion = (resource: Resource): number =>
  "user_version" in resource
    ? resource.user_version
    : resource.preferences_version;
export const accountDraftDirty = (state: AccountEditState) =>
  state.saved !== null && state.draft !== accountSavedValue(state.saved);
export const accountReviewRequired = (state: AccountEditState) =>
  state.saved !== null &&
  accountDraftDirty(state) &&
  state.baseVersion !== accountResourceVersion(state.saved);
const unresolved = (operation: AccountOperation) =>
  operation.kind === "saving" || operation.kind === "uncertain";
export function canSaveAccountEdit(state: AccountEditState) {
  return (
    state.saved !== null &&
    accountDraftDirty(state) &&
    !accountReviewRequired(state) &&
    state.read !== "refreshing" &&
    state.readFailure !== "authorization" &&
    !unresolved(state.operation) &&
    state.operation.kind !== "conflict" &&
    !(
      state.operation.kind === "rejected" &&
      ["transaction", "authorization"].includes(state.operation.reason)
    ) &&
    !(
      state.operation.kind === "confirmed" &&
      state.operation.publication === "pending"
    )
  );
}

export function validateAccountDisplayName(input: string): {
  value: string;
  error: string | null;
} {
  const value = input
    .normalize("NFC")
    .replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");
  const scalars = Array.from(value);
  const invalid = scalars.some((scalar) => {
    const point = scalar.codePointAt(0) ?? 0;
    return (
      point <= 31 ||
      (point >= 127 && point <= 159) ||
      (point >= 0xd800 && point <= 0xdfff)
    );
  });
  return {
    value,
    error:
      value === ""
        ? "Enter a display name."
        : invalid
          ? "Remove control characters from the display name."
          : scalars.length > 256
            ? "Use 256 characters or fewer."
            : null,
  };
}

/** Two account-local scalar edits. The session controller remains the preferences authority. */
export class AccountSettingsController {
  private snapshot = initialSnapshot();
  private readonly listeners = new Set<() => void>();
  private readonly observations = new Map<string, () => void>();
  private unsubscribe: (() => void) | null = null;
  private lifetime: string | null = null;
  private actor: string | null = null;
  private epoch = 0;
  private disposed = false;
  private readonly ports: Ports;
  constructor(
    private readonly session: AppSessionController,
    ports: Partial<Ports> = {},
  ) {
    this.ports = {
      readProfile: loadAccountProfile,
      patchProfile: patchAccountProfile,
      putPreferences: putAccountPreferences,
      newTransactionId: (kind) => clientTxnID(`account-${kind}`),
      ...ports,
    };
  }
  getSnapshot = () => this.snapshot;
  bind(kind: AccountResourceKind, lifetime: string | null) {
    const current = () =>
      lifetime !== null && lifetime === this.lifetime && this.current();
    return {
      open: () => {
        if (current()) this.open(kind);
      },
      change: (value: Value) => {
        if (current()) this.change(kind, value);
      },
      submit: () => {
        if (current()) this.submit(kind);
      },
      refresh: () => {
        if (current()) void this.refresh(kind);
      },
      discard: () => {
        if (current()) this.discard(kind);
      },
      review: () => {
        if (current()) this.review(kind);
      },
      replay: () => {
        if (current()) this.replay(kind);
      },
      retryPublication: () => {
        if (current()) this.retryPublication(kind);
      },
    };
  }
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  start = () => {
    if (this.disposed || this.unsubscribe !== null) return;
    this.unsubscribe = this.session.subscribe(this.syncSession);
    this.syncSession();
  };
  stop = () => {
    this.unsubscribe?.();
    this.unsubscribe = null;
  };
  retireLifetime = () => {
    ++this.epoch;
    this.lifetime = null;
    this.actor = null;
    for (const cancel of [...this.observations.values()]) cancel();
    this.snapshot = initialSnapshot();
    this.emit();
  };
  dispose = () => {
    this.stop();
    this.retireLifetime();
    this.disposed = true;
    this.listeners.clear();
  };
  private current(epoch = this.epoch) {
    const current = this.session.getSnapshot();
    return (
      !this.disposed &&
      epoch === this.epoch &&
      this.lifetime !== null &&
      current.lifetime === this.lifetime &&
      current.session?.user_id === this.actor
    );
  }
  private syncSession = () => {
    const current = this.session.getSnapshot();
    if (current.lifetime !== this.lifetime) {
      this.retireLifetime();
      this.lifetime = current.lifetime;
      this.actor = current.session?.user_id ?? null;
    }
    if (!this.current()) return;
    const preferences = current.preferences;
    if (preferences.kind === "ready") {
      this.acceptResource("appearance", preferences.value);
      this.update("appearance", {
        read: preferences.refreshing
          ? "refreshing"
          : preferences.refreshError
            ? "failed"
            : "ready",
        readFailure: preferences.refreshError
          ? this.failure(preferences.refreshError.code)
          : null,
      });
    } else if (preferences.kind === "failed") {
      this.update("appearance", {
        read: "failed",
        readFailure: this.failure(preferences.error.code),
      });
    } else
      this.update("appearance", { read: preferences.kind, readFailure: null });
  };
  open = (kind: AccountResourceKind) => {
    this.syncSession();
    if (!this.current()) return;
    const state = this.snapshot[kind];
    if (
      state.read === "unresolved" ||
      state.read === "ready" ||
      (state.read === "failed" && state.saved !== null)
    )
      void this.refresh(kind);
  };
  change = (kind: AccountResourceKind, draft: Value) => {
    if (!this.current() || this.snapshot[kind].saved === null) return;
    if (kind === "profile" && typeof draft !== "string") return;
    if (
      kind === "appearance" &&
      draft !== null &&
      !["compact", "default", "comfortable"].includes(draft)
    )
      return;
    const state = this.snapshot[kind];
    if (draft === state.draft) return;
    this.update(kind, {
      draft,
      revision: state.revision + 1,
      followSaved: false,
      fieldError: null,
    });
  };
  discard = (kind: AccountResourceKind) => {
    if (!this.current()) return;
    const state = this.snapshot[kind];
    this.update(kind, {
      draft: accountSavedValue(state.saved),
      baseVersion:
        state.saved === null ? null : accountResourceVersion(state.saved),
      followSaved: true,
      revision: state.revision + 1,
      fieldError: null,
      operation:
        unresolved(state.operation) || state.operation.kind === "confirmed"
          ? state.operation
          : { kind: "idle" },
    });
  };
  review = (kind: AccountResourceKind) => {
    if (!this.current()) return;
    const state = this.snapshot[kind];
    if (
      state.saved === null ||
      state.read !== "ready" ||
      unresolved(state.operation)
    )
      return;
    this.update(kind, {
      baseVersion: accountResourceVersion(state.saved),
      operation: { kind: "idle" },
      fieldError: null,
    });
  };
  refresh = async (kind: AccountResourceKind) => {
    if (!this.current()) return;
    if (kind === "appearance") {
      await this.session.refreshPreferences();
      return;
    }
    const epoch = this.epoch;
    this.update(kind, {
      read: this.snapshot[kind].saved === null ? "loading" : "refreshing",
      readFailure: null,
    });
    const observation = await this.observe("profile-read", (signal) =>
      this.ports.readProfile({ signal }),
    );
    if (!this.current(epoch) || observation.kind === "cancelled") return;
    const result = observation.kind === "completed" ? observation.value : null;
    if (
      result?.ok &&
      isAccountProfileResource(result.payload.data) &&
      this.acceptResource(kind, result.payload.data)
    ) {
      this.update(kind, { read: "ready", readFailure: null });
      return;
    }
    if (
      result &&
      !result.ok &&
      result.status === 401 &&
      result.payload.error?.code === "session_required"
    ) {
      this.session.sessionLost();
      return;
    }
    this.update(kind, {
      read: "failed",
      readFailure: result?.ok
        ? "contract"
        : result && !result.ok
          ? this.failure(result.payload.error?.code, result.status)
          : observation.kind === "timeout"
            ? "timeout"
            : "transport",
    });
  };
  submit = (kind: AccountResourceKind) => {
    if (!this.current()) return;
    const state = this.snapshot[kind];
    if (
      !canSaveAccountEdit(state) ||
      state.baseVersion === null ||
      this.actor === null ||
      this.lifetime === null
    )
      return;
    const validated =
      kind === "profile"
        ? validateAccountDisplayName(state.draft ?? "")
        : { value: state.draft, error: null };
    if (validated.error !== null) {
      this.update(kind, {
        fieldError: validated.error,
        operation: { kind: "rejected", reason: "validation" },
      });
      return;
    }
    try {
      const attempt = Object.freeze({
        actor: this.actor,
        lifetime: this.lifetime,
        kind,
        baseVersion: state.baseVersion,
        desired: validated.value,
        revision: state.revision,
        clientTxnId: this.ports.newTransactionId(kind),
      });
      void this.dispatch(attempt, false);
    } catch {
      this.update(kind, {
        operation: { kind: "rejected", reason: "preparation" },
      });
    }
  };
  replay = (kind: AccountResourceKind) => {
    if (!this.current()) return;
    const operation = this.snapshot[kind].operation;
    if (operation.kind === "uncertain")
      void this.dispatch(operation.attempt, true);
  };
  private async dispatch(attempt: AccountAttempt, replay: boolean) {
    const { kind } = attempt;
    if (
      !this.current() ||
      attempt.lifetime !== this.lifetime ||
      attempt.actor !== this.actor ||
      this.snapshot[kind].operation.kind === "saving"
    )
      return;
    const epoch = this.epoch;
    this.observations.get(`${kind}-publication`)?.();
    this.update(kind, {
      operation: { kind: "saving", attempt, replay },
      fieldError: null,
    });
    const observation = await this.observe(
      `${kind}-write`,
      (
        signal,
      ): Promise<
        HTTPOperationResult<{ data: Resource; meta: { request_id: string } }>
      > =>
        kind === "profile"
          ? this.ports.patchProfile({
              baseUserVersion: attempt.baseVersion,
              displayName: attempt.desired ?? "",
              clientTxnId: attempt.clientTxnId,
              signal,
            })
          : this.ports.putPreferences({
              basePreferencesVersion: attempt.baseVersion,
              densityMode: attempt.desired as DensityMode | null,
              clientTxnId: attempt.clientTxnId,
              signal,
            }),
    );
    if (!this.current(epoch) || observation.kind === "cancelled") return;
    const uncertain = (reason: Failure) =>
      this.update(kind, { operation: { kind: "uncertain", attempt, reason } });
    if (observation.kind !== "completed") {
      uncertain(observation.kind);
      return;
    }
    const result = observation.value;
    if (result.ok) {
      const resource = result.payload.data;
      if (
        !(kind === "profile"
          ? isAccountProfileResource(resource)
          : isAccountPreferencesResource(resource)) ||
        resource.user_id !== attempt.actor ||
        accountResourceVersion(resource) < attempt.baseVersion ||
        accountSavedValue(resource) !== attempt.desired
      ) {
        uncertain("contract");
        return;
      }
      // Invalidating an older read is separate from accepting a replay's older version.
      if (kind === "profile") this.observations.get("profile-read")?.();
      if (kind === "appearance") {
        const accepted = this.session.preferencesChanged(
          resource as AccountPreferencesResource,
          attempt.lifetime,
        );
        if (accepted === "invalid") {
          uncertain("contract");
          return;
        }
      } else if (!this.acceptResource(kind, resource)) {
        uncertain("contract");
        return;
      }
      if (!this.current(epoch)) return;
      const state = this.snapshot[kind];
      const latestVersion =
        state.saved === null ? 0 : accountResourceVersion(state.saved);
      const acknowledge =
        state.revision === attempt.revision &&
        accountResourceVersion(resource) >= latestVersion;
      this.update(kind, {
        ...(acknowledge
          ? {
              draft: accountSavedValue(resource),
              baseVersion: latestVersion,
              followSaved: true,
            }
          : {}),
        // An obsolete Appearance replay must leave a newer session read visible.
        ...(kind === "profile"
          ? { read: "ready" as const, readFailure: null }
          : {}),
        operation: {
          kind: "confirmed",
          attempt,
          resource,
          publication: kind === "profile" ? "pending" : "ready",
        },
      });
      if (kind === "profile") void this.publishProfile(attempt);
      return;
    }
    const error = result.payload.error;
    if (result.status === 401 && error?.code === "session_required") {
      this.session.sessionLost();
      return;
    }
    const conflictCode =
      kind === "profile"
        ? "user_version_conflict"
        : "preferences_version_conflict";
    if (result.status === 409 && error?.code === conflictCode) {
      this.update(kind, {
        operation: { kind: "conflict", code: conflictCode },
      });
      void this.refresh(kind);
    } else if (result.status === 409 && error?.code === "client_txn_conflict") {
      this.update(kind, {
        operation: { kind: "rejected", reason: "transaction" },
      });
      void this.refresh(kind);
    } else if (
      result.status === 400 &&
      error?.code === "invalid_mutation_payload" &&
      !replay
    ) {
      this.update(kind, {
        operation: { kind: "rejected", reason: "validation" },
        fieldError:
          error.details?.field ===
          (kind === "profile" ? "display_name" : "density_mode")
            ? kind === "profile"
              ? "Check the display name: use 1–256 characters without control characters."
              : "Select one of the listed density choices."
            : null,
      });
    } else if (result.status === 403 && !replay) {
      this.update(kind, {
        operation: { kind: "rejected", reason: "authorization" },
        readFailure: "authorization",
      });
    } else uncertain(this.failure(error?.code, result.status));
  }
  retryPublication = (kind: AccountResourceKind) => {
    const operation = this.snapshot[kind].operation;
    if (
      this.current() &&
      kind === "profile" &&
      operation.kind === "confirmed" &&
      operation.publication === "failed"
    )
      void this.publishProfile(operation.attempt);
  };
  private async publishProfile(attempt: AccountAttempt) {
    const epoch = this.epoch;
    const isCurrent = () => {
      const operation = this.snapshot.profile.operation;
      return (
        this.current(epoch) &&
        operation.kind === "confirmed" &&
        operation.attempt === attempt
      );
    };
    if (!isCurrent()) return;
    const operation = this.snapshot.profile.operation;
    if (operation.kind !== "confirmed") return;
    this.update("profile", {
      operation: { ...operation, publication: "pending" },
    });
    const observation = await this.observe("profile-publication", (signal) =>
      this.session.refreshSessionForAccountEdit(
        attempt.lifetime,
        signal,
        isCurrent,
      ),
    );
    if (!isCurrent() || observation.kind === "cancelled") return;
    const result = observation.kind === "completed" ? observation.value : null;
    // Session reads may already contain a later, externally accepted display name.
    const ready = result?.kind === "accepted";
    this.update("profile", {
      operation: { ...operation, publication: ready ? "ready" : "failed" },
    });
  }
  private acceptResource(kind: AccountResourceKind, resource: Resource) {
    if (resource.user_id !== this.actor) return false;
    const state = this.snapshot[kind];
    const version = accountResourceVersion(resource);
    if (state.saved !== null) {
      const priorVersion = accountResourceVersion(state.saved);
      if (version < priorVersion) return true;
      if (
        version === priorVersion &&
        (accountSavedValue(resource) !== accountSavedValue(state.saved) ||
          resource.created_at !== state.saved.created_at ||
          resource.updated_at !== state.saved.updated_at ||
          ("email" in resource &&
            "email" in state.saved &&
            resource.email !== state.saved.email))
      )
        return false;
    }
    this.update(kind, {
      saved: resource,
      ...(state.followSaved
        ? { draft: accountSavedValue(resource), baseVersion: version }
        : {}),
    });
    return true;
  }
  private failure(code?: string, status?: number): Failure {
    return code === "invalid_public_contract_response"
      ? "contract"
      : status === 403 || code === "authorization_denied"
        ? "authorization"
        : "unavailable";
  }
  private observe<T>(
    key: string,
    task: (signal: AbortSignal) => Promise<T>,
  ): Promise<Observation<T>> {
    this.observations.get(key)?.();
    return new Promise((resolve) => {
      const controller = new AbortController();
      let settled = false;
      const finish = (result: Observation<T>) => {
        if (settled) return;
        settled = true;
        clearTimeout(timer);
        if (this.observations.get(key) === cancel)
          this.observations.delete(key);
        controller.abort();
        resolve(result);
      };
      const cancel = () => finish({ kind: "cancelled" });
      const timer = setTimeout(() => finish({ kind: "timeout" }), 30_000);
      this.observations.set(key, cancel);
      // Dispatch starts synchronously, after the operation lock was published.
      try {
        void task(controller.signal).then(
          (value) => finish({ kind: "completed", value }),
          () => finish({ kind: "transport" }),
        );
      } catch {
        finish({ kind: "transport" });
      }
    });
  }
  private update(kind: AccountResourceKind, patch: Partial<AccountEditState>) {
    const previous = this.snapshot[kind];
    if (
      Object.entries(patch).every(
        ([key, value]) => previous[key as keyof AccountEditState] === value,
      )
    )
      return;
    this.snapshot = { ...this.snapshot, [kind]: { ...previous, ...patch } };
    this.emit();
  }
  private emit() {
    for (const listener of this.listeners) listener();
  }
}
