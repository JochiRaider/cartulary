import type {
  IncidentCollaborationEvent,
  IncidentCollaborationStatus,
} from "../../collaboration/IncidentCollaborationSession";
import type {
  AuthorizationRecoveryPort,
  AuthorizationRecoveryResult,
} from "../../shared/authorizationRecovery";
import {
  type WorkbookSheetRef,
  workbookSheetRefKey,
} from "../../shared/workbookSheetRef";
import type {
  WorkbookDependentPresentationInvalidationReason,
  WorkbookInvalidationReason,
  WorkbookQueryInvalidationReason,
} from "../lifecycle/workbookInvalidation";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import {
  isPresenceRecord,
  type PresenceRecord,
  presenceMatchesSheet,
} from "../utils/workbookPresence";
import {
  buildWorkbookPresenceInput,
  isRecordChangedMessage,
  type WorkbookPresenceDraft,
} from "./workbookCollaborationMessages";
import type { WorkbookActiveSurfacePort } from "./workbookSurfacePort";

type CollaborationProjectionListener = () => void;

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

function recordValue(value: unknown): Record<string, unknown> {
  if (value !== null && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

/**
 * Shell-lifetime interpretation and reconciliation authority for collaboration.
 * The incident session remains the sole transport owner.
 */
export class WorkbookCollaborationCoordinator {
  private activePort: WorkbookActiveSurfacePort | null = null;
  private activeSheetRef: WorkbookSheetRef;
  private authorizationRecoveryController: AbortController | null = null;
  private authorizationRecoveryTimer: ReturnType<typeof setTimeout> | null =
    null;
  private canResumeMutations = true;
  private disposed = false;
  private readonly dirtySurfaceKeys = new Set<string>();
  private readonly listeners = new Set<CollaborationProjectionListener>();
  private readonly clientTxnResolvers = new Set<
    (clientTxnId: string | null | undefined) => boolean
  >();
  private presenceByConnectionId = new Map<string, PresenceRecord>();
  private presenceDraft: WorkbookPresenceDraft = {
    fieldKey: null,
    mode: "viewing",
    recordId: null,
  };
  private presenceTimer: ReturnType<typeof setTimeout> | null = null;
  private retainCount = 0;
  private retainGeneration = 0;
  private session: WorkbookCollaborationSessionPort | null = null;
  private sessionUnsubscribe: (() => void) | null = null;
  private snapshot: WorkbookCollaborationSnapshot;

  constructor(
    private readonly options: {
      readonly authorizationRecovery: AuthorizationRecoveryPort;
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
      readonly initialSheetRef: WorkbookSheetRef;
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
      readonly queryInvalidation: (
        reason: WorkbookQueryInvalidationReason,
      ) => void;
    },
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
    this.sessionUnsubscribe?.();
    this.session = session;
    this.sessionUnsubscribe = session.subscribe((event) => {
      this.handleEvent(event);
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

  setActiveSheet(sheetRef: WorkbookSheetRef): void {
    if (
      workbookSheetRefKey(this.activeSheetRef) === workbookSheetRefKey(sheetRef)
    ) {
      return;
    }
    this.activeSheetRef = { ...sheetRef };
    this.presenceDraft = {
      fieldKey: null,
      mode: "viewing",
      recordId: null,
    };
    this.publishPresenceNow();
    this.emit();
  }

  registerActiveSurface(port: WorkbookActiveSurfacePort): () => void {
    this.activePort = port;
    const key = workbookSheetRefKey(port.identity.sheetRef);
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
    if (this.disposed) return;
    this.invalidateProtectedState({ kind: "session_unavailable" });
    this.scheduleAuthorizationRecovery();
  }

  publishPresence(presence: WorkbookPresenceDraft): void {
    this.presenceDraft = { ...presence };
    if (this.presenceTimer !== null) clearTimeout(this.presenceTimer);
    this.presenceTimer = setTimeout(() => {
      this.presenceTimer = null;
      this.publishPresenceNow();
    }, 150);
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
    if (this.presenceTimer !== null) clearTimeout(this.presenceTimer);
    if (this.authorizationRecoveryTimer !== null) {
      clearTimeout(this.authorizationRecoveryTimer);
    }
    this.authorizationRecoveryController?.abort();
    this.presenceTimer = null;
    this.authorizationRecoveryTimer = null;
    this.authorizationRecoveryController = null;
    this.presenceByConnectionId.clear();
    this.listeners.clear();
    this.clientTxnResolvers.clear();
    const reason = { kind: "runtime_disposed" } as const;
    this.options.mutationRuntime.invalidate(reason);
    this.options.queryInvalidation(reason);
    this.activePort?.invalidate(reason);
    this.options.inspectorInvalidation(reason);
    this.options.continuityInvalidation(reason);
    this.options.evidenceInvalidation(reason);
  }

  private emit(): void {
    const connectionId = this.session?.connectionId ?? null;
    const activeSheetPresenceRecords = Array.from(
      this.presenceByConnectionId.values(),
    )
      .filter((presence) => presenceMatchesSheet(presence, this.activeSheetRef))
      .filter((presence) => presence.connection_id !== connectionId)
      .sort((left, right) => {
        const byName = left.display_name.localeCompare(right.display_name);
        return byName === 0
          ? left.connection_id.localeCompare(right.connection_id)
          : byName;
      });
    this.snapshot = {
      activeSheetPresenceRecords,
      connectionId,
      status: this.session?.status ?? "disconnected",
    };
    for (const listener of this.listeners) listener();
  }

  private publishPresenceNow(): void {
    this.session?.publishPresence(
      buildWorkbookPresenceInput(this.presenceDraft, this.activeSheetRef),
    );
  }

  private applyPresenceSnapshot(payload: Record<string, unknown>): void {
    const values = Array.isArray(payload.presences) ? payload.presences : [];
    this.presenceByConnectionId = new Map(
      values
        .filter(isPresenceRecord)
        .map((presence) => [
          presence.connection_id,
          { ...presence, sheet_ref: { ...presence.sheet_ref } },
        ]),
    );
    this.emit();
  }

  private applyPresenceDelta(payload: Record<string, unknown>): void {
    const presence = recordValue(payload.presence);
    const connectionId = presence.connection_id;
    if (typeof connectionId !== "string") return;
    if (payload.delta_kind === "remove") {
      this.presenceByConnectionId.delete(connectionId);
      this.emit();
      return;
    }
    if (payload.delta_kind !== "upsert" || !isPresenceRecord(presence)) return;
    this.presenceByConnectionId.set(connectionId, {
      ...presence,
      sheet_ref: { ...presence.sheet_ref },
    });
    this.emit();
  }

  private clearPresence(): void {
    if (this.presenceByConnectionId.size === 0) return;
    this.presenceByConnectionId.clear();
    this.emit();
  }

  private refreshActiveSurface(
    reason: string,
    recordId?: string,
  ): Promise<void> {
    const port = this.activePort;
    if (
      port === null ||
      workbookSheetRefKey(port.identity.sheetRef) !==
        workbookSheetRefKey(this.activeSheetRef)
    ) {
      this.dirtySurfaceKeys.add(workbookSheetRefKey(this.activeSheetRef));
      return Promise.resolve();
    }
    return port.refresh({ reason, ...(recordId ? { recordId } : {}) });
  }

  private scheduleAuthorizationRecovery(): void {
    if (this.disposed || this.authorizationRecoveryTimer !== null) return;
    this.authorizationRecoveryTimer = setTimeout(() => {
      this.authorizationRecoveryTimer = null;
      void this.recoverAuthorization();
    }, 1000);
  }

  private async recoverAuthorization(): Promise<void> {
    if (
      this.disposed ||
      !this.options.mutationRuntime.getSnapshot().authPaused
    ) {
      return;
    }
    this.authorizationRecoveryController?.abort();
    const controller = new AbortController();
    this.authorizationRecoveryController = controller;
    let result: AuthorizationRecoveryResult;
    try {
      result = await this.options.authorizationRecovery.recover({
        incidentId: this.options.incidentId,
        signal: controller.signal,
      });
    } catch {
      if (!controller.signal.aborted) this.scheduleAuthorizationRecovery();
      return;
    }
    if (
      this.disposed ||
      controller.signal.aborted ||
      this.authorizationRecoveryController !== controller
    ) {
      return;
    }
    this.authorizationRecoveryController = null;
    if (result.kind === "unavailable") {
      this.scheduleAuthorizationRecovery();
      return;
    }
    if (result.kind === "access_lost") {
      this.invalidateProtectedState({ kind: "incident_access_lost" });
      this.options.onIncidentAccessLost?.();
      return;
    }
    this.options.onAuthorizationRecovered(result);
    this.canResumeMutations =
      result.role === "editor" ||
      result.role === "reviewer" ||
      result.role === "admin";
    if (!this.canResumeMutations) {
      const reason = {
        kind: "incident_role_changed",
        role: result.role,
      } as const;
      this.options.mutationRuntime.invalidate(reason);
      this.options.inspectorInvalidation(reason);
    }
    await this.refreshActiveSurface("authorization_recovered");
    if (this.disposed) return;
    if (this.canResumeMutations) {
      this.options.mutationRuntime.resumeAfterAuthRecovery();
    }
    this.session?.reconnect();
  }

  private invalidateProtectedState(
    reason:
      | Extract<
          WorkbookInvalidationReason,
          { readonly kind: "session_unavailable" }
        >
      | Extract<
          WorkbookInvalidationReason,
          { readonly kind: "incident_access_lost" }
        >,
  ): void {
    this.canResumeMutations = false;
    this.options.mutationRuntime.invalidate(reason);
    if (this.authorizationRecoveryTimer !== null) {
      clearTimeout(this.authorizationRecoveryTimer);
      this.authorizationRecoveryTimer = null;
    }
    this.authorizationRecoveryController?.abort();
    this.authorizationRecoveryController = null;
    this.options.queryInvalidation(reason);
    this.activePort?.invalidate(reason);
    this.options.extensionInvalidation(reason);
    this.clearPresence();
    this.options.inspectorInvalidation(reason);
    this.options.continuityInvalidation(reason);
    this.options.evidenceInvalidation(reason);
  }

  private handleEvent(event: IncidentCollaborationEvent): void {
    if (event.kind === "established") {
      this.emit();
      if (event.payload.status !== "reset_required") {
        if (this.canResumeMutations) {
          this.options.mutationRuntime.resumeAfterAuthRecovery();
        }
        this.publishPresenceNow();
      }
      return;
    }
    if (event.kind === "reset_required") {
      this.clearPresence();
      const reason = { kind: "collaboration_reset_required" } as const;
      this.options.queryInvalidation(reason);
      this.activePort?.invalidate(reason);
      void this.refreshActiveSurface(event.reason).then(() => {
        if (this.session?.completeReset(event.generation)) {
          if (this.canResumeMutations) {
            this.options.mutationRuntime.resumeAfterAuthRecovery();
          }
          this.publishPresenceNow();
        }
      });
      return;
    }
    if (
      event.kind === "authorization_lost" ||
      event.kind === "session_revoked"
    ) {
      this.requestAuthorizationRecovery();
      return;
    }
    if (event.kind === "incident_closed") {
      const reason = { kind: "incident_closed" } as const;
      this.options.mutationRuntime.invalidate(reason);
      this.options.queryInvalidation(reason);
      this.activePort?.invalidate(reason);
      this.clearPresence();
      this.options.inspectorInvalidation(reason);
      this.options.continuityInvalidation(reason);
      this.options.evidenceInvalidation(reason);
      return;
    }
    const message = event.message;
    if (message.type === "presence_snapshot") {
      this.applyPresenceSnapshot(recordValue(message.payload));
      return;
    }
    if (message.type === "presence_delta") {
      this.applyPresenceDelta(recordValue(message.payload));
      return;
    }
    if (!isRecordChangedMessage(message)) return;
    const payload = message.payload;
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
    const port = this.activePort;
    if (port === null) {
      this.dirtySurfaceKeys.add(workbookSheetRefKey(this.activeSheetRef));
      return;
    }
    const result = port.applyRecordChanged(payload);
    if (result.kind === "refresh_required") {
      void this.refreshActiveSurface("record_changed", payload.record_id);
    }
  }
}
