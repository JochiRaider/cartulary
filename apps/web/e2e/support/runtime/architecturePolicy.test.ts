// @vitest-environment node

import { readdirSync, readFileSync } from "node:fs";
import { basename, extname, join, relative, resolve } from "node:path";
import { listViewContracts } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";

const repositoryRoot = resolve(import.meta.dirname, "../../../../..");
const e2eRoot = join(repositoryRoot, "apps/web/e2e");
const supportRoot = join(e2eRoot, "support");
const testUtilsRoot = join(repositoryRoot, "packages/test-utils");
const currentViewSchemaIds = new Set(
  listViewContracts().map((contract) => contract.viewSchemaId),
);

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
