import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { copyFileSync, existsSync, mkdirSync, readFileSync, symlinkSync } from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import test from "node:test";

import { createFixture, repository } from "./fixtures.mjs";

const require = createRequire(import.meta.url);
const biomeCLI = require.resolve("@biomejs/biome/bin/biome");
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
  copyFileSync(path.join(repository, "tools/task_surface.generated.mk"), path.join(fixture.root, "tasks.mk"));
  fixture.write("Makefile", `NODE_BIN := ${process.execPath}\nQ := @\ninclude tasks.mk\n`);
  mkdirSync(path.join(fixture.root, "node_modules/@biomejs"), { recursive: true });
  symlinkSync(path.dirname(path.dirname(biomeCLI)), path.join(fixture.root, "node_modules/@biomejs/biome"));
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
  const initialized = command("git", ["init", "--quiet", fixture.root]);
  assert.equal(initialized.status, 0, initialized.stderr);
  return {
    ...fixture,
    lint: () => command("make", ["--no-print-directory", "lint-markdown"], {
      cwd: fixture.root,
      env: { ...process.env, MAKEFLAGS: "", MFLAGS: "", MAKEOVERRIDES: "" },
    }),
  };
}

test("Markdown Make target lints directly with automatic fixes and isolated scope", (t) => {
  const fixture = fixtureFor(t);
  const first = fixture.lint();
  assert.equal(first.status, 0, first.stdout + first.stderr);
  assert.match(first.stdout + first.stderr, /Checked 5 files/);
  assert.equal(readFileSync(path.join(fixture.root, "example.md"), "utf8"), valid);
  assert.ok(!existsSync(path.join(fixture.root, ".cartulary")), "no retained harness reports");
  assert.ok(!existsSync(path.join(fixture.root, ".cache/cartulary")), "no lint cache or stamps");

  fixture.write("example.md", "# Example\n\n### Skipped level\n");
  const heading = fixture.lint();
  assert.notEqual(heading.status, 0);
  assert.match(heading.stdout + heading.stderr, /useConsistentHeadingLevel/);

  fixture.write("example.md", "# Example\n\n```\nexample\n```\n");
  const fence = fixture.lint();
  assert.notEqual(fence.status, 0);
  assert.match(fence.stdout + fence.stderr, /useFencedCodeLanguage/);

  fixture.write("example.md", valid);
  assert.equal(fixture.lint().status, 0);

  fixture.write(".gitignore", ".cache/\n");
  const changedIgnore = fixture.lint();
  assert.notEqual(changedIgnore.status, 0, "newly admitted invalid input must be checked");
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
