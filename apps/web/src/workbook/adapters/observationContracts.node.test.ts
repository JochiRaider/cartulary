import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import { jsonResponse } from "../../testing/fetchMockTestSupport";
import {
  observationAuthority,
  observationCreateIntent,
  observationTargetId,
  testObservation,
} from "../../testing/observationTestSupport";
import { resolveIndicatorInspectorHandler } from "../features/indicators/indicatorInspectorHandlers";
import { createObservationReader } from "./createObservationReader";
import { createObservationTransport } from "./createObservationTransport";

afterEach(() => vi.unstubAllGlobals());
describe("Observation public contracts", () => {
  it("uses the typed source-span operation without provenance fields", () => {
    const port = createObservationTransport({
      apiBase: undefined,
      incidentId: observationAuthority.incidentId,
    });
    const a = port.capture(
      observationAuthority,
      1,
      { ...observationCreateIntent, parsedType: "domain_name" },
      "original-id",
    );
    expect(a.operation).toBe("createManualIndicatorObservation");
    expect(JSON.parse(a.body)).toEqual({
      client_txn_id: "original-id",
      base_row_version: 4,
      source_field_key: "timeline.raw_activity_text",
      span_start_byte: 2,
      span_end_byte: 12,
      parsed_indicator_type: "domain_name",
    });
  });
  it("binds observation transitions to typed child operations", () => {
    const port = createObservationTransport({
      apiBase: undefined,
      incidentId: observationAuthority.incidentId,
    });
    for (const action of ["resolve", "dismiss", "restore"] as const) {
      const a = port.capture(
        observationAuthority,
        1,
        action === "resolve"
          ? {
              action,
              observation: testObservation,
              targetId: observationTargetId,
            }
          : { action, observation: testObservation },
        "original-id",
      );
      expect(a.operation).toBe(`${action}IndicatorObservation`);
      expect(JSON.parse(a.body)).toEqual({
        client_txn_id: "original-id",
        base_row_version: 1,
        ...(action === "resolve"
          ? { resolved_indicator_record_id: observationTargetId }
          : {}),
      });
    }
  });
  it("preserves paging metadata for load-more requests", async () => {
    const fetch = vi.fn(async (_input: RequestInfo | URL) =>
      jsonResponse({
        data: { observations: [testObservation] },
        meta: {
          request_id: "page",
          paging: { limit: 100, has_more: true, next_cursor: "next" },
        },
      }),
    );
    vi.stubGlobal("fetch", fetch);
    const result = await createObservationReader({
      apiBase: undefined,
      incidentId: observationAuthority.incidentId,
    }).observations(
      { kind: "source", recordId: testObservation.source_record_id },
      "opaque",
      new AbortController().signal,
    );
    expect(result).toEqual({
      kind: "accepted",
      value: { items: [testObservation], hasMore: true, nextCursor: "next" },
    });
    const url = String(fetch.mock.calls[0]?.[0]);
    expect(url).toContain("cursor_token=opaque");
    expect(url).toContain("limit=100");
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
