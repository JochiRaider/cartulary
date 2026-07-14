import {
  evidenceViewSchemaId,
  partiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  defineWorkbookSurfacePolicy,
  referenceRequirement,
  type WorkbookSurfacePolicyDefinition,
} from "./workbookSurfacePolicy";

export const evidenceSurfacePolicies = [
  {
    viewSchemaId: evidenceViewSchemaId,
    ownerId: "evidence",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      ownerBindings: ["evidence_lifecycle"],
      createMinimumFieldSets: [
        ["evidence.title"],
        ["evidence.storage_ref"],
        ["evidence.collector_party_text"],
        ["evidence.source_party_text"],
      ],
      createMinimumMessage:
        "Evidence needs a title, storage ref, collector, or source.",
      referenceRequirements: [referenceRequirement(partiesViewSchemaId)],
    }),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
