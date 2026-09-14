import { listWorkbookSurfaceContracts } from "@cartulary/view-contracts";
import type { OrdinaryCreateContribution } from "../ordinary/ordinaryCreateContract";
import { prepareOrdinaryCreateFields } from "../ordinary/prepareOrdinaryCreateFields";
import { coordinationReferenceView } from "./coordinationCreateModel";
import { initialCoordinationCreateErrors } from "./initialCoordinationCreateRules";

export const coordinationOrdinaryCreate: OrdinaryCreateContribution = {
  views: [
    "task_requests",
    "decisions",
    "comm_log",
    "handoff",
    "status_review",
    "lesson",
  ].map((name) => `cartulary.view.${name}.v1`),
  prepare: (contract, values, id) =>
    prepareOrdinaryCreateFields(contract, values, id, {
      validate: (request, errors) => {
        Object.assign(
          errors,
          initialCoordinationCreateErrors(contract.viewSchemaId, request),
        );
      },
    }),
  defaults: (contract, actor) => ({
    ...(contract.viewSchemaId === "cartulary.view.task_requests.v1"
      ? { "task.status": "open", "task.priority": "normal" }
      : {}),
    ...(contract.viewSchemaId === "cartulary.view.decisions.v1"
      ? { "decision.status": "proposed" }
      : {}),
    ...(contract.viewSchemaId === "cartulary.view.handoff.v1"
      ? { "handoff.outgoing_owner_user_id": actor }
      : {}),
    ...(contract.viewSchemaId === "cartulary.view.lesson.v1"
      ? { "lesson.closure_state": "open" }
      : {}),
  }),
  referenceViews: (field) => {
    const coordination = coordinationReferenceView(field);
    if (coordination) return [coordination];
    if (field.directReferenceContractId === "same_incident_party_ref_v1")
      return ["cartulary.view.parties.v1"];
    if (field.directReferenceContractId === "same_incident_decision_ref_v1")
      return ["cartulary.view.decisions.v1"];
    if (
      [
        "task.linked_record_ids",
        "decision.support_refs",
        "decision.affected_record_ids",
      ].includes(field.fieldKey)
    )
      return listWorkbookSurfaceContracts().map((view) => view.viewSchemaId);
    return [];
  },
};
