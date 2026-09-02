import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const workbookDirectory = path.dirname(fileURLToPath(import.meta.url));

const roots = [
  "components/GenericWorkbookSurface.tsx",
  "components/EntityWorkbookSurface.tsx",
  "components/AssessmentWorkbookSurface.tsx",
] as const;

describe("Workbook inspector owner boundaries", () => {
  it.each(roots)("keeps inspector construction out of %s", (relativePath) => {
    const source = readFileSync(
      path.join(workbookDirectory, relativePath),
      "utf8",
    );

    for (const forbidden of [
      "buildWorkbookInspectorSubject",
      "useWorkbookInspectorCoordinator",
      "useInspectorCreateRelatedWorkflow",
      "<GenericWorkbookInspector",
      "<EntityWorkbookInspector",
      "<AssessmentWorkbookInspector",
      "workflowContent=",
      "relationshipsContent=",
      "detailsContent=",
    ]) {
      expect(source, `${relativePath}: ${forbidden}`).not.toContain(forbidden);
    }
  });

  it("keeps Timeline inspector sections and model assembly in its owner unit", () => {
    const broadPresentation = readFileSync(
      path.join(
        workbookDirectory,
        "timeline/presentation/useTimelineWorkbookPresentation.tsx",
      ),
      "utf8",
    );
    const inspectorPresentation = readFileSync(
      path.join(
        workbookDirectory,
        "timeline/presentation/useTimelineInspectorPresentation.tsx",
      ),
      "utf8",
    );

    expect(broadPresentation).not.toContain(
      "useTimelineWorkbookInspectorSections",
    );
    expect(broadPresentation).not.toContain("rowHistoryRecordId:");
    expect(inspectorPresentation).toContain(
      "useTimelineWorkbookInspectorSections(sections)",
    );
    expect(inspectorPresentation).toContain("rowHistoryRecordId:");
  });

  it("keeps inspector focus and geometry on semantic owner registries", () => {
    const semanticFocusPaths = [
      "components/GenericWorkbookSurface.tsx",
      "components/EntityWorkbookSurface.tsx",
      "timeline/composition/useTimelineGridEnvironment.ts",
      "timeline/hooks/useTimelineViewportContinuityController.ts",
      "timeline/hooks/useTimelineKeyboardController.ts",
      "timeline/hooks/useTimelineInspectorSelection.ts",
      "timeline/components/useTimelineCollectionRenderer.tsx",
    ] as const;
    for (const relativePath of semanticFocusPaths) {
      const source = readFileSync(
        path.join(workbookDirectory, relativePath),
        "utf8",
      );
      expect(source, relativePath).not.toContain("dataTestIdSelector");
    }

    const editorRegistry = readFileSync(
      path.join(
        workbookDirectory,
        "timeline/editing/useTimelineEditorDraftRegistry.ts",
      ),
      "utf8",
    );
    expect(editorRegistry).not.toContain("inputTestId");
    expect(editorRegistry).not.toContain("dataTestId");

    const gridHandle = readFileSync(
      path.resolve(
        workbookDirectory,
        "../../../../packages/grid-adapter/src/core.ts",
      ),
      "utf8",
    );
    expect(gridHandle).toContain(
      "readonly focusDraftCell: (fieldKey: string) => boolean",
    );
    expect(gridHandle).toContain(
      "readonly getAnchorRect: (anchor: GridCellAnchor) => DOMRectReadOnly | null",
    );
    expect(gridHandle).not.toContain("readonly focusDraftCell?:");
    expect(gridHandle).not.toContain("readonly getAnchorRect?:");
  });
});
