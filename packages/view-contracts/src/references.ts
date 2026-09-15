import { projectedReferenceRegistry } from "./projection.js";

export type ReferenceIdentityKind = "record" | "party" | "incident_member";
export type ReferenceFieldContract = Readonly<{
  viewSchemaId: string;
  fieldKey: string;
  kind: "direct" | "collection";
  identityKind: ReferenceIdentityKind;
  targetRecordTypes: readonly string[];
  targetViewSchemaIds: readonly string[];
  excludeSource: boolean;
  addOperation: "add_record_ref" | "add_party_ref" | null;
  removeOperation: "remove_record_ref" | "remove_party_ref" | null;
}>;

const fields: readonly ReferenceFieldContract[] = Object.freeze(
  projectedReferenceRegistry.fields.map((field): ReferenceFieldContract => {
    const domain =
      projectedReferenceRegistry.target_domains[field.target_domain];
    return Object.freeze({
      viewSchemaId: field.view_schema_id,
      fieldKey: field.field_key,
      kind: field.kind,
      identityKind: domain.identity_kind,
      targetRecordTypes: domain.record_types,
      targetViewSchemaIds: domain.view_schema_ids,
      excludeSource: field.exclude_source,
      addOperation: "add_operation" in field ? field.add_operation : null,
      removeOperation:
        "remove_operation" in field ? field.remove_operation : null,
    });
  }),
);

export function listReferenceFieldContracts(): readonly ReferenceFieldContract[] {
  return fields;
}

export function getReferenceFieldContract(
  viewSchemaId: string,
  fieldKey: string,
): ReferenceFieldContract | undefined {
  return fields.find(
    (field) =>
      field.viewSchemaId === viewSchemaId && field.fieldKey === fieldKey,
  );
}
