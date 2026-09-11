import { afterEach, expect, it, vi } from "vitest";
import {
  errorResponse,
  jsonResponse,
} from "../../testing/fetchMockTestSupport";
import {
  observationAuthority,
  observationCreateIntent,
  observationSource,
  observationTargetId,
  testObservation,
  testObservationReceipt,
} from "../../testing/observationTestSupport";
import type { ObservationIntent } from "../features/indicators/observationOperation";
import { createObservationTransport } from "./createObservationTransport";

const port = () =>
  createObservationTransport({
    apiBase: "https://original.test",
    incidentId: observationAuthority.incidentId,
  });
const capture = (intent: ObservationIntent = observationCreateIntent) =>
  port().capture(observationAuthority, 1, intent, "secure-original");
afterEach(() => vi.unstubAllGlobals());
it("Observation transport admits only child request members preserves omissions and exact replay bytes", async () => {
  const a = capture(),
    fetch = vi.fn(async () =>
      jsonResponse(
        {
          data: { ...testObservationReceipt, replayed: true },
          meta: { request_id: "replay" },
        },
        200,
      ),
    );
  expect(JSON.parse(a.body)).toEqual({
    client_txn_id: "secure-original",
    base_row_version: 4,
    source_field_key: observationSource.fieldKey,
    span_start_byte: 2,
    span_end_byte: 12,
  });
  vi.stubGlobal("fetch", fetch);
  for (let n = 0; n < 2; n++)
    expect((await port().send(a, new AbortController().signal)).kind).toBe(
      "acknowledged",
    );
  for (const [url, init] of fetch.mock.calls as unknown as [
    string,
    RequestInit,
  ][]) {
    expect(url).toBe(a.path);
    expect(init.body).toBe(a.body);
    expect(init.credentials).toBe("include");
  }
  expect(
    JSON.parse(
      capture({
        action: "resolve",
        observation: testObservation,
        targetId: observationTargetId,
      }).body,
    ),
  ).toEqual({
    client_txn_id: "secure-original",
    base_row_version: 1,
    resolved_indicator_record_id: observationTargetId,
  });
  expect(
    JSON.parse(
      capture({ action: "dismiss", observation: testObservation }).body,
    ),
  ).toEqual({ client_txn_id: "secure-original", base_row_version: 1 });
});
it("Observation receipt validation rejects incomplete inconsistent and malformed success as uncertainty", async () => {
  const a = capture(),
    valid = testObservationReceipt;
  const invalid = [
    { ...valid, change_set_id: "" },
    { ...valid, replayed: true },
    { ...valid, affected_records: [] },
    {
      ...valid,
      affected_records: [...valid.affected_records, ...valid.affected_records],
    },
    {
      ...valid,
      affected_records: [
        { record_id: observationSource.recordId, row_version: 4 },
      ],
    },
    ...[
      { incident_id: observationTargetId },
      { source_record_id: observationTargetId },
      { source_field_key: "other" },
      { row_version: 2 },
      { observed_text: "different" },
      { origin_kind: "inline_parser" },
      { created_by_user_id: observationTargetId },
      { created_at: "yesterday" },
      { origin_locator: "" },
      { resolved_indicator_record_id: observationTargetId },
      { resolved_at: valid.observation.created_at },
    ].map((patch) => ({
      ...valid,
      observation: { ...valid.observation, ...patch },
    })),
  ];
  for (const data of [valid, ...invalid]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({ data, meta: { request_id: "receipt" } }, 201),
      ),
    );
    expect((await port().send(a, new AbortController().signal)).kind).toBe(
      data === valid ? "acknowledged" : "uncertain",
    );
  }
  for (const response of [
    new Response("broken", { status: 201 }),
    new Response("broken", { status: 409 }),
    errorResponse("internal_error", 500),
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => response),
    );
    expect((await port().send(a, new AbortController().signal)).kind).toBe(
      "uncertain",
    );
  }
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => errorResponse("row_version_conflict", 409)),
  );
  expect((await port().send(a, new AbortController().signal)).kind).toBe(
    "rejected",
  );
});
it("Observation transition receipts preserve historical provenance and require source old and new targets", async () => {
  const oldTarget = "50000000-0000-4000-8000-000000000002";
  const previous = {
    ...testObservation,
    created_at: "2026-09-11T08:00:00-04:00",
    row_version: 7,
    resolution_status: "resolved" as const,
    resolved_indicator_record_id: oldTarget,
    resolved_by_user_id: observationAuthority.actorId,
    resolved_at: testObservation.created_at,
    resolution_method: "indicators.observations.resolve",
    origin_locator: "opaque parser locator",
  };
  const intents: ObservationIntent[] = [
    { action: "resolve", observation: previous, targetId: observationTargetId },
    { action: "dismiss", observation: previous },
    {
      action: "restore",
      observation: {
        ...previous,
        resolution_status: "dismissed",
        resolved_indicator_record_id: null,
      },
    },
  ];
  for (const intent of intents) {
    const a = capture(intent),
      restored = intent.action === "restore",
      resolved = intent.action === "resolve";
    const observation = {
      ...previous,
      created_at: "2026-09-11T12:00:00Z",
      row_version: 8,
      resolution_status: restored
        ? "unresolved"
        : resolved
          ? "resolved"
          : "dismissed",
      resolved_indicator_record_id: resolved ? observationTargetId : null,
      resolved_by_user_id: restored ? null : observationAuthority.actorId,
      resolved_at: restored ? null : previous.created_at,
      resolution_method: restored
        ? null
        : `indicators.observations.${intent.action}`,
    };
    const affected_records = [
      { record_id: observationSource.recordId, row_version: 23 },
      ...(resolved
        ? [{ record_id: observationTargetId, row_version: 99 }]
        : []),
      ...(!restored ? [{ record_id: oldTarget, row_version: 3 }] : []),
    ];
    const data = { ...testObservationReceipt, observation, affected_records };
    for (const receipt of [
      data,
      { ...data, observation: { ...observation, observed_text: "rewritten" } },
      { ...data, affected_records: affected_records.slice(1) },
    ]) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () =>
          jsonResponse(
            { data: receipt, meta: { request_id: "transition" } },
            200,
          ),
        ),
      );
      expect((await port().send(a, new AbortController().signal)).kind).toBe(
        receipt === data ? "acknowledged" : "uncertain",
      );
    }
  }
});
