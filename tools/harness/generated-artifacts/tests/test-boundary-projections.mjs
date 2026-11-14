import assert from "node:assert/strict";
import { existsSync, readFileSync, mkdtempSync, writeFileSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { createContractTestContext } from "../../tests/contract-test-context.mjs";
import { registerJSONShapeTests } from "./test-json-shapes.mjs";
import { catalogDDLSource } from "../../../database-migrations/catalog-ddl-source.mjs";
import { collectForeignKeyCoverage, requireForeignKeyCoverage } from "../../../database-migrations/foreign-key-coverage.mjs";
import { validateSchemaObjectOwnershipManifestShape } from "../database-contract-drift/schema-object-ownership.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}
export function registerBoundariesTests() {
  registerJSONShapeTests();
  test("migration_fk_coverage: derived key prefixes and exact manifest claims", () => {
    const table = `CREATE TABLE public.child (a int, b int, c int,
      CONSTRAINT child_fk FOREIGN KEY(a,b) REFERENCES public.parent(a,b));`;
    const collect = (...sources) => collectForeignKeyCoverage(sources.map((source, i) => ({ filename: `${i}.sql`, source })));
    for (const [definition, covered] of [
      ["(a,b)", true], ["(a DESC,b,c)", true], ["(b,a)", false], ["(a,c,b)", false],
      ["(a) INCLUDE (b)", false], ["(a,b) WHERE a IS NOT NULL", false], ["((a+0),b)", false],
      ["USING hash (a)", false], ["(a,b) NULLS NOT DISTINCT", true],
      ["(a,b) INCLUDE (c)", true], ["(a,b,(c+1))", true],
    ]) {
      const facts = collect(table, `CREATE INDEX candidate ON public.child ${definition};`);
      assert.equal(facts[0].coveringIndexes.length > 0, covered, definition);
      if (covered) assert.doesNotThrow(() => requireForeignKeyCoverage(facts));
      else assert.throws(() => requireForeignKeyCoverage(facts), /Uncovered foreign key public.child.child_fk/u);
    }
    for (const kind of ["PRIMARY KEY", "UNIQUE"]) {
      assert.doesNotThrow(() => requireForeignKeyCoverage(collect(table,
        `ALTER TABLE public.child ADD CONSTRAINT child_key ${kind}(a,b);`)));
      assert.throws(() => requireForeignKeyCoverage(collect(table,
        `ALTER TABLE public.child ADD CONSTRAINT child_key ${kind}(a,b); ALTER TABLE public.child DROP CONSTRAINT child_key;`)), /Uncovered/u);
      assert.throws(() => collect(table, `ALTER TABLE public.child ADD CONSTRAINT child_key ${kind}(a,b) INCLUDE (c);`), /backing index .* suffix/u);
    }
    const index = "CREATE INDEX candidate ON public.child(a,b);";
    assert.doesNotThrow(() => requireForeignKeyCoverage(collect(table, index, "ALTER TABLE public.child DROP COLUMN c;")));
    for (const definition of ["(a,b) INCLUDE (c)", "(a,b,(c+1))"]) {
      const dependentIndex = `CREATE INDEX candidate ON public.child ${definition};`;
      assert.throws(() => collect(table, dependentIndex, "ALTER TABLE public.child DROP COLUMN c;"), /drop referenced or unresolved key column/u);
      assert.throws(() => collect(table, dependentIndex, "ALTER TABLE public.child RENAME COLUMN c TO renamed; ALTER TABLE public.child DROP COLUMN renamed;"), /drop referenced or unresolved key column/u);
      assert.doesNotThrow(() => requireForeignKeyCoverage(collect(table, dependentIndex,
        "DROP INDEX public.candidate; ALTER TABLE public.child DROP COLUMN c;", index)));
    }
    assert.throws(() => requireForeignKeyCoverage(collect(table, index, "DROP INDEX public.candidate;")), /Uncovered/u);
    const facts = collect(table, index);
    const otherIndex = "CREATE INDEX other ON public.child(c);";
    assert.deepEqual(collect(table, index, otherIndex), collect(table, otherIndex, index));
    assert.deepEqual(collect(table, index, "DROP INDEX public.candidate;", index), facts);
    assert.deepEqual(collect(table, index, "DROP TABLE public.child;"), []);
    assert.deepEqual(collect(table, index, "ALTER TABLE public.child DROP CONSTRAINT child_fk;"), []);
    const renamed = collect(table, index, "ALTER TABLE public.child RENAME COLUMN a TO renamed;");
    assert.deepEqual(renamed[0].columns, ["renamed", "b"]);
    assert.doesNotThrow(() => requireForeignKeyCoverage(renamed));
    const entries = [{ object_kind: "constraint", qualified_name: "public.child.child_fk", foreign_key_index_status: "covered" }];
    requireForeignKeyCoverage(facts, entries);
    assert.throws(() => requireForeignKeyCoverage(facts, []), /Missing foreign keys/u);
    assert.throws(() => requireForeignKeyCoverage(facts, [{ ...entries[0], foreign_key_index_status: "not_applicable" }]), /status mismatch/u);
    assert.throws(() => requireForeignKeyCoverage([], entries), /status mismatch/u);
    // Source assertions are checked only at the final head, not at each prefix.
    assert.doesNotThrow(() => requireForeignKeyCoverage(collect(table, index)));
    assert.deepEqual(collect(table, index, "-- DROP INDEX public.candidate;\nDO $$ BEGIN DROP INDEX public.candidate; END $$;"), facts);
    assert.deepEqual(collect("CREATE TABLE public.copy AS WITH x AS (SELECT 1) SELECT * FROM x;"), []);
    assert.deepEqual(collect("CREATE TABLE public.inline (id int CONSTRAINT inline_fk REFERENCES public.parent(id), CONSTRAINT inline_pk PRIMARY KEY(id));")[0].coveringIndexes, ["public.inline_pk"]);
    for (const invalid of ["CREATE INDEX candidate ON public.child (a,b", "CREATE INDEX candidate ON public.child(a,b) SOMETHING;", "ALTER TABLE public.child RENAME TO other;", 'CREATE INDEX "quoted" ON public.child(a,b);']) {
      assert.throws(() => collect(table, invalid), /unsupported key-bearing DDL/u);
    }
  });
  test("migration_ownership_schema: full schema and canonical allocation agree", () => {
    const manifest = readJSON("tools/schema_object_ownership_manifest.json");
    validateSchemaSync(manifest.schema_id, manifest);
    validateSchemaObjectOwnershipManifestShape(path.join(root, "tools/schema_object_ownership_manifest.json"));
    const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-ownership-schema-"));
    try {
      writeFileSync(path.join(scratch, "migration_history_manifest.json"), readFileSync(path.join(root, "tools/migration_history_manifest.json")));
      const file = path.join(scratch, "schema_object_ownership_manifest.json");
      for (const mutate of [
        (value) => { value.unrecognized = true; },
        (value) => { value.entries.find((entry) => entry.migration_version !== null).migration_version = 99999; },
        (value) => { value.entries.find((entry) => entry.migration_version !== null).migration_file = "00001_wrong.sql"; },
        (value) => { value.entries.find((entry) => entry.migration_version !== null).migration_version = 0; },
      ]) {
        const candidate = structuredClone(manifest);
        mutate(candidate);
        writeFileSync(file, JSON.stringify(candidate));
        assert.throws(() => validateSchemaObjectOwnershipManifestShape(file));
      }
    } finally { rmSync(scratch, { recursive: true, force: true }); }
    const association = manifest.entries.find((entry) => entry.object_id === "table.public.import-committed-source-records");
    assert.equal(association.runtime_access_class, "table_append_only");
    assert.equal(association.recovery_access_class, "table_restore");
  });
  test("migration_catalog_ddl_lexing: only top-level declarations enter the catalog", () => {
    const before = "CREATE TABLE public.real (id int CONSTRAINT real_pk PRIMARY KEY);\n";
    const after = "ALTER TABLE public.real ADD CONSTRAINT real_ck CHECK (id > 0);";
    const ignored = [
      "-- bulk import constraint path\n",
      "/* CREATE TABLE public.fake; /* CONSTRAINT nested */ DROP TABLE public.real; */\n",
      "'CREATE TABLE public.fake; '' CONSTRAINT literal'\n",
      String.raw`E'escaped \' CONSTRAINT escaped'` + "\n",
      '"CONSTRAINT quoted""identifier"\n',
      "$$ BEGIN -- CONSTRAINT body\n DROP TABLE public.real; END $$\n",
      "$routine$ BEGIN '$$'; CONSTRAINT tagged; END $routine$\n",
      "-- Unicode comment 😀 CONSTRAINT astral\r\n",
    ];
    for (const token of ignored) {
      const source = before + token + after;
      const projected = catalogDDLSource(source);
      assert.equal(projected, before + token.replace(/[^\r\n]/g, " ") + after);
      assert.equal(projected.indexOf(after), source.indexOf(after));
      assert.deepEqual([...projected.matchAll(/CONSTRAINT\s+(\w+)/gu)].map((match) => match[1]), ["real_pk", "real_ck"]);
    }
    for (const malformed of ["/* open", "/* /* nested */", "'open", '"open', "$$open", "$tag$open$$"]) {
      assert.throws(() => catalogDDLSource(malformed), /Unterminated SQL/u);
    }
    assert.equal(catalogDDLSource("-- end of file"), "              ");
    assert.equal(catalogDDLSource("SELECT $1, name$tag$ FROM public.real"), "SELECT $1, name$tag$ FROM public.real");
  });
  test("schema_attachment_uniqueness: active schema attachments are unique", async () => {
    {
      const attachments = readJSON("tools/harness_schema_attachments.json");
      const ids = attachments.attachments.map((entry) => entry.schema_id);
      assert.equal(new Set(ids).size, ids.length);
      return;
    }
  });
  test("current_schema_presence: current harness schemas are present", async () => {
    for (const schema of [
      "cartulary.harness_work_graph.v8.schema.json",
      "cartulary.harness_execution_context.v3.schema.json",
      "cartulary.harness_runtime_plan.v6.schema.json",
      "cartulary.harness_run_manifest.v3.schema.json",
      "cartulary.harness_unit_event.v4.schema.json",
      "cartulary.harness_run_summary.v1.schema.json",
    ])
      assert.ok(existsSync(path.join(root, "tools/schemas", schema)));
  });
  test("retired_schema_absence: retired harness schemas remain absent", async () => {
    for (const oldSchema of [
      "cartulary.harness_execution_context.v1.schema.json",
      "cartulary.harness_execution_context.v2.schema.json",
      "cartulary.harness_runtime_plan.v3.schema.json",
      "cartulary.harness_work_graph.v5.schema.json",
      "cartulary.harness_invocation_start.v1.schema.json",
      "cartulary.harness_observability_index.v1.schema.json",
      "cartulary.harness_sequence_event.v1.schema.json",
      "cartulary.task_surface_owner.v1.schema.json",
      "cartulary.test_family_manifest.v2.schema.json",
      "cartulary.test_family_manifest.v4.schema.json",
      "cartulary.scheduler_manifest.v2.schema.json",
    ])
      assert.equal(existsSync(path.join(root, "tools/schemas", oldSchema)), false);
  });
}
