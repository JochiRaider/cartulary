import type { GridHandle } from "@cartulary/grid-adapter";
import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useSyncExternalStore,
} from "react";
import { WorkbookQueryBrowser } from "./WorkbookQueryBrowser";
import type { WorkbookViewQueryPort } from "./WorkbookViewQueryPort";

class WorkbookQueryBrowsingRegistry {
  private entries = new Map<
    string,
    { port: WorkbookViewQueryPort; browser: WorkbookQueryBrowser }
  >();
  private listeners = new Set<() => void>();
  private revision = 0;
  private readers = new Map<string, () => Promise<void>>();
  private reverters = new Map<string, () => void>();
  private grids = new Map<string, { current: GridHandle | null }>();
  bindRead(view: string, read: () => Promise<void>) {
    this.readers.set(view, read);
    return () => {
      if (this.readers.get(view) === read) this.readers.delete(view);
    };
  }
  bindRevert(view: string, revert: () => void) {
    this.reverters.set(view, revert);
    return () => {
      if (this.reverters.get(view) === revert) this.reverters.delete(view);
    };
  }
  bindGrid(view: string, grid: { current: GridHandle | null }) {
    this.grids.set(view, grid);
    return () => {
      if (this.grids.get(view) === grid) this.grids.delete(view);
    };
  }
  grid(view: string) {
    return this.grids.get(view)?.current ?? null;
  }
  prepareBrowse(view: string) {
    const grid = this.grids.get(view)?.current;
    const anchor = grid?.getActiveCell?.();
    if (anchor?.rowIdentity.kind === "core_record")
      this.find(view)?.rememberAnchor(anchor.rowIdentity.recordId);
    grid?.detachEdit?.();
  }
  async activate(
    view: string,
    action: Parameters<WorkbookQueryBrowser["activate"]>[0],
  ) {
    const read = this.readers.get(view),
      browser = this.find(view);
    if (!read || !browser) return;
    this.prepareBrowse(view);
    await browser.activate(action, read);
  }
  revert(view: string) {
    this.reverters.get(view)?.();
  }
  invalidateAll() {
    for (const { browser } of this.entries.values()) browser.invalidate();
  }
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.revision;
  find = (viewSchemaId: string) => this.entries.get(viewSchemaId)?.browser;
  get(port: WorkbookViewQueryPort, viewSchemaId: string) {
    const existing = this.entries.get(viewSchemaId);
    if (existing?.port.query === port.query) return existing.browser;
    existing?.browser.invalidate();
    const browser = new WorkbookQueryBrowser(port, viewSchemaId);
    browser.subscribe(() => {
      this.revision += 1;
      for (const listener of this.listeners) listener();
    });
    this.entries.set(viewSchemaId, { port, browser });
    return browser;
  }
  dispose() {
    for (const { browser } of this.entries.values()) browser.invalidate();
  }
}

const Context = createContext<WorkbookQueryBrowsingRegistry | null>(null);
const noSubscribe = () => () => {};
const noSnapshot = () => 0;

export function WorkbookQueryBrowsingProvider({
  children,
}: {
  readonly children: ReactNode;
}) {
  const registry = useMemo(() => new WorkbookQueryBrowsingRegistry(), []);
  useEffect(() => () => registry.dispose(), [registry]);
  return <Context.Provider value={registry}>{children}</Context.Provider>;
}

export function useWorkbookQueryPresentation() {
  const registry = useContext(Context);
  useSyncExternalStore(
    registry?.subscribe ?? noSubscribe,
    registry?.getSnapshot ?? noSnapshot,
  );
  return registry;
}

export function useWorkbookBrowsingRegistry() {
  return useContext(Context);
}

export function useWorkbookBrowsingRead(
  view: string,
  read: () => Promise<void>,
  active = true,
) {
  const registry = useContext(Context);
  useEffect(
    () => (active ? registry?.bindRead(view, read) : undefined),
    [registry, view, read, active],
  );
}

export function useWorkbookQueryBrowser(
  port: WorkbookViewQueryPort,
  viewSchemaId: string,
  active = true,
) {
  const registry = useContext(Context);
  const query = port.query;
  const browser = useMemo(
    () =>
      (active ? registry?.get({ query }, viewSchemaId) : undefined) ??
      new WorkbookQueryBrowser({ query }, viewSchemaId),
    [query, registry, viewSchemaId, active],
  );
  const snapshot = useSyncExternalStore(browser.subscribe, browser.getSnapshot);
  useEffect(() => () => browser.detach(), [browser]);
  return { browser, snapshot };
}

/** An explicit UI refresh always retires continuation and starts cursor-free. */
export function useWorkbookQueryRestart(
  view: string,
  fallback: () => void | Promise<void>,
) {
  const registry = useWorkbookBrowsingRegistry();
  return useCallback(async () => {
    if (registry?.find(view)) await registry.activate(view, "restart");
    else await fallback();
  }, [registry, view, fallback]);
}
