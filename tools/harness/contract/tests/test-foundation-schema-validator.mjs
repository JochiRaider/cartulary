#!/usr/bin/env node
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import {
  cpSync,
  mkdirSync,
  mkdtempSync,
  rmSync,
  readFileSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { createRequire } from "node:module";
import vm from "node:vm";
import Ajv2020 from "ajv/dist/2020.js";
import { compileStandaloneValidators, validatorFilename } from "../../generated-artifacts/schema-validator-compiler.mjs";

const testDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testDir, "..", "..", "..", "..");
const ownerRegistry = JSON.parse(readFileSync(path.join(repoRoot, "tools/test_catalog_owner.json"), "utf8"));
const attachments = JSON.parse(readFileSync(path.join(repoRoot, "tools/harness_schema_attachments.json"), "utf8")).attachments;
// Compare the generated closure with the pinned interpreter/compiler using its
// original inline-reference policy. Expected boundary cases are independently
// authored; the production compiler does not calculate their expected results.
const reference = new Ajv2020({ allErrors: true, inlineRefs: true, strict: false, validateFormats: false });
for (const entry of attachments) reference.addSchema(JSON.parse(readFileSync(path.join(repoRoot, entry.path), "utf8")));
const require = createRequire(import.meta.url);
const generated = require("../generated/foundation-schema-validators.cjs").foundationSchemaValidators;
for (const entry of attachments) {
  const baseline = reference.getSchema(entry.schema_id), candidate = generated[entry.schema_id];
  const values = [null, {}, [], true, 0, "", { schema_id: entry.schema_id }, ownerRegistry];
  for (const value of values) {
    const original = JSON.stringify(value);
    assert.equal(candidate(value), baseline(value), entry.schema_id);
    assert.equal(JSON.stringify(value), original, "validation must not mutate inputs");
  }
}
const probes = JSON.parse(readFileSync(path.join(repoRoot, "internal/testutil/harnessdiag/testdata/schema-probes.json"), "utf8"));
// Each repeated reference must share a validator while preserving composition,
// Unicode lengths, escaped JSON pointers and nullable/conditional semantics.
const referenced = probes.map(({ schema, values }) => ({
  schema: { $id: `${schema.$id}.pair`, type: "array", minItems: 2, maxItems: 2,
    prefixItems: [{ $ref: schema.$id }, { $ref: schema.$id }], items: false },
  values: values.map((value) => [value, value]),
}));
const structural = [
  { schema: { $id: "probe.recursive", $defs: { node: { type: "object", required: ["name", "next"],
    properties: { name: { type: "string", minLength: 1, maxLength: 2 }, next: { anyOf: [{ type: "null" }, { $ref: "#/$defs/node" }] } },
    additionalProperties: false } }, $ref: "#/$defs/node" },
    values: [{ name: "😀", next: null }, { name: "a", next: { name: "bc", next: null } }, { name: "😀😀😀", next: null },
      { name: "a", next: { name: "", next: null } }, { name: "a", next: null, extra: true }], expected: [true, true, false, false, false] },
  { schema: { $id: "probe.evaluated", $defs: { left: { properties: { a: { type: "integer" } }, required: ["a"] },
    right: { properties: { b: { type: ["string", "null"] } }, required: ["b"] } }, type: "object",
    allOf: [{ $ref: "#/$defs/left" }, { $ref: "#/$defs/right" }], unevaluatedProperties: false },
    values: [{ a: 1, b: null }, { a: 2, b: "😀" }, { a: 1 }, { a: "1", b: null }, { a: 1, b: null, c: 0 }],
    expected: [true, true, false, false, false] },
  { schema: { $id: "probe.shared", type: "object", required: ["value"],
    properties: { value: { const: { name: "😀" } } }, additionalProperties: false }, values: [] },
  ...["left", "right"].map((side) => ({ schema: { $id: `probe.shared.${side}`, $ref: "probe.shared" },
    values: [{ value: { name: "😀" } }, { value: { name: "x" } }, { value: { name: "😀" }, extra: 1 }], expected: [true, false, false] })),
  ...["a", "b"].map((side) => ({ schema: { $id: `probe.mutual.${side}`, type: "object", required: [side, "next"],
    properties: { [side]: { type: "integer" }, next: { anyOf: [{ type: "null" }, { $ref: `probe.mutual.${side === "a" ? "b" : "a"}` }] } }, additionalProperties: false },
    values: [{ [side]: 1, next: null }, { [side]: 1, next: { [side === "a" ? "b" : "a"]: 2, next: null } },
      { [side]: 1, next: { [side]: 2, next: null } }], expected: [true, true, false] })),
  { schema: { $id: "https://cartulary.test/tree", $dynamicAnchor: "node", type: "object",
    properties: { data: true, children: { type: "array", items: { $dynamicRef: "#node" } } } },
    values: [{ data: 1, children: [{ data: 2 }] }, { data: 1, children: [2] }], expected: [true, false] },
  { schema: { $id: "https://cartulary.test/strict-tree", $dynamicAnchor: "node", $ref: "https://cartulary.test/tree", unevaluatedProperties: false },
    values: [{ data: 1, children: [{ data: 2 }] }, { data: 1, children: [{ extra: 2 }] }, { extra: 2 }], expected: [true, false, false] },
];
const cases = [...probes, ...referenced, ...structural];
const probeReference = new Ajv2020({ allErrors: true, inlineRefs: true, strict: false, validateFormats: false });
for (const { schema } of cases) probeReference.addSchema(schema);
const compiled = compileStandaloneValidators(cases.map(({ schema }) => schema));
// A closed CommonJS loader supplies only generated siblings, all in one realm.
// Loading opposite orders exercises both entry directions of reference cycles.
for (const ordered of [cases, [...cases].reverse()]) {
  const context = vm.createContext({});
  const sources = new Map([...compiled].map(([id, source]) => [`./${validatorFilename(id)}`, source]));
  const cache = new Map();
  const load = (file) => {
    if (cache.has(file)) return cache.get(file).exports;
    assert.ok(sources.has(file), `unexpected runtime dependency: ${file}`);
    const module = { exports: {} }; cache.set(file, module);
    new vm.Script(`(function(module, exports, require) {\n${sources.get(file)}\n})`, { filename: file })
      .runInContext(context)(module, module.exports, load);
    return module.exports;
  };
  for (const { schema, values, expected } of ordered) {
    const validate = load(`./${validatorFilename(schema.$id)}`), baseline = probeReference.getSchema(schema.$id);
    context.validate = validate;
    for (const [index, value] of values.entries()) {
      const accepted = baseline(value);
      if (expected) assert.equal(accepted, expected[index], `${schema.$id}: authored expectation`);
      context.valueJSON = JSON.stringify(value);
      assert.equal(new vm.Script("validate(JSON.parse(valueJSON))").runInContext(context), accepted, `${schema.$id}: standalone parity`);
      assert.equal(validate.errors === null, accepted, `${schema.$id}: errors reset on success`);
    }
  }
  const shared = `./${validatorFilename("probe.shared")}`;
  assert.ok(cache.has(shared), "referenced root loads through the shared module boundary");
  for (const side of ["left", "right"]) assert.ok(sources.get(`./${validatorFilename(`probe.shared.${side}`)}`).includes(`require(${JSON.stringify(shared)})`));
}
const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-foundation-validator-"));
const summaryClosure = ["cartulary.tool_run_summary.v5"].map(validatorFilename);
try {
  const contractRoot = path.join(scratch, "tools", "harness", "contract");
  mkdirSync(path.join(contractRoot, "generated"), { recursive: true });
  for (const relative of [
    "tools/harness/contract/harness-contract.mjs",
    "tools/harness/contract/artifact-writer.mjs",
    "tools/harness/contract/generated/foundation-schema-validators.cjs",
    "tools/harness/contract/generated/schema-validators",
  ]) {
    cpSync(path.join(repoRoot, relative), path.join(scratch, relative), { recursive: true });
  }
  // An unregistered file must not silently become a runtime contract.
  mkdirSync(path.join(scratch, "tools/schemas"), { recursive: true });
  writeFileSync(path.join(scratch, "tools/schemas/cartulary.unregistered.v1.schema.json"), JSON.stringify({ type: "object" }));
  const probe = path.join(scratch, "probe.mjs");
  writeFileSync(
    probe,
    `import { validateSchema, validateSchemaSync } from "./tools/harness/contract/harness-contract.mjs";\n` +
      `import { createRequire } from "node:module";\n` +
      `const require = createRequire(import.meta.url);\n` +
      `const zeroCounts = { steps: 0, tests: 0, failed: 0, non_test: 0, non_test_failed: 0, packages: 0 };\n` +
      `const zeroAccounting = { authoritative: 0, support: 0, raw: 0, tooling_support: 0, unowned_regression: 0, unmapped: 0, authoritative_failed: 0, support_failed: 0, raw_failed: 0, tooling_support_failed: 0, unowned_regression_failed: 0, unmapped_failed: 0, missing: 0 };\n` +
      `validateSchemaSync("cartulary.tool_run_summary.v5", { schema_id: "cartulary.tool_run_summary.v5", target: "bootstrap-node-runtime", command: { cwd: ".", argv: ["make", "bootstrap-node-runtime"], make_target: "bootstrap-node-runtime", env: {} }, status: "pass", exit_code: 0, started_at: "2026-01-01T00:00:00Z", completed_at: "2026-01-01T00:00:00Z", duration_ms: 0, output_mode: "summary", result_root: ".cartulary/test-results", run_id: "foundation", run_root: ".cartulary/test-results/foundation", summary_artifacts: [], log_artifacts: [], work_units: [], evidence_targets: [], helper_units: [], counts: zeroCounts, step_accounting: zeroAccounting, failure_class: null, failure_reason: null, failures: [], slowest: [], warnings: [], rerun_commands: [], scheduler_timing: null, extensions: {} });\n` +
      `const loaded = () => Object.keys(require.cache).filter((file) => file.includes("/schema-validators/")).map((file) => file.split("/").at(-1)).sort();\n` +
      `if (JSON.stringify(loaded()) !== ${JSON.stringify(JSON.stringify(summaryClosure))}) throw Error("summary reference closure mismatch");\n` +
      `const registry = require("./tools/harness/contract/generated/foundation-schema-validators.cjs").foundationSchemaValidators;\n` +
      `if (JSON.stringify(Object.keys(registry).sort()) !== ${JSON.stringify(JSON.stringify(attachments.map((entry) => entry.schema_id).sort()))}) throw Error("incomplete registered schema closure");\n` +
      `const owner = ${JSON.stringify(ownerRegistry)}; validateSchemaSync(owner.schema_id, owner); await validateSchema(owner.schema_id, owner);\n` +
      `if (JSON.stringify(loaded()) !== ${JSON.stringify(JSON.stringify([...summaryClosure, validatorFilename(ownerRegistry.schema_id)].sort()))}) throw Error("unrelated validators loaded by owner validation");\n` +
      `for (const validate of [validateSchemaSync, validateSchema]) {\n` +
      `  for (const [id, value, expected] of [[owner.schema_id, {}, "validation failed"], ["cartulary.unregistered.v1", {}, "unregistered schema attachment"]]) {\n` +
      `    let rejected = false; try { await validate(id, value); } catch (error) { if (!error.message.includes(expected)) throw error; rejected = true; }\n` +
      `    if (!rejected) throw Error("invalid or unregistered schema accepted");\n` +
      `  }\n` +
      `}\n` +
      `try { validateSchemaSync("cartulary.tool_run_summary.v5", {}); } catch (error) {\n` +
      `  if (!String(error.message).includes("cartulary.tool_run_summary.v5 validation failed")) throw error;\n` +
      `  process.stdout.write("foundation validator accepted a valid summary and rejected an invalid summary\\n");\n` +
      `  process.exit(0);\n` +
      `}\nthrow new Error("foundation validator accepted an invalid summary");\n`,
  );
  const output = execFileSync(process.execPath, [probe], {
    cwd: scratch,
    env: { HOME: scratch, PATH: process.env.PATH ?? "" },
    encoding: "utf8",
  });
  assert.match(output, /accepted a valid summary and rejected an invalid summary/u);
} finally {
  rmSync(scratch, { recursive: true, force: true });
}

process.stdout.write("foundation schema validator checks passed\n");
