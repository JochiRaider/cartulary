import { clientTxnID } from "../services/browserApi";
import type { AuthorizationRecoveryResult } from "../shared/authorizationRecovery";
import {
  type IncidentResource,
  incidentResourceOrder,
} from "../shared/incidentResource";
import { observeAccountOperation } from "./accountOperation";
import type {
  LifecycleResult,
  mutateIncidentLifecycle,
  readLifecycleIncident,
} from "./api/incidentLifecycleClient";
import {
  initialLifecycleState,
  type LifecycleAction,
  type LifecycleAttempt,
  type LifecycleAuthority,
  type LifecycleState,
  lifecycleAllowed,
  lifecycleBinding,
  lifecycleProblemMessage,
} from "./incidentLifecycleModel";

export type IncidentLifecyclePorts = {
  read: typeof readLifecycleIncident;
  mutate: typeof mutateIncidentLifecycle;
  isCurrent: (authority: LifecycleAuthority) => boolean;
  recover: (
    authority: LifecycleAuthority,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<AuthorizationRecoveryResult>;
  lost: (reason: "session" | "incident", authority: LifecycleAuthority) => void;
  publishResource?:
    | ((resource: IncidentResource, authority: LifecycleAuthority) => void)
    | undefined;
};
type Observation = { cancel: () => void };

/** Owns one action and exact recovery record. Observation is never acknowledgement. */
export class IncidentLifecycleController {
  private state = initialLifecycleState();
  private listeners = new Set<() => void>();
  private epoch = 0;
  private sequence = 0;
  private read: Observation | null = null;
  private write: Observation | null = null;
  private transport: object | null = null;
  private leave: ((accepted: boolean) => void) | null = null;
  private disposed = false;
  constructor(private readonly ports: IncidentLifecyclePorts) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(change: Partial<LifecycleState>) {
    if (this.disposed) return;
    this.state = { ...this.state, ...change };
    for (const listener of this.listeners) listener();
  }
  private current(authority = this.state.authority, epoch = this.epoch) {
    return (
      !this.disposed &&
      authority !== null &&
      this.state.authority !== null &&
      epoch === this.epoch &&
      lifecycleBinding(authority) === lifecycleBinding(this.state.authority) &&
      this.ports.isCurrent(authority)
    );
  }
  private cancelRead() {
    const read = this.read;
    this.read = null;
    read?.cancel();
  }
  retire = () => {
    ++this.epoch;
    this.cancelRead();
    const write = this.write;
    this.write = null;
    write?.cancel();
    const leave = this.leave;
    this.leave = null;
    leave?.(false);
    this.publish({
      ...initialLifecycleState(),
      transportPending: this.transport !== null,
    });
  };
  dispose = () => {
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  };
  setAuthority = (authority: LifecycleAuthority | null, observed = false) => {
    const previous = this.state.authority;
    if (!previous && !authority) return;
    if (
      !previous ||
      !authority ||
      lifecycleBinding(previous) !== lifecycleBinding(authority)
    ) {
      const active = this.state.active;
      this.retire();
      this.publish({ authority, active });
      if (authority && active) this.refresh();
      return;
    }
    if (previous.role !== authority.role) {
      const operation = this.state.operation;
      if (!observed) {
        this.cancelRead();
        if (
          operation.kind === "pending" &&
          operation.stage === "authorization"
        ) {
          const write = this.write;
          this.write = null;
          write?.cancel();
          this.publish({
            operation: operation.replay
              ? { kind: "uncertain", attempt: operation.attempt }
              : { kind: "idle" },
          });
        }
      }
      this.publish({
        ...(!observed
          ? { access: "checking" as const, read: "idle" as const }
          : {}),
        authority,
        review: null,
        notice:
          authority.role === "admin"
            ? "Administrator access changed. Refresh and review before acting."
            : "Only current incident admins can close, reopen or replay. Your local reason and recovery are retained.",
      });
      if (!observed && this.state.active && !this.write) this.refresh();
    }
  };
  setActive = (active: boolean) => {
    if (this.disposed || active === this.state.active) return;
    this.cancelRead();
    const operation = this.state.operation;
    if (
      !active &&
      operation.kind === "pending" &&
      operation.stage === "authorization"
    ) {
      const write = this.write;
      this.write = null;
      write?.cancel();
      this.publish({
        operation: operation.replay
          ? { kind: "uncertain", attempt: operation.attempt }
          : { kind: "idle" },
      });
    }
    this.publish({ active, access: "checking", read: "idle", review: null });
    if (active && this.state.authority) this.refresh();
  };
  private acceptAccess(
    result: AuthorizationRecoveryResult,
    authority: LifecycleAuthority,
  ) {
    if (result.kind === "session_lost" || result.kind === "access_lost") {
      this.retire();
      this.ports.lost(
        result.kind === "session_lost" ? "session" : "incident",
        authority,
      );
      return false;
    }
    if (result.kind !== "authorized" || result.userId !== authority.actorId) {
      this.publish({ access: "unavailable", read: "failed", review: null });
      return false;
    }
    this.setAuthority({ ...authority, role: result.role }, true);
    this.publish({ access: "ready" });
    return true;
  }
  private signalLoss(
    result: Exclude<LifecycleResult, { ok: true }>,
    authority: LifecycleAuthority,
  ) {
    if (
      result.status === 401 ||
      (result.status === 404 && result.problem.code === "incident_not_found")
    ) {
      this.retire();
      this.ports.lost(
        result.status === 401 ? "session" : "incident",
        authority,
      );
      return true;
    }
    return false;
  }
  /** Ingestion from the neutral App resource owner does not echo publication. */
  acceptResource = (resource: IncidentResource, broadcast = true): boolean => {
    const authority = this.state.authority;
    if (!authority || !this.current()) return false;
    const order = incidentResourceOrder(
      this.state.resource,
      resource,
      authority.incidentId,
    );
    if (order === "invalid") return false;
    if (order === "new") {
      this.publish({ resource: Object.freeze({ ...resource }), review: null });
      if (broadcast) this.ports.publishResource?.(resource, authority);
    }
    return true;
  };
  refresh = () => {
    const authority = this.state.authority;
    if (!authority || !this.current() || !this.state.active || this.write)
      return;
    this.cancelRead();
    const epoch = this.epoch;
    const admission: Observation = { cancel: () => {} };
    this.read = admission;
    const current = () =>
      this.read === admission &&
      this.current(authority, epoch) &&
      this.state.active;
    this.publish({
      read: this.state.resource ? "refreshing" : "loading",
      review: null,
    });
    const observation = observeAccountOperation(async (signal) => {
      const access = await this.ports.recover(authority, signal, current);
      if (!current() || !this.acceptAccess(access, authority) || !current())
        return;
      const result = await this.ports.read({ authority, signal });
      if (!current()) return;
      if (!result.ok) {
        if (!this.signalLoss(result, authority))
          this.publish({
            read: "failed",
            ...(result.status === 403
              ? { access: "unavailable" as const }
              : {}),
          });
        return;
      }
      this.publish({
        read: this.acceptResource(result.resource) ? "ready" : "failed",
      });
    });
    admission.cancel = observation.cancel;
    void observation.result.then((outcome) => {
      if (!current()) return;
      this.read = null;
      if (outcome.kind !== "completed") this.publish({ read: "failed" });
    });
  };
  changeReason = (reason: string) => {
    if (
      !this.current() ||
      !this.state.active ||
      this.state.authority?.role !== "admin" ||
      reason === this.state.draft.reason
    )
      return;
    this.publish({
      draft: {
        ...this.state.draft,
        reason,
        revision: this.state.draft.revision + 1,
      },
      review: null,
      fieldError: null,
    });
  };
  private unresolved() {
    const op = this.state.operation;
    return (
      op.kind === "pending" ||
      op.kind === "uncertain" ||
      (op.kind === "rejected" && op.problem.code === "client_txn_conflict")
    );
  }
  private adminReady() {
    return (
      this.current() &&
      this.state.active &&
      this.state.access === "ready" &&
      this.state.authority?.role === "admin"
    );
  }
  canPropose = (action: LifecycleAction) =>
    this.adminReady() &&
    this.state.read === "ready" &&
    !this.write &&
    !this.transport &&
    !this.unresolved() &&
    this.state.resource !== null &&
    lifecycleAllowed(action, this.state.resource);
  propose = (action: LifecycleAction) => {
    if (!this.canPropose(action)) return;
    this.publish({
      draft: { ...this.state.draft, action },
      review: null,
      notice: "",
    });
    this.review();
  };
  review = () => {
    const { draft, resource, authority } = this.state;
    if (
      !draft.action ||
      !resource ||
      !authority ||
      !this.canPropose(draft.action)
    )
      return;
    this.publish({
      review: Object.freeze({
        resource,
        revision: draft.revision,
        role: authority.role,
        action: draft.action,
      }),
      notice:
        "Current incident reviewed. Confirm explicitly to send this action.",
    });
  };
  cancelReview = () => {
    this.publish({
      review: null,
      draft: { ...this.state.draft, action: null },
      notice: "Review cancelled. Your reason is retained.",
    });
  };
  canConfirm = () => {
    const { review, draft, resource, authority } = this.state;
    return (
      !!review &&
      !!resource &&
      !!authority &&
      this.canPropose(review.action) &&
      draft.reason !== "" &&
      review.revision === draft.revision &&
      review.role === authority.role &&
      review.resource.incident_version === resource.incident_version &&
      draft.action === review.action
    );
  };
  confirm = () => {
    const { review, authority, draft } = this.state;
    if (!this.canConfirm() || !review || !authority) return;
    let txn: string;
    try {
      txn = clientTxnID(`incident-${review.action}`);
    } catch {
      this.publish({
        notice: "Secure action identity is unavailable. No action was sent.",
      });
      return;
    }
    const attempt: LifecycleAttempt = Object.freeze({
      id: ++this.sequence,
      authority: Object.freeze({ ...authority }),
      action: review.action,
      resource: Object.freeze({ ...review.resource }),
      revision: draft.revision,
      payload: Object.freeze({
        base_incident_version: review.resource.incident_version,
        client_txn_id: txn,
        reason: draft.reason,
      }),
    });
    this.dispatch(attempt, false);
  };
  canReplay = () =>
    this.adminReady() &&
    !this.write &&
    !this.transport &&
    this.state.operation.kind === "uncertain" &&
    this.state.operation.problem?.code !== "client_txn_conflict";
  replay = () => {
    const op = this.state.operation;
    if (this.canReplay() && op.kind === "uncertain")
      this.dispatch(op.attempt, true);
  };
  private dispatch(attempt: LifecycleAttempt, replay: boolean) {
    if (
      !this.adminReady() ||
      this.write ||
      this.transport ||
      !this.current(attempt.authority)
    )
      return;
    const epoch = this.epoch;
    const admission: Observation = { cancel: () => {} };
    this.write = admission;
    this.cancelRead();
    this.publish({
      operation: { kind: "pending", attempt, replay, stage: "authorization" },
      review: null,
      fieldError: null,
      notice: "",
    });
    const current = () =>
      this.write === admission && this.current(attempt.authority, epoch);
    const attemptCurrent = () =>
      this.current(attempt.authority, epoch) &&
      "attempt" in this.state.operation &&
      this.state.operation.attempt.id === attempt.id &&
      (this.state.operation.kind === "pending" ||
        this.state.operation.kind === "uncertain");
    let sent = false;
    const observation = observeAccountOperation(async (signal) => {
      const access = await this.ports.recover(
        attempt.authority,
        signal,
        current,
      );
      if (
        !current() ||
        !this.acceptAccess(access, attempt.authority) ||
        !current() ||
        !this.state.active ||
        this.state.authority?.role !== "admin"
      )
        return;
      // Fresh review can become stale during preflight. Replay must keep its old base.
      if (
        !replay &&
        (!this.state.resource ||
          this.state.resource.incident_version !==
            attempt.payload.base_incident_version ||
          !lifecycleAllowed(attempt.action, this.state.resource))
      )
        return;
      const transport = {};
      this.transport = transport;
      sent = true;
      this.publish({
        transportPending: true,
        operation: { kind: "pending", attempt, replay, stage: "write" },
      });
      try {
        const result = await this.ports.mutate({ attempt, signal });
        if (attemptCurrent()) {
          if (this.write === admission) this.write = null;
          this.complete(attempt, result, replay);
        }
      } finally {
        if (this.transport === transport) {
          this.transport = null;
          this.publish({ transportPending: false });
        }
      }
    });
    admission.cancel = observation.cancel;
    void observation.result.then((outcome) => {
      if (!current()) return;
      this.write = null;
      if (sent || replay)
        this.publish({ operation: { kind: "uncertain", attempt } });
      else
        this.publish({
          operation: { kind: "idle" },
          notice:
            "The action was not sent. Refresh and review your retained reason before acting.",
        });
      if (!sent && outcome.kind !== "completed")
        this.publish({ access: "unavailable", read: "failed" });
    });
  }
  private complete(
    attempt: LifecycleAttempt,
    result: LifecycleResult,
    replay: boolean,
  ) {
    if (!result.ok) {
      if (this.signalLoss(result, attempt.authority)) return;
      const { problem } = result;
      const definite =
        (result.status === 400 &&
          problem.code === "invalid_incident_lifecycle_request") ||
        (result.status === 409 &&
          [
            "incident_version_conflict",
            "illegal_transition",
            "client_txn_conflict",
            "credential_bootstrap_rejected",
          ].includes(problem.code)) ||
        (result.status === 403 &&
          ["authorization_denied", "csrf_verification_failed"].includes(
            problem.code,
          ));
      this.publish({
        operation: {
          kind: !replay && definite ? "rejected" : "uncertain",
          attempt,
          problem,
        },
        review: null,
        fieldError:
          !replay &&
          problem.code === "invalid_incident_lifecycle_request" &&
          problem.field === "reason" &&
          this.state.draft.revision === attempt.revision
            ? {
                revision: attempt.revision,
                message: lifecycleProblemMessage(problem),
              }
            : null,
      });
      if (definite && problem.code !== "invalid_incident_lifecycle_request")
        this.refresh();
      return;
    }
    const receipt = Object.freeze({ ...result.resource });
    const draft = this.state.draft;
    this.publish({
      operation: { kind: "confirmed", attempt, receipt, replay },
      review: null,
      draft:
        draft.revision === attempt.revision
          ? { reason: "", action: null, revision: draft.revision + 1 }
          : draft,
      fieldError:
        draft.revision === attempt.revision ? null : this.state.fieldError,
    });
    // Replay receipts are historical even when no newer resource is locally materialized.
    if (!replay) this.acceptResource(receipt);
    this.refresh();
  }
  discardReason = () => {
    if (!this.current()) return;
    this.publish({
      draft: {
        reason: "",
        action: null,
        revision: this.state.draft.revision + 1,
      },
      review: null,
      fieldError: null,
      notice: this.unresolved()
        ? "Local reason discarded. The captured server action and its recovery remain unresolved."
        : "Local reason discarded.",
    });
  };
  canForget = () =>
    this.current() && !this.write && !this.transport && this.unresolved();
  forgetRecovery = () => {
    if (!this.canForget()) return;
    this.publish({
      operation: { kind: "idle" },
      review: null,
      notice:
        "Recovery forgotten. This cannot cancel or undo the earlier server action. Refresh and review before any new action.",
    });
    this.refresh();
  };
  hasDepartureWork = () =>
    this.current() &&
    (this.state.draft.reason !== "" ||
      this.state.draft.action !== null ||
      this.unresolved());
  requestLeave = (): Promise<boolean> => {
    if (!this.hasDepartureWork()) return Promise.resolve(true);
    if (this.leave) return Promise.resolve(false);
    this.publish({ departure: true });
    return new Promise((resolve) => {
      this.leave = resolve;
    });
  };
  resolveDeparture = (choice: "stay" | "discard") => {
    const leave = this.leave;
    if (!leave) return;
    this.leave = null;
    this.publish({ departure: false });
    if (choice === "discard") this.retire();
    leave(choice === "discard");
  };
}
