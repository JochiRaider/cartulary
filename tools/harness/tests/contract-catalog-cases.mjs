import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import childProcess, { execFileSync } from "node:child_process";
import fs, { chmodSync, copyFileSync, mkdirSync, mkdtempSync, readFileSync, readlinkSync, renameSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import { tmpdir } from "node:os";
import path from "node:path";
import { createRowSelectorResolver, discoverRunnerSources, validatePostgresFixturePolicy } from "../test-support/owner-fixtures.mjs";
import { groupRowsByPerformanceFixture, postgresMigrationDigest, snapshotKey } from "../performance-fixture/index.mjs";
import { buildSourceSnapshot, loadSelectedTestRows, loadTestCatalog } from "../test-catalog/index.mjs";

export function assertSourceSnapshotIdentity() {
  const root = mkdtempSync(path.join(tmpdir(), "cartulary-source-identity."));
  const hash = (bytes) => `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
  const large = Buffer.from(Array.from({ length: 2 * 65536 + 17 }, (_, index) => index % 251));
  const contents = new Map([["empty", Buffer.alloc(0)], ["large", large], ["unicode", Buffer.from("α😀\0\r\n")]]);
  const target = path.join(root, "large");
  const originalRead = fs.readSync;
  try {
    execFileSync("git", ["init", "-q"], { cwd: root });
    for (const [name, bytes] of contents) writeFileSync(path.join(root, name), bytes, { mode: 0o600 });
    const link = "../absent-external-input";
    symlinkSync(link, path.join(root, "link"));
    const entries = [
      ...[...contents].map(([name, bytes]) => ({ byte_digest: hash(bytes), kind: "file", mode: "0600", path: name })),
      { byte_digest: hash(Buffer.from(link)), kind: "symlink", mode: "0777", path: "link" },
    ].sort((left, right) => left.path < right.path ? -1 : 1);
    const result = buildSourceSnapshot(root, { restrictedRoots: [] });
    assert.deepEqual(result.entries, entries);
    // Keys are authored in canonical order, independently of the production
    // serializer. Binary, empty, multi-chunk and link bytes retain v2 identity.
    assert.equal(result.digest, hash(Buffer.from(JSON.stringify({ entries, schema_id: "cartulary.source_snapshot_digest.v2" }))));
    assert.equal(result.file_count, entries.length);

    // Check admission before Git enumeration, then prohibit filesystem probes
    // of excluded material. Include an executable-looking untracked file under
    // a restricted root and mixed-case Markdown beside ordinary source.
    const excludedRoot = "human-material";
    mkdirSync(path.join(root, excludedRoot, "untracked"), { recursive: true });
    writeFileSync(path.join(root, excludedRoot, "untracked", "hidden.test.ts"), "throw new Error('not executable');");
    writeFileSync(path.join(root, "owner.Md"), "not executable");
    writeFileSync(path.join(root, "README"), "not executable");
    const originalExec = childProcess.execFileSync;
    const originalStat = fs.lstatSync;
    const originalOpen = fs.openSync;
    const assertAdmitted = (file) => assert.ok(!String(file).startsWith(path.join(root, excludedRoot)) && !/\/(?:owner\.Md$|README$)/u.test(String(file)), `probed excluded input: ${file}`);
    let checkedEnumeration = false;
    childProcess.execFileSync = (command, args, options) => {
      if (command === "git" && args[0] === "ls-files") {
        const pathspec = args.slice(args.indexOf("--") + 1);
        for (const excluded of [`:(top,exclude,literal)${excludedRoot}`, ":(exclude,icase)*.md", ":(exclude,icase)README"])
          assert.ok(pathspec.includes(excluded), `missing pre-enumeration exclusion ${excluded}`);
        checkedEnumeration = true;
      }
      return originalExec(command, args, options);
    };
    fs.lstatSync = (file, ...args) => { assertAdmitted(file); return originalStat(file, ...args); };
    fs.openSync = (file, ...args) => { assertAdmitted(file); return originalOpen(file, ...args); };
    syncBuiltinESMExports();
    try {
      assert.deepEqual(buildSourceSnapshot(root, { restrictedRoots: [excludedRoot] }), result);
      assert.equal(checkedEnumeration, true);
    } finally {
      childProcess.execFileSync = originalExec; fs.lstatSync = originalStat; fs.openSync = originalOpen;
      syncBuiltinESMExports();
      rmSync(path.join(root, excludedRoot), { recursive: true });
      rmSync(path.join(root, "owner.Md")); rmSync(path.join(root, "README"));
    }

    for (const mutate of [
      () => fs.appendFileSync(target, "growth"),
      () => fs.truncateSync(target, 1),
      () => { renameSync(target, `${target}.old`); writeFileSync(target, large, { mode: 0o600 }); },
      () => chmodSync(target, 0o644),
      () => { throw new Error("injected source read failure"); },
    ]) {
      let observedDescriptor, injected = false;
      fs.readSync = (...args) => {
        const count = originalRead(...args);
        if (!injected && readlinkSync(`/proc/self/fd/${args[0]}`) === target) {
          injected = true; observedDescriptor = args[0]; mutate();
        }
        return count;
      };
      syncBuiltinESMExports();
      try {
        assert.throws(() => buildSourceSnapshot(root, { restrictedRoots: [] }), /source snapshot file|injected source read/u);
        assert.equal(injected, true);
        assert.throws(() => fs.fstatSync(observedDescriptor), { code: "EBADF" });
      } finally {
        fs.readSync = originalRead; syncBuiltinESMExports();
        rmSync(`${target}.old`, { force: true });
        writeFileSync(target, large); chmodSync(target, 0o600);
      }
    }

    // A tracked descendant remains in Git's index after its directory becomes
    // a link. Reject the ancestor before any stat/open of that descendant.
    mkdirSync(path.join(root, "nested"));
    writeFileSync(path.join(root, "nested", "source.ts"), "export const value = 1;");
    execFileSync("git", ["add", "nested/source.ts"], { cwd: root });
    renameSync(path.join(root, "nested"), path.join(root, "original-nested"));
    symlinkSync("original-nested", path.join(root, "nested"));
    const ancestorOriginalStat = fs.lstatSync;
    fs.lstatSync = (file, ...args) => {
      assert.notEqual(String(file), path.join(root, "nested", "source.ts"), "probed a descendant through a symbolic link");
      return ancestorOriginalStat(file, ...args);
    };
    syncBuiltinESMExports();
    try {
      assert.throws(() => buildSourceSnapshot(root, { restrictedRoots: [] }), /ancestor.*symbolic links/u);
    } finally {
      fs.lstatSync = ancestorOriginalStat;
      syncBuiltinESMExports();
    }
  } finally {
    fs.readSync = originalRead; syncBuiltinESMExports();
    rmSync(root, { recursive: true, force: true });
  }
}

export function assertPrivateRowSelection(context) {
  const catalog = context.catalog;
  const selected = ["go", "vitest", "shell"].map((runner) => catalog.rows.find((row) => row.runner === runner));
  assert.deepEqual(loadSelectedTestRows(context.root, selected.map((row) => row.row_id).reverse()),
    [...selected].sort((left, right) => left.row_id < right.row_id ? -1 : 1));

  // All metadata is real and complete. Only the selected Go source exists, so
  // full preflight must fail while a private selected-source recheck can pass.
  const fixture = mkdtempSync(path.join(tmpdir(), "cartulary-private-row-selection."));
  const copy = (relative) => {
    const destination = path.join(fixture, relative);
    mkdirSync(path.dirname(destination), { recursive: true });
    copyFileSync(path.join(context.root, relative), destination);
  };
  const read = (relative) => JSON.parse(readFileSync(path.join(fixture, relative), "utf8"));
  const write = (relative, value) => writeFileSync(path.join(fixture, relative), JSON.stringify(value));
  try {
    for (const relative of [
      "tools/test_catalog_owner.json", "contracts/verification/registry.json",
      "tools/test_runner_registry.json", "tools/execution_topology_manifest.json",
      "tools/scheduler_resource_registry.json", "tools/performance_fixture_snapshot_owner.json",
      "tools/task_surface_owner.json", "tools/test_catalog_row_migrations.json",
      "tools/postgres_fixture_policy_registry.json", "tools/toolchain_pins.json",
      "tools/generated_artifact_policy.json",
      ...catalog.registry.owners.map((owner) => owner.manifest_path),
      ...catalog.verification.registry.owners.map((owner) => owner.contract_path),
      ...catalog.runners.runners.map((runner) => runner.adapter_path),
      ...catalog.fixtureProfiles.registry.profiles.flatMap((profile) => profile.source_contract_refs.map((ref) => ref.path)),
    ]) copy(relative);
    for (const runner of catalog.runners.runners) {
      for (const approved of runner.approved_roots) mkdirSync(path.join(fixture, approved), { recursive: true });
    }

    const goRow = selected[0];
    const familyPath = `tools/test_families/${goRow.owner_id}.json`;
    const family = read(familyPath);
    const row = family.rows.find((entry) => entry.row_id === goRow.row_id);
    row.selector = { package: "./internal/fixture", tests: ["TestSelected"] };
    write(familyPath, family);
    const packageRoot = path.join(fixture, "internal/fixture");
    mkdirSync(packageRoot, { recursive: true });
    const source = path.join(packageRoot, "fixture_test.go");
    const sourceText = "package fixture\nfunc TestSelected() {}\n";
    writeFileSync(source, sourceText);
    const load = (ids = [row.row_id]) => loadSelectedTestRows(fixture, ids);
    assert.deepEqual(load(), [row]);
    assert.throws(() => loadTestCatalog(fixture), /does not exist|ENOENT/u);
    for (const ids of [[], [row.row_id, row.row_id], [""], [null]]) {
      assert.throws(() => load(ids), /invalid exact row selection/u);
    }
    rmSync(source);
    assert.throws(() => load([row.row_id, "harness.unknown.row"]), /unknown active row/u);
    assert.throws(() => load(), /must resolve exactly once/u);
    writeFileSync(source, sourceText.replace("TestSelected", "TestChanged"));
    assert.throws(() => load(), /must resolve exactly once/u);
    writeFileSync(source, sourceText);
    const duplicate = path.join(packageRoot, "duplicate_test.go");
    writeFileSync(duplicate, sourceText);
    assert.throws(() => load(), /must resolve exactly once/u);
    rmSync(duplicate);
    renameSync(packageRoot, `${packageRoot}-original`);
    symlinkSync(`${packageRoot}-original`, packageRoot);
    assert.throws(() => load(), /symbolic link/u);
    rmSync(packageRoot); renameSync(`${packageRoot}-original`, packageRoot);

    const overlapping = family.rows.find((entry) => entry.runner === "go" && entry.row_id !== row.row_id);
    assert.ok(overlapping);
    const oldSelector = overlapping.selector;
    overlapping.selector = structuredClone(row.selector); write(familyPath, family);
    assert.throws(() => load([row.row_id, overlapping.row_id]), /selector overlaps/u);
    overlapping.selector = oldSelector;
    // Unselected malformed metadata remains a hard failure at the child boundary.
    overlapping.runtime_profile_id = "unknown_profile"; write(familyPath, family);
    assert.throws(() => load(), /runtime_profile_id/u);
  } finally {
    rmSync(fixture, { recursive: true, force: true });
  }
}

export function assertSelectorResolutionScope(context) {
  const directory = mkdtempSync(path.join(tmpdir(), "cartulary-selector-scope."));
  const webDirectory = path.join(directory, "apps/web/src");
  try {
    mkdirSync(webDirectory, { recursive: true });
    const tools = path.join(directory, "tools"), packageRoot = path.join(directory, "internal/fixture");
    mkdirSync(tools); mkdirSync(packageRoot, { recursive: true });
    const pins = path.join(tools, "toolchain_pins.json");
    writeFileSync(pins, JSON.stringify({ go_version: "1.27.2" }));
    writeFileSync(path.join(tools, "generated_artifact_policy.json"), JSON.stringify({ generated_roots: [], generated_files: [] }));
    const source = path.join(packageRoot, "fixture_test.go");
    writeFileSync(source, "//go:build go1.27\n\npackage fixture\nfunc TestCurrent() {}\n");
    const goSelection = { row: { row_id: "fixture.go", runner: "go", selector: { package: "./internal/fixture", tests: ["TestCurrent"] } }, runner: { approved_roots: ["internal"] }, taskSurfaceCommandIDs: new Set() };
    const resolveGo = createRowSelectorResolver(directory);
    assert.deepEqual(resolveGo(goSelection), ["go:./internal/fixture:TestCurrent"]);
    assert.deepEqual(resolveGo(goSelection), ["go:./internal/fixture:TestCurrent"]);

    for (const expression of ["cgo", "amd64.v2", "race", "private"]) {
      writeFileSync(source, `//go:build ${expression}\n\npackage fixture\nfunc TestCurrent() {}\n`);
      assert.throws(() => createRowSelectorResolver(directory)(goSelection), /excluded from the Go runner build context/u);
    }
    writeFileSync(source, "//go:build linux && amd64.v1 && gc && !cgo && !race\n\npackage fixture\nfunc TestCurrent() {}\n");
    assert.deepEqual(createRowSelectorResolver(directory)(goSelection), ["go:./internal/fixture:TestCurrent"]);
    for (const expression of ["goexperiment.jsonv2", "!goexperiment.jsonv2", "linux || goexperiment.jsonv2"]) {
      writeFileSync(source, `//go:build ${expression}\n\npackage fixture\nfunc TestCurrent() {}\n`);
      assert.throws(() => createRowSelectorResolver(directory)(goSelection), /unsupported experimental Go build context/u);
    }
    writeFileSync(source, "//go:build go1.27\n\npackage fixture\nfunc TestCurrent() {}\n");

    // A new load must observe pin changes, not a process-global build context.
    writeFileSync(pins, JSON.stringify({ go_version: "1.26.0" }));
    assert.throws(() => createRowSelectorResolver(directory)(goSelection), /excluded from the Go runner build context/u);
    writeFileSync(pins, JSON.stringify({ go_version: "1.27.2" }));
    writeFileSync(path.join(packageRoot, "duplicate_test.go"), "package fixture\nfunc TestCurrent() {}\n");
    assert.throws(() => createRowSelectorResolver(directory)(goSelection), /must resolve exactly once/u);
    rmSync(path.join(packageRoot, "duplicate_test.go"));
    writeFileSync(source, "package fixture\nfunc TestReplacement() {}\n");
    assert.throws(() => createRowSelectorResolver(directory)(goSelection), /must resolve exactly once/u);

    // Reusing facts must not bypass path checks on a later selector.
    renameSync(packageRoot, `${packageRoot}-original`);
    symlinkSync(`${packageRoot}-original`, packageRoot);
    assert.throws(() => resolveGo(goSelection), /symbolic link/u);

    const main = path.join(webDirectory, "scope.test.ts"), suite = path.join(webDirectory, "suite/register.ts");
    mkdirSync(path.dirname(suite));
    writeFileSync(main, 'import { registerFixtureSuite } from "./suite/register";\nregisterFixtureSuite();\n');
    writeFileSync(suite, 'export function registerFixtureSuite() { it("before", () => {}); }\n');
    const webSelection = (title) => ({ row: { row_id: "fixture.web", runner: "vitest", selector: { file: path.relative(directory, main), titles: [title] } }, runner: { approved_roots: ["apps/web"] }, taskSurfaceCommandIDs: new Set() });
    const resolveWeb = createRowSelectorResolver(directory);
    assert.deepEqual(resolveWeb(webSelection("before")), [`vitest::${path.relative(directory, main)}:before`]);
    // Only the imported suite changes. Main-source-keyed global caches are stale.
    writeFileSync(suite, 'export function registerFixtureSuite() { it("after", () => {}); }\n');
    assert.deepEqual(createRowSelectorResolver(directory)(webSelection("after")), [`vitest::${path.relative(directory, main)}:after`]);
    assert.throws(() => createRowSelectorResolver(directory)(webSelection("before")), /must resolve exactly once/u);
    for (const calls of ["", "registerFixtureSuite(); registerFixtureSuite();", "if (enabled) registerFixtureSuite();", "registerFixtureSuite(options);"]) {
      writeFileSync(main, `import { registerFixtureSuite } from "./suite/register";\n${calls}`);
      assert.throws(() => createRowSelectorResolver(directory)(webSelection("after")), /exactly one unconditional/u);
    }
    writeFileSync(main, 'import { registerFixtureSuite as install } from "./suite/register"; install();');
    assert.deepEqual(createRowSelectorResolver(directory)(webSelection("after")), [`vitest::${path.relative(directory, main)}:after`]);
    writeFileSync(suite, 'export function wrongExport() { it("after", () => {}); }');
    assert.throws(() => createRowSelectorResolver(directory)(webSelection("after")), /indirect test registration|missing registration suite/u);
    rmSync(main); symlinkSync(suite, main);
    assert.throws(() => resolveWeb(webSelection("before")), /symbolic link/u);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
  assertSourceDiscoveryBoundary();
}

function assertSourceDiscoveryBoundary() {
  const root = mkdtempSync(path.join(tmpdir(), "cartulary-source-discovery."));
  const write = (relative, source) => {
    mkdirSync(path.dirname(path.join(root, relative)), { recursive: true });
    writeFileSync(path.join(root, relative), source);
  };
  const go = { runner: "go", approved_roots: ["internal"] };
  const web = { runner: "vitest", approved_roots: ["apps/web"] };
  const selection = { row: { row_id: "fixture.web", runner: "vitest", selector: { file: "apps/web/src/fixture.test.ts", project_id: "browser-unit", titles: ["suite row alpha", "suite row beta"] } }, runner: web, taskSurfaceCommandIDs: new Set() };
  try {
    write("tools/generated_artifact_policy.json", JSON.stringify({ generated_roots: [{ path: "internal/projected" }], generated_files: [{ path: "internal/fixture/generated_test.go" }] }));
    write("tools/toolchain_pins.json", JSON.stringify({ go_version: "1.27.2" }));
    write("internal/fixture/new_test.go", "package fixture\nfunc TestNew() {}\nfunc TestMain() {}\n");
    symlinkSync("/unreachable-discovery-sentinel", path.join(root, "internal/fixture/generated_test.go"));
    write("internal/fixture/excluded_windows_test.go", "package fixture\nfunc TestWindows() {}\n");
    write("internal/fixture/_ignored_test.go", "package fixture\nfunc TestIgnored() {}\n");
    write("internal/_ignored/unreachable_test.go", "package fixture\nfunc TestIgnoredPackage() {}\n");
    write("apps/web/src/fixture.test.ts", 'const rows = ["alpha", "beta"] as const satisfies readonly string[]; describe("suite", () => { it.each(rows)("row %s", () => {}); });');
    for (const excluded of ["docs", "README.md", "node_modules", "vendor", "generated", "projected"]) {
      symlinkSync("/unreachable-discovery-sentinel", path.join(root, "internal", excluded));
    }
    for (const excluded of ["dist", "dist-measurement", ".frontend-publish-fixture"]) {
      symlinkSync("/unreachable-discovery-sentinel", path.join(root, "apps/web", excluded));
    }
    assert.deepEqual(discoverRunnerSources(root, web), ["apps/web/src/fixture.test.ts"]);
    assert.deepEqual(discoverRunnerSources(root, go), ["internal/fixture/excluded_windows_test.go", "internal/fixture/new_test.go"]);
    const runners = new Map([["go", go], ["vitest", web]]);
    const resolver = createRowSelectorResolver(root);
    const keys = resolver(selection);
    assert.deepEqual(keys, ["vitest:browser-unit:apps/web/src/fixture.test.ts:suite row alpha", "vitest:browser-unit:apps/web/src/fixture.test.ts:suite row beta"]);
    assert.throws(() => resolver.assertSourceCoverage(runners, new Set(keys)), /go:.\/internal\/fixture:TestNew/u);
    const selected = new Set([...keys, "go:./internal/fixture:TestNew"]);
    resolver.assertSourceCoverage(runners, selected);
    write("internal/fixture/unicode_test.go", "package fixture\nfunc Test界() {}\n");
    assert.throws(() => createRowSelectorResolver(root).assertSourceCoverage(runners, selected), /unsupported Go test identity/u);
    rmSync(path.join(root, "internal/fixture/unicode_test.go"));
    write("internal/fixture/bare_test.go", "package fixture\nfunc Test() {}\nfunc TestlowerHelper() {}\n");
    assert.throws(() => createRowSelectorResolver(root).assertSourceCoverage(runners, selected), /go:.\/internal\/fixture:Test(?:\n|$)/u);
    rmSync(path.join(root, "internal/fixture/bare_test.go"));
    selected.delete(keys[1]);
    assert.throws(() => resolver.assertSourceCoverage(runners, selected), /suite row beta/u);

    const old = fs.lstatSync;
    fs.lstatSync = (file, ...args) => {
      if (String(file).includes("/docs") || String(file).endsWith(".md")) throw new Error("prohibited metadata read");
      return old(file, ...args);
    };
    syncBuiltinESMExports();
    try {
      discoverRunnerSources(root, go);
      // Negative selector values belong to the external scratch repository;
      // they are not executable input references to repository documentation.
      for (const file of [path.posix.join("apps/web", "docs", "missing.test.ts"), path.posix.join("apps/web", "README.md"), "../outside.test.ts"]) {
        assert.throws(() => resolver({ ...selection, row: { ...selection.row, selector: { ...selection.row.selector, file } } }), /excluded source|traversal/u);
      }
    } finally { fs.lstatSync = old; syncBuiltinESMExports(); }

    for (const source of ['it(runtimeTitle(), () => {});', 'it.each(runtimeRows())("row %s", () => {});', 'describe(runtimeTitle(), () => { it("child", () => {}); });']) {
      write("apps/web/src/fixture.test.ts", source);
      assert.throws(() => createRowSelectorResolver(root)(selection), /unresolved/u);
    }
    const resolveTitles = (source, titles) => {
      write("apps/web/src/fixture.test.ts", source);
      return createRowSelectorResolver(root)({ ...selection, row: { ...selection.row, selector: { ...selection.row.selector, titles } } });
    };
    assert.deepEqual(resolveTitles('import { it as check, describe as group } from "vitest"; const prefix = "outer"; const title = prefix; group("suite", () => { const prefix = "inner"; check(title, () => {}); });', ["suite outer"]), ["vitest:browser-unit:apps/web/src/fixture.test.ts:suite outer"]);
    assert.deepEqual(resolveTitles('import { it as check } from "vitest"; const another = check; another("aliased", () => {});', ["aliased"]), ["vitest:browser-unit:apps/web/src/fixture.test.ts:aliased"]);
    assert.deepEqual(resolveTitles('it.each([["12.5units", "$untouched"]])("%i %s %# %$", () => {});', ["12 $untouched 0 1"]), ["vitest:browser-unit:apps/web/src/fixture.test.ts:12 $untouched 0 1"]);
    for (const source of [
      'if (enabled) it("hidden", () => {});',
      'function registerHidden() { it("hidden", () => {}); }',
      '[1, 2].forEach(() => it("hidden", () => {}));',
      'it.skipIf(enabled)("hidden", () => {});',
      'it.for([1])("hidden", () => {});',
      'it.each([1])("hidden %o", () => {});',
      'it.each([1])("hidden %% %s", () => {});',
      'it.each([{ value: "known" }])("hidden $missing", () => {});',
      'let title = "hidden"; it(title, () => {});',
      'const it = unknownFactory(); it("hidden", () => {});',
      'it["each"]([1])("hidden", () => {});',
      'const { only: check } = it; check("hidden", () => {});',
      'const holder = { register: it }; holder.register("hidden", () => {});',
      'registerUnknown(it);',
    ]) assert.throws(() => resolveTitles(source, ["hidden"]), /unsupported|unresolved|runner aliases/u);
    write("apps/web/src/fixture.test.ts", 'it("duplicate", () => {}); it("duplicate", () => {});');
    assert.throws(() => createRowSelectorResolver(root)({ ...selection, row: { ...selection.row, selector: { ...selection.row.selector, titles: ["duplicate"] } } }), /exactly once/u);
    const browserFile = "apps/web/e2e/fixture.spec.ts";
    const browserRunner = { runner: "playwright", approved_roots: ["apps/web/e2e"], project_ids: ["chromium"], stages: ["measurement"] };
    const browserSelection = { row: { row_id: "fixture.browser", runner: "playwright", selector: { file: browserFile, project_id: "chromium", stage: "measurement", titles: ["outer same"], scenario_ids: ["fixture_case"] } }, runner: browserRunner, taskSurfaceCommandIDs: new Set() };
    const resolveBrowser = (source, title = "outer same") => {
      write(browserFile, `import { test } from "@playwright/test"; ${source}`);
      return createRowSelectorResolver(root)({ ...browserSelection, row: { ...browserSelection.row, selector: { ...browserSelection.row.selector, titles: [title] } } });
    };
    assert.deepEqual(resolveBrowser('test.describe("outer", () => { test("same", { tag: "@fast" }, () => {}); });'), [
      "playwright:chromium:measurement:fixture_case", `playwright-source:chromium:${browserFile}:outer same`,
    ]);
    for (const title of [" leading", "trailing ", "line\nbreak", "delimiter › component"]) {
      assert.throws(() => resolveBrowser(`test(${JSON.stringify(title)}, () => {});`, title), /test-list interface/u);
    }
    assert.throws(() => resolveBrowser('test("outer", () => {}); test.describe("outer", () => { test("child", () => {}); });', "outer"), /suite prefix/u);
    write("apps/web/e2e/suites/register.ts", 'import { test } from "@playwright/test"; export function registerSharedSuite() { test("shared", () => {}); }');
    assert.throws(() => resolveBrowser('import { registerSharedSuite } from "./suites/register"; registerSharedSuite();', "shared"), /imported registration suites/u);
    symlinkSync("fixture", path.join(root, "internal/linked"));
    assert.throws(() => discoverRunnerSources(root, go), /symbolic link/u);
    assert.throws(() => discoverRunnerSources(root, { ...go, approved_roots: ["internal/../outside"] }), /invalid discovery root/u);
  } finally { rmSync(root, { recursive: true, force: true }); }
}

export function assertPostgresCatalogClosure(context) {
  const { rows, postgresFixturePolicy: actual } = context.catalogMetadata;
  const active = rows.filter((row) => row.status === "active");
  const classified = new Set();
  for (const policy of actual.registry.capability_policies) {
    const members = active.filter((row) => row.fixture_capability === policy.capability);
    assert.equal(actual.counts[policy.capability], members.length, policy.capability);
    for (const row of members) {
      assert.equal(classified.has(row.row_id), false, row.row_id);
      classified.add(row.row_id);
    }
  }
  assert.deepEqual([...classified].sort(), active
    .filter((row) => row.fixture_capability.startsWith("postgres_"))
    .map((row) => row.row_id).sort());
  assert.equal(actual.row_count, classified.size);
  assert.deepEqual(actual.registry.transaction_row_approvals, active
    .filter((row) => row.fixture_capability === "postgres_transaction")
    .map((row) => row.row_id).sort());
}

export function assertPostgresPolicyFixtures(context) {
  const base = context.readJSON("tools/postgres_fixture_policy_registry.json");
  const directory = mkdtempSync(path.join(tmpdir(), "cartulary-postgres-policy."));
  const policyPath = path.join(directory, "policy.json");
  const row = (name, capability, status = "active") => ({
    row_id: `harness.test_catalog.fixture.${name}`,
    fixture_capability: capability,
    status,
    service_dependencies: capability === "none" ? [] : ["postgres"],
  });
  const transaction = row("transaction", "postgres_transaction");
  const rows = [
    row("dedicated", "postgres_dedicated"),
    row("migration", "postgres_migration"), transaction,
    row("inactive", "postgres_dedicated", "inactive"),
    row("browser", "browser_stack"), row("none", "none"),
  ];
  const validate = (input, approvals = [transaction.row_id]) => {
    writeFileSync(policyPath, JSON.stringify({ ...base, transaction_row_approvals: approvals }));
    return validatePostgresFixturePolicy(context.root, input, { policyPath });
  };
  try {
    assert.deepEqual(validate(rows).counts, {
      postgres_dedicated: 1, postgres_migration: 1, postgres_transaction: 1,
    });
    assert.equal(validate(rows).row_count, 3);
    const grown = [...rows, row("another_dedicated", "postgres_dedicated")];
    assert.deepEqual(validate(grown).counts, {
      postgres_dedicated: 2, postgres_migration: 1, postgres_transaction: 1,
    });
    assert.deepEqual(validate([], []).counts, {
      postgres_dedicated: 0, postgres_migration: 0, postgres_transaction: 0,
    });
    assert.throws(() => validate([{ ...rows[0], service_dependencies: [] }], []),
      /lacks the postgres service/u);
    assert.throws(() => validate(rows, []), /must exactly cover current transaction rows/u);
    assert.throws(() => validate(rows, ["harness.test_catalog.fixture.extra", transaction.row_id]),
      /must exactly cover current transaction rows/u);
    assert.throws(() => validate(rows, [transaction.row_id, transaction.row_id]),
      /duplicate|unique/u);
    const other = row("another_transaction", "postgres_transaction");
    assert.throws(() => validate([...rows, other], [transaction.row_id, other.row_id]),
      /ASCII-sorted/u);
    assert.deepEqual(validate([...rows, other], [other.row_id, transaction.row_id]).counts, {
      postgres_dedicated: 1, postgres_migration: 1, postgres_transaction: 2,
    });
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}

export function assertFixtureBuilderClosure(context) {
  const { catalog, fixtureBuilderPolicy } = context;
  const activeProfiles = [...catalog.fixtureProfiles.profiles.values()]
    .filter((profile) => profile.status === "active");
  assert.deepEqual([...fixtureBuilderPolicy.byFixtureProfileID.keys()].sort(),
    activeProfiles.map((profile) => profile.fixture_profile_id).sort());
  const groups = groupRowsByPerformanceFixture(context.root, catalog.rows, {
    registry: catalog.fixtureProfiles,
  });
  const migrationDigest = postgresMigrationDigest(context.root);
  // Each group and their union exercise O(groups) plans, independent of row count.
  const selections = [...groups.map((group) => [group]), ...(groups.length ? [groups] : [])];
  for (const selected of selections) {
    const graph = context.compiler.compile({ kind: "rows", row_ids: selected.flatMap((group) => group.row_ids) });
    const builders = graph.units.filter((unit) => unit.kind === "fixture_builder");
    assert.deepEqual(builders.map((unit) => unit.snapshot_key).sort(),
      selected.map((group) => snapshotKey(group.profile, migrationDigest)).sort());
    assert.equal(new Set(builders.map((unit) => unit.unit_id)).size, builders.length);
  }
  const unprofiled = catalog.rows.find((row) => !row.fixture_profile_id && row.fixture_capability === "none");
  assert.ok(unprofiled);
  const plain = context.compiler.compile({ kind: "rows", row_ids: [unprofiled.row_id] });
  assert.equal(plain.units.some((unit) => unit.kind === "fixture_builder"), false);
}
