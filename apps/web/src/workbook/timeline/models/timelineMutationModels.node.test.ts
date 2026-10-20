import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { planTimelineAcceptedMutationEffects } from "./timelineAcceptedMutationEffects";
import { projectAcceptedTimelineRow } from "./timelineAcceptedProjection";
import { projectTimelineCollectionPresentation } from "./timelineCollectionPresentation";
import { createTimelineCommittedVersionLedger } from "./timelineCommittedVersionLedger";
import { timelineFieldBinding } from "./timelineFieldRegistry";
import {
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

describe("Timeline mutation models", () => {
  it("releases evicted query observations while retaining drafts, the inspector source, and mutation version floors", () => {
    const ledger = createTimelineCommittedVersionLedger();
    const first = Array.from({ length: 300 }, (_, index) =>
      savedRow(`row-${index}`, 2),
    );
    ledger.replaceQueryRows(first, []);
    ledger.retainInspectorRecord("row-0");
    ledger.accept(savedRow("row-1", 5), first);
    const original = first[2];
    if (!original) throw new Error("Missing original row");
    const draft = {
      ...original,
      values: { ...original.values, activitySynopsisText: "unsubmitted" },
    };
    const next = Array.from({ length: 300 }, (_, index) =>
      savedRow(`row-${index + 100}`, 2),
    );
    ledger.replaceQueryRows(
      next,
      first.map((row) => (row.recordId === draft.recordId ? draft : row)),
    );
    expect(ledger.current("row-0", [])?.rowVersion).toBe(2);
    expect(ledger.current("row-1", [])?.rowVersion).toBe(5);
    expect(
      ledger.current("row-2", [])?.committedValues.activitySynopsisText,
    ).toBe(first[2]?.committedValues.activitySynopsisText);
    expect(ledger.current("row-3", [])).toBeNull();
    expect(ledger.knownVersion("row-3")).toBeUndefined();
    ledger.retainInspectorRecord(null);
    ledger.replaceQueryRows(next, next);
    expect(ledger.current("row-0", [])).toBeNull();
    expect(ledger.accept(savedRow("row-1", 4), next).stale).toBe(true);
    expect(ledger.current("row-399", [])?.rowVersion).toBe(2);
  });
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
        row: changed,
      }),
    ).toMatchObject({ kind: "rejected", outcome: { kind: "conflict" } });
    expect(
      planTimelineCollectionMutation({
        clientTxnId: "txn-3",
        draftValue: "host-a",
        effectiveRow: row,
        fieldKey: "timeline.host_refs",
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
    const refreshedBeforeAcknowledgement = projectAcceptedTimelineRow({
      committed,
      currentRows: [committed, draft],
      nextDraftIndex: () => 2,
      rowKey: draft.key,
    });
    expect(refreshedBeforeAcknowledgement.rows.map((row) => row.key)).toEqual([
      "timeline-1",
      "draft-2",
    ]);
    expect(refreshedBeforeAcknowledgement.createdFromDraft).toBe(true);
    expect(
      planTimelineAcceptedMutationEffects({
        committed,
        continueOnFreshDraft: true,
        projection,
        promoteToCommittedRowInspect: false,
        selectedRowId: null,
      }),
    ).toMatchObject({
      continuity: { kind: "fresh_draft", recordId: "timeline-1" },
      createdRecordId: "timeline-1",
    });
  });

  it("maintains committed row versions as a monotonic reference-preserving ledger", () => {
    const ledger = createTimelineCommittedVersionLedger();
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

  it("preserves canonical null empty and absent cells independently of display strings and queued work", () => {
    const ledger = createTimelineCommittedVersionLedger();
    const row = savedRow();
    if (!row.rawRow) throw new Error("Missing canonical row");
    const canonical = {
      ...row.rawRow,
      cells: {
        "timeline.raw_activity_text": { value: null },
        "timeline.activity_synopsis_text": { value: "" },
      },
    };
    const withDraft = {
      ...row,
      rawRow: canonical,
      values: { ...row.values, rawActivityText: "Unsubmitted" },
      pendingSignature: "pending",
    };
    ledger.accept(withDraft, [withDraft]);
    const saved = ledger.current(row.recordId ?? "", []);
    expect(saved?.rawRow).toBe(canonical);
    expect(saved?.values.rawActivityText).toBe("");
    expect(
      saved?.rawRow?.cells["timeline.raw_activity_text"]?.value,
    ).toBeNull();
    expect(saved?.rawRow?.cells["timeline.activity_synopsis_text"]?.value).toBe(
      "",
    );
    expect(saved?.rawRow?.cells).not.toHaveProperty(
      "timeline.data_source_text",
    );
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
            entityMentionId: "public-mention-1",
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
          {
            itemRef: "tag-2",
            itemKind: "tag",
            displayText: "hidden tag Ω",
            rawText: "hidden tag Ω",
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
      hiddenItemCount: 1,
      overflowRecordId: row.recordId,
      firstHiddenItemRef: "tag-2",
      visibleItems: [
        { kind: "tag", itemRef: "tag-1", displayText: "credential-access" },
      ],
    });
  });
});
