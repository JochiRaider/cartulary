import type { GridColumn } from "@cartulary/grid-adapter";
import type { ViewContract } from "@cartulary/view-contracts";
import {
  buildSavedViewLayoutJson,
  type WorkbookLayoutState,
} from "../models/workbookQuery";

export type WorkbookResolvedLayoutState = {
  readonly columnOrder: readonly string[];
  readonly columnWidths: Readonly<Record<string, number>>;
  readonly hiddenFieldKeys: readonly string[];
};

export function resolveWorkbookLayoutState(
  contract: ViewContract,
  state: WorkbookLayoutState = {},
): WorkbookResolvedLayoutState {
  const normalized = buildSavedViewLayoutJson(contract, state);
  return {
    columnOrder: normalized.column_order,
    columnWidths: Object.fromEntries(
      normalized.column_widths.map((entry) => [
        entry.field_key,
        entry.width_px,
      ]),
    ),
    hiddenFieldKeys: normalized.hidden_field_keys,
  };
}

export function defaultWorkbookLayoutState(
  contract: ViewContract,
): WorkbookResolvedLayoutState {
  return resolveWorkbookLayoutState(contract);
}

export function applyWorkbookLayoutToColumns<Row>(
  contract: ViewContract,
  columns: readonly GridColumn<Row>[],
  state: WorkbookLayoutState,
): readonly GridColumn<Row>[] {
  const layout = resolveWorkbookLayoutState(contract, state);
  const hidden = new Set(layout.hiddenFieldKeys);
  const byKey = new Map(columns.map((column) => [column.fieldKey, column]));
  return layout.columnOrder.flatMap((fieldKey) => {
    if (hidden.has(fieldKey)) {
      return [];
    }
    const column = byKey.get(fieldKey);
    if (column === undefined) {
      return [];
    }
    const width = layout.columnWidths[fieldKey];
    return [width === undefined ? column : { ...column, width }];
  });
}

export function reorderWorkbookColumns(
  contract: ViewContract,
  state: WorkbookLayoutState,
  sourceFieldKey: string,
  targetFieldKey: string,
): WorkbookResolvedLayoutState {
  const current = resolveWorkbookLayoutState(contract, state);
  if (
    sourceFieldKey === targetFieldKey ||
    !current.columnOrder.includes(sourceFieldKey) ||
    !current.columnOrder.includes(targetFieldKey)
  ) {
    return current;
  }
  const withoutSource = current.columnOrder.filter(
    (fieldKey) => fieldKey !== sourceFieldKey,
  );
  const targetIndex = withoutSource.indexOf(targetFieldKey);
  return resolveWorkbookLayoutState(contract, {
    ...current,
    columnOrder: [
      ...withoutSource.slice(0, targetIndex),
      sourceFieldKey,
      ...withoutSource.slice(targetIndex),
    ],
  });
}

export function moveWorkbookColumn(
  contract: ViewContract,
  state: WorkbookLayoutState,
  fieldKey: string,
  direction: "earlier" | "later",
): WorkbookResolvedLayoutState {
  const current = resolveWorkbookLayoutState(contract, state);
  const index = current.columnOrder.indexOf(fieldKey);
  const nextIndex = direction === "earlier" ? index - 1 : index + 1;
  if (index < 0 || nextIndex < 0 || nextIndex >= current.columnOrder.length) {
    return current;
  }
  const nextOrder = [...current.columnOrder];
  [nextOrder[index], nextOrder[nextIndex]] = [
    nextOrder[nextIndex] as string,
    nextOrder[index] as string,
  ];
  return resolveWorkbookLayoutState(contract, {
    ...current,
    columnOrder: nextOrder,
  });
}

export function setWorkbookColumnHidden(
  contract: ViewContract,
  state: WorkbookLayoutState,
  fieldKey: string,
  hidden: boolean,
): WorkbookResolvedLayoutState {
  if (!contract.fieldMap[fieldKey]) {
    return resolveWorkbookLayoutState(contract, state);
  }
  const current = resolveWorkbookLayoutState(contract, state);
  const hiddenKeys = new Set(current.hiddenFieldKeys);
  if (hidden) {
    hiddenKeys.add(fieldKey);
  } else {
    hiddenKeys.delete(fieldKey);
  }
  return resolveWorkbookLayoutState(contract, {
    ...current,
    hiddenFieldKeys: [...hiddenKeys],
  });
}

export function setWorkbookColumnWidth(
  contract: ViewContract,
  state: WorkbookLayoutState,
  fieldKey: string,
  width: number,
): WorkbookResolvedLayoutState {
  const current = resolveWorkbookLayoutState(contract, state);
  if (
    !contract.fieldMap[fieldKey] ||
    !Number.isSafeInteger(width) ||
    width < 40 ||
    width > 4096
  ) {
    return current;
  }
  return resolveWorkbookLayoutState(contract, {
    ...current,
    columnWidths: { ...current.columnWidths, [fieldKey]: width },
  });
}
