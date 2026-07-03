#!/usr/bin/env node

import { runGoTargetCLI } from "../tools/harness/backend/go-target-runner.mjs";

const status = await runGoTargetCLI(process.argv.slice(2));
process.exit(status);
