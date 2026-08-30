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
      referenceRequirements: [referenceRequirement(partiesViewSchemaId)],
    }),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
