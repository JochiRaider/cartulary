import { describe, expect, it } from "vitest";
import type { NetworkFlowFilter } from "../services/networkFlowContractAdapter";
import {
  decodeNetworkFlowFilter,
  networkFlowPresentationMetadata,
} from "../services/networkFlowContractAdapter";
import {
  acceptedInitialRequest,
  canonicalFilterKey,
  compileAcceptedDraft,
  compilePredicate,
  compileRejectedDraft,
  graphInitialRequest,
  type QueryPredicateDraft,
  queryField,
  reconstructAcceptedQuery,
  reconstructRejectedQuery,
} from "./networkFlowQueryModel";
import {
  parseQueryScalar,
  parseQueryTimestamp,
} from "./networkFlowQueryValues";

const predicate = (
  field: QueryPredicateDraft["field"],
  op: QueryPredicateDraft["op"],
  input: QueryPredicateDraft["input"],
): QueryPredicateDraft => ({ id: "p", field, op, input, slot: "advanced" });
describe("Network Flow typed query compilation", () => {
  it("round trips lists equality both bounds null tests and repeated fields without rewriting", () => {
    const filters: NetworkFlowFilter[] = [
      {
        field_key: "network_flow.endpoint_ip",
        op: "in",
        value: ["192.0.2.1", "2001:db8::1"],
      },
      { field_key: "network_flow.ip_protocol", op: "in", value: [6, 17] },
      {
        field_key: "network_flow.bytes_count",
        op: "eq",
        value: "18446744073709551615",
      },
      {
        field_key: "network_flow.packets_count",
        op: "range",
        value: { lte: "42" },
      },
      {
        field_key: "network_flow.bytes_count",
        op: "range",
        value: { gte: "1", lte: "18446744073709551615" },
      },
      {
        field_key: "network_flow.bytes_count",
        op: "range",
        value: { gte: "2", lte: null },
      },
      { field_key: "network_flow.src_port", op: "is_null" },
      {
        field_key: "network_flow.input_interface",
        op: "in",
        value: ["01", "1", "a,b", ""],
      },
    ];
    const query = { filters, sort: [], timeWindow: null };
    const draft = reconstructAcceptedQuery(query);
    expect(draft.predicates.filter((p) => p.slot !== "advanced")).toHaveLength(
      1,
    );
    const result = compileAcceptedDraft(draft);
    expect(result).toEqual({ ok: true, value: query });
    if (!result.ok) throw new Error("compile");
    result.value.filters.forEach((f, i) => {
      expect(f).toBe(filters[i]);
    });
    const unrelated = compileAcceptedDraft({
      ...draft,
      startUTC: "2026-07-10T12:00:00Z",
    });
    expect(unrelated.ok && unrelated.value.filters).toEqual(filters);
  });
  it("validates exact uint64 bounds and rejects numeric coercion", () => {
    for (const text of ["0", "9007199254740993", "18446744073709551615"]) {
      const result = compilePredicate(
        predicate("network_flow.bytes_count", "eq", { kind: "scalar", text }),
      );
      expect(result.ok && result.value).toEqual({
        field_key: "network_flow.bytes_count",
        op: "eq",
        value: text,
      });
    }
    for (const text of [
      "18446744073709551616",
      "01",
      "-1",
      "1e3",
      "1.0",
      " 1",
      "",
    ])
      expect(parseQueryScalar("counter", text).ok, text).toBe(false);
    for (const text of ["256", "", "1,2", "0x10", " 6", "6.0"])
      expect(parseQueryScalar("protocol", text).ok, text).toBe(false);
    expect(parseQueryScalar("positive_integer", "9007199254740993").ok).toBe(
      false,
    );
  });
  it("uses inclusive numeric and half-open timestamp ranges without epsilon", () => {
    expect(
      compilePredicate(
        predicate("network_flow.bytes_count", "range", {
          kind: "range",
          lower: "5",
          upper: "5",
        }),
      ).ok,
    ).toBe(true);
    expect(
      compilePredicate(
        predicate("network_flow.bytes_count", "range", {
          kind: "range",
          lower: "",
          upper: "5",
        }),
      ),
    ).toEqual({
      ok: true,
      value: {
        field_key: "network_flow.bytes_count",
        op: "range",
        value: { gte: null, lte: "5" },
      },
    });
    const lower = "2026-07-10T12:00:00.000001Z";
    const upper = "2026-07-10T12:00:00.000002Z";
    expect(
      compilePredicate(
        predicate("network_flow.flow_start_utc", "range", {
          kind: "range",
          lower,
          upper,
        }),
      ),
    ).toEqual({
      ok: true,
      value: {
        field_key: "network_flow.flow_start_utc",
        op: "range",
        value: { gte: lower, lt: upper },
      },
    });
    for (const [a, b] of [
      [lower, lower],
      [upper, lower],
      ["", ""],
    ])
      expect(
        compilePredicate(
          predicate("network_flow.flow_start_utc", "range", {
            kind: "range",
            lower: a ?? "",
            upper: b ?? "",
          }),
        ).ok,
      ).toBe(false);
    expect(
      compilePredicate(
        predicate("network_flow.flow_start_utc", "eq", {
          kind: "scalar",
          text: lower,
        }),
      ).ok,
    ).toBe(false);
    for (const text of [
      "2026-02-30T00:00:00Z",
      "0000-01-01T00:00:00Z",
      "2026-07-10T12:00:60Z",
      "2026-07-10T12:00:00.1234567Z",
      "2026-07-10T12:00:00",
    ])
      expect(parseQueryTimestamp(text), text).toBeNull();
    expect(parseQueryTimestamp("0001-01-01T00:00:00Z")).not.toBeNull();
  });
  it("rejects canonical duplicates and invalid list members without removing entries", () => {
    for (const members of [
      ["6", ""],
      ["6", "6"],
      ["6", "6e0"],
    ])
      expect(
        compilePredicate(
          predicate("network_flow.ip_protocol", "in", {
            kind: "list",
            members,
          }),
        ).ok,
      ).toBe(false);
    expect(
      compilePredicate(
        predicate("network_flow.src_ip", "in", {
          kind: "list",
          members: ["2001:DB8::1", "2001:db8:0:0:0:0:0:1"],
        }),
      ).ok,
    ).toBe(false);
    const filters: NetworkFlowFilter[] = [
      {
        field_key: "network_flow.src_ip",
        op: "cidr_contains",
        value: "192.0.2.1/24",
      },
      {
        field_key: "network_flow.src_ip",
        op: "cidr_contains",
        value: "192.0.2.99/24",
      },
    ];
    const draft = reconstructAcceptedQuery({
      filters,
      sort: [],
      timeWindow: null,
    });
    expect(compileAcceptedDraft(draft).ok).toBe(false);
    expect(draft.predicates).toHaveLength(2);
    const tooMany = {
      ...draft,
      predicates: Array.from({ length: 17 }, (_, index) => ({
        ...predicate("network_flow.bytes_count", "eq", {
          kind: "scalar",
          text: String(index),
        }),
        id: "limit-" + index,
      })),
    };
    expect(compileAcceptedDraft(tooMany).ok).toBe(false);
    expect(tooMany.predicates).toHaveLength(17);

    const [first, second] = filters;
    if (!first || !second) throw new Error("Missing fixture");
    expect(canonicalFilterKey(first)).toBe(canonicalFilterKey(second));
    expect(parseQueryScalar("ip", "::ffff:192.0.2.1").ok).toBe(true);
    expect(parseQueryScalar("ip", "192.000.2.1").ok).toBe(false);
    expect(parseQueryScalar("ip", "fe80::1%eth0").ok).toBe(false);
    expect(parseQueryScalar("ip", "192.0.2.1/33", true).ok).toBe(false);
  });
  it("validates overlap filters together while keeping graph time range separate", () => {
    const query = {
      filters: [
        {
          field_key: "network_flow.flow_end_utc",
          op: "range",
          value: { gte: "2026-07-10T12:00:00.000000Z" },
        },
      ] satisfies NetworkFlowFilter[],
      sort: [],
      timeWindow: { startUTC: "2026-07-10T12:00:00Z", endUTC: null },
    };
    expect(
      compileAcceptedDraft(reconstructAcceptedQuery(query), "rows").ok,
    ).toBe(false);
    expect(
      compileAcceptedDraft(reconstructAcceptedQuery(query), "graph").ok,
    ).toBe(true);
    const plain = {
      filters: [],
      sort: [],
      timeWindow: {
        startUTC: "2026-07-10T12:00:00Z",
        endUTC: "2026-07-10T13:00:00Z",
      },
    };
    expect(
      graphInitialRequest({
        filters: [],
        tableScope: {
          mode: "active_table",
          active_table_id: "nft_" + "a".repeat(32),
        },
        aggregation: { mode: "time_bucket_v1", bucket_width_seconds: 60 },
        timeRange: {
          start_utc: "2026-07-10T12:00:00.000000Z",
          end_utc: "2026-07-10T12:00:00.000001Z",
        },
      }).time_range,
    ).toEqual({
      start_utc: "2026-07-10T12:00:00Z",
      end_utc: "2026-07-10T12:00:00.000001Z",
    });
    expect(acceptedInitialRequest(plain).filters).toEqual([
      {
        field_key: "network_flow.flow_end_utc",
        op: "range",
        value: { gte: plain.timeWindow.startUTC, lt: null },
      },
      {
        field_key: "network_flow.flow_start_utc",
        op: "range",
        value: { gte: null, lt: plain.timeWindow.endUTC },
      },
    ]);
  });
  it("preserves diagnostic lists and rejects unknown duplicate empty and reversed inputs", () => {
    const query = {
      errorCodes: ["network_flow_invalid_ip"] as const,
      fieldKeys: ["network_flow.src_ip", "network_flow.dst_ip"] as const,
      sourceRowRange: { gte: 2, lte: 40 },
    };
    expect(compileRejectedDraft(reconstructRejectedQuery(query))).toEqual({
      ok: true,
      value: query,
    });
    for (const draft of [
      { ...reconstructRejectedQuery(query), fieldKeys: ["unknown"] },
      {
        ...reconstructRejectedQuery(query),
        errorCodes: ["network_flow_invalid_ip", "network_flow_invalid_ip"],
      },
      { ...reconstructRejectedQuery(query), lower: "41" },
      { ...reconstructRejectedQuery(query), lower: "2.5" },
      { ...reconstructRejectedQuery(query), fieldKeys: [""] },
    ])
      expect(compileRejectedDraft(draft).ok).toBe(false);
  });
  it("compiles every offered presentation operator into an owner-valid filter shape", () => {
    const grid = networkFlowPresentationMetadata.grid_schemas[0];
    for (const column of grid.columns) {
      const metadata = queryField(column.field_key);
      for (const op of column.filter_operators) {
        if (!metadata)
          throw new Error("Unregistered offered field: " + column.field_key);
        const scalar =
          metadata.kind === "ip"
            ? "192.0.2.1"
            : metadata.kind === "timestamp"
              ? "2026-07-10T12:00:00Z"
              : metadata.kind === "text"
                ? "a,b"
                : "6";
        const input: QueryPredicateDraft["input"] =
          op === "range"
            ? { kind: "range", lower: scalar, upper: "" }
            : op === "in"
              ? { kind: "list", members: [scalar] }
              : op === "is_null" || op === "not_null"
                ? { kind: "null" }
                : {
                    kind: "scalar",
                    text: op === "cidr_contains" ? scalar + "/24" : scalar,
                  };
        const result = compilePredicate(
          predicate(metadata.field_key, op, input),
        );
        expect(result.ok, column.field_key + " " + op).toBe(true);
        if (result.ok)
          expect(decodeNetworkFlowFilter(result.value).ok).toBe(true);
      }
    }
    expect(
      decodeNetworkFlowFilter({
        field_key: "network_flow.src_port",
        op: "is_null",
        value: null,
      }).ok,
    ).toBe(false);
    expect(
      decodeNetworkFlowFilter({
        field_key: "network_flow.flow_start_utc",
        op: "eq",
        value: "2026-07-10T12:00:00Z",
      }).ok,
    ).toBe(false);
  });
});
