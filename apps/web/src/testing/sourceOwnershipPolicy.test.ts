import { lstatSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
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

describe("frontend source ownership policy", () => {
  it("accounts for every live TypeScript path exactly once without Markdown input", () => {
    const manifest = JSON.parse(
      readFileSync(manifestPath, "utf8"),
    ) as SourceOwnershipManifest;
    expect(manifest.schema_id).toBe("cartulary.frontend_source_ownership.v1");
    expect(manifest.source_root).toBe("apps/web/src");
    expect(manifest.included_extensions).toEqual([".ts", ".tsx"]);

    const ownerIds = manifest.entries.map((entry) => entry.owner_id);
    expect(ownerIds).toEqual([...ownerIds].sort());
    expect(new Set(ownerIds).size).toBe(ownerIds.length);

    const accountedPaths = manifest.entries.flatMap((entry) => {
      expect(entry.paths).toEqual([...entry.paths].sort());
      return entry.paths;
    });
    expect(new Set(accountedPaths).size).toBe(accountedPaths.length);
    for (const sourcePath of accountedPaths) {
      expect(sourcePath.startsWith(`${manifest.source_root}/`)).toBe(true);
      expect(
        manifest.included_extensions.includes(path.posix.extname(sourcePath)),
      ).toBe(true);
      expect(path.posix.normalize(sourcePath)).toBe(sourcePath);
      const sourceStats = lstatSync(path.join(repoRoot, sourcePath));
      expect(sourceStats.isFile()).toBe(true);
      expect(sourceStats.isSymbolicLink()).toBe(false);
    }

    expect(accountedPaths.sort()).toEqual(
      liveTypeScriptPaths(path.join(repoRoot, manifest.source_root)),
    );
  });

  it("keeps transaction identity and mutation wire intents behind Workbook command assembly", () => {
    const workbookRoot = path.join(repoRoot, "apps/web/src/workbook");
    const productionPaths = liveTypeScriptPaths(workbookRoot).filter(
      (sourcePath) =>
        !sourcePath.endsWith(".test.ts") && !sourcePath.endsWith(".test.tsx"),
    );
    const transactionIdCallers = productionPaths.filter((sourcePath) =>
      /\bclientTxnID\b/u.test(
        readFileSync(path.join(repoRoot, sourcePath), "utf8"),
      ),
    );
    expect(transactionIdCallers).toEqual([
      "apps/web/src/workbook/mutations/secureTransactionId.ts",
    ]);

    const presentationRoots = [
      path.join(workbookRoot, "components"),
      path.join(workbookRoot, "features"),
    ];
    const presentationWireIntents = presentationRoots
      .flatMap(liveTypeScriptPaths)
      .filter((sourcePath) => {
        const source = readFileSync(path.join(repoRoot, sourcePath), "utf8");
        return source.includes("client_txn_id");
      });
    expect(presentationWireIntents).toEqual([]);
  });

  it("keeps authorization recovery and collaboration transport dependencies below Workbook reconciliation", () => {
    const workbookProduction = liveTypeScriptPaths(
      path.join(repoRoot, "apps/web/src/workbook"),
    ).filter(
      (sourcePath) =>
        !sourcePath.endsWith(".test.ts") && !sourcePath.endsWith(".test.tsx"),
    );
    const sources = workbookProduction.map((sourcePath) => ({
      sourcePath,
      source: readFileSync(path.join(repoRoot, sourcePath), "utf8"),
    }));
    expect(
      sources
        .filter(({ source }) => source.includes("CollaborationSheetRef"))
        .map(({ sourcePath }) => sourcePath),
    ).toEqual([]);
    expect(
      sources
        .filter(({ source }) => source.includes("/api/v1/auth/session"))
        .map(({ sourcePath }) => sourcePath),
    ).toEqual([]);

    const coordinatorSource = readFileSync(
      path.join(
        repoRoot,
        "apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts",
      ),
      "utf8",
    );
    expect(coordinatorSource).not.toMatch(
      /from\s+["'][^"']*services\/browserApi["']/u,
    );
  });
});
