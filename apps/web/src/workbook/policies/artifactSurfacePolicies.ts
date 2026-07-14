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
      createMinimumFieldSets: [["note.title"], ["note.body"]],
      createMinimumMessage: "Title or body is required.",
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
      createMinimumFieldSets: [["finding.statement"]],
      createMinimumMessage: "Statement is required.",
      referenceRequirements: allRecordRequirements,
    }),
  },
  {
    viewSchemaId: investigativeQueriesViewSchemaId,
    ownerId: "artifacts",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      createMinimumFieldSets: [
        [
          "investigative_query.platform",
          "investigative_query.purpose",
          "investigative_query.query_text",
        ],
      ],
      createMinimumMessage: "Platform, purpose, and query text are required.",
    }),
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
      createMinimumFieldSets: [
        ["forensic_keyword.pattern", "forensic_keyword.reason"],
      ],
      createMinimumMessage: "Pattern and reason are required.",
    }),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
