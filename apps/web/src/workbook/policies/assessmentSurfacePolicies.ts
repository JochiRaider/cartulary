import { assessmentsViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  defineWorkbookSurfacePolicy,
  type WorkbookSurfacePolicyDefinition,
} from "./workbookSurfacePolicy";

export const assessmentSurfacePolicies = [
  {
    viewSchemaId: assessmentsViewSchemaId,
    ownerId: "assessments",
    renderer: "assessment",
    policy: defineWorkbookSurfacePolicy(),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
