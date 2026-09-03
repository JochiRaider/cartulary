import { describe, expect, it } from "vitest";
import {
  decideWorkbookApplicationShortcut,
  type WorkbookApplicationShortcutContext,
} from "./workbookApplicationShortcuts";

const eligibleContext: WorkbookApplicationShortcutContext = {
  capabilities: {
    closeInspector: false,
    history: true,
    linkedEvidence: true,
    quickLink: true,
  },
  focusOwner: "grid_navigation",
  previewableEvidenceCount: 1,
  rowKind: "committed",
  selectionIdentity: "record-1",
};

describe("Workbook application shortcuts", () => {
  it("returns explicit commands and consumption decisions", () => {
    for (const event of [
      { ctrlKey: true, key: "k" },
      { metaKey: true, key: "K" },
    ]) {
      expect(decideWorkbookApplicationShortcut(event, eligibleContext)).toEqual(
        {
          kind: "quick_link",
          preventDefault: true,
          stopPropagation: true,
        },
      );
    }
    expect(
      decideWorkbookApplicationShortcut({ key: " " }, eligibleContext),
    ).toEqual({
      destination: "sole_previewable_item",
      kind: "preview_linked_evidence",
      preventDefault: true,
      stopPropagation: true,
    });
    for (const key of ["h", "H"]) {
      expect(
        decideWorkbookApplicationShortcut(
          { altKey: true, key },
          eligibleContext,
        ),
      ).toEqual({
        kind: "open_history",
        preventDefault: true,
        stopPropagation: true,
      });
    }
    expect(
      decideWorkbookApplicationShortcut(
        { key: "Escape" },
        {
          ...eligibleContext,
          capabilities: {
            ...eligibleContext.capabilities,
            closeInspector: true,
          },
          focusOwner: "editor",
        },
      ),
    ).toEqual({
      kind: "close_inspector",
      preventDefault: true,
      stopPropagation: true,
    });
  });

  it("does not consume an unavailable command", () => {
    for (const focusOwner of [
      "editor",
      "inspector",
      "menu",
      "overlay",
    ] as const) {
      expect(
        decideWorkbookApplicationShortcut(
          { ctrlKey: true, key: "k" },
          { ...eligibleContext, focusOwner },
        ),
      ).toEqual({
        kind: "none",
        preventDefault: false,
        stopPropagation: false,
      });
    }

    for (const context of [
      { ...eligibleContext, rowKind: "draft" as const },
      { ...eligibleContext, selectionIdentity: null },
      {
        ...eligibleContext,
        capabilities: {
          ...eligibleContext.capabilities,
          quickLink: false,
        },
      },
    ]) {
      expect(
        decideWorkbookApplicationShortcut({ ctrlKey: true, key: "k" }, context),
      ).toEqual({
        kind: "none",
        preventDefault: false,
        stopPropagation: false,
      });
    }
    expect(
      decideWorkbookApplicationShortcut(
        { altKey: true, ctrlKey: true, key: "k" },
        eligibleContext,
      ),
    ).toEqual({
      kind: "none",
      preventDefault: false,
      stopPropagation: false,
    });
    expect(
      decideWorkbookApplicationShortcut(
        { altKey: true, key: "h" },
        {
          ...eligibleContext,
          capabilities: {
            ...eligibleContext.capabilities,
            history: false,
          },
        },
      ),
    ).toEqual({
      kind: "none",
      preventDefault: false,
      stopPropagation: false,
    });
  });

  it("does not claim Grid Adapter or native paste and navigation keys", () => {
    for (const event of [
      { key: "ArrowDown" },
      { key: "Enter" },
      { key: "Tab" },
      { ctrlKey: true, key: "v" },
      { metaKey: true, key: "v" },
    ]) {
      expect(decideWorkbookApplicationShortcut(event, eligibleContext)).toEqual(
        {
          kind: "none",
          preventDefault: false,
          stopPropagation: false,
        },
      );
    }
  });
});
