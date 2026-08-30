import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
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
      referenceRequirements: [
        ...allRecordRequirements,
        referenceRequirement(partiesViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: decisionsViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      ownerBindings: ["decision_supersede"],
      currentUserDefaultFields: ["decision.owner_user_id"],
      referenceRequirements: allRecordRequirements,
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
      referenceRequirements: [
        referenceRequirement(partiesViewSchemaId),
        referenceRequirement(decisionsViewSchemaId),
        referenceRequirement(taskRequestsViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: handoffViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      collectionActions: { "handoff.open_risk_refs": "risk" },
      referenceRequirements: [
        referenceRequirement(decisionsViewSchemaId),
        referenceRequirement(taskRequestsViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: statusReviewViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      currentUserDefaultFields: ["status_review.review_owner_user_id"],
      referenceRequirements: [
        referenceRequirement(decisionsViewSchemaId),
        referenceRequirement(taskRequestsViewSchemaId),
        referenceRequirement(evidenceViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: lessonViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      currentUserDefaultFields: ["lesson.owner_user_id"],
      referenceRequirements: [
        referenceRequirement(taskRequestsViewSchemaId),
        referenceRequirement(evidenceViewSchemaId),
      ],
    }),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
