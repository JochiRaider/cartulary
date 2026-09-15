import {
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  investigativeQueriesViewSchemaId,
  notesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  defineWorkbookSurfacePolicy,
  type WorkbookSurfacePolicyDefinition,
} from "./workbookSurfacePolicy";

export const artifactSurfacePolicies = [
  {
    viewSchemaId: notesViewSchemaId,
    ownerId: "artifacts",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      ownerBindings: ["linked_note_create"],
      collectionActions: { "note.tags": "tag" },
    }),
  },
  {
    viewSchemaId: findingsViewSchemaId,
    ownerId: "artifacts",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      createDefaults: { "finding.kind": "finding", "finding.state": "open" },
      currentUserDefaultFields: ["finding.owner_user_id"],
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
