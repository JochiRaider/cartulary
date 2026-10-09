import type { GridHandle } from "@cartulary/grid-adapter";
import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useSyncExternalStore,
} from "react";
import {
  emptyWorkbookBrowsingSnapshot,
  type WorkbookQueryAcceptance,
  WorkbookQueryBrowser,
} from "./WorkbookQueryBrowser";
import type {
  WorkbookViewQueryAccepted,
  WorkbookViewQueryPort,
} from "./WorkbookViewQueryPort";

type QueryBinding = {
  readonly viewSchemaId: string;
  readonly query: WorkbookViewQueryPort["query"];
  readonly token: symbol;
  readonly active: boolean;
  readonly currentBrowser: () => WorkbookQueryBrowser | undefined;
};
type BrowsingEntry = {
  readonly query: WorkbookViewQueryPort["query"];
  readonly browser: WorkbookQueryBrowser;
  token: symbol | null;
  unsubscribe: (() => void) | undefined;
};

class WorkbookQueryBrowsingRegistry {
  private entries = new Map<string, BrowsingEntry>();
  private listeners = new Set<() => void>();
  private presentationListeners = new Set<() => void>();
  private revision = 0;
  private navigationPage: {
    page: WorkbookViewQueryAccepted;
    query: WorkbookViewQueryPort["query"];
    adopted: WorkbookQueryBrowser | null;
    rejected: boolean;
  } | null = null;
  private releaseNavigation: (() => void) | null = null;
  private gridSubscriptions = new Map<
    string,
    { handle: GridHandle; unsubscribe: () => void }
  >();
  private presentations = new Map<
    string,
    { token: symbol; ready: boolean; detach: () => void }
  >();

  updatePresentation(
    view: string,
    token: symbol,
    ready: boolean,
    detach: () => void,
  ) {
    const previous = this.presentations.get(view);
    this.presentations.set(view, { token, ready, detach });
    if (previous?.ready !== ready || previous.token !== token) this.publish();
  }
  detachPresentation(view: string) {
    this.presentations.get(view)?.detach();
  }
  unbindPresentation(view: string, token: symbol) {
    if (this.presentations.get(view)?.token !== token) return;
    this.presentations.delete(view);
    this.publish();
  }
  presentationReady(view: string) {
    return this.presentations.get(view)?.ready ?? false;
  }
  presentationState(view: string) {
    return this.presentations.get(view);
  }
  selectionBinding(view: string) {
    return this.selections.get(view);
  }
  refreshGridBinding(view: string) {
    const grid = this.grid(view);
    const previous = this.gridSubscriptions.get(view);
    if (previous?.handle !== grid) {
      previous?.unsubscribe();
      this.gridSubscriptions.delete(view);
      if (grid?.presentation)
        this.gridSubscriptions.set(view, {
          handle: grid,
          unsubscribe: grid.presentation.subscribe(this.notifyPresentation),
        });
    }
    this.notifyPresentation();
  }
  private readers = new Map<
    string,
    { binding: QueryBinding; read: () => Promise<void> }
  >();
  private reverters = new Map<string, () => void>();
  private selections = new Map<
    string,
    { token: symbol; select: (recordId: string, committed: () => void) => void }
  >();
  private grids = new Map<string, { ref: { current: GridHandle | null } }>();

  // Render prepares identity only. Ownership and browser construction start at commit.
  prepare(
    query: WorkbookViewQueryPort["query"],
    viewSchemaId: string,
    active: boolean,
  ): QueryBinding {
    const token = Symbol(viewSchemaId);
    return {
      query,
      viewSchemaId,
      token,
      active,
      currentBrowser: () => {
        const entry = this.entries.get(viewSchemaId);
        return active && entry?.token === token ? entry.browser : undefined;
      },
    };
  }
  commit(binding: QueryBinding) {
    if (!binding.active) return;
    const previous = this.entries.get(binding.viewSchemaId);
    previous?.unsubscribe?.();
    this.readers.delete(binding.viewSchemaId);
    let browser: WorkbookQueryBrowser;
    if (previous?.query === binding.query) {
      browser = previous.browser;
      if (previous.token !== null) browser.detach();
    } else {
      previous?.browser.invalidate();
      browser = new WorkbookQueryBrowser(
        { query: binding.query },
        binding.viewSchemaId,
      );
    }
    if (
      this.navigationPage?.page.viewSchemaId === binding.viewSchemaId &&
      this.navigationPage.query === binding.query
    ) {
      this.navigationPage.adopted = browser;
      this.navigationPage.rejected = !browser.adoptNavigation(
        this.navigationPage.page,
      );
    }
    const entry: BrowsingEntry = {
      query: binding.query,
      browser,
      token: binding.token,
      unsubscribe: undefined,
    };
    this.entries.set(binding.viewSchemaId, entry);
    entry.unsubscribe = browser.subscribe(this.publish);
    this.publish();
    return () => {
      if (
        entry.token !== binding.token ||
        this.entries.get(binding.viewSchemaId) !== entry
      )
        return;
      entry.token = null;
      entry.unsubscribe?.();
      entry.unsubscribe = undefined;
      this.readers.delete(binding.viewSchemaId);
      browser.detach();
      this.publish();
    };
  }
  stageNavigation(
    page: WorkbookViewQueryAccepted,
    query: WorkbookViewQueryPort["query"],
    signal: AbortSignal,
  ): WorkbookQueryAcceptance {
    this.releaseNavigation?.();
    const staged = {
      page,
      query,
      adopted: null as WorkbookQueryBrowser | null,
      rejected: false,
    };
    let released = false;
    const release = () => {
      if (released) return;
      released = true;
      signal.removeEventListener("abort", release);
      if (this.navigationPage === staged) this.navigationPage = null;
      if (this.releaseNavigation === release) this.releaseNavigation = null;
      staged.adopted?.discardNavigation(page);
      this.publish();
    };
    this.navigationPage = staged;
    this.releaseNavigation = release;
    signal.addEventListener("abort", release, { once: true });
    const entry = this.entries.get(page.viewSchemaId);
    if (entry?.token && entry.query === query) {
      staged.adopted = entry.browser;
      staged.rejected = !entry.browser.adoptNavigation(page);
    }
    if (signal.aborted) release();
    return {
      state: () => {
        if (released || signal.aborted) return "cancelled";
        if (staged.rejected) return "failed";
        const browser = this.find(page.viewSchemaId);
        if (!browser) return staged.adopted ? "cancelled" : "pending";
        if (staged.adopted && staged.adopted !== browser) return "cancelled";
        return browser.navigationState(page);
      },
      hasRecord: (id) =>
        this.find(page.viewSchemaId)
          ?.getSnapshot()
          .accepted?.rows.some((row) => row.record_id === id) ?? false,
      release,
    };
  }
  checkpoint(
    view: string,
    query: import("../models/workbookQuery").WorkbookQueryState,
  ) {
    return this.entries.get(view)?.browser.checkpointFor(query) ?? null;
  }
  clearNavigation() {
    this.releaseNavigation?.();
  }
  bindRead(binding: QueryBinding, read: () => Promise<void>) {
    if (!binding.currentBrowser()) return;
    const reader = { binding, read };
    this.readers.set(binding.viewSchemaId, reader);
    return () => {
      if (this.readers.get(binding.viewSchemaId) === reader)
        this.readers.delete(binding.viewSchemaId);
    };
  }
  bindRevert(view: string, revert: () => void) {
    this.reverters.set(view, revert);
    return () => {
      if (this.reverters.get(view) === revert) this.reverters.delete(view);
    };
  }
  bindGrid(
    view: string,
    grid: { current: GridHandle | null },
    selectRecord?: (recordId: string, committed: () => void) => void,
  ) {
    const binding = { ref: grid };
    this.grids.set(view, binding);
    if (selectRecord)
      this.selections.set(view, { token: Symbol(view), select: selectRecord });
    this.refreshGridBinding(view);
    return () => {
      if (this.grids.get(view) === binding) {
        this.grids.delete(view);
        this.selections.delete(view);
        this.gridSubscriptions.get(view)?.unsubscribe();
        this.gridSubscriptions.delete(view);
        this.notifyPresentation();
      }
    };
  }
  selectLoadedRecord(view: string, recordId: string) {
    const select = this.selections.get(view)?.select;
    if (
      !select ||
      !this.grid(view)
        ?.presentation?.getSnapshot()
        ?.rowIdentities.some(
          (row) => row.kind === "core_record" && row.recordId === recordId,
        )
    )
      return false;
    select(recordId, () => {});
    return true;
  }
  grid(view: string) {
    return this.grids.get(view)?.ref.current ?? null;
  }
  prepareBrowse(view: string) {
    const grid = this.grids.get(view)?.ref.current;
    const anchor = grid?.getActiveCell?.();
    if (anchor?.rowIdentity.kind === "core_record")
      this.find(view)?.rememberAnchor(anchor.rowIdentity.recordId);
    grid?.detachEdit?.();
  }
  async activate(
    view: string,
    action: Parameters<WorkbookQueryBrowser["activate"]>[0],
  ) {
    const reader = this.readers.get(view);
    const browser = reader?.binding.currentBrowser();
    if (!reader || !browser) return;
    this.prepareBrowse(view);
    await browser.activate(action, async () => {
      if (reader.binding.currentBrowser() === browser) await reader.read();
    });
  }
  revert(view: string) {
    this.reverters.get(view)?.();
  }
  invalidateAll() {
    this.clearNavigation();
    for (const { browser } of this.entries.values()) browser.invalidate();
  }
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  subscribePresentation = (listener: () => void) => {
    this.presentationListeners.add(listener);
    return () => {
      this.presentationListeners.delete(listener);
    };
  };
  private notifyPresentation = () => {
    for (const listener of this.presentationListeners) listener();
  };
  getSnapshot = () => this.revision;
  private publish = () => {
    this.notifyPresentation();
    this.revision += 1;
    for (const listener of this.listeners) listener();
  };
  find = (viewSchemaId: string) => {
    const entry = this.entries.get(viewSchemaId);
    return entry?.token ? entry.browser : undefined;
  };
  dispose() {
    this.clearNavigation();
    for (const entry of this.entries.values()) {
      entry.unsubscribe?.();
      entry.token = null;
      entry.browser.invalidate();
    }
    this.entries.clear();
    this.readers.clear();
    this.reverters.clear();
    this.grids.clear();
    for (const subscription of this.gridSubscriptions.values())
      subscription.unsubscribe();
    this.gridSubscriptions.clear();
    this.presentations.clear();
    this.selections.clear();
    this.publish();
  }
}

const Context = createContext<WorkbookQueryBrowsingRegistry | null>(null);

export function WorkbookQueryBrowsingProvider({
  children,
}: {
  readonly children: ReactNode;
}) {
  const registry = useMemo(() => new WorkbookQueryBrowsingRegistry(), []);
  useEffect(() => () => registry.dispose(), [registry]);
  return <Context.Provider value={registry}>{children}</Context.Provider>;
}

export function useWorkbookBrowsingRegistry() {
  const registry = useContext(Context);
  if (!registry) throw new Error("WorkbookQueryBrowsingProvider is required");
  return registry;
}

export function useWorkbookQueryPresentation() {
  const registry = useWorkbookBrowsingRegistry();
  useSyncExternalStore(registry.subscribe, registry.getSnapshot);
  return registry;
}

export function useWorkbookBrowsingRead(
  binding: QueryBinding,
  read: () => Promise<void>,
) {
  const registry = useWorkbookBrowsingRegistry();
  useLayoutEffect(
    () => registry.bindRead(binding, read),
    [registry, binding, read],
  );
}

export function useWorkbookQueryBrowser(
  port: WorkbookViewQueryPort,
  viewSchemaId: string,
  active = true,
) {
  const registry = useWorkbookBrowsingRegistry();
  const query = port.query;
  const binding = useMemo(
    () => registry.prepare(query, viewSchemaId, active),
    [query, registry, viewSchemaId, active],
  );
  useLayoutEffect(() => registry.commit(binding), [registry, binding]);
  useSyncExternalStore(registry.subscribe, registry.getSnapshot);
  const browser = binding.currentBrowser();
  return {
    binding,
    browser,
    snapshot: browser?.getSnapshot() ?? emptyWorkbookBrowsingSnapshot,
  };
}

/** An explicit UI refresh always retires continuation and starts cursor-free. */
export function useWorkbookQueryRestart(view: string) {
  const registry = useWorkbookBrowsingRegistry();
  return useCallback(
    () => registry.activate(view, "restart"),
    [registry, view],
  );
}
