import {
  timelineDraftEvidenceFileInputTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { taskAuthority } from "../../../testing/taskWorkbookTestSupport";
import { TimelineFileContext } from "../../features/evidence/EvidenceAttachmentContext";
import { WorkbookTimelineFileOwner } from "../../features/evidence/WorkbookTimelineFileOwner";
import type { WorkbookRow } from "../models/timelineRowModel";
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
    const owner = new WorkbookTimelineFileOwner(
      taskAuthority.incidentId,
      { create: () => "unused" },
      {
        coordinate: async () => ({ kind: "settled", minimumRowVersion: 0 }),
        accepted: vi.fn(),
        refresh: async () => {},
      },
    );
    owner.setAuthority(taskAuthority);
    const selected = vi.fn();
    render(
      <TimelineFileContext value={owner}>
        <TimelineEvidencePanel
          countDisplay={{ displayCount: "1", stateKey: "available" }}
          row={workbookRow("record-1")}
          onFilesSelected={selected}
        />
      </TimelineFileContext>,
    );

    expect(
      screen.getByTestId(timelineEvidenceFileInputTestId("record-1")),
    ).toBeInstanceOf(HTMLInputElement);
    expect(
      screen.queryByTestId(timelineDraftEvidenceFileInputTestId()),
    ).toBeNull();
    const region = screen.getByRole("region", {
      name: "File attachment for this Timeline record",
    });
    const button = screen.getByRole("button", {
      name: "Attach file to this Timeline record",
    });
    const file = new File(["content"], "evidence.txt", { type: "text/plain" });
    fireEvent.click(button);
    fireEvent.change(
      screen.getByTestId(timelineEvidenceFileInputTestId("record-1")),
      { target: { files: [file] } },
    );
    fireEvent.drop(region, { dataTransfer: { files: [file] } });
    button.focus();
    fireEvent.paste(button, { clipboardData: { files: [file] } });
    expect(selected).toHaveBeenCalledTimes(3);
    for (const call of selected.mock.calls) expect(call[1]).toEqual([file]);
    act(() => owner.setAuthority({ ...taskAuthority, role: "viewer" }));
    fireEvent.drop(region, { dataTransfer: { files: [file] } });
    fireEvent.paste(button, { clipboardData: { files: [file] } });
    expect(selected).toHaveBeenCalledTimes(3);
    expect((button as HTMLButtonElement).disabled).toBe(true);
    expect(
      screen.getByRole("region", { name: "Evidence information" }).textContent,
    ).toContain("Attached evidence count: 1");
  });
});
