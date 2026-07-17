import { describe, expect, it } from "vitest";

import { uniqueEmail, uniqueIncidentKey, uniqueTxn } from "./fixtureIdentity";

describe("fixture identity", () => {
  it("supports deterministic transaction identifiers", () => {
    expect(uniqueTxn("incident", { now: () => 42, random: () => 0.5 })).toBe(
      "incident-42-8",
    );
  });

  it("derives email and incident identities from the injected clock", () => {
    expect(uniqueEmail("member", { now: () => 36 })).toBe(
      "member-10@example.test",
    );
    expect(uniqueIncidentKey("member", { now: () => 36 })).toBe("IR-member-10");
  });
});
