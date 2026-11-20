#!/usr/bin/env node
import { lifecyclePolicy } from "../../contract/generated/lifecycle-policy.mjs";
import { preparationDiagnostic } from "../../contract/index.mjs";
import { closeSync, fstatSync, writeSync } from "node:fs";
import { prepareRuntimePlan } from "./prepare-plan.mjs";

// Descriptor 3 is an inherited pipe owned by the Go supervisor. There is no
// pathname or arbitrary command input and no retained preparation artifact.
// It carries the plan, or one closed preparation diagnostic when preparation
// is rejected (TH-HARNESS-REQ-854). Exception text never crosses it.
function publish(bytes) {
  let offset = 0;
  while (offset < bytes.length) offset += writeSync(3, bytes, offset, bytes.length - offset);
  closeSync(3);
}

let descriptor = false;
try {
  if (!fstatSync(3).isFIFO()) throw Error("private compiler descriptor required");
  descriptor = true;
  const bytes = Buffer.from(JSON.stringify(prepareRuntimePlan(process.argv.slice(2))));
  if (bytes.length > lifecyclePolicy.limits.plan) throw Error("private compiler plan exceeds bound");
  publish(bytes);
} catch (error) {
  process.exitCode = 2;
  try {
    if (!descriptor) throw error;
    publish(Buffer.from(JSON.stringify(preparationDiagnostic(error))));
  } catch {
    process.stderr.write("graph preparation failed\n");
  }
}
