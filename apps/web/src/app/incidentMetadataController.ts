import type { AuthorizationRecoveryResult } from "../shared/authorizationRecovery";
import { observeAccountOperation } from "./accountOperation";
import type {
  MetadataResult,
  patchIncidentMetadata,
  readIncidentMetadata,
} from "./api/incidentMetadataClient";
import {
  buildIncidentMetadataPatch,
  changedMetadataFields,
  type IncidentMetadataField,
  type IncidentMetadataResource,
  initialMetadataState,
  type MetadataAttempt,
  type MetadataAuthority,
  type MetadataState,
  metadataBinding,
  metadataDirty,
  metadataEditRole,
  metadataFieldError,
  metadataFields,
  metadataValues,
  newMetadataDraft,
} from "./incidentMetadataModel";

export type IncidentMetadataPorts = {
  read: typeof readIncidentMetadata;
  patch: typeof patchIncidentMetadata;
  isCurrent: (authority: MetadataAuthority) => boolean;
  recover: (
    authority: MetadataAuthority,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<AuthorizationRecoveryResult>;
  lost: (reason: "session" | "incident", authority: MetadataAuthority) => void;
  publishResource?:
    | ((
        resource: IncidentMetadataResource,
        authority: MetadataAuthority,
      ) => void)
    | undefined;
};
type Observation = { token: number; cancel: () => void };

/** One incident/session draft and operation; App owns session acceptance and navigation. */
export class IncidentMetadataController {
  private state = initialMetadataState();
  private listeners = new Set<() => void>();
  private epoch = 0;
  private sequence = 0;
  private read: Observation | null = null;
  private write: Observation | null = null;
  private transport: object | null = null;
  private leave: ((accepted: boolean) => void) | null = null;
  private disposed = false;
  constructor(private readonly ports: IncidentMetadataPorts) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(change: Partial<MetadataState>) {
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
      metadataBinding(authority) === metadataBinding(this.state.authority) &&
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
    this.state = {
      ...initialMetadataState(),
      transportPending: this.transport !== null,
    };
    for (const listener of this.listeners) listener();
  };
  dispose = () => {
    this.retire();
    this.disposed = true;
    this.listeners.clear();
  };
  setAuthority = (authority: MetadataAuthority | null) => {
    const previous = this.state.authority;
    if (!previous && !authority) return;
    if (
      !previous ||
      !authority ||
      metadataBinding(previous) !== metadataBinding(authority)
    ) {
      const active = this.state.active;
      this.retire();
      this.publish({ authority, active });
      if (authority && active) this.refresh();
      return;
    }
    if (previous.role !== authority.role)
      this.publish({
        authority,
        draft: this.state.draft
          ? {
              ...this.state.draft,
              reviewRequired:
                metadataDirty(this.state.draft) ||
                this.state.draft.reviewRequired,
            }
          : null,
      });
  };
  setActive = (active: boolean) => {
    if (this.disposed || active === this.state.active) return;
    this.cancelRead();
    this.publish({ active, access: "checking", read: "idle" });
    if (active && this.state.authority) this.refresh();
  };
  private acceptAccess(
    result: AuthorizationRecoveryResult,
    authority: MetadataAuthority,
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
      this.publish({ access: "unavailable", read: "failed" });
      return false;
    }
    this.setAuthority({ ...authority, role: result.role });
    this.publish({ access: "ready" });
    return true;
  }
  private signalLoss(
    result: Exclude<MetadataResult, { ok: true }>,
    authority: MetadataAuthority,
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
  /** Accept external/lifecycle observations without changing an outstanding draft's base. */
  acceptResource = (resource: IncidentMetadataResource, broadcast = true) => {
    const authority = this.state.authority;
    if (
      !authority ||
      !this.current(authority) ||
      resource.incident_id !== authority.incidentId ||
      !Number.isSafeInteger(resource.incident_version) ||
      resource.incident_version < 1 ||
      (this.state.resource &&
        resource.incident_version < this.state.resource.incident_version)
    )
      return;
    const accepted = Object.freeze({ ...resource });
    const draft = this.state.draft;
    const op = this.state.operation;
    const retain =
      draft &&
      (metadataDirty(draft) ||
        op.kind === "pending" ||
        op.kind === "uncertain" ||
        op.kind === "conflicted");
    this.publish({
      resource: accepted,
      draft: retain
        ? {
            ...draft,
            reviewRequired:
              draft.reviewRequired ||
              accepted.incident_version !== draft.base.incident_version ||
              accepted.status !== draft.base.status,
          }
        : newMetadataDraft(accepted),
    });
    if (broadcast) this.ports.publishResource?.(accepted, authority);
  };
  refresh = () => {
    const authority = this.state.authority;
    if (!authority || !this.current() || !this.state.active) return;
    this.cancelRead();
    const epoch = this.epoch;
    const admission: Observation = { token: ++this.sequence, cancel: () => {} };
    this.read = admission;
    const current = () =>
      this.read === admission &&
      this.current(authority, epoch) &&
      this.state.active;
    this.publish({
      read: this.state.resource ? "refreshing" : "loading",
      readGeneration: admission.token,
    });
    const observation = observeAccountOperation(async (signal) => {
      const access = await this.ports.recover(authority, signal, current);
      if (!current() || !this.acceptAccess(access, authority) || !current())
        return;
      const result = await this.ports.read({ authority, signal });
      if (!current()) return;
      if (!result.ok) {
        if (!this.signalLoss(result, authority))
          this.publish({ read: "failed" });
        return;
      }
      this.acceptResource(result.resource);
      this.publish({ read: "ready" });
    });
    admission.cancel = observation.cancel;
    void observation.result.then((outcome) => {
      if (!current()) return;
      this.read = null;
      if (outcome.kind !== "completed") this.publish({ read: "failed" });
    });
  };
  change = (field: IncidentMetadataField, value: string) => {
    const draft = this.state.draft;
    if (
      !draft ||
      !this.current() ||
      !this.state.active ||
      !metadataEditRole(this.state.authority?.role) ||
      this.state.resource?.status !== "active"
    )
      return;
    if (draft.values[field] === value) return;
    const revision = draft.revision + 1;
    const fieldErrors = { ...this.state.fieldErrors };
    delete fieldErrors[field];
    this.publish({
      draft: {
        ...draft,
        values: { ...draft.values, [field]: value },
        revisions: { ...draft.revisions, [field]: revision },
        revision,
      },
      fieldErrors,
    });
  };
  canSave = () =>
    this.current() &&
    this.state.active &&
    this.state.access === "ready" &&
    this.state.read === "ready" &&
    metadataEditRole(this.state.authority?.role) &&
    this.state.resource?.status === "active" &&
    metadataDirty(this.state.draft) &&
    !this.state.draft?.reviewRequired &&
    !this.write &&
    !this.transport &&
    this.state.operation.kind !== "uncertain" &&
    this.state.operation.kind !== "conflicted";
  save = () => {
    const authority = this.state.authority;
    const draft = this.state.draft;
    if (!this.canSave() || !authority || !draft) return;
    let payload: MetadataAttempt["payload"];
    try {
      payload = buildIncidentMetadataPatch(draft);
    } catch {
      this.publish({
        fieldErrors: {
          tlp: {
            revision: draft.revisions.tlp,
            message: "Select a listed TLP value or Unset.",
          },
        },
      });
      return;
    }
    const attempt: MetadataAttempt = Object.freeze({
      id: ++this.sequence,
      authority: Object.freeze({ ...authority }),
      base: Object.freeze({ ...draft.base }),
      revision: draft.revision,
      revisions: Object.freeze({ ...draft.revisions }),
      values: Object.freeze({ ...draft.values }),
      fields: Object.freeze(changedMetadataFields(draft)),
      payload,
    });
    const epoch = this.epoch;
    const admission: Observation = { token: attempt.id, cancel: () => {} };
    this.write = admission;
    this.cancelRead();
    this.publish({
      operation: { kind: "pending", attempt, stage: "authorization" },
      fieldErrors: {},
      reviewNotice: "",
    });
    const current = () =>
      this.current(authority, epoch) && this.write === admission;
    const attemptCurrent = () =>
      this.current(authority, epoch) &&
      "attempt" in this.state.operation &&
      this.state.operation.attempt.id === attempt.id &&
      (this.state.operation.kind === "pending" ||
        this.state.operation.kind === "uncertain");
    let sent = false;
    const observation = observeAccountOperation(async (signal) => {
      const access = await this.ports.recover(authority, signal, current);
      if (
        !current() ||
        !this.acceptAccess(access, authority) ||
        !current() ||
        !this.state.active ||
        !metadataEditRole(this.state.authority?.role) ||
        this.state.resource?.status !== "active" ||
        this.state.draft?.reviewRequired
      )
        return;
      const transport = {};
      this.transport = transport;
      sent = true;
      this.publish({
        transportPending: true,
        operation: { kind: "pending", attempt, stage: "write" },
      });
      try {
        const result = await this.ports.patch({ authority, payload, signal });
        // A deadline is not a server outcome. A later valid acknowledgement still belongs to this attempt.
        if (attemptCurrent()) {
          if (this.write === admission) this.write = null;
          this.complete(attempt, result);
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
      if (sent) {
        if (outcome.kind !== "completed")
          this.publish({ operation: { kind: "uncertain", attempt } });
      } else {
        this.publish({
          operation: { kind: "idle" },
          reviewNotice:
            "Save was not sent. Check current access and review your retained changes before saving.",
        });
        if (outcome.kind !== "completed")
          this.publish({ access: "unavailable", read: "failed" });
      }
    });
  };
  private complete(attempt: MetadataAttempt, result: MetadataResult) {
    if (!result.ok) {
      if (this.signalLoss(result, attempt.authority)) return;
      if (
        result.status === 409 &&
        result.problem.code === "incident_version_conflict"
      ) {
        this.publish({
          operation: { kind: "conflicted", attempt },
          draft: this.state.draft
            ? { ...this.state.draft, reviewRequired: true }
            : null,
        });
        this.refresh();
        return;
      }
      if (
        (result.status === 400 &&
          result.problem.code === "invalid_incident_patch") ||
        (result.status === 409 && result.problem.code === "incident_closed") ||
        (result.status === 403 &&
          (result.problem.code === "authorization_denied" ||
            result.problem.code === "csrf_failed"))
      ) {
        const field = result.problem.field;
        const draft = this.state.draft;
        this.publish({
          operation: { kind: "rejected", attempt, problem: result.problem },
          fieldErrors:
            field &&
            attempt.fields.includes(field) &&
            draft &&
            draft.revisions[field] === attempt.revisions[field]
              ? {
                  [field]: {
                    revision: attempt.revisions[field],
                    message: metadataFieldError(result.problem),
                  },
                }
              : {},
          ...(result.problem.code === "incident_closed" ||
          result.problem.code === "authorization_denied"
            ? {
                access: "checking",
                draft: draft ? { ...draft, reviewRequired: true } : null,
              }
            : {}),
        });
        if (
          result.problem.code === "incident_closed" ||
          result.problem.code === "authorization_denied"
        )
          this.refresh();
        return;
      }
      this.publish({ operation: { kind: "uncertain", attempt } });
      return;
    }
    const resource = Object.freeze({ ...result.resource });
    const latest =
      this.state.resource &&
      this.state.resource.incident_version > resource.incident_version
        ? this.state.resource
        : resource;
    const draft = this.state.draft;
    const values = draft ? { ...draft.values } : metadataValues(latest);
    const saved = metadataValues(latest);
    if (draft)
      for (const field of metadataFields)
        if (
          draft.revisions[field] === attempt.revisions[field] ||
          (draft.resetRevision >= attempt.revision &&
            draft.revisions[field] <= draft.resetRevision)
        )
          values[field] = saved[field];
    this.publish({
      resource: latest,
      draft: draft
        ? {
            ...draft,
            base: latest,
            values,
            reviewRequired:
              (latest.incident_version > resource.incident_version ||
                this.state.authority?.role !== attempt.authority.role ||
                latest.status !== attempt.base.status) &&
              metadataFields.some((field) => values[field] !== saved[field]),
          }
        : newMetadataDraft(latest),
      operation: { kind: "confirmed", attempt, resource },
      fieldErrors: {},
    });
    this.ports.publishResource?.(latest, attempt.authority);
    this.refresh();
  }
  canReview = () =>
    this.current() &&
    this.state.active &&
    this.state.access === "ready" &&
    this.state.read === "ready" &&
    metadataEditRole(this.state.authority?.role) &&
    this.state.resource?.status === "active" &&
    !this.write &&
    !this.transport;
  review = () => {
    const draft = this.state.draft;
    const resource = this.state.resource;
    const op = this.state.operation;
    if (
      !draft ||
      !resource ||
      !this.canReview() ||
      (!draft.reviewRequired &&
        op.kind !== "uncertain" &&
        op.kind !== "conflicted")
    )
      return;
    const values = metadataValues(resource);
    for (const field of changedMetadataFields(draft))
      values[field] = draft.values[field];
    this.publish({
      draft: { ...draft, base: resource, values, reviewRequired: false },
      operation: { kind: "idle" },
      fieldErrors: {},
      reviewNotice:
        op.kind === "uncertain"
          ? "The previous save remains unconfirmed. Current values were reviewed; save explicitly to make a new attempt."
          : "Current values were reviewed. Save explicitly to apply your changes.",
    });
  };
  discard = () => {
    const draft = this.state.draft;
    const resource = this.state.resource;
    if (!draft || !resource || !this.current()) return;
    const revision = draft.revision + 1;
    const revisions = { ...draft.revisions };
    for (const field of metadataFields) revisions[field] = revision;
    this.publish({
      draft: {
        ...newMetadataDraft(resource),
        revision,
        revisions,
        resetRevision: revision,
      },
      fieldErrors: {},
      reviewNotice:
        this.state.operation.kind === "pending" ||
        this.state.operation.kind === "uncertain"
          ? "Local edits were discarded. This cannot cancel the previous server operation; its result still needs observation."
          : "Local edits discarded.",
    });
  };
  hasDepartureWork = () =>
    this.current() &&
    (metadataDirty(this.state.draft) ||
      this.state.operation.kind === "pending" ||
      this.state.operation.kind === "uncertain" ||
      this.state.operation.kind === "conflicted");
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
