import type {
  IncidentCollaborationEvent,
  IncidentCollaborationStatus,
} from "../../collaboration/IncidentCollaborationSession";
import { apiPath } from "../../services/browserApi";
import { fetchWorkbookJSON } from "../../services/workbookApi";
import {
  type WorkbookSheetRef,
  workbookSheetRefKey,
} from "../../shared/workbookSheetRef";
import {
  isPresenceRecord,
  type PresenceRecord,
  presenceMatchesSheet,
} from "../utils/workbookPresence";
import type { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";
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
export class WorkbookCollaborationProjection {
  private activePort: WorkbookActiveSurfacePort | null = null;
  private activeSheetRef: WorkbookSheetRef;
  private authProbeTimer: ReturnType<typeof setTimeout> | null = null;
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
  private session: WorkbookCollaborationSessionPort | null = null;
  private sessionUnsubscribe: (() => void) | null = null;
  private snapshot: WorkbookCollaborationSnapshot;

  constructor(
    private readonly options: {
      readonly apiBase: string | undefined;
      readonly initialSheetRef: WorkbookSheetRef;
      readonly mutationRuntime: WorkbookMutationRuntime;
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

  attachSession(session: WorkbookCollaborationSessionPort): () => void {
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

  reconnect(): void {
    this.session?.reconnect();
  }

  publishPresence(presence: WorkbookPresenceDraft): void {
    this.presenceDraft = { ...presence };
    if (this.presenceTimer !== null) clearTimeout(this.presenceTimer);
    this.presenceTimer = setTimeout(() => {
      this.presenceTimer = null;
      this.publishPresenceNow();
    }, 150);
  }

  presenceForRow(recordId: string | null): readonly PresenceRecord[] {
    if (recordId === null) return [];
    return this.snapshot.activeSheetPresenceRecords.filter(
      (presence) => presence.record_id === recordId,
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
    this.sessionUnsubscribe?.();
    this.sessionUnsubscribe = null;
    this.session = null;
    if (this.presenceTimer !== null) clearTimeout(this.presenceTimer);
    if (this.authProbeTimer !== null) clearTimeout(this.authProbeTimer);
    this.presenceTimer = null;
    this.authProbeTimer = null;
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

  private scheduleAuthRecoveryProbe(): void {
    if (this.authProbeTimer !== null) return;
    this.authProbeTimer = setTimeout(async () => {
      this.authProbeTimer = null;
      if (!this.options.mutationRuntime.getSnapshot().authPaused) return;
      try {
        const result = await fetchWorkbookJSON(
          apiPath(this.options.apiBase, "/api/v1/auth/session"),
        );
        if (!result.ok) {
          this.scheduleAuthRecoveryProbe();
          return;
        }
        this.options.mutationRuntime.resumeAfterAuthRecovery();
        this.session?.reconnect();
      } catch {
        this.scheduleAuthRecoveryProbe();
      }
    }, 1000);
  }

  private handleEvent(event: IncidentCollaborationEvent): void {
    if (event.kind === "established") {
      this.emit();
      if (event.payload.status !== "reset_required") {
        this.options.mutationRuntime.resumeAfterAuthRecovery();
        this.publishPresenceNow();
      }
      return;
    }
    if (event.kind === "reset_required") {
      this.clearPresence();
      void this.refreshActiveSurface(event.reason).then(() => {
        if (this.session?.completeReset(event.generation)) {
          this.options.mutationRuntime.resumeAfterAuthRecovery();
          this.publishPresenceNow();
        }
      });
      return;
    }
    if (
      event.kind === "authorization_lost" ||
      event.kind === "session_revoked"
    ) {
      this.clearPresence();
      this.options.mutationRuntime.pauseForAuthRecovery();
      this.activePort?.clearAuthorizedRows();
      this.scheduleAuthRecoveryProbe();
      return;
    }
    if (event.kind === "incident_closed") {
      this.clearPresence();
      this.options.mutationRuntime.pauseForTerminalLifecycle();
      this.activePort?.clearAuthorizedRows();
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
