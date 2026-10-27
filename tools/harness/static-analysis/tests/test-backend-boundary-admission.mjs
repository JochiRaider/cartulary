import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { prepareBoundaryPolicy, inspectBackendBoundaries } from "../backend-module-boundary-check.mjs";
import { emptyPolicy, inspect, repositoryPolicy, evaluateBoundaryPolicy } from "./backend-boundary-fixtures.mjs";
import { createFixture, repository } from "./fixtures.mjs";

const rawPolicy = () => JSON.parse(readFileSync(path.join(repository, "tools/backend_module_boundaries.json"), "utf8"));
const tokenRule = (scan_paths) => ({
  id: "literal", tokens: ["forbidden"], match_kind: "source_text", scan_paths,
  allowed_paths: [], production_only: false,
});

export function registerBoundariesTests() {
  test("backend admission: real manifest and every rule family reject unknown fields", () => {
    assert.doesNotThrow(() => prepareBoundaryPolicy(rawPolicy()));
    const sample = rawPolicy();
    const families = Object.keys(sample).filter((key) => Array.isArray(sample[key]) && sample[key][0]?.id);
    for (const family of families) {
      const value = rawPolicy();
      value[family][0].ignored_typo = true;
      assert.throws(() => prepareBoundaryPolicy(value), /validation failed/u, family);
    }
    for (const family of ["restricted_imports", "forbidden_declarations", "exact_calls"]) {
      const value = rawPolicy();
      value.migration_source_boundary[family][0].ignored_typo = true;
      assert.throws(() => prepareBoundaryPolicy(value), /validation failed/u, family);
    }
    for (const family of ["generated_root_writes", "command_root_shape"]) {
      const value = rawPolicy();
      value[family].ignored_typo = true;
      assert.throws(() => prepareBoundaryPolicy(value), /validation failed/u, family);
    }
  });

  test("backend admission: malformed policy never becomes a silent no-op", (t) => {
    const invalidPolicies = [
      {}, emptyPolicy({ schema_id: "cartulary.backend_module_boundaries.v3" }),
      emptyPolicy({ generated_root_writes: [] }),
      emptyPolicy({ generated_root_writes: { scan_roots: [] } }),
      emptyPolicy({ forbidden_source_tokens: [{ ...tokenRule(["src/**"]), tokens: undefined, imports: ["example/private"] }] }),
      emptyPolicy({ forbidden_source_tokens: [{ ...tokenRule(["src/**"]), tokens: [] }] }),
      emptyPolicy({ forbidden_source_tokens: [{ ...tokenRule(["src/**"]), tokens: [" "] }] }),
      emptyPolicy({ forbidden_source_tokens: [{ ...tokenRule(["src/**"]), id: " " }] }),
      emptyPolicy({ forbidden_source_tokens: [{ ...tokenRule(["src/**"]), match_kind: undefined }] }),
      emptyPolicy({ forbidden_source_tokens: [{ ...tokenRule(["src/**"]), match_kind: "subtree" }] }),
      emptyPolicy({ forbidden_source_tokens: [{ ...tokenRule(["src/**"]), exact_file_set_id: "unknown" }] }),
      emptyPolicy({ forbidden_source_tokens: [{ ...tokenRule(undefined), exact_file_set_id: "unknown" }] }),
      emptyPolicy({ forbidden_go_imports: [{ id: "import", imports: ["example/private"], scan_paths: ["src/**"], allowed_paths: [], production_only: true }] }),
      emptyPolicy({ forbidden_go_calls: [{ id: "call", symbols: ["Forbidden("], allowed_paths: [], production_only: true }] }),
    ];
    const fixture = createFixture(t);
    for (const value of invalidPolicies) {
      assert.throws(() => prepareBoundaryPolicy(value));
      fixture.write("invalid.json", value);
      // A missing support inventory must not mask invalid policy admission.
      assert.throws(() => inspectBackendBoundaries({
        root: fixture.root, manifest: path.join(fixture.root, "invalid.json"), supportInventory: "absent.json",
      }), (error) => !error.message.includes("test support inventory not found"));
    }
  });

  test("backend paths: exact paths, component stars and terminal subtrees are distinct", () => {
    for (const [selector, matched, rejected] of [
      ["src/source.go", ["src/source.go"], ["src/source.go/child.go", "src/source.go2"]],
      ["src/*.go", ["src/source.go", "src/renamed.go"], ["src/nested/source.go", "src/source.go.sql", "srcpeer/source.go"]],
      ["src/*_adapters.go", ["src/entity_adapters.go", "src/_adapters.go"], ["src/adapters.go", "src/nested/entity_adapters.go"]],
      ["src/**", ["src/source.go", "src/nested/source.go"], ["srcpeer/source.go"]],
      ["src.v1/*.go", ["src.v1/source.go"], ["srcXv1/source.go"]],
    ]) {
      const policy = prepareBoundaryPolicy(emptyPolicy({ forbidden_source_tokens: [tokenRule([selector])] }));
      for (const [paths, denied] of [[matched, true], [rejected, false]]) {
        for (const relative of paths) assert.equal(evaluateBoundaryPolicy(policy, {
          files: [{ relative, content: "package fixture\n// forbidden" }],
        }).violations.length, Number(denied), `${selector}: ${relative}`);
      }
    }
    for (const selector of ["**", "src/**/*.go", "src/**/leaf", "src/a**b.go", "src/?.go", "src/[ab].go", "src/{a,b}.go", "/src/a.go", "C:/src/a.go", "src\\a.go", "src/../a.go", "./src/a.go", "src//a.go", "src/", " src/a.go"]) {
      assert.throws(() => prepareBoundaryPolicy(emptyPolicy({ forbidden_source_tokens: [tokenRule([selector])] })), /selector/u, selector);
    }
  });

  test("backend paths: subtree excludes prune descendants and preserve siblings", (t) => {
    const fixture = createFixture(t);
    fixture.write("policy.json", emptyPolicy({ scan_roots: ["src"], scan_excludes: ["src/generated/**"], forbidden_source_tokens: [tokenRule(["src/**"])] }));
    fixture.write("support.json", { schema_id: "cartulary.test_support_inventory.v1", go_support_roots: [{ path: "support", runtime_scan: "excluded", support_scan: "included" }] });
    fixture.write("src/generated/nested/skip.go", "package fixture\n// forbidden");
    fixture.write("src/generatedpeer/check.go", "package fixture\n// forbidden");
    const result = inspectBackendBoundaries({ root: fixture.root, manifest: path.join(fixture.root, "policy.json"), supportInventory: "support.json" });
    assert.deepEqual(result.violations.map((entry) => entry.path), ["src/generatedpeer/check.go"]);
  });

  test("backend policy: newly effective root declaration, import and token rules", () => {
    const manifest = repositoryPolicy();
    const declarations = manifest.migrationSourceBoundary.forbiddenDeclarations.filter((rule) => rule.id === "links-retired-production-declarations");
    assert.equal(declarations.length, 1);
    const declarationPolicy = { ...prepareBoundaryPolicy(emptyPolicy()), migrationSourceBoundary: { ...manifest.migrationSourceBoundary, restrictedImports: [], forbiddenDeclarations: declarations, exactCalls: [] } };
    for (const [content, count] of [["type RecordLink struct{}", 1], ["type CurrentLink struct{}", 0]]) {
      const result = evaluateBoundaryPolicy(declarationPolicy, { migrationBoundaryFiles: [{ relative: "internal/modules/links/renamed.go", content: `package links\n${content}` }] });
      assert.equal(result.violations.length, count);
    }
    for (const [id, relative, prohibited] of [
      ["incidents-root-is-transport-neutral", "internal/modules/incidents/renamed.go", "net/http"],
      ["timeline-no-tabular-parser-implementations", "internal/modules/timeline/renamed.go", "archive/zip"],
    ]) {
      const rules = manifest.forbiddenGoImports.filter((rule) => rule.id === id);
      assert.equal(rules.length, 1);
      for (const [imported, count] of [[prohibited, 1], ["fmt", 0]]) assert.equal(inspect([{ relative, content: `package fixture\nimport "${imported}"` }], { forbiddenGoImports: rules }).length, count, id);
    }
    for (const [id, relative, prohibited] of [
      ["retired-incidents-http-root-api", "internal/modules/incidents/renamed.go", "type RouteOptions struct{}"],
      ["assessments-retired-incident-bundle-helpers", "internal/modules/assessments/renamed.go", "func ExportIncidentBundleFiles() {}"],
      ["retired-workbook-legacy-mutation-surfaces", "internal/modules/workbook/renamed.go", "type Store struct{}"],
      ["retired-workbook-provider-callback-and-restore-aliases", "internal/modules/workbook/renamed.go", "type QueryProviderFunc func()"],
      ["retired-workbook-adapter-admission-dispatch-and-hash-override", "internal/app/workbookassembly/entity_adapters.go", "var value command.Admission"],
      ["tasks-decisions-retired-forwarders-and-dead-surfaces", "internal/modules/tasksdecisions/renamed.go", "func ValidateTaskCreateParams() {}"],
    ]) {
      const rules = manifest.forbiddenSourceTokens.filter((rule) => rule.id === id);
      assert.equal(rules.length, 1);
      for (const [content, count] of [[prohibited, 1], ["func CurrentSurface() {}", 0]]) {
        assert.equal(inspect([{ relative, content: `package fixture\n${content}` }], { forbiddenSourceTokens: rules }).length, count, id);
      }
    }
  });

  test("backend policy: Timeline contracts and source tables retain narrow ownership", () => {
    const manifest = repositoryPolicy();
    const imports = manifest.goImportAllowlists.filter((rule) => rule.id === "timeline-approved-owner-imports");
    const tables = manifest.sqlTableAllowlists.filter((rule) => rule.id === "timeline-canonical-source-repository-tables-only");
    assert.equal(imports.length, 1);
    assert.equal(tables.length, 1);
    const prefix = "github.com/JochiRaider/cartulary/internal/modules/";
    for (const contract of ["revisions/historycontract", "imports/ownerfacade", "incidentbundles/sourceport", "records/subtypepresence"]) {
      for (const [suffix, count] of [["", 0], ["/private", 1]]) {
        assert.equal(inspect([{ relative: "internal/modules/timeline/renamed.go", content: `package timeline\nimport "${prefix}${contract}${suffix}"` }], { goImportAllowlists: imports }).length, count, contract + suffix);
      }
    }
    for (const [sql, count] of [
      ["SELECT * FROM timeline_source_provenance JOIN records ON true", 0],
      ["SELECT * FROM timeline_events JOIN timeline_time_conversion_profiles ON true", 0],
      ["SELECT * FROM jobs", 1],
      ["SELECT * FROM timeline_grid_projection", 1],
    ]) assert.equal(inspect([{ relative: "internal/modules/timeline/provenance.go", content: `package timeline\nconst query = \`${sql}\`` }], { sqlTableAllowlists: tables }).length, count, sql);
    const records = manifest.sqlTableAccess.filter((rule) => rule.id === "records-current-envelope-access");
    assert.equal(records.length, 1);
    for (const [relative, sql, count] of [
      ["internal/modules/timeline/incident_bundle_source_port.go", "SELECT * FROM records", 0],
      ["internal/modules/timeline/incident_bundle_source_port.go", "UPDATE records SET row_version = 2", 1],
      ["internal/modules/timeline/unapproved.go", "SELECT * FROM records", 1],
    ]) assert.equal(inspect([{ relative, content: `package timeline\nconst query = \`${sql}\`` }], { sqlTableAccess: records }).length, count, relative + sql);
  });

  test("backend policy: Recovery imports enforce current boundaries and exact adapter permission", () => {
    const manifest = repositoryPolicy();
    const prefix = "github.com/JochiRaider/cartulary/";
    for (const [id, relative, imported, count] of [
      ["recovery-application-no-operator-transport-imports", "internal/modules/recovery/application/restore.go", "internal/app/operator", 1],
      ["recovery-application-no-operator-transport-imports", "internal/modules/recovery/application/restore.go", "internal/app/operator/internal/recoverycli", 1],
      ["recovery-application-no-operator-transport-imports", "internal/modules/recovery/application/restore.go", "internal/app/operatorpeer", 0],
      ["recovery-sqlc-confined-to-private-store-adapter", "internal/modules/recovery/store.go", "internal/gen/sql", 0],
      ["recovery-sqlc-confined-to-private-store-adapter", "internal/modules/recovery/another_store.go", "internal/gen/sql", 1],
      ["recovery-application-no-audit-adapter-import", "internal/modules/recovery/application/restore.go", "internal/platform/administrativeaudit", 1],
      ["recovery-application-no-audit-adapter-import", "internal/modules/recovery/application/restore.go", "internal/platform/administrativeaudit/private", 1],
      ["recovery-application-no-audit-adapter-import", "internal/modules/recovery/application/restore.go", "internal/platform/postgres", 0],
    ]) {
      const rules = manifest.forbiddenGoImports.filter((rule) => rule.id === id);
      assert.equal(rules.length, 1);
      const violations = inspect([{ relative, content: `package recovery\nimport "${prefix}${imported}"` }], { forbiddenGoImports: rules });
      assert.equal(violations.length, count, id + imported);
      if (count) assert.equal(violations[0].rule_id, id);
    }
  });

  test("backend admission: optional SQL permissions retain their distinct meanings", () => {
    const rule = { id: "owner-projection-storage-access", table: "records", scan_paths: ["src/**"], read_allowed_paths: ["src/read.go"], write_allowed_paths: [] };
    const files = [{ relative: "src/read.go", content: "package fixture\nconst sql = `SELECT * FROM records FOR UPDATE`" }];
    const omitted = prepareBoundaryPolicy(emptyPolicy({ sql_table_access: [rule] }));
    assert.deepEqual(evaluateBoundaryPolicy(omitted, { files }).violations, []);
    const forbidden = prepareBoundaryPolicy(emptyPolicy({ sql_table_access: [{ ...rule, lock_allowed_paths: [] }] }));
    assert.equal(evaluateBoundaryPolicy(forbidden, { files }).violations[0].code, "sql_table_lock_access");
    assert.equal(evaluateBoundaryPolicy(omitted, { testSupportFiles: files }).violations.length, 2);
  });
}
