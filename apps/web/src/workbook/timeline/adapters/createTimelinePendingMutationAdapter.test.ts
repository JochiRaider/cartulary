import { afterEach, expect, it, vi } from "vitest";
import {
  successEnvelope,
  timelineRow,
} from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  type PendingReplayKind,
  type PendingReplayUnitState,
  WorkbookPendingQueueModel,
} from "../../utils/workbookPendingQueue";
import { createTimelinePendingMutationAdapter } from "./createTimelinePendingMutationAdapter";

const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000001";
const changeSetId = "30000000-0000-4000-8000-000000000001";

function pendingUnit(kind: PendingReplayKind): PendingReplayUnitState {
  const clientTxnId = `txn-${kind}`;
  const queue = new WorkbookPendingQueueModel({
    incidentId,
    clientInstanceId: "client-instance",
  });
  const admission = queue.admit({
    id: `pending-${kind}`,
    kind,
    source: "autosave",
    incidentId,
    clientInstanceId: "client-instance",
    viewSchemaId: timelineViewSchemaId,
    rowKey: kind === "create" ? "draft-1" : recordId,
    recordId: kind === "create" ? null : recordId,
    payloadIntent:
      kind === "create"
        ? {
            client_txn_id: clientTxnId,
            "timeline.activity_synopsis_text": "Created summary",
          }
        : {
            view_schema_id: timelineViewSchemaId,
            base_row_version: 2,
            client_txn_id: clientTxnId,
            changes: [
              {
                field_key: "timeline.activity_synopsis_text",
                value: "Patched summary",
              },
            ],
          },
    clientTxnId,
    coalesceKey: kind === "create" ? "draft:draft-1" : `record:${recordId}`,
    enqueueOrder: 1,
  });
  if (!admission.accepted) throw new Error("expected pending admission");
  return admission.unit;
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

it("materializes fresh and replayed Timeline creates through the generated binding", async () => {
  const fetchMock = vi.fn().mockResolvedValueOnce(
    successEnvelope({
      change_set_id: changeSetId,
      row: timelineRow({
        captureState: "rough",
        recordId,
        rowVersion: 1,
        summary: "Created summary",
      }),
      view_schema_id: timelineViewSchemaId,
    }),
  );
  vi.stubGlobal("fetch", fetchMock);
  const recordTiming = vi.fn();

  const outcome = await createTimelinePendingMutationAdapter({
    apiBase: "/base",
    recordTiming,
  }).execute({ committedRowVersion: null, unit: pendingUnit("create") });

  expect(outcome).toMatchObject({
    kind: "accepted",
    value: { changeSetId, row: { record_id: recordId, row_version: 1 } },
  });
  expect(fetchMock).toHaveBeenCalledWith(
    `/base/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
    expect.objectContaining({
      method: "POST",
      body: JSON.stringify({
        "timeline.activity_synopsis_text": "Created summary",
        client_txn_id: "txn-create",
      }),
    }),
  );
  expect(recordTiming).toHaveBeenCalledWith(
    "pending_fetch_response",
    expect.objectContaining({ clientTxnId: "txn-create", status: 200 }),
  );
  expect(recordTiming).toHaveBeenCalledWith(
    "pending_fetch_json_parsed",
    expect.objectContaining({ clientTxnId: "txn-create" }),
  );
});

it("materializes Timeline patches with the dispatch-time committed row version", async () => {
  const fetchMock = vi.fn().mockResolvedValueOnce(
    successEnvelope({
      change_set_id: changeSetId,
      row: timelineRow({
        captureState: "enriched",
        recordId,
        rowVersion: 8,
        summary: "Patched summary",
      }),
      view_schema_id: timelineViewSchemaId,
    }),
  );
  vi.stubGlobal("fetch", fetchMock);

  await expect(
    createTimelinePendingMutationAdapter({ apiBase: undefined }).execute({
      committedRowVersion: 7,
      unit: pendingUnit("patch"),
    }),
  ).resolves.toMatchObject({
    kind: "accepted",
    value: { row: { record_id: recordId, row_version: 8 } },
  });
  expect(JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body))).toEqual({
    view_schema_id: timelineViewSchemaId,
    base_row_version: 7,
    client_txn_id: "txn-patch",
    changes: [
      {
        field_key: "timeline.activity_synopsis_text",
        value: "Patched summary",
      },
    ],
  });
});

it("fails closed for malformed or target-inconsistent replay success", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { row: {} } }), {
        status: 200,
        headers: { "content-type": "application/json" },
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        change_set_id: changeSetId,
        row: timelineRow({
          captureState: "enriched",
          recordId: "20000000-0000-4000-8000-000000000002",
          rowVersion: 8,
        }),
        view_schema_id: timelineViewSchemaId,
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        change_set_id: changeSetId,
        row: {
          record_id: recordId,
          row_version: 8,
          cells: {},
        },
        view_schema_id: timelineViewSchemaId,
      }),
    );
  vi.stubGlobal("fetch", fetchMock);
  const adapter = createTimelinePendingMutationAdapter({ apiBase: undefined });

  await expect(
    adapter.execute({ committedRowVersion: 7, unit: pendingUnit("patch") }),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { kind: "invalid_contract" },
  });
  await expect(
    adapter.execute({ committedRowVersion: 7, unit: pendingUnit("patch") }),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { kind: "invalid_contract" },
  });
  await expect(
    adapter.execute({ committedRowVersion: 7, unit: pendingUnit("patch") }),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { kind: "invalid_contract" },
  });
});
