import { spawn, spawnSync } from "node:child_process";
import { randomUUID } from "node:crypto";
import { closeSync, mkdirSync, openSync, readFileSync } from "node:fs";
import { createServer } from "node:net";
import path from "node:path";

function run(command, args, { cwd, environment }) {
  const result = spawnSync(command, args, {
    cwd,
    env: { ...process.env, ...environment },
    encoding: "utf8",
    maxBuffer: 32 * 1024 * 1024,
  });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    throw new Error(
      `${path.basename(command)} ${args[0] ?? ""} failed: ${(result.stderr || result.stdout || `exit ${result.status}`).trim()}`,
    );
  }
}

function readEnvironmentFile(file) {
  const environment = JSON.parse(readFileSync(file, "utf8"));
  if (!environment || typeof environment !== "object" || Array.isArray(environment)) {
    throw new Error(`invalid environment object in ${file}`);
  }
  for (const [name, value] of Object.entries(environment)) {
    if (!/^[A-Z][A-Z0-9_]*$/u.test(name)) {
      throw new Error(`invalid environment name in ${file}`);
    }
    if (typeof value !== "string") {
      throw new Error(`invalid environment value for ${name} in ${file}`);
    }
  }
  return environment;
}

function availableLoopbackPort() {
  return new Promise((resolve, reject) => {
    const server = createServer();
    server.unref();
    server.once("error", reject);
    server.listen(0, "127.0.0.1", () => {
      const address = server.address();
      server.close((error) => {
        if (error) reject(error);
        else resolve(address.port);
      });
    });
  });
}

function commandSucceeded(command, args, options) {
  const result = spawnSync(command, args, {
    ...options,
    encoding: "utf8",
    maxBuffer: 32 * 1024 * 1024,
  });
  return !result.error && result.status === 0;
}

function objectStoreNamespaceProvider({ root, runRoot, suiteController }) {
  const proxyRoot = path.join(runRoot, "_shared", "object-store-proxy");
  const proxyBinary = path.join(proxyRoot, "s3corsproxy");
  let binaryReady = false;
  return {
    async acquire() {
      const suite = suiteController.ensure();
      const upstreamEndpoint = suite.environment.CARTULARY_S3TEST_ENDPOINT;
      if (!upstreamEndpoint) throw new Error("test-services suite has no object-store endpoint");
      mkdirSync(proxyRoot, { recursive: true, mode: 0o700 });
      if (!binaryReady) {
        run(process.env.GO || "go", ["build", "-o", proxyBinary, "./tools/s3corsproxy"], {
          cwd: root,
          environment: {
            GOCACHE: process.env.GOCACHE || process.env.GO_CACHE_DIR || "/tmp/cartulary-go-build",
            GOMODCACHE: process.env.GOMODCACHE || process.env.GO_MOD_CACHE_DIR || "/tmp/cartulary-go-mod",
          },
        });
        binaryReady = true;
      }
      const instanceID = randomUUID().replaceAll("-", "");
      const instanceRoot = path.join(proxyRoot, instanceID);
      mkdirSync(instanceRoot, { recursive: true, mode: 0o700 });
      const attemptFile = path.join(instanceRoot, "start-attempt.json");
      const leaseFile = path.join(instanceRoot, "ready-lease.json");
      const logFile = path.join(instanceRoot, "proxy.log");
      const listen = `127.0.0.1:${await availableLoopbackPort()}`;
      const upstream = `http://${upstreamEndpoint}`;
      const origin = "http://localhost:5173";
      const common = ["--listen", listen, "--upstream", upstream, "--origin", origin];
      run(
        proxyBinary,
        [
          "attempt",
          ...common,
          "--attempt-file",
          attemptFile,
          "--instance-id",
          instanceID,
          "--log-path",
          logFile,
        ],
        { cwd: root, environment: {} },
      );
      const logFD = openSync(logFile, "a", 0o600);
      const child = spawn(
        proxyBinary,
        [
          "serve",
          ...common,
          "--attempt-file",
          attemptFile,
          "--instance-id",
          instanceID,
          "--log-path",
          logFile,
        ],
        { cwd: root, detached: true, stdio: ["ignore", logFD, logFD] },
      );
      child.unref();
      closeSync(logFD);
      try {
        let ready = false;
        for (let attempt = 0; attempt < 200 && !ready; attempt += 1) {
          ready = commandSucceeded(
            proxyBinary,
            ["status", ...common, "--state-file", attemptFile],
            { cwd: root },
          );
          if (!ready) await new Promise((resolve) => setTimeout(resolve, 25));
        }
        if (!ready) {
          throw new Error(`run-scoped object-store proxy failed to become ready; inspect ${logFile}`);
        }
        run(
          proxyBinary,
          ["promote", ...common, "--attempt-file", attemptFile, "--lease-file", leaseFile],
          { cwd: root, environment: {} },
        );
      } catch (error) {
        commandSucceeded(proxyBinary, ["stop", ...common, "--state-file", attemptFile], {
          cwd: root,
        });
        throw error;
      }
      const stop = () =>
        run(proxyBinary, ["stop", ...common, "--state-file", leaseFile], {
          cwd: root,
          environment: {},
        });
      return {
        ownership: "owned",
        resource_ids: [
          "suite:object_store_namespace",
          `object-store-proxy:${instanceID}`,
        ],
        resource: {
          environment: {
            ...suite.environment,
            OBJECT_STORE_ENDPOINT: listen,
            SEAWEEDFS_S3_ACCESS_KEY_ID: suite.environment.CARTULARY_S3TEST_ACCESS_KEY_ID,
            SEAWEEDFS_S3_SECRET_ACCESS_KEY: suite.environment.CARTULARY_S3TEST_SECRET_ACCESS_KEY,
            OBJECT_STORE_SECURE: "false",
          },
        },
        release: stop,
        quarantine: stop,
        destroy: stop,
      };
    },
    close: () => suiteController.close(),
  };
}

export function startManagedSuite({ root, runRoot, target, environment = {} }) {
  const suiteRoot = path.join(runRoot, "_shared", "work-graph-suite");
  mkdirSync(suiteRoot, { recursive: true, mode: 0o700 });
  const envFile = path.join(suiteRoot, "suite-environment.json");
  const leaseFile = path.join(suiteRoot, "suite-lease.json");
  const executable =
    process.env.TEST_SERVICES_BIN ||
    process.env.CARTULARY_TEST_SERVICES_BIN ||
    path.join(root, "tmp/toolbin/cartulary-test-services");
  const suiteID = `work-graph-${target}`.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
  run(executable, ["start-suite", "--env-file", envFile, "--lease-file", leaseFile], {
    cwd: root,
    environment: {
      ...environment,
      CARTULARY_TEST_SUITE_ID: suiteID,
      CARTULARY_TEST_TARGET: target,
    },
  });
  const suiteEnvironment = JSON.parse(readFileSync(envFile, "utf8"));
  if (!suiteEnvironment || typeof suiteEnvironment !== "object" || Array.isArray(suiteEnvironment)) {
    throw new Error("test-services suite environment is invalid");
  }
  return {
    environment: suiteEnvironment,
    leaseFile,
    executable,
    close() {
      run(executable, ["terminate-suite", "--lease", leaseFile], {
        cwd: root,
        environment: { ...environment, ...suiteEnvironment },
      });
    },
  };
}

function borrowedProvider(suiteController, resourceID, environmentForSuite = (value) => value) {
  return {
    async acquire() {
      const suite = suiteController.ensure();
      const environment = environmentForSuite(suite.environment);
      return {
        ownership: "borrowed",
        resource_ids: [resourceID],
        resource: { environment },
        detach() {},
      };
    },
    close: () => suiteController.close(),
  };
}

function browserSuiteEnvironment(environment) {
  const exact = new Set([
    "CARTULARY_TEST_SERVICES_ACTIVE",
    "CARTULARY_TEST_SERVICES_CALL_MODE",
    "CARTULARY_TEST_SERVICES_LIFECYCLE_MODE",
    "CARTULARY_TEST_SUITE_ID",
  ]);
  return Object.fromEntries(
    Object.entries(environment).filter(([name]) =>
      exact.has(name) ||
      name.startsWith("CARTULARY_PGTEST_") ||
      name.startsWith("CARTULARY_S3TEST_"),
    ),
  );
}

export function productionFixtureProviders({ root, runRoot, suiteController }) {
  const sharedProviders = Object.fromEntries(
    [
      "postgres_transaction",
      "postgres_group",
      "postgres_dedicated",
      "postgres_migration",
      "managed_process",
    ].map((capability) => [
      capability,
      borrowedProvider(suiteController, `suite:${capability}`),
    ]),
  );
  return {
    ...sharedProviders,
    object_store_namespace: objectStoreNamespaceProvider({ root, runRoot, suiteController }),
    browser_stack: {
      async acquire({ affinityKey, runtimeProfileID = "default" }) {
        const suite = suiteController.ensure();
        const suiteEnvironment = suite.environment;
        const safeAffinity = String(affinityKey).replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
        const sessionRoot = path.join(runRoot, "_shared", "browser-stack-leases", safeAffinity);
        mkdirSync(sessionRoot, { recursive: true, mode: 0o700 });
        const envFile = path.join(sessionRoot, "stack.env");
        const leaseFile = path.join(sessionRoot, "stack.lease");
        const lifecycle = path.join(root, "tools/harness/browser/start-web-e2e.sh");
        const environment = {
          ...suiteEnvironment,
          CARTULARY_BROWSER_RUNTIME_PROFILE_ID: runtimeProfileID,
          CARTULARY_BROWSER_SERVICE_REQUIREMENT: "test-services",
          CARTULARY_BROWSER_SESSION_GROUP: safeAffinity,
          CARTULARY_TEST_SUITE_ID: suiteEnvironment.CARTULARY_TEST_SUITE_ID || `work-graph-${safeAffinity}`,
        };
        run(lifecycle, ["--session-start", "--env-file", envFile, "--lease-file", leaseFile], {
          cwd: root,
          environment,
        });
        const stackEnvironment = readEnvironmentFile(envFile);
        const unitEnvironment = {
          ...browserSuiteEnvironment(suiteEnvironment),
          ...stackEnvironment,
        };
        const close = () =>
          run(lifecycle, ["--session-stop", "--lease-file", leaseFile], {
            cwd: root,
            environment: { ...environment, ...unitEnvironment },
          });
        return {
          ownership: "owned",
          resource_ids: [`browser-stack:${safeAffinity}`],
          resource: { environment: unitEnvironment },
          environment: unitEnvironment,
          release: close,
          quarantine: close,
          destroy: close,
        };
      },
      close: () => suiteController.close(),
    },
  };
}
