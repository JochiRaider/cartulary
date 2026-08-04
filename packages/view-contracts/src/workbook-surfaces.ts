import type { ViewSchemaRegistryEntry } from "@cartulary/protocol-ts/view-schemas";
import { viewSchemaRegistry } from "@cartulary/protocol-ts/view-schemas";
import type {
  ViewContract,
  WorkbookSurfaceContract,
  WorkbookSurfaceStatus,
} from "./types.js";
import { listViewContracts } from "./view-contracts.js";

const workbookSurfaceRegistryEntries: readonly ViewSchemaRegistryEntry[] =
  viewSchemaRegistry.view_schemas;

function sameOrderedValues(
  left: readonly string[],
  right: readonly string[],
): boolean {
  return (
    left.length === right.length &&
    left.every((value, index) => value === right[index])
  );
}

export function buildWorkbookSurfaceContracts(
  sourceContracts: readonly ViewContract[] = listViewContracts(),
): readonly WorkbookSurfaceContract[] {
  const contractsById = new Map(
    sourceContracts.map((contract) => [contract.viewSchemaId, contract]),
  );
  return Object.freeze(
    workbookSurfaceRegistryEntries.flatMap((entry) => {
      const contract = contractsById.get(entry.view_schema_id);
      if (!contract) {
        if (entry.surface_status === "standardized_optional_workbook_surface") {
          return [];
        }
        throw new Error(
          `Missing workbook surface contract: ${entry.view_schema_id}`,
        );
      }
      if (contract.surfaceKind !== entry.surface_kind) {
        throw new Error(
          `Workbook surface ${entry.view_schema_id} has surface_kind ${contract.surfaceKind}, expected ${entry.surface_kind}`,
        );
      }
      if (
        !sameOrderedValues(
          contract.requiredReferencePackKeys,
          entry.required_reference_pack_keys,
        )
      ) {
        throw new Error(
          `Workbook surface ${entry.view_schema_id} required_reference_pack_keys do not match its registry entry`,
        );
      }
      return [
        Object.freeze({
          contract,
          requiredReferencePackKeys: Object.freeze([
            ...entry.required_reference_pack_keys,
          ]),
          sourceRecordTypes: Object.freeze([...entry.source_record_types]),
          surfaceKind: entry.surface_kind,
          surfaceStatus: entry.surface_status,
          title: entry.title,
          viewSchemaId: entry.view_schema_id,
        }),
      ];
    }),
  );
}

const workbookSurfaceContracts = buildWorkbookSurfaceContracts();
const workbookSurfaceContractIndex = new Map(
  workbookSurfaceContracts.map((entry) => [entry.viewSchemaId, entry]),
);

export function listWorkbookSurfaceContracts(): readonly WorkbookSurfaceContract[] {
  return workbookSurfaceContracts;
}

export function getWorkbookSurfaceContract(
  viewSchemaId: string,
): WorkbookSurfaceContract | undefined {
  return workbookSurfaceContractIndex.get(viewSchemaId);
}

export function requireWorkbookSurfaceContract(
  viewSchemaId: string,
): WorkbookSurfaceContract {
  const entry = getWorkbookSurfaceContract(viewSchemaId);
  if (!entry) {
    throw new Error(`Unknown workbook surface contract: ${viewSchemaId}`);
  }
  return entry;
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
