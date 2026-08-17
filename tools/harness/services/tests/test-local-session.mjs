import assert from "node:assert/strict";
import { chmodSync, mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";

import {
  attachLocalSession,
  createLocalSession,
  localSessionStatus,
  readLocalSessionDescriptor,
  resolveLocalSessionFile,
  resolveServiceSessionMode,
  stopLocalSession,
} from "../local-session.mjs";

const repositoryRoot = path.resolve(import.meta.dirname, "../../../..");
const taskSurface = JSON.parse(
  readFileSync(path.join(repositoryRoot, "tools/task_surface_owner.json"), "utf8"),
);
const postgresID = "a".repeat(64);
const objectStoreID = "b".repeat(64);
const postgresImage = "postgres:17.5-alpine";
const objectStoreImage = "chrislusf/seaweedfs:3.89";

function privateDirectory(directory) {
  mkdirSync(directory, { recursive: true, mode: 0o700 });
  chmodSync(directory, 0o700);
}

function privateJSON(file, value) {
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`, { encoding: "utf8", mode: 0o600 });
  chmodSync(file, 0o600);
}

function sessionFixture({ ageMS = 0 } = {}) {
  const temporary = mkdtempSync(path.join(tmpdir(), "cartulary-local-session-test-"));
  chmodSync(temporary, 0o700);
  const binary = path.join(temporary, "testservices");
  writeFileSync(binary, "test-services fixture\n", { mode: 0o700 });
  chmodSync(binary, 0o700);
  const sessionFile = path.join(temporary, "cache", "test-services", "session.json");
  const now = new Date(Date.now() + ageMS);
  const containers = new Map();
  let terminateCount = 0;
  const dependencies = {
    now: () => now,
    sessionID: () => "0123456789abcdef01234567",
    testservicesOutput(args) {
      if (args[0] === "schema-hash") return "c".repeat(64);
      if (args[0] === "images") return `${postgresImage}\n${objectStoreImage}\n`;
      throw new Error(`unexpected test-services output command: ${args.join(" ")}`);
    },
    inspectImage(reference) {
      if (reference === postgresImage) return { Id: `sha256:${"d".repeat(64)}` };
      if (reference === objectStoreImage) return { Id: `sha256:${"e".repeat(64)}` };
      throw new Error(`unexpected image reference: ${reference}`);
    },
    inspectContainer(identity) {
      const value = containers.get(identity);
      if (!value) throw new Error("container missing");
      return value;
    },
    runTestservices(args, { environment }) {
      if (args[0] === "terminate-suite") {
        assert.match(environment.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT, /suite-runtime/u);
        assert.match(
          environment.CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID,
          /^[0-9a-f-]{36}$/u,
        );
        assert.equal(environment.CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID, "session-up");
        terminateCount += 1;
        return;
      }
      assert.equal(args[0], "start-suite");
      const envFile = args[args.indexOf("--env-file") + 1];
      const leaseFile = args[args.indexOf("--lease-file") + 1];
      const labels = (service) => ({
        "cartulary.test-services.managed": "true",
        "cartulary.test-services.session-id": environment.CARTULARY_TEST_SUITE_ID,
        "cartulary.test-services.session-expires-at":
          environment.CARTULARY_TEST_SERVICES_SESSION_EXPIRES_AT,
        "cartulary.test-services.service": service,
      });
      const environmentValue = {
        CARTULARY_PGTEST_ADMIN_DSN: "postgres://admin:session-password@127.0.0.1:5432/postgres",
        CARTULARY_PGTEST_DSN_TEMPLATE: "postgres://app:session-password@127.0.0.1:5432/%s",
        CARTULARY_PGTEST_SCHEMA_HASH: "c".repeat(64),
        CARTULARY_PGTEST_TEMPLATE_DB: "cartulary_template_session",
        CARTULARY_S3TEST_ACCESS_KEY_ID: "session-access-key",
        CARTULARY_S3TEST_ENDPOINT: "http://127.0.0.1:8333",
        CARTULARY_S3TEST_PROBE_BUCKET: "session-probe",
        CARTULARY_S3TEST_SECRET_ACCESS_KEY: "session-secret-key",
        CARTULARY_S3TEST_SECURE: "false",
      };
      const resources = [
        {
          kind: "container",
          service: "postgres",
          container_id: postgresID,
          name: "cartulary-postgres-session",
          image: postgresImage,
          labels: labels("postgres"),
        },
        {
          kind: "container",
          service: "object_store",
          container_id: objectStoreID,
          name: "cartulary-object-store-session",
          image: objectStoreImage,
          labels: labels("object_store"),
        },
      ];
      containers.set(postgresID, {
        Id: postgresID,
        Image: `sha256:${"d".repeat(64)}`,
        Config: { Image: postgresImage, Labels: labels("postgres") },
        State: { Running: true },
      });
      containers.set(objectStoreID, {
        Id: objectStoreID,
        Image: `sha256:${"e".repeat(64)}`,
        Config: { Image: objectStoreImage, Labels: labels("object_store") },
        State: { Running: true },
      });
      privateJSON(envFile, environmentValue);
      privateJSON(leaseFile, {
        schema_id: "cartulary.test_services.lease.v1",
        lease_id: "session-service-lease",
        suite_id: environment.CARTULARY_TEST_SUITE_ID,
        run_id: environment.CARTULARY_TEST_RUN_ID,
        result_root: environment.CARTULARY_TEST_RESULTS_DIR,
        run_root: path.join(environment.CARTULARY_TEST_RESULTS_DIR, "session-up"),
        target: environment.CARTULARY_TEST_TARGET,
        mode: "owned",
        ownership_mode: "owned",
        owner_pid: process.pid,
        created_at: now.toISOString(),
        expires_at: environment.CARTULARY_TEST_SERVICES_SESSION_EXPIRES_AT,
        resources,
        proof_labels: { "cartulary.test-services.session-id": environment.CARTULARY_TEST_SUITE_ID },
        proof_prefixes: {},
        cleanup_state: "active",
      });
    },
  };
  return {
    binary,
    containers,
    dependencies,
    sessionFile,
    temporary,
    terminateCount: () => terminateCount,
    cleanup() {
      rmSync(temporary, { recursive: true, force: true });
    },
  };
}

function createFixtureSession(fixture) {
  return createLocalSession({
    root: repositoryRoot,
    binary: fixture.binary,
    sessionFile: fixture.sessionFile,
    environment: {},
    dependencies: fixture.dependencies,
  });
}

function rewriteDescriptor(fixture, mutate) {
  const value = JSON.parse(readFileSync(fixture.sessionFile, "utf8"));
  mutate(value);
  privateJSON(fixture.sessionFile, value);
}

test("session mode is explicit, closed, and limited to approved leaf targets", () => {
  assert.deepEqual(resolveServiceSessionMode({ target: "check", environment: {} }), {
    mode: "owned",
    sessionFile: null,
  });
  const sessionFile = path.join(tmpdir(), "cartulary-explicit-session.json");
  assert.deepEqual(
    resolveServiceSessionMode({
      target: "service-backed-test-slice",
      environment: {
        CARTULARY_TEST_SERVICES_MODE: "attach",
        CARTULARY_TEST_SERVICES_SESSION_FILE: sessionFile,
      },
    }),
    { mode: "attach", sessionFile },
  );
  for (const target of ["service-backed-test-slice", "browser-e2e-webserver-backed", "browser-e2e-stateful"]) {
    assert.equal(
      resolveServiceSessionMode({
        target,
        environment: {
          CARTULARY_TEST_SERVICES_MODE: "attach",
          CARTULARY_TEST_SERVICES_SESSION_FILE: sessionFile,
        },
      }).mode,
      "attach",
    );
  }
  assert.throws(
    () => resolveServiceSessionMode({ target: "check", environment: { CARTULARY_TEST_SERVICES_MODE: "attach", CARTULARY_TEST_SERVICES_SESSION_FILE: sessionFile } }),
    /owned-only/u,
  );
  assert.throws(
    () => resolveServiceSessionMode({ target: "test-fast", environment: { CARTULARY_TEST_SERVICES_ACTIVE: "1" } }),
    /reserved internal service state/u,
  );
  assert.throws(
    () =>
      resolveServiceSessionMode({
        target: "test-fast",
        environment: { CARTULARY_TEST_SERVICES_PERSISTENT_BORROWER: "1" },
      }),
    /reserved internal service state/u,
  );
  assert.throws(
    () => resolveServiceSessionMode({ target: "test-fast", environment: { CARTULARY_TEST_SERVICES_SESSION_FILE: sessionFile } }),
    /only in attach mode/u,
  );
  assert.throws(
    () => resolveServiceSessionMode({ target: "test-fast", environment: { CARTULARY_TEST_SERVICES_MODE: "legacy" } }),
    /owned or attach/u,
  );
  assert.throws(
    () => resolveLocalSessionFile({ CARTULARY_TEST_SERVICES_SESSION_FILE: "relative.json" }),
    /absolute/u,
  );

  const allowed = [
    "browser-e2e-stateful",
    "browser-e2e-webserver-backed",
    "service-backed-test-slice",
  ];
  const declared = taskSurface.targets
    .filter((entry) =>
      (entry.input_contract?.inputs ?? []).some(
        (input) => input.name === "CARTULARY_TEST_SERVICES_MODE",
      ),
    )
    .map((entry) => entry.name)
    .sort();
  assert.deepEqual(declared, allowed);
  for (const entry of taskSurface.targets.filter(
    (candidate) => candidate.target_class === "public" && !allowed.includes(candidate.name),
  )) {
    assert.throws(
      () =>
        resolveServiceSessionMode({
          target: entry.name,
          environment: {
            CARTULARY_TEST_SERVICES_MODE: "attach",
            CARTULARY_TEST_SERVICES_SESSION_FILE: sessionFile,
          },
        }),
      /owned-only/u,
    );
  }
});

test("session lifecycle is redacted, concurrent-borrower safe, and idempotent", () => {
  const fixture = sessionFixture();
  try {
    const created = createFixtureSession(fixture);
    assert.equal(created.state, "active");
    const descriptor = readLocalSessionDescriptor(fixture.sessionFile);
    assert.equal(descriptor.containers[0].service, "object_store");
    assert.equal(descriptor.containers[1].service, "postgres");
    assert.equal(readFileSync(fixture.sessionFile).length > 0, true);

    const runtime = (suffix) => ({
      root: path.join(fixture.temporary, `borrower-runtime-${suffix}`),
      leaseID: `00000000-0000-4000-8000-00000000000${suffix}`,
      registerEnvironment(environment) {
        assert.equal(environment.CARTULARY_TEST_SERVICES_CALL_MODE, "attach");
      },
    });
    const first = attachLocalSession({
      root: repositoryRoot,
      binary: fixture.binary,
      sessionFile: fixture.sessionFile,
      target: "service-backed-test-slice",
      runID: "borrower-one",
      suiteRuntime: runtime("1"),
      dependencies: { ...fixture.dependencies, borrowerID: () => "borrower-one" },
    });
    const second = attachLocalSession({
      root: repositoryRoot,
      binary: fixture.binary,
      sessionFile: fixture.sessionFile,
      target: "browser-e2e-stateful",
      runID: "borrower-two",
      suiteRuntime: runtime("2"),
      dependencies: { ...fixture.dependencies, borrowerID: () => "borrower-two" },
    });
    assert.notEqual(first.environment.CARTULARY_TEST_SUITE_ID, second.environment.CARTULARY_TEST_SUITE_ID);
    const status = localSessionStatus({
      root: repositoryRoot,
      binary: fixture.binary,
      sessionFile: fixture.sessionFile,
      dependencies: fixture.dependencies,
    });
    assert.equal(status.borrower_count, 2);
    const output = JSON.stringify(status);
    for (const forbidden of ["session-password", "session-secret-key", postgresID, objectStoreID, descriptor.state_root]) {
      assert.equal(output.includes(forbidden), false);
    }
    assert.throws(
      () => stopLocalSession({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile, dependencies: fixture.dependencies }),
      /live borrower/u,
    );
    first.close();
    second.close();
    assert.equal(
      localSessionStatus({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile, dependencies: fixture.dependencies }).borrower_count,
      0,
    );
    assert.equal(
      stopLocalSession({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile, dependencies: fixture.dependencies }).state,
      "absent",
    );
    assert.equal(fixture.terminateCount(), 1);
    assert.equal(
      stopLocalSession({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile, dependencies: fixture.dependencies }).state,
      "absent",
    );
    assert.equal(fixture.terminateCount(), 1);
  } finally {
    fixture.cleanup();
  }
});

test("status rejects malformed, permissive, wrong-owner, and symlink descriptors", () => {
  const fixture = sessionFixture();
  const directory = path.dirname(fixture.sessionFile);
  try {
    assert.equal(
      localSessionStatus({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile }).state,
      "absent",
    );
    privateDirectory(directory);
    writeFileSync(fixture.sessionFile, "{not-json\n", { mode: 0o600 });
    assert.equal(localSessionStatus({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile }).state, "invalid");
    chmodSync(fixture.sessionFile, 0o644);
    assert.equal(localSessionStatus({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile }).state, "invalid");
    rmSync(fixture.sessionFile);
    const target = path.join(fixture.temporary, "descriptor-target.json");
    privateJSON(target, {});
    symlinkSync(target, fixture.sessionFile);
    assert.equal(localSessionStatus({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile }).state, "invalid");
    rmSync(fixture.sessionFile);
    privateJSON(fixture.sessionFile, {});
    chmodSync(directory, 0o755);
    assert.equal(localSessionStatus({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile }).state, "invalid");
  } finally {
    chmodSync(directory, 0o700);
    fixture.cleanup();
  }
});

test("status detects owner, digest, expiry, and container invalidation", () => {
  const ownerFixture = sessionFixture();
  try {
    createFixtureSession(ownerFixture);
    rewriteDescriptor(ownerFixture, (descriptor) => {
      descriptor.owner_uid += 1;
    });
    assert.equal(localSessionStatus({ root: repositoryRoot, binary: ownerFixture.binary, sessionFile: ownerFixture.sessionFile }).state, "invalid");
  } finally {
    ownerFixture.cleanup();
  }

  const digestFixture = sessionFixture();
  try {
    createFixtureSession(digestFixture);
    rewriteDescriptor(digestFixture, (descriptor) => {
      descriptor.config_digest = `sha256:${"f".repeat(64)}`;
    });
    const status = localSessionStatus({ root: repositoryRoot, binary: digestFixture.binary, sessionFile: digestFixture.sessionFile, dependencies: digestFixture.dependencies });
    assert.equal(status.state, "stale");
    assert.equal(status.reason, "configuration_changed");
  } finally {
    digestFixture.cleanup();
  }

  const expiredFixture = sessionFixture({ ageMS: -25 * 60 * 60 * 1000 });
  try {
    createFixtureSession(expiredFixture);
    assert.equal(localSessionStatus({ root: repositoryRoot, binary: expiredFixture.binary, sessionFile: expiredFixture.sessionFile, dependencies: expiredFixture.dependencies }).state, "expired");
  } finally {
    expiredFixture.cleanup();
  }

  const containerFixture = sessionFixture();
  try {
    createFixtureSession(containerFixture);
    containerFixture.containers.get(objectStoreID).State.Running = false;
    const status = localSessionStatus({ root: repositoryRoot, binary: containerFixture.binary, sessionFile: containerFixture.sessionFile, dependencies: containerFixture.dependencies });
    assert.equal(status.state, "stale");
    assert.equal(status.reason, "container_not_running");
  } finally {
    containerFixture.cleanup();
  }
});
