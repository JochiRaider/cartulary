// @vitest-environment node

import { readdirSync, readFileSync } from "node:fs";
import { basename, extname, join, relative, resolve } from "node:path";
import { describe, expect, it } from "vitest";

const repositoryRoot = resolve(import.meta.dirname, "../../../../..");
const e2eRoot = join(repositoryRoot, "apps/web/e2e");
const supportRoot = join(e2eRoot, "support");
const testUtilsRoot = join(repositoryRoot, "packages/test-utils");

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

describe("web E2E semantic support policy", () => {
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

  it("removes legacy root support implementations", () => {
    const forbiddenRootNames = new Set([
      "a11yPhaseMap.ts",
      "authRuntime.ts",
      "evidenceFixtureHelpers.ts",
      "harnessState.ts",
      "helpers.ts",
      "authenticationPage.ts",
      "entity_linkingHelpers.ts",
      "collaborationHarness.ts",
      "network_flowNetworkFlowHarness.ts",
      "sessionSupport.ts",
      "visualFixtureHelpers.ts",
      "workbookRequestSupport.ts",
    ]);
    const present = sourceFiles(e2eRoot)
      .filter((path) => relative(e2eRoot, path).split("/").length === 1)
      .map((path) => basename(path))
      .filter((name) => forbiddenRootNames.has(name));

    expect(present).toEqual([]);
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
