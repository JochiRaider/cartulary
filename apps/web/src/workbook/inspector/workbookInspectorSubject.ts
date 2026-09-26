import type { InspectorConfig } from "@cartulary/view-contracts";

import type { WorkbookRecordSubject } from "../ports/WorkbookRecordSubject";

export type WorkbookInspectorLiveSubject = Extract<
  WorkbookRecordSubject,
  { readonly kind: "live" }
>;

export type WorkbookInspectorLiveRowBinding = {
  readonly cells: Readonly<Record<string, { readonly value: unknown }>>;
  readonly subject: WorkbookInspectorLiveSubject;
};

export function buildWorkbookInspectorSubject({
  config,
  kind,
  label,
  recordId,
  rowVersion,
  stateLabel,
  surfaceLabel,
}: {
  readonly config: InspectorConfig;
  readonly kind: WorkbookRecordSubject["kind"];
  readonly label: string;
  readonly recordId: string | null | undefined;
  readonly rowVersion: number | null | undefined;
  readonly stateLabel?: string | undefined;
  readonly surfaceLabel: string;
}): WorkbookRecordSubject | null {
  return validatedWorkbookInspectorSubject({
    kind,
    label,
    recordId,
    rowVersion,
    stateLabel,
    surfaceLabel,
    viewSchemaId: config.viewSchemaId,
  });
}

export function updateWorkbookInspectorSubject(
  subject: WorkbookRecordSubject,
  identity: {
    readonly kind: WorkbookRecordSubject["kind"];
    readonly recordId: string | null | undefined;
    readonly rowVersion: number | null | undefined;
  },
): WorkbookRecordSubject | null {
  if (
    identity.kind === subject.kind &&
    identity.recordId?.trim() === subject.recordId &&
    identity.rowVersion === subject.rowVersion
  ) {
    return subject;
  }
  return validatedWorkbookInspectorSubject({
    ...subject,
    ...identity,
    stateLabel:
      identity.kind === "deleted"
        ? "Deleted"
        : subject.kind === "deleted"
          ? undefined
          : subject.stateLabel,
  });
}

export function workbookInspectorSubjectsEqual(
  left: WorkbookRecordSubject | null,
  right: WorkbookRecordSubject | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      left.kind === right.kind &&
      left.viewSchemaId === right.viewSchemaId &&
      left.recordId === right.recordId &&
      left.rowVersion === right.rowVersion)
  );
}

/** Version replacement invalidates review, but does not detach a live source. */
export function workbookInspectorSubjectChange(
  left: WorkbookRecordSubject | null,
  right: WorkbookRecordSubject | null,
): "record_updated" | "retarget" | null {
  if (workbookInspectorSubjectsEqual(left, right)) return null;
  return left?.kind === "live" &&
    right?.kind === "live" &&
    left.viewSchemaId === right.viewSchemaId &&
    left.recordId === right.recordId
    ? "record_updated"
    : "retarget";
}

function validatedWorkbookInspectorSubject({
  kind,
  label,
  recordId,
  rowVersion,
  stateLabel,
  surfaceLabel,
  viewSchemaId,
}: {
  readonly kind: WorkbookRecordSubject["kind"];
  readonly label: string;
  readonly recordId: string | null | undefined;
  readonly rowVersion: number | null | undefined;
  readonly stateLabel?: string | undefined;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
}): WorkbookRecordSubject | null {
  const normalizedRecordId = recordId?.trim() ?? "";
  const normalizedLabel = label.trim();
  const normalizedSurfaceLabel = surfaceLabel.trim();
  const normalizedViewSchemaId = viewSchemaId.trim();
  if (
    normalizedRecordId === "" ||
    normalizedLabel === "" ||
    normalizedSurfaceLabel === "" ||
    normalizedViewSchemaId === "" ||
    typeof rowVersion !== "number" ||
    !Number.isInteger(rowVersion) ||
    rowVersion <= 0
  ) {
    return null;
  }
  const context = {
    label: normalizedLabel,
    recordId: normalizedRecordId,
    rowVersion,
    surfaceLabel: normalizedSurfaceLabel,
    viewSchemaId: normalizedViewSchemaId,
  };
  if (kind === "deleted") {
    return {
      ...context,
      kind,
      stateLabel: stateLabel?.trim() || "Deleted",
    };
  }
  const normalizedStateLabel = stateLabel?.trim();
  return normalizedStateLabel
    ? { ...context, kind, stateLabel: normalizedStateLabel }
    : { ...context, kind };
}
