import {
  listViewContracts,
  type ViewContract,
} from "@cartulary/view-contracts";
import type { SystemViewSwitcherGroupToken } from "@cartulary/ui-contracts";

export const timelineViewSchemaId = "cartulary.view.timeline.v1";
export const hostsViewSchemaId = "cartulary.view.hosts.v1";
export const identitiesViewSchemaId = "cartulary.view.identities.v1";
export const evidenceViewSchemaId = "cartulary.view.evidence.v1";
export const notesViewSchemaId = "cartulary.view.notes.v1";
export const indicatorsViewSchemaId = "cartulary.view.indicators.v1";
export const assessmentsViewSchemaId = "cartulary.view.assessments.v1";
export const taskRequestsViewSchemaId = "cartulary.view.task_requests.v1";
export const decisionsViewSchemaId = "cartulary.view.decisions.v1";
export const partiesViewSchemaId = "cartulary.view.parties.v1";
export const commLogViewSchemaId = "cartulary.view.comm_log.v1";
export const handoffViewSchemaId = "cartulary.view.handoff.v1";
export const statusReviewViewSchemaId = "cartulary.view.status_review.v1";
export const lessonViewSchemaId = "cartulary.view.lesson.v1";
export const findingsViewSchemaId = "cartulary.view.findings.v1";
export const investigativeQueriesViewSchemaId =
  "cartulary.view.investigative_queries.v1";
export const forensicKeywordsViewSchemaId =
  "cartulary.view.forensic_keywords.v1";

export const requiredBuiltInWorkbookSurfaceIds = [
  timelineViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  evidenceViewSchemaId,
  notesViewSchemaId,
] as const;

export const requiredSystemWorkbookSurfaceIds = [
  indicatorsViewSchemaId,
  assessmentsViewSchemaId,
  taskRequestsViewSchemaId,
  decisionsViewSchemaId,
  partiesViewSchemaId,
  commLogViewSchemaId,
  handoffViewSchemaId,
  statusReviewViewSchemaId,
  lessonViewSchemaId,
] as const;

export const optionalStandardizedWorkbookSurfaceIds = [
  findingsViewSchemaId,
  investigativeQueriesViewSchemaId,
  forensicKeywordsViewSchemaId,
] as const;

export type BuiltInWorkbookSurfaceId =
  (typeof requiredBuiltInWorkbookSurfaceIds)[number];

export type WorkbookSurfaceStatus =
  | "required_built_in_sheet"
  | "required_system_view"
  | "standardized_optional_workbook_surface";

export type WorkbookSurfaceKind = "built_in_sheet" | "system_view";

export type WorkbookSurfaceRegistryEntry = {
  readonly contract: ViewContract;
  readonly requiredReferencePackKeys: readonly string[];
  readonly surfaceKind: WorkbookSurfaceKind;
  readonly surfaceStatus: WorkbookSurfaceStatus;
  readonly viewSchemaId: string;
};

export type SystemWorkbookSurfaceGroup = {
  readonly entries: readonly WorkbookSurfaceRegistryEntry[];
  readonly label: string;
  readonly token: SystemViewSwitcherGroupToken;
};

const builtInSurfaceIdSet = new Set<string>(requiredBuiltInWorkbookSurfaceIds);

const systemWorkbookSurfaceGroupDefinitions = [
  {
    label: "Scope and assessment",
    token: "scope-assessment",
    viewSchemaIds: [
      indicatorsViewSchemaId,
      assessmentsViewSchemaId,
      partiesViewSchemaId,
    ],
  },
  {
    label: "Coordination",
    token: "coordination",
    viewSchemaIds: [
      taskRequestsViewSchemaId,
      decisionsViewSchemaId,
      commLogViewSchemaId,
      handoffViewSchemaId,
    ],
  },
  {
    label: "Review and learning",
    token: "review-learning",
    viewSchemaIds: [statusReviewViewSchemaId, lessonViewSchemaId],
  },
  {
    label: "Optional artifact surfaces",
    token: "optional-artifact-surfaces",
    viewSchemaIds: [
      findingsViewSchemaId,
      investigativeQueriesViewSchemaId,
      forensicKeywordsViewSchemaId,
    ],
  },
] as const satisfies ReadonlyArray<{
  readonly label: string;
  readonly token: SystemViewSwitcherGroupToken;
  readonly viewSchemaIds: readonly string[];
}>;

function contractIndex(
  contracts: readonly ViewContract[],
): ReadonlyMap<string, ViewContract> {
  return new Map(
    contracts.map((contract) => [contract.viewSchemaId, contract]),
  );
}

function expectedSurfaceKind(
  status: WorkbookSurfaceStatus,
): WorkbookSurfaceKind {
  return status === "required_built_in_sheet"
    ? "built_in_sheet"
    : "system_view";
}

function registryEntry(
  contractsById: ReadonlyMap<string, ViewContract>,
  viewSchemaId: string,
  surfaceStatus: WorkbookSurfaceStatus,
): WorkbookSurfaceRegistryEntry {
  const contract = contractsById.get(viewSchemaId);
  if (!contract) {
    throw new Error(`Missing workbook surface contract: ${viewSchemaId}`);
  }
  const surfaceKind = expectedSurfaceKind(surfaceStatus);
  if (contract.surfaceKind !== surfaceKind) {
    throw new Error(
      `Workbook surface ${viewSchemaId} has surface_kind ${contract.surfaceKind}, expected ${surfaceKind}`,
    );
  }
  return Object.freeze({
    contract,
    requiredReferencePackKeys: contract.requiredReferencePackKeys,
    surfaceKind,
    surfaceStatus,
    viewSchemaId,
  });
}

export function buildWorkbookSurfaceRegistry(
  contracts: readonly ViewContract[] = listViewContracts(),
): readonly WorkbookSurfaceRegistryEntry[] {
  const contractsById = contractIndex(contracts);
  const requiredBuiltIns = requiredBuiltInWorkbookSurfaceIds.map((id) =>
    registryEntry(contractsById, id, "required_built_in_sheet"),
  );
  const requiredSystemViews = requiredSystemWorkbookSurfaceIds.map((id) =>
    registryEntry(contractsById, id, "required_system_view"),
  );
  const optionalStandardized = optionalStandardizedWorkbookSurfaceIds
    .filter((id) => contractsById.has(id))
    .map((id) =>
      registryEntry(
        contractsById,
        id,
        "standardized_optional_workbook_surface",
      ),
    );
  return Object.freeze([
    ...requiredBuiltIns,
    ...requiredSystemViews,
    ...optionalStandardized,
  ]);
}

const workbookSurfaceRegistry = buildWorkbookSurfaceRegistry();
const workbookSurfaceRegistryIndex = new Map(
  workbookSurfaceRegistry.map((entry) => [entry.viewSchemaId, entry]),
);

export function listWorkbookSurfaceRegistryEntries(): readonly WorkbookSurfaceRegistryEntry[] {
  return workbookSurfaceRegistry;
}

export function listBuiltInWorkbookSurfaceRegistryEntries(): readonly WorkbookSurfaceRegistryEntry[] {
  return workbookSurfaceRegistry.filter(
    (entry) => entry.surfaceStatus === "required_built_in_sheet",
  );
}

export function listSystemWorkbookSurfaceRegistryEntries(): readonly WorkbookSurfaceRegistryEntry[] {
  return workbookSurfaceRegistry.filter(
    (entry) => entry.surfaceStatus !== "required_built_in_sheet",
  );
}

export function listSystemWorkbookSurfaceGroups(): readonly SystemWorkbookSurfaceGroup[] {
  return systemWorkbookSurfaceGroupDefinitions.flatMap((definition) => {
    const entries = definition.viewSchemaIds.flatMap((viewSchemaId) => {
      const entry = workbookSurfaceRegistryIndex.get(viewSchemaId);
      return entry === undefined ? [] : [entry];
    });
    if (entries.length === 0) {
      return [];
    }
    return [
      Object.freeze({
        entries: Object.freeze(entries),
        label: definition.label,
        token: definition.token,
      }),
    ];
  });
}

export function getWorkbookSurfaceRegistryEntry(
  viewSchemaId: string,
): WorkbookSurfaceRegistryEntry | undefined {
  return workbookSurfaceRegistryIndex.get(viewSchemaId);
}

export function isBuiltInWorkbookSurfaceId(
  viewSchemaId: string,
): viewSchemaId is BuiltInWorkbookSurfaceId {
  return builtInSurfaceIdSet.has(viewSchemaId);
}

export function isStandardizedWorkbookViewSchemaId(
  viewSchemaId: string,
): boolean {
  return workbookSurfaceRegistryIndex.has(viewSchemaId);
}

export function knownWorkbookViewSchemaId(viewSchemaId: string): string {
  return isStandardizedWorkbookViewSchemaId(viewSchemaId)
    ? viewSchemaId
    : timelineViewSchemaId;
}
