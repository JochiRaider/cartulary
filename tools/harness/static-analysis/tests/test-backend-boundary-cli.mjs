import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import path from "node:path";
import test from "node:test";
import { mkdirSync, rmSync, symlinkSync } from "node:fs";
import * as fs from "node:fs";
import { createServer } from "node:net";
import { once } from "node:events";
import { createSourceAdmission } from "../source-admission.mjs";
import { inspectBackendBoundaries } from "../backend-module-boundary-check.mjs";
import { spawnSync } from "../../workspace/child-process.mjs";
import { createFixture, repository } from "./fixtures.mjs";
import { emptyPolicy } from "./backend-boundary-fixtures.mjs";

export function registerBoundariesTests() {
  test("source admission propagates IO failures and rejects nonregular inputs before reads",(t)=>{
    const fixture=createFixture(t);fixture.write('src/ok.go','package p');
    for(const operation of ['lstatSync','readdirSync','readFileSync']) {
      const admission=createSourceAdmission(fixture.root,{...fs,[operation](){throw Object.assign(new Error('permission denied'),{code:'EACCES'});}});
      assert.throws(()=>operation==='readdirSync'?admission.entries('src','fixture scope'):admission.file('src/ok.go','fixture scope'),/fixture scope: src.*permission denied/);
    }
    let reads=0;
    const admission=createSourceAdmission(fixture.root,{...fs,
      lstatSync(file){return file.endsWith('ok.go')?{isSymbolicLink:()=>false,isFile:()=>false,isDirectory:()=>false}:fs.lstatSync(file);},
      readFileSync(){reads++;throw new Error('must not read nonregular source');},
    });
    assert.throws(()=>admission.file('src/ok.go'),/regular file/);assert.equal(reads,0);
    const entries=createSourceAdmission(fixture.root,{...fs,readdirSync(){return [{name:'pipe.go',isSymbolicLink:()=>false,isFile:()=>false,isDirectory:()=>false}];}});
    assert.throws(()=>entries.entries('src'),/nonregular source src\/pipe.go/);
    const normal=createSourceAdmission(fixture.root);
    assert.deepEqual(normal.entries('src').map(e=>e.name),['ok.go']);
    mkdirSync(path.join(fixture.root,'empty'));
    assert.deepEqual(normal.entries('empty'),[]);
  });
  test("backend source admission rejects missing and linked inputs before completeness", async(t)=>{
    for(const scenario of ['regular','missing authored','missing migration','missing support','missing generated','linked file','dangling file','linked directory','linked root','linked ancestor','nonregular file','ordinary tmp directory','overlap']) {
      await t.test(scenario,async(t)=>{
        const fixture=createFixture(t);
        fixture.write('internal/gen/types.go','package generated');
        fixture.write('support/helper.go','package support');
        fixture.write('migration/main.go','package migration');
        fixture.write('src/read.go','package p\nimport "github.com/jackc/pgx/v5"\nfunc Read(tx pgx.Tx){tx.Exec(ctx,"DELETE FROM host_grid_projection")}');
        const policy=emptyPolicy({scan_roots:['src'],sql_table_access:[{id:'projection-storage-access',table:'host_grid_projection',scan_paths:['src/**'],read_allowed_paths:[],write_allowed_paths:[],lock_allowed_paths:[]} ]});
        policy.migration_source_boundary.scan_roots=['migration'];
        fixture.write('support.json',{schema_id:'cartulary.test_support_inventory.v1',go_support_roots:[{path:'support',runtime_scan:'excluded',support_scan:'included',service_starting:false}]});
        if(scenario.startsWith('missing')) rmSync(path.join(fixture.root,{'missing authored':'src','missing migration':'migration','missing support':'support','missing generated':'internal/gen'}[scenario]),{recursive:true});
        if(scenario==='linked file' || scenario==='dangling file') {rmSync(path.join(fixture.root,'src/read.go'));symlinkSync(path.join(fixture.root,scenario==='linked file'?'migration/main.go':'absent.go'),path.join(fixture.root,'src/read.go'));}
        if(scenario==='linked directory') symlinkSync(path.join(fixture.root,'migration'),path.join(fixture.root,'src/linked'));
        if(scenario==='linked root') {symlinkSync(path.join(fixture.root,'src'),path.join(fixture.root,'linked'));policy.scan_roots=['linked'];}
        if(scenario==='linked ancestor') {symlinkSync(fixture.root,path.join(fixture.root,'linked'));policy.scan_roots=['linked/src'];}
        if(scenario==='nonregular file') {
          const server=createServer();
          server.listen(path.join(fixture.root,'src/pipe.go'));
          await once(server,'listening');
          t.after(()=>new Promise((resolve,reject)=>server.close(error=>error?reject(error):resolve())));
        }
        if(scenario==='ordinary tmp directory') {const content='package p\nconst query="DELETE FROM host_grid_projection"';rmSync(path.join(fixture.root,'src/read.go'));fixture.write('src/tmp/read.go',content);}
        if(scenario==='overlap') {fixture.write('src/nested/read.go','package p\nconst query="DELETE FROM host_grid_projection"');policy.scan_roots.push('src/nested');}
        fixture.write('policy.json',policy);
        const inspect=()=>inspectBackendBoundaries({root:fixture.root,manifest:path.join(fixture.root,'policy.json'),supportInventory:'support.json'});
        if(['regular','ordinary tmp directory','overlap'].includes(scenario)) {
          const result=inspect();assert.ok(result.violations.some(v=>v.code==='sql_table_write_access'));
          if(scenario==='overlap') assert.equal(result.violations.filter(v=>v.path==='src/nested/read.go').length,1);
        } else {
          assert.throws(inspect,/source|root|directory|linked|ENOENT/);
          const child=spawnSync(process.execPath,[path.join(repository,'tools/harness/static-analysis/backend-module-boundary-check-cli.mjs'),'--root',fixture.root,'--manifest','policy.json','--support-inventory','support.json'],{encoding:'utf8',timeout:10000});
          assert.equal(child.status,1);assert.equal(child.stdout,'');assert.match(child.stderr,/backend module boundary check failed:/);
        }
      });
    }
  });
  test("backend CLI: success, sorted violations and invalid input preserve process contract", (t) => {
    const fixture = createFixture(t);
    fixture.write("internal/gen/fixture/types.go", "package fixture");
    const policy = emptyPolicy({ scan_roots: ["src"], forbidden_go_imports: [{
      id: "private-owner", match_kind: "subtree", imports: ["example/private"],
      scan_paths: ["src/**"], allowed_paths: [], production_only: true,
    }] });
    const support = { schema_id: "cartulary.test_support_inventory.v1", go_support_roots: [{
      path: "support", runtime_scan: "excluded", support_scan: "included", service_starting: false,
    }] };
    fixture.write("boundaries.json", policy);
    fixture.write("support.json", support);
    fixture.write("src/allowed.go", "package allowed");
    mkdirSync(path.join(fixture.root,"support"));
    const invoke = (args = []) => spawnSync(process.execPath, [
      path.join(repository, "tools/harness/static-analysis/backend-module-boundary-check-cli.mjs"),
      "--root", fixture.root, "--manifest", "boundaries.json", "--support-inventory", "support.json", ...args,
    ], {
      encoding: "utf8", timeout: 10000,
      env: { ...process.env, CARTULARY_REPO_ROOT: "/absent", BACKEND_MODULE_BOUNDARIES_MANIFEST: "/absent", TEST_SUPPORT_INVENTORY: "/absent" },
    });
    const success = invoke();
    assert.equal(success.error, undefined);
    assert.equal(success.status, 0, success.stderr);
    assert.equal(success.stderr, "");
    const summary = JSON.parse(success.stdout);
    const digest = (value) => `sha256:${createHash("sha256").update(JSON.stringify(value)).digest("hex")}`;
    assert.match(summary.checked_at, /^\d{4}-\d{2}-\d{2}T.+Z$/u);
    assert.deepEqual(summary, {
      schema_id: "cartulary.backend_module_boundary_summary.v1", checked_at: summary.checked_at,
      repo_root: fixture.root, manifest_digest: digest(policy), support_inventory_digest: digest(support),
      effective_scan_excludes: ["support/**"], dynamic_relation_sources: [], violations: [], result: "pass",
    });
    for (const name of ["z", "a"]) fixture.write(`src/${name}.go`, 'package forbidden\nimport "example/private/child"');
    const failure = invoke();
    assert.equal(failure.status, 1);
    assert.equal(failure.stderr, "");
    const failedSummary = JSON.parse(failure.stdout);
    assert.equal(failedSummary.result, "fail");
    assert.deepEqual(failedSummary.violations, ["a", "z"].map((name) => ({
      code: "forbidden_go_import", path: `src/${name}.go`, symbol_or_import: "example/private/child",
      rule_id: "private-owner", match_kind: "subtree", result: "fail",
    })));
    fixture.write("boundaries.json", "{invalid");
    const invalid = invoke();
    assert.equal(invalid.status, 1);
    assert.equal(invalid.stdout, "");
    assert.match(invalid.stderr, /^backend module boundary check failed:/u);
    const usage = invoke(["--unknown"]);
    assert.equal(usage.status, 1);
    assert.equal(usage.stdout, "");
    assert.match(usage.stderr, /^backend module boundary check failed: usage:/u);
  });
}
