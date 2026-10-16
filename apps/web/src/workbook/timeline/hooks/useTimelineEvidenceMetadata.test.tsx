import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import type {
  EvidenceMetadataResult,
  WorkbookEvidenceMetadataPort,
} from "../../evidence/WorkbookEvidenceMetadataPort";
import type { WorkbookReadScope } from "../../query/WorkbookQueryRow";
import { createDraftRow, type WorkbookRow } from "../models/timelineRowModel";
import { useTimelineEvidenceMetadata } from "./useTimelineEvidenceMetadata";

const scope: WorkbookReadScope = {
  actorId: "actor",
  sessionIdentity: "session",
  incidentId: "incident",
  epoch: 1,
};
const ids = Array.from({ length: 12 }, (_, i) => `evidence-${i}`);
function row(version = 1): WorkbookRow {
  return {
    ...createDraftRow(0),
    key: "source",
    recordId: "source",
    rowVersion: version,
    rawRow: {
      record_id: "source",
      row_version: version,
      view_schema_id: "cartulary.view.timeline.v2",
      observation: { recordId: "source", rowVersion: version, scope },
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
const accepted = (recordIds: readonly string[]): EvidenceMetadataResult => ({
  kind: "accepted",
  items: recordIds.map((recordId) => ({
    recordId,
    rowVersion: 1,
    title: recordId,
    hasBlob: false,
    uploadState: null,
    lifecycleState: "requested",
    previewEligibility: { eligible: false, reasonCode: "no_visible_blob" },
  })),
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

it("reads only the current Evidence window and suppresses late page source and authority outcomes", async () => {
  const pending: {
    ids: readonly string[];
    signal: AbortSignal;
    settle: (value: EvidenceMetadataResult) => void;
  }[] = [];
  const query = vi.fn<WorkbookEvidenceMetadataPort["query"]>(
    ({ recordIds, signal }) =>
      new Promise<EvidenceMetadataResult>((settle) =>
        pending.push({ ids: recordIds, signal, settle }),
      ),
  );
  const props = {
    port: { query },
    row: row(),
    readScope: scope,
    canRead: true,
    inspectorOpen: true,
    scopeKey: "scope",
    onAccessFailure: vi.fn(),
  };
  const view = renderHook(useTimelineEvidenceMetadata, { initialProps: props });
  expect(pending[0]?.ids).toEqual(ids.slice(0, 5));
  act(() => view.result.current.next?.());
  expect(pending[0]?.signal.aborted).toBe(true);
  expect(pending[1]?.ids).toEqual(ids.slice(5, 10));
  await act(async () => {
    pending[0]?.settle(accepted(ids.slice(0, 5)));
  });
  expect(view.result.current.kind).toBe("loading");
  await act(async () => {
    pending[1]?.settle(accepted(ids.slice(5, 10)));
  });
  expect(view.result.current.kind).toBe("ready");
  expect(view.result.current.items.map((item) => item.recordId)).toEqual(
    ids.slice(5, 10),
  );
  view.rerender({ ...props, inspectorOpen: false });
  expect(view.result.current.items).toEqual([]);
  view.rerender(props);
  expect(view.result.current.kind).toBe("loading");
  expect(view.result.current.items).toEqual([]);
  await act(async () => {
    pending[2]?.settle(accepted(ids.slice(5, 10)));
  });
  act(() => view.result.current.next?.());
  expect(pending[3]?.ids).toEqual(ids.slice(10));
  view.rerender({ ...props, row: row(2) });
  expect(view.result.current.start).toBe(0);
  expect(view.result.current.items).toEqual([]);
  await act(async () => {
    pending[3]?.settle(accepted(ids.slice(10)));
  });
  expect(view.result.current.kind).toBe("loading");
  view.rerender({ ...props, canRead: false });
  await act(async () => {
    pending[4]?.settle(accepted(ids.slice(0, 5)));
  });
  expect(view.result.current.kind).toBe("unavailable");
  expect(view.result.current.items).toEqual([]);
  expect(view.result.current.links).toEqual([]);
  expect(query).toHaveBeenCalledTimes(5);
});

it("requires explicit metadata retry and forwards current authority loss", async () => {
  const query = vi
    .fn<WorkbookEvidenceMetadataPort["query"]>()
    .mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "unavailable" },
    });
  const onAccessFailure = vi.fn();
  const props = {
    port: { query },
    row: row(),
    readScope: scope,
    canRead: true,
    inspectorOpen: true,
    scopeKey: "scope",
    onAccessFailure,
  };
  const view = renderHook(useTimelineEvidenceMetadata, { initialProps: props });
  await waitFor(() => expect(view.result.current.kind).toBe("failed"));
  view.rerender(props);
  expect(query).toHaveBeenCalledTimes(1);
  query.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "authorization_lost", message: "denied" },
  });
  act(() => view.result.current.retry());
  await waitFor(() => expect(onAccessFailure).toHaveBeenCalledTimes(1));
  expect(view.result.current.items).toEqual([]);
});
