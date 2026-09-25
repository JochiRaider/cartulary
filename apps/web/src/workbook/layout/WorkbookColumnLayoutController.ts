import type {
  GridColumnSizingIntent,
  GridColumnSizingPort,
  GridFrozenColumnPort,
  GridFrozenColumnStatus,
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
  readonly frozenColumns?: GridFrozenColumnPort | undefined;
  readonly defaultWidth: (fieldKey: string) => number | undefined;
  readonly port: GridColumnSizingPort | undefined;
};
export type WorkbookColumnSizingDescriptor = {
  readonly defaultWidth: number | undefined;
  readonly width: number | undefined;
  readonly overridden: boolean;
  readonly unavailableReason: string | null;
};
export type WorkbookColumnSizingOutcome =
  | { readonly kind: "completed"; readonly widthPx: number | undefined }
  | { readonly kind: "unavailable"; readonly reason: string }
  | { readonly kind: "cancelled" };
export type WorkbookColumnSizingControls = {
  readonly read: (fieldKey: string) => WorkbookColumnSizingDescriptor;
  readonly applyWidth: (
    fieldKey: string,
    widthPx: number,
  ) => WorkbookColumnSizingOutcome;
  readonly fitVisible: (
    fieldKey: string,
  ) => Promise<WorkbookColumnSizingOutcome>;
  readonly restoreDefault: (fieldKey: string) => WorkbookColumnSizingOutcome;
  readonly cancel: () => void;
  readonly pendingField: string | null;
  readonly notice: string | null;
};

export type WorkbookFrozenColumnControls = {
  readonly status: GridFrozenColumnStatus | null;
  readonly onBoundaryChange: (field: string | null) => void;
};

/** The single working-layout store; bindings contain capabilities/defaults only. */
export class WorkbookColumnLayoutController {
  private entries = new Map<string, WorkbookResolvedLayoutState>();
  private bindings = new Map<string, WorkbookColumnSizingBinding>();
  private listeners = new Set<() => void>();
  private pending: AbortController | null = null;
  private pendingFit: {
    readonly id: string;
    readonly field: string;
    readonly context: string;
    readonly outcome: Promise<WorkbookColumnSizingOutcome>;
  } | null = null;
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
  freeze = (id: string, field: string | null) => {
    if (
      field !== null &&
      !this.currentLayoutStateForSurface(id).columnOrder.includes(field)
    )
      return;
    this.update(id, (_c, state) => ({
      ...state,
      frozenThroughFieldKey: field,
    }));
  };
  readFreezing = (id: string) =>
    this.bindings.get(id)?.frozenColumns?.getSnapshot() ?? null;
  hide = (id: string, field: string, hidden: boolean) =>
    this.update(id, (c, s) => setWorkbookColumnHidden(c, s, field, hidden));
  move = (id: string, field: string, direction: "earlier" | "later") =>
    this.update(id, (c, s) => moveWorkbookColumn(c, s, field, direction));
  reorder = (id: string, from: string, to: string) =>
    this.update(id, (c, s) => reorderWorkbookColumns(c, s, from, to));
  reset = (id: string) => this.update(id, (c) => defaultWorkbookLayoutState(c));
  restoreDefault = (id: string, field: string): WorkbookColumnSizingOutcome => {
    if (!this.contractFor(id).fieldMap[field])
      return { kind: "unavailable", reason: "Column is unavailable." };
    this.update(id, (c, s) => restoreWorkbookColumnDefault(c, s, field));
    this.publish({
      notice: `${this.label(id, field)} default width restored.`,
    });
    return { kind: "completed", widthPx: this.read(id, field).width };
  };
  applyWidth = (
    id: string,
    field: string,
    widthPx: number,
  ): WorkbookColumnSizingOutcome => this.setWidth(id, field, widthPx, true);
  private setWidth(
    id: string,
    field: string,
    widthPx: number,
    announce: boolean,
  ): WorkbookColumnSizingOutcome {
    if (!this.contractFor(id).fieldMap[field])
      return { kind: "unavailable", reason: "Column is unavailable." };
    if (!isWorkbookColumnWidth(widthPx)) {
      this.cancel();
      return {
        kind: "unavailable",
        reason: "Enter a whole number from 40 to 4096.",
      };
    }
    this.update(id, (c, s) => setWorkbookColumnWidth(c, s, field, widthPx));
    if (announce)
      this.publish({
        notice: `${this.label(id, field)} width set to ${widthPx} px.`,
      });
    return { kind: "completed", widthPx };
  }
  onIntent = async (
    id: string,
    intent: GridColumnSizingIntent,
  ): Promise<void> => {
    if (intent.kind === "fit_visible") {
      await this.fitVisible(id, intent.fieldKey);
      return;
    }
    this.setWidth(id, intent.fieldKey, intent.widthPx, false);
  };
  bind = (id: string, binding: WorkbookColumnSizingBinding) => {
    this.cancel();
    this.bindings.set(id, binding);
    const unsubscribe = binding.port?.subscribe(() => this.publish());
    const unsubscribeFreezing = binding.frozenColumns?.subscribe(() =>
      this.publish(),
    );
    this.publish();
    return () => {
      unsubscribe?.();
      unsubscribeFreezing?.();
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
    this.pendingFit = null;
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
  fitVisible = (
    id: string,
    field: string,
  ): Promise<WorkbookColumnSizingOutcome> => {
    if (!this.contractFor(id).fieldMap[field])
      return Promise.resolve({
        kind: "unavailable",
        reason: "Column is unavailable.",
      });
    if (
      this.pending !== null &&
      this.pendingFit?.id === id &&
      this.pendingFit.field === field &&
      this.pendingFit.context === this.context
    )
      return this.pendingFit.outcome;
    this.cancel();
    const binding = this.bindings.get(id);
    const reason = this.read(id, field).unavailableReason;
    if (reason !== null || !binding?.port) {
      const unavailableReason = reason ?? "Column measurement is unavailable.";
      this.publish({ notice: unavailableReason });
      return Promise.resolve({
        kind: "unavailable",
        reason: unavailableReason,
      });
    }
    const abort = new AbortController();
    this.pending = abort;
    const context = this.context;
    const layout = this.currentLayoutStateForSurface(id);
    this.publish({ pendingField: field, notice: null });
    const outcome = this.completeFit(
      id,
      field,
      binding,
      abort,
      context,
      layout,
    );
    this.pendingFit = { id, field, context, outcome };
    return outcome;
  };
  private completeFit = async (
    id: string,
    field: string,
    binding: WorkbookColumnSizingBinding,
    abort: AbortController,
    context: string,
    layout: WorkbookResolvedLayoutState,
  ): Promise<WorkbookColumnSizingOutcome> => {
    if (!binding.port) return { kind: "cancelled" };
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
      return { kind: "cancelled" };
    this.pending = null;
    this.pendingFit = null;
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
      return { kind: "completed", widthPx: result.widthPx };
    }
    if (result.kind === "cancelled") {
      this.publish({ pendingField: null, notice: null });
      return { kind: "cancelled" };
    }
    const unavailableReason =
      result.kind === "unavailable"
        ? result.reason
        : "Column measurement returned an invalid width. Try Fit again.";
    this.publish({ pendingField: null, notice: unavailableReason });
    return { kind: "unavailable", reason: unavailableReason };
  };
}
