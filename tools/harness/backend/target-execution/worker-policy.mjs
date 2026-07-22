export function resolveBackendWorkerPool(availableParallelism, groupCount) {
  if (!Number.isInteger(availableParallelism) || availableParallelism < 1) {
    throw new Error(`invalid available parallelism ${availableParallelism}`);
  }
  if (!Number.isInteger(groupCount) || groupCount < 0) {
    throw new Error(`invalid backend group count ${groupCount}`);
  }
  if (groupCount === 0) return { workers: 0, goMaxProcs: 1 };
  const workers = Math.min(
    groupCount,
    Math.max(1, Math.min(8, Math.floor(availableParallelism / 4))),
  );
  return {
    workers,
    goMaxProcs: Math.max(1, Math.floor(availableParallelism / workers)),
  };
}

export function resolveBackendCapturePool(
  target,
  availableParallelism,
  groupCount,
) {
  const pool = resolveBackendWorkerPool(availableParallelism, groupCount);
  if (target === "backend-unit") return pool;
  return {
    workers: pool.workers,
    goMaxProcs: availableParallelism,
  };
}
