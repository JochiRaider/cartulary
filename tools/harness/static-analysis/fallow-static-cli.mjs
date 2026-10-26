#!/usr/bin/env node

import { repoRoot } from "../contract/index.mjs";
import { runFallowStatic } from "./fallow-static.mjs";

process.umask(0o077);
process.exit(await runFallowStatic({ root: repoRoot }));
