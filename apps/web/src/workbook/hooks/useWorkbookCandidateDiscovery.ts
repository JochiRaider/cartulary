import {
  createContext,
  useContext,
  useLayoutEffect,
  useMemo,
  useRef,
  useSyncExternalStore,
} from "react";
import type { WorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type {
  WorkbookCandidate,
  WorkbookCandidateReader,
} from "../ports/WorkbookCandidateReadPort";
import { WorkbookCandidateDiscovery } from "../services/WorkbookCandidateDiscovery";

/** Supplied by the existing shell authority owner, including recovery attachments. */
export const WorkbookCandidateAuthorityContext = createContext({
  identity: "standalone",
  canRead: true,
  onAuthorityFailure: (_failure: WorkbookOperationFailure) => {},
});

export function useWorkbookCandidateDiscovery<T extends WorkbookCandidate>(
  read: WorkbookCandidateReader<T>,
  queryState: WorkbookQueryState,
  target: string,
  revision: string | number,
  enabled = true,
) {
  const authority = useContext(WorkbookCandidateAuthorityContext);
  const key = JSON.stringify([
    authority.identity,
    target,
    queryState,
    revision,
    enabled,
    authority.canRead,
  ]);
  const current = useRef<WorkbookCandidateDiscovery<T> | null>(null);
  const latest = useRef(authority);
  latest.current = authority;
  const admitted = useRef(enabled && authority.canRead);
  admitted.current = enabled && authority.canRead;
  const controller = useMemo(() => {
    const instance: WorkbookCandidateDiscovery<T> =
      new WorkbookCandidateDiscovery({
        scope: key,
        queryState,
        read,
        isCurrent: () => current.current === instance && admitted.current,
        onAuthorityFailure: (failure) =>
          latest.current.onAuthorityFailure(failure),
      });
    return instance;
  }, [key, queryState, read]);
  current.current = controller;
  const snapshot = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  useLayoutEffect(() => {
    if (enabled && authority.canRead) void controller.start();
    return () => controller.dispose();
  }, [controller, enabled, authority.canRead]);
  return {
    ...snapshot,
    controller,
    enabled,
    canRead: authority.canRead && !snapshot.concealed,
  };
}
