import type {
  GridClipboardInput,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import type { WorkbookClipboardPasteInput } from "../../adapters/WorkbookClipboardPastePort";
import { workbookPasteResolutionMatchesSurface } from "../../models/workbookClipboardPaste";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { timelineScalarBindingForField } from "./timelineFieldRegistry";

const timelineContract = requireViewContract(timelineViewSchemaId);

export type TimelinePasteAuthority = {
  readonly canCreateRows: boolean;
  readonly editable: boolean;
  readonly grouped: boolean;
};

export type TimelinePastePlanAdmission =
  | { readonly kind: "accepted" }
  | {
      readonly kind: "rejected";
      readonly reason:
        | "create_unavailable"
        | "grouped_create"
        | "invalid_fields"
        | "invalid_shape"
        | "read_only"
        | "wrong_surface";
    };

export function timelinePastePlanAdmission(
  resolution: GridPasteTargetResolution,
  input: Extract<GridClipboardInput, { readonly kind: "table" }>,
  authority: TimelinePasteAuthority,
): TimelinePastePlanAdmission {
  if (!authority.editable) {
    return { kind: "rejected", reason: "read_only" };
  }
  if (
    resolution.rowTargets.length !== input.values.length ||
    input.values.length === 0 ||
    input.values.some(
      (row) => row.length === 0 || row.length > resolution.columns.length,
    )
  ) {
    return { kind: "rejected", reason: "invalid_shape" };
  }
  if (
    !workbookPasteResolutionMatchesSurface(resolution, timelineViewSchemaId)
  ) {
    return { kind: "rejected", reason: "wrong_surface" };
  }
  if (
    resolution.columns.some(
      (fieldKey) =>
        timelineScalarBindingForField(fieldKey) === null ||
        timelineContract.fieldMap[fieldKey]?.gridEditable !== true,
    )
  ) {
    return { kind: "rejected", reason: "invalid_fields" };
  }
  const createsRows = resolution.rowTargets.some(
    (target) => target.kind === "create",
  );
  if (createsRows && authority.grouped) {
    return { kind: "rejected", reason: "grouped_create" };
  }
  if (createsRows && !authority.canCreateRows) {
    return { kind: "rejected", reason: "create_unavailable" };
  }
  return { kind: "accepted" };
}

export function timelinePasteTargetPlansMatch(
  left: GridPasteTargetResolution,
  right: GridPasteTargetResolution,
): boolean {
  if (
    left.columns.length !== right.columns.length ||
    left.rowTargets.length !== right.rowTargets.length ||
    left.columns.some((column, index) => column !== right.columns[index])
  ) {
    return false;
  }
  return left.rowTargets.every((target, index) => {
    const current = right.rowTargets[index];
    if (
      current === undefined ||
      target.kind !== current.kind ||
      target.surface.viewSchemaId !== current.surface.viewSchemaId
    ) {
      return false;
    }
    return target.kind === "create"
      ? current.kind === "create" && target.createIndex === current.createIndex
      : current.kind === "record" &&
          target.rowIdentity.recordId === current.rowIdentity.recordId;
  });
}

export function timelinePasteRequestTargetsMatchResolution(
  targets: readonly WorkbookClipboardPasteInput["targets"][number][],
  resolution: GridPasteTargetResolution,
): boolean {
  if (targets.length !== resolution.rowTargets.length) return false;
  return targets.every((target, index) => {
    const current = resolution.rowTargets[index];
    if (current === undefined || target.kind !== current.kind) return false;
    return target.kind === "create"
      ? current.kind === "create"
      : current.kind === "record" &&
          target.record_id === current.rowIdentity.recordId &&
          target.base_row_version === current.mutationIdentity.baseRowVersion;
  });
}
