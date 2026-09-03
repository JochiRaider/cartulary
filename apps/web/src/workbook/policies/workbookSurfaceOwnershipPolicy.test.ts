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

    for (const [
      rootPath,
      compositionPath,
      ownerModulePath,
      compositionHook,
      ownerComponent,
    ] of [
      [
        "components/GenericWorkbookSurface.tsx",
        "features/generic/useGenericWorkbookInspectorComposition.tsx",
        "features/generic/GenericWorkbookInspector.tsx",
        "useGenericWorkbookInspectorComposition",
        "GenericWorkbookInspector",
      ],
      [
        "components/EntityWorkbookSurface.tsx",
        "features/entities/useEntityWorkbookInspectorComposition.tsx",
        "features/entities/EntityWorkbookInspector.tsx",
        "useEntityWorkbookInspectorComposition",
        "EntityWorkbookInspector",
      ],
      [
        "components/AssessmentWorkbookSurface.tsx",
        "features/assessments/useAssessmentWorkbookInspectorComposition.tsx",
        "features/assessments/AssessmentWorkbookInspector.tsx",
        "useAssessmentWorkbookInspectorComposition",
        "AssessmentWorkbookInspector",
      ],
    ] as const) {
      const rootSource = readFileSync(
        path.join(workbookDirectory, rootPath),
        "utf8",
      );
      const ownerSource = readFileSync(
        path.join(workbookDirectory, ownerModulePath),
        "utf8",
      );
      const compositionSource = readFileSync(
        path.join(workbookDirectory, compositionPath),
        "utf8",
      );
      expect(rootSource, rootPath).toContain(`${compositionHook}({`);
      expect(rootSource, rootPath).not.toContain(`<${ownerComponent}`);
      expect(compositionSource, compositionPath).toContain(
        `<${ownerComponent}`,
      );
      expect(rootSource, rootPath).not.toContain("inspectorConfig.panels");
      expect(rootSource, rootPath).not.toContain(
        "WorkbookInspectorPanelSection",
      );
      expect(rootSource, rootPath).not.toContain("WorkbookInspectorShell");
      expect(rootSource, rootPath).not.toContain(
        "InspectorContextualCapability",
      );
      expect(rootSource, rootPath).not.toContain("capability.kind");
      expect(rootSource, rootPath).not.toContain("InspectorPanelId");
      expect(rootSource, rootPath).not.toContain("inspectorPanelIsDeclared");
      expect(ownerSource, ownerModulePath).toContain(
        "WorkbookInspectorRecordHistory",
      );
      expect(ownerSource, ownerModulePath).toContain(
        "WorkbookInspectorDeclaredPanelList",
      );
    }

    const panelListSource = readFileSync(
      path.join(
        workbookDirectory,
        "inspector/WorkbookInspectorDeclaredPanelList.tsx",
      ),
      "utf8",
    );
    expect(panelListSource).toContain("config.panels.map");
    for (const forbiddenOwnerFacade of [
      "mutationCommands",
      "onRefresh",
      "persistence",
      "selectedRow",
      "socket",
    ]) {
      expect(panelListSource).not.toContain(forbiddenOwnerFacade);
    }
    for (const ownerModulePath of [
      "features/generic/GenericWorkbookInspector.tsx",
      "features/entities/EntityWorkbookInspector.tsx",
      "features/assessments/AssessmentWorkbookInspector.tsx",
      "timeline/components/TimelineWorkbookInspector.tsx",
    ]) {
      const ownerSource = readFileSync(
        path.join(workbookDirectory, ownerModulePath),
        "utf8",
      );
      expect(ownerSource, ownerModulePath).not.toContain("config.panels.map");
      expect(ownerSource, ownerModulePath).not.toContain(
        "inspectorConfig.panels.map",
      );
      expect(ownerSource, ownerModulePath).toContain(
        "WorkbookInspectorDeclaredPanelList",
      );
    }
  });

  it("keeps assessment semantics out of Timeline and vendor grid imports behind the adapter", () => {
    const timelineModel = [
      "timeline/models/timelineFieldRegistry.ts",
      "timeline/models/timelineLayoutPolicy.ts",
      "timeline/models/timelineMutationIntents.ts",
      "timeline/models/timelineRowModel.ts",
    ]
      .map((modelPath) =>
        readFileSync(path.join(workbookDirectory, modelPath), "utf8"),
      )
      .join("\n");
    for (const assessmentOwnedSymbol of [
      "AssessmentApiRow",
      "AssessmentConfidenceBand",
      "AssessmentCreateDraft",
      "AssessmentSubjectType",
      "AssessmentSupportCandidate",
      "buildAssessmentCreatePayload",
      "confidenceScoreFromBand",
      "initialAssessmentDraft",
    ]) {
      expect(timelineModel).not.toContain(assessmentOwnedSymbol);
    }

    for (const assessmentOwnedFile of [
      "components/AssessmentWorkbookSurface.tsx",
      "components/WorkbookRecordCandidatePicker.tsx",
      "models/assessmentWorkbookModel.ts",
    ]) {
      const source = readFileSync(
        path.join(workbookDirectory, assessmentOwnedFile),
        "utf8",
      );
      expect(source, assessmentOwnedFile).not.toContain("react-data-grid");
      expect(source, assessmentOwnedFile).not.toContain(
        "timeline/models/timeline",
      );
    }
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
