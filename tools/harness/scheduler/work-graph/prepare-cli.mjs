#!/usr/bin/env node
import { lifecyclePolicy } from "../../contract/generated/lifecycle-policy.mjs";
import { closeSync, fstatSync, writeSync } from "node:fs";
import { prepareRuntimePlan } from "./prepare-plan.mjs";

// Descriptor 3 is an inherited pipe owned by the Go supervisor. There is no
// pathname or arbitrary command input and no retained preparation artifact.
try {
  if (!fstatSync(3).isFIFO()) throw Error("private compiler descriptor required");
  const bytes = Buffer.from(JSON.stringify(prepareRuntimePlan(process.argv.slice(2))));
  if (bytes.length > lifecyclePolicy.limits.plan) throw Error("private compiler plan exceeds bound");
  let offset = 0;
  while (offset < bytes.length) offset += writeSync(3, bytes, offset, bytes.length - offset);
  closeSync(3);
} catch {
  process.stderr.write("graph preparation failed\n");
  process.exitCode = 2;
}
