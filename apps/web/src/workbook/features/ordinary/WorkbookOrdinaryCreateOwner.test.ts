import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { ordinaryCreateContributions } from "./ordinaryCreateContributions";
import { WorkbookOrdinaryCreateOwner } from "./WorkbookOrdinaryCreateOwner";

const actor = "99999999-9999-4999-8999-999999999999";
const incident = "10000000-0000-4000-8000-000000000001";
const authority = {
  actorId: actor,
  incidentId: incident,
  sessionIdentity: "same-account",
  role: "admin" as const,
  closed: false,
};
const view = (name: string) => `cartulary.view.${name}.v1`;
const owner = () => {
  const result = new WorkbookOrdinaryCreateOwner(
    incident,
    ordinaryCreateContributions,
  );
  result.setAuthority(authority);
  return result;
};
const prepare = (name: string, values: Record<string, string | null>) => {
  const contract = requireViewContract(view(name));
  const contribution = ordinaryCreateContributions.find((item) =>
    item.views.includes(contract.viewSchemaId),
  );
  if (!contribution) throw new Error("Missing ordinary contribution.");
  return contribution.prepare(contract, values, "ordinary-test");
};
const minimumDrafts = {
  hosts: { "host.hostname": " host.example " },
  identities: { "identity.email": " Analyst@Example.test " },
  parties: {
    "party.display_name": " Operator ",
    "party.party_kind": "person",
  },
  task_requests: {
    "task.title": " Investigate ",
    "task.task_kind": "question",
  },
  decisions: {
    "decision.summary": " Isolate ",
    "decision.decision_type": "containment",
    "decision.rationale": " Rationale ",
  },
  comm_log: {
    "comm_log.comm_type": "briefing",
    "comm_log.audience": " Operations ",
    "comm_log.channel_or_meeting": " Bridge ",
    "comm_log.summary": " Update ",
  },
  handoff: {
    "handoff.incoming_owner_user_id": actor,
    "handoff.current_state_summary": " Current state ",
  },
  status_review: {
    "status_review.current_state_summary": " Current state ",
  },
  lesson: { "lesson.summary": " Summary " },
  findings: { "finding.statement": " Finding " },
  investigative_queries: {
    "investigative_query.platform": " Platform ",
    "investigative_query.purpose": " Purpose ",
    "investigative_query.query_text": " select\r\n\tvalue ",
  },
  forensic_keywords: {
    "forensic_keyword.pattern": " pattern ",
    "forensic_keyword.reason": " why ",
  },
  evidence: { "evidence.title": " Evidence " },
  indicators: {
    "indicator.indicator_type": "domain_name",
    "indicator.value_kind": "atomic",
    "indicator.display_value": "EXAMPLE.TEST",
  },
};

describe("ordinary workbook authoring", () => {
  it("covers every declared create field and input with explicit clear and source lifecycle dispositions", () => {
    const special: Record<string, string> = {
      "party.primary_email": "author@example.test",
      "party.timezone_name": "America/New_York",
      "identity.email": "author@example.test",
      "host.aad_device_id": actor,
      "identity.aad_object_id": actor,
      "identity.sid": "S-1-5-21",
      "indicator.indicator_type": "domain_name",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "EXAMPLE.TEST",
      "indicator.hash_algorithm": "sha256",
      "indicator.hash_value": "a".repeat(64),
    };
    // These declared fields require a matching Task state. The baseline draft is open.
    const initialRejections = new Set([
      "task.blocked_reason",
      "task.completed_at",
    ]);
    let covered = 0;
    for (const [name, base] of Object.entries(minimumDrafts)) {
      const contract = requireViewContract(view(name));
      for (const field of contract.fields.filter(
        (field) => field.createWritable,
      )) {
        const key = field.fieldKey;
        const raw =
          special[key] ??
          (field.readKind === "timestamp"
            ? "2026-09-13T12:34:56.123456789Z"
            : field.directReferenceContractId
              ? actor
              : (field.enumValues?.[0] ??
                (field.readKind === "number"
                  ? "50"
                  : field.readKind === "boolean"
                    ? "false"
                    : field.writeKind === "action_payload" &&
                        !key.endsWith("aliases") &&
                        key !== "handoff.open_risk_refs"
                      ? actor
                      : "Authored value")));
        const values: Record<string, string> = { ...base, [key]: raw };
        if (key.startsWith("indicator.hash_"))
          Object.assign(values, {
            "indicator.indicator_type": "sha256",
            "indicator.hash_algorithm": "sha256",
            "indicator.hash_value": "a".repeat(64),
            "indicator.display_value": "a".repeat(64),
          });
        const result = prepare(name, values);
        if (initialRejections.has(key)) expect(result.request, key).toBeNull();
        else {
          expect(result.errors, key).toEqual({});
          expect(result.request, key).toHaveProperty(key);
        }
        const cleared = prepare(name, { ...base, [key]: null });
        if (field.clearable)
          expect(cleared.request, `${key} explicit null`).toHaveProperty(
            key,
            null,
          );
        else expect(cleared.request, `${key} rejects null`).toBeNull();
        covered++;
      }
      for (const input of contract.createInputs) {
        expect(
          prepare(name, { ...base, [input.inputKey]: actor }).request,
          input.inputKey,
        ).toBeNull();
      }
    }
    expect(covered).toBeGreaterThan(100);
  });
  it("covers the frozen fourteen-schema minimum matrix without manufactured defaults", () => {
    const runtime = owner();
    expect(Object.keys(runtime.getSnapshot().schemas)).toHaveLength(14);
    for (const [name, values] of Object.entries(minimumDrafts)) {
      expect(prepare(name, {}).request, name).toBeNull();
      expect(runtime.getSnapshot().schemas[view(name)]?.ready, name).toBe(
        false,
      );
      expect(prepare(name, values).errors, name).toEqual({});
      expect(prepare(name, values).request, name).not.toBeNull();
    }
  });
  it("preserves omission, explicit clears, exact raw text and contract-specific normalization", () => {
    const omitted = prepare("evidence", { "evidence.title": " Cafe\u0301 " });
    expect(omitted.request).toEqual({
      client_txn_id: "ordinary-test",
      "evidence.title": "Café",
    });
    expect(
      prepare("evidence", {
        "evidence.title": "Title",
        "evidence.requested_at": null,
      }).request,
    ).toHaveProperty("evidence.requested_at", null);
    expect(
      prepare("evidence", { "evidence.requested_at": " 2026-09-13T00:00:00Z " })
        .request,
    ).toBeNull();
    expect(
      prepare("evidence", { "evidence.requested_at": "2026-02-30T00:00:00Z" })
        .request,
    ).toBeNull();
    expect(
      prepare("evidence", { "evidence.collector_party_id": actor }).request,
    ).toBeNull();
    expect(
      prepare("evidence", { "evidence.storage_ref": "object://reserved" })
        .request,
    ).toBeNull();
    expect(
      prepare("evidence", { "evidence.lifecycle_state": "available" }).request,
    ).toBeNull();
    expect(
      prepare("evidence", { "evidence.lifecycle_state": "released" }).request,
    ).toBeNull();
    expect(
      prepare("evidence", {
        "evidence.title": "Title",
        "evidence.initial_object_blob_id": actor,
      }).request,
    ).toBeNull();
    expect(
      prepare("status_review", {
        "status_review.current_state_summary": "Summary",
      }).request,
    ).not.toHaveProperty("coordination.source_record_id");
    expect(
      prepare("hosts", { "host.aad_device_id": actor }).request,
    ).not.toBeNull();
    expect(prepare("hosts", { "host.aliases": "alias" }).request).toBeNull();
    expect(
      prepare("identities", { "identity.sid": "S-1-5-21" }).request,
    ).not.toBeNull();
    expect(
      prepare("parties", {
        "party.display_name": "Party",
        "party.party_kind": "person",
        "party.primary_email": " bad @example.test ",
      }).request,
    ).toBeNull();
    expect(
      prepare("forensic_keywords", {
        "forensic_keyword.pattern": "x",
        "forensic_keyword.reason": " a\r\n\tb ",
      }).request,
    ).toHaveProperty("forensic_keyword.reason", "a\n\tb");
    const runtime = owner();
    runtime.update(view("evidence"), "evidence.title", " exact\r\n draft ");
    expect(
      runtime.getSnapshot().schemas[view("evidence")]?.draft.values[
        "evidence.title"
      ],
    ).toBe(" exact\r\n draft ");
    for (const timezone of ["America/New_York", "US/Eastern"]) {
      expect(
        prepare("parties", {
          "party.display_name": "Party",
          "party.party_kind": "person",
          "party.timezone_name": timezone,
        }).request,
      ).toHaveProperty("party.timezone_name", timezone);
    }
    for (const timezone of ["america/new_york", "+05:00", "../UTC"])
      expect(
        prepare("parties", {
          "party.display_name": "Party",
          "party.party_kind": "person",
          "party.timezone_name": timezone,
        }).request,
      ).toBeNull();
    expect(
      prepare("task_requests", {
        "task.title": "Task",
        "task.task_kind": "question",
        "task.status": "blocked",
      }).request,
    ).toBeNull();
    expect(
      prepare("task_requests", {
        "task.title": "Task",
        "task.task_kind": "question",
        "task.blocked_reason": "unexpected",
      }).request,
    ).toBeNull();
    expect(
      prepare("decisions", {
        "decision.summary": "Decision",
        "decision.decision_type": "scope",
        "decision.rationale": "why",
        "decision.status": "superseded",
      }).request,
    ).toBeNull();
    const indicator = {
      "indicator.indicator_type": "ipv4_addr",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "192.0.2.1",
    };
    expect(prepare("indicators", indicator).request).not.toBeNull();
    expect(
      prepare("indicators", { ...indicator, "indicator.value_kind": "pattern" })
        .request,
    ).toBeNull();
    expect(
      prepare("indicators", {
        ...indicator,
        "indicator.hash_algorithm": "sha256",
      }).request,
    ).toBeNull();
  });
  it("shares one draft across detachments and retains chosen identities off-page", () => {
    const runtime = owner(),
      schema = view("handoff");
    const detachGrid = runtime.attach(schema, Symbol("grid"));
    const detachInspector = runtime.attach(schema, Symbol("inspector"));
    runtime.update(schema, "handoff.current_state_summary", "  retained  ");
    runtime.selectReferences(schema, "handoff.incoming_owner_user_id", [
      {
        recordId: actor,
        displayText: "Analyst",
        viewSchemaId: "incident_members",
      },
    ]);
    const before = runtime.getSnapshot().schemas[schema]?.draft;
    detachGrid();
    detachInspector();
    runtime.attach(schema, Symbol("saved view"));
    expect(runtime.getSnapshot().schemas[schema]?.draft).toBe(before);
    expect(
      before?.references["handoff.incoming_owner_user_id"]?.[0]?.recordId,
    ).toBe(actor);
    runtime.update(view("evidence"), "evidence.title", "Independent sheet");
    expect(runtime.getSnapshot().schemas[schema]?.draft).toBe(before);
  });
  it("retains copyable work for role loss and closure but conceals and retires by security lifetime", () => {
    const runtime = owner(),
      schema = view("evidence");
    runtime.update(schema, "evidence.title", "private draft");
    runtime.setAuthority({ ...authority, role: "viewer" });
    expect(
      runtime.getSnapshot().schemas[schema]?.values["evidence.title"],
    ).toBe("private draft");
    expect(runtime.canAuthor()).toBe(false);
    runtime.update(schema, "evidence.title", "unauthorized");
    runtime.closeIncident();
    runtime.setAuthority({ ...authority, closed: true });
    expect(
      runtime.getSnapshot().schemas[schema]?.values["evidence.title"],
    ).toBe("private draft");
    expect(runtime.getSnapshot().schemas[schema]?.ready).toBe(false);
    runtime.suspend();
    expect(runtime.getSnapshot()).toEqual({ authority: null, schemas: {} });
    runtime.setAuthority(authority);
    expect(
      runtime.getSnapshot().schemas[schema]?.values["evidence.title"],
    ).toBe("private draft");
    runtime.setAuthority({
      ...authority,
      actorId: "88888888-8888-4888-8888-888888888888",
    });
    expect(
      runtime.getSnapshot().schemas[schema]?.values["evidence.title"],
    ).toBeUndefined();
    runtime.retire();
    expect(runtime.getSnapshot().schemas).toEqual({});
  });
});
