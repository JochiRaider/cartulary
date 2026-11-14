import assert from "node:assert/strict";
import test from "node:test";
import { emptyPolicy, inspect, repositoryPolicy, repositoryPolicyInput } from "./backend-boundary-fixtures.mjs";

const checkSQLTableAccess = (files, rules, scope) => scope === "test_fixture"
  ? inspect([], { sqlTableAccess: rules }, { testSupportFiles: files }) : inspect(files, { sqlTableAccess: rules });
const checkForbiddenSourceTokens = (files, rules, exactFilesByID) => inspect(files, { forbiddenSourceTokens: rules }, { exactFilesByID });
const checkForbiddenGoImports = (files, rules) => inspect(files, { forbiddenGoImports: rules });
const checkOwnerPortOnlyImports = (files, rules) => inspect(files, { ownerPortOnlyImports: rules });
const checkGoImportAllowlists = (files, rules) => inspect(files, { goImportAllowlists: rules });

import { prepareBoundaryPolicy } from "../backend-module-boundary-check.mjs";
import { sqlStatements } from "../sql-relations.mjs";

export function registerBoundariesTests() {
  test("backend policy: Entities portable joins require exact owner shapes", () => {
    const owner="internal/modules/entities";
    const rule=repositoryPolicy().sqlTableAccess.find((entry)=>entry.id==="records-current-envelope-access");
    const grants=repositoryPolicyInput().sql_table_access.find((entry)=>entry.id===rule.id).read_allowed_statements.filter((entry)=>entry.package===owner);
    assert.equal(grants.length,2);
    assert.ok(!rule.readAllowedPaths.includes(`${owner}/*.go`));
    const scoped={...rule,readAllowedStatements:rule.readAllowedStatements.filter((entry)=>entry.package===owner)};
    const source=(entries)=>'package entities\nimport "context"\nimport "github.com/jackc/pgx/v5"\n'+entries.map((entry)=>`func ${entry.declaration}(ctx context.Context,db pgx.Tx) { db.Query(ctx,${JSON.stringify(entry.sql)}) }`).join("\n");
    const evaluate=(content,file)=>checkSQLTableAccess([{relative:`${owner}/${file}`,content}],[scoped]);
    for(const file of ["renamed.go","split.go"]) {
      assert.deepEqual(evaluate(source(grants),file),[]);
      for(const grant of grants) {
        assert.ok(evaluate(source(grants.map((entry)=>entry===grant?{...entry,sql:entry.sql.replace("$1","$9")}:entry)),file).some((entry)=>entry.code==="sql_table_read_access"));
      }
      for(const query of ["SELECT incident_id FROM records WHERE record_id=$1","SELECT record_id FROM records FOR UPDATE"]) {
        assert.ok(evaluate(source(grants)+`\nfunc bypass(ctx context.Context,db pgx.Tx) { db.Query(ctx,${JSON.stringify(query)}) }`,file).some((entry)=>entry.code==="sql_table_read_access"));
      }
    }
    assert.ok(evaluate(source(grants),"child/read.go").some((entry)=>entry.code==="sql_table_read_access"));
  });

  test("backend policy: entity delete/restore permits only exact envelope snapshots", () => {
    const owner = "internal/modules/entities/hostidentity/deleterestore";
    const rule = repositoryPolicy().sqlTableAccess.find((entry) => entry.id === "records-current-envelope-access");
    const grants = repositoryPolicyInput().sql_table_access.find((entry) => entry.id === rule.id).read_allowed_statements.filter((entry) => entry.package === owner);
    assert.equal(grants.length, 2);
    assert.ok(!rule.readAllowedPaths.includes(`${owner}/*.go`));
    const scoped = {...rule,readAllowedStatements:rule.readAllowedStatements.filter((entry) => entry.package === owner)};
    const source = (entries) => 'package deleterestore\nimport "context"\nimport "github.com/jackc/pgx/v5"\n' + entries.map((entry) => {
      const [type,method] = entry.declaration.split(".");
      return `type ${type} struct{}\nfunc (s ${type}) ${method}(ctx context.Context, db pgx.Tx) { db.Query(ctx,${JSON.stringify(entry.sql)}) }`;
    }).join("\n");
    const evaluate = (content,file="renamed.go") => checkSQLTableAccess([{relative:`${owner}/${file}`,content}], [scoped]);
    for (const file of ["renamed.go","split.go"]) {
      assert.deepEqual(evaluate(source(grants),file), []);
      for (const grant of grants) {
        assert.ok(evaluate(source(grants.map((entry) => entry===grant ? {...entry,sql:entry.sql.replace("$1","$2")} : entry)),file).some((entry) => entry.code==="sql_table_read_access"));
      }
      for (const query of ["SELECT * FROM records", "UPDATE hosts h SET row_version=r.row_version FROM records r WHERE h.record_id=r.record_id"]) {
        assert.ok(evaluate(source(grants)+`\nfunc bypass(ctx context.Context,db pgx.Tx) { db.Exec(ctx,${JSON.stringify(query)}) }`,file).some((entry) => entry.code==="sql_table_read_access"));
      }
    }
    assert.ok(evaluate(source(grants),"child/read.go").some((entry) => entry.code==="sql_table_read_access"));
  });

  test("backend policy: entity projection envelope joins require exact point/page shapes", () => {
    const owner = "internal/modules/entities/hostidentity/projectionprovider";
    const rule = repositoryPolicy().sqlTableAccess.find((entry) => entry.id === "records-current-envelope-access");
    const grants = repositoryPolicyInput().sql_table_access.find((entry) => entry.id === rule.id).read_allowed_statements.filter((entry) => entry.package === owner);
    assert.equal(grants.length, 6);
    assert.ok(!rule.readAllowedPaths.includes(`${owner}/*.go`));
    const scoped = {...rule, readAllowedStatements: rule.readAllowedStatements.filter((entry) => entry.package === owner)};
    const source = (entries) => 'package projectionprovider\nimport "context"\nimport "github.com/jackc/pgx/v5"\ntype source struct{}\n' + entries.map((entry) => entry.declaration.includes(".")
      ? `func (s *source) ${entry.declaration.split(".")[1]}(ctx context.Context, db pgx.Tx) { db.Query(ctx, ${JSON.stringify(entry.sql)}) }`
      : `const ${entry.declaration} = ${JSON.stringify(entry.sql)}`).join("\n");
    const evaluate = (content, file) => checkSQLTableAccess([{relative:`${owner}/${file}`,content}], [scoped]);
    for (const file of ["renamed.go", "split.go"]) {
      assert.deepEqual(evaluate(source(grants),file), []);
      for (const grant of grants) for (const [before,after] of [["$1","$9"],["deleted_at IS NULL","deleted_at IS NOT NULL"]].filter(([before]) => grant.sql.includes(before))) {
        assert.ok(evaluate(source(grants.map((entry) => entry === grant ? {...entry,sql:entry.sql.replace(before,after)} : entry)),file).some((entry) => entry.code === "sql_table_read_access"), `${file}: ${grant.declaration}: ${before}`);
      }
      assert.ok(evaluate(source(grants)+'\nfunc bypass(ctx context.Context,db pgx.Tx) { db.Query(ctx,"SELECT * FROM records") }',file).some((entry) => entry.code === "sql_table_read_access"));
    }
    assert.ok(evaluate(source(grants),"child/read.go").some((entry) => entry.code === "sql_table_read_access"));
  });

  test("backend policy: entity collection rollback reads envelopes only through Records", () => {
    const rules = repositoryPolicy().sqlTableAccess.filter((rule) => rule.id === "records-current-envelope-access");
    assert.equal(rules.length, 1);
    for (const name of ["renamed.go", "split.go", "child/read.go"]) {
      const relative = `internal/modules/entities/hostidentity/rollbackprovider/${name}`;
      const evaluate = (query) => checkSQLTableAccess([{relative, content: `package rollbackprovider\nconst query = ${JSON.stringify(query)}`}], rules);
      for (const query of ["SELECT incident_id, record_type FROM records WHERE record_id=$1", "SELECT record_id FROM records FOR UPDATE"]) {
        assert.ok(evaluate(query).some((entry) => entry.code === "sql_table_read_access"), `${relative}: ${query}`);
      }
      assert.ok(evaluate("DELETE FROM records").some((entry) => entry.code === "sql_table_write_access"), relative);
      assert.deepEqual(evaluate("SELECT entity_alias_id FROM entity_aliases"), []);
    }
  });

  test("backend policy: Merge reference dependencies bind through the exact application package", () => {
    const rules = repositoryPolicy().forbiddenGoImports.filter((rule) => rule.id === "entities-root-direct-consumers");
    assert.equal(rules.length, 1);
    const content = 'package assembly\nimport "github.com/JochiRaider/cartulary/internal/modules/entities"';
    for (const [relative, allowed] of [
      ["internal/app/entitymergeassembly/renamed.go", true],
      ["internal/app/entitymergeassembly/split.go", true],
      ["internal/app/entitymergeassembly/child/read.go", false],
      ["internal/app/peer/read.go", false],
      ["internal/modules/entities/merge/read.go", false],
    ]) {
      assert.equal(checkForbiddenGoImports([{relative, content}], rules).length === 0, allowed, relative);
    }
  });

  test("backend policy: HostIdentity scalar conflicts consume only the owner-local entity contract", () => {
    const rules = repositoryPolicy().forbiddenGoImports.filter((rule) => rule.id === "entities-entitycontract-remains-owner-local");
    assert.equal(rules.length, 1);
    const content = 'package boundary\nimport "github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"';
    assert.deepEqual(checkForbiddenGoImports([{relative: "internal/modules/entities/hostidentity/scalar_conflict.go", content}], rules), []);
    for (const relative of ["internal/modules/entities/hostidentity/child/scalar_conflict.go", "internal/modules/timeline/scalar_conflict.go", "internal/modules/workbook/scalar_conflict.go"]) {
      assert.ok(checkForbiddenGoImports([{relative, content}], rules).some((finding) => finding.code === "forbidden_go_import"), relative);
    }
  });
  test("backend policy: HostIdentity Records reads require exact source joins", () => {
    const rule = repositoryPolicy().sqlTableAccess.find((entry) => entry.id === "records-current-envelope-access");
    const owner = "internal/modules/entities/hostidentity";
    const grants = repositoryPolicyInput().sql_table_access.find((entry) => entry.id === rule.id).read_allowed_statements.filter((entry) => entry.package === owner);
    assert.equal(grants.length, 8);
    assert.ok(!rule.readAllowedPaths.includes(`${owner}/*.go`));
    const scoped = {...rule, readAllowedStatements: rule.readAllowedStatements.filter((entry) => entry.package === owner)};
    const source = (entries) => {
      const declarations = new Map();
      for (const entry of entries) {
        const calls = declarations.get(entry.declaration) ?? [];
        calls.push(`db.Query(ctx, ${JSON.stringify(entry.sql)})`);
        declarations.set(entry.declaration, calls);
      }
      const types = new Set([...declarations.keys()].filter((name) => name.includes(".")).map((name) => name.split(".")[0]));
      return [...types].map((name) => `type ${name} struct{}`).join("\n") + "\n" + [...declarations].map(([name, calls]) => {
        const parts = name.split(".");
        const declaration = parts.length === 2 ? `(receiver *${parts[0]}) ${parts[1]}` : name;
        return `func ${declaration}(ctx context.Context, db pgx.Tx) { ${calls.join("; ")} }`;
      }).join("\n");
    };
    const evaluate = (content, relative = `${owner}/renamed.go`) => checkSQLTableAccess([{relative, content: `package hostidentity\nimport "context"\nimport "github.com/jackc/pgx/v5"\n${content}`}], [scoped]);
    assert.deepEqual(evaluate(source(grants)), []);
    for (const query of ["SELECT record_id FROM records", "SELECT r.record_id FROM records r JOIN decisions d ON d.record_id=r.record_id FOR SHARE OF r", "DELETE FROM records", "SELECT record_id FROM records FOR UPDATE NOWAIT"]) {
      assert.ok(evaluate(`${source(grants)}\nfunc bypass(ctx context.Context, db pgx.Tx) { db.Query(ctx, ${JSON.stringify(query)}) }`).some((entry) => entry.code.startsWith("sql_table_")), query);
    }
    for (const entry of grants) {
      assert.ok(evaluate(source(grants.map((other) => other === entry ? {...entry, sql: entry.sql.replace("$1", "$9")} : other))).some((item) => item.code === "sql_table_read_access"), entry.declaration);
    }
    assert.ok(evaluate(source(grants.map((entry) => ({...entry, declaration: "foreignRead"})))).some((entry) => entry.code === "sql_table_read_access"));
    assert.ok(evaluate(source(grants), `${owner}/child/read.go`).some((entry) => entry.code === "sql_table_read_access"));
  });

  test("backend policy: Task source Records reads have exact joined shapes and no locks", () => {
    const rule = repositoryPolicy().sqlTableAccess.find((entry) => entry.id === "records-current-envelope-access");
    const owner = "internal/modules/tasksdecisions/internal/source";
    const grants = repositoryPolicyInput().sql_table_access.find((entry) => entry.id === rule.id).read_allowed_statements.filter((entry) => entry.package === owner);
    assert.equal(grants.length, 5);
    assert.ok(!rule.readAllowedPaths.includes(`${owner}/*.go`));
    const scoped = {...rule, readAllowedStatements: rule.readAllowedStatements.filter((entry) => entry.package === owner)};
    const source = (entries) => {
      const declarations = new Map();
      for (const entry of entries) {
        const calls = declarations.get(entry.declaration) ?? [];
        calls.push(`db.Query(ctx, ${JSON.stringify(entry.sql)})`);
        declarations.set(entry.declaration, calls);
      }
      return [...declarations].map(([name, calls]) => `func ${name}(ctx context.Context, db pgx.Tx) { ${calls.join("; ")} }`).join("\n");
    };
    const evaluate = (content, relative = `${owner}/renamed.go`) => checkSQLTableAccess([{relative, content: `package source\nimport "context"\nimport "github.com/jackc/pgx/v5"\n${content}`}], [scoped]);
    assert.deepEqual(evaluate(source(grants)), []);
    for (const query of ["SELECT record_id FROM records", "SELECT r.record_id FROM records r JOIN decisions d ON d.record_id=r.record_id FOR SHARE OF r", "DELETE FROM records", "SELECT record_id FROM records FOR UPDATE NOWAIT"]) {
      assert.ok(evaluate(`${source(grants)}\nfunc bypass(ctx context.Context, db pgx.Tx) { db.Query(ctx, ${JSON.stringify(query)}) }`).some((entry) => entry.code.startsWith("sql_table_")), query);
    }
    for (const entry of grants) {
      assert.ok(evaluate(source(grants.map((other) => other === entry ? {...entry, sql: entry.sql.replace("$1", "$9")} : other))).some((item) => item.code === "sql_table_read_access"), entry.declaration);
    }
    assert.ok(evaluate(source(grants), `${owner}/child/read.go`).some((entry) => entry.code === "sql_table_read_access"));
  });

  test("backend policy: Task portable envelope reads use the consumer contract", () => {
    const manifest = repositoryPolicy();
    for (const [id, consumer] of [
      ["tasks-decisions-incident-bundle-provider-exact-imports", "providers/incidentbundle"],
      ["tasks-decisions-private-source-exact-imports", "source"],
    ]) {
      const rules = manifest.goImportAllowlists.filter((rule) => rule.id === id);
      assert.equal(rules.length, 1);
      for (const file of ["renamed.go", "split.go", "child/read.go"]) {
        const relative = `internal/modules/tasksdecisions/internal/${consumer}/${file}`;
        for (const [suffix, allowed] of [
          ["tasksdecisions/internal/envelopefacts", true],
          ["records", false],
          ["records/enveloperead", false],
          ["tasksdecisions/internal/envelopefacts/child", false],
        ]) {
          const content = `package source\nimport "github.com/JochiRaider/cartulary/internal/modules/${suffix}"`;
          assert.equal(checkGoImportAllowlists([{relative, content}], rules).length === 0, allowed, `${relative}: ${suffix}`);
        }
      }
    }
  });

  test("backend policy: Task and Decision inverses consume active Links through ports", () => {
    const rules = repositoryPolicy().sqlTableAccess.filter((rule) => rule.id === "tasksdecisions-inverse-link-facts-port");
    assert.equal(rules.length, 1);
    const evaluate = (relative, query) => checkSQLTableAccess([{ relative, content: `package inverse\nconst query = ${JSON.stringify(query)}` }], rules);
    for (const name of ["renamed.go", "split.go", "child/read.go"]) {
      const relative = `internal/modules/tasksdecisions/internal/providers/rollback/${name}`;
      for (const query of ["SELECT count(*) FROM active_record_links_v1", "SELECT d.status,(SELECT count(*) FROM active_record_links_v1 l WHERE l.dst_record_id=d.record_id) FROM decisions d"]) {
        assert.ok(evaluate(relative, query).some((entry) => entry.code === "sql_table_read_access"), `${relative}: ${query}`);
      }
      assert.deepEqual(evaluate(relative, "SELECT status FROM decisions"), []);
    }
    assert.deepEqual(evaluate("internal/modules/links/facts/renamed.go", "SELECT * FROM active_record_links_v1"), []);
  });

  test("backend policy: Reference Data fixture imports require explicit test package authority", () => {
    const manifest = repositoryPolicy();
    const rule = manifest.ownerPortOnlyImports.find((entry) => entry.id === "reference-data-facade-import");
    assert.ok(rule);
    assert.deepEqual(rule.allowedImporters, []);
    for (const name of ["renamed", "split"]) for (const [receiver, suffix, isTest, allowed] of [
      ["internal/modules/incidentbundles", "", true, true],
      ["internal/modules/indicators", "/testsupport", true, true],
      ["internal/modules/revisions", "/testsupport", true, true],
      ["internal/modules/indicators", "/testsupport", false, false],
      ["internal/modules/revisions", "/testsupport", false, false],
      ["internal/modules/indicators/child", "/testsupport", true, false],
      ["internal/modules/peer", "/testsupport", true, false],
      ["internal/modules/revisions", "/testsupport/private", true, false],
      ["internal/modules/revisions", "/testsupport/fixturehttp", true, false],
      ["internal/modules/incidentbundles", "/testsupport/fixturehttp", true, true],
      ["internal/modules/incidentbundles", "/testsupport/fixturehttp", false, false],
      ["internal/app/recoveryassembly", "/recoverycontribution", false, true],
      ["internal/modules/recovery", "/recoverycontribution", false, false],
      ["internal/modules/reference_data/contractadmission", "/internal/packformat", false, true],
      ["internal/modules/reporting", "/internal/packformat", false, false],
      ["internal/app/referenceassembly", "/contractadmission", false, true],
      ["internal/app/referenceassembly/child", "/contractadmission", false, false],
    ]) {
      const relative = `${receiver}/${name}${isTest ? "_test" : ""}.go`;
      const files = [{ relative, content: `package fixture\nimport "${rule.import}${suffix}"` }];
      assert.equal(checkOwnerPortOnlyImports(files, [rule]).length === 0, allowed, `${relative}: ${suffix}`);
    }
    const productionRule = manifest.forbiddenGoImports.find((entry) => entry.id === "production-no-reference-pack-fixtures");
    assert.ok(productionRule);
    for (const relative of ["internal/modules/indicators/testsupport/fixtures.go", "internal/modules/indicators/testsupport/split.go", "internal/modules/peer/read.go"]) {
      const files = [{ relative, content: `package fixture\nimport "${rule.import}/testsupport"` }];
      assert.ok(checkForbiddenGoImports(files, [productionRule]).some((entry) => entry.code === "forbidden_go_import"), relative);
    }
    assert.deepEqual(checkForbiddenGoImports([{ relative: "internal/modules/indicators/fixture_test.go", content: `package fixture\nimport "${rule.import}/testsupport"` }], [productionRule]), []);
  });

  test("backend policy: source snapshot contract grants survive file changes without wider authority", () => {
    const rule = repositoryPolicy().ownerPortOnlyImports.find((entry) => entry.id === "revisions-delete-restore-source-contract-import");
    assert.ok(rule);
    assert.deepEqual(rule.allowedImporters, []);
    const evaluate = (relative, imported = rule.import) => checkOwnerPortOnlyImports([{ relative, content: `package fixture\nimport "${imported}"` }], [rule]);
    for (const name of ["renamed.go", "split.go"]) {
      for (const receiver of ["artifacts/internal/providers/deleterestore", "assessments/internal/providers/deleterestore", "entities/hostidentity/deleterestore", "evidence/internal/providers/deleterestore", "evidence/internal/sourcevalue", "indicators/internal/providers/deleterestore", "indicators/internal/sourcevalue", "parties/internal/providers/deleterestore", "tasksdecisions/internal/providers/deleterestore", "timeline/deleterestore", "timeline/internal/sourcevalue", "revisions"]) {
        const relative = `internal/modules/${receiver}/${name}`;
        assert.deepEqual(evaluate(relative), []);
        assert.ok(evaluate(relative, `${rule.import}/private`).length > 0);
        assert.ok(evaluate(`internal/modules/${receiver}/child/${name}`).length > 0);
      }
    }
    assert.deepEqual(evaluate("internal/modules/revisions/renamed_test.go"), []);
    for (const relative of ["internal/modules/evidence/internal/sourcevalue/renamed_test.go", "internal/modules/artifacts/internal/providers/deleterestore/split_test.go", "internal/modules/revisions/peer/read.go", "internal/modules/evidence/internal/peer/read.go", "internal/app/revisionassembly/read.go"]) {
      assert.ok(evaluate(relative).length > 0, relative);
    }
  });

  test("backend policy: IncidentBundles grants distinguish facade and exact portable ports", () => {
    const manifest = repositoryPolicy();
    const rules = manifest.ownerPortOnlyImports.filter((rule) => ["incidentbundles-facade-import", "incident-bundle-import-finalizer-port-import"].includes(rule.id));
    assert.equal(rules.length, 2);
    for (const rule of rules) assert.deepEqual(rule.allowedImporters, []);
    const artifactRules = manifest.forbiddenGoImports.filter((rule) => rule.id === "artifact-subtree-no-consumer-implementations");
    assert.equal(artifactRules.length, 1);
    const prefix = "github.com/JochiRaider/cartulary/internal/modules/incidentbundles";
    for (const name of ["renamed", "split"]) for (const [receiver, suffix, isTest, allowed] of [
      ["internal/modules/artifacts", "/sourceport", false, true],
      ["internal/modules/artifacts", "/sourceport", true, true],
      ["internal/modules/tasksdecisions/internal/providers/incidentbundle", "/sourceport", false, true],
      ["internal/modules/tasksdecisions/internal/providers/incidentbundle", "/sourceport", true, false],
      ["internal/modules/artifacts/child", "/sourceport", false, false],
      ["internal/modules/artifacts", "/sourceport/private", false, false],
      ["internal/modules/artifacts", "", false, false],
      ["internal/modules/evidence", "", false, false],
      ["internal/modules/reporting", "/artifactport", false, true],
      ["internal/modules/reporting", "/sourceport", false, false],
      ["internal/modules/incidents", "/importfinalizerport", false, true],
      ["internal/modules/incidents", "/importfinalizerport", true, true],
      ["internal/modules/incidents/child", "/importfinalizerport", false, false],
      ["internal/app/incidentportabilityassembly", "", false, true],
      ["internal/app/incidentportabilityassembly/child", "", false, false],
      ["internal/modules/peer", "/sourceport", false, false],
    ]) {
      const files = [{ relative: `${receiver}/${name}${isTest ? "_test" : ""}.go`, content: `package fixture\nimport "${prefix}${suffix}"` }];
      assert.equal(inspect(files, { ownerPortOnlyImports: rules, forbiddenGoImports: artifactRules }).length === 0, allowed, `${receiver}: ${suffix}: ${isTest}`);
    }
  });

  test("backend policy: retained Links field facts are an exact read-only capability", () => {
    const manifest = repositoryPolicy();
    const sourceTableAccess = manifest.sourceTableAccess.filter((rule) => rule.id === "links-source-tables");
    const sqlTableAccess = manifest.sqlTableAccess.filter((rule) => rule.id === "links-retained-field-facts-read-only");
    assert.equal(sourceTableAccess.length, 1);
    assert.equal(sqlTableAccess.length, 1);
    const evaluate = (relative, query) => {
      const files = [{ relative, content: `package facts\nconst query = ${JSON.stringify(query)}` }];
      return inspect(files, { sourceTableAccess, sqlTableAccess });
    };
    for (const name of ["renamed.go", "split.go"]) {
      const file = `internal/modules/links/facts/${name}`;
      assert.deepEqual(evaluate(file, "SELECT dst_record_id FROM record_links"), []);
      for (const query of ["INSERT INTO record_links (record_link_id) VALUES ($1)", "UPDATE record_links SET deleted_at = now()", "DELETE FROM record_links"]) {
        assert.ok(evaluate(file, query).some((entry) => entry.code === "sql_table_write_access"), query);
      }
      assert.ok(evaluate(file, "SELECT * FROM record_links FOR UPDATE").some((entry) => entry.code === "sql_table_lock_access"));
    }
    for (const file of ["internal/modules/links/facts/child/read.go", "internal/modules/links/facts/testsupport/read.go", "internal/modules/links/peer/read.go", "internal/modules/peer/read.go"]) {
      assert.ok(evaluate(file, "SELECT * FROM record_links").some((entry) => entry.code === "source_table_access"), file);
    }
    assert.deepEqual(evaluate("internal/modules/links/renamed.go", "UPDATE record_links SET deleted_at = now()"), []);
  });

  test("backend policy: package capabilities survive rename and split without granting descendants", () => {
    const manifest = repositoryPolicy();
    for (const id of ["evidence-private-projection-provider-import", "indicator-projection-provider-import", "party-projection-provider-import", "tasks-decisions-projection-provider-import"]) {
      const rule = manifest.ownerPortOnlyImports.find((entry) => entry.id === id);
      assert.ok(rule, id);
      const owner = rule.import.split("/internal/providers/projection")[0].split("cartulary/")[1];
      for (const name of ["renamed.go", "split.go"]) {
        assert.deepEqual(checkOwnerPortOnlyImports([{relative: `${owner}/${name}`, content: `package fixture\nimport alias "${rule.import}"`}], [rule]), []);
      }
      for (const relative of [`${owner}/child/bypass.go`, "internal/app/projectionassembly/bypass.go", "internal/modules/peer/bypass.go"]) {
        assert.ok(checkOwnerPortOnlyImports([{relative, content: `package fixture\nimport "${rule.import}"`}], [rule]).length > 0, relative);
      }
    }
    for (const id of ["entities-projectioncontract-direct-consumers", "entities-projectionports-direct-consumers"]) {
      const rule = manifest.forbiddenGoImports.find((entry) => entry.id === id);
      for (const pattern of rule.allowedPaths) {
        const owner = pattern.slice(0, -"/*.go".length);
        for (const [relative, rejected] of [[`${owner}/renamed.go`, false], [`${owner}/split.go`, false], [`${owner}/child/bypass.go`, true], ["internal/modules/peer/use.go", true]]) {
          const violations = checkForbiddenGoImports([{relative, content: `package fixture\nimport "${rule.imports[0]}"`}], [rule]);
          assert.equal(violations.length > 0, rejected, `${id}: ${relative}`);
        }
      }
    }
    const rule = manifest.forbiddenGoImports.find((entry) => entry.id === "recovery-sqlc-confined-to-private-store-adapter");
    for (const [relative, rejected] of [["internal/modules/recovery/renamed.go", false], ["internal/modules/recovery/split.go", false], ["internal/modules/recovery/child/bypass.go", true]]) {
      const violations = checkForbiddenGoImports([{relative, content: 'package recovery\nimport sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"'}], [rule]);
      assert.equal(violations.length > 0, rejected, relative);
    }
  });

  test("backend policy: Links Records grants constrain exact SQL shapes", () => {
    const rule = repositoryPolicy().sqlTableAccess.find((entry) => entry.id === "records-current-envelope-access");
    assert.equal(rule.readAllowedStatements.filter((entry) => entry.package.startsWith("internal/modules/links")).length, 3);
    assert.ok(!rule.readAllowedPaths.some((entry) => entry.startsWith("internal/modules/links/")));
    const rawRule = {
      id: "exact-records", table: "records", scan_paths: ["internal/modules/**"],
      read_allowed_paths: [], write_allowed_paths: [],
      read_allowed_statements: [{ package: "internal/modules/links", declaration: "query", sql: "UPDATE record_links SET deleted_at = $2 WHERE EXISTS (SELECT 1 FROM public.\"records\" r WHERE r.record_id = $1 AND r.record_type = 'party')" }],
    };
    const prepare = (value) => prepareBoundaryPolicy(emptyPolicy({ sql_table_access: [value] })).sqlTableAccess;
    const rules = prepare(rawRule);
    const query = rawRule.read_allowed_statements[0].sql;
    const evaluate = (content, relative = `${rawRule.read_allowed_statements[0].package}/field_refs.go`) => inspect([{ relative, content }], { sqlTableAccess: rules });
    const go = (sql) => `package fixture\nconst query = ${JSON.stringify(sql)}`;
    assert.deepEqual(evaluate(go(query.replace('SELECT 1', 'select /* reviewed */ 1').replace('WHERE EXISTS', 'WHERE\n EXISTS'))), []);
    for (const sql of [
      query.replace("'party'", "'host'"), query.replace('$1', '$3'), query.replace(' = $1', ' <> $1'),
      query.replace('record_id = $1', 'incident_id = $1'), query.replace('SELECT 1', 'SELECT 10'),
      query + '; SELECT * FROM records', query.replace("'party')", "'party' FOR UPDATE OF r)"),
      'SELECT * FROM public."records"', 'UPDATE records SET row_version = 1',
      query.replace('SELECT 1', 'SELECT (SELECT count(*) FROM records)'),
    ]) assert.ok(evaluate(go(sql)).length > 0, sql);
    assert.deepEqual(evaluate(go(query), 'internal/modules/links/renamed.go'), []);
    assert.deepEqual(evaluate(go(query), 'internal/modules/links/split.go'), []);
    for (const file of ['internal/modules/links/child/read.go', 'internal/modules/peer/read.go']) assert.ok(evaluate(go(query), file).some(entry=>entry.code==='sql_table_read_access'));
    assert.ok(evaluate(go(query).replace('const query', 'const extracted')).some(entry=>entry.code==='sql_statement_grant_stale'));
    for (const content of [
      `package fixture\nconst query = ${JSON.stringify(query)} + fragment`,
      `package fixture\nconst query = fragment + ${JSON.stringify(query)}`,
      `package fixture\nfunc Read() { query := ${JSON.stringify(query)}; query += fragment; _ = query }`,
      `package fixture\nconst template = ${JSON.stringify(query)}\nvar query = template + fragment`,
    ]) assert.ok(evaluate(content).some(entry=>entry.code==='sql_construction_unresolved'),content);
    assert.ok(evaluate(go(query.replace('SELECT 1', 'SELECT %s'))).some(entry=>entry.code==='sql_statement_grant_stale'));
    assert.ok(evaluate('package fixture').some((entry) => entry.code === 'sql_statement_grant_stale'));
    assert.ok(inspect([], { sqlTableAccess: rules }, { completeSourceSet: true }).some((entry) => entry.code === 'sql_statement_grant_stale'));
    assert.ok(evaluate(go(query.replace("'party')", "'party' FOR UPDATE OF r)"))).some((entry) => entry.code === 'sql_table_lock_access'));
    for (const sql of ['SELECT * FROM records FOR UPDATE', 'UPDATE records SET row_version = 1', 'SELECT * FROM %s']) {
      assert.throws(() => prepare({ ...rawRule, read_allowed_statements: [{ ...rawRule.read_allowed_statements[0], sql }] }));
    }
    assert.throws(() => prepare({ ...rawRule, read_allowed_paths: ['internal/modules/links/**'] }));
    assert.throws(() => prepare({ ...rawRule, read_allowed_statements: [rawRule.read_allowed_statements[0], rawRule.read_allowed_statements[0]] }));
    assert.throws(() => prepare({ ...rawRule, read_allowed_statements: [{ package: 'internal/modules/links/**', declaration: 'query', sql: query }] }));
  });
  test("backend policy: normalized SQL preserves literals, parameters and boundaries", () => {
    const key = (sql) => sqlStatements(sql)[0].key;
    const base = "SELECT $1, 'Case', $$body$$, 10, a->>'b' FROM public.\"records\"";
    assert.equal(key(base), key(base.replace('SELECT', 'select /* nested /* comment */ fine */')));
    for (const changed of [base.replace('Case','case'), base.replace('body','BODY'), base.replace('10','1 0'), base.replace('$1','$ 1'), base.replace('->>','- >>'), base + ';', base.replace('records','Records')]) {
      assert.notEqual(key(base), key(changed));
    }
    assert.notEqual(key("SELECT E'foo\\n' FROM records"), key("SELECT E'foo\\t' FROM records"));
  });

  test("backend policy: module relation ownership is closed and external grants are read only", () => {
    const manifest = repositoryPolicy();
    const rules = manifest.sqlTableAllowlists.filter((rule) => ["entities-merge-relation-ownership", "reporting-root-relation-ownership", "revisions-history-and-envelope-tables-only"].includes(rule.id));
    assert.equal(rules.length, 3);
    const evaluate = (relative, sql) => inspect([{relative, content: `package fixture\nconst query = ${JSON.stringify(sql)}`}], {sqlTableAllowlists: rules, sqlTableAccess: manifest.sqlTableAccess});
    for (const [file, sql] of [
      ["internal/modules/entities/merge/renamed.go", 'UPDATE /* spaced */ "public"."entity_mentions" SET row_version=2'],
      ["internal/modules/entities/merge/admission.go", 'SELECT * FROM records'],
      ["internal/modules/entities/merge/renamed.go", 'WITH changed AS (DELETE FROM host_grid_projection RETURNING *) SELECT * FROM changed'],
      ["internal/modules/reporting/renamed.go", 'select * from "public"."records"'],
      ["internal/modules/reporting/renamed.go", 'WITH x AS (SELECT * FROM reference_pack_unknown) SELECT * FROM x'],
      ["internal/modules/revisions/extracted.go", 'SELECT * FROM source'],
      ["internal/modules/reporting/renamed.go", 'SELECT * FROM %s'],
    ]) assert.ok(evaluate(file,sql).some((entry) => entry.code === "sql_table_allowlist"), sql);
    for (const [file, sql] of [
      ["internal/modules/entities/merge/renamed.go", 'SELECT * FROM entity_aliases'],
      ["internal/modules/reporting/renamed.go", 'WITH records AS (SELECT * FROM reporting_snapshots) SELECT * FROM records'],
      ["internal/modules/reporting/store.go", 'SELECT * FROM route_idempotency'],
      ["internal/modules/reporting/graph_source.go", 'SELECT * FROM jobs'],
      ["internal/modules/revisions/extracted.go", 'SELECT * FROM change_set_mutations'],
    ]) assert.deepEqual(evaluate(file,sql),[],sql);
    for (const table of ["jobs", "route_idempotency", "pg_class", "report_composition_preview_attempts"]) {
      const file = table === "route_idempotency" ? "store.go" : "cutover.go";
      for (const [sql,code] of [[`UPDATE ${table} SET id=1`,"sql_table_write_access"],[`SELECT * FROM ${table} x FOR UPDATE OF x`,"sql_table_lock_access"]])
        assert.ok(evaluate(`internal/modules/reporting/${file}`,sql).some((entry)=>entry.code===code),sql);
      assert.deepEqual(evaluate("internal/modules/reporting/renamed.go",`SELECT * FROM ${table}`),[]);
      assert.deepEqual(evaluate("internal/modules/reporting/split.go",`SELECT * FROM ${table}`),[]);
      assert.ok(evaluate("internal/modules/reporting/child/read.go",`SELECT * FROM ${table}`).some((entry)=>entry.code==="sql_table_read_access"));
    }
    for (const relative of ["internal/modules/entities/merge/renamed.go", "internal/modules/entities/merge/split.go", "internal/modules/entities/merge/child/read.go"]) {
      for (const [sql,code] of [["SELECT record_id FROM records","sql_table_read_access"],["UPDATE records SET row_version=2","sql_table_write_access"],["SELECT record_id FROM records FOR UPDATE","sql_table_read_access"]]) {
        assert.ok(evaluate(relative,sql).some(entry=>entry.code===code),`${relative}: ${sql}`);
      }
    }
    assert.ok(evaluate("internal/modules/entities/merge/child/read.go","SELECT * FROM records").some(entry=>entry.code==="sql_table_read_access"));
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
        relative: "internal/modules/entities/peer/fixture_query.go",
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
test('SQL forwarding is exact declaration transport authority, never construction authority',()=>{
 const sqlForwardingCapabilities=[{package:'internal/platform/transport',declaration:'Forward',testSupport:false}];
 const file=body=>({relative:'internal/platform/transport/renamed.go',content:`package transport\nimport "github.com/jackc/pgx/v5"\nfunc Forward(tx pgx.Tx,query string){tx.Query(ctx,${body})}`});
 assert.deepEqual(inspect([file('query')],{sqlForwardingCapabilities}),[]);
 for(const body of ['query+` UNION SELECT * FROM users`','opaque(query)']) assert.ok(inspect([file(body)],{sqlForwardingCapabilities}).some(v=>v.code==='sql_construction_unresolved'));
 assert.ok(inspect([{...file('query'),relative:'internal/platform/transport/child/read.go'}],{sqlForwardingCapabilities}).some(v=>v.code==='sql_construction_unresolved'));
 assert.ok(inspect([file('query')],{},{completeSourceSet:true}).some(v=>v.code==='sql_construction_unresolved'));
 assert.ok(inspect([],{sqlForwardingCapabilities},{completeSourceSet:true}).some(v=>v.code==='stale_sql_forwarder'));
});

  test('SQL session authority requires the exact package, declaration and complete command',()=>{
    const grant={package:'internal/platform/session',declaration:'Initialize',sql:'SET ROLE cartulary_runtime'};
    const policy=prepareBoundaryPolicy(emptyPolicy({sql_control_capabilities:[grant]}));
    const file=sql=>({relative:'internal/platform/session/renamed.go',content:`package session\nimport "github.com/jackc/pgx/v5"\nfunc Initialize(tx pgx.Tx){tx.Exec(ctx,${JSON.stringify(sql)})}`});
    assert.deepEqual(inspect([file(grant.sql)],policy,{completeSourceSet:true}),[]);
    for(const sql of ['SET ROLE cartulary_schema_owner','SET ROLE cartulary_runtime;SET LOCAL row_security = off']) assert.ok(inspect([file(sql)],policy,{completeSourceSet:true}).some(v=>v.code==='sql_control_access'));
    assert.ok(inspect([{...file(grant.sql),relative:'internal/platform/session/child/read.go'}],policy,{completeSourceSet:true}).some(v=>v.code==='sql_control_access'));
  });

}
