import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import test from "node:test";
import { spawnSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";
import { canonicalTimingAccounting, projectResourcePressure, reduceCanonicalUnitIntervals } from "../../evidence-accounting/index.mjs";
import { compactJSONString, validateSchemaSync, redactString, redactValue } from "../../contract/index.mjs";
import { resolveExactRunDir } from "../observability.mjs";
import { deliver, headersFromFile, validatedEndpoint } from "../otel-export-cli.mjs";
import { renderInstrumentationDefinitions } from "../../generated-artifacts/instrumentation-definitions.mjs";
import { renderNativeRedactionPolicy } from "../../generated-artifacts/redaction-policy.mjs";
import { readFileSync } from "node:fs";
import { createResourceAccumulator } from "../resource-accumulator.mjs";
import Ajv2020 from "ajv/dist/2020.js";
import { createRuntime } from "../../runtime/session-client.mjs";
import { createLaunchContext, readLaunchContext } from "../../runtime/launch-context.mjs";
import { openExecutionChannel } from "../../test-support/execution-channel-fixture.mjs";
import { readExecutionIndex } from "../execution-reader.mjs";
import { instrumentationPolicy } from "../instrumentation-policy.mjs";
import { activityTiming, beginActivity } from "../../runtime/local-activity.mjs";


test("URL redaction removes complete userinfo and preserves unrelated URL components", () => {
  for (const [input, expected] of [
    ["https://PrivateFixtureAlpha@example.test/path", "https://<redacted>@example.test/path"],
    ["https://PrivateUser:PrivatePassword@example.test:443/path?q=1", "https://<redacted>@example.test:443/path?q=1"],
    ["https://PrivateUser:@example.test", "https://<redacted>@example.test"],
    ["https://Private%3AToken@example.test", "https://<redacted>@example.test"],
    ["postgresql://PrivateUser:PrivatePassword@localhost/db", "postgresql://<redacted>@localhost/db"],
    ["https://example.test/path@item?q=a@b#c@d", "https://example.test/path@item?q=a@b#c@d"],
    ["https://example.test?q=a@b", "https://example.test?q=a@b"],
    ["https://example.test#c@d", "https://example.test#c@d"],
  ]) {
    assert.equal(redactString(input), expected);
    assert.equal(redactString(expected), expected);
    assert.deepEqual(redactValue({ url: input, count: 1, available: true, absent: null }),
      { url: expected, count: 1, available: true, absent: null });
  }
});

test("local activities retain independent overlapping durations, unavailable clocks and incomplete ends", () => {
  const options = { repoRoot: ".", environment: {}, activity: "report_parse" };
  const tick = (value, identity = "clock:a") => ({ clock: "node_monotonic", clock_identity: identity, resolution_ms: 0.000001, monotonic_ms: value });
  let now = 10;
  const read = () => tick(now);
  const outer = beginActivity(options, read);
  now = 12; const inner = beginActivity(options, read);
  now = 15; assert.equal(inner.finish("failed").duration_ms, 3);
  now = 20; assert.equal(outer.finish("passed").duration_ms, 10);
  assert.equal(outer.finish("failed"), null);
  for (const end of [tick(5), tick(20, "clock:b"), null]) {
    assert.equal(activityTiming(tick(10), end).duration_ms, null);
  }
  const unavailable = beginActivity(options, () => { throw new Error("clock unavailable"); }).finish("cancelled");
  assert.equal(unavailable.availability, "unavailable");
  assert.equal(unavailable.duration_ms, null);
});

test("neutral launch records retain packed parentage, unlaunched failure and one terminal outcome", (t) => {
  const repoRoot = path.resolve(import.meta.dirname, "../../../..");
  const runRoot = temporary(t), runID = path.basename(runRoot);
  const runtime = createRuntime({ repoRoot, runRoot, runID });
  t.after(() => runtime.close());
  const digest = `sha256:${"a".repeat(64)}`;
  const { policy, digest: policyDigest } = instrumentationPolicy();
  const manifest = { run_id: runID, source_digest: digest, graph_digest: digest, instrumentation: { mode: "basic", policy_digest: policyDigest } };
  const session = openExecutionChannel(runtime);
  const environment = { CARTULARY_TEST_RESULTS_DIR: path.dirname(runRoot), CARTULARY_TEST_RUN_ID: runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID, ...session.environment };
  const parent = createLaunchContext({ repoRoot, environment, unitID: "u", rowIDs: ["a", "b"], producer: "unit" });
  parent.spawned();
  const child = createLaunchContext({ repoRoot, environment: { ...environment, ...parent.environment }, producer: "go" });
  const activity = beginActivity({ repoRoot, environment, launch: child.identity, activity: "migration" });
  activity.finish("failed");
  beginActivity({ repoRoot, environment, launch: parent.identity, activity: "fixture_reset" });
  child.settle("spawn_failed"); child.settle("passed"); parent.settle("passed");
  session.close();
  createLaunchContext({ repoRoot, environment, unitID: "u" }).settle("passed");
  const index = {records:session.records().map(v=>v.record??v)};
  assert.equal(index.records.length, 4, "closed producers cannot publish late records");
  assert.equal(index.records.find((record) => record.activity === "migration").outcome, "failed");
  assert.equal(index.records.find((record) => record.activity === "fixture_reset").outcome, "incomplete");
  const recorded = index.records.find((record) => record.kind === "launch" && record.invocation_id === child.identity.invocation_id);
  assert.deepEqual(recorded.row_ids, ["a", "b"]);
  assert.equal(recorded.parent_invocation_id, parent.identity.invocation_id);
  assert.equal(recorded.outcome, "spawn_failed");
  assert.equal(recorded.launched, false);
  const stale = { ...environment, ...parent.environment, CARTULARY_TEST_RUN_ID: "other" };
  assert.throws(() => readLaunchContext(stale), /run mismatch/u);
  const fresh = createLaunchContext({ repoRoot, environment: stale });
  assert.equal(fresh.identity.parent_invocation_id, null);
  assert.equal(fresh.identity.run_id, "other");
  assert.notEqual(fresh.identity.invocation_id, parent.identity.invocation_id);
});

test("late attribution counts only complete jointly observed CPU intervals after registration", () => {
  const accumulator = createResourceAccumulator();
  const roster = new Map([["process:1", { unit_id: "a", allocation_ref: null, attribution_start_ms: 500 }]]);
  for (const [elapsed_ms, user, system] of [[0, 100, 20], [1000, 1000, 30], [2000, 1200, 40]]) {
    accumulator.consume({ scope: "process", scope_ref: "process:1", identity_digest: "proof", elapsed_ms, segment: 0, availability: "available",
      metrics: { cpu_user_us: { value: user, availability: "available" }, cpu_system_us: { value: system, availability: "available" } } }, roster);
  }
  assert.deepEqual([...accumulator.finish().unitCPU], [["a", 210]]);
});

const temporary = (t) => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-foundation-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  return root;
};
const graph = { units: [
  { unit_id: "a", kind: "runner", needs: [], resource_claims: { cpu: 1 } },
  { unit_id: "b", kind: "finalizer", needs: [], resource_claims: { cpu: 1 } },
] };
const facts = {
  starts: new Map([["a", 2], ["b", 5]]),
  terminals: new Map([["a", { event: "completed", monotonic_ms: 9, seq: 11 }], ["b", { event: "failed", monotonic_ms: 10, seq: 12 }]]),
  waits: new Map([["a", { count: 1, resource_intervals: [], resource_holders: new Set() }], ["b", { count: 1, resource_intervals: [{ start: 3, end: 5 }], resource_holders: new Set(["a"]) }]]),
  admissions: new Map([["a", { time: 2, seq: 4 }], ["b", { time: 5, seq: 9 }]]),
  grants: new Map(),
  fixtureAcquired: new Map([["a", 4]]), cleanupStarted: 10, finalMonotonicMs: 12,
};

test("independent overlapping accounting expectations include precedence and reservations", () => {
  assert.deepEqual(canonicalTimingAccounting(facts, graph, 12), {
    setup_ms: 2, fixture_ms: 2, execution_ms: 1, collation_ms: 5,
    wrapper_ms: 2, unattributed_ms: 0, resource_blocking_ms: 2, process_count: 2,
  });
  assert.deepEqual(canonicalTimingAccounting(facts, graph, 7, { selectedUnitIDs: ["a"], includeRunEnvelope: false }), {
    setup_ms: 0, fixture_ms: 2, execution_ms: 5, collation_ms: 0,
    wrapper_ms: 0, unattributed_ms: 0, resource_blocking_ms: 0, process_count: 1,
  });
  const pressure = projectResourcePressure(facts, graph, { cpu: 2, process: 2, io: 1, memory_mb: 1, postgres: 1, object_store: 1, port_lane: 1, browser_stack: 1, service_stack: 1, volume: 1 });
  assert.deepEqual(pressure.peak_use, { cpu: 2 });
  assert.deepEqual(pressure.saturation_ms, { cpu: 4 });
  // A worker that waits in host admission after graph admission holds its
  // physical claims only from its grant, here after a's terminal.
  const granted = projectResourcePressure({ ...facts,
    terminals: new Map([["a", { event: "completed", monotonic_ms: 9, seq: 11 }], ["b", { event: "failed", monotonic_ms: 10, seq: 13 }]]),
    grants: new Map([["b", { time: 9, seq: 12 }]]) }, graph, { cpu: 2 });
  assert.deepEqual(granted.peak_use, { cpu: 1 });
  assert.deepEqual(granted.saturation_ms, {});
  assert.equal(pressure.wait_events, 2);
  assert.deepEqual(pressure.blocked_units, ["b"]);
  assert.deepEqual(pressure.resource_holders, ["a"]);
});

test("canonical phases retain incomplete ends and reject the 129th start", async (t) => {
  const root = temporary(t), file = path.join(root, "events.ndjson");
  const entries = [{ event: "queued", unit_kind: "runner", row_ids: [], status: "pending" }];
  for (let i = 1; i <= 128; i++) entries.push({ event: "phase_started", interval_id: `phase:${i}`, phase: "runner" });
  const publish = () => writeFileSync(file, entries.map((entry, i) => JSON.stringify({
    schema_id: "cartulary.harness_unit_event.v5", seq: i + 1, monotonic_ms: i,
    unit_id: "a", status: "running", needs: [], resource_claims: {}, service_dependencies: [], ...entry,
  })).join("\n") + "\n");
  publish();
  const facts = await reduceCanonicalUnitIntervals(file);
  assert.equal(facts.phases.size, 128);
  assert.ok([...facts.phases.values()].every((phase) => phase.end_ms === null));
  entries.push({ event: "phase_started", interval_id: "phase:129", phase: "runner" });
  publish();
  await assert.rejects(reduceCanonicalUnitIntervals(file), /phase start/u);
});

test("sample acceptance preserves nullable availability and closed vocabulary", () => {
  const sample = { schema_id: "cartulary.harness_resource_sample.v2", seq: 1, elapsed_ms: 0,
    segment: 0, scope: "process", scope_ref: "process:1", identity_digest: null,
    availability: "available", metrics: {} };
  for (const availability of ["available", "unsupported", "permission_denied", "not_observed", "process_gone", "counter_reset", "paused_measurement", "truncated", "collector_failed"]) {
    const value = { ...sample, metrics: { cpu_user_us: { value: availability === "available" ? 0 : null, availability } } };
    validateSchemaSync(sample.schema_id, value);
    value.metrics.cpu_user_us.value = availability === "available" ? null : 0;
    assert.throws(() => validateSchemaSync(sample.schema_id, value));
  }
  for (const invalid of [{ ...sample, unknown: 1 }, { ...sample, metrics: { mystery: { value: 0, availability: "available" } } }, { ...sample, scope: "container" }]) {
    assert.throws(() => validateSchemaSync(sample.schema_id, invalid));
  }
});

test("exact selection never chooses a newest sibling", (t) => {
  const root = temporary(t);
  assert.throws(() => resolveExactRunDir(root), /RUN_ID/u);
  assert.throws(() => resolveExactRunDir(root, "missing"), /RUN_ID/u);
});

test("explicit export validates endpoints, private headers and transport failures", async (t) => {
  assert.equal(validatedEndpoint("http://127.0.0.1:4318").protocol, "http:");
  for (const endpoint of ["http://example.com", "https://user:secret@example.com", "https://example.com/?secret=x"]) assert.throws(() => validatedEndpoint(endpoint));
  const file = path.join(temporary(t), "headers.json");
  writeFileSync(file, JSON.stringify({ authorization: "test-only" }), { mode: 0o600 });
  assert.deepEqual(headersFromFile(file), { authorization: "test-only" });
  writeFileSync(file, '{"authorization":"first","authorization":"second"}');
  assert.throws(() => headersFromFile(file), /duplicate/u);
  writeFileSync(file, JSON.stringify({ host: "forbidden" }));
  assert.throws(() => headersFromFile(file));
  let requests = 0;
  await assert.rejects(deliver(new URL("https://example.com"), {}, {}, 100, async (_url, options) => {
    requests++;
    assert.equal(options.redirect, "error");
    assert.ok(options.signal instanceof AbortSignal);
    return { redirected: false, status: 302, ok: false };
  }), /redirect/u);
  assert.equal(requests, 1);
});

test("generated definitions preserve acceptance and admit isolated ordinary extensions", async () => {
  const catalog = JSON.parse(readFileSync(new URL("../../../harness_instrumentation_definitions.json", import.meta.url), "utf8"));
  const lease = JSON.parse(readFileSync(new URL("../../../schemas/cartulary.harness_fixture_lease.v5.schema.json", import.meta.url), "utf8"));
  const ajv = new Ajv2020({ strict: false, validateFormats: false });
  ajv.addSchema(lease);
  const original = renderInstrumentationDefinitions(catalog);
  ajv.addSchema(original.schema);
  const value = ajv.compile({ $ref: `${original.schema.$id}#/$defs/sample_metrics` });
  for (const signal of catalog.signals) {
    assert.equal(value({ [signal.name]: { availability: "available", value: 0 } }), true);
    assert.equal(value({ [signal.name]: { availability: "not_observed", value: null } }), true);
    assert.equal(value({ [signal.name]: { availability: "available", value: null } }), false);
  }
  assert.equal(value({ synthetic_counter: { availability: "available", value: 0 } }), false);
  catalog.phases.push("synthetic_phase");
  catalog.signals.push({ name: "synthetic_counter", kind: "counter", unit: "count", producer_scope: "process", aggregation: "segmented_delta" },
    { name: "synthetic_gauge", kind: "gauge", unit: "count", producer_scope: "process", aggregation: "sampled_maximum" });
  ajv.removeSchema(original.schema.$id);
  const extended = renderInstrumentationDefinitions(catalog);
  ajv.addSchema(extended.schema);
  const extendedValue = ajv.compile({ $ref: `${extended.schema.$id}#/$defs/sample_metrics` });
  assert.equal(extendedValue({ synthetic_counter: { availability: "available", value: 2 }, synthetic_gauge: { availability: "available", value: 3 } }), true);
  assert.ok(extended.schema.$defs.phase.enum.includes("synthetic_phase"));
  const descriptors = await import(`data:text/javascript;base64,${Buffer.from(extended.runtime).toString("base64")}`);
  const accumulator = createResourceAccumulator({ signals: descriptors.signals });
  for (const [elapsed_ms, counter, gauge] of [[0, 2, 12], [10, 7, 8]]) accumulator.consume({
    scope_ref: "guest:synthetic", scope: "guest_context", availability: "available", segment: 0, identity_digest: null, elapsed_ms,
    metrics: { synthetic_counter: { value: counter, availability: "available" }, synthetic_gauge: { value: gauge, availability: "available" } },
  }, new Map());
  const metrics = accumulator.finish().rows[0].metrics;
  assert.deepEqual(metrics.synthetic_counter, { maximum: 7, observed_delta: 5, observed_interval_ms: 10, unavailable: 0 });
  assert.deepEqual(metrics.synthetic_gauge, { maximum: 12, observed_delta: null, observed_interval_ms: 0, unavailable: 0 });

  const duplicate = structuredClone(catalog); duplicate.signals.push(duplicate.signals[0]);
  assert.throws(() => renderInstrumentationDefinitions(duplicate), /duplicate/u);
});

// TH-HARNESS-AC-160: the script redactor yields the same independently
// specified output as the native redactor (scan_linux_test.go) for every
// manifest rule and structured secret-key class in every scan context.
function redactionMatrix() {
  const secret = "fixture-secret-1234";
  const cases = [
    ["rule/bearer_token", false, `token type Bearer ${secret} end`, "token type Bearer <redacted> end", secret],
    ["rule/basic_auth", false, "scheme Basic Zml4dHVyZS1zZWNyZXQtMTIzNA== end", "scheme Basic <redacted> end", "Zml4dHVyZS1zZWNyZXQtMTIzNA=="],
    ["rule/authorization_header", false, `Authorization: Digest ${secret}`, "Authorization: <redacted>", secret],
    ["rule/authorization_bearer_header", false, `Authorization: Bearer ${secret}`, "Authorization: <redacted>", secret],
    ["rule/cookie_header", false, `Set-Cookie: sid=${secret}; Path=/`, "Set-Cookie: <redacted>", secret],
    ["rule/route_token_header", false, `X-Cartulary-Test-Route-Token: ${secret}`, "X-Cartulary-Test-Route-Token: <redacted>", secret],
    ["rule/jwt", false, "jwt eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.fixturesignature1234 end", "jwt <redacted> end", "fixturesignature1234"],
    ["rule/generic_url_userinfo", false, `https://user:${secret}@example.test/path`, "https://<redacted>@example.test/path", secret],
    ["rule/postgres_url_userinfo_query", false, "postgres://u:fixture?secret-1234@h/db", "postgres://u:<redacted>@h/db", "fixture?secret-1234"],
    ["rule/postgres_url_userinfo_fragment", false, "postgres://u:fixture#secret-1234@h/db", "postgres://u:<redacted>@h/db", "fixture#secret-1234"],
    ["rule/cli_secret_argument_equals", false, `--client-secret=${secret}`, "--client-secret=<redacted>", secret],
    ["rule/cli_secret_argument_space", false, `--token ${secret}`, "--token <redacted>", secret],
    ["rule/secret_assignment", false, `DB_PASSWORD=${secret}`, "DB_PASSWORD=<redacted>", secret],
    ["rule/json_like_secret_assignment", false, `{"pwd":"${secret}"}`, '{"pwd":"<redacted>"}', secret],
    ["rule/private_key_block", false, `-----BEGIN PRIVATE KEY-----\n${secret}\n-----END PRIVATE KEY-----`, "<redacted>", secret],
    ["rule/json_member_closure", false, `{"database_dsn":"${secret}"}`, '{"database_dsn":"<redacted>"}', secret],
    ["rule/json_member_closure_url", false, `{"database_dsn":"postgres://u:${secret}@h"}`, '{"database_dsn":"<redacted>"}', secret],
    ["assignment/PASSWORD", false, `PASSWORD=${secret}`, "PASSWORD=<redacted>", secret],
    ["assignment/ACCESS_TOKEN", false, `ACCESS_TOKEN=${secret}`, "ACCESS_TOKEN=<redacted>", secret],
    ["assignment/CARTULARY_S3_ACCESS_KEY", false, `CARTULARY_S3_ACCESS_KEY=${secret}`, "CARTULARY_S3_ACCESS_KEY=<redacted>", secret],
    ["assignment/AWS_ACCESS_KEY_ID", false, `AWS_ACCESS_KEY_ID=${secret}`, "AWS_ACCESS_KEY_ID=<redacted>", secret],
    ["assignment/SESSION_SECRET", false, `SESSION_SECRET=${secret}`, "SESSION_SECRET=<redacted>", secret],
    ["assignment/TLS_PRIVATE_KEY", false, `TLS_PRIVATE_KEY=${secret}`, "TLS_PRIVATE_KEY=<redacted>", secret],
    ["assignment/COOKIE", false, `COOKIE=${secret}`, "COOKIE=<redacted>", secret],
    ["assignment/CARTULARY_DSN", false, `CARTULARY_DSN=postgres://u:${secret}@h/db`, "CARTULARY_DSN=<redacted>", secret],
    ["assignment/export_prefix", false, `export DB_PASSWORD=${secret}`, "export DB_PASSWORD=<redacted>", secret],
    ["assignment/lowercase_dsn_keyword", false, `host=db password=${secret} dbname=cartulary`, "host=db password=<redacted> dbname=cartulary", secret],
    ["assignment/spaced", false, `client_secret = ${secret}`, "client_secret=<redacted>", secret],
    ["url/postgres", false, `postgres://cartulary:${secret}@127.0.0.1:5432/db`, "postgres://<redacted>@127.0.0.1:5432/db", secret],
    ["url/postgresql_query", false, `postgresql://u:${secret}@h/db?sslmode=disable`, "postgresql://<redacted>@h/db?sslmode=disable", secret],
    ["url/postgres_query_at", false, `postgres://u:${secret}@h:5432?user=a@b`, "postgres://<redacted>@h:<redacted>@b", secret],
    ["assignment/SESSION", false, `SESSION=${secret}`, "SESSION=<redacted>", secret],
    ["assignment/APP_SESSION", false, `APP_SESSION=${secret}`, "APP_SESSION=<redacted>", secret],
    ["assignment/export_session", false, `export CARTULARY_SESSION=${secret}`, "export CARTULARY_SESSION=<redacted>", secret],
    ["assignment/session_query", false, `https://example.test/cb?session=${secret}`, "https://example.test/cb?session=<redacted>", secret],
    ["nbsp/assignment", false, "DB_PASSWORD=fixture\u00a0secret-1234", "DB_PASSWORD=<redacted>", "secret-1234"],
    ["nbsp/cli_secret_argument_equals", false, "--token=fixture\u00a0secret-1234", "--token=<redacted>", "secret-1234"],
    ["preserve/SESSION_ID", false, "SESSION_ID=abc123", "SESSION_ID=abc123", ""],
    ["preserve/raw_session_member", false, '{"session":"browser-session-name"}', '{"session":"browser-session-name"}', ""],
    ["preserve/TOKENS", false, "TOKENS=ordinary", "TOKENS=ordinary", ""],
    ["preserve/max_tokens", false, "max_tokens=5", "max_tokens=5", ""],
    ["preserve/url_at_outside_authority", false, "https://example.test/path@item?q=a@b#c@d", "https://example.test/path@item?q=a@b#c@d", ""],
    ["preserve/json_tokens", false, '{"tokens":"ordinary"}', '{"tokens":"ordinary"}', ""],
  ];
  for (const key of [
    "password", "Password", "PASS", "pwd", "secret", "SECRET", "client_secret", "session_secret", "secret_ref",
    "dsn", "DSN", "database_dsn", "recovery_dsn", "dsn_ref", "private_key", "private_key_path", "api_key",
    "access_key", "secret_key", "jwt", "bearer", "Authorization", "Cookie", "Set-Cookie", "X-Cartulary-Test-Route-Token",
    "token", "session_token", "cookies", "tokenizer", "passwordless", "CARTULARY_S3TEST_ACCESS_KEY_ID",
    "dbPassword", "clientSecret", "secretKey", "accessToken", "refreshToken", "sessionCookie",
  ]) cases.push([`member/${key}`, true, `{"${key}":"${secret}"}`, `{"${key}":"<redacted>"}\n`, secret]);
  cases.push(
    ["member/dsn_url", true, `{"dsn":"postgres://u:${secret}@h/db"}`, '{"dsn":"<redacted>"}\n', secret],
    ["member/database_dsn_url", true, `{"database_dsn":"postgres://u:${secret}@h/db"}`, '{"database_dsn":"<redacted>"}\n', secret],
    ["member/cookies_assignment", true, `{"cookies":"session=${secret}"}`, '{"cookies":"<redacted>"}\n', secret],
    ["member/session_secret_bearer", true, `{"session_secret":"Bearer ${secret}"}`, '{"session_secret":"<redacted>"}\n', secret],
    ["member/nested", true, `{"nested":{"SECRET_KEY":"${secret}","api_key":"${secret}"}}`, '{"nested":{"SECRET_KEY":"<redacted>","api_key":"<redacted>"}}\n', secret],
    ["member/message_url", true, `{"message":"postgres://u:${secret}@h/db"}`, '{"message":"postgres://<redacted>@h/db"}\n', secret],
    ["member/message_assignment", true, `{"message":"run DB_PASSWORD=${secret} now"}`, '{"message":"run DB_PASSWORD=<redacted> now"}\n', secret],
    ["member/whole_string_assignment", true, `{"env":["DB_PASSWORD=fixture secret 1234","TOKEN_FILE=/run/${secret}","PLAIN=value"]}`, '{"env":["DB_PASSWORD=<redacted>","TOKEN_FILE=<redacted>","PLAIN=value"]}\n', secret],
    ["member/cli_arguments", true, `{"args":["--token","${secret}","--dsn=postgres://u:${secret}@h/db"]}`, '{"args":["--token","<redacted>","--dsn=<redacted>"]}\n', secret],
    ["member/message_session", true, `{"message":"SESSION=${secret}"}`, '{"message":"SESSION=<redacted>"}\n', secret],
    ["member/env_session", true, `{"env":["SESSION=${secret} tail"]}`, '{"env":["SESSION=<redacted> tail"]}\n', secret],
    ["member/long_s_secret", true, `{"\u017fecret":"${secret}"}`, '{"\u017fecret":"<redacted>"}\n', secret],
    ["member/non_ascii_prefix_secret", true, `{"t\u00f6kenSecret":"${secret}"}`, '{"t\u00f6kenSecret":"<redacted>"}\n', secret],
    ["preserve/session_member", true, '{"session":"browser-session-name"}', '{"session":"browser-session-name"}\n', ""],
    ["preserve/sharp_s_key", true, '{"pa\u00dfword":"fixture-plain"}', '{"pa\u00dfword":"fixture-plain"}\n', ""],
    ["preserve/session_target", true, '{"session_target":"browser-stage-token-name"}', '{"session_target":"browser-stage-token-name"}\n', ""],
    ["preserve/cpu_tokens", true, '{"cpu_tokens":"2"}', '{"cpu_tokens":"2"}\n', ""],
    ["preserve/closure_container", true, '{"tokenizer":{"kind":"bpe"}}', '{"tokenizer":{"kind":"bpe"}}\n', ""],
    ["preserve/closure_scalar", true, '{"cookies":3,"tokenizer_count":2}', '{"cookies":3,"tokenizer_count":2}\n', ""],
    ["preserve/service_sessions", true, '{"service_sessions":[{"cleanup_status":"pass","setup_duration_ms":12}]}', '{"service_sessions":[{"cleanup_status":"pass","setup_duration_ms":12}]}\n', ""],
  );
  return cases;
}

// TH-HARNESS-REQ-617: every redactor is linear in its input. Adversarial runs
// of 4 MiB finish in milliseconds when linear; a quadratic redactor cannot meet
// the bound. A synchronous regular-expression hang cannot be interrupted in
// process, so the work runs in a child with its own deadline.
test("script redaction completes in linear time on adversarial inputs", () => {
  const child = `
const { compactJSONString, redactString } = await import(process.env.CARTULARY_REDACTION_CONTRACT_URL);
const size = 4 * 1024 * 1024;
const fill = (unit) => unit.repeat(Math.ceil(size / unit.length)).slice(0, size);
const runs = {
  word_without_assignment: fill("Password_Token_Session_9"),
  unquoted_terms: fill("passwordsecrettokencookiedsn"),
  scheme_without_separator: fill("ab0+.-"),
  token_without_dot: fill("eyJ-_aZ9"),
};
const cases = [];
for (const [name, run] of Object.entries(runs)) {
  cases.push([\`raw/\${name}\`, () => redactString(run)]);
  cases.push([\`unterminated_quote/\${name}\`, () => redactString(\`"\${run}\`)]);
  cases.push([\`json_value/\${name}\`, () => redactString(\`{"message":"\${run}","next":"x"}\`)]);
  cases.push([\`json_key/\${name}\`, () => redactString(\`{"\${run}":"value"}\`)]);
  cases.push([\`structured/\${name}\`, () => compactJSONString({ [run]: "value", message: run, args: [run] })]);
}
const timings = {};
for (const [name, redact] of cases) {
  const started = performance.now();
  redact();
  timings[name] = Math.round(performance.now() - started);
}
process.stdout.write(JSON.stringify(timings));
`;
  const result = spawnSync(process.execPath, ["--input-type=module", "-e", child], {
    encoding: "utf8",
    timeout: 120_000,
    env: { ...process.env, CARTULARY_REDACTION_CONTRACT_URL: pathToFileURL(path.resolve(import.meta.dirname, "../../contract/index.mjs")).href },
  });
  assert.equal(result.error?.code, undefined, "adversarial redaction exceeded its 120 s deadline and was stopped");
  assert.equal(result.status, 0, "adversarial redaction child failed");
  const slow = Object.entries(JSON.parse(result.stdout)).filter(([, milliseconds]) => milliseconds >= 5000);
  assert.deepEqual(slow, [], "adversarial redaction exceeded 5 s for a 4 MiB input");
});

test("script redaction matrix matches the independently specified expectations", () => {
  // Report case names only: a failing output may still hold its fixture
  // secret, and retained logs must pass the retained-secret scan.
  const failures = [];
  const cases = redactionMatrix();
  for (const [name, structured, input, want, secret] of cases) {
    const got = structured ? compactJSONString(JSON.parse(input)) : redactString(input);
    if (got !== want) failures.push(`${name}:differs`);
    if (secret && got.includes(secret)) failures.push(`${name}:leaks`);
    if (/REDACTED[\]:]/u.test(got)) failures.push(`${name}:marker`);
  }
  assert.deepEqual(failures, [], `${failures.length} findings over ${cases.length} cases`);
});

test("native redaction generation renders the frozen scan identity and rejects partial, ambiguous or uncovered policy", () => {
  const policy = JSON.parse(readFileSync(path.resolve(import.meta.dirname, "../../../harness_redaction_manifest.json"), "utf8"));
  const rendered = renderNativeRedactionPolicy(policy);
  // The manifest owns the single marker; the native policy exports it (TH-HARNESS-REQ-617).
  assert.match(rendered, /^const Marker = "<redacted>"$/mu);
  assert.match(rendered, /replacement: Marker,/u);
  // The generated scan identity reproduces the frozen policy digest.
  const identity = JSON.parse(/^const RetainedScanPolicyIdentity = (".*")$/mu.exec(rendered)[1]);
  assert.equal(`sha256:${createHash("sha256").update(identity).digest("hex")}`,
    "sha256:eaa74efae69ea73a6e0f18823a94ccc70a19b242e2cb2880e446611c3084ee11");
  for (const mutate of [
    () => null,
    (p) => ({ ...p, schema_id: "foreign" }),
    (p) => ({ ...p, replacement: "" }),
    (p) => ({ ...p, value_patterns: [] }),
    (p) => ({ ...p, value_patterns: [...p.value_patterns, p.value_patterns[0]] }),
    (p) => { delete p.value_patterns[0].replacement; return p; },
    (p) => { p.value_patterns[0].pattern = "["; return p; },
    // A rule must emit the marker, and only the marker: no second or typed form.
    (p) => { p.value_patterns[0].replacement = "$1"; return p; },
    (p) => { p.value_patterns[0].replacement = "$1[REDACTED]"; return p; },
    (p) => { p.value_patterns[0].replacement = "$1<redacted>:dsn"; return p; },
    (p) => { p.value_patterns[0].replacement = "$1<redacted> [REDACTED:token]"; return p; },
    // The structured and CLI vocabularies are required, closed and canonical.
    (p) => { delete p.structured_secret_key_tokens; return p; },
    (p) => { p.structured_secret_key_tokens.push("password"); return p; },
    (p) => { p.secret_cli_flag_names = []; return p; },
    (p) => { p.whole_string_assignment = "([A-Za-z]+)="; return p; },
    // A scan vocabulary change cannot keep the frozen policy identity.
    (p) => { p.retained_scan.json_member_terms.push({ term: "credential" }); return p; },
    (p) => { p.retained_scan.assignment_names[2] = { name: "TOKEN" }; return p; },
    (p) => { p.retained_scan.url_userinfo_schemes.push("mysql"); return p; },
    (p) => { p.retained_scan.forbidden_names.pop(); return p; },
    // The authored rules must cover the scan and CLI vocabularies.
    (p) => { p.value_patterns = p.value_patterns.filter((rule) => rule.name !== "secret_assignment"); return p; },
    (p) => { p.value_patterns = p.value_patterns.filter((rule) => rule.name !== "postgres_url_userinfo"); return p; },
    (p) => { p.secret_cli_flag_names.push("credential"); return p; },
    // Shared patterns use explicit ASCII classes, never engine-dependent ones.
    ...["\\s", "\\S", "\\d", "\\D", "\\w", "\\W", "\\b", "\\B", "."].map((token) =>
      (p) => { p.value_patterns[0].pattern = `(?i)Bearer${token}[A-Za-z]+`; return p; }),
    (p) => { p.whole_string_assignment = "^([A-Za-z]\\w*)="; return p; },
  ]) assert.throws(() => renderNativeRedactionPolicy(mutate(structuredClone(policy))));
  // A shared pattern begins with a literal or with a consumed ASCII boundary
  // group that its replacement re-emits, so it stays linear (TH-HARNESS-REQ-617).
  const rule = (p, name) => p.value_patterns.find((entry) => entry.name === name);
  for (const [mutate, message] of [
    [(p) => { rule(p, "secret_assignment").pattern = String.raw`(?i)([A-Za-z0-9_]*?(?:PASSWORD|SECRET))=[^\t\n\f\r ]+`; }, /must begin with a literal or a consumed boundary/u],
    [(p) => { rule(p, "bearer_token").pattern = String.raw`(?i)B+earer[\t ]+[A-Za-z]+`; }, /must begin with a literal/u],
    [(p) => { rule(p, "cookie_header").pattern = String.raw`(?i)((?:Cookie|[A-Z]+-Cookie)[\t ]*:[\t ]*)([^\r\n]+)`; }, /must begin with a literal/u],
    [(p) => { rule(p, "jwt").replacement = "<redacted>"; }, /re-emit its consumed boundary/u],
  ]) {
    const copy = structuredClone(policy);
    mutate(copy);
    assert.throws(() => renderNativeRedactionPolicy(copy), message);
  }
});

test("foundation generation compiles before publication and rolls back partial installation", async (t) => {
  const { chmodSync, cpSync, linkSync, mkdirSync, renameSync, statSync, symlinkSync } = await import("node:fs");
  const { generateFoundationArtifacts } = await import("../../generated-artifacts/generate-foundation-schema-validators.mjs");
  const { normalizeGeneratedSourceModes, publishGeneratedTransaction } = await import("../../generated-artifacts/generated-transaction.mjs");
  const root = temporary(t), repo = path.resolve(import.meta.dirname, "../../../..");
  for (const relative of ["tools/schemas", "tools/harness_schema_attachments.json", "tools/harness_instrumentation_definitions.json", "tools/harness_lifecycle_policy.json", "tools/harness_redaction_manifest.json", "tools/scheduler_resource_registry.json", "tools/task_surface_owner.json", "internal/testutil/harnessdiag/testdata/schema-probes.json"]) {
    mkdirSync(path.dirname(path.join(root, relative)), { recursive: true });
    cpSync(path.join(repo, relative), path.join(root, relative), { recursive: true });
  }
  const priorMask = process.umask(0o077);
  try { generateFoundationArtifacts(root); }
  finally { process.umask(priorMask); }
  const shared = path.join(root, "tools/schemas/cartulary.harness_instrumentation_defs.v2.schema.json");
  assert.equal(statSync(shared).mode & 0o777, 0o644, "generated source modes are independent of the caller's umask");
  const inode = statSync(shared).ino;
  chmodSync(shared, 0o600);
  normalizeGeneratedSourceModes(root, [path.relative(root, shared)]);
  assert.equal(statSync(shared).mode & 0o777, 0o644);
  assert.equal(statSync(shared).ino, inode, "mode normalization preserves the source inode");
  const privateEvidence = path.join(root, "private-evidence.json");
  writeFileSync(privateEvidence, "private", { mode: 0o600 });
  normalizeGeneratedSourceModes(root, [path.relative(root, shared)]);
  assert.equal(statSync(privateEvidence).mode & 0o777, 0o600, "only declared source receives source permissions");
  symlinkSync(privateEvidence, path.join(root, "linked-source.json"));
  assert.throws(() => normalizeGeneratedSourceModes(root, ["linked-source.json"]), /symbolic link/u);
  symlinkSync(root, path.join(root, "linked-parent"));
  assert.throws(() => normalizeGeneratedSourceModes(root, ["linked-parent/private-evidence.json"]), /ancestor/u);
  linkSync(privateEvidence, path.join(root, "hardlinked-source.json"));
  assert.throws(() => normalizeGeneratedSourceModes(root, ["hardlinked-source.json"]), /unsafe generated source inode/u);
  assert.equal(statSync(privateEvidence).mode & 0o777, 0o600);
  const humanRoot = ["do", "cs"].join("");
  assert.throws(() => normalizeGeneratedSourceModes(root, [`${humanRoot}/never-probe.json`]), /unsafe generated path/u);
  const before = readFileSync(shared);
  const nativePolicy = path.join(root, "internal/platform/harnessredact/policy_generated.go");
  const policyBefore = readFileSync(nativePolicy);
  const policyFile = path.join(root, "tools/harness_redaction_manifest.json");
  const policy = JSON.parse(readFileSync(policyFile));
  writeFileSync(policyFile, JSON.stringify({ ...policy, value_patterns: [] }));
  assert.throws(() => generateFoundationArtifacts(root), /invalid native redaction policy/u);
  assert.deepEqual(readFileSync(nativePolicy), policyBefore);
  assert.deepEqual(readFileSync(shared), before);
  writeFileSync(policyFile, JSON.stringify(policy));
  const catalogFile = path.join(root, "tools/harness_instrumentation_definitions.json");
  const catalog = JSON.parse(readFileSync(catalogFile)); catalog.phases.push("synthetic_phase"); writeFileSync(catalogFile, JSON.stringify(catalog));
  // Catalog is valid, but a dependent foundation schema cannot compile.
  writeFileSync(path.join(root, "tools/schemas/cartulary.harness_resource_sample.v2.schema.json"), "{");
  assert.throws(() => generateFoundationArtifacts(root));
  assert.deepEqual(readFileSync(shared), before);
  const rendered = temporary(t), relative = "tools/transaction-fixture.json";
  writeFileSync(path.join(root, relative), "old"); mkdirSync(path.join(rendered, "tools")); writeFileSync(path.join(rendered, relative), "new");
  assert.throws(() => publishGeneratedTransaction({ repoRoot: root, renderedRoot: rendered, generatedPaths: [relative],
    rename: (from, to) => { if (from.includes(`${path.sep}staging${path.sep}`)) throw new Error("install failed"); renameSync(from, to); },
  }), /install failed/u);
  assert.equal(readFileSync(path.join(root, relative), "utf8"), "old");
});
