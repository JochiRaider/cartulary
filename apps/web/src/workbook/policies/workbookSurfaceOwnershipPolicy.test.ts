import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { listWorkbookSurfaceRegistrations } from "../models/workbookSurfaceRegistration";

const currentDirectory = path.dirname(fileURLToPath(import.meta.url));
const workbookDirectory = path.resolve(currentDirectory, "..");

describe("workbook surface ownership policy", () => {
  it("keeps owner policy definitions pure and transport-free", () => {
    for (const file of [
      "artifactSurfacePolicies.ts",
      "assessmentSurfacePolicies.ts",
      "captureTimelineSurfacePolicies.ts",
      "coordinationSurfacePolicies.ts",
      "entitiesObservationsSurfacePolicies.ts",
      "evidenceSurfacePolicies.ts",
    ]) {
      const source = readFileSync(path.join(currentDirectory, file), "utf8");
      expect(source).not.toMatch(/\bfetch\s*\(/u);
      expect(source).not.toContain("apiPath(");
      expect(source).not.toContain("authorization");
    }
  });

  it("keeps domain workflow branches out of the common contract surface", () => {
    const source = readFileSync(
      path.join(workbookDirectory, "components/GenericWorkbookSurface.tsx"),
      "utf8",
    );
    expect(source).not.toContain(
      'ownerBindings.includes("evidence_lifecycle")',
    );
    expect(source).not.toContain('ownerBindings.includes("task_lifecycle")');
    expect(source).not.toContain(
      'ownerBindings.includes("decision_supersede")',
    );
    expect(source).not.toContain('field_key: "task.status"');
    expect(source).not.toContain("/supersede");
    expect(source).not.toContain("createAndAttachEvidenceBlob");
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
