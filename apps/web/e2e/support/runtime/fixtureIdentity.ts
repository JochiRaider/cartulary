export type FixtureIdentitySource = {
  now?: () => number;
  random?: () => number;
};

export function uniqueTxn(prefix: string, source: FixtureIdentitySource = {}) {
  const now = source.now ?? Date.now;
  const random = source.random ?? Math.random;
  return `${prefix}-${now()}-${random().toString(16).slice(2, 8)}`;
}

export function uniqueEmail(
  prefix: string,
  source: Pick<FixtureIdentitySource, "now"> = {},
) {
  const now = source.now ?? Date.now;
  return `${prefix}-${now().toString(36)}@example.test`;
}

export function uniqueIncidentKey(
  prefix: string,
  source: Pick<FixtureIdentitySource, "now"> = {},
) {
  const now = source.now ?? Date.now;
  return `IR-${prefix}-${now().toString(36).toUpperCase()}`;
}
