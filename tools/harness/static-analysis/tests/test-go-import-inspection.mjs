import assert from "node:assert/strict";
import { chmodSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "../../workspace/child-process.mjs";
import { inspectGoImports, inspectGoDeclarations, inspectGoSQL } from "../go-import-inspection.mjs";
import { evaluateBoundaryPolicy, prepareBoundaryPolicy } from "../backend-module-boundary-check.mjs";
import { emptyPolicy } from "./backend-boundary-fixtures.mjs";
import { createFixture } from "./fixtures.mjs";

export function registerBoundariesTests() {
  test("Go declaration inspection: stable binding identities and complete bodies", () => {
    const source = 'package fixture\nconst (query = `SELECT * FROM records`; number = 1)\nvar a,b = 1,2\ntype R[T any] struct{}\nfunc (r *R[T]) Read() string { return query }\nfunc Free() {}';
    const declarations = inspectGoDeclarations([[{relative:"renamed.go",content:source}]]).get("renamed.go");
    assert.deepEqual(declarations.map((value)=>value.name), ["query", "number", "a,b", "R.Read", "Free"]);
    assert.equal(declarations[0].source, 'const query = `SELECT * FROM records`');
    assert.match(declarations[3].source, /return query/u);
    assert.throws(()=>inspectGoDeclarations([[{relative:"bad.go",content:"package p\nfunc Bad( {"}]]), /parse declarations/u);
  });

  test("Go SQL inspection: located dispositions are complete and internally consistent", (t) => {
    const fixture = createFixture(t);
    const files = [{relative:"a.go",content:"package a"}];
    const valid = {kind:"execution",text:"SELECT 1",dynamic:false,unresolved:[],line:1,column:1,disposition:"analyzed"};
    const invalid = [null, {...valid,line:0}, {...valid,column:undefined}, {...valid,extra:true}, {...valid,disposition:"transport"},
      {...valid,dynamic:true}, {...valid,unresolved:["unknown"]}, {...valid,kind:"forward"}, {...valid,kind:"closed"}, {...valid,kind:"empty"},
      {...valid,kind:"legacy"}, {...valid,dynamic:true,unresolved:[""],disposition:"rejected"}];
    for (const [index, expression] of [valid,...invalid].entries()) {
      const output = {files:[{path:"a.go",imports:[],sql:[{name:"Read",sql:[expression]}]}]};
      const file = `sql-parser-${index}.mjs`;
      fixture.write(file, `#!${process.execPath}\nprocess.stdin.resume();process.stdin.on("end",()=>process.stdout.write(${JSON.stringify(JSON.stringify(output))}));\n`);
      const binary=path.join(fixture.root,file);chmodSync(binary,0o700);
      if (index===0) assert.deepEqual(inspectGoSQL([files],{binary}).get("a.go")[0].sql,[valid]);
      else assert.throws(()=>inspectGoSQL([files],{binary}),/invalid Go SQL inspection response/u);
    }
  });

  test("Go import inspection: all declarations, build tags and source isolation", () => {
    const files = [{ relative: "z.go", content: '//go:build impossible\n\npackage fixture\nimport alias "one/owner"\nimport (_ `two/owner`; . "three/owner")\nimport ("four/owner")\n// import "fake/comment"\nvar text = `import "fake/string"`' },
      { relative: "a_test.go", content: 'package fixture\nimport "test/owner"' },
      { relative: "README.md", content: 'import "not/source"' }];
    assert.deepEqual([...inspectGoImports([files, [files[0]]])], [["a_test.go", ["test/owner"]], ["z.go", ["one/owner", "two/owner", "three/owner", "four/owner"]]]);
    const policy = prepareBoundaryPolicy(emptyPolicy({ forbidden_go_imports: [{ id: "guard", match_kind: "exact", imports: ["four/owner"], scan_paths: ["*.go"], allowed_paths: [], production_only: true }] }));
    const inputs = { files, importsByPath: inspectGoImports([files]) };
    assert.equal(evaluateBoundaryPolicy(policy, inputs).violations[0].symbol_or_import, "four/owner");
    assert.throws(() => evaluateBoundaryPolicy(policy, { files }), /missing parsed imports/u);
    assert.throws(() => inspectGoImports([[{relative:"broken.go", content:"package p\nimport ("}]]), /broken.go/u);
    assert.throws(() => inspectGoImports([files, [{...files[0], content:"package other"}]]), /conflicting import source/u);
  });
  test("Go import inspection: missing tool and incomplete or malformed batches fail", (t) => {
    const fixture = createFixture(t);
    const files = [{relative:"a.go", content:"package a"}];
    assert.throws(() => inspectGoImports([files], {binary:"relative"}), /installed absolute/u);
    assert.throws(() => inspectGoImports([files], {binary:path.join(fixture.root,"missing")}), /inspection failed/u);
    for (const [index, output] of ["not json", "{}", '{"files":[]}', '{"files":[{"path":"other.go","imports":[]}]}', '{"files":[{"path":"a.go","imports":[0]}]}'].entries()) {
      const file = `parser-${index}.mjs`;
      fixture.write(file, `#!${process.execPath}\nprocess.stdin.resume();process.stdin.on("end",()=>process.stdout.write(${JSON.stringify(output)}));\n`);
      const binary = path.join(fixture.root, file); chmodSync(binary, 0o700);
      assert.throws(() => inspectGoImports([files], { binary }), /inspection response/u);
    }
    for (const input of ['{"files":[{"path":"a.go","source":"package a"},{"path":"a.go","source":"package a"}]}', '{"files":[]}{"files":[]}', '{}', '{"files":[],"unknown":true}']) {
      const child = spawnSync(process.env.CARTULARY_SOURCECHECK_BIN, [], {input,encoding:"utf8",timeout:10000});
      assert.equal(child.status, 1, child.stderr); assert.equal(child.stdout, "");
    }
  });
}
