import { parentPort, workerData } from "node:worker_threads";

const { handleGoStep } = await import("../../output/test-output/runners/go.mjs");

const results = [];
for (const task of workerData.tasks) {
  const previous = new Map();
  for (const [name, value] of Object.entries(task.env)) {
    previous.set(name, Object.hasOwn(process.env, name) ? process.env[name] : undefined);
    process.env[name] = value;
  }
  let status = 1;
  try {
    status = handleGoStep({ catalogAware: task.catalogAware });
  } finally {
    for (const [name, value] of previous) {
      if (value === undefined) delete process.env[name];
      else process.env[name] = value;
    }
  }
  results.push({ index: task.index, familyIndex: task.familyIndex, status });
}
parentPort.postMessage(results);
