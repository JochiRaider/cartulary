// Controlled resource-owner port. Production provider acquisition and stop
// still run through their real subprocess boundary and exact private lease.
import { spawn } from "node:child_process";
import { createServer } from "node:net";
import { existsSync, readFileSync, writeFileSync, rmSync, appendFileSync } from "node:fs";
import path from "node:path";
import { ownedProcess, stopOwnedProcess } from "../runtime/owned-process.mjs";

const args = process.argv.slice(2);
if (args[0] === "serve") {
  const server = createServer();
  server.listen(0, "127.0.0.1", () => writeFileSync(args[1], String(server.address().port), { mode: 0o600 }));
  process.on("SIGTERM", () => server.close(() => process.exit(0)));
} else {
  const value = (key) => args[args.indexOf(key) + 1];
  const leaseFile = value("--lease-file");
  const directory = path.dirname(leaseFile);
  if (args[0] === "--session-start") {
    const portFile = path.join(directory, "port");
    const child = spawn(process.execPath, [import.meta.filename, "serve", portFile], { detached: true, stdio: "ignore" });
    child.unref();
    for (let i = 0; i < 200 && !existsSync(portFile); i += 1) await new Promise((resolve) => setTimeout(resolve, 10));
    const lease = { proof: ownedProcess(child.pid, { group: true }), port: Number(readFileSync(portFile, "utf8")) };
    const stack = path.join(directory, "stack-v7.json");
    writeFileSync(stack, "{}", { mode: 0o600 });
    writeFileSync(path.join(directory, "service-admission.json"), "{}", { mode: 0o600 });
    writeFileSync(leaseFile, JSON.stringify(lease), { mode: 0o600 });
    writeFileSync(value("--env-file"), JSON.stringify({ CARTULARY_WEB_E2E_STACK_JSON_FILE: stack }), { mode: 0o600 });
  } else if (args[0] === "--session-stop") {
    appendFileSync(path.join(directory, "stop-count"), "stop\n", { mode: 0o600 });
    const lease = JSON.parse(readFileSync(leaseFile, "utf8"));
    if (process.env.CARTULARY_FIXTURE_TEST_STOP_FAILURE === "1") throw new Error("deliberate owner stop failure");
    await stopOwnedProcess(lease.proof);
    if (process.env.CARTULARY_WEB_E2E_RETAIN_SESSION_LEASE !== "1") rmSync(leaseFile);
  } else throw new Error("unknown controlled stack operation");
}
