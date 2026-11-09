import assert from "node:assert/strict";
import test from "node:test";
import { decodeGoPackages, validateModuleClosure, validateApplicationClosure } from "./closure.mjs";

const app = "github.com/JochiRaider/cartulary";
const format = `${app}/referencepacks/format`;
const context = { modulePath: format, sourceRoot: "/isolated/source", moduleCache: "/cache/modules", goRoot: "/go" };
const own = () => ({ Dir: "/isolated/source/dataset", ImportPath: `${format}/dataset`, Imports: ["encoding/json"], Module: { Path: format, Main: true }, GoFiles: ["types.go"], EmbedFiles: [] });
const dependency = () => ({ Dir: "/cache/modules/example.org/text@v1.0.0", ImportPath: "example.org/text", Module: { Path: "example.org/text", Version: "v1.0.0" }, GoFiles: ["text.go"] });

test("exported module closure accepts only its versioned independent inputs", () => {
  validateModuleClosure([own(), dependency(), { Dir: "/go/src/io", ImportPath: "io", GoFiles: ["io.go"] }], context);
  const producer = { ...context, modulePath: `${app}/referencepacks/windows` };
  validateModuleClosure([{ ...own(), Module: { Path: format, Version: "v0.1.0" }, Dir: "/cache/modules/format@v0.1.0/dataset" }], producer);
});
for (const [name, mutate] of [
  ["direct application import", pkg => { pkg.Imports.push(`${app}/internal/platform/postgres`); }],
  ["sibling producer", pkg => { pkg.Imports.push(`${app}/referencepacks/entra`); }],
  ["transitive application dependency", pkg => { pkg.ImportPath = `${app}/internal/platform/postgres`; }],
  ["module replacement", pkg => { pkg.Module.Replace = { Dir: "/application" }; }],
  ["outside source directory", pkg => { pkg.Dir = "/isolated/source-escape"; }],
  ["escaping embedded file", pkg => { pkg.EmbedFiles = ["../../application/contracts.json"]; }],
  ["prose embedding", pkg => { pkg.EmbedFiles = [["docs", "spec.json"].join("/")]; }],
  ["README embedding", pkg => { pkg.EmbedFiles = ["README.md"]; }],
  ["unversioned dependency", pkg => { pkg.Module.Main = false; }],
  ["incomplete graph", pkg => { pkg.Incomplete = true; }],
]) test(`module qualification rejects ${name}`, () => {
  const pkg = own(); mutate(pkg); assert.throws(() => validateModuleClosure([pkg], context));
});

test("application closure rejects transitive signing and producer imports", () => {
  validateApplicationClosure([{ ImportPath: `${app}/internal/modules/reference_data`, Imports: [`${format}/dataset`] }, { ImportPath: `${format}/dataset`, Imports: ["encoding/json"] }]);
  for (const forbidden of [`${format}/publication`, `${app}/referencepacks/attack`, `${app}/referencepacks/windows/internal/source`, `${app}/internal/modules/reference_data/producer`]) {
    assert.throws(() => validateApplicationClosure([{ ImportPath: "example.org/indirect", Imports: [forbidden] }]));
  }
});

test("application closure rejects indirect corpus and private-key embeddings", () => {
  for (const name of ["reference-pack-producers/attack/source.json", "dataset-v1/logical/manifest.json", "signing-context/release.json", "../../source.json", ["docs", "contract.json"].join("/"), "README.md"]) {
    assert.throws(() => validateApplicationClosure([{ ImportPath: "example.org/transitive", EmbedFiles: [name] }]));
  }
  validateApplicationClosure([{ ImportPath: `${app}/internal/platform/httpapi/webassets`, EmbedFiles: ["assets/index.html"] }]);
});

test("Go package stream decoding preserves escaped delimiters and rejects truncation", () => {
  const expected = [{ ImportPath: 'example/"quoted{}', Imports: [] }, { ImportPath: "second", Module: { Version: "v1.0.0" } }];
  assert.deepEqual(decodeGoPackages(expected.map(value => JSON.stringify(value)).join("\n")), expected);
  for (const raw of ["{", "[]broken", "}", '{"x":"unterminated', "true"]) assert.throws(() => decodeGoPackages(raw));
});

test("only compiler-generated test mains may use absolute build cache inputs", () => {
  const generated = { ...own(), Name: "main", ImportPath: `${format}/dataset.test`, GoFiles: [`/cache/build/ab/${"a".repeat(64)}-d`] };
  validateModuleClosure([generated], { ...context, buildCache: "/cache/build" });
  validateModuleClosure([{ ...generated, ImportPath: `${format}.test` }], { ...context, buildCache: "/cache/build" });
  assert.throws(() => validateModuleClosure([generated], context));
  assert.throws(() => validateModuleClosure([{ ...generated, Name: "library" }], { ...context, buildCache: "/cache/build" }));
  assert.throws(() => validateModuleClosure([{ ...generated, GoFiles: ["/application/main.go"] }], { ...context, buildCache: "/cache/build" }));
});
