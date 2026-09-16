import { requireViewContract } from "@cartulary/view-contracts";
import type { AuthorizationRecoveryResult } from "../../shared/authorizationRecovery";
import { validateDisplayName } from "../../shared/displayName";
import {
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  emptyWorkbookQueryState,
} from "../models/workbookQuery";
import { savedViewChanges } from "../models/workbookSavedViewRuntime";
import {
  canMutateSavedView,
  type SavedViewResource,
  savedViewLayoutJsonForPersistence,
  savedViewQueryJsonForPersistence,
} from "../models/workbookSavedViews";
import type {
  SavedViewProblem,
  SavedViewResult,
  WorkbookSavedViewDefinition,
} from "../ports/WorkbookSavedViewPort";
import {
  emptySavedViewDiscovery,
  SavedViewDiscovery,
} from "./SavedViewDiscovery";
import { SavedViewResourceObserver } from "./SavedViewResourceObserver";
import {
  type SavedViewAttempt,
  type SavedViewAuthority,
  type SavedViewBinding,
  type SavedViewControllerPorts,
  type SavedViewDraft,
  type SavedViewIntent,
  type SavedViewSnapshot,
  type SavedViewSubject,
  savedViewSubjectKey,
} from "./savedViewOperationModel";

const initial = (): SavedViewSnapshot => ({
  authority: null,
  access: "checking",
  observations: new Map(),
  discovery: emptySavedViewDiscovery(),
  activationId: null,
  refreshing: false,
  resourceProblem: null,
  observation: null,
  transportPending: false,
  operation: { kind: "idle" },
  drafts: new Map(),
  notice: null,
  reviewName: null,
});
const defaultDraft = (saved: SavedViewResource | null): SavedViewDraft => ({
  displayName: saved?.display_name ?? "Saved view",
  scope: saved?.scope === "shared" ? "shared" : "private",
  generation: 0,
  edited: false,
});
function freeze<T>(value: T): T {
  if (value !== null && typeof value === "object") {
    for (const entry of Object.values(value)) freeze(entry);
    Object.freeze(value);
  }
  return value;
}
function accessProblem(): SavedViewProblem {
  return {
    kind: "transport",
    message:
      "Current access could not be checked. Retry the access check before writing.",
  };
}

/** One current incident/session owns admission, receipts and recovery; UI lifetimes do not. */
export class WorkbookSavedViewController {
  private state = initial();
  private listeners = new Set<() => void>();
  private epoch = 0;
  private sequence = 0;
  private activationGeneration = 0;
  private resourceReadGeneration = 0;
  private binding: SavedViewBinding | null = null;
  private write: { cancel: () => void } | null = null;
  private transport: object | null = null;
  private accessRead: { cancel: () => void } | null = null;
  readonly discovery: SavedViewDiscovery;
  readonly resources: SavedViewResourceObserver;
  constructor(private readonly ports: SavedViewControllerPorts) {
    const port = () =>
      this.current() && this.state.authority
        ? this.ports.port(this.state.authority)
        : null;
    const changed = () => {
      const id = this.binding?.subject.savedViewId;
      const before = id ? this.state.observations.get(id)?.resource : null;
      const after = id ? this.resources.get(id)?.resource : null;
      this.publish({
        observations: this.resources.getSnapshot(),
        discovery: this.discovery.getSnapshot(),
        ...(before &&
        after &&
        after.saved_view_version > before.saved_view_version &&
        !this.write &&
        !this.transport &&
        this.state.activationId !== id
          ? {
              notice:
                "The saved configuration changed. Your working query and layout are unchanged; Reset restores the latest saved configuration.",
            }
          : {}),
      });
    };
    this.discovery = new SavedViewDiscovery({
      port,
      observe: ports.observe,
      changed,
      failed: (problem) => this.readFailed(null, problem),
    });
    this.resources = new SavedViewResourceObserver({
      port,
      observe: ports.observe,
      changed,
      visible: (resource) => this.current() && this.visible(resource),
      failed: (id, problem) => this.readFailed(id, problem),
    });
  }
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(patch: Partial<SavedViewSnapshot>) {
    this.state = { ...this.state, ...patch };
    if (patch.operation) {
      const operation = patch.operation;
      this.resources.retain(
        "operation",
        operation.kind === "idle" || operation.kind === "reviewed"
          ? null
          : operation.kind === "confirmed"
            ? (operation.resource?.saved_view_id ??
              operation.attempt.base?.saved_view_id ??
              null)
            : (operation.attempt.base?.saved_view_id ?? null),
      );
    }
    for (const listener of this.listeners) listener();
  }
  private current(
    authority = this.state.authority,
    epoch = this.epoch,
  ): authority is SavedViewAuthority {
    const now = this.state.authority;
    return (
      authority !== null &&
      now !== null &&
      epoch === this.epoch &&
      authority.incidentId === now.incidentId &&
      authority.actorId === now.actorId &&
      authority.lifetime === now.lifetime &&
      authority.apiBase === now.apiBase &&
      this.ports.isCurrent(authority)
    );
  }
  setAuthority = (authority: SavedViewAuthority | null) => {
    const previous = this.state.authority;
    if (!authority) {
      if (previous) this.retire();
      return;
    }
    if (
      !previous ||
      previous.incidentId !== authority.incidentId ||
      previous.actorId !== authority.actorId ||
      previous.lifetime !== authority.lifetime ||
      previous.apiBase !== authority.apiBase
    ) {
      this.retire();
      this.publish({ authority: freeze({ ...authority }), access: "ready" });
    } else if (previous.role !== authority.role) {
      const hidden = [...this.state.observations.values()]
        .flatMap((o) => (o.resource ? [o.resource] : []))
        .filter((resource) => !this.visible(resource, authority));
      const drafts = new Map(this.state.drafts);
      for (const resource of hidden)
        drafts.delete(
          savedViewSubjectKey({
            viewSchemaId: resource.view_schema_id,
            savedViewId: resource.saved_view_id,
          }),
        );
      const operation = this.state.operation;
      this.publish({
        authority: freeze({ ...authority }),
        access: "ready",
        drafts,
        ...(operation.kind !== "idle" &&
        operation.attempt.base &&
        !this.visible(operation.attempt.base, authority)
          ? { operation: { kind: "idle" } as const, reviewName: null }
          : {}),
        observation: null,
      });
      this.cancelRead();
      this.discovery.clear();
      this.discovery.setSchema(this.binding?.subject.viewSchemaId ?? null);
      for (const resource of hidden)
        this.resources.unavailable(resource.saved_view_id, {
          kind: "unavailable_target",
          message: "This saved view is no longer visible.",
        });
      if (
        hidden.some(
          (r) => r.saved_view_id === this.binding?.subject.savedViewId,
        )
      )
        this.binding?.unavailable();
      void this.refresh();
    }
  };
  private visible(
    resource: SavedViewResource,
    authority = this.state.authority,
  ) {
    return (
      authority !== null &&
      resource.incident_id === authority.incidentId &&
      (resource.scope !== "private" ||
        resource.owner_user_id === authority.actorId ||
        authority.role === "admin")
    );
  }
  setBinding = (binding: SavedViewBinding | null) => {
    if (binding && binding.incidentId !== this.state.authority?.incidentId)
      return;
    this.binding = binding;
    this.discovery.setSchema(binding?.subject.viewSchemaId ?? null);
    this.resources.retain("selected", binding?.subject.savedViewId ?? null);
    if (binding) {
      this.ensureDraft(binding.subject);
      const id = binding.subject.savedViewId;
      const observation = this.resources.get(id);
      if (
        id &&
        observation?.status === "unobserved" &&
        !observation.pending &&
        !observation.problem
      )
        void this.resources.read(id);
    } else this.closeDiscovery();
  };
  private selected(subject: SavedViewSubject) {
    const resource = this.resources.get(subject.savedViewId)?.resource;
    return resource?.view_schema_id === subject.viewSchemaId ? resource : null;
  }
  private ensureDraft(subject: SavedViewSubject) {
    const key = savedViewSubjectKey(subject);
    const current = this.state.drafts.get(key);
    const saved = this.selected(subject);
    const defaults = defaultDraft(saved);
    if (
      !current ||
      (!current.edited &&
        (current.displayName !== defaults.displayName ||
          current.scope !== defaults.scope))
    ) {
      const drafts = new Map(this.state.drafts);
      drafts.set(key, {
        ...defaults,
        generation: (current?.generation ?? -1) + 1,
      });
      this.publish({ drafts });
    }
  }
  draftFor = (subject: SavedViewSubject) =>
    this.state.drafts.get(savedViewSubjectKey(subject)) ??
    defaultDraft(this.selected(subject));
  changeDraft = (
    subject: SavedViewSubject,
    patch: Partial<Pick<SavedViewDraft, "displayName" | "scope">>,
  ) => {
    if (!this.subjectCurrent(subject)) return;
    const key = savedViewSubjectKey(subject);
    const old = this.draftFor(subject);
    const drafts = new Map(this.state.drafts);
    drafts.set(key, {
      ...old,
      ...patch,
      generation: old.generation + 1,
      edited: true,
    });
    this.publish({ drafts, notice: null });
  };
  private subjectCurrent(subject: SavedViewSubject) {
    const now = this.binding?.subject;
    return (
      this.current() &&
      now !== undefined &&
      now.viewSchemaId === subject.viewSchemaId &&
      now.savedViewId === subject.savedViewId &&
      now.savedViewVersion === subject.savedViewVersion
    );
  }
  private mutationBusy() {
    return (
      this.write !== null ||
      this.transport !== null ||
      ["pending", "conflict", "uncertain"].includes(this.state.operation.kind)
    );
  }
  unavailableReason = (
    kind: SavedViewIntent["kind"],
    subject: SavedViewSubject,
  ): string | null => {
    if (!this.subjectCurrent(subject))
      return "The saved-view action subject is no longer current.";
    const selected = this.selected(subject);
    if (
      subject.savedViewId !== null &&
      (!selected || selected.saved_view_version !== subject.savedViewVersion)
    )
      return "Select an available saved view or the base surface.";
    if (kind === "reset" && this.state.access === "ready") return null;
    if (this.mutationBusy())
      return "Finish the current saved-view action or review its recovery state first.";
    if (this.state.access !== "ready")
      return "Check current access before changing saved views.";
    if (kind === "create" || kind === "reset") return null;
    if (!selected) return "Select a saved view first.";
    if (
      (kind === "update" || kind === "delete") &&
      !canMutateSavedView(
        selected,
        this.state.authority?.actorId ?? null,
        this.state.authority?.role ?? null,
      )
    )
      return selected.scope === "system"
        ? "System saved views are immutable. You can duplicate this view."
        : "Only the owner or incident administrator can change this view.";
    return null;
  };
  run = (intent: SavedViewIntent, subject: SavedViewSubject) => {
    const unavailable = this.unavailableReason(intent.kind, subject);
    if (unavailable) {
      this.publish({ notice: unavailable });
      return;
    }
    const binding = this.binding;
    const authority = this.state.authority;
    if (!binding || !authority) return;
    const base = this.selected(subject);
    const contract = requireViewContract(subject.viewSchemaId);
    if (intent.kind === "reset") {
      binding.applyConfiguration(
        subject.viewSchemaId,
        base
          ? savedViewQueryJsonForPersistence(contract, base.query_json)
          : buildSavedViewQueryJson(contract, emptyWorkbookQueryState()),
        base
          ? savedViewLayoutJsonForPersistence(contract, base.layout_json)
          : buildSavedViewLayoutJson(contract),
      );
      this.publish({
        notice: base
          ? "Saved configuration restored. Source rows are unchanged."
          : "Default configuration restored. Source rows are unchanged.",
      });
      return;
    }
    const draft = this.draftFor(subject);
    const definition: WorkbookSavedViewDefinition =
      intent.kind === "duplicate" && base
        ? {
            displayName: `${base.display_name} Copy`,
            scope: "private",
            viewSchemaId: base.view_schema_id,
            queryJson: savedViewQueryJsonForPersistence(
              contract,
              base.query_json,
            ),
            layoutJson: savedViewLayoutJsonForPersistence(
              contract,
              base.layout_json,
            ),
          }
        : {
            displayName: draft.displayName,
            scope: draft.scope,
            viewSchemaId: subject.viewSchemaId,
            queryJson: binding.queryJson,
            layoutJson: binding.layoutJson,
          };
    const attempt = freeze(
      structuredClone({
        id: ++this.sequence,
        kind: intent.kind,
        authority,
        subject,
        base,
        definition,
        changes: base ? savedViewChanges(base, definition) : {},
        selectionGeneration: binding.selectionGeneration,
        workingGeneration: binding.workingGeneration,
        formGeneration: draft.generation,
      } satisfies SavedViewAttempt),
    );
    this.start(attempt);
  };
  private start(attempt: SavedViewAttempt) {
    this.publish({ reviewName: null });
    const name = validateDisplayName(attempt.definition.displayName);
    if (attempt.kind !== "delete" && name.error) {
      this.publish({
        operation: {
          kind: "rejected",
          attempt,
          problem: {
            kind: "validation",
            field: "display_name",
            message: name.error,
          },
        },
        notice: null,
      });
      return;
    }
    const authority = attempt.authority;
    const epoch = this.epoch;
    if (!this.current(authority, epoch) || this.transport || this.write) return;
    this.cancelRead();
    const admission = { cancel: () => {} };
    this.write = admission;
    this.publish({
      operation: { kind: "pending", attempt, stage: "authorization" },
      observation: null,
      notice: null,
    });
    const current = () => this.current(authority, epoch);
    const attemptCurrent = () =>
      current() &&
      this.state.operation.kind !== "idle" &&
      this.state.operation.attempt.id === attempt.id;
    let sent = false;
    const observation = this.ports.observe(async (signal) => {
      const admitted = () =>
        current() && this.write === admission && !signal.aborted;
      const access = await this.ports.recover(authority, signal, admitted);
      if (!admitted() || !this.acceptAccess(access, authority) || !admitted())
        return;
      const now = this.state.authority;
      // A role recheck can hide the captured source and retire its recovery state.
      if (attempt.base && !this.visible(attempt.base)) return;
      if (
        (attempt.kind === "update" || attempt.kind === "delete") &&
        !canMutateSavedView(
          attempt.base,
          now?.actorId ?? null,
          now?.role ?? null,
        )
      ) {
        if (attemptCurrent())
          this.publish({
            operation: {
              kind: "rejected",
              attempt,
              problem: {
                kind: "authorization_denied",
                message:
                  "Current access does not permit changing this saved view.",
              },
            },
          });
        return;
      }
      sent = true;
      const transport = {};
      this.transport = transport;
      this.publish({
        transportPending: true,
        operation: { kind: "pending", attempt, stage: "write" },
      });
      try {
        const port = this.ports.port(authority);
        const result =
          attempt.kind === "update" && attempt.base
            ? await port.patch({
                base: attempt.base,
                changes: attempt.changes,
                signal,
              })
            : attempt.kind === "delete" && attempt.base
              ? await port.delete({
                  savedViewId: attempt.base.saved_view_id,
                  scope: attempt.base.scope,
                  signal,
                })
              : await port.create({ definition: attempt.definition, signal });
        if (attemptCurrent()) this.complete(attempt, result);
      } finally {
        if (current() && this.transport === transport) {
          this.transport = null;
          this.publish({ transportPending: false });
        }
      }
    });
    admission.cancel = observation.cancel;
    void observation.result.then((outcome) => {
      if (!current() || this.write !== admission) return;
      this.write = null;
      if (
        outcome.kind !== "completed" &&
        sent &&
        attemptCurrent() &&
        this.state.operation.kind === "pending"
      ) {
        this.publish({
          operation: {
            kind: "uncertain",
            attempt,
            problem: {
              kind: "transport",
              message:
                outcome.kind === "timeout"
                  ? "The observation deadline expired. This does not establish server cancellation."
                  : "The write could not be confirmed.",
            },
          },
          observation: null,
        });
      } else if (
        !sent &&
        attemptCurrent() &&
        this.state.operation.kind === "pending"
      ) {
        this.publish({
          access: "unavailable",
          operation: { kind: "rejected", attempt, problem: accessProblem() },
        });
      }
      if (sent && this.transport === null) this.followup(attempt);
    });
    void observation.settled.then(() => {
      if (current() && sent && this.write === null && this.transport === null)
        this.followup(attempt);
    });
  }
  private followedAttempt = 0;
  private followup(attempt: SavedViewAttempt) {
    if (this.followedAttempt >= attempt.id) return;
    this.followedAttempt = attempt.id;
    const op = this.state.operation;
    if (op.kind === "idle" || op.attempt.id !== attempt.id) return;
    if (
      op.kind === "rejected" &&
      [
        "authentication_required",
        "authorization_denied",
        "unavailable_target",
      ].includes(op.problem.kind)
    )
      void this.recheckAccess();
    else void this.refresh();
  }
  private complete(
    attempt: SavedViewAttempt,
    result: SavedViewResult<SavedViewResource | undefined>,
  ) {
    this.cancelRead();
    if (result.kind !== "accepted") {
      this.publish({
        operation: {
          kind:
            result.kind === "uncertain"
              ? "uncertain"
              : result.failure.kind === "conflict"
                ? "conflict"
                : "rejected",
          attempt,
          problem: result.failure,
        },
        observation: null,
      });
      return;
    }
    const resource = result.value ?? null;
    if (
      attempt.kind !== "delete" &&
      (!resource ||
        !this.visible(resource) ||
        ((attempt.kind === "create" || attempt.kind === "duplicate") &&
          resource.owner_user_id !== attempt.authority.actorId))
    ) {
      this.publish({
        operation: {
          kind: "uncertain",
          attempt,
          problem: {
            kind: "invalid_contract",
            message:
              "The acknowledgement does not match the current saved-view authority.",
          },
        },
        observation: null,
      });
      return;
    }
    const existing = resource
      ? this.resources.get(resource.saved_view_id)?.resource
      : null;
    const isNewer =
      !resource ||
      !existing ||
      existing.saved_view_version <= resource.saved_view_version;
    const apply = this.canApply(attempt);
    if (resource && isNewer) {
      this.resources.retain("operation", resource.saved_view_id);
      this.resources.accept(resource);
    }
    if (attempt.kind === "delete" && attempt.base) {
      this.resources.unavailable(attempt.base.saved_view_id, {
        kind: "unavailable_target",
        message: "Saved view deleted.",
      });
      this.discovery.invalidate(attempt.base.saved_view_id);
    }
    this.publish({
      operation: {
        kind: "confirmed",
        attempt,
        resource: resource ? freeze(structuredClone(resource)) : null,
      },
      observation: null,
    });
    // UI effects follow confirmation and are guarded independently from it.
    if (
      attempt.kind === "delete" &&
      attempt.base &&
      this.binding?.subject.savedViewId === attempt.base.saved_view_id &&
      this.binding.subject.viewSchemaId === attempt.base.view_schema_id
    ) {
      // Identity fallback leaves newer working/form edits intact.
      this.binding.deleted(attempt.base);
    } else if (apply && isNewer && this.binding) {
      if (resource && attempt.kind === "update")
        this.binding.applyConfiguration(
          resource.view_schema_id,
          savedViewQueryJsonForPersistence(
            requireViewContract(resource.view_schema_id),
            resource.query_json,
          ),
          savedViewLayoutJsonForPersistence(
            requireViewContract(resource.view_schema_id),
            resource.layout_json,
          ),
        );
      else if (resource) {
        this.resources.retain("selected", resource.saved_view_id);
        this.binding.select(resource);
      }
    }
    const key = savedViewSubjectKey(attempt.subject);
    const draft = this.state.drafts.get(key);
    if (
      resource &&
      draft?.generation === attempt.formGeneration &&
      attempt.kind !== "duplicate"
    ) {
      const drafts = new Map(this.state.drafts);
      drafts.set(key, {
        displayName: resource.display_name,
        scope: resource.scope === "shared" ? "shared" : "private",
        generation: draft.generation + 1,
        edited: false,
      });
      this.publish({ drafts });
    }
  }
  private canApply(attempt: SavedViewAttempt) {
    const binding = this.binding;
    return (
      binding !== null &&
      this.subjectCurrent(attempt.subject) &&
      binding.selectionGeneration === attempt.selectionGeneration &&
      binding.workingGeneration === attempt.workingGeneration &&
      this.draftFor(attempt.subject).generation === attempt.formGeneration
    );
  }
  acceptResource = (resource: SavedViewResource) => {
    if (!this.current() || !this.visible(resource)) return;
    this.resources.retain("selected", resource.saved_view_id);
    this.resources.accept(resource);
  };
  openDiscovery = () => this.discovery.open();
  closeDiscovery = () => {
    ++this.activationGeneration;
    const id = this.state.activationId;
    if (id) this.resources.cancel(id);
    this.resources.retain("activation", null);
    this.publish({ activationId: null });
    this.discovery.close();
  };
  activateResource = async (
    id: string,
    schema = this.binding?.subject.viewSchemaId,
  ): Promise<boolean> => {
    const binding = this.binding;
    if (!binding || !this.current()) return false;
    const generation = ++this.activationGeneration;
    const epoch = this.epoch;
    this.resources.retain("activation", id);
    this.publish({ activationId: id, notice: null });
    const resource = await this.resources.read(id);
    if (generation !== this.activationGeneration || epoch !== this.epoch)
      return false;
    const current = this.binding;
    const apply =
      resource &&
      resource.view_schema_id === schema &&
      current &&
      this.current() &&
      current.selectionGeneration === binding.selectionGeneration &&
      current.workingGeneration === binding.workingGeneration;
    if (apply) {
      this.resources.retain("selected", id);
      current.select(resource);
    } else if (resource)
      this.publish({
        notice:
          "The working configuration changed while this view was loading. Select the view again to apply it.",
      });
    this.resources.retain("activation", null);
    this.publish({ activationId: null });
    return !!apply;
  };
  observePreference = (slot: "home" | "default", id: string | null) => {
    this.resources.retain(slot, id);
    const observation = this.resources.get(id);
    if (
      id &&
      observation?.status === "unobserved" &&
      !observation.pending &&
      !observation.problem
    )
      void this.resources.read(id);
  };
  private readFailed(id: string | null, problem: SavedViewProblem) {
    if (!this.current()) return;
    if (
      [
        "authentication_required",
        "authorization_denied",
        "unavailable_target",
      ].includes(problem.kind)
    )
      void this.recheckAccess();
    if (id !== null && id === this.state.activationId)
      this.publish({ notice: problem.message });
  }
  private acceptAccess(
    result: AuthorizationRecoveryResult,
    authority: SavedViewAuthority,
  ) {
    if (!this.current(authority)) return false;
    if (result.kind === "authorized" && result.userId === authority.actorId) {
      this.setAuthority({ ...authority, role: result.role });
      this.publish({ access: "ready" });
      this.binding?.authorizationRecovered(result);
      return true;
    }
    if (
      result.kind === "session_lost" ||
      result.kind === "access_lost" ||
      (result.kind === "authorized" && result.userId !== authority.actorId)
    ) {
      this.retire();
      this.ports.lost(
        result.kind === "access_lost" ? "incident" : "session",
        authority,
      );
      return false;
    }
    if (result.kind !== "cancelled") this.publish({ access: "unavailable" });
    return false;
  }
  recheckAccess = async () => {
    const authority = this.state.authority;
    const epoch = this.epoch;
    if (!authority || !this.current(authority, epoch)) return;
    if (this.accessRead) return;
    const admission = { cancel: () => {} };
    this.accessRead = admission;
    this.publish({ access: "checking", observation: null });
    const observation = this.ports.observe((signal) =>
      this.ports.recover(
        authority,
        signal,
        () => this.current(authority, epoch) && !signal.aborted,
      ),
    );
    admission.cancel = observation.cancel;
    const outcome = await observation.result;
    if (!this.current(authority, epoch) || this.accessRead !== admission)
      return;
    this.accessRead = null;
    if (
      outcome.kind === "completed" &&
      this.acceptAccess(outcome.value, authority)
    )
      this.fallbackUnavailableSelection();
    else if (this.current(authority, epoch))
      this.publish({ access: "unavailable" });
  };
  private fallbackUnavailableSelection() {
    const id = this.binding?.subject.savedViewId;
    if (id && this.resources.get(id)?.status === "unavailable") {
      this.binding?.unavailable();
      this.publish({
        notice:
          "The selected saved view is no longer available. Showing the base surface with your working configuration.",
      });
    }
  }
  private cancelRead() {
    ++this.resourceReadGeneration;
    this.resources.invalidate();
    this.discovery.invalidate();
    ++this.activationGeneration;
    this.resources.retain("activation", null);
    this.publish({ activationId: null, refreshing: false });
  }
  refresh = async () => {
    if (!this.current()) return;
    const generation = ++this.resourceReadGeneration;
    const op = this.state.operation;
    const id =
      op.kind !== "idle" && op.kind !== "reviewed"
        ? op.kind === "confirmed"
          ? op.resource?.saved_view_id
          : op.attempt.base?.saved_view_id
        : this.binding?.subject.savedViewId;
    this.publish({ observation: null, resourceProblem: null });
    if (!id || (op.kind === "confirmed" && op.attempt.kind === "delete"))
      return;
    this.publish({ refreshing: true });
    const epoch = this.epoch;
    const resource = await this.resources.read(id);
    if (epoch !== this.epoch || generation !== this.resourceReadGeneration)
      return;
    const observation = this.resources.get(id);
    this.publish({
      refreshing: false,
      resourceProblem: observation?.problem ?? null,
      observation:
        resource && !this.transport && !this.write
          ? (observation?.revision ?? null)
          : null,
    });
    if (this.binding) this.ensureDraft(this.binding.subject);
  };
  canReview = (attemptId: number, _observation?: number | null) => {
    const operation = this.state.operation;
    return (
      this.current() &&
      this.state.access === "ready" &&
      !this.transport &&
      !this.write &&
      ["conflict", "uncertain", "rejected"].includes(operation.kind) &&
      operation.kind !== "idle" &&
      operation.attempt.id === attemptId
    );
  };
  setReviewName = (attemptId: number, value: string) => {
    const operation = this.state.operation;
    if (
      this.current() &&
      operation.kind !== "idle" &&
      operation.attempt.id === attemptId &&
      ["conflict", "uncertain", "rejected"].includes(operation.kind)
    )
      this.publish({ reviewName: value });
  };
  review = (
    attemptId: number,
    observation: number | null,
    choice: "apply" | "create" | "keep",
    name?: string,
  ) => {
    if (!this.canReview(attemptId, observation)) return;
    const operation = this.state.operation;
    if (operation.kind === "idle" || !this.state.authority) return;
    const attempt = operation.attempt;
    if (choice === "keep") {
      this.publish({
        operation: { kind: "reviewed", attempt },
        notice: "Recovery ended. Your working configuration was not changed.",
      });
      return;
    }
    const observed = attempt.base
      ? this.resources.get(attempt.base.saved_view_id)?.resource
      : null;
    if (
      choice === "apply" &&
      (!this.canApplyReview(attemptId, observation) ||
        (attempt.kind !== "update" && attempt.kind !== "delete") ||
        !observed ||
        !canMutateSavedView(
          observed,
          this.state.authority.actorId,
          this.state.authority.role,
        ))
    )
      return;
    let definition = attempt.definition;
    if (choice === "apply" && observed) {
      const contract = requireViewContract(observed.view_schema_id);
      definition = {
        viewSchemaId: observed.view_schema_id,
        displayName: attempt.changes.displayName ?? observed.display_name,
        scope:
          attempt.changes.scope ??
          (observed.scope === "shared" ? "shared" : "private"),
        queryJson:
          attempt.changes.queryJson ??
          savedViewQueryJsonForPersistence(contract, observed.query_json),
        layoutJson:
          attempt.changes.layoutJson ??
          savedViewLayoutJsonForPersistence(contract, observed.layout_json),
      };
    }
    if (name !== undefined) definition = { ...definition, displayName: name };
    // A new explicitly reviewed request; never an idempotent replay or silent rebase.
    const next = freeze(
      structuredClone({
        ...attempt,
        id: ++this.sequence,
        kind:
          choice === "create"
            ? "create"
            : attempt.kind === "delete"
              ? "delete"
              : "update",
        authority: this.state.authority,
        base: choice === "create" ? null : (observed ?? null),
        definition,
        changes:
          choice === "apply" && observed
            ? savedViewChanges(observed, definition)
            : {},
        // Recovery itself never navigates or applies over current working edits.
        selectionGeneration: -1,
        workingGeneration: -1,
      } satisfies SavedViewAttempt),
    );
    this.start(next);
  };
  dismiss = () => {
    if (
      !this.transport &&
      !this.write &&
      !["pending", "conflict", "uncertain"].includes(this.state.operation.kind)
    )
      this.publish({
        operation: { kind: "idle" },
        notice: null,
        reviewName: null,
      });
  };
  canApplyReview = (attemptId: number, observation: number | null) => {
    const op = this.state.operation;
    const id =
      op.kind === "idle" ? null : (op.attempt.base?.saved_view_id ?? null);
    const current = this.resources.get(id);
    return (
      this.canReview(attemptId) &&
      observation !== null &&
      observation === this.state.observation &&
      observation === current?.revision &&
      current.status === "ready" &&
      !current.pending &&
      !current.problem
    );
  };
  openConfirmed = async () => {
    const operation = this.state.operation;
    if (operation.kind === "confirmed" && operation.resource && this.current())
      await this.activateResource(
        operation.resource.saved_view_id,
        operation.resource.view_schema_id,
      );
  };
  retire = () => {
    this.epoch += 1;
    this.cancelRead();
    this.accessRead?.cancel();
    this.accessRead = null;
    this.write?.cancel();
    this.write = null;
    this.transport = null;
    this.binding = null;
    this.resources.clear();
    this.discovery.clear();
    this.publish(initial());
  };
  dispose = () => {
    this.retire();
    this.listeners.clear();
  };
}
