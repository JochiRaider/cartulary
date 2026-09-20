import {
  type SheetRef,
  sheetRefKey,
  sheetRefsEqual,
} from "../../shared/sheetRef";
import { workbookEditRecoveryPresentation } from "../utils/workbookEditRecoveryPresentation";
import {
  deriveWorkbookSaveState,
  type PendingQueueStatusFacts,
} from "../utils/workbookPendingQueue";
import {
  selectWorkbookStatusSecondary,
  type WorkbookStatusAction,
  type WorkbookStatusSecondaryCandidate,
} from "../utils/workbookStatusSecondary";
import {
  type WorkbookConflictEntry,
  workbookConflictQueueKey,
} from "./workbookConflictModel";

export type WorkbookSaveLabel = "Conflict" | "Saved" | "Syncing";

export type WorkbookMutationSnapshot = {
  readonly authPaused: boolean;
  readonly blockedEdit: {
    readonly kind: "client_txn_conflict" | "terminal_replay_failure";
    readonly message: string;
    readonly unitId: string;
  } | null;
  readonly conflicts: readonly WorkbookConflictEntry[];
  readonly explicitInFlightCount: number;
  readonly queuedCount: number;
  readonly inFlightCount: number;
  readonly primaryLabel: WorkbookSaveLabel;
  readonly unresolvedConflictCount: number;
  readonly overflowMessage: string | null;
  readonly secondaryCandidates: readonly WorkbookStatusSecondaryCandidate[];
};

export type WorkbookStatusPresentation = WorkbookMutationSnapshot & {
  readonly affectedConflictCount: number;
  readonly secondary: WorkbookStatusSecondaryCandidate | null;
  readonly action: WorkbookStatusAction | null;
};

export type WorkbookRefreshStatusFact = {
  readonly sheetRef: SheetRef;
  readonly count: number;
};

type WorkbookMutationStatusInput = {
  readonly conflicts: readonly WorkbookConflictEntry[];
  readonly explicitInFlightCount: number;
  readonly queue: PendingQueueStatusFacts;
  readonly refreshes?: readonly WorkbookRefreshStatusFact[];
  readonly refreshDebts?: readonly string[];
  readonly authorityEpoch?: number;
};

/** Memoizes only the owner's explicit status inputs, never execution events. */
export function createWorkbookMutationStatusObservation() {
  let previous: WorkbookMutationStatusInput | null = null;
  let snapshot: WorkbookMutationSnapshot;
  return (input: WorkbookMutationStatusInput): WorkbookMutationSnapshot => {
    if (
      previous &&
      previous.conflicts === input.conflicts &&
      previous.explicitInFlightCount === input.explicitInFlightCount &&
      previous.queue === input.queue &&
      previous.refreshes === input.refreshes &&
      previous.refreshDebts === input.refreshDebts &&
      previous.authorityEpoch === input.authorityEpoch
    )
      return snapshot;
    previous = input;
    snapshot = projectWorkbookMutationStatus(input);
    return snapshot;
  };
}

export function projectWorkbookMutationStatus({
  conflicts,
  explicitInFlightCount,
  queue,
  refreshes = [],
  refreshDebts = [],
}: WorkbookMutationStatusInput): WorkbookMutationSnapshot {
  // The conflict store replaces an entry by its existing record/field key.
  // Prefer that current entry over a queue report of an older token/version.
  const storedKeys = new Set(conflicts.map((entry) => entry.key));
  const queueConflicts = queue.sameFieldConflicts.filter(
    (entry) => !storedKeys.has(entry.key),
  );
  const pendingCount =
    queue.queuedCount + queue.inFlightCount + explicitInFlightCount;
  const derived = deriveWorkbookSaveState({
    queuedCount: queue.queuedCount,
    inFlightCount: queue.inFlightCount,
    pendingMutationCount: explicitInFlightCount,
    authPaused: queue.authPaused && pendingCount > 0,
    refreshPaused: pendingCount > 0 && refreshes.some((fact) => fact.count > 0),
    halted: queue.halted,
    overflow: queue.overflow,
    sameFieldConflicts: queueConflicts,
    localDraftConflicts: conflicts.map((entry) => entry.conflict),
  });
  const candidates: WorkbookStatusSecondaryCandidate[] = [];
  const recovery =
    queue.halted === null
      ? null
      : workbookEditRecoveryPresentation({
          errorCode: queue.halted.error_code,
        });
  const blockedEdit =
    recovery === null || queue.halted === null
      ? null
      : {
          ...recovery,
          unitId: queue.halted.unit_id,
        };
  if (blockedEdit !== null)
    candidates.push({
      kind: blockedEdit.kind,
      scope: { kind: "workbook" },
      count: 1,
      message: blockedEdit.message,
      action: {
        kind:
          blockedEdit.kind === "client_txn_conflict"
            ? "transaction_recovery"
            : "terminal_failure",
        unitId: blockedEdit.unitId,
      },
    });
  const overflowMessage =
    queue.overflow === null
      ? null
      : "The local pending queue is full. Existing queued edits are retained; the current edit remains unsaved local work.";
  if (overflowMessage !== null)
    candidates.push({
      kind: "queue_overflow",
      scope: { kind: "workbook" },
      count: 1,
      message: overflowMessage,
      action: { kind: "overflow" },
    });
  const conflictGroups = new Map<
    string,
    { sheetRef: SheetRef; keys: Set<string>; firstKey: string }
  >();
  const addConflict = (sheetRef: SheetRef | undefined, key: string) => {
    if (sheetRef === undefined) return;
    const scopeKey = sheetRefKey(sheetRef);
    let group = conflictGroups.get(scopeKey);
    if (group === undefined) {
      group = { sheetRef, keys: new Set(), firstKey: key };
      conflictGroups.set(scopeKey, group);
    }
    group.keys.add(key);
  };
  for (const entry of conflicts) addConflict(entry.origin.sheetRef, entry.key);
  for (const entry of queueConflicts)
    addConflict(entry.sheetRef, workbookConflictQueueKey(entry));
  for (const group of conflictGroups.values()) {
    const count = group.keys.size;
    candidates.push({
      kind: "same_field_conflict",
      scope: { kind: "surface", sheetRef: group.sheetRef },
      count,
      message:
        count === 1
          ? "1 same-field conflict needs review."
          : `${count} same-field conflicts need review.`,
      action: { kind: "same_field_resolver", conflictKey: group.firstKey },
    });
  }
  if (queue.authPaused && pendingCount > 0)
    candidates.push({
      kind: "authentication_required",
      scope: { kind: "workbook" },
      count: pendingCount,
      message:
        queue.queuedCount + queue.inFlightCount > 0
          ? "Authentication is required before queued edits can replay."
          : "Authentication is required before pending work can continue.",
      action: { kind: "session_recovery" },
    });
  for (const view of refreshDebts)
    candidates.push({
      kind: "refresh_paused",
      scope: { kind: "surface", sheetRef: { kind: "view_schema", id: view } },
      count: 1,
      message: "Saved changes; view refresh pending.",
      action: { kind: "surface_refresh", viewSchemaId: view },
    });
  if (pendingCount > 0) {
    for (const refresh of refreshes) {
      if (refresh.count <= 0) continue;
      candidates.push({
        kind: "refresh_paused",
        scope: { kind: "surface", sheetRef: refresh.sheetRef },
        count: refresh.count,
        message:
          queue.queuedCount + queue.inFlightCount > 0
            ? "Queued edits are waiting for workbook refresh."
            : "Pending work is waiting for workbook refresh.",
        action: { kind: "recovery_list" },
      });
    }
    candidates.push({
      kind: "queued_or_in_flight",
      scope: { kind: "workbook" },
      count: pendingCount,
      message:
        queue.queuedCount > 0
          ? "Queued edits are waiting to replay."
          : "Workbook edits are syncing.",
      action: null,
    });
  }
  return {
    authPaused: queue.authPaused,
    blockedEdit,
    conflicts,
    explicitInFlightCount,
    queuedCount: queue.queuedCount,
    inFlightCount: queue.inFlightCount,
    primaryLabel: derived.primaryLabel,
    unresolvedConflictCount: derived.conflictAnchors.length,
    overflowMessage,
    secondaryCandidates: candidates,
  };
}

export function projectWorkbookStatusForSurface(
  snapshot: WorkbookMutationSnapshot,
  sheetRef?: SheetRef,
): WorkbookStatusPresentation {
  const secondary = selectWorkbookStatusSecondary(
    snapshot.secondaryCandidates,
    sheetRef,
  );
  const affected = snapshot.secondaryCandidates.find(
    (candidate) =>
      candidate.kind === "same_field_conflict" &&
      candidate.scope.kind === "surface" &&
      sheetRef !== undefined &&
      sheetRefsEqual(candidate.scope.sheetRef, sheetRef),
  );
  return {
    ...snapshot,
    affectedConflictCount: affected?.count ?? 0,
    secondary,
    action:
      secondary?.action ??
      (snapshot.primaryLabel === "Conflict" ? { kind: "recovery_list" } : null),
  };
}
