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

export function useAppRouteRuntime(): AppRouteRuntime {
  const [route, setRoute] = useState<AppRouteState>(() => readAppRouteState());
  const routeRef = useRef(route);
  routeRef.current = route;

  const commitRoute = useCallback(
    (next: AppRouteState, mode: AppRouteWriteMode) => {
      writeAppRouteState(next, mode);
      startTransition(() => {
        setRoute(next);
      });
    },
    [],
  );

  useEffect(() => {
    const handlePopState = () => {
      setRoute(readAppRouteState());
    };
    window.addEventListener("popstate", handlePopState);
    return () => {
      window.removeEventListener("popstate", handlePopState);
    };
  }, []);

  return {
    commitRoute,
    route,
    routeRef,
  };
}
