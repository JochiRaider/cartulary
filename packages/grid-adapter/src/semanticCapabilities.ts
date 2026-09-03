import type {
  GridInteractionMode,
  GridSurfaceIdentity,
  SemanticDataGridProps,
} from "./core";

export type SemanticGridCapabilities<Row> = {
  readonly allowPasteCreateRows: boolean;
  readonly bulkSelection:
    | SemanticDataGridProps<Row>["coreRecordBulkSelection"]
    | undefined;
  readonly draftRow: SemanticDataGridProps<Row>["draftRow"] | undefined;
  readonly editable: boolean;
  readonly interactionMode: GridInteractionMode;
};

type SemanticGridCapabilityInput<Row> = Pick<
  SemanticDataGridProps<Row>,
  | "actionsColumn"
  | "allowPasteCreateRows"
  | "clipboardPaste"
  | "coreRecordBulkSelection"
  | "draftRow"
  | "interactionMode"
  | "onFillCells"
  | "surface"
>;

export function resolveSemanticGridCapabilities<Row>(
  input: SemanticGridCapabilityInput<Row>,
): SemanticGridCapabilities<Row> {
  assertAdmittedCapabilities(input);
  const interactionMode =
    input.interactionMode ??
    (input.surface.kind === "extension_grid"
      ? { kind: "read_only" as const, label: "Read only" }
      : { kind: "editable" as const });
  const editable = interactionMode.kind === "editable";
  return {
    allowPasteCreateRows: editable && input.allowPasteCreateRows === true,
    bulkSelection: editable ? input.coreRecordBulkSelection : undefined,
    draftRow: editable ? input.draftRow : undefined,
    editable,
    interactionMode,
  };
}

function assertAdmittedCapabilities<Row>(
  input: SemanticGridCapabilityInput<Row>,
): void {
  if (
    input.surface.kind !== "extension_grid" ||
    !hasCoreMutationCapability(input)
  ) {
    return;
  }
  throw new Error(
    "Extension grid surfaces cannot enable Core mutation or bulk-selection capabilities.",
  );
}

function hasCoreMutationCapability<Row>(
  input: SemanticGridCapabilityInput<Row>,
): boolean {
  return (
    input.actionsColumn !== undefined ||
    input.allowPasteCreateRows === true ||
    input.coreRecordBulkSelection !== undefined ||
    input.draftRow !== undefined ||
    input.interactionMode?.kind === "editable" ||
    input.onFillCells !== undefined ||
    input.clipboardPaste !== undefined
  );
}

export function isSameViewSchemaSurface(
  left: GridSurfaceIdentity,
  right: GridSurfaceIdentity,
): boolean {
  return (
    left.kind === "view_schema" &&
    right.kind === "view_schema" &&
    left.viewSchemaId === right.viewSchemaId
  );
}
