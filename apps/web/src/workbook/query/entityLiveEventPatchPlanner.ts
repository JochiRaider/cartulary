import {
  normalizeViewRowPatchV1,
  requireViewContract,
} from "@cartulary/view-contracts";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import type { EntityRow } from "../models/entityWorkbookModel";
import { entityRowFromApi } from "../models/entityWorkbookModel";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { applyWorkbookQueryRowPatch } from "./workbookQueryRowPatch";

const contracts = {
  host: requireViewContract(hostsViewSchemaId),
  identity: requireViewContract(identitiesViewSchemaId),
} as const;

export type EntityLiveEventPatchPlan =
  | {
      readonly kind: "apply";
      readonly entityType: EntityRow["entityType"];
      readonly rows: readonly EntityRow[];
    }
  | { readonly kind: "stale_noop" }
  | { readonly kind: "refresh_required" };

function entityTypeForSurface(
  viewSchemaId: string,
): EntityRow["entityType"] | null {
  if (viewSchemaId === hostsViewSchemaId) return "host";
  if (viewSchemaId === identitiesViewSchemaId) return "identity";
  return null;
}

export function planEntityLiveEventPatch(input: {
  readonly hostRows: readonly EntityRow[];
  readonly identityRows: readonly EntityRow[];
  readonly payload: RecordChangedPayload;
  readonly viewSchemaId: string;
}): EntityLiveEventPatchPlan {
  const entityType = entityTypeForSurface(input.viewSchemaId);
  if (entityType === null) return { kind: "refresh_required" };
  const affected = input.payload.affected_views.filter(
    (candidate) => candidate.view_schema_id === input.viewSchemaId,
  );
  const [change] = affected;
  if (
    affected.length !== 1 ||
    change?.change_kind !== "patch" ||
    change.patch_cells === undefined
  ) {
    return { kind: "refresh_required" };
  }
  let patch: ReturnType<typeof normalizeViewRowPatchV1>;
  try {
    patch = normalizeViewRowPatchV1(
      contracts[entityType],
      change.patch_cells,
      "record_changed patch_cells",
    );
  } catch {
    return { kind: "refresh_required" };
  }
  if (
    patch.recordId !== input.payload.record_id ||
    patch.rowVersion !== input.payload.row_version ||
    !Number.isSafeInteger(input.payload.row_version) ||
    input.payload.row_version < 1
  ) {
    return { kind: "refresh_required" };
  }
  const rows = entityType === "host" ? input.hostRows : input.identityRows;
  const existing = rows.find((row) => row.recordId === patch.recordId);
  if (existing === undefined || existing.entityType !== entityType) {
    return { kind: "refresh_required" };
  }
  if (existing.rowVersion >= patch.rowVersion) {
    return { kind: "stale_noop" };
  }
  return {
    kind: "apply",
    entityType,
    rows: rows.map((row) =>
      row.recordId === patch.recordId
        ? entityRowFromApi(
            applyWorkbookQueryRowPatch(row.rawRow, patch),
            entityType,
          )
        : row,
    ),
  };
}
