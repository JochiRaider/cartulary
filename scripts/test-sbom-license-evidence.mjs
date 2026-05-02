#!/usr/bin/env node
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  artifactRows,
  deterministicLicenseFilename,
  findLicenseFiles,
  licenseReviewFlags,
  makeCycloneDxBom,
  normalizePackageName,
  parseGoModuleGraph,
  parseJSONStream,
} from "./generate-sbom-license-evidence.mjs";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const tmp = mkdtempSync(path.join(os.tmpdir(), "cartulary-sbom-test-"));

try {
  assert.equal(
    deterministicLicenseFilename("npm", "@scope/pkg name", "1.2.3-beta.1"),
    "npm__scope_pkg_name__1.2.3-beta.1__LICENSE.txt",
  );
  assert.equal(normalizePackageName("@scope/pkg name"), "scope_pkg_name");

  const packageDir = path.join(tmp, "package");
  writeFileSync(path.join(tmp, "a.txt"), "alpha");
  writeFileSync(path.join(tmp, "b.txt"), "beta");
  await import("node:fs").then(({ mkdirSync }) => mkdirSync(packageDir));
  writeFileSync(path.join(packageDir, "LICENSE"), "MIT fixture");
  writeFileSync(path.join(packageDir, "NOTICE.txt"), "notice fixture");
  assert.deepEqual(
    findLicenseFiles(packageDir).map((file) => path.basename(file)),
    ["LICENSE", "NOTICE.txt"],
  );

  assert.deepEqual(licenseReviewFlags("GPL-3.0-only", null), ["copyleft_review", "legal_review"]);
  assert.deepEqual(licenseReviewFlags("MIT", null), ["notice_or_attribution_review"]);

  assert.deepEqual(parseJSONStream('{"a":1}\n{"b":{"c":2}}'), [{ a: 1 }, { b: { c: 2 } }]);
  assert.deepEqual(parseGoModuleGraph("a@v1 b@v2\nb@v2 c@v3\n"), [
    ["a@v1", "b@v2"],
    ["b@v2", "c@v3"],
  ]);

  const rows = artifactRows([path.join(tmp, "b.txt"), path.join(tmp, "a.txt")]);
  assert.equal(rows.length, 2);
  assert.equal(rows[0].bytes, 5);
  assert.match(rows[0].sha256, /^[a-f0-9]{64}$/);

  const validBom = path.join(tmp, "valid.cyclonedx.json");
  const bom = makeCycloneDxBom({
    name: "fixture",
    version: "0.0.0",
    bomRefValue: "fixture:application",
    components: [],
    dependencies: [],
    tools: [{ type: "application", name: "fixture-tool", version: "0.0.0" }],
    timestamp: "2026-05-02T00:00:00.000Z",
    completeness: "complete",
  });
  writeFileSync(validBom, `${JSON.stringify(bom, null, 2)}\n`);
  const valid = spawnSync(process.execPath, [path.join(repoRoot, "scripts", "validate-cyclonedx.mjs"), validBom], {
    cwd: repoRoot,
    encoding: "utf8",
  });
  assert.equal(valid.status, 0, valid.stderr);

  const invalidBom = path.join(tmp, "invalid.cyclonedx.json");
  writeFileSync(invalidBom, "{}\n");
  const invalid = spawnSync(process.execPath, [path.join(repoRoot, "scripts", "validate-cyclonedx.mjs"), invalidBom], {
    cwd: repoRoot,
    encoding: "utf8",
  });
  assert.notEqual(invalid.status, 0);
  assert.match(invalid.stderr, /missing specVersion/);
} finally {
  rmSync(tmp, { recursive: true, force: true });
}
