// Interrupt only the disposable package's database while the actual backup
// process is reading application state. No application fault injection exists.
import assert from "node:assert/strict";
import { writeFileSync } from "node:fs";
import path from "node:path";
import { spawn, spawnSync } from "node:child_process";
import { setTimeout as delay } from "node:timers/promises";

const [work, project, artifacts] = process.argv.slice(2);
assert.match(project, /^cartularymvprecoverysmk[0-9]+$/);
assert.ok(path.isAbsolute(work) && path.isAbsolute(artifacts));
const args = ["compose", "--project-name", project, "--env-file", path.join(work, ".env"), "-f", path.join(work, "release/assets/docker-compose.yml")];
function docker(...arguments_) {
  const r = spawnSync("docker", arguments_, { encoding: "utf8", timeout: 60000, maxBuffer: 1024 * 1024 });
  assert.equal(r.status, 0, `owned recovery interruption command failed: ${arguments_[0]}`);
  return r.stdout.trim();
}
const compose = (...arguments_) => docker(...args, ...arguments_);
const sql = (query) => compose("exec", "-T", "--user", "postgres", "postgres", "psql", "-v", "ON_ERROR_STOP=1", "-U", "postgres", "-d", "cartulary", "-Atc", query);
async function until(check, message) {
  for (let i = 0; i < 60; i++) {
    if (check()) return;
    await delay(500);
  }
  throw new Error(message);
}
const app = compose("ps", "-q", "app");
const database = compose("ps", "-q", "postgres");
for (const id of [app, database]) assert.match(id, /^[a-f0-9]{64}$/);
const previousBackups = sql("SELECT count(*) FROM backup_sets");
const previousJournals = sql("SELECT count(*) FROM operator_recovery_journal WHERE operation='backup_create' AND result='succeeded'");
docker("stop", app);
// The lock provides a deterministic observation point in the genuine snapshot
// reader. Its lifetime ends with this exact owned database service interruption.
const lock = spawn("docker", [...args, "exec", "-T", "--user", "postgres", "postgres", "psql", "-v", "ON_ERROR_STOP=1", "-U", "postgres", "-d", "cartulary", "-c", "BEGIN; SET LOCAL application_name='package-interruption-lock'; LOCK TABLE users IN ACCESS EXCLUSIVE MODE; SELECT pg_sleep(120); ROLLBACK;"], { stdio: "ignore" });
const exited = new Promise((resolve) => { lock.once("exit", resolve); lock.once("error", resolve); });
try {
  await until(() => sql("SELECT count(*) FROM pg_locks l JOIN pg_stat_activity a USING(pid) WHERE a.application_name='package-interruption-lock' AND l.relation='users'::regclass AND l.mode='AccessExclusiveLock' AND l.granted") === "1", "snapshot lock did not become ready");
  const operator = compose("run", "-d", "--no-deps", "--entrypoint", "/usr/local/bin/cartulary-operator", "recovery-operator", "backup", "create", "--source-config-file", "/etc/cartulary/config.toml");
  assert.match(operator, /^[a-f0-9]{64}$/);
  await until(() => sql("SELECT count(*) FROM pg_stat_activity WHERE usename='cartulary_recovery_login' AND wait_event_type='Lock' AND query LIKE '%users%'") !== "0", "real backup did not reach the blocked snapshot read");
  docker("stop", "--time", "1", database);
  await exited;
  await until(() => docker("inspect", "--format", "{{.State.Running}}", operator) === "false", "interrupted backup did not exit");
  const exit = Number(docker("inspect", "--format", "{{.State.ExitCode}}", operator));
  assert.notEqual(exit, 0, "interrupted backup reported success");
  docker("start", database);
  await until(() => docker("inspect", "--format", "{{.State.Health.Status}}", database) === "healthy", "owned database did not recover");
  assert.equal(sql("SELECT count(*) FROM backup_sets"), previousBackups, "interrupted backup published metadata");
  assert.equal(sql("SELECT count(*) FROM operator_recovery_journal WHERE operation='backup_create' AND result='succeeded'"), previousJournals, "interrupted backup published successful evidence");
  docker("rm", operator);
  docker("start", app);
  writeFileSync(path.join(artifacts, "interruption.json"), JSON.stringify({ snapshot_read_observed: true, owned_database_interrupted: true, operator_exit: exit, successful_backups: 0, successful_journals: 0, database_restarted: true, application_restarted: true }, null, 2) + "\n", { mode: 0o600 });
} finally {
  lock.kill("SIGTERM");
  await exited;
}
