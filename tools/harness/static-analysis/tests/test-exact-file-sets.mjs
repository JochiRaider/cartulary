import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";
import { normalizeExactFileSets, resolveExactFileSets } from "../exact-file-sets.mjs";
import { createContractTestContext } from "../../tests/contract-test-context.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
function assertExactFileSetContract() {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-exact-file-set."));
  const importsRoot = path.join(fixtureRoot, "internal/modules/imports");
  const handlerPath = path.join(importsRoot, "http_handlers.go");
  const definition = (paths = ["internal/modules/imports/http_handlers.go"]) => [{
      id: "imports-transport-bindings",
      paths,
      discovery: {
        scan_roots: ["internal/modules/imports"],
        tokens: ["httpapi.BindOwnerRoutes("],
        production_only: true,
      },
    }];
  try {
    mkdirSync(importsRoot, { recursive: true });
    writeFileSync(handlerPath, "package imports\nfunc bind() { httpapi.BindOwnerRoutes(nil, nil, \"module.imports\", nil) }\n");
    const resolved = resolveExactFileSets(fixtureRoot, normalizeExactFileSets(definition()));
    assert.deepEqual(resolved.get("imports-transport-bindings").map((file) => file.relative), ["internal/modules/imports/http_handlers.go"]);
    for (const invalid of [
      "/absolute.go",
      "C:/absolute.go",
      "internal/modules/imports/../routes.go",
      "internal\\modules\\imports\\routes.go",
      "internal/modules/imports/*.go",
    ]) {
      assert.throws(() => normalizeExactFileSets(definition([invalid])));
    }
    assert.throws(() => normalizeExactFileSets(definition([])), /non-empty array/u);
    assert.throws(() => normalizeExactFileSets(definition([
      "internal/modules/imports/http_handlers.go",
      "internal/modules/imports/http_handlers.go",
    ])), /unique/u);
    assert.throws(() => resolveExactFileSets(fixtureRoot, normalizeExactFileSets(definition(["internal/modules/imports/missing.go"]))), /does not exist/u);
    const stalePath = path.join(importsRoot, "stale.go");
    writeFileSync(stalePath, "package imports\n");
    assert.throws(() => resolveExactFileSets(fixtureRoot, normalizeExactFileSets(definition(["internal/modules/imports/stale.go"]))), /missing=\[internal\/modules\/imports\/http_handlers.go\].*stale=\[internal\/modules\/imports\/stale.go\]/u);
    rmSync(stalePath);
    const omittedPath = path.join(importsRoot, "new_routes.go");
    writeFileSync(omittedPath, "package imports\nfunc bindMore() { httpapi.BindOwnerRoutes(nil, nil, \"module.imports\", nil) }\n");
    assert.throws(() => resolveExactFileSets(fixtureRoot, normalizeExactFileSets(definition())), /missing=\[internal\/modules\/imports\/new_routes.go\]/u);
    rmSync(omittedPath);
    const directoryPath = path.join(importsRoot, "directory.go");
    mkdirSync(directoryPath);
    assert.throws(() => resolveExactFileSets(fixtureRoot, normalizeExactFileSets(definition(["internal/modules/imports/directory.go"]))), /regular file/u);
    rmSync(directoryPath, { recursive: true });
    const symlinkPath = path.join(importsRoot, "binding_link.go");
    symlinkSync("http_handlers.go", symlinkPath);
    assert.throws(() => resolveExactFileSets(fixtureRoot, normalizeExactFileSets(definition(["internal/modules/imports/binding_link.go"]))), /symlink/u);
  }
  finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}
export function registerBoundariesTests() {
  test("exact_boundary_file_sets: exact boundary file sets reject stale, unsafe, and omitted paths", async () => {
    assertExactFileSetContract();
  });
}
