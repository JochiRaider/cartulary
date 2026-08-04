import type { ViewSchemaRegistryEntry } from "@cartulary/protocol-ts/view-schemas";
import { viewSchemaRegistry } from "@cartulary/protocol-ts/view-schemas";

import {
  listProjectedViewContracts,
  listProjectedWorkbookSurfaceContracts,
} from "./projection.js";
import type {
  ViewContract,
  WorkbookSurfaceContract,
  WorkbookSurfaceStatus,
} from "./types.js";

const contracts = listProjectedViewContracts();
const contractIndex = Object.freeze(
  Object.fromEntries(
    contracts.map((contract) => [contract.viewSchemaId, contract]),
  ) as Record<string, ViewContract>,
);

export function listViewContracts(): readonly ViewContract[] {
  return contracts;
}

export function getViewContract(
  viewSchemaId: string,
): ViewContract | undefined {
  return contractIndex[viewSchemaId];
}

export function requireViewContract(viewSchemaId: string): ViewContract {
  const contract = getViewContract(viewSchemaId);
  if (!contract) {
    throw new Error(`Unknown view schema contract: ${viewSchemaId}`);
  }
  return contract;
}

export function resolveHeaderSortFieldKey(
  contract: ViewContract,
  fieldKey: string,
): string | null {
  const field = contract.fieldMap[fieldKey];
  if (!field) {
    return null;
  }
  return field.headerSortFieldKey ?? field.fieldKey;
}

const workbookSurfaceRegistryEntries: readonly ViewSchemaRegistryEntry[] =
  viewSchemaRegistry.view_schemas;
const workbookSurfaceContracts = listProjectedWorkbookSurfaceContracts();

export function listWorkbookSurfaceContracts(): readonly WorkbookSurfaceContract[] {
  return workbookSurfaceContracts;
}

function requiredRegistryViewSchemaId(viewSchemaId: string): string {
  const entry = workbookSurfaceRegistryEntries.find(
    (candidate) => candidate.view_schema_id === viewSchemaId,
  );
  if (!entry) {
    throw new Error(`Missing workbook surface registry entry: ${viewSchemaId}`);
  }
  return entry.view_schema_id;
}

export const timelineViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.timeline.v2",
);
export const hostsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.hosts.v1",
);
export const identitiesViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.identities.v1",
);
export const evidenceViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.evidence.v1",
);
export const notesViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.notes.v1",
);
export const indicatorsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.indicators.v1",
);
export const assessmentsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.assessments.v1",
);
export const taskRequestsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.task_requests.v1",
);
export const decisionsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.decisions.v1",
);
export const partiesViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.parties.v1",
);
export const commLogViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.comm_log.v1",
);
export const handoffViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.handoff.v1",
);
export const statusReviewViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.status_review.v1",
);
export const lessonViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.lesson.v1",
);
export const findingsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.findings.v1",
);
export const investigativeQueriesViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.investigative_queries.v1",
);
export const forensicKeywordsViewSchemaId = requiredRegistryViewSchemaId(
  "cartulary.view.forensic_keywords.v1",
);

function workbookSurfaceIdsByStatus(
  status: WorkbookSurfaceStatus,
): readonly string[] {
  return Object.freeze(
    workbookSurfaceRegistryEntries
      .filter((entry) => entry.surface_status === status)
      .map((entry) => entry.view_schema_id),
  );
}

export const requiredBuiltInWorkbookSurfaceIds = workbookSurfaceIdsByStatus(
  "required_built_in_sheet",
);
export const requiredSystemWorkbookSurfaceIds = workbookSurfaceIdsByStatus(
  "required_system_view",
);
export const optionalStandardizedWorkbookSurfaceIds =
  workbookSurfaceIdsByStatus("standardized_optional_workbook_surface");
