import type {
  GridColumnSizingIntent,
  GridColumnSizingPort,
} from "@cartulary/grid-adapter";
import type { ViewContract } from "@cartulary/view-contracts";
import { isWorkbookColumnWidth } from "../models/workbookColumnSizing";
import type { WorkbookLayoutState } from "../models/workbookQuery";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import {
  defaultWorkbookLayoutState,
  moveWorkbookColumn,
  reorderWorkbookColumns,
  resolveWorkbookLayoutState,
  restoreWorkbookColumnDefault,
  setWorkbookColumnHidden,
  setWorkbookColumnWidth,
  type WorkbookResolvedLayoutState,
} from "./workbookColumnLayout";

export type WorkbookColumnSizingBinding = {
  readonly defaultWidth: (fieldKey: string) => number | undefined;
  readonly port: GridColumnSizingPort | undefined;
};
export type WorkbookColumnSizingDescriptor = {
  readonly defaultWidth: number | undefined;
  readonly width: number | undefined;
  readonly overridden: boolean;
  readonly unavailableReason: string | null;
};
export type WorkbookColumnSizingControls = {
  readonly read: (fieldKey: string) => WorkbookColumnSizingDescriptor;
  readonly onIntent: (intent: GridColumnSizingIntent) => void;
  readonly restoreDefault: (fieldKey: string) => void;
  readonly cancel: () => void;
  readonly pendingField: string | null;
  readonly notice: string | null;
};

/** The single working-layout store; bindings contain capabilities/defaults only. */
export class WorkbookColumnLayoutController {
  private entries = new Map<string, WorkbookResolvedLayoutState>();
  private bindings = new Map<string, WorkbookColumnSizingBinding>();
  private listeners = new Set<() => void>();
  private pending: AbortController | null = null;
  private context = "";
  private revision = 0;
  private snapshot = {
    revision: 0,
    pendingField: null as string | null,
    notice: null as string | null,
  };
  constructor(
    private readonly contractFor: (
      id: string,
    ) => ViewContract = workbookContractForViewSchemaId,
  ) {}
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.snapshot;
  private publish(patch: Partial<typeof this.snapshot> = {}) {
    this.snapshot = { ...this.snapshot, ...patch, revision: ++this.revision };
    for (const listener of this.listeners) listener();
  }
  activate = (context: string) => {
    if (context === this.context) return;
    this.context = context;
    this.cancel();
    this.publish({ notice: null });
  };
  currentLayoutStateForSurface = (id: string): WorkbookResolvedLayoutState => {
    let state = this.entries.get(id);
    if (!state) {
      state = defaultWorkbookLayoutState(this.contractFor(id));
      this.entries.set(id, state);
    }
    return state;
  };
  applyLayoutStateForSurface = (id: string, state: WorkbookLayoutState) => {
    this.cancel();
    this.entries.set(
      id,
      resolveWorkbookLayoutState(this.contractFor(id), state),
    );
    this.publish({ notice: null });
  };
  private update(
    id: string,
    update: (
      contract: ViewContract,
      state: WorkbookResolvedLayoutState,
    ) => WorkbookResolvedLayoutState,
  ) {
    this.cancel();
    this.entries.set(
      id,
      update(this.contractFor(id), this.currentLayoutStateForSurface(id)),
    );
    this.publish({ notice: null });
  }
  hide = (id: string, field: string, hidden: boolean) =>
    this.update(id, (c, s) => setWorkbookColumnHidden(c, s, field, hidden));
  move = (id: string, field: string, direction: "earlier" | "later") =>
    this.update(id, (c, s) => moveWorkbookColumn(c, s, field, direction));
  reorder = (id: string, from: string, to: string) =>
    this.update(id, (c, s) => reorderWorkbookColumns(c, s, from, to));
  reset = (id: string) => this.update(id, (c) => defaultWorkbookLayoutState(c));
  restoreDefault = (id: string, field: string) => {
    this.update(id, (c, s) => restoreWorkbookColumnDefault(c, s, field));
    this.publish({
      notice: `${this.label(id, field)} default width restored.`,
    });
  };
  onIntent = async (
    id: string,
    intent: GridColumnSizingIntent,
  ): Promise<void> => {
    if (!this.contractFor(id).fieldMap[intent.fieldKey]) return;
    if (intent.kind === "fit_visible") {
      return this.fitVisible(id, intent.fieldKey);
    }
    this.cancel();
    if (!isWorkbookColumnWidth(intent.widthPx)) return;
    this.update(id, (c, s) =>
      setWorkbookColumnWidth(c, s, intent.fieldKey, intent.widthPx),
    );
  };
  bind = (id: string, binding: WorkbookColumnSizingBinding) => {
    this.cancel();
    this.bindings.set(id, binding);
    const unsubscribe = binding.port?.subscribe(() => this.publish());
    this.publish();
    return () => {
      unsubscribe?.();
      if (this.bindings.get(id) !== binding) return;
      this.cancel();
      this.bindings.delete(id);
      this.publish();
    };
  };
  refresh = () => this.publish();
  read = (id: string, field: string): WorkbookColumnSizingDescriptor => {
    const state = this.currentLayoutStateForSurface(id);
    const binding = this.bindings.get(id);
    const defaultWidth = binding?.defaultWidth(field);
    const override = state.columnWidths[field];
    return {
      defaultWidth,
      width: override ?? defaultWidth,
      overridden: override !== undefined,
      unavailableReason: state.hiddenFieldKeys.includes(field)
        ? "Show this column before fitting it."
        : (binding?.port?.unavailableReason(field) ??
          (binding?.port ? null : "Column measurement is unavailable.")),
    };
  };
  cancel = () => {
    const pending = this.pending;
    this.pending = null;
    pending?.abort();
    if (this.snapshot.pendingField !== null)
      this.publish({ pendingField: null });
  };
  dispose = () => {
    this.cancel();
    this.bindings.clear();
    this.listeners.clear();
  };
  private label(id: string, field: string) {
    return this.contractFor(id).fieldMap[field]?.label ?? field;
  }
  private async fitVisible(id: string, field: string) {
    this.cancel();
    const binding = this.bindings.get(id);
    const reason = this.read(id, field).unavailableReason;
    if (reason !== null || !binding?.port) {
      this.publish({ notice: reason ?? "Column measurement is unavailable." });
      return;
    }
    const abort = new AbortController();
    this.pending = abort;
    const context = this.context;
    const layout = this.currentLayoutStateForSurface(id);
    this.publish({ pendingField: field, notice: null });
    const result = await binding.port
      .measureVisibleContent(field, {
        signal: abort.signal,
      })
      .catch(() => ({
        kind: "unavailable" as const,
        reason: "Column measurement is unavailable. Try Fit again.",
      }));
    if (
      abort.signal.aborted ||
      this.pending !== abort ||
      context !== this.context ||
      binding !== this.bindings.get(id) ||
      layout !== this.currentLayoutStateForSurface(id)
    )
      return;
    this.pending = null;
    if (result.kind === "measured" && isWorkbookColumnWidth(result.widthPx)) {
      this.entries.set(
        id,
        setWorkbookColumnWidth(
          this.contractFor(id),
          layout,
          field,
          result.widthPx,
        ),
      );
      this.publish({
        pendingField: null,
        notice: `${this.label(id, field)} fitted to ${result.widthPx} px${result.cellCount === 0 ? " using the header" : ""}.${result.capped ? " Maximum width reached; content may still be clipped." : ""}`,
      });
    } else
      this.publish({
        pendingField: null,
        notice: result.kind === "unavailable" ? result.reason : null,
      });
  }
}
