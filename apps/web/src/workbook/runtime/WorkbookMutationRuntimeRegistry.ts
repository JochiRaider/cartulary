import type { PendingReplayScope } from "../utils/workbookPendingQueue";
import type { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";

type RuntimeEntry = {
  readonly key: string;
  readonly runtime: WorkbookMutationRuntime;
};

function scopeKey(scope: PendingReplayScope): string {
  return `${scope.incidentId}\u0000${scope.clientInstanceId}`;
}

/**
 * Browser-runtime owner for the one active incident-scoped mutation queue.
 *
 * Authenticated Workbook shells borrow this authority. An Auth-shell
 * transition therefore detaches presentation and collaboration consumers but
 * does not destroy same-runtime unsent work. Entering a different incident
 * retires the old authority, and disposing the App runtime retires all state.
 */
export class WorkbookMutationRuntimeRegistry {
  private entry: RuntimeEntry | null = null;
  private disposed = false;

  acquire(
    scope: PendingReplayScope,
    create: () => WorkbookMutationRuntime,
  ): WorkbookMutationRuntime {
    if (this.disposed) {
      throw new Error("workbook mutation runtime registry is disposed");
    }
    const key = scopeKey(scope);
    if (this.entry?.key === key) return this.entry.runtime;

    this.retireCurrent(
      this.entry === null
        ? null
        : {
            kind: "incident_changed",
            nextIncidentId: scope.incidentId,
          },
    );
    const runtime = create();
    if (scopeKey(runtime.scope) !== key) {
      runtime.invalidate({ kind: "runtime_disposed" });
      throw new Error("workbook mutation runtime factory returned wrong scope");
    }
    this.entry = { key, runtime };
    return runtime;
  }

  dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    this.retireCurrent(null);
  }

  private retireCurrent(
    reason: {
      readonly kind: "incident_changed";
      readonly nextIncidentId: string;
    } | null,
  ): void {
    const current = this.entry;
    this.entry = null;
    if (current === null) return;
    if (reason !== null) current.runtime.invalidate(reason);
    current.runtime.invalidate({ kind: "runtime_disposed" });
  }
}
