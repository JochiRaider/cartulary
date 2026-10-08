export const fixtureLifecycleAttachment: string;
export type BrowserFixtureFailure = {
  phase: "startup" | "body" | "cleanup";
  failure_class: "harness" | "product" | "interrupted";
  failure_reason: "tool_diagnostic_failure" | "test_assertion_failure" | "cleanup_error" | "cancelled_or_interrupted";
  error: unknown;
};
export function createBrowserFixtureProcess(options: {
  command: string;
  args: string[];
  cwd?: string;
  env?: NodeJS.ProcessEnv;
  startupMs?: number;
  cleanupMs?: number;
  signalMs?: number;
}): {
  ready: Promise<unknown>;
  stop(): Promise<void>;
  report(failures?: BrowserFixtureFailure[]): {
    schema_id: string;
    attempt_id: string;
    duration_ms: number;
    process: { exit_code: number | null; signal: string | null; closed: boolean; forced: boolean };
    events: Array<{ stage: string; phase: string; outcome: string; elapsed_ms: number; deadline_ms: number; message: string }>;
    stderr: string;
    failures: Array<Omit<BrowserFixtureFailure, "error"> & { message: string }>;
  };
}
