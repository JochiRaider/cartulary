import assert from "node:assert/strict";
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";

import {
  attachLocalSession,
  detachLocalBorrower,
  createLocalSession,
  localSessionStatus,
  readLocalSessionDescriptor,
  resolveLocalSessionFile,
  resolveServiceSessionMode,
  stopLocalSession,
} from "../local-session.mjs";

import { acquireHostAdmission, hostClaims, processIdentity, transactAdmission } from "../../runtime/host-admission.mjs";
import { lifecyclePolicy } from "../../contract/generated/lifecycle-policy.mjs";

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
  const reservations = new Set();
  const dependencies = {
    async acquireAdmission({ residents }) {
      for (const allocation of residents) reservations.add(allocation.token);
      return { token: "a".repeat(32), async release() {} };
    },
    admission(request) {
      if (request.operation === "admit") { for (const allocation of request.residents) reservations.add(allocation.token); return true; }
      if (request.operation === "resident_granted") return reservations.has(request.token);
      if (request.operation === "release_resident") reservations.delete(request.token);
      return false;
    },
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
      const resultFile = args[args.indexOf("--result-file") + 1];
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
 CARTULARY_PGTEST_GENERATION: "12345",
 CARTULARY_PGTEST_ISOLATION: "shared",
 CARTULARY_PGTEST_ALLOCATION_REF: `session:${environment.CARTULARY_TEST_SUITE_ID}`,
 CARTULARY_PGTEST_PROVISIONING_DIGEST: "d".repeat(64),
 CARTULARY_PGTEST_TEMPLATE_DIGEST: "e".repeat(64),
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
        schema_id: "cartulary.test_services.lease.v3",
 postgres_provisioning: {schema_id:"cartulary.test_services.postgres_provisioning.v2", generation:"12345",postgres_isolation:"shared",allocation_ref:`session:${environment.CARTULARY_TEST_SUITE_ID}`,policy_digest:"d".repeat(64),template_digest:"e".repeat(64)},
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
      privateJSON(resultFile, {
        schema_id: "cartulary.test_services.start_result.v2",
        recovery_files: [],
        status: "ready",
        run_id: "session-up",
        target: "test-services-session-up",
        suite_id: environment.CARTULARY_TEST_SUITE_ID,
        service_scope_ref: `_shared/test-services/${environment.CARTULARY_TEST_SUITE_ID}/service-scope.json`,
        failure_class: null,
        failure_reason: null,
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

test("local lifecycle holds its lock and publishes no service until queued admission succeeds", async () => {
  const fixture = sessionFixture();
  const acquire = fixture.dependencies.acquireAdmission;
  const gate = Promise.withResolvers();
  fixture.dependencies.acquireAdmission = async (options) => {
    assert.equal(options.timeoutMs, lifecyclePolicy.limits.preparation);
    assert.equal(options.residents[0].parent, null);
    await gate.promise;
    return acquire(options);
  };
  const creating = createFixtureSession(fixture);
  try {
    assert.equal(fixture.containers.size, 0);
    assert.equal(existsSync(fixture.sessionFile), false);
    await assert.rejects(createFixtureSession(fixture), /lifecycle is already active/u);
    gate.resolve();
    assert.equal((await creating).state, "active");
    const stopping = Promise.withResolvers();
    fixture.dependencies.acquireAdmission = async (options) => {
      assert.equal(options.mode, "drain");
      assert.equal(options.timeoutMs, lifecyclePolicy.limits.drain);
      await stopping.promise;
      return acquire(options);
    };
    const closed = stopLocalSession({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile, dependencies: fixture.dependencies });
    assert.equal(fixture.terminateCount(), 0);
    assert.equal(existsSync(fixture.sessionFile), true);
    stopping.resolve();
    assert.equal((await closed).state, "absent");
    assert.equal(fixture.terminateCount(), 1);
  } finally { gate.resolve(); await creating.catch(() => {}); fixture.cleanup(); }
});

test("nested persistent startup shares its ancestor envelope and retains residency after the ancestor exits", async () => {
  const fixture = sessionFixture();
  const root = path.join(fixture.temporary, "admission");
  const capacities = hostClaims({ cpu: 2, io: 4, memory_mb: 1536, process: 32, object_store: 1, postgres: 1, port_lane: 2, volume: 1 });
  const leases = [];
  const pendingToken = "c".repeat(32), contenderToken = "d".repeat(32);
  let resident;
  const transaction = request => transactAdmission({ ...request, root });
  try {
    const parent = await acquireHostAdmission({ root, capacities, claims: { cpu: 1, memory_mb: 768, process: 20 } }); leases.push(parent);
    const sibling = await acquireHostAdmission({ root, capacities, claims: { cpu: 1, memory_mb: 768, process: 12 } }); leases.push(sibling);
    fixture.dependencies.admission = transaction;
    fixture.dependencies.acquireAdmission = async options => {
      assert.equal(options.parent, parent.token);
      resident = { ...options.residents[0], capacities };
      assert.equal(resident.parent, parent.token);
      if (options.mode === "shared") {
        // The old root-resident shape cannot fit while the enclosing test and
        // an unrelated sibling hold their otherwise sufficient reservations.
        assert.equal(transaction({ operation: "admit", token: pendingToken, owner: processIdentity(), mode: "shared", parent: parent.token,
          claims: hostClaims(options.claims), capacities, residents: [{ ...resident, parent: null }] }), false);
        transaction({ operation: "release", token: pendingToken });
        assert.equal(fixture.containers.size, 0);
      }
      return acquireHostAdmission({ ...options, root, capacities, residents: [resident], timeoutMs: 1000 });
    };
    await createLocalSession({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile,
      environment: { CARTULARY_HOST_ADMISSION_LEASE: parent.token }, dependencies: fixture.dependencies });
    assert.equal(readLocalSessionDescriptor(fixture.sessionFile).resident_allocation.allocation.parent, parent.token);
    await parent.release();
    assert.equal(transaction({ operation: "resident_granted", token: resident.token, reference: resident.reference }), true);
    assert.equal(transaction({ operation: "admit", token: contenderToken, owner: processIdentity(), mode: "shared", parent: null,
      claims: hostClaims({ memory_mb: 600, process: 20 }), capacities }), false);
    await stopLocalSession({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile, dependencies: fixture.dependencies });
    assert.equal(fixture.terminateCount(), 1);
    assert.equal(transaction({ operation: "resident_granted", token: resident.token, reference: resident.reference }), false);
    assert.equal(transaction({ operation: "admit", token: contenderToken, owner: processIdentity(), mode: "shared", parent: null,
      claims: hostClaims({ memory_mb: 600, process: 20 }), capacities }), true);
  } finally {
    for (const token of [pendingToken, contenderToken]) transaction({ operation: "release", token });
    if (resident) transaction({ operation: "release_resident", token: resident.token, reference: resident.reference });
    for (const lease of leases.reverse()) await lease.release();
    fixture.cleanup();
  }
});

test("queued physical admission withdraws on cancellation and retains residents until exact release", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "cartulary-session-admission-"));
  const capacities = hostClaims({ cpu: 2, io: 1, memory_mb: 1024, process: 20, object_store: 1, postgres: 1, port_lane: 2 });
  const resident = { token: "f".repeat(32), reference: path.join(root, "resident.json"), claims: hostClaims({ memory_mb: 512, process: 10, object_store: 1, postgres: 1, port_lane: 2 }), capacities, parent: null };
  const leases = [];
  const state = () => JSON.parse(readFileSync(path.join(root, "state.json")));
  try {
    const holder = await acquireHostAdmission({ root, capacities, claims: { cpu: 2 } }); leases.push(holder);
    const controller = new AbortController();
    const waiting = acquireHostAdmission({ root, capacities, claims: { cpu: 1 }, residents: [resident], signal: controller.signal });
    const rejected = assert.rejects(waiting, /cancel queued session/u);
    // Observe registration, not a guessed scheduler delay.
    const deadline = performance.now() + 5000;
    while (state().leases.length < 2) {
      assert.ok(performance.now() < deadline, "pending admission was not registered");
      await new Promise((resolve) => setTimeout(resolve, 10));
    }
    assert.equal(state().residents.length, 0);
    controller.abort(new Error("cancel queued session"));
    await rejected;
    assert.equal(state().leases.length, 1);
    const acquiring = acquireHostAdmission({ root, capacities, claims: { cpu: 1 }, residents: [resident] });
    await holder.release();
    const active = await acquiring; leases.push(active);
    await active.release();
    assert.equal(state().residents.length, 1);
    const drain = await acquireHostAdmission({ root, capacities, claims: { cpu: 1 }, mode: "drain", residents: [resident] }); leases.push(drain);
    assert.equal(state().residents.length, 1);
    transactAdmission({ root, operation: "release_resident", token: resident.token, reference: resident.reference });
    await drain.release();
    assert.equal(state().residents.length, 0);
  } finally { for (const lease of leases) await lease.release(); rmSync(root, { recursive: true, force: true }); }
});

test("session mode is explicit, closed, and limited to approved leaf targets", () => {
  assert.deepEqual(resolveServiceSessionMode({ target: "check", environment: {} }), {
    mode: "owned",
    sessionFile: null,
  });
  const sessionFile = path.join(tmpdir(), "cartulary-explicit-session.json");
  assert.deepEqual(
    resolveServiceSessionMode({
      target: "test-slice", selection: { SCOPE: "services" },
      environment: {
        SCOPE: "services",
        CARTULARY_TEST_SERVICES_MODE: "attach",
        CARTULARY_TEST_SERVICES_SESSION_FILE: sessionFile,
      },
    }),
    { mode: "attach", sessionFile },
  );
  for (const [target, selection] of [["test-slice", { SCOPE: "services" }], ["browser-e2e", { SUITE: "functional" }], ["browser-e2e", { SUITE: "stateful" }]]) {
    assert.equal(
      resolveServiceSessionMode({
        target,
        environment: {
          ...selection,
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

  const allowed = ["browser-e2e", "test-slice"];
  for (const [target, selection] of [["test-slice", { SCOPE: "all" }], ...["all", "measurement", "visual", "accessibility", "support"].map((SUITE) => ["browser-e2e", { SUITE }])]) {
    assert.throws(() => resolveServiceSessionMode({ target, environment: { ...selection, CARTULARY_TEST_SERVICES_MODE: "attach" } }), /owned-only/u);
  }
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

test("session lifecycle is redacted, concurrent-borrower safe, and idempotent", async () => {
  const fixture = sessionFixture();
  try {
    const created = await createFixtureSession(fixture);
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
      target: "test-slice", selection: { SCOPE: "services" },
      runID: "borrower-one",
      suiteRuntime: runtime("1"),
      dependencies: { ...fixture.dependencies, borrowerID: () => "borrower-one" },
    });
    const second = attachLocalSession({
      root: repositoryRoot,
      binary: fixture.binary,
      sessionFile: fixture.sessionFile,
      target: "browser-e2e", selection: { SUITE: "stateful" },
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
    const parent=processIdentity(process.ppid);
    const nativeOptions={root:repositoryRoot,binary:fixture.binary,sessionFile:fixture.sessionFile,target:"test-slice",selection:{SCOPE:"services"},runID:"borrower-native",suiteRuntime:runtime("3"),borrowerOwner:parent,dependencies:{...fixture.dependencies,borrowerID:()=>"f".repeat(32)}};
    assert.throws(()=>attachLocalSession({...nativeOptions,borrowerOwner:{...parent,start:"0"}}),/invalid native borrower/u);
    const native=attachLocalSession(nativeOptions);
    assert.equal(native.borrower_proof.pid,process.ppid);
    assert.throws(()=>detachLocalBorrower({sessionFile:fixture.sessionFile,borrower:{...native.borrower_proof,run_id:"forged"}}),/replaced/u);
    detachLocalBorrower({sessionFile:fixture.sessionFile,borrower:native.borrower_proof});
    detachLocalBorrower({sessionFile:fixture.sessionFile,borrower:native.borrower_proof});
    assert.equal(fixture.terminateCount(),0);

    const output = JSON.stringify(status);
    for (const forbidden of ["session-password", "session-secret-key", postgresID, objectStoreID, descriptor.state_root]) {
      assert.equal(output.includes(forbidden), false);
    }
    await assert.rejects(
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
      (await stopLocalSession({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile, dependencies: fixture.dependencies })).state,
      "absent",
    );
    assert.equal(fixture.terminateCount(), 1);
    assert.equal(
      (await stopLocalSession({ root: repositoryRoot, binary: fixture.binary, sessionFile: fixture.sessionFile, dependencies: fixture.dependencies })).state,
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

test("status detects owner, digest, expiry, and container invalidation", async () => {
  const ownerFixture = sessionFixture();
  try {
    await createFixtureSession(ownerFixture);
    rewriteDescriptor(ownerFixture, (descriptor) => {
      descriptor.owner_uid += 1;
    });
    assert.equal(localSessionStatus({ root: repositoryRoot, binary: ownerFixture.binary, sessionFile: ownerFixture.sessionFile }).state, "invalid");
  } finally {
    ownerFixture.cleanup();
  }

  const digestFixture = sessionFixture();
  try {
    await createFixtureSession(digestFixture);
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
    await createFixtureSession(expiredFixture);
    assert.equal(localSessionStatus({ root: repositoryRoot, binary: expiredFixture.binary, sessionFile: expiredFixture.sessionFile, dependencies: expiredFixture.dependencies }).state, "expired");
  } finally {
    expiredFixture.cleanup();
  }

  const containerFixture = sessionFixture();
  try {
    await createFixtureSession(containerFixture);
    containerFixture.containers.get(objectStoreID).State.Running = false;
    const status = localSessionStatus({ root: repositoryRoot, binary: containerFixture.binary, sessionFile: containerFixture.sessionFile, dependencies: containerFixture.dependencies });
    assert.equal(status.state, "stale");
    assert.equal(status.reason, "container_not_running");
  } finally {
    containerFixture.cleanup();
  }
});
