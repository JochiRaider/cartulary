export function parseFixtureRuntimeBundle(
  contents: string,
  expected: {
    fixtureProfileId: string;
    snapshotKey: string;
    credentialSetId: string;
    accountCount: number;
  },
) {
  let decoded: unknown;
  try {
    decoded = JSON.parse(contents);
  } catch {
    throw new Error(
      "Performance fixture private runtime bundle contains invalid JSON",
    );
  }
  const bundle = requireRecord(
    decoded,
    "Performance fixture private runtime bundle",
  );
  requireExactKeys(
    bundle,
    ["schema_id", "fixture_profile_id", "snapshot_key", "credential_sets"],
    "Performance fixture private runtime bundle",
  );
  if (
    bundle.schema_id !== "cartulary.performance_fixture_runtime.v2" ||
    bundle.fixture_profile_id !== expected.fixtureProfileId ||
    bundle.snapshot_key !== expected.snapshotKey ||
    !/^[a-f0-9]{64}$/u.test(expected.snapshotKey)
  ) {
    throw new Error(
      "Performance fixture private runtime bundle identity is inconsistent",
    );
  }
  if (
    !Array.isArray(bundle.credential_sets) ||
    bundle.credential_sets.length !== 1
  ) {
    throw new Error(
      "Performance fixture private runtime bundle credential sets are inconsistent",
    );
  }
  const credentialSet = requireRecord(
    bundle.credential_sets[0],
    "Performance fixture private runtime credential set",
  );
  requireExactKeys(
    credentialSet,
    ["set_id", "credential_kind", "credentials"],
    "Performance fixture private runtime credential set",
  );
  if (
    credentialSet.set_id !== expected.credentialSetId ||
    credentialSet.credential_kind !== "email_password" ||
    !Array.isArray(credentialSet.credentials)
  ) {
    throw new Error(
      "Performance fixture private runtime credential set is incompatible",
    );
  }
  const emails = new Set<string>();
  const accounts = credentialSet.credentials.map((rawAccount, index) => {
    const label = `Performance fixture private runtime account ${index + 1}`;
    const account = requireRecord(rawAccount, label);
    requireExactKeys(account, ["principal", "secret"], label);
    const email = requirePrivateString(account.principal, `${label} email`);
    const password = requirePrivateString(account.secret, `${label} password`);
    if (!email.includes("@") || email.length > 254) {
      throw new Error(`${label} email is invalid`);
    }
    if (password.length < 24 || password.length > 256) {
      throw new Error(`${label} password is invalid`);
    }
    if (emails.has(email)) {
      throw new Error(
        "Performance fixture private runtime bundle contains duplicate accounts",
      );
    }
    emails.add(email);
    return { email, password };
  });
  if (accounts.length !== expected.accountCount) {
    throw new Error(
      `Performance fixture private runtime bundle expected ${expected.accountCount} accounts, got ${accounts.length}`,
    );
  }
  return {
    accounts,
    fixtureProfileId: expected.fixtureProfileId,
    snapshotKey: expected.snapshotKey,
  } as const;
}

function requireRecord(value: unknown, label: string) {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value as Record<string, unknown>;
}

function requireExactKeys(
  value: Record<string, unknown>,
  keys: readonly string[],
  label: string,
) {
  if (
    JSON.stringify(Object.keys(value).sort()) !==
    JSON.stringify([...keys].sort())
  ) {
    throw new Error(`${label} has an invalid field set`);
  }
}

function requirePrivateString(value: unknown, label: string) {
  if (typeof value !== "string" || value === "") {
    throw new Error(`${label} is invalid`);
  }
  return value;
}
