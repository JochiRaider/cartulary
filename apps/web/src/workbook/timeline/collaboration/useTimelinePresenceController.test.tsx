import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import {
  emptyWorkbookPresence,
  projectWorkbookPresence,
} from "../../collaboration/workbookPresencePresentation";
import type { PresenceRecord } from "../../utils/workbookPresence";
import { useTimelinePresenceController } from "./useTimelinePresenceController";

const firstRecordId = "11111111-1111-4111-8111-111111111111";
const secondRecordId = "22222222-2222-4222-8222-222222222222";
const summaryFieldKey = "timeline.activity_synopsis_text";

function presence({
  connectionId,
  fieldKey,
  mode = "viewing",
  recordId = firstRecordId,
  sheetKind = "view_schema",
}: {
  readonly connectionId: string;
  readonly fieldKey?: string;
  readonly mode?: PresenceRecord["mode"];
  readonly recordId?: string;
  readonly sheetKind?: "saved_view" | "view_schema";
}): PresenceRecord {
  return {
    connection_id: connectionId,
    display_name: connectionId,
    expires_at: "2026-08-14T12:01:00Z",
    ...(fieldKey === undefined ? {} : { field_key: fieldKey }),
    mode,
    observed_at: "2026-08-14T12:00:00Z",
    record_id: recordId,
    sheet_ref: {
      kind: sheetKind,
      id:
        sheetKind === "view_schema"
          ? "cartulary.view.timeline.v2"
          : "saved-view-1",
    },
    user_id: `user-${connectionId}`,
  };
}

describe("useTimelinePresenceController", () => {
  it("derives row and editing-cell markers only from stable record and field identities", () => {
    const editing = presence({
      connectionId: "editing",
      fieldKey: summaryFieldKey,
      mode: "editing",
    });
    const viewing = presence({ connectionId: "viewing" });
    const wrongField = presence({
      connectionId: "wrong-field",
      fieldKey: "timeline.raw_activity_text",
      mode: "editing",
    });
    const wrongRow = presence({
      connectionId: "wrong-row",
      fieldKey: summaryFieldKey,
      mode: "editing",
      recordId: secondRecordId,
    });
    const alreadyScopedSavedView = presence({
      connectionId: "saved-view",
      fieldKey: summaryFieldKey,
      mode: "editing",
      sheetKind: "saved_view",
    });
    const records = [
      editing,
      viewing,
      wrongField,
      wrongRow,
      alreadyScopedSavedView,
      editing,
    ];
    const { result, rerender } = renderHook(
      ({ presenceRecords }) =>
        useTimelinePresenceController({
          presence: projectWorkbookPresence({
            records: presenceRecords,
            activeSheetRef: {
              kind: "view_schema",
              id: "cartulary.view.timeline.v2",
            },
            connectionId: null,
            nowMs: Date.parse("2026-08-14T12:00:00Z"),
          }).presentation,
          publishPresence: vi.fn(),
          resetKey: "surface-1",
        }),
      { initialProps: { presenceRecords: records } },
    );

    expect(
      result.current.snapshot
        .presenceForRow(firstRecordId)
        .users.map((record) => record.connection_id),
    ).toEqual(["editing", "wrong-field", "viewing"]);
    expect(
      result.current.snapshot
        .editingPresenceForCell(firstRecordId, summaryFieldKey)
        .users.map((record) => record.connection_id),
    ).toEqual(["editing"]);
    expect(result.current.snapshot.presenceForRow(null).users).toEqual([]);
    expect(
      result.current.snapshot.editingPresenceForCell(null, summaryFieldKey)
        .users,
    ).toEqual([]);

    rerender({ presenceRecords: [...records].reverse() });
    expect(
      result.current.snapshot
        .editingPresenceForCell(firstRecordId, summaryFieldKey)
        .users.map((record) => record.connection_id)
        .sort(),
    ).toEqual(["editing"]);
  });

  it("publishes coherent viewing and editing transitions and keeps selection from overriding an edit", () => {
    const publishPresence = vi.fn();
    const { result } = renderHook(() =>
      useTimelinePresenceController({
        presence: emptyWorkbookPresence,
        publishPresence,
        resetKey: "surface-1",
      }),
    );
    expect(result.current.snapshot.currentPresence).toEqual({
      fieldKey: null,
      mode: "viewing",
      recordId: null,
    });

    act(() => result.current.commands.publishViewingPresence(firstRecordId));
    expect(publishPresence).toHaveBeenLastCalledWith({
      fieldKey: null,
      mode: "viewing",
      recordId: firstRecordId,
    });
    act(() =>
      result.current.commands.publishEditModePresence(
        firstRecordId,
        summaryFieldKey,
        true,
      ),
    );
    expect(result.current.snapshot.currentPresence).toEqual({
      fieldKey: summaryFieldKey,
      mode: "editing",
      recordId: firstRecordId,
    });

    act(() => result.current.commands.publishViewingPresence(secondRecordId));
    expect(publishPresence).toHaveBeenCalledTimes(2);
    expect(result.current.snapshot.currentPresence.recordId).toBe(
      firstRecordId,
    );

    act(() =>
      result.current.commands.publishEditModePresence(
        null,
        summaryFieldKey,
        false,
      ),
    );
    expect(publishPresence).toHaveBeenLastCalledWith({
      fieldKey: null,
      mode: "viewing",
      recordId: firstRecordId,
    });
  });

  it("clears the local draft on surface or authorization lifecycle reset without publishing", () => {
    const publishPresence = vi.fn();
    const { result, rerender } = renderHook(
      ({ resetKey }) =>
        useTimelinePresenceController({
          presence: emptyWorkbookPresence,
          publishPresence,
          resetKey,
        }),
      { initialProps: { resetKey: "surface-1:authorized" } },
    );
    act(() =>
      result.current.commands.publishEditModePresence(
        firstRecordId,
        summaryFieldKey,
        true,
      ),
    );
    publishPresence.mockClear();
    rerender({ resetKey: "surface-1:access-lost" });
    expect(result.current.snapshot.currentPresence).toEqual({
      fieldKey: null,
      mode: "viewing",
      recordId: null,
    });
    expect(result.current.snapshot.currentPresenceRef.current).toEqual(
      result.current.snapshot.currentPresence,
    );
    expect(publishPresence).not.toHaveBeenCalled();
  });
});
