import { useCallback, useEffect, useRef, useState } from "react";
import { boundedRead } from "../../../services/asyncObservation";
import type { PartyLinkReadPort, PartyPage } from "./partyLinkModel";

type State = PartyPage & {
  key: string;
  revision: number;
  phase: "loading" | "ready" | "failed";
  error: string | null;
};
const empty = { rows: [], nextCursor: null, hasMore: false } as const;
export function usePartyCandidates(
  reader: PartyLinkReadPort,
  key: string,
  revision = 0,
) {
  const [state, setState] = useState<State>({
    ...empty,
    key,
    revision,
    phase: "loading",
    error: null,
  });
  const active = useRef<AbortController | null>(null);
  const current = useRef({ key, revision });
  current.current = { key, revision };
  const cursors = useRef(new Set<string>());
  const failedCursor = useRef<string | null>(null);
  const load = useCallback(
    async (cursor: string | null) => {
      if (active.current) return;
      const controller = new AbortController();
      active.current = controller;
      failedCursor.current = cursor;
      setState((previous) => ({
        ...(previous.key === key ? previous : { ...empty, key }),
        revision,
        phase: "loading",
        error: null,
      }));
      try {
        const page = await boundedRead(
          (signal) => reader.page(cursor, signal),
          controller.signal,
        );
        if (
          controller.signal.aborted ||
          current.current.key !== key ||
          current.current.revision !== revision
        )
          return;
        if (
          page.hasMore !== (page.nextCursor !== null) ||
          (page.nextCursor &&
            (page.nextCursor === cursor ||
              cursors.current.has(page.nextCursor)))
        )
          throw new Error("Party paging changed. Reload Parties.");
        if (cursor) cursors.current.add(cursor);
        setState((previous) => {
          const rows = new Map(
            (cursor ? previous.rows : []).map((row) => [row.record_id, row]),
          );
          for (const row of page.rows)
            if (row.row_version >= (rows.get(row.record_id)?.row_version ?? 0))
              rows.set(row.record_id, row);
          return {
            ...page,
            rows: [...rows.values()],
            key,
            revision,
            phase: "ready",
            error: null,
          };
        });
      } catch {
        if (
          !controller.signal.aborted &&
          current.current.key === key &&
          current.current.revision === revision
        )
          setState((previous) => ({
            ...previous,
            phase: "failed",
            error:
              "Parties could not be refreshed. Previously loaded Parties may be stale.",
          }));
      } finally {
        if (active.current === controller) active.current = null;
      }
    },
    [reader, key, revision],
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
    ...(state.key === key
      ? state.revision === revision
        ? state
        : { ...state, phase: "loading" as const, error: null }
      : { ...empty, key, phase: "loading" as const, error: null }),
    loadMore: () => load(state.nextCursor),
    retry: () => load(failedCursor.current),
    reload: () => {
      cursors.current.clear();
      return load(null);
    },
  };
}
