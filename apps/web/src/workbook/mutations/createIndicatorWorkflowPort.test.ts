import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it, vi } from "vitest";
import type { WorkbookOperationExecutor } from "../adapters/workbookOperationContract";
import { resolveIndicatorInspectorHandler } from "../features/indicators/indicatorInspectorHandlers";
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

const mutationEnvelope = {
  data: {
    affected_records: [
      { record_id: observation.source_record_id, row_version: 6 },
    ],
    change_set_id: "60000000-0000-4000-8000-000000000001",
    observation,
    replayed: false,
  },
  meta: { request_id: "request-indicator-mutation" },
};

function executorReturning(value: unknown) {
  const execute = vi.fn(async (_input: unknown) => ({
    kind: "accepted",
    value,
  }));
  return {
    execute,
    operations: { execute } as unknown as WorkbookOperationExecutor,
  };
}

describe("createIndicatorWorkflowPort", () => {
  it("uses the typed source-span operation without provenance fields", async () => {
    const { execute, operations } = executorReturning(mutationEnvelope);
    const port = createIndicatorWorkflowPort({
      createMutationID: () => "client-observation-1",
      operations,
    });

    const outcome = await port.createManualObservation({
      baseRowVersion: 5,
      parsedIndicatorType: "ipv4_addr",
      sourceFieldKey: "timeline.raw_activity_text",
      sourceRecordId: observation.source_record_id,
      spanStartByte: 0,
      spanEndByte: 7,
    });

    expect(outcome).toEqual({
      kind: "accepted",
      value: {
        affectedRecords: mutationEnvelope.data.affected_records,
        changeSetId: mutationEnvelope.data.change_set_id,
        replayed: false,
        resource: observation,
      },
    });
    expect(execute).toHaveBeenCalledWith({
      operationID: "createManualIndicatorObservation",
      pathParameters: { source_record_id: observation.source_record_id },
      request: {
        client_txn_id: "client-observation-1",
        base_row_version: 5,
        source_field_key: "timeline.raw_activity_text",
        span_start_byte: 0,
        span_end_byte: 7,
        parsed_indicator_type: "ipv4_addr",
      },
    });
    const call = execute.mock.calls[0]?.[0];
    if (call === undefined || typeof call !== "object" || call === null) {
      throw new Error("Expected a typed Indicator operation request");
    }
    const request = (call as { readonly request: unknown }).request;
    expect(request).not.toHaveProperty("origin_kind");
    expect(request).not.toHaveProperty("observed_text");
    expect(request).not.toHaveProperty("origin_locator");
  });

  it("binds observation transitions to typed child operations", async () => {
    const { execute, operations } = executorReturning(mutationEnvelope);
    const port = createIndicatorWorkflowPort({
      createMutationID: () => "client-transition-1",
      operations,
    });

    await port.transitionObservation({
      action: "resolve",
      baseRowVersion: 2,
      observationId: observation.observation_id,
      resolvedIndicatorRecordId: "50000000-0000-4000-8000-000000000001",
    });

    expect(execute).toHaveBeenCalledWith({
      operationID: "resolveIndicatorObservation",
      pathParameters: { observation_id: observation.observation_id },
      request: {
        client_txn_id: "client-transition-1",
        base_row_version: 2,
        resolved_indicator_record_id: "50000000-0000-4000-8000-000000000001",
      },
    });
  });

  it("preserves paging metadata for load-more requests", async () => {
    const envelope = {
      data: { observations: [observation] },
      meta: {
        request_id: "request-indicator-list",
        paging: { has_more: true, limit: 25, next_cursor: "cursor-next" },
      },
    };
    const { execute, operations } = executorReturning(envelope);
    const port = createIndicatorWorkflowPort({
      createMutationID: () => "unused",
      operations,
    });

    const outcome = await port.listObservations({
      cursorToken: "cursor-current",
      indicatorRecordId: "50000000-0000-4000-8000-000000000001",
      limit: 25,
    });

    expect(outcome).toEqual({
      kind: "accepted",
      value: {
        items: [observation],
        paging: envelope.meta.paging,
      },
    });
    expect(execute).toHaveBeenCalledWith({
      operationID: "listIndicatorObservations",
      pathParameters: {
        indicator_id: "50000000-0000-4000-8000-000000000001",
      },
      query: { cursor_token: "cursor-current", limit: 25 },
    });
  });

  it("resolves all four Indicator handlers by their complete semantic tuple", () => {
    const expected = [
      [
        "cartulary.view.timeline.v2",
        "indicator.observations.manage",
        "relationships",
      ],
      [
        "cartulary.view.indicators.v1",
        "indicator.observations.pivot",
        "relationships",
      ],
      ["cartulary.view.indicators.v1", "indicator.lifecycle.read", "history"],
      ["cartulary.view.indicators.v1", "indicator.lifecycle.manage", "history"],
    ] as const;

    for (const [viewSchemaId, action, panelId] of expected) {
      const featureGroup = requireViewContract(
        viewSchemaId,
      ).inspectorConfig.featureGroups.find(
        (candidate) => candidate.featureGroupKey === action,
      );
      if (featureGroup === undefined) {
        throw new Error(`Missing characterized feature group ${action}`);
      }
      expect(
        resolveIndicatorInspectorHandler(viewSchemaId, featureGroup),
      ).toEqual({ action, panelId });
    }
  });

  it("omits mismatched and generic-patch Indicator tuples", () => {
    const featureGroup = requireViewContract(
      "cartulary.view.indicators.v1",
    ).inspectorConfig.featureGroups.find(
      (candidate) => candidate.featureGroupKey === "indicator.lifecycle.manage",
    );
    if (featureGroup === undefined) {
      throw new Error("Missing characterized Indicator lifecycle feature");
    }
    expect(
      resolveIndicatorInspectorHandler("cartulary.view.timeline.v2", {
        ...featureGroup,
        routeBinding: {
          ...featureGroup.routeBinding,
          kind: "record_patch",
          owner: "record_patch_route",
        },
      }),
    ).toBeNull();
    expect(
      resolveIndicatorInspectorHandler("cartulary.view.indicators.v1", {
        ...featureGroup,
        panelId: "relationships",
      }),
    ).toBeNull();
  });
});
