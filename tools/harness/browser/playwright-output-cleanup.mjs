import { readdirSync } from "node:fs";
import path from "node:path";
import { removePrivateTree } from "../runtime/secure-local-files.mjs";

// Pinned producer scratch, never a report attachment. Only call after the owned
// runner group has drained; completed per-test attachments remain untouched.
export function removePlaywrightWorkingTraces(outputRoot) {
  let entries;
  try { entries = readdirSync(outputRoot); }
  catch (error) { if (error.code === "ENOENT") return; throw error; }
  for (const entry of entries) {
    if (/^\.playwright-artifacts-\d+$/u.test(entry)) removePrivateTree(path.join(outputRoot, entry));
  }
}
