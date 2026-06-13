import {
  fieldCapability,
  requireViewContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";

import {
  assessmentsViewSchemaId,
  buildWorkbookSurfaceRegistry,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  listBuiltInWorkbookSurfaceRegistryEntries,
  listSystemWorkbookSurfaceGroups,
  listSystemWorkbookSurfaceRegistryEntries,
  listWorkbookSurfaceRegistryEntries,
  optionalStandardizedWorkbookSurfaceIds,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  requiredSystemWorkbookSurfaceIds,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "./workbookSurfaceRegistry";

describe("workbook surface registry", () => {
  it("FE-U-P2-02 registers built-in and system workbook surfaces by stable IDs", () => {
    const builtIns = listBuiltInWorkbookSurfaceRegistryEntries();
    expect(builtIns.map((entry) => entry.viewSchemaId)).toEqual([
      ...requiredBuiltInWorkbookSurfaceIds,
    ]);
    expect(builtIns.map((entry) => entry.surfaceKind)).toEqual(
      requiredBuiltInWorkbookSurfaceIds.map(() => "built_in_sheet"),
    );

    const systemEntries = listSystemWorkbookSurfaceRegistryEntries().filter(
      (entry) => entry.surfaceStatus === "required_system_view",
    );
    expect(systemEntries.map((entry) => entry.viewSchemaId)).toEqual([
      ...requiredSystemWorkbookSurfaceIds,
    ]);
    expect(systemEntries.map((entry) => entry.surfaceKind)).toEqual(
      requiredSystemWorkbookSurfaceIds.map(() => "system_view"),
    );

    const allIds = listWorkbookSurfaceRegistryEntries().map(
      (entry) => entry.viewSchemaId,
    );
    expect(new Set(allIds).size).toBe(allIds.length);
    for (const entry of listWorkbookSurfaceRegistryEntries()) {
      expect(entry.contract.viewSchemaId).toBe(entry.viewSchemaId);
      expect(entry.contract.title).not.toBe(entry.viewSchemaId);
    }
  });

  it("FE-U-P2-02 keeps optional standardized surfaces additive after required surfaces", () => {
    const entries = listWorkbookSurfaceRegistryEntries();
    const ids = entries.map((entry) => entry.viewSchemaId);
    const requiredIds = [
      ...requiredBuiltInWorkbookSurfaceIds,
      ...requiredSystemWorkbookSurfaceIds,
    ];
    expect(ids.slice(0, requiredIds.length)).toEqual(requiredIds);
    expect(ids.slice(requiredIds.length)).toEqual([
      ...optionalStandardizedWorkbookSurfaceIds,
    ]);

    const optionalEntries = entries.slice(requiredIds.length);
    expect(optionalEntries.map((entry) => entry.surfaceStatus)).toEqual(
      optionalStandardizedWorkbookSurfaceIds.map(
        () => "standardized_optional_workbook_surface",
      ),
    );
    expect(
      entries
        .filter(
          (entry) =>
            entry.surfaceStatus !== "standardized_optional_workbook_surface",
        )
        .map((entry) => entry.viewSchemaId),
    ).toEqual(requiredIds);

    const shuffled = buildWorkbookSurfaceRegistry(
      [...entries].reverse().map((entry) => entry.contract),
    );
    expect(shuffled.map((entry) => entry.viewSchemaId)).toEqual(ids);
  });

  it("FE-U-P2-02 groups System views by stable group tokens and registry-backed IDs", () => {
    const groups = listSystemWorkbookSurfaceGroups();

    expect(groups.map((group) => group.token)).toEqual([
      "scope-assessment",
      "coordination",
      "review-learning",
      "optional-artifact-surfaces",
    ]);
    expect(
      groups.map((group) => group.entries.map((entry) => entry.viewSchemaId)),
    ).toEqual([
      [indicatorsViewSchemaId, assessmentsViewSchemaId, partiesViewSchemaId],
      [
        taskRequestsViewSchemaId,
        decisionsViewSchemaId,
        commLogViewSchemaId,
        handoffViewSchemaId,
      ],
      [statusReviewViewSchemaId, lessonViewSchemaId],
      [
        findingsViewSchemaId,
        investigativeQueriesViewSchemaId,
        forensicKeywordsViewSchemaId,
      ],
    ]);
    expect(
      groups.flatMap((group) =>
        group.entries.map((entry) => entry.contract.viewSchemaId),
      ),
    ).toEqual([
      indicatorsViewSchemaId,
      assessmentsViewSchemaId,
      partiesViewSchemaId,
      taskRequestsViewSchemaId,
      decisionsViewSchemaId,
      commLogViewSchemaId,
      handoffViewSchemaId,
      statusReviewViewSchemaId,
      lessonViewSchemaId,
      ...optionalStandardizedWorkbookSurfaceIds,
    ]);
  });

  it("FE-U-P2-02 remains keyed by stable IDs when registry labels are relabeled", () => {
    const entries = listWorkbookSurfaceRegistryEntries();
    const relabeledContracts = entries.map((entry) => ({
      ...entry.contract,
      title: `Surface ${entry.viewSchemaId}`,
    }));

    const relabeled = buildWorkbookSurfaceRegistry(relabeledContracts);

    expect(relabeled.map((entry) => entry.viewSchemaId)).toEqual(
      entries.map((entry) => entry.viewSchemaId),
    );
    expect(relabeled.map((entry) => entry.contract.title)).toEqual(
      relabeled.map((entry) => `Surface ${entry.viewSchemaId}`),
    );
    expect(relabeled.map((entry) => entry.surfaceStatus)).toEqual(
      entries.map((entry) => entry.surfaceStatus),
    );
  });

  it("FE-U-P2-02 tolerates absent optional standardized surfaces while requiring required surfaces", () => {
    const entries = listWorkbookSurfaceRegistryEntries();
    const requiredIds = [
      ...requiredBuiltInWorkbookSurfaceIds,
      ...requiredSystemWorkbookSurfaceIds,
    ];
    const requiredIdSet = new Set<string>(requiredIds);
    const optionalIdSet = new Set<string>(
      optionalStandardizedWorkbookSurfaceIds,
    );
    const requiredContracts = entries
      .filter((entry) => requiredIdSet.has(entry.viewSchemaId))
      .map((entry) => entry.contract);

    const requiredOnly = buildWorkbookSurfaceRegistry(requiredContracts);

    expect(requiredOnly.map((entry) => entry.viewSchemaId)).toEqual(
      requiredIds,
    );
    expect(
      requiredOnly.some((entry) => optionalIdSet.has(entry.viewSchemaId)),
    ).toBe(false);
    expect(() =>
      buildWorkbookSurfaceRegistry(
        requiredContracts.filter(
          (contract) =>
            contract.viewSchemaId !== requiredBuiltInWorkbookSurfaceIds[0],
        ),
      ),
    ).toThrow(/Missing workbook surface contract/);
  });

  it("FE-U-P2-02 exposes required reference-pack keys from view contracts", () => {
    const entries = listWorkbookSurfaceRegistryEntries();
    const packBoundContracts = entries.map((entry) =>
      entry.viewSchemaId === findingsViewSchemaId
        ? {
            ...entry.contract,
            requiredReferencePackKeys: ["mitre_attack_enterprise"],
          }
        : entry.contract,
    );

    expect(entries.map((entry) => entry.requiredReferencePackKeys)).toEqual(
      entries.map(() => []),
    );
    expect(
      buildWorkbookSurfaceRegistry(packBoundContracts).find(
        (entry) => entry.viewSchemaId === findingsViewSchemaId,
      )?.requiredReferencePackKeys,
    ).toEqual(["mitre_attack_enterprise"]);
  });

  it("FE-U-P10-01 Verify coordination and review system-view registrations, field mappings, and closed vocabulary options use stable IDs and contract metadata.", () => {
    const expectedSurfaces = [
      {
        viewSchemaId: taskRequestsViewSchemaId,
        title: "Task Requests",
        visibleFields: [
          "task.title",
          "task.status",
          "task.owner_user_id",
          "task.priority",
          "task.task_kind",
          "task.workstream",
          "task.due_at",
          "task.requester_party_text",
          "task.blocked_reason",
          "task.completed_at",
          "task.external_ticket_ref",
          "task.linked_record_count",
          "task.updated_at",
        ],
        enumFields: {
          "task.priority": ["low", "normal", "high", "urgent"],
          "task.status": ["open", "in_progress", "blocked", "done", "canceled"],
          "task.task_kind": [
            "question",
            "request",
            "collection",
            "containment",
            "follow_up",
          ],
        },
        writableFieldMetadata: {
          "task.decision_record_id": {
            directReferenceContractId: "same_incident_decision_ref_v1",
          },
          "task.due_at": {
            directScalarContractId: "timestamp_instant_v1",
          },
          "task.linked_record_ids": {
            writeAction: "task linked record collection actions",
          },
          "task.requester_party_id": {
            directReferenceContractId: "same_incident_party_ref_v1",
          },
          "task.title": { stringContractId: "single_line_title_v1" },
        },
      },
      {
        viewSchemaId: decisionsViewSchemaId,
        title: "Decisions",
        visibleFields: [
          "decision.summary",
          "decision.status",
          "decision.owner_user_id",
          "decision.decision_type",
          "decision.decided_at",
          "decision.rationale",
          "decision.support_refs",
          "decision.affected_record_count",
          "decision.supersedes_record_id",
          "decision.updated_at",
        ],
        enumFields: {
          "decision.decision_type": [
            "scope",
            "containment",
            "communication",
            "evidence",
            "reporting",
          ],
          "decision.status": [
            "proposed",
            "approved",
            "rejected",
            "superseded",
            "executed",
          ],
        },
        writableFieldMetadata: {
          "decision.affected_record_ids": {
            writeAction: "decision affected record collection actions",
          },
          "decision.decided_at": {
            directScalarContractId: "timestamp_instant_v1",
          },
          "decision.rationale": { stringContractId: "multiline_body_v1" },
          "decision.summary": { stringContractId: "single_line_title_v1" },
          "decision.support_refs": {
            writeAction: "decision support reference collection actions",
          },
        },
      },
      {
        viewSchemaId: partiesViewSchemaId,
        title: "Parties",
        visibleFields: [
          "party.display_name",
          "party.party_kind",
          "party.organization_name",
          "party.role_title",
          "party.primary_email",
          "party.timezone_name",
          "party.external_ref",
          "party.updated_at",
        ],
        enumFields: {
          "party.party_kind": [
            "person",
            "team",
            "organization",
            "distribution_list",
            "other",
          ],
        },
        writableFieldMetadata: {
          "party.display_name": { stringContractId: "display_name_line_v1" },
          "party.external_ref": { stringContractId: "locator_text_v1" },
          "party.notes": { stringContractId: "multiline_body_v1" },
          "party.primary_email": { stringContractId: "email_address_v1" },
          "party.timezone_name": { stringContractId: "timezone_name_v1" },
        },
      },
      {
        viewSchemaId: commLogViewSchemaId,
        title: "Communications Log",
        visibleFields: [
          "comm_log.timestamp_utc",
          "comm_log.comm_type",
          "comm_log.audience",
          "comm_log.channel_or_meeting",
          "comm_log.summary",
          "comm_log.next_report_at",
          "comm_log.updated_at",
        ],
        enumFields: {
          "comm_log.comm_type": [
            "meeting",
            "notification",
            "approval",
            "briefing",
            "handoff",
          ],
        },
        writableFieldMetadata: {
          "comm_log.action_task_ids": {
            writeAction: "comm_log task reference collection actions",
          },
          "comm_log.attendee_party_ids": {
            writeAction: "comm_log attendee party collection actions",
          },
          "comm_log.audience": { stringContractId: "party_text_v1" },
          "comm_log.audience_party_ids": {
            writeAction: "comm_log audience party collection actions",
          },
          "comm_log.decision_ids": {
            writeAction: "comm_log decision reference collection actions",
          },
          "comm_log.next_report_at": {
            directScalarContractId: "timestamp_instant_v1",
          },
          "comm_log.timestamp_utc": {
            directScalarContractId: "timestamp_instant_v1",
          },
        },
      },
      {
        viewSchemaId: handoffViewSchemaId,
        title: "Handoff",
        visibleFields: [
          "handoff.timestamp_utc",
          "handoff.outgoing_owner_user_id",
          "handoff.incoming_owner_user_id",
          "handoff.current_state_summary",
          "handoff.next_checks",
          "handoff.acknowledged_at",
          "handoff.updated_at",
        ],
        enumFields: {
          "handoff.ack_state": ["pending", "acknowledged"],
        },
        writableFieldMetadata: {
          "handoff.acknowledged_at": {
            directScalarContractId: "timestamp_instant_v1",
          },
          "handoff.current_state_summary": {
            stringContractId: "multiline_body_v1",
          },
          "handoff.open_decision_ids": {
            writeAction: "handoff decision reference collection actions",
          },
          "handoff.open_risk_refs": {
            writeAction: "handoff risk reference collection actions",
          },
          "handoff.open_task_ids": {
            writeAction: "handoff task reference collection actions",
          },
          "handoff.timestamp_utc": {
            directScalarContractId: "timestamp_instant_v1",
          },
        },
      },
      {
        viewSchemaId: statusReviewViewSchemaId,
        title: "Status Review",
        visibleFields: [
          "status_review.timestamp_utc",
          "status_review.review_owner_user_id",
          "status_review.current_state_summary",
          "status_review.active_risks_summary",
          "status_review.next_report_at",
          "status_review.updated_at",
        ],
        enumFields: {},
        writableFieldMetadata: {
          "status_review.active_risks_summary": {
            stringContractId: "multiline_body_v1",
          },
          "status_review.blocked_task_ids": {
            writeAction:
              "status_review blocked task reference collection actions",
          },
          "status_review.current_state_summary": {
            stringContractId: "multiline_body_v1",
          },
          "status_review.next_report_at": {
            directScalarContractId: "timestamp_instant_v1",
          },
          "status_review.open_decision_ids": {
            writeAction: "status_review decision reference collection actions",
          },
          "status_review.pending_evidence_ids": {
            writeAction:
              "status_review pending evidence reference collection actions",
          },
          "status_review.timestamp_utc": {
            directScalarContractId: "timestamp_instant_v1",
          },
        },
      },
      {
        viewSchemaId: lessonViewSchemaId,
        title: "Lesson",
        visibleFields: [
          "lesson.timestamp_utc",
          "lesson.summary",
          "lesson.owner_user_id",
          "lesson.closure_state",
          "lesson.follow_up_task_ids",
          "lesson.evidence_refs",
          "lesson.updated_at",
        ],
        enumFields: {
          "lesson.closure_state": ["open", "closed"],
        },
        writableFieldMetadata: {
          "lesson.evidence_refs": {
            writeAction: "lesson evidence reference collection actions",
          },
          "lesson.follow_up_task_ids": {
            writeAction: "lesson follow-up task reference collection actions",
          },
          "lesson.summary": { stringContractId: "single_line_title_v1" },
          "lesson.timestamp_utc": {
            directScalarContractId: "timestamp_instant_v1",
          },
        },
      },
    ] as const;

    const entriesById = new Map(
      listSystemWorkbookSurfaceRegistryEntries().map((entry) => [
        entry.viewSchemaId,
        entry,
      ]),
    );

    for (const expected of expectedSurfaces) {
      const entry = entriesById.get(expected.viewSchemaId);
      expect(entry, expected.viewSchemaId).toBeDefined();
      expect(entry?.surfaceKind).toBe("system_view");
      expect(entry?.surfaceStatus).toBe("required_system_view");
      expect(entry?.contract.title).toBe(expected.title);
      expect(entry?.contract.viewSchemaId).toBe(expected.viewSchemaId);
      expect(entry?.requiredReferencePackKeys).toEqual([]);

      const contract = requireViewContract(expected.viewSchemaId);
      expect(contract).toBe(entry?.contract);
      expect(contract.defaultVisibleFields).toEqual([
        ...expected.visibleFields,
      ]);

      const fieldKeys = contract.fields.map((field) => field.fieldKey);
      expect(new Set(fieldKeys).size).toBe(fieldKeys.length);
      for (const fieldKey of expected.visibleFields) {
        expect(contract.fieldMap[fieldKey]?.fieldKey).toBe(fieldKey);
      }
      for (const fieldKey of [
        ...contract.defaultVisibleFields,
        ...contract.filterFields,
        ...contract.groupingFields,
        ...contract.sortFields,
      ]) {
        expect(contract.fieldMap[fieldKey]?.fieldKey).toBe(fieldKey);
      }
      for (const [fieldKey, values] of Object.entries(expected.enumFields)) {
        expect(contract.fieldMap[fieldKey]?.readKind).toBe("enum");
        expect(contract.fieldMap[fieldKey]?.enumValues).toEqual(values);
        expect(fieldCapability(contract, fieldKey).editable).toBe(
          contract.fieldMap[fieldKey]?.writeKind !== "read_only",
        );
      }
      for (const [fieldKey, metadata] of Object.entries(
        expected.writableFieldMetadata,
      )) {
        expect(contract.fieldMap[fieldKey]).toMatchObject({
          fieldKey,
          ...metadata,
        });
        expect(contract.fieldMap[fieldKey]?.writeKind).not.toBe("read_only");
      }
    }
  });
});
