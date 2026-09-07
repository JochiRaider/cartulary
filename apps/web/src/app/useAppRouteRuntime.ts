import {
  type MutableRefObject,
  startTransition,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import {
  type AppRouteState,
  type AppRouteWriteMode,
  readAppRouteState,
  writeAppRouteState,
} from "./routeState";
export type AppRouteRuntime = {
  commitRoute: (next: AppRouteState, mode: AppRouteWriteMode) => void;
  route: AppRouteState;
  routeRef: MutableRefObject<AppRouteState>;
};
type LeavePolicy = {
  hasPendingEdits: () => boolean;
  requestLeave: () => Promise<boolean>;
  beforeCommit: (next: AppRouteState) => void;
};
const indexKey = "cartularyRouteIndex";
const record = (value: unknown): Record<string, unknown> =>
  typeof value === "object" && value !== null
    ? (value as Record<string, unknown>)
    : {};
const readIndex = () => {
  const value = record(window.history.state)[indexKey];
  return typeof value === "number" && Number.isSafeInteger(value)
    ? value
    : null;
};

/** The application route owner guards only the single deployment-user draft. */
export function useAppRouteRuntime(policy?: LeavePolicy): AppRouteRuntime {
  const [route, setRoute] = useState<AppRouteState>(() => readAppRouteState());
  const routeRef = useRef(route);
  const policyRef = useRef(policy);
  policyRef.current = policy;
  const index = useRef(readIndex() ?? 0);
  const request = useRef(0);
  const pending = useRef(false);
  const entryURL = useRef(
    `${window.location.pathname}${window.location.search}${window.location.hash}`,
  );
  const publish = useCallback((next: AppRouteState) => {
    policyRef.current?.beforeCommit(next);
    routeRef.current = next;
    entryURL.current = `${window.location.pathname}${window.location.search}${window.location.hash}`;
    startTransition(() => setRoute(next));
  }, []);
  const write = useCallback(
    (next: AppRouteState, mode: AppRouteWriteMode) => {
      if (mode === "push") ++index.current;
      writeAppRouteState(next, mode, {
        history: {
          pushState: (data, unused, url) =>
            window.history.pushState(
              { ...record(data), [indexKey]: index.current },
              unused,
              url,
            ),
          replaceState: (data, unused, url) =>
            window.history.replaceState(
              { ...record(data), [indexKey]: index.current },
              unused,
              url,
            ),
        },
      });
      publish(next);
    },
    [publish],
  );
  const commitRoute = useCallback(
    (next: AppRouteState, mode: AppRouteWriteMode) => {
      if (!policyRef.current?.hasPendingEdits()) {
        ++request.current;
        pending.current = false;
        write(next, mode);
        return;
      }
      if (pending.current) return;
      pending.current = true;
      const token = ++request.current;
      void policyRef.current.requestLeave().then((accepted) => {
        if (token !== request.current) return;
        pending.current = false;
        if (accepted) write(next, mode);
      });
    },
    [write],
  );
  useEffect(() => {
    window.history.replaceState(
      { ...record(window.history.state), [indexKey]: index.current },
      "",
      window.location.href,
    );
    let restoration: {
      oldIndex: number;
      targetIndex: number;
      token: number;
    } | null = null;
    let approvedIndex: number | null = null;
    const askForHistoryLeave = (intent: {
      oldIndex: number;
      targetIndex: number;
      token: number;
    }) => {
      if (intent.token !== request.current) return;
      void (policyRef.current?.requestLeave() ?? Promise.resolve(true)).then(
        (accepted) => {
          if (intent.token !== request.current) return;
          pending.current = false;
          if (accepted) {
            approvedIndex = intent.targetIndex;
            window.history.go(intent.targetIndex - intent.oldIndex);
          }
        },
      );
    };
    const handlePopState = () => {
      const targetIndex = readIndex();
      if (restoration) {
        const intent = restoration;
        if (targetIndex !== intent.oldIndex && targetIndex !== null) {
          window.history.go(intent.oldIndex - targetIndex);
          return;
        }
        restoration = null;
        askForHistoryLeave(intent);
        return;
      }
      const next = readAppRouteState();
      if (targetIndex !== null && approvedIndex === targetIndex) {
        approvedIndex = null;
        index.current = targetIndex;
        publish(next);
        return;
      }
      if (!policyRef.current?.hasPendingEdits()) {
        ++request.current;
        pending.current = false;
        if (targetIndex !== null) index.current = targetIndex;
        publish(next);
        return;
      }
      if (pending.current) {
        if (targetIndex !== null && targetIndex !== index.current)
          window.history.go(index.current - targetIndex);
        return;
      }
      pending.current = true;
      const token = ++request.current;
      if (targetIndex !== null && targetIndex !== index.current) {
        restoration = { oldIndex: index.current, targetIndex, token };
        // Restore the accepted entry before presenting the guard; approval replays traversal once.
        window.history.go(index.current - targetIndex);
        return;
      }
      // A synthetic/unowned same-document entry has no traversal distance. Keep
      // the accepted presentation until the user resolves it, without inferring direction.
      const acceptedURL = entryURL.current;
      void policyRef.current.requestLeave().then((accepted) => {
        if (token !== request.current) return;
        pending.current = false;
        if (accepted) publish(next);
        else
          window.history.replaceState(
            { ...record(window.history.state), [indexKey]: index.current },
            "",
            acceptedURL,
          );
      });
    };
    window.addEventListener("popstate", handlePopState);
    return () => {
      ++request.current;
      pending.current = false;
      window.removeEventListener("popstate", handlePopState);
    };
  }, [publish]);
  return { commitRoute, route, routeRef };
}
