import { writeFileSync } from "node:fs";
import type { TestInfo } from "@playwright/test";
import type { JsonRequestContextLike } from "../transport/publicJsonClient";
import { createEnvironmentTestControlClient } from "../transport/testControlEnvironment";

/** The owned work that failed to settle, and the last state it reported. */
export type StallSubject = {
  readonly kind: "job";
  readonly job_id: string;
  readonly last_status: string | null;
};

type StallRecorderTestInfo = Pick<TestInfo, "attach" | "outputPath">;

const jobIdentity =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u;

/**
 * TH-HARNESS-REQ-864: while a stall is still live, keep the harness server's
 * own view of it (database pool, lock waits, goroutine stacks) beside the
 * failed test. Recording is evidence only: it never throws and never changes
 * the test's failure.
 */
export async function recordStallDiagnostics(
  request: JsonRequestContextLike,
  endpointOrigin: string,
  testInfo: StallRecorderTestInfo,
  subject: StallSubject,
): Promise<void> {
  let diagnostics: unknown;
  try {
    const result = await createEnvironmentTestControlClient(request, {
      endpointOrigin,
    }).request({
      method: "GET",
      path: "/api/v1/test/runtime/stall-diagnostics",
    });
    const data =
      result.ok && result.body && typeof result.body === "object"
        ? (result.body as { data?: unknown }).data
        : undefined;
    diagnostics = data ?? { unavailable: `status_${result.status}` };
  } catch {
    diagnostics = { unavailable: "request_failed" };
  }
  try {
    const name = jobIdentity.test(subject.job_id) ? subject.job_id : "unknown";
    const file = testInfo.outputPath(`stall-diagnostics-${name}.json`);
    writeFileSync(
      file,
      `${JSON.stringify({ subject, diagnostics }, null, 2)}\n`,
      {
        mode: 0o600,
      },
    );
    await testInfo.attach("stall-diagnostics", {
      path: file,
      contentType: "application/json",
    });
  } catch {
    // The failed settlement remains the reported failure.
  }
}
