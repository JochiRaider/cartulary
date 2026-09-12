import { useCallback, useEffect, useRef, useState } from "react";
import { boundedRead } from "../../../services/asyncObservation";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";
import type {
  AssessmentCandidatePage,
  AssessmentCandidateQuery,
} from "./assessmentCandidatePort";

type State = AssessmentCandidatePage & {
  readonly key: string;
  readonly phase: "loading" | "ready" | "failed";
  readonly error: string | null;
  readonly loaded: boolean;
};
const empty = {
  candidates: [],
  hasMore: false,
  nextCursor: null,
  loaded: false,
} as const;
export function useAssessmentCandidates(
  read: (
    input: AssessmentCandidateQuery,
  ) => Promise<WorkbookPortResult<AssessmentCandidatePage>>,
  queryState: WorkbookQueryState,
  revision: string | number,
  enabled = true,
) {
  const key = JSON.stringify([queryState, revision]);
  const current = useRef(key);
  current.current = key;
  const [state, setState] = useState<State>({
    ...empty,
    key,
    phase: "loading",
    error: null,
  });
  const active = useRef<AbortController | null>(null);
  const failedCursor = useRef<string | null>(null);
  const cursors = useRef(new Set<string>());
  const load = useCallback(
    async (cursor: string | null) => {
      if (!enabled || active.current) return;
      const controller = new AbortController();
      active.current = controller;
      failedCursor.current = cursor;
      setState((prior) => ({ ...prior, phase: "loading", error: null }));
      try {
        const result = await boundedRead(
          (signal) => read({ queryState, cursor, signal }),
          controller.signal,
        );
        if (
          controller.signal.aborted ||
          current.current !== key ||
          result.kind === "aborted"
        )
          return;
        if (result.kind === "rejected") throw new Error(result.failure.message);
        const page = result.value;
        if (
          page.hasMore !== (page.nextCursor !== null) ||
          (page.nextCursor &&
            (page.nextCursor === cursor ||
              cursors.current.has(page.nextCursor)))
        )
          throw new Error("Candidate paging changed. Reload candidates.");
        if (cursor) cursors.current.add(cursor);
        setState((prior) => {
          const rows = new Map(
            (cursor ? prior.candidates : []).map((candidate) => [
              candidate.recordId,
              candidate,
            ]),
          );
          for (const row of page.candidates) rows.set(row.recordId, row);
          return {
            ...page,
            candidates: [...rows.values()],
            key,
            loaded: true,
            phase: "ready",
            error: null,
          };
        });
      } catch (error) {
        if (!controller.signal.aborted && current.current === key)
          setState((prior) => ({
            ...prior,
            key,
            phase: "failed",
            error:
              error instanceof Error
                ? error.message
                : "Candidates could not be loaded.",
          }));
      } finally {
        if (active.current === controller) active.current = null;
      }
    },
    [enabled, key, queryState, read],
  );
  useEffect(() => {
    active.current?.abort();
    active.current = null;
    cursors.current.clear();
    void load(null);
    return () => {
      active.current?.abort();
      active.current = null;
    };
  }, [load]);
  return {
    ...state,
    stale: state.loaded && (state.key !== key || state.phase !== "ready"),
    phase: state.key !== key ? ("loading" as const) : state.phase,
    reload: () => {
      cursors.current.clear();
      return load(null);
    },
    retry: () => load(failedCursor.current),
    loadMore: () =>
      state.nextCursor === null ? Promise.resolve() : load(state.nextCursor),
  };
}
