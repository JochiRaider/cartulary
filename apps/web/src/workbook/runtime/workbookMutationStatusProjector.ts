import { workbookEditRecoveryPresentation } from "../utils/workbookEditRecoveryPresentation";
import type { PendingQueueSnapshot } from "../utils/workbookPendingQueue";
import type { WorkbookStatusSecondaryCandidate } from "../utils/workbookStatusSecondary";
import type { WorkbookConflictEntry } from "./workbookConflictModel";

export type WorkbookMutationSnapshot = {
  readonly authPaused: boolean;
  readonly blockedEdit: {
    readonly kind: "client_txn_conflict" | "terminal_replay_failure";
    readonly message: string;
    readonly unitId: string;
  } | null;
  readonly conflictPanelOpen: boolean;
  readonly conflicts: readonly WorkbookConflictEntry[];
  readonly explicitInFlightCount: number;
  readonly primaryLabel: "Conflict" | "Saved" | "Syncing";
  readonly overflowMessage: string | null;
  readonly secondaryMessage: string | null;
  readonly secondaryCandidates: readonly WorkbookStatusSecondaryCandidate[];
};

export type WorkbookSurfaceSaveStateProjection = {
  readonly primaryLabel: "Conflict" | "Saved" | "Syncing";
  readonly secondaryMessage: string | null;
};

type WorkbookMutationStatusInput = {
  readonly conflictPanelOpen: boolean;
  readonly conflicts: readonly WorkbookConflictEntry[];
  readonly explicitInFlightCount: number;
  readonly queue: PendingQueueSnapshot;
  readonly surfaceSaveStates: ReadonlyMap<
    string,
    WorkbookSurfaceSaveStateProjection
  >;
};

function haltedCandidate(
  queue: PendingQueueSnapshot,
): WorkbookStatusSecondaryCandidate | null {
  if (queue.halted === null) return null;
  const haltedSurfaceId =
    queue.units.find((unit) => unit.id === queue.halted?.unit_id)
      ?.viewSchemaId ?? null;
  if (haltedSurfaceId === null) return null;
  return {
    kind:
      queue.halted.error_code === "client_txn_conflict"
        ? "client_txn_conflict"
        : "terminal_replay_failure",
    message: queue.halted.message,
    surfaceId: haltedSurfaceId,
  };
}

function overflowCandidate(
  queue: PendingQueueSnapshot,
): WorkbookStatusSecondaryCandidate | null {
  return queue.overflow?.view_schema_id === undefined
    ? null
    : {
        kind: "queue_overflow",
        message: queue.overflow.message,
        surfaceId: queue.overflow.view_schema_id,
      };
}

function queueConflictCandidates(
  queue: PendingQueueSnapshot,
): WorkbookStatusSecondaryCandidate[] {
  return queue.sameFieldConflicts.flatMap((conflict) =>
    conflict.view_schema_id === undefined
      ? []
      : [
          {
            kind: "same_field_conflict",
            message:
              queue.saveStatePresentation.secondaryMessage ??
              "Same-field conflict requires review.",
            surfaceId: conflict.view_schema_id,
          },
        ],
  );
}

function localConflictCandidates(
  conflicts: readonly WorkbookConflictEntry[],
): WorkbookStatusSecondaryCandidate[] {
  return conflicts.map((conflict) => ({
    kind: "same_field_conflict",
    message: "Same-field conflict requires review.",
    surfaceId: conflict.origin.viewSchemaId,
  }));
}

function queuedCandidates(
  queue: PendingQueueSnapshot,
): WorkbookStatusSecondaryCandidate[] {
  return Array.from(
    new Set(queue.units.map((unit) => unit.viewSchemaId)),
    (surfaceId): WorkbookStatusSecondaryCandidate => ({
      kind: queue.authPaused
        ? "authentication_required"
        : "queued_or_in_flight",
      message: queue.authPaused
        ? "Authentication is required before queued edits can replay."
        : "Queued edits are waiting to replay.",
      surfaceId,
    }),
  );
}

function surfaceCandidates(
  surfaceSaveStates: WorkbookMutationStatusInput["surfaceSaveStates"],
): WorkbookStatusSecondaryCandidate[] {
  return Array.from(surfaceSaveStates).flatMap(([surfaceId, state]) =>
    state.secondaryMessage === null
      ? []
      : [
          {
            kind:
              state.primaryLabel === "Conflict"
                ? "terminal_replay_failure"
                : "refresh_paused",
            message: state.secondaryMessage,
            surfaceId,
          },
        ],
  );
}

function secondaryCandidates({
  conflicts,
  queue,
  surfaceSaveStates,
}: Pick<
  WorkbookMutationStatusInput,
  "conflicts" | "queue" | "surfaceSaveStates"
>): WorkbookStatusSecondaryCandidate[] {
  return [
    haltedCandidate(queue),
    overflowCandidate(queue),
    ...queueConflictCandidates(queue),
    ...localConflictCandidates(conflicts),
    ...queuedCandidates(queue),
    ...surfaceCandidates(surfaceSaveStates),
  ].filter(
    (candidate): candidate is WorkbookStatusSecondaryCandidate =>
      candidate !== null,
  );
}

function primaryLabel(
  input: WorkbookMutationStatusInput,
): WorkbookMutationSnapshot["primaryLabel"] {
  const hasConflict =
    input.conflicts.length > 0 ||
    input.queue.halted !== null ||
    input.queue.overflow !== null ||
    input.queue.sameFieldConflicts.length > 0 ||
    Array.from(input.surfaceSaveStates.values()).some(
      (state) => state.primaryLabel === "Conflict",
    );
  if (hasConflict) return "Conflict";
  const hasPending =
    input.queue.queuedCount > 0 ||
    input.queue.inFlightCount > 0 ||
    input.explicitInFlightCount > 0 ||
    Array.from(input.surfaceSaveStates.values()).some(
      (state) => state.primaryLabel === "Syncing",
    );
  return hasPending ? "Syncing" : "Saved";
}

function projectedSecondaryMessage(
  input: WorkbookMutationStatusInput,
  projectedPrimaryLabel: WorkbookMutationSnapshot["primaryLabel"],
): string | null {
  if (
    input.queue.saveStatePresentation.primaryLabel === projectedPrimaryLabel &&
    input.queue.saveStatePresentation.secondaryMessage !== null
  ) {
    return input.queue.saveStatePresentation.secondaryMessage;
  }
  const surfaceMessage = Array.from(input.surfaceSaveStates.values()).find(
    (state) => state.primaryLabel === projectedPrimaryLabel,
  )?.secondaryMessage;
  if (surfaceMessage !== undefined && surfaceMessage !== null) {
    return surfaceMessage;
  }
  if (input.explicitInFlightCount === 0) return null;
  const suffix = input.explicitInFlightCount === 1 ? "" : "s";
  return `${input.explicitInFlightCount} explicit change${suffix} in flight`;
}

export function projectWorkbookMutationStatus({
  conflictPanelOpen,
  conflicts,
  explicitInFlightCount,
  queue,
  surfaceSaveStates,
}: WorkbookMutationStatusInput): WorkbookMutationSnapshot {
  const input = {
    conflictPanelOpen,
    conflicts,
    explicitInFlightCount,
    queue,
    surfaceSaveStates,
  };
  const projectedPrimaryLabel = primaryLabel(input);
  const blockedEdit =
    queue.halted === null
      ? null
      : workbookEditRecoveryPresentation({
          errorCode: queue.halted.error_code,
        });
  return {
    authPaused: queue.authPaused,
    blockedEdit:
      blockedEdit === null || queue.halted === null
        ? null
        : {
            kind: blockedEdit.kind,
            message: blockedEdit.message,
            unitId: queue.halted.unit_id,
          },
    conflictPanelOpen,
    conflicts,
    explicitInFlightCount,
    primaryLabel: projectedPrimaryLabel,
    overflowMessage: queue.overflow?.message ?? null,
    secondaryMessage: projectedSecondaryMessage(input, projectedPrimaryLabel),
    secondaryCandidates: secondaryCandidates({
      conflicts,
      queue,
      surfaceSaveStates,
    }),
  };
}
