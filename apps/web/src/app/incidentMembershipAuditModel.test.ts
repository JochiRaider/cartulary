import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import {
  formatAuditJSON,
  normalizeAuditInstant,
} from "../shared/auditReadValues";
import {
  emptyMembershipAuditQuery,
  membershipAuditActions,
  membershipAuditBinding,
  membershipAuditTargets,
  validateMembershipAuditInputs,
} from "./incidentMembershipAuditModel";

describe("Incident membership audit query", () => {
  it("matches authored incident request vocabulary without admitting additive response choices", () => {
    const owner = JSON.parse(
      readFileSync(
        "../../contracts/openapi-source/owners/module.incidents/openapi.json",
        "utf8",
      ),
    );
    const operation =
      owner.paths["/api/v1/incidents/{incident_id}/membership-audit-events"]
        .get;
    expect(membershipAuditActions).toEqual(
      operation.parameters.find(
        (p: { name: string }) => p.name === "action_code",
      ).schema.enum,
    );
    expect(membershipAuditTargets).toEqual(
      operation.parameters.find(
        (p: { name: string }) => p.name === "target_kind",
      ).schema.enum,
    );
    expect(
      validateMembershipAuditInputs({
        ...emptyMembershipAuditQuery,
        action_code: "future_membership_action",
      }).errors.action_code,
    ).toBeTruthy();
  });
  it("preserves nanosecond instants explicit offsets and inclusive exclusive ordering", () => {
    const result = validateMembershipAuditInputs({
      ...emptyMembershipAuditQuery,
      occurred_at_gte: "2026-05-24T02:00:00.000000001+02:00",
      occurred_at_lt: "2026-05-24T00:00:00.000000002Z",
    });
    expect(result.errors).toEqual({});
    expect(result.query.occurred_at_gte).toBe("2026-05-24T00:00:00.000000001Z");
    expect(
      normalizeAuditInstant(result.query.occurred_at_gte)?.nanoseconds,
    ).toBeLessThan(
      required(normalizeAuditInstant(result.query.occurred_at_lt)).nanoseconds,
    );
    for (const upper of [
      result.query.occurred_at_gte,
      "2026-05-24T00:00:00",
      "2026-02-29T00:00:00Z",
      "2026-05-24",
    ])
      expect(
        validateMembershipAuditInputs({
          ...result.query,
          occurred_at_lt: upper,
        }).errors.occurred_at_lt,
      ).toBeTruthy();
  });
  it("requires exact targets and binds incident actor session query and route context", () => {
    expect(
      validateMembershipAuditInputs({
        ...emptyMembershipAuditQuery,
        target_id: "member",
      }).errors.target_kind,
    ).toBeTruthy();
    expect(
      validateMembershipAuditInputs({
        ...emptyMembershipAuditQuery,
        target_kind: "incident_membership",
        target_id: "member",
      }).errors,
    ).toEqual({});
    for (const actor_user_id of ["null", "[]", "a,b", "not-a-uuid"])
      expect(
        validateMembershipAuditInputs({
          ...emptyMembershipAuditQuery,
          actor_user_id,
        }).errors.actor_user_id,
      ).toBeTruthy();
    const a = { lifetime: "one", actorId: "actor", incidentId: "incident" };
    const initial = membershipAuditBinding(a, emptyMembershipAuditQuery);
    for (const changed of [
      { ...a, lifetime: "two" },
      { ...a, actorId: "other" },
      { ...a, incidentId: "other" },
      { ...a, apiBase: "https://other.test" },
    ])
      expect(
        membershipAuditBinding(changed, emptyMembershipAuditQuery),
      ).not.toBe(initial);
  });
  it("formats all JSON kinds faithfully without interpreting strings as markup", () => {
    expect(
      [null, "null", "", false, 0, [false, 0], { a: "<img src=x>" }].map(
        formatAuditJSON,
      ),
    ).toEqual([
      "null",
      '"null"',
      '""',
      "false",
      "0",
      "[\n  false,\n  0\n]",
      '{\n  "a": "<img src=x>"\n}',
    ]);
  });
});

function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Expected audit fixture value");
  return value;
}
