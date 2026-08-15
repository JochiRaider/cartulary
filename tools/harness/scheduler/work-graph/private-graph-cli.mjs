#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import { canonicalJSONString } from "../../contract/index.mjs";
import { WorkGraphCompiler } from "./compiler.mjs";

function usage() {
  return "usage: private-graph-cli.mjs target|aggregate|owner|rows <value>";
}

function main() {
  const [kind, value, ...extra] = process.argv.slice(2);
  if (!kind || value === undefined || extra.length > 0) throw new Error(usage());
  const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../../..");
  const compiler = new WorkGraphCompiler(root);
  let selection;
  if (kind === "target" || kind === "aggregate") selection = { kind, target: value };
  else if (kind === "owner") selection = { kind, owner_id: value };
  else if (kind === "rows") selection = { kind, row_ids: value.split(",") };
  else throw new Error(usage());
  process.stdout.write(`${canonicalJSONString(compiler.compile(selection))}\n`);
}

try {
  main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 2;
}
