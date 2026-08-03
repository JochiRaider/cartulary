import { afterEach, describe, expect, it, vi } from "vitest";
import { createIndicatorWorkflowPort } from "./createIndicatorWorkflowPort";
import type { IndicatorObservation } from "./workbookMutationCommandPorts";

const observation = {
  observation_id: "30000000-0000-4000-8000-000000000001",
  incident_id: "10000000-0000-4000-8000-000000000001",
  source_record_id: "20000000-0000-4000-8000-000000000001",
  source_field_key: "timeline.raw_activity_text",
  origin_kind: "manual_entry",
  origin_locator: "bytes:0-7",
  observed_text: "1.2.3.4",
  parsed_indicator_type: "ipv4_addr",
  normalized_candidate: "1.2.3.4",
  resolution_status: "unresolved",
  resolved_indicator_record_id: null,
  row_version: 1,
  created_by_user_id: "40000000-0000-4000-8000-000000000001",
  created_at: "2026-08-03T12:00:00Z",
  resolved_by_user_id: null,
  resolved_at: null,
  resolution_method: null,
} as const satisfies IndicatorObservation;

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("createIndicatorWorkflowPort", () => {
  it("uses the source-span route without exposing provenance fields", async () => {
    const requests: Array<{ readonly path: string; readonly body: unknown }> =
      [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        requests.push({
          path: String(input),
          body: init?.body ? JSON.parse(String(init.body)) : null,
        });
        return new Response(
          JSON.stringify({ data: { observation }, meta: {} }),
          { status: 201, headers: { "Content-Type": "application/json" } },
        );
      }),
    );
    const port = createIndicatorWorkflowPort({
      apiBase: "https://cartulary.test",
      createMutationID: () => "client-observation-1",
    });

    const outcome = await port.createManualObservation({
      baseRowVersion: 5,
      parsedIndicatorType: "ipv4_addr",
      sourceFieldKey: "timeline.raw_activity_text",
      sourceRecordId: observation.source_record_id,
      spanStartByte: 0,
      spanEndByte: 7,
    });

    expect(outcome).toEqual({ kind: "accepted", value: observation });
    expect(requests).toEqual([
      {
        path: `https://cartulary.test/api/v1/records/${observation.source_record_id}/indicator-observations`,
        body: {
          client_txn_id: "client-observation-1",
          base_row_version: 5,
          source_field_key: "timeline.raw_activity_text",
          span_start_byte: 0,
          span_end_byte: 7,
          parsed_indicator_type: "ipv4_addr",
        },
      },
    ]);
    expect(requests[0]?.body).not.toHaveProperty("origin_kind");
    expect(requests[0]?.body).not.toHaveProperty("observed_text");
    expect(requests[0]?.body).not.toHaveProperty("origin_locator");
  });

  it("binds observation transition actions to the child row version", async () => {
    const request = vi.fn(
      async (_input: RequestInfo | URL, _init?: RequestInit) =>
        new Response(JSON.stringify({ data: { observation }, meta: {} }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    );
    vi.stubGlobal("fetch", request);
    const port = createIndicatorWorkflowPort({
      apiBase: undefined,
      createMutationID: () => "client-transition-1",
    });

    await port.transitionObservation({
      action: "resolve",
      baseRowVersion: 2,
      observationId: observation.observation_id,
      resolvedIndicatorRecordId: "50000000-0000-4000-8000-000000000001",
    });

    expect(request).toHaveBeenCalledOnce();
    expect(request.mock.calls[0]?.[0]).toBe(
      `/api/v1/indicator-observations/${observation.observation_id}/resolve`,
    );
    expect(JSON.parse(String(request.mock.calls[0]?.[1]?.body))).toEqual({
      client_txn_id: "client-transition-1",
      base_row_version: 2,
      resolved_indicator_record_id: "50000000-0000-4000-8000-000000000001",
    });
  });
});
