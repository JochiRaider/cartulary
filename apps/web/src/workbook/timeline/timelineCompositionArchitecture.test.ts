import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const timelineDirectory = path.dirname(fileURLToPath(import.meta.url));
const rootPath = path.join(
  timelineDirectory,
  "components",
  "TimelineWorkbook.tsx",
);
const presentationDirectory = path.join(timelineDirectory, "presentation");

function productionTypeScriptPaths(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const absolutePath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      return productionTypeScriptPaths(absolutePath);
    }
    if (
      !entry.isFile() ||
      (!entry.name.endsWith(".ts") && !entry.name.endsWith(".tsx")) ||
      entry.name.endsWith(".test.ts") ||
      entry.name.endsWith(".test.tsx")
    ) {
      return [];
    }
    return [absolutePath];
  });
}

describe("Timeline workbook composition architecture", () => {
  it("keeps the public root slim and makes it the sole composition-hook caller", () => {
    const rootSource = readFileSync(rootPath, "utf8");
    expect(rootSource).toContain("TimelineCollaborationBoundary");
    expect(rootSource).toContain("useTimelineWorkbookComposition({ runtime })");
    expect(rootSource).toContain("useTimelineWorkbookPresentation({");
    expect(rootSource).toContain("composition: composition.presentation");
    expect(rootSource).toContain(
      "<TimelineWorkbookView model={presentation} />",
    );
    expect(rootSource).not.toMatch(
      /\buse(?:State|Ref|Effect|LayoutEffect|Memo|Callback)\s*\(/u,
    );
    expect(rootSource).not.toMatch(/\b(?:document|window)\b/u);

    const compositionCallers = productionTypeScriptPaths(
      timelineDirectory,
    ).filter((file) =>
      /\buseTimelineWorkbookComposition\s*\(\s*\{\s*runtime\s*\}\s*\)/u.test(
        readFileSync(file, "utf8"),
      ),
    );
    expect(compositionCallers).toEqual([rootPath]);
  });

  it("keeps presentation regions stateless and excludes privileged runtime capabilities", () => {
    const presentationPaths = productionTypeScriptPaths(presentationDirectory);
    const regionPaths = presentationPaths.filter(
      (file) =>
        !file.endsWith("useTimelineWorkbookPresentation.tsx") &&
        !file.endsWith("useTimelineInspectorPresentation.tsx"),
    );
    for (const file of regionPaths) {
      expect(readFileSync(file, "utf8"), file).not.toMatch(
        /\buse(?:State|Ref|Effect|LayoutEffect|Memo|Callback)\s*\(/u,
      );
    }

    const presentationSource = readFileSync(
      path.join(presentationDirectory, "useTimelineWorkbookPresentation.tsx"),
      "utf8",
    );
    for (const forbiddenCapability of [
      "mutationRuntime",
      "pendingMutationPort",
      "collaborationProjection",
      "onIncidentAccessLost",
    ]) {
      expect(presentationSource).not.toContain(forbiddenCapability);
    }
    expect(presentationSource).toContain("useLayoutEffect(() => {");
    expect(presentationSource).toContain(
      "grid.commands.registerVisibleColumns(visibleTimelineColumns)",
    );
    expect(presentationSource).toContain("useTimelineInspectorPresentation({");
    expect(presentationSource).not.toContain(
      "useTimelineWorkbookInspectorSections",
    );

    const workbookDirectory = path.resolve(timelineDirectory, "..");
    const feedbackModelSource = readFileSync(
      path.join(workbookDirectory, "inspector/workbookInspectorErrorModel.ts"),
      "utf8",
    );
    expect(feedbackModelSource).toContain('readonly kind: "message"');
    expect(feedbackModelSource).toContain('readonly kind: "error"');
    expect(feedbackModelSource).not.toMatch(
      /WorkbookInspectorFeedback\s*=\s*(?:\r?\n\s*)?\|?\s*string\b/u,
    );

    for (const relativePath of [
      "features/indicators/IndicatorInspectorWorkflow.tsx",
      "timeline/components/TimelineMentionsPanel.tsx",
      "timeline/components/TimelineWorkbookInspector.tsx",
      "timeline/composition/useTimelineInspectorWorkflowComposition.ts",
      "timeline/hooks/useTimelineCreateRelatedWorkflow.ts",
      "timeline/hooks/useTimelineEvidenceAttach.ts",
      "timeline/hooks/useTimelineInspectorFeatureController.ts",
      "timeline/hooks/useTimelineInspectorSelection.ts",
      "timeline/hooks/useTimelineKeyboardController.ts",
      "timeline/hooks/useTimelineMentionActions.ts",
    ]) {
      const source = readFileSync(
        path.join(workbookDirectory, relativePath),
        "utf8",
      );
      expect(source, relativePath).not.toMatch(
        /\b(?:onSetInspectorMessage|setInspectorMessage|setMessage)\(\s*(?:"|'|`)/u,
      );
      expect(source, relativePath).not.toMatch(/\btypeof\s+message\b/u);
      expect(source, relativePath).not.toContain("function InspectorMessage");
    }
  });
});
