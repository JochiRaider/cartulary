import type { GridDraftPasteIntent } from "@cartulary/grid-adapter";
import { renderHook } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { useTimelineClipboardPasteController } from "./useTimelineClipboardPasteController";

it("reserves creation presentation only for an admitted paste gesture", () => {
  const intent: GridDraftPasteIntent = {
    fieldKey: "timeline.activity_synopsis_text",
    input: {
      kind: "table",
      format: "tsv",
      rawText: "Alpha\tSource",
      values: [["Alpha", "Source"]],
    },
    targetResolution: {
      columns: ["timeline.activity_synopsis_text", "timeline.data_source_text"],
      rowTargets: [
        {
          kind: "create",
          createIndex: 0,
          surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
        },
      ],
    },
  };
  const paste = vi
    .fn<() => string | null>()
    .mockReturnValueOnce("accepted")
    .mockReturnValue(null);
  const cancel = vi.fn();
  const observeReceipt = vi.fn(() => cancel);
  const captureCreatedBatchPresentation = vi.fn(() => () => true);
  const { result, unmount } = renderHook(() =>
    useTimelineClipboardPasteController({
      canCreateRows: true,
      editable: true,
      grouped: false,
      clipboardPaste: { paste, observeReceipt },
      captureCreatedBatchPresentation,
      captureCreatedPasteReveal: () => vi.fn(),
      pendingSavesRefs: { saveQueueRef: { current: Promise.resolve() } },
      resolveTimelinePasteTargetResolution: () => null,
      resolveTimelineDraftPasteTargetResolution: () => ({
        anchor: null,
        targetResolution: intent.targetResolution,
      }),
      setError: vi.fn(),
    }),
  );
  expect(result.current.commands.handleDraftPaste(intent)).toBe(true);
  expect(result.current.commands.handleDraftPaste({ ...intent })).toBe(false);
  // Rejected admission must not invalidate the older accepted gesture's
  // presentation generation while its retained receipt is still pending.
  expect(paste).toHaveBeenCalledTimes(2);
  expect(captureCreatedBatchPresentation).toHaveBeenCalledTimes(1);
  expect(observeReceipt).toHaveBeenCalledTimes(1);
  unmount();
  expect(cancel).toHaveBeenCalledTimes(1);
  performance.clearMarks("cartulary.workbook.timeline_paste_accepted");
});
