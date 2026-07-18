import {
  timelineDraftEvidenceFileInputTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import { TimelineEvidencePanel } from "./TimelineEvidencePanel";

const rowValues = {
  dateEnteredText: "",
  analystText: "",
  mitreStageText: "",
  deviceObjectText: "",
  ipAddressText: "",
  activityUTCText: "",
  activityLocalText: "",
  rawActivityText: "",
  activitySynopsisText: "",
  dataSourceText: "",
};

function workbookRow(recordId: string | null): WorkbookRow {
  return {
    key: recordId ?? "draft",
    recordId,
    rowVersion: recordId === null ? null : 1,
    viewSchemaId: "cartulary.view.timeline.v2",
    captureState: "rough",
    values: rowValues,
    committedValues: rowValues,
    collectionValues: {
      hostRefs: [],
      identityRefs: [],
      tags: [],
    },
    collectionDrafts: {
      hostRefs: "",
      identityRefs: "",
      tags: "",
    },
    pendingSignature: null,
    rawRow: null,
  };
}

describe("TimelineEvidencePanel", () => {
  afterEach(() => {
    cleanup();
  });

  it("does not render draft rows as row-bound inspector evidence", () => {
    render(
      <TimelineEvidencePanel
        countDisplay={{ displayCount: "0", stateKey: "empty" }}
        row={workbookRow(null)}
        onFilesSelected={vi.fn()}
      />,
    );

    expect(
      screen.queryByTestId(timelineInspectorSectionTestId("evidence")),
    ).toBeNull();
    expect(
      screen.queryByTestId(timelineDraftEvidenceFileInputTestId()),
    ).toBeNull();
  });

  it("uses row-specific file inputs for committed row evidence", () => {
    render(
      <TimelineEvidencePanel
        countDisplay={{ displayCount: "1", stateKey: "attached" }}
        row={workbookRow("record-1")}
        onFilesSelected={vi.fn()}
      />,
    );

    expect(
      screen.getByTestId(timelineEvidenceFileInputTestId("record-1")),
    ).toBeInstanceOf(HTMLInputElement);
    expect(
      screen.queryByTestId(timelineDraftEvidenceFileInputTestId()),
    ).toBeNull();
  });
});
