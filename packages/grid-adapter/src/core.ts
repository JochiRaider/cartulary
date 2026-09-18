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
  /** Creation eligibility is independent of existing-record patch eligibility. */
  readonly draftWritable?: boolean | undefined;
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
  readonly maxWidth?: number | undefined;
  /** False when the displayed value includes local, uncommitted authoring. */
  readonly isCellContentCommitted?: ((row: Row) => boolean) | undefined;
  readonly width?: number | undefined;
};

export type GridColumnSizingIntent =
  | {
      readonly kind: "set_width";
      readonly fieldKey: string;
      readonly widthPx: number;
    }
  | { readonly kind: "fit_visible"; readonly fieldKey: string };

export type GridColumnMeasurement =
  | {
      readonly kind: "measured";
      readonly widthPx: number;
      readonly capped: boolean;
      readonly cellCount: number;
    }
  | { readonly kind: "unavailable"; readonly reason: string }
  | { readonly kind: "cancelled" };

/** Observational capability: no width store, vendor nodes, or persistence policy. */
export type GridColumnSizingPort = {
  readonly unavailableReason: (fieldKey: string) => string | null;
  readonly measureVisibleContent: (
    fieldKey: string,
    options: { readonly signal: AbortSignal },
  ) => Promise<GridColumnMeasurement>;
  readonly subscribe: (listener: () => void) => () => void;
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
  readonly findMatch?: "match" | "current" | undefined;
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
  | {
      readonly kind: "initial_loading";
      readonly generationKey: string | number;
      readonly surfaceLabel: string;
    }
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
  /** Explicit adoption; scope changes invalidate selection without remounting editors. */
  readonly cellRangeSelection?:
    | {
        readonly kind: "contiguous";
        readonly scopeKey: string;
        /** Explicitly adopted cyclic traversal and retained keyboard edit entry. */
        readonly keyboardEntry?: "cycle" | undefined;
      }
    | undefined;
  readonly keyboardNavigation?: "region" | "spreadsheet" | undefined;
  readonly accessibleLabel?: string | undefined;
  readonly allowPasteCreateRows?: boolean | undefined;
  readonly activeRowIdentity?: GridRowIdentity | null | undefined;
  readonly actionsColumn?: GridActionsColumn<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly coreRecordBulkSelection?:
    | GridCoreRecordBulkSelection<Row>
    | undefined;
  readonly cellRange?: GridCellRange | null | undefined;
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
  readonly clipboardPaste?: GridClipboardPasteContract | undefined;
  readonly onSortChange?:
    | ((sort: readonly GridSortEntry[]) => void)
    | undefined;
  readonly onColumnReorder?:
    | ((sourceFieldKey: string, targetFieldKey: string) => void)
    | undefined;
  readonly onColumnSizingIntent?:
    | ((intent: GridColumnSizingIntent) => void)
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
      | "clipboardPaste"
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
      readonly clipboardPaste?: never;
    });

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
  readonly focusTargetRef: RefCallback<GridEditorFocusTarget>;
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
  | { readonly kind: "superseded"; readonly message: string }
  | { readonly kind: "validation_error"; readonly message: string }
  | { readonly kind: "conflict"; readonly message: string }
  | { readonly kind: "stale_target"; readonly message: string }
  | { readonly kind: "rejected_mutation"; readonly message: string };

export type GridEditorActivation = {
  readonly selectionRange?:
    | { readonly start: number; readonly end: number }
    | undefined;
  readonly source:
    | "clear"
    | "enter"
    | "f2"
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
  /** Declares the mounted primary control; the adapter owns its focus and reveal. */
  readonly focusTargetRef: RefCallback<GridEditorFocusTarget>;
  readonly target: GridCellTarget;
};

export type GridEditorAdapter<Row> = {
  /** The source owner retains raw authoring; the adapter owns mounted editor mechanics. */
  readonly retainDraft?:
    | ((row: Row, value: unknown, target: GridCellTarget) => void)
    | undefined;
  readonly discardDraft?: ((row: Row) => void) | undefined;
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
  readonly compareValues?:
    | ((left: GridGroupingScalar, right: GridGroupingScalar) => number)
    | undefined;
  readonly getTestId?:
    | ((
        fieldKey: string,
        value: GridGroupingScalar,
        label: string | null,
      ) => string | undefined)
    | undefined;
};

export type GridFocusTarget =
  | { readonly kind: "draft"; readonly fieldKey: string }
  | { readonly kind: "cell"; readonly anchor: GridCellAnchor }
  | { readonly kind: "root" };

export type GridFocusResult = "focused" | "unavailable" | "cancelled";

/** Current presented membership, including virtualized cells; no values or coordinates. */
export type GridPresentationSnapshot = {
  readonly surface: GridSurfaceIdentity;
  readonly rowIdentities: readonly GridRowIdentity[];
  readonly fieldKeys: readonly string[];
  readonly revision: number;
};

export type GridPresentationPort = {
  readonly getSnapshot: () => GridPresentationSnapshot | null;
  readonly subscribe: (listener: () => void) => () => void;
};

export type GridCellNavigationOptions = {
  readonly signal?: AbortSignal;
  /** Recheck source-owned eligibility after editor settlement, before reveal. */
  readonly isCurrent?: () => boolean;
  readonly beforeFocus?: () => void;
};

export type GridCellNavigationResult = GridFocusResult | "rejected";

export type GridHandle = {
  /** Caller context composed with adapter-owned keyboard guidance. */
  readonly setAccessibleDescription?: (description: string | undefined) => void;
  /** True only for the grid navigation cell itself, never a nested editor/control. */
  readonly ownsNavigationFocus?: (target: EventTarget | null) => boolean;
  readonly presentation?: GridPresentationPort;
  readonly navigateToCell?: (
    anchor: GridCellAnchor,
    options?: GridCellNavigationOptions,
  ) => Promise<GridCellNavigationResult>;
  readonly columnSizing?: GridColumnSizingPort | undefined;
  readonly activateEdit: (
    anchor: GridCellAnchor,
    seed?:
      | {
          readonly value: unknown;
          readonly selectionRange?: GridEditorActivation["selectionRange"];
        }
      | undefined,
  ) => boolean;
  readonly cancelEdit: (anchor: GridCellAnchor) => boolean;
  /** Release editor presentation while leaving retained source drafts untouched. */
  readonly detachEdit?: () => void;
  readonly getActiveCell?: () => GridCellAnchor | null;
  readonly requestFocus: (
    target: GridFocusTarget,
    options?: { readonly signal?: AbortSignal },
  ) => Promise<GridFocusResult>;
  readonly focusAdjacentRegion?: (backwards: boolean) => boolean;
  readonly getAnchorRect: (anchor: GridCellAnchor) => DOMRectReadOnly | null;
  readonly getScrollElement: () => HTMLDivElement | null;
  readonly isAnchorRendered: (anchor: GridCellAnchor) => boolean;
  readonly moveFocus: (
    current: GridCellAnchor,
    intent: GridNavigationIntent,
  ) => GridCellAnchor | null;
  readonly planPasteTargets: (
    current: GridCellAnchor,
    dimensions: GridClipboardDimensions,
  ) => GridPasteTargetResolution | null;
  readonly scrollToAnchor: (anchor: GridCellAnchor) => boolean;
};

export type GridCellCopyIntent = {
  readonly anchor: GridCellAnchor;
  readonly range: GridCellRange;
  readonly expandedRange: GridExpandedCellRange;
};

export type GridClipboardDimensions = {
  /** Optional ordered mapping validated by the source clipboard owner. */
  readonly fieldKeys?: readonly string[] | undefined;
  readonly columnCount: number;
  readonly rowCount: number;
};

export type GridClipboardInput =
  | {
      readonly kind: "scalar";
      readonly rawText: string;
      readonly value: string;
    }
  | {
      readonly fieldKeys?: readonly string[] | undefined;
      readonly format: "csv" | "tsv";
      readonly headerMode?: "auto" | "none";
      readonly kind: "table";
      readonly rawText: string;
      readonly values: readonly (readonly string[])[];
    };

export type GridCellPasteIntent = {
  readonly input: GridClipboardInput;
  readonly range: GridCellRange;
  readonly targetResolution: GridPasteTargetResolution;
  readonly target: GridCellTarget;
};

export type GridClipboardPasteContract = {
  readonly decode: (
    offered: import("./clipboardCodec").ClipboardRepresentations,
  ) => import("./clipboardCodec").ClipboardDecodeResult;
  readonly onPaste: (
    intent: GridCellPasteIntent,
  ) => boolean | undefined | Promise<boolean | undefined>;
  readonly onError?: (message: string) => void;
};

export type GridFillIntent = {
  readonly range: GridCellRange;
  readonly source: GridCellTarget;
  readonly target: GridCellTarget;
  readonly targets: readonly GridCellTarget[];
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

export type GridPasteTargetResolution = {
  readonly columns: readonly string[];
  readonly rowTargets: readonly GridPasteRowTarget[];
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
