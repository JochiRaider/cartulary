import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, copyFileSync, writeFileSync, readFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";

const root = path.resolve(import.meta.dirname, "../../..");
const work = mkdtempSync(path.join(tmpdir(), "cartulary-package-operation-"));
try {
  mkdirSync(path.join(work, "bin"));
  mkdirSync(path.join(work, "scripts"));
  writeFileSync(path.join(work, ".env"), "# isolated operator fixture\n");
  writeFileSync(path.join(work, "config.toml"), "# supplied to the fake container boundary\n");
  const log = path.join(work, "operations.log");
  writeFileSync(path.join(work, "bin/docker"), `#!/usr/bin/env bash
set -eu
printf '%s\\n' "$*" >>"$OPERATION_LOG"
if [[ "$1" == info ]]; then exit "\${DAEMON_STATUS:-0}"; fi
if [[ "$*" == *'ps --status running -q app'* ]]; then
  if [[ "\${WAS_RUNNING:-1}" == 1 ]]; then echo aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa; fi
elif [[ "$*" == *'backup inspect latest'* ]]; then
  exit "\${BACKUP_STATUS:-0}"
elif [[ "$*" == *'backup create'* ]]; then
  exit "\${BACKUP_STATUS:-0}"
elif [[ "$*" == *'up -d app'* ]]; then
  exit "\${START_STATUS:-0}"
fi
`, { mode: 0o700 });
  for (const name of ["operation-start.sh", "backup-capture.sh"]) copyFileSync(path.join(root, "deploy/mvp/scripts", name), path.join(work, "scripts", name));
  writeFileSync(path.join(work, "scripts/restore-verify-due.sh"), `#!/usr/bin/env bash
printf '%s\\n' due-verification >>"$OPERATION_LOG"
exit "\${DUE_STATUS:-0}"
`, { mode: 0o700 });
  function run(script, extra = {}) {
    writeFileSync(log, "");
    const result = spawnSync("bash", [path.join(work, "scripts", script)], {
      encoding: "utf8", timeout: 10000,
      env: { PATH: `${path.join(work, "bin")}:${process.env.PATH}`, CARTULARY_MVP_DIR: work, CARTULARY_MVP_COMPOSE_PROJECT_NAME: "isolated-operator", OPERATION_LOG: log, ...extra },
    });
    assert.equal(result.error, undefined);
    return { status: result.status, calls: readFileSync(log, "utf8") };
  }
  for (const script of ["operation-start.sh", "backup-capture.sh"]) {
    const unavailable = run(script, { DAEMON_STATUS: "1" });
    assert.notEqual(unavailable.status, 0);
    assert.doesNotMatch(unavailable.calls, /up -d|stop app|backup (inspect|create)|due-verification/);
  }
  const failedStart = run("operation-start.sh", { START_STATUS: "1" });
  assert.notEqual(failedStart.status, 0);
  assert.doesNotMatch(failedStart.calls, /backup inspect|due-verification/);
  for (const extra of [{ BACKUP_STATUS: "1" }, { DUE_STATUS: "1" }, {}]) {
    const restarted = run("operation-start.sh", extra);
    assert.equal(restarted.status === 0, Object.keys(extra).length === 0);
    assert.match(restarted.calls, /--project-name isolated-operator/);
    assert.match(restarted.calls, /backup inspect latest/);
    assert.match(restarted.calls, /due-verification/);
    assert.doesNotMatch(restarted.calls, /backup create|systemctl|wsl/);
  }
  for (const running of ["0", "1"]) for (const backupStatus of ["0", "1"]) {
    const captured = run("backup-capture.sh", { WAS_RUNNING: running, BACKUP_STATUS: backupStatus });
    assert.equal(captured.status, Number(backupStatus));
    assert.match(captured.calls, /backup create/);
    assert.equal(/^start [a-f0-9]{64}$/m.test(captured.calls), running === "1", "cleanup restores only an app that was already running");
    assert.doesNotMatch(captured.calls, /up -d app/, "cleanup must not create a replacement or initialize dependencies");
  }
} finally {
  rmSync(work, { recursive: true, force: true });
}
