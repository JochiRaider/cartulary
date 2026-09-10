import { useCallback, useRef, useState } from "react";
import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";
import type { NetworkFlowRequestError } from "./networkFlowErrors";
import {
  acceptedQueryIdentity,
  compileAcceptedDraft,
  compileRejectedDraft,
  defaultGraphQuerySettings,
  emptyNetworkFlowAcceptedQuery,
  emptyNetworkFlowRejectedQuery,
  type GraphQuerySettings,
  graphQueryIdentity,
  type NetworkFlowAcceptedDraft,
  type NetworkFlowAcceptedQuery,
  type NetworkFlowRejectedDraft,
  type NetworkFlowRejectedQuery,
  type QueryIssue,
  reconstructAcceptedQuery,
  reconstructRejectedQuery,
  validateGraphDraft,
} from "./networkFlowQueryModel";

type QueryStatus = "applied" | "pending" | "failed";
type State = {
  contextKey: string | null;
  navigation: string;
  submittedAccepted: NetworkFlowAcceptedDraft | null;
  submittedRejected: NetworkFlowRejectedDraft | null;
  acceptedDraft: NetworkFlowAcceptedDraft;
  rejectedDraft: NetworkFlowRejectedDraft;
  graphDraft: GraphQuerySettings;
  acceptedQuery: NetworkFlowAcceptedQuery;
  rejectedQuery: NetworkFlowRejectedQuery;
  graphSettings: GraphQuerySettings;
  appliedAccepted: NetworkFlowAcceptedQuery;
  appliedRejected: NetworkFlowRejectedQuery;
  appliedGraph: GraphQuerySettings;
  graphApplicationRevision: number;
  acceptedRevision: number;
  rejectedRevision: number;
  acceptedStatus: QueryStatus;
  rejectedStatus: QueryStatus;
  acceptedIssues: readonly QueryIssue[];
  rejectedIssues: readonly QueryIssue[];
};
const initial = (contextKey: string | null, navigation: string): State => ({
  contextKey,
  navigation,
  submittedAccepted: null,
  submittedRejected: null,
  acceptedDraft: reconstructAcceptedQuery(emptyNetworkFlowAcceptedQuery),
  rejectedDraft: reconstructRejectedQuery(emptyNetworkFlowRejectedQuery),
  graphDraft: defaultGraphQuerySettings,
  graphSettings: defaultGraphQuerySettings,
  appliedGraph: defaultGraphQuerySettings,
  acceptedQuery: emptyNetworkFlowAcceptedQuery,
  appliedAccepted: emptyNetworkFlowAcceptedQuery,
  rejectedQuery: emptyNetworkFlowRejectedQuery,
  appliedRejected: emptyNetworkFlowRejectedQuery,
  graphApplicationRevision: 0,
  acceptedRevision: 0,
  rejectedRevision: 0,
  acceptedStatus: "applied",
  rejectedStatus: "applied",
  acceptedIssues: [],
  rejectedIssues: [],
});
/** Volatile authoring owner. Controllers execute captured queries; controls edit drafts. */
export function useNetworkFlowQueryAuthoring(options: {
  readonly contextKey: string | null;
  readonly activeTableId: string | null;
  readonly mode: "rows" | "rejected" | "graph";
  readonly tables: readonly NetworkFlowTable[];
}) {
  const navigation = JSON.stringify([options.mode, options.activeTableId]);
  const [stored, setStored] = useState(() =>
    initial(options.contextKey, navigation),
  );
  let state = stored;
  if (stored.contextKey !== options.contextKey) {
    state = initial(options.contextKey, navigation);
    // A new context cannot share revision IDs with a response from its predecessor.
    state.graphApplicationRevision = stored.graphApplicationRevision + 1;
    state.acceptedRevision = stored.acceptedRevision + 1;
    state.rejectedRevision = stored.rejectedRevision + 1;
    setStored(state);
  } else if (stored.navigation !== navigation) {
    state = {
      ...stored,
      navigation,
      acceptedQuery: stored.appliedAccepted,
      rejectedQuery: stored.appliedRejected,
      graphSettings: stored.appliedGraph,
      acceptedRevision: stored.acceptedRevision + 1,
      rejectedRevision: stored.rejectedRevision + 1,
      acceptedStatus: "applied",
      rejectedStatus: "applied",
      submittedAccepted: null,
      submittedRejected: null,
    };
    setStored(state);
  }
  const current = useRef(state);
  current.current = state;
  const context = useRef(options);
  context.current = options;
  const update = useCallback((change: (value: State) => State) => {
    const next = change(current.current);
    current.current = next;
    setStored(next);
  }, []);
  const setAcceptedDraft = useCallback(
    (draft: NetworkFlowAcceptedDraft) =>
      update((s) => ({ ...s, acceptedDraft: draft, acceptedIssues: [] })),
    [update],
  );
  const setRejectedDraft = useCallback(
    (draft: NetworkFlowRejectedDraft) =>
      update((s) => ({ ...s, rejectedDraft: draft, rejectedIssues: [] })),
    [update],
  );
  const setGraphDraft = useCallback(
    (draft: GraphQuerySettings) =>
      update((s) => ({ ...s, graphDraft: draft, acceptedIssues: [] })),
    [update],
  );
  const applyAccepted = useCallback(() => {
    const s = current.current;
    if (s.contextKey === null) return false;
    const surface = context.current.mode === "graph" ? "graph" : "rows";
    const result = compileAcceptedDraft(s.acceptedDraft, surface);
    const issues =
      result.ok && surface === "graph"
        ? validateGraphDraft(
            s.graphDraft,
            result.value,
            context.current.activeTableId,
            context.current.tables,
          )
        : result.ok
          ? []
          : result.issues;
    if (!result.ok || issues.length > 0) {
      update((v) => ({ ...v, acceptedIssues: issues }));
      return false;
    }
    const graphSettings = surface === "graph" ? s.graphDraft : s.appliedGraph;
    const unchanged =
      acceptedQueryIdentity(result.value) ===
        acceptedQueryIdentity(s.appliedAccepted) &&
      (surface !== "graph" ||
        graphQueryIdentity(graphSettings) ===
          graphQueryIdentity(s.appliedGraph));
    if (unchanged && s.acceptedStatus === "applied") return true;
    update((v) => ({
      ...v,
      acceptedQuery: result.value,
      graphSettings,
      submittedAccepted: s.acceptedDraft,
      acceptedStatus: "pending",
      acceptedIssues: [],
      graphApplicationRevision:
        surface === "graph"
          ? v.graphApplicationRevision + 1
          : v.graphApplicationRevision,
      acceptedRevision: v.acceptedRevision + 1,
    }));
    return true;
  }, [update]);
  const applyRejected = useCallback(() => {
    const s = current.current;
    if (s.contextKey === null) return false;
    const result = compileRejectedDraft(s.rejectedDraft);
    if (!result.ok) {
      update((v) => ({ ...v, rejectedIssues: result.issues }));
      return false;
    }
    if (
      s.rejectedStatus === "applied" &&
      JSON.stringify(result.value) === JSON.stringify(s.appliedRejected)
    )
      return true;
    update((v) => ({
      ...v,
      rejectedQuery: result.value,
      submittedRejected: s.rejectedDraft,
      rejectedStatus: "pending",
      rejectedIssues: [],
      rejectedRevision: v.rejectedRevision + 1,
    }));
    return true;
  }, [update]);
  const acceptedResult = useCallback(
    (revision: number, error: NetworkFlowRequestError | null) => {
      const s = current.current;
      if (revision !== s.acceptedRevision || s.contextKey === null) return;
      if (error) {
        const issue = queryErrorIssue(
          error,
          s.submittedAccepted,
          s.acceptedDraft,
          s.acceptedQuery,
        );
        update((v) => ({
          ...v,
          acceptedStatus: "failed",
          acceptedIssues: [issue],
        }));
      } else
        update((v) => ({
          ...v,
          appliedAccepted: v.acceptedQuery,
          appliedGraph:
            context.current.mode === "graph" ? v.graphSettings : v.appliedGraph,
          acceptedStatus: "applied",
          acceptedIssues:
            v.acceptedDraft === v.submittedAccepted ? [] : v.acceptedIssues,
        }));
    },
    [update],
  );
  const rejectedResult = useCallback(
    (revision: number, error: NetworkFlowRequestError | null) => {
      const s = current.current;
      if (revision !== s.rejectedRevision || s.contextKey === null) return;
      update((v) =>
        error
          ? {
              ...v,
              rejectedStatus: "failed",
              rejectedIssues: [
                { path: error.field ?? "query", message: error.message },
              ],
            }
          : {
              ...v,
              appliedRejected: v.rejectedQuery,
              rejectedStatus: "applied",
              rejectedIssues:
                v.rejectedDraft === v.submittedRejected ? [] : v.rejectedIssues,
            },
      );
    },
    [update],
  );
  const clearAccepted = useCallback(
    () =>
      update((s) => ({
        ...s,
        acceptedDraft: reconstructAcceptedQuery(emptyNetworkFlowAcceptedQuery),
        acceptedQuery: emptyNetworkFlowAcceptedQuery,
        appliedAccepted: emptyNetworkFlowAcceptedQuery,
        graphDraft: defaultGraphQuerySettings,
        graphSettings: defaultGraphQuerySettings,
        appliedGraph: defaultGraphQuerySettings,
        acceptedStatus: "pending",
        submittedAccepted: null,
        acceptedIssues: [],
        graphApplicationRevision: s.graphApplicationRevision + 1,
        acceptedRevision: s.acceptedRevision + 1,
      })),
    [update],
  );
  const clearRejected = useCallback(
    () =>
      update((s) => ({
        ...s,
        rejectedDraft: reconstructRejectedQuery(emptyNetworkFlowRejectedQuery),
        rejectedQuery: emptyNetworkFlowRejectedQuery,
        appliedRejected: emptyNetworkFlowRejectedQuery,
        rejectedStatus: "pending",
        submittedRejected: null,
        rejectedIssues: [],
        rejectedRevision: s.rejectedRevision + 1,
      })),
    [update],
  );
  const purge = useCallback(
    () =>
      update((s) => ({
        ...initial(null, s.navigation),
        graphApplicationRevision: s.graphApplicationRevision + 1,
        acceptedRevision: s.acceptedRevision + 1,
        rejectedRevision: s.rejectedRevision + 1,
      })),
    [update],
  );
  const sortAccepted = useCallback(
    (sort: NetworkFlowAcceptedQuery["sort"]) =>
      update((s) => ({
        ...s,
        acceptedQuery: { ...s.appliedAccepted, sort },
        acceptedDraft: { ...s.acceptedDraft, sort },
        acceptedStatus: "pending",
        submittedAccepted: reconstructAcceptedQuery({
          ...s.appliedAccepted,
          sort,
        }),
        acceptedRevision: s.acceptedRevision + 1,
      })),
    [update],
  );
  return {
    ...state,
    graphDirty:
      graphQueryIdentity(state.graphDraft) !==
      graphQueryIdentity(state.appliedGraph),
    setAcceptedDraft,
    setRejectedDraft,
    setGraphDraft,
    applyAccepted,
    applyRejected,
    acceptedResult,
    rejectedResult,
    clearAccepted,
    clearRejected,
    purge,
    sortAccepted,
  };
}
function queryErrorIssue(
  error: NetworkFlowRequestError,
  submitted: NetworkFlowAcceptedDraft | null,
  draft: NetworkFlowAcceptedDraft,
  query: NetworkFlowAcceptedQuery,
): QueryIssue {
  const index = error.filterIndex;
  const entry = index === null ? undefined : submitted?.predicates[index];
  const matching =
    entry &&
    draft.predicates.some(
      (p) =>
        p.id === entry.id &&
        JSON.stringify(p.input) === JSON.stringify(entry.input) &&
        p.field === entry.field &&
        p.op === entry.op,
    );
  const candidatePath = matching
    ? entry.id + ".value"
    : index !== null && index >= query.filters.length
      ? query.timeWindow?.startUTC && index === query.filters.length
        ? "startUTC"
        : "endUTC"
      : error.field === "time_range" || error.field === "start_utc"
        ? "startUTC"
        : error.field === "end_utc"
          ? "endUTC"
          : "query";
  const path =
    (candidatePath === "startUTC" || candidatePath === "endUTC") &&
    submitted?.[candidatePath] !== draft[candidatePath]
      ? "query"
      : candidatePath;
  return {
    path,
    message:
      error.message +
      (error.reasonCode
        ? " (" + error.reasonCode.replaceAll("_", " ") + ")."
        : ""),
  };
}
