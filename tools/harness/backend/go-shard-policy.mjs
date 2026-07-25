const defaultServiceExactShardProfile = Object.freeze({
  max_symbols: 8,
  max_estimated_test_work_ms: 12_000,
});

const targetServiceExactShardProfiles = Object.freeze({
  "backend-integration": Object.freeze({
    max_symbols: 8,
    max_estimated_test_work_ms: 6_000,
  }),
  "backend-process": Object.freeze({
    max_symbols: 16,
    max_estimated_test_work_ms: 24_000,
  }),
});

export function serviceExactShardProfile(target) {
  return targetServiceExactShardProfiles[target] ?? defaultServiceExactShardProfile;
}
