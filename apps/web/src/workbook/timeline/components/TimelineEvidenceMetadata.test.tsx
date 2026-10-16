import {
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import type { WorkbookEvidenceMetadataPort } from "../../evidence/WorkbookEvidenceMetadataPort";
import type { EvidenceHandleOutcome } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookReadScope } from "../../query/WorkbookQueryRow";
import { createTimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import { useTimelineEvidenceMetadata } from "../hooks/useTimelineEvidenceMetadata";
import { useTimelineLinkedEvidenceReview } from "../hooks/useTimelineLinkedEvidenceReview";
import { createDraftRow } from "../models/timelineRowModel";
import { TimelineEvidencePanel } from "./TimelineEvidencePanel";

const scope: WorkbookReadScope = {
  actorId: "actor",
  sessionIdentity: "session",
  incidentId: "incident",
  epoch: 1,
};
const ids = Array.from({ length: 12 }, (_, i) => `evidence-${i}`);
function row() {
  return {
    ...createDraftRow(0),
    key: "source",
    recordId: "source",
    rowVersion: 1,
    rawRow: {
      record_id: "source",
      row_version: 1,
      view_schema_id: "cartulary.view.timeline.v2",
      observation: { recordId: "source", rowVersion: 1, scope },
      cells: {
        "timeline.attached_evidence_ids": {
          value: {
            items: ids.map((id) => ({
              item_ref: `link-${id}`,
              item_kind: "record_ref",
              display_text: id,
              linked_record_id: id,
            })),
          },
        },
      },
    },
  };
}
afterEach(cleanup);
it("renders bounded linked Evidence metadata and distinct preview download eligibility without issuing handles", async () => {
  const source = row();
  const rowsRef = { current: [source] };
  const issueHandle = vi.fn(
    async (): Promise<EvidenceHandleOutcome> => ({
      kind: "rejected",
      failure: { kind: "terminal", message: "unused" },
    }),
  );
  const port = { issueHandle };
  const metadataPort: WorkbookEvidenceMetadataPort = {
    query: vi.fn<WorkbookEvidenceMetadataPort["query"]>(
      async ({ recordIds }) => ({
        kind: "accepted",
        items: recordIds.map((recordId) => ({
          recordId,
          rowVersion: 7,
          title: `Current ${recordId}`,
          hasBlob: true,
          uploadState: "available",
          lifecycleState: "available",
          previewEligibility:
            recordId === ids[0]
              ? { eligible: false, reasonCode: "unsupported_preview" }
              : recordId === ids[1]
                ? { eligible: false, reasonCode: "blob_missing" }
                : { eligible: true, reasonCode: null },
        })),
      }),
    ),
  };
  const registry = createTimelineInspectorElementRegistry({
    lifecycleKey: "incident",
    reviewGeneration: 1,
    subject: null,
  });
  function Panel() {
    const review = useTimelineLinkedEvidenceReview({
      port,
      rowsRef,
      selectedRowId: "source",
      readScope: scope,
      canRead: true,
      inspectorOpen: true,
      scopeKey: "scope",
      elementRegistry: registry,
      onAccessFailure: vi.fn(),
      onRestoreGridFocus: vi.fn(),
    });
    return (
      <TimelineEvidencePanel
        countDisplay={{ displayCount: "10", stateKey: "available" }}
        row={source}
        review={{
          ...review,
          metadata: useTimelineEvidenceMetadata({
            port: metadataPort,
            row: source,
            readScope: scope,
            canRead: true,
            inspectorOpen: true,
            scopeKey: "scope",
            onAccessFailure: () => {},
          }),
        }}
        onFilesSelected={vi.fn()}
      />
    );
  }
  render(<Panel />);
  await screen.findByText("Current evidence-0");
  const list = screen.getByRole("list", { name: "Linked Evidence" });
  expect(within(list).getAllByRole("listitem")).toHaveLength(5);
  expect(screen.getByText("Linked evidence: 12")).toBeTruthy();
  expect(screen.getByText("Attached evidence count: 10")).toBeTruthy();
  expect(
    screen
      .getByRole("button", { name: "Preview Current evidence-0" })
      .getAttribute("aria-disabled"),
  ).toBe("true");
  expect(
    screen
      .getByRole("button", { name: "Download Current evidence-0" })
      .getAttribute("aria-disabled"),
  ).toBe("false");
  expect(
    screen
      .getByRole("button", { name: "Download Current evidence-1" })
      .getAttribute("aria-disabled"),
  ).toBe("true");
  expect(
    screen
      .getByRole("button", { name: "Preview Current evidence-2" })
      .getAttribute("aria-disabled"),
  ).toBe("false");
  const next = screen.getByRole("button", { name: "Next evidence" });
  next.focus();
  fireEvent.click(next);
  await screen.findByText("Current evidence-5");
  expect(document.activeElement).toBe(next);
  expect(screen.queryByText("Current evidence-0")).toBeNull();
  expect(metadataPort.query).toHaveBeenCalledTimes(2);
  expect(issueHandle).not.toHaveBeenCalled();
});
