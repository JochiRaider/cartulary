import type {
  GridCellAnchor,
  GridCellRange,
  GridColumn,
  GridDataRow,
  GridEditorActivation,
  GridNavigationKey,
} from "./core";
import { gridRowIdentitiesEqual, isGridColumnEditable } from "./core";
import type { PendingEditorSeed } from "./editorSessionPolicy";
import {
  coreRowVersion,
  type GridSemanticPresentationModel,
  navigateSemanticPresentation,
  resolveSemanticCellRange,
  semanticPresentationContainsAnchor,
} from "./semanticPresentation";
import { extendSemanticCellRange } from "./semanticSelectionPolicy";

export type NormalizedGridKey = {
  readonly altKey: boolean;
  readonly ctrlOrMetaKey: boolean;
  readonly key: string;
  readonly shiftKey: boolean;
};

export type SemanticGridDecision =
  | { readonly kind: "ignore" }
  | { readonly kind: "focus_draft"; readonly fieldKey: string }
  | { readonly announcement: string; readonly kind: "reject" }
  | { readonly backwards: boolean; readonly kind: "exit_grid" }
  | { readonly kind: "collapse_range"; readonly target: GridCellAnchor }
  | {
      readonly kind: "begin_edit";
      readonly range: GridCellRange | null;
      readonly seed: PendingEditorSeed;
      readonly timelineMeasurement: boolean;
    }
  | {
      readonly kind: "navigate";
      readonly range: GridCellRange | null;
      readonly target: GridCellAnchor;
      readonly timelineMeasurement: boolean;
    }
  | { readonly kind: "copy" }
  | { readonly kind: "paste" }
  | { readonly kind: "fill"; readonly range: GridCellRange | null };

export function normalizeGridKey(input: {
  readonly altKey: boolean;
  readonly ctrlKey: boolean;
  readonly key: string;
  readonly metaKey: boolean;
  readonly shiftKey: boolean;
}): NormalizedGridKey {
  return {
    altKey: input.altKey,
    ctrlOrMetaKey: input.ctrlKey || input.metaKey,
    key: input.key,
    shiftKey: input.shiftKey,
  };
}

export function decideSemanticGridKey<Row>({
  anchor,
  column,
  editable,
  input,
  keyboardNavigation = "region",
  draftFieldKeys = [],
  model,
  pageSize,
  range,
  rangeKeyboardEntry,
  readOnlyLabel,
  row,
}: {
  readonly anchor: GridCellAnchor;
  readonly column: GridColumn<Row> | undefined;
  readonly editable: boolean;
  readonly input: NormalizedGridKey;
  readonly keyboardNavigation?: "region" | "spreadsheet" | undefined;
  readonly draftFieldKeys?: readonly string[] | undefined;
  readonly model: GridSemanticPresentationModel<Row>;
  readonly pageSize: number;
  readonly range: GridCellRange | null;
  readonly rangeKeyboardEntry?: "cycle" | undefined;
  readonly readOnlyLabel: string;
  readonly row: GridDataRow<Row>;
}): SemanticGridDecision {
  const retainedRange =
    rangeKeyboardEntry === "cycle"
      ? activeMultiCellRange(model, range, anchor)
      : null;
  if (!input.altKey && !input.ctrlOrMetaKey && !input.shiftKey) {
    if (input.key === "Escape" && retainedRange !== null)
      return { kind: "collapse_range", target: anchor };
    if (input.key === "F2" && rangeKeyboardEntry === "cycle")
      return editDecision({
        anchor,
        column,
        editable,
        input,
        readOnlyLabel,
        row,
        range: retainedRange,
      });
  }
  if (input.ctrlOrMetaKey && input.key.toLowerCase() === "d") {
    return { kind: "fill", range };
  }
  if (
    keyboardNavigation === "spreadsheet" &&
    !input.altKey &&
    !input.ctrlOrMetaKey &&
    (input.key === "Enter" || input.key === "Tab")
  ) {
    return decideSpreadsheetNavigation(
      model,
      anchor,
      input,
      draftFieldKeys,
      retainedRange,
    );
  }
  if (input.key === "Tab") {
    return { backwards: input.shiftKey, kind: "exit_grid" };
  }
  if (input.key === "Enter") {
    return editDecision({
      anchor,
      column,
      editable,
      input,
      readOnlyLabel,
      row,
    });
  }
  if (isPrintableGridEntry(input)) {
    return editDecision({
      anchor,
      column,
      editable,
      input,
      readOnlyLabel,
      row,
      seed: input.key,
      range: retainedRange,
    });
  }
  if (input.key === "Backspace" || input.key === "Delete") {
    if (column?.editor?.clearDraftValue === undefined) {
      return {
        announcement: `${column?.label ?? "This field"} cannot be cleared.`,
        kind: "reject",
      };
    }
    return editDecision({
      anchor,
      column,
      editable,
      input,
      readOnlyLabel,
      row,
      seed: column.editor.clearDraftValue,
    });
  }
  if (!isSemanticNavigationKey(input.key)) return { kind: "ignore" };
  const target = navigateSemanticPresentation(model, anchor, {
    ctrlOrMetaKey: input.ctrlOrMetaKey,
    key: input.key,
    pageSize,
    shiftKey: input.shiftKey,
  });
  if (target === null) return { kind: "ignore" };
  return {
    kind: "navigate",
    range:
      input.shiftKey && input.key.startsWith("Arrow")
        ? extendSemanticCellRange(model, anchor, range, target)
        : null,
    target,
    timelineMeasurement:
      input.key === "ArrowDown" &&
      !input.shiftKey &&
      !input.ctrlOrMetaKey &&
      isTimelineSummary(anchor),
  };
}

export function decideSpreadsheetNavigation<Row>(
  model: GridSemanticPresentationModel<Row>,
  anchor: GridCellAnchor,
  input: Pick<NormalizedGridKey, "key" | "shiftKey">,
  draftFieldKeys: readonly string[] = [],
  range: GridCellRange | null = null,
): SemanticGridDecision {
  const retained = activeMultiCellRange(model, range, anchor);
  const members =
    retained === null ? null : resolveSemanticCellRange(model, retained);
  if (members !== null) {
    const rows = members.rowIdentities;
    const fields = members.fieldKeys;
    const row = rows.findIndex((identity) =>
      gridRowIdentitiesEqual(identity, anchor.rowIdentity),
    );
    const column = fields.indexOf(anchor.fieldKey);
    const rowMajor = input.key === "Tab";
    const stride = rowMajor ? fields.length : rows.length;
    const index = rowMajor ? row * stride + column : column * stride + row;
    const size = rows.length * fields.length;
    const next = (index + (input.shiftKey ? -1 : 1) + size) % size;
    const rowIdentity =
      rows[rowMajor ? Math.floor(next / stride) : next % stride];
    const fieldKey =
      fields[rowMajor ? next % stride : Math.floor(next / stride)];
    if (rowIdentity !== undefined && fieldKey !== undefined)
      return {
        kind: "navigate",
        range: retained,
        target: { surface: anchor.surface, rowIdentity, fieldKey },
        timelineMeasurement: false,
      };
  }
  const backwards = input.shiftKey;
  const rowIndex = model.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, anchor.rowIdentity),
  );
  const columnIndex = model.fieldKeys.indexOf(anchor.fieldKey);
  if (rowIndex < 0 || columnIndex < 0) return { kind: "ignore" };
  let nextRow = rowIndex;
  let nextColumn = columnIndex;
  if (input.key === "Enter") nextRow += backwards ? -1 : 1;
  else {
    nextColumn += backwards ? -1 : 1;
    if (nextColumn < 0) {
      nextRow -= 1;
      nextColumn = model.fieldKeys.length - 1;
    } else if (nextColumn === model.fieldKeys.length) {
      nextRow += 1;
      nextColumn = 0;
    }
  }
  const fieldKey = model.fieldKeys[nextColumn];
  const rowIdentity = model.rowIdentities[nextRow];
  if (fieldKey !== undefined && rowIdentity !== undefined) {
    return {
      kind: "navigate",
      range: null,
      target: { ...anchor, fieldKey, rowIdentity },
      timelineMeasurement: false,
    };
  }
  if (nextRow === model.rowIdentities.length && draftFieldKeys.length > 0) {
    const draftField =
      input.key === "Enter" && draftFieldKeys.includes(anchor.fieldKey)
        ? anchor.fieldKey
        : draftFieldKeys[0];
    if (draftField !== undefined)
      return { kind: "focus_draft", fieldKey: draftField };
  }
  return input.key === "Tab"
    ? { kind: "exit_grid", backwards }
    : { kind: "ignore" };
}

export function activeMultiCellRange<Row>(
  model: GridSemanticPresentationModel<Row>,
  range: GridCellRange | null,
  active: GridCellAnchor,
): GridCellRange | null {
  if (range === null || !semanticPresentationContainsAnchor(model, active))
    return null;
  const members = resolveSemanticCellRange(model, range);
  return members !== null &&
    members.fieldKeys.length * members.rowIdentities.length > 1 &&
    members.fieldKeys.includes(active.fieldKey) &&
    members.rowIdentities.some((identity) =>
      gridRowIdentitiesEqual(identity, active.rowIdentity),
    )
    ? range
    : null;
}

function editDecision<Row>({
  anchor,
  column,
  editable,
  input,
  readOnlyLabel,
  row,
  seed,
  range = null,
}: {
  readonly anchor: GridCellAnchor;
  readonly column: GridColumn<Row> | undefined;
  readonly editable: boolean;
  readonly input: NormalizedGridKey;
  readonly readOnlyLabel: string;
  readonly row: GridDataRow<Row>;
  readonly seed?: unknown;
  readonly range?: GridCellRange | null;
}): SemanticGridDecision {
  const rejection = editRejection(column, editable, readOnlyLabel);
  if (rejection !== null) return { announcement: rejection, kind: "reject" };
  const baseRowVersion = coreRowVersion(row);
  if (baseRowVersion === null) {
    return { announcement: "This row cannot be edited.", kind: "reject" };
  }
  const hasSeed = seed !== undefined;
  const timelineMeasurement =
    input.key === "Enter" && isTimelineSummary(anchor);
  return {
    kind: "begin_edit",
    range,
    seed: {
      activation: editActivation(input, hasSeed, timelineMeasurement),
      anchor,
      baseRowVersion,
      hasValue: hasSeed,
      value: seed,
    },
    timelineMeasurement,
  };
}

function editRejection<Row>(
  column: GridColumn<Row> | undefined,
  editable: boolean,
  readOnlyLabel: string,
): string | null {
  if (!editable) return readOnlyLabel;
  if (column === undefined || !isGridColumnEditable(column)) {
    return `${column?.label ?? "Cell"} is read-only.`;
  }
  return null;
}

function editActivation(
  input: NormalizedGridKey,
  hasSeed: boolean,
  timelineMeasurement: boolean,
): GridEditorActivation {
  return {
    initialSelection:
      input.key === "F2"
        ? "end"
        : initialEditSelection(hasSeed, input.shiftKey, timelineMeasurement),
    source: editSource(input, hasSeed),
  };
}

function initialEditSelection(
  hasSeed: boolean,
  shiftKey: boolean,
  timelineMeasurement: boolean,
): GridEditorActivation["initialSelection"] {
  if (hasSeed) return "seed";
  return shiftKey || timelineMeasurement ? "end" : "all";
}

function editSource(
  input: NormalizedGridKey,
  hasSeed: boolean,
): GridEditorActivation["source"] {
  if (input.key === "F2") return "f2";
  if (!hasSeed) return input.shiftKey ? "shift_enter" : "enter";
  return input.key === "Backspace" || input.key === "Delete"
    ? "clear"
    : "printable";
}

function isPrintableGridEntry(input: NormalizedGridKey): boolean {
  return input.key.length === 1 && !input.altKey && !input.ctrlOrMetaKey;
}

function isSemanticNavigationKey(
  key: string,
): key is Exclude<GridNavigationKey, "Enter" | "Tab"> {
  return (
    key === "ArrowDown" ||
    key === "ArrowLeft" ||
    key === "ArrowRight" ||
    key === "ArrowUp" ||
    key === "End" ||
    key === "Home" ||
    key === "PageDown" ||
    key === "PageUp"
  );
}

function isTimelineSummary(anchor: GridCellAnchor): boolean {
  return (
    anchor.surface.kind === "view_schema" &&
    anchor.surface.viewSchemaId === "cartulary.view.timeline.v2" &&
    anchor.fieldKey === "timeline.activity_synopsis_text"
  );
}
