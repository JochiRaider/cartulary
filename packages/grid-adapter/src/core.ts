import type {
  CSSProperties,
  PropsWithChildren,
  ReactNode,
  RefCallback,
} from "react";

export type GridSortDirection = "asc" | "desc";

export type GridDensity = "compact" | "default" | "comfortable";

export type GridSortEntry = {
  readonly fieldKey: string;
  readonly direction: GridSortDirection;
};

export type GridColumn<Row> = {
  readonly contractWritable?: boolean | undefined;
  readonly fieldKey: string;
  readonly headerTestId?: string | undefined;
  readonly label: string;
  readonly renderCell: (context: GridCellRenderContext<Row>) => ReactNode;
  readonly getClipboardValue?: ((row: Row) => unknown) | undefined;
  readonly renderDraftCell?:
    | ((context: GridDraftCellRenderContext<Row>) => ReactNode)
    | undefined;
  readonly editor?: GridEditorAdapter<Row> | undefined;
  readonly sortableFieldKey?: string | null;
  readonly sortDisabled?: boolean | undefined;
  readonly sortDisabledReason?: string | null | undefined;
  readonly valueKind?: string | undefined;
  readonly align?: "left" | "center" | "right" | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridSurfaceIdentity =
  | { readonly kind: "view_schema"; readonly viewSchemaId: string }
  | {
      readonly kind: "extension_grid";
      readonly extensionProfileId: string;
      readonly workspaceKey: string;
      readonly gridSchemaId: string;
    };

export type GridRowIdentity =
  | { readonly kind: "core_record"; readonly recordId: string }
  | {
      readonly kind: "extension_resource";
      readonly extensionProfileId: string;
      readonly resourceKind: string;
      readonly resourceId: string;
    };

export type GridMutationIdentity = {
  readonly kind: "core_row_version";
  readonly baseRowVersion: number;
};

export type GridDataRow<Row> = {
  readonly kind: "data";
  readonly rowIdentity: GridRowIdentity;
  readonly mutationIdentity?: GridMutationIdentity | undefined;
  readonly data: Row;
  readonly gutterContent?: ReactNode | undefined;
  readonly gutterLabel?: string | undefined;
  readonly gutterTestId?: string | undefined;
  readonly testId?: string | undefined;
};

export type GridStateValidation = {
  readonly message: string;
};

/**
 * Vendor-neutral semantic state supplied by workbook owners. The adapter adds
 * active-cell, bulk-selection, inspector, read-only, stale-query, and saved
 * defaults before compiling private RDG classes and accessibility attributes.
 */
export type GridSemanticStateInput = {
  readonly active?: boolean | undefined;
  readonly bulkSelected?: boolean | undefined;
  readonly conflicted?: boolean | undefined;
  readonly inspectorActive?: boolean | undefined;
  readonly invalid?: GridStateValidation | false | undefined;
  readonly pending?: boolean | undefined;
  readonly readOnlyOrDerived?: boolean | undefined;
  readonly saved?: boolean | undefined;
  readonly stale?: boolean | undefined;
};

export type GridCellStateInput = GridSemanticStateInput;
export type GridRowStateInput = GridSemanticStateInput;

export type GridCellStateContext<Row> = GridCellRenderContext<Row> & {
  readonly mutationIdentity?: GridMutationIdentity | undefined;
};

export type GridCoreRecordBulkSelection<Row> = {
  readonly isRecordSelectable?:
    | ((row: GridDataRow<Row>) => boolean)
    | undefined;
  readonly onSelectedRecordIdsChange: (recordIds: ReadonlySet<string>) => void;
  readonly selectedRecordIds: ReadonlySet<string>;
};

export type GridDataStateAction = {
  readonly label: string;
  readonly onInvoke: () => void;
};

export type GridDataState =
  | { readonly kind: "ready" }
  | { readonly kind: "initial_loading"; readonly surfaceLabel: string }
  | { readonly kind: "refreshing"; readonly surfaceLabel: string }
  | {
      readonly kind: "empty";
      readonly message: string;
      readonly action?: GridDataStateAction | undefined;
    }
  | {
      readonly kind: "filtered_empty";
      readonly action: GridDataStateAction;
    }
  | {
      readonly kind: "stale_error";
      readonly message: string;
      readonly action?: GridDataStateAction | undefined;
    }
  | {
      readonly kind: "unavailable";
      readonly message: string;
      readonly action?: GridDataStateAction | undefined;
    }
  | {
      readonly kind: "permission_denied";
      readonly message?: string | undefined;
    };

export type GridInteractionMode =
  | { readonly kind: "editable" }
  | { readonly kind: "read_only"; readonly label: string };

export type GridDraftRow<Row> = {
  readonly kind: "draft";
  readonly data: Row;
  readonly gutterContent?: ReactNode | undefined;
  readonly gutterLabel?: string | undefined;
  readonly testId?: string | undefined;
};

export type GridSemanticRow<Row> = GridDraftRow<Row> | GridDataRow<Row>;

export type GridRowGutter = {
  readonly headerTestId?: string | undefined;
  readonly label?: ReactNode | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridActionsColumn<Row> = {
  readonly headerTestId?: string | undefined;
  readonly label: string;
  readonly renderCell: (row: GridDataRow<Row>) => ReactNode;
  readonly renderDraftCell?:
    | ((row: GridDraftRow<Row>) => ReactNode)
    | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridChrome = "sheet" | "framed";
export type GridBlockSizing = "standalone" | "fill";

export type GridViewportProps = PropsWithChildren<{
  readonly blockSizing?: GridBlockSizing | undefined;
  readonly className?: string | undefined;
  readonly chrome?: GridChrome | undefined;
  readonly style?: CSSProperties | undefined;
  readonly testId?: string | undefined;
}>;

type SemanticDataGridBaseProps<Row> = {
  readonly accessibleLabel?: string | undefined;
  readonly allowPasteCreateRows?: boolean | undefined;
  readonly activeRowIdentity?: GridRowIdentity | null | undefined;
  readonly actionsColumn?: GridActionsColumn<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly coreRecordBulkSelection?:
    | GridCoreRecordBulkSelection<Row>
    | undefined;
  readonly cellRange?: GridCellRange | null | undefined;
  readonly columnWidths?: Readonly<Record<string, number>> | undefined;
  readonly dataState?: GridDataState | undefined;
  readonly density?: GridDensity | undefined;
  readonly draftRow?: GridDraftRow<Row> | undefined;
  readonly fillViewportInline?: boolean | undefined;
  readonly grouping?: GridGroupingDescriptor<Row> | null | undefined;
  readonly getCellState?:
    | ((context: GridCellStateContext<Row>) => GridCellStateInput)
    | undefined;
  readonly getRowState?:
    | ((row: GridDataRow<Row>) => GridRowStateInput)
    | undefined;
  readonly interactionMode?: GridInteractionMode | undefined;
  readonly onActiveCellChange?:
    | ((anchor: GridCellAnchor | null) => void)
    | undefined;
  readonly onCellRangeChange?:
    | ((range: GridCellRange | null) => void)
    | undefined;
  readonly onCopyCell?: ((intent: GridCellCopyIntent) => void) | undefined;
  readonly onFillCells?: ((intent: GridFillIntent) => void) | undefined;
  readonly onPasteCell?: ((intent: GridCellMutationIntent) => void) | undefined;
  readonly onSortChange?:
    | ((sort: readonly GridSortEntry[]) => void)
    | undefined;
  readonly onColumnReorder?:
    | ((sourceFieldKey: string, targetFieldKey: string) => void)
    | undefined;
  readonly onColumnWidthChange?:
    | ((fieldKey: string, width: number) => void)
    | undefined;
  readonly onSelectRow?: ((rowIdentity: GridRowIdentity) => void) | undefined;
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly rowGutter?: GridRowGutter | undefined;
  readonly sort?: readonly GridSortEntry[] | undefined;
  readonly surface: GridSurfaceIdentity;
};

export type SemanticDataGridProps<Row> =
  | (SemanticDataGridBaseProps<Row> & {
      readonly surface: Extract<
        GridSurfaceIdentity,
        { readonly kind: "view_schema" }
      >;
    })
  | (Omit<
      SemanticDataGridBaseProps<Row>,
      | "actionsColumn"
      | "allowPasteCreateRows"
      | "coreRecordBulkSelection"
      | "draftRow"
      | "interactionMode"
      | "onFillCells"
      | "onPasteCell"
    > & {
      readonly surface: Extract<
        GridSurfaceIdentity,
        { readonly kind: "extension_grid" }
      >;
      readonly actionsColumn?: never;
      readonly allowPasteCreateRows?: never;
      readonly coreRecordBulkSelection?: never;
      readonly draftRow?: never;
      readonly interactionMode?: {
        readonly kind: "read_only";
        readonly label: string;
      };
      readonly onFillCells?: never;
      readonly onPasteCell?: never;
    });

export type GridPresentationGroupRow = {
  readonly groupBy: string;
  readonly groupLabel: string | null;
  readonly key: string;
  readonly kind: "group";
  readonly testId?: string | undefined;
};

export type GridPresentationDataRow<Row> = {
  readonly gridRow: GridDataRow<Row>;
  readonly key: string;
  readonly kind: "data";
};

export type GridPresentationRow<Row> =
  | GridPresentationGroupRow
  | GridPresentationDataRow<Row>;

export type GridCellAnchor = {
  readonly surface: GridSurfaceIdentity;
  readonly rowIdentity: GridRowIdentity;
  readonly fieldKey: string;
};

export type GridCellTarget = GridCellAnchor & {
  readonly mutationIdentity: GridMutationIdentity;
};

export type GridCellRange = {
  readonly end: GridCellAnchor;
  readonly start: GridCellAnchor;
};

export type GridExpandedCellRange = {
  readonly fieldKeys: readonly string[];
  readonly rowIdentities: readonly GridRowIdentity[];
};

export type GridCellRenderContext<Row> = {
  readonly anchor: GridCellAnchor;
  readonly row: Row;
};

export type GridDraftCellRenderContext<Row> = {
  readonly fieldKey: string;
  readonly row: Row;
  readonly surface: Extract<
    GridSurfaceIdentity,
    { readonly kind: "view_schema" }
  >;
};

export type GridEditCommitIntent<Row> = {
  readonly draftValue: unknown;
  readonly row: Row;
  readonly target: GridCellTarget;
};

export type GridEditCommitOutcome =
  | { readonly kind: "accepted" }
  | { readonly kind: "validation_error"; readonly message: string }
  | { readonly kind: "conflict"; readonly message: string }
  | { readonly kind: "stale_target"; readonly message: string }
  | { readonly kind: "rejected_mutation"; readonly message: string };

export type GridEditorActivation = {
  readonly source:
    | "clear"
    | "enter"
    | "pointer"
    | "printable"
    | "programmatic"
    | "shift_enter";
  readonly initialSelection: "all" | "end" | "seed";
};

export type GridEditorFocusTarget =
  | HTMLInputElement
  | HTMLSelectElement
  | HTMLTextAreaElement;

export type GridEditorRenderContext<Row> = {
  readonly activation: GridEditorActivation;
  readonly cancel: () => void;
  readonly commit: (draftValueOverride?: unknown) => Promise<void>;
  readonly draftValue: unknown;
  readonly outcome: GridEditCommitOutcome | null;
  readonly pending: boolean;
  readonly row: Row;
  readonly setDraftValue: (value: unknown) => void;
  readonly focusTargetRef: RefCallback<GridEditorFocusTarget>;
  readonly target: GridCellTarget;
};

export type GridEditorAdapter<Row> = {
  /** Draft value used by Backspace/Delete entry when this field permits clear. */
  readonly clearDraftValue?: unknown;
  readonly commit: (
    intent: GridEditCommitIntent<Row>,
  ) => Promise<GridEditCommitOutcome>;
  readonly initialDraftValue: (row: Row) => unknown;
  readonly renderEditor: (context: GridEditorRenderContext<Row>) => ReactNode;
};

export type GridGroupingScalar = boolean | number | string | null;

export type GridGroupingDescriptor<Row> = {
  readonly fieldKey: string;
  readonly label?: string | undefined;
  readonly getValue: (row: Row) => GridGroupingScalar;
  readonly formatLabel: (value: GridGroupingScalar) => string | null;
  readonly getTestId?:
    | ((
        fieldKey: string,
        value: GridGroupingScalar,
        label: string | null,
      ) => string | undefined)
    | undefined;
};

export type GridHandle = {
  readonly activateEdit: (
    anchor: GridCellAnchor,
    seed?: { readonly value: unknown } | undefined,
  ) => boolean;
  readonly cancelEdit: (anchor: GridCellAnchor) => boolean;
  readonly focusAnchor: (anchor: GridCellAnchor) => boolean;
  readonly focusRoot: () => boolean;
  readonly getScrollElement: () => HTMLDivElement | null;
  readonly scrollToAnchor: (anchor: GridCellAnchor) => boolean;
};

export type GridCellCopyIntent = {
  readonly anchor: GridCellAnchor;
  readonly range: GridCellRange;
  readonly expandedRange: GridExpandedCellRange;
};

export type GridCellMutationIntent = {
  readonly clipboardText?: string | undefined;
  readonly range?: GridCellRange | undefined;
  readonly targetResolution?: GridPasteTargetResolution | undefined;
  readonly target: GridCellTarget;
};

export type GridFillIntent = {
  readonly range: GridCellRange;
  readonly source: GridCellTarget;
  readonly target: GridCellTarget;
  readonly targets: readonly GridCellTarget[];
};

export type ResolveGridCellRangeProps<Row> = {
  readonly columns: readonly GridColumn<Row>[];
  readonly presentationRows: readonly GridPresentationRow<Row>[];
  readonly range: GridCellRange;
};

export type GridCellSelection = {
  readonly fieldKey: string;
  readonly rowIndex: number;
};

export type GridNavigationKey =
  | "ArrowDown"
  | "ArrowLeft"
  | "ArrowRight"
  | "ArrowUp"
  | "End"
  | "Enter"
  | "Home"
  | "PageDown"
  | "PageUp"
  | "Tab";

export type GridNavigationIntent = {
  readonly ctrlOrMetaKey?: boolean | undefined;
  readonly key: GridNavigationKey;
  /** Semantic page step, normally visible body rows minus one. */
  readonly pageSize?: number | undefined;
  readonly shiftKey?: boolean | undefined;
};

export type ResolveGridCellAnchorProps<Row> = {
  readonly columns: readonly GridColumn<Row>[];
  readonly presentationRows: readonly GridPresentationRow<Row>[];
  readonly selection: GridCellSelection;
  readonly surface: GridSurfaceIdentity;
};

export type NavigateGridCellAnchorProps<Row> = {
  readonly columns: readonly GridColumn<Row>[];
  readonly current: GridCellAnchor;
  readonly intent: GridNavigationIntent;
  readonly presentationRows: readonly GridPresentationRow<Row>[];
};

export type GridPasteCreateRowTarget = {
  readonly createIndex: number;
  readonly kind: "create";
  readonly surface: Extract<
    GridSurfaceIdentity,
    { readonly kind: "view_schema" }
  >;
};

export type GridPasteRecordRowTarget = {
  readonly mutationIdentity: GridMutationIdentity;
  readonly kind: "record";
  readonly rowIdentity: Extract<
    GridRowIdentity,
    { readonly kind: "core_record" }
  >;
  readonly surface: Extract<
    GridSurfaceIdentity,
    { readonly kind: "view_schema" }
  >;
};

export type GridPasteRowTarget =
  | GridPasteCreateRowTarget
  | GridPasteRecordRowTarget;

export type ResolveGridPasteTargetsProps<Row> = {
  readonly allowCreateRows?: boolean | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly current: GridCellAnchor;
  readonly pastedColumnCount: number;
  readonly pastedRowCount: number;
  readonly presentationRows: readonly GridPresentationRow<Row>[];
};

export type GridPasteTargetResolution = {
  readonly columns: readonly string[];
  readonly rowTargets: readonly GridPasteRowTarget[];
};

type BuildGridPresentationRowsProps<Row> = {
  readonly grouping?: GridGroupingDescriptor<Row> | null | undefined;
  readonly rows: readonly GridSemanticRow<Row>[];
};

export const gridUnassignedGroupLabel = "Unassigned";

export function isGridColumnEditable<Row>(column: GridColumn<Row>): boolean {
  return column.contractWritable === true && column.editor !== undefined;
}

export function assertGridRows<Row>(rows: readonly GridDataRow<Row>[]) {
  const seen = new Set<string>();
  for (const row of rows) {
    const key = gridRowIdentityKey(row.rowIdentity);
    if (!isValidGridRowIdentity(row.rowIdentity)) {
      throw new Error(
        "Grid adapter invariant failed: a data row has an invalid semantic identity.",
      );
    }
    if (seen.has(key)) {
      throw new Error(
        "Grid adapter invariant failed: duplicate semantic row identity.",
      );
    }
    seen.add(key);
  }
}

export function buildGridPresentationRows<Row>({
  grouping,
  rows,
}: BuildGridPresentationRowsProps<Row>): readonly GridPresentationRow<Row>[] {
  if (grouping === null || grouping === undefined) {
    return rows.flatMap((row) =>
      row.kind === "draft"
        ? []
        : [
            {
              gridRow: row,
              key: gridRowIdentityKey(row.rowIdentity),
              kind: "data" as const,
            },
          ],
    );
  }

  const groupBy = grouping.fieldKey;
  const buckets: Array<{
    groupValue: GridGroupingScalar;
    groupKeyValue: string;
    groupLabel: string | null;
    rows: Array<GridDataRow<Row>>;
  }> = [];
  const bucketsByKey = new Map<
    string,
    {
      groupKeyValue: string;
      groupLabel: string | null;
      groupValue: GridGroupingScalar;
      rows: Array<GridDataRow<Row>>;
    }
  >();
  for (const row of rows) {
    if (row.kind === "draft") {
      continue;
    }
    const groupValue = grouping.getValue(row.data);
    const nextGroupLabel = normalizeGroupLabel(
      grouping.formatLabel(groupValue),
    );
    const bucketMapKey = encodeGridGroupingScalar(groupValue);
    let bucket = bucketsByKey.get(bucketMapKey);
    if (bucket === undefined) {
      bucket = {
        groupKeyValue: bucketMapKey,
        groupLabel: nextGroupLabel,
        groupValue,
        rows: [],
      };
      bucketsByKey.set(bucketMapKey, bucket);
      buckets.push(bucket);
    }
    bucket.rows.push(row);
  }

  const presentationRows: GridPresentationRow<Row>[] = [];
  for (const bucket of buckets) {
    presentationRows.push({
      groupBy,
      groupLabel: bucket.groupLabel,
      key: `group:${groupBy}:${bucket.groupKeyValue}:0`,
      kind: "group",
      testId:
        grouping.getTestId === undefined
          ? undefined
          : grouping.getTestId(groupBy, bucket.groupValue, bucket.groupLabel),
    });
    for (const row of bucket.rows) {
      presentationRows.push({
        gridRow: row,
        key: gridRowIdentityKey(row.rowIdentity),
        kind: "data",
      });
    }
  }

  return presentationRows;
}

export function resolveGridCellAnchor<Row>({
  columns,
  presentationRows,
  selection,
  surface,
}: ResolveGridCellAnchorProps<Row>): GridCellAnchor | null {
  if (
    !Number.isInteger(selection.rowIndex) ||
    selection.rowIndex < 0 ||
    !columns.some((column) => column.fieldKey === selection.fieldKey)
  ) {
    return null;
  }
  const row = presentationRows[selection.rowIndex];
  if (row === undefined || row.kind !== "data") {
    return null;
  }
  if (!isValidGridRowIdentity(row.gridRow.rowIdentity)) {
    return null;
  }
  return {
    fieldKey: selection.fieldKey,
    rowIdentity: row.gridRow.rowIdentity,
    surface,
  };
}

export function navigateGridCellAnchor<Row>({
  columns,
  current,
  intent,
  presentationRows,
}: NavigateGridCellAnchorProps<Row>): GridCellAnchor | null {
  const dataRows = presentationRows.filter((row) => row.kind === "data");
  const currentRowIndex = dataRows.findIndex((row) =>
    gridRowIdentitiesEqual(row.gridRow.rowIdentity, current.rowIdentity),
  );
  const currentColumnIndex = columns.findIndex(
    (column) => column.fieldKey === current.fieldKey,
  );
  if (currentRowIndex < 0 || currentColumnIndex < 0) {
    return null;
  }

  const target = navigateGridCellCoordinates({
    columnIndex: currentColumnIndex,
    columnCount: columns.length,
    intent,
    rowIndex: currentRowIndex,
    rowCount: dataRows.length,
  });
  if (target === null) {
    return null;
  }
  const targetColumn = columns[target.columnIndex];
  if (targetColumn === undefined) {
    return null;
  }
  const targetRow = dataRows[target.rowIndex];
  if (targetRow === undefined) return null;
  return {
    fieldKey: targetColumn.fieldKey,
    rowIdentity: targetRow.gridRow.rowIdentity,
    surface: current.surface,
  };
}

export function resolveGridCellRange<Row>({
  columns,
  presentationRows,
  range,
}: ResolveGridCellRangeProps<Row>): GridExpandedCellRange | null {
  if (!gridSurfaceIdentitiesEqual(range.start.surface, range.end.surface)) {
    return null;
  }
  const startColumnIndex = columns.findIndex(
    (column) => column.fieldKey === range.start.fieldKey,
  );
  const endColumnIndex = columns.findIndex(
    (column) => column.fieldKey === range.end.fieldKey,
  );
  const startRowIndex = presentationRows.findIndex(
    (row) =>
      row.kind === "data" &&
      gridRowIdentitiesEqual(row.gridRow.rowIdentity, range.start.rowIdentity),
  );
  const endRowIndex = presentationRows.findIndex(
    (row) =>
      row.kind === "data" &&
      gridRowIdentitiesEqual(row.gridRow.rowIdentity, range.end.rowIdentity),
  );
  if (
    startColumnIndex < 0 ||
    endColumnIndex < 0 ||
    startRowIndex < 0 ||
    endRowIndex < 0
  ) {
    return null;
  }
  const firstColumnIndex = Math.min(startColumnIndex, endColumnIndex);
  const lastColumnIndex = Math.max(startColumnIndex, endColumnIndex);
  const firstRowIndex = Math.min(startRowIndex, endRowIndex);
  const lastRowIndex = Math.max(startRowIndex, endRowIndex);
  const fieldKeys = columns
    .slice(firstColumnIndex, lastColumnIndex + 1)
    .map((column) => column.fieldKey);
  const rowIdentities = presentationRows
    .slice(firstRowIndex, lastRowIndex + 1)
    .flatMap((row) => (row.kind === "group" ? [] : [row.gridRow.rowIdentity]));
  return fieldKeys.length === 0 || rowIdentities.length === 0
    ? null
    : { fieldKeys, rowIdentities };
}

export function parseGridClipboardTable(text: string): string[][] {
  const normalized = text.replace(/\r\n/g, "\n").replace(/\r/g, "\n");
  const trimmed = normalized.replace(/\n+$/u, "");
  if (trimmed === "") return [[""]];
  const delimiter = trimmed.includes("\t") ? "\t" : ",";
  const rows: string[][] = [];
  let row: string[] = [];
  let cell = "";
  let quoted = false;
  for (let index = 0; index < trimmed.length; index += 1) {
    const char = trimmed[index];
    if (char === '"') {
      if (quoted && trimmed[index + 1] === '"') {
        cell += '"';
        index += 1;
      } else {
        quoted = !quoted;
      }
      continue;
    }
    if (!quoted && char === delimiter) {
      row.push(cell);
      cell = "";
      continue;
    }
    if (!quoted && char === "\n") {
      row.push(cell);
      rows.push(row);
      row = [];
      cell = "";
      continue;
    }
    cell += char;
  }
  row.push(cell);
  rows.push(row);
  return rows;
}

export function gridClipboardDimensions(text: string): {
  readonly columnCount: number;
  readonly rowCount: number;
} {
  const rows = parseGridClipboardTable(text);
  return {
    columnCount: rows.reduce((max, row) => Math.max(max, row.length), 1),
    rowCount: Math.max(1, rows.length),
  };
}

export function formatGridClipboardTSV(
  values: readonly (readonly unknown[])[],
): string {
  return values
    .map((row) => row.map(formatGridClipboardCell).join("\t"))
    .join("\n");
}

function formatGridClipboardCell(value: unknown): string {
  const text =
    value === null || value === undefined
      ? ""
      : typeof value === "boolean" || typeof value === "number"
        ? String(value)
        : String(value);
  return /[\t\n\r"]/u.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
}

export function resolveGridPasteTargets<Row>({
  allowCreateRows = true,
  columns,
  current,
  pastedColumnCount,
  pastedRowCount,
  presentationRows,
}: ResolveGridPasteTargetsProps<Row>): GridPasteTargetResolution | null {
  if (
    current.rowIdentity.kind !== "core_record" ||
    current.surface.kind !== "view_schema" ||
    current.rowIdentity.recordId.trim() === "" ||
    current.fieldKey.trim() === "" ||
    !Number.isInteger(pastedRowCount) ||
    pastedRowCount < 1 ||
    !Number.isInteger(pastedColumnCount) ||
    pastedColumnCount < 1
  ) {
    return null;
  }
  const startColumnIndex = columns.findIndex(
    (column) => column.fieldKey === current.fieldKey,
  );
  if (startColumnIndex < 0) {
    return null;
  }
  const targetColumns = columns
    .slice(startColumnIndex, startColumnIndex + pastedColumnCount)
    .map((column) => column.fieldKey);
  if (targetColumns.length !== pastedColumnCount) {
    return null;
  }

  const startRowIndex = presentationRows.findIndex(
    (row) =>
      row.kind === "data" &&
      gridRowIdentitiesEqual(row.gridRow.rowIdentity, current.rowIdentity),
  );
  if (startRowIndex < 0) {
    return null;
  }

  const rowTargets: GridPasteRowTarget[] = [];
  let createIndex = 0;
  for (let offset = 0; offset < pastedRowCount; offset += 1) {
    const presentationRow = presentationRows[startRowIndex + offset];
    if (presentationRow === undefined) {
      if (!allowCreateRows) {
        return null;
      }
      rowTargets.push({
        createIndex,
        kind: "create",
        surface: current.surface,
      });
      createIndex += 1;
      continue;
    }
    if (presentationRow.kind !== "data") {
      return null;
    }
    const { mutationIdentity, rowIdentity } = presentationRow.gridRow;
    if (
      rowIdentity.kind !== "core_record" ||
      rowIdentity.recordId.trim() === "" ||
      mutationIdentity === undefined
    ) {
      return null;
    }
    rowTargets.push({
      kind: "record",
      mutationIdentity,
      rowIdentity,
      surface: current.surface,
    });
  }

  return {
    columns: targetColumns,
    rowTargets,
  };
}

export function gridRowIdentitiesEqual(
  left: GridRowIdentity,
  right: GridRowIdentity,
): boolean {
  return gridRowIdentityKey(left) === gridRowIdentityKey(right);
}

export function gridSurfaceIdentitiesEqual(
  left: GridSurfaceIdentity,
  right: GridSurfaceIdentity,
): boolean {
  return gridSurfaceIdentityKey(left) === gridSurfaceIdentityKey(right);
}

/** Package-internal key used only to bridge semantic identities to the vendor. */
export function gridRowIdentityKey(identity: GridRowIdentity): string {
  return identity.kind === "core_record"
    ? identity.recordId
    : JSON.stringify([
        identity.kind,
        identity.extensionProfileId,
        identity.resourceKind,
        identity.resourceId,
      ]);
}

/** Package-internal key used only to scope semantic state. */
export function gridSurfaceIdentityKey(identity: GridSurfaceIdentity): string {
  return identity.kind === "view_schema"
    ? JSON.stringify([identity.kind, identity.viewSchemaId])
    : JSON.stringify([
        identity.kind,
        identity.extensionProfileId,
        identity.workspaceKey,
        identity.gridSchemaId,
      ]);
}

function isValidGridRowIdentity(identity: GridRowIdentity): boolean {
  return identity.kind === "core_record"
    ? identity.recordId.trim() !== ""
    : identity.extensionProfileId.trim() !== "" &&
        identity.resourceKind.trim() !== "" &&
        identity.resourceId.trim() !== "";
}

function navigateGridCellCoordinates({
  columnCount,
  columnIndex,
  intent,
  rowCount,
  rowIndex,
}: {
  readonly columnCount: number;
  readonly columnIndex: number;
  readonly intent: GridNavigationIntent;
  readonly rowCount: number;
  readonly rowIndex: number;
}): { columnIndex: number; rowIndex: number } | null {
  let nextColumnIndex = columnIndex;
  let nextRowIndex = rowIndex;
  const pageSize = Math.max(1, Math.floor(intent.pageSize ?? 1));
  switch (intent.key) {
    case "ArrowDown":
      nextRowIndex += 1;
      break;
    case "ArrowUp":
      nextRowIndex -= 1;
      break;
    case "ArrowLeft":
      nextColumnIndex -= 1;
      break;
    case "ArrowRight":
      nextColumnIndex += 1;
      break;
    case "PageDown":
      nextRowIndex = Math.min(rowCount - 1, rowIndex + pageSize);
      break;
    case "PageUp":
      nextRowIndex = Math.max(0, rowIndex - pageSize);
      break;
    case "Home":
      if (intent.ctrlOrMetaKey === true) nextRowIndex = 0;
      nextColumnIndex = 0;
      break;
    case "End":
      if (intent.ctrlOrMetaKey === true) nextRowIndex = rowCount - 1;
      nextColumnIndex = columnCount - 1;
      break;
    case "Enter":
      nextRowIndex += intent.shiftKey === true ? -1 : 1;
      break;
    case "Tab":
      nextColumnIndex += intent.shiftKey === true ? -1 : 1;
      break;
  }
  if (
    nextRowIndex < 0 ||
    nextRowIndex >= rowCount ||
    nextColumnIndex < 0 ||
    nextColumnIndex >= columnCount
  ) {
    return null;
  }
  return {
    columnIndex: nextColumnIndex,
    rowIndex: nextRowIndex,
  };
}

function normalizeGroupLabel(value: string | null | undefined): string | null {
  if (value === null || value === undefined) {
    return null;
  }
  const normalized = value.trim();
  return normalized === "" ? null : normalized;
}

function encodeGridGroupingScalar(value: GridGroupingScalar): string {
  if (value === null) {
    return "n:null";
  }
  if (typeof value === "boolean") {
    return value ? "b:true" : "b:false";
  }
  if (typeof value === "number") {
    if (!Number.isFinite(value)) {
      throw new Error("Grid grouping values must be finite numbers.");
    }
    return `d:${String(value)}`;
  }
  return `s:${value}`;
}
