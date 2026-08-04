import type {
  ViewInspectorDisabledCondition,
  ViewInspectorFailureResultBehavior,
  ViewInspectorIncidentRole,
  ViewInspectorPanelID,
  ViewInspectorRouteKind,
  ViewInspectorRouteOwner,
  ViewInspectorSeedSourceKind,
  ViewInspectorSpecializedActionKey,
  ViewInspectorSuccessResultBehavior,
} from "@cartulary/protocol-ts/view-schemas";

export type SortEntry = {
  readonly fieldKey: string;
  readonly direction: "asc" | "desc";
};

export type ViewFieldContract = {
  readonly fieldKey: string;
  readonly label: string;
  readonly createWritable: boolean;
  readonly defaultHidden: boolean;
  readonly stringContractId: string | null;
  readonly directScalarContractId: string | null;
  readonly directReferenceContractId: string | null;
  readonly writeAction: string | null;
  readonly enumValues: readonly string[] | null;
  readonly headerSortFieldKey: string | null;
  readonly filterOps: readonly string[];
  readonly groupable: boolean;
  readonly sortable: boolean;
  readonly readKind: string;
  readonly writeKind: "read_only" | "direct_value" | "action_payload";
  readonly gridEditable: boolean;
  readonly clearable: boolean;
  readonly conflictResolutionClass: string | null;
  readonly entityBindingMode: string | null;
};

export type ViewFieldCapability = {
  readonly editable: boolean;
  readonly filterable: boolean;
  readonly groupable: boolean;
  readonly sortable: boolean;
};

export type InspectorPanelId = ViewInspectorPanelID;
export type InspectorRouteBindingKind = ViewInspectorRouteKind;
export type InspectorRouteBindingOwner = ViewInspectorRouteOwner;
export type InspectorDisabledCondition = ViewInspectorDisabledCondition;
export type InspectorSuccessResultBehavior = ViewInspectorSuccessResultBehavior;
export type InspectorFailureResultBehavior = ViewInspectorFailureResultBehavior;
export type InspectorSeedSourceKind = ViewInspectorSeedSourceKind;
export type InspectorSpecializedActionKey = ViewInspectorSpecializedActionKey;

export type InspectorPanel = {
  readonly panelId: InspectorPanelId;
  readonly label: string;
};

export type InspectorRouteBinding = {
  readonly kind: InspectorRouteBindingKind;
  readonly owner: InspectorRouteBindingOwner;
  readonly targetViewSchemaId?: string | undefined;
  readonly actionKey?: string | undefined;
};

export type InspectorSeedSource = {
  readonly kind: InspectorSeedSourceKind;
  readonly sourceFieldKey?: string | undefined;
  readonly value?: unknown;
};

export type InspectorSeedBinding = {
  readonly targetFieldKey: string;
  readonly source: InspectorSeedSource;
};

export type InspectorFeatureGroup = {
  readonly featureGroupKey: string;
  readonly panelId: InspectorPanelId;
  readonly label: string;
  readonly minimumIncidentRole: ViewInspectorIncidentRole | null;
  readonly mutates: boolean;
  readonly requiresConfirmation: boolean;
  readonly routeBinding: InspectorRouteBinding;
  readonly seedBindings: readonly InspectorSeedBinding[];
  readonly disabledWhen: readonly InspectorDisabledCondition[];
  readonly successResultBehavior: InspectorSuccessResultBehavior;
  readonly failureResultBehavior: InspectorFailureResultBehavior;
};

export type InspectorConfig = {
  readonly inspectorConfigSchemaId: "cartulary.inspector_config.v1";
  readonly viewSchemaId: string;
  readonly defaultOpen: false;
  readonly subjectBinding: {
    readonly kind: "selected_record";
  };
  readonly noRowState: "no_row_selected";
  readonly unsupportedFeatureBehavior: "omit_feature";
  readonly panels: readonly InspectorPanel[];
  readonly featureGroups: readonly InspectorFeatureGroup[];
};

export type ViewContract = {
  readonly viewSchemaId: string;
  readonly title: string;
  readonly surfaceKind: string;
  readonly defaultHiddenFields: readonly string[];
  readonly defaultSort: readonly SortEntry[];
  readonly defaultVisibleFields: readonly string[];
  readonly filterFields: readonly string[];
  readonly fields: readonly ViewFieldContract[];
  readonly fieldMap: Readonly<Record<string, ViewFieldContract>>;
  readonly groupingFields: readonly string[];
  readonly inspectorConfig: InspectorConfig;
  readonly minimumCreateFieldSets: readonly (readonly string[])[];
  readonly permitsZeroFieldCreate: boolean;
  readonly requiredReferencePackKeys: readonly string[];
  readonly sortableFieldMap: Readonly<Record<string, true>>;
  readonly filterableFieldMap: Readonly<Record<string, true>>;
  readonly groupableFieldMap: Readonly<Record<string, true>>;
  readonly sortFields: readonly string[];
  readonly sortNullOrder: "last";
  readonly technicalFields: readonly string[];
};

export type WorkbookSurfaceStatus =
  | "required_built_in_sheet"
  | "required_system_view"
  | "standardized_optional_workbook_surface";

export type WorkbookSurfaceKind = "built_in_sheet" | "system_view";

export type WorkbookSurfaceContract = {
  readonly contract: ViewContract;
  readonly requiredReferencePackKeys: readonly string[];
  readonly sourceRecordTypes: readonly string[];
  readonly surfaceKind: WorkbookSurfaceKind;
  readonly surfaceStatus: WorkbookSurfaceStatus;
  readonly title: string;
  readonly viewSchemaId: string;
};

export type ViewRowCellV1 = {
  readonly value: unknown;
};

type NormalizedViewRowDataV1 = {
  readonly recordId: string;
  readonly rowVersion: number;
  readonly viewSchemaId: string;
  readonly cells: Readonly<Record<string, ViewRowCellV1>>;
  readonly groupValues?: Readonly<Record<string, unknown>> | undefined;
};

declare const normalizedViewRowFullBrand: unique symbol;
declare const normalizedViewRowPatchBrand: unique symbol;

export type NormalizedViewRowV1 = NormalizedViewRowDataV1 & {
  readonly [normalizedViewRowFullBrand]: "full";
};

export type NormalizedViewRowPatchV1 = NormalizedViewRowDataV1 & {
  readonly [normalizedViewRowPatchBrand]: "patch";
};
