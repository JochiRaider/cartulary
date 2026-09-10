import { networkFlowContractEqual } from "../services/networkFlowContractAdapter";
import type {
  NetworkFlowGraphEdge,
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
  NetworkFlowGraphVertex,
} from "./networkFlowClient";

export const explorationVertexLimit = 500;
export const explorationEdgeLimit = 1_000;
type VertexSelector = Extract<NetworkFlowGraphSelector, { kind: "vertex" }>;
type EdgeSelector = Exclude<NetworkFlowGraphSelector, VertexSelector>;
export type ExplorationVertex = {
  readonly object: NetworkFlowGraphVertex;
  readonly selector: VertexSelector;
  readonly label: string;
};
export type ExplorationEdge = {
  readonly object: NetworkFlowGraphEdge;
  readonly selector: EdgeSelector;
  readonly label: string;
};
type Bucket = Extract<
  NetworkFlowGraphResult["result_variant"],
  { kind: "time_bucket_v1" }
>["time_buckets"][number];
type ObjectSet = {
  readonly vertices: readonly ExplorationVertex[];
  readonly edges: readonly ExplorationEdge[];
};
export type ExplorationIndex = ObjectSet & {
  readonly vertexById: ReadonlyMap<string, ExplorationVertex>;
  readonly edgeById: ReadonlyMap<string, ExplorationEdge>;
  readonly buckets: readonly Bucket[];
  readonly bucketObjects: readonly ObjectSet[];
  readonly temporal: boolean;
};
export type ExplorationFocus = {
  readonly contextKey: string;
  readonly resultIdentity: string | null;
  readonly generation: number;
  readonly selector: NetworkFlowGraphSelector;
};
export type ExplorationNavigation = {
  readonly contextKey: string;
  readonly identity: string | null;
  readonly result: NetworkFlowGraphResult | null;
  readonly index: ExplorationIndex | null;
  readonly bucketIndex: number;
  readonly vertexPage: number;
  readonly edgePage: number;
  readonly selection: NetworkFlowGraphSelector | null;
  readonly selectionRevision: number;
  readonly interaction: number;
  readonly active: boolean;
  readonly focus: ExplorationFocus | null;
  readonly notice: string;
};
export function explorationFocusCurrent(
  state: ExplorationNavigation,
  intent: ExplorationFocus,
): boolean {
  return (
    state.active &&
    state.focus !== null &&
    state.contextKey === intent.contextKey &&
    state.identity === intent.resultIdentity &&
    state.interaction === intent.generation
  );
}
export type ExplorationAction =
  | {
      readonly type: "context";
      readonly contextKey: string;
      readonly active: boolean;
    }
  | { readonly type: "accept"; readonly result: NetworkFlowGraphResult }
  | { readonly type: "clear" }
  | { readonly type: "bucket"; readonly index: number }
  | {
      readonly type: "page";
      readonly kind: "vertices" | "edges";
      readonly page: number;
    }
  | {
      readonly type: "select";
      readonly selector: NetworkFlowGraphSelector | null;
    }
  | { readonly type: "close" }
  | { readonly type: "reveal" }
  | { readonly type: "active"; readonly active: boolean }
  | { readonly type: "interact" };

export function emptyExploration(
  contextKey: string,
  active = true,
): ExplorationNavigation {
  return {
    contextKey,
    active,
    identity: null,
    result: null,
    index: null,
    bucketIndex: 0,
    vertexPage: 0,
    edgePage: 0,
    selection: null,
    selectionRevision: 0,
    interaction: 0,
    focus: null,
    notice: "",
  };
}

export function explorationResultIdentity(
  result: NetworkFlowGraphResult,
): string {
  const projection = result.graph_projection_result;
  return JSON.stringify([
    result.graph_query_digest,
    projection.projection_result_id,
    projection.source_snapshot_id,
    projection.projection_schema_id,
    projection.projection_version,
    projection.normalized_configuration_sha256,
    projection.normalized_source_sha256,
    projection.canonical_output_sha256,
  ]);
}
export function explorationSelectionId(
  selector: NetworkFlowGraphSelector,
): string {
  return selector.kind === "vertex"
    ? selector.source_vertex_id
    : selector.source_edge_id;
}
const compare = (a: string, b: string) => (a < b ? -1 : a > b ? 1 : 0);
export function explorationEdgeLabel(selector: EdgeSelector): string {
  return `${selector.source_endpoint_value} → ${selector.destination_endpoint_value} · protocol ${selector.protocol} · port ${selector.destination_port_present ? selector.destination_port : "—"}`;
}
export function buildExplorationIndex(
  result: NetworkFlowGraphResult,
): ExplorationIndex {
  const projectedVertices = new Map(
    result.graph_projection_result.vertices.map((v) => [v.vertex_id, v]),
  );
  const projectedEdges = new Map(
    result.graph_projection_result.edges.map((e) => [e.edge_id, e]),
  );
  const vertices = result.vertex_selectors
    .flatMap((binding) => {
      const object = projectedVertices.get(binding.projected_vertex_id);
      return object
        ? [
            {
              object,
              selector: binding.selector,
              label: binding.selector.endpoint_value,
            },
          ]
        : [];
    })
    .sort(
      (a, b) =>
        compare(a.selector.endpoint_value, b.selector.endpoint_value) ||
        compare(a.selector.source_vertex_id, b.selector.source_vertex_id),
    );
  const edges = result.edge_annotations
    .flatMap((annotation) => {
      const object = projectedEdges.get(annotation.projected_edge_id);
      return object
        ? [
            {
              object,
              selector: annotation.selector,
              label: explorationEdgeLabel(annotation.selector),
            },
          ]
        : [];
    })
    .sort(
      (a, b) =>
        compare(
          a.selector.source_endpoint_value,
          b.selector.source_endpoint_value,
        ) ||
        compare(
          a.selector.destination_endpoint_value,
          b.selector.destination_endpoint_value,
        ) ||
        a.selector.protocol - b.selector.protocol ||
        Number(a.selector.destination_port_present) -
          Number(b.selector.destination_port_present) ||
        (a.selector.destination_port_present
          ? a.selector.destination_port
          : 0) -
          (b.selector.destination_port_present
            ? b.selector.destination_port
            : 0) ||
        compare(a.selector.source_edge_id, b.selector.source_edge_id),
    );
  const temporal = result.result_variant.kind === "time_bucket_v1";
  const buckets =
    result.result_variant.kind === "time_bucket_v1"
      ? result.result_variant.time_buckets
      : [];
  const bucketEdges = new Map<string, ExplorationEdge[]>();
  for (const edge of edges) {
    if (edge.selector.kind !== "time_bucket_edge") continue;
    const key = JSON.stringify([
      edge.selector.bucket_start_utc,
      edge.selector.bucket_end_utc,
    ]);
    const items = bucketEdges.get(key) ?? [];
    items.push(edge);
    bucketEdges.set(key, items);
  }
  return {
    vertices,
    edges,
    temporal,
    buckets,
    vertexById: new Map(vertices.map((v) => [v.selector.source_vertex_id, v])),
    edgeById: new Map(edges.map((e) => [e.selector.source_edge_id, e])),
    bucketObjects: buckets.map((bucket) => {
      const bucketItems =
        bucketEdges.get(JSON.stringify([bucket.start_utc, bucket.end_utc])) ??
        [];
      const endpoints = new Set(
        bucketItems.flatMap((edge) => [
          edge.object.src_vertex_id,
          edge.object.dst_vertex_id,
        ]),
      );
      return {
        edges: bucketItems,
        vertices: vertices.filter((vertex) =>
          endpoints.has(vertex.object.vertex_id),
        ),
      };
    }),
  };
}
const noObjects: ObjectSet = { vertices: [], edges: [] };
export function explorationObjects(state: ExplorationNavigation): ObjectSet {
  const index = state.index;
  if (index === null) return noObjects;
  return index.temporal
    ? (index.bucketObjects[state.bucketIndex] ?? noObjects)
    : index;
}
const pageBound = (page: number, count: number, limit: number) =>
  Math.min(
    Math.max(Number.isSafeInteger(page) ? page : 0, 0),
    Math.max(0, Math.ceil(count / limit) - 1),
  );
function selectedEntry(state: ExplorationNavigation) {
  const selection = state.selection;
  if (!selection) return null;
  const entry =
    selection.kind === "vertex"
      ? state.index?.vertexById.get(selection.source_vertex_id)
      : state.index?.edgeById.get(selection.source_edge_id);
  return entry && networkFlowContractEqual(entry.selector, selection)
    ? entry
    : null;
}
export function explorationPresentation(state: ExplorationNavigation) {
  const objects = explorationObjects(state);
  const vertexPage = pageBound(
    state.vertexPage,
    objects.vertices.length,
    explorationVertexLimit,
  );
  const edgePage = pageBound(
    state.edgePage,
    objects.edges.length,
    explorationEdgeLimit,
  );
  const vertices = objects.vertices.slice(
    vertexPage * explorationVertexLimit,
    (vertexPage + 1) * explorationVertexLimit,
  );
  const edges = objects.edges.slice(
    edgePage * explorationEdgeLimit,
    (edgePage + 1) * explorationEdgeLimit,
  );
  const selected = selectedEntry(state);
  return {
    vertices,
    edges,
    vertexPage,
    edgePage,
    vertexCount: objects.vertices.length,
    edgeCount: objects.edges.length,
    selected,
    offPage: selected !== null && ![...vertices, ...edges].includes(selected),
    bucket: state.index?.buckets[state.bucketIndex] ?? null,
    bucketIndex: state.bucketIndex,
    bucketCount: state.index?.buckets.length ?? 0,
    temporal: state.index?.temporal ?? false,
  };
}
function reconcile(state: ExplorationNavigation): ExplorationNavigation {
  const bucketIndex = pageBound(
    state.bucketIndex,
    state.index?.buckets.length ?? 0,
    1,
  );
  const bounded = { ...state, bucketIndex };
  const objects = explorationObjects(bounded);
  const entry = selectedEntry(bounded);
  const validSelection =
    entry !== null && [...objects.vertices, ...objects.edges].includes(entry);
  return {
    ...bounded,
    vertexPage: pageBound(
      state.vertexPage,
      objects.vertices.length,
      explorationVertexLimit,
    ),
    edgePage: pageBound(
      state.edgePage,
      objects.edges.length,
      explorationEdgeLimit,
    ),
    selection: validSelection ? state.selection : null,
    selectionRevision:
      state.selectionRevision +
      (state.selection !== null && !validSelection ? 1 : 0),
    focus: null,
  };
}

/** All navigation changes are synchronous and independent of React/DOM lifetime. */
export function transitionExploration(
  state: ExplorationNavigation,
  action: ExplorationAction,
): ExplorationNavigation {
  if (action.type === "context") {
    if (action.contextKey === state.contextKey) return state;
    return {
      ...emptyExploration(action.contextKey, action.active),
      selectionRevision: state.selectionRevision + 1,
      interaction: state.interaction + 1,
    };
  }
  if (action.type === "interact")
    return state.focus
      ? { ...state, focus: null, interaction: state.interaction + 1 }
      : state;
  if (action.type === "active")
    return action.active === state.active
      ? state
      : {
          ...state,
          active: action.active,
          focus: null,
          notice: "",
          interaction: state.interaction + 1,
        };
  if (action.type === "clear")
    return {
      ...emptyExploration(state.contextKey, state.active),
      selectionRevision: state.selectionRevision + 1,
      interaction: state.interaction + 1,
    };
  if (action.type === "accept") {
    const identity = explorationResultIdentity(action.result);
    const same =
      identity === state.identity &&
      networkFlowContractEqual(
        state.result?.semantic_query,
        action.result.semantic_query,
      );
    return reconcile({
      ...(same ? state : transitionExploration(state, { type: "clear" })),
      result: action.result,
      identity,
      index: buildExplorationIndex(action.result),
    });
  }
  if (!state.active || !state.result) return state;
  const next = { ...state, focus: null, interaction: state.interaction + 1 };
  const objects = explorationObjects(state);
  switch (action.type) {
    case "bucket": {
      if (
        !state.index?.temporal ||
        !Number.isSafeInteger(action.index) ||
        action.index < 0 ||
        action.index >= state.index.buckets.length ||
        action.index === state.bucketIndex
      )
        return state;
      return {
        ...next,
        bucketIndex: action.index,
        vertexPage: 0,
        edgePage: 0,
        selection: null,
        selectionRevision: state.selectionRevision + 1,
        notice: `Bucket ${action.index + 1} of ${state.index.buckets.length}.${state.selection ? " Selection cleared." : ""}`,
      };
    }
    case "page": {
      const limit =
        action.kind === "vertices"
          ? explorationVertexLimit
          : explorationEdgeLimit;
      if (
        !Number.isSafeInteger(action.page) ||
        action.page !==
          pageBound(action.page, objects[action.kind].length, limit)
      )
        return state;
      const key = action.kind === "vertices" ? "vertexPage" : "edgePage";
      if (state[key] === action.page) return state;
      const moved = { ...next, [key]: action.page };
      return {
        ...moved,
        notice: `${action.kind === "vertices" ? "Vertices" : "Edges"} page ${action.page + 1} of ${Math.max(1, Math.ceil(objects[action.kind].length / limit))}.${explorationPresentation(moved).offPage ? " Selected object is off-page." : ""}`,
      };
    }
    case "select": {
      if (networkFlowContractEqual(state.selection, action.selector))
        return state;
      if (
        action.selector !== null &&
        ![...objects.vertices, ...objects.edges].some((entry) =>
          networkFlowContractEqual(entry.selector, action.selector),
        )
      )
        return state;
      const selected = {
        ...next,
        selection: action.selector,
        selectionRevision: state.selectionRevision + 1,
      };
      return {
        ...selected,
        notice: action.selector
          ? `${action.selector.kind === "vertex" ? "Vertex" : "Edge"} selected. ${selectedEntry(selected)?.label ?? ""}`
          : "Selection cleared.",
      };
    }
    case "close":
      return state.selection === null
        ? state
        : {
            ...next,
            selection: null,
            selectionRevision: state.selectionRevision + 1,
            focus: {
              contextKey: state.contextKey,
              resultIdentity: state.identity,
              generation: next.interaction,
              selector: state.selection,
            },
            notice: "Selection cleared.",
          };
    case "reveal": {
      if (!state.selection) return state;
      const selection = state.selection;
      const kind = selection.kind === "vertex" ? "vertices" : "edges";
      const position = objects[kind].findIndex((entry) =>
        networkFlowContractEqual(entry.selector, selection),
      );
      if (position < 0) return state;
      return {
        ...next,
        [kind === "vertices" ? "vertexPage" : "edgePage"]: Math.floor(
          position /
            (kind === "vertices"
              ? explorationVertexLimit
              : explorationEdgeLimit),
        ),
        focus: {
          contextKey: state.contextKey,
          resultIdentity: state.identity,
          generation: next.interaction,
          selector: selection,
        },
        notice: "Selected object revealed.",
      };
    }
  }
}
