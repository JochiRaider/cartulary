import { useCallback, useContext, useEffect, useRef, useState } from "react";
import { boundedRead } from "../../../services/asyncObservation";
import { WorkbookCandidateAuthorityContext } from "../../hooks/useWorkbookCandidateDiscovery";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type {
  MentionCandidate,
  TimelineMentionCandidatePort,
} from "./TimelineMentionCandidatePort";

type Read = Readonly<{
  kind: "initial" | "continuation" | "restart";
  cursor: string | null;
  chain: number;
}>;
type ReadAction = "initial" | "loadMore" | "retry" | "restart";
type FailureKind = "ordinary" | "unusable_continuation" | "authority";
type State = Readonly<{
  key: string;
  phase: "idle" | "loading" | "ready" | "failed";
  /** Only candidates accepted into the current cursor chain are eligible. */
  candidates: readonly MentionCandidate[];
  /** Authorized observations from before an admitted restart, shown as stale. */
  staleCandidates: readonly MentionCandidate[];
  nextCursor: string | null;
  hasMore: boolean;
  error: string | null;
  failure: FailureKind | null;
  failedRead: Read | null;
  pendingAction: ReadAction | null;
}>;

function empty(key: string): State {
  return {
    key,
    phase: "idle",
    candidates: [],
    staleCandidates: [],
    nextCursor: null,
    hasMore: false,
    error: null,
    failure: null,
    failedRead: null,
    pendingAction: null,
  };
}

function unusableContinuation(failure: WorkbookOperationFailure, read: Read) {
  return (
    read.cursor !== null &&
    (failure.kind === "invalid_contract" ||
      (failure.publicCode === "invalid_view_query" &&
        [
          "invalid_cursor_token",
          "cursor_query_mismatch",
          "cursor_snapshot_unavailable",
        ].includes(failure.publicReason ?? "")))
  );
}

function validPage(
  page: Extract<
    Awaited<ReturnType<TimelineMentionCandidatePort["page"]>>,
    { kind: "accepted" }
  >,
  entityType: MentionCandidate["entityType"],
  read: Read,
  loadedCursors: ReadonlySet<string>,
) {
  const { candidates, hasMore, nextCursor } = page.value;
  return (
    candidates.length <= 100 &&
    candidates.every(
      (candidate) =>
        !!candidate.recordId &&
        candidate.entityType === entityType &&
        typeof candidate.displayText === "string",
    ) &&
    new Set(candidates.map((candidate) => candidate.recordId)).size ===
      candidates.length &&
    hasMore === (nextCursor !== null) &&
    (nextCursor === null ||
      (typeof nextCursor === "string" &&
        nextCursor.length > 0 &&
        nextCursor !== read.cursor &&
        !loadedCursors.has(nextCursor)))
  );
}

/** Accumulated Timeline pages and their retry identity have one read owner. */
export function useTimelineMentionCandidates(
  port: TimelineMentionCandidatePort,
  entityType: MentionCandidate["entityType"],
  scopeKey: string,
  enabled: boolean,
  selectedTargetId = "",
) {
  const authority = useContext(WorkbookCandidateAuthorityContext);
  const admitted = enabled && authority.canRead;
  const key = JSON.stringify([
    scopeKey,
    entityType,
    admitted,
    authority.identity,
  ]);
  const [state, setState] = useState<State>(() => empty(key));
  const stateRef = useRef(state);
  const current = useRef({
    key,
    port,
    admitted,
    onAuthorityFailure: authority.onAuthorityFailure,
  });
  current.current = {
    key,
    port,
    admitted,
    onAuthorityFailure: authority.onAuthorityFailure,
  };
  const selectedId = useRef(selectedTargetId);
  selectedId.current = selectedTargetId;
  const active = useRef<AbortController | null>(null);
  const loadedCursors = useRef(new Set<string>());
  const chain = useRef(0);
  const publish = useCallback((next: State) => {
    stateRef.current = next;
    setState(next);
  }, []);
  const load = useCallback(
    async (read: Read, action: ReadAction) => {
      if (
        !admitted ||
        active.current ||
        stateRef.current.key !== key ||
        (action === "loadMore" &&
          (stateRef.current.phase !== "ready" ||
            !stateRef.current.hasMore ||
            stateRef.current.nextCursor !== read.cursor)) ||
        (action === "retry" &&
          (stateRef.current.phase !== "failed" ||
            stateRef.current.failedRead !== read))
      )
        return;
      const controller = new AbortController();
      active.current = controller;
      let previous = stateRef.current;
      if (action === "restart") {
        chain.current++;
        read = { kind: "restart", cursor: null, chain: chain.current };
        loadedCursors.current.clear();
        const stale = new Map(
          previous.staleCandidates
            .filter(
              (candidate) =>
                previous.candidates.length === 0 ||
                candidate.recordId === selectedId.current,
            )
            .map((candidate) => [candidate.recordId, candidate]),
        );
        for (const candidate of previous.candidates)
          stale.set(candidate.recordId, candidate);
        previous = {
          ...previous,
          candidates: [],
          staleCandidates: [...stale.values()],
          nextCursor: null,
          hasMore: false,
        };
      }
      publish({
        ...previous,
        phase: "loading",
        error: null,
        failure: null,
        failedRead: null,
        pendingAction: action,
      });
      const isCurrent = () =>
        !controller.signal.aborted &&
        active.current === controller &&
        current.current.key === key &&
        current.current.port === port &&
        current.current.admitted &&
        read.chain === chain.current;
      const fail = (failure: WorkbookOperationFailure) => {
        if (!isCurrent()) return;
        const authorityLost =
          workbookFailureLifecycle(failure).kind === "authority_unavailable";
        const unusable = unusableContinuation(failure, read);
        const before = stateRef.current;
        publish({
          ...before,
          ...(authorityLost
            ? {
                candidates: [],
                staleCandidates: [],
                nextCursor: null,
                hasMore: false,
              }
            : {}),
          phase: "failed",
          pendingAction: null,
          failedRead: authorityLost ? null : read,
          failure: authorityLost
            ? "authority"
            : unusable
              ? "unusable_continuation"
              : "ordinary",
          error: authorityLost
            ? "Target access needs to be checked."
            : unusable
              ? "This target page could not be used. Retry it or restart target discovery."
              : failure.message,
        });
        if (authorityLost) current.current.onAuthorityFailure(failure);
      };
      try {
        const result = await boundedRead(
          (signal) => port.page(entityType, read.cursor, signal),
          controller.signal,
        );
        if (!isCurrent()) return;
        if (result.kind === "aborted") {
          publish({
            ...stateRef.current,
            phase: stateRef.current.candidates.length ? "ready" : "idle",
            pendingAction: null,
          });
          return;
        }
        if (result.kind === "rejected") return fail(result.failure);
        if (!validPage(result, entityType, read, loadedCursors.current))
          return fail({
            kind: "invalid_contract",
            message: "Target paging changed. Restart target discovery.",
          });
        if (read.cursor !== null) loadedCursors.current.add(read.cursor);
        const before = stateRef.current;
        const candidates = new Map(
          (read.kind === "continuation" ? before.candidates : []).map(
            (candidate) => [candidate.recordId, candidate],
          ),
        );
        for (const candidate of result.value.candidates)
          candidates.set(candidate.recordId, candidate);
        publish({
          ...before,
          phase: "ready",
          candidates: [...candidates.values()],
          staleCandidates: before.staleCandidates.filter(
            (candidate) => !candidates.has(candidate.recordId),
          ),
          nextCursor: result.value.nextCursor,
          hasMore: result.value.hasMore,
          error: null,
          failure: null,
          failedRead: null,
          pendingAction: null,
        });
      } catch {
        fail({
          kind: "retryable",
          message: "Targets could not be loaded. Retry the read.",
        });
      } finally {
        if (active.current === controller) active.current = null;
      }
    },
    [admitted, entityType, key, port, publish],
  );
  useEffect(() => {
    active.current?.abort();
    active.current = null;
    loadedCursors.current.clear();
    chain.current++;
    publish(empty(key));
    if (admitted)
      void load(
        { kind: "initial", cursor: null, chain: chain.current },
        "initial",
      );
    return () => {
      active.current?.abort();
      active.current = null;
    };
  }, [admitted, key, load, publish]);
  const visible = state.key === key ? state : empty(key);
  return {
    ...visible,
    canRead: admitted,
    loadMore: () => {
      const snapshot = stateRef.current;
      if (snapshot.key === key && snapshot.nextCursor)
        return load(
          {
            kind: "continuation",
            cursor: snapshot.nextCursor,
            chain: chain.current,
          },
          "loadMore",
        );
      return undefined;
    },
    retry: () => {
      const snapshot = stateRef.current;
      if (snapshot.key === key && snapshot.failedRead)
        return load(snapshot.failedRead, "retry");
      return undefined;
    },
    restart: () =>
      load(
        { kind: "restart", cursor: null, chain: chain.current + 1 },
        "restart",
      ),
  };
}
