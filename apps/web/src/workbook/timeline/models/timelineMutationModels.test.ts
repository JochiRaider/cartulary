import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { PendingReplayUnitState } from "../../utils/workbookPendingQueue";
import { planTimelineAcceptedMutationEffects } from "./timelineAcceptedMutationEffects";
import { projectAcceptedTimelineRow } from "./timelineAcceptedProjection";
import { projectTimelineCollectionPresentation } from "./timelineCollectionPresentation";
import { TimelineCommittedVersionLedger } from "./timelineCommittedVersionLedger";
import type { TimelineReplayContext } from "./timelineControllerPorts";
import { reconcileDiscardedTimelineUnit } from "./timelineDiscardedReconciliation";
import { timelineFieldBinding } from "./timelineFieldRegistry";
import {
  decideTimelineCollectionCommit,
  planTimelineCollectionMutation,
  planTimelineScalarMutation,
} from "./timelineMutationQueueAdmission";
import { createDraftRow, type WorkbookRow } from "./timelineRowModel";

function savedRow(recordId = "timeline-1", rowVersion = 4): WorkbookRow {
  const draft = createDraftRow(1);
  return {
    ...draft,
    key: recordId,
    recordId,
    rowVersion,
    rawRow: {
      view_schema_id: timelineViewSchemaId,
      record_id: recordId,
      row_version: rowVersion,
      cells: {
        "timeline.activity_synopsis_text": {
          value: draft.values.activitySynopsisText,
        },
      },
    },
  };
}

function pendingPatch(
  id: string,
  enqueueOrder: number,
  value: string,
): PendingReplayUnitState {
  return {
    id,
    kind: "patch",
    source: "autosave",
    incidentId: "incident-1",
    clientInstanceId: "client-1",
    viewSchemaId: timelineViewSchemaId,
    rowKey: "timeline-1",
    recordId: "timeline-1",
    payloadIntent: {
      base_row_version: 4,
      changes: [{ field_key: "timeline.activity_synopsis_text", value }],
    },
    clientTxnId: `txn-${id}`,
    mutationSignature: `signature-${id}`,
    coalesceKey: "record:timeline-1",
    enqueueOrder,
    operationClass: "hot_path",
    status: "queued",
    identity: {
      kind: "patch",
      route_scope: { record_id: "timeline-1" },
      record_id: "timeline-1",
      client_txn_id: `txn-${id}`,
      view_schema_id: timelineViewSchemaId,
      base_row_version: 4,
      changes: [{ field_key: "timeline.activity_synopsis_text", value }],
    },
  };
}

describe("Timeline mutation models", () => {
  it("admits exact scalar and collection intents and deduplicates keyboard blur", () => {
    const row = savedRow();
    const changed = {
      ...row,
      values: { ...row.values, activitySynopsisText: "changed" },
    };
    expect(
      planTimelineScalarMutation({
        allowZeroFieldCreate: false,
        clientTxnId: "txn-1",
        focusField: "activitySynopsisText",
        hasConflict: false,
        pendingSignature: undefined,
        row: changed,
      }),
    ).toMatchObject({
      kind: "admit",
      visibleEdit: {
        fieldKey: "timeline.activity_synopsis_text",
        value: "changed",
      },
    });
    expect(
      planTimelineScalarMutation({
        allowZeroFieldCreate: false,
        clientTxnId: "txn-2",
        focusField: "activitySynopsisText",
        hasConflict: true,
        pendingSignature: undefined,
        row: changed,
      }),
    ).toMatchObject({ kind: "rejected", outcome: { kind: "conflict" } });
    expect(
      decideTimelineCollectionCommit({
        draftValue: "host-a",
        priorKeyboardCommitValue: "host-a",
        source: "blur",
      }),
    ).toEqual({ admit: false, nextKeyboardCommitValue: null });
    expect(
      planTimelineCollectionMutation({
        clientTxnId: "txn-3",
        draftValue: "host-a",
        effectiveRow: row,
        fieldKey: "timeline.host_refs",
        pendingSignature: undefined,
      }),
    ).toMatchObject({ kind: "admit" });
  });

  it("projects accepted draft replacement and its deterministic continuation", () => {
    const draft = createDraftRow(1);
    const committed = savedRow();
    const projection = projectAcceptedTimelineRow({
      committed,
      currentRows: [draft],
      nextDraftIndex: () => 2,
      rowKey: draft.key,
    });
    expect(projection.rows.map((row) => row.key)).toEqual([
      "timeline-1",
      "draft-2",
    ]);
    expect(projection.createdFromDraft).toBe(true);
    expect(
      planTimelineAcceptedMutationEffects({
        committed,
        continueOnFreshDraft: true,
        detectAutoResolution: true,
        projection,
        promoteToCommittedRowInspect: false,
        selectedRowId: null,
      }),
    ).toMatchObject({
      continuity: { kind: "fresh_draft", recordId: "timeline-1" },
      createdRecordId: "timeline-1",
    });
  });

  it("reconciles discarded units from committed state and reapplies later FIFO work", () => {
    const row = savedRow();
    const discarded = pendingPatch("discarded", 1, "discarded");
    const remaining = pendingPatch("remaining", 2, "retained");
    const context: TimelineReplayContext = {
      focusField: "activitySynopsisText",
      focusKey: "timeline-1:activitySynopsisText:grid",
      surface: "grid",
      rowSnapshot: row,
      continueOnFreshDraft: false,
      detectAutoResolution: false,
      promoteToCommittedRowInspect: false,
      viewportContinuityToken: 1,
    };
    const plan = reconcileDiscardedTimelineUnit({
      committedRow: row,
      contextByUnitId: new Map([
        [discarded.id, context],
        [remaining.id, context],
      ]),
      currentRows: [
        {
          ...row,
          values: { ...row.values, activitySynopsisText: "discarded" },
        },
      ],
      discardedUnit: discarded,
      nextDraftIndex: () => 2,
      remainingUnits: [remaining],
    });
    expect(plan.rows?.[0]).toMatchObject({
      pendingSignature: "signature-remaining",
      values: { activitySynopsisText: "retained" },
    });
    expect(plan.cancelEdit).toEqual({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: "timeline-1",
    });
  });

  it("maintains committed row versions as a monotonic reference-preserving ledger", () => {
    const ledger = new TimelineCommittedVersionLedger();
    const versionFour = savedRow("timeline-1", 4);
    const accepted = ledger.accept(versionFour, [versionFour]);
    expect(accepted).toMatchObject({
      accepted: true,
      stale: false,
    });
    const epoch = ledger.currentEpoch();
    expect(ledger.accept(accepted.row, [accepted.row]).row).toBe(accepted.row);
    expect(ledger.currentEpoch()).toBe(epoch);
    expect(
      ledger.accept(savedRow("timeline-1", 3), [accepted.row]),
    ).toMatchObject({ accepted: false, row: accepted.row, stale: true });
    expect(ledger.knownVersion("timeline-1")).toBe(4);
  });

  it("discriminates relationship and tag presentation without union downcasts", () => {
    const hostBinding = timelineFieldBinding("timeline.host_refs");
    const tagBinding = timelineFieldBinding("timeline.tags");
    if (
      hostBinding.kind !== "collection" ||
      hostBinding.collectionKind !== "relationship" ||
      tagBinding.kind !== "collection" ||
      tagBinding.collectionKind !== "tag"
    ) {
      throw new Error("Timeline collection registry is incomplete.");
    }
    const row: WorkbookRow = {
      ...savedRow(),
      collectionValues: {
        hostRefs: [
          {
            itemRef: "mention-1",
            entityType: "host",
            itemKind: "unresolved_mention",
            displayText: "host-a",
            rawText: "host-a",
            resolvedRecordId: null,
            mentionRowVersion: 1,
            resolutionMethod: null,
            autoResolved: false,
            provenance: null,
            confidence: null,
            matchedAliasText: null,
          },
        ],
        identityRefs: [],
        tags: [
          {
            itemRef: "tag-1",
            itemKind: "future_tag_member",
            displayText: "credential-access",
            rawText: "credential-access",
          },
        ],
      },
    };
    expect(
      projectTimelineCollectionPresentation({
        binding: hostBinding,
        entityIndex: {},
        row,
      }),
    ).toMatchObject({
      kind: "relationship",
      visibleItems: [{ kind: "relationship", itemRef: "mention-1" }],
    });
    expect(
      projectTimelineCollectionPresentation({
        binding: tagBinding,
        entityIndex: {},
        row,
      }),
    ).toMatchObject({
      kind: "tag",
      visibleItems: [
        { kind: "tag", itemRef: "tag-1", displayText: "credential-access" },
      ],
    });
  });
});
