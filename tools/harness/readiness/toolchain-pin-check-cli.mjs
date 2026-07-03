import { readFileSync } from "node:fs";
import path from "node:path";

const toolchainPinsSchemaID = "cartulary.toolchain_pins.v1";

function usage() {
  process.stderr.write("usage: check-toolchain-pins.mjs [--root <path>]\n");
}

function parseArgs(argv) {
  let root = process.cwd();
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--root") {
      const value = argv[i + 1];
      if (!value) {
        usage();
        process.exit(2);
      }
      root = value;
      i += 1;
      continue;
    }
    usage();
    process.exit(2);
  }
  return path.resolve(root);
}

function readRepoFile(root, relativePath) {
  return readFileSync(path.join(root, relativePath), "utf8");
}

function requireString(record, field, file) {
  const value = record?.[field];
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${file}: ${field} must be a non-empty string`);
  }
  return value;
}

function requireObject(record, field, file) {
  const value = record?.[field];
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${file}: ${field} must be an object`);
  }
  return value;
}

function loadExpected(root) {
  const file = "tools/toolchain_pins.json";
  const pins = JSON.parse(readRepoFile(root, file));
  if (pins.schema_id !== toolchainPinsSchemaID) {
    throw new Error(`${file}: schema_id must be ${toolchainPinsSchemaID}`);
  }
  const tools = requireObject(pins, "tools", file);
  return {
    modulePath: requireString(pins, "module_path", file),
    goVersion: requireString(pins, "go_version", file),
    goToolchain: requireString(pins, "go_toolchain", file),
    nodeVersion: requireString(pins, "node_version", file),
    pnpmVersion: requireString(pins, "pnpm_version", file),
    sqlcTool: requireString(tools, "sqlc", file),
    gooseTool: requireString(tools, "goose", file),
    staticcheckTool: requireString(tools, "staticcheck", file),
    govulncheckTool: requireString(tools, "govulncheck", file),
    gosecTool: requireString(tools, "gosec", file),
    cyclonedxGomodTool: requireString(tools, "cyclonedx_gomod", file),
    syftTool: requireString(tools, "syft", file),
    shellcheckVersion: requireString(pins, "shellcheck_version", file),
    testcontainersGoVersion: requireString(
      pins,
      "testcontainers_go_version",
      file,
    ),
  };
}

function reportMismatch(mismatches, file, field, expectedValue, actualValue) {
  mismatches.push({
    file,
    field,
    expected: expectedValue,
    actual: actualValue ?? "(missing)",
  });
}

function matchLine(text, regex) {
  const match = regex.exec(text);
  return match?.[1];
}

function parseMakeVariable(makefile, name) {
  const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  return matchLine(makefile, new RegExp(`^${escaped}\\s*(?:\\?=|:=)\\s*(\\S+)\\s*$`, "m"));
}

function checkEqual(mismatches, file, field, expectedValue, actualValue) {
  if (actualValue !== expectedValue) {
    reportMismatch(mismatches, file, field, expectedValue, actualValue);
  }
}

function checkMakefile(root, mismatches, expected) {
  const file = "Makefile";
  const makefile = readRepoFile(root, file);
  checkEqual(
    mismatches,
    file,
    "NODE_VERSION",
    expected.nodeVersion,
    parseMakeVariable(makefile, "NODE_VERSION"),
  );
  checkEqual(
    mismatches,
    file,
    "PNPM_VERSION",
    expected.pnpmVersion,
    parseMakeVariable(makefile, "PNPM_VERSION"),
  );
  checkEqual(
    mismatches,
    file,
    "SQLC_TOOL",
    expected.sqlcTool,
    parseMakeVariable(makefile, "SQLC_TOOL"),
  );
  checkEqual(
    mismatches,
    file,
    "GOOSE_TOOL",
    expected.gooseTool,
    parseMakeVariable(makefile, "GOOSE_TOOL"),
  );
  checkEqual(
    mismatches,
    file,
    "STATICCHECK_TOOL",
    expected.staticcheckTool,
    parseMakeVariable(makefile, "STATICCHECK_TOOL"),
  );
  checkEqual(
    mismatches,
    file,
    "GOVULNCHECK_TOOL",
    expected.govulncheckTool,
    parseMakeVariable(makefile, "GOVULNCHECK_TOOL"),
  );
  checkEqual(
    mismatches,
    file,
    "GOSEC_TOOL",
    expected.gosecTool,
    parseMakeVariable(makefile, "GOSEC_TOOL"),
  );
  checkEqual(
    mismatches,
    file,
    "CYCLONEDX_GOMOD_TOOL",
    expected.cyclonedxGomodTool,
    parseMakeVariable(makefile, "CYCLONEDX_GOMOD_TOOL"),
  );
  checkEqual(
    mismatches,
    file,
    "SYFT_TOOL",
    expected.syftTool,
    parseMakeVariable(makefile, "SYFT_TOOL"),
  );
  checkEqual(
    mismatches,
    file,
    "SHELLCHECK_VERSION",
    expected.shellcheckVersion,
    parseMakeVariable(makefile, "SHELLCHECK_VERSION"),
  );
  checkEqual(
    mismatches,
    file,
    "TESTCONTAINERS_GO_VERSION",
    expected.testcontainersGoVersion,
    parseMakeVariable(makefile, "TESTCONTAINERS_GO_VERSION"),
  );
}

function checkPackageJson(root, mismatches, expected) {
  const file = "package.json";
  const packageJson = JSON.parse(readRepoFile(root, file));
  checkEqual(
    mismatches,
    file,
    "packageManager",
    `pnpm@${expected.pnpmVersion}`,
    packageJson.packageManager,
  );
  checkEqual(
    mismatches,
    file,
    "engines.node",
    expected.nodeVersion,
    packageJson.engines?.node,
  );
}

function checkGoMod(root, mismatches, expected) {
  const file = "go.mod";
  const goMod = readRepoFile(root, file);
  checkEqual(
    mismatches,
    file,
    "module",
    expected.modulePath,
    matchLine(goMod, /^module\s+(\S+)\s*$/m),
  );
  checkEqual(
    mismatches,
    file,
    "go",
    expected.goVersion,
    matchLine(goMod, /^go\s+(\S+)\s*$/m),
  );
  checkEqual(
    mismatches,
    file,
    "toolchain",
    expected.goToolchain,
    matchLine(goMod, /^toolchain\s+(\S+)\s*$/m),
  );
  checkEqual(
    mismatches,
    file,
    "github.com/pressly/goose/v3",
    expected.gooseTool.split("@")[1],
    matchLine(goMod, /^\s*github\.com\/pressly\/goose\/v3\s+(\S+)/m),
  );
  checkEqual(
    mismatches,
    file,
    "github.com/testcontainers/testcontainers-go",
    expected.testcontainersGoVersion,
    matchLine(goMod, /^\s*github\.com\/testcontainers\/testcontainers-go\s+(\S+)/m),
  );
}

function checkBootstrapNodeRuntime(root, mismatches, expected) {
  const file = "tools/harness/readiness/bootstrap-node-runtime.sh";
  const script = readRepoFile(root, file);
  checkEqual(
    mismatches,
    file,
    "NODE_VERSION default",
    expected.nodeVersion,
    matchLine(script, /^NODE_VERSION="\$\{NODE_VERSION:-(.+)\}"\s*$/m),
  );
}

function checkBootstrapShellcheck(root, mismatches, expected) {
  const file = "tools/harness/readiness/bootstrap-shellcheck.sh";
  const script = readRepoFile(root, file);
  checkEqual(
    mismatches,
    file,
    "SHELLCHECK_VERSION default",
    expected.shellcheckVersion,
    matchLine(script, /^SHELLCHECK_VERSION="\$\{SHELLCHECK_VERSION:-(.+)\}"\s*$/m),
  );
}

function checkReadme(root, mismatches, expected) {
  const file = "README.md";
  const readme = readRepoFile(root, file);
  const goLine = `- Go \`${expected.goVersion}\` with toolchain \`${expected.goToolchain}\``;
  const nodeLine = `- Node.js \`${expected.nodeVersion}\``;
  const pnpmLine = `- pnpm \`${expected.pnpmVersion}\``;
  const staticcheckLine = `- Staticcheck \`${expected.staticcheckTool.split("@")[1]}\``;
  const govulncheckLine = `- Govulncheck \`${expected.govulncheckTool.split("@")[1]}\``;
  const gosecLine = `- Gosec \`${expected.gosecTool.split("@")[1]}\``;
  const shellcheckLine = `- ShellCheck \`${expected.shellcheckVersion}\``;
  checkEqual(
    mismatches,
    file,
    "Go pin line",
    goLine,
    readme.split("\n").find((line) => line.startsWith("- Go `")),
  );
  checkEqual(
    mismatches,
    file,
    "Node.js pin line",
    nodeLine,
    readme.split("\n").find((line) => line.startsWith("- Node.js `")),
  );
  checkEqual(
    mismatches,
    file,
    "pnpm pin line",
    pnpmLine,
    readme.split("\n").find((line) => line.startsWith("- pnpm `")),
  );
  checkEqual(
    mismatches,
    file,
    "Staticcheck pin line",
    staticcheckLine,
    readme.split("\n").find((line) => line.startsWith("- Staticcheck `")),
  );
  checkEqual(
    mismatches,
    file,
    "Govulncheck pin line",
    govulncheckLine,
    readme.split("\n").find((line) => line.startsWith("- Govulncheck `")),
  );
  checkEqual(
    mismatches,
    file,
    "Gosec pin line",
    gosecLine,
    readme.split("\n").find((line) => line.startsWith("- Gosec `")),
  );
  checkEqual(
    mismatches,
    file,
    "ShellCheck pin line",
    shellcheckLine,
    readme.split("\n").find((line) => line.startsWith("- ShellCheck `")),
  );
}

function checkAgents(root, mismatches, expected) {
  const file = "AGENTS.md";
  const agents = readRepoFile(root, file);
  const pinnedToolsLine = `- Pinned bootstrap tools: \`${expected.sqlcTool}\`, \`${expected.gooseTool}\`, \`${expected.staticcheckTool}\`, \`${expected.govulncheckTool}\`, \`${expected.gosecTool}\`, \`${expected.cyclonedxGomodTool}\`, \`${expected.syftTool}\`, ShellCheck \`${expected.shellcheckVersion}\`, and \`github.com/testcontainers/testcontainers-go ${expected.testcontainersGoVersion}\`.`;
  checkEqual(
    mismatches,
    file,
    "Pinned bootstrap tools line",
    pinnedToolsLine,
    agents.split("\n").find((line) => line.startsWith("- Pinned bootstrap tools:")),
  );
}

function main() {
  const root = parseArgs(process.argv.slice(2));
  const expected = loadExpected(root);
  const mismatches = [];

  checkMakefile(root, mismatches, expected);
  checkPackageJson(root, mismatches, expected);
  checkGoMod(root, mismatches, expected);
  checkBootstrapNodeRuntime(root, mismatches, expected);
  checkBootstrapShellcheck(root, mismatches, expected);
  checkReadme(root, mismatches, expected);
  checkAgents(root, mismatches, expected);

  if (mismatches.length > 0) {
    for (const mismatch of mismatches) {
      process.stderr.write(
        `${mismatch.file}: ${mismatch.field} mismatch: expected ${mismatch.expected}, got ${mismatch.actual}\n`,
      );
    }
    process.exit(1);
  }

  console.log("toolchain pins verified");
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
