import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { NetworkFlowRequestError } from "./networkFlowErrors";
import {
  type NetworkFlowAcceptedQuery,
  reconstructAcceptedQuery,
} from "./networkFlowQueryModel";
import { tableFixture } from "./tableLifecycleTestFixtures";
import { useNetworkFlowQueryAuthoring } from "./useNetworkFlowQueryAuthoring";

afterEach(cleanup);
const options = {
  contextKey: "incident:actor:session:availability",
  activeTableId: tableFixture().network_flow_table_id,
  mode: "rows" as "rows" | "graph" | "rejected",
  tables: [tableFixture()],
};
const query: NetworkFlowAcceptedQuery = {
  filters: [
    { field_key: "network_flow.ip_protocol", op: "in", value: [6, 17] },
  ],
  sort: [],
  timeWindow: null,
};
const failure = new NetworkFlowRequestError({
  code: "network_flow_invalid_filter",
  filterIndex: 0,
  field: "network_flow.ip_protocol",
  safeMessage: "Correct this filter.",
  retryAction: "correct_request",
  retryable: false,
  status: 400,
});
describe("Network Flow query application ownership", () => {
  it("captures attempts retains successful state and ignores superseded success and failure", () => {
    const hook = renderHook(() => useNetworkFlowQueryAuthoring(options));
    act(() =>
      hook.result.current.setAcceptedDraft(reconstructAcceptedQuery(query)),
    );
    act(() => {
      expect(hook.result.current.applyAccepted()).toBe(true);
    });
    const revision = hook.result.current.acceptedRevision;
    const captured = hook.result.current.acceptedQuery;
    expect(hook.result.current.appliedAccepted.filters).toEqual([]);
    act(() => hook.result.current.acceptedResult(revision, null));
    expect(hook.result.current.appliedAccepted).toBe(captured);
    act(() => {
      expect(hook.result.current.applyAccepted()).toBe(true);
    });
    expect(hook.result.current.acceptedRevision).toBe(revision);
    act(() =>
      hook.result.current.setAcceptedDraft({
        ...hook.result.current.acceptedDraft,
        endUTC: "invalid",
      }),
    );
    act(() => {
      expect(hook.result.current.applyAccepted()).toBe(false);
    });
    expect(hook.result.current.acceptedQuery).toBe(captured);
    act(() =>
      hook.result.current.setAcceptedDraft({
        ...hook.result.current.acceptedDraft,
        endUTC: "2026-07-10T13:00:00Z",
      }),
    );
    act(() => {
      hook.result.current.applyAccepted();
    });
    const rejectedRevision = hook.result.current.acceptedRevision;
    act(() => hook.result.current.acceptedResult(rejectedRevision, failure));
    expect(hook.result.current.appliedAccepted).toBe(captured);
    expect(hook.result.current.acceptedDraft.endUTC).toBe(
      "2026-07-10T13:00:00Z",
    );
    expect(hook.result.current.acceptedStatus).toBe("failed");
    act(() => hook.result.current.clearAccepted());
    act(() => {
      hook.result.current.acceptedResult(rejectedRevision, null);
      hook.result.current.acceptedResult(rejectedRevision, failure);
    });
    expect(hook.result.current.acceptedStatus).toBe("pending");
    expect(hook.result.current.acceptedQuery.filters).toEqual([]);
  });
  it("attributes rejection to the submitted entry without rewriting a newer draft", () => {
    const hook = renderHook(() => useNetworkFlowQueryAuthoring(options));
    act(() =>
      hook.result.current.setAcceptedDraft(reconstructAcceptedQuery(query)),
    );
    act(() => {
      hook.result.current.applyAccepted();
    });
    const revision = hook.result.current.acceptedRevision;
    act(() =>
      hook.result.current.setAcceptedDraft({
        ...hook.result.current.acceptedDraft,
        predicates: [],
      }),
    );
    act(() => hook.result.current.acceptedResult(revision, failure));
    expect(hook.result.current.acceptedIssues[0]?.path).toBe("query");
    expect(hook.result.current.acceptedDraft.predicates).toEqual([]);
  });
  it("retains separate drafts through navigation and purges authority replacements", () => {
    const hook = renderHook((props) => useNetworkFlowQueryAuthoring(props), {
      initialProps: options,
    });
    act(() =>
      hook.result.current.setAcceptedDraft(reconstructAcceptedQuery(query)),
    );
    act(() =>
      hook.result.current.setRejectedDraft({
        errorCodes: ["network_flow_invalid_ip"],
        fieldKeys: ["network_flow.src_ip", "network_flow.dst_ip"],
        lower: "1",
        upper: "4",
      }),
    );
    hook.rerender({
      ...options,
      mode: "rejected",
      activeTableId: tableFixture("b").network_flow_table_id,
    });
    expect(hook.result.current.acceptedDraft.predicates).toHaveLength(1);
    expect(hook.result.current.acceptedQuery.filters).toEqual([]);
    act(() => {
      expect(hook.result.current.applyRejected()).toBe(true);
    });
    const revision = hook.result.current.rejectedRevision;
    hook.rerender({ ...options, contextKey: "replacement" });
    act(() => hook.result.current.rejectedResult(revision, null));
    expect(hook.result.current.acceptedDraft.predicates).toEqual([]);
    expect(hook.result.current.rejectedDraft.fieldKeys).toEqual([]);
    expect(hook.result.current.appliedRejected.fieldKeys).toEqual([]);
  });
  it("applies graph scope and aggregation atomically and retains invalid selected IDs", () => {
    const hook = renderHook((props) => useNetworkFlowQueryAuthoring(props), {
      initialProps: options,
    });
    act(() =>
      hook.result.current.setGraphDraft({
        scopeMode: "selected_tables",
        selectedTableIds: [options.activeTableId],
        aggregation: {
          mode: "time_bucket_v1",
          bucket_width_seconds: 3600,
          include_example_row_refs: true,
        },
      }),
    );
    act(() =>
      hook.result.current.setAcceptedDraft(reconstructAcceptedQuery(query)),
    );
    act(() => {
      hook.result.current.applyAccepted();
    });
    expect(hook.result.current.graphSettings.scopeMode).toBe("active_table");
    act(() =>
      hook.result.current.acceptedResult(
        hook.result.current.acceptedRevision,
        null,
      ),
    );
    hook.rerender({ ...options, mode: "graph" });
    act(() => {
      expect(hook.result.current.applyAccepted()).toBe(false);
    });
    act(() =>
      hook.result.current.setAcceptedDraft({
        ...hook.result.current.acceptedDraft,
        startUTC: "2026-07-10T00:00:00Z",
        endUTC: "2026-07-10T01:00:00Z",
      }),
    );
    act(() => {
      expect(hook.result.current.applyAccepted()).toBe(true);
    });
    expect(hook.result.current.graphSettings).toBe(
      hook.result.current.graphDraft,
    );
    expect(hook.result.current.appliedGraph.scopeMode).toBe("active_table");
    act(() =>
      hook.result.current.acceptedResult(
        hook.result.current.acceptedRevision,
        null,
      ),
    );
    hook.rerender({
      ...options,
      mode: "graph",
      tables: [tableFixture("b")],
      activeTableId: tableFixture("b").network_flow_table_id,
    });
    act(() => {
      expect(hook.result.current.applyAccepted()).toBe(false);
    });
    expect(hook.result.current.graphDraft.selectedTableIds).toEqual([
      options.activeTableId,
    ]);
    act(() => hook.result.current.clearAccepted());
    expect(hook.result.current.graphDraft.scopeMode).toBe("active_table");
    expect(hook.result.current.graphDraft.aggregation.mode).toBe(
      "default_flow_edge_v1",
    );
    expect(hook.result.current.acceptedDraft.startUTC).toBe("");
  });
});
