#!/usr/bin/env node

import {
  readBoundedRegularFile,
  RetirementAttestationError,
  validateRetirementAttestationBytes,
  writeRetirementAttestationResult,
} from "./incident-bundle-v1-retirement-attestation.mjs";

async function main() {
  const args = process.argv.slice(2);
  const attestationPath = args[0] === "--attestation" ? args[1] : "";
  if (!attestationPath || args.length !== 2) {
    process.stderr.write(
      "incident bundle v1 retirement attestation failed failure_class=config failure_reason=usage_error diagnostic=missing-attestation\n",
    );
    process.exitCode = 2;
    return;
  }
  try {
    const bytes = readBoundedRegularFile(attestationPath);
    const result = await validateRetirementAttestationBytes(bytes);
    const retained = writeRetirementAttestationResult(
      result,
      process.env.CARTULARY_STEP_ARTIFACT_DIR,
    );
    process.stdout.write(
      `incident bundle v1 retirement attestation passed gates=5 input_digest=${result.input_digest} result_digest=${retained.digest}\n`,
    );
  } catch (error) {
    const diagnostic =
      error instanceof RetirementAttestationError ? error.code : "validator_internal_error";
    process.stderr.write(
      `incident bundle v1 retirement attestation failed failure_class=artifact failure_reason=artifact_error diagnostic=${diagnostic}\n`,
    );
    process.exitCode = 11;
  }
}

await main();
