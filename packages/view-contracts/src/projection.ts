import { viewContractProjection } from "./generated/view-contract-projection.js";
import type {
  ViewContract,
  ViewFieldContract,
  WorkbookSurfaceContract,
} from "./types.js";

function truthMap(values: readonly string[]): Readonly<Record<string, true>> {
  return Object.freeze(
    Object.fromEntries(values.map((value) => [value, true])) as Record<
      string,
      true
    >,
  );
}

const projectedContracts: readonly ViewContract[] = Object.freeze(
  [...viewContractProjection]
    .sort((left, right) => left.viewSchemaId.localeCompare(right.viewSchemaId))
    .map((entry): ViewContract => {
      const fields = entry.fields as readonly ViewFieldContract[];
      const syntheticFilterFields =
        entry.syntheticFilterFields as readonly ViewFieldContract[];
      const fieldMap = Object.freeze(
        Object.fromEntries(
          [...fields, ...syntheticFilterFields].map((field) => [
            field.fieldKey,
            field,
          ]),
        ) as Record<string, ViewFieldContract>,
      );
      return Object.freeze({
        defaultHiddenFields: entry.defaultHiddenFields,
        defaultSort: entry.defaultSort,
        defaultVisibleFields: entry.defaultVisibleFields,
        fieldMap,
        fields,
        filterableFieldMap: truthMap(entry.filterFields),
        filterFields: entry.filterFields,
        groupableFieldMap: truthMap(entry.groupingFields),
        groupingFields: entry.groupingFields,
        inspectorConfig: entry.inspectorConfig,
        minimumCreateFieldSets: entry.minimumCreateFieldSets,
        permitsZeroFieldCreate: entry.permitsZeroFieldCreate,
        requiredReferencePackKeys: entry.requiredReferencePackKeys,
        sortableFieldMap: truthMap(entry.sortFields),
        sortFields: entry.sortFields,
        sortNullOrder: entry.sortNullOrder,
        surfaceKind: entry.surfaceKind,
        technicalFields: entry.technicalFields,
        title: entry.title,
        viewSchemaId: entry.viewSchemaId,
      });
    }),
);

const projectedContractIndex = new Map(
  projectedContracts.map((contract) => [contract.viewSchemaId, contract]),
);

const projectedWorkbookSurfaces: readonly WorkbookSurfaceContract[] =
  Object.freeze(
    viewContractProjection.map((entry): WorkbookSurfaceContract => {
      const contract = projectedContractIndex.get(entry.viewSchemaId);
      if (contract === undefined) {
        throw new Error(
          `Generated view-contract projection is missing ${entry.viewSchemaId}`,
        );
      }
      return Object.freeze({
        contract,
        requiredReferencePackKeys: entry.requiredReferencePackKeys,
        sourceRecordTypes: entry.sourceRecordTypes,
        surfaceKind: entry.surfaceKind,
        surfaceStatus: entry.surfaceStatus,
        title: entry.title,
        viewSchemaId: entry.viewSchemaId,
      });
    }),
  );

export function listProjectedViewContracts(): readonly ViewContract[] {
  return projectedContracts;
}

export function listProjectedWorkbookSurfaceContracts(): readonly WorkbookSurfaceContract[] {
  return projectedWorkbookSurfaces;
}
