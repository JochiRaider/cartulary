import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import * as ts from "typescript";
import { describe, expect, it } from "vitest";
import { workbookGridRows } from "../models/workbookContractRows";

const layoutDirectory = path.dirname(fileURLToPath(import.meta.url));
const workbookDirectory = path.resolve(layoutDirectory, "..");

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

describe("workbook layout policy", () => {
  it("preserves exactly the supplied record population without layout filler", () => {
    for (const count of [0, 1, 3, 50]) {
      const records = Array.from({ length: count }, (_, index) => ({
        id: `record-${index}`,
        version: index + 1,
      }));
      const rendered = workbookGridRows({
        rows: records,
        surface: "cartulary.view.timeline.v2",
        getRecordId: (row) => row.id,
        getRowVersion: (row) => row.version,
      });
      expect(rendered).toHaveLength(count);
      for (const [index, row] of rendered.entries()) {
        expect(row.kind).toBe("data");
        expect(row.data).toBe(records[index]);
        expect(row.rowIdentity).toEqual({
          kind: "core_record",
          recordId: records[index]?.id,
        });
        expect(row.mutationIdentity).toEqual({
          kind: "core_row_version",
          baseRowVersion: records[index]?.version,
        });
      }
    }
  });
  it("keeps surface geometry in the shared layout owner", () => {
    for (const file of productionTypeScriptPaths(workbookDirectory)) {
      if (file.startsWith(`${layoutDirectory}${path.sep}`)) continue;
      expect(surfaceGeometryProblems(readFileSync(file, "utf8")), file).toEqual(
        [],
      );
    }
    expect(
      surfaceGeometryProblems(
        'import {WorkbookSurfaceLayout as Renamed} from "../layout/WorkbookSurfaceLayout"; const view = <Renamed primaryGrid={grid}/>;',
      ),
    ).toEqual([]);
    expect(
      surfaceGeometryProblems(
        'const view = <Surface primaryGrid={grid} style={{["minHeight"]: 400}}/>;',
      ),
    ).toEqual(["minHeight"]);
    expect(
      surfaceGeometryProblems(
        "// minHeight: 400\nconst view = <Surface primaryGrid={grid}/>;",
      ),
    ).toEqual([]);
  });
});

function surfaceGeometryProblems(source: string): string[] {
  const syntax = ts.createSourceFile(
    "renamable.tsx",
    source,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  );
  let surface = false;
  const dimensions = new Set<string>();
  const visit = (node: ts.Node) => {
    if (
      ts.isJsxAttribute(node) &&
      ts.isIdentifier(node.name) &&
      node.name.text === "primaryGrid"
    )
      surface = true;
    if (ts.isPropertyAssignment(node)) {
      const key = ts.isComputedPropertyName(node.name)
        ? node.name.expression
        : node.name;
      if (
        (ts.isIdentifier(key) || ts.isStringLiteral(key)) &&
        (key.text === "minHeight" || key.text === "minBlockSize")
      )
        dimensions.add(key.text);
    }
    ts.forEachChild(node, visit);
  };
  visit(syntax);
  return surface ? [...dimensions].sort() : [];
}
