import type { InspectorConfig } from "@cartulary/view-contracts";

type WorkbookInspectorSubjectContext = {
  readonly label: string;
  readonly recordId: string;
  readonly rowVersion: number;
  readonly surfaceLabel: string;
  readonly viewSchemaId: string;
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
  return validatedWorkbookInspectorSubject({
    ...subject,
    ...identity,
    stateLabel: identity.kind === "deleted" ? "Deleted" : subject.stateLabel,
  });
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
  if (
    normalizedRecordId === "" ||
    normalizedLabel === "" ||
    normalizedSurfaceLabel === "" ||
    !Number.isInteger(rowVersion) ||
    (rowVersion ?? 0) <= 0
  ) {
    return null;
  }
  const context = {
    label: normalizedLabel,
    recordId: normalizedRecordId,
    rowVersion: rowVersion as number,
    surfaceLabel: normalizedSurfaceLabel,
    viewSchemaId,
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
