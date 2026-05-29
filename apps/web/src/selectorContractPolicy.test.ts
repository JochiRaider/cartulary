import { readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const thisFile = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(thisFile), "../../..");

const scannedRoots = [
  "apps/web/src",
  "apps/web/e2e",
  "packages/test-utils/src",
] as const;

const ignoredFilePatterns = [
  /\.d\.ts$/u,
  /selectorContractPolicy\.test\.ts$/u,
] as const;

const ignoredPathSegments = new Set(["node_modules", "dist", "coverage"]);

const forbiddenPatterns = [
  {
    message: "dynamic data-testid template",
    pattern: /data-testid=\{`[^`]*\$\{[^`]*`\}/u,
  },
  {
    message: "dynamic getByTestId template",
    pattern: /getByTestId\(\s*`[^`]*\$\{[^`]*`/u,
  },
  {
    message: "raw dynamic data-testid selector",
    pattern: /\[data-testid(?:[*^$|~]?=)"?\$\{/u,
  },
  {
    message: "raw data-testid prefix or suffix selector",
    pattern: /\[data-testid(?:[*^$|~]=)["'][^"']+["']\]/u,
  },
  {
    message: "shared selector literal outside ui-contracts builder",
    pattern:
      /\b(?:getByTestId|findByTestId|queryByTestId)\(\s*["'](?:save-state|reference-pack-(?:admin-panel|file|import|job-status|reload|cancel|refresh-all|refresh-selected|row|error))["']\s*\)/u,
  },
  {
    message: "shared data-testid literal outside ui-contracts builder",
    pattern:
      /data-testid=["'](?:save-state|reference-pack-(?:admin-panel|file|import|job-status|reload|cancel|refresh-all|refresh-selected|row|error))["']/u,
  },
  {
    message: "Timeline workbook shell readiness text",
    pattern: /getByText\(\s*["']Timeline workbook shell["']/u,
  },
  {
    message: "Timeline mutation substrate readiness text",
    pattern: /getByText\(\s*["']Timeline mutation substrate["']/u,
  },
  {
    message: "Current incident role readiness text",
    pattern: /getByText\(\s*["']Current incident role:/u,
  },
  {
    message: "heading-name readiness wait",
    pattern: /getByRole\(\s*["']heading["']\s*,\s*\{\s*name:/u,
  },
] as const;

function listSourceFiles(relativeRoot: string): string[] {
  const absoluteRoot = path.join(repoRoot, relativeRoot);
  const entries = readdirSync(absoluteRoot, { withFileTypes: true });
  const files: string[] = [];

  for (const entry of entries) {
    const absolutePath = path.join(absoluteRoot, entry.name);
    const relativePath = path.relative(repoRoot, absolutePath);
    if (entry.isDirectory()) {
      if (
        !ignoredPathSegments.has(entry.name) &&
        !entry.name.endsWith("-snapshots")
      ) {
        files.push(...listSourceFiles(relativePath));
      }
      continue;
    }
    if (!entry.isFile()) {
      continue;
    }
    if (!/\.[cm]?[tj]sx?$/u.test(entry.name)) {
      continue;
    }
    if (shouldIgnoreFile(entry.name)) {
      continue;
    }
    files.push(relativePath);
  }

  return files;
}

function shouldIgnoreFile(fileName: string): boolean {
  return ignoredFilePatterns.some((pattern) => pattern.test(fileName));
}

function lineNumberForOffset(content: string, offset: number): number {
  return content.slice(0, offset).split("\n").length;
}

describe("selector contract policy", () => {
  it("keeps cross-boundary selector templates behind shared builders", () => {
    const violations: string[] = [];

    for (const file of scannedRoots.flatMap(listSourceFiles)) {
      const absolutePath = path.join(repoRoot, file);
      if (!statSync(absolutePath).isFile()) {
        continue;
      }
      const content = readFileSync(absolutePath, "utf8");
      for (const { message, pattern } of forbiddenPatterns) {
        const flags = pattern.flags.includes("g")
          ? pattern.flags
          : `${pattern.flags}g`;
        const globalPattern = new RegExp(pattern.source, flags);
        for (const match of content.matchAll(globalPattern)) {
          if (match.index === undefined) {
            continue;
          }
          violations.push(
            `${file}:${lineNumberForOffset(content, match.index)} ${message}`,
          );
        }
      }
    }

    expect(violations).toEqual([]);
  });
});
