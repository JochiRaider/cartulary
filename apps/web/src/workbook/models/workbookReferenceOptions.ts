import type { ViewFieldContract } from "@cartulary/view-contracts";
import { isPartyRefCollection } from "./genericWorkbookModel";

export type GenericReferenceOption = {
  recordId: string;
  label: string;
  viewSchemaId: string;
};

export type GenericReferenceOptions = {
  incidentMembers: GenericReferenceOption[];
  parties: GenericReferenceOption[];
  taskRequests: GenericReferenceOption[];
  decisions: GenericReferenceOption[];
  evidence: GenericReferenceOption[];
  hosts: GenericReferenceOption[];
  identities: GenericReferenceOption[];
  notes: GenericReferenceOption[];
  timeline: GenericReferenceOption[];
  noteSourceRecords: GenericReferenceOption[];
  allRecords: GenericReferenceOption[];
};

export function emptyGenericReferenceOptions(): GenericReferenceOptions {
  return {
    incidentMembers: [],
    parties: [],
    taskRequests: [],
    decisions: [],
    evidence: [],
    hosts: [],
    identities: [],
    notes: [],
    timeline: [],
    noteSourceRecords: [],
    allRecords: [],
  };
}

export function referenceOptionsForField(
  field: ViewFieldContract,
  options: GenericReferenceOptions,
): GenericReferenceOption[] {
  if (field.directReferenceContractId === "same_incident_party_ref_v1") {
    return options.parties;
  }
  if (field.directReferenceContractId === "same_incident_decision_ref_v1") {
    return options.decisions;
  }
  if (field.directReferenceContractId === "incident_member_user_ref_v1") {
    return options.incidentMembers;
  }
  if (isPartyRefCollection(field.fieldKey)) {
    return options.parties;
  }
  switch (field.fieldKey) {
    case "comm_log.decision_ids":
    case "handoff.open_decision_ids":
    case "status_review.open_decision_ids":
      return options.decisions;
    case "comm_log.action_task_ids":
    case "handoff.open_task_ids":
    case "status_review.blocked_task_ids":
    case "lesson.follow_up_task_ids":
      return options.taskRequests;
    case "status_review.pending_evidence_ids":
    case "lesson.evidence_refs":
      return options.evidence;
    case "task.linked_record_ids":
    case "decision.support_refs":
    case "decision.affected_record_ids":
    case "finding.supporting_refs":
    case "finding.contradictory_refs":
      return options.allRecords;
    default:
      return [];
  }
}

export function genericFieldUsesReferenceOptions(
  field: ViewFieldContract,
): boolean {
  return (
    referenceOptionsForField(field, {
      ...emptyGenericReferenceOptions(),
      parties: [{ recordId: "party", label: "party", viewSchemaId: "party" }],
      decisions: [
        { recordId: "decision", label: "decision", viewSchemaId: "decision" },
      ],
      taskRequests: [{ recordId: "task", label: "task", viewSchemaId: "task" }],
      evidence: [
        { recordId: "evidence", label: "evidence", viewSchemaId: "evidence" },
      ],
      allRecords: [
        { recordId: "record", label: "record", viewSchemaId: "record" },
      ],
      incidentMembers: [
        { recordId: "user", label: "user", viewSchemaId: "incident_member" },
      ],
    }).length > 0
  );
}
