// @vitest-environment node

import { readdirSync, readFileSync } from "node:fs";
import { basename, dirname, extname, join, relative, resolve } from "node:path";
import { listViewContracts } from "@cartulary/view-contracts";
import * as ts from "typescript";
import { describe, expect, it } from "vitest";

const repositoryRoot = resolve(import.meta.dirname, "../../../../..");
const e2eRoot = join(repositoryRoot, "apps/web/e2e");
const supportRoot = join(e2eRoot, "support");
const testUtilsRoot = join(repositoryRoot, "packages/test-utils");
const currentViewSchemaIds = new Set(
  listViewContracts().map((contract) => contract.viewSchemaId),
);
const rootEntrypointNames = new Set([
  "fixtures.ts",
  "fixtures.test.ts",
  "global-setup.ts",
  "global-teardown.ts",
]);

type E2ESourceGraphResult = {
  readonly ambiguousLocalImports: readonly string[];
  readonly rootViolations: readonly string[];
  readonly unreachableModules: readonly string[];
  readonly unusedExports: readonly string[];
};

function rawPublicJsonUsageViolations(
  sources: ReadonlyMap<string, string>,
  root: string,
): readonly string[] {
  const sourcePaths = new Set(sources.keys());
  const violations: string[] = [];
  for (const [path, source] of sources) {
    const localPath = relative(root, path);
    if (localPath.startsWith("support/transport/")) {
      continue;
    }
    const sourceFile = sourceFileFor(path, source);
    const localBindings = new Set<string>();
    for (const statement of sourceFile.statements) {
      if (
        !ts.isImportDeclaration(statement) ||
        !ts.isStringLiteral(statement.moduleSpecifier)
      ) {
        continue;
      }
      const target = resolveLocalModule(
        path,
        statement.moduleSpecifier.text,
        sourcePaths,
      );
      if (
        !target ||
        relative(root, target) !== "support/transport/publicJsonClient.ts"
      ) {
        continue;
      }
      const bindings = statement.importClause?.namedBindings;
      if (!bindings || !ts.isNamedImports(bindings)) {
        continue;
      }
      for (const element of bindings.elements) {
        if (
          (element.propertyName ?? element.name).text === "requestPublicJson"
        ) {
          localBindings.add(element.name.text);
          violations.push(`${localPath}: import requestPublicJson`);
        }
      }
    }
    const visit = (node: ts.Node) => {
      if (
        ts.isCallExpression(node) &&
        ts.isIdentifier(node.expression) &&
        localBindings.has(node.expression.text)
      ) {
        violations.push(`${localPath}: call requestPublicJson`);
      }
      ts.forEachChild(node, visit);
    };
    ts.forEachChild(sourceFile, visit);
  }
  return violations.sort();
}

function sourceFiles(root: string): string[] {
  return readdirSync(root, { withFileTypes: true }).flatMap((entry) => {
    const path = join(root, entry.name);
    if (entry.isDirectory()) {
      return sourceFiles(path);
    }
    return [".ts", ".tsx"].includes(extname(entry.name)) ? [path] : [];
  });
}

function relativePaths(paths: readonly string[]): readonly string[] {
  return paths.map((path) => relative(repositoryRoot, path));
}

function hasModifier(node: ts.Node, kind: ts.SyntaxKind): boolean {
  return ts.canHaveModifiers(node)
    ? (ts.getModifiers(node)?.some((modifier) => modifier.kind === kind) ??
        false)
    : false;
}

function bindingNames(name: ts.BindingName): string[] {
  if (ts.isIdentifier(name)) {
    return [name.text];
  }
  return name.elements.flatMap((element) =>
    ts.isOmittedExpression(element) ? [] : bindingNames(element.name),
  );
}

function exportedNames(sourceFile: ts.SourceFile): ReadonlySet<string> {
  const names = new Set<string>();
  for (const statement of sourceFile.statements) {
    if (ts.isExportAssignment(statement)) {
      names.add("default");
      continue;
    }
    if (ts.isExportDeclaration(statement)) {
      if (statement.exportClause && ts.isNamedExports(statement.exportClause)) {
        for (const element of statement.exportClause.elements) {
          names.add(element.name.text);
        }
      }
      continue;
    }
    if (!hasModifier(statement, ts.SyntaxKind.ExportKeyword)) {
      continue;
    }
    if (hasModifier(statement, ts.SyntaxKind.DefaultKeyword)) {
      names.add("default");
      continue;
    }
    if (ts.isVariableStatement(statement)) {
      for (const declaration of statement.declarationList.declarations) {
        for (const name of bindingNames(declaration.name)) {
          names.add(name);
        }
      }
      continue;
    }
    if (
      (ts.isFunctionDeclaration(statement) ||
        ts.isClassDeclaration(statement) ||
        ts.isInterfaceDeclaration(statement) ||
        ts.isTypeAliasDeclaration(statement) ||
        ts.isEnumDeclaration(statement) ||
        ts.isModuleDeclaration(statement)) &&
      statement.name
    ) {
      names.add(statement.name.text);
    }
  }
  return names;
}

function sourceFileFor(path: string, source: string): ts.SourceFile {
  return ts.createSourceFile(
    path,
    source,
    ts.ScriptTarget.Latest,
    true,
    path.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
  );
}

function resolveLocalModule(
  importer: string,
  specifier: string,
  sourcePaths: ReadonlySet<string>,
): string | null {
  if (!specifier.startsWith(".")) {
    return null;
  }
  const base = resolve(dirname(importer), specifier);
  const candidates = extname(base)
    ? [base]
    : [base, `${base}.ts`, `${base}.tsx`, join(base, "index.ts")];
  return candidates.find((candidate) => sourcePaths.has(candidate)) ?? null;
}

function isSupportOrPage(path: string, root: string): boolean {
  const localPath = relative(root, path);
  return localPath.startsWith("support/") || localPath.startsWith("pages/");
}

function isE2EEntrypoint(path: string, root: string): boolean {
  const name = basename(path);
  return (
    name.endsWith(".spec.ts") ||
    name.endsWith(".test.ts") ||
    (dirname(path) === root && rootEntrypointNames.has(name))
  );
}

function analyzeE2ESourceGraph(
  sources: ReadonlyMap<string, string>,
  root: string,
): E2ESourceGraphResult {
  const sourcePaths = new Set(sources.keys());
  const sourceFiles = new Map(
    Array.from(sources, ([path, source]) => [
      path,
      sourceFileFor(path, source),
    ]),
  );
  const edges = new Map<string, Set<string>>();
  const importedNames = new Map<string, Set<string>>();
  const ambiguousLocalImports: string[] = [];
  const addEdge = (from: string, to: string) => {
    const targets = edges.get(from) ?? new Set<string>();
    targets.add(to);
    edges.set(from, targets);
  };
  const addImportedName = (target: string, name: string) => {
    const names = importedNames.get(target) ?? new Set<string>();
    names.add(name);
    importedNames.set(target, names);
  };

  for (const [path, sourceFile] of sourceFiles) {
    for (const statement of sourceFile.statements) {
      if (ts.isImportDeclaration(statement)) {
        const specifier = ts.isStringLiteral(statement.moduleSpecifier)
          ? statement.moduleSpecifier.text
          : "";
        const target = resolveLocalModule(path, specifier, sourcePaths);
        if (!target) {
          continue;
        }
        addEdge(path, target);
        const importClause = statement.importClause;
        if (!importClause) {
          continue;
        }
        if (importClause.name) {
          addImportedName(target, "default");
        }
        if (importClause.namedBindings) {
          if (ts.isNamespaceImport(importClause.namedBindings)) {
            ambiguousLocalImports.push(
              `${relative(root, path)}: namespace import ${specifier}`,
            );
          } else {
            for (const element of importClause.namedBindings.elements) {
              addImportedName(
                target,
                (element.propertyName ?? element.name).text,
              );
            }
          }
        }
      }
      if (
        ts.isExportDeclaration(statement) &&
        statement.moduleSpecifier &&
        ts.isStringLiteral(statement.moduleSpecifier)
      ) {
        const specifier = statement.moduleSpecifier.text;
        const target = resolveLocalModule(path, specifier, sourcePaths);
        if (!target) {
          continue;
        }
        addEdge(path, target);
        if (
          !statement.exportClause ||
          ts.isNamespaceExport(statement.exportClause)
        ) {
          ambiguousLocalImports.push(
            `${relative(root, path)}: wildcard re-export ${specifier}`,
          );
        } else {
          for (const element of statement.exportClause.elements) {
            addImportedName(
              target,
              (element.propertyName ?? element.name).text,
            );
          }
        }
      }
    }

    const visit = (node: ts.Node) => {
      if (
        ts.isCallExpression(node) &&
        node.expression.kind === ts.SyntaxKind.ImportKeyword
      ) {
        const [argument] = node.arguments;
        if (!argument || !ts.isStringLiteral(argument)) {
          ambiguousLocalImports.push(
            `${relative(root, path)}: non-literal dynamic import`,
          );
        } else {
          const target = resolveLocalModule(path, argument.text, sourcePaths);
          if (target) {
            addEdge(path, target);
            let consumer: ts.Node = node;
            while (
              consumer.parent &&
              (ts.isAwaitExpression(consumer.parent) ||
                ts.isParenthesizedExpression(consumer.parent))
            ) {
              consumer = consumer.parent;
            }
            const parent = consumer.parent;
            if (
              parent &&
              ts.isPropertyAccessExpression(parent) &&
              parent.expression === consumer
            ) {
              addImportedName(target, parent.name.text);
            } else if (
              parent &&
              ts.isVariableDeclaration(parent) &&
              parent.initializer === consumer &&
              ts.isObjectBindingPattern(parent.name)
            ) {
              for (const element of parent.name.elements) {
                addImportedName(
                  target,
                  (element.propertyName ?? element.name).getText(sourceFile),
                );
              }
            } else if (
              parent &&
              ts.isVariableDeclaration(parent) &&
              parent.initializer === consumer &&
              ts.isIdentifier(parent.name)
            ) {
              const binding = parent.name;
              let scope: ts.Node = parent;
              while (
                scope.parent &&
                !ts.isBlock(scope.parent) &&
                !ts.isSourceFile(scope.parent)
              ) {
                scope = scope.parent;
              }
              const usageScope = scope.parent ?? sourceFile;
              let precise = true;
              const inspectUsage = (candidate: ts.Node) => {
                if (
                  ts.isIdentifier(candidate) &&
                  candidate.text === binding.text &&
                  candidate !== binding
                ) {
                  if (
                    ts.isPropertyAccessExpression(candidate.parent) &&
                    candidate.parent.expression === candidate
                  ) {
                    addImportedName(target, candidate.parent.name.text);
                  } else {
                    precise = false;
                  }
                }
                ts.forEachChild(candidate, inspectUsage);
              };
              ts.forEachChild(usageScope, inspectUsage);
              if (!precise) {
                ambiguousLocalImports.push(
                  `${relative(root, path)}: imprecise dynamic import ${argument.text}`,
                );
              }
            } else {
              ambiguousLocalImports.push(
                `${relative(root, path)}: unobserved dynamic import ${argument.text}`,
              );
            }
          }
        }
      }
      ts.forEachChild(node, visit);
    };
    ts.forEachChild(sourceFile, visit);
  }

  const reachable = new Set<string>();
  const pending = Array.from(sourcePaths).filter((path) =>
    isE2EEntrypoint(path, root),
  );
  while (pending.length > 0) {
    const path = pending.pop();
    if (!path || reachable.has(path)) {
      continue;
    }
    reachable.add(path);
    for (const target of edges.get(path) ?? []) {
      pending.push(target);
    }
  }

  const rootViolations = Array.from(sourcePaths)
    .filter((path) => dirname(path) === root)
    .filter((path) => {
      const name = basename(path);
      return !name.endsWith(".spec.ts") && !rootEntrypointNames.has(name);
    })
    .map((path) => relative(root, path))
    .sort();
  const unreachableModules = Array.from(sourcePaths)
    .filter((path) => isSupportOrPage(path, root) && !reachable.has(path))
    .map((path) => relative(root, path))
    .sort();
  const unusedExports = Array.from(sourceFiles).flatMap(
    ([path, sourceFile]) => {
      if (!isSupportOrPage(path, root)) {
        return [];
      }
      const used = importedNames.get(path) ?? new Set<string>();
      return Array.from(exportedNames(sourceFile))
        .filter((name) => !used.has(name))
        .map((name) => `${relative(root, path)}: ${name}`);
    },
  );

  return {
    ambiguousLocalImports: ambiguousLocalImports.sort(),
    rootViolations,
    unreachableModules,
    unusedExports: unusedExports.sort(),
  };
}

function currentViewSchemaLiteralBindings(
  source: string,
  sourceLabel: string,
): string[] {
  const pattern =
    /\b(?:const|let|var)\s+[A-Za-z_$][\w$]*\s*(?::[^=;\n]+)?=\s*["'](cartulary\.view\.[^"']+)["']/gu;
  return Array.from(source.matchAll(pattern)).flatMap((match) => {
    const viewSchemaId = match[1] ?? "";
    return currentViewSchemaIds.has(viewSchemaId)
      ? [`${sourceLabel}: ${viewSchemaId}`]
      : [];
  });
}

function rawGridVendorSelectors(source: string, sourceLabel: string): string[] {
  const vendorSelectorPattern = new RegExp(
    `${String.raw`\.`}${"rdg-"}[A-Za-z0-9_-]+`,
    "gu",
  );
  return Array.from(
    source.matchAll(vendorSelectorPattern),
    (match) => `${sourceLabel}: ${match[0]}`,
  );
}

describe("web E2E semantic support policy", () => {
  it("retains deliberate invalid view-schema probes as explicit negative coverage", () => {
    const source = readFileSync(join(e2eRoot, "workbook.spec.ts"), "utf8");

    expect(source).toContain('"?view_schema_id=cartulary.view.unknown.v1"');
    expect(source).toContain("invalidExplicitStartup.selected_view_schema_id");
  });

  it("requires current view-schema identities to come from the view-contract facade", () => {
    expect(
      currentViewSchemaLiteralBindings(
        'const localTimeline = "cartulary.view.timeline.v2";',
        "seed.ts",
      ),
    ).toEqual(["seed.ts: cartulary.view.timeline.v2"]);

    const violations = sourceFiles(e2eRoot)
      .filter((path) => !path.endsWith("architecturePolicy.test.ts"))
      .flatMap((path) =>
        currentViewSchemaLiteralBindings(
          readFileSync(path, "utf8"),
          relative(repositoryRoot, path),
        ),
      );
    expect(violations).toEqual([]);
  });

  it("rejects raw grid-vendor selectors outside the grid adapter", () => {
    const seededVendorSelector = [".", "rdg-cell-drag-handle"].join("");
    expect(
      rawGridVendorSelectors(
        `page.locator("${seededVendorSelector}")`,
        "seed.ts",
      ),
    ).toEqual([`seed.ts: ${seededVendorSelector}`]);

    const violations = sourceFiles(e2eRoot)
      .filter((path) => !path.endsWith("architecturePolicy.test.ts"))
      .flatMap((path) =>
        rawGridVendorSelectors(
          readFileSync(path, "utf8"),
          relative(repositoryRoot, path),
        ),
      );
    expect(violations).toEqual([]);
  });

  it("keeps phase and catch-all names out of semantic support", () => {
    const forbidden = sourceFiles(supportRoot).filter((path) => {
      const name = basename(path);
      return (
        /^(?:common|helpers|misc|utils)\.(?:ts|tsx)$/u.test(name) ||
        /^phase[0-9]/u.test(name) ||
        relative(supportRoot, path) === "index.ts"
      );
    });

    expect(relativePaths(forbidden)).toEqual([]);
  });

  it("enforces the closed E2E root and statically reachable support modules", () => {
    const seededRoot = "/repo/apps/web/e2e";
    const seeded = analyzeE2ESourceGraph(
      new Map([
        [`${seededRoot}/workbook.spec.ts`, 'import "./support/live";'],
        [`${seededRoot}/legacyShim.ts`, "export const shim = true;"],
        [`${seededRoot}/support/live.ts`, "export {};"],
        [`${seededRoot}/support/unreachable.ts`, "export {};"],
      ]),
      seededRoot,
    );
    expect(seeded.rootViolations).toEqual(["legacyShim.ts"]);
    expect(seeded.unreachableModules).toEqual(["support/unreachable.ts"]);

    const sources = new Map(
      sourceFiles(e2eRoot).map((path) => [path, readFileSync(path, "utf8")]),
    );
    const result = analyzeE2ESourceGraph(sources, e2eRoot);
    expect(result.rootViolations).toEqual([]);
    expect(result.unreachableModules).toEqual([]);
  });

  it("requires every support and page export to have an external importer", () => {
    const seededRoot = "/repo/apps/web/e2e";
    const seeded = analyzeE2ESourceGraph(
      new Map([
        [
          `${seededRoot}/workbook.spec.ts`,
          'import { used } from "./support/client"; void used;',
        ],
        [
          `${seededRoot}/support/client.ts`,
          "export const used = true; export const unnecessary = true;",
        ],
      ]),
      seededRoot,
    );
    expect(seeded.unusedExports).toEqual(["support/client.ts: unnecessary"]);

    const sources = new Map(
      sourceFiles(e2eRoot).map((path) => [path, readFileSync(path, "utf8")]),
    );
    const unusedExports = analyzeE2ESourceGraph(sources, e2eRoot).unusedExports;
    if (unusedExports.length > 0) {
      throw new Error(`unused E2E exports:\n${unusedExports.join("\n")}`);
    }
  });

  it("rejects ambiguous local namespace and dynamic imports", () => {
    const seededRoot = "/repo/apps/web/e2e";
    const seeded = analyzeE2ESourceGraph(
      new Map([
        [
          `${seededRoot}/workbook.spec.ts`,
          'import * as client from "./support/client"; const loaded = await import("./support/client"); consume(loaded); void client;',
        ],
        [`${seededRoot}/support/client.ts`, "export const used = true;"],
      ]),
      seededRoot,
    );
    expect(seeded.ambiguousLocalImports).toEqual([
      "workbook.spec.ts: imprecise dynamic import ./support/client",
      "workbook.spec.ts: namespace import ./support/client",
    ]);

    const sources = new Map(
      sourceFiles(e2eRoot).map((path) => [path, readFileSync(path, "utf8")]),
    );
    expect(
      analyzeE2ESourceGraph(sources, e2eRoot).ambiguousLocalImports,
    ).toEqual([]);
  });

  it("keeps requestPublicJson private to the transport boundary", () => {
    const seededRoot = "/repo/apps/web/e2e";
    const seeded = rawPublicJsonUsageViolations(
      new Map([
        [
          `${seededRoot}/workbook.spec.ts`,
          'import { requestPublicJson as raw } from "./support/transport/publicJsonClient"; void raw({});',
        ],
        [
          `${seededRoot}/support/transport/publicJsonClient.ts`,
          "export function requestPublicJson() {}",
        ],
      ]),
      seededRoot,
    );
    expect(seeded).toEqual([
      "workbook.spec.ts: call requestPublicJson",
      "workbook.spec.ts: import requestPublicJson",
    ]);

    const sources = new Map(
      sourceFiles(e2eRoot).map((path) => [path, readFileSync(path, "utf8")]),
    );
    expect(rawPublicJsonUsageViolations(sources, e2eRoot)).toEqual([]);
  });

  it("keeps application routes and payload semantics out of test-utils", () => {
    const violations = sourceFiles(testUtilsRoot).flatMap((path) => {
      const source = readFileSync(path, "utf8");
      return /\/api\/v1\/|X-Cartulary-Test-Route-Token|CARTULARY_TEST_ROUTE_TOKEN/u.test(
        source,
      )
        ? [relative(repositoryRoot, path)]
        : [];
    });

    expect(violations).toEqual([]);
  });

  it("keeps E2E consumers on contract and package facades", () => {
    const violations = sourceFiles(e2eRoot)
      .filter((path) => !path.endsWith("architecturePolicy.test.ts"))
      .flatMap((path) => {
        const source = readFileSync(path, "utf8");
        const reasons = [
          source.includes("src/workbook/models/workbookSurfaceRegistry")
            ? "app workbook registry"
            : null,
          source.includes('from "@cartulary/test-utils"')
            ? "root test-utils export"
            : null,
          source.includes('from "@cartulary/test-utils/accessibility"') ||
          source.includes('from "@cartulary/test-utils/visual"')
            ? "removed test-utils alias"
            : null,
          source.includes("react-data-grid") ? "grid vendor" : null,
          source.includes("/src/generated/")
            ? "protected generated root"
            : null,
        ].filter((reason): reason is string => reason !== null);
        return reasons.map(
          (reason) => `${relative(repositoryRoot, path)}: ${reason}`,
        );
      });

    expect(violations).toEqual([]);
  });
});
