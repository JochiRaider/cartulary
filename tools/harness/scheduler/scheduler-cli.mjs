export function parseSchedulerRunnerArgs(
  argv,
  {
    defaultManifestPath,
    parseResourceLimitOverride,
    usageText,
    allowDeferSummary = false,
  },
) {
  const options = {
    manifest: defaultManifestPath,
    target: "",
    deferSummary: false,
    resourceLimitOverrides: new Map(),
  };
  const usage = () => {
    process.stderr.write(usageText);
    process.exit(2);
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") {
      options.target = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--manifest") {
      options.manifest = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (allowDeferSummary && arg === "--defer-summary") {
      options.deferSummary = true;
      continue;
    }
    if (arg === "--resource-limit") {
      const value = argv[index + 1] ?? "";
      const [resource, amount] = parseResourceLimitOverride(value);
      options.resourceLimitOverrides.set(resource.trim(), amount);
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.target || !options.manifest) {
    usage();
  }
  return options;
}
