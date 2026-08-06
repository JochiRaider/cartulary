#!/usr/bin/env node
import assert from "node:assert/strict";
import {
  chmodSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";

import { runCleanup } from "../harness-contract.mjs";

const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-cleanup-contract-"));
const externalScratch = mkdtempSync(
  path.join(os.tmpdir(), "cartulary-cleanup-external-"),
);
try {
  const owned = path.join(scratch, ".cache", "cartulary");
  const readOnly = path.join(owned, "download", "module");
  const unrelated = path.join(scratch, ".cache", "dx-remediation", "keep.txt");
  const external = path.join(externalScratch, "external-target");
  const linked = path.join(owned, "external-link");
  mkdirSync(readOnly, { recursive: true });
  mkdirSync(path.dirname(unrelated), { recursive: true });
  mkdirSync(external);
  writeFileSync(path.join(readOnly, "artifact"), "read-only\n");
  writeFileSync(unrelated, "unrelated\n");
  writeFileSync(path.join(external, "marker"), "external\n");
  symlinkSync(external, linked, "dir");
  chmodSync(path.join(readOnly, "artifact"), 0o400);
  chmodSync(readOnly, 0o500);
  chmodSync(path.dirname(readOnly), 0o500);

  runCleanup({
    scope: "distclean",
    candidates: [owned, path.join(scratch, "absent")],
    includeTmp: false,
    root: scratch,
  });
  assert.equal(existsSync(owned), false, "owned read-only cache must be removed");
  assert.equal(readFileSync(unrelated, "utf8"), "unrelated\n", "unrelated cache must survive");
  assert.equal(
    readFileSync(path.join(external, "marker"), "utf8"),
    "external\n",
    "external symlink target must survive",
  );

  runCleanup({
    scope: "distclean",
    candidates: [owned, path.join(scratch, "absent")],
    includeTmp: false,
    root: scratch,
  });

  const embedded = path.join(
    scratch,
    "internal",
    "platform",
    "httpapi",
    "webassets",
    "dist",
  );
  mkdirSync(path.join(embedded, "readonly"), { recursive: true });
  writeFileSync(path.join(embedded, ".keep"), "keep\n");
  writeFileSync(path.join(embedded, "readonly", "asset"), "asset\n");
  chmodSync(path.join(embedded, "readonly"), 0o500);
  runCleanup({
    scope: "distclean",
    candidates: [],
    includeTmp: false,
    root: scratch,
    embeddedWebAssetsDir: embedded,
  });
  assert.equal(readFileSync(path.join(embedded, ".keep"), "utf8"), "keep\n");
  assert.equal(existsSync(path.join(embedded, "readonly")), false);

  const directLink = path.join(scratch, "owned-link");
  symlinkSync(external, directLink, "dir");
  runCleanup({
    scope: "distclean",
    candidates: [directLink],
    includeTmp: false,
    root: scratch,
  });
  assert.equal(existsSync(directLink), false, "cleanup must unlink a candidate symlink");
  assert.equal(readFileSync(path.join(external, "marker"), "utf8"), "external\n");
} finally {
  chmodSync(scratch, 0o700);
  rmSync(scratch, { recursive: true, force: true });
  rmSync(externalScratch, { recursive: true, force: true });
}

process.stdout.write("cleanup contract checks passed\n");
