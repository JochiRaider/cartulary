import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, mkdirSync, writeFileSync, symlinkSync, linkSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { isolated, executableInputs } from "./isolation.mjs";
import { validateGenerationOutputs } from "./generation-outputs.mjs";
import { execFileSync } from "node:child_process";

test("filesystem capabilities reject undeclared discovery and generator inputs", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "reference-isolation-"));
  try {
    const allowed = path.join(root, "selected"); mkdirSync(allowed);
    const denied = path.join(root, "sibling"); mkdirSync(denied);
    const secret = path.join(denied, "signing.json"); writeFileSync(secret, "boundary-probe");
    writeFileSync(path.join(allowed, "public.json"), "permitted");
    symlinkSync(secret, path.join(allowed, "escape"));
    const manifest = { readonly: [...executableInputs(process.execPath), allowed], cwd: allowed, environment: { DISCOVER_INPUT: secret } };
    const run = script => isolated({ ...manifest, command: [process.execPath, "--input-type=module", "-e", script] });
    assert.equal(run('import fs from "node:fs"; process.stdout.write(fs.readFileSync("public.json","utf8"));'), "permitted");
    for (const [probe, expression] of [
      ["absolute", JSON.stringify(secret)],
      ["relative", '"../sibling/signing.json"'],
      ["environment discovery", "process.env.DISCOVER_INPUT"],
      ["symlink escape", '"escape"'],
      ["generator input", JSON.stringify(secret)],
      ["sibling corpus", '"../sibling/signing.json"'],
      ["indirect signing", 'new URL("../sibling/signing.json", `file://${process.cwd()}/`)'],
    ]) {
      assert.equal(run(`import fs from "node:fs";try{fs.readFileSync(${expression});process.exit(9)}catch(e){if(!["ENOENT","EACCES"].includes(e.code))throw e;process.stdout.write("denied")}`), "denied", probe);
    }
    assert.equal(run('process.stdout.write(String(process.env.HOME===undefined && process.env.PATH===undefined))'), "true");
    assert.throws(() => run('import fs from "node:fs";fs.writeFileSync("public.json","changed")'), /qualification command failed/u);
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test("generation output directories cannot overwrite overlaid authored inputs", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "reference-generation-overlay-"));
  try {
    const input = path.join(root, "immutable-release.json");
    writeFileSync(input, "retained");
    const manifest = {readonly: executableInputs(process.execPath), writable: [root], readonly_overlays: [input], cwd: root};
    const run = code => isolated({...manifest, command: [process.execPath, "--input-type=module", "-e", code]});
    assert.equal(run('import fs from "node:fs";fs.writeFileSync("output.json","generated");process.stdout.write(fs.readFileSync("immutable-release.json","utf8"));'), "retained");
    for (const operation of ['fs.writeFileSync("immutable-release.json","changed")', 'fs.unlinkSync("immutable-release.json")', 'fs.renameSync("output.json","immutable-release.json")']) {
      assert.throws(() => run(`import fs from "node:fs";${operation}`), /qualification command failed/u);
    }
  } finally { rmSync(root, {recursive: true, force: true}); }
});


test("generation acceptance rejects linked, missing, special and undeclared outputs before comparison", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "reference-generation-outputs-"));
  const source = path.join(root, "source"), baseline = path.join(root, "baseline");
  const inventory = {inputs: ["generated/authored.json"], outputs: ["generated/nested/result.ts"], output_directories: ["generated"]};
  try {
    for (const dir of [source, baseline]) {
      mkdirSync(path.join(dir, "generated/nested"), {recursive: true});
      writeFileSync(path.join(dir, "generated/authored.json"), "authored");
      writeFileSync(path.join(dir, "generated/nested/result.ts"), "generated");
    }
    linkSync(path.join(baseline, inventory.outputs[0]), path.join(root, "cached-reference"));
    const validate = () => validateGenerationOutputs(source, baseline, inventory);
    assert.deepEqual(Object.keys(validate()), inventory.outputs);
    const result = path.join(source, inventory.outputs[0]);
    for (const target of [path.join(baseline, inventory.outputs[0]), "missing.ts"]) {
      rmSync(result);
      // The host answer is absent from the generator's filesystem, but it can
      // still emit a dangling absolute link that resolves after sandbox exit.
      isolated({readonly: executableInputs(process.execPath), writable: [source], cwd: source, command: [process.execPath, "--input-type=module", "-e", `import fs from "node:fs";fs.symlinkSync(${JSON.stringify(target)}, "generated/nested/result.ts");try{fs.readFileSync("generated/nested/result.ts");process.exit(9)}catch(error){if(error.code!=="ENOENT")throw error}`]});
      assert.throws(validate, /linked or nonregular/u, "byte-identical and dangling leaf links cannot qualify");
      rmSync(result); writeFileSync(result, "generated");
    }
    rmSync(result); linkSync(path.join(baseline, inventory.outputs[0]), result);
    assert.throws(validate, /linked or nonregular/u);
    rmSync(result); execFileSync("mkfifo", [result]);
    assert.throws(validate, /linked or nonregular/u, "special files cannot block or supply output bytes");
    rmSync(result);
    assert.throws(validate, /ENOENT/u, "every declared output is required");
    writeFileSync(result, "generated");
    writeFileSync(path.join(source, "generated/extra.ts"), "undeclared");
    assert.throws(validate, /undeclared generation output/u);
    rmSync(path.join(source, "generated/extra.ts"));
    mkdirSync(path.join(source, "generated/extra"));
    assert.throws(validate, /undeclared generation directory/u);
    rmSync(path.join(source, "generated/extra"), {recursive: true});
    rmSync(path.join(source, "generated/nested"), {recursive: true});
    symlinkSync(path.join(baseline, "generated/nested"), path.join(source, "generated/nested"));
    assert.throws(validate, /undeclared generation output|linked or nonregular/u, "linked parent cannot expose matching host answers");
    rmSync(path.join(source, "generated/nested"));
    mkdirSync(path.join(source, "generated/nested")); writeFileSync(result, "changed");
    assert.throws(validate, /content mismatch/u);
    writeFileSync(result, "generated");
    writeFileSync(path.join(source, "generated/authored.json"), "changed");
    assert.throws(validate, /content mismatch/u, "authored inputs sharing output directories are preserved");
    assert.throws(() => validateGenerationOutputs(source, baseline, {...inventory, outputs: ["../escape"]}), /invalid generation path/u);
  } finally { rmSync(root, {recursive: true, force: true}); }
});
