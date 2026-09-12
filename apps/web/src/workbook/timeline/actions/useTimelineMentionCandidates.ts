import { useCallback, useEffect, useRef, useState } from "react";
import { boundedRead } from "../../../services/asyncObservation";
import type {
  MentionCandidate,
  TimelineMentionCandidatePort,
} from "./TimelineMentionCandidatePort";

type State = Readonly<{
  key: string;
  phase: "idle" | "loading" | "ready" | "failed";
  candidates: readonly MentionCandidate[];
  nextCursor: string | null;
  hasMore: boolean;
  error: string | null;
}>;
export function useTimelineMentionCandidates(
  port: TimelineMentionCandidatePort,
  entityType: MentionCandidate["entityType"],
  scopeKey: string,
  enabled: boolean,
) {
  const key = `${scopeKey}:${entityType}:${enabled}`;
  const [state, setState] = useState<State>({
    key,
    phase: "idle",
    candidates: [],
    nextCursor: null,
    hasMore: false,
    error: null,
  });
  const current = useRef({ key, port, enabled });
  current.current = { key, port, enabled };
  const active = useRef<AbortController | null>(null);
  const loadedCursors = useRef(new Set<string>());
  const load = useCallback(
    async (cursor: string | null) => {
      if (!enabled || active.current) return;
      const controller = new AbortController();
      active.current = controller;
      setState((previous) => ({
        ...(previous.key === key
          ? previous
          : { key, candidates: [], nextCursor: null, hasMore: false }),
        phase: "loading",
        error: null,
      }));
      try {
        const result = await boundedRead(
          (signal) => port.page(entityType, cursor, signal),
          controller.signal,
        );
        if (
          controller.signal.aborted ||
          current.current.key !== key ||
          current.current.port !== port
        )
          return;
        if (result.kind !== "accepted")
          throw new Error(
            result.kind === "rejected"
              ? result.failure.message
              : "Target read interrupted.",
          );
        if (
          result.value.nextCursor &&
          loadedCursors.current.has(result.value.nextCursor)
        )
          throw new Error("Target paging changed. Reload targets.");
        if (cursor) loadedCursors.current.add(cursor);
        setState((previous) => {
          const candidates = new Map(
            (cursor === null ? [] : previous.candidates).map((candidate) => [
              candidate.recordId,
              candidate,
            ]),
          );
          for (const candidate of result.value.candidates)
            candidates.set(candidate.recordId, candidate);
          return {
            key,
            phase: "ready",
            candidates: [...candidates.values()],
            nextCursor: result.value.nextCursor,
            hasMore: result.value.hasMore,
            error: null,
          };
        });
      } catch (error) {
        if (!controller.signal.aborted && current.current.key === key)
          setState((previous) => ({
            ...previous,
            phase: "failed",
            error:
              error instanceof Error
                ? error.message
                : "Targets could not be loaded.",
          }));
      } finally {
        if (active.current === controller) active.current = null;
      }
    },
    [enabled, entityType, key, port],
  );
  useEffect(() => {
    active.current?.abort();
    active.current = null;
    loadedCursors.current.clear();
    setState({
      key,
      phase: "idle",
      candidates: [],
      nextCursor: null,
      hasMore: false,
      error: null,
    });
    if (enabled) void load(null);
    return () => {
      active.current?.abort();
      active.current = null;
    };
  }, [enabled, key, load]);
  const visible =
    state.key === key
      ? state
      : {
          key,
          phase: "idle" as const,
          candidates: [],
          nextCursor: null,
          hasMore: false,
          error: null,
        };
  return {
    ...visible,
    loadMore: () => load(visible.nextCursor),
    retry: () => load(visible.nextCursor),
    reload: () => {
      loadedCursors.current.clear();
      return load(null);
    },
  };
}
