import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import type { PendingReplayPayloadIntent } from "../../utils/workbookPendingQueue";
import { buildStableMutationSignature } from "../../utils/workbookPendingQueue";
import {
  type CollectionFieldKey,
  type RowValues,
  timelineScalarBindingForValueKey,
} from "./timelineFieldRegistry";
import {
  buildCollectionPatchIntent,
  buildCreatePayload,
  buildScalarPatchIntent,
} from "./timelineMutationIntents";
import type { WorkbookRow } from "./timelineRowModel";

export type TimelineMutationAdmission =
  | {
      readonly kind: "admit";
      readonly mutationSignature: string;
      readonly payloadIntent: PendingReplayPayloadIntent;
      readonly visibleEdit: {
        readonly fieldKey: string;
        readonly value: unknown;
      };
    }
  | { readonly kind: "accepted_duplicate" }
  | { readonly kind: "accepted_no_change" }
  | {
      readonly kind: "rejected";
      readonly outcome: GridEditCommitOutcome;
    };

export function planTimelineScalarMutation({
  allowZeroFieldCreate,
  clientTxnId,
  focusField,
  hasConflict,
  pendingSignature,
  row,
}: {
  readonly allowZeroFieldCreate: boolean;
  readonly clientTxnId: string;
  readonly focusField: keyof RowValues;
  readonly hasConflict: boolean;
  readonly pendingSignature: string | undefined;
  readonly row: WorkbookRow;
}): TimelineMutationAdmission {
  if (hasConflict) {
    return {
      kind: "rejected",
      outcome: {
        kind: "conflict",
        message: "Resolve the existing field conflict before editing.",
      },
    };
  }
  const payloadIntent =
    row.recordId === null
      ? buildCreatePayload(row, clientTxnId, { allowZeroFieldCreate })
      : buildScalarPatchIntent(row, clientTxnId);
  if (payloadIntent === null) return { kind: "accepted_no_change" };
  const mutationSignature = buildStableMutationSignature(payloadIntent);
  if (mutationSignature === pendingSignature) {
    return { kind: "accepted_duplicate" };
  }
  const binding = timelineScalarBindingForValueKey(focusField);
  return {
    kind: "admit",
    mutationSignature,
    payloadIntent,
    visibleEdit: {
      fieldKey: binding.fieldKey,
      value: row.values[focusField],
    },
  };
}

export function planTimelineCollectionMutation({
  clientTxnId,
  draftValue,
  effectiveRow,
  fieldKey,
}: {
  readonly clientTxnId: string;
  readonly draftValue: string;
  readonly effectiveRow: WorkbookRow;
  readonly fieldKey: CollectionFieldKey;
}): TimelineMutationAdmission {
  const payloadIntent =
    effectiveRow.recordId === null
      ? buildCreatePayload(effectiveRow, clientTxnId)
      : buildCollectionPatchIntent(fieldKey, draftValue, clientTxnId);
  if (payloadIntent === null) return { kind: "accepted_no_change" };
  const mutationSignature = buildStableMutationSignature(payloadIntent);
  return {
    kind: "admit",
    mutationSignature,
    payloadIntent,
    visibleEdit: { fieldKey, value: draftValue },
  };
}
