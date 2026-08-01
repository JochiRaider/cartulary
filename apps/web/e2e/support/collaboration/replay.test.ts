// @vitest-environment jsdom

import {
  gridRowTestId,
  gridShellTestId,
  rowCellTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../contracts/workbookSurfaces";
import { currentTimelineSummary } from "./replay";

type RowState = {
  recordId: string;
  visible?: boolean;
  editorValue?: string;
  displayValue?: string;
};

type LocatorRef = {
  element: HTMLElement;
  row?: RowState;
  visible: boolean;
};

class FakeLocator {
  constructor(private readonly resolve: () => LocatorRef[]) {}

  async count() {
    return this.resolve().length;
  }

  async isVisible() {
    return this.resolve()[0]?.visible === true;
  }

  getByTestId(testId: string) {
    return new FakeLocator(() =>
      this.resolve().flatMap((reference) => {
        const row = reference.row;
        if (!row) return [];
        const editorID = timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: row.recordId,
          surface: "grid",
        });
        if (testId === editorID && row.editorValue !== undefined) {
          const input = document.createElement("input");
          input.value = row.editorValue;
          return [{ element: input, row, visible: reference.visible }];
        }
        if (
          testId ===
            rowCellTestId(row.recordId, "timeline.activity_synopsis_text") &&
          row.displayValue !== undefined
        ) {
          const display = document.createElement("span");
          display.textContent = row.displayValue;
          return [{ element: display, row, visible: reference.visible }];
        }
        return [];
      }),
    );
  }

  locator(selector: string) {
    if (selector !== '[role="row"][data-grid-record-id]') {
      throw new Error(`unsupported fake locator selector ${selector}`);
    }
    return this;
  }

  async evaluate<T>(callback: (element: HTMLElement) => T) {
    const reference = this.resolve()[0];
    if (!reference) throw new Error("fake locator resolved no element");
    return callback(reference.element);
  }

  async evaluateAll<T>(callback: (elements: HTMLElement[]) => T) {
    return callback(this.resolve().map((reference) => reference.element));
  }
}

function rowReference(row: RowState): LocatorRef {
  const element = document.createElement("div");
  element.setAttribute("role", "row");
  element.setAttribute("data-grid-record-id", row.recordId);
  return { element, row, visible: row.visible !== false };
}

function fakeTimelinePage(model: { rows: RowState[] }) {
  const getByTestId = (testId: string) => {
    if (testId === gridShellTestId(timelineViewSchemaId)) {
      return new FakeLocator(() => model.rows.map(rowReference));
    }
    return new FakeLocator(() =>
      model.rows
        .filter(
          (row) => testId === gridRowTestId(timelineViewSchemaId, row.recordId),
        )
        .map(rowReference),
    );
  };
  return { getByTestId } as unknown as Page;
}

describe("record-stable collaboration summary polling", () => {
  it("reacquires the requested row after reorder and replacement", async () => {
    const model: { rows: RowState[] } = {
      rows: [
        { recordId: "neighbor", editorValue: "neighbor base" },
        { recordId: "requested", displayValue: "requested first" },
      ],
    };
    const page = fakeTimelinePage(model);

    await expect(currentTimelineSummary(page, "requested")).resolves.toBe(
      "requested first",
    );
    model.rows = [
      { recordId: "requested", displayValue: "requested replacement" },
      { recordId: "neighbor", editorValue: "neighbor refreshed" },
    ];
    await expect(currentTimelineSummary(page, "requested")).resolves.toBe(
      "requested replacement",
    );
  });

  it("selects the requested editor when neighboring rows also have editors", async () => {
    const model: { rows: RowState[] } = {
      rows: [
        { recordId: "requested", editorValue: "requested queued" },
        { recordId: "neighbor", editorValue: "neighbor active" },
      ],
    };

    await expect(
      currentTimelineSummary(fakeTimelinePage(model), "requested"),
    ).resolves.toBe("requested queued");
  });

  it("follows the requested row from editor to display and reports mounted identities", async () => {
    const model: { rows: RowState[] } = {
      rows: [{ recordId: "requested", editorValue: "queued value" }],
    };
    const page = fakeTimelinePage(model);

    await expect(currentTimelineSummary(page, "requested")).resolves.toBe(
      "queued value",
    );
    model.rows = [
      { recordId: "requested", displayValue: "committed value" },
      { recordId: "mounted-neighbor", displayValue: "neighbor value" },
    ];
    await expect(currentTimelineSummary(page, "requested")).resolves.toBe(
      "committed value",
    );
    model.rows = [
      { recordId: "requested" },
      { recordId: "mounted-neighbor", displayValue: "neighbor value" },
    ];
    await expect(currentTimelineSummary(page, "requested")).rejects.toThrow(
      /timeline row requested.*mounted record IDs: requested,mounted-neighbor/,
    );
    await expect(
      currentTimelineSummary(page, "missing-record"),
    ).rejects.toThrow(
      /timeline row missing-record.*mounted record IDs: requested,mounted-neighbor/,
    );
  });
});
