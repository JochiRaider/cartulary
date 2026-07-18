#!/usr/bin/env node

import { writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";

const argv = process.argv.slice(2);
if (argv.length !== 2 || argv[0] !== "--output" || !argv[1]) {
  process.stderr.write("usage: owner-slice-finalizer-cli.mjs --output <path>\n");
  process.exitCode = 2;
} else if (process.env.CARTULARY_TEST_SLICE_FINALIZER_FAULT === "1") {
  process.stderr.write("injected owner-slice finalizer failure\n");
  process.exitCode = 12;
} else {
  writeFileSync(path.resolve(argv[1]), `${JSON.stringify({
    cleanup_state: "complete",
    completed_at: new Date().toISOString(),
  }, null, 2)}\n`, "utf8");
}
