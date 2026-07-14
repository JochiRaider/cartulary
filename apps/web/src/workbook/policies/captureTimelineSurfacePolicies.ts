import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  defineWorkbookSurfacePolicy,
  type WorkbookSurfacePolicyDefinition,
} from "./workbookSurfacePolicy";

export const captureTimelineSurfacePolicies = [
  {
    viewSchemaId: timelineViewSchemaId,
    ownerId: "capture_timeline",
    renderer: "timeline",
    policy: defineWorkbookSurfacePolicy(),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
