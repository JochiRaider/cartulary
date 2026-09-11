import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, expect, it, vi } from "vitest";
import { timelineCaptureReview } from "../../testing/timelineCaptureActionTestSupport";
import { successEnvelope } from "../../testing/timelineWorkbookTestSupport";
import { inspectorContextualCapabilities } from "../inspector/inspectorCapabilityResolver";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { createTimelineRecordActionAdapter } from "./adapters/createTimelineRecordActionAdapter";

afterEach(() => vi.unstubAllGlobals());

it("Timeline declares both capture actions in History with distinct confirmation policies", () => {
  const capabilities = inspectorContextualCapabilities({
    config: requireViewContract(timelineViewSchemaId).inspectorConfig,
    panelId: "history",
  });
  expect(
    capabilities.map(({ featureGroup: group }) => [
      group.featureGroupKey,
      group.requiresConfirmation,
    ]),
  ).toEqual([
    ["timeline.mark_reviewed", false],
    ["timeline.supersede", true],
  ]);
});

it("Timeline supersession preserves authored reason and omits an absent replacement", async () => {
  const input = {
    action: "supersede" as const,
    baseRowVersion: 3,
    clientTxnId: "txn-characterize-supersede",
    recordId: "20000000-0000-4000-8000-000000000001",
    replacementRecordId: null,
    reason: "Duplicate source entry",
  };
  const fetch = vi.fn(async () =>
    successEnvelope({
      record_id: input.recordId,
      incident_id: "10000000-0000-4000-8000-000000000001",
      row_version: 4,
      capture_state: "superseded",
      change_set_id: "30000000-0000-4000-8000-000000000001",
      reason: input.reason,
      replacement_record_id: null,
    }),
  );
  vi.stubGlobal("fetch", fetch);
  const adapter = createTimelineRecordActionAdapter({ apiBase: undefined });
  const base = timelineCaptureReview();
  const attempt = adapter.capture(
    timelineCaptureReview({
      action: "supersede",
      target: { ...base.target, rowVersion: 3 },
      reason: input.reason,
    }),
    input.clientTxnId,
  );
  await expect(
    adapter.send(attempt, new AbortController().signal),
  ).resolves.toMatchObject({
    kind: "acknowledged",
    receipt: { operation: "supersede", data: { replacement_record_id: null } },
  });
  expect(
    JSON.parse(
      String(
        (fetch.mock.calls as unknown as [string, RequestInit][])[0]?.[1].body,
      ),
    ),
  ).toEqual({
    base_row_version: 3,
    client_txn_id: input.clientTxnId,
    reason: input.reason,
  });
});
