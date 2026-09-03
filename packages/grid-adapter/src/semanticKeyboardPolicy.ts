import type {
  GridCellAnchor,
  GridCellRange,
  GridColumn,
  GridDataRow,
  GridEditorActivation,
  GridNavigationKey,
} from "./core";
import { isGridColumnEditable } from "./core";
import type { PendingEditorSeed } from "./editorSessionPolicy";
import {
  coreRowVersion,
  type GridSemanticPresentationModel,
  navigateSemanticPresentation,
} from "./semanticPresentation";

export type NormalizedGridKey = {
  readonly altKey: boolean;
  readonly ctrlOrMetaKey: boolean;
  readonly key: string;
  readonly shiftKey: boolean;
};

export type SemanticGridDecision =
  | { readonly kind: "ignore" }
  | { readonly announcement: string; readonly kind: "reject" }
  | { readonly backwards: boolean; readonly kind: "exit_grid" }
  | {
      readonly kind: "begin_edit";
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
  model,
  pageSize,
  range,
  readOnlyLabel,
  row,
}: {
  readonly anchor: GridCellAnchor;
  readonly column: GridColumn<Row> | undefined;
  readonly editable: boolean;
  readonly input: NormalizedGridKey;
  readonly model: GridSemanticPresentationModel<Row>;
  readonly pageSize: number;
  readonly range: GridCellRange | null;
  readonly readOnlyLabel: string;
  readonly row: GridDataRow<Row>;
}): SemanticGridDecision {
  if (input.ctrlOrMetaKey && input.key.toLowerCase() === "d") {
    return { kind: "fill", range };
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
        ? { start: range?.start ?? anchor, end: target }
        : null,
    target,
    timelineMeasurement:
      input.key === "ArrowDown" &&
      !input.shiftKey &&
      !input.ctrlOrMetaKey &&
      isTimelineSummary(anchor),
  };
}

function editDecision<Row>({
  anchor,
  column,
  editable,
  input,
  readOnlyLabel,
  row,
  seed,
}: {
  readonly anchor: GridCellAnchor;
  readonly column: GridColumn<Row> | undefined;
  readonly editable: boolean;
  readonly input: NormalizedGridKey;
  readonly readOnlyLabel: string;
  readonly row: GridDataRow<Row>;
  readonly seed?: unknown;
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
    initialSelection: initialEditSelection(
      hasSeed,
      input.shiftKey,
      timelineMeasurement,
    ),
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
