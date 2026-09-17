import type {
  GridColumnSizingIntent,
  GridColumnSizingPort,
} from "@cartulary/grid-adapter";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  type NetworkFlowGridSchemaId,
  type NetworkFlowPresentationColumn,
  networkFlowPresentationColumns,
} from "./networkFlowPresentation";

const sessionLayouts = new Map<string, NetworkFlowGridLayout>();

type LayoutGridSchemaId = NetworkFlowGridSchemaId;

export type NetworkFlowGridLayout = {
  readonly order: readonly string[];
  readonly visible: ReadonlySet<string>;
  readonly widths: Readonly<Record<string, number>>;
};

export function useNetworkFlowGridLayout(gridSchemaId: LayoutGridSchemaId) {
  const key = layoutKey(gridSchemaId);
  const pending = useRef<AbortController | null>(null);
  const cancelSizing = useCallback(() => {
    pending.current?.abort();
    pending.current = null;
  }, []);
  useEffect(() => cancelSizing, [cancelSizing]);
  const [layout, setLayout] = useState<NetworkFlowGridLayout>(() => {
    const stored = sessionLayouts.get(key);
    if (stored !== undefined) {
      return stored;
    }
    const initial = defaultLayout(gridSchemaId);
    sessionLayouts.set(key, initial);
    return initial;
  });
  const metadata = useMemo(
    () => configurableColumns(gridSchemaId),
    [gridSchemaId],
  );

  const update = useCallback(
    (mutate: (current: NetworkFlowGridLayout) => NetworkFlowGridLayout) => {
      cancelSizing();
      setLayout((current) => {
        const next = mutate(current);
        sessionLayouts.set(key, next);
        return next;
      });
    },
    [key, cancelSizing],
  );

  const onColumnReorder = useCallback(
    (sourceFieldKey: string, targetFieldKey: string) => {
      update((current) => ({
        ...current,
        order: reorder(current.order, sourceFieldKey, targetFieldKey),
      }));
    },
    [update],
  );

  const onColumnSizingIntent = useCallback(
    (intent: GridColumnSizingIntent, port?: GridColumnSizingPort) => {
      cancelSizing();
      const column = metadata.find(
        (candidate) => candidate.field_key === intent.fieldKey,
      );
      if (!column) return;
      const apply = (width: number) => {
        if (!Number.isSafeInteger(width) || width < column.minimum_width_px)
          return;
        update((current) => ({
          ...current,
          widths: { ...current.widths, [intent.fieldKey]: width },
        }));
      };
      if (intent.kind === "set_width") {
        apply(intent.widthPx);
        return;
      }
      if (!port) return;
      const abort = new AbortController();
      pending.current = abort;
      void port
        .measureVisibleContent(intent.fieldKey, { signal: abort.signal })
        .then((result) => {
          if (pending.current !== abort || abort.signal.aborted) return;
          pending.current = null;
          if (result.kind === "measured") apply(result.widthPx);
        });
    },
    [metadata, update, cancelSizing],
  );

  const setColumnVisible = useCallback(
    (fieldKey: string, visible: boolean) => {
      if (!metadata.some((candidate) => candidate.field_key === fieldKey)) {
        return;
      }
      update((current) => {
        const nextVisible = new Set(current.visible);
        if (visible) {
          nextVisible.add(fieldKey);
        } else {
          nextVisible.delete(fieldKey);
        }
        return { ...current, visible: nextVisible };
      });
    },
    [metadata, update],
  );

  const reset = useCallback(() => {
    cancelSizing();
    const next = defaultLayout(gridSchemaId);
    sessionLayouts.set(key, next);
    setLayout(next);
  }, [gridSchemaId, key, cancelSizing]);

  return {
    allColumns: metadata,
    columnWidths: layout.widths,
    onColumnReorder,
    onColumnSizingIntent,
    orderedVisibleFieldKeys: layout.order.filter((fieldKey) =>
      layout.visible.has(fieldKey),
    ),
    reset,
    setColumnVisible,
    visibleFieldKeys: layout.visible,
  };
}

function defaultLayout(
  gridSchemaId: LayoutGridSchemaId,
): NetworkFlowGridLayout {
  const columns = configurableColumns(gridSchemaId);
  return {
    order: columns.map((column) => column.field_key),
    visible: new Set(
      columns
        .filter((column) => column.default_visible)
        .map((column) => column.field_key),
    ),
    widths: Object.fromEntries(
      columns.map((column) => [column.field_key, column.default_width_px]),
    ),
  };
}

function configurableColumns(
  gridSchemaId: LayoutGridSchemaId,
): readonly NetworkFlowPresentationColumn[] {
  return networkFlowPresentationColumns(gridSchemaId).filter(
    (column) =>
      !column.inspector_only &&
      !(
        (gridSchemaId === "network_flow.accepted_rows.v1" ||
          gridSchemaId === "network_flow.graph_contributors.v1") &&
        column.field_key === "source_row_number"
      ),
  );
}

function reorder(
  order: readonly string[],
  sourceFieldKey: string,
  targetFieldKey: string,
): readonly string[] {
  const sourceIndex = order.indexOf(sourceFieldKey);
  const targetIndex = order.indexOf(targetFieldKey);
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) {
    return order;
  }
  const next = [...order];
  next.splice(sourceIndex, 1);
  next.splice(targetIndex, 0, sourceFieldKey);
  return next;
}

function layoutKey(gridSchemaId: LayoutGridSchemaId): string {
  return `network_flow_activity:network_analysis:${gridSchemaId}`;
}
