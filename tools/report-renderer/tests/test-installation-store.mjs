#!/usr/bin/env node

// TH-HARNESS-AC-174: a recorded renderer installation is reused only while
// its manifest, its root and the host libraries it copied are unchanged.
import assert from "node:assert/strict";
import { chmod, mkdir, mkdtemp, readFile, rename, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import { fileDigest, inventory, publishReceipt, recordKey, reusableInstallation, sha } from "../installation-store.mjs";

const key = "a".repeat(64);

async function store(t) {
  const scratch = await mkdtemp(path.join(os.tmpdir(), "renderer-store-"));
  t.after(() => rm(scratch, { recursive: true, force: true }));
  const output = path.join(scratch, "build");
  const libraries = path.join(scratch, "host-lib");
  await mkdir(libraries, { recursive: true });
  await writeFile(path.join(libraries, "libfixture.so.1"), "host library v1");
  const staged = path.join(scratch, "root");
  await mkdir(path.join(staged, "bin"), { recursive: true });
  await mkdir(path.join(staged, "lib/x86_64-linux-gnu"), { recursive: true });
  await writeFile(path.join(staged, "bin/tool"), "tool");
  await writeFile(path.join(staged, "lib/x86_64-linux-gnu/libfixture.so.1"), "host library v1");
  const files = await inventory(staged);
  const bytes = JSON.stringify({ schema_id: "cartulary.reporting_runtime_installation.v1", files });
  const identity = sha(bytes);
  await mkdir(output, { recursive: true });
  await writeFile(path.join(output, `${identity}.json`), bytes);
  await rename(staged, path.join(output, `${identity}.root`));
  const work = path.join(scratch, "work");
  await mkdir(work);
  await recordKey(output, key, identity, work);
  const options = { libraryDirectories: [libraries], loader: path.join(libraries, "absent-loader") };
  return { scratch, output, identity, libraries, work, options };
}

test("an intact recorded installation is reused and receives the receipt", async (t) => {
  const { output, identity, work, options } = await store(t);
  assert.equal(await reusableInstallation(output, key, options), identity);
  await publishReceipt(output, identity, work);
  const receipt = JSON.parse(await readFile(path.join(output, "installation.json"), "utf8"));
  assert.deepEqual(receipt, { schema_id: "cartulary.reporting_runtime_receipt.v1", manifest_sha256: identity, root: `${identity}.root`, manifest: `${identity}.json` });
  assert.equal(await reusableInstallation(output, "b".repeat(64), options), null, "an unrecorded key is never reused");
});

test("a changed manifest, root or copied host library is not reused", async (t) => {
  await t.test("manifest", async (t) => {
    const { output, identity, options } = await store(t);
    await writeFile(path.join(output, `${identity}.json`), JSON.stringify({ files: [] }));
    assert.equal(await reusableInstallation(output, key, options), null);
  });
  await t.test("root content", async (t) => {
    const { output, identity, options } = await store(t);
    const file = path.join(output, `${identity}.root`, "bin/tool");
    await chmod(file, 0o644);
    await writeFile(file, "tampered");
    assert.equal(await reusableInstallation(output, key, options), null);
  });
  await t.test("root membership", async (t) => {
    const { output, identity, options } = await store(t);
    await writeFile(path.join(output, `${identity}.root`, "bin/extra"), "extra");
    assert.equal(await reusableInstallation(output, key, options), null);
  });
  await t.test("host library", async (t) => {
    const { output, libraries, options } = await store(t);
    await writeFile(path.join(libraries, "libfixture.so.1"), "host library v2");
    assert.equal(await reusableInstallation(output, key, options), null);
    await rm(path.join(libraries, "libfixture.so.1"));
    assert.equal(await reusableInstallation(output, key, options), null, "a vanished host library is not reusable");
  });
  await t.test("key record", async (t) => {
    const { output, options } = await store(t);
    await writeFile(path.join(output, "keys", `${key}.json`), JSON.stringify({ input_key: key, manifest_sha256: "../escape" }));
    assert.equal(await reusableInstallation(output, key, options), null);
  });
});

test("the inventory binds every file by digest in byte order", async (t) => {
  const { output, identity } = await store(t);
  const files = await inventory(path.join(output, `${identity}.root`));
  assert.deepEqual(files.map((file) => file.path), ["bin/tool", "lib/x86_64-linux-gnu/libfixture.so.1"]);
  assert.equal(files[0].sha256, await fileDigest(path.join(output, `${identity}.root`, "bin/tool")));
});
