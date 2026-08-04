import type { SystemViewSwitcherGroupToken } from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  listWorkbookSurfaceContracts,
  notesViewSchemaId,
  optionalStandardizedWorkbookSurfaceIds,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  requiredSystemWorkbookSurfaceIds,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
  type WorkbookSurfaceContract,
} from "@cartulary/view-contracts";

export {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  optionalStandardizedWorkbookSurfaceIds,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  requiredSystemWorkbookSurfaceIds,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
};

export type WorkbookSurfaceRegistryEntry = WorkbookSurfaceContract;

export type SystemWorkbookSurfaceGroup = {
  readonly entries: readonly WorkbookSurfaceRegistryEntry[];
  readonly label: string;
  readonly token: SystemViewSwitcherGroupToken;
};

const systemWorkbookSurfaceGroupDefinitions = [
  {
    label: "Scope and indicators",
    token: "scope-indicators",
    viewSchemaIds: [indicatorsViewSchemaId, assessmentsViewSchemaId],
  },
  {
    label: "Coordination",
    token: "coordination",
    viewSchemaIds: [
      taskRequestsViewSchemaId,
      decisionsViewSchemaId,
      partiesViewSchemaId,
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

const workbookSurfaceRegistry = listWorkbookSurfaceContracts();
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
