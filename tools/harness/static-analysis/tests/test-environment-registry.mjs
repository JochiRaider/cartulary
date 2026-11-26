import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";

import { childEnvironment } from "../../runtime/child-environment.mjs";
import { environmentRoles, globPattern, inspectEnvironmentRegistry, readEnvironmentRegistry, unitFixtureMarker } from "../environment-registry.mjs";

// Names are assembled so this file mentions no registry name itself.
const name = (suffix) => `CARTULARY${"_"}${suffix}`;

function fixture(t, files) {
  const root = mkdtempSync(path.join(tmpdir(), "cartulary-environment-registry."));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  for (const [relative, text] of Object.entries(files)) {
    mkdirSync(path.dirname(path.join(root, relative)), { recursive: true });
    writeFileSync(path.join(root, relative), text);
  }
  return { root, paths: Object.keys(files).sort() };
}

function registry(roles = {}, owners = {}) {
  return {
    schema_id: "cartulary.harness_environment_registry.v1",
    scopes: { excluded: ["notes/**", "**/*.generated.*"], harness: ["tools/**", "**/*_test.go"] },
    inherited: { names: ["HOME", "PATH"], prefixes: ["LC_"] },
    explicit_launchers: ["tools/launcher.go", "tools/launcher.mjs"],
    owners,
    roles: Object.fromEntries(environmentRoles.map((role) => [role, roles[role] ?? []])),
  };
}

test("globs match within and across path segments", () => {
  assert.ok(globPattern("tools/**").test("tools/a/b.mjs"));
  assert.ok(globPattern("**/*_test.go").test("internal/x/y_test.go"));
  assert.ok(globPattern("**/*_test.go").test("top_test.go"));
  assert.ok(!globPattern("tools/*.mjs").test("tools/a/b.mjs"));
  assert.ok(globPattern("apps/web/*.config.ts").test("apps/web/vitest.config.ts"));
});

test("every harness name is registered or declared, and nothing registered is stale", (t) => {
  const { root, paths } = fixture(t, {
    "tools/run.sh": `echo "$${name("RUN_ID")}" "$${name("UNKNOWN")}" "$${name("PUBLIC")}"\n`,
    "tools/generated.generated.mk": `${name("ONLY_GENERATED")} := 1\n`,
    "notes/guide.txt": `${name("IN_NOTES")}\n`,
    "internal/app/main.go": `os.Getenv("${name("PRODUCT_ONLY")}")\n`,
  });
  const result = inspectEnvironmentRegistry(root, {
    registry: registry({ invocation_context: [name("RUN_ID"), name("STALE")] }),
    declared: new Set([name("PUBLIC")]),
    paths,
  });
  assert.deepEqual(result.findings, [
    `${name("STALE")} (invocation_context) is stale: no harness code mentions it`,
    `${name("UNKNOWN")} is unregistered (first harness mention: tools/run.sh)`,
  ]);
  assert.equal(result.summary.harness_names, 3);
});

test("a name both sides mention must be a product interface, and only then", (t) => {
  const { root, paths } = fixture(t, {
    "tools/stack.sh": `export ${name("HTTP_PORT")}=1 ${name("SHARED")}=2 ${name("CLAIMED")}=3\n`,
    "internal/app/server.go": `os.Getenv("${name("HTTP_PORT")}"); os.Getenv("${name("SHARED")}")\n`,
  });
  const result = inspectEnvironmentRegistry(root, {
    registry: registry({ product_interface: [name("CLAIMED"), name("HTTP_PORT")], tool_option: [name("SHARED")] }),
    declared: new Set(),
    paths,
  });
  assert.deepEqual(result.findings, [
    `${name("CLAIMED")} is a product_interface that no product code mentions`,
    `${name("SHARED")} is tool_option but product code mentions it (internal/app/server.go)`,
  ]);
});

test("a name has one role and a declared input is not registered again", (t) => {
  const { root, paths } = fixture(t, { "tools/a.sh": `${name("TWICE")} ${name("PUBLIC")}\n` });
  const result = inspectEnvironmentRegistry(root, {
    registry: registry({ mode_switch: [name("TWICE")], tool_option: [name("PUBLIC"), name("TWICE")] }),
    declared: new Set([name("PUBLIC")]),
    paths,
  });
  assert.deepEqual(result.findings, [
    `${name("PUBLIC")} is a declared public input; remove it from roles.tool_option`,
    `${name("TWICE")} is registered as both mode_switch and tool_option`,
  ]);
});

test("product configuration and partial names are not harness names", (t) => {
  const { root, paths } = fixture(t, {
    "tools/a.mjs": `const a = "CARTULARY__SERVER__PORT"; const b = \`${name("PREFIX")}_\${x}\`; const c = "${name("LOWER")}x";\n`,
  });
  const result = inspectEnvironmentRegistry(root, { registry: registry(), declared: new Set(), paths });
  assert.deepEqual(result.findings, []);
  assert.equal(result.summary.harness_names, 0);
});

test("the registry's shape is closed and inherited names are not harness names", (t) => {
  const write = (registryValue) => fixture(t, { "tools/harness_environment_registry.json": JSON.stringify(registryValue) }).root;
  const valid = { ...registry(), inherited: { names: ["HOME", "PATH"], prefixes: ["LC_"] } };
  assert.doesNotThrow(() => readEnvironmentRegistry(write(valid)));
  assert.throws(() => readEnvironmentRegistry(write({ ...valid, extra: true })), /exactly schema_id, scopes, inherited, explicit_launchers, owners and roles/u);
  const owner = (names) => ({ paths: ["tools/owner/**"], names });
  assert.doesNotThrow(() => readEnvironmentRegistry(write({ ...valid, owners: { suite: owner([name("SUITE_ID")]) } })));
  assert.throws(() => readEnvironmentRegistry(write({ ...valid, owners: { suite: { ...owner([]), extra: [] } } })), /owners\.suite must hold exactly paths and names/u);
  assert.throws(() => readEnvironmentRegistry(write({ ...valid, owners: { suite: owner([name("SUITE_ID"), name("CALL_MODE")]) } })), /ASCII-sorted/u);
  assert.throws(() => readEnvironmentRegistry(write({ ...valid, owners: { a: owner([name("SUITE_ID")]), b: owner([name("SUITE_ID")]) } })), /has two owners: a and b/u);
  assert.throws(() => readEnvironmentRegistry(write({ ...valid, inherited: { names: ["PATH", "HOME"], prefixes: [] } })), /inherited\.names must be ASCII-sorted/u);
  assert.throws(() => readEnvironmentRegistry(write({ ...valid, inherited: { names: [name("RUN_ID")], prefixes: [] } })), /not a valid name/u);
  assert.throws(() => readEnvironmentRegistry(write({ ...valid, inherited: { names: [], prefixes: ["LC"] } })), /not a valid name/u);
});

test("a child environment carries only basics, forwarded names and set values", () => {
  const parent = {
    PATH: "/bin", HOME: "/home/operator", LC_ALL: "C.UTF-8", CARTULARY_MACHINE_CACHE_DIR: "/cache",
    [name("HOST_ADMISSION_LEASE")]: "unit-lease", [name("HARNESS_SERVICE_DEPENDENCIES")]: "postgres",
    AGENT_TOKEN: "operator secret", GOFLAGS: "-mod=mod",
  };
  const child = childEnvironment(parent, { pass: [name("HOST_ADMISSION_LEASE")], set: { [name("TEST_SUITE_ID")]: "session", HOME: undefined } });
  assert.deepEqual(child, {
    PATH: "/bin", LC_ALL: "C.UTF-8", CARTULARY_MACHINE_CACHE_DIR: "/cache",
    [name("HOST_ADMISSION_LEASE")]: "unit-lease", [name("TEST_SUITE_ID")]: "session",
  });
});

test("an explicit launcher never passes its environment through", (t) => {
  const { root, paths } = fixture(t, {
    "tools/launcher.mjs": [
      "spawn(a, [], { env: { ...process.env, X: 1 } });",
      "spawn(b, [], { env: process.env });",
      "run({ environment: { ...environment, Y: 2 } });",
      `spawn(c, [], { env: process.env }); // ${unitFixtureMarker}`,
      "spawn(d, [], { env: childEnvironment(process.env, { set: { Z: 3 } }) });",
      "const local = { ...descriptor.environment, ...startEnvironment };",
    ].join("\n"),
    "tools/other.mjs": "spawn(e, [], { env: { ...process.env } });\n",
  });
  const result = inspectEnvironmentRegistry(root, { registry: registry(), declared: new Set(), paths });
  assert.deepEqual(result.findings, [
    "tools/launcher.mjs:1 passes an environment through; build it with childEnvironment",
    "tools/launcher.mjs:2 passes an environment through; build it with childEnvironment",
    "tools/launcher.mjs:3 passes an environment through; build it with childEnvironment",
  ]);
});

test("a Go launcher builds a child's environment with childenv, never from os.Environ", (t) => {
  const { root, paths } = fixture(t, {
    "tools/launcher.go": [
      "cmd.Env = append(os.Environ(), extra...)",
      "for _, pair := range os.Environ() {",
      "cmd.Env = os.Environ()",
      "cmd.Env, err = childenv.Build(root, os.Environ(), nil, set)",
    ].join("\n"),
  });
  const result = inspectEnvironmentRegistry(root, { registry: registry(), declared: new Set(), paths });
  assert.deepEqual(result.findings, [
    "tools/launcher.go:1 passes an environment through; build it with childEnvironment",
    "tools/launcher.go:2 passes an environment through; build it with childEnvironment",
    "tools/launcher.go:3 passes an environment through; build it with childEnvironment",
  ]);
});


test("only an owned name's owner writes it; others read or forward it", (t) => {
  const suite = name("SUITE_ID");
  const { root, paths } = fixture(t, {
    "tools/owner/attach.mjs": `export const attach = (id) => ({ ${suite}: id });\n`,
    "tools/writer.mjs": [
      `env["${suite}"] = id;`,
      `process.env.${suite} = id;`,
      `spawn(a, [], { env: { ${suite}: id } });`,
      `run(["${suite}=1"]);`,
      `const value = env.${suite} === id;`,
      `childEnvironment(env, { pass: ["${suite}"] });`,
      `switch (key) { case "${suite}": break; }`,
    ].join("\n"),
    "tools/writer.go": [
      `environment["${suite}"] = id`,
      `os.Setenv("${suite}", id)`,
      `return map[string]string{"${suite}": id}`,
      `value := environment["${suite}"]`,
      `case "${suite}":`,
    ].join("\n"),
    "tools/writer_test.go": `t.Setenv("${suite}", "fixture")\n`,
  });
  const result = inspectEnvironmentRegistry(root, {
    registry: registry({ invocation_context: [suite] }, { suite_services: { paths: ["tools/owner/**"], names: [suite] } }),
    declared: new Set(),
    paths,
  });
  const writes = (line) => `${line} writes ${suite}, which suite_services owns; call its function or forward the name`;
  assert.deepEqual(result.findings, [
    writes("tools/writer.go:1"), writes("tools/writer.go:2"), writes("tools/writer.go:3"),
    writes("tools/writer.mjs:1"), writes("tools/writer.mjs:2"), writes("tools/writer.mjs:3"), writes("tools/writer.mjs:4"),
  ]);
});

test("a Go constant for an owned name stands for it wherever it is written", (t) => {
  const suite = name("SUITE_ID");
  const { root, paths } = fixture(t, {
    "tools/owner/env.go": `const SuiteIDEnv = "${suite}"\n\nfunc Attach(id string) map[string]string { return map[string]string{SuiteIDEnv: id} }\n`,
    "tools/consumer.go": [
      "environment[owner.SuiteIDEnv] = id",
      "return map[string]string{owner.SuiteIDEnv: id}",
      "os.Setenv(owner.SuiteIDEnv, id)",
      "value := environment[owner.SuiteIDEnv]",
      "case owner.SuiteIDEnv:",
      "maps.Copy(environment, owner.Attach(id))",
    ].join("\n"),
  });
  const result = inspectEnvironmentRegistry(root, {
    registry: registry({ invocation_context: [suite] }, { suite_services: { paths: ["tools/owner/**"], names: [suite] } }),
    declared: new Set(),
    paths,
  });
  assert.deepEqual(result.findings, [1, 2, 3].map((line) => `tools/consumer.go:${line} writes ${suite}, which suite_services owns; call its function or forward the name`));
});
