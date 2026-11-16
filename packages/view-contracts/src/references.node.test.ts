import { describe, expect, it } from "vitest";
import {
  listViewContracts,
  listWorkbookSurfaceContracts,
  requireViewContract,
} from "./contracts.js";
import {
  getReferenceFieldContract,
  listReferenceFieldContracts,
} from "./references.js";

describe("ordinary reference field contracts", () => {
  it("accounts for every ordinary direct reference and declared reference action without widening specialized fields", () => {
    const fields = listReferenceFieldContracts();
    expect(fields.length).toBeGreaterThan(0);
    expect(
      new Set(fields.map((f) => `${f.viewSchemaId}:${f.fieldKey}`)).size,
    ).toBe(fields.length);
    for (const contract of listViewContracts()) {
      for (const field of contract.fields.filter(
        (f) => f.patchWritable && f.directReferenceContractId,
      ))
        expect(
          getReferenceFieldContract(contract.viewSchemaId, field.fieldKey)
            ?.kind,
        ).toBe("direct");
    }
    for (const field of fields) {
      const schema = requireViewContract(field.viewSchemaId).fieldMap[
        field.fieldKey
      ];
      expect(schema?.patchWritable).toBe(true);
      expect(schema?.writeKind).toBe(
        field.kind === "direct" ? "direct_value" : "action_payload",
      );
      for (const id of field.targetViewSchemaIds) {
        const target = listWorkbookSurfaceContracts().find(
          (surface) => surface.viewSchemaId === id,
        );
        expect(
          target?.sourceRecordTypes.every((kind) =>
            field.targetRecordTypes.includes(kind),
          ),
        ).toBe(true);
      }
    }
    // Every action-payload field is an ordinary collection reference or one
    // of these specialized fields, so a new field must be classified.
    const specializedActionFields = new Set([
      "cartulary.view.assessments.v1:assessment.support_refs",
      "cartulary.view.handoff.v1:handoff.open_risk_refs",
      "cartulary.view.hosts.v1:host.aliases",
      "cartulary.view.identities.v1:identity.aliases",
      "cartulary.view.notes.v1:note.tags",
      "cartulary.view.timeline.v2:timeline.attached_evidence_ids",
      "cartulary.view.timeline.v2:timeline.host_refs",
      "cartulary.view.timeline.v2:timeline.identity_refs",
      "cartulary.view.timeline.v2:timeline.tags",
    ]);
    const unclassified: string[] = [];
    for (const contract of listViewContracts())
      for (const field of contract.fields.filter(
        (f) => f.writeKind === "action_payload",
      )) {
        const key = `${contract.viewSchemaId}:${field.fieldKey}`;
        const reference = getReferenceFieldContract(
          contract.viewSchemaId,
          field.fieldKey,
        );
        if (specializedActionFields.has(key))
          expect(reference, key).toBeUndefined();
        else if (reference?.kind !== "collection") unclassified.push(key);
      }
    expect(unclassified).toEqual([]);
    expect(
      getReferenceFieldContract(
        "cartulary.view.assessments.v1",
        "assessment.support_refs",
      ),
    ).toBeUndefined();
    expect(
      getReferenceFieldContract(
        "cartulary.view.handoff.v1",
        "handoff.open_risk_refs",
      ),
    ).toBeUndefined();
    expect(
      getReferenceFieldContract(
        "cartulary.view.timeline.v2",
        "timeline.host_refs",
      ),
    ).toBeUndefined();
  });

  it("keeps member identity separate and includes all eligible first-class record surfaces", () => {
    const member = getReferenceFieldContract(
      "cartulary.view.task_requests.v1",
      "task.owner_user_id",
    );
    expect(member?.identityKind).toBe("incident_member");
    expect(member?.targetViewSchemaIds).toEqual([]);
    expect(
      getReferenceFieldContract(
        "cartulary.view.evidence.v1",
        "evidence.owner_user_id",
      ),
    ).toEqual({
      viewSchemaId: "cartulary.view.evidence.v1",
      fieldKey: "evidence.owner_user_id",
      kind: "direct",
      identityKind: "incident_member",
      targetRecordTypes: [],
      targetViewSchemaIds: [],
      excludeSource: false,
      addOperation: null,
      removeOperation: null,
    });
    expect(
      requireViewContract("cartulary.view.evidence.v1").fieldMap[
        "evidence.owner_user_id"
      ],
    ).toMatchObject({
      clearable: true,
      createWritable: true,
      patchWritable: true,
      defaultHidden: true,
      directReferenceContractId: "incident_member_user_ref_v1",
      conflictResolutionClass: "atomic_replace",
    });
    const records = getReferenceFieldContract(
      "cartulary.view.decisions.v1",
      "decision.support_refs",
    );
    expect(records?.targetViewSchemaIds).toHaveLength(17);
    expect(records?.targetRecordTypes).toHaveLength(10);
    expect(records?.excludeSource).toBe(true);
    expect(
      getReferenceFieldContract(
        "cartulary.view.comm_log.v1",
        "comm_log.attendee_party_ids",
      )?.addOperation,
    ).toBe("add_party_ref");
  });
});
