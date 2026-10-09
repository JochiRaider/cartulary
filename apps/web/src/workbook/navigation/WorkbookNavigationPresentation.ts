import type { GridFocusTarget, GridHandle } from "@cartulary/grid-adapter";
import type { WorkbookQueryAcceptance } from "../query/WorkbookQueryBrowser";

type Selection = {
  readonly token: symbol;
  readonly select: (id: string, committed: () => void) => void;
};
type Source = {
  readonly subscribePresentation: (listener: () => void) => () => void;
  readonly grid: (view: string) => GridHandle | null;
  readonly selectionBinding: (view: string) => Selection | undefined;
  readonly presentationState: (
    view: string,
  ) => { readonly token: symbol; readonly ready: boolean } | undefined;
};
type Destination = {
  readonly view: string;
  readonly recordId?: string;
  readonly fieldKey?: string;
  readonly mode: "record" | "entry" | "extension";
  readonly inspect: boolean;
  readonly acceptance?: WorkbookQueryAcceptance;
  readonly signal: AbortSignal;
  readonly admitted: () => boolean;
  readonly committed: () => boolean;
  readonly entryReady: () => boolean;
  readonly settled: (outcome: "succeeded" | "failed" | "cancelled") => void;
};
type Pending = {
  destination: Destination;
  removeAbort: () => void;
  mount: symbol | null;
  grid: GridHandle | null;
  selection: Selection | undefined;
  selected: boolean;
  selectionCommitted: boolean;
  focused: boolean;
  focus: AbortController | null;
  opened: object | null;
};

/** Sequences semantic presentation only. Query pages and source work stay with owners. */
export class WorkbookNavigationPresentation {
  private active: Pending | null = null;
  private completed: (() => boolean) | null = null;
  private unsubscribe: (() => void) | null = null;
  private scheduled = false;
  private openers = new Map<string, { open: () => void }>();
  private inspectors = new Map<
    string,
    { recordId: string; focus: () => boolean }
  >();
  constructor(private readonly source: Source) {}
  connect() {
    this.unsubscribe ??= this.source.subscribePresentation(this.resume);
    return () => {
      this.cancel();
      this.unsubscribe?.();
      this.unsubscribe = null;
    };
  }
  isCurrent = () => this.completed?.() ?? false;
  registerInspector = (view: string, open: () => void) => {
    const binding = { open };
    this.openers.set(view, binding);
    this.resume();
    return () => {
      if (this.openers.get(view) === binding) {
        this.openers.delete(view);
        this.resume();
      }
    };
  };
  registerInspectorFocus = (
    view: string,
    recordId: string,
    focus: () => boolean,
  ) => {
    const binding = { recordId, focus };
    this.inspectors.set(view, binding);
    this.resume();
    return () => {
      if (this.inspectors.get(view) === binding) {
        this.inspectors.delete(view);
        this.resume();
      }
    };
  };
  start(destination: Destination) {
    this.cancel();
    const pending: Pending = {
      destination,
      removeAbort: () => {},
      mount: null,
      grid: null,
      selection: undefined,
      selected: false,
      selectionCommitted: false,
      focused: false,
      focus: null,
      opened: null,
    };
    this.active = pending;
    const abort = () => {
      if (this.active === pending) this.finish(pending, "cancelled");
    };
    destination.signal.addEventListener("abort", abort, { once: true });
    pending.removeAbort = () =>
      destination.signal.removeEventListener("abort", abort);
    if (destination.signal.aborted) abort();
    this.resume();
  }
  cancel = () => {
    this.completed = null;
    if (this.active) this.finish(this.active, "cancelled");
  };
  private finish(
    pending: Pending,
    outcome: "succeeded" | "failed" | "cancelled",
  ) {
    if (this.active !== pending) return;
    this.active = null;
    pending.removeAbort();
    pending.focus?.abort();
    pending.destination.acceptance?.release();
    this.completed =
      outcome === "succeeded" ? pending.destination.committed : null;
    pending.destination.settled(outcome);
  }
  resume = () => {
    if (this.scheduled || !this.active) return;
    this.scheduled = true;
    queueMicrotask(() => {
      this.scheduled = false;
      this.reconcile();
    });
  };
  private reconcile() {
    const pending = this.active;
    if (!pending) return;
    const destination = pending.destination;
    if (destination.signal.aborted) return this.finish(pending, "cancelled");
    if (!destination.admitted()) return;
    const acceptance = destination.acceptance?.state() ?? "accepted";
    if (acceptance === "failed" || acceptance === "cancelled")
      return this.finish(pending, acceptance);
    if (acceptance !== "accepted" || !destination.committed()) return;
    if (destination.mode === "extension")
      return this.finish(pending, "succeeded");
    const presentation = this.source.presentationState(destination.view);
    if (!presentation) {
      if (pending.mount) this.finish(pending, "cancelled");
      return;
    }
    const grid = this.source.grid(destination.view);
    const selection = this.source.selectionBinding(destination.view);
    if (
      pending.mount !== presentation.token ||
      pending.grid !== grid ||
      pending.selection !== selection
    ) {
      pending.focus?.abort();
      pending.focus = null;
      // A new grid handle fences its pending focus, but does not undo a committed
      // source selection or completed grid focus while Record is attaching.
      if (
        pending.mount !== presentation.token ||
        pending.selection !== selection
      ) {
        pending.selected = pending.selectionCommitted = pending.focused = false;
        pending.opened = null;
      }
      pending.mount = presentation.token;
      pending.grid = grid;
      pending.selection = selection;
    }
    if (pending.focused) {
      if (!destination.inspect) return this.finish(pending, "succeeded");
      const inspector = this.inspectors.get(destination.view);
      if (
        inspector &&
        inspector.recordId === destination.recordId &&
        inspector.focus()
      ) {
        if (this.inspectors.get(destination.view) === inspector)
          this.finish(pending, "succeeded");
        return;
      }
      const opener = this.openers.get(destination.view);
      if (opener && pending.opened !== opener) {
        pending.opened = opener;
        opener.open();
      }
      return;
    }
    if (!presentation.ready) {
      pending.focus?.abort();
      pending.focus = null;
      return;
    }
    if (destination.mode === "entry") {
      if (destination.entryReady()) this.finish(pending, "succeeded");
      return;
    }
    if (!grid?.presentation || pending.focus) return;
    if (destination.recordId) {
      if (!destination.acceptance?.hasRecord(destination.recordId))
        return this.finish(pending, "failed");
      grid.revealRecord?.(destination.recordId);
    }
    const snapshot = grid.presentation.getSnapshot();
    if (!snapshot) return;
    const row = destination.recordId
      ? snapshot.rowIdentities.find(
          (row) =>
            row.kind === "core_record" && row.recordId === destination.recordId,
        )
      : snapshot.rowIdentities.find((row) => row.kind === "core_record");
    if (destination.recordId && !row) return;
    if (row?.kind === "core_record" && !pending.selected) {
      if (!selection) return;
      pending.selected = true;
      selection.select(row.recordId, () => {
        if (
          this.active !== pending ||
          this.source.selectionBinding(destination.view) !== selection
        )
          return;
        pending.selectionCommitted = true;
        this.resume();
      });
      return;
    }
    if (pending.selected && !pending.selectionCommitted) return;
    const fields = [
      ...new Set([
        ...(destination.fieldKey &&
        snapshot.fieldKeys.includes(destination.fieldKey)
          ? [destination.fieldKey]
          : []),
        ...snapshot.fieldKeys,
      ]),
    ];
    const targets: GridFocusTarget[] = [
      ...(row
        ? fields.map((fieldKey) => ({
            kind: "cell" as const,
            anchor: { surface: snapshot.surface, rowIdentity: row, fieldKey },
          }))
        : []),
      { kind: "root" },
    ];
    const focus = new AbortController();
    pending.focus = focus;
    const current = () =>
      this.active === pending &&
      !focus.signal.aborted &&
      this.source.grid(destination.view) === grid &&
      this.source.presentationState(destination.view)?.token === pending.mount;
    void (async () => {
      try {
        for (const target of targets) {
          if (!current()) return;
          const outcome = await grid.requestFocus(target, {
            signal: focus.signal,
          });
          if (!current()) return;
          if (outcome === "cancelled") return this.finish(pending, "cancelled");
          if (outcome === "focused") {
            pending.focus = null;
            pending.focused = true;
            this.resume();
            return;
          }
        }
        if (current()) this.finish(pending, "failed");
      } catch {
        if (current()) this.finish(pending, "failed");
      }
    })();
  }
}
