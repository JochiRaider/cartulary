import fs, { mkdirSync, writeFileSync, writeSync } from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import path from "node:path";

const root = path.resolve(import.meta.dirname, "../../..");
const mode = process.argv[2];
if (mode === "compiler-failure") {
  process.exitCode = 7;
} else if (mode === "producer" || mode === "producer-race") {
  const { claimFrontendProducer, publishFrontendOutput, sealFrontendArtifact } = await import("../readiness/frontend-artifact.mjs");
  const { borrowSuiteRuntime } = await import("../runtime/suite-runtime.mjs");
  const runRoot = path.resolve(process.env.CARTULARY_TEST_RESULTS_DIR, process.env.CARTULARY_TEST_RUN_ID);
  const runtime = borrowSuiteRuntime({ repoRoot: root, runRoot });
  const id = process.argv[3] || "production";
  const profile = { id, producer_target: id === "production" ? "build-web" : "build-web-measurement", entries: id === "production" ? ["index.html"] : ["index.html", "measurement.html", "presence.html"] };
  const admit = () => { try {
    const claim = claimFrontendProducer(runtime, profile);
    process.send({ state: "admitted" });
    process.once("message", () => {
      const boundary = process.argv[4];
      const publication = process.argv[5];
      if (boundary) {
        const rename = fs.renameSync;
        fs.renameSync = (source, destination) => {
          if ((boundary === "receipt" && path.basename(destination) === "frontend-artifact.json") ||
            (boundary === "conventional" && destination === publication)) {
            // Synchronous witness before stopping at the real rename boundary.
            // Only the parent SIGKILL releases this fixture; no timing sleeps.
            writeSync(4, boundary);
            Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0);
          }
          return rename(source, destination);
        };
        syncBuiltinESMExports();
      }
      writeFileSync(runtime.privatePath(`build-start-${id}-${process.pid}`), "started", { mode: 0o600, flag: "wx" });
      const staging = runtime.privatePath(`worker-${process.pid}`);
      mkdirSync(staging, { mode: 0o700 });
      for (const entry of profile.entries) writeFileSync(path.join(staging, entry), "sealed");
      const directory = sealFrontendArtifact({ repoRoot: root, runRoot, runtime, profile, staging, claim });
      if (publication) publishFrontendOutput(directory, publication);
      process.send({ state: "sealed" });
      process.disconnect();
    });
  } catch (error) {
    process.send({ state: "rejected", failure_class: error.failure_class, failure_reason: error.failure_reason });
    process.exitCode = 11;
    process.disconnect();
  } };
  if (mode === "producer-race") {
    process.send({ state: "ready" });
    process.once("message", admit);
  } else admit();
}
