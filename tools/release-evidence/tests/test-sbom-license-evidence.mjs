#!/usr/bin/env node
import assert from "node:assert/strict";
import "./test-license-inputs.mjs";
import { spawnSync } from "node:child_process";
import fs, { existsSync, lstatSync, mkdirSync, readFileSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { enforcePrivateProcessUmask } from "../../harness/runtime/private-process.mjs";
import { removePrivateTree } from "../../harness/runtime/secure-local-files.mjs";
import { validateFile } from "../validate-cyclonedx.mjs";

import {
  collectGoEvidence,
  runInventoryCommand,
  withInventoryWorkspace,
  goTestDependencyListArgs,
  goInventoryDependencies,
  licenseReviewFlags,
  makeCycloneDxBom,
  parseGoModuleGraph,
  parseJSONStream,
} from "../generate-sbom-license-evidence.mjs";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
enforcePrivateProcessUmask();

function assertPrivateTree(directory) {
  assert.equal(lstatSync(directory).mode & 0o777, 0o700);
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const file = path.join(directory, entry.name);
    if (entry.isDirectory()) assertPrivateTree(file);
    else {
      assert.ok(entry.isFile(), "workspace contains only regular files and directories");
      assert.equal(lstatSync(file).mode & 0o777, 0o600);
    }
  }
}

test("canonical CycloneDX version admission precedes every schema probe", (t) => {
  const scratch = createScratch(t, "cartulary-schema-admission-");
  const input = path.join(scratch, "bom.json");
  const originals = Object.fromEntries(["existsSync", "statSync", "lstatSync", "openSync", "readFileSync"].map(name => [name, fs[name]]));
  const invalid = ["1.6", "1.8", 1.7, {}, [], true, "../../private-input", "../untrusted/schema.json", "1.7/../../../elsewhere", "x".repeat(100000)];
  for (const version of [...invalid, undefined, null, ""]) {
    writeFileSync(input, JSON.stringify({ specVersion: version }));
    const probes = [];
    try {
      for (const [name, original] of Object.entries(originals)) fs[name] = (file, ...args) => {
        probes.push([name, file]);
        // Reject before the original API can stat/open any supplied or schema path.
        assert.equal(file, input, `unexpected ${name} before version admission`);
        return original(file, ...args);
      };
      syncBuiltinESMExports();
      assert.throws(() => validateFile(input), error => {
        assert.match(error.message, version === undefined || version === null || version === "" ? /missing specVersion/u : /unsupported CycloneDX specVersion \(expected 1.7\)/u);
        assert.ok(error.message.length < 512, "untrusted version text must not enter diagnostics");
        return true;
      });
      assert.ok(probes.length > 0);
      assert.ok(probes.every(([, file]) => file === input));
    } finally {
      Object.assign(fs, originals); syncBuiltinESMExports();
    }
  }
  writeFileSync(input, JSON.stringify({ specVersion: "1.7", bomFormat: "invalid" }));
  assert.throws(() => validateFile(input), /CycloneDX 1.7 validation failed/u);
});

test("scanner streams admit only complete objects and graph records", () => {
  for (const empty of ["", " \t\r\n"]) {
    assert.deepEqual(parseJSONStream(empty), []);
    assert.deepEqual(parseGoModuleGraph(empty), []);
  }
  assert.deepEqual(parseJSONStream(String.raw`{"value":"} { \" \\" ,"nested":[{"ok":true},null]}{"next":2}`), [
    { value: '} { " \\', nested: [{ ok: true }, null] }, { next: 2 },
  ]);
  assert.deepEqual(parseGoModuleGraph("main go@1.27.3\r\n\tdependency@v1.2.3\tchild@v2.3.4 \n"), [
    ["main", "go@1.27.3"], ["dependency@v1.2.3", "child@v2.3.4"],
  ]);
  const invalidObjects = [
    "{", '{}{"unfinished":', '{"text":"escape\\', '{"array":[}', '{"x":1,}',
    '}{"ok":1}', '{}garbage', 'garbage{}', '{}"extra"', '[]', 'null', '1', '{}[]',
    '{}\u00a0{}', '{"x":1,"x":2}', '{"nested":{"x":1,"x":2}}',
  ];
  const invalidGraphs = ["single", "a b extra", "a b\npartial", "a\u0000 b", "a b\rgarbage", "a\u00a0b", "\u000b"];
  for (const [parser, format, vectors] of [[parseJSONStream, "JSON stream", invalidObjects], [parseGoModuleGraph, "Go module graph", invalidGraphs]]) {
    for (const input of vectors) {
      assert.throws(() => parser(input), error => {
        assert.equal(error.failure_reason, "tool_diagnostic_failure");
        assert.match(error.message, new RegExp(`^invalid inventory ${format} at offset [0-9]+$`, "u"));
        return true;
      }, `must reject ${JSON.stringify(input)}`);
    }
  }
});

test("release inventory helpers and isolated Go discovery", (t) => {
  const tmp = createScratch(t, "cartulary-sbom-test-");
  assert.deepEqual(licenseReviewFlags("GPL-3.0-only", null), ["copyleft_review", "legal_review"]);
  assert.deepEqual(licenseReviewFlags("MIT", null), ["notice_or_attribution_review"]);

  assert.deepEqual(parseJSONStream('{"a":1}\n{"b":{"c":2}}'), [{ a: 1 }, { b: { c: 2 } }]);
  assert.deepEqual(parseGoModuleGraph("a@v1 b@v2\nb@v2 c@v3\n"), [
    ["a@v1", "b@v2"],
    ["b@v2", "c@v3"],
  ]);

  const goListArgs = goTestDependencyListArgs();
  assert.equal(goListArgs.includes("./..."), false);
  assert.deepEqual(goListArgs.slice(-4), [
    "./cmd/...",
    "./db/...",
    "./internal/...",
    "./tools/...",
  ]);
  const discoveryRoot = path.join(tmp, "discovery");
  mkdirSync(discoveryRoot);
  const pins = JSON.parse(readFileSync(path.join(repoRoot, "tools/toolchain_pins.json"), "utf8"));
  writeFileSync(path.join(discoveryRoot, "go.mod"), `module example.test/inventory\n\ngo ${pins.go_version}\n`);
  const includedRoots = ["cmd", "db", "internal", "tools"];
  for (const root of includedRoots) {
    mkdirSync(path.join(discoveryRoot, root));
    writeFileSync(path.join(discoveryRoot, root, "fixture.go"), "package fixture\n");
    writeFileSync(path.join(discoveryRoot, root, "fixture_test.go"), "package fixture\nimport \"testing\"\nfunc TestFixture(t *testing.T) {}\n");
  }
  for (const root of ["tmp/transient", ".cartulary/test-results", "runtime", "cache"]) {
    mkdirSync(path.join(discoveryRoot, root), { recursive: true });
    writeFileSync(path.join(discoveryRoot, root, "broken.go"), "not a Go package\n");
  }
  const discover = () => spawnSync(process.env.GO || "go", goListArgs, {
    cwd: discoveryRoot,
    env: {
      ...process.env,
      GOFLAGS: "-mod=readonly",
      GOWORK: "off",
      GOTOOLCHAIN: pins.go_toolchain,
      ...(process.env.GO_CACHE_DIR ? { GOCACHE: process.env.GO_CACHE_DIR } : {}),
      ...(process.env.GO_MOD_CACHE_DIR ? { GOMODCACHE: process.env.GO_MOD_CACHE_DIR } : {}),
      ...(process.env.GO_TMP_DIR ? { GOTMPDIR: process.env.GO_TMP_DIR } : {}),
    },
    encoding: "utf8",
    maxBuffer: 1024 * 1024,
  });
  const stableList = discover();
  assert.ifError(stableList.error);
  assert.equal(
    stableList.status,
    0,
    `stable Go package discovery entered a repository transient root: ${stableList.stderr}`,
  );
  assert.match(stableList.stdout, /example\.test\/inventory@/u);
  for (const root of includedRoots) {
    const broken = path.join(discoveryRoot, root, "broken.go");
    writeFileSync(broken, "not a Go package\n");
    const rejected = discover();
    assert.ifError(rejected.error);
    assert.notEqual(rejected.status, 0, `${root} must be discovered`);
    assert.ok(rejected.stderr.includes(`${root}/broken.go`), rejected.stderr);
    rmSync(broken);
  }
  const validBom = path.join(tmp, "valid.cyclonedx.json");
  const bom = makeCycloneDxBom({
    name: "fixture",
    version: "0.0.0",
    bomRefValue: "fixture:application",
    components: [],
    dependencies: [],
    tools: [{ type: "application", name: "fixture-tool", version: "0.0.0" }],
    semanticInputDigest: `sha256:${"a".repeat(64)}`,
    completeness: "complete",
  });
  assert.equal(bom.serialNumber, "urn:uuid:aaaaaaaa-aaaa-5aaa-8aaa-aaaaaaaaaaaa");
  assert.equal(Object.hasOwn(bom.metadata, "timestamp"), false);
  assert.equal(
    bom.metadata.component.properties.find((entry) => entry.name === "cartulary:semantic_input_digest")?.value,
    `sha256:${"a".repeat(64)}`,
  );
  writeFileSync(validBom, `${JSON.stringify(bom, null, 2)}\n`);
  const valid = spawnSync(process.execPath, [path.join(repoRoot, "tools", "release-evidence", "validate-cyclonedx.mjs"), validBom], {
    cwd: repoRoot,
    encoding: "utf8",
  });
  assert.equal(valid.status, 0, valid.stderr);

  const invalidBom = path.join(tmp, "invalid.cyclonedx.json");
  writeFileSync(invalidBom, "{}\n");
  const invalid = spawnSync(process.execPath, [path.join(repoRoot, "tools", "release-evidence", "validate-cyclonedx.mjs"), invalidBom], {
    cwd: repoRoot,
    encoding: "utf8",
  });
  assert.notEqual(invalid.status, 0);
  assert.match(invalid.stderr, /missing specVersion/);
});

function moduleFixture(t) {
  const scratch = createScratch(t, "cartulary-go-inventory-");
  const root = path.join(scratch, "source");
  const modulePath = "github.com/JochiRaider/cartulary";
  const pins = JSON.parse(readFileSync(path.join(repoRoot, "tools/toolchain_pins.json"), "utf8"));
  const selections = [".", "referencepacks/format", "referencepacks/attack"];
  for (const directory of selections) {
    mkdirSync(path.join(root, directory), { recursive: true });
    writeFileSync(path.join(root, directory, "go.mod"), `module ${modulePath}${directory === "." ? "" : `/${directory}`}\n\ngo ${pins.go_version}\n`);
    writeFileSync(path.join(root, directory, "go.sum"), "");
  }
  mkdirSync(path.join(root, "tools"));
  writeFileSync(path.join(root, "tools/go_modules.json"), JSON.stringify({
    schema_id: "cartulary.go_modules.v1",
    modules: selections.slice(1).map(directory => ({ root: directory, module_path: `${modulePath}/${directory}`, version: "v0.1.0", kind: "reference_pack_producer" })),
  }));
  const manifests = selections.flatMap(directory => ["go.mod", "go.sum"].map(file => path.join(root, directory, file)));
  const original = manifests.map(file => readFileSync(file, "utf8"));
  return { scratch, root, modulePath, selections,
    collect(go, inspect = () => {}) {
      return withInventoryWorkspace(async ctx => {
        ctx.config.go = go;
        const evidence = await collectGoEvidence(ctx, root);
        inspect(ctx);
        return evidence;
      });
    },
    assertUnchanged() { assert.deepEqual(manifests.map(file => readFileSync(file, "utf8")), original); },
  };
}

test("Go inventory visits each admitted independent module with real Go and leaves manifests unchanged", async (t) => {
  const f = moduleFixture(t);
  for (const directory of ["cmd/server", "cmd/migrate", "cmd/operator", "db", "internal", "tools", "referencepacks/format", "referencepacks/attack"]) {
    mkdirSync(path.join(f.root, directory), { recursive: true });
    writeFileSync(path.join(f.root, directory, "fixture.go"), "package fixture\n");
    writeFileSync(path.join(f.root, directory, "fixture_test.go"), "package fixture\nimport \"testing\"\nfunc TestFixture(t *testing.T) {}\n");
  }
  for (const directory of ["tmp/broken", "referencepacks/unregistered"]) {
    mkdirSync(path.join(f.root, directory), { recursive: true });
    writeFileSync(path.join(f.root, directory, "broken.go"), "not a Go package\n");
  }
  const result = await f.collect(process.env.GO || "go", ctx => assertPrivateTree(ctx.workspace));
  assert.deepEqual([...result.mainPaths].sort(), [f.modulePath, `${f.modulePath}/referencepacks/attack`, `${f.modulePath}/referencepacks/format`]);
  assert.deepEqual(result.modules, []);
  for (const name of result.mainPaths) assert.ok(result.testSet.has(`${name}@`), `${name} test imports are observed`);
  assert.deepEqual([...result.runtimeSet], [`${f.modulePath}@`]);
  f.assertUnchanged();
  writeFileSync(path.join(f.root, "referencepacks/format/fixture.go"), "not a Go package\n");
  await assert.rejects(f.collect(process.env.GO || "go"), /test deps failed with exit code/u);
  f.assertUnchanged();
});

test("Go inventory merges exact versions, directness and origins without collapsing first-party dependency edges", async (t) => {
  const f = moduleFixture(t);
  const main = f.modulePath, format = `${main}/referencepacks/format`, attack = `${main}/referencepacks/attack`;
  const shared = "example.test/shared", unique = "example.test/pack-only";
  const dep = (Path, Version, Indirect = false) => ({ Path, Version, Indirect });
  const config = path.join(f.scratch, "responses.json");
  const responses = {
    [main]: { modules: [dep(shared, "v1.0.0", true), dep(format, "v0.1.0")], tests: [`${shared}@v1.0.0`, `${format}@v0.1.0`],
      graph: [[main, `${shared}@v1.0.0`], [main, `${format}@v0.1.0`], [`${format}@v0.1.0`, `${shared}@v1.0.0`]], runtime: [`${shared}@v1.0.0`] },
    [attack]: { modules: [dep(shared, "v1.0.0"), dep(unique, "v2.0.0")], tests: [`${unique}@v2.0.0`],
      graph: [[attack, `${shared}@v1.0.0`], [attack, `${unique}@v2.0.0`], [`${format}@v0.1.0`, `${shared}@v1.0.0`]], runtime: [] },
    [format]: { modules: [dep(shared, "v1.1.0", true)], tests: [`${shared}@v1.1.0`],
      graph: [[format, `${shared}@v1.1.0`], ["example.test/unselected@v0.0.1", `${shared}@v1.1.0`], [format, "go@1.27.3"]], runtime: [] },
  };
  writeFileSync(config, JSON.stringify(responses));
  const calls = path.join(f.scratch, "calls.jsonl");
  writeFileSync(calls, "");
  const fakeGo = path.join(f.scratch, "go");
  writeFileSync(fakeGo, `#!${process.execPath}
import { appendFileSync, readFileSync } from "node:fs";
if (process.env.GOWORK !== "off") throw new Error("ambient workspaces forbidden");
const args = process.argv.slice(2);
appendFileSync(${JSON.stringify(calls)}, JSON.stringify(args) + "\\n");
if (args[0] === "version") { console.log("go version fixture"); process.exit(0); }
const name = readFileSync("go.mod", "utf8").match(/^module (.+)$/m)[1];
const all = JSON.parse(readFileSync(${JSON.stringify(config)}, "utf8")), data = all[name];
if (all.fail === name && args.includes("-test")) { console.error("selected module failed"); process.exit(19); }
if (args[0] === "mod" && args[1] === "download") {
  appendFileSync("go.sum", "download must only alter the private copy\\n");
  for (const module of data.modules) console.log(JSON.stringify(module));
} else if (args.includes("-m")) {
  for (const module of [{ Path: name, Main: true }, ...data.modules]) console.log(JSON.stringify(module));
} else if (args[0] === "mod" && args[1] === "graph") {
  for (const edge of data.graph) console.log(edge.join(" "));
} else if (args[0] === "list") {
  for (const module of args.includes("-test") ? data.tests : data.runtime) console.log(module);
} else { throw new Error("unexpected Go invocation"); }
`, { mode: 0o700 });
  const evidence = await f.collect(fakeGo, ctx => assertPrivateTree(ctx.workspace));
  assert.equal(readFileSync(calls, "utf8").trim().split("\n").length, 14, "one version query, four queries per admitted module and one runtime query");
  assert.deepEqual(evidence.modules.map(module => [module.Path, module.Version, module.Indirect]), [
    [unique, "v2.0.0", false], [shared, "v1.0.0", false], [shared, "v1.1.0", true], [format, "v0.1.0", false],
  ]);
  assert.deepEqual([...evidence.downloads.keys()].sort(), [`${unique}@v2.0.0`, `${shared}@v1.0.0`, `${shared}@v1.1.0`, `${format}@v0.1.0`]);
  assert.deepEqual([...evidence.testSet].sort(), [`${unique}@v2.0.0`, `${shared}@v1.0.0`, `${shared}@v1.1.0`, `${format}@v0.1.0`]);
  assert.deepEqual([...evidence.runtimeSet], [`${shared}@v1.0.0`]);
  const sharedOrigins = evidence.modules.find(module => module.Path === shared && module.Version === "v1.0.0").inventoryEvidence;
  assert.deepEqual(sharedOrigins, ["go -C . list -deps -test ./cmd/... ./db/... ./internal/... ./tools/...", "go -C . list -m -json all", "go -C . mod graph",
    "go -C referencepacks/attack list -m -json all", "go -C referencepacks/attack mod graph", "go.mod", "go.sum", "referencepacks/attack/go.mod", "referencepacks/attack/go.sum"]);
  assert.ok(evidence.modules.find(module => module.Path === unique).inventoryEvidence.includes("go -C referencepacks/attack list -deps -test ./..."));
  const graph = goInventoryDependencies(evidence, new Map([
    [`go:${unique}@v2.0.0`, {}], [`go:${shared}@v1.0.0`, {}], [`go:${shared}@v1.1.0`, {}], [`go:${format}@v0.1.0`, {}],
  ]));
  assert.deepEqual([...graph].map(([from, targets]) => [from, [...targets].sort()]).sort(), [
    ["cartulary:go-backend", [`go:${unique}@v2.0.0`, `go:${shared}@v1.0.0`, `go:${shared}@v1.1.0`, `go:${format}@v0.1.0`]],
    [`go:${format}@v0.1.0`, [`go:${shared}@v1.0.0`]],
  ]);
  assert.deepEqual(await f.collect(fakeGo), evidence, "merge is deterministic across private captures");
  f.assertUnchanged();
  writeFileSync(config, JSON.stringify({ ...responses, fail: format }));
  await assert.rejects(f.collect(fakeGo), /test deps failed with exit code 19/u);
  f.assertUnchanged();
  writeFileSync(path.join(f.root, "referencepacks/format/go.mod"), `module ${format}\nreplace example.test/a => ../attack\n`);
  const previousCalls = readFileSync(calls, "utf8");
  await assert.rejects(f.collect(fakeGo), /dependency override mismatch/u);
  assert.equal(readFileSync(calls, "utf8"), previousCalls, "invalid module admission fails before any scanner");
});

test("inventory workspace consumes complete owned output, cleans failures and preserves its borrowed runtime", async () => {
  const parent = process.env.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT;
  assert.ok(parent, "Make supplies the borrowed private runtime");
  let workspace;
  const output = await withInventoryWorkspace(async ctx => {
    workspace = ctx.workspace;
    assert.ok(workspace.startsWith(`${parent}${path.sep}`));
    assertPrivateTree(workspace);
    return runInventoryCommand(ctx, "large complete stream", process.execPath,
      ["-e", 'process.stdout.write("BEGIN" + "x".repeat(512 * 1024) + "END")']);
  });
  assert.equal(output.stdout, `BEGIN${"x".repeat(512 * 1024)}END`, "scanner output is not the diagnostic tail");
  assert.equal(existsSync(workspace), false);
  assert.ok(existsSync(parent));
  for (const [script, reason] of [
    ['process.stderr.write("fixture failure\\n"); process.exit(19)', "tool_diagnostic_failure"],
    ['process.kill(process.pid, "SIGTERM")', "cancelled_or_interrupted"],
  ]) {
    await assert.rejects(withInventoryWorkspace(async ctx => {
      workspace = ctx.workspace;
      await runInventoryCommand(ctx, "failing scanner", process.execPath, ["-e", script]);
    }), error => error.failure_reason === reason);
    assert.equal(existsSync(workspace), false);
    assert.ok(existsSync(parent), "a failed producer cannot dispose the borrowed runtime");
  }
  await assert.rejects(withInventoryWorkspace(() => assert.fail("unbound work must not run"), {}), /requires Make run identity/u);
  await assert.rejects(withInventoryWorkspace(() => assert.fail("unbound runtime must not run"), {
    CARTULARY_TEST_RUN_ID: process.env.CARTULARY_TEST_RUN_ID,
  }), /requires a bound private runtime/u);
  const unsettled = Object.assign(new Error("fixture unresolved writer"), { processSettled: false });
  try {
    await assert.rejects(withInventoryWorkspace(ctx => {
      workspace = ctx.workspace;
      throw unsettled;
    }), error => error === unsettled);
    assert.ok(existsSync(workspace), "unresolved physical ownership forbids working-directory deletion");
    assert.ok(existsSync(parent));
  } finally {
    // This fixture supplied a result flag, not a real writer; native settlement
    // itself is independently qualified by the shared capture contract.
    removePrivateTree(workspace);
  }
});

test("inventory workspace preserves primary and independent cleanup failures", async () => {
  for (const failWork of [false, true]) {
    const primary = new Error("fixture work failed");
    const original = fs.rmSync;
    let failure, cleanup;
    try {
      fs.rmSync = (file, options) => {
        if (String(file).includes("/.deleting-")) throw new Error("fixture cleanup failed");
        return original(file, options);
      };
      syncBuiltinESMExports();
      try {
        await withInventoryWorkspace(() => { if (failWork) throw primary; });
      } catch (error) { failure = error; }
      assert.ok(failure);
      if (failWork) {
        assert.equal(failure, primary);
        assert.equal(failure.cleanupFailures.length, 1);
        cleanup = failure.cleanupFailures[0];
      } else {
        assert.equal(failure.failure_reason, "cleanup_error");
        assert.equal(failure.cause.errors.length, 1);
        cleanup = failure.cause.errors[0];
      }
      assert.match(cleanup.message, /fixture cleanup failed/u);
      assert.ok(existsSync(cleanup.cleanupPath), "failed owned disposal retains the exact quarantined path");
    } finally {
      fs.rmSync = original;
      syncBuiltinESMExports();
      if (cleanup?.cleanupPath) removePrivateTree(cleanup.cleanupPath);
    }
  }
});

test("inventory CLI publishes its validated pair and leaves existing outputs intact on scanner failure", async t => {
  const scratch = createScratch(t, "cartulary-inventory-cli-");
  const fakeGo = path.join(scratch, "go"), fakePnpm = path.join(scratch, "pnpm");
  writeFileSync(fakeGo, `#!${process.execPath}
import { readFileSync } from "node:fs";
if ((process.umask() & 0o077) !== 0o077) throw new Error("private producer umask missing");
if (process.env.INVENTORY_FIXTURE_FAIL === "1") { console.error("fixture scanner failure"); process.exit(19); }
const args = process.argv.slice(2);
const malformed = process.env.INVENTORY_FIXTURE_MALFORMED;
if (malformed === "json" && args.includes("-m")) { process.stdout.write('{"Path":"accepted"}{"Path":'); process.exit(0); }
if (malformed === "graph" && args[0] === "mod" && args[1] === "graph") { console.log("accepted@v1 child@v2\\nmalformed extra tokens"); process.exit(0); }
if (args[0] === "version") console.log("go version inventory-fixture");
else if (args.includes("-m")) console.log(JSON.stringify({ Path: readFileSync("go.mod", "utf8").match(/^module (.+)$/m)[1], Main: true }));
else if (args[0] !== "mod" && args[0] !== "list") throw new Error("unexpected scanner command");
`, { mode: 0o700 });
  writeFileSync(fakePnpm, `#!${process.execPath}
const args = process.argv.slice(2);
if (args[0] === "--version") console.log("inventory-fixture");
else if (args[0] === "list") console.log("[]");
else if (args[0] === "licenses") console.log("{}");
else throw new Error("unused scanners must not be probed");
`, { mode: 0o700 });
  const destination = path.join(scratch, "published");
  const sbomPath = path.join(destination, "sbom.cyclonedx.json"), licensePath = path.join(destination, "license-report.json");
  const environment = { GO: fakeGo, PNPM: fakePnpm, RELEASE_ARTIFACT_DIR: destination, SBOM_ARTIFACT: sbomPath, LICENSE_REPORT_ARTIFACT: licensePath };
  const producer = path.join(repoRoot, "tools/release-evidence/generate-sbom-license-evidence.mjs");
  await withInventoryWorkspace(ctx => runInventoryCommand(ctx, "inventory CLI", process.execPath, [producer], { env: environment }));
  assert.deepEqual(readdirSync(destination).sort(), ["license-report.json", "sbom.cyclonedx.json"]);
  for (const file of [sbomPath, licensePath]) assert.equal(lstatSync(file).mode & 0o777, 0o644);
  const before = [readFileSync(sbomPath, "utf8"), readFileSync(licensePath, "utf8")];
  const [bom, license] = before.map(bytes => JSON.parse(bytes));
  assert.equal(license.schema_id, "cartulary.license_report.v2");
  assert.equal(bom.metadata.component.properties.find(property => property.name === "cartulary:semantic_input_digest").value, license.semantic_input_digest);
  assert.deepEqual(bom.metadata.tools.components.map(tool => tool.name), ["cartulary-release-inventory-generator", "node", "pnpm", "go"]);
  await assert.rejects(withInventoryWorkspace(ctx => runInventoryCommand(ctx, "failed inventory CLI", process.execPath, [producer], {
    env: { ...environment, INVENTORY_FIXTURE_FAIL: "1" },
  })), /fixture scanner failure/u);
  assert.deepEqual([readFileSync(sbomPath, "utf8"), readFileSync(licensePath, "utf8")], before);
  assert.deepEqual(readdirSync(destination).sort(), ["license-report.json", "sbom.cyclonedx.json"]);
  for (const malformed of ["json", "graph"]) {
    await assert.rejects(withInventoryWorkspace(ctx => runInventoryCommand(ctx, "malformed inventory CLI", process.execPath, [producer], {
      env: { ...environment, INVENTORY_FIXTURE_MALFORMED: malformed },
    })), /invalid inventory (?:JSON stream|Go module graph) at offset [0-9]+/u);
    assert.deepEqual([readFileSync(sbomPath, "utf8"), readFileSync(licensePath, "utf8")], before);
    assert.deepEqual(readdirSync(destination).sort(), ["license-report.json", "sbom.cyclonedx.json"]);
  }
});
