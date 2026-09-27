import { fork } from "node:child_process";
import { fileURLToPath } from "node:url";
import { ReviewFailure, limits, failureMappings } from "./contract.mjs";

export function runImageWork(job, signal) {
  signal.throwIfAborted();
  return new Promise((resolve, reject) => {
    const worker = fork(fileURLToPath(import.meta.url), ["--worker"], { execArgv: [], serialization: "advanced", stdio: ["ignore", "ignore", "ignore", "ipc"], env: { PATH: "/usr/bin:/bin", LANG: "C.UTF-8", TZ: "UTC" } });
    let response, failure, settled = false;
    const abort = () => { failure ??= signal.reason; worker.kill("SIGKILL"); };
    const timer = setTimeout(() => { failure ??= new ReviewFailure("operation_expired"); worker.kill("SIGKILL"); }, limits.operation);
    signal.addEventListener("abort", abort, { once: true });
    const finish = () => {
      if (settled) return; settled = true; clearTimeout(timer); signal.removeEventListener("abort", abort);
      if (failure) reject(failure); else if (response) resolve(response); else reject(new ReviewFailure("analysis_failed"));
    };
    worker.once("error", (cause) => { failure = new ReviewFailure("analysis_failed", { cause }); finish(); });
    worker.once("message", (message) => {
      if (message.error) failure = new ReviewFailure(Object.hasOwn(failureMappings, message.error) ? message.error : "analysis_failed");
      else response = message.value;
    });
    worker.once("exit", finish);
    worker.send(job, (error) => { if (error) { failure ??= new ReviewFailure("analysis_failed"); worker.kill("SIGKILL"); } });
  });
}
if (process.argv[2] === "--worker") {
  let completed = false;
  process.once("disconnect", () => process.exit(completed ? 0 : 15));
  const reply = (message) => process.send(message, () => { completed = true; process.disconnect(); });
  process.once("message", async (job) => {
    try {
      const { computeImages } = await import("./image-algorithms.mjs");
      const value = await computeImages(job);
      reply({ value });
    } catch (error) { reply({ error: error instanceof ReviewFailure ? error.diagnostic : "analysis_failed" }); }
  });
}
