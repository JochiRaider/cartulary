import type { SheetRef } from "./sheetRef";

export type RecoveryAttention =
  | "attention"
  | "progress"
  | "draft"
  | "completed";

/** Safe navigation metadata only. Payloads, permissions and actions stay with the source. */
export type WorkbookRecoveryItem = {
  readonly id: string;
  readonly label: string;
  readonly summary: string;
  readonly origin: string;
  readonly sheetRef: SheetRef | null;
  readonly attention: RecoveryAttention;
  readonly order: number;
  readonly priority?: number;
  readonly refreshViews?: readonly string[];
  readonly refreshOnlyView?: string;
  readonly conflictKeys?: readonly string[];
  readonly conflictKey?: string;
  readonly operationIds?: readonly string[];
  readonly conflictOperationId?: string;
};
export type WorkbookRecoveryEntry = WorkbookRecoveryItem & {
  readonly key: string;
  readonly source: string;
};
export type WorkbookRecoverySource = {
  /** Attaches existing owner presentation; must not submit, replay or discard. */
  readonly activate?: (id: string) => boolean;
  /** Invalidates presentation/review according to the source owner's policy. */
  readonly detach?: () => void;
};
export type WorkbookRecoverySnapshot = {
  readonly entries: readonly WorkbookRecoveryEntry[];
  readonly count: number;
  readonly open: boolean;
  readonly selected: string | null;
  readonly message: string | null;
  readonly activation: number;
};
type Registration = {
  readonly token: symbol;
  readonly callbacks: WorkbookRecoverySource;
  items: readonly WorkbookRecoveryItem[];
};
const attentionOrder: Record<RecoveryAttention, number> = {
  attention: 0,
  progress: 1,
  draft: 2,
  completed: 3,
};
export function workbookRecoveryKey(source: string, id: string): string {
  return JSON.stringify([source, id]);
}

/** Resolve the parent obligation without importing feature execution owners. */
export function workbookConflictRecoveryKey(
  entries: readonly WorkbookRecoveryEntry[],
  conflict: {
    readonly key: string;
    readonly batchOperationId?: string | undefined;
    readonly compoundOperationId?: string | undefined;
  },
): string {
  return (
    entries.find(
      (entry) =>
        entry.conflictKeys?.includes(conflict.key) ||
        (conflict.compoundOperationId !== undefined &&
          entry.operationIds?.includes(conflict.compoundOperationId)),
    )?.key ??
    (conflict.batchOperationId
      ? workbookRecoveryKey("batch", conflict.batchOperationId)
      : workbookRecoveryKey("core", `conflict:${conflict.key}`))
  );
}

/** One live shell's navigation and attachment state. Never retains owner payloads. */
export class WorkbookRecoveryNavigation {
  private readonly sources = new Map<string, Registration>();
  private readonly listeners = new Set<() => void>();
  private readonly dismissed = new Set<string>();
  private disposed = false;
  private external: { key: symbol; detach: () => void } | null = null;
  private snapshot: WorkbookRecoverySnapshot = {
    entries: [],
    count: 0,
    open: false,
    selected: null,
    message: null,
    activation: 0,
  };
  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    if (this.disposed) return () => {};
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  register(source: string, callbacks: WorkbookRecoverySource) {
    const token = Symbol(source);
    const previous = this.sources.get(source);
    if (previous) this.remove(source, previous.token);
    if (!this.disposed)
      this.sources.set(source, { token, callbacks, items: [] });
    return {
      update: (items: readonly WorkbookRecoveryItem[]) => {
        const registration = this.sources.get(source);
        if (this.disposed || registration?.token !== token) return;
        if (new Set(items.map((item) => item.id)).size !== items.length)
          throw new Error("Duplicate recovery work identity");
        if (JSON.stringify(registration.items) === JSON.stringify(items))
          return;
        registration.items = items;
        this.reconcile();
      },
      unregister: () => this.remove(source, token),
    };
  }
  private remove(source: string, token: symbol) {
    const registration = this.sources.get(source);
    if (!registration || registration.token !== token) return;
    const selected = this.selectedEntry();
    this.sources.delete(source);
    if (selected?.source === source) {
      this.snapshot = {
        ...this.snapshot,
        selected: null,
        message: "This recovery item is no longer available.",
      };
      registration.callbacks.detach?.();
    }
    this.reconcile();
  }
  private selectedEntry() {
    return this.snapshot.entries.find(
      (entry) => entry.key === this.snapshot.selected,
    );
  }
  private detach() {
    const entry = this.selectedEntry();
    this.snapshot = { ...this.snapshot, selected: null };
    if (entry) this.sources.get(entry.source)?.callbacks.detach?.();
  }
  private leaveExternal() {
    const previous = this.external;
    this.external = null;
    previous?.detach();
  }
  /** Used only by actual secondary-panel activation, never background outcomes. */
  attachExternal(key: symbol, detach: () => void) {
    if (this.disposed || this.external?.key === key) return;
    this.close();
    this.leaveExternal();
    this.external = { key, detach };
  }
  detachExternal(key: symbol) {
    if (this.external?.key === key) this.external = null;
  }
  openList = () => {
    if (this.disposed) return;
    this.leaveExternal();
    this.detach();
    this.snapshot = {
      ...this.snapshot,
      open: true,
      message: null,
      activation: this.snapshot.activation + 1,
    };
    this.emit();
  };
  activate = (key: string) => {
    if (this.disposed) return;
    this.leaveExternal();
    if (this.snapshot.selected !== key) this.detach();
    const entry = this.snapshot.entries.find((item) => item.key === key);
    this.snapshot = {
      ...this.snapshot,
      open: true,
      selected: entry?.key ?? null,
      message: entry ? null : "This recovery item is no longer available.",
      activation: this.snapshot.activation + 1,
    };
    if (
      entry &&
      this.sources.get(entry.source)?.callbacks.activate?.(entry.id) === false
    ) {
      this.detach();
      this.snapshot = {
        ...this.snapshot,
        message: "This recovery item is no longer available.",
      };
    }
    this.emit();
  };
  close = () => {
    if (!this.snapshot.open) return;
    this.snapshot = { ...this.snapshot, open: false, message: null };
    this.detach();
    this.emit();
  };
  /** Hides a completed notice, without calling a record/receipt mutation. */
  dismissCompleted(key: string) {
    if (
      this.snapshot.entries.find((entry) => entry.key === key)?.attention !==
      "completed"
    )
      return;
    this.dismissed.add(key);
    this.reconcile();
  }
  private reconcile() {
    const previous = this.selectedEntry();
    const all = [...this.sources].flatMap(([source, registration]) =>
      registration.items.map((item) => ({
        ...item,
        source,
        key: workbookRecoveryKey(source, item.id),
      })),
    );
    const keys = new Set(all.map((entry) => entry.key));
    for (const key of this.dismissed)
      if (!keys.has(key)) this.dismissed.delete(key);
    const representedConflicts = new Set(
      all.flatMap((entry) => entry.conflictKeys ?? []),
    );
    const representedOperations = new Set(
      all.flatMap((entry) => entry.operationIds ?? []),
    );
    const representedRefreshes = new Set(
      all.flatMap((entry) => entry.refreshViews ?? []),
    );
    const entries = all
      .filter(
        (entry) =>
          !entry.conflictKey || !representedConflicts.has(entry.conflictKey),
      )
      .filter(
        (entry) =>
          !entry.conflictOperationId ||
          !representedOperations.has(entry.conflictOperationId),
      )
      .filter(
        (entry) =>
          !entry.refreshOnlyView ||
          !representedRefreshes.has(entry.refreshOnlyView),
      )
      .filter(
        (entry) =>
          entry.attention !== "completed" || !this.dismissed.has(entry.key),
      )
      .sort(
        (a, b) =>
          attentionOrder[a.attention] - attentionOrder[b.attention] ||
          (a.priority ?? 10) - (b.priority ?? 10) ||
          a.order - b.order ||
          (a.key < b.key ? -1 : a.key > b.key ? 1 : 0),
      );
    const selected = entries.some(
      (entry) => entry.key === this.snapshot.selected,
    )
      ? this.snapshot.selected
      : null;
    this.snapshot = {
      ...this.snapshot,
      entries,
      count: entries.filter((entry) => entry.attention !== "completed").length,
      selected,
      message:
        previous && !selected
          ? "This recovery item is no longer available."
          : this.snapshot.message,
    };
    if (previous && !selected)
      this.sources.get(previous.source)?.callbacks.detach?.();
    this.emit();
  }
  private emit() {
    if (!this.disposed) for (const listener of this.listeners) listener();
  }
  dispose() {
    if (this.disposed) return;
    this.close();
    this.leaveExternal();
    this.disposed = true;
    this.sources.clear();
    this.dismissed.clear();
    this.listeners.clear();
    this.snapshot = {
      entries: [],
      count: 0,
      open: false,
      selected: null,
      message: null,
      activation: 0,
    };
  }
}
