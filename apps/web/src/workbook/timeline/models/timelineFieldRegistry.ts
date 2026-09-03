import { requireViewContract } from "@cartulary/view-contracts";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { RelationshipFieldKey } from "./workbookMentionChips";

export type TimelineEditableFieldKey =
  | "timeline.date_entered_text"
  | "timeline.analyst_text"
  | "timeline.mitre_stage_text"
  | "timeline.device_object_text"
  | "timeline.ip_address_text"
  | "timeline.activity_utc_text"
  | "timeline.activity_local_text"
  | "timeline.raw_activity_text"
  | "timeline.activity_synopsis_text"
  | "timeline.data_source_text";

export type RelationshipDraftKey = "hostRefs" | "identityRefs";
export type CollectionFieldKey = RelationshipFieldKey | "timeline.tags";
export type CollectionDraftKey = RelationshipDraftKey | "tags";

export type RowValues = {
  dateEnteredText: string;
  analystText: string;
  mitreStageText: string;
  deviceObjectText: string;
  ipAddressText: string;
  activityUTCText: string;
  activityLocalText: string;
  rawActivityText: string;
  activitySynopsisText: string;
  dataSourceText: string;
};

export type FocusFieldKey = keyof RowValues | CollectionDraftKey;

export function isTimelineCollectionDraftKey(
  field: FocusFieldKey,
): field is CollectionDraftKey {
  return field === "hostRefs" || field === "identityRefs" || field === "tags";
}

export type TimelineScalarBinding = {
  readonly kind: "scalar";
  readonly fieldKey: TimelineEditableFieldKey;
  readonly key: keyof RowValues;
  readonly multiline?: boolean;
};

export type TimelineRelationshipBinding = {
  readonly kind: "collection";
  readonly fieldKey: RelationshipFieldKey;
  readonly draftKey: RelationshipDraftKey;
  readonly collectionKind: "relationship";
  readonly entityType: "host" | "identity";
};

export type TimelineTagBinding = {
  readonly kind: "collection";
  readonly fieldKey: "timeline.tags";
  readonly draftKey: "tags";
  readonly collectionKind: "tag";
};

export type TimelineCollectionBinding =
  | TimelineRelationshipBinding
  | TimelineTagBinding;

export type TimelineReadonlyBinding = {
  readonly kind: "readonly";
  readonly fieldKey: string;
};

export type TimelineFieldBinding =
  | TimelineScalarBinding
  | TimelineCollectionBinding
  | TimelineReadonlyBinding;

export type TimelineScalarEditorSurface = "grid" | "inspector";

export const timelineScalarEditorSurfaces: readonly TimelineScalarEditorSurface[] =
  ["grid", "inspector"];

export const timelineScalarBindings = [
  {
    kind: "scalar",
    fieldKey: "timeline.date_entered_text",
    key: "dateEnteredText",
  },
  {
    kind: "scalar",
    fieldKey: "timeline.analyst_text",
    key: "analystText",
  },
  {
    kind: "scalar",
    fieldKey: "timeline.mitre_stage_text",
    key: "mitreStageText",
  },
  {
    kind: "scalar",
    fieldKey: "timeline.device_object_text",
    key: "deviceObjectText",
  },
  {
    kind: "scalar",
    fieldKey: "timeline.ip_address_text",
    key: "ipAddressText",
  },
  {
    kind: "scalar",
    fieldKey: "timeline.activity_utc_text",
    key: "activityUTCText",
  },
  {
    kind: "scalar",
    fieldKey: "timeline.activity_local_text",
    key: "activityLocalText",
  },
  {
    kind: "scalar",
    fieldKey: "timeline.raw_activity_text",
    key: "rawActivityText",
    multiline: true,
  },
  {
    kind: "scalar",
    fieldKey: "timeline.activity_synopsis_text",
    key: "activitySynopsisText",
    multiline: true,
  },
  {
    kind: "scalar",
    fieldKey: "timeline.data_source_text",
    key: "dataSourceText",
  },
] as const satisfies readonly TimelineScalarBinding[];

export const timelineCollectionBindings = [
  {
    kind: "collection",
    fieldKey: "timeline.host_refs",
    draftKey: "hostRefs",
    collectionKind: "relationship",
    entityType: "host",
  },
  {
    kind: "collection",
    fieldKey: "timeline.identity_refs",
    draftKey: "identityRefs",
    collectionKind: "relationship",
    entityType: "identity",
  },
  {
    kind: "collection",
    fieldKey: "timeline.tags",
    draftKey: "tags",
    collectionKind: "tag",
  },
] as const satisfies readonly TimelineCollectionBinding[];

const scalarBindingByField = new Map<string, TimelineScalarBinding>(
  timelineScalarBindings.map((binding) => [binding.fieldKey, binding]),
);
const scalarBindingByValueKey = new Map<keyof RowValues, TimelineScalarBinding>(
  timelineScalarBindings.map((binding) => [binding.key, binding]),
);
const collectionBindingByField = new Map<string, TimelineCollectionBinding>(
  timelineCollectionBindings.map((binding) => [binding.fieldKey, binding]),
);
const timelineContract = requireViewContract(timelineViewSchemaId);

export function timelineFieldBinding(fieldKey: string): TimelineFieldBinding {
  return (
    scalarBindingByField.get(fieldKey) ??
    collectionBindingByField.get(fieldKey) ?? { kind: "readonly", fieldKey }
  );
}

export function timelineScalarBindingForField(
  fieldKey: string,
): TimelineScalarBinding | null {
  return scalarBindingByField.get(fieldKey) ?? null;
}

export function timelineScalarBindingForValueKey(
  key: keyof RowValues,
): TimelineScalarBinding {
  const binding = scalarBindingByValueKey.get(key);
  if (binding === undefined) {
    throw new Error(`Missing Timeline scalar value binding for ${key}.`);
  }
  return binding;
}

function requireScalarBinding(fieldKey: TimelineEditableFieldKey) {
  const binding = scalarBindingByField.get(fieldKey);
  if (binding === undefined) {
    throw new Error(`Missing Timeline scalar binding for ${fieldKey}.`);
  }
  return binding;
}

export const timelineVisibleBindings: readonly TimelineFieldBinding[] =
  timelineContract.fields.map((field) => timelineFieldBinding(field.fieldKey));

export const timelineInspectorBindings: readonly TimelineScalarBinding[] = [
  requireScalarBinding("timeline.raw_activity_text"),
  requireScalarBinding("timeline.activity_synopsis_text"),
];

export const timelineObservationSourceFields = timelineScalarBindings.map(
  (binding) => ({
    fieldKey: binding.fieldKey,
    label:
      timelineContract.fieldMap[binding.fieldKey]?.label ?? binding.fieldKey,
  }),
);

export function inputFocusKey(
  rowKey: string,
  field: FocusFieldKey,
  surface: TimelineScalarEditorSurface = "grid",
) {
  return `${rowKey}:${field}:${surface}`;
}

export function timelineRelationshipLabel(fieldKey: RelationshipFieldKey) {
  return fieldKey === "timeline.identity_refs" ? "Identities" : "Hosts";
}
