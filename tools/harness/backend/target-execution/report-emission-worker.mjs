import { parentPort, workerData } from "node:worker_threads";

import { handleGoStep } from "../../output/test-output/runners/go.mjs";

function errorRecord(error) {
  return {
    message: error instanceof Error ? error.message : String(error),
    stack: error instanceof Error ? error.stack : undefined,
    exitCode: Number.isInteger(error?.publicExitCode)
      ? error.publicExitCode
      : Number.isInteger(error?.exitCode) ? error.exitCode : undefined,
  };
}

function window(startedMs) {
  const endedMs = Date.now();
  return {
    startTime: new Date(startedMs).toISOString(),
    endTime: new Date(endedMs).toISOString(),
    durationMs: Math.max(0, endedMs - startedMs),
  };
}

const results = [];
for (const entry of workerData.entries) {
  const startedMs = Date.now();
  let status = 0;
  let error = null;
  try {
    for (const emission of entry.request.emissions) {
      Object.assign(process.env, emission.env);
      const result = handleGoStep({ catalogAware: emission.catalogAware });
      if (result !== 0) status = result;
    }
  } catch (caught) {
    error = errorRecord(caught);
  }
  results.push({
    index: entry.index,
    status: error ? null : status,
    error,
    window: window(startedMs),
  });
}
parentPort.postMessage({ results });
