import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import type { Dispatch, SetStateAction } from "react";
import { describe, expect, it, vi } from "vitest";
import {
  fullWorkbookViewRow,
  workbookCollectionValue,
} from "../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  timelineMentionForAutoResolutionNotice,
  useTimelineMentionActions,
} from "./hooks/useTimelineMentionActions";
import type { DismissedMention } from "./models/workbookMentionChips";
import {
  normalizeTimelineFullRow,
  rowFromApi,
} from "./models/workbookTimelineModel";
import type { TimelineMentionPort } from "./ports/TimelineMentionPort";

const timelineContract = requireViewContract(timelineViewSchemaId);
const recordId = "11111111-1111-4111-8111-111111111111";
const resolvedRecordId = "22222222-2222-4222-8222-222222222222";
const mentionId = "33333333-3333-4333-8333-333333333333";
const itemRef = `entity_mention:${mentionId}`;
const notice = {
  entityType: "host" as const,
  fieldKey: "timeline.host_refs" as const,
  itemRef,
  matchedAliasText: "notice alias",
  rawText: " notice raw text ",
  resolvedRecordId,
  rowRecordId: recordId,
};

function mentionRow({
  itemKind = "resolved_ref",
  mentionRowVersion = 7,
  rawText = " current raw text ",
  rowVersion = 4,
}: {
  readonly itemKind?: string;
  readonly mentionRowVersion?: number;
  readonly rawText?: string;
  readonly rowVersion?: number;
} = {}) {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timelineContract, recordId, rowVersion, {
        "timeline.activity_synopsis_text": "Mention row",
        "timeline.host_refs": workbookCollectionValue(false, [
          {
            auto_resolved: true,
            confidence: 97,
            display_text: "Current display",
            entity_type: "host",
            item_kind: itemKind,
            item_ref: itemRef,
            matched_alias_text: "current alias",
            mention_row_version: mentionRowVersion,
            provenance: "auto_match",
            raw_text: rawText,
            resolution_method: "auto_match",
            resolved_record_id: resolvedRecordId,
          },
        ]),
      }),
      "mention action controller fixture",
    ),
  );
}

function controller({
  portResult,
  rowsRef = { current: [mentionRow()] },
  waitRow = mentionRow({ mentionRowVersion: 9, rowVersion: 5 }),
}: {
  readonly portResult: Awaited<ReturnType<TimelineMentionPort["resolve"]>>;
  readonly rowsRef?: { current: ReturnType<typeof mentionRow>[] };
  readonly waitRow?: ReturnType<typeof mentionRow>;
}) {
  let queuedWork: (() => Promise<void>) | null = null;
  const resolve = vi
    .fn<TimelineMentionPort["resolve"]>()
    .mockResolvedValue(portResult);
  const loadRows = vi.fn(
    async (options: { afterProjectionCommit?: () => void }) => {
      options.afterProjectionCommit?.();
    },
  );
  const setDismissedMentions = vi.fn();
  const mocks = {
    beginSave: vi.fn(),
    beginViewportContinuity: vi.fn(() => 41),
    clearViewportContinuity: vi.fn(),
    enqueueSaveWork: vi.fn((work: () => Promise<void>) => {
      queuedWork = work;
    }),
    finishSave: vi.fn(),
    loadRows,
    mentionPort: {
      createEntity: vi.fn<TimelineMentionPort["createEntity"]>(),
      resolve,
    },
    nextClientTxnId: vi.fn(() => "txn-undo"),
    requireViewportContinuitySourceRecord: vi.fn(),
    resolvePendingSocketTxn: vi.fn(),
    setDismissedMentionsByRow: setDismissedMentions as Dispatch<
      SetStateAction<Record<string, DismissedMention[]>>
    >,
    setInspectorMessage: vi.fn(),
    settleViewportContinuityFollowUp: vi.fn(),
    trackPendingSocketTxn: vi.fn(),
    waitForCommittedRecordIdle: vi.fn(async () => ({
      row: waitRow,
      rowVersion: waitRow.rowVersion ?? 0,
    })),
  };
  const rendered = renderHook(() =>
    useTimelineMentionActions({
      ...mocks,
      rowsRef,
    }),
  );
  return {
    ...rendered,
    getQueuedWork: () => queuedWork,
    mocks,
    resolve,
    setDismissedMentions,
  };
}

describe("useTimelineMentionActions auto-resolution undo", () => {
  it("converts the current resolved item without conflating notice, mention, and entity identity", () => {
    expect(
      timelineMentionForAutoResolutionNotice([mentionRow()], notice),
    ).toEqual({
      anchor: {
        entityMentionId: mentionId,
        fieldKey: "timeline.host_refs",
        itemRef,
        recordId,
        targetEntityRecordId: resolvedRecordId,
      },
      autoResolved: true,
      chipState: "auto_resolved",
      confidence: 97,
      displayText: "Current display",
      entityType: "host",
      fieldKey: "timeline.host_refs",
      isActiveRelationshipValue: true,
      itemRef,
      matchedAliasText: "current alias",
      mentionRowVersion: 7,
      priorTargetEntityRecordId: null,
      provenance: "auto_match",
      rawText: " current raw text ",
      resolutionMethod: "auto_match",
      resolvedRecordId,
      rowRecordId: recordId,
      sourceKind: "entity_mention",
      status: "resolved",
    });
    expect(timelineMentionForAutoResolutionNotice([], notice)).toBeNull();
    expect(
      timelineMentionForAutoResolutionNotice(
        [mentionRow({ itemKind: "unresolved_mention" })],
        notice,
      ),
    ).toBeNull();
    expect(
      timelineMentionForAutoResolutionNotice([mentionRow()], {
        ...notice,
        fieldKey: "timeline.identity_refs",
      }),
    ).toBeNull();
  });

  it("re-reads the committed mention version and preserves accepted refresh and continuity sequencing", async () => {
    const accepted = {
      kind: "accepted" as const,
      value: {
        entityMention: {
          entityType: "host" as const,
          rawText: " latest raw text ",
          resolutionMethod: null,
          rowVersion: 10,
          sourceFieldKey: "timeline.host_refs",
        },
        sourceRecord: { recordId, rowVersion: 6 },
      },
    };
    const { getQueuedWork, mocks, resolve, result, setDismissedMentions } =
      controller({ portResult: accepted });
    act(() => result.current.handleUndoAutoResolutionNotice(notice));
    expect(mocks.beginViewportContinuity).toHaveBeenCalledWith({
      kind: "row-inspect",
      recordId,
    });
    expect(mocks.beginSave).toHaveBeenCalledOnce();
    expect(mocks.enqueueSaveWork).toHaveBeenCalledOnce();

    await act(async () => getQueuedWork()?.());
    expect(resolve).toHaveBeenCalledWith({
      action: "revert_to_unresolved",
      baseMentionRowVersion: 9,
      clientTxnId: "txn-undo",
      expectedSourceRecordId: recordId,
      mentionId,
    });
    expect(mocks.trackPendingSocketTxn).toHaveBeenCalledWith("txn-undo");
    expect(mocks.requireViewportContinuitySourceRecord).toHaveBeenCalledWith(
      41,
      { recordId, minimumRowVersion: 6 },
    );
    expect(mocks.loadRows).toHaveBeenCalledWith({
      afterProjectionCommit: expect.any(Function),
      showLoading: false,
      sourceRecordRequirement: { recordId, minimumRowVersion: 6 },
      viewportContinuityToken: 41,
    });
    expect(mocks.finishSave).toHaveBeenLastCalledWith("Saved");
    expect(mocks.clearViewportContinuity).not.toHaveBeenCalled();

    const updateDismissed = setDismissedMentions.mock.calls[0]?.[0];
    if (typeof updateDismissed !== "function") {
      throw new Error("Expected the mention follow-up state updater");
    }
    expect(
      updateDismissed({
        [recordId]: [
          {
            autoResolved: true,
            entityType: "host",
            fieldKey: "timeline.host_refs",
            itemRef,
            mentionRowVersion: 7,
            rawText: " current raw text ",
            resolvedRecordId,
            resolutionMethod: "auto_match",
            rowRecordId: recordId,
          },
        ],
      }),
    ).toEqual({});
  });

  it("no-ops missing items and retains the existing rejection conflict path", async () => {
    const rejected = {
      kind: "rejected" as const,
      failure: { kind: "terminal" as const, message: "Stale mention version" },
    };
    const missing = controller({
      portResult: rejected,
      rowsRef: { current: [] },
    });
    act(() => missing.result.current.handleUndoAutoResolutionNotice(notice));
    expect(missing.mocks.beginSave).not.toHaveBeenCalled();
    expect(missing.resolve).not.toHaveBeenCalled();

    const active = controller({ portResult: rejected });
    act(() => active.result.current.handleUndoAutoResolutionNotice(notice));
    await act(async () => active.getQueuedWork()?.());
    expect(active.mocks.resolvePendingSocketTxn).toHaveBeenCalledWith(
      "txn-undo",
    );
    expect(active.mocks.clearViewportContinuity).toHaveBeenCalledWith(41);
    expect(active.mocks.setInspectorMessage).toHaveBeenCalledWith({
      primaryMessage: "Stale mention version",
      technicalFields: [],
    });
    expect(active.mocks.finishSave).toHaveBeenLastCalledWith("Conflict");
    expect(active.mocks.loadRows).not.toHaveBeenCalled();
  });
});
