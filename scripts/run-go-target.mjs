#!/usr/bin/env node

import { runGoTargetCLI } from "./lib/go-target-runner.mjs";

const status = await runGoTargetCLI(process.argv.slice(2));
process.exit(status);
