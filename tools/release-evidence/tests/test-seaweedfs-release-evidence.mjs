#!/usr/bin/env node
import assert from "node:assert/strict";
import { Buffer } from "node:buffer";
import fs, { writeFileSync } from "node:fs";
import childProcess from "node:child_process";
import { syncBuiltinESMExports } from "node:module";
import test from "node:test";
import path from "node:path";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import "./test-seaweedfs-artifacts.mjs";

import {
  buildDependencyBoundary,
  buildSeaweedfsReleaseReport,
  buildBackupIntegrityCoverage,
  buildOccurrenceInventoryFromEntries,
  buildReleaseManifestExposure,
  buildSbomLicenseClassification,
  buildStorageRefOwnerCoverage,
  buildThreatModelCoverage,
  globToRegExp,
  isDefaultExcluded,
  isTextBuffer,
  occurrenceTokens,
  readRepositorySource,
  repositorySourceFiles,
} from "../seaweedfs-release-evidence.mjs";

test("SeaweedFS producer assembly projects each boundary and cannot hide unqualified recovery", () => {
  const context = { manifest: { run_id: "assembly-fixture", source_commit: "a".repeat(40),
    source_digest: `sha256:${"a".repeat(64)}`, toolchain_digest: `sha256:${"b".repeat(64)}`, graph_digest: `sha256:${"c".repeat(64)}` }, target: "seaweedfs-release-gate" };
  const source = {
    occurrence: { result: "pass", occurrences: [] }, exposure: { result: "pass", findings: [] },
    dependency: { result: "pass", sdk_dependency: { disallowed_imports: [] }, release_server_artifacts: [] },
    policies: { threat: { result: "pass" }, storageRefOwner: { result: "pass" }, backupIntegrity: { result: "pass" } },
  };
  // Assembly consumes already-admitted snapshots; their actual file/shape
  // admission is independently exercised by test-seaweedfs-artifacts.
  const inventory = { sbom: { components: [] }, licenseReport: { entries: [] }, evidence_refs: ["inventory/sbom.json", "inventory/license.json"] };
  const compatible = { unit: { status: "passed" }, report: { result: "pass" }, evidence_refs: ["current/unit.json", "current/compatibility.json"] };
  const build = input => buildSeaweedfsReleaseReport(context, { source, inventory, compatible, ...input });
  const report = build({});
  assert.equal(report.status, "blocked");
  assert.deepEqual(report.checks.filter(item => item.status !== "pass").map(item => [item.check_id, item.status]), [["recovery", "fail"]]);
  assert.deepEqual(report.findings.map(item => item.code), ["current-recovery-unqualified"]);
  const recovery = { row: { terminal_state: "passed" }, unit: { status: "passed" }, evidence_refs: ["current/recovery-unit.json", "current/recovery-row.json"] };
  const qualified = build({ recovery });
  assert.equal(qualified.status, "pass");
  assert.deepEqual(qualified.checks.find(item => item.check_id === "recovery").evidence_refs, recovery.evidence_refs);
  assert.ok(qualified.checks.find(item => item.check_id === "redaction").evidence_refs.includes("current/recovery-row.json"));
  const recoveryLeak = build({ recovery: { ...recovery, row: { endpoint: "https://source.internal" } } });
  assert.equal(recoveryLeak.status, "blocked");
  assert.equal(recoveryLeak.checks.find(item => item.check_id === "redaction").status, "fail");
  assert.doesNotMatch(JSON.stringify(recoveryLeak), /source\.internal/u);
  const missing = build({ inventory: null, compatible: null });
  assert.deepEqual(missing.checks.filter(item => item.status !== "pass").map(item => item.check_id), ["inventory_classification", "compatibility", "recovery"]);
  const changed = structuredClone(source);
  changed.exposure = { result: "fail", findings: [{ check_id: "runtime-seaweedfs-admin-api", path: "source/file.go", line: 1, column: 2, excerpt: "sensitive-fixture-value" }] };
  changed.dependency.result = "fail";
  changed.dependency.sdk_dependency.disallowed_imports.push({ path: "source/other.go", line: 3, column: 4 });
  const rejected = build({ source: changed });
  assert.deepEqual(rejected.findings.slice(0, 2), [
    { check_id: "release_surface", code: "runtime-seaweedfs-admin-api", path: "source/file.go", line: 1, column: 2, json_path: null },
    { check_id: "dependency_boundary", code: "sdk-import-boundary", path: "source/other.go", line: 3, column: 4, json_path: null },
  ]);
  assert.doesNotMatch(JSON.stringify(rejected), /sensitive-fixture-value/u);
  const leaked = build({ compatible: { ...compatible, report: { diagnostic: "seaweedfs-secret" } } });
  assert.equal(leaked.checks.find(item => item.check_id === "redaction").status, "fail");
  assert.doesNotMatch(JSON.stringify(leaked), /seaweedfs-secret/u);
});

test("release source admission prunes human inputs before probes and rejects unsafe source objects", t => {
  const root = createScratch(t, "cartulary-release-source-");
  fs.mkdirSync(path.join(root, "source"));
  writeFileSync(path.join(root, "source/accepted.go"), "package accepted\n");
  fs.symlinkSync("accepted.go", path.join(root, "source/alias.go"));
  fs.symlinkSync("source", path.join(root, "alias"));
  const forbidden = ["human/private.go", "README", "nested/Readme.txt", "guide.MD", "nested/notes.MarkDown"];
  const originals = { spawnSync: childProcess.spawnSync, lstatSync: fs.lstatSync, openSync: fs.openSync };
  let enumerations = 0;
  let listed = ["source/accepted.go", "source/deleted.go", ...forbidden];
  let gitStatus = 0;
  try {
    childProcess.spawnSync = (command, args, options) => {
      enumerations++;
      assert.equal(command, "git");
      assert.equal(options.cwd, root);
      assert.deepEqual(args, ["ls-files", "--cached", "--others", "--exclude-standard", "-z", "--", ".",
        ":(top,exclude,literal)human", ":(exclude,icase)*.md", ":(exclude,icase)*.markdown",
        ":(exclude,icase)README", ":(exclude,icase)README.*", ":(exclude,icase)**/README", ":(exclude,icase)**/README.*"]);
      // Deliberately return excluded entries too: defensive lexical admission
      // must protect probes even if enumeration supplies surplus paths.
      return { status: gitStatus, stdout: listed.join("\0") + "\0" };
    };
    for (const name of ["lstatSync", "openSync"]) fs[name] = (file, ...args) => {
      assert.ok(!forbidden.some(relative => String(file) === path.join(root, relative)));
      assert.notEqual(String(file), path.join(root, "human"));
      return originals[name](file, ...args);
    };
    syncBuiltinESMExports();
    assert.deepEqual(repositorySourceFiles(root, ["human"]), ["source/accepted.go"]);
    assert.equal(enumerations, 1);
    for (const relative of forbidden) assert.throws(() => readRepositorySource(relative, root, ["human"]), /outside executable scope/u);
    assert.equal(readRepositorySource("source/accepted.go", root, ["human"]).toString(), "package accepted\n");
    listed = ["source/alias.go"];
    assert.throws(() => repositorySourceFiles(root, ["human"]), /must be a regular file/u);
    listed = ["alias/accepted.go"];
    assert.throws(() => repositorySourceFiles(root, ["human"]), /ancestor is not a directory/u);
    gitStatus = 1;
    assert.throws(() => repositorySourceFiles(root, ["human"]), /cannot enumerate/u);
  } finally {
    childProcess.spawnSync = originals.spawnSync;
    fs.lstatSync = originals.lstatSync; fs.openSync = originals.openSync;
    syncBuiltinESMExports();
  }
  for (const relative of ["source/alias.go", "alias/accepted.go", "source", "../outside", "/absolute", "source/../accepted.go"]) {
    assert.throws(() => readRepositorySource(relative, root, ["human"]), undefined, relative);
  }
  const bounded = path.join(root, "bounded.bin");
  writeFileSync(bounded, ""); fs.truncateSync(bounded, 64 * 1024 * 1024);
  assert.equal(readRepositorySource("bounded.bin", root, ["human"]).length, 64 * 1024 * 1024);
  fs.truncateSync(bounded, 64 * 1024 * 1024 + 1);
  assert.throws(() => readRepositorySource("bounded.bin", root, ["human"]), /byte limit/u);
});

const legacyStem = "min" + "io";
const legacyTitle = "Min" + "IO";
const upperLegacyTitle = "MIN" + "IO";
const bucketToken = `${legacyStem}_bucket`;
const endpointToken = `${legacyStem}_endpoint`;
const sdkModule = `github.com/${legacyStem}/${legacyStem}-go/v7`;
const serverImage = `${legacyStem}/${legacyStem}`;
const generatedPolicy = {
  roots: ["internal/gen", "packages/protocol-ts/src/generated"],
  files: ["tools/task_surface.generated.mk"],
};

const rules = [
  {
    id: "sdk",
    pathRegexps: [globToRegExp("adapter/**")],
    classification: "sdk_only",
    owner: "fixture sdk owner",
    rationale: "fixture SDK boundary",
  },
  {
    id: "legacy",
    pathRegexps: [globToRegExp("testdata/history/**")],
    classification: "historical_changelog",
    owner: "fixture archive owner",
    rationale: "fixture archive boundary",
  },
  {
    id: "invalid",
    pathRegexps: [globToRegExp("bad/**")],
    classification: "invalid",
  },
];

assert.deepEqual(occurrenceTokens, [
  legacyTitle,
  legacyStem,
  upperLegacyTitle,
  bucketToken,
  endpointToken,
]);

assert.equal(isDefaultExcluded(".cartulary/test-results/run.json", generatedPolicy), true);
assert.equal(isDefaultExcluded("internal/gen/contracts/contracts_gen.go", generatedPolicy), true);
assert.equal(isDefaultExcluded("tools/task_surface.generated.mk", generatedPolicy), true);
assert.equal(isDefaultExcluded("internal/platform/objectstore/objectstore.go", generatedPolicy), false);
assert.equal(isTextBuffer(Buffer.from("plain text")), true);
assert.equal(isTextBuffer(Buffer.from([0x61, 0x00, 0x62])), false);

const inventory = buildOccurrenceInventoryFromEntries({
  rules,
  repoCommitValue: "abc123",
  scannedAt: "2026-06-04T00:00:00.000Z",
  entries: [
    {
      path: "testdata/history/legacy.txt",
      text: `${legacyTitle} history\n`,
    },
    {
      path: "adapter/client.go",
      text: `import "${sdkModule}"\nconst a = "${bucketToken}"\n`,
    },
    {
      path: "bad/default.txt",
      text: `${upperLegacyTitle} server default\n`,
    },
    {
      path: "unclassified.txt",
      text: `${endpointToken}\n`,
    },
  ],
});
assert.equal(inventory.schema_id, "cartulary.seaweedfs_occurrence_inventory.v1");
assert.equal(inventory.result, "fail");
assert.equal(inventory.occurrences.length, 8);
assert.deepEqual(
  inventory.occurrences.map((entry) => [entry.path, entry.line, entry.column, entry.token, entry.classification]),
  [
    ["adapter/client.go", 1, 20, legacyStem, "sdk_only"],
    ["adapter/client.go", 1, 26, legacyStem, "sdk_only"],
    ["adapter/client.go", 2, 12, legacyStem, "sdk_only"],
    ["adapter/client.go", 2, 12, bucketToken, "sdk_only"],
    ["bad/default.txt", 1, 1, upperLegacyTitle, "invalid"],
    ["testdata/history/legacy.txt", 1, 1, legacyTitle, "historical_changelog"],
    ["unclassified.txt", 1, 1, legacyStem, "unclassified"],
    ["unclassified.txt", 1, 1, endpointToken, "unclassified"],
  ].sort((a, b) => a[0].localeCompare(b[0]) || a[1] - b[1] || a[2] - b[2] || a[3].localeCompare(b[3])),
);
assert.equal(
  inventory.occurrences.every((entry) => !Object.hasOwn(entry, "line_excerpt")),
  true,
);

const passingInventory = buildOccurrenceInventoryFromEntries({
  rules,
  entries: [
    {
      path: "adapter/client.go",
      text: `import "${sdkModule}"\n`,
    },
  ],
});
assert.equal(passingInventory.result, "pass");

const releaseFixture = new Map([
  ["docker-compose.release-test.yml", `services:\n  object:\n    image: ${serverImage}:latest\n`],
  ["configs/release-test.toml", "cors_allowed_origins = \"*\"\n"],
  ["internal/modules/evidence/leak.go", 'const api = "/cluster/status"\n'],
]);
const exposure = buildReleaseManifestExposure({
  files: [...releaseFixture.keys()],
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  readFile: (rel) => releaseFixture.get(rel),
});
assert.equal(exposure.result, "fail");
assert.deepEqual(
  new Set(exposure.findings.map((finding) => finding.check_id)),
  new Set(["legacy-server-image", "wildcard-direct-upload-cors", "runtime-seaweedfs-admin-api"]),
);

const cleanExposure = buildReleaseManifestExposure({
  files: ["docker-compose.dev.yml"],
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  readFile: () =>
    "services:\n  object:\n    image: docker.io/chrislusf/seaweedfs:4.17@sha256:abc\n    ports:\n      - \"8333:8333\"\n    command: [\"server\", \"-s3\", \"-webdav=false\"]\n",
});
assert.equal(cleanExposure.result, "pass");

const dependencyFixture = new Map([
  ["go.mod", `module fixture\nrequire ${sdkModule} v7.0.100\n`],
  ["internal/platform/objectstore/objectstore.go", `import "${sdkModule}"\n`],
  ["internal/modules/evidence/evidence.go", `import "${sdkModule}"\n`],
]);
const dependency = buildDependencyBoundary({
  files: [...dependencyFixture.keys()],
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  goModText: dependencyFixture.get("go.mod"),
  readFile: (rel) => dependencyFixture.get(rel),
});
assert.equal(dependency.result, "fail");
assert.equal(dependency.sdk_dependency.disallowed_imports.length, 1);

const dependencyServer = buildDependencyBoundary({
  files: ["go.mod", "internal/platform/objectstore/objectstore.go"],
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  goModText: dependencyFixture.get("go.mod"),
  readFile: (rel) => dependencyFixture.get(rel) ?? `import "${sdkModule}"\n`,
  sbom: { components: [{ name: serverImage }] },
  licenseReport: { dependencies: [{ package: sdkModule }] },
});
assert.equal(dependencyServer.result, "fail");
assert.equal(dependencyServer.release_server_artifacts.length, 1);

const sbomLicense = buildSbomLicenseClassification({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  sbomPath: "fixture-sbom.json",
  licensePath: "fixture-license.json",
  sbom: { components: [{ name: sdkModule, purl: `pkg:golang/${sdkModule}` }] },
  licenseReport: { dependencies: [{ package_name: sdkModule, license: "Apache-2.0" }] },
});
assert.equal(sbomLicense.result, "pass");
assert.equal(sbomLicense.sdk_dependency.classification, "sdk_only");

const sbomLicenseServer = buildSbomLicenseClassification({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  sbomPath: "fixture-sbom.json",
  licensePath: "fixture-license.json",
  sbom: { components: [{ name: serverImage }] },
  licenseReport: { dependencies: [] },
});
assert.equal(sbomLicenseServer.result, "fail");
assert.equal(sbomLicenseServer.release_server_artifacts.length, 1);

const threat = buildThreatModelCoverage({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
});
assert.equal(threat.result, "pass");

const storageRefOwner = buildStorageRefOwnerCoverage({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
});
assert.equal(storageRefOwner.result, "pass");

const backupIntegrity = buildBackupIntegrityCoverage({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
});
assert.equal(backupIntegrity.result, "pass");
assert.equal(backupIntegrity.schema_id, "cartulary.seaweedfs_backup_integrity_coverage.v1");
assert.deepEqual(backupIntegrity.required_schema_ids, [
  "cartulary.backup_artifact_envelope.v3",
  "cartulary.backup_integrity_manifest.v4",
  "cartulary.object_store_backup_manifest.v2",
  "cartulary.object_store_backup_summary.v2",
]);
