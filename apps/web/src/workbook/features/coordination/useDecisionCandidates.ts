import { useCallback, useEffect, useRef, useState } from "react";
import { boundedRead } from "../../../services/asyncObservation";
import { workbookOperationFailureIsAccessLoss } from "../../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { DecisionSupersessionOwnerPort } from "./decisionSupersessionOperation";

type Candidates = {
  rows: readonly WorkbookQueryRow[];
  state: "loading" | "ready" | "failed" | "access_lost";
  hasMore: boolean;
  nextCursor: string | null;
  error: string | null;
};
const initial: Candidates = {
  rows: [],
  state: "loading",
  hasMore: true,
  nextCursor: null,
  error: null,
};
export function useDecisionCandidates(
  owner: DecisionSupersessionOwnerPort,
  lifecycleKey: string,
) {
  const [value, setValue] = useState<Candidates>(initial);
  const current = useRef(value);
  const request = useRef<AbortController | null>(null);
  const lastCursor = useRef<string | null>(null);
  const visited = useRef(new Set<string>());
  const load = useCallback(
    async (cursor: string | null) => {
      if (request.current) return;
      const authority = owner.getSnapshot();
      if (!authority.authority) return;
      const controller = new AbortController();
      request.current = controller;
      lastCursor.current = cursor;
      setValue((state) => ({ ...state, state: "loading", error: null }));
      try {
        const result = await boundedRead(
          (signal) => owner.page(cursor, signal),
          controller.signal,
        );
        if (
          controller.signal.aborted ||
          authority.generation !== owner.getSnapshot().generation ||
          result.kind === "aborted"
        )
          return;
        if (result.kind === "rejected") {
          setValue((state) =>
            workbookOperationFailureIsAccessLoss(result.failure)
              ? {
                  ...initial,
                  rows: [],
                  state: "access_lost",
                  error: "Decision access is unavailable.",
                }
              : { ...state, state: "failed", error: result.failure.message },
          );
          return;
        }
        if (cursor === null) visited.current.clear();
        if (cursor !== null) visited.current.add(cursor);
        if (
          result.value.hasMore &&
          result.value.nextCursor &&
          visited.current.has(result.value.nextCursor)
        )
          throw new Error("Repeated candidate cursor");
        const rows = result.value.rows.flatMap((row) => {
          const accepted = owner.acceptRow(row);
          return accepted ? [accepted] : [];
        });
        setValue((state) => ({
          rows: [
            ...new Map(
              [...(cursor === null ? [] : state.rows), ...rows].map((row) => [
                row.record_id,
                row,
              ]),
            ).values(),
          ],
          state: "ready",
          error: null,
          hasMore: result.value.hasMore,
          nextCursor: result.value.nextCursor,
        }));
      } catch {
        if (!controller.signal.aborted)
          setValue((state) => ({
            ...state,
            state: "failed",
            error: "Decision candidates could not be loaded. Retry the read.",
          }));
      } finally {
        if (request.current === controller) request.current = null;
      }
    },
    [owner],
  );
  current.current = value;
  useEffect(() => {
    void lifecycleKey;
    setValue(initial);
    void load(null);
    return () => {
      request.current?.abort();
      request.current = null;
    };
  }, [lifecycleKey, load]);
  return {
    ...value,
    retry: () => load(lastCursor.current),
    refresh: () => load(null),
    more: () =>
      current.current.hasMore && current.current.nextCursor
        ? load(current.current.nextCursor)
        : Promise.resolve(),
  };
}
