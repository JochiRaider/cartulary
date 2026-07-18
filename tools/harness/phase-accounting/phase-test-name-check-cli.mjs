#!/usr/bin/env node

// Temporary pre-cutover entry point. WS-09 replaces the public target and WS-10
// deletes this compatibility-shaped path with the rest of phase accounting.
import { main } from "../test-catalog/semantic-identity-check-cli.mjs";

try {
  main();
} catch (error) {
  process.stderr.write(`semantic identity check failed: ${error.message}\n`);
  process.exitCode = 1;
}
