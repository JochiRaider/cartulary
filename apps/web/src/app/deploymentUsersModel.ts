import { useLayoutEffect, useSyncExternalStore } from "react";
import {
  type APIError,
  clientTxnID,
  type HTTPOperationResult,
} from "../services/browserApi";
import {
  validateAccountDisplayName,
  validateAccountEmail,
  validateProvisioningPassword,
} from "./accountInputValidation";
import {
  accountOperationError,
  observeAccountOperation,
} from "./accountOperation";
import { listEnterpriseAuthProviders } from "./api/authAccountClient";
import type { DeploymentUserChanges } from "./api/deploymentUserClient";
import * as userApi from "./api/deploymentUserClient";
import type {
  EnterpriseAuthProvider,
  UserResource,
} from "./api/publicHttpTypes";

type Identity = Readonly<{ actor: string; lifetime: string; admin: boolean }>;
export type DeploymentUsersSessionPort = {
  identity: () => Identity | null;
  subscribe: (listener: () => void) => () => void;
  refresh: (
    identity: Identity,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<void> | void;
  revoked: (identity: Identity, message: string) => void;
  lost: (identity: Identity) => void;
};
export type DeploymentUsersPanelProps = {
  controller: DeploymentUsersController;
  autoLoadUsers?: boolean;
  enterpriseAuthClaimed?: boolean;
  active?: boolean;
};
type Binding = UserResource["auth_bindings"][number];
export function isEnterpriseAuthBinding(
  binding: Binding,
): binding is Extract<Binding, { provider_type: "oidc" | "saml" }> {
  return binding.provider_type !== "local";
}
type Query = Readonly<{
  search: string;
  isActive: boolean | null;
  isAdmin: boolean | null;
}>;
type MutableUser = Required<DeploymentUserChanges>;
type Draft = Readonly<{
  changes: DeploymentUserChanges;
  baseVersion: number;
  revision: number;
}>;
type CredentialDialog = "password" | "revoke" | "totp";
type Action =
  | "create"
  | "patch"
  | "password"
  | "totp"
  | "revoke"
  | "bindingCreate"
  | "bindingRotate"
  | "bindingRetire";
type RequestInput<F extends (...args: never[]) => unknown> = Readonly<
  Omit<Parameters<F>[0], "signal" | "apiBase">
>;
type ReplayAttempt =
  | {
      readonly action: "totp";
      readonly input: RequestInput<typeof userApi.adminResetTotp>;
    }
  | {
      readonly action: "revoke";
      readonly input: RequestInput<typeof userApi.adminRevokeAllSessions>;
    }
  | {
      readonly action: "bindingCreate";
      readonly input: RequestInput<typeof userApi.createEnterpriseAuthBinding>;
    }
  | {
      readonly action: "bindingRotate";
      readonly input: RequestInput<typeof userApi.rotateEnterpriseAuthBinding>;
    }
  | {
      readonly action: "bindingRetire";
      readonly input: RequestInput<typeof userApi.retireEnterpriseAuthBinding>;
    };
type Intent = Readonly<{
  identity: Identity;
  action: Action;
  target: string | null;
  label: string;
  selection: number;
  form: number;
  draftRevision: number;
  baseVersion: number | null;
}>;
type Operation =
  | { kind: "idle" }
  | {
      kind: "pending" | "uncertain";
      intent: Intent;
      attempt: ReplayAttempt | null;
    }
  | { kind: "rejected"; intent: Intent; review: boolean }
  | {
      kind: "confirmed";
      intent: Intent;
      propagation: "pending" | "ready" | "failed";
    };
type MutationReply =
  | { ok: true; userId: string; user: UserResource | null }
  | { ok: false; status: number; error: APIError | null };
type CreateDraft = {
  email: string;
  displayName: string;
  password: string;
  mfaRequired: boolean;
  admin: boolean;
};
type BindingDraft = {
  providerKey: string;
  subject: string;
  target: string;
  newSubject: string;
  reason: string;
};
type State = {
  authorized: boolean;
  statusText: string;
  error: APIError | null;
  queryDraft: Query;
  acceptedQuery: Query | null;
  queryStatus: "idle" | "pending" | "loading" | "failed";
  users: UserResource[];
  paging: { next: string | null; hasMore: boolean; invalid: boolean };
  pagePending: boolean;
  selected: UserResource | null;
  draft: Draft | null;
  targetStatus: "idle" | "loading" | "failed";
  create: CreateDraft;
  createOpen: boolean;
  credentialDialog: CredentialDialog | null;
  password: string;
  reason: string;
  binding: BindingDraft;
  enterpriseClaimed: boolean;
  providers: EnterpriseAuthProvider[];
  providersStatus: "idle" | "loading" | "ready" | "failed";
  operation: Operation;
  operationObserved: boolean;
  transportPending: boolean;
  leavePrompt: boolean;
  fieldErrors: Record<string, string>;
};
const emptyQuery = (): Query => ({ search: "", isActive: null, isAdmin: null });
const emptyCreate = (): CreateDraft => ({
  email: "",
  displayName: "",
  password: "",
  mfaRequired: true,
  admin: false,
});
const emptyBinding = (): BindingDraft => ({
  providerKey: "",
  subject: "",
  target: "",
  newSubject: "",
  reason: "",
});
const initial = (): State => ({
  authorized: false,
  statusText: "Deployment admin access is required for user administration.",
  error: null,
  queryDraft: emptyQuery(),
  acceptedQuery: null,
  queryStatus: "idle",
  users: [],
  paging: { next: null, hasMore: false, invalid: false },
  pagePending: false,
  selected: null,
  draft: null,
  targetStatus: "idle",
  create: emptyCreate(),
  createOpen: false,
  credentialDialog: null,
  password: "",
  reason: "",
  binding: emptyBinding(),
  enterpriseClaimed: false,
  providers: [],
  providersStatus: "idle",
  operation: { kind: "idle" },
  operationObserved: false,
  transportPending: false,
  leavePrompt: false,
  fieldErrors: {},
});
const mutable = (user: UserResource): MutableUser => ({
  email: user.email,
  display_name: user.display_name,
  mfa_required: user.mfa_required,
  is_active: user.is_active,
  is_deployment_admin: user.is_deployment_admin,
});
const equalQuery = (a: Query, b: Query | null) =>
  b !== null &&
  a.search === b.search &&
  a.isActive === b.isActive &&
  a.isAdmin === b.isAdmin;
const sameVersionFacts = (a: UserResource, b: UserResource) =>
  JSON.stringify({
    ...mutable(a),
    created_at: a.created_at,
    bindings: a.auth_bindings.map((binding) =>
      isEnterpriseAuthBinding(binding)
        ? {
            auth_binding_id: binding.auth_binding_id,
            provider_type: binding.provider_type,
            provider_key: binding.provider_key,
            provider_subject: binding.provider_subject,
            created_at: binding.created_at,
          }
        : binding,
    ),
  }) ===
  JSON.stringify({
    ...mutable(b),
    created_at: b.created_at,
    bindings: b.auth_bindings.map((binding) =>
      isEnterpriseAuthBinding(binding)
        ? {
            auth_binding_id: binding.auth_binding_id,
            provider_type: binding.provider_type,
            provider_key: binding.provider_key,
            provider_subject: binding.provider_subject,
            created_at: binding.created_at,
          }
        : binding,
    ),
  });
const labels: Record<Action, [string, string, string]> = {
  create: [
    "Creating local user",
    "Created local user",
    "Create local user failed",
  ],
  patch: [
    "Patching local user",
    "Patched local user",
    "Patch local user failed",
  ],
  password: [
    "Resetting user password",
    "Reset user password",
    "Reset user password failed",
  ],
  totp: ["Resetting user TOTP", "Reset user TOTP", "Reset user TOTP failed"],
  revoke: [
    "Revoking every user session",
    "Revoked every user session",
    "Revoke-all failed",
  ],
  bindingCreate: [
    "Creating enterprise auth binding",
    "Created enterprise auth binding",
    "Create enterprise auth binding failed",
  ],
  bindingRotate: [
    "Rotating enterprise auth binding",
    "Rotated enterprise auth binding",
    "Rotate enterprise auth binding failed",
  ],
  bindingRetire: [
    "Retiring enterprise auth binding",
    "Retired enterprise auth binding",
    "Retire enterprise auth binding failed",
  ],
};
function resourceReply(
  result: HTTPOperationResult<{ data: UserResource }>,
): MutationReply {
  return result.ok
    ? {
        ok: true,
        userId: result.payload.data.user_id,
        user: result.payload.data,
      }
    : { ok: false, status: result.status, error: result.payload.error ?? null };
}

export class DeploymentUsersController {
  private state = initial();
  private readonly listeners = new Set<() => void>();
  private identity: Identity | null = null;
  private epoch = 0;
  private selection = 0;
  private form = 0;
  private active = false;
  private autoLoad = false;
  private disposed = false;
  private unsubscribe: (() => void) | null = null;
  private queryGeneration = 0;
  private readRequest = 0;
  private pageRequest = 0;
  private targetRequest = 0;
  private providerRequest = 0;
  private operation: object | null = null;
  private transportCount = 0;
  private propagation: object | null = null;
  private readonly observations = new Map<string, () => void>();
  private debounce: ReturnType<typeof setTimeout> | undefined;
  private leave: ((accepted: boolean) => void) | null = null;
  constructor(
    private readonly session: DeploymentUsersSessionPort,
    private readonly api = { ...userApi, listEnterpriseAuthProviders },
  ) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(patch: Partial<State>) {
    this.state = { ...this.state, ...patch };
    for (const listener of this.listeners) listener();
  }
  start = () => {
    if (this.disposed || this.unsubscribe) return;
    this.unsubscribe = this.session.subscribe(this.sync);
    this.sync();
  };
  stop = () => {
    this.unsubscribe?.();
    this.unsubscribe = null;
  };
  retire = () => {
    ++this.epoch;
    ++this.selection;
    ++this.form;
    clearTimeout(this.debounce);
    for (const cancel of [...this.observations.values()]) cancel();
    this.observations.clear();
    this.leave?.(false);
    this.leave = null;
    this.operation = null;
    this.propagation = null;
    this.identity = null;
    this.transportCount = 0;
    this.publish(initial());
  };
  dispose = () => {
    this.stop();
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  };
  private sync = () => {
    const identity = this.session.identity();
    if (!identity?.admin) {
      if (this.identity !== null || this.state.authorized) this.retire();
      return;
    }
    if (
      identity.actor !== this.identity?.actor ||
      identity.lifetime !== this.identity?.lifetime
    ) {
      this.retire();
      this.identity = identity;
      this.publish({
        authorized: true,
        statusText: "Deployment user administration is ready.",
      });
    }
  };
  private current = (identity = this.identity, epoch = this.epoch) => {
    const next = this.session.identity();
    return (
      !this.disposed &&
      epoch === this.epoch &&
      identity !== null &&
      next?.admin === true &&
      identity.actor === next.actor &&
      identity.lifetime === next.lifetime
    );
  };
  open = () => {
    this.start();
    this.active = true;
  };
  close = () => {
    this.active = false;
    this.autoLoad = false;
    ++this.form;
    ++this.readRequest;
    ++this.pageRequest;
    ++this.targetRequest;
    ++this.providerRequest;
    clearTimeout(this.debounce);
    for (const key of ["query", "page", "target", "providers"])
      this.observations.get(key)?.();
    this.leave?.(false);
    this.leave = null;
    this.publish({
      createOpen: false,
      create: { ...this.state.create, password: "" },
      password: "",
      credentialDialog: null,
      leavePrompt: false,
      pagePending: false,
      targetStatus: "idle",
      fieldErrors: {},
    });
  };
  configure = (autoLoad: boolean, enterpriseClaimed: boolean) => {
    const newlyActive = autoLoad && !this.autoLoad;
    this.autoLoad = autoLoad;
    if (this.state.enterpriseClaimed !== enterpriseClaimed) {
      this.publish({
        enterpriseClaimed,
        ...(enterpriseClaimed
          ? {}
          : {
              providers: [],
              providersStatus: "idle",
              binding: emptyBinding(),
            }),
      });
      ++this.providerRequest;
      this.observations.get("providers")?.();
    }
    if (this.active && this.current()) {
      if (autoLoad && (newlyActive || this.state.acceptedQuery === null)) {
        const epoch = this.epoch;
        const form = this.form;
        queueMicrotask(() => {
          if (
            this.active &&
            form === this.form &&
            this.current(undefined, epoch)
          )
            void this.refreshUsers();
        });
      }
      if (enterpriseClaimed && this.state.providersStatus === "idle")
        void this.discoverProviders();
    }
  };
  private observe<T>(key: string, task: (signal: AbortSignal) => Promise<T>) {
    this.observations.get(key)?.();
    const observation = observeAccountOperation(task);
    this.observations.set(key, observation.cancel);
    void observation.result.then(() => {
      if (this.observations.get(key) === observation.cancel)
        this.observations.delete(key);
    });
    return observation;
  }
  hasDirtyDraft = () =>
    this.state.draft !== null &&
    Object.keys(this.state.draft.changes).length > 0;
  needsReview = () =>
    this.state.selected !== null &&
    this.state.draft !== null &&
    (this.state.draft.baseVersion !== this.state.selected.user_version ||
      (this.state.operation.kind === "rejected" &&
        this.state.operation.review &&
        this.state.operation.intent.target === this.state.selected.user_id) ||
      (this.state.operation.kind === "uncertain" &&
        this.state.operation.attempt === null &&
        this.state.operation.intent.target === this.state.selected.user_id));
  private canMutate = () =>
    this.active &&
    this.current() &&
    this.state.authorized &&
    this.operation === null &&
    this.transportCount === 0 &&
    this.state.operation.kind !== "uncertain" &&
    !(
      this.state.operation.kind === "rejected" &&
      this.state.operation.review &&
      this.state.operation.intent.target === this.state.selected?.user_id
    ) &&
    !(
      this.state.operation.kind === "confirmed" &&
      this.state.operation.propagation === "pending"
    );
  canTargetAction = () =>
    this.canMutate() &&
    this.state.selected !== null &&
    this.state.targetStatus !== "loading";
  canSave = () =>
    this.canTargetAction() && this.hasDirtyDraft() && !this.needsReview();
  canPage = () =>
    this.active &&
    this.current() &&
    this.state.queryStatus === "idle" &&
    !this.state.pagePending &&
    equalQuery(this.state.queryDraft, this.state.acceptedQuery) &&
    this.state.paging.hasMore &&
    !this.state.paging.invalid;
  requestLeave = (): Promise<boolean> => {
    if (!this.current() || !this.hasDirtyDraft()) return Promise.resolve(true);
    if (this.leave) return Promise.resolve(false);
    this.publish({ leavePrompt: true });
    return new Promise((resolve) => {
      this.leave = resolve;
    });
  };
  resolveLeave = async (choice: "stay" | "discard" | "save") => {
    const leave = this.leave;
    if (!leave) return;
    if (choice === "save") {
      const revision = this.state.draft?.revision;
      const selection = this.selection;
      const saved = await this.save();
      if (this.leave !== leave || !this.current()) return;
      if (
        !saved ||
        selection !== this.selection ||
        this.state.draft?.revision !== revision ||
        this.hasDirtyDraft()
      )
        return;
    } else if (choice === "discard") this.discard();
    if (this.leave !== leave) return;
    this.leave = null;
    this.publish({ leavePrompt: false });
    leave(choice !== "stay");
  };
  changeDraft = <K extends keyof MutableUser>(
    key: K,
    value: MutableUser[K],
  ) => {
    if (
      !this.active ||
      !this.current() ||
      !this.state.selected ||
      !this.state.draft
    )
      return;
    const changes = { ...this.state.draft.changes };
    if (value === this.state.selected[key]) delete changes[key];
    else changes[key] = value;
    this.publish({
      draft: {
        changes,
        baseVersion: this.hasDirtyDraft()
          ? this.state.draft.baseVersion
          : this.state.selected.user_version,
        revision: this.state.draft.revision + 1,
      },
      fieldErrors: {},
    });
  };
  discard = () => {
    if (this.state.selected && this.state.draft)
      this.publish({
        draft: {
          changes: {},
          baseVersion: this.state.selected.user_version,
          revision: this.state.draft.revision + 1,
        },
        fieldErrors: {},
      });
  };
  review = () => {
    const op = this.state.operation;
    if (
      !this.current() ||
      this.operation ||
      this.transportCount ||
      this.state.targetStatus === "loading" ||
      (op.kind === "uncertain" &&
        (op.attempt !== null || !this.state.operationObserved))
    )
      return;
    if (op.kind === "uncertain" && op.intent.action === "create") {
      this.publish({
        operation: { kind: "idle" },
        operationObserved: false,
        error: null,
        statusText:
          "Current users reviewed. The earlier create remains unconfirmed; explicitly review any new create.",
      });
      return;
    }
    if (
      !this.state.selected ||
      !this.state.draft ||
      this.state.targetStatus !== "idle" ||
      (op.kind === "uncertain" &&
        op.intent.target !== this.state.selected.user_id)
    )
      return;
    const selected = this.state.selected;
    const changes = Object.fromEntries(
      Object.entries(this.state.draft.changes).filter(
        ([key, value]) => selected[key as keyof MutableUser] !== value,
      ),
    );
    this.publish({
      draft: {
        ...this.state.draft,
        changes,
        baseVersion: selected.user_version,
      },
      operation: { kind: "idle" },
      operationObserved: false,
      error: null,
      statusText:
        "Current user reviewed. Remaining changes can be submitted as a new action.",
    });
  };
  changeQuery = (patch: Partial<Query>) => {
    if (!this.active || !this.current()) return;
    const queryDraft = { ...this.state.queryDraft, ...patch };
    if (equalQuery(queryDraft, this.state.queryDraft)) return;
    ++this.queryGeneration;
    ++this.pageRequest;
    ++this.readRequest;
    this.observations.get("page")?.();
    this.observations.get("query")?.();
    clearTimeout(this.debounce);
    this.publish({ queryDraft, queryStatus: "pending", pagePending: false });
    if (this.autoLoad)
      this.debounce = setTimeout(() => {
        void this.refreshUsers();
      }, 180);
  };
  private acceptResource(user: UserResource) {
    const previous =
      this.state.selected?.user_id === user.user_id
        ? this.state.selected
        : this.state.users.find((row) => row.user_id === user.user_id);
    if (previous && user.user_version < previous.user_version) return previous;
    if (
      previous &&
      user.user_version === previous.user_version &&
      !sameVersionFacts(previous, user)
    )
      return null;
    if (this.state.selected?.user_id === user.user_id)
      this.publish({
        selected: user,
        ...(this.state.draft && !this.hasDirtyDraft()
          ? { draft: { ...this.state.draft, baseVersion: user.user_version } }
          : {}),
      });
    return user;
  }
  private deny = async (error: APIError | null, current: () => boolean) => {
    if (!current() || !this.identity) return;
    const identity = this.identity;
    if (error?.code === "session_required") {
      this.retire();
      this.session.lost(identity);
      return;
    }
    if (error?.status === 403 || error?.code === "authorization_denied") {
      this.retire();
      this.identity = identity;
      this.publish({
        error: accountOperationError(error),
        statusText:
          "Deployment administration access is unavailable. Check access before continuing.",
      });
      const epoch = this.epoch;
      const valid = () => this.current(identity, epoch);
      await this.observe("authorization", (signal) =>
        Promise.resolve(this.session.refresh(identity, signal, valid)),
      ).result;
    }
  };
  refreshUsers = async () => {
    if (!this.active || !this.current()) return;
    clearTimeout(this.debounce);
    const generation = ++this.queryGeneration;
    const request = ++this.readRequest;
    ++this.pageRequest;
    this.observations.get("page")?.();
    const epoch = this.epoch;
    const query = Object.freeze({
      ...this.state.queryDraft,
      search: this.state.queryDraft.search.trim(),
    });
    const current = () =>
      this.active &&
      this.current(undefined, epoch) &&
      generation === this.queryGeneration &&
      request === this.readRequest;
    this.publish({ queryStatus: "loading", pagePending: false, error: null });
    const outcome = await this.observe("query", (signal) =>
      this.api.listUsers({
        limit: 100,
        search: query.search,
        isActive: query.isActive,
        isDeploymentAdmin: query.isAdmin,
        signal,
      }),
    ).result;
    if (!current()) return;
    if (
      outcome.kind === "completed" &&
      outcome.value.ok &&
      outcome.value.payload.meta.paging
    ) {
      const paging = outcome.value.payload.meta.paging;
      const payload = outcome.value.payload;
      const users: UserResource[] = [];
      for (const resource of payload.data.users) {
        const duplicate = users.find((row) => row.user_id === resource.user_id);
        const accepted =
          duplicate &&
          duplicate.user_version === resource.user_version &&
          !sameVersionFacts(duplicate, resource)
            ? null
            : this.acceptResource(
                duplicate && duplicate.user_version > resource.user_version
                  ? duplicate
                  : resource,
              );
        if (!accepted) {
          this.publish({
            queryStatus: "failed",
            error: { code: "invalid_deployment_user_response", status: 502 },
          });
          return;
        }
        const index = users.findIndex(
          (row) => row.user_id === accepted.user_id,
        );
        if (index < 0) users.push(accepted);
        else users[index] = accepted;
      }
      this.publish({
        authorized: true,
        users: users.sort((a, b) =>
          a.user_id < b.user_id ? -1 : a.user_id > b.user_id ? 1 : 0,
        ),
        acceptedQuery: query,
        queryDraft: query,
        operationObserved:
          this.state.operationObserved ||
          (this.state.operation.kind === "uncertain" &&
            this.state.operation.intent.action === "create"),
        queryStatus: "idle",
        paging: {
          next: paging.next_cursor,
          hasMore: paging.has_more,
          invalid: false,
        },
        statusText:
          this.state.operation.kind === "idle"
            ? "Deployment users loaded"
            : this.state.statusText,
      });
      return;
    }
    const error =
      outcome.kind === "completed" && !outcome.value.ok
        ? (outcome.value.payload.error ?? null)
        : null;
    this.publish({
      queryStatus: "failed",
      error: accountOperationError(error) ?? {
        code: "deployment_users_unavailable",
        status: 503,
      },
      statusText:
        this.state.operation.kind === "idle"
          ? "Deployment users unavailable"
          : this.state.statusText,
    });
    await this.deny(error, current);
  };
  loadMore = async () => {
    if (!this.canPage() || !this.state.acceptedQuery || !this.state.paging.next)
      return;
    const query = this.state.acceptedQuery;
    const cursor = this.state.paging.next;
    const generation = this.queryGeneration;
    const request = ++this.pageRequest;
    const epoch = this.epoch;
    const current = () =>
      this.active &&
      this.current(undefined, epoch) &&
      generation === this.queryGeneration &&
      request === this.pageRequest &&
      this.state.paging.next === cursor &&
      equalQuery(query, this.state.acceptedQuery);
    this.publish({ pagePending: true, error: null });
    const outcome = await this.observe("page", (signal) =>
      this.api.listUsers({
        limit: 100,
        cursorToken: cursor,
        search: query.search,
        isActive: query.isActive,
        isDeploymentAdmin: query.isAdmin,
        signal,
      }),
    ).result;
    if (!current()) return;
    if (
      outcome.kind === "completed" &&
      outcome.value.ok &&
      outcome.value.payload.meta.paging
    ) {
      const rows = new Map(this.state.users.map((row) => [row.user_id, row]));
      for (const resource of outcome.value.payload.data.users) {
        const previous = rows.get(resource.user_id);
        const accepted =
          previous &&
          previous.user_version === resource.user_version &&
          !sameVersionFacts(previous, resource)
            ? null
            : this.acceptResource(
                previous && previous.user_version > resource.user_version
                  ? previous
                  : resource,
              );
        if (!accepted) {
          this.publish({
            pagePending: false,
            paging: { ...this.state.paging, invalid: true },
            error: { code: "invalid_deployment_user_response", status: 502 },
          });
          return;
        }
        const prior = rows.get(accepted.user_id);
        if (!prior || accepted.user_version >= prior.user_version)
          rows.set(accepted.user_id, accepted);
      }
      const paging = outcome.value.payload.meta.paging;
      this.publish({
        users: [...rows.values()].sort((a, b) =>
          a.user_id < b.user_id ? -1 : a.user_id > b.user_id ? 1 : 0,
        ),
        paging: {
          next: paging.next_cursor,
          hasMore: paging.has_more,
          invalid: false,
        },
        pagePending: false,
        statusText:
          this.state.operation.kind === "idle"
            ? "Loaded more deployment users"
            : this.state.statusText,
      });
      return;
    }
    const error =
      outcome.kind === "completed" && !outcome.value.ok
        ? (outcome.value.payload.error ?? null)
        : null;
    this.publish({
      pagePending: false,
      paging: {
        ...this.state.paging,
        invalid: error?.code === "invalid_pagination_request",
      },
      error: accountOperationError(error) ?? {
        code: "deployment_users_unavailable",
        status: 503,
      },
      statusText:
        "Load more users failed. Refresh users to start from the first page.",
    });
    await this.deny(error, current);
  };
  select = async (userId: string) => {
    const epoch = this.epoch;
    if (!this.active || !this.current()) return;
    if (
      userId !== this.state.selected?.user_id &&
      this.hasDirtyDraft() &&
      !(await this.requestLeave())
    )
      return;
    if (!this.active || !this.current(undefined, epoch)) return;
    if (!userId) {
      this.clearSelection();
      return;
    }
    if (this.state.selected?.user_id !== userId) this.clearSelection();
    await this.loadTarget(userId);
  };
  private clearSelection() {
    ++this.selection;
    ++this.targetRequest;
    ++this.form;
    this.observations.get("target")?.();
    this.publish({
      selected: null,
      draft: null,
      binding: emptyBinding(),
      password: "",
      reason: "",
      credentialDialog: null,
      fieldErrors: {},
      targetStatus: "idle",
    });
  }
  refreshTarget = async () => {
    if (this.state.selected) await this.loadTarget(this.state.selected.user_id);
  };
  private loadTarget = async (target: string) => {
    if (!this.active || !this.current()) return;
    const selection = this.selection;
    const request = ++this.targetRequest;
    const epoch = this.epoch;
    const current = () =>
      this.active &&
      this.current(undefined, epoch) &&
      selection === this.selection &&
      request === this.targetRequest;
    this.publish({
      targetStatus: "loading",
      ...(this.state.operation.kind === "idle"
        ? { error: null, statusText: "Loading target user" }
        : {}),
    });
    const outcome = await this.observe("target", (signal) =>
      this.api.loadUser({ userId: target, signal }),
    ).result;
    if (!current()) return;
    if (
      outcome.kind === "completed" &&
      outcome.value.ok &&
      outcome.value.payload.data.user_id === target
    ) {
      const user = this.acceptResource(outcome.value.payload.data);
      if (user) {
        this.publish({
          authorized: true,
          selected: user,
          targetStatus: "idle",
          operationObserved:
            this.state.operationObserved ||
            (this.state.operation.kind === "uncertain" &&
              this.state.operation.intent.target === target),
          draft: this.state.draft ?? {
            changes: {},
            baseVersion: user.user_version,
            revision: 0,
          },
          binding: {
            ...this.state.binding,
            target:
              user.auth_bindings.find(isEnterpriseAuthBinding)
                ?.auth_binding_id ?? "",
          },
          statusText:
            this.state.operation.kind === "idle"
              ? "Loaded target user"
              : this.state.statusText,
        });
        return;
      }
    }
    const error =
      outcome.kind === "completed" && !outcome.value.ok
        ? (outcome.value.payload.error ?? null)
        : null;
    this.publish({
      targetStatus: "failed",
      error: accountOperationError(error) ?? {
        code: "deployment_user_unavailable",
        status: 503,
      },
      statusText: "Load target user failed",
    });
    await this.deny(error, current);
  };
  discoverProviders = async () => {
    if (!this.active || !this.current() || !this.state.enterpriseClaimed)
      return;
    const request = ++this.providerRequest;
    const epoch = this.epoch;
    this.publish({ providersStatus: "loading" });
    const outcome = await this.observe("providers", (signal) =>
      this.api.listEnterpriseAuthProviders({ signal }),
    ).result;
    if (
      !this.active ||
      !this.current(undefined, epoch) ||
      request !== this.providerRequest ||
      !this.state.enterpriseClaimed
    )
      return;
    this.publish(
      outcome.kind === "completed" && outcome.value.ok
        ? {
            providers: outcome.value.payload.data.providers,
            providersStatus: "ready",
          }
        : { providersStatus: "failed" },
    );
  };
  changeCreate = <K extends keyof CreateDraft>(
    key: K,
    value: CreateDraft[K],
  ) => {
    if (this.active && this.current())
      this.publish({
        create: { ...this.state.create, [key]: value },
        fieldErrors: {},
      });
  };
  createDialog = async (open: boolean) => {
    if (!open) {
      ++this.form;
      this.publish({
        createOpen: false,
        create: { ...this.state.create, password: "" },
        fieldErrors: {},
      });
      return;
    }
    const epoch = this.epoch;
    if (
      (this.hasDirtyDraft() && !(await this.requestLeave())) ||
      !this.current(undefined, epoch) ||
      !this.canMutate()
    )
      return;
    ++this.form;
    this.publish({ createOpen: true, create: emptyCreate(), fieldErrors: {} });
  };
  credentialDialog = (dialog: CredentialDialog | null) => {
    if (dialog && !this.canTargetAction()) return;
    ++this.form;
    this.publish({
      credentialDialog: dialog,
      password: "",
      reason: "",
      fieldErrors: {},
    });
  };
  changePassword = (password: string) => {
    if (this.active) this.publish({ password, fieldErrors: {} });
  };
  changeReason = (reason: string) => {
    if (this.active) this.publish({ reason });
  };
  changeBinding = <K extends keyof BindingDraft>(
    key: K,
    value: BindingDraft[K],
  ) => {
    if (this.active && this.current() && this.state.enterpriseClaimed)
      this.publish({
        binding: { ...this.state.binding, [key]: value },
        fieldErrors: {},
      });
  };
  private intent(action: Action): Intent | null {
    if (!this.identity) return null;
    return Object.freeze({
      identity: this.identity,
      action,
      target:
        action === "create" ? null : (this.state.selected?.user_id ?? null),
      label:
        action === "create"
          ? this.state.create.email
          : (this.state.selected?.display_name ?? ""),
      selection: this.selection,
      form: this.form,
      draftRevision: this.state.draft?.revision ?? 0,
      baseVersion: this.state.selected?.user_version ?? null,
    });
  }
  private async dispatch(
    intent: Intent,
    task: (signal: AbortSignal) => Promise<MutationReply>,
    attempt: ReplayAttempt | null = null,
    replay = false,
  ) {
    if (
      !this.active ||
      !this.current(intent.identity) ||
      !this.state.authorized ||
      this.operation ||
      this.transportCount > 0 ||
      (!replay && !this.canMutate()) ||
      (intent.action.startsWith("binding") && !this.state.enterpriseClaimed)
    )
      return false;
    const token = {};
    const epoch = this.epoch;
    this.operation = token;
    ++this.transportCount;
    const current = () =>
      this.current(intent.identity, epoch) && this.operation === token;
    this.publish({
      operation: { kind: "pending", intent, attempt },
      operationObserved: false,
      transportPending: true,
      error: null,
      statusText: labels[intent.action][0],
      fieldErrors: {},
    });
    const observation = this.observe("write", task);
    this.publish({
      password: "",
      create: { ...this.state.create, password: "" },
    });
    void observation.settled.then(() => {
      if (this.current(intent.identity, epoch)) {
        this.transportCount = Math.max(0, this.transportCount - 1);
        this.publish({ transportPending: this.transportCount > 0 });
      }
    });
    try {
      const outcome = await observation.result;
      if (!current()) return false;
      if (
        outcome.kind !== "completed" ||
        (!outcome.value.ok && outcome.value.status >= 500)
      ) {
        this.publish({
          operation: { kind: "uncertain", intent, attempt },
          statusText: `${labels[intent.action][0]}: outcome uncertain for ${intent.target ?? intent.label}. ${attempt ? "Replay the exact action to check the original request." : "Refresh current state and explicitly review a new action. Matching state is not a receipt."}`,
        });
        return false;
      }
      const result = outcome.value;
      if (!result.ok) {
        const review =
          result.error?.code === "user_version_conflict" ||
          result.error?.code === "client_txn_conflict";
        this.publish({
          operation: { kind: "rejected", intent, review },
          error: accountOperationError(result.error),
          statusText: labels[intent.action][2],
        });
        await this.deny(
          result.error && { ...result.error, status: result.status },
          current,
        );
        if (
          current() &&
          review &&
          intent.target === this.state.selected?.user_id
        )
          await this.refreshTarget();
        return false;
      }
      if (
        (intent.target !== null && result.userId !== intent.target) ||
        (result.user !== null &&
          intent.baseVersion !== null &&
          intent.action !== "create" &&
          result.user.user_version < intent.baseVersion)
      ) {
        this.publish({
          operation: { kind: "uncertain", intent, attempt },
          statusText:
            "Administrator response did not match the captured action. Review current state.",
        });
        return false;
      }
      if (result.user) {
        const accepted = this.acceptResource(result.user);
        if (!accepted) {
          this.publish({
            operation: { kind: "uncertain", intent, attempt },
            statusText:
              "Administrator response was inconsistent with accepted state. Review current state.",
          });
          return false;
        }
        if (
          intent.action === "create" &&
          intent.selection === this.selection &&
          intent.form === this.form &&
          this.state.createOpen &&
          !this.hasDirtyDraft()
        ) {
          ++this.selection;
          this.publish({
            selected: accepted,
            draft: {
              changes: {},
              baseVersion: accepted.user_version,
              revision: 0,
            },
          });
        }
        if (
          intent.action === "patch" &&
          intent.selection === this.selection &&
          this.state.selected?.user_id === intent.target &&
          this.state.draft?.revision === intent.draftRevision &&
          result.user.user_version >= accepted.user_version
        )
          this.publish({
            draft: {
              changes: {},
              baseVersion: accepted.user_version,
              revision: intent.draftRevision,
            },
          });
        // A mutation can update an existing accepted row, but cannot invent query membership.
        this.publish({
          users: this.state.users.map((user) =>
            user.user_id === accepted.user_id ? accepted : user,
          ),
        });
      }
      if (intent.form === this.form)
        this.publish({
          createOpen: false,
          credentialDialog: null,
          password: "",
          reason: "",
        });
      this.publish({
        operation: { kind: "confirmed", intent, propagation: "pending" },
        statusText: labels[intent.action][1],
        error: null,
      });
      if (
        intent.target === intent.identity.actor &&
        ["password", "totp", "revoke"].includes(intent.action)
      ) {
        this.session.revoked(
          intent.identity,
          `${labels[intent.action][1]}. Sign in again.`,
        );
        return true;
      }
      await this.refreshPropagation(intent, current);
      return true;
    } finally {
      if (this.operation === token) {
        this.operation = null;
        this.publish({});
      }
    }
  }
  private async refreshPropagation(intent: Intent, guard: () => boolean) {
    if (!guard()) return;
    const token = {};
    this.propagation = token;
    const current = () =>
      guard() &&
      this.propagation === token &&
      this.state.operation.kind === "confirmed" &&
      this.state.operation.intent === intent;
    const outcome = await this.observe("propagation", (signal) =>
      Promise.resolve(this.session.refresh(intent.identity, signal, current)),
    ).result;
    if (!current()) return;
    this.publish({
      operation: {
        kind: "confirmed",
        intent,
        propagation: outcome.kind === "completed" ? "ready" : "failed",
      },
      statusText:
        outcome.kind === "completed"
          ? labels[intent.action][1]
          : `${labels[intent.action][1]}. Session refresh failed; retry refresh without repeating the action.`,
    });
  }
  retryRefresh = async () => {
    const op = this.state.operation;
    if (
      op.kind !== "confirmed" ||
      op.propagation !== "failed" ||
      !this.current(op.intent.identity)
    )
      return;
    const epoch = this.epoch;
    this.publish({ operation: { ...op, propagation: "pending" } });
    await this.refreshPropagation(op.intent, () =>
      this.current(op.intent.identity, epoch),
    );
  };
  save = async () => {
    if (!this.canSave() || !this.state.selected || !this.state.draft)
      return false;
    const changes = { ...this.state.draft.changes };
    const errors: Record<string, string> = {};
    if (changes.email !== undefined) {
      const v = validateAccountEmail(changes.email);
      if (v.error) errors.email = v.error;
      else changes.email = v.value;
    }
    if (changes.display_name !== undefined) {
      const v = validateAccountDisplayName(changes.display_name);
      if (v.error) errors.display_name = v.error;
      else changes.display_name = v.value;
    }
    if (Object.keys(errors).length) {
      this.publish({ fieldErrors: errors });
      return false;
    }
    for (const key of Object.keys(changes) as (keyof MutableUser)[])
      if (changes[key] === this.state.selected[key]) delete changes[key];
    if (!Object.keys(changes).length) {
      this.publish({
        draft: {
          ...this.state.draft,
          changes: {},
          baseVersion: this.state.selected.user_version,
        },
        fieldErrors: {},
      });
      return true;
    }
    const intent = this.intent("patch");
    if (!intent) return false;
    const input = {
      userId: this.state.selected.user_id,
      baseUserVersion: this.state.draft.baseVersion,
      changes: Object.freeze(changes),
    };
    return this.dispatch(intent, async (signal) =>
      resourceReply(await this.api.patchLocalUser({ ...input, signal })),
    );
  };
  create = () => {
    if (!this.canMutate() || !this.state.createOpen) return;
    const draft = this.state.create;
    const email = validateAccountEmail(draft.email);
    const name = validateAccountDisplayName(draft.displayName);
    const passwordError = validateProvisioningPassword(draft.password);
    const errors: Record<string, string> = {};
    if (email.error) errors.createEmail = email.error;
    if (name.error) errors.createDisplayName = name.error;
    if (passwordError) errors.createPassword = passwordError;
    if (Object.keys(errors).length) {
      this.publish({ fieldErrors: errors });
      return;
    }
    const intent = this.intent("create");
    if (!intent) return;
    const input = {
      email: email.value,
      displayName: name.value,
      initialPassword: draft.password,
      mfaRequired: draft.mfaRequired,
      isDeploymentAdmin: draft.admin,
      clientTxnId: clientTxnID("deployment-create"),
    };
    void this.dispatch(intent, async (signal) =>
      resourceReply(await this.api.createLocalUser({ ...input, signal })),
    );
  };
  resetPassword = () => {
    if (!this.canTargetAction() || !this.state.selected) return;
    const error = validateProvisioningPassword(this.state.password);
    if (error) {
      this.publish({ fieldErrors: { password: error } });
      return;
    }
    const intent = this.intent("password");
    if (!intent) return;
    const input = {
      userId: this.state.selected.user_id,
      baseUserVersion: this.state.selected.user_version,
      newPassword: this.state.password,
      reason: this.state.reason,
      clientTxnId: clientTxnID("deployment-password"),
    };
    void this.dispatch(intent, async (signal) =>
      resourceReply(await this.api.adminResetPassword({ ...input, signal })),
    );
  };
  safeAction = (action: ReplayAttempt["action"]) => {
    if (
      !this.canTargetAction() ||
      !this.state.selected ||
      (action.startsWith("binding") && !this.state.enterpriseClaimed)
    )
      return;
    const selected = this.state.selected;
    const binding = this.state.binding;
    const base = {
      userId: selected.user_id,
      baseUserVersion: selected.user_version,
      reason: action.startsWith("binding") ? binding.reason : this.state.reason,
      clientTxnId: clientTxnID(`deployment-${action}`),
    };
    let attempt: ReplayAttempt;
    if (action === "totp") attempt = { action, input: base };
    else if (action === "revoke") {
      const { baseUserVersion: _, ...input } = base;
      attempt = { action, input };
    } else if (action === "bindingCreate") {
      if (!binding.providerKey.trim()) {
        this.publish({
          fieldErrors: { providerKey: "Enter the configured provider key." },
        });
        return;
      }
      attempt = {
        action,
        input: {
          ...base,
          providerKey: binding.providerKey.trim(),
          providerSubject: binding.subject,
        },
      };
    } else {
      if (
        !selected.auth_bindings.some(
          (value) =>
            isEnterpriseAuthBinding(value) &&
            value.auth_binding_id === binding.target,
        )
      )
        return;
      attempt =
        action === "bindingRotate"
          ? {
              action,
              input: {
                ...base,
                authBindingId: binding.target,
                newProviderSubject: binding.newSubject,
              },
            }
          : { action, input: { ...base, authBindingId: binding.target } };
    }
    Object.freeze(attempt.input);
    Object.freeze(attempt);
    const intent = this.intent(action);
    if (intent)
      void this.dispatch(
        intent,
        (signal) => this.sendReplay(attempt, signal),
        attempt,
      );
  };
  private async sendReplay(
    attempt: ReplayAttempt,
    signal: AbortSignal,
  ): Promise<MutationReply> {
    switch (attempt.action) {
      case "totp":
        return resourceReply(
          await this.api.adminResetTotp({ ...attempt.input, signal }),
        );
      case "bindingCreate":
        return resourceReply(
          await this.api.createEnterpriseAuthBinding({
            ...attempt.input,
            signal,
          }),
        );
      case "bindingRotate":
        return resourceReply(
          await this.api.rotateEnterpriseAuthBinding({
            ...attempt.input,
            signal,
          }),
        );
      case "bindingRetire":
        return resourceReply(
          await this.api.retireEnterpriseAuthBinding({
            ...attempt.input,
            signal,
          }),
        );
      case "revoke": {
        const result = await this.api.adminRevokeAllSessions({
          ...attempt.input,
          signal,
        });
        return result.ok
          ? { ok: true, userId: result.payload.data.user_id, user: null }
          : {
              ok: false,
              status: result.status,
              error: result.payload.error ?? null,
            };
      }
    }
  }
  replay = () => {
    const op = this.state.operation;
    if (op.kind === "uncertain" && op.attempt !== null) {
      const attempt = op.attempt;
      void this.dispatch(
        op.intent,
        (signal) => this.sendReplay(attempt, signal),
        attempt,
        true,
      );
    }
  };
}

export function useDeploymentUsers({
  controller,
  autoLoadUsers = false,
  enterpriseAuthClaimed = false,
  active = true,
}: DeploymentUsersPanelProps) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  useLayoutEffect(() => {
    if (active) controller.open();
    else controller.close();
    return controller.close;
  }, [controller, active]);
  useLayoutEffect(() => {
    controller.configure(active && autoLoadUsers, enterpriseAuthClaimed);
  }, [controller, autoLoadUsers, enterpriseAuthClaimed, active]);
  const selected = state.selected;
  const values = selected
    ? { ...mutable(selected), ...state.draft?.changes }
    : null;
  return {
    ...state,
    selectedUser: selected,
    userFilter: state.queryDraft.search,
    userActiveFilter:
      state.queryDraft.isActive === null
        ? "all"
        : String(state.queryDraft.isActive),
    userAdminFilter:
      state.queryDraft.isAdmin === null
        ? "all"
        : String(state.queryDraft.isAdmin),
    usersHasMore: state.paging.hasMore,
    setUserFilter: (search: string) => controller.changeQuery({ search }),
    setUserActiveFilter: (v: string) =>
      controller.changeQuery({ isActive: v === "all" ? null : v === "true" }),
    setUserAdminFilter: (v: string) =>
      controller.changeQuery({ isAdmin: v === "all" ? null : v === "true" }),
    createEmail: state.create.email,
    createDisplayName: state.create.displayName,
    createInitialPassword: state.create.password,
    createMfaRequired: state.create.mfaRequired,
    createIsDeploymentAdmin: state.create.admin,
    createDialogOpen: state.createOpen,
    setCreateEmail: (v: string) => controller.changeCreate("email", v),
    setCreateDisplayName: (v: string) =>
      controller.changeCreate("displayName", v),
    setCreateInitialPassword: (v: string) =>
      controller.changeCreate("password", v),
    setCreateMfaRequired: (v: boolean) =>
      controller.changeCreate("mfaRequired", v),
    setCreateIsDeploymentAdmin: (v: boolean) =>
      controller.changeCreate("admin", v),
    setCreateDialogOpen: controller.createDialog,
    patchEmail: values?.email ?? "",
    patchDisplayName: values?.display_name ?? "",
    patchMfaRequired: values?.mfa_required ?? true,
    patchIsActive: values?.is_active ?? true,
    patchIsDeploymentAdmin: values?.is_deployment_admin ?? false,
    setPatchEmail: (v: string) => controller.changeDraft("email", v),
    setPatchDisplayName: (v: string) =>
      controller.changeDraft("display_name", v),
    setPatchMfaRequired: (v: boolean) =>
      controller.changeDraft("mfa_required", v),
    setPatchIsActive: (v: boolean) => controller.changeDraft("is_active", v),
    setPatchIsDeploymentAdmin: (v: boolean) =>
      controller.changeDraft("is_deployment_admin", v),
    adminNewPassword: state.password,
    adminReason: state.reason,
    setAdminNewPassword: controller.changePassword,
    setAdminReason: controller.changeReason,
    setCredentialDialog: controller.credentialDialog,
    enterpriseProviders: state.providers,
    bindingProviderKey: state.binding.providerKey,
    bindingProviderSubject: state.binding.subject,
    bindingTargetID: state.binding.target,
    bindingNewSubject: state.binding.newSubject,
    bindingReason: state.binding.reason,
    setBindingProviderKey: (v: string) =>
      controller.changeBinding("providerKey", v),
    setBindingProviderSubject: (v: string) =>
      controller.changeBinding("subject", v),
    setBindingTargetID: (v: string) => controller.changeBinding("target", v),
    setBindingNewSubject: (v: string) =>
      controller.changeBinding("newSubject", v),
    setBindingReason: (v: string) => controller.changeBinding("reason", v),
    clearSelectedUser: () => {
      void controller.select("");
    },
    loadSelectedUser: controller.select,
    handleCreateUser: controller.create,
    handlePatchUser: controller.save,
    handleAdminPasswordReset: controller.resetPassword,
    handleAdminTotpReset: () => controller.safeAction("totp"),
    handleAdminRevokeAll: () => controller.safeAction("revoke"),
    handleCreateEnterpriseBinding: () => controller.safeAction("bindingCreate"),
    handleRotateEnterpriseBinding: () => controller.safeAction("bindingRotate"),
    handleRetireEnterpriseBinding: () => controller.safeAction("bindingRetire"),
    refreshUsers: controller.refreshUsers,
    loadMoreUsers: controller.loadMore,
    targetOperationPending:
      state.targetStatus === "loading" || state.operation.kind === "pending",
    canSubmitTargetAction: controller.canTargetAction(),
    canSave: controller.canSave(),
    canPage: controller.canPage(),
    dirty: controller.hasDirtyDraft(),
    reviewRequired: controller.needsReview(),
  };
}
