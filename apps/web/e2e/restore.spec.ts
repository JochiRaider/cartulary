import { type ChildProcessWithoutNullStreams, spawn } from "node:child_process";
import { fileURLToPath } from "node:url";

import { workbookShellReadyTestId } from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { waitForCommittedRowSummary } from "./measurement/timingSupport";
import { applyCookies, requireCookie } from "./support/auth/browserSession";

type RestoreTarget = {
  backup_set_id: string;
  consistency_point_at: string;
  incident_id: string;
  origin: string;
  restored_incident_ids: string[];
  schema_id: string;
  timeline_summary: string;
  user_email: string;
  user_password: string;
};

test("restore recovers workbook surface and executes a built-in workbook query", async ({
  page,
}) => {
  test.setTimeout(120_000);

  const runtimeRoot = process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT;
  if (!runtimeRoot) {
    throw new Error("CARTULARY_WEB_E2E_RUNTIME_ROOT is required");
  }

  const target = await startRestoreTarget(runtimeRoot);
  try {
    const incidentId = target.ready.incident_id;
    const summary = target.ready.timeline_summary;
    expect(target.ready.schema_id).toBe(
      "cartulary.restore.browser_restore_target.v1",
    );
    expect(target.ready.backup_set_id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u,
    );
    expect(target.ready.restored_incident_ids).toContain(incidentId);

    await loginTargetLocalSession(
      page,
      target.ready.origin,
      target.ready.user_email,
      target.ready.user_password,
    );
    await page.goto(`${target.ready.origin}/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await waitForCommittedRowSummary(page, {
      expectedSummary: summary,
      surface: timelineViewSchemaId,
      timeoutMs: 5_000,
    });

    const queryResponse = await page.request.post(
      `${target.ready.origin}/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
      { data: {} },
    );
    expect(queryResponse.ok(), await queryResponse.text()).toBeTruthy();
    const query = (await queryResponse.json()) as {
      data: { rows: Array<{ cells: Record<string, { value: unknown }> }> };
    };
    expect(
      query.data.rows.some(
        (row) =>
          row.cells["timeline.activity_synopsis_text"]?.value === summary,
      ),
    ).toBe(true);
  } finally {
    try {
      if (!page.isClosed()) await page.goto("about:blank");
    } finally {
      await stopRestoreTarget(target.process);
    }
  }
});

async function loginTargetLocalSession(
  page: Parameters<typeof applyCookies>[0],
  origin: string,
  email: string,
  password: string,
) {
  await page.context().clearCookies();
  const response = await page.request.post(`${origin}/api/v1/auth/login`, {
    data: {
      password,
      username: email,
    },
  });
  expect(response.ok(), await response.text()).toBeTruthy();
  await applyCookies(
    page,
    requireCookie(response, "cartulary_session"),
    requireCookie(response, "cartulary_csrf"),
  );
}

async function startRestoreTarget(runtimeRoot: string): Promise<{
  process: ChildProcessWithoutNullStreams;
  ready: RestoreTarget;
}> {
  const repoRoot = fileURLToPath(new URL("../../..", import.meta.url));
  const child = spawn(
    "go",
    ["run", "./tools/recoverybrowserrestore", "--runtime-root", runtimeRoot],
    {
      cwd: repoRoot,
      detached: true,
      env: process.env,
      stdio: ["pipe", "pipe", "pipe"],
    },
  );

  let stdout = "";
  let stderr = "";
  try {
    return await new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error(`restore target timed out\n${stderr}`));
      }, 90_000);

      child.stderr.on("data", (chunk: Buffer) => {
        stderr = (stderr + chunk.toString("utf8")).slice(-32_768);
      });
      child.stdout.on("data", (chunk: Buffer) => {
        stdout += chunk.toString("utf8");
        const newline = stdout.indexOf("\n");
        if (newline < 0) {
          return;
        }
        clearTimeout(timeout);
        const line = stdout.slice(0, newline);
        try {
          resolve({
            process: child,
            ready: JSON.parse(line) as RestoreTarget,
          });
        } catch (error) {
          reject(
            new Error(
              `decode restore target ready payload: ${String(error)}\nstderr=${stderr}`,
            ),
          );
        }
      });
      child.once("exit", (code, signal) => {
        clearTimeout(timeout);
        if (stdout.includes("\n")) {
          return;
        }
        reject(
          new Error(
            `restore target exited before ready code=${code} signal=${signal}\nstderr=${stderr}`,
          ),
        );
      });
      child.once("error", (error) => {
        clearTimeout(timeout);
        reject(error);
      });
    });
  } catch (error) {
    try {
      await stopRestoreTarget(child);
    } catch (cleanupError) {
      throw new AggregateError(
        [error, cleanupError],
        "Restore fixture startup and cleanup failed",
      );
    }
    throw error;
  }
}

async function stopRestoreTarget(child: ChildProcessWithoutNullStreams) {
  if (child.pid === undefined) {
    return;
  }
  const exited = () => child.exitCode !== null || child.signalCode !== null;
  if (!exited()) {
    // The helper owns cleanup; EOF reaches it through the go-run supervisor.
    child.stdin.end();
    // Two five-second HTTP drains and two ten-second database retirements,
    // plus bounded runtime closure. Forced retirement remains a failed fixture.
    await waitForExit(child, 35_000);
  }
  if (!exited()) {
    // Retire the whole owned process group, including a go-run child.
    process.kill(-child.pid, "SIGTERM");
    await waitForExit(child, 5_000);
    if (!exited()) {
      process.kill(-child.pid, "SIGKILL");
      await waitForExit(child, 5_000);
    }
    throw new Error(
      "Restore fixture did not finish its cleanup before the deadline",
    );
  }
  if (child.exitCode !== 0) {
    throw new Error(
      `Restore fixture cleanup failed code=${child.exitCode} signal=${child.signalCode}`,
    );
  }
}

async function waitForExit(
  child: ChildProcessWithoutNullStreams,
  timeoutMs: number,
) {
  if (child.exitCode !== null || child.signalCode !== null) {
    return;
  }
  await new Promise<void>((resolve) => {
    const finish = () => {
      clearTimeout(timeout);
      child.off("exit", finish);
      resolve();
    };
    const timeout = setTimeout(finish, timeoutMs);
    child.once("exit", finish);
  });
}
