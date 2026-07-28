import type {
  ApplyImportSessionRequest,
  ApplyImportSessionResponse,
  CancelJobRequest,
  CancelJobResponse,
  CreateImportSessionResponse,
  ExtensionMappingPreviewRequest,
  GetImportSessionResponse,
  GetImportUnitPreviewResponse,
  GetImportUnitResponse,
  GetJobResponse,
  ImportSourceColumnMapping,
  ListImportUnitsResponse,
  PreviewImportUnitExtensionMappingResponse,
  PutImportUnitMappingRequest,
  PutImportUnitMappingResponse,
  SelectImportUnitRequest,
  SelectImportUnitResponse,
  SkipImportUnitRequest,
  SkipImportUnitResponse,
} from "@cartulary/protocol-ts";

export type {
  ApplyImportSessionRequest,
  ApplyImportSessionResponse,
  CancelJobRequest,
  CancelJobResponse,
  CreateImportSessionResponse,
  ExtensionMappingPreviewRequest,
  GetImportSessionResponse,
  GetImportUnitPreviewResponse,
  GetImportUnitResponse,
  GetJobResponse,
  ImportSourceColumnMapping,
  ListImportUnitsResponse,
  PreviewImportUnitExtensionMappingResponse,
  PutImportUnitMappingRequest,
  PutImportUnitMappingResponse,
  SelectImportUnitRequest,
  SelectImportUnitResponse,
  SkipImportUnitRequest,
  SkipImportUnitResponse,
};

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
