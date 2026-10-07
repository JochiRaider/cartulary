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
  private revision = 0;
  private navigationPage: {
    page: WorkbookViewQueryAccepted;
    query: WorkbookViewQueryPort["query"];
  } | null = null;
  private navigationFocus: {
    view: string;
    page: WorkbookViewQueryAccepted;
    recordId: string;
    fieldKey?: string;
    onFocused?: (() => void) | undefined;
  } | null = null;
  private focusController: AbortController | null = null;
  private focusSubscription: (() => void) | undefined;
  cancelNavigationFocus() {
    this.navigationFocus = null;
    this.focusController?.abort();
    this.focusController = null;
    this.focusSubscription?.();
    this.focusSubscription = undefined;
  }
  private attachNavigationFocus(view: string) {
    this.focusSubscription?.();
    this.focusSubscription = undefined;
    const grid = this.grid(view);
    if (!grid?.presentation || this.navigationFocus?.view !== view) return;
    const focus = () => {
      const intent = this.navigationFocus;
      if (!intent || intent.view !== view) return;
      // A matching old row is not evidence that the destination owner accepted this read.
      if (
        this.find(view)?.getSnapshot().accepted?.producingRequest !==
        intent.page.producingRequest
      )
        return;
      grid.revealRecord?.(intent.recordId);
      const snapshot = grid.presentation?.getSnapshot();
      const rowIdentity = snapshot?.rowIdentities.find(
        (row) => row.kind === "core_record" && row.recordId === intent.recordId,
      );
      const fieldKey =
        intent.fieldKey && snapshot?.fieldKeys.includes(intent.fieldKey)
          ? intent.fieldKey
          : snapshot?.fieldKeys[0];
      if (!snapshot || !rowIdentity || !fieldKey) return;
      this.navigationFocus = null;
      this.navigationSelections.get(view)?.(intent.recordId);
      this.focusSubscription?.();
      this.focusSubscription = undefined;
      const controller = new AbortController();
      this.focusController = controller;
      void grid
        .requestFocus(
          {
            kind: "cell",
            anchor: { surface: snapshot.surface, rowIdentity, fieldKey },
          },
          { signal: controller.signal },
        )
        .then((result) => {
          if (result === "focused" && !controller.signal.aborted)
            intent.onFocused?.();
        })
        .finally(() => {
          if (this.focusController === controller) this.focusController = null;
        });
    };
    this.focusSubscription = grid.presentation.subscribe(focus);
    focus();
  }
  private readers = new Map<
    string,
    { binding: QueryBinding; read: () => Promise<void> }
  >();
  private reverters = new Map<string, () => void>();
  private navigationSelections = new Map<string, (recordId: string) => void>();
  private grids = new Map<string, { current: GridHandle | null }>();

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
      browser.adoptNavigation(this.navigationPage.page);
      this.navigationPage = null;
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
    anchor: { recordId: string; fieldKey?: string } | null = null,
    onFocused?: (() => void) | undefined,
  ) {
    this.cancelNavigationFocus();
    if (anchor)
      this.navigationFocus = {
        view: page.viewSchemaId,
        page,
        ...anchor,
        onFocused,
      };
    this.navigationPage = { page, query };
    const entry = this.entries.get(page.viewSchemaId);
    if (entry?.token && entry.query === query) {
      entry.browser.adoptNavigation(page);
      this.navigationPage = null;
      queueMicrotask(() => this.attachNavigationFocus(page.viewSchemaId));
    }
  }
  checkpoint(
    view: string,
    query: import("../models/workbookQuery").WorkbookQueryState,
  ) {
    return this.entries.get(view)?.browser.checkpointFor(query) ?? null;
  }
  clearNavigation() {
    this.navigationPage = null;
    this.cancelNavigationFocus();
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
    selectRecord?: (recordId: string) => void,
  ) {
    this.grids.set(view, grid);
    if (selectRecord) this.navigationSelections.set(view, selectRecord);
    this.attachNavigationFocus(view);
    return () => {
      if (this.grids.get(view) === grid) {
        this.grids.delete(view);
        this.navigationSelections.delete(view);
      }
    };
  }
  selectLoadedRecord(view: string, recordId: string) {
    const select = this.navigationSelections.get(view);
    if (
      !select ||
      !this.grid(view)
        ?.presentation?.getSnapshot()
        ?.rowIdentities.some(
          (row) => row.kind === "core_record" && row.recordId === recordId,
        )
    )
      return false;
    select(recordId);
    return true;
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
  getSnapshot = () => this.revision;
  private publish = () => {
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
    this.navigationSelections.clear();
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
