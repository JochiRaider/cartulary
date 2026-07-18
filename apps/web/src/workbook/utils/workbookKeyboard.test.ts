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
        { quickLink: true },
      ),
    ).toEqual({ kind: "quick-link", preventDefault: true });
    expect(
      mapWorkbookKeyboardCommand(
        { key: " ", shiftKey: false },
        { previewLinkedEvidence: true },
      ),
    ).toEqual({ kind: "preview-linked-evidence", preventDefault: true });
    expect(
      mapWorkbookKeyboardCommand({ key: "h", altKey: true }, { history: true }),
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
});
