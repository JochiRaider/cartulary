import { lstatSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import * as ts from "typescript";
import { describe, expect, it } from "vitest";

type SourceOwnershipManifest = {
  readonly schema_id: string;
  readonly source_root: string;
  readonly included_extensions: readonly string[];
  readonly entries: readonly {
    readonly owner_id: string;
    readonly paths: readonly string[];
  }[];
};

const thisFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(thisFile), "../../../..");
const manifestPath = path.join(
  repoRoot,
  "tools/frontend_source_ownership.json",
);

function normalizedRepoPath(absolutePath: string): string {
  return path.relative(repoRoot, absolutePath).split(path.sep).join("/");
}

function liveTypeScriptPaths(directory: string): string[] {
  const paths: string[] = [];
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const absolutePath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      paths.push(...liveTypeScriptPaths(absolutePath));
      continue;
    }
    if (
      entry.isFile() &&
      (entry.name.endsWith(".ts") || entry.name.endsWith(".tsx"))
    ) {
      paths.push(normalizedRepoPath(absolutePath));
    }
  }
  return paths.sort();
}

function sourceOwnershipProblems(
  manifest: SourceOwnershipManifest,
  livePaths: readonly string[],
): string[] {
  const problems: string[] = [];
  const owners = new Set<string>();
  const declared = new Map<string, string[]>();
  const live = new Set(livePaths);
  const ownerIds = manifest.entries.map((entry) => entry.owner_id);
  if (ownerIds.join("\n") !== [...ownerIds].sort().join("\n"))
    problems.push("Unsorted owner IDs");
  for (const entry of manifest.entries) {
    if (owners.has(entry.owner_id))
      problems.push(`Duplicate owner: ${entry.owner_id}`);
    owners.add(entry.owner_id);
    if (entry.paths.join("\n") !== [...entry.paths].sort().join("\n"))
      problems.push(`Unsorted paths: ${entry.owner_id}`);
    for (const sourcePath of entry.paths) {
      if (
        !sourcePath.startsWith(`${manifest.source_root}/`) ||
        sourcePath.includes("\\") ||
        path.posix.normalize(sourcePath) !== sourcePath ||
        !manifest.included_extensions.includes(path.posix.extname(sourcePath))
      )
        problems.push(`Unsafe path: ${sourcePath}`);
      const pathOwners = declared.get(sourcePath) ?? [];
      pathOwners.push(entry.owner_id);
      declared.set(sourcePath, pathOwners);
    }
  }
  for (const [sourcePath, pathOwners] of declared) {
    if (pathOwners.length > 1)
      problems.push(`Duplicate path: ${sourcePath} (${pathOwners.join(", ")})`);
    if (!live.has(sourcePath)) problems.push(`Stale path: ${sourcePath}`);
  }
  for (const sourcePath of live)
    if (!declared.has(sourcePath)) problems.push(`Missing path: ${sourcePath}`);
  return problems.sort();
}

function commandOwnershipProblems(file: string, source: string): string[] {
  const syntax = ts.createSourceFile(
    file,
    source,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  );
  const problems = new Set<string>();
  const inspect = (node: ts.Node) => {
    if (
      ((ts.isIdentifier(node) && node.text === "clientTxnID") ||
        (ts.isElementAccessExpression(node) &&
          ts.isStringLiteral(node.argumentExpression) &&
          node.argumentExpression.text === "clientTxnID")) &&
      !file.includes("/workbook/mutations/")
    )
      problems.add("transaction identity outside mutation owner");
    if (
      ts.isPropertyAssignment(node) ||
      ts.isPropertySignature(node) ||
      ts.isShorthandPropertyAssignment(node)
    ) {
      const name = node.name;
      const key = ts.isComputedPropertyName(name) ? name.expression : name;
      if (
        (ts.isIdentifier(key) || ts.isStringLiteral(key)) &&
        key.text === "client_txn_id" &&
        file.includes("/workbook/components/")
      )
        problems.add("wire intent in presentation");
    }
    ts.forEachChild(node, inspect);
  };
  inspect(syntax);
  return [...problems].sort();
}

describe("frontend source ownership policy", () => {
  it("accounts for every live TypeScript path exactly once without Markdown input", () => {
    const manifest = JSON.parse(
      readFileSync(manifestPath, "utf8"),
    ) as SourceOwnershipManifest;
    expect(manifest.schema_id).toBe("cartulary.frontend_source_ownership.v1");
    expect(manifest.source_root).toBe("apps/web/src");
    expect(manifest.included_extensions).toEqual([".ts", ".tsx"]);

    expect(
      sourceOwnershipProblems(
        manifest,
        liveTypeScriptPaths(path.join(repoRoot, manifest.source_root)),
      ),
    ).toEqual([]);
    for (const entry of manifest.entries) {
      for (const sourcePath of entry.paths) {
        const sourceStats = lstatSync(path.join(repoRoot, sourcePath));
        expect(sourceStats.isFile(), sourcePath).toBe(true);
        expect(sourceStats.isSymbolicLink(), sourcePath).toBe(false);
      }
    }
  });

  it("reports missing duplicate stale and unsafe paths explicitly", () => {
    const live = ["apps/web/src/example.ts"];
    const fixture = (paths: readonly string[]): SourceOwnershipManifest => ({
      schema_id: "cartulary.frontend_source_ownership.v1",
      source_root: "apps/web/src",
      included_extensions: [".ts", ".tsx"],
      entries: [{ owner_id: "web.example", paths }],
    });
    expect(sourceOwnershipProblems(fixture(live), live)).toEqual([]);
    expect(sourceOwnershipProblems(fixture([]), live)).toEqual([
      "Missing path: apps/web/src/example.ts",
    ]);
    expect(sourceOwnershipProblems(fixture([...live, ...live]), live)).toEqual([
      "Duplicate path: apps/web/src/example.ts (web.example, web.example)",
    ]);
    expect(
      sourceOwnershipProblems(
        fixture([...live, "apps/web/src/removed.ts"]),
        live,
      ),
    ).toEqual(["Stale path: apps/web/src/removed.ts"]);
    for (const unsafe of [
      "/tmp/outside.ts",
      "apps/web/src/../outside.ts",
      "apps/web/src/file.js",
      "apps/web/src/dir\\file.ts",
    ]) {
      expect(sourceOwnershipProblems(fixture([unsafe]), live)).toContain(
        `Unsafe path: ${unsafe}`,
      );
    }
    expect(
      sourceOwnershipProblems(fixture(["apps/web/src/z.ts", ...live]), live),
    ).toContain("Unsorted paths: web.example");
    const duplicateOwner = {
      ...fixture(live),
      entries: [...fixture(live).entries, ...fixture([]).entries],
    };
    expect(sourceOwnershipProblems(duplicateOwner, live)).toContain(
      "Duplicate owner: web.example",
    );
    const crossOwner = {
      ...fixture(live),
      entries: [
        ...fixture(live).entries,
        { owner_id: "web.other", paths: live },
      ],
    };
    expect(sourceOwnershipProblems(crossOwner, live)).toContain(
      "Duplicate path: apps/web/src/example.ts (web.example, web.other)",
    );
    expect(
      sourceOwnershipProblems(
        { ...crossOwner, entries: [...crossOwner.entries].reverse() },
        live,
      ),
    ).toContain("Unsorted owner IDs");
  });

  it("centralizes transaction identity and confines wire intents to owner-local command modules", () => {
    const workbookRoot = path.join(repoRoot, "apps/web/src/workbook");
    const productionPaths = liveTypeScriptPaths(workbookRoot).filter(
      (sourcePath) =>
        !sourcePath.endsWith(".test.ts") && !sourcePath.endsWith(".test.tsx"),
    );
    for (const sourcePath of productionPaths) {
      expect(
        commandOwnershipProblems(
          sourcePath,
          readFileSync(path.join(repoRoot, sourcePath), "utf8"),
        ),
        sourcePath,
      ).toEqual([]);
    }
    // File renames, local aliases, formatting and comments are not authority.
    expect(
      commandOwnershipProblems(
        "apps/web/src/workbook/mutations/renamed.ts",
        'import { clientTxnID as create } from "../../services/browserApi"; void create;',
      ),
    ).toEqual([]);
    expect(
      commandOwnershipProblems(
        "apps/web/src/workbook/features/notes/renamed.ts",
        'const request = { ["client_txn_id"] : id };',
      ),
    ).toEqual([]);
    expect(
      commandOwnershipProblems(
        "apps/web/src/workbook/components/renamed.tsx",
        '// client_txn_id clientTxnID\nconst label = "client_txn_id";',
      ),
    ).toEqual([]);
    expect(
      commandOwnershipProblems(
        "apps/web/src/workbook/components/renamed.tsx",
        'import { clientTxnID as create } from "../../services/browserApi";',
      ),
    ).toEqual(["transaction identity outside mutation owner"]);
    expect(
      commandOwnershipProblems(
        "apps/web/src/workbook/components/renamed.tsx",
        'const request = { ["client_txn_id"] : id };',
      ),
    ).toEqual(["wire intent in presentation"]);
  });

  it("keeps authorization recovery and collaboration transport dependencies below Workbook reconciliation", () => {
    const workbookProduction = liveTypeScriptPaths(
      path.join(repoRoot, "apps/web/src/workbook"),
    ).filter(
      (sourcePath) =>
        !sourcePath.endsWith(".test.ts") && !sourcePath.endsWith(".test.tsx"),
    );
    for (const sourcePath of workbookProduction) {
      const syntax = ts.createSourceFile(
        sourcePath,
        readFileSync(path.join(repoRoot, sourcePath), "utf8"),
        ts.ScriptTarget.Latest,
        true,
        ts.ScriptKind.TSX,
      );
      const violations: string[] = [];
      const visit = (node: ts.Node) => {
        if (ts.isIdentifier(node) && node.text === "CollaborationSheetRef")
          violations.push("transport identity in Workbook");
        if (
          ts.isStringLiteralLike(node) &&
          node.text.includes("/api/v1/auth/session")
        )
          violations.push("session transport in Workbook");
        if (
          ts.isImportDeclaration(node) &&
          ts.isStringLiteral(node.moduleSpecifier) &&
          sourcePath.includes("/workbook/collaboration/") &&
          /(?:^|\/)services\//u.test(node.moduleSpecifier.text)
        )
          violations.push("service import in reconciliation owner");
        ts.forEachChild(node, visit);
      };
      visit(syntax);
      expect(violations, sourcePath).toEqual([]);
    }
  });
});
