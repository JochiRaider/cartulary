import { describe, expect, it } from "vitest";
import { explorationFixture } from "./explorationTestFixtures";
import {
  emptyExploration,
  explorationFocusCurrent,
  explorationObjects,
  explorationPresentation,
  transitionExploration as move,
} from "./networkFlowExplorationNavigation";
import {
  networkFlowEdgeLinkCandidate,
  networkFlowVertexLinkCandidate,
} from "./networkFlowIndicatorLinkModel";

describe("Exploration navigation", () => {
  it("clears selection and both pages on populated and empty bucket transitions", () => {
    const graph = explorationFixture(501, 1001, true);
    const initial = move(emptyExploration("reader"), {
      type: "accept",
      result: graph,
    });
    for (const index of [1, 2]) {
      let state = move(initial, {
        type: "select",
        selector: required(graph.edge_annotations[0]).selector,
      });
      state = move(state, { type: "page", kind: "vertices", page: 1 });
      state = move(state, { type: "page", kind: "edges", page: 1 });
      state = move(state, { type: "bucket", index });
      expect(state.selection).toBeNull();
      expect([state.vertexPage, state.edgePage]).toEqual([0, 0]);
      expect(explorationObjects(state).edges).toHaveLength(
        index === 1 ? 0 : 1001,
      );
      expect(state.notice).toContain("Selection cleared");
    }
  });
  it("retains exact off-page selection and computes reveal and close focus semantically", () => {
    const graph = explorationFixture();
    for (const kind of ["vertices", "edges"] as const) {
      let state = move(emptyExploration("reader"), {
        type: "accept",
        result: graph,
      });
      const entry = required(explorationObjects(state)[kind][0]);
      state = move(state, { type: "select", selector: entry.selector });
      const revision = state.selectionRevision;
      state = move(state, { type: "page", kind, page: 1 });
      expect(explorationPresentation(state).offPage).toBe(true);
      expect(state.selection).toEqual(entry.selector);
      expect(state.selectionRevision).toBe(revision);
      const revealed = move(state, { type: "reveal" });
      expect(explorationPresentation(revealed).offPage).toBe(false);
      expect(revealed.focus?.selector).toEqual(entry.selector);
      expect(revealed.selectionRevision).toBe(revision);
      const closed = move(state, { type: "close" });
      expect(closed.selection).toBeNull();
      expect(closed.focus?.selector).toEqual(entry.selector);
      expect(move(closed, { type: "interact" }).focus).toBeNull();
      const oldIntent = required(revealed.focus);
      let returned = move(revealed, {
        type: "context",
        contextKey: "another reader",
        active: false,
      });
      returned = move(returned, {
        type: "context",
        contextKey: "reader",
        active: true,
      });
      returned = move(returned, { type: "accept", result: graph });
      returned = move(returned, { type: "select", selector: entry.selector });
      returned = move(returned, { type: "reveal" });
      expect(explorationFocusCurrent(returned, oldIntent)).toBe(false);
      expect(explorationFocusCurrent(returned, required(returned.focus))).toBe(
        true,
      );
    }
  });
  it("bounds every presentation synchronously at exact limits and after shrinking replacement", () => {
    for (const [vertices = 0, edges = 0] of [
      [500, 1000],
      [501, 1001],
      [0, 0],
    ]) {
      const graph = explorationFixture(vertices, edges);
      let state = move(emptyExploration("reader"), {
        type: "accept",
        result: graph,
      });
      expect(explorationPresentation(state).vertices).toHaveLength(
        Math.min(vertices, 500),
      );
      expect(explorationPresentation(state).edges).toHaveLength(
        Math.min(edges, 1000),
      );
      state = move(state, { type: "page", kind: "edges", page: 1 });
      const small = explorationFixture(2, 1);
      state = move(state, { type: "accept", result: small });
      expect(state.edgePage).toBe(0);
      expect(explorationPresentation(state).edges).toHaveLength(1);
    }
    let state = move(emptyExploration("reader"), {
      type: "accept",
      result: explorationFixture(2, 1, true),
    });
    state = move(state, { type: "bucket", index: 2 });
    expect(
      explorationPresentation({ ...state, bucketIndex: 999 }).edges,
    ).toHaveLength(0);
    const graph = explorationFixture(2, 1, true);
    if (graph.result_variant.kind === "time_bucket_v1") {
      state = move(state, {
        type: "accept",
        result: {
          ...graph,
          result_variant: {
            ...graph.result_variant,
            time_buckets: [required(graph.result_variant.time_buckets[1])],
          },
        },
      });
    }
    expect(explorationPresentation(state).edges).toHaveLength(0);
    expect(move(state, { type: "bucket", index: -1 })).toBe(state);
  });
  it("preserves metadata and subview continuity but resets a different accepted result", () => {
    const graph = explorationFixture(501, 1001, true);
    let state = move(emptyExploration("reader"), {
      type: "accept",
      result: graph,
    });
    state = move(state, { type: "bucket", index: 2 });
    state = move(state, { type: "page", kind: "vertices", page: 1 });
    state = move(state, {
      type: "select",
      selector: required(explorationPresentation(state).vertices[0]).selector,
    });
    const selection = state.selection;
    state = move(move(state, { type: "active", active: false }), {
      type: "active",
      active: true,
    });
    const refreshed = structuredClone(graph);
    required(refreshed.source_table_refs[0]).table_version = 2;
    state = move(state, { type: "accept", result: refreshed });
    expect([state.bucketIndex, state.vertexPage, state.selection]).toEqual([
      2,
      1,
      selection,
    ]);
    state = move(state, {
      type: "accept",
      result: { ...graph, graph_query_digest: "new-result" },
    });
    expect([state.bucketIndex, state.vertexPage, state.selection]).toEqual([
      0,
      0,
      null,
    ]);
  });
  it("uses complete selectors and stable ordering without inventing link or contributor meaning", () => {
    const graph = explorationFixture(501, 1001, true);
    let state = move(emptyExploration("reader"), {
      type: "accept",
      result: graph,
    });
    const vertex = required(explorationObjects(state).vertices[0]);
    const edge = required(explorationObjects(state).edges[0]);
    expect(
      networkFlowVertexLinkCandidate(graph, vertex.object)?.candidateValue,
    ).toBe(vertex.selector.endpoint_value);
    expect(
      networkFlowEdgeLinkCandidate({
        graph,
        edge: edge.object,
        fieldKey: "network_flow.src_ip",
      }),
    ).toBeNull();
    expect(
      move(state, {
        type: "select",
        selector: { ...vertex.selector, endpoint_value: "203.0.113.99" },
      }),
    ).toBe(state);
    state = move(state, { type: "select", selector: vertex.selector });
    const shuffled = {
      ...graph,
      vertex_selectors: [...graph.vertex_selectors].reverse(),
      edge_annotations: [...graph.edge_annotations].reverse(),
    };
    const reordered = move(state, { type: "accept", result: shuffled });
    expect(
      explorationObjects(reordered).vertices.map((v) => v.selector),
    ).toEqual(explorationObjects(state).vertices.map((v) => v.selector));
    expect(explorationObjects(reordered).edges.map((e) => e.selector)).toEqual(
      explorationObjects(state).edges.map((e) => e.selector),
    );
    const defaultGraph = explorationFixture(2, 1);
    expect(
      networkFlowEdgeLinkCandidate({
        graph: defaultGraph,
        edge: required(defaultGraph.graph_projection_result.edges[0]),
        fieldKey: "network_flow.src_ip",
      })?.candidateValue,
    ).toBe("192.0.0.0");
  });
});

function required<T>(value: T | null | undefined): T {
  if (value === undefined || value === null)
    throw new Error("Missing fixture member");
  return value;
}
