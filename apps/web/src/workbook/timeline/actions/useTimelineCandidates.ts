import { useCallback, useEffect, useRef, useState } from "react";
import { boundedRead } from "../../../services/asyncObservation";
import { workbookOperationFailureIsAccessLoss } from "../../ports/WorkbookPortResult";
import type { TimelineCandidatePort } from "./TimelineCandidatePort";
import type { TimelineCaptureSubject } from "./timelineCaptureActionModel";

type Page = Readonly<{
  rows: readonly TimelineCaptureSubject[];
  state: "loading" | "ready" | "failed";
  cursor: string | null;
  nextCursor: string | null;
  error: string | null;
}>;
const empty: Page = {
  rows: [],
  state: "loading",
  cursor: null,
  nextCursor: null,
  error: null,
};

export function useTimelineCandidates(
  port: TimelineCandidatePort,
  scopeKey: string,
  onAccessLost: () => void,
) {
  const [page, setPage] = useState<Page>(empty);
  const request = useRef<AbortController | null>(null);
  const previous = useRef<(string | null)[]>([]);
  const latestScope = useRef(scopeKey);
  latestScope.current = scopeKey;
  const load = useCallback(
    async (cursor: string | null) => {
      request.current?.abort();
      const controller = new AbortController();
      request.current = controller;
      const current = () =>
        !controller.signal.aborted &&
        request.current === controller &&
        latestScope.current === scopeKey;
      setPage((value) => ({ ...value, state: "loading", cursor, error: null }));
      try {
        const result = await boundedRead(
          (signal) => port.page(cursor, signal),
          controller.signal,
        );
        if (!current() || result.kind === "aborted") return;
        if (result.kind === "rejected") {
          if (workbookOperationFailureIsAccessLoss(result.failure)) {
            setPage({ ...empty, state: "failed" });
            onAccessLost();
            return;
          }
          throw new Error("Candidate read failed");
        }
        if (
          result.value.nextCursor !== null &&
          previous.current.includes(result.value.nextCursor)
        )
          throw new Error("Repeated candidate cursor");
        setPage({
          rows: result.value.rows,
          cursor,
          nextCursor: result.value.nextCursor,
          state: "ready",
          error: null,
        });
      } catch {
        if (current())
          setPage((value) => ({
            ...value,
            state: "failed",
            error:
              "Replacements could not be loaded. Retry the read or choose no replacement.",
          }));
      } finally {
        if (request.current === controller) request.current = null;
      }
    },
    [port, scopeKey, onAccessLost],
  );
  useEffect(() => {
    previous.current = [];
    setPage(empty);
    void load(null);
    return () => request.current?.abort();
  }, [load]);
  return {
    ...page,
    canPrevious: previous.current.length > 0,
    retry: () => void load(page.cursor),
    next: () => {
      if (page.state === "ready" && page.nextCursor) {
        previous.current.push(page.cursor);
        void load(page.nextCursor);
      }
    },
    previous: () => {
      if (page.state !== "loading" && previous.current.length)
        void load(previous.current.pop() ?? null);
    },
  };
}
