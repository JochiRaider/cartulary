import { evidenceViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  defineWorkbookSurfacePolicy,
  type WorkbookSurfacePolicyDefinition,
} from "./workbookSurfacePolicy";

export const evidenceSurfacePolicies = [
  {
    viewSchemaId: evidenceViewSchemaId,
    ownerId: "evidence",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      ownerBindings: ["evidence_lifecycle"],
    }),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
