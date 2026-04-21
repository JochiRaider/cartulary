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
});
