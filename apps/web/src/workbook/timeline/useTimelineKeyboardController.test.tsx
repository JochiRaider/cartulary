import type { GridCellAnchor } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import type { KeyboardEvent as ReactKeyboardEvent } from "react";
import { describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { initialWorkbookRecordHistoryState } from "../inspector/workbookRecordHistoryModel";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { useTimelineKeyboardController } from "./hooks/useTimelineKeyboardController";
import {
  normalizeTimelineFullRow,
  rowFromApi,
} from "./models/timelineRowModel";

const recordId = "11111111-1111-4111-8111-111111111111";
const summaryFieldKey = "timeline.activity_synopsis_text";
const timelineContract = requireViewContract(timelineViewSchemaId);
const anchor: GridCellAnchor = {
  fieldKey: summaryFieldKey,
  rowIdentity: { kind: "core_record", recordId },
  surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
};

function workbookRow() {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timelineContract, recordId, 3, {
        [summaryFieldKey]: "Keyboard row",
      }),
      "keyboard controller fixture",
    ),
  );
}

function keyboardEvent<ElementType extends HTMLElement>({
  altKey = false,
  ctrlKey = false,
  currentTarget,
  key,
  metaKey = false,
  shiftKey = false,
  target = currentTarget,
}: {
  readonly altKey?: boolean;
  readonly ctrlKey?: boolean;
  readonly currentTarget: ElementType;
  readonly key: string;
  readonly metaKey?: boolean;
  readonly shiftKey?: boolean;
  readonly target?: EventTarget;
}) {
  const event = {
    altKey,
    ctrlKey,
    currentTarget,
    defaultPrevented: false,
    key,
    metaKey,
    preventDefault: vi.fn(() => {
      event.defaultPrevented = true;
    }),
    shiftKey,
    stopPropagation: vi.fn(),
    target,
  };
  return event as unknown as ReactKeyboardEvent<ElementType>;
}

function controller(
  overrides: {
    readonly currentTimelineAnchorFor?: () => GridCellAnchor | null;
    readonly row?: ReturnType<typeof workbookRow> | null;
    readonly selectedRowId?: string | null;
    readonly workbookFocusAnchor?: {
      readonly viewSchemaId: string;
      readonly recordId: string;
      readonly fieldKey: string;
    } | null;
  } = {},
) {
  const calls: string[] = [];
  const mocks = {
    clearRowHistory: vi.fn(() => calls.push("clear-history")),
    currentTimelineAnchorFor: vi.fn(
      overrides.currentTimelineAnchorFor ?? (() => anchor),
    ),
    elementRegistry: {
      containsActiveElement: vi.fn(() => false),
      focusMention: vi.fn(() => true),
      focusPanel: vi.fn(() => true),
      registerMention: vi.fn(),
      registerPanel: vi.fn(),
      registerRoot: vi.fn(),
      updateScope: vi.fn(),
    },
    handleTimelineGridContextKeyDown: vi.fn(),
    navigateTimelineFocusAnchor: vi.fn((_anchor, intent) =>
      calls.push(`navigate-${intent.key}-${intent.shiftKey}`),
    ),
    openRowHistory: vi.fn(() => calls.push("open-history")),
    queueCollectionSave: vi.fn(() => calls.push("save-collection")),
    queueScalarSave: vi.fn(() => calls.push("save-scalar")),
    recordTiming: vi.fn(() => calls.push("timing")),
    restoreTimelineFocusAnchor: vi.fn(() => calls.push("restore-focus")),
    setInspectorMessage: vi.fn((feedback) =>
      calls.push(
        `message-${
          feedback === null
            ? "null"
            : feedback.kind === "message"
              ? feedback.message
              : feedback.error.primaryMessage
        }`,
      ),
    ),
    setIsInspectorOpen: vi.fn((isOpen) => calls.push(`inspector-${isOpen}`)),
    setSelectedMentionRef: vi.fn((itemRef) => calls.push(`mention-${itemRef}`)),
    setSelectedRowId: vi.fn((selectedRecordId) =>
      calls.push(`row-${selectedRecordId}`),
    ),
    timelineRowForEventTarget: vi.fn(() => overrides.row ?? null),
  };
  const rendered = renderHook(() =>
    useTimelineKeyboardController({
      ...mocks,
      rowHistory: initialWorkbookRecordHistoryState(),
      selectedRowId: overrides.selectedRowId ?? null,
      workbookFocusAnchorRef: {
        current: overrides.workbookFocusAnchor ?? null,
      },
    }),
  );
  return { calls, mocks, result: rendered.result };
}

describe("useTimelineKeyboardController", () => {
  it("owns scalar commit, navigation, range, draft, unavailable, and inspector focus order", () => {
    const input = document.createElement("input");
    input.value = "Edited summary";
    const { calls, mocks, result } = controller();
    const enter = keyboardEvent({ currentTarget: input, key: "Enter" });
    act(() =>
      result.current.commands.onScalarEditorKeyDown(
        enter,
        "row-key",
        "activitySynopsisText",
        "grid",
      ),
    );
    expect(enter.preventDefault).toHaveBeenCalledOnce();
    expect(enter.stopPropagation).toHaveBeenCalledOnce();
    expect(calls).toEqual(["save-scalar", "navigate-Enter-false"]);
    expect(mocks.queueScalarSave).toHaveBeenCalledWith(
      "row-key",
      "activitySynopsisText",
      {
        continueOnFreshDraft: true,
        preserveInputFocus: false,
        surface: "grid",
      },
      "Edited summary",
    );

    calls.length = 0;
    const tab = keyboardEvent({ currentTarget: input, key: "Tab" });
    act(() =>
      result.current.commands.onScalarEditorKeyDown(
        tab,
        "row-key",
        "activitySynopsisText",
        "grid",
      ),
    );
    expect(calls).toEqual(["save-scalar"]);

    calls.length = 0;
    const range = keyboardEvent({
      currentTarget: input,
      key: "ArrowDown",
      shiftKey: true,
    });
    act(() =>
      result.current.commands.onScalarEditorKeyDown(
        range,
        "row-key",
        "activitySynopsisText",
        "grid",
      ),
    );
    expect(range.preventDefault).toHaveBeenCalledOnce();
    expect(range.stopPropagation).toHaveBeenCalledOnce();
    expect(calls).toEqual([]);

    const draftController = controller({
      currentTimelineAnchorFor: () => null,
    });
    const draftEnter = keyboardEvent({ currentTarget: input, key: "Enter" });
    act(() =>
      draftController.result.current.commands.onScalarEditorKeyDown(
        draftEnter,
        "draft-1",
        "activitySynopsisText",
        "grid",
      ),
    );
    expect(draftController.calls).toEqual(["timing", "save-scalar"]);
    expect(draftController.mocks.queueScalarSave).toHaveBeenCalledWith(
      "draft-1",
      "activitySynopsisText",
      {
        continueOnFreshDraft: true,
        preserveInputFocus: true,
        surface: "grid",
      },
      "Edited summary",
    );

    calls.length = 0;
    const unavailable = keyboardEvent({
      ctrlKey: true,
      currentTarget: input,
      key: "k",
    });
    act(() =>
      result.current.commands.onScalarEditorKeyDown(
        unavailable,
        "row-key",
        "activitySynopsisText",
        "grid",
      ),
    );
    expect(unavailable.preventDefault).not.toHaveBeenCalled();
    expect(unavailable.stopPropagation).not.toHaveBeenCalled();
    expect(calls).toEqual([]);

    const focusAnchor = {
      fieldKey: summaryFieldKey,
      recordId,
      viewSchemaId: timelineViewSchemaId,
    };
    const inspectorController = controller({
      workbookFocusAnchor: focusAnchor,
    });
    const escapeEvent = keyboardEvent({
      currentTarget: input,
      key: "Escape",
    });
    act(() =>
      inspectorController.result.current.commands.onScalarEditorKeyDown(
        escapeEvent,
        "row-key",
        "activitySynopsisText",
        "inspector",
      ),
    );
    expect(escapeEvent.preventDefault).toHaveBeenCalledOnce();
    expect(escapeEvent.stopPropagation).toHaveBeenCalledOnce();
    expect(
      inspectorController.mocks.restoreTimelineFocusAnchor,
    ).toHaveBeenCalledWith(focusAnchor);
  });

  it("owns collection commit and close behavior without unreachable editor shortcuts", () => {
    const input = document.createElement("input");
    input.value = "host mention";
    const { calls, mocks, result } = controller({ selectedRowId: recordId });
    const enter = keyboardEvent({ currentTarget: input, key: "Enter" });
    act(() =>
      result.current.commands.onCollectionEditorKeyDown(
        enter,
        "row-key",
        "timeline.host_refs",
        "hostRefs",
      ),
    );
    expect(enter.preventDefault).toHaveBeenCalledOnce();
    expect(enter.stopPropagation).toHaveBeenCalledOnce();
    expect(calls).toEqual(["save-collection", "navigate-Enter-false"]);
    expect(mocks.queueCollectionSave).toHaveBeenCalledWith(
      "row-key",
      "timeline.host_refs",
      "hostRefs",
      "host mention",
      "keyboard",
    );

    calls.length = 0;
    const escapeEvent = keyboardEvent({
      currentTarget: input,
      key: "Escape",
    });
    act(() =>
      result.current.commands.onCollectionEditorKeyDown(
        escapeEvent,
        "row-key",
        "timeline.host_refs",
        "hostRefs",
      ),
    );
    expect(escapeEvent.preventDefault).toHaveBeenCalledOnce();
    expect(escapeEvent.stopPropagation).toHaveBeenCalledOnce();
    expect(calls).toEqual([
      "row-null",
      "mention-null",
      "message-null",
      "clear-history",
      "restore-focus",
    ]);

    calls.length = 0;
    const quickLink = keyboardEvent({
      ctrlKey: true,
      currentTarget: input,
      key: "k",
    });
    act(() =>
      result.current.commands.onCollectionEditorKeyDown(
        quickLink,
        "row-key",
        "timeline.host_refs",
        "hostRefs",
      ),
    );
    expect(quickLink.preventDefault).not.toHaveBeenCalled();
    expect(quickLink.stopPropagation).not.toHaveBeenCalled();
    expect(calls).toEqual([]);
  });

  it("owns work-area shortcuts and leaves controls, menus, dialogs, and unavailable actions unconsumed", () => {
    const target = document.createElement("div");
    target.dataset.gridFieldKey = "timeline.host_refs";
    const row = workbookRow();
    const { calls, mocks, result } = controller({ row });
    const quickLink = keyboardEvent({
      ctrlKey: true,
      currentTarget: target,
      key: "k",
    });
    act(() => result.current.commands.onWorkAreaKeyDown(quickLink));
    expect(quickLink.preventDefault).toHaveBeenCalledOnce();
    expect(quickLink.stopPropagation).toHaveBeenCalledOnce();
    expect(calls).toEqual([
      `row-${recordId}`,
      "inspector-true",
      "message-No unresolved mention is available for quick link.",
    ]);

    calls.length = 0;
    const history = keyboardEvent({
      altKey: true,
      currentTarget: target,
      key: "h",
    });
    act(() => result.current.commands.onWorkAreaKeyDown(history));
    expect(history.preventDefault).toHaveBeenCalledOnce();
    expect(history.stopPropagation).toHaveBeenCalledOnce();
    expect(calls).toEqual(["open-history"]);
    expect(mocks.elementRegistry.focusPanel).toHaveBeenCalledWith(
      {
        recordId,
        rowVersion: 3,
        viewSchemaId: timelineViewSchemaId,
      },
      "history",
    );

    const menu = document.createElement("div");
    menu.setAttribute("role", "menu");
    const dialog = document.createElement("div");
    dialog.setAttribute("role", "dialog");
    for (const owner of [
      document.createElement("input"),
      document.createElement("button"),
      menu,
      dialog,
    ]) {
      const owned = keyboardEvent({
        altKey: true,
        currentTarget: owner,
        key: "h",
      });
      act(() =>
        result.current.commands.onWorkAreaKeyDown(
          owned as unknown as ReactKeyboardEvent<HTMLDivElement>,
        ),
      );
      expect(owned.preventDefault).not.toHaveBeenCalled();
      expect(owned.stopPropagation).not.toHaveBeenCalled();
    }

    target.dataset.gridFieldKey = summaryFieldKey;
    const unavailable = keyboardEvent({
      ctrlKey: true,
      currentTarget: target,
      key: "k",
    });
    act(() => result.current.commands.onWorkAreaKeyDown(unavailable));
    expect(unavailable.preventDefault).not.toHaveBeenCalled();
    expect(unavailable.stopPropagation).not.toHaveBeenCalled();
    expect(mocks.setSelectedMentionRef).not.toHaveBeenCalled();
  });
});
