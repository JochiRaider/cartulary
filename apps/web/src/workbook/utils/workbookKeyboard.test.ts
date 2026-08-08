import { describe, expect, it } from "vitest";

import { mapWorkbookKeyboardCommand } from "./workbookKeyboard";

describe("keyboard command contract", () => {
  it("maps required workbook keys without hidden paste macro behavior", () => {
    for (const key of ["ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight"]) {
      expect(mapWorkbookKeyboardCommand({ key })).toEqual({
        intent: { key, shiftKey: false },
        kind: "navigate",
        preventDefault: true,
      });
    }

    expect(mapWorkbookKeyboardCommand({ key: "Enter" })).toEqual({
      intent: { key: "Enter", shiftKey: false },
      kind: "navigate",
      preventDefault: true,
    });
    expect(
      mapWorkbookKeyboardCommand({ key: "Enter", shiftKey: true }),
    ).toEqual({
      intent: { key: "Enter", shiftKey: true },
      kind: "navigate",
      preventDefault: true,
    });
    expect(mapWorkbookKeyboardCommand({ key: "Tab" })).toEqual({
      intent: { key: "Tab", shiftKey: false },
      kind: "navigate",
      preventDefault: true,
    });

    expect(mapWorkbookKeyboardCommand({ key: "v", ctrlKey: true })).toEqual({
      kind: "paste-intent",
      preventDefault: false,
    });
    expect(mapWorkbookKeyboardCommand({ key: "v", metaKey: true })).toEqual({
      kind: "paste-intent",
      preventDefault: false,
    });

    expect(
      mapWorkbookKeyboardCommand(
        { key: "k", ctrlKey: true },
        {
          cell: { linkResolveCapability: true },
          committedRowIdentity: "record-1",
          mode: "grid_navigation",
          rowKind: "committed",
        },
      ),
    ).toEqual({ kind: "quick-link", preventDefault: true });
    expect(
      mapWorkbookKeyboardCommand(
        { key: " ", shiftKey: false },
        {
          inspectorGroups: ["evidence"],
          committedRowIdentity: "record-1",
          mode: "grid_navigation",
          previewableEvidenceCount: 0,
          rowKind: "committed",
        },
      ),
    ).toEqual({
      destination: "list_or_empty",
      kind: "preview-linked-evidence",
      preventDefault: true,
    });
    expect(
      mapWorkbookKeyboardCommand(
        { key: " " },
        {
          committedRowIdentity: "record-1",
          inspectorGroups: ["evidence"],
          mode: "grid_navigation",
          previewableEvidenceCount: 1,
          rowKind: "committed",
        },
      ),
    ).toEqual({
      destination: "sole_previewable_item",
      kind: "preview-linked-evidence",
      preventDefault: true,
    });
    expect(
      mapWorkbookKeyboardCommand(
        { key: "h", altKey: true },
        {
          inspectorGroups: ["history"],
          committedRowIdentity: "record-1",
          mode: "grid_navigation",
          rowKind: "committed",
        },
      ),
    ).toEqual({ kind: "open-history", preventDefault: true });
    expect(
      mapWorkbookKeyboardCommand({ key: "Escape" }, { closeInspector: true }),
    ).toEqual({ kind: "close-inspector", preventDefault: true });
  });

  it("fails closed when optional shortcut actions are unavailable", () => {
    expect(mapWorkbookKeyboardCommand({ key: "k", ctrlKey: true })).toEqual({
      kind: "none",
      preventDefault: false,
    });
    expect(mapWorkbookKeyboardCommand({ key: " " })).toEqual({
      kind: "none",
      preventDefault: false,
    });
    expect(mapWorkbookKeyboardCommand({ key: "h", altKey: true })).toEqual({
      kind: "none",
      preventDefault: false,
    });
    expect(mapWorkbookKeyboardCommand({ key: "Escape" })).toEqual({
      kind: "none",
      preventDefault: false,
    });
  });

  it("does not consume application shortcuts outside committed grid navigation", () => {
    const shortcutCases = [
      { event: { ctrlKey: true, key: "k" }, kind: "quick-link" },
      { event: { key: " " }, kind: "preview-linked-evidence" },
      { event: { altKey: true, key: "h" }, kind: "open-history" },
    ] as const;
    const eligibleContext = {
      cell: { linkResolveCapability: true },
      committedRowIdentity: "record-1",
      inspectorGroups: ["evidence", "history"] as const,
      mode: "grid_navigation" as const,
      previewableEvidenceCount: 1,
      rowKind: "committed" as const,
    };

    for (const mode of ["editor", "inspector", "menu"] as const) {
      for (const shortcut of shortcutCases) {
        expect(
          mapWorkbookKeyboardCommand(shortcut.event, {
            ...eligibleContext,
            mode,
          }),
        ).toEqual({ kind: "none", preventDefault: false });
      }
    }
    for (const rowKind of ["draft", "group", "none"] as const) {
      for (const shortcut of shortcutCases) {
        expect(
          mapWorkbookKeyboardCommand(shortcut.event, {
            ...eligibleContext,
            committedRowIdentity: null,
            rowKind,
          }),
        ).toEqual({ kind: "none", preventDefault: false });
      }
    }
    for (const shortcut of shortcutCases) {
      expect(
        mapWorkbookKeyboardCommand(shortcut.event, {
          ...eligibleContext,
          committedRowIdentity: null,
        }),
      ).toEqual({ kind: "none", preventDefault: false });
    }
    expect(
      mapWorkbookKeyboardCommand(
        { ctrlKey: true, key: "k" },
        { ...eligibleContext, cell: { linkResolveCapability: false } },
      ),
    ).toEqual({ kind: "none", preventDefault: false });
    expect(
      mapWorkbookKeyboardCommand(
        { key: " " },
        { ...eligibleContext, inspectorGroups: ["history"] },
      ),
    ).toEqual({ kind: "none", preventDefault: false });
    expect(
      mapWorkbookKeyboardCommand(
        { altKey: true, key: "h" },
        { ...eligibleContext, inspectorGroups: ["evidence"] },
      ),
    ).toEqual({ kind: "none", preventDefault: false });
  });
});
