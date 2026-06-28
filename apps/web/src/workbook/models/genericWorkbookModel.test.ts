import {
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  buildGenericCreatePayload,
  buildGenericPatchChange,
  collectionItemLabels,
  extractEmailFromPartyText,
  genericCellLabel,
  genericCellLabelForField,
  genericCollectionItems,
  genericContractColumnWidth,
  genericCreateMinimumMessage,
  genericReferenceOptionsFromRows,
  genericRowLabel,
  initialGenericCreateDraft,
  isMultilineGenericField,
  parseMutationError,
  partyLinkPairsForContract,
  workbookCreateMinimumSatisfied,
} from "./genericWorkbookModel";
import { workbookGridRows } from "./workbookContractRows";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
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
    const rows = [
      { record_id: "party-1", cells: {} },
      { record_id: "party-2", cells: {} },
    ];

    expect(
      workbookGridRows({
        getRecordId: (row) => row.record_id,
        rows,
        surface: partiesViewSchemaId,
      }),
    ).toEqual([
      {
        key: "party-1",
        recordId: "party-1",
        data: rows[0],
        testId: "grid-row-cartulary.view.parties.v1-party-1",
      },
      {
        key: "party-2",
        recordId: "party-2",
        data: rows[1],
        testId: "grid-row-cartulary.view.parties.v1-party-2",
      },
    ]);

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

  it("uses contract create minima and canonical alias actions for entity sheets", () => {
    const hosts = requireViewContract(hostsViewSchemaId);
    const identities = requireViewContract(identitiesViewSchemaId);

    expect(
      workbookCreateMinimumSatisfied(hosts, {
        "host.aliases": "VPN Gateway",
      }),
    ).toBe(false);
    expect(
      buildGenericCreatePayload(
        hosts,
        {
          "host.aliases": "VPN Gateway",
        },
        "txn-host-alias-only",
      ),
    ).toBeNull();
    expect(
      buildGenericCreatePayload(
        hosts,
        {
          "host.location": "Datacenter A",
        },
        "txn-host-operational-only",
      ),
    ).toBeNull();
    expect(
      buildGenericCreatePayload(
        hosts,
        {
          "host.hostname": " GATEWAY-01 ",
          "host.aliases": " VPN Gateway ",
        },
        "txn-host-create",
      ),
    ).toMatchObject({
      client_txn_id: "txn-host-create",
      "host.hostname": "GATEWAY-01",
      "host.aliases": {
        kind: "collection_actions_v1",
        actions: [{ op: "add_alias", alias_text: "VPN Gateway" }],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(hosts, "host.aliases"),
        "entity_alias:host-1",
        "remove",
      ),
    ).toEqual({
      field_key: "host.aliases",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "remove_alias", item_ref: "entity_alias:host-1" }],
      },
    });

    expect(
      buildGenericCreatePayload(
        identities,
        {
          "identity.mfa_state": "enabled",
        },
        "txn-identity-operational-only",
      ),
    ).toBeNull();
    expect(
      buildGenericCreatePayload(
        identities,
        {
          "identity.email": " alex.analyst@example.test ",
          "identity.aliases": " Analyst Alex ",
        },
        "txn-identity-create",
      ),
    ).toMatchObject({
      client_txn_id: "txn-identity-create",
      "identity.email": "alex.analyst@example.test",
      "identity.aliases": {
        kind: "collection_actions_v1",
        actions: [{ op: "add_alias", alias_text: "Analyst Alex" }],
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

  it("keeps display labels, widths, and minimum messages contract-driven", () => {
    const evidence = requireViewContract(evidenceViewSchemaId);
    const notes = requireViewContract(notesViewSchemaId);

    expect(genericCellLabel(null)).toBe("None");
    expect(genericCellLabel(true)).toBe("Yes");
    expect(
      genericCellLabel({
        items: [
          { display_text: "Display" },
          { alias_text: "Alias" },
          { tag_name: "Tag" },
        ],
      }),
    ).toBe("Display, Alias, Tag");
    expect(
      genericCellLabelForField(
        evidenceViewSchemaId,
        "evidence.storage_ref",
        "object://12345678-1234-4234-8234-123456789abc",
      ),
    ).toBe("Managed object");
    expect(
      collectionItemLabels([
        { raw_text: "raw" },
        { linked_record_id: "record-1" },
        { item_ref: "item-1" },
      ]),
    ).toEqual(["raw", "record-1", "item-1"]);
    expect(genericCreateMinimumMessage(evidenceViewSchemaId)).toBe(
      "Evidence needs a title, storage ref, collector, or source.",
    );
    expect(genericCreateMinimumMessage("cartulary.view.unknown.v1")).toBe(
      "At least one value is required.",
    );
    expect(genericContractColumnWidth(requireField(notes, "note.body"))).toBe(
      320,
    );
    expect(
      genericContractColumnWidth(requireField(evidence, "evidence.edited_at")),
    ).toBe(180);
  });

  it("builds reference labels, party-link pairs, and collection item view models", () => {
    const evidence = requireViewContract(evidenceViewSchemaId);
    const commLog = requireViewContract(commLogViewSchemaId);
    const row = {
      record_id: "evidence-1",
      row_version: 2,
      cells: {
        "evidence.title": { value: "Endpoint package" },
        "evidence.related_records": {
          value: {
            items: [
              { item_ref: "record-ref-1", display_text: "Record one" },
              { item_ref: "record-ref-2", display_text: "" },
              { display_text: "missing ref" },
            ],
          },
        },
      },
    };

    expect(genericRowLabel(evidence, row)).toBe(
      "Endpoint package (evidence-1)",
    );
    expect(
      genericReferenceOptionsFromRows(evidenceViewSchemaId, [row]),
    ).toEqual([
      {
        recordId: "evidence-1",
        label: "Endpoint package (evidence-1)",
        viewSchemaId: evidenceViewSchemaId,
      },
    ]);
    expect(
      partyLinkPairsForContract(evidence).map((pair) => pair.label),
    ).toEqual(["Collector", "Source"]);
    expect(
      partyLinkPairsForContract(commLog).some(
        (pair) => pair.label === "Requester",
      ),
    ).toBe(false);
    expect(genericCollectionItems(row, "evidence.related_records")).toEqual([
      { itemRef: "record-ref-1", displayText: "Record one" },
      { itemRef: "record-ref-2", displayText: "record-ref-2" },
    ]);
  });

  it("keeps multiline detection, party email extraction, and mutation error parsing stable", () => {
    const notes = requireViewContract(notesViewSchemaId);
    expect(isMultilineGenericField(requireField(notes, "note.body"))).toBe(
      true,
    );
    expect(extractEmailFromPartyText("Alice <alice@example.test>")).toBe(
      "alice@example.test",
    );
    expect(extractEmailFromPartyText("No address")).toBeNull();
    expect(
      parseMutationError({
        error: {
          code: "row_version_conflict",
          conflict: { field_key: "note.body" },
        },
      }),
    ).toBe("row_version_conflict: note.body");
    expect(
      parseMutationError({
        error: {
          code: "validation_failed",
          details: { reason_code: "missing_required_field" },
        },
      }),
    ).toBe("validation_failed: missing_required_field");
  });
});
