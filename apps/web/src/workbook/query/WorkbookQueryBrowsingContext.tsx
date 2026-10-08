import type { GridFocusTarget, GridHandle } from "@cartulary/grid-adapter";
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
    recordId?: string;
    fieldKey?: string;
    selected: boolean;
    selectionCommitted: boolean;
    focused: boolean;
    onFocused?: (() => boolean) | undefined;
    onUnavailable?: (() => void) | undefined;
    onSettled?:
      | ((outcome: "ready" | "failed" | "cancelled") => void)
      | undefined;
  } | null = null;
  private focusController: AbortController | null = null;
  private removeNavigationAbort: (() => void) | undefined;
  private focusScheduled = false;
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
    this.resumeNavigation();
  }
  detachPresentation(view: string) {
    this.presentations.get(view)?.detach();
  }
  unbindPresentation(view: string, token: symbol) {
    if (this.presentations.get(view)?.token !== token) return;
    this.presentations.delete(view);
    if (this.navigationFocus?.view === view) this.cancelNavigationFocus();
    this.publish();
  }
  presentationReady(view: string) {
    return this.presentations.get(view)?.ready ?? false;
  }
  navigationPresentationReady(
    view: string,
    page: WorkbookViewQueryAccepted | null,
  ) {
    return (
      this.presentationReady(view) &&
      this.navigationFocus === null &&
      (!page ||
        this.find(view)?.getSnapshot().accepted?.producingRequest ===
          page.producingRequest)
    );
  }
  cancelNavigationFocus(
    outcome: "ready" | "failed" | "cancelled" = "cancelled",
  ) {
    const intent = this.navigationFocus;
    const changed = intent !== null;
    this.navigationFocus = null;
    this.focusController?.abort();
    this.focusController = null;
    this.removeNavigationAbort?.();
    this.removeNavigationAbort = undefined;
    intent?.onSettled?.(outcome);
    if (changed) this.publish();
  }
  // Read acceptance, portal/layout commit and semantic registration are separate signals.
  resumeNavigation = () => {
    if (this.focusScheduled || !this.navigationFocus) return;
    this.focusScheduled = true;
    queueMicrotask(() => {
      this.focusScheduled = false;
      this.attachNavigationFocus();
    });
  };
  private attachNavigationFocus() {
    const intent = this.navigationFocus;
    if (!intent) return;
    if (
      this.find(intent.view)?.getSnapshot().accepted?.producingRequest !==
      intent.page.producingRequest
    ) {
      if (intent.selected || intent.focused || this.focusController)
        this.cancelNavigationFocus();
      return;
    }
    if (intent.focused) {
      if (intent.onFocused?.() !== false && this.navigationFocus === intent)
        this.cancelNavigationFocus("ready");
      return;
    }
    if (this.focusController || !this.presentationReady(intent.view)) return;
    const grid = this.grid(intent.view);
    if (!grid?.presentation) return;
    if (intent.recordId) grid.revealRecord?.(intent.recordId);
    const snapshot = grid.presentation.getSnapshot();
    if (!snapshot) return;
    const rowIdentity = intent.recordId
      ? snapshot.rowIdentities.find(
          (row) =>
            row.kind === "core_record" && row.recordId === intent.recordId,
        )
      : snapshot.rowIdentities.find((row) => row.kind === "core_record");
    if (intent.recordId && !rowIdentity) {
      // Reveal can publish a new presentation on the next commit. A disappeared
      // accepted record, however, is terminal rather than an unrelated-row success.
      if (
        !this.find(intent.view)
          ?.getSnapshot()
          .accepted?.rows.some((row) => row.record_id === intent.recordId)
      ) {
        this.cancelNavigationFocus("failed");
        intent.onUnavailable?.();
      }
      return;
    }
    if (rowIdentity?.kind === "core_record" && !intent.selected) {
      const select = this.navigationSelections.get(intent.view);
      if (!select) return;
      intent.selected = true;
      select(rowIdentity.recordId, () => {
        if (this.navigationFocus !== intent) return;
        intent.selectionCommitted = true;
        this.resumeNavigation();
      });
      return;
    }
    if (intent.selected && !intent.selectionCommitted) return;
    const fields = [
      ...new Set([
        ...(intent.fieldKey && snapshot.fieldKeys.includes(intent.fieldKey)
          ? [intent.fieldKey]
          : []),
        ...snapshot.fieldKeys,
      ]),
    ];
    const targets: GridFocusTarget[] = [
      ...(rowIdentity
        ? fields.map((fieldKey) => ({
            kind: "cell" as const,
            anchor: { surface: snapshot.surface, rowIdentity, fieldKey },
          }))
        : []),
      { kind: "root" },
    ];
    const controller = new AbortController();
    this.focusController = controller;
    void (async () => {
      for (const target of targets) {
        if (controller.signal.aborted || this.navigationFocus !== intent)
          return;
        const result = await grid.requestFocus(target, {
          signal: controller.signal,
        });
        if (controller.signal.aborted || this.navigationFocus !== intent)
          return;
        if (result === "cancelled") {
          this.cancelNavigationFocus();
          return;
        }
        if (result === "focused") {
          this.focusController = null;
          intent.focused = true;
          this.resumeNavigation();
          return;
        }
      }
      if (this.navigationFocus === intent && !controller.signal.aborted) {
        this.cancelNavigationFocus("failed");
        intent.onUnavailable?.();
      }
    })();
  }
  refreshGridBinding(view: string) {
    const grid = this.grid(view);
    const previous = this.gridSubscriptions.get(view);
    if (previous?.handle !== grid) {
      previous?.unsubscribe();
      this.gridSubscriptions.delete(view);
      if (this.navigationFocus?.view === view) {
        this.focusController?.abort();
        this.focusController = null;
      }
      if (grid?.presentation)
        this.gridSubscriptions.set(view, {
          handle: grid,
          unsubscribe: grid.presentation.subscribe(this.resumeNavigation),
        });
    }
    this.resumeNavigation();
  }
  private readers = new Map<
    string,
    { binding: QueryBinding; read: () => Promise<void> }
  >();
  private reverters = new Map<string, () => void>();
  private navigationSelections = new Map<
    string,
    (recordId: string, committed: () => void) => void
  >();
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
    onFocused?: (() => boolean) | undefined,
    handoff?: {
      readonly signal: AbortSignal;
      readonly navigationOnly: boolean;
      readonly onUnavailable: () => void;
      readonly onSettled?: (outcome: "ready" | "failed" | "cancelled") => void;
    },
  ) {
    this.cancelNavigationFocus();
    if (anchor || handoff?.navigationOnly)
      this.navigationFocus = {
        view: page.viewSchemaId,
        page,
        ...anchor,
        selected: false,
        selectionCommitted: false,
        focused: false,
        onFocused,
        onUnavailable: handoff?.onUnavailable,
        onSettled: handoff?.onSettled,
      };
    if (handoff) {
      const abort = () => this.cancelNavigationFocus();
      handoff.signal.addEventListener("abort", abort, { once: true });
      this.removeNavigationAbort = () =>
        handoff.signal.removeEventListener("abort", abort);
      if (handoff.signal.aborted) abort();
    }
    this.navigationPage = { page, query };
    const entry = this.entries.get(page.viewSchemaId);
    if (entry?.token && entry.query === query) {
      entry.browser.adoptNavigation(page);
      this.navigationPage = null;
      this.resumeNavigation();
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
    selectRecord?: (recordId: string, committed: () => void) => void,
  ) {
    this.grids.set(view, grid);
    if (selectRecord) this.navigationSelections.set(view, selectRecord);
    this.refreshGridBinding(view);
    return () => {
      if (this.grids.get(view) === grid) {
        this.grids.delete(view);
        this.navigationSelections.delete(view);
        this.gridSubscriptions.get(view)?.unsubscribe();
        this.gridSubscriptions.delete(view);
        if (this.navigationFocus?.view === view) this.cancelNavigationFocus();
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
    select(recordId, () => {});
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
    this.resumeNavigation();
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
