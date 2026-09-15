import { renderHook } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { createTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { useTimelineSourceWriteCoordination } from "../hooks/useTimelineSourceWriteCoordination";
import { createDraftRow } from "../models/timelineRowModel";

it("coordinates an evicted original source and refuses its retained unsubmitted draft without scanning query pages", async () => {
  const draft = createDraftRow(1);
  const row = {
    ...draft,
    key: "original",
    recordId: "original",
    rowVersion: 3,
  };
  const clean = { ...row, values: { ...row.committedValues } };
  let coordinate:
    | ((recordId: string, signal: AbortSignal) => Promise<boolean>)
    | null = null;
  const waitForIdle = vi.fn(async () => ({ row: clean, rowVersion: 3 }));
  const registry = createTimelineEditorDraftRegistry();
  const materializeRow = vi.spyOn(registry, "materializeRow").mockReturnValue({
    ...clean,
    values: { ...clean.values, activitySynopsisText: "unsubmitted" },
  });
  const hook = renderHook(() =>
    useTimelineSourceWriteCoordination({
      owner: {
        registerSourceCoordinator: (callback) => {
          coordinate = callback;
          return () => {
            coordinate = null;
          };
        },
      },
      available: true,
      rows: { current: [] },
      drafts: registry,
      committedRow: (id) => (id === "original" ? clean : null),
      waitForIdle,
    }),
  );
  const registered = coordinate as
    | ((recordId: string, signal: AbortSignal) => Promise<boolean>)
    | null;
  if (!registered) throw new Error("Missing source coordinator");
  const signal = new AbortController().signal;
  expect(await registered("original", signal)).toBe(false);
  expect(waitForIdle).not.toHaveBeenCalled();
  materializeRow.mockReturnValue(clean);
  expect(await registered("original", signal)).toBe(true);
  expect(waitForIdle).toHaveBeenCalledExactlyOnceWith("original", {
    signal,
    refreshIfMissing: false,
  });
  hook.unmount();
});
