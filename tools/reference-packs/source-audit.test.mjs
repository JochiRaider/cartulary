import assert from "node:assert/strict";
import { chmodSync, readFileSync, existsSync, mkdirSync, mkdtempSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import test from "node:test";
import { createHash } from "node:crypto";

const tools = path.resolve("tools");
const canonical = value => Array.isArray(value) ? `[${value.map(canonical).join(",")}]` : value !== null && typeof value === "object" ? `{${Object.keys(value).sort().map(key => `${JSON.stringify(key)}:${canonical(value[key])}`).join(",")}}` : JSON.stringify(value);
const digest = "a".repeat(64);
const sha = raw => createHash("sha256").update(raw).digest("hex");
const builtAt = "2026-11-09T00:00:00Z";
const owned = producer => `referencepacks/${producer}/inputs/`;
// Explicit examples of the current report contracts, independent of the schema
// compiler and producer encoder. The wrapper checks delivery, not source facts.
const report = (producer, root) => {
  const base = owned(producer);
  const binary = readFileSync(path.join(root, `build/bin/reference-pack-${producer}`));
  const common = {source_lock_sha256: sha(readFileSync(path.join(root, producer === "attack" ? "input.json" : "source/source-lock.json"))), source_profile_sha256: sha(readFileSync(`${base}dataset-profile.v${producer === "attack" ? 2 : 1}.json`)), builder: {builder_id: `cartulary.reference-pack.${producer}`, builder_version: "dataset-v1", builder_source_sha256: sha(binary), size_bytes: binary.length, go_version: "go1.26.0"}};
  if (producer === "attack") return {schema_id: "cartulary.attack_source_audit.v2", ...common, source_profile_id: "cartulary.attack.all_domains.dataset.v2", field_policy_sha256: sha(readFileSync(`${base}field-policy.native.v1.json`)), status: "source_audit_passed", built_at: builtAt, domains: {}, unique_records: 0, unique_types: {}, cross_domain_duplicate_ids: 0, properties: {}, record_locators: {}, alias_collisions: [], lookup_aliases: {}, tactic_memberships: [], alias_occurrences: 0, max_aliases_per_object: 0, max_external_refs_per_record: 0, citations_without_external_id: 0, urls_with_fragments: 0, non_https_urls: 0, source_version_values: {}, lifecycle_flags: {}, embedded_reference_counts: {}, embedded_references: [], max_canonical_source_record_bytes: 0, issues: []};
  if (producer === "veris") return {schema_id: "cartulary.veris_audit.v1", ...common, rules_sha256: sha(readFileSync(`${base}rules.v1.json`)), dispositions_sha256: sha(readFileSync(`${base}dispositions.v1.json`)), counts: {}, objects_sha256: digest, relationships_sha256: digest};
  if (producer === "entra") return {schema_id: "cartulary.entra_reconciliation.v2", ...common, counts: {}, assertions: [], unselected_rows: [], alias_omissions: {}, field_accounting: [], entries_sha256: digest, entries_bytes: 0, maximum_label_scalars: 0, maximum_aliases: 0, maximum_line_bytes: 0};
  return {schema_id: "cartulary.windows_event_reconciliation.v2", ...common, editorial_sha256: sha(readFileSync(`${base}editorial.v2.json`)), status: "accepted", entry_count: 0, payload_file_sha256: digest, evidence_sha256: digest, maximum_description_scalars: 0, maximum_line_bytes: 0, coverage: [], dispositions: [], issues: [], limitations: []};
};

test("source wrapper requires a valid report and preserves the producer failure", () => {
  const root = mkdtempSync(path.join(tmpdir(), "source-audit-command-"));
  try {
    mkdirSync(path.join(root, "build/bin"), {recursive: true});
    mkdirSync(path.join(root, "source"));
    symlinkSync(tools, path.join(root, "tools"));
    symlinkSync(path.resolve("referencepacks"), path.join(root, "referencepacks"));
    for (const producer of ["attack", "veris", "windows", "entra"]) {
      const file = path.join(root, `build/bin/reference-pack-${producer}`);
      writeFileSync(file, `#!${process.execPath}\nimport fs from 'node:fs';
const report = process.argv[process.argv.indexOf('--report-file') + 1];
if (process.env.REPORT_MODE === 'missing') fs.unlinkSync(report);
else fs.writeFileSync(report, process.env.REPORT_BYTES ?? '');
if (process.env.PRODUCER_EXIT !== '0') { process.stdout.write('producer failure report\\n'); process.stderr.write('producer diagnostic\\n'); }
process.exit(Number(process.env.PRODUCER_EXIT));\n`);
      chmodSync(file, 0o700);
    }
    const select = producer => {
      const schema = {attack: "cartulary.attack_source_lock.v1", veris: "cartulary.veris_inventory.v1", windows: "cartulary.windows_event_inventory.v2", entra: "cartulary.entra_inventory.v2"}[producer];
      const input = canonical({schema_id: schema});
      writeFileSync(path.join(root, "input.json"), input);
      if (producer !== "attack") writeFileSync(path.join(root, "source/source-lock.json"), canonical({schema_id: {veris: "cartulary.veris_source_lock.v1", windows: "cartulary.windows_event_source_lock.v2", entra: "cartulary.entra_source_lock.v2"}[producer], inventory_sha256: sha(input), sources: []}));
    };
    const invoke = (raw, status = 0, producer = "attack", mode = "write") => spawnSync("/bin/bash", [path.join(tools, "reference-packs/source.sh")], {
      cwd: root, encoding: "utf8", env: {PATH: process.env.PATH, NODE_BIN: process.execPath, PRODUCER: producer, ACTION: "audit", BUILT_AT: builtAt, SOURCE_LOCK: "input.json", SOURCE_DIR: "source", REPORT_BYTES: raw, PRODUCER_EXIT: String(status), REPORT_MODE: mode},
    });
    writeFileSync(path.join(root, "input.json"), '{"schema_id":"cartulary.attack_source_lock.v1"}');
    const valid = report("attack", root);
    assert.equal(invoke(canonical(valid)).status, 0);
    for (const raw of ["", "not json", "{}", "null", canonical({...valid, schema_id: "wrong"}), canonical({...valid, status: "source_audit_failed"}), canonical({...valid, issues: ["failure"]}), canonical({...valid, builder: {}})]) assert.notEqual(invoke(raw).status, 0, raw);
    assert.notEqual(invoke("", 0, "attack", "missing").status, 0);
    for (const producer of ["attack", "veris", "windows", "entra"]) {
      select(producer);
      const good = report(producer, root);
      for (const key of Object.keys(good)) {
        const missing = {...good}; delete missing[key];
        assert.match(invoke(canonical(missing), 0, producer).stderr, /invalid source-audit report contract/u, `${producer}: required ${key}`);
      }
      for (const key of Object.keys(good).filter(key => key.endsWith("_sha256"))) {
        assert.match(invoke(canonical({...good, [key]: [good[key]]}), 0, producer).stderr, /invalid source-audit report contract/u, `${producer}: digest array ${key}`);
      }
      for (const bad of [{...good, extra: true}, {...good, builder: {...good.builder, size_bytes: "1"}}, {...good, builder: {...good.builder, builder_source_sha256: [digest]}}]) {
        assert.match(invoke(canonical(bad), 0, producer).stderr, /invalid source-audit report contract/u);
      }
      const bindings = ["source_lock_sha256", "source_profile_sha256", ...({attack: ["field_policy_sha256"], veris: ["rules_sha256", "dispositions_sha256"], windows: ["editorial_sha256"], entra: []}[producer])];
      for (const key of bindings) assert.match(invoke(canonical({...good, [key]: digest}), 0, producer).stderr, /invocation mismatch/u, key);
      for (const patch of [{builder_source_sha256: digest}, {builder_id: "unrelated"}, {builder_version: "old"}, {size_bytes: 1}]) assert.match(invoke(canonical({...good, builder: {...good.builder, ...patch}}), 0, producer).stderr, /builder mismatch/u);
    }
    select("attack");
    for (const patch of [{built_at: "2026-11-08T00:00:00Z"}, {source_profile_id: "wrong"}]) assert.match(invoke(canonical({...valid, ...patch})).stderr, /invocation mismatch/u);
    assert.match(invoke(canonical({...valid, unique_records: "1"})).stderr, /invalid source-audit report contract/u);
    assert.match(invoke(canonical({...valid, embedded_references: [{source_id: "a"}]})).stderr, /invalid source-audit report contract/u);

    for (const raw of ["", "malformed", canonical(valid)]) {
      const result = invoke(raw, 37);
      assert.equal(result.status, 37);
      assert.equal(result.stdout, "producer failure report\n");
      assert.equal(result.stderr, "producer diagnostic\n");
    }
    assert(!existsSync(path.join(root, "source/witnesses")));
    for (const [producer, input] of [
      ["veris", "cartulary.veris_inventory.v1"],
      ["windows", "cartulary.windows_event_inventory.v2"],
      ["entra", "cartulary.entra_inventory.v2"],
    ]) {
      select(producer);
      const result = invoke(canonical(report(producer, root)), 0, producer);
      assert.notEqual(result.status, 0, `${producer} cannot pass without witness inputs`);
      assert.match(result.stderr, /ENOENT/u);
      assert(!result.stdout.includes("Independent witness retained"));
      assert(!existsSync(path.join(root, "source/witnesses")));
      const saved = report(producer, root);
      writeFileSync(path.join(root, "input.json"), canonical({schema_id: input, changed: true}));
      assert.match(invoke(canonical(saved), 0, producer).stderr, /inventory binding mismatch/u);
      const failed = invoke("malformed", 37, producer);
      assert.equal(failed.status, 37);
      assert.equal(failed.stdout, "producer failure report\n");
      assert.equal(failed.stderr, "producer diagnostic\n");
    }
    writeFileSync(path.join(root, "input.json"), '{"schema_id":"cartulary.attack_source_lock.v1"}');
    writeFileSync(path.join(root, "report.json"), canonical(valid));
    const unsupported = spawnSync(process.execPath, [path.join(tools, "reference-packs/source-witness.mjs"), "source-audit", "input.json", "report.json", "source"], {cwd: root, encoding: "utf8"});
    assert.notEqual(unsupported.status, 0);
    assert.match(unsupported.stderr, /unsupported witness input schema/u);
    assert(!existsSync(path.join(root, "source/witnesses")));
    for (const action of ["assemble", "envelope", "renew"]) {
      const result = spawnSync(process.execPath, [path.join(tools, "reference-packs/source-witness.mjs"), action, "input.json", "report.json", "source"], {cwd: root, encoding: "utf8"});
      assert.notEqual(result.status, 0);
      assert.match(result.stderr, /unsupported witness action/u);
      assert(!existsSync(path.join(root, "source/witnesses")));
    }
  } finally { rmSync(root, {recursive: true, force: true}); }
});
