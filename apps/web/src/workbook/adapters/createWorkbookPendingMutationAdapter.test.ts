import { afterEach, expect, it, vi } from "vitest";
import {
  successEnvelope,
  timelineRow,
} from "../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  createWorkbookPendingQueueModel,
  type PendingReplayKind,
  type PendingReplayUnitState,
} from "../utils/workbookPendingQueue";
import { createWorkbookPendingMutationAdapter } from "./createWorkbookPendingMutationAdapter";

const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000001";
const changeSetId = "30000000-0000-4000-8000-000000000001";

function pendingUnit(
  kind: PendingReplayKind,
  overrides: { readonly incidentId?: string } = {},
): PendingReplayUnitState {
  const clientTxnId = `txn-${kind}`;
  const unitIncidentId = overrides.incidentId ?? incidentId;
  const queue = createWorkbookPendingQueueModel({
    clientInstanceId: "client-instance",
    incidentId: unitIncidentId,
  });
  const admission = queue.admit({
    clientInstanceId: "client-instance",
    clientTxnId,
    coalesceKey: kind === "create" ? "draft:draft-1" : `record:${recordId}`,
    enqueueOrder: 1,
    id: `pending-${kind}`,
    incidentId: unitIncidentId,
    kind,
    payloadIntent:
      kind === "create"
        ? {
            client_txn_id: "untrusted-payload-txn",
            "timeline.activity_synopsis_text": "Created summary",
          }
        : {
            base_row_version: 2,
            changes: [
              {
                field_key: "timeline.activity_synopsis_text",
                value: "Patched summary",
              },
            ],
            client_txn_id: "untrusted-payload-txn",
            view_schema_id: timelineViewSchemaId,
          },
    recordId: kind === "create" ? null : recordId,
    rowKey: kind === "create" ? "draft-1" : recordId,
    source: "autosave",
    viewSchemaId: timelineViewSchemaId,
  });
  if (!admission.accepted) throw new Error("expected pending admission");
  return admission.unit;
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

it("executes creates through the incident-bound generated operation", async () => {
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

  await expect(
    createWorkbookPendingMutationAdapter({
      apiBase: "/base",
      incidentId,
      recordTiming,
    }).execute({ committedRowVersion: null, unit: pendingUnit("create") }),
  ).resolves.toMatchObject({
    kind: "accepted",
    value: {
      changeSetId,
      row: { record_id: recordId, row_version: 1 },
      viewSchemaId: timelineViewSchemaId,
    },
  });
  expect(fetchMock).toHaveBeenCalledWith(
    `/base/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
    expect.objectContaining({
      body: JSON.stringify({
        "timeline.activity_synopsis_text": "Created summary",
        client_txn_id: "txn-create",
      }),
      method: "POST",
    }),
  );
  expect(recordTiming).toHaveBeenCalledWith("pending_fetch_request", {
    kind: "create",
  });
  expect(recordTiming).toHaveBeenCalledWith("pending_fetch_response", {
    kind: "create",
    status: 200,
  });
  expect(recordTiming).toHaveBeenCalledWith("pending_fetch_json_parsed", {
    kind: "create",
  });
});

it("builds patches from dispatch-time identity and version", async () => {
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
    createWorkbookPendingMutationAdapter({
      apiBase: undefined,
      incidentId,
    }).execute({ committedRowVersion: 7, unit: pendingUnit("patch") }),
  ).resolves.toMatchObject({
    kind: "accepted",
    value: { row: { record_id: recordId, row_version: 8 } },
  });
  expect(JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body))).toEqual({
    base_row_version: 7,
    changes: [
      {
        field_key: "timeline.activity_synopsis_text",
        value: "Patched summary",
      },
    ],
    client_txn_id: "txn-patch",
    view_schema_id: timelineViewSchemaId,
  });
});

it("fails closed for malformed, mismatched, or invalid success responses", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { row: {} } }), {
        headers: { "content-type": "application/json" },
        status: 200,
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        change_set_id: changeSetId,
        row: timelineRow({
          captureState: "rough",
          recordId: "20000000-0000-4000-8000-000000000002",
          rowVersion: 8,
        }),
        view_schema_id: timelineViewSchemaId,
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        change_set_id: changeSetId,
        row: timelineRow({ captureState: "rough", recordId, rowVersion: 0 }),
        view_schema_id: timelineViewSchemaId,
      }),
    )
    .mockResolvedValueOnce(
      successEnvelope({
        change_set_id: "",
        row: timelineRow({ captureState: "rough", recordId, rowVersion: 8 }),
        view_schema_id: timelineViewSchemaId,
      }),
    );
  vi.stubGlobal("fetch", fetchMock);
  const adapter = createWorkbookPendingMutationAdapter({
    apiBase: undefined,
    incidentId,
  });

  for (let index = 0; index < 4; index += 1) {
    await expect(
      adapter.execute({ committedRowVersion: 7, unit: pendingUnit("patch") }),
    ).resolves.toMatchObject({
      failure: { kind: "invalid_contract" },
      kind: "rejected",
    });
  }
});

it("rejects cross-incident and undispatchable units before transport", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  const adapter = createWorkbookPendingMutationAdapter({
    apiBase: undefined,
    incidentId,
  });

  await expect(
    adapter.execute({
      committedRowVersion: 7,
      unit: pendingUnit("patch", {
        incidentId: "10000000-0000-4000-8000-000000000002",
      }),
    }),
  ).resolves.toMatchObject({
    failure: { kind: "invalid_contract" },
    kind: "rejected",
  });
  await expect(
    adapter.execute({ committedRowVersion: null, unit: pendingUnit("patch") }),
  ).resolves.toMatchObject({
    failure: { kind: "stale_target" },
    kind: "rejected",
  });
  expect(fetchMock).not.toHaveBeenCalled();
});
