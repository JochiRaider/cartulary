import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { workbookQueryMeta } from "../../testing/workbookQueryTestSupport";
import {
  buildQueryRequest,
  buildSavedViewQueryJson,
  emptyWorkbookQueryState,
} from "../models/workbookQuery";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  canonicalAuthoredQuery,
  readWorkbookQueryMetadata,
} from "./workbookQueryMetadata";

const contract = requireViewContract(timelineViewSchemaId);
const empty = emptyWorkbookQueryState();
const validate = (meta: unknown) =>
  readWorkbookQueryMetadata(contract, meta, empty, 100, undefined);

describe("Workbook canonical query metadata", () => {
  it("requires complete paging and the effective default sort without assuming a short page is terminal", () => {
    const meta = workbookQueryMeta(timelineViewSchemaId);
    expect(
      validate({
        ...meta,
        paging: { limit: 100, has_more: true, next_cursor: " opaque +/= " },
      }).paging,
    ).toEqual({ limit: 100, hasMore: true, nextCursor: " opaque +/= " });
    for (const invalid of [
      { ...meta, paging: undefined },
      { ...meta, query: undefined },
      { ...meta, query: { filters: [], sort: [] } },
      { ...meta, paging: { limit: 99, has_more: false, next_cursor: null } },
      { ...meta, paging: { limit: 100, has_more: true, next_cursor: null } },
      {
        ...meta,
        paging: { limit: 100, has_more: false, next_cursor: "cursor" },
      },
      { ...meta, query: { ...meta.query, group_by: null } },
    ])
      expect(() => validate(invalid)).toThrow();
  });

  it("adopts canonical filter operands without copying effective tails into saved or continuation overrides", () => {
    const requested = {
      ...empty,
      groupBy: "timeline.capture_state",
      filters: [
        {
          fieldKey: "timeline.tags",
          op: "contains_any" as const,
          arg: { values: ["B", "a", "a"] },
        },
      ],
    };
    const canonical = {
      ...requested,
      filters: [
        {
          fieldKey: "timeline.tags",
          op: "contains_any" as const,
          arg: { values: ["a", "b"] },
        },
      ],
    };
    const meta = workbookQueryMeta(timelineViewSchemaId, canonical);
    const result = readWorkbookQueryMetadata(
      contract,
      meta,
      requested,
      100,
      undefined,
    );
    const intent = canonicalAuthoredQuery(requested, result.canonicalQuery);
    expect(intent.sort).toEqual([]);
    expect(buildSavedViewQueryJson(contract, intent)).toEqual({
      sort: [],
      filters: meta.query.filters,
      group_by: requested.groupBy,
    });
    expect(buildQueryRequest(contract, intent)).toEqual({
      filters: meta.query.filters,
      group_by: requested.groupBy,
    });
    expect(
      buildQueryRequest(contract, { ...intent, groupBy: null }),
    ).not.toHaveProperty("group_by");
    expect(
      readWorkbookQueryMetadata(
        contract,
        meta,
        intent,
        100,
        "exact +/=",
        result.canonicalQuery,
      ).canonicalQuery,
    ).toEqual(result.canonicalQuery);
    const different = workbookQueryMeta(timelineViewSchemaId, {
      ...canonical,
      filters: [
        {
          fieldKey: "timeline.tags",
          op: "contains_any" as const,
          arg: { values: ["c"] },
        },
      ],
    });
    expect(() =>
      readWorkbookQueryMetadata(
        contract,
        different,
        intent,
        100,
        "exact +/=",
        result.canonicalQuery,
      ),
    ).toThrow("Continuation query mismatch");
  });

  it("correlates grouping and authored sort order and permits eight sorts without group injection", () => {
    const sort = contract.fields
      .filter((field) => contract.sortableFieldMap[field.fieldKey])
      .slice(0, 8)
      .map((field) => ({
        fieldKey: field.fieldKey,
        direction: "desc" as const,
      }));
    const requested = { ...empty, sort, groupBy: "timeline.capture_state" };
    const meta = workbookQueryMeta(timelineViewSchemaId, requested);
    expect(
      readWorkbookQueryMetadata(
        contract,
        meta,
        requested,
        100,
        undefined,
      ).canonicalQuery.sort.slice(0, sort.length),
    ).toEqual(sort);
    expect(buildQueryRequest(contract, requested).sort).toHaveLength(
      sort.length,
    );
    expect(() =>
      readWorkbookQueryMetadata(
        contract,
        meta,
        { ...requested, groupBy: null },
        100,
        undefined,
      ),
    ).toThrow();
    expect(() =>
      readWorkbookQueryMetadata(
        contract,
        meta,
        { ...requested, sort: [...sort].reverse() },
        100,
        undefined,
      ),
    ).toThrow();
  });
});
