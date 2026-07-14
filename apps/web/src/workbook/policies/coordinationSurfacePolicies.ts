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
    policy: defineWorkbookSurfacePolicy({
      createMinimumFieldSets: [["party.display_name", "party.party_kind"]],
      createMinimumMessage: "Display name and kind are required.",
    }),
  },
  {
    viewSchemaId: taskRequestsViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy({
      ownerBindings: ["task_lifecycle"],
      currentUserDefaultFields: ["task.owner_user_id"],
      createMinimumFieldSets: [["task.title", "task.task_kind"]],
      createMinimumMessage: "Title and task kind are required.",
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
      createMinimumFieldSets: [
        ["decision.summary", "decision.decision_type", "decision.rationale"],
      ],
      createMinimumMessage:
        "Summary, decision type, and rationale are required.",
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
      createMinimumFieldSets: [
        [
          "comm_log.comm_type",
          "comm_log.audience",
          "comm_log.channel_or_meeting",
          "comm_log.summary",
        ],
      ],
      createMinimumMessage:
        "Type, audience, channel or meeting, and summary are required.",
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
      createMinimumFieldSets: [
        ["handoff.incoming_owner_user_id", "handoff.current_state_summary"],
      ],
      createMinimumMessage:
        "Incoming owner and current state summary are required.",
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
      createMinimumFieldSets: [["status_review.current_state_summary"]],
      createMinimumMessage: "Current state summary is required.",
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
      createMinimumFieldSets: [["lesson.summary"]],
      createMinimumMessage: "Summary is required.",
      referenceRequirements: [
        referenceRequirement(taskRequestsViewSchemaId),
        referenceRequirement(evidenceViewSchemaId),
      ],
    }),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
