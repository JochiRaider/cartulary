import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import {
  auditActionCodes,
  auditBinding,
  auditTargetKinds,
  emptyAuditQuery,
  normalizeAuditInstant,
  validateAuditInputs,
} from "./administrativeAuditModel";

describe("Administrative audit exact query", () => {
  it("projects current filter vocabulary from the authored operation", () => {
    const document = JSON.parse(
      readFileSync(
        resolve(
          process.cwd(),
          "../../contracts/openapi-source/owners/module.auth/openapi.json",
        ),
        "utf8",
      ),
    );
    const parameters = document.paths["/api/v1/administrative-audit-events"].get
      .parameters as { name: string; schema: { enum?: string[] } }[];
    expect(auditActionCodes).toEqual(
      parameters.find((p) => p.name === "action_code")?.schema.enum,
    );
    expect(auditTargetKinds).toEqual(
      parameters.find((p) => p.name === "target_kind")?.schema.enum,
    );
  });
  it("validates exact filters and the target dependency without aliases", () => {
    expect(
      validateAuditInputs({ ...emptyAuditQuery, target_id: "target" }).errors
        .target_kind,
    ).toBeTruthy();
    for (const target_id of ["null", "a,b", "[a]", "a]"])
      expect(
        validateAuditInputs({
          ...emptyAuditQuery,
          target_kind: "user",
          target_id,
        }).errors.target_id,
      ).toBeTruthy();
    for (const action_code of ["future_action", "constructor", "User created"])
      expect(
        validateAuditInputs({ ...emptyAuditQuery, action_code }).errors
          .action_code,
      ).toBeTruthy();
    expect(
      validateAuditInputs({
        ...emptyAuditQuery,
        actor_user_id: "someone@example.test",
      }).errors.actor_user_id,
    ).toBeTruthy();
    expect(
      validateAuditInputs({
        ...emptyAuditQuery,
        target_kind: "user",
        target_id: " ExactCase ",
      }),
    ).toMatchObject({ query: { target_id: "ExactCase" }, errors: {} });
  });
  it("normalizes explicit instants and compares nanosecond range boundaries", () => {
    const lower = "2026-05-24T02:00:00.000000001+02:00";
    const upper = "2026-05-24T00:00:00.000000002Z";
    expect(
      validateAuditInputs({
        ...emptyAuditQuery,
        occurred_at_gte: lower,
        occurred_at_lt: upper,
      }),
    ).toMatchObject({
      query: { occurred_at_gte: "2026-05-24T00:00:00.000000001Z" },
      errors: {},
    });
    expect(normalizeAuditInstant(lower)?.nanoseconds).toBe(
      normalizeAuditInstant("2026-05-23T20:00:00.000000001-04:00")?.nanoseconds,
    );
    for (const occurred_at_lt of [lower, "2026-05-24T00:00:00Z"])
      expect(
        validateAuditInputs({
          ...emptyAuditQuery,
          occurred_at_gte: lower,
          occurred_at_lt,
        }).errors.occurred_at_lt,
      ).toBeTruthy();
    for (const input of [
      "2026-02-29T00:00:00Z",
      "2026-05-24",
      "2026-05-24T00:00:00",
      "2026-05-24T25:00:00Z",
      "2026-05-24T00:00:00+25:00",
    ])
      expect(normalizeAuditInstant(input)).toBeNull();
    expect(normalizeAuditInstant("2024-02-29T00:00:00Z")).not.toBeNull();
  });
  it("binds continuations to query actor and application lifetime", () => {
    const authority = { actorId: "actor", lifetime: "session-1" };
    const key = auditBinding(authority, emptyAuditQuery);
    expect(
      auditBinding({ ...authority, lifetime: "session-2" }, emptyAuditQuery),
    ).not.toBe(key);
    expect(
      auditBinding({ ...authority, actorId: "other" }, emptyAuditQuery),
    ).not.toBe(key);
    expect(
      auditBinding(authority, {
        ...emptyAuditQuery,
        action_code: "user_created",
      }),
    ).not.toBe(key);
  });
});
