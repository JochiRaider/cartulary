#!/usr/bin/env node

import { runHarnessContractCLI } from "../tools/harness/core/harness-contract-cli.mjs";

await runHarnessContractCLI(process.argv.slice(2), {
  programName: "harness-contract.mjs",
});
