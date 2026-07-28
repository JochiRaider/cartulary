import type { WorkbookSheetRef } from "../models/workbookStartup";
import type { RecordChangedPayload } from "./workbookCollaborationMessages";

export type WorkbookSurfaceIdentity = {
  readonly sheetRef: WorkbookSheetRef;
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
  clearAuthorizedRows(): void;
  refresh(options?: {
    readonly recordId?: string | undefined;
    readonly reason?: string | undefined;
  }): Promise<void>;
};
