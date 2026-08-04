import type {
  ApplyImportSessionRequest,
  ApplyImportSessionResponse,
  CancelJobRequest,
  CancelJobResponse,
  CreateImportSessionResponse,
  CreateImportUnitRegionRequest,
  CreateImportUnitRegionResponse,
  GetImportSessionResponse,
  GetImportUnitPreviewResponse,
  GetImportUnitResponse,
  GetJobResponse,
  ListImportUnitsResponse,
  PreviewImportUnitExtensionMappingRequest,
  PreviewImportUnitExtensionMappingResponse,
  PutImportUnitMappingRequest,
  PutImportUnitMappingResponse,
  SelectImportUnitRequest,
  SelectImportUnitResponse,
  SkipImportUnitRequest,
  SkipImportUnitResponse,
} from "@cartulary/protocol-ts/http";

export type {
  ApplyImportSessionRequest,
  ApplyImportSessionResponse,
  CancelJobRequest,
  CancelJobResponse,
  CreateImportSessionResponse,
  CreateImportUnitRegionRequest,
  CreateImportUnitRegionResponse,
  GetImportSessionResponse,
  GetImportUnitPreviewResponse,
  GetImportUnitResponse,
  GetJobResponse,
  ListImportUnitsResponse,
  PreviewImportUnitExtensionMappingResponse,
  PutImportUnitMappingRequest,
  PutImportUnitMappingResponse,
  SelectImportUnitRequest,
  SelectImportUnitResponse,
  SkipImportUnitRequest,
  SkipImportUnitResponse,
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
export type ImportResourceRef = NonNullable<
  NonNullable<GetJobResponse["data"]["result_summary"]>["resource_refs"]
>[number];
export type ImportSessionResource = GetImportSessionResponse["data"];
export type WorkbookSourceColumnMapping = ImportSourceColumnMapping;

export type ExtensionMappingPreviewResource<OwnerResult> = Omit<
  PreviewImportUnitExtensionMappingResponse["data"],
  "owner_result"
> & {
  readonly owner_result: OwnerResult;
};
