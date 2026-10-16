import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import type { GridCellAnchor, GridHandle } from "@cartulary/grid-adapter";
import * as ts from "typescript";
import { describe, expect, expectTypeOf, it } from "vitest";

const workbookDirectory = path.dirname(fileURLToPath(import.meta.url));

const roots = [
  "components/GenericWorkbookSurface.tsx",
  "components/EntityWorkbookSurface.tsx",
  "components/AssessmentWorkbookSurface.tsx",
] as const;

function syntax(source: string) {
  return ts.createSourceFile(
    "boundary.tsx",
    source,
    ts.ScriptTarget.Latest,
    false,
    ts.ScriptKind.TSX,
  );
}

function runtimeImports(source: string) {
  return syntax(source).statements.flatMap((statement) => {
    if (
      !ts.isImportDeclaration(statement) ||
      !ts.isStringLiteral(statement.moduleSpecifier) ||
      statement.importClause?.isTypeOnly
    )
      return [];
    const bindings = statement.importClause?.namedBindings;
    const names =
      bindings && ts.isNamedImports(bindings)
        ? bindings.elements
            .filter((element) => !element.isTypeOnly)
            .map((element) => (element.propertyName ?? element.name).text)
        : ["*"];
    const module = statement.moduleSpecifier.text;
    return names.map((name) => ({ module, name }));
  });
}

function forbiddenConstruction(source: string) {
  const forbidden = new Set([
    "buildWorkbookInspectorSubject",
    "useWorkbookInspectorCoordinator",
    "useInspectorCreateRelatedWorkflow",
    "GenericWorkbookInspector",
    "EntityWorkbookInspector",
    "AssessmentWorkbookInspector",
  ]);
  return runtimeImports(source).filter(
    ({ module, name }) =>
      forbidden.has(name) ||
      /\/(?:useWorkbookInspectorCoordinator|useInspectorCreateRelatedWorkflow|(?:Generic|Entity|Assessment)WorkbookInspector)$/u.test(
        module,
      ),
  );
}

describe("Workbook inspector owner boundaries", () => {
  it.each(roots)("keeps inspector construction out of %s", (relativePath) => {
    const source = readFileSync(
      path.join(workbookDirectory, relativePath),
      "utf8",
    );

    expect(forbiddenConstruction(source), relativePath).toEqual([]);
    expect(
      forbiddenConstruction(
        'import { useWorkbookInspectorCoordinator as local } from "../inspector/useWorkbookInspectorCoordinator";',
      ),
    ).toHaveLength(1);
    expect(
      forbiddenConstruction(
        'import * as local from "../inspector/useWorkbookInspectorCoordinator";',
      ),
    ).toHaveLength(1);
    expect(
      forbiddenConstruction(
        '// useWorkbookInspectorCoordinator\nimport { useOwnerComposition as renamed } from "../features/owner/composition"; function extracted() { return renamed(); }',
      ),
    ).toEqual([]);
  });

  it("keeps Timeline inspector sections and model assembly in its owner unit", () => {
    const broadPresentation = readFileSync(
      path.join(
        workbookDirectory,
        "timeline/presentation/useTimelineWorkbookPresentation.tsx",
      ),
      "utf8",
    );
    // The boundary is dependency direction. Local names, the call expression,
    // argument spelling and model-property placement are private realization.
    const sectionDependencies = (source: string) =>
      runtimeImports(source).filter(({ module }) =>
        module.endsWith("/TimelineWorkbookInspectorSections"),
      );
    expect(sectionDependencies(broadPresentation)).toEqual([]);
    expect(
      sectionDependencies(
        'import { useTimelineWorkbookInspectorSections as local } from "../components/TimelineWorkbookInspectorSections";',
      ),
    ).toHaveLength(1);
    expect(
      sectionDependencies(
        'import { useTimelineInspectorPresentation as local } from "./useTimelineInspectorPresentation"; function extracted(input) { return local(input); }',
      ),
    ).toEqual([]);
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
      // Row/editor identity must come from semantic registries. Shell fallback
      // controls have separate fixed identities and are outside this boundary.
      const gridTestSelectors = new Set([
        "rowCellTestId",
        "draftCellTestId",
        "timelineScalarEditorTestId",
        "rowInspectTestId",
      ]);
      expect(
        runtimeImports(source).filter(({ name }) =>
          gridTestSelectors.has(name),
        ),
        relativePath,
      ).toEqual([]);
    }

    const editorRegistry = readFileSync(
      path.join(
        workbookDirectory,
        "timeline/editing/useTimelineEditorDraftRegistry.ts",
      ),
      "utf8",
    );
    expect(
      runtimeImports(editorRegistry).filter(({ name }) =>
        /TestId|dataTestId/u.test(name),
      ),
    ).toEqual([]);
    expectTypeOf<GridHandle["getAnchorRect"]>().toEqualTypeOf<
      (anchor: GridCellAnchor) => DOMRectReadOnly | null
    >();
    expectTypeOf<
      Extract<GridHandle["requestFocus"], undefined>
    >().toEqualTypeOf<never>();
  });
  it("keeps presentation source free of state orchestration hooks", () => {
    const directory = path.join(workbookDirectory, "inspector/presentation");
    // Presentation-local disclosure/highlight state is legitimate. Domain
    // orchestration belongs to the owning coordinator and mutation adapters.
    const stateHooks = (source: string) => [
      ...forbiddenConstruction(source),
      ...runtimeImports(source).filter(({ module }) =>
        /\/(?:services|mutations|runtime)\//u.test(module),
      ),
    ];
    for (const filename of readdirSync(directory)) {
      if (!/\.tsx?$/u.test(filename) || /\.test\.tsx?$/u.test(filename))
        continue;
      expect(
        stateHooks(readFileSync(path.join(directory, filename), "utf8")),
        filename,
      ).toEqual([]);
    }
    expect(
      stateHooks(
        'import { useWorkbookInspectorCoordinator as local } from "../inspector/useWorkbookInspectorCoordinator";',
      ),
    ).toHaveLength(1);
    expect(
      stateHooks(
        'import { useState as local } from "react"; // useWorkbookInspectorCoordinator',
      ),
    ).toEqual([]);
  });
});
