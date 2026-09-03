import type {
  IncidentCollaborationEvent,
  IncidentCollaborationStatus,
} from "../../collaboration/IncidentCollaborationSession";
import type {
  AuthorizationRecoveryPort,
  AuthorizationRecoveryResult,
} from "../../shared/authorizationRecovery";
import type { SheetRef } from "../../shared/sheetRef";
import { sheetRefKey } from "../../shared/sheetRef";
import type {
  WorkbookDependentPresentationInvalidationReason,
  WorkbookInvalidationReason,
  WorkbookQueryInvalidationReason,
} from "../lifecycle/workbookInvalidation";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { PresenceRecord } from "../utils/workbookPresence";
import {
  beginWorkbookAuthorizationRecovery,
  completeWorkbookAuthorizationRecovery,
  initialWorkbookAuthorizationRecoveryMachine,
  planWorkbookAuthorizationRecoveryResult,
  retryWorkbookAuthorizationRecovery,
  scheduleWorkbookAuthorizationRecovery,
  terminateWorkbookAuthorizationRecovery,
  type WorkbookAuthorizationRecoveryAdmission,
  type WorkbookAuthorizationRecoveryResultPlan,
} from "./workbookAuthorizationRecoveryMachine";
import {
  planWorkbookCollaborationEvent,
  type WorkbookCollaborationEventPlan,
} from "./workbookCollaborationEventPlan";
import {
  planWorkbookCollaborationInvalidation,
  type WorkbookCollaborationInvalidationEffect,
} from "./workbookCollaborationInvalidationPlan";
import {
  buildWorkbookPresenceInput,
  type RecordChangedPayload,
  type WorkbookPresenceDraft,
} from "./workbookCollaborationMessages";
import {
  beginWorkbookCollaborationReset,
  cancelWorkbookCollaborationReset,
  initialWorkbookCollaborationResetMachine,
  type WorkbookCollaborationResetAdmission,
  workbookCollaborationResetIsCurrent,
} from "./workbookCollaborationResetMachine";
import type {
  WorkbookCollaborationClock,
  WorkbookCollaborationScheduledTask,
  WorkbookCollaborationScheduler,
} from "./workbookCollaborationTiming";
import {
  activeWorkbookPresenceRecords,
  applyWorkbookPresenceDelta,
  clearWorkbookPresenceProjection,
  initialWorkbookPresenceProjection,
  replaceWorkbookPresenceSnapshot,
} from "./workbookPresenceProjection";
import {
  cancelWorkbookPresencePublication,
  initialWorkbookPresencePublicationMachine,
  scheduleWorkbookPresencePublication,
  settleWorkbookPresencePublication,
} from "./workbookPresencePublicationMachine";
import type { WorkbookActiveSurfacePort } from "./workbookSurfacePort";

type CollaborationProjectionListener = () => void;
type AuthorizedRecoveryPlan = Extract<
  WorkbookAuthorizationRecoveryResultPlan,
  { readonly kind: "authorized" }
>;
type LifecycleEventPlan = Extract<
  WorkbookCollaborationEventPlan,
  {
    readonly kind:
      | "established"
      | "reset"
      | "recover_authorization"
      | "incident_closed";
  }
>;
type PresenceEventPlan = Extract<
  WorkbookCollaborationEventPlan,
  { readonly kind: "presence_snapshot" | "presence_delta" }
>;

type WorkbookCollaborationSessionPort = {
  readonly completeReset: (generation: number) => boolean;
  readonly connectionId: string | null;
  readonly publishPresence: (
    presence: ReturnType<typeof buildWorkbookPresenceInput>,
  ) => void;
  readonly reconnect: () => void;
  readonly status: IncidentCollaborationStatus;
  readonly subscribe: (
    listener: (event: IncidentCollaborationEvent) => void,
  ) => () => void;
};

export type WorkbookCollaborationSnapshot = {
  readonly activeSheetPresenceRecords: readonly PresenceRecord[];
  readonly connectionId: string | null;
  readonly status: IncidentCollaborationStatus;
};

type WorkbookCollaborationCoordinatorOptions = {
  readonly authorizationRecovery: AuthorizationRecoveryPort;
  readonly clock: WorkbookCollaborationClock;
  readonly continuityInvalidation: (
    reason: WorkbookDependentPresentationInvalidationReason,
  ) => void;
  readonly evidenceInvalidation: (
    reason: WorkbookDependentPresentationInvalidationReason,
  ) => void;
  readonly extensionInvalidation: (
    reason:
      | Extract<
          WorkbookInvalidationReason,
          { readonly kind: "session_unavailable" }
        >
      | Extract<
          WorkbookInvalidationReason,
          { readonly kind: "incident_access_lost" }
        >,
  ) => void;
  readonly incidentId: string;
  readonly initialSheetRef: SheetRef;
  readonly inspectorInvalidation: (
    reason: WorkbookDependentPresentationInvalidationReason,
  ) => void;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onAuthorizationRecovered: (
    result: Extract<
      AuthorizationRecoveryResult,
      { readonly kind: "authorized" }
    >,
  ) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly queryInvalidation: (reason: WorkbookQueryInvalidationReason) => void;
  readonly scheduler: WorkbookCollaborationScheduler;
};

/**
 * Shell-lifetime effect authority for collaboration interpretation.
 * The incident session remains the sole transport and replay-sequence owner.
 */
class WorkbookCollaborationCoordinatorRuntime {
  private activePort: WorkbookActiveSurfacePort | null = null;
  private activeSheetRef: SheetRef;
  private authorizationRecoveryController: AbortController | null = null;
  private authorizationRecoveryMachine =
    initialWorkbookAuthorizationRecoveryMachine();
  private authorizationRecoveryTask: WorkbookCollaborationScheduledTask | null =
    null;
  private disposed = false;
  private readonly dirtySurfaceKeys = new Set<string>();
  private readonly listeners = new Set<CollaborationProjectionListener>();
  private readonly clientTxnResolvers = new Set<
    (clientTxnId: string | null | undefined) => boolean
  >();
  private presenceProjection = initialWorkbookPresenceProjection();
  private presenceDraft: WorkbookPresenceDraft = {
    fieldKey: null,
    mode: "viewing",
    recordId: null,
  };
  private presencePublicationMachine =
    initialWorkbookPresencePublicationMachine();
  private presenceTask: WorkbookCollaborationScheduledTask | null = null;
  private resetMachine = initialWorkbookCollaborationResetMachine();
  private resetRetryTask: WorkbookCollaborationScheduledTask | null = null;
  private retainCount = 0;
  private retainGeneration = 0;
  private session: WorkbookCollaborationSessionPort | null = null;
  private sessionGeneration = 0;
  private sessionOwnerSubscribe:
    | WorkbookCollaborationSessionPort["subscribe"]
    | null = null;
  private sessionUnsubscribe: (() => void) | null = null;
  private snapshot: WorkbookCollaborationSnapshot;

  constructor(
    private readonly options: WorkbookCollaborationCoordinatorOptions,
  ) {
    this.activeSheetRef = { ...options.initialSheetRef };
    this.snapshot = {
      activeSheetPresenceRecords: [],
      connectionId: null,
      status: "disconnected",
    };
  }

  getSnapshot = (): WorkbookCollaborationSnapshot => this.snapshot;

  subscribe = (listener: CollaborationProjectionListener): (() => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };

  retain(): () => void {
    if (this.disposed) return () => undefined;
    this.retainCount += 1;
    this.retainGeneration += 1;
    let released = false;
    return () => {
      if (released) return;
      released = true;
      this.retainCount = Math.max(0, this.retainCount - 1);
      const releaseGeneration = ++this.retainGeneration;
      queueMicrotask(() => {
        if (
          this.retainCount === 0 &&
          this.retainGeneration === releaseGeneration
        ) {
          this.dispose();
        }
      });
    };
  }

  attachSession(session: WorkbookCollaborationSessionPort): () => void {
    if (this.disposed) return () => undefined;
    if (this.sessionOwnerSubscribe !== session.subscribe) {
      this.sessionGeneration += 1;
      this.sessionOwnerSubscribe = session.subscribe;
      this.cancelReset();
    }
    this.sessionUnsubscribe?.();
    this.session = session;
    this.sessionUnsubscribe = session.subscribe((event) => {
      this.handleEventPlan(planWorkbookCollaborationEvent(event));
    });
    this.emit();
    this.publishPresenceNow();
    return () => {
      if (this.session !== session) return;
      this.sessionUnsubscribe?.();
      this.sessionUnsubscribe = null;
      this.session = null;
    };
  }

  setActiveSheet(sheetRef: SheetRef): void {
    if (sheetRefKey(this.activeSheetRef) === sheetRefKey(sheetRef)) return;
    this.activeSheetRef = { ...sheetRef };
    this.presenceDraft = {
      fieldKey: null,
      mode: "viewing",
      recordId: null,
    };
    this.cancelPresenceTask();
    this.cancelReset();
    this.publishPresenceNow();
    this.emit();
  }

  registerActiveSurface(port: WorkbookActiveSurfacePort): () => void {
    this.activePort = port;
    const key = sheetRefKey(port.identity.sheetRef);
    if (this.dirtySurfaceKeys.delete(key)) {
      void port
        .refresh({ reason: "inactive_surface_reconciliation" })
        .catch(() => {
          this.dirtySurfaceKeys.add(key);
        });
    }
    return () => {
      if (this.activePort === port) this.activePort = null;
    };
  }

  registerClientTxnResolver(
    resolver: (clientTxnId: string | null | undefined) => boolean,
  ): () => void {
    this.clientTxnResolvers.add(resolver);
    return () => {
      this.clientTxnResolvers.delete(resolver);
    };
  }

  requestAuthorizationRecovery(): void {
    if (
      this.disposed ||
      !this.authorizationRecoveryMachine.authorizationConfirmed
    ) {
      return;
    }
    this.cancelReset();
    this.cancelAuthorizationWork();
    this.applyInvalidationPlan({ kind: "session_unavailable" });
    this.authorizationRecoveryMachine = scheduleWorkbookAuthorizationRecovery(
      this.authorizationRecoveryMachine,
      this.options.clock.nowMs(),
    );
    this.scheduleCurrentAuthorizationRecovery();
  }

  publishPresence(presence: WorkbookPresenceDraft): void {
    if (this.disposed) return;
    this.presenceDraft = { ...presence };
    if (!this.authorizationRecoveryMachine.authorizationConfirmed) return;
    const scheduled = scheduleWorkbookPresencePublication(
      this.presencePublicationMachine,
      this.options.clock.nowMs(),
    );
    this.presencePublicationMachine = scheduled.machine;
    this.presenceTask?.cancel();
    this.presenceTask = this.options.scheduler.schedule(
      scheduled.dueAtMs - this.options.clock.nowMs(),
      () => {
        this.presenceTask = null;
        const settled = settleWorkbookPresencePublication(
          this.presencePublicationMachine,
          scheduled.generation,
        );
        if (settled.kind === "stale" || this.disposed) return;
        this.presencePublicationMachine = settled.machine;
        this.publishPresenceNow();
      },
    );
  }

  editingPresenceForCell(
    recordId: string | null,
    fieldKey: string,
  ): readonly PresenceRecord[] {
    if (recordId === null) return [];
    return this.snapshot.activeSheetPresenceRecords.filter(
      (presence) =>
        presence.record_id === recordId &&
        presence.field_key === fieldKey &&
        presence.mode === "editing",
    );
  }

  dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    this.sessionUnsubscribe?.();
    this.sessionUnsubscribe = null;
    this.session = null;
    this.sessionOwnerSubscribe = null;
    this.cancelPresenceTask();
    this.cancelReset();
    this.cancelAuthorizationWork();
    this.authorizationRecoveryMachine = terminateWorkbookAuthorizationRecovery(
      this.authorizationRecoveryMachine,
    );
    this.presenceProjection = clearWorkbookPresenceProjection(
      this.presenceProjection,
    );
    this.applyInvalidationPlan({ kind: "runtime_disposed" });
    this.listeners.clear();
    this.clientTxnResolvers.clear();
  }

  private emit(): void {
    const connectionId = this.session?.connectionId ?? null;
    this.snapshot = {
      activeSheetPresenceRecords: activeWorkbookPresenceRecords({
        activeSheetRef: this.activeSheetRef,
        connectionId,
        projection: this.presenceProjection,
      }),
      connectionId,
      status: this.session?.status ?? "disconnected",
    };
    for (const listener of this.listeners) listener();
  }

  private publishPresenceNow(): void {
    if (this.disposed) return;
    this.session?.publishPresence(
      buildWorkbookPresenceInput(this.presenceDraft, this.activeSheetRef),
    );
  }

  private cancelPresenceTask(): void {
    this.presenceTask?.cancel();
    this.presenceTask = null;
    this.presencePublicationMachine = cancelWorkbookPresencePublication(
      this.presencePublicationMachine,
    );
  }

  private refreshActiveSurface(
    reason: string,
    recordId?: string,
  ): Promise<void> {
    const port = this.activePort;
    if (
      port === null ||
      sheetRefKey(port.identity.sheetRef) !== sheetRefKey(this.activeSheetRef)
    ) {
      this.dirtySurfaceKeys.add(sheetRefKey(this.activeSheetRef));
      return Promise.resolve();
    }
    return port.refresh({ reason, ...(recordId ? { recordId } : {}) });
  }

  private queueActiveSurfaceRefresh(reason: string, recordId?: string): void {
    const key = sheetRefKey(this.activeSheetRef);
    void this.refreshActiveSurface(reason, recordId).catch(() => {
      this.dirtySurfaceKeys.add(key);
    });
  }

  private applyInvalidationPlan(reason: WorkbookInvalidationReason): void {
    for (const effect of planWorkbookCollaborationInvalidation(reason)) {
      this.applyInvalidationEffect(effect);
    }
  }

  private applyInvalidationEffect(
    effect: WorkbookCollaborationInvalidationEffect,
  ): void {
    switch (effect.kind) {
      case "mutation":
        this.options.mutationRuntime.invalidate(effect.reason);
        return;
      case "query":
        this.options.queryInvalidation(effect.reason);
        return;
      case "active_surface":
        this.activePort?.invalidate(effect.reason);
        return;
      case "extension":
        this.options.extensionInvalidation(effect.reason);
        return;
      case "presence":
        this.presenceProjection = clearWorkbookPresenceProjection(
          this.presenceProjection,
        );
        this.emit();
        return;
      case "inspector":
        this.options.inspectorInvalidation(effect.reason);
        return;
      case "continuity":
        this.options.continuityInvalidation(effect.reason);
        return;
      case "evidence":
        this.options.evidenceInvalidation(effect.reason);
        return;
    }
  }

  private cancelAuthorizationWork(): void {
    this.authorizationRecoveryTask?.cancel();
    this.authorizationRecoveryTask = null;
    this.authorizationRecoveryController?.abort();
    this.authorizationRecoveryController = null;
  }

  private scheduleCurrentAuthorizationRecovery(): void {
    const machine = this.authorizationRecoveryMachine;
    if (machine.phase !== "scheduled" || machine.scheduledForMs === null)
      return;
    const generation = machine.generation;
    this.authorizationRecoveryTask = this.options.scheduler.schedule(
      machine.scheduledForMs - this.options.clock.nowMs(),
      () => {
        this.authorizationRecoveryTask = null;
        const begun = beginWorkbookAuthorizationRecovery(
          this.authorizationRecoveryMachine,
          generation,
        );
        this.authorizationRecoveryMachine = begun.machine;
        if (begun.kind === "recover") {
          void this.runAuthorizationRecovery(begun.admission);
        }
      },
    );
  }

  private async runAuthorizationRecovery(
    admission: WorkbookAuthorizationRecoveryAdmission,
  ): Promise<void> {
    const result = await this.loadAuthorizationRecovery();
    if (result === null) return;
    this.handleAuthorizationRecoveryResult(admission, result);
  }

  private async loadAuthorizationRecovery(): Promise<AuthorizationRecoveryResult | null> {
    const controller = new AbortController();
    this.authorizationRecoveryController = controller;
    let result: AuthorizationRecoveryResult = { kind: "unavailable" };
    try {
      result = await this.options.authorizationRecovery.recover({
        incidentId: this.options.incidentId,
        signal: controller.signal,
      });
    } catch {
      if (controller.signal.aborted) return null;
    }
    if (this.authorizationRecoveryController === controller) {
      this.authorizationRecoveryController = null;
    }
    return this.disposed || controller.signal.aborted ? null : result;
  }

  private handleAuthorizationRecoveryResult(
    admission: WorkbookAuthorizationRecoveryAdmission,
    result: AuthorizationRecoveryResult,
  ): void {
    const plan = planWorkbookAuthorizationRecoveryResult(
      this.authorizationRecoveryMachine,
      admission,
      result,
      this.options.clock.nowMs(),
    );
    this.authorizationRecoveryMachine = plan.machine;
    if (plan.kind === "stale") return;
    if (plan.kind === "retry") {
      this.scheduleCurrentAuthorizationRecovery();
      return;
    }
    if (plan.kind === "access_lost") {
      this.cancelReset();
      this.applyInvalidationPlan({ kind: "incident_access_lost" });
      this.options.onIncidentAccessLost?.();
      return;
    }
    this.options.onAuthorizationRecovered(plan.result);
    if (!plan.canResumeMutations) {
      this.applyInvalidationPlan({
        kind: "incident_role_changed",
        role: plan.result.role,
      });
    }
    void this.settleAuthorizedRecovery(plan);
  }

  private async settleAuthorizedRecovery(
    plan: AuthorizedRecoveryPlan,
  ): Promise<void> {
    try {
      await this.refreshActiveSurface("authorization_recovered");
    } catch {
      this.authorizationRecoveryMachine = retryWorkbookAuthorizationRecovery(
        this.authorizationRecoveryMachine,
        plan.admission,
        this.options.clock.nowMs(),
      );
      this.scheduleCurrentAuthorizationRecovery();
      return;
    }
    const completed = completeWorkbookAuthorizationRecovery(
      this.authorizationRecoveryMachine,
      plan.admission,
    );
    this.authorizationRecoveryMachine = completed.machine;
    if (completed.kind === "stale" || this.disposed) return;
    if (completed.canResumeMutations) {
      this.options.mutationRuntime.applyAuthorizationRecoveryState("resumed");
    }
    this.session?.reconnect();
  }

  private cancelReset(): void {
    this.resetRetryTask?.cancel();
    this.resetRetryTask = null;
    this.resetMachine = cancelWorkbookCollaborationReset(this.resetMachine);
  }

  private beginReset(
    eventGeneration: number,
    reason: "resume_reset" | "sequence_gap",
  ): void {
    if (
      this.disposed ||
      !this.authorizationRecoveryMachine.authorizationConfirmed
    ) {
      return;
    }
    this.cancelReset();
    this.applyInvalidationPlan({ kind: "collaboration_reset_required" });
    const started = beginWorkbookCollaborationReset(this.resetMachine, {
      eventGeneration,
      sessionGeneration: this.sessionGeneration,
      sheetKey: sheetRefKey(this.activeSheetRef),
    });
    this.resetMachine = started.machine;
    void this.runResetRefresh(started.admission, reason);
  }

  private resetIsCurrent(admission: WorkbookCollaborationResetAdmission) {
    return (
      !this.disposed &&
      workbookCollaborationResetIsCurrent(this.resetMachine, admission, {
        sessionGeneration: this.sessionGeneration,
        sheetKey: sheetRefKey(this.activeSheetRef),
      })
    );
  }

  private async runResetRefresh(
    admission: WorkbookCollaborationResetAdmission,
    reason: "resume_reset" | "sequence_gap",
  ): Promise<void> {
    try {
      await this.refreshActiveSurface(reason);
    } catch {
      if (!this.resetIsCurrent(admission)) return;
      this.dirtySurfaceKeys.add(admission.sheetKey);
      this.resetRetryTask = this.options.scheduler.schedule(1_000, () => {
        this.resetRetryTask = null;
        if (this.resetIsCurrent(admission)) {
          void this.runResetRefresh(admission, reason);
        }
      });
      return;
    }
    if (!this.resetIsCurrent(admission)) return;
    const completed = this.session?.completeReset(admission.eventGeneration);
    this.resetMachine = cancelWorkbookCollaborationReset(this.resetMachine);
    if (completed !== true) return;
    if (this.authorizationRecoveryMachine.canResumeMutations) {
      this.options.mutationRuntime.applyAuthorizationRecoveryState("resumed");
    }
    this.publishPresenceNow();
  }

  private handleRecordChanged(payload: RecordChangedPayload): void {
    if (
      this.options.mutationRuntime.resolveSocketClientTxn(
        payload.client_txn_id,
      ) ||
      Array.from(this.clientTxnResolvers).some((resolver) =>
        resolver(payload.client_txn_id),
      )
    ) {
      return;
    }
    const activeKey = sheetRefKey(this.activeSheetRef);
    const port = this.activePort;
    if (port === null || sheetRefKey(port.identity.sheetRef) !== activeKey) {
      this.dirtySurfaceKeys.add(activeKey);
      return;
    }
    const result = port.applyRecordChanged(payload);
    if (result.kind === "refresh_required") {
      this.queueActiveSurfaceRefresh("record_changed", payload.record_id);
    }
  }

  private handleEventPlan(plan: WorkbookCollaborationEventPlan): void {
    if (this.disposed) return;
    if (plan.kind === "record_changed") {
      if (this.authorizationRecoveryMachine.authorizationConfirmed) {
        this.handleRecordChanged(plan.payload);
      }
      return;
    }
    if (plan.kind === "presence_snapshot" || plan.kind === "presence_delta") {
      if (this.authorizationRecoveryMachine.authorizationConfirmed) {
        this.handlePresenceEventPlan(plan);
      }
      return;
    }
    if (plan.kind === "ignore") return;
    this.handleLifecycleEventPlan(plan);
  }

  private handlePresenceEventPlan(plan: PresenceEventPlan): void {
    const next =
      plan.kind === "presence_snapshot"
        ? replaceWorkbookPresenceSnapshot(this.presenceProjection, plan.payload)
        : applyWorkbookPresenceDelta(this.presenceProjection, plan.payload);
    if (next === this.presenceProjection) return;
    this.presenceProjection = next;
    this.emit();
  }

  private handleLifecycleEventPlan(plan: LifecycleEventPlan): void {
    switch (plan.kind) {
      case "established":
        this.emit();
        if (
          plan.mayResume &&
          this.authorizationRecoveryMachine.authorizationConfirmed
        ) {
          if (this.authorizationRecoveryMachine.canResumeMutations) {
            this.options.mutationRuntime.applyAuthorizationRecoveryState(
              "resumed",
            );
          }
          this.publishPresenceNow();
        }
        return;
      case "reset":
        this.beginReset(plan.eventGeneration, plan.reason);
        return;
      case "recover_authorization":
        this.requestAuthorizationRecovery();
        return;
      case "incident_closed":
        this.cancelReset();
        this.cancelAuthorizationWork();
        this.authorizationRecoveryMachine =
          terminateWorkbookAuthorizationRecovery(
            this.authorizationRecoveryMachine,
          );
        this.applyInvalidationPlan({ kind: "incident_closed" });
        return;
    }
  }
}

export function createWorkbookCollaborationCoordinator(
  options: WorkbookCollaborationCoordinatorOptions,
) {
  const runtime = new WorkbookCollaborationCoordinatorRuntime(options);
  return {
    attachSession: (session: WorkbookCollaborationSessionPort) =>
      runtime.attachSession(session),
    dispose: () => runtime.dispose(),
    editingPresenceForCell: (recordId: string | null, fieldKey: string) =>
      runtime.editingPresenceForCell(recordId, fieldKey),
    getSnapshot: runtime.getSnapshot,
    publishPresence: (presence: WorkbookPresenceDraft) =>
      runtime.publishPresence(presence),
    registerActiveSurface: (port: WorkbookActiveSurfacePort) =>
      runtime.registerActiveSurface(port),
    registerClientTxnResolver: (
      resolver: (clientTxnId: string | null | undefined) => boolean,
    ) => runtime.registerClientTxnResolver(resolver),
    requestAuthorizationRecovery: () => runtime.requestAuthorizationRecovery(),
    retain: () => runtime.retain(),
    setActiveSheet: (sheetRef: SheetRef) => runtime.setActiveSheet(sheetRef),
    subscribe: runtime.subscribe,
  };
}

export type WorkbookCollaborationCoordinator = ReturnType<
  typeof createWorkbookCollaborationCoordinator
>;
