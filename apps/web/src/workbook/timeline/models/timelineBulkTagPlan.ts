import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "./timelineRowModel";

export type TimelineBulkTagContext = {
  readonly authorized: boolean;
  readonly capabilityAvailable: boolean;
};

export type TimelineBulkTagPlan =
  | {
      readonly kind: "dispatch";
      readonly normalizedTagName: string;
      readonly targets: readonly {
        readonly baseRowVersion: number;
        readonly recordId: string;
      }[];
    }
  | {
      readonly kind: "reject";
      readonly reason:
        | "authorization_lost"
        | "capability_unavailable"
        | "empty_tag"
        | "empty_selection"
        | "partial_selection"
        | "invalid_target";
    };

export function planTimelineBulkTag(options: {
  readonly context: TimelineBulkTagContext;
  readonly rows: readonly WorkbookRow[];
  readonly queryMembers: ReadonlySet<string> | null;
  readonly selectedRecordIds: ReadonlySet<string>;
  readonly tagName: string;
}): TimelineBulkTagPlan {
  if (!options.context.authorized) {
    return { kind: "reject", reason: "authorization_lost" };
  }
  if (!options.context.capabilityAvailable) {
    return { kind: "reject", reason: "capability_unavailable" };
  }
  const normalizedTagName = options.tagName.trim();
  if (normalizedTagName === "") return { kind: "reject", reason: "empty_tag" };
  if (options.selectedRecordIds.size === 0) {
    return { kind: "reject", reason: "empty_selection" };
  }
  const selectedRows = options.rows.filter(
    (row) =>
      row.recordId !== null && options.selectedRecordIds.has(row.recordId),
  );
  if (selectedRows.length !== options.selectedRecordIds.size) {
    return { kind: "reject", reason: "partial_selection" };
  }
  if (
    selectedRows.some(
      (row) => !timelineBulkTagMember(row, options.queryMembers),
    )
  ) {
    return { kind: "reject", reason: "invalid_target" };
  }
  return {
    kind: "dispatch",
    normalizedTagName,
    targets: selectedRows.map((row) => ({
      baseRowVersion: row.rowVersion as number,
      recordId: row.recordId as string,
    })),
  };
}

export function timelineBulkTagMember(
  row: WorkbookRow,
  queryMembers: ReadonlySet<string> | null,
): boolean {
  return (
    row.viewSchemaId === timelineViewSchemaId &&
    row.recordId !== null &&
    row.rowVersion !== null &&
    Number.isSafeInteger(row.rowVersion) &&
    row.rowVersion > 0 &&
    (queryMembers === null || queryMembers.has(row.recordId))
  );
}
