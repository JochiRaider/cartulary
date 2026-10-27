import assert from "node:assert/strict";
import test from "node:test";
import { inspect, repositoryPolicy } from "./backend-boundary-fixtures.mjs";

const checkSQLTableAccess = (files, rules, scope) => scope === "test_fixture"
  ? inspect([], { sqlTableAccess: rules }, { testSupportFiles: files }) : inspect(files, { sqlTableAccess: rules });
const checkForbiddenSourceTokens = (files, rules, exactFilesByID) => inspect(files, { forbiddenSourceTokens: rules }, { exactFilesByID });
const checkForbiddenGoImports = (files, rules) => inspect(files, { forbiddenGoImports: rules });
const checkOwnerPortOnlyImports = (files, rules) => inspect(files, { ownerPortOnlyImports: rules });
const checkGoImportAllowlists = (files, rules) => inspect(files, { goImportAllowlists: rules });

export function registerBoundariesTests() {
  test("backend policy: module relation ownership is closed and external grants are read only", () => {
    const manifest = repositoryPolicy();
    const rules = manifest.sqlTableAllowlists.filter((rule) => ["entities-merge-relation-ownership", "reporting-root-relation-ownership", "revisions-history-and-envelope-tables-only"].includes(rule.id));
    assert.equal(rules.length, 3);
    const evaluate = (relative, sql) => inspect([{relative, content: `package fixture\nconst query = ${JSON.stringify(sql)}`}], {sqlTableAllowlists: rules, sqlTableAccess: manifest.sqlTableAccess});
    for (const [file, sql] of [
      ["internal/modules/entities/merge/renamed.go", 'UPDATE /* spaced */ "public"."entity_mentions" SET row_version=2'],
      ["internal/modules/entities/merge/renamed.go", 'WITH changed AS (DELETE FROM host_grid_projection RETURNING *) SELECT * FROM changed'],
      ["internal/modules/reporting/renamed.go", 'select * from "public"."records"'],
      ["internal/modules/reporting/renamed.go", 'WITH x AS (SELECT * FROM reference_pack_unknown) SELECT * FROM x'],
      ["internal/modules/revisions/extracted.go", 'SELECT * FROM source'],
      ["internal/modules/reporting/renamed.go", 'SELECT * FROM %s'],
    ]) assert.ok(evaluate(file,sql).some((entry) => entry.code === "sql_table_allowlist"), sql);
    for (const [file, sql] of [
      ["internal/modules/entities/merge/renamed.go", 'SELECT * FROM entity_aliases'],
      ["internal/modules/entities/merge/admission.go", 'SELECT * FROM records'],
      ["internal/modules/reporting/renamed.go", 'WITH records AS (SELECT * FROM reporting_snapshots) SELECT * FROM records'],
      ["internal/modules/reporting/store.go", 'SELECT * FROM route_idempotency'],
      ["internal/modules/reporting/graph_source.go", 'SELECT * FROM jobs'],
      ["internal/modules/revisions/extracted.go", 'SELECT * FROM change_set_mutations'],
    ]) assert.deepEqual(evaluate(file,sql),[],sql);
    for (const table of ["jobs", "route_idempotency", "pg_class", "report_composition_preview_attempts"]) {
      const file = table === "route_idempotency" ? "store.go" : "cutover.go";
      for (const [sql,code] of [[`UPDATE ${table} SET id=1`,"sql_table_write_access"],[`SELECT * FROM ${table} x FOR UPDATE OF x`,"sql_table_lock_access"]])
        assert.ok(evaluate(`internal/modules/reporting/${file}`,sql).some((entry)=>entry.code===code),sql);
      assert.ok(evaluate("internal/modules/reporting/renamed.go",`SELECT * FROM ${table}`).some((entry)=>entry.code==="sql_table_read_access"));
    }
    assert.ok(evaluate("internal/modules/entities/merge/renamed.go","SELECT * FROM records").some((entry)=>entry.code==="sql_table_read_access"));
    assert.deepEqual(evaluate("internal/modules/reporting/renamed.go","SELECT 'FROM records' /* JOIN timeline_events */ FROM reporting_snapshots"),[]);
  });

  test("backend policy: retired Collaboration lifecycle names use identifier semantics", () => {
    const rules = repositoryPolicy().forbiddenSourceTokens.filter((rule) =>
      rule.id === "collaboration-no-retired-mutable-runtime-lifecycle-state");
    assert.equal(rules.length, 1);
    for (const [content, count] of [
      ["func ConfigureTelemetry() {}", 1],
      ['const example = "ConfigureTelemetry()" // collaboration_dispatcher', 0],
    ]) {
      assert.equal(checkForbiddenSourceTokens([{
        relative: "internal/platform/jobs/fixture_test.go", content: `package jobs\n${content}`,
      }], rules).length, count);
    }
  });
  test('backend policy: Records ownership', () => {
    const manifest = repositoryPolicy();
  const recordsRule = manifest.sqlTableAccess.find(
    (rule) => rule.id === "records-current-envelope-access",
  );
  if (!recordsRule) {
    throw new Error("records-current-envelope-access boundary rule is required");
  }
  const fixtureFiles = [
    {
      label: "records owner write",
      file: {
        relative: "internal/modules/records/fixture_store.go",
        content: "package records\nconst sql = `UPDATE records SET row_version = row_version + 1`",
      },
      wantCode: null,
    },
    {
      label: "revisions write",
      file: {
        relative: "internal/modules/revisions/fixture_store.go",
        content: "package revisions\nconst sql = `UPDATE records SET row_version = 2`",
      },
      wantCode: "sql_table_write_access",
    },
    {
      label: "unapproved source-owner read join",
      file: {
        relative: "internal/modules/entities/fixture_query.go",
        content: "package entities\nconst sql = `SELECT r.record_id FROM records r`",
      },
      wantCode: "sql_table_read_access",
    },
  ];
  for (const fixture of fixtureFiles) {
    const violations = checkSQLTableAccess([fixture.file], [recordsRule]);
    if (fixture.wantCode === null && violations.length !== 0) {
      throw new Error(`${fixture.label} boundary fixture must pass`);
    }
    if (
      fixture.wantCode !== null &&
      (violations.length !== 1 || violations[0].code !== fixture.wantCode)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${fixture.wantCode}`);
    }
  }


  });
  test('backend policy: Projection storage and support scopes', () => {
    const manifest = repositoryPolicy();
  const projectionRules = manifest.sqlTableAccess.filter((rule) =>
    rule.id.endsWith("-projection-storage-access"),
  );
  const timelineProjectionRule = projectionRules.find(
    (rule) => rule.table === "timeline_grid_projection",
  );
  const assessmentProjectionRule = projectionRules.find(
    (rule) => rule.table === "assessment_grid_projection",
  );
  const projectionFixtures = [
    {
      label: "private projection query read",
      scope: "production",
      rule: timelineProjectionRule,
      file: {
        relative: "internal/modules/projections/internal/queryengine/timeline.go",
        content: "package queryengine\nconst sql = `SELECT record_id FROM timeline_grid_projection`",
      },
      wantCode: null,
    },
    {
      label: "source-owner projection read",
      scope: "production",
      rule: timelineProjectionRule,
      file: {
        relative: "internal/modules/timeline/store.go",
        content: "package timeline\nconst sql = `SELECT record_id FROM timeline_grid_projection`",
      },
      wantCode: "sql_table_read_access",
    },
    {
      label: "Timeline test-support projection read",
      scope: "test_fixture",
      rule: timelineProjectionRule,
      file: {
        relative: "internal/modules/timeline/testsupport/asserttest/assertions.go",
        content: "package asserttest\nconst sql = `SELECT COUNT(*) FROM timeline_grid_projection`",
      },
      wantCode: null,
    },
    {
      label: "Projections test-support assessment projection write",
      scope: "test_fixture",
      rule: assessmentProjectionRule,
      file: {
        relative: "internal/modules/projections/testsupport/capability.go",
        content: "package testsupport\nconst sql = `INSERT INTO assessment_grid_projection (record_id) VALUES ($1)`",
      },
      wantCode: null,
    },
    {
      label: "Projections test-support path is not a production permission",
      scope: "production",
      rule: assessmentProjectionRule,
      file: {
        relative: "internal/modules/projections/testsupport/capability.go",
        content: "package testsupport\nconst sql = `INSERT INTO assessment_grid_projection (record_id) VALUES ($1)`",
      },
      wantCode: "sql_table_write_access",
    },
    {
      label: "unapproved projection fixture write",
      scope: "test_fixture",
      rule: assessmentProjectionRule,
      file: {
        relative: "internal/modules/entities/testsupport/fixture.go",
        content: "package testsupport\nconst sql = `DELETE FROM assessment_grid_projection`",
      },
      wantCode: "sql_table_write_access",
    },
  ];
  for (const fixture of projectionFixtures) {
    if (!fixture.rule) {
      throw new Error(`${fixture.label} boundary rule is required`);
    }
    const violations = checkSQLTableAccess([fixture.file], [fixture.rule], fixture.scope);
    if (fixture.wantCode === null && violations.length !== 0) {
      throw new Error(`${fixture.label} boundary fixture must pass`);
    }
    if (
      fixture.wantCode !== null &&
      (violations.length !== 1 || violations[0].code !== fixture.wantCode)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${fixture.wantCode}`);
    }
  }


  });
  test('backend policy: Jobs storage and locks', () => {
    const manifest = repositoryPolicy();
  const jobsStorageRule = manifest.sqlTableAccess.find(
    (rule) => rule.id === "jobs-storage-owner-write-and-lock-access",
  );
  const jobsForeignStorageRuleIDs = [
    "jobs-no-auth-route-idempotency-storage-access",
    "jobs-no-extension-cancellation-storage-access",
    "jobs-no-collaboration-intent-storage-access",
  ];
  const jobsForeignStorageRules = jobsForeignStorageRuleIDs.map((id) => {
    const rule = manifest.sqlTableAccess.find((candidate) => candidate.id === id);
    if (!rule) {
      throw new Error(`${id} boundary rule is required`);
    }
    return rule;
  });
  if (!jobsStorageRule) {
    throw new Error("jobs-storage-owner-write-and-lock-access boundary rule is required");
  }
  const jobsStorageFixtures = [
    {
      label: "Jobs owner write",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content: "package jobs\nconst sql = `UPDATE jobs SET status = 'running'`",
      },
      rule: jobsStorageRule,
      wantCode: null,
    },
    {
      label: "consumer Jobs read",
      file: {
        relative: "internal/modules/jobapi/fixture.go",
        content: "package jobapi\nconst sql = `SELECT status FROM jobs`",
      },
      rule: jobsStorageRule,
      wantCode: null,
    },
    {
      label: "non-owner Jobs write",
      file: {
        relative: "internal/modules/imports/fixture.go",
        content: "package imports\nconst sql = `UPDATE jobs SET status = 'failed'`",
      },
      rule: jobsStorageRule,
      wantCode: "sql_table_write_access",
    },
    {
      label: "non-owner Jobs transition lock",
      file: {
        relative: "internal/platform/extensionstore/fixture.go",
        content: "package extensionstore\nconst sql = `SELECT status FROM jobs FOR UPDATE`",
      },
      rule: jobsStorageRule,
      wantCode: "sql_table_lock_access",
    },
    {
      label: "auth owner adapter storage access",
      file: {
        relative: "internal/app/server/job_owner_ports.go",
        content: "package server\nconst sql = `SELECT response_json FROM route_idempotency`",
      },
      rule: jobsForeignStorageRules[0],
      wantCode: null,
    },
    {
      label: "Jobs auth storage access",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content: "package jobs\nconst sql = `SELECT response_json FROM route_idempotency`",
      },
      rule: jobsForeignStorageRules[0],
      wantCode: "sql_table_read_access",
    },
    {
      label: "Jobs extension cancellation storage access",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content:
          "package jobs\nconst sql = `INSERT INTO extension_job_cancellation_observations (job_id) VALUES ($1)`",
      },
      rule: jobsForeignStorageRules[1],
      wantCode: "sql_table_write_access",
    },
    {
      label: "Jobs Collaboration storage access",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content: "package jobs\nconst sql = `INSERT INTO collaboration_event_intents (intent_key) VALUES ($1)`",
      },
      rule: jobsForeignStorageRules[2],
      wantCode: "sql_table_write_access",
    },
    {
      label: "Jobs test storage access",
      file: {
        relative: "internal/platform/jobs/fixture_test.go",
        content: "package jobs\nconst sql = `UPDATE route_idempotency SET status_code = 200`",
      },
      rule: jobsForeignStorageRules[0],
      wantCode: null,
    },
    {
      label: "Jobs migration write",
      file: {
        relative: "db/migrations/fixture.sql",
        content: "UPDATE jobs SET job_kind = extension_job_kind;",
      },
      rule: jobsStorageRule,
      wantCode: null,
    },
  ];
  for (const fixture of jobsStorageFixtures) {
    const violations = checkSQLTableAccess([fixture.file], [fixture.rule]);
    if (fixture.wantCode === null && violations.length !== 0) {
      throw new Error(`${fixture.label} boundary fixture must pass`);
    }
    if (
      fixture.wantCode !== null &&
      (violations.length !== 1 ||
        violations[0].code !== fixture.wantCode ||
        violations[0].rule_id !== fixture.rule.id)
    ) {
      throw new Error(
        `${fixture.label} boundary fixture must fail with ${fixture.rule.id}/${fixture.wantCode}`,
      );
    }
  }


  });
  test('backend policy: Jobs capabilities', () => {
    const manifest = repositoryPolicy();
  const jobsCapabilityRule = manifest.forbiddenSourceTokens.find(
    (rule) => rule.id === "jobs-consumers-use-narrow-capabilities",
  );
  if (!jobsCapabilityRule) {
    throw new Error("jobs-consumers-use-narrow-capabilities boundary rule is required");
  }
  const jobsCapabilityFixtures = [
    {
      label: "application Jobs composition",
      file: {
        relative: "internal/app/server/fixture.go",
        content: "package server\ntype dependencies struct { Jobs *jobs.Manager }",
      },
      wantViolation: false,
    },
    {
      label: "consumer broad Jobs manager",
      file: {
        relative: "internal/modules/imports/fixture.go",
        content: "package imports\ntype store struct { jobs *jobs.Manager }",
      },
      wantViolation: true,
    },
    {
      label: "consumer raw Jobs transition lock",
      file: {
        relative: "internal/modules/imports/fixture.go",
        content: "package imports\nfunc lock() { jobs.LockTransitionTx(nil, nil, id) }",
      },
      wantViolation: true,
    },
    {
      label: "consumer test broad Jobs manager",
      file: {
        relative: "internal/modules/imports/fixture_test.go",
        content: "package imports\ntype fixture struct { jobs *jobs.Manager }",
      },
      wantViolation: false,
    },
  ];
  for (const fixture of jobsCapabilityFixtures) {
    const violations = checkForbiddenSourceTokens([fixture.file], [jobsCapabilityRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 || violations[0].code !== "forbidden_source_token")
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with the capability rule`);
    }
  }


  });
  test('backend policy: Jobs peer imports', () => {
    const manifest = repositoryPolicy();
  const jobsPeerImportRule = manifest.forbiddenGoImports.find(
    (rule) => rule.id === "jobs-use-owner-transaction-ports-not-peer-persistence",
  );
  if (!jobsPeerImportRule) {
    throw new Error(
      "jobs-use-owner-transaction-ports-not-peer-persistence boundary rule is required",
    );
  }
  const jobsPeerImportFixtures = [
    {
      label: "application Auth owner adapter",
      file: {
        relative: "internal/app/server/job_owner_ports.go",
        content:
          'package server\nimport "github.com/JochiRaider/cartulary/internal/platform/authn"',
      },
      wantViolation: false,
    },
    {
      label: "Jobs peer persistence import",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content:
          'package jobs\nimport "github.com/JochiRaider/cartulary/internal/platform/authn"',
      },
      wantViolation: true,
    },
  ];
  for (const fixture of jobsPeerImportFixtures) {
    const violations = checkForbiddenGoImports([fixture.file], [jobsPeerImportRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 ||
        violations[0].code !== "forbidden_go_import" ||
        violations[0].rule_id !== jobsPeerImportRule.id)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with the peer import rule`);
    }
  }


  });
  test('backend policy: Delete/restore metadata', () => {
    const manifest = repositoryPolicy();
  const genericDeleteRestoreRule = manifest.forbiddenSourceTokens.find(
    (rule) => rule.id === "delete-restore-sources-no-generic-sql-metadata",
  );
  if (!genericDeleteRestoreRule) {
    throw new Error("delete-restore-sources-no-generic-sql-metadata boundary rule is required");
  }
  const genericProvider = {
    relative: "internal/modules/artifacts/internal/providers/deleterestore/fixture.go",
    content: "package deleterestore\ntype TableProvider struct { SourceTable string }",
  };
  const genericViolations = checkForbiddenSourceTokens(
    [genericProvider],
    [genericDeleteRestoreRule],
  );
  if (
    genericViolations.length !== 2 ||
    genericViolations.some((entry) => entry.code !== "forbidden_source_token")
  ) {
    throw new Error("generic delete/restore provider boundary fixture must fail closed");
  }


  });
  test('backend policy: Delete/restore contracts', () => {
    const manifest = repositoryPolicy();
  const deleteRestoreContractRule = manifest.ownerPortOnlyImports.find(
    (rule) => rule.id === "revisions-delete-restore-source-contract-import",
  );
  if (!deleteRestoreContractRule) {
    throw new Error("revisions-delete-restore-source-contract-import boundary rule is required");
  }
  const contractImport =
    'github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract';
  const contractFixtures = [
    {
      label: "source owner adapter contract import",
      file: {
        relative: "internal/modules/artifacts/internal/providers/deleterestore/provider.go",
        content: `package deleterestore\nimport "${contractImport}"`,
      },
      wantViolation: false,
    },
    {
      label: "unapproved contract import",
      file: {
        relative: "internal/modules/workbook/fixture.go",
        content: `package workbook\nimport "${contractImport}"`,
      },
      wantViolation: true,
    },
  ];
  for (const fixture of contractFixtures) {
    const violations = checkOwnerPortOnlyImports([fixture.file], [deleteRestoreContractRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
  }


  });
  test('backend policy: Config kernel imports', () => {
    const manifest = repositoryPolicy();
  const configImportRule = manifest.goImportAllowlists.find(
    (rule) => rule.id === "platform-config-kernel-internal-imports",
  );
  if (!configImportRule) {
    throw new Error("platform-config-kernel-internal-imports boundary rule is required");
  }
  const configImportFixtures = [
    {
      label: "platform-neutral external import",
      imported: "github.com/BurntSushi/toml",
      wantViolation: false,
    },
    {
      label: "application import",
      imported: "github.com/JochiRaider/cartulary/internal/app/configassembly",
      wantViolation: true,
    },
    {
      label: "module owner import",
      imported: "github.com/JochiRaider/cartulary/internal/modules/extensions",
      wantViolation: true,
    },
    {
      label: "platform owner import",
      imported: "github.com/JochiRaider/cartulary/internal/platform/telemetry",
      wantViolation: true,
    },
  ];
  for (const fixture of configImportFixtures) {
    const file = {
      relative: "internal/platform/config/fixture.go",
      content: `package config\nimport "${fixture.imported}"`,
    };
    const violations = checkGoImportAllowlists([file], [configImportRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} config-kernel boundary fixture produced an unexpected result`);
    }
  }


  });
  test('backend policy: Config registry', () => {
    const manifest = repositoryPolicy();
  const configRegistryRule = manifest.forbiddenSourceTokens.find(
    (rule) => rule.id === "platform-config-no-owner-profile-registry",
  );
  if (!configRegistryRule) {
    throw new Error("platform-config-no-owner-profile-registry boundary rule is required");
  }
  const profileRegistry = {
    relative: "internal/platform/config/fixture.go",
    content: 'package config\nvar claimPaths = []string{"import.claimed"}',
  };
  const profileRegistryViolations = checkForbiddenSourceTokens(
    [profileRegistry],
    [configRegistryRule],
  );
  if (
    profileRegistryViolations.length !== 1 ||
    profileRegistryViolations[0].code !== "forbidden_source_token"
  ) {
    throw new Error("platform config profile-registry boundary fixture must fail closed");
  }


  });
  test('backend policy: Config contracted surfaces', () => {
    const manifest = repositoryPolicy();
  const contractedConfigRuleIDs = [
    "platform-config-no-retired-document-snapshot-access",
    "platform-config-no-in-memory-toml-snapshot-admission",
    "configassembly-no-in-memory-admission-definition",
    "application-no-configassembly-in-memory-admission-callers",
    "configassembly-deployment-no-wire-tags",
  ];
  const contractedConfigRules = contractedConfigRuleIDs.map((id) => {
    const rule = manifest.forbiddenSourceTokens.find((candidate) => candidate.id === id);
    if (!rule) {
      throw new Error(`${id} boundary rule is required`);
    }
    return rule;
  });
  const retiredConfigFixtures = [
    {
      label: "retired full-document source",
      relative: "internal/platform/config/source.go",
      content: "package config\ntype Source interface { Decode(path []string, target any) error }",
      ruleID: "platform-config-no-retired-document-snapshot-access",
    },
    {
      label: "retired reflective owner clone",
      relative: "internal/platform/config/snapshot.go",
      content: "package config\nfunc cloneConfig(value any) any { return value }",
      ruleID: "platform-config-no-retired-document-snapshot-access",
    },
    {
      label: "retired in-memory TOML snapshot loading",
      relative: "internal/platform/config/snapshot_test.go",
      content: "package config\nfunc fixture() { LoadSnapshotFromTOML(nil, Catalog{}) }",
      ruleID: "platform-config-no-in-memory-toml-snapshot-admission",
    },
    {
      label: "retired configassembly admission definition",
      relative: "internal/app/configassembly/configuration.go",
      content: "package configassembly\nfunc Admit(value Deployment) (Loaded, error) { return Loaded{}, nil }",
      ruleID: "configassembly-no-in-memory-admission-definition",
    },
    {
      label: "retired configassembly admission caller",
      relative: "internal/app/server/runtime_test.go",
      content: "package server\nfunc fixture() { _, _ = configassembly.Admit(configassembly.Deployment{}) }",
      ruleID: "application-no-configassembly-in-memory-admission-callers",
    },
    {
      label: "wire-tagged application deployment projection",
      relative: "internal/app/configassembly/deployment.go",
      content: 'package configassembly\ntype Deployment struct { Profile string `toml:"deployment_profile"` }',
      ruleID: "configassembly-deployment-no-wire-tags",
    },
  ];
  for (const fixture of retiredConfigFixtures) {
    const rule = contractedConfigRules.find((candidate) => candidate.id === fixture.ruleID);
    const violations = checkForbiddenSourceTokens(
      [{ relative: fixture.relative, content: fixture.content }],
      [rule],
    );
    if (violations.length !== 1 || violations[0].code !== "forbidden_source_token") {
      throw new Error(`${fixture.label} boundary fixture must fail closed`);
    }
  }

  const supportedConfigFixtures = [
    {
      relative: "internal/platform/config/catalog.go",
      content:
        "package config\ntype NamespaceDecoder interface { Decode(target any) error }\ntype NamespacePresence interface { Defined(path ...string) bool }",
    },
    {
      relative: "internal/platform/config/config.go",
      content:
        'package config\nimport "reflect"\nfunc applyOverlay(target reflect.Value) reflect.Value { return target }',
    },
    {
      relative: "internal/app/configassembly/configuration.go",
      content:
        "package configassembly\nvar owner = Definition[Configuration]{Clone: func(value Configuration) Configuration { return value }}",
    },
  ];
  const supportedConfigViolations = checkForbiddenSourceTokens(
    supportedConfigFixtures,
    contractedConfigRules,
  );
  if (supportedConfigViolations.length !== 0) {
    throw new Error("supported config seams must remain accepted by contracted boundary rules");
  }
  });
  test('backend policy: ImportsOwnerFacadeBoundary', () => {
    const manifest = repositoryPolicy();
  const rule = manifest.forbiddenGoImports.find(
    (candidate) => candidate.id === "imports-use-owner-facades-not-peer-implementations",
  );
  if (!rule) {
    throw new Error("imports-use-owner-facades-not-peer-implementations boundary rule is required");
  }
  const fixtures = [
    {
      label: "Imports capability imports",
      file: {
        relative: "internal/modules/imports/future_source_capability.go",
        content:
          'package imports\nimport "github.com/JochiRaider/cartulary/internal/modules/incidents/admission"',
      },
      wantViolation: false,
    },
    {
      label: "Imports peer source-owner implementation import",
      file: {
        relative: "internal/modules/imports/future_owner_dispatch.go",
        content:
          'package imports\nimport "github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/providers"',
      },
      wantViolation: true,
    },
    {
      label: "Imports Workbook implementation import",
      file: {
        relative: "internal/modules/imports/future_workbook_dispatch.go",
        content:
          'package imports\nimport "github.com/JochiRaider/cartulary/internal/modules/workbook"',
      },
      wantViolation: true,
    },
    {
      label: "Imports Projections implementation import",
      file: {
        relative: "internal/modules/imports/future_projection_dispatch.go",
        content:
          'package imports\nimport "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"',
      },
      wantViolation: true,
    },
  ];
  for (const fixture of fixtures) {
    const violations = checkForbiddenGoImports([fixture.file], [rule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 ||
        violations[0].code !== "forbidden_go_import" ||
        violations[0].rule_id !== rule.id)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${rule.id}`);
    }
  }
  });
  test('backend policy: ExactFileSetBoundary', () => {
    const manifest = repositoryPolicy();
  const rule = manifest.forbiddenSourceTokens.find(
    (candidate) => candidate.id === "retired-timeline-facade-calls",
  );
  if (!rule || rule.exactFileSetID !== "imports-transport-bindings") {
    throw new Error("retired-timeline-facade-calls must use the Imports transport exact file set");
  }
  const relative = "internal/modules/imports/http_handlers.go";
  const conforming = new Map([[rule.exactFileSetID, [{
    relative,
    content: "package imports\nfunc bind() { httpapi.BindOwnerRoutes(nil, nil, \"module.imports\", nil) }",
  }]]]);
  if (checkForbiddenSourceTokens([], [rule], conforming).length !== 0) {
    throw new Error("conforming Imports transport exact-file fixture must pass");
  }
  const forbidden = new Map([[rule.exactFileSetID, [{
    relative,
    content: "package imports\nfunc bind() { timeline.CreateTimelineRow(ctx) }",
  }]]]);
  const violations = checkForbiddenSourceTokens([], [rule], forbidden);
  if (
    violations.length !== 1 ||
    violations[0].code !== "forbidden_source_token" ||
    violations[0].path !== relative
  ) {
    throw new Error("forbidden Imports transport call exact-file fixture must fail closed");
  }
  });
  test('backend policy: ArtifactBoundary', () => {
    const manifest = repositoryPolicy();
  const httpRule = manifest.forbiddenGoImports.find(
    (rule) => rule.id === "artifact-subtree-no-http-or-auth-transport",
  );
  const consumerRule = manifest.forbiddenGoImports.find(
    (rule) => rule.id === "artifact-subtree-no-consumer-implementations",
  );
  const idempotencyRule = manifest.forbiddenSourceTokens.find(
    (rule) => rule.id === "artifact-mutation-idempotency-lookup-is-centralized",
  );
  if (!httpRule || !consumerRule || !idempotencyRule) {
    throw new Error("Artifact subtree boundary rules are required");
  }

  const httpFixtures = [
    {
      label: "Artifact transport-neutral source",
      file: { relative: "internal/modules/artifacts/future_surface.go", content: "package artifacts" },
      wantViolation: false,
    },
    {
      label: "Artifact future HTTP import",
      file: {
        relative: "internal/modules/artifacts/future_surface.go",
        content: 'package artifacts\nimport "net/http"',
      },
      wantViolation: true,
    },
  ];
  for (const fixture of httpFixtures) {
    const violations = checkForbiddenGoImports([fixture.file], [httpRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 ||
        violations[0].code !== "forbidden_go_import" ||
        violations[0].rule_id !== httpRule.id)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${httpRule.id}`);
    }
  }

  const consumerFixtures = [
    {
      label: "Artifact published projection contribution contract",
      file: {
        relative: "internal/modules/artifacts/workbookprojection/contribution.go",
        content:
          'package workbookprojection\nimport "github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"',
      },
      wantViolation: false,
    },
    {
      label: "Artifact future Workbook implementation import",
      file: {
        relative: "internal/modules/artifacts/future_surface.go",
        content:
          'package artifacts\nimport "github.com/JochiRaider/cartulary/internal/modules/workbook"',
      },
      wantViolation: true,
    },
  ];
  for (const fixture of consumerFixtures) {
    const violations = checkForbiddenGoImports([fixture.file], [consumerRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 ||
        violations[0].code !== "forbidden_go_import" ||
        violations[0].rule_id !== consumerRule.id)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${consumerRule.id}`);
    }
  }

  const idempotencyFixtures = [
    {
      label: "Artifact centralized idempotency lookup",
      file: {
        relative: "internal/modules/artifacts/mutation_idempotency.go",
        content: "package artifacts\nfunc lookup() { f.idempotency.Get(ctx, key, hash) }",
      },
      wantViolation: false,
    },
    {
      label: "Artifact future direct idempotency lookup",
      file: {
        relative: "internal/modules/artifacts/future_surface.go",
        content: "package artifacts\nfunc lookup() { f.idempotency.Get(ctx, key, hash) }",
      },
      wantViolation: true,
    },
  ];
  for (const fixture of idempotencyFixtures) {
    const violations = checkForbiddenSourceTokens([fixture.file], [idempotencyRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 || violations[0].code !== "forbidden_source_token")
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${idempotencyRule.id}`);
    }
  }
  });
}
