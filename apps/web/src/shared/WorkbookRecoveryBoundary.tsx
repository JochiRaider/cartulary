import {
  createContext,
  type ReactNode,
  type RefObject,
  useCallback,
  useContext,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import { createPortal } from "react-dom";
import {
  type WorkbookRecoveryItem,
  type WorkbookRecoveryNavigation,
  type WorkbookRecoverySource,
  workbookRecoveryKey,
} from "./workbookRecoveryNavigation";

const subscribeEmpty = () => () => {};
const Context = createContext<{
  navigation: WorkbookRecoveryNavigation;
  detailHost: HTMLElement | null;
  invokerRef?: RefObject<HTMLElement | null>;
} | null>(null);

export function WorkbookRecoveryBoundary({
  navigation,
  detailHost,
  invokerRef,
  children,
}: {
  readonly navigation: WorkbookRecoveryNavigation;
  readonly detailHost: HTMLElement | null;
  readonly invokerRef?: RefObject<HTMLElement | null>;
  readonly children: ReactNode;
}) {
  return (
    <Context.Provider
      value={{ navigation, detailHost, ...(invokerRef ? { invokerRef } : {}) }}
    >
      {children}
    </Context.Provider>
  );
}

export function useWorkbookRecoveryNavigation() {
  return useContext(Context)?.navigation ?? null;
}

/** Captures an explicit shell/grid invoker before attaching recovery. */
export function useWorkbookRecoveryActivation() {
  const context = useContext(Context);
  const navigation = context?.navigation;
  const invokerRef = context?.invokerRef;
  return useCallback(
    (key?: string) => {
      const active = document.activeElement;
      if (
        invokerRef &&
        active instanceof HTMLElement &&
        active !== document.body &&
        !active.closest("#workbook-recovery-panel")
      )
        invokerRef.current = active;
      if (key === undefined) navigation?.openList();
      else navigation?.activate(key);
    },
    [navigation, invokerRef],
  );
}

/** Adapts one owner subscription's safe projection; never subscribes to unrelated owners. */
export function useWorkbookRecoverySource(
  source: string,
  items: readonly WorkbookRecoveryItem[],
  callbacks: WorkbookRecoverySource = {},
) {
  const navigation = useWorkbookRecoveryNavigation();
  const latest = useRef(callbacks);
  latest.current = callbacks;
  const registration = useRef<ReturnType<
    WorkbookRecoveryNavigation["register"]
  > | null>(null);
  useLayoutEffect(() => {
    const detach = latest.current.detach;
    registration.current =
      navigation?.register(source, {
        activate: (id) => latest.current.activate?.(id) ?? true,
        detach: () => detach?.(),
      }) ?? null;
    return () => {
      registration.current?.unregister();
      registration.current = null;
    };
  }, [navigation, source]);
  useLayoutEffect(() => {
    registration.current?.update(items);
  });
  const getSelected = useCallback(() => {
    const snapshot = navigation?.getSnapshot();
    return snapshot?.open
      ? (snapshot.entries.find(
          (entry) => entry.source === source && entry.key === snapshot.selected,
        )?.id ?? null)
      : null;
  }, [navigation, source]);
  return useSyncExternalStore(
    navigation?.subscribe ?? subscribeEmpty,
    getSelected,
  );
}

export function WorkbookRecoveryDetail({
  source,
  item,
  children,
}: {
  readonly source: string;
  readonly item: string | null;
  readonly children: ReactNode;
}) {
  const context = useContext(Context);
  const navigation = context?.navigation;
  const attached = useSyncExternalStore(
    navigation?.subscribe ?? subscribeEmpty,
    () => {
      const snapshot = navigation?.getSnapshot();
      return (
        !!snapshot?.open &&
        item !== null &&
        snapshot.selected === workbookRecoveryKey(source, item)
      );
    },
  );
  if (!context?.detailHost || !attached) return null;
  return createPortal(children, context.detailHost);
}

/** Coordinates an existing inspector/dialog without taking ownership of its state. */
export function useWorkbookSecondaryPanel(open: boolean, detach: () => void) {
  const navigation = useWorkbookRecoveryNavigation();
  const key = useRef(Symbol("secondary panel")).current;
  const current = useRef(detach);
  current.current = detach;
  const coordinated = useRef(false);
  useLayoutEffect(() => {
    if (open) {
      coordinated.current = false;
      navigation?.attachExternal(key, () => {
        coordinated.current = true;
        current.current();
      });
    } else navigation?.detachExternal(key);
    return () => navigation?.detachExternal(key);
  }, [navigation, key, open]);
  return coordinated;
}

/** Bridges explicit owner dialog activation; phase and outcome updates keep attachment unchanged. */
export function useWorkbookRecoveryPresentation(
  source: string,
  presentedId: string | null,
  selected: string | null,
) {
  const context = useContext(Context);
  const navigation = context?.navigation;
  const invokerRef = context?.invokerRef;
  const previous = useRef<string | null>(null);
  useLayoutEffect(() => {
    const before = previous.current;
    previous.current = presentedId;
    if (presentedId !== null && presentedId !== before) {
      const active = document.activeElement;
      if (
        invokerRef &&
        active instanceof HTMLElement &&
        active !== document.body &&
        !active.closest("#workbook-recovery-panel")
      )
        invokerRef.current = active;
      navigation?.activate(workbookRecoveryKey(source, presentedId));
    } else if (presentedId === null && before !== null && selected === before)
      navigation?.close();
  }, [navigation, source, presentedId, selected, invokerRef]);
}
