import assert from "node:assert/strict";
import { chmodSync, existsSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "../../workspace/child-process.mjs";
import { createFixture, repository } from "./fixtures.mjs";

export const authoredPackages = ["cmd/operator", "cmd/server", "internal/app/server", "internal/modules/auth", "db/migrations", "tools/testservices"]
  .map((name) => `github.com/JochiRaider/cartulary/${name}`);
export const discoveryPackages = [...authoredPackages, "github.com/JochiRaider/cartulary/internal/gen/contracts", "github.com/JochiRaider/cartulary/internal/gen/sql"];
export const analyzerDirectory = path.join(repository, "tools/go-analysis");
export const analyzerPackage = "github.com/JochiRaider/cartulary/tools/go-analysis/errordeclaudit";
export const discoveryArgs = ["list", "./cmd/...", "./internal/...", "./db/...", "./tools/..."];

// Small fake executables record the real wrapper's boundary. They do not model
// wrapper policy, parse findings or compute expected scanner arguments.
export function scannerFixture(t, { packages = discoveryPackages, goExit = 0, output = "", scannerExit = 0, failArgument, vetExit = 0, analysis = {} } = {}) {
  const fixture = createFixture(t);
  const moduleTools = (listed, listStatus, vetStatus, scanner) => ({
    go: { output: listed.map((name) => `${name}\n`).join(""), status: listStatus }, vet: { output: "", status: vetStatus }, scanner,
  });
  fixture.write("tool-config.json", {
    root: moduleTools(packages, goExit, vetExit, { output, status: scannerExit, failArgument }),
    analysis: moduleTools(analysis.packages ?? [analyzerPackage], analysis.goExit ?? 0, analysis.vetExit ?? 0,
      { output: "", status: analysis.scannerExit ?? 0 }),
  });
  for (const tool of ["go", "scanner"]) {
    fixture.write(`${tool}.mjs`, `#!${process.execPath}
import { appendFileSync, readFileSync } from "node:fs";
const modules = JSON.parse(readFileSync(new URL("./tool-config.json", import.meta.url)));
const cwd = process.cwd();
const configByTool = cwd === ${JSON.stringify(analyzerDirectory)} ? modules.analysis : modules.root;
const args = process.argv.slice(2);
const config = configByTool[${JSON.stringify(tool)} === "go" && args[0] === "vet" ? "vet" : ${JSON.stringify(tool)}];
const env = Object.fromEntries(["GOCACHE", "GOMODCACHE", "GOTMPDIR", "GOMAXPROCS", "PATH", "GOWORK", "GOFLAGS", "GOPROXY"].map(key => [key, process.env[key] ?? ""]));
appendFileSync(new URL("./${tool}-calls.jsonl", import.meta.url), JSON.stringify({ args, env, cwd }) + "\\n");
process.stdout.write(config.output);
process.exit(config.failArgument && !args.includes(config.failArgument) ? 0 : config.status);
`);
    chmodSync(path.join(fixture.root, `${tool}.mjs`), 0o700);
  }
  for (const directory of ["cache", "modules", "tmp"]) fixture.write(`${directory}/.keep`, "");
  const env = { ...process.env };
  for (const key of Object.keys(env)) {
    if (/^(?:GOSEC|GOVULNCHECK|STATICCHECK|FAKE_)/u.test(key) ||
      ["CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT", "CARTULARY_STEP_ARTIFACT_DIR", "TEST_SUPPORT_INVENTORY", "CARTULARY_TEST_SUPPORT_INVENTORY"].includes(key)) delete env[key];
  }
  Object.assign(env, { GO: path.join(fixture.root, "go.mjs"), NODE_BIN: process.execPath,
    GO_CACHE_DIR: path.join(fixture.root, "cache"), GO_MOD_CACHE_DIR: path.join(fixture.root, "modules"), GO_TMP_DIR: path.join(fixture.root, "tmp"),
    TMPDIR: path.join(fixture.root, "tmp"), STATICCHECK_BIN: path.join(fixture.root, "scanner.mjs"),
    GOSEC_BIN: path.join(fixture.root, "scanner.mjs"), GOVULNCHECK_BIN: path.join(fixture.root, "scanner.mjs"),
    CARTULARY_SEQUENCE_HOST_CPU_LIMIT: "4", CARTULARY_STEP_ARTIFACT_DIR: path.join(fixture.root, "artifacts") });
  return { ...fixture, env,
    invoke(wrapper, overrides = {}) {
      const child = spawnSync("bash", ["-c", 'umask 022; exec bash "$1"', "scanner-fixture", path.join(repository, `tools/harness/static-analysis/${wrapper}.sh`)],
        { cwd: repository, env: { ...env, ...overrides }, encoding: "utf8", timeout: 15000 });
      assert.equal(child.error, undefined);
      assert.equal(child.signal, null);
      return child;
    },
    calls(tool = "scanner") {
      const file = path.join(fixture.root, `${tool}-calls.jsonl`);
      return existsSync(file) ? readFileSync(file, "utf8").trim().split("\n").map((line) => JSON.parse(line)) : [];
    },
    profiles() { return readFileSync(path.join(fixture.root, "artifacts/security-profiles.jsonl"), "utf8").trim().split("\n").map((line) => JSON.parse(line)); },
  };
}

export function assertToolEnvironment(fixture, call, withGoPath = true) {
  assert.equal(call.env.GOCACHE, fixture.env.GO_CACHE_DIR);
  assert.equal(call.env.GOMODCACHE, fixture.env.GO_MOD_CACHE_DIR);
  assert.equal(call.env.GOTMPDIR, fixture.env.GO_TMP_DIR);
  if (withGoPath) assert.equal(call.env.PATH.split(path.delimiter)[0], fixture.root);
}

export function assertPrivateArtifacts(fixture, files) {
  assert.equal(statSync(path.join(fixture.root, "artifacts")).mode & 0o777, 0o700);
  for (const file of files) assert.equal(statSync(path.join(fixture.root, "artifacts", file)).mode & 0o777, 0o600, file);
}
