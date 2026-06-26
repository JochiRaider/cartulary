import { type ChildProcessWithoutNullStreams, spawn } from "node:child_process";
import { fileURLToPath } from "node:url";

import { workbookShellReadyTestId } from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import {
  applyCookies,
  requireCookie,
  waitForCommittedRowSummary,
} from "./helpers";
import { timelineViewSchemaId } from "./phase4Helpers";

type Phase10RestoreTarget = {
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

test("Phase 10 E-10-02 restore recovers workbook surface and executes a built-in workbook query", async ({
  page,
}) => {
  test.setTimeout(120_000);

  const runtimeRoot = process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT;
  if (!runtimeRoot) {
    throw new Error("CARTULARY_WEB_E2E_RUNTIME_ROOT is required");
  }

  const target = await startPhase10RestoreTarget(runtimeRoot);
  try {
    const incidentId = target.ready.incident_id;
    const summary = target.ready.timeline_summary;
    expect(target.ready.schema_id).toBe(
      "cartulary.phase10.browser_restore_target.v1",
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
    await stopPhase10RestoreTarget(target.process);
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

async function startPhase10RestoreTarget(runtimeRoot: string): Promise<{
  process: ChildProcessWithoutNullStreams;
  ready: Phase10RestoreTarget;
}> {
  const repoRoot = fileURLToPath(new URL("../../..", import.meta.url));
  const child = spawn(
    "go",
    ["run", "./tools/phase10browserrestore", "--runtime-root", runtimeRoot],
    {
      cwd: repoRoot,
      env: process.env,
      stdio: ["pipe", "pipe", "pipe"],
    },
  );

  let stdout = "";
  let stderr = "";
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      child.kill("SIGTERM");
      reject(new Error(`phase10 restore target timed out\n${stderr}`));
    }, 90_000);

    child.stderr.on("data", (chunk: Buffer) => {
      stderr += chunk.toString("utf8");
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
          ready: JSON.parse(line) as Phase10RestoreTarget,
        });
      } catch (error) {
        child.kill("SIGTERM");
        reject(
          new Error(
            `decode phase10 restore target ready payload: ${String(error)}\nstdout=${stdout}\nstderr=${stderr}`,
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
          `phase10 restore target exited before ready code=${code} signal=${signal}\nstderr=${stderr}`,
        ),
      );
    });
    child.once("error", (error) => {
      clearTimeout(timeout);
      reject(error);
    });
  });
}

async function stopPhase10RestoreTarget(child: ChildProcessWithoutNullStreams) {
  if (child.exitCode !== null || child.signalCode !== null) {
    return;
  }
  const done = new Promise<void>((resolve) => {
    child.once("exit", () => resolve());
  });
  child.kill("SIGTERM");
  await Promise.race([
    done,
    new Promise<void>((resolve) => setTimeout(resolve, 5_000)),
  ]);
  if (child.exitCode === null && child.signalCode === null) {
    child.kill("SIGKILL");
  }
}
