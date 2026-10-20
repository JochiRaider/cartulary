import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import * as ts from "typescript";
import { describe, expect, it } from "vitest";

const webSourceDirectory = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);

function runtimeSourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      return runtimeSourceFiles(target);
    }
    return /\.(?:ts|tsx)$/u.test(entry.name) &&
      !entry.name.includes(".test.") &&
      !entry.name.endsWith("TestSupport.ts") &&
      !entry.name.endsWith("TestSupport.tsx")
      ? [target]
      : [];
  });
}

function rawTransportReferences(source: string): boolean {
  const syntax = ts.createSourceFile(
    "subject.tsx",
    source,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  );
  let checker: ts.TypeChecker | undefined;
  const symbols = () => {
    if (checker) return checker;
    const options = { noResolve: true, noLib: true };
    const host = ts.createCompilerHost(options);
    host.getSourceFile = (file) =>
      file === syntax.fileName ? syntax : undefined;
    host.fileExists = (file) => file === syntax.fileName;
    host.readFile = (file) => (file === syntax.fileName ? source : undefined);
    checker = ts
      .createProgram([syntax.fileName], options, host)
      .getTypeChecker();
    return checker;
  };
  const ambient = (name: ts.Identifier) => {
    const symbol = ts.isShorthandPropertyAssignment(name.parent)
      ? symbols().getShorthandAssignmentValueSymbol(name.parent)
      : symbols().getSymbolAtLocation(name);
    return !symbol?.declarations?.length;
  };
  const browserGlobal = (
    value: ts.Expression,
    seen = new Set<ts.Node>(),
  ): boolean => {
    if (seen.has(value)) return false;
    seen.add(value);
    if (ts.isParenthesizedExpression(value))
      return browserGlobal(value.expression, seen);
    if (!ts.isIdentifier(value)) return false;
    if (["globalThis", "window", "self"].includes(value.text) && ambient(value))
      return true;
    const declaration = symbols().getSymbolAtLocation(value)?.valueDeclaration;
    return (
      !!declaration &&
      ts.isVariableDeclaration(declaration) &&
      !!declaration.initializer &&
      browserGlobal(declaration.initializer, seen)
    );
  };
  let found = false;
  const visit = (node: ts.Node) => {
    if (
      ts.isIdentifier(node) &&
      node.text === "fetch" &&
      !(
        ts.isPropertyAccessExpression(node.parent) && node.parent.name === node
      ) &&
      ambient(node)
    )
      found = true;
    if (
      (ts.isPropertyAccessExpression(node) ||
        ts.isElementAccessExpression(node)) &&
      (ts.isPropertyAccessExpression(node)
        ? node.name.text === "fetch"
        : ts.isStringLiteral(node.argumentExpression) &&
          node.argumentExpression.text === "fetch") &&
      browserGlobal(node.expression)
    )
      found = true;
    if (
      ts.isBindingElement(node) &&
      ts.isObjectBindingPattern(node.parent) &&
      ts.isVariableDeclaration(node.parent.parent) &&
      node.parent.parent.initializer
    ) {
      const property = node.propertyName ?? node.name;
      if (
        (ts.isIdentifier(property) || ts.isStringLiteral(property)) &&
        property.text === "fetch" &&
        browserGlobal(node.parent.parent.initializer)
      )
        found = true;
    }
    ts.forEachChild(node, visit);
  };
  visit(syntax);
  return found;
}

describe("frontend transport boundary policy", () => {
  it("confines raw transport references to the service adapter owner", () => {
    const manifest = JSON.parse(
      readFileSync(
        path.resolve(
          webSourceDirectory,
          "../../../tools/frontend_source_ownership.json",
        ),
        "utf8",
      ),
    ) as { entries: { owner_id: string; paths: string[] }[] };
    const adapters = new Set(
      manifest.entries.find((entry) => entry.owner_id === "web.services")
        ?.paths,
    );
    expect(adapters.size).toBeGreaterThan(0);
    for (const file of runtimeSourceFiles(webSourceDirectory)) {
      if (rawTransportReferences(readFileSync(file, "utf8"))) {
        expect(
          adapters.has(
            `apps/web/src/${path.relative(webSourceDirectory, file).split(path.sep).join("/")}`,
          ),
          file,
        ).toBe(true);
      }
    }
    for (const source of [
      "fetch (url)",
      "globalThis.fetch(url)",
      'const send = window["fetch"]; send(url)',
      "const {fetch: send} = globalThis; send(url)",
      'const renamed = globalThis; const send = renamed["fetch"]; send(url)',
    ])
      expect(rawTransportReferences(source), source).toBe(true);
    expect(
      rawTransportReferences('// fetch(url)\nconst label = "fetch(url)";'),
    ).toBe(false);
    for (const source of [
      "const view = { fetch: 'a label' }; void view.fetch;",
      "function use(fetch: () => void) { fetch(); }",
      "function use(window: {fetch: () => void}) { window.fetch(); }",
      "const holder = { fetch: () => undefined }; holder.fetch();",
    ])
      expect(rawTransportReferences(source), source).toBe(false);
  });
});
