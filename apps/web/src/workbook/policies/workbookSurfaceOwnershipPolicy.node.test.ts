import { describe, expect, it } from "vitest";
import { listWorkbookSurfaceRegistrations } from "../models/workbookSurfaceRegistration";

describe("workbook surface ownership policy", () => {
  it("exposes immutable declarative owner policies without executable capabilities", () => {
    const plainData = (value: unknown): boolean => {
      if (
        value === null ||
        ["string", "boolean", "number"].includes(typeof value)
      )
        return true;
      if (typeof value !== "object" || !Object.isFrozen(value)) return false;
      if (
        !Array.isArray(value) &&
        Object.getPrototypeOf(value) !== Object.prototype
      )
        return false;
      return Object.values(value).every(plainData);
    };
    for (const registration of listWorkbookSurfaceRegistrations()) {
      expect(plainData(registration.policy), registration.viewSchemaId).toBe(
        true,
      );
      expect(Object.keys(registration.policy).sort()).toEqual([
        "collectionActions",
        "createDefaults",
        "currentUserDefaultFields",
        "ownerBindings",
        "publicErrorPresentation",
      ]);
    }
    expect(
      plainData(Object.freeze({ authorization: "a harmless label" })),
    ).toBe(true);
    expect(plainData(Object.freeze({ renamed: () => undefined }))).toBe(false);
    expect(plainData(Object.freeze({ nested: { mutable: true } }))).toBe(false);
    expect(plainData(Object.freeze(new Date()))).toBe(false);
  });
  it("assigns every registration to one declared bounded-context owner", () => {
    const registrations = listWorkbookSurfaceRegistrations();
    expect(registrations).toHaveLength(17);
    const view = (name: string, version = 1) =>
      `cartulary.view.${name}.v${version}`;
    expect(
      Object.fromEntries(
        registrations.map((entry) => [entry.viewSchemaId, entry.ownerId]),
      ),
    ).toEqual({
      [view("notes")]: "artifacts",
      [view("findings")]: "artifacts",
      [view("investigative_queries")]: "artifacts",
      [view("forensic_keywords")]: "artifacts",
      [view("assessments")]: "assessments",
      [view("timeline", 2)]: "capture_timeline",
      [view("parties")]: "coordination",
      [view("task_requests")]: "coordination",
      [view("decisions")]: "coordination",
      [view("comm_log")]: "coordination",
      [view("handoff")]: "coordination",
      [view("status_review")]: "coordination",
      [view("lesson")]: "coordination",
      [view("hosts")]: "entities_observations",
      [view("identities")]: "entities_observations",
      [view("indicators")]: "entities_observations",
      [view("evidence")]: "evidence",
    });
    const owners = new Set(registrations.map((entry) => entry.ownerId));
    expect([...owners].sort()).toEqual([
      "artifacts",
      "assessments",
      "capture_timeline",
      "coordination",
      "entities_observations",
      "evidence",
    ]);
  });
});
