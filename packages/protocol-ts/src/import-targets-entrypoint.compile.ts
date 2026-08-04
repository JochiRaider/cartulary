import {
  type ImportTargetFrontendDisposition,
  type ImportTargetFrontendProjection,
  type ImportTargetFrontendRow,
  importTargetRegistry,
} from "@cartulary/protocol-ts/import-targets";

export const typedImportTargetRegistry: ImportTargetFrontendProjection =
  importTargetRegistry;
export const firstImportTarget: ImportTargetFrontendRow =
  importTargetRegistry.targets[0];
export const firstImportTargetDisposition: ImportTargetFrontendDisposition =
  firstImportTarget.public_projection_disposition;
