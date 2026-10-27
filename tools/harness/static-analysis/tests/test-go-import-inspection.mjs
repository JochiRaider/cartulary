import assert from "node:assert/strict";
import { chmodSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "../../workspace/child-process.mjs";
import { inspectGoImports } from "../go-import-inspection.mjs";
import { evaluateBoundaryPolicy, prepareBoundaryPolicy } from "../backend-module-boundary-check.mjs";
import { emptyPolicy } from "./backend-boundary-fixtures.mjs";
import { createFixture } from "./fixtures.mjs";

export function registerBoundariesTests() {
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
