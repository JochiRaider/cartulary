import assert from "node:assert/strict";
import { readFileSync, mkdirSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { checkFrontendImportBoundaries } from "../frontend-import-boundary-check.mjs";
import { createFixture, repository } from "./fixtures.mjs";

const owner = JSON.parse(readFileSync(path.join(repository, "tools/frontend_import_boundaries.json"), "utf8"));

function frontendFixture(t) {
  const fixture = createFixture(t);
  for (const name of ["grid-adapter", "protocol-ts", "test-utils", "ui-contracts", "view-contracts"]) {
    fixture.write(`packages/${name}/package.json`, {
      name: `@cartulary/${name}`,
      exports: name === "test-utils" ? { "./grid": "./src/grid.ts" } : {
        ".": "./src/index.ts", ...(name === "grid-adapter" ? { "./test-support": "./src/test-support.tsx" } : {}),
      },
    });
  }
  mkdirSync(path.join(fixture.root, "apps/web/src"), { recursive: true });
  fixture.write("packages/grid-adapter/src/index.tsx", 'import "react-data-grid/lib/styles.css";');
  fixture.write("contracts/design/tokens.v1.json", readFileSync(path.join(repository, "contracts/design/tokens.v1.json"), "utf8"));
  return fixture;
}

test("frontend support boundaries distinguish executable dependencies from comments", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/workbook.spec.ts", '// page.locator(".rdg-cell");\npage.getByRole("grid");');
  fixture.write("packages/test-utils/src/grid.ts", '// /api/v1/ X-Cartulary-Test-Route-Token CARTULARY_TEST_ROUTE_TOKEN\nexport const role = "grid";');
  fixture.write("packages/grid-adapter/src/selectors.ts", 'export const cell = ".rdg-cell";');
  const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

for (const [name, file, source, ruleID] of [
  ["vendor selector", "apps/web/e2e/workbook.spec.ts", 'page.locator(".rdg-cell");', "web-e2e-grid-vendor-selector-boundary"],
  ["vendor template", "apps/web/e2e/workbook.spec.ts", 'page.locator(`.rdg-cell-${column}`);', "web-e2e-grid-vendor-selector-boundary"],
  ["application route", "packages/test-utils/src/grid.ts", 'fetch(`/api/v1/incidents/${id}`);', "frontend-test-utils-route-control-boundary"],
  ["control header", "packages/test-utils/src/grid.ts", 'const headers = { "X-Cartulary-Test-Route-Token": token };', "frontend-test-utils-route-control-boundary"],
  ["control environment", "packages/test-utils/src/grid.ts", 'const token = process.env.CARTULARY_TEST_ROUTE_TOKEN;', "frontend-test-utils-route-control-boundary"],
]) {
  test(`frontend support boundaries reject ${name}`, (t) => {
    const fixture = frontendFixture(t);
    fixture.write(file, source);
    const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
    const matches = result.diagnostics.filter((entry) => entry.ruleID === ruleID);
    assert.equal(matches.length, 1, JSON.stringify(result.diagnostics));
    assert.equal(matches[0].file, file);
    assert.equal(matches[0].line, 1);
  });
}

for (const [name, source] of [
  ["named import", 'import { requestPublicJsonObserved as raw } from "./support/transport/rawPublicJson";'],
  ["namespace import", 'import * as raw from "./support/transport/rawPublicJson";'],
  ["re-export", 'export * from "./support/transport/rawPublicJson";'],
  ["dynamic import", 'void import("./support/transport/rawPublicJson");'],
]) {
  test(`frontend transport boundary rejects raw ${name}`, (t) => {
    const fixture = frontendFixture(t);
    fixture.write("apps/web/e2e/workbook.spec.ts", source);
    const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
    assert.ok(result.diagnostics.some((entry) => entry.ruleID === "web-e2e-raw-json-transport-boundary"), JSON.stringify(result.diagnostics));
  });
}

test("frontend transport boundary admits operation and test consumers without exposing raw execution through context", (t) => {
  const fixture = frontendFixture(t);
  for (const file of ["publicHttpOperationClient.ts", "publicJsonClient.test.ts"]) {
    fixture.write(`apps/web/e2e/support/transport/${file}`, 'import { requestPublicJsonObserved } from "./rawPublicJson";');
  }
  fixture.write("apps/web/e2e/workbook.spec.ts", 'import { atJsonOrigin } from "./support/transport/publicJsonClient";');
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  assert.equal(checkFrontendImportBoundaries({ root: fixture.root, config }).errorCount, 0);
  fixture.write("apps/web/e2e/support/transport/publicJsonClient.ts", 'export * from "./rawPublicJson";');
  assert.ok(checkFrontendImportBoundaries({ root: fixture.root, config }).diagnostics.some((entry) => entry.ruleID === "web-e2e-raw-json-transport-boundary"));
});

for (const file of ["publicJsonClient.ts", "publicHttpOperationClient.ts", "rawPublicJson.ts"]) {
  test(`frontend transport boundary confines privileged control away from ${file}`, (t) => {
    const fixture = frontendFixture(t);
    fixture.write(`apps/web/e2e/support/transport/${file}`, 'import { TestControlClient } from "./testControlClient";');
    const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
    assert.ok(result.diagnostics.some((entry) => entry.ruleID === "web-e2e-public-control-transport-boundary"));
  });
}


test("frontend policy: allowed-grid", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/grid-adapter/src/index.tsx", "import { DataGrid } from \"react-data-grid\";\nimport \"react-data-grid/lib/styles.css\";\n\nexport const grid = DataGrid;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: blocked-grid", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/GridLeak.tsx", "import { DataGrid } from \"react-data-grid\";\n\nexport const grid = DataGrid;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-grid-vendor-boundary"));
  assert.ok(JSON.stringify(result.diagnostics).includes("apps/web/src/GridLeak.tsx"));
});

test("frontend policy: blocked-assessment-timeline", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx", "import { normalizeTimelineFullRow } from \"../timeline/models/timelineRowModel\";\n\nexport const leaked = normalizeTimelineFullRow;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-assessment-owner-no-timeline-imports"));
  assert.ok(JSON.stringify(result.diagnostics).includes("apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx"));
});

test("frontend policy: allowed-assessment-candidate", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/workbook/hooks/useAssessmentSupportCandidates.ts", "import { normalizeTimelineFullRow } from \"../timeline/models/timelineRowModel\";\n\nexport const candidateAdapter = normalizeTimelineFullRow;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: allowed-dom-unit", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/grid-adapter/src/domUnitBinding.test.tsx", "import { SemanticDataGrid } from \"./domUnitBinding\";\n\nexport const grid = SemanticDataGrid;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: blocked-dom-unit", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/GridDomUnitLeak.test.tsx", "import { SemanticDataGrid } from \"../../../packages/grid-adapter/src/domUnitBinding\";\n\nexport const grid = SemanticDataGrid;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-grid-dom-unit-binding-boundary"));
  assert.ok(JSON.stringify(result.diagnostics).includes("apps/web/src/GridDomUnitLeak.test.tsx"));
});

test("frontend policy: missing-stylesheet", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/grid-adapter/src/index.tsx", "export const gridStylesheetLoaded = false;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-rdg-stylesheet-singleton"));
  assert.ok(JSON.stringify(result.diagnostics).includes("expected exactly 1, found 0"));
});

test("frontend policy: duplicate-stylesheet", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/grid-adapter/src/duplicate.tsx", "import \"react-data-grid/lib/styles.css\";\n\nexport const duplicateGridStylesheetLoaded = true;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-rdg-stylesheet-singleton"));
  assert.ok(JSON.stringify(result.diagnostics).includes("expected exactly 1, found 2"));
});

test("frontend policy: outside-stylesheet", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/GridStylesheetLeak.tsx", "import \"react-data-grid/lib/styles.css\";\n\nexport const appGridStylesheetLoaded = true;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-rdg-stylesheet-singleton"));
  assert.ok(JSON.stringify(result.diagnostics).includes("apps/web/src/GridStylesheetLeak.tsx"));
});

test("frontend policy: generated-package", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/generatedProtocol.ts", "import { contractArtifactIndex } from \"@cartulary/protocol-ts/generated\";\n\nexport const contracts = contractArtifactIndex;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-boundary"));
});

test("frontend policy: generated-relative", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/support.ts", "import { contractArtifactIndex } from \"../../../packages/protocol-ts/src/generated/contracts\";\n\nexport const contracts = contractArtifactIndex;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-boundary"));
  assert.ok(JSON.stringify(result.diagnostics).includes("apps/web/e2e/support.ts"));
});

test("frontend policy: warning-fixture", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/warningFixture.ts", "import { fixture } from \"cartulary-warning-only-fixture/subpath\";\n\nexport const warningFixture = fixture;\n");
  const config = { schema_id: owner.schema_id, scan_roots: ["apps/web"], rules: [{ id: "synthetic-warning", level: "warning", message: "synthetic warning", applies_to: { include: ["**"] }, restricted_imports: [{ kind: "package", name: "cartulary-warning-only-fixture", include_subpaths: true }] }] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
  assert.equal(result.warningCount, 1);
});

test("frontend policy: family", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/package.json", "{\n  \"name\": \"@cartulary/protocol-ts\",\n  \"exports\": {\n    \"./http\": \"./src/entrypoints/http.ts\"\n  }\n}\n");
  fixture.write("apps/web/src/services/contracts.ts", "import { buildHTTPOperationPath } from \"@cartulary/protocol-ts/http\";\n\nexport const buildPath = buildHTTPOperationPath;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
  assert.equal(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-boundary"), false);
});

test("frontend policy: unsupported-root", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/package.json", "{\n  \"name\": \"@cartulary/protocol-ts\",\n  \"exports\": {\n    \"./http\": \"./src/entrypoints/http.ts\"\n  }\n}\n");
  fixture.write("apps/web/src/contracts.ts", "import { legacy } from \"@cartulary/protocol-ts\";\n\nexport const leaked = legacy;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-workspace-package-facade-boundary"));
});

test("frontend policy: protocol-entrypoint", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/src/entrypoints/errors.ts", "import { errorRegistry } from \"../generated/error-registry.js\";\n\nexport { errorRegistry };\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: protocol-root-bypass", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/src/index.ts", "import { errorArtifacts } from \"./generated/errors-artifacts.js\";\n\nexport const artifacts = errorArtifacts;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-boundary"));
});

test("frontend policy: protocol-wrong-entrypoint", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/src/entrypoints/collaboration.ts", "import { errorRegistry } from \"../generated/error-registry.js\";\n\nexport { errorRegistry };\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-contract-artifact-owner"));
});

test("frontend policy: protocol-barrel", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/src/entrypoints/errors.ts", "import { errorArtifacts } from \"../generated/index.js\";\n\nexport const artifacts = errorArtifacts;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-barrel-deny"));
});

test("frontend policy: protocol-obsolete-artifact", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/src/entrypoints/errors.ts", "import { errorArtifacts } from \"../generated/errors-artifacts.js\";\n\nexport { errorArtifacts };\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-obsolete-artifact-projection-deny"));
});

test("frontend policy: protocol-validator-barrel", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/src/entrypoints/http.ts", "export { validateCartularyCoreHttpErrorEnvelopeV1 } from \"../generated/protocol-validators.js\";\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-validator-barrel-deny"));
});

test("frontend policy: protocol-audit", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/src/entrypoints/errors.ts", "import { auditArtifacts } from \"../generated/audit-artifacts.js\";\n\nexport const artifacts = auditArtifacts;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-audit-deny"));
});

test("frontend policy: protocol-revisions", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/src/entrypoints/errors.ts", "import { revisionsArtifacts } from \"../generated/revisions-artifacts.js\";\n\nexport const artifacts = revisionsArtifacts;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-protocol-revisions-deny"));
});

test("frontend policy: design-token-owner", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/ui-contracts/src/designTokens.ts", "export { cartularyDesignTokenVars } from \"./generated/design-tokens\";\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: design-token-relative", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/designTokenBypass.ts", "import { cartularyDesignTokenVars } from \"../../../packages/ui-contracts/src/generated/design-tokens\";\n\nexport const vars = cartularyDesignTokenVars;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-generated-design-token-boundary"));
  assert.ok(JSON.stringify(result.diagnostics).includes("apps/web/src/designTokenBypass.ts"));
});

test("frontend policy: raw-design-literal", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/rawDesignLiteral.tsx", "export const accent = \"#FACC15\";\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-runtime-raw-design-color-literals"));
  assert.ok(JSON.stringify(result.diagnostics).includes("#FACC15"));
});

test("frontend policy: node-runtime", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/nodeRuntimeLeak.ts", "import path from \"node:path\";\n\nexport const joined = path.join(\"browser\", \"runtime\");\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-runtime-node-boundary"));
  assert.ok(JSON.stringify(result.diagnostics).includes("node:path"));
});

test("frontend policy: test-utils-runtime-dependency", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/test-utils/src/grid.ts", "import type { Page } from \"@playwright/test\";\nimport path from \"node:path\";\n\nexport const leaked = { path, page: undefined as Page | undefined };\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-test-utils-runtime-dependency-boundary"));
  assert.ok(JSON.stringify(result.diagnostics).includes("@playwright/test"));
  assert.ok(JSON.stringify(result.diagnostics).includes("node:path"));
});

test("frontend policy: test-utils-cycle", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/test-utils/src/grid.ts", "import { action } from \"./grid-actions\";\n\nexport const grid = action;\n");
  fixture.write("packages/test-utils/src/grid-actions.ts", "import { grid } from \"./grid\";\n\nexport const action = grid;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-test-utils-production-acyclic-import-graph"));
});

test("frontend policy: test-utils-observer-direction", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/test-utils/src/grid-actions.ts", "export const mutateGrid = true;\n");
  fixture.write("packages/test-utils/src/grid-setup.ts", "export const prepareGrid = true;\n");
  fixture.write("packages/test-utils/src/marker.ts", "import { mutateGrid } from \"./grid-actions\";\nimport { prepareGrid } from \"./grid-setup\";\n\nexport const observer = { mutateGrid, prepareGrid };\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-test-utils-observer-direction"));
  assert.ok(JSON.stringify(result.diagnostics).includes("./grid-actions"));
  assert.ok(JSON.stringify(result.diagnostics).includes("./grid-setup"));
});

test("frontend policy: node-e2e", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/nodeHarness.ts", "import path from \"node:path\";\n\nexport const joined = path.join(\"e2e\", \"harness\");\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: package-subpath", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/viewContractBypass.ts", "import { internal } from \"@cartulary/view-contracts/src/index\";\n\nexport const leaked = internal;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-workspace-package-facade-boundary"));
  assert.ok(JSON.stringify(result.diagnostics).includes("@cartulary/view-contracts/src/index"));
});

test("frontend policy: relative-package", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/viewContractRelativeBypass.ts", "import { internal } from \"../../../packages/view-contracts/src/index\";\n\nexport const leaked = internal;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-workspace-package-facade-boundary"));
  assert.ok(JSON.stringify(result.diagnostics).includes("apps/web/src/viewContractRelativeBypass.ts"));
});

test("frontend policy: declared-package-export", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/GridAdapter.test.tsx", "import { SemanticDataGrid } from \"@cartulary/grid-adapter/test-support\";\nimport { render } from \"@testing-library/react\";\nimport { it } from \"vitest\";\n\nexport const testGrid = { SemanticDataGrid, render, it };\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: test-helper-runtime", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/testing/timelineWorkbookTestSupport.ts", "export const support = true;\n");
  fixture.write("apps/web/src/runtimeHelperLeak.tsx", "import { helper } from \"@cartulary/test-utils\";\nimport { SemanticDataGrid } from \"@cartulary/grid-adapter/test-support\";\nimport { test } from \"@playwright/test\";\nimport { render } from \"@testing-library/react\";\nimport { describe } from \"vitest\";\nimport { support } from \"./testing/timelineWorkbookTestSupport\";\n\nexport const leaked = { SemanticDataGrid, describe, helper, render, support, test };\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-runtime-test-helper-boundary"));
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-runtime-test-helper-boundary" && item.specifier === "@cartulary/test-utils"));
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-runtime-test-helper-boundary" && item.specifier === "@cartulary/grid-adapter/test-support"));
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-runtime-test-helper-boundary" && item.specifier === "@playwright/test"));
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-runtime-test-helper-boundary" && item.specifier === "@testing-library/react"));
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-runtime-test-helper-boundary" && item.specifier === "vitest"));
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-runtime-test-helper-boundary" && item.specifier === "./testing/timelineWorkbookTestSupport"));
});

test("frontend policy: test-helper-allowed", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/runtimeHelperAllowed.test.tsx", "import { helper } from \"@cartulary/test-utils/grid\";\nimport { render } from \"@testing-library/react\";\nimport { describe } from \"vitest\";\n\nexport const allowed = { describe, helper, render };\n");
  fixture.write("apps/web/e2e/runtimeHarness.ts", "import { test } from \"@playwright/test\";\nimport { helper } from \"@cartulary/test-utils/grid\";\n\nexport const allowed = { helper, test };\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: e2e-semantic-allowed", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/runtimeHarness.ts", "import { helper } from \"@cartulary/test-utils/grid\";\nimport { metadata } from \"./support/workbook/query\";\n\nexport const allowed = { helper, metadata };\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: e2e-root-test-utils", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/runtimeHarness.ts", "import { helper } from \"@cartulary/test-utils\";\n\nexport const leaked = helper;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("web-e2e-test-utils-subpath-boundary"));
});

test("frontend policy: e2e-removed-test-utils-alias", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/runtimeHarness.ts", "import { helper } from \"@cartulary/test-utils/visual\";\n\nexport const leaked = helper;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-workspace-package-facade-boundary"));
});

test("frontend policy: e2e-private-test-utils-path", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/runtimeHarness.ts", "import { helper } from \"@cartulary/test-utils/browser\";\n\nexport const leaked = helper;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-workspace-package-facade-boundary"));
});

test("frontend policy: e2e-app-registry", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/runtimeHarness.ts", "import { registry } from \"../src/workbook/models/workbookSurfaceRegistry\";\n\nexport const leaked = registry;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("web-e2e-workbook-contract-boundary"));
});

test("frontend policy: e2e-production-source", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/runtimeHarness.ts", "import type { PrivateControllerState } from \"../src/app/privateController\";\n\nexport type Leaked = PrivateControllerState;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("web-e2e-production-source-boundary"));
});

test("frontend policy: e2e-public-control", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/support/transport/publicJsonClient.ts", "import { TestControlClient } from \"./testControlClient\";\n\nexport const leaked = TestControlClient;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("web-e2e-public-control-transport-boundary"));
});

test("frontend policy: e2e-semantic-visual", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/support/evidence/fixtures.ts", "import { visualFixture } from \"../visual/fixtures\";\n\nexport const leaked = visualFixture;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("web-e2e-semantic-fixture-direction"));
});

test("frontend policy: e2e-core-evidence", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/e2e/support/workbook/query.ts", "import { evidenceFixture } from \"../evidence/fixtures\";\n\nexport const leaked = evidenceFixture;\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("web-e2e-core-support-direction"));
});

test("frontend policy: protocol-entrypoint-allowed", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("contracts/protocol-ts/frontend-entrypoints.v2.json", "{\n  \"schema_id\": \"cartulary.protocol_ts_frontend_entrypoints.v2\",\n  \"entrypoints\": [\n    {\n      \"specifier\": \"@cartulary/protocol-ts/collaboration\",\n      \"authored_path\": \"packages/protocol-ts/src/entrypoints/collaboration.ts\",\n      \"generated_module_allowlist\": [\n        \"packages/protocol-ts/src/generated/collaboration-types.ts\"\n      ]\n    }\n  ]\n}\n");
  fixture.write("packages/protocol-ts/src/entrypoints/collaboration.ts", "export type { IncidentStreamMessage } from \"../generated/collaboration-types.js\";\n");
  fixture.write("packages/protocol-ts/package.json", "{\n  \"name\": \"@cartulary/protocol-ts\",\n  \"exports\": {\n    \"./collaboration\": \"./src/entrypoints/collaboration.ts\"\n  }\n}\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});

test("frontend policy: protocol-entrypoint-forbidden", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/src/entrypoints/collaboration.ts", "export { validateCartularyCoreHttpErrorEnvelopeV1 } from \"../generated/core-http-validators.js\";\n");
  fixture.write("contracts/protocol-ts/frontend-entrypoints.v2.json", "{\n  \"schema_id\": \"cartulary.protocol_ts_frontend_entrypoints.v2\",\n  \"entrypoints\": [\n    {\n      \"specifier\": \"@cartulary/protocol-ts/collaboration\",\n      \"authored_path\": \"packages/protocol-ts/src/entrypoints/collaboration.ts\",\n      \"generated_module_allowlist\": [\n        \"packages/protocol-ts/src/generated/collaboration-types.ts\"\n      ]\n    }\n  ]\n}\n");
  fixture.write("packages/protocol-ts/package.json", "{\n  \"name\": \"@cartulary/protocol-ts\",\n  \"exports\": {\n    \"./collaboration\": \"./src/entrypoints/collaboration.ts\"\n  }\n}\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-protocol-ts-entrypoint-generated-allowlist"));
});

test("frontend policy: protocol-export-map-mismatch", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/protocol-ts/package.json", "{\n  \"name\": \"@cartulary/protocol-ts\",\n  \"exports\": {\n    \".\": \"./src/index.ts\",\n    \"./collaboration\": \"./src/entrypoints/collaboration.ts\"\n  }\n}\n");
  fixture.write("packages/protocol-ts/src/entrypoints/collaboration.ts", "export type { IncidentStreamMessage } from \"../generated/collaboration-types.js\";\n");
  fixture.write("contracts/protocol-ts/frontend-entrypoints.v2.json", "{\n  \"schema_id\": \"cartulary.protocol_ts_frontend_entrypoints.v2\",\n  \"entrypoints\": [\n    {\n      \"specifier\": \"@cartulary/protocol-ts/collaboration\",\n      \"authored_path\": \"packages/protocol-ts/src/entrypoints/collaboration.ts\",\n      \"generated_module_allowlist\": [\n        \"packages/protocol-ts/src/generated/collaboration-types.ts\"\n      ]\n    }\n  ]\n}\n");
  const config = { ...owner, scan_roots: ["apps/web", "packages"] };
  const result = checkFrontendImportBoundaries({ root: fixture.root, config });
  assert.ok(result.errorCount > 0);
  assert.ok(JSON.stringify(result.diagnostics).includes("frontend-protocol-ts-entrypoint-export-map"));
});

test("frontend owner: type alias", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/workbook/timeline/models/renamed.ts", "import type { AssessmentCreateDraft as Local } from \"../../models/assessmentWorkbookModel\"; export type Extracted = Local;");
  const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-timeline-no-assessment-owner"));
});

test("frontend owner: re-export", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx", "export { row as Renamed } from \"../timeline/models/row\";");
  const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-assessment-owner-no-timeline-imports"));
});

test("frontend owner: dynamic import", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/networkFlow/renamed.ts", "export const load = () => import(\"react-data-grid\");");
  const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-grid-vendor-boundary"));
});

test("frontend owner: transport alias", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/workbook/policies/renamed.ts", "import * as local from \"../../services/browserApi\"; export const capability = local;");
  const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-workbook-policy-transport-boundary"));
});

test("frontend owner: comments", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("apps/web/src/workbook/timeline/models/renamed.ts", "// AssessmentCreateDraft react-data-grid authorization\nimport type { Row as Local } from \"./row\"; export type Extracted = Local;");
  const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
  assert.equal(result.errorCount, 0);
});

test("frontend policy: generated tokens reject same-package bypass", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/ui-contracts/src/index.ts", 'export { cartularyDesignTokenVars } from "./generated/design-tokens";');
  const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
  assert.ok(result.diagnostics.some((item) => item.ruleID === "frontend-generated-design-token-boundary"));
});

test("frontend policy: non-design generated contracts retain their owner", (t) => {
  const fixture = frontendFixture(t);
  fixture.write("packages/ui-contracts/src/performanceContracts.ts", 'export { contract } from "./generated/performance-contracts";');
  const result = checkFrontendImportBoundaries({ root: fixture.root, config: { ...owner, scan_roots: ["apps/web", "packages"] } });
  assert.equal(result.errorCount, 0, JSON.stringify(result.diagnostics));
});
