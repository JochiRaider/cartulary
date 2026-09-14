import type { AuthorizationRecoveryResult } from "../../shared/authorizationRecovery";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";
import type { RecordChangedPayload } from "./workbookCollaborationMessages";

export type WorkbookSurfaceIdentity = {
  readonly sheetRef: SheetRef;
  readonly viewSchemaId: string | null;
};

export type WorkbookSurfaceRecordChangeResult =
  | { readonly kind: "applied" }
  | { readonly kind: "stale" }
  | { readonly kind: "refresh_required" };

/**
 * Renderer-neutral reconciliation capabilities owned by an active Base
 * Workbook surface. Transport and socket state remain outside this port.
 */
export type WorkbookActiveSurfacePort = {
  readonly identity: WorkbookSurfaceIdentity;
  applyRecordChanged(
    payload: RecordChangedPayload,
  ): WorkbookSurfaceRecordChangeResult;
  invalidate(reason: WorkbookQueryInvalidationReason): void;
  refresh(options?: {
    readonly recordId?: string | undefined;
    readonly reason?: string | undefined;
  }): Promise<void>;
};

/** A reconciliation refresh must acknowledge an accepted, current query before replay. */
export class WorkbookSurfaceRefreshError extends Error {
  constructor(
    readonly recovery: Exclude<
      AuthorizationRecoveryResult,
      { kind: "authorized" }
    >,
  ) {
    super(
      "Workbook surface refresh did not establish current authorized state.",
    );
  }
}

export function requireWorkbookSurfaceAcceptance(
  result: WorkbookPortResult<unknown>,
): void {
  if (result.kind === "accepted") return;
  if (result.kind === "aborted")
    throw new WorkbookSurfaceRefreshError({ kind: "cancelled" });
  const kind = result.failure.kind;
  throw new WorkbookSurfaceRefreshError(
    kind === "authentication_required"
      ? { kind: "session_lost" }
      : {
          kind: "unavailable",
          failure: kind === "invalid_contract" ? "contract" : "transient",
        },
  );
}
