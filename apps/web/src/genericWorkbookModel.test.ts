import {
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  buildGenericCreatePayload,
  buildGenericPatchChange,
  initialGenericCreateDraft,
  workbookCreateMinimumSatisfied,
} from "./genericWorkbookModel";
import {
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  notesViewSchemaId,
} from "./workbookSurfaceRegistry";

function requireField(
  contract: ViewContract,
  fieldKey: string,
): ViewFieldContract {
  const field = contract.fieldMap[fieldKey];
  if (!field) {
    throw new Error(`Missing field ${fieldKey} on ${contract.viewSchemaId}`);
  }
  return field;
}

describe("genericWorkbookModel", () => {
  it("builds generic creates with omitted fields, trims, explicit clears, and minimum checks", () => {
    const evidence = requireViewContract(evidenceViewSchemaId);
    expect(
      buildGenericCreatePayload(evidence, {}, "txn-evidence-missing"),
    ).toBeNull();
    expect(
      buildGenericCreatePayload(
        evidence,
        {
          "evidence.title": " Endpoint package ",
          "evidence.requested_at": "2026-04-24T12:00:00Z",
          "evidence.collector_party_id": "",
        },
        "txn-evidence-create",
      ),
    ).toMatchObject({
      client_txn_id: "txn-evidence-create",
      "evidence.title": "Endpoint package",
      "evidence.requested_at": "2026-04-24T12:00:00Z",
      "evidence.collector_party_id": null,
    });
  });

  it("builds typed direct and collection patch changes", () => {
    const evidence = requireViewContract(evidenceViewSchemaId);
    const notes = requireViewContract(notesViewSchemaId);
    const decisions = requireViewContract(decisionsViewSchemaId);

    expect(
      buildGenericPatchChange(
        requireField(evidence, "evidence.source_party_id"),
        "",
      ),
    ).toEqual({ field_key: "evidence.source_party_id", value: null });
    expect(
      buildGenericPatchChange(requireField(notes, "note.tags"), " urgent "),
    ).toEqual({
      field_key: "note.tags",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "add_tag", tag_name: "urgent" }],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(decisions, "decision.support_refs"),
        "record-1",
      ),
    ).toEqual({
      field_key: "decision.support_refs",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "add_record_ref", linked_record_id: "record-1" }],
      },
    });
  });

  it("rejects invalid number and boolean direct payload values", () => {
    const forensicKeywords = requireViewContract(forensicKeywordsViewSchemaId);
    const findings = requireViewContract(findingsViewSchemaId);
    const booleanField = requireField(
      forensicKeywords,
      "forensic_keyword.case_sensitive",
    );
    const confidenceScoreField = requireField(
      findings,
      "finding.confidence_score",
    );

    expect(buildGenericPatchChange(booleanField, "false")).toEqual({
      field_key: "forensic_keyword.case_sensitive",
      value: false,
    });
    expect(buildGenericPatchChange(booleanField, "yes")).toBeNull();
    expect(buildGenericPatchChange(confidenceScoreField, "10")).toEqual({
      field_key: "finding.confidence_score",
      value: 10,
    });
    expect(buildGenericPatchChange(confidenceScoreField, "10.5")).toBeNull();
    expect(
      buildGenericPatchChange(
        confidenceScoreField,
        String(Number.MAX_SAFE_INTEGER + 1),
      ),
    ).toBeNull();
  });

  it("seeds owner defaults and known generic create minimums", () => {
    const findings = requireViewContract(findingsViewSchemaId);
    const forensicKeywords = requireViewContract(forensicKeywordsViewSchemaId);

    expect(initialGenericCreateDraft(findings, "user-1")).toMatchObject({
      "finding.kind": "finding",
      "finding.owner_user_id": "user-1",
      "finding.state": "open",
    });
    expect(initialGenericCreateDraft(forensicKeywords, null)).toMatchObject({
      "forensic_keyword.case_sensitive": "false",
      "forensic_keyword.match_mode": "literal",
    });
    expect(
      workbookCreateMinimumSatisfied(notesViewSchemaId, {
        "note.body": "body",
      }),
    ).toBe(true);
    expect(
      workbookCreateMinimumSatisfied(notesViewSchemaId, {
        "note.body": "",
      }),
    ).toBe(false);
    expect(
      workbookCreateMinimumSatisfied("cartulary.view.unknown.v1", {
        "unknown.field": "value",
      }),
    ).toBe(true);
  });
});
