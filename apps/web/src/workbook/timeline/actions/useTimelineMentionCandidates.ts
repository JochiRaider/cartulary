import { useCallback, useContext, useEffect, useRef, useState } from "react";
import { boundedRead } from "../../../services/asyncObservation";
import { WorkbookCandidateAuthorityContext } from "../../hooks/useWorkbookCandidateDiscovery";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import { workbookFailureLifecycle } from "../../ports/WorkbookPortResult";
import type {
  MentionCandidate,
  MentionCandidateRequest,
  TimelineMentionCandidatePort,
} from "./TimelineMentionCandidatePort";

type ReadAction = "initial" | "next" | "previous" | "retry" | "restart";
type Read = MentionCandidateRequest &
  Readonly<{
    kind: "initial" | "continuation" | "previous" | "restart";
    previous: readonly (string | null)[];
  }>;
type State = Readonly<{
  key: string;
  phase: "idle" | "loading" | "ready" | "failed";
  candidates: readonly MentionCandidate[];
  staleCandidates: readonly MentionCandidate[];
  selected: MentionCandidate | null;
  selectionStale: boolean;
  cursor: string | null;
  previous: readonly (string | null)[];
  nextCursor: string | null;
  hasMore: boolean;
  error: string | null;
  failure: "ordinary" | "unusable_continuation" | "authority" | null;
  failedRead: Read | null;
  pendingAction: ReadAction | null;
}>;
function empty(key: string): State {
  return {
    key,
    phase: "idle",
    candidates: [],
    staleCandidates: [],
    selected: null,
    selectionStale: false,
    cursor: null,
    previous: [],
    nextCursor: null,
    hasMore: false,
    error: null,
    failure: null,
    failedRead: null,
    pendingAction: null,
  };
}

/** One page, bounded checkpoints and an independent selected observation. */
export function useTimelineMentionCandidates(
  port: TimelineMentionCandidatePort,
  entityType: MentionCandidate["entityType"],
  scopeKey: string,
  enabled: boolean,
  selectedTargetId = "",
) {
  const authority = useContext(WorkbookCandidateAuthorityContext);
  const admitted = enabled && authority.canRead;
  const [input, setInput] = useState({
    scopeKey,
    search: "",
    composing: false,
    edited: false,
  });
  const search = input.scopeKey === scopeKey ? input.search : "";
  const edited = input.scopeKey === scopeKey && input.edited;
  const composing = input.scopeKey === scopeKey && input.composing;
  const key = JSON.stringify([
    scopeKey,
    entityType,
    admitted,
    authority.identity,
    search,
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
  const publish = useCallback((next: State) => {
    stateRef.current = next;
    setState(next);
  }, []);
  const load = useCallback(
    async (read: Read, action: ReadAction) => {
      const before = stateRef.current;
      if (
        !admitted ||
        active.current ||
        before.key !== key ||
        (action === "retry" && before.failedRead !== read) ||
        before.failure === "authority"
      )
        return;
      const controller = new AbortController();
      active.current = controller;
      publish({
        ...before,
        phase: "loading",
        error: null,
        failure: null,
        failedRead: null,
        pendingAction: action,
        ...(action === "restart"
          ? {
              staleCandidates: before.candidates,
              candidates: [],
              selectionStale: true,
              previous: [],
              hasMore: false,
              nextCursor: null,
            }
          : {}),
      });
      const isCurrent = () =>
        !controller.signal.aborted &&
        active.current === controller &&
        current.current.key === key &&
        current.current.port === port &&
        current.current.admitted;
      const fail = (failure: WorkbookOperationFailure) => {
        if (!isCurrent()) return;
        const lost =
          workbookFailureLifecycle(failure).kind === "authority_unavailable";
        const unusable =
          read.cursor !== null &&
          (failure.kind === "invalid_contract" ||
            [
              "invalid_cursor_token",
              "cursor_query_mismatch",
              "cursor_snapshot_unavailable",
            ].includes(failure.publicReason ?? ""));
        publish({
          ...stateRef.current,
          ...(lost ? empty(key) : {}),
          phase: "failed",
          pendingAction: null,
          failedRead: lost ? null : read,
          failure: lost
            ? "authority"
            : unusable
              ? "unusable_continuation"
              : "ordinary",
          error: lost
            ? "Target access needs to be checked."
            : unusable
              ? "This continuation is unavailable. Restart target discovery."
              : failure.message,
        });
        if (lost) current.current.onAuthorityFailure(failure);
      };
      try {
        const result = await boundedRead(
          (signal) => port.page(read, signal),
          controller.signal,
        );
        if (!isCurrent()) return;
        if (result.kind === "aborted") {
          publish({
            ...stateRef.current,
            phase: before.phase,
            pendingAction: null,
          });
          return;
        }
        if (result.kind === "rejected") {
          fail(result.failure);
          return;
        }
        const page = result.value;
        if (
          page.candidates.length > port.policy.pageSize ||
          page.candidates.some(
            (candidate) =>
              !candidate.recordId ||
              candidate.entityType !== entityType ||
              !Number.isSafeInteger(candidate.rowVersion) ||
              candidate.rowVersion < 1 ||
              typeof candidate.displayText !== "string",
          ) ||
          new Set(page.candidates.map((candidate) => candidate.recordId))
            .size !== page.candidates.length ||
          page.hasMore !== (page.nextCursor !== null) ||
          (page.nextCursor !== null &&
            (!page.nextCursor ||
              page.nextCursor === read.cursor ||
              (read.kind !== "previous" &&
                read.previous.includes(page.nextCursor))))
        ) {
          fail({
            kind: "invalid_contract",
            message: "Target paging changed. Restart target discovery.",
          });
          return;
        }
        const revalidated = page.candidates.find(
          (candidate) => candidate.recordId === selectedId.current,
        );
        const selected = revalidated ?? stateRef.current.selected;
        publish({
          ...empty(key),
          phase: "ready",
          candidates: page.candidates,
          selected,
          selectionStale: !revalidated && stateRef.current.selectionStale,
          cursor: read.cursor,
          previous: read.previous,
          hasMore: page.hasMore,
          nextCursor: page.nextCursor,
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
    publish(empty(key));
    if (!admitted || composing) return;
    const read: Read = {
      entityType,
      search,
      cursor: null,
      previous: [],
      kind: "initial",
    };
    const timer = edited
      ? setTimeout(() => {
          void load(read, "initial");
        }, port.policy.settledInputMs)
      : null;
    if (!edited) void load(read, "initial");
    return () => {
      if (timer !== null) clearTimeout(timer);
      active.current?.abort();
      active.current = null;
    };
  }, [
    admitted,
    composing,
    edited,
    entityType,
    key,
    load,
    port,
    publish,
    search,
  ]);
  const visible = state.key === key ? state : empty(key);
  const navigate = (direction: "next" | "previous") => {
    const value = stateRef.current;
    if (value.key !== key || value.phase !== "ready") return;
    if (direction === "next" && value.hasMore && value.nextCursor)
      return load(
        {
          entityType,
          search,
          cursor: value.nextCursor,
          previous: [...value.previous, value.cursor].slice(
            -port.policy.previousCursorLimit,
          ),
          kind: "continuation",
        },
        "next",
      );
    if (direction === "previous" && value.previous.length)
      return load(
        {
          entityType,
          search,
          cursor: value.previous.at(-1) ?? null,
          previous: value.previous.slice(0, -1),
          kind: "previous",
        },
        "previous",
      );
    return undefined;
  };
  return {
    ...visible,
    search,
    composing,
    canRead: admitted,
    selected:
      !visible.selectionStale && visible.selected?.recordId === selectedTargetId
        ? visible.selected
        : null,
    selectedObservation:
      visible.selected?.recordId === selectedTargetId ? visible.selected : null,
    changeSearch: (value: string) => {
      active.current?.abort();
      active.current = null;
      setInput({ scopeKey, search: value, composing, edited: true });
    },
    setComposing: (value: boolean) =>
      setInput({ scopeKey, search, composing: value, edited }),
    select: (id: string) => {
      if (stateRef.current.key !== key) return;
      publish({
        ...stateRef.current,
        selected:
          stateRef.current.candidates.find(
            (candidate) => candidate.recordId === id,
          ) ?? null,
        selectionStale: false,
      });
    },
    next: () => navigate("next"),
    previousPage: () => navigate("previous"),
    retry: () => {
      const value = stateRef.current;
      if (
        value.key === key &&
        value.failedRead &&
        value.failure !== "unusable_continuation"
      )
        return load(value.failedRead, "retry");
      return undefined;
    },
    restart: () =>
      load(
        { entityType, search, cursor: null, previous: [], kind: "restart" },
        "restart",
      ),
  };
}
