import { writeFileSync } from "node:fs";
import path from "node:path";

const [mode, readyFile] = process.argv.slice(2);
const phase = (name) => process.stdout.write(`${JSON.stringify({ phase: name, at: performance.timeOrigin + performance.now() })}\n`);
phase("started");
if (mode === "diagnostic" || mode === "contradictory" || mode === "conflict") {
  const { publishCommandFailure } = await import("../runtime/command-failure.mjs");
  publishCommandFailure(path.resolve(import.meta.dirname, "../../.."), { failure_class: "artifact", failure_reason: "artifact_error" });
  if (mode === "conflict") {
    try { publishCommandFailure(path.resolve(import.meta.dirname, "../../.."), { failure_class: "security", failure_reason: "security_finding" }); }
    catch { process.exitCode = 11; }
  }
  phase("published");
  process.exitCode ??= mode === "diagnostic" ? 2 : 0;
  phase("completed");
} else if (mode === "assertion") {
  process.exitCode = 10;
} else if (mode === "wait" || mode === "cooperative") {
  process.on("SIGTERM", () => { if (mode === "cooperative") process.exit(0); });
  setInterval(() => {}, 1000);
  writeFileSync(readyFile, String(process.pid), { mode: 0o600, flag: "wx" });
} else throw new Error("unknown diagnostic fixture mode");
