import assert from "node:assert/strict";
import { mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import test from "node:test";

const command = path.resolve("tools/reference-packs/explicit-input.mjs");
test("explicit producer inputs survive empty working directories without widening admission", () => {
  const root = mkdtempSync(path.join(tmpdir(), "cartulary-producer-input-"));
  try {
    let sequence = 0;
    const invoke = raw => {
      const input = path.join(root, `input-${sequence}`), output = path.join(root, `output-${sequence++}`);
      writeFileSync(input, raw);
      return { output, result: spawnSync(process.execPath, [command, input, output], { cwd: root, env: {}, encoding: "utf8" }) };
    };
    const input = { source_root: "corpus", output_root: "output", ledger_root: "ledger", untouched: "value", envelope: { logical_root: "logical", signing_context: "keys/key.json", previous_envelope: "", metadata_version: 1 } };
    const { output, result } = invoke(JSON.stringify(input));
    assert.equal(result.status, 0, result.stderr);
    const resolved = JSON.parse(readFileSync(output));
    assert.deepEqual(resolved, { ...input, source_root: path.join(root, "corpus"), output_root: path.join(root, "output"), ledger_root: path.join(root, "ledger"), envelope: { ...input.envelope, logical_root: path.join(root, "logical"), signing_context: path.join(root, "keys/key.json") } });
    for (const raw of ['{"version":1,"version":2}', '{"version":1e0}', '{"version":1.0}', '{"version":-0}', '{"version":9007199254740992}', '{"path":"\\ud800"}', Buffer.from([0xff]), Buffer.alloc(4194305, 32)]) assert.notEqual(invoke(raw).result.status, 0);
    const link = path.join(root, "link");
    symlinkSync(output, link);
    assert.notEqual(spawnSync(process.execPath, [command, link, path.join(root, "linked-output")], { cwd: root, env: {} }).status, 0);
  } finally { rmSync(root, { recursive: true, force: true }); }
});
