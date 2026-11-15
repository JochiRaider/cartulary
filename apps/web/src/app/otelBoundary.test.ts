import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../../..",
);
const workspacePackageDirectories: Record<string, string> = {
  "@cartulary/grid-adapter": "packages/grid-adapter",
  "@cartulary/protocol-ts": "packages/protocol-ts",
  "@cartulary/ui-contracts": "packages/ui-contracts",
  "@cartulary/view-contracts": "packages/view-contracts",
};
const forbiddenBrowserTelemetryPackages = [
  /^@opentelemetry\//u,
  /sentry|datadog|newrelic|honeycomb|logrocket|fullstory|posthog/iu,
  /segment|mixpanel|amplitude|rrweb|session-replay|google-analytics|gtag/iu,
];

function runtimeDependencies(directory: string): string[] {
  const manifest = JSON.parse(
    readFileSync(path.join(repoRoot, directory, "package.json"), "utf8"),
  ) as { dependencies?: Record<string, string> };
  return Object.keys(manifest.dependencies ?? {});
}

function browserPackageClosure(): Set<string> {
  const closure = new Set<string>();
  const pending = runtimeDependencies("apps/web");
  while (pending.length > 0) {
    const name = pending.pop() as string;
    if (closure.has(name)) continue;
    closure.add(name);
    const workspace = workspacePackageDirectories[name];
    if (workspace !== undefined)
      pending.push(...runtimeDependencies(workspace));
  }
  return closure;
}

describe("OpenTelemetry browser boundary", () => {
  it("keeps the browser package closure free of telemetry exporters, vendor SDKs, session replay and analytics", () => {
    const closure = browserPackageClosure();
    // Control: the closure is the real browser graph, not an empty read.
    for (const name of [...Object.keys(workspacePackageDirectories), "react"])
      expect(closure).toContain(name);
    const forbidden = [...closure].filter((name) =>
      forbiddenBrowserTelemetryPackages.some((pattern) => pattern.test(name)),
    );
    expect(forbidden).toEqual([]);
  });
});
