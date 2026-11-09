import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { chmodSync, copyFileSync, existsSync, mkdirSync, readFileSync } from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import test from "node:test";

import { createFixture, repository } from "./fixtures.mjs";

const require = createRequire(import.meta.url);
const biomeCLI = require.resolve("@biomejs/biome/bin/biome");
const wrapper = "tools/harness/static-analysis/markdown-biome.sh";
const configuration = "biome.markdown.json";
const valid = "---\nname: example\n---\n# Example\n\n## Details\n\n```js\nconst invalidExample = ;\n```\n";
const invalid = "# Example\n\n### Skipped level\n\n```\nexample\n```\n";

function command(file, args, options = {}) {
  const result = spawnSync(file, args, { encoding: "utf8", timeout: 30000, ...options });
  assert.ifError(result.error);
  return result;
}

function fixtureFor(t) {
  const fixture = createFixture(t);
  for (const relative of [wrapper, "tools/harness/static-analysis/markdown-biome-runner.sh",
    "tools/harness/readiness/cache-artifact.sh", "tools/workspace_layout.generated.sh",
    "package.json", "pnpm-lock.yaml"]) {
    const target = path.join(fixture.root, relative);
    mkdirSync(path.dirname(target), { recursive: true });
    copyFileSync(path.join(repository, relative), target);
  }
  const config = JSON.parse(readFileSync(path.join(repository, configuration), "utf8"));
  const policy = JSON.parse(readFileSync(path.join(repository, "tools/executable_input_policy.json"), "utf8"));
  // Exercise the real selection with neutral synthetic roots, never repository prose.
  config.files.includes = config.files.includes.map((pattern) => {
    for (const root of policy.restricted_roots) pattern = pattern.replace(`${root}/`, "manuals/");
    return pattern;
  });
  fixture.write(configuration, config);
  fixture.write(".gitignore", ".cache/\nignored.md\n");
  fixture.write("example.md", valid);
  fixture.write("manuals/spec/example.md", valid);
  fixture.write("manuals/spec/large.md", `${valid}\n${"Example paragraph.\n\n".repeat(60000)}`);
  fixture.write(".agents/skills/example/SKILL.md", valid);
  fixture.write("ignored.md", invalid);
  fixture.write("unselected/example.md", invalid);
  fixture.write("manuals/archive/example.md", invalid);
  fixture.write("manuals/handoffs/unselected.md", invalid);
  fixture.write("apps/web/src/example/README.md", valid);
  fixture.write(".cache/excluded.md", invalid);
  fixture.write("pnpm-test.sh", `#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == --dir && "$3" == exec && "$4" == biome ]]
cd "$2"
shift 4
printf 'run\\n' >>"$BIOME_TEST_RUNS"
exec "$BIOME_TEST_NODE" "$BIOME_TEST_CLI" "$@"
`);
  chmodSync(path.join(fixture.root, "pnpm-test.sh"), 0o755);
  const initialized = command("git", ["init", "--quiet", fixture.root]);
  assert.equal(initialized.status, 0, initialized.stderr);
  const runs = path.join(fixture.root, "biome-runs.log");
  return {
    ...fixture,
    config,
    stamp: path.join(fixture.root, ".cache/cartulary/static-analysis/outputs/lint-markdown.ok"),
    count: () => existsSync(runs) ? readFileSync(runs, "utf8").trim().split("\n").length : 0,
    lint: () => command("bash", [path.join(fixture.root, wrapper)], {
      cwd: fixture.root,
      env: { ...process.env, NODE_RUNTIME_DIR: path.dirname(path.dirname(process.execPath)),
        PNPM: path.join(fixture.root, "pnpm-test.sh"), BIOME_TEST_NODE: process.execPath,
        BIOME_TEST_CLI: biomeCLI, BIOME_TEST_RUNS: runs,
        CARTULARY_STATIC_ANALYSIS_CACHE_DIR: path.join(fixture.root, ".cache/cartulary/static-analysis"),
        CARTULARY_STATIC_ANALYSIS_DISABLE_CACHE: "0", CARTULARY_STATIC_ANALYSIS_FORCE: "0" },
    }),
  };
}

test("Markdown lint preserves scope and content and caches only successful inputs", (t) => {
  const fixture = fixtureFor(t);
  const first = fixture.lint();
  assert.equal(first.status, 0, first.stdout + first.stderr);
  assert.match(first.stdout + first.stderr, /Checked 5 files/);
  assert.equal(fixture.count(), 1);
  assert.ok(existsSync(fixture.stamp));
  assert.equal(readFileSync(path.join(fixture.root, "example.md"), "utf8"), valid);
  assert.equal(fixture.lint().status, 0);
  assert.equal(fixture.count(), 1, "unchanged inputs reuse the successful result");

  fixture.write("example.md", "# Example\n\n### Skipped level\n");
  const heading = fixture.lint();
  assert.notEqual(heading.status, 0);
  assert.match(heading.stdout + heading.stderr, /useConsistentHeadingLevel/);
  assert.equal(fixture.count(), 2);
  assert.ok(!existsSync(fixture.stamp), "a failed run removes the stale success stamp");

  fixture.write("example.md", "# Example\n\n```\nexample\n```\n");
  const fence = fixture.lint();
  assert.notEqual(fence.status, 0);
  assert.match(fence.stdout + fence.stderr, /useFencedCodeLanguage/);
  assert.equal(fixture.count(), 3);
  assert.ok(!existsSync(fixture.stamp));

  fixture.write("example.md", valid);
  assert.equal(fixture.lint().status, 0);
  assert.equal(fixture.count(), 4);
  fixture.write("example.md", "# First\n\n# Second\n");
  assert.equal(fixture.lint().status, 0);
  assert.equal(fixture.count(), 5);
  fixture.config.linter.rules.nursery.useSingleTopLevelHeading = "error";
  fixture.write(configuration, fixture.config);
  const enabledRule = fixture.lint();
  assert.notEqual(enabledRule.status, 0);
  assert.match(enabledRule.stdout + enabledRule.stderr, /useSingleTopLevelHeading/);
  assert.equal(fixture.count(), 6, "configuration changes invalidate successful cached results");
  fixture.config.linter.rules.nursery.useSingleTopLevelHeading = "off";
  fixture.write(configuration, fixture.config);
  const changedConfiguration = fixture.lint();
  assert.equal(changedConfiguration.status, 0, changedConfiguration.stdout + changedConfiguration.stderr);
  assert.equal(fixture.count(), 7);

  fixture.write(".gitignore", ".cache/\n");
  const changedIgnore = fixture.lint();
  assert.notEqual(changedIgnore.status, 0, "newly admitted invalid input must be checked");
  assert.equal(fixture.count(), 8);
});

test("code lint force-ignores prose while continuing to lint source", (t) => {
  const fixture = createFixture(t);
  copyFileSync(path.join(repository, "biome.json"), path.join(fixture.root, "biome.json"));
  fixture.write(".gitignore", "");
  fixture.write("apps/web/src/example.ts", "debugger;\n");
  fixture.write("apps/web/src/README.md", invalid);
  fixture.write("apps/web/src/example.markdown", invalid);
  fixture.write("apps/web/src/example.mD", invalid);
  fixture.write("apps/web/src/rEaDmE.js", "debugger;\n");
  assert.equal(command("git", ["init", "--quiet", fixture.root]).status, 0);
  const result = command(process.execPath, [biomeCLI, "lint", "--verbose", "apps/web/src"], { cwd: fixture.root });
  const output = result.stdout + result.stderr;
  assert.notEqual(result.status, 0);
  assert.match(output, /noDebugger/);
  assert.doesNotMatch(output, /README\.md|example\.markdown|example\.mD|rEaDmE\.js/);
});
