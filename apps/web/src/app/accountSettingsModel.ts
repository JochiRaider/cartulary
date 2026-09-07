import { clientTxnID, type HTTPOperationResult } from "../services/browserApi";
import { validateAccountDisplayName } from "./accountInputValidation";
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
import type { AccountSettingsSessionPort } from "./appSessionController";

export type AccountResourceKind = "profile" | "appearance";
type Failure =
  | "transport"
  | "timeout"
  | "contract"
  | "authorization"
  | "unavailable";
type Attempt<V> = Readonly<{
  actor: string;
  lifetime: string;
  kind: AccountResourceKind;
  baseVersion: number;
  desired: V;
  revision: number;
  clientTxnId: string;
}>;
type Operation<V> =
  | { kind: "idle" }
  | { kind: "saving"; attempt: Attempt<V>; replay: boolean }
  | { kind: "uncertain"; attempt: Attempt<V>; reason: Failure }
  | {
      kind: "rejected";
      reason: "validation" | "transaction" | "authorization" | "preparation";
    }
  | {
      kind: "conflict";
      code: "user_version_conflict" | "preferences_version_conflict";
    }
  | { kind: "confirmed"; attempt: Attempt<V> };
type ReadState = "unresolved" | "loading" | "refreshing" | "ready" | "failed";
type EditorState<R, V> = Readonly<{
  saved: R | null;
  draft: V | undefined;
  revision: number;
  baseVersion: number | null;
  followSaved: boolean;
  read: ReadState;
  readFailure: Failure | null;
  fieldError: string | null;
  operation: Operation<V>;
}>;
type Publication = "pending" | "ready" | "failed";
export type ProfileEditState = Omit<
  EditorState<AccountProfileResource, string>,
  "operation"
> & {
  readonly kind: "profile";
  readonly operation:
    | Exclude<Operation<string>, { kind: "confirmed" }>
    | { kind: "confirmed"; attempt: Attempt<string>; publication: Publication };
};
export type AppearanceEditState = EditorState<
  AccountPreferencesResource,
  DensityMode | null
> & { readonly kind: "appearance" };
export type AccountEditState = ProfileEditState | AppearanceEditState;
type AccountSettingsSnapshot = Readonly<{
  profile: ProfileEditState;
  appearance: AppearanceEditState;
}>;
type Ports = {
  readProfile: typeof loadAccountProfile;
  patchProfile: typeof patchAccountProfile;
  putPreferences: typeof putAccountPreferences;
  newTransactionId: (kind: AccountResourceKind) => string;
};
type ResourceIdentity = {
  user_id: string;
  created_at: string;
  updated_at: string;
};
type Observation<T> =
  | { kind: "completed"; value: T }
  | { kind: "cancelled" | "timeout" | "transport" };
const unresolved = (operation: { kind: string }) =>
  operation.kind === "saving" || operation.kind === "uncertain";
const failure = (code?: string, status?: number): Failure =>
  code === "invalid_public_contract_response"
    ? "contract"
    : status === 403 || code === "authorization_denied"
      ? "authorization"
      : "unavailable";
export const accountSavedValue = (state: AccountEditState) =>
  state.kind === "profile"
    ? state.saved?.display_name
    : state.saved?.density_mode;
export const accountDraftDirty = (state: AccountEditState) =>
  state.saved !== null && state.draft !== accountSavedValue(state);
export const accountReviewRequired = (state: AccountEditState) =>
  state.saved !== null &&
  accountDraftDirty(state) &&
  state.baseVersion !==
    (state.kind === "profile"
      ? state.saved.user_version
      : state.saved.preferences_version);
export function canSaveAccountEdit(state: AccountEditState) {
  return (
    state.saved !== null &&
    accountDraftDirty(state) &&
    !accountReviewRequired(state) &&
    writable(state) &&
    !(
      state.kind === "profile" &&
      state.operation.kind === "confirmed" &&
      state.operation.publication === "pending"
    )
  );
}
export function accountEditActions(state: AccountEditState) {
  const operation = state.operation;
  const recovery =
    accountReviewRequired(state) ||
    operation.kind === "conflict" ||
    (operation.kind === "rejected" &&
      (operation.reason === "transaction" ||
        operation.reason === "authorization"));
  return {
    save: canSaveAccountEdit(state),
    review: recovery && !unresolved(operation),
    reviewEnabled: state.read === "ready",
    discard: accountDraftDirty(state),
    replay: operation.kind === "uncertain",
    publicationRetry:
      state.kind === "profile" &&
      state.operation.kind === "confirmed" &&
      state.operation.publication === "failed",
  };
}
function writable(state: {
  read: ReadState;
  readFailure: Failure | null;
  operation: { kind: string; reason?: string };
}) {
  return (
    state.read !== "refreshing" &&
    state.readFailure !== "authorization" &&
    !unresolved(state.operation) &&
    state.operation.kind !== "conflict" &&
    !(
      state.operation.kind === "rejected" &&
      ["transaction", "authorization"].includes(state.operation.reason ?? "")
    )
  );
}

/** Bounded mechanics for the two scalar account editors, with typed resource policies. */
class AccountEditor<R extends ResourceIdentity, V> {
  state: EditorState<R, V> = this.initial();
  private readonly observations = new Map<string, () => void>();
  private epoch = 0;
  constructor(
    private readonly policy: {
      kind: AccountResourceKind;
      current: () => { actor: string; lifetime: string } | null;
      version: (resource: R) => number;
      value: (resource: R) => V;
      equal: (a: R, b: R) => boolean;
      validate: (value: V) => { value: V; error: string | null };
      resourceValid: (value: unknown) => value is R;
      write: (
        attempt: Attempt<V>,
        signal: AbortSignal,
      ) => Promise<
        HTTPOperationResult<{ data: R; meta: { request_id: string } }>
      >;
      newTransactionId: () => string;
      refresh: () => Promise<void>;
      sessionLost: () => void;
      confirmed: (resource: R, attempt: Attempt<V>) => boolean;
      changed: () => void;
      canSubmit: () => boolean;
    },
  ) {}
  private initial(): EditorState<R, V> {
    return {
      saved: null,
      draft: undefined,
      revision: 0,
      baseVersion: null,
      followSaved: true,
      read: "unresolved",
      readFailure: null,
      fieldError: null,
      operation: { kind: "idle" },
    };
  }
  reset() {
    ++this.epoch;
    for (const cancel of [...this.observations.values()]) cancel();
    this.state = this.initial();
  }
  update(patch: Partial<EditorState<R, V>>) {
    if (
      Object.entries(patch).every(
        ([key, value]) => this.state[key as keyof EditorState<R, V>] === value,
      )
    )
      return;
    this.state = { ...this.state, ...patch };
    this.policy.changed();
  }
  private current(epoch: number) {
    return epoch === this.epoch && this.policy.current() !== null;
  }
  private dirty() {
    return (
      this.state.saved !== null &&
      this.state.draft !== this.policy.value(this.state.saved)
    );
  }
  change(draft: V) {
    if (
      !this.policy.current() ||
      this.state.saved === null ||
      draft === this.state.draft
    )
      return;
    this.update({
      draft,
      revision: this.state.revision + 1,
      followSaved: false,
      fieldError: null,
    });
  }
  discard() {
    if (!this.policy.current()) return;
    const { saved, operation } = this.state;
    this.update({
      draft: saved === null ? undefined : this.policy.value(saved),
      baseVersion: saved === null ? null : this.policy.version(saved),
      followSaved: true,
      revision: this.state.revision + 1,
      fieldError: null,
      operation:
        unresolved(operation) || operation.kind === "confirmed"
          ? operation
          : { kind: "idle" },
    });
  }
  review() {
    const { saved, read, operation } = this.state;
    if (
      !this.policy.current() ||
      saved === null ||
      read !== "ready" ||
      unresolved(operation)
    )
      return;
    this.update({
      baseVersion: this.policy.version(saved),
      operation: { kind: "idle" },
      fieldError: null,
    });
  }
  accept(resource: R) {
    if (resource.user_id !== this.policy.current()?.actor) return false;
    const previous = this.state.saved;
    const version = this.policy.version(resource);
    if (previous !== null) {
      if (version < this.policy.version(previous)) return true;
      if (
        version === this.policy.version(previous) &&
        !this.policy.equal(previous, resource)
      )
        return false;
    }
    this.update({
      saved: resource,
      ...(this.state.followSaved
        ? { draft: this.policy.value(resource), baseVersion: version }
        : {}),
    });
    return true;
  }
  submit() {
    const identity = this.policy.current();
    const { saved, draft, baseVersion } = this.state;
    if (
      !identity ||
      saved === null ||
      draft === undefined ||
      baseVersion === null ||
      !this.dirty() ||
      baseVersion !== this.policy.version(saved) ||
      !writable(this.state) ||
      !this.policy.canSubmit()
    )
      return;
    const validated = this.policy.validate(draft);
    if (validated.error !== null) {
      this.update({
        fieldError: validated.error,
        operation: { kind: "rejected", reason: "validation" },
      });
      return;
    }
    try {
      const attempt = Object.freeze({
        ...identity,
        kind: this.policy.kind,
        baseVersion,
        desired: validated.value,
        revision: this.state.revision,
        clientTxnId: this.policy.newTransactionId(),
      });
      void this.dispatch(attempt, false);
    } catch {
      this.update({ operation: { kind: "rejected", reason: "preparation" } });
    }
  }
  replay() {
    const operation = this.state.operation;
    if (operation.kind === "uncertain")
      void this.dispatch(operation.attempt, true);
  }
  private async dispatch(attempt: Attempt<V>, replay: boolean) {
    const identity = this.policy.current();
    if (
      !identity ||
      identity.actor !== attempt.actor ||
      identity.lifetime !== attempt.lifetime ||
      this.state.operation.kind === "saving"
    )
      return;
    const epoch = this.epoch;
    this.cancel("publication");
    this.update({
      operation: { kind: "saving", attempt, replay },
      fieldError: null,
    });
    const observation = await this.observe("write", (signal) =>
      this.policy.write(attempt, signal),
    );
    if (!this.current(epoch) || observation.kind === "cancelled") return;
    const uncertain = (reason: Failure) =>
      this.update({ operation: { kind: "uncertain", attempt, reason } });
    if (observation.kind !== "completed") {
      uncertain(observation.kind);
      return;
    }
    const result = observation.value;
    if (result.ok) {
      const resource = result.payload.data;
      if (
        !this.policy.resourceValid(resource) ||
        resource.user_id !== attempt.actor ||
        this.policy.version(resource) < attempt.baseVersion ||
        this.policy.value(resource) !== attempt.desired
      ) {
        uncertain("contract");
        return;
      }
      if (this.policy.kind === "profile") this.cancel("read");
      if (!this.policy.confirmed(resource, attempt)) {
        uncertain("contract");
        return;
      }
      if (!this.current(epoch)) return;
      const latest = this.state.saved;
      const version = latest === null ? 0 : this.policy.version(latest);
      const acknowledge =
        this.state.revision === attempt.revision &&
        this.policy.version(resource) >= version;
      this.update({
        ...(acknowledge
          ? {
              draft: this.policy.value(resource),
              baseVersion: version,
              followSaved: true,
            }
          : {}),
        ...(this.policy.kind === "profile"
          ? { read: "ready" as const, readFailure: null }
          : {}),
        operation: { kind: "confirmed", attempt },
      });
      return;
    }
    const error = result.payload.error;
    if (result.status === 401 && error?.code === "session_required") {
      this.policy.sessionLost();
      return;
    }
    const conflict =
      this.policy.kind === "profile"
        ? "user_version_conflict"
        : "preferences_version_conflict";
    if (result.status === 409 && error?.code === conflict) {
      this.update({ operation: { kind: "conflict", code: conflict } });
      void this.policy.refresh();
    } else if (result.status === 409 && error?.code === "client_txn_conflict") {
      this.update({ operation: { kind: "rejected", reason: "transaction" } });
      void this.policy.refresh();
    } else if (
      result.status === 400 &&
      error?.code === "invalid_mutation_payload" &&
      !replay
    ) {
      const profile = this.policy.kind === "profile";
      this.update({
        operation: { kind: "rejected", reason: "validation" },
        fieldError:
          error.details?.field === (profile ? "display_name" : "density_mode")
            ? profile
              ? "Check the display name: use 1–256 characters without control characters."
              : "Select one of the listed density choices."
            : null,
      });
    } else if (result.status === 403 && !replay)
      this.update({
        operation: { kind: "rejected", reason: "authorization" },
        readFailure: "authorization",
      });
    else uncertain(failure(error?.code, result.status));
  }
  cancel(key: string) {
    this.observations.get(key)?.();
  }
  observe<T>(
    key: string,
    task: (signal: AbortSignal) => Promise<T>,
  ): Promise<Observation<T>> {
    this.cancel(key);
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
}

export class AccountSettingsController {
  private readonly profile: AccountEditor<AccountProfileResource, string>;
  private readonly appearance: AccountEditor<
    AccountPreferencesResource,
    DensityMode | null
  >;
  private snapshot: AccountSettingsSnapshot;
  private readonly listeners = new Set<() => void>();
  private lifetime: string | null = null;
  private actor: string | null = null;
  private epoch = 0;
  private disposed = false;
  private unsubscribe: (() => void) | null = null;
  private publication: Publication = "ready";
  private publishing: Attempt<string> | null = null;
  private readonly ports: Ports;
  constructor(
    private readonly session: AccountSettingsSessionPort,
    ports: Partial<Ports> = {},
  ) {
    this.ports = {
      readProfile: loadAccountProfile,
      patchProfile: patchAccountProfile,
      putPreferences: putAccountPreferences,
      newTransactionId: (kind) => clientTxnID(`account-${kind}`),
      ...ports,
    };
    const common = {
      current: this.current,
      sessionLost: () => session.sessionLost(),
      changed: () => this.changed(),
    };
    this.profile = new AccountEditor({
      ...common,
      kind: "profile",
      version: (r) => r.user_version,
      value: (r) => r.display_name,
      equal: (a, b) =>
        a.display_name === b.display_name &&
        a.email === b.email &&
        a.created_at === b.created_at &&
        a.updated_at === b.updated_at,
      validate: validateAccountDisplayName,
      resourceValid: isAccountProfileResource,
      write: (a, signal) =>
        this.ports.patchProfile({
          baseUserVersion: a.baseVersion,
          displayName: a.desired,
          clientTxnId: a.clientTxnId,
          signal,
        }),
      newTransactionId: () => this.ports.newTransactionId("profile"),
      refresh: () => this.refresh("profile"),
      confirmed: (r) => {
        if (!this.profile.accept(r)) return false;
        this.publication = "pending";
        return true;
      },
      canSubmit: () =>
        this.profile.state.operation.kind !== "confirmed" ||
        this.publication !== "pending",
    });
    this.appearance = new AccountEditor({
      ...common,
      kind: "appearance",
      version: (r) => r.preferences_version,
      value: (r) => r.density_mode,
      equal: (a, b) =>
        a.density_mode === b.density_mode &&
        a.created_at === b.created_at &&
        a.updated_at === b.updated_at,
      validate: (value) => ({ value, error: null }),
      resourceValid: isAccountPreferencesResource,
      write: (a, signal) =>
        this.ports.putPreferences({
          basePreferencesVersion: a.baseVersion,
          densityMode: a.desired,
          clientTxnId: a.clientTxnId,
          signal,
        }),
      newTransactionId: () => this.ports.newTransactionId("appearance"),
      refresh: () => this.refresh("appearance"),
      confirmed: (r, a) =>
        session.preferencesChanged(r, a.lifetime) !== "invalid",
      canSubmit: () => true,
    });
    this.snapshot = this.project();
  }
  private project(): AccountSettingsSnapshot {
    const state = this.profile.state;
    return {
      profile: {
        ...state,
        kind: "profile",
        operation:
          state.operation.kind === "confirmed"
            ? { ...state.operation, publication: this.publication }
            : state.operation,
      },
      appearance: { ...this.appearance.state, kind: "appearance" },
    };
  }
  private changed() {
    this.snapshot = this.project();
    for (const listener of this.listeners) listener();
    const op = this.profile.state.operation;
    if (
      op.kind === "confirmed" &&
      this.publication === "pending" &&
      this.publishing !== op.attempt
    )
      void this.publishProfile(op.attempt);
  }
  private current = () => {
    const state = this.session.getSnapshot();
    return !this.disposed &&
      this.lifetime !== null &&
      state.lifetime === this.lifetime &&
      this.actor !== null &&
      state.session?.user_id === this.actor
      ? { actor: this.actor, lifetime: this.lifetime }
      : null;
  };
  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  start = () => {
    if (this.disposed || this.unsubscribe) return;
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
    this.publishing = null;
    this.publication = "ready";
    this.profile.reset();
    this.appearance.reset();
    this.changed();
  };
  dispose = () => {
    this.stop();
    this.retireLifetime();
    this.disposed = true;
    this.listeners.clear();
  };
  private syncSession = () => {
    const state = this.session.getSnapshot();
    if (state.lifetime !== this.lifetime) {
      this.retireLifetime();
      this.lifetime = state.lifetime;
      this.actor = state.session?.user_id ?? null;
    }
    if (!this.current()) return;
    const preferences = state.preferences;
    if (preferences.kind === "ready") {
      this.appearance.accept(preferences.value);
      this.appearance.update({
        read: preferences.refreshing
          ? "refreshing"
          : preferences.refreshError
            ? "failed"
            : "ready",
        readFailure: preferences.refreshError
          ? failure(preferences.refreshError.code)
          : null,
      });
    } else if (preferences.kind === "failed")
      this.appearance.update({
        read: "failed",
        readFailure: failure(preferences.error.code),
      });
    else this.appearance.update({ read: preferences.kind, readFailure: null });
  };
  bind<K extends AccountResourceKind>(kind: K, lifetime: string | null) {
    const active = () =>
      lifetime !== null &&
      lifetime === this.lifetime &&
      this.current() !== null;
    return {
      open: () => {
        if (active()) this.open(kind);
      },
      change: (
        value: { profile: string; appearance: DensityMode | null }[K],
      ) => {
        if (active()) this.change(kind, value);
      },
      submit: () => {
        if (active()) this.submit(kind);
      },
      refresh: () => {
        if (active()) void this.refresh(kind);
      },
      discard: () => {
        if (active()) this.discard(kind);
      },
      review: () => {
        if (active()) this.review(kind);
      },
      replay: () => {
        if (active()) this.replay(kind);
      },
      retryPublication: () => {
        if (active()) this.retryPublication(kind);
      },
    };
  }
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
  change = <K extends AccountResourceKind>(
    kind: K,
    value: { profile: string; appearance: DensityMode | null }[K],
  ) => {
    if (kind === "profile" && typeof value === "string")
      this.profile.change(value);
    else if (
      kind === "appearance" &&
      (value === null ||
        value === "compact" ||
        value === "default" ||
        value === "comfortable")
    )
      this.appearance.change(value);
  };
  submit = (kind: AccountResourceKind) => this[kind].submit();
  discard = (kind: AccountResourceKind) => this[kind].discard();
  review = (kind: AccountResourceKind) => this[kind].review();
  replay = (kind: AccountResourceKind) => this[kind].replay();
  refresh = async (kind: AccountResourceKind) => {
    if (!this.current()) return;
    if (kind === "appearance") {
      await this.session.refreshPreferences();
      return;
    }
    const epoch = this.epoch;
    this.profile.update({
      read: this.profile.state.saved === null ? "loading" : "refreshing",
      readFailure: null,
    });
    const observation = await this.profile.observe("read", (signal) =>
      this.ports.readProfile({ signal }),
    );
    if (
      epoch !== this.epoch ||
      !this.current() ||
      observation.kind === "cancelled"
    )
      return;
    const result = observation.kind === "completed" ? observation.value : null;
    if (
      result?.ok &&
      isAccountProfileResource(result.payload.data) &&
      this.profile.accept(result.payload.data)
    ) {
      this.profile.update({ read: "ready", readFailure: null });
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
    this.profile.update({
      read: "failed",
      readFailure: result?.ok
        ? "contract"
        : result && !result.ok
          ? failure(result.payload.error?.code, result.status)
          : observation.kind === "timeout"
            ? "timeout"
            : "transport",
    });
  };
  retryPublication = (kind: AccountResourceKind) => {
    const operation = this.profile.state.operation;
    if (
      kind === "profile" &&
      this.current() &&
      operation.kind === "confirmed" &&
      this.publication === "failed"
    )
      void this.publishProfile(operation.attempt);
  };
  private async publishProfile(attempt: Attempt<string>) {
    const epoch = this.epoch;
    const current = () => {
      const op = this.profile.state.operation;
      return (
        epoch === this.epoch &&
        this.current() !== null &&
        op.kind === "confirmed" &&
        op.attempt === attempt
      );
    };
    if (!current()) return;
    this.publishing = attempt;
    this.publication = "pending";
    this.changed();
    const observation = await this.profile.observe("publication", (signal) =>
      this.session.observeOperationSession(attempt.lifetime, signal, current),
    );
    if (!current() || observation.kind === "cancelled") return;
    this.publication =
      observation.kind === "completed" && observation.value.kind === "accepted"
        ? "ready"
        : "failed";
    this.changed();
  }
}
