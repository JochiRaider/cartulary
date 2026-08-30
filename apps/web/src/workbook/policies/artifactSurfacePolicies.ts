import {
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  investigativeQueriesViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  defineWorkbookSurfacePolicy,
  referenceRequirement,
  type WorkbookSurfacePolicyDefinition,
} from "./workbookSurfacePolicy";

const allRecordRequirements = [
  timelineViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  evidenceViewSchemaId,
  notesViewSchemaId,
  findingsViewSchemaId,
  investigativeQueriesViewSchemaId,
  forensicKeywordsViewSchemaId,
  taskRequestsViewSchemaId,
  decisionsViewSchemaId,
  partiesViewSchemaId,
].map(referenceRequirement);

export const artifactSurfacePolicies = [
  {
    viewSchemaId: notesViewSchemaId,
    ownerId: "artifacts",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      ownerBindings: ["linked_note_create"],
      collectionActions: { "note.tags": "tag" },
      referenceRequirements: [
        referenceRequirement(timelineViewSchemaId),
        referenceRequirement(hostsViewSchemaId),
        referenceRequirement(identitiesViewSchemaId),
        referenceRequirement(evidenceViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: findingsViewSchemaId,
    ownerId: "artifacts",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      createDefaults: { "finding.kind": "finding", "finding.state": "open" },
      currentUserDefaultFields: ["finding.owner_user_id"],
      referenceRequirements: allRecordRequirements,
    }),
  },
  {
    viewSchemaId: investigativeQueriesViewSchemaId,
    ownerId: "artifacts",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({}),
  },
  {
    viewSchemaId: forensicKeywordsViewSchemaId,
    ownerId: "artifacts",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      createDefaults: {
        "forensic_keyword.match_mode": "literal",
        "forensic_keyword.case_sensitive": "false",
      },
    }),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
