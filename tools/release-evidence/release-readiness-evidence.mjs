#!/usr/bin/env node

import path from "node:path";
import { packageRunContext, readPackageReceipt, verifyPackageIdentity } from "./package-identity.mjs";
import { fileURLToPath } from "node:url";

export function validateLiveReleaseReadiness(environment = process.env) {
  if (environment.CARTULARY_HARNESS_GRAPH_CHILD !== "1") {
    throw new Error("release readiness is a private graph operation; inspect retained runs with harness-observability-check CHECK=all");
  }
  const { runRoot } = packageRunContext(environment);
  // Dependencies are owned by the work graph. The marker checks one package
  // identity after its consumers finish; canonical summaries publish afterward.
  const producer = readPackageReceipt(runRoot);
  const consumers = [];
  for (const consumer of ["standup-package-smoke", "standup-reference-pack-smoke", "standup-operational-recovery-smoke", "credential-capacity-assessment"]) {
    const receipt = readPackageReceipt(runRoot, `${consumer}/artifacts/package.json`);
    consumers.push([consumer, receipt]);
  }
  verifyPackageIdentity(producer, consumers);
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    validateLiveReleaseReadiness();
    process.stdout.write("release readiness closure reached; canonical projection pending\n");
  } catch {
    process.stderr.write("release readiness evidence admission failed\n");
    process.exitCode = 1;
  }
}
