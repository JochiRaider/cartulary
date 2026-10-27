import assert from "node:assert/strict";
import test from "node:test";
import { prepareBoundaryPolicy } from "../backend-module-boundary-check.mjs";
import { emptyPolicy, inspect, repositoryPolicy, evaluateBoundaryPolicy } from "./backend-boundary-fixtures.mjs";

export function registerBoundariesTests() {
  test("backend mechanics: import matching is explicit and segment bounded", () => {
    for (const match_kind of ["exact", "subtree"]) {
      const policy = prepareBoundaryPolicy(emptyPolicy({ forbidden_go_imports: [{
        id: "private-owner", match_kind, imports: ["example/owner"], scan_paths: ["internal/**"],
        allowed_paths: ["internal/assembly.go"], production_only: true,
      }] }));
      for (const [relative, imported, denied] of [
        ["internal/source.go", "example/owner", true],
        ["internal/source.go", "example/owner/private", match_kind === "subtree"],
        ["internal/source.go", "example/ownerpeer", false],
        ["internal/assembly.go", "example/owner", false],
        ["internal/source_test.go", "example/owner", false],
        ["elsewhere/source.go", "example/owner", false],
      ]) {
        const { violations } = evaluateBoundaryPolicy(policy, { files: [{ relative, content: `package fixture\nimport "${imported}"` }] });
        assert.equal(violations.length, Number(denied), `${match_kind} ${relative} ${imported}`);
        if (denied) assert.deepEqual(violations[0], {
          code: "forbidden_go_import", path: relative, symbol_or_import: imported,
          rule_id: "private-owner", match_kind, result: "fail",
        });
      }
    }
  });

  test("backend mechanics: identifier and source-text rules have independent meaning", () => {
    const definition = {
      id: "retired-name", tokens: ["ConfigureTelemetry", "collaboration_dispatcher"],
      scan_paths: ["internal/**"], allowed_paths: [], production_only: false, match_kind: "go_identifier",
    };
    const policy = prepareBoundaryPolicy(emptyPolicy({ forbidden_source_tokens: [definition] }));
    for (const [source, expected] of [
      ["func ConfigureTelemetry () {}", 1],
      ["func f() { runtime.ConfigureTelemetry /* comment */ () }", 1],
      ["var collaboration_dispatcher int", 1],
      ['const fixture = "func ConfigureTelemetry() {}"', 0],
      ['const fixture = `ConfigureTelemetry()`', 0],
      ["// ConfigureTelemetry()\n/* collaboration_dispatcher */", 0],
      ['const fixture = "escaped \\\" ConfigureTelemetry()"', 0],
      ["const fixture = '\\'' // ConfigureTelemetry", 0],
      ["var ConfigureTelemetrySuffix, πConfigureTelemetry int", 0],
    ]) {
      assert.equal(evaluateBoundaryPolicy(policy, { files: [{
        relative: "internal/fixture_test.go", content: `package fixture\n${source}`,
      }] }).violations.length, expected, source);
    }
    const textPolicy = prepareBoundaryPolicy(emptyPolicy({ forbidden_source_tokens: [{ ...definition, match_kind: "source_text" }] }));
    assert.equal(evaluateBoundaryPolicy(textPolicy, { files: [{
      relative: "internal/fixture_test.go", content: '// ConfigureTelemetry',
    }] }).violations.length, 1);
    for (const invalid of [{ match_kind: undefined }, { match_kind: "unknown" }, { tokens: ["bad.name"] }]) {
      assert.throws(() => prepareBoundaryPolicy(emptyPolicy({ forbidden_source_tokens: [{ ...definition, ...invalid }] })));
    }
  });

  test("backend mechanics: closed SQL allowlists reject physical and unresolved targets", () => {
    for (const sql of [
      "WITH allowed AS (SELECT * FROM jobs) SELECT * FROM allowed",
      "WITH records AS (VALUES (1)) SELECT * FROM public.jobs",
      "SELECT * FROM %s",
    ]) {
      assert.equal(inspect([{ relative: "fixture.sql", content: sql }], {
        sqlTableAllowlists: [{ scanPaths: ["fixture.sql"], allowedTables: new Set(["records"]) }],
      }).length, 1, sql);
    }
  });

  test("backend mechanics: analysis state and dynamic observations belong to one inspection", () => {
    const files = [{ relative: "fixture.go", content: "package fixture\nconst query = `SELECT * FROM %s`" }];
    const policy = prepareBoundaryPolicy(emptyPolicy());
    const constrained = { ...policy, sqlTableAllowlists: [{ scanPaths: ["fixture.go"], allowedTables: new Set() }] };
    const first = evaluateBoundaryPolicy(constrained, { files });
    assert.deepEqual(first.dynamicRelationSources, [{ path: "fixture.go", operation: "read" }]);
    assert.deepEqual(evaluateBoundaryPolicy(policy, { files }), { violations: [], dynamicRelationSources: [] });
    assert.deepEqual(evaluateBoundaryPolicy(constrained, { files }), first);
  });

  test("backend policy: migration source ownership uses concrete sources", () => {
    const boundary = repositoryPolicy().migrationSourceBoundary;
    // Literal cases are independent of rule fields and path patterns.
    const approved = [
      { relative: "db/migrations/source.go", content: "package migrations\nfunc Source() { database_migrations.BuildCanonicalEmbedded(files, embeddedPath) }" },
      { relative: "internal/modules/database_migrations/migrations.go", content: 'package migrations\nimport "github.com/JochiRaider/cartulary/internal/modules/database_migrations/sourcecatalog"\nfunc BuildCanonicalEmbedded() {}' },
      { relative: "internal/testutil/pgtest/pgtest.go", content: 'package pgtest\nimport "github.com/JochiRaider/cartulary/internal/modules/database_migrations/sourcecatalog"\nfunc (m *MigrationDatabase) ApplyThrough() {}\nfunc (m *MigrationDatabase) RollbackThrough() {}' },
      { relative: "internal/app/server/runtime_test.go", content: 'package server\nimport "github.com/JochiRaider/cartulary/internal/testutil/pgtest"' },
    ];
    const policy = { ...prepareBoundaryPolicy(emptyPolicy()), migrationSourceBoundary: boundary };
    assert.deepEqual(evaluateBoundaryPolicy(policy, { migrationBoundaryFiles: approved }).violations, []);
    for (const [relative, content, code, rule] of [
      ["internal/app/server/forbidden.go", 'import "github.com/JochiRaider/cartulary/internal/modules/database_migrations/sourcecatalog"', "forbidden_go_import", "database-migrations-sourcecatalog-import"],
      ["internal/app/server/forbidden.go", 'import "github.com/JochiRaider/cartulary/internal/testutil/pgtest"', "forbidden_go_import", "production-no-pgtest-import"],
      ["db/migrations/forbidden.go", "var Files int", "forbidden_go_declaration", "database-migrations-adapter-no-raw-exports"],
      ["internal/modules/database_migrations/forbidden.go", "func NewSource() {}", "forbidden_go_declaration", "database-migrations-no-generic-source-constructor"],
      ["internal/modules/database_migrations/forbidden_test.go", "func HasVersion() {}", "forbidden_go_declaration", "database-migrations-no-retired-state-helpers"],
      ["internal/modules/database_migrations/forbidden.go", "const migrationAdvisoryLockID = 1", "forbidden_go_declaration", "database-migrations-single-lock-policy-owner"],
      ["internal/modules/database_migrations/migrationevidence/forbidden.go", "type ManifestFailure struct{}", "forbidden_go_declaration", "migration-evidence-no-exported-manifest-failure"],
      ["internal/testutil/pgtest/forbidden.go", "func NewMigrationDatabase() {}", "forbidden_go_declaration", "pgtest-no-raw-migration-scratch-surface"],
      ["internal/app/server/forbidden.go", "func ApplyThrough() {}", "forbidden_go_declaration", "targeted-migrations-remain-pgtest-only"],
      ["internal/app/server/forbidden.go", "func LoadCollectionFieldsChangedTx() {}", "forbidden_go_declaration", "links-iteration3-retired-conflict-declarations"],
      ["internal/modules/links/internal/valuecodec/forbidden.go", "func DecodeMutationValue() {}", "forbidden_go_declaration", "links-iteration3-retired-codec-declarations"],
      ["internal/modules/links/internal/revisionprovider/forbidden.go", "func ValidateRecordLinkValue() {}", "forbidden_go_declaration", "links-iteration3-retired-provider-wrappers"],
      ["internal/modules/links/forbidden.go", "func validateTagCatalog() {}", "forbidden_go_declaration", "links-iteration3-retired-source-state-helpers"],
      ["internal/app/server/forbidden.go", "func build() { database_migrations.BuildCanonicalEmbedded() }", "forbidden_go_call", "canonical-migration-source-single-adapter-construction"],
    ]) {
      const violations = evaluateBoundaryPolicy(policy, { migrationBoundaryFiles: [...approved, { relative, content: `package fixture\n${content}` }] }).violations;
      assert.equal(violations.length, 1, `${rule}: ${JSON.stringify(violations)}`);
      assert.equal(violations[0].code, code);
      assert.equal(violations[0].rule_id, rule);
      assert.equal(violations[0].path, relative);
    }
    assert.equal(evaluateBoundaryPolicy(policy, { migrationBoundaryFiles: approved.slice(1) }).violations[0].code, "required_go_call_count");
  });
}
