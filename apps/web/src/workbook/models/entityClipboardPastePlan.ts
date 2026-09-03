import type {
  GridCellPasteIntent,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import type { WorkbookClipboardPasteInput } from "../adapters/WorkbookClipboardPastePort";
import type { EntityRow } from "./entityWorkbookModel";
import {
  workbookPasteColumns,
  workbookPasteResolutionMatchesSurface,
  workbookPasteTargets,
  workbookPasteViewSchemaId,
} from "./workbookClipboardPaste";

export type EntityClipboardPastePlan =
  | {
      readonly fieldKey: string;
      readonly kind: "scalar";
      readonly target: {
        readonly baseRowVersion: number;
        readonly recordId: string;
      };
      readonly value: string;
    }
  | {
      readonly input: WorkbookClipboardPasteInput;
      readonly kind: "batch";
    }
  | {
      readonly kind: "rejected";
      readonly message: string;
    };

type EntityPasteAuthority = {
  readonly canCreateRows: boolean;
  readonly grouped: boolean;
  readonly rows: readonly Pick<EntityRow, "recordId" | "rowVersion">[];
  readonly viewSchemaId: string;
  readonly writableFieldKeys: ReadonlySet<string>;
};

const incompleteTargetsMessage =
  "Paste targets are incomplete or incompatible.";
const changedTargetsMessage =
  "Paste targets changed or are unavailable for this surface.";

function inputValues(
  intent: GridCellPasteIntent,
): readonly (readonly string[])[] {
  return intent.input.kind === "scalar"
    ? [[intent.input.value]]
    : intent.input.values;
}

function resolutionHasInputShape(
  resolution: GridPasteTargetResolution,
  values: readonly (readonly string[])[],
): boolean {
  return (
    resolution.columns.length > 0 &&
    resolution.rowTargets.length === values.length &&
    values.every(
      (row) => row.length > 0 && row.length <= resolution.columns.length,
    )
  );
}

function scalarPastePlan(
  intent: GridCellPasteIntent,
  values: readonly (readonly string[])[],
): EntityClipboardPastePlan | null {
  if (values.length !== 1 || values[0]?.length !== 1) return null;
  const rowTarget = intent.targetResolution.rowTargets[0];
  if (rowTarget?.kind !== "record") {
    return {
      kind: "rejected",
      message: "Scalar paste requires an existing record target.",
    };
  }
  return {
    fieldKey: intent.targetResolution.columns[0] ?? intent.target.fieldKey,
    kind: "scalar",
    target: {
      baseRowVersion: rowTarget.mutationIdentity.baseRowVersion,
      recordId: rowTarget.rowIdentity.recordId,
    },
    value: values[0][0] ?? "",
  };
}

function recordTargetsAreCurrent(
  resolution: GridPasteTargetResolution,
  rows: EntityPasteAuthority["rows"],
): boolean {
  const currentVersions = new Map(
    rows.map((row) => [row.recordId, row.rowVersion]),
  );
  return resolution.rowTargets.every(
    (target) =>
      target.kind === "create" ||
      currentVersions.get(target.rowIdentity.recordId) ===
        target.mutationIdentity.baseRowVersion,
  );
}

function batchPastePlan(
  intent: GridCellPasteIntent,
  authority: EntityPasteAuthority,
): EntityClipboardPastePlan {
  if (authority.grouped) {
    return {
      kind: "rejected",
      message:
        "Rectangular entity creation paste is unavailable while grouped.",
    };
  }
  if (!authority.canCreateRows) {
    return {
      kind: "rejected",
      message: "Row creation is unavailable in the current view mode.",
    };
  }
  const resolution = intent.targetResolution;
  const viewSchemaId = workbookPasteViewSchemaId(authority.viewSchemaId);
  const columns = workbookPasteColumns(resolution.columns);
  const targets = workbookPasteTargets(
    resolution.rowTargets.map(() => ({ kind: "create" })),
  );
  const targetSurfaceMatches =
    intent.target.surface.kind === "view_schema" &&
    intent.target.surface.viewSchemaId === authority.viewSchemaId &&
    workbookPasteResolutionMatchesSurface(resolution, authority.viewSchemaId);
  const columnsAreWritable = resolution.columns.every((fieldKey) =>
    authority.writableFieldKeys.has(fieldKey),
  );
  if (
    viewSchemaId === null ||
    columns === null ||
    targets === null ||
    !targetSurfaceMatches ||
    !columnsAreWritable ||
    !recordTargetsAreCurrent(resolution, authority.rows)
  ) {
    return { kind: "rejected", message: changedTargetsMessage };
  }
  return {
    input: {
      clipboard_text: intent.input.rawText,
      columns,
      format: intent.input.kind === "table" ? intent.input.format : "csv",
      start_field_key: intent.target.fieldKey,
      targets,
      view_schema_id: viewSchemaId,
    },
    kind: "batch",
  };
}

export function entityClipboardPastePlan(
  intent: GridCellPasteIntent,
  authority: EntityPasteAuthority,
): EntityClipboardPastePlan {
  const values = inputValues(intent);
  if (!resolutionHasInputShape(intent.targetResolution, values)) {
    return { kind: "rejected", message: incompleteTargetsMessage };
  }
  return scalarPastePlan(intent, values) ?? batchPastePlan(intent, authority);
}
