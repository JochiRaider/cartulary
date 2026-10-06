// Controlled resource-owner port. Production provider acquisition and stop
// still run through their real subprocess boundary and exact private lease.
import { spawn } from "node:child_process";
import { createServer } from "node:net";
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { ownedProcess } from "../runtime/owned-process.mjs";

import { createAcquisitionLaunch, recordAcquisitionProcess } from "../browser/browser-acquisition.mjs";

if (process.env.CARTULARY_FIXTURE_PARENT_ONLY !== undefined) {
  throw new Error("parent-only binding leaked into the owned fixture");
}

const args = process.argv.slice(2);
if (args[0] === "serve") {
  recordAcquisitionProcess(process.env.CARTULARY_BROWSER_ACQUISITION_FILE, args[2]);
  const server = createServer();
  server.listen(0, "127.0.0.1", () => writeFileSync(args[1], String(server.address().port), { mode: 0o600 }));
  process.on("SIGTERM", () => server.close(() => process.exit(0)));
} else {
  const value = (key) => args[args.indexOf(key) + 1];
  const leaseFile = value("--lease-file");
  const directory = path.dirname(leaseFile);
  if (args[0] === "--session-start") {
    const portFile = path.join(directory, "port");
    const launchID = createAcquisitionLaunch(process.env.CARTULARY_BROWSER_ACQUISITION_FILE, "backend");
    const child = spawn(process.execPath, [import.meta.filename, "serve", portFile, launchID], { detached: true, stdio: "ignore" });
    recordAcquisitionProcess(process.env.CARTULARY_BROWSER_ACQUISITION_FILE, launchID, child.pid);
    child.unref();
    for (let i = 0; i < 200 && !existsSync(portFile); i += 1) await new Promise((resolve) => setTimeout(resolve, 10));
    const lease = { proof: ownedProcess(child.pid, { group: true }), port: Number(readFileSync(portFile, "utf8")) };
    if (process.env.CARTULARY_FIXTURE_TEST_ACQUISITION_FAILURE === "before-ready") throw new Error("controlled failure before ready publication");
    if (process.env.CARTULARY_FIXTURE_TEST_ACQUISITION_FAILURE === "late-ready") await new Promise((resolve) => {
      const active = setInterval(() => {}, 1000);
      process.once("SIGTERM", () => { clearInterval(active); resolve(); });
      writeFileSync(path.join(directory, "cancellable"), "ready", { mode: 0o600 });
    });
    const stack = path.join(directory, "stack-v8.json");
    writeFileSync(stack, "{}", { mode: 0o600 });
    writeFileSync(path.join(directory, "service-admission.json"), "{}", { mode: 0o600 });
    writeFileSync(leaseFile, JSON.stringify(lease), { mode: 0o600 });
    writeFileSync(value("--env-file"), JSON.stringify({ CARTULARY_WEB_E2E_STACK_JSON_FILE: stack }), { mode: 0o600 });
  } else throw new Error("unknown controlled stack operation");
}
