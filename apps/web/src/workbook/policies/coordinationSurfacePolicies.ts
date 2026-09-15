import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  handoffViewSchemaId,
  lessonViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  defineWorkbookSurfacePolicy,
  type WorkbookSurfacePolicyDefinition,
} from "./workbookSurfacePolicy";

export const coordinationSurfacePolicies = [
  {
    viewSchemaId: partiesViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({}),
  },
  {
    viewSchemaId: taskRequestsViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      ownerBindings: ["task_lifecycle"],
      currentUserDefaultFields: ["task.owner_user_id"],
    }),
  },
  {
    viewSchemaId: decisionsViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      ownerBindings: ["decision_supersede"],
      currentUserDefaultFields: ["decision.owner_user_id"],
    }),
  },
  {
    viewSchemaId: commLogViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      collectionActions: {
        "comm_log.audience_party_ids": "party",
        "comm_log.attendee_party_ids": "party",
      },
    }),
  },
  {
    viewSchemaId: handoffViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      collectionActions: { "handoff.open_risk_refs": "risk" },
    }),
  },
  {
    viewSchemaId: statusReviewViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      currentUserDefaultFields: ["status_review.review_owner_user_id"],
    }),
  },
  {
    viewSchemaId: lessonViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      currentUserDefaultFields: ["lesson.owner_user_id"],
    }),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
