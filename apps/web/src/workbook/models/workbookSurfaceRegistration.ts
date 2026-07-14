import {
  listViewContracts,
  type ViewContract,
} from "@cartulary/view-contracts";
import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

export type WorkbookSurfaceOwnerId =
  | "capture_timeline"
  | "entities_observations"
  | "assessments"
  | "evidence"
  | "artifacts"
  | "coordination";

export type WorkbookSurfaceRenderer =
  | "timeline"
  | "entity_hosts"
  | "entity_identities"
  | "assessment"
  | "contract";

export type ReferenceRequirement = {
  readonly requirementId: string;
  readonly resourceId: string;
  readonly viewSchemaId: string;
};

export type WorkbookSurfacePolicy = {
  readonly capabilities: Readonly<{
    decisionSupersede?: true;
    evidenceLifecycle?: true;
    linkedNoteCreate?: true;
    taskLifecycle?: true;
  }>;
  readonly collectionActions: Readonly<
    Record<string, "alias" | "party" | "record" | "risk" | "tag">
  >;
  readonly createDefaults: Readonly<Record<string, string>>;
  readonly currentUserDefaultFields: readonly string[];
  readonly createMinimumFieldSets: readonly (readonly string[])[];
  readonly createMinimumMessage: string;
  readonly referenceRequirements: readonly ReferenceRequirement[];
};

export type WorkbookSurfaceRegistration = {
  readonly contract: ViewContract;
  readonly ownerId: WorkbookSurfaceOwnerId;
  readonly policy: WorkbookSurfacePolicy;
  readonly renderer: WorkbookSurfaceRenderer;
  readonly viewSchemaId: string;
};

const requirement = (viewSchemaId: string): ReferenceRequirement => ({
  requirementId: `workbook-reference:${viewSchemaId}`,
  resourceId: `view:${viewSchemaId}:rows`,
  viewSchemaId,
});

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
].map(requirement);

const policy = (
  overrides: Partial<WorkbookSurfacePolicy> = {},
): WorkbookSurfacePolicy =>
  Object.freeze({
    capabilities: Object.freeze(overrides.capabilities ?? {}),
    collectionActions: Object.freeze(overrides.collectionActions ?? {}),
    createDefaults: Object.freeze(overrides.createDefaults ?? {}),
    currentUserDefaultFields: Object.freeze(
      overrides.currentUserDefaultFields ?? [],
    ),
    createMinimumFieldSets: Object.freeze(
      overrides.createMinimumFieldSets ?? [],
    ),
    createMinimumMessage:
      overrides.createMinimumMessage ?? "At least one value is required.",
    referenceRequirements: Object.freeze(overrides.referenceRequirements ?? []),
  });

type RegistrationDefinition = Omit<
  WorkbookSurfaceRegistration,
  "contract" | "viewSchemaId"
> & {
  readonly viewSchemaId: string;
};

const definitions: readonly RegistrationDefinition[] = [
  {
    viewSchemaId: timelineViewSchemaId,
    ownerId: "capture_timeline",
    renderer: "timeline",
    policy: policy(),
  },
  {
    viewSchemaId: hostsViewSchemaId,
    ownerId: "entities_observations",
    renderer: "entity_hosts",
    policy: policy({
      collectionActions: { "host.aliases": "alias" },
    }),
  },
  {
    viewSchemaId: identitiesViewSchemaId,
    ownerId: "entities_observations",
    renderer: "entity_identities",
    policy: policy({
      collectionActions: { "identity.aliases": "alias" },
    }),
  },
  {
    viewSchemaId: indicatorsViewSchemaId,
    ownerId: "entities_observations",
    renderer: "contract",
    policy: policy(),
  },
  {
    viewSchemaId: assessmentsViewSchemaId,
    ownerId: "assessments",
    renderer: "assessment",
    policy: policy(),
  },
  {
    viewSchemaId: evidenceViewSchemaId,
    ownerId: "evidence",
    renderer: "contract",
    policy: policy({
      capabilities: { evidenceLifecycle: true },
      createMinimumFieldSets: [
        ["evidence.title"],
        ["evidence.storage_ref"],
        ["evidence.collector_party_text"],
        ["evidence.source_party_text"],
      ],
      createMinimumMessage:
        "Evidence needs a title, storage ref, collector, or source.",
      referenceRequirements: [requirement(partiesViewSchemaId)],
    }),
  },
  {
    viewSchemaId: notesViewSchemaId,
    ownerId: "artifacts",
    renderer: "contract",
    policy: policy({
      capabilities: { linkedNoteCreate: true },
      collectionActions: { "note.tags": "tag" },
      createMinimumFieldSets: [["note.title"], ["note.body"]],
      createMinimumMessage: "Title or body is required.",
      referenceRequirements: [
        requirement(timelineViewSchemaId),
        requirement(hostsViewSchemaId),
        requirement(identitiesViewSchemaId),
        requirement(evidenceViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: findingsViewSchemaId,
    ownerId: "artifacts",
    renderer: "contract",
    policy: policy({
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
    policy: policy({
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
    policy: policy({
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
  {
    viewSchemaId: partiesViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: policy({
      createMinimumFieldSets: [["party.display_name", "party.party_kind"]],
      createMinimumMessage: "Display name and kind are required.",
    }),
  },
  {
    viewSchemaId: taskRequestsViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: policy({
      capabilities: { taskLifecycle: true },
      currentUserDefaultFields: ["task.owner_user_id"],
      createMinimumFieldSets: [["task.title", "task.task_kind"]],
      createMinimumMessage: "Title and task kind are required.",
      referenceRequirements: [
        ...allRecordRequirements,
        requirement(partiesViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: decisionsViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: policy({
      capabilities: { decisionSupersede: true },
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
    policy: policy({
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
        requirement(partiesViewSchemaId),
        requirement(decisionsViewSchemaId),
        requirement(taskRequestsViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: handoffViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: policy({
      collectionActions: { "handoff.open_risk_refs": "risk" },
      createMinimumFieldSets: [
        ["handoff.incoming_owner_user_id", "handoff.current_state_summary"],
      ],
      createMinimumMessage:
        "Incoming owner and current state summary are required.",
      referenceRequirements: [
        requirement(decisionsViewSchemaId),
        requirement(taskRequestsViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: statusReviewViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: policy({
      currentUserDefaultFields: ["status_review.review_owner_user_id"],
      createMinimumFieldSets: [["status_review.current_state_summary"]],
      createMinimumMessage: "Current state summary is required.",
      referenceRequirements: [
        requirement(decisionsViewSchemaId),
        requirement(taskRequestsViewSchemaId),
        requirement(evidenceViewSchemaId),
      ],
    }),
  },
  {
    viewSchemaId: lessonViewSchemaId,
    ownerId: "coordination",
    renderer: "contract",
    policy: policy({
      currentUserDefaultFields: ["lesson.owner_user_id"],
      createMinimumFieldSets: [["lesson.summary"]],
      createMinimumMessage: "Summary is required.",
      referenceRequirements: [
        requirement(taskRequestsViewSchemaId),
        requirement(evidenceViewSchemaId),
      ],
    }),
  },
];

export function buildWorkbookSurfaceRegistrations(
  contracts: readonly ViewContract[] = listViewContracts(),
  registrationDefinitions: readonly RegistrationDefinition[] = definitions,
): readonly WorkbookSurfaceRegistration[] {
  const contractsById = new Map(
    contracts.map((contract) => [contract.viewSchemaId, contract]),
  );
  const seen = new Set<string>();
  const registrations = registrationDefinitions.map((definition) => {
    if (seen.has(definition.viewSchemaId)) {
      throw new Error(
        `Duplicate workbook surface registration: ${definition.viewSchemaId}`,
      );
    }
    seen.add(definition.viewSchemaId);
    const contract = contractsById.get(definition.viewSchemaId);
    if (!contract) {
      throw new Error(
        `Workbook surface registration has no contract: ${definition.viewSchemaId}`,
      );
    }
    return Object.freeze({ ...definition, contract });
  });
  const missing = [...contractsById.keys()].filter((id) => !seen.has(id));
  if (missing.length > 0) {
    throw new Error(
      `Missing workbook surface registration: ${missing.sort().join(", ")}`,
    );
  }
  return Object.freeze(registrations);
}

const registrations = buildWorkbookSurfaceRegistrations();
const registrationsById = new Map(
  registrations.map((registration) => [
    registration.viewSchemaId,
    registration,
  ]),
);

export function listWorkbookSurfaceRegistrations() {
  return registrations;
}

export function requireWorkbookSurfaceRegistration(
  viewSchemaId: string,
): WorkbookSurfaceRegistration {
  const registration = registrationsById.get(viewSchemaId);
  if (!registration) {
    throw new Error(`Unknown workbook surface registration: ${viewSchemaId}`);
  }
  return registration;
}
