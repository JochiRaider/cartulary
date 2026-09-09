import { describe, expect, it, vi } from "vitest";
import type {
  NetworkFlowSavedGraphContributorResult,
  NetworkFlowSavedGraphResult,
} from "../services/networkFlowContractAdapter";
import {
  initialSavedGraphResultState,
  SavedGraphResultNavigation,
  type SavedGraphResultPort,
} from "./SavedGraphResultNavigation";
import {
  deferredSavedGraph,
  savedGraphBindingFixture,
  savedGraphFixture,
  savedGraphResultFixture,
  savedGraphSelectorFixture,
} from "./savedGraphTestFixtures";

const selected = () =>
  savedGraphFixture({ selected_result_binding: savedGraphBindingFixture });
const page = (
  graph = selected(),
  cursor: string | null = null,
): NetworkFlowSavedGraphContributorResult => ({
  schema_id: "cartulary.network_flow.graph_view_contributor_query_result.v2",
  graph_view_id: graph.graph_view_id,
  projection_result_id: savedGraphBindingFixture.projection_result_id,
  selector: savedGraphSelectorFixture,
  contributors: [],
  meta: {
    paging: { limit: 100, returned_count: 0, next_cursor_token: cursor },
  },
});
function setup() {
  let state = initialSavedGraphResultState(),
    current = true;
  const port: SavedGraphResultPort = {
    result: vi.fn(async () => savedGraphResultFixture()),
    contributors: vi.fn(async () => page()),
  };
  const navigation = new SavedGraphResultNavigation((next) => {
    state = next;
  });
  navigation.bind(port);
  navigation.authorize(selected(), () => current);
  return {
    navigation,
    port,
    state: () => state,
    loseAuthority: () => {
      current = false;
      navigation.clear();
    },
  };
}
describe("Saved graph semantic result continuity", () => {
  it("retains the exact result and contributor context through rename and unchanged declaration polling", async () => {
    const { navigation, port, state } = setup();
    const large = savedGraphResultFixture();
    large.result.graph_projection_result.vertices = Array.from(
      { length: 501 },
      (_, i) => ({
        vertex_id: `vx_${i.toString(16).padStart(64, "0")}`,
        vertex_kind: "network_flow.ip_endpoint.v1",
        vertex_family: "network_flow.ip_endpoint.v1",
        labels: [],
        properties: {},
        metadata: {
          mapping_rule_id: null,
          aggregation_rule_id: null,
          aggregation_source_refs: [],
          mapped_metadata: {},
        },
        source_entity_ref: null,
        sort_key: String(i),
      }),
    );
    vi.mocked(port.result).mockResolvedValue(large);

    await navigation.loadResult();
    const first = state().result;
    navigation.setPage("vertexPage", 1);
    await navigation.selectObject(savedGraphSelectorFixture);
    navigation.authorize(
      { ...selected(), display_name: "Renamed", graph_view_version: 2 },
      () => true,
    );
    navigation.authorize(structuredClone(selected()), () => true);
    expect(state().result).toBe(first);
    expect(state().vertexPage).toBe(1);
    expect(state().selection).toEqual(savedGraphSelectorFixture);
    expect(port.result).toHaveBeenCalledTimes(1);
    expect(port.contributors).toHaveBeenCalledTimes(1);
  });
  it("keeps a prior result during pending refresh and failed reads but clears removed bindings immediately", async () => {
    const { navigation, port, state } = setup();
    await navigation.loadResult();
    const first = state().result;
    navigation.authorize(
      {
        ...selected(),
        latest_job_id: "00000000-0000-4000-8000-000000000099",
        graph_view_version: 2,
        materialization_generation: 2,
      },
      () => true,
    );
    vi.mocked(port.result).mockRejectedValueOnce(new Error("unavailable"));
    await navigation.loadResult();
    expect(state().resultState).toBe("error");
    expect(state().result).toBe(first);
    navigation.authorize(
      {
        ...selected(),
        last_failure_code: "network_flow_graph_materialization_timeout",
        last_failed_at: "2026-09-09T12:00:01Z",
      },
      () => true,
    );
    expect(state().result).toBe(first);
    navigation.authorize(savedGraphFixture(), () => true);
    expect(state().result).toBeNull();
    expect(state().selection).toBeNull();
  });
  it("fences deferred results across A to B to A and protected-scope replacement", async () => {
    const { navigation, port, state, loseAuthority } = setup();
    const a = deferredSavedGraph<NetworkFlowSavedGraphResult>();
    vi.mocked(port.result).mockReturnValueOnce(a.promise);
    const loading = navigation.loadResult();
    const b = { ...selected(), graph_view_id: `nfgv_${"b".repeat(32)}` };
    navigation.authorize(b, () => true);
    navigation.authorize(selected(), () => true);
    a.resolve(savedGraphResultFixture());
    await loading;
    expect(state().result).toBeNull();
    const late = deferredSavedGraph<NetworkFlowSavedGraphResult>();
    vi.mocked(port.result).mockReturnValueOnce(late.promise);
    const next = navigation.loadResult();
    loseAuthority();
    late.resolve(savedGraphResultFixture());
    await next;
    expect(state().result).toBeNull();
  });
  it("fences contributors by graph binding selector and request generation and keeps only one page", async () => {
    const { navigation, port, state } = setup();
    await navigation.loadResult();
    const late = deferredSavedGraph<NetworkFlowSavedGraphContributorResult>();
    vi.mocked(port.contributors).mockReturnValueOnce(late.promise);
    const selecting = navigation.selectObject(savedGraphSelectorFixture);
    navigation.authorize(
      { ...selected(), graph_view_id: `nfgv_${"b".repeat(32)}` },
      () => true,
    );
    late.resolve(page());
    await selecting;
    expect(state().selection).toBeNull();
    expect(state().contributors).toEqual([]);
    navigation.authorize(selected(), () => true);
    await navigation.loadResult();
    vi.mocked(port.contributors).mockResolvedValueOnce(
      page(selected(), "opaque-next"),
    );
    await navigation.selectObject(savedGraphSelectorFixture);
    await navigation.loadContributors(true);
    expect(vi.mocked(port.contributors).mock.calls.at(-1)?.[2]).toBe(
      "opaque-next",
    );
    expect(state().contributorPage).toBe(1);
    await navigation.loadContributors();
    expect(vi.mocked(port.contributors).mock.calls.at(-1)?.[2]).toBeUndefined();
    expect(state().contributorPage).toBe(0);
  });
  it("distinguishes initial failure from a failed reread and rejects cross-binding result labeling", async () => {
    const { navigation, port, state } = setup();
    const other = savedGraphResultFixture({
      ...selected(),
      selected_result_binding: {
        ...savedGraphBindingFixture,
        canonical_output_sha256: "e".repeat(64),
      },
    });
    vi.mocked(port.result).mockResolvedValueOnce(other);
    await navigation.loadResult();
    expect(state().resultState).toBe("error");
    expect(state().result).toBeNull();
    navigation.authorize(selected(), () => true);
    await navigation.loadResult();
    expect(state().result?.graph_view.graph_view_id).toBe(
      selected().graph_view_id,
    );
  });
});
