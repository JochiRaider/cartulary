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
  let found = false;
  const visit = (node: ts.Node) => {
    if (ts.isIdentifier(node) && node.text === "fetch") found = true;
    if (
      ts.isElementAccessExpression(node) &&
      ts.isStringLiteral(node.argumentExpression) &&
      node.argumentExpression.text === "fetch"
    )
      found = true;
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
    ])
      expect(rawTransportReferences(source), source).toBe(true);
    expect(
      rawTransportReferences('// fetch(url)\nconst label = "fetch(url)";'),
    ).toBe(false);
  });
});
