import type {
  ApplyImportSessionRequest,
  CancelJobRequest,
  CreateImportSessionRequest,
  CreateImportUnitRegionRequest,
  GetImportSessionResponse,
  GetImportUnitPreviewResponse,
  GetImportUnitResponse,
  GetJobResponse,
  ListImportUnitsResponse,
  PreviewImportUnitExtensionMappingRequest,
  PreviewImportUnitExtensionMappingResponse,
  SelectImportUnitRequest,
  SelectImportUnitResponse,
} from "@cartulary/protocol-ts/http";

export type {
  ApplyImportSessionRequest,
  CancelJobRequest,
  CreateImportSessionRequest,
  CreateImportUnitRegionRequest,
  GetImportSessionResponse,
  GetImportUnitPreviewResponse,
  GetImportUnitResponse,
  GetJobResponse,
  ListImportUnitsResponse,
  PreviewImportUnitExtensionMappingResponse,
  SelectImportUnitRequest,
  SelectImportUnitResponse,
};

export type ExtensionMappingPreviewRequest =
  PreviewImportUnitExtensionMappingRequest;
export type ImportSourceColumnMapping = NonNullable<
  GetImportUnitResponse["data"]["approved_mapping"]
>["source_columns"][number];

export type DiscoveredImportColumn =
  GetImportUnitPreviewResponse["data"]["columns"][number];
export type DiscoveredImportPreview = GetImportUnitPreviewResponse["data"];
export type DiscoveredImportUnit = GetImportUnitResponse["data"];
export type ImportJobResource = GetJobResponse["data"];
export type ImportSessionResource = GetImportSessionResponse["data"];
export type WorkbookSourceColumnMapping = ImportSourceColumnMapping;
export type ImportSelectionReceipt = SelectImportUnitResponse["data"];
// The generated oneOf request is an open object; refine its adopted alternatives
// at the typed transport boundary rather than letting drafts send arbitrary JSON.
export type ImportMappingRequest = {
  readonly client_txn_id: string;
  readonly header_row_ref: number;
  readonly data_start_row_ref: number;
  readonly source_columns: readonly WorkbookSourceColumnMapping[];
} & (
  | {
      readonly target_view_schema_id: string;
      readonly unknown_column_policy:
        | "preserve_raw_capture"
        | "preserve_custom_attrs"
        | "reject_if_unmapped";
    }
  | {
      readonly target_kind: string;
      readonly extension_profile_id: string;
      readonly owner_mapping_schema_id: string;
      readonly owner_mapping: Record<string, unknown>;
    }
);
export type ImportUploadMetadata = CreateImportSessionRequest["metadata"];

export type ExtensionMappingPreviewResource<OwnerResult> = Omit<
  PreviewImportUnitExtensionMappingResponse["data"],
  "owner_result"
> & {
  readonly owner_result: OwnerResult;
};
