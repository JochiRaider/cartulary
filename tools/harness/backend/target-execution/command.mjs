export function shellQuote(value) {
  const text = String(value);
  if (/^[A-Za-z0-9_@%+=:,./-]+$/u.test(text)) {
    return text;
  }
  return `'${text.replaceAll("'", "'\"'\"'")}'`;
}

export function renderCommand(args) {
  return args.map(shellQuote).join(" ");
}

export function fixtureEnv(policy = {}) {
  return {
    CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS: policy.tests ?? "",
    CARTULARY_POSTGRES_FIXTURE_POLICY_PACKAGES: policy.packages ?? "",
    CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT: policy.defaultPolicy ?? "",
    CARTULARY_POSTGRES_RESET_TABLES_TESTS: policy.resetTests ?? "",
    CARTULARY_POSTGRES_RESET_TABLES_PACKAGES: policy.resetPackages ?? "",
  };
}

export function goTestEnvAssignments(ctx, policy = {}) {
  const assignments = [
    `GOCACHE=${ctx.goCacheDir}`,
    `GOMODCACHE=${ctx.goModCacheDir}`,
  ];
  const values = fixtureEnv(policy);
  for (const [name, value] of Object.entries(values)) {
    if (value) {
      assignments.push(`${name}=${value}`);
    }
  }
  return assignments;
}

export function renderGoTestCommand(ctx, regex, args, policy = {}) {
  return renderCommand([
    "env",
    ...goTestEnvAssignments(ctx, policy),
    ctx.goBin,
    "test",
    "-json",
    "-run",
    regex,
    ...args,
  ]);
}

export function goChildEnv(ctx, policy = {}) {
  return {
    ...ctx.env,
    GOCACHE: ctx.goCacheDir,
    GOMODCACHE: ctx.goModCacheDir,
    ...fixtureEnv(policy),
  };
}
