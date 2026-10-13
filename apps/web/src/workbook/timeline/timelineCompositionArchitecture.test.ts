import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import * as ts from "typescript";
import { describe, expect, expectTypeOf, it } from "vitest";
import type { TimelineWorkbookPresentationRuntime } from "./presentation/useTimelineWorkbookPresentation";

const timelineDirectory = path.dirname(fileURLToPath(import.meta.url));
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

function parsed(source: string): ts.SourceFile {
  return ts.createSourceFile(
    "subject.tsx",
    source,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  );
}
function stateImports(source: ts.SourceFile): string[] {
  const hooks = new Set([
    "useState",
    "useRef",
    "useEffect",
    "useLayoutEffect",
    "useMemo",
    "useCallback",
    "useReducer",
    "useSyncExternalStore",
  ]);
  const result: string[] = [];
  for (const statement of source.statements) {
    if (
      !ts.isImportDeclaration(statement) ||
      !ts.isStringLiteral(statement.moduleSpecifier) ||
      statement.moduleSpecifier.text !== "react" ||
      statement.importClause?.isTypeOnly
    )
      continue;
    const bindings = statement.importClause?.namedBindings;
    if (bindings && ts.isNamedImports(bindings)) {
      for (const element of bindings.elements)
        if (
          !element.isTypeOnly &&
          hooks.has((element.propertyName ?? element.name).text)
        )
          result.push((element.propertyName ?? element.name).text);
    } else if (bindings || statement.importClause?.name) {
      // Namespace/default React imports obscure the region's state ownership.
      result.push("React namespace");
    }
  }
  return result;
}
function exportsFunction(
  source: ts.SourceFile,
  matches: (name: string) => boolean,
): boolean {
  return source.statements.some(
    (node) =>
      ts.isFunctionDeclaration(node) &&
      node.name &&
      matches(node.name.text) &&
      node.modifiers?.some(
        (modifier) => modifier.kind === ts.SyntaxKind.ExportKeyword,
      ),
  );
}

describe("Timeline workbook composition architecture", () => {
  it("keeps the public root slim and makes it the sole composition-hook caller", () => {
    // Import direction and sole composition authority are enforced by the
    // existing frontend import policy; local aliases and JSX spelling are free.
    const roots = productionTypeScriptPaths(timelineDirectory)
      .map((file) => ({ file, syntax: parsed(readFileSync(file, "utf8")) }))
      .filter(({ syntax }) =>
        exportsFunction(syntax, (name) => name === "TimelineWorkbook"),
      );
    expect(roots).toHaveLength(1);
    for (const { file, syntax } of roots) {
      expect(stateImports(syntax), file).toEqual([]);
      const dom: string[] = [];
      const visit = (node: ts.Node) => {
        if (
          ts.isIdentifier(node) &&
          (node.text === "window" || node.text === "document")
        )
          dom.push(node.text);
        ts.forEachChild(node, visit);
      };
      visit(syntax);
      expect(dom, file).toEqual([]);
    }
    expect(
      stateImports(parsed('import {useState as localState} from "react";')),
    ).toEqual(["useState"]);
    expect(
      stateImports(
        parsed(
          '// useState()\nimport type {ReactNode} from "react"; export function RenamedView() {return <div/>;}',
        ),
      ),
    ).toEqual([]);
  });

  it("keeps presentation regions stateless and excludes privileged runtime capabilities", () => {
    for (const file of productionTypeScriptPaths(presentationDirectory)) {
      const syntax = parsed(readFileSync(file, "utf8"));
      // Presentation hooks own binding lifecycle; component regions do not.
      if (!exportsFunction(syntax, (name) => /^use[A-Z]/u.test(name)))
        expect(stateImports(syntax), file).toEqual([]);
    }
    type Privileged =
      | "mutationRuntime"
      | "pendingMutationPort"
      | "collaborationProjection"
      | "onAuthorityUncertain";
    expectTypeOf<
      Extract<keyof TimelineWorkbookPresentationRuntime, Privileged>
    >().toEqualTypeOf<never>();
  });
});
