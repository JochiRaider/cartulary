import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import {
  initialGenericCreateDraft,
  workbookCreationAvailable,
} from "../models/genericWorkbookModel";
import { stringifyGridValue } from "../utils/workbookValueFormat";
import type { WorkbookInspectorErrorPresentation } from "./workbookInspectorErrorModel";

export type InspectorRelatedRecordSubject = {
  readonly cells: Readonly<Record<string, { readonly value: unknown }>>;
  readonly recordId: string;
};

export type InspectorRelatedRecordDraftResult =
  | {
      readonly kind: "ready";
      readonly draft: Record<string, string>;
    }
  | {
      readonly kind: "invalid_target";
      readonly reason: "creation_unavailable" | "semantic_mismatch";
    };

export type InspectorRelatedRecordFormModel = {
  readonly draft: Record<string, string>;
  readonly featureGroup: InspectorFeatureGroup;
  readonly isSubmitting: boolean;
  readonly error: WorkbookInspectorErrorPresentation | null;
  readonly targetContract: ViewContract;
};

export function buildInspectorRelatedRecordDraft({
  currentUserId,
  featureGroup,
  subject,
  targetContract,
}: {
  readonly currentUserId: string | null;
  readonly featureGroup: InspectorFeatureGroup;
  readonly subject: InspectorRelatedRecordSubject;
  readonly targetContract: ViewContract;
}): InspectorRelatedRecordDraftResult {
  const route = featureGroup.routeBinding;
  if (
    route.kind !== "view_row_create" ||
    route.owner !== "view_row_create_route" ||
    route.targetViewSchemaId !== targetContract.viewSchemaId
  ) {
    return { kind: "invalid_target", reason: "semantic_mismatch" };
  }
  if (!workbookCreationAvailable(targetContract)) {
    return { kind: "invalid_target", reason: "creation_unavailable" };
  }

  const draft = initialGenericCreateDraft(targetContract, currentUserId);
  for (const binding of featureGroup.seedBindings) {
    const value = inspectorRelatedRecordSeedValue(binding.source, subject);
    if (value !== null) draft[binding.targetFieldKey] = value;
  }
  return { kind: "ready", draft };
}

function inspectorRelatedRecordSeedValue(
  source: InspectorFeatureGroup["seedBindings"][number]["source"],
  subject: InspectorRelatedRecordSubject,
): string | null {
  if (source.kind === "selected_record_id") return subject.recordId;
  if (source.kind === "selected_field_value") {
    if (source.sourceFieldKey === undefined) return null;
    const text = stringifyGridValue(
      subject.cells[source.sourceFieldKey]?.value,
    ).trim();
    return text === "" ? null : text;
  }
  if (source.value === null || source.value === undefined) return null;
  return typeof source.value === "string"
    ? source.value
    : JSON.stringify(source.value);
}
