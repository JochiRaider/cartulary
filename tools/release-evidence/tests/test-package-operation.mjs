import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { spawn } from "node:child_process";
import { mkdtempSync, mkdirSync, copyFileSync, writeFileSync, readFileSync, rmSync, existsSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";

const root = path.resolve(import.meta.dirname, "../../..");
const work = mkdtempSync(path.join(tmpdir(), "cartulary-package-operation-"));
try {
  mkdirSync(path.join(work, "bin"));
  mkdirSync(path.join(work, "release/assets/scripts"), {recursive:true});
  mkdirSync(path.join(work, "tls"));
  for (const name of ["ca.pem", "postgres.crt", "postgres.key", "seaweed.crt", "seaweed.key", "migration.crt", "migration.key", "recovery.crt", "recovery.key", "runtime.crt", "runtime.key", "restore-migration.crt", "restore-migration.key", "restore-recovery.crt", "restore-recovery.key", "application.crt", "application.key"]) writeFileSync(path.join(work, "tls", name), "isolated fixture");
  const validSettings = readFileSync(path.join(root, "deploy/mvp/.env.example"), "utf8")
    .replaceAll(/replace-[^\n]+/g, "fixture-value")
    .replace("/etc/cartulary/tls", path.join(work, "tls"))
    .replace("CARTULARY_MVP_COMPOSE_PROJECT_NAME=cartulary-mvp", "CARTULARY_MVP_COMPOSE_PROJECT_NAME=isolated-operator");
  writeFileSync(path.join(work, ".env"), validSettings);
  writeFileSync(path.join(work, "release/assets/docker-compose.yml"), "services: {}\n");
  writeFileSync(path.join(work, "restore-verification-target.toml"), "# fixture\n");
  writeFileSync(path.join(work, "config.toml"), "# supplied to the fake container boundary\n");
  const log = path.join(work, "operations.log");
  writeFileSync(path.join(work, "bin/docker"), `#!/usr/bin/env bash
set -eu
printf '%s\\n' "$*" >>"$OPERATION_LOG"
if [[ "$1" == info ]]; then
  if [[ "$*" == *--format* ]]; then echo "$FIXTURE_DAEMON"; fi
  exit "\${DAEMON_STATUS:-0}"
fi
if [[ "$1" == ps ]]; then
  if [[ "\${LIVE_OWNED:-0}" == 1 ]]; then echo cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc; fi
elif [[ "$*" == *'ps --status running -q app'* ]]; then
  if [[ "\${WAS_RUNNING:-1}" == 1 || -f "$STARTED_FILE" ]]; then echo aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa; fi
elif [[ "$*" == *'up -d --wait --force-recreate postgres seaweedfs-s3'* ]]; then
  exit "\${DEPENDENCY_STATUS:-0}"
elif [[ "$*" == *'run'* && "$*" == *' object-store-init' ]]; then
  if [[ "\${OBJECT_FAILURE:-}" == transient && ! -f "$OBJECT_READY" ]]; then touch "$OBJECT_READY"; echo 'object-store init failed: reason_code=retry_exhausted' >&2; exit 1; fi
  if [[ "\${OBJECT_FAILURE:-}" == fatal ]]; then echo 'object-store init failed: reason_code=invalid_deployment_config' >&2; exit 1; fi
elif [[ "$*" == *'backup inspect latest'* ]]; then
  if [[ "\${INSPECT_STATUS:-0}" != 0 ]]; then echo '{"error":{"code":"backup_set_not_found","reason_code":"no_successful_retained_backup"}}'; fi
  exit "\${INSPECT_STATUS:-0}"
elif [[ "$*" == *'backup create'* ]]; then
  if [[ -n "\${BACKUP_RELEASE:-}" ]]; then
    touch "$BACKUP_READY"
    for ((attempt=0; attempt<600; attempt++)); do
      [[ ! -f "$BACKUP_RELEASE" ]] || break
      sleep 0.1
    done
    [[ -f "$BACKUP_RELEASE" ]] || exit 1
  fi
  exit "\${BACKUP_STATUS:-0}"
elif [[ "$*" == *'cartulary.package.bootstrap=true app'* ]]; then
  echo bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
elif [[ "$*" == *'up -d --no-deps --force-recreate app'* ]]; then
  touch "$STARTED_FILE"
  exit "\${START_STATUS:-0}"
elif [[ "$1" == start ]]; then exit "\${RESTART_STATUS:-0}"
elif [[ "$*" == *'package readiness'* ]]; then exit "\${READY_STATUS:-0}"
elif [[ "$1" == inspect ]]; then echo false
fi
`, { mode: 0o700 });
  writeFileSync(path.join(work, "bin/curl"), '#!/usr/bin/env bash\nexit 0\n', { mode: 0o700 });
  for (const name of ["package.sh", "package-common.sh"]) copyFileSync(path.join(root, "deploy/mvp/scripts", name), path.join(work, "release/assets/scripts", name));
  writeFileSync(path.join(work, "release/assets/scripts/package-verification.sh"), `#!/usr/bin/env bash
package_verify() {
  printf '%s\\n' due-verification >>"$OPERATION_LOG"
  return "\${DUE_STATUS:-0}"
}
`);
  writeFileSync(path.join(work,"release/assets/scripts/release-verify.sh"),"#!/usr/bin/env bash\nexit 0\n",{mode:0o755});
  const fixtureEnv = {
    PATH: `${path.join(work, "bin")}:${process.env.PATH}`, CARTULARY_MVP_DIR: work,
    OPERATION_LOG: log, FIXTURE_DAEMON: path.basename(work), STARTED_FILE: path.join(work, "started"), BACKUP_READY: path.join(work, "backup-ready"), OBJECT_READY:path.join(work,"object-ready"),
  };
  function run(operation, extra = {}, args = []) {
    writeFileSync(log, "");
    rmSync(fixtureEnv.STARTED_FILE, { force: true });
    const result = spawnSync("bash", [path.join(work, "release/assets/scripts/package.sh"), operation, ...args], {
      encoding: "utf8", timeout: 10000, env: { ...fixtureEnv, ...extra },
    });
    assert.equal(result.error, undefined);
    return { status: result.status, calls: readFileSync(log, "utf8"), stderr: result.stderr };
  }
  for (const invalid of [
    validSettings + "POSTGRES_DB=other\n",
    validSettings.replace("RESTORE_VERIFY_POSTGRES_DB=cartulary_restore_verify", "RESTORE_VERIFY_POSTGRES_DB=cartulary"),
    validSettings + "BASH_ENV=/tmp/never\n",
    validSettings.replace("POSTGRES_DB=cartulary", "POSTGRES_DB=$(touch injected)"),
    validSettings.replace("CARTULARY_S3_RESTORE_VERIFY_BUCKET=cartulary-restore-verify", "CARTULARY_S3_RESTORE_VERIFY_BUCKET=cartulary"),
  ]) {
    writeFileSync(path.join(work, ".env"), invalid);
    const result = run("backup-create");
    assert.notEqual(result.status, 0);
    assert.equal(result.calls, "", "unsafe settings reject before Docker or provisioning");
  }
  writeFileSync(path.join(work, ".env"), validSettings);
  for (const script of ["start", "backup-create"]) {
    const unavailable = run(script, { DAEMON_STATUS: "1" });
    assert.notEqual(unavailable.status, 0);
    assert.doesNotMatch(unavailable.calls, /up -d|stop app|backup (inspect|create)|due-verification/);
  }
  const exported = run("backup-export", {}, [path.join(work,"export")]);
  assert.equal(exported.status, 0);
  assert.match(exported.calls, /backup export latest.*--output-directory \/transfer-output\/export/);
  assert.match(exported.calls, /--project-name isolated-operator/);
  assert.doesNotMatch(exported.calls, /stop |up -d/, "export leaves serving lifecycle unchanged");
  const backupID="00000000-0000-4000-8000-000000000001";
  mkdirSync(path.join(work,"bundle"));
  const restored=run("restore-bundle",{WAS_RUNNING:"0"},[path.join(work,"bundle"),backupID,"--acknowledge-stale-backup",backupID]);
  assert.equal(restored.status,0,restored.stderr);
  const restoreLog=readFileSync(log,"utf8");
  assert.match(restoreLog,/restore bundle --bundle-directory \/transfer-input --target-config-file \/etc\/cartulary\/config.toml/);
  assert.doesNotMatch(restoreLog,/up -d --no-deps --force-recreate app|restore-verify due|backup create/);
  const replay=run("restore-bundle",{WAS_RUNNING:"0"},[path.join(work,"bundle"),backupID,"--operation-id",backupID]);
  assert.equal(replay.status,0,replay.stderr);
  assert.equal((replay.calls.match(/--operation-id/g)||[]).length,1);
  for (const flag of ["--operation-id","--acknowledge-stale-backup"]) {
    const duplicate=run("restore-bundle",{WAS_RUNNING:"0"},[path.join(work,"bundle"),backupID,flag,backupID,flag,backupID]);
    assert.notEqual(duplicate.status,0);
    assert.equal(duplicate.calls,"");
  }
  const servingRestore=run("restore-bundle",{},[path.join(work,"bundle"),backupID]);
  assert.notEqual(servingRestore.status,0);
  assert.doesNotMatch(readFileSync(log,"utf8"),/up -d --wait|restore bundle --/);
  const relativeExport = run("backup-export", {}, ["relative"]);
  assert.notEqual(relativeExport.status,0);
  assert.equal(relativeExport.calls,"");
  const failedDependencies=run("start",{DEPENDENCY_STATUS:"1"});
  assert.notEqual(failedDependencies.status,0);
  assert.doesNotMatch(failedDependencies.calls,/backup create|up -d --no-deps --force-recreate app|cartulary.package.bootstrap=true/);
  const transient=run("start",{OBJECT_FAILURE:"transient"});
  assert.equal(transient.status,0,transient.stderr);
  assert.equal((transient.calls.match(/ object-store-init/g)||[]).length,2);
  const fatal=run("start",{OBJECT_FAILURE:"fatal"});
  assert.notEqual(fatal.status,0);
  assert.equal((fatal.calls.match(/ object-store-init/g)||[]).length,1);
  assert.match(fatal.stderr,/"gate":"object_initialization"/);
  assert.doesNotMatch(fatal.calls,/cartulary.package.bootstrap=true|backup create|up -d --no-deps --force-recreate app/);
  const renewed=run("start");
  assert.equal(renewed.status,0,renewed.stderr);
  assert.match(renewed.calls,/up -d --wait --force-recreate postgres seaweedfs-s3/);
  assert.ok(renewed.calls.indexOf("--force-recreate postgres") < renewed.calls.indexOf("backup inspect latest"));
  const failedStart = run("start", { READY_STATUS: "1" });
  assert.notEqual(failedStart.status, 0);
  assert.doesNotMatch(failedStart.calls, /backup inspect|due-verification|up -d --no-deps --force-recreate app/);
  assert.match(failedStart.calls, /rm -f b{64}/, "private bootstrap is cleaned on failed readiness");
  for (const extra of [{ DUE_STATUS: "1" }, { INSPECT_STATUS: "1", BACKUP_STATUS: "1" }, {}]) {
    const restarted = run("start", extra);
    assert.equal(restarted.status === 0, Object.keys(extra).length === 0);
    assert.match(restarted.calls, /--project-name isolated-operator/);
    assert.match(restarted.calls, /backup inspect latest/);
    if (extra.BACKUP_STATUS !== "1") assert.match(restarted.calls, /due-verification/);
    assert.doesNotMatch(restarted.calls, /systemctl|wsl|--service-ports/);
    if (Object.keys(extra).length) assert.doesNotMatch(restarted.calls, /up -d --no-deps --force-recreate app/);
  }
  const repaired = run("start", { INSPECT_STATUS: "1", WAS_RUNNING: "0" });
  assert.equal(repaired.status, 0);
  assert.match(repaired.calls, /backup create/);
  assert.ok(repaired.calls.indexOf("backup create") < repaired.calls.indexOf("due-verification"));
  assert.ok(repaired.calls.indexOf("due-verification") < repaired.calls.indexOf("up -d --no-deps --force-recreate app"));
  const restartFailed = run("backup-create", { RESTART_STATUS: "1" });
  assert.notEqual(restartFailed.status, 0);
  assert.match(restartFailed.stderr, /"restart":"failed"/);
  for (const running of ["0", "1"]) for (const backupStatus of ["0", "1"]) {
    const captured = run("backup-create", { WAS_RUNNING: running, BACKUP_STATUS: backupStatus });
    assert.equal(captured.status, Number(backupStatus));
    assert.match(captured.calls, /backup create/);
    assert.equal(/^start [a-f0-9]{64}$/m.test(captured.calls), running === "1", "cleanup restores only an app that was already running");
    assert.doesNotMatch(captured.calls, /up -d --no-deps --force-recreate app/, "cleanup must not create a replacement or initialize dependencies");
  }
  const incomplete = run("backup-create", { LIVE_OWNED: "1" });
  assert.notEqual(incomplete.status, 0);
  assert.match(incomplete.stderr, /"restart":"blocked"/);
  assert.doesNotMatch(incomplete.calls, /^start /m, "live interrupted children prevent restart");
  const interruptedRetry = run("start", { LIVE_OWNED: "1" });
  assert.notEqual(interruptedRetry.status, 0);
  assert.doesNotMatch(interruptedRetry.calls, /stop |up -d|package preflight/);
  const recovered = run("backup-create");
  assert.equal(recovered.status, 0, "quiescent interrupted operation permits fresh admission");
  writeFileSync(log, "");
  const release = path.join(work, "backup-release");
  const running = spawn("bash", [path.join(work, "release/assets/scripts/package.sh"), "backup-create"], {
    env: { ...fixtureEnv, BACKUP_RELEASE: release }, stdio: "ignore",
  });
  const finished = new Promise((resolve, reject) => { running.on("error", reject); running.on("exit", resolve); });
  try {
    const deadline = Date.now() + 10000;
    while (!existsSync(fixtureEnv.BACKUP_READY) && Date.now() < deadline && running.exitCode === null) {
      await new Promise((resolve) => setTimeout(resolve, 10));
    }
    assert.ok(existsSync(fixtureEnv.BACKUP_READY), "first operation acquired exclusion");
    const contested = run("start");
    assert.notEqual(contested.status, 0);
    assert.doesNotMatch(contested.calls, /stop |up -d|package preflight/, "contender cannot change services before exclusion");
  } finally {
    writeFileSync(release, "release\n");
    assert.equal(await finished, 0, "capture child exits and is reaped after the barrier");
  }
} finally {
  const lock = createHash("sha256").update(`${path.basename(work)}\nisolated-operator\n`).digest("hex");
  rmSync(`/tmp/cartulary-package-operations/${lock}.lock`, { force: true });
  rmSync(`/tmp/cartulary-package-operations/${lock}.pending`, { force: true });
  rmSync(work, { recursive: true, force: true });
}
