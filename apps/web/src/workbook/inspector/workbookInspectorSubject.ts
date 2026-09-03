import type { InspectorConfig } from "@cartulary/view-contracts";

type WorkbookInspectorSubjectIdentity = {
  readonly kind: "live" | "deleted";
  readonly recordId: string;
  readonly rowVersion: number;
  readonly viewSchemaId: string;
};

type WorkbookInspectorSubjectContext = WorkbookInspectorSubjectIdentity & {
  readonly label: string;
  readonly surfaceLabel: string;
};

export type WorkbookInspectorSubject =
  | (WorkbookInspectorSubjectContext & {
      readonly kind: "live";
      readonly stateLabel?: string | undefined;
    })
  | (WorkbookInspectorSubjectContext & {
      readonly kind: "deleted";
      readonly stateLabel: string;
    });

export type WorkbookInspectorLiveSubject = Extract<
  WorkbookInspectorSubject,
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
  readonly kind: WorkbookInspectorSubject["kind"];
  readonly label: string;
  readonly recordId: string | null | undefined;
  readonly rowVersion: number | null | undefined;
  readonly stateLabel?: string | undefined;
  readonly surfaceLabel: string;
}): WorkbookInspectorSubject | null {
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
  subject: WorkbookInspectorSubject,
  identity: {
    readonly kind: WorkbookInspectorSubject["kind"];
    readonly recordId: string | null | undefined;
    readonly rowVersion: number | null | undefined;
  },
): WorkbookInspectorSubject | null {
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
  left: WorkbookInspectorSubject | null,
  right: WorkbookInspectorSubject | null,
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

function validatedWorkbookInspectorSubject({
  kind,
  label,
  recordId,
  rowVersion,
  stateLabel,
  surfaceLabel,
  viewSchemaId,
}: {
  readonly kind: WorkbookInspectorSubject["kind"];
  readonly label: string;
  readonly recordId: string | null | undefined;
  readonly rowVersion: number | null | undefined;
  readonly stateLabel?: string | undefined;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
}): WorkbookInspectorSubject | null {
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
