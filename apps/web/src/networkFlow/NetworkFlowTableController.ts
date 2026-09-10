import {
  boundedRead,
  browserObservationClock,
  type ObservationClock,
} from "../services/asyncObservation";
import { importContractFailure } from "../services/importClient";
import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";
import type { NetworkFlowExtensionResourceChange } from "./networkFlowCollaborationInterpreter";
import {
  initialNetworkFlowControllerState,
  type NetworkFlowControllerState,
  networkFlowControllerReducer,
} from "./networkFlowController";
import {
  NetworkFlowRequestError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";
import {
  type NetworkFlowImportHandoffRequest,
  type NetworkFlowImportHandoffResult,
  networkFlowImportTableMatches,
} from "./networkFlowImportState";
import {
  canMutateTable,
  canReadTables,
  captureTableAttempt,
  normalizeTableDisplayName,
  type TableAction,
  type TableAttempt,
  type TableAuthority,
  TableWriteError,
  tableAuthorityKey,
  tableWriteFailure,
} from "./networkFlowTableOperation";

export type TableTransport = {
  readonly list: (signal: AbortSignal) => Promise<NetworkFlowTable[]>;
  readonly submit: (
    attempt: TableAttempt,
    signal: AbortSignal,
    authorizeDispatch: () => void,
  ) => Promise<NetworkFlowTable>;
};
export type TableDraft = {
  readonly id: number;
  readonly action: TableAction;
  readonly target: NetworkFlowTable;
  readonly name: string;
  readonly confirmation: string;
  readonly reviewRequired: boolean;
  readonly error: string | null;
};
export type TableOperation = {
  readonly attempt: TableAttempt;
  readonly dialogId: number;
  readonly status: "pending" | "uncertain" | "rejected" | "acknowledged";
  readonly receipt: NetworkFlowTable | null;
  readonly failure: TableWriteError | null;
};
export type TableSnapshot = NetworkFlowControllerState & {
  readonly loadState: "loading" | "refreshing" | "ready" | "error";
  readonly error: NetworkFlowRequestError | null;
  readonly draft: TableDraft | null;
  readonly operation: TableOperation | null;
  readonly presentation: "draft" | "operation" | null;
  readonly hidden: boolean;
  readonly canRename: boolean;
  readonly canDelete: boolean;
  readonly notice: string | null;
  readonly observationRevision: number;
};
const initial = (): TableSnapshot => ({
  ...initialNetworkFlowControllerState,
  loadState: "loading",
  error: null,
  draft: null,
  operation: null,
  presentation: null,
  hidden: true,
  canRename: false,
  canDelete: false,
  notice: null,
  observationRevision: 0,
});

/** Sole owner of table observations, selection, captured intent and volatile recovery. */
export class NetworkFlowTableController {
  private state = initial();
  private snapshot = this.state;
  private listeners = new Set<() => void>();
  private changes = new Set<
    (change: NetworkFlowExtensionResourceChange) => void
  >();
  private authorityReader: (() => TableAuthority) | null = null;
  private authority: TableAuthority | null = null;
  private retainedActor: string | null = null;
  private retainedIncident: string | null = null;
  private blockedAuthority: string | null = null;
  private blockedWrites: string | null = null;
  private transport: TableTransport | null = null;
  private active = false;
  private closed = false;
  private listRequest: AbortController | null = null;
  private listGeneration = 0;
  private write: {
    readonly stop: AbortController;
    readonly attempt: TableAttempt;
    readonly authorityKey: string;
    dispatched: boolean;
    readonly replay: boolean;
  } | null = null;
  private nextDialogId = 0;
  private selectionRevision = 0;
  private onAdmitted: (() => void) | null = null;
  private localChange:
    | ((change: NetworkFlowExtensionResourceChange) => void)
    | null = null;
  constructor(
    private readonly clock: ObservationClock = browserObservationClock,
  ) {}
  readonly getSnapshot = (): TableSnapshot => this.snapshot;
  /** Query drafts survive metadata refreshes, but never a replaced reader. */
  readonly queryContextIdentity = (): string | null => {
    const authority = this.authority;
    return authority && canReadTables(authority) && !this.snapshot.hidden
      ? JSON.stringify([
          authority.incidentId,
          authority.actorId,
          authority.sessionIdentity,
          authority.availabilityTag?.epochId,
        ])
      : null;
  };
  readonly subscribe = (listener: () => void): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };
  readonly subscribeChanges = (
    listener: (change: NetworkFlowExtensionResourceChange) => void,
  ): (() => void) => {
    this.changes.add(listener);
    return () => this.changes.delete(listener);
  };
  private update(patch: Partial<TableSnapshot>): void {
    this.state = { ...this.state, ...patch };
    this.snapshot = this.state.hidden
      ? {
          ...this.state,
          tables: [],
          activeTableId: null,
          removedTableIds: [],
          draft: null,
          operation: null,
          presentation: null,
          notice: null,
        }
      : this.state;
    for (const listener of this.listeners) listener();
  }
  bind(
    transport: TableTransport,
    authority: () => TableAuthority,
    localChange?: (change: NetworkFlowExtensionResourceChange) => void,
    onAdmitted?: () => void,
  ): void {
    this.transport = transport;
    this.authorityReader = authority;
    this.localChange = localChange ?? null;
    this.onAdmitted = onAdmitted ?? null;
    this.revalidateAuthority();
  }
  readonly revalidateAuthority = (): void => {
    const next = this.authorityReader?.();
    if (!next) return;
    const previous = this.authority;
    const key = tableAuthorityKey(next);
    if (previous && tableAuthorityKey(previous) === key) return;
    this.authority = next;
    this.closed = !next.open;
    this.invalidateLists();
    this.stopObservation();
    if (
      (this.retainedIncident !== null &&
        this.retainedIncident !== next.incidentId) ||
      (next.actorId !== null &&
        this.retainedActor !== null &&
        this.retainedActor !== next.actorId)
    )
      this.purge();
    this.retainedIncident = next.incidentId;
    if (next.actorId !== null) this.retainedActor = next.actorId;
    if (this.blockedAuthority !== key) this.blockedAuthority = null;
    if (this.blockedWrites !== key) this.blockedWrites = null;
    if (next.sessionIdentity === null || next.actorId === null) {
      this.update({ hidden: true, canRename: false, canDelete: false });
      return;
    }
    if (
      !next.profileAvailable ||
      !["viewer", "editor", "reviewer", "admin"].includes(next.role ?? "")
    ) {
      this.purge();
      return;
    }
    if (!next.available) {
      this.update({ hidden: true, canRename: false, canDelete: false });
      return;
    }
    const read = canReadTables(next);
    if (!read)
      this.update({ ...initialNetworkFlowControllerState, loadState: "ready" });
    const draft = this.state.draft;
    this.update({
      hidden: false,
      canRename: canMutateTable(next, "rename"),
      canDelete: canMutateTable(next, "delete"),
      draft: draft
        ? {
            ...draft,
            reviewRequired: true,
            confirmation: "",
            error:
              "Access changed. Review the current table before submitting. Your draft remains available to copy.",
          }
        : null,
    });
    if (this.active && read) void this.loadTables();
  };
  readonly setActive = (active: boolean): void => {
    if (this.active === active) return;
    this.active = active;
    if (active) {
      this.revalidateAuthority();
      void this.loadTables();
    } else {
      this.invalidateLists();
      this.closeDialog();
    }
  };
  private readable(): boolean {
    this.revalidateAuthority();
    return (
      this.authority !== null &&
      canReadTables(this.authority) &&
      this.blockedAuthority !== tableAuthorityKey(this.authority)
    );
  }
  private writable(action: TableAction): boolean {
    return (
      this.readable() &&
      this.authority !== null &&
      canMutateTable(this.authority, action) &&
      this.blockedWrites !== tableAuthorityKey(this.authority)
    );
  }
  private invalidateLists(): void {
    this.listRequest?.abort();
    this.listRequest = null;
    this.listGeneration++;
  }
  private stopObservation(): void {
    const run = this.write;
    if (!run) return;
    this.write = null;
    run.stop.abort();
    const operation = this.state.operation;
    if (operation?.attempt === run.attempt && operation.status === "pending")
      this.update({
        operation: {
          ...operation,
          status: run.dispatched || run.replay ? "uncertain" : "rejected",
          failure:
            run.dispatched || run.replay
              ? tableWriteFailure(null)
              : new TableWriteError(
                  "rejected",
                  "This request was not submitted. Review the current table before trying again.",
                ),
        },
      });
  }
  private purge(): void {
    this.invalidateLists();
    this.stopObservation();
    this.selectionRevision++;
    this.state = initial();
    this.update({ loadState: "ready" });
  }
  readonly clearAuthorization = (): void => {
    if (
      this.state.hidden &&
      this.blockedAuthority ===
        (this.authority ? tableAuthorityKey(this.authority) : null)
    )
      return;
    this.blockedAuthority = this.authority
      ? tableAuthorityKey(this.authority)
      : null;
    this.purge();
  };
  readonly dispose = (): void => {
    this.active = false;
    this.purge();
    this.authority = null;
  };
  readonly loadTables = async (): Promise<boolean> => {
    if (!this.readable() || !this.transport) return false;
    this.invalidateLists();
    const stop = new AbortController();
    this.listRequest = stop;
    const generation = this.listGeneration;
    const key = tableAuthorityKey(this.authority as TableAuthority);
    this.update({
      loadState: this.state.tables.length ? "refreshing" : "loading",
      error: null,
    });
    try {
      const tables = await boundedRead(
        this.transport.list,
        stop.signal,
        30_000,
        this.clock,
      );
      if (
        stop.signal.aborted ||
        generation !== this.listGeneration ||
        !this.readable() ||
        key !== tableAuthorityKey(this.authority as TableAuthority)
      )
        return false;
      this.acceptTables(tables);
      this.update({ loadState: "ready", error: null });
      return true;
    } catch (caught) {
      if (stop.signal.aborted || generation !== this.listGeneration)
        return false;
      const error = networkFlowErrorFromUnknown(
        caught,
        "Table metadata could not be refreshed. Previously authorized data remains visible.",
      );
      this.handleFailure(error);
      this.update({ loadState: "error", error });
      return false;
    }
  };
  private acceptTables(tables: readonly NetworkFlowTable[]): void {
    const removed = this.state.tables.filter(
      (table) =>
        !tables.some(
          (next) => next.network_flow_table_id === table.network_flow_table_id,
        ),
    );
    for (const table of removed) {
      this.removeTable(
        table.network_flow_table_id,
        this.state.operation?.status === "acknowledged" &&
          this.state.operation.attempt.action === "delete",
      );
      const change: NetworkFlowExtensionResourceChange = {
        resourceKind: "network_flow_table",
        resourceId: table.network_flow_table_id,
        changeKind: "remove",
        reasonCode: "soft_deleted",
      };
      this.localChange?.(change);
      this.emit(change);
    }
    const catalog = networkFlowControllerReducer(this.state, {
      type: "replace_tables",
      tables,
    });
    const draft = this.state.draft;
    const current = draft
      ? catalog.tables.find(
          (table) =>
            table.network_flow_table_id === draft.target.network_flow_table_id,
        )
      : null;
    this.update({
      observationRevision: this.state.observationRevision + 1,
      ...catalog,
      draft:
        draft &&
        current &&
        (draft.target.table_version !== current.table_version ||
          draft.target.display_name !== current.display_name)
          ? {
              ...draft,
              reviewRequired: true,
              confirmation: "",
              error:
                "The table changed. Review the current name and version before submitting.",
            }
          : draft,
    });
  }

  readonly selectTable = (tableId: string): void => {
    if (!this.readable()) return;
    this.selectionRevision++;
    this.update({
      ...networkFlowControllerReducer(this.state, {
        type: "select_table",
        tableId,
      }),
      presentation: null,
    });
  };
  readonly handoffImportedTable = async (
    request: NetworkFlowImportHandoffRequest,
  ): Promise<NetworkFlowImportHandoffResult> => {
    if (
      !this.readable() ||
      request.incidentId !== this.authority?.incidentId ||
      !request.current() ||
      !this.transport
    )
      return { kind: "superseded" };
    this.invalidateLists();
    const stop = new AbortController();
    this.listRequest = stop;
    const generation = this.listGeneration,
      selection = this.selectionRevision;
    const signal = AbortSignal.any([stop.signal, request.signal]);
    const current = () =>
      !signal.aborted &&
      generation === this.listGeneration &&
      selection === this.selectionRevision &&
      request.current() &&
      this.readable();
    try {
      const tables = await boundedRead(
        this.transport.list,
        signal,
        30_000,
        this.clock,
      );
      if (!current()) return { kind: "superseded" };
      const table = tables.find(
        (table) => table.network_flow_table_id === request.tableId,
      );
      if (!table) return { kind: "unavailable" };
      if (!networkFlowImportTableMatches(table, request))
        return { kind: "failed", failure: importContractFailure() };
      this.acceptTables(tables);
      this.update({
        ...networkFlowControllerReducer(this.state, {
          type: "select_table",
          tableId: request.tableId,
        }),
        loadState: "ready",
        error: null,
      });
      return this.state.activeTableId === request.tableId
        ? { kind: "selected", table }
        : { kind: "unavailable" };
    } catch (caught) {
      if (!current()) return { kind: "superseded" };
      const error = networkFlowErrorFromUnknown(
        caught,
        "The imported table could not be loaded.",
      );
      return {
        kind: "failed",
        failure: {
          kind: error.status ? "public" : "transport",
          status: error.status,
          code: error.code,
          reason: error.reasonCode,
          field: null,
          retryable: error.retryable,
        },
      };
    }
  };
  readonly openAction = (action: TableAction, tableId: string): void => {
    if (!this.writable(action)) return;
    const target = this.state.tables.find(
      (table) => table.network_flow_table_id === tableId,
    );
    if (!target) return;
    this.update({
      draft: {
        id: ++this.nextDialogId,
        action,
        target: Object.freeze({ ...target }),
        name: target.display_name,
        confirmation: "",
        reviewRequired: false,
        error: null,
      },
      presentation: "draft",
      notice: null,
    });
  };
  readonly setName = (name: string): void => {
    if (this.state.draft)
      this.update({ draft: { ...this.state.draft, name, error: null } });
  };
  readonly setConfirmation = (confirmation: string): void => {
    if (this.state.draft)
      this.update({
        draft: { ...this.state.draft, confirmation, error: null },
      });
  };
  readonly closeDialog = (): void => {
    this.stopObservation();
    this.update({ presentation: null });
  };
  readonly reopenDraft = (): void => {
    if (!this.state.hidden && this.state.draft)
      this.update({ presentation: "draft" });
  };
  readonly reopenOperation = (): void => {
    if (!this.state.hidden && this.state.operation)
      this.update({ presentation: "operation" });
  };
  readonly reviewCurrent = (): void => {
    const draft = this.state.draft;
    if (
      !draft ||
      !this.writable(draft.action) ||
      this.state.loadState !== "ready"
    )
      return;
    const target = this.state.tables.find(
      (table) =>
        table.network_flow_table_id === draft.target.network_flow_table_id,
    );
    if (!target) return;
    this.update({
      draft: {
        ...draft,
        id: ++this.nextDialogId,
        target: Object.freeze({ ...target }),
        reviewRequired: false,
        confirmation: "",
        error: null,
      },
      presentation: "draft",
    });
  };
  readonly submit = async (): Promise<void> => {
    const draft = this.state.draft;
    if (!draft || !this.writable(draft.action) || this.unresolved()) return;
    const current = this.state.tables.find(
      (table) =>
        table.network_flow_table_id === draft.target.network_flow_table_id,
    );
    if (
      !current ||
      draft.reviewRequired ||
      current.table_version !== draft.target.table_version ||
      current.display_name !== draft.target.display_name
    ) {
      this.update({
        draft: {
          ...draft,
          reviewRequired: true,
          error: "Review the current table before submitting.",
        },
      });
      return;
    }
    if (
      draft.action === "delete" &&
      draft.confirmation !== draft.target.display_name
    ) {
      this.update({
        draft: {
          ...draft,
          error: "Type the captured table name exactly to confirm deletion.",
        },
      });
      return;
    }
    if (draft.action === "rename") {
      const name = normalizeTableDisplayName(draft.name);
      const error = !name.ok
        ? name.message
        : this.state.tables.some(
              (table) =>
                table.network_flow_table_id !== current.network_flow_table_id &&
                table.display_name === name.name,
            )
          ? "Another active table has this name. Choose a unique name."
          : null;
      if (error) {
        this.update({ draft: { ...draft, error } });
        return;
      }
    }
    try {
      const attempt = captureTableAttempt(
        draft.action,
        this.authority as TableAuthority,
        draft.target,
        draft.name,
      );
      this.update({
        operation: {
          attempt,
          dialogId: draft.id,
          status: "pending",
          receipt: null,
          failure: null,
        },
      });
      this.onAdmitted?.();
      await this.execute(attempt, false);
    } catch (caught) {
      this.update({
        draft: {
          ...draft,
          error:
            caught instanceof Error
              ? caught.message
              : "This request was not submitted.",
        },
      });
    }
  };
  private unresolved(): boolean {
    return (
      this.state.operation?.status === "pending" ||
      this.state.operation?.status === "uncertain"
    );
  }
  readonly replay = async (): Promise<void> => {
    const operation = this.state.operation;
    if (
      !operation ||
      operation.status !== "uncertain" ||
      !this.writable(operation.attempt.action) ||
      this.write
    )
      return;
    if (
      operation.attempt.authority.actorId !== this.authority?.actorId ||
      operation.attempt.authority.incidentId !== this.authority.incidentId
    )
      return;
    this.update({
      operation: { ...operation, status: "pending", failure: null },
    });
    this.onAdmitted?.();
    await this.execute(operation.attempt, true);
  };
  private async execute(attempt: TableAttempt, replay: boolean): Promise<void> {
    if (!this.transport || !this.authority) return;
    const run = {
      stop: new AbortController(),
      attempt,
      authorityKey: tableAuthorityKey(this.authority),
      dispatched: false,
      replay,
    };
    this.write = run;
    const current = () =>
      this.write === run &&
      this.state.operation?.attempt === attempt &&
      this.authorityReader !== null &&
      tableAuthorityKey(this.authorityReader()) === run.authorityKey;
    try {
      const receipt = await boundedRead(
        (signal) =>
          (this.transport as TableTransport).submit(attempt, signal, () => {
            if (
              !current() ||
              !this.writable(attempt.action) ||
              run.stop.signal.aborted
            )
              throw new Error(
                "Table mutation authority changed before dispatch.",
              );
            if (!replay) {
              const target = this.state.tables.find(
                (table) =>
                  table.network_flow_table_id ===
                  attempt.target.network_flow_table_id,
              );
              if (
                !target ||
                target.table_version !== attempt.target.table_version ||
                this.state.draft?.reviewRequired
              )
                throw new Error(
                  "Table changed before dispatch. Review the current table.",
                );
            }
            run.dispatched = true;
          }),
        run.stop.signal,
        30_000,
        this.clock,
      );
      if (!current()) return;
      this.write = null;
      this.invalidateLists();
      const operation = this.state.operation as TableOperation;
      const sameDialog =
        this.state.presentation === "operation" ||
        (this.state.presentation === "draft" &&
          this.state.draft?.id === operation.dialogId);
      this.update({
        draft:
          this.state.draft?.id === operation.dialogId ? null : this.state.draft,
        operation: {
          ...operation,
          status: "acknowledged",
          receipt,
          failure: null,
        },
        presentation: sameDialog ? null : this.state.presentation,
        notice:
          attempt.action === "delete"
            ? `${receipt.display_name} was deleted.`
            : receipt.table_version === attempt.target.table_version
              ? "The table name is unchanged."
              : "The table was renamed.",
      });
      if (attempt.action === "delete") {
        const change: NetworkFlowExtensionResourceChange = {
          resourceKind: "network_flow_table",
          resourceId: receipt.network_flow_table_id,
          changeKind: "remove",
          reasonCode: "soft_deleted",
        };
        this.removeTable(receipt.network_flow_table_id, true);
        this.localChange?.(change);
        this.emit(change);
      } else if (!replay)
        this.update(
          networkFlowControllerReducer(this.state, {
            type: "replace_table",
            table: receipt,
          }),
        );
      // A receipt remains acknowledged independently of current metadata observation.
      await this.loadTables();
    } catch (caught) {
      if (!current()) return;
      this.write = null;
      const observedFailure = tableWriteFailure(caught);
      const failure =
        !run.dispatched && !replay
          ? new TableWriteError(
              "rejected",
              "This request was not submitted. Review the current table before trying again.",
              observedFailure.detail,
            )
          : observedFailure;
      const operation = this.state.operation as TableOperation;
      this.update({
        operation: {
          ...operation,
          status:
            replay || (run.dispatched && failure.certainty === "uncertain")
              ? "uncertain"
              : "rejected",
          failure,
        },
        draft:
          this.state.draft?.id === operation.dialogId
            ? {
                ...this.state.draft,
                error: failure.message,
                reviewRequired:
                  failure.detail?.code ===
                    "network_flow_table_version_conflict" || !run.dispatched,
              }
            : this.state.draft,
      });
      if (failure.detail)
        this.handleFailure(
          failure.detail,
          attempt.target.network_flow_table_id,
        );
      if (
        failure.detail?.code === "network_flow_table_version_conflict" ||
        failure.detail?.code === "authorization_denied"
      )
        await this.loadTables();
    }
  }
  readonly onProtectedFailure = (error: NetworkFlowRequestError): void => {
    this.handleFailure(error);
  };
  private revokeRead(): void {
    this.clearAuthorization();
    this.localChange?.({
      resourceKind: "*",
      resourceId: "*",
      changeKind: "remove",
      reasonCode: "authorization_lost",
    });
  }
  private handleFailure(
    error: NetworkFlowRequestError,
    tableId?: string,
  ): void {
    if (error.status === 401 || error.code === "session_required") {
      this.blockedAuthority = this.authority
        ? tableAuthorityKey(this.authority)
        : null;
      this.stopObservation();
      this.update({ hidden: true, canRename: false, canDelete: false });
      this.localChange?.({
        resourceKind: "*",
        resourceId: "*",
        changeKind: "remove",
        reasonCode: "session_revoked",
      });
    } else if (
      error.code === "incident_not_found" ||
      error.code === "extension_profile_not_claimed"
    )
      this.revokeRead();
    else if (error.code === "incident_closed") {
      if (this.closed) return;
      this.closed = true;
      this.blockedAuthority = this.authority
        ? tableAuthorityKey(this.authority)
        : null;
      this.stopObservation();
      this.update({
        ...initialNetworkFlowControllerState,
        canRename: false,
        canDelete: false,
      });
      const change: NetworkFlowExtensionResourceChange = {
        resourceKind: "*",
        resourceId: "*",
        changeKind: "remove",
        reasonCode: "incident_closed",
      };
      this.localChange?.(change);
      this.emit(change);
    } else if (error.code === "authorization_denied") {
      if (tableId) {
        this.blockedWrites = this.authority
          ? tableAuthorityKey(this.authority)
          : null;
        this.update({ canRename: false, canDelete: false });
      } else this.revokeRead();
    } else if (
      tableId &&
      (error.code === "network_flow_table_not_active" ||
        error.code === "network_flow_table_not_found")
    ) {
      this.removeTable(tableId);
      const change: NetworkFlowExtensionResourceChange = {
        resourceKind: "network_flow_table",
        resourceId: tableId,
        changeKind: "remove",
        reasonCode: "soft_deleted",
      };
      this.localChange?.(change);
      this.emit(change);
    }
  }
  private removeTable(tableId: string, keepReceipt = false): void {
    this.invalidateLists();
    const operation = this.state.operation;
    if (
      !keepReceipt &&
      operation?.attempt.target.network_flow_table_id === tableId
    )
      this.stopObservation();
    const draft =
      this.state.draft?.target.network_flow_table_id === tableId
        ? null
        : this.state.draft;
    this.update({
      ...networkFlowControllerReducer(this.state, {
        type: "remove_table",
        tableId,
      }),
      draft,
      operation:
        !keepReceipt &&
        operation?.attempt.target.network_flow_table_id === tableId
          ? null
          : this.state.operation,
      presentation:
        this.state.presentation === "draft" && draft === null
          ? null
          : this.state.presentation,
    });
  }
  private emit(change: NetworkFlowExtensionResourceChange): void {
    for (const listener of this.changes) listener(change);
  }
  readonly onResourceChange = async (
    change: NetworkFlowExtensionResourceChange,
  ): Promise<void> => {
    if (change.resourceKind === "network_flow_graph_view") return;
    if (change.resourceKind === "*" && change.changeKind === "remove") {
      if (change.reasonCode === "session_revoked")
        this.handleFailure(
          new NetworkFlowRequestError({
            code: "session_required",
            status: 401,
            retryAction: "do_not_retry",
            retryable: false,
            safeMessage: "Session recovery is required.",
          }),
        );
      else if (change.reasonCode === "incident_closed")
        this.handleFailure(
          new NetworkFlowRequestError({
            code: "incident_closed",
            status: 409,
            retryAction: "do_not_retry",
            retryable: false,
            safeMessage: "The incident is closed.",
          }),
        );
      else this.clearAuthorization();
      this.emit(change);
      return;
    }
    this.invalidateLists();
    if (change.changeKind === "remove") {
      const removedName = this.state.tables.find(
        (table) => table.network_flow_table_id === change.resourceId,
      )?.display_name;
      if (removedName) this.update({ notice: `${removedName} was deleted.` });
      this.removeTable(
        change.resourceId,
        this.state.operation?.status === "acknowledged" &&
          this.state.operation.attempt.action === "delete",
      );
    } else if (
      this.state.draft &&
      (change.resourceId === "*" ||
        change.resourceId === this.state.draft.target.network_flow_table_id)
    )
      this.update({
        draft: {
          ...this.state.draft,
          reviewRequired: true,
          confirmation: "",
          error: "The table changed. Refresh and review before submitting.",
        },
      });
    this.emit(change);
    if (this.active) await this.loadTables();
  };
}
