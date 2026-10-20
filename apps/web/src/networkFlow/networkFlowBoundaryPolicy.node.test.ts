import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { cartularyDesignTokenVars } from "@cartulary/ui-contracts";
import * as ts from "typescript";
import { describe, expect, it } from "vitest";

const networkFlowDirectory = path.dirname(fileURLToPath(import.meta.url));
const webSourceDirectory = path.resolve(networkFlowDirectory, "..");

function sourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      return sourceFiles(target);
    }
    return /\.(?:ts|tsx)$/u.test(entry.name) && !entry.name.includes(".test.")
      ? [target]
      : [];
  });
}

describe("Network Flow frontend boundary policy", () => {
  it("checks contract references and browser evidence APIs independently of helper names", () => {
    for (const file of sourceFiles(webSourceDirectory))
      expect(contractReferences(readFileSync(file, "utf8")), file).toEqual([]);
    for (const file of [
      path.resolve(webSourceDirectory, "../e2e/network-flow.spec.ts"),
      ...sourceFiles(path.resolve(webSourceDirectory, "../e2e/support")),
    ])
      expect(
        syntheticContentReferences(readFileSync(file, "utf8")),
        file,
      ).toEqual([]);
    for (const source of [
      "renamed.setContent(markup)",
      'const replace = extracted["setContent"]; replace(markup)',
      "const {setContent: replace} = renamed; replace(markup)",
    ])
      expect(syntheticContentReferences(source), source).toEqual([
        "setContent",
      ]);
    expect(
      syntheticContentReferences(
        '// page.setContent\nconst label = "setContent"; await renamedOpen();',
      ),
    ).toEqual([]);
    expect(
      contractReferences(
        'import type { GraphProjectionEphemeralInput as Local } from "contract";',
      ),
    ).toEqual(["GraphProjectionEphemeralInput"]);
    expect(contractReferences('const schema = {["$defs"]: {}};')).toEqual([
      "$defs",
    ]);
    expect(
      contractReferences(
        '// GraphProjectionEphemeralInput $defs\nconst label = "schema $defs reference";',
      ),
    ).toEqual([]);
  });
  it("uses only authored Cartulary design-token variables in Network Flow source", () => {
    const authored = new Set(Object.keys(cartularyDesignTokenVars));
    const unknown = sourceFiles(networkFlowDirectory).flatMap((file) => {
      const source = readFileSync(file, "utf8");
      return [...source.matchAll(/--ct-[A-Za-z0-9-]+/gu)]
        .map((match) => match[0])
        .filter((token) => !authored.has(token))
        .map((token) => `${path.basename(file)}: ${token}`);
    });
    expect([...new Set(unknown)].sort()).toEqual([]);
  });

  it("rejects legacy light-theme variable fallbacks in Network Flow source", () => {
    for (const file of sourceFiles(networkFlowDirectory)) {
      const source = readFileSync(file, "utf8");
      expect(source, file).not.toMatch(/var\(--color-/u);
      expect(source, file).not.toMatch(/#[fF]{3,6}\b/u);
    }
  });
});

function sourceSyntax(source: string) {
  return ts.createSourceFile(
    "renamable.tsx",
    source,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  );
}

function propertyIdentity(node: ts.Node): string | undefined {
  const key = ts.isPropertyAccessExpression(node)
    ? node.name
    : ts.isElementAccessExpression(node)
      ? node.argumentExpression
      : ts.isBindingElement(node)
        ? (node.propertyName ?? node.name)
        : ts.isPropertyAssignment(node) || ts.isPropertySignature(node)
          ? node.name
          : undefined;
  const value = key && ts.isComputedPropertyName(key) ? key.expression : key;
  return value && (ts.isIdentifier(value) || ts.isStringLiteralLike(value))
    ? value.text
    : undefined;
}

function syntheticContentReferences(source: string): string[] {
  const found = new Set<string>();
  const visit = (node: ts.Node) => {
    if (propertyIdentity(node) === "setContent") found.add("setContent");
    ts.forEachChild(node, visit);
  };
  visit(sourceSyntax(source));
  return [...found];
}

function contractReferences(source: string): string[] {
  const found = new Set<string>();
  const visit = (node: ts.Node) => {
    if (propertyIdentity(node) === "$defs") found.add("$defs");
    if (ts.isIdentifier(node) && node.text === "GraphProjectionEphemeralInput")
      found.add(node.text);
    if (
      ts.isStringLiteralLike(node) &&
      node.text === "graph_projection_ephemeral_input"
    )
      found.add(node.text);
    ts.forEachChild(node, visit);
  };
  visit(sourceSyntax(source));
  return [...found].sort();
}
