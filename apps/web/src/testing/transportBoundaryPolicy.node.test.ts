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

type TransportKind = "http" | "socket";

const transportGlobals = new Map<string, TransportKind>([
  ["fetch", "http"],
  ["XMLHttpRequest", "http"],
  ["EventSource", "http"],
  ["WebSocket", "socket"],
]);

function rawTransportReferences(source: string): Set<TransportKind> {
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
  const navigatorReference = (value: ts.Expression): boolean =>
    (ts.isIdentifier(value) && value.text === "navigator" && ambient(value)) ||
    (ts.isPropertyAccessExpression(value) &&
      value.name.text === "navigator" &&
      browserGlobal(value.expression));
  const typePosition = (node: ts.Node) =>
    ts.isTypeReferenceNode(node.parent) ||
    ts.isQualifiedName(node.parent) ||
    ts.isExpressionWithTypeArguments(node.parent);
  const found = new Set<TransportKind>();
  const visit = (node: ts.Node) => {
    const identifierKind = ts.isIdentifier(node)
      ? transportGlobals.get(node.text)
      : undefined;
    if (
      identifierKind !== undefined &&
      !(
        ts.isPropertyAccessExpression(node.parent) && node.parent.name === node
      ) &&
      !typePosition(node) &&
      ambient(node as ts.Identifier)
    )
      found.add(identifierKind);
    if (
      ts.isPropertyAccessExpression(node) ||
      ts.isElementAccessExpression(node)
    ) {
      const name = ts.isPropertyAccessExpression(node)
        ? node.name.text
        : ts.isStringLiteral(node.argumentExpression)
          ? node.argumentExpression.text
          : undefined;
      const memberKind =
        name === undefined ? undefined : transportGlobals.get(name);
      if (memberKind !== undefined && browserGlobal(node.expression))
        found.add(memberKind);
      if (name === "sendBeacon" && navigatorReference(node.expression))
        found.add("http");
    }
    if (
      ts.isBindingElement(node) &&
      ts.isObjectBindingPattern(node.parent) &&
      ts.isVariableDeclaration(node.parent.parent) &&
      node.parent.parent.initializer
    ) {
      const property = node.propertyName ?? node.name;
      const bindingKind =
        ts.isIdentifier(property) || ts.isStringLiteral(property)
          ? transportGlobals.get(property.text)
          : undefined;
      if (
        bindingKind !== undefined &&
        browserGlobal(node.parent.parent.initializer)
      )
        found.add(bindingKind);
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
    const ownerPaths = (ownerID: string) =>
      new Set(
        manifest.entries.find((entry) => entry.owner_id === ownerID)?.paths,
      );
    const allowed: Record<TransportKind, Set<string>> = {
      http: ownerPaths("web.services"),
      socket: ownerPaths("web.collaboration"),
    };
    expect(allowed.http.size).toBeGreaterThan(0);
    expect(allowed.socket.size).toBeGreaterThan(0);
    for (const file of runtimeSourceFiles(webSourceDirectory)) {
      const relative = `apps/web/src/${path.relative(webSourceDirectory, file).split(path.sep).join("/")}`;
      for (const kind of rawTransportReferences(readFileSync(file, "utf8")))
        expect(allowed[kind].has(relative), `${kind} ${file}`).toBe(true);
    }
    for (const [source, kind] of [
      ["fetch (url)", "http"],
      ["globalThis.fetch(url)", "http"],
      ['const send = window["fetch"]; send(url)', "http"],
      ["const {fetch: send} = globalThis; send(url)", "http"],
      [
        'const renamed = globalThis; const send = renamed["fetch"]; send(url)',
        "http",
      ],
      ["new XMLHttpRequest()", "http"],
      ["new window.EventSource(url)", "http"],
      ["const {EventSource: Stream} = globalThis; new Stream(url)", "http"],
      ["navigator.sendBeacon(url, body)", "http"],
      ["self.navigator.sendBeacon(url, body)", "http"],
      ["new WebSocket(url)", "socket"],
      ['new globalThis["WebSocket"](url)', "socket"],
    ] as const)
      expect([...rawTransportReferences(source)], source).toEqual([kind]);
    for (const source of [
      '// fetch(url)\nconst label = "fetch(url)";',
      "const view = { fetch: 'a label' }; void view.fetch;",
      "function use(fetch: () => void) { fetch(); }",
      "function use(window: {fetch: () => void}) { window.fetch(); }",
      "const holder = { fetch: () => undefined }; holder.fetch();",
      "type Socket = WebSocket; let socket: Socket | null = null;",
      "class WebSocket {} new WebSocket();",
      "function use(navigator: {sendBeacon: () => void}) { navigator.sendBeacon(); }",
    ])
      expect([...rawTransportReferences(source)], source).toEqual([]);
  });
});
