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
    const owners = new Set(
      listWorkbookSurfaceRegistrations().map((entry) => entry.ownerId),
    );
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
