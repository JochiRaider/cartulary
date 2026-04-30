import { describe, expect, it } from "vitest";

import {
  fieldCapability,
  requireViewContract,
  resolveHeaderSortFieldKey,
  visibleFields,
} from "./index";

describe("view-contracts", () => {
  it("parses sortable, filterable, and zero-field create metadata", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v1");

    expect(timeline.permitsZeroFieldCreate).toBe(true);
    expect(timeline.sortFields).toContain("timeline.sort_ts");
    expect(timeline.filterFields).toContain("timeline.capture_state");
    expect(timeline.groupingFields).toContain("timeline.capture_state");
  });

  it("resolves header sort keys and field capabilities from contract metadata", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v1");

    expect(resolveHeaderSortFieldKey(timeline, "timeline.occurred_at")).toBe(
      "timeline.sort_ts",
    );
    expect(fieldCapability(timeline, "timeline.occurred_at")).toEqual({
      editable: true,
      filterable: false,
      groupable: false,
      sortable: true,
    });
  });

  it("returns the contract-default visible field order", () => {
    const hosts = requireViewContract("cartulary.view.hosts.v1");
    expect(visibleFields(hosts).map((field) => field.fieldKey)).toEqual(
      hosts.defaultVisibleFields,
    );
  });

  it("parses enum values for contract-backed controls", () => {
    const assessments = requireViewContract("cartulary.view.assessments.v1");

    expect(
      assessments.fieldMap["assessment.assessment_state"]?.enumValues,
    ).toEqual(["unknown", "suspected", "confirmed", "disproven", "cleared"]);
    expect(
      assessments.fieldMap["assessment.confidence_band"]?.enumValues,
    ).toEqual(["unset", "low", "medium", "high"]);
  });

  it("exposes mutation metadata needed by workbook controls", () => {
    const evidence = requireViewContract("cartulary.view.evidence.v1");

    expect(evidence.fieldMap["evidence.title"]?.stringContractId).toBe(
      "single_line_title_v1",
    );
    expect(
      evidence.fieldMap["evidence.requested_at"]?.directScalarContractId,
    ).toBe("timestamp_instant_v1");
    expect(
      evidence.fieldMap["evidence.collector_party_id"]
        ?.directReferenceContractId,
    ).toBe("same_incident_party_ref_v1");
    expect(evidence.fieldMap["evidence.lifecycle_state"]?.enumValues).toEqual([
      "requested",
      "pending_receipt",
      "received",
      "available",
      "quarantined",
      "released",
    ]);
  });
});
