import {
  getReferenceFieldContract,
  type ReferenceFieldContract,
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import { workbookFormInputStyle } from "../../components/workbookFormStyles";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { genericCellLabel } from "../../models/genericWorkbookModel";
import { useWorkbookWorkbench } from "../../navigation/WorkbookWorkbenchContext";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

/** Reference membership and permitted target kinds come from authored fields.
 * These controls select the existing ordinary edit owner; they never invent a
 * writable field or infer an empty relationship list from missing projection. */
export function GenericInspectorReferenceSummary({
  contract,
  row,
  evidenceOnly = false,
  canEdit,
  onEdit,
}: {
  readonly contract: ViewContract;
  readonly row: WorkbookQueryRow;
  readonly evidenceOnly?: boolean;
  readonly canEdit: boolean;
  readonly onEdit: (fieldKey: string, trigger: HTMLElement) => void;
}) {
  const fields = contract.fields.filter((field) => {
    const reference = getReferenceFieldContract(
      contract.viewSchemaId,
      field.fieldKey,
    );
    return (
      reference &&
      (!evidenceOnly || reference.targetRecordTypes.includes("evidence"))
    );
  });
  if (!fields.length)
    return (
      <p>
        {evidenceOnly
          ? "This record has no evidence reference field. Evidence linked through other records remains with those records."
          : "This record has no reference fields. Links from other records are managed on the referring record."}
      </p>
    );
  return (
    <>
      {evidenceOnly ? (
        <p>Evidence can be linked through these reference fields.</p>
      ) : null}
      <dl style={{ minInlineSize: 0 }}>
        {fields.map((field) => {
          const cell = row.cells[field.fieldKey];
          const action = field.readKind === "collection" ? "Manage" : "Edit";
          return (
            <div key={field.fieldKey} style={{ minInlineSize: 0 }}>
              <dt style={{ overflowWrap: "anywhere" }}>{field.label}</dt>
              <dd style={{ marginInlineStart: 0, overflowWrap: "anywhere" }}>
                {cell?.value === undefined
                  ? "Not available in this row."
                  : genericCellLabel(cell.value)}
              </dd>
              <ReferenceDestinations
                reference={getReferenceFieldContract(
                  contract.viewSchemaId,
                  field.fieldKey,
                )}
                value={cell?.value}
              />
              {field.patchWritable ? (
                <dd style={{ marginInlineStart: 0, minInlineSize: 0 }}>
                  <Button
                    tone="quiet"
                    disabled={!canEdit || !cell}
                    style={{ maxInlineSize: "100%", overflowWrap: "anywhere" }}
                    onClick={(event) =>
                      onEdit(field.fieldKey, event.currentTarget)
                    }
                  >
                    {action} {field.label}
                  </Button>
                </dd>
              ) : null}
            </div>
          );
        })}
      </dl>
    </>
  );
}

/** The read contract does not supply a target schema for polymorphic links.
 * Require an explicit admitted surface; never parse a display label as identity
 * or scan candidate surfaces to infer a record's type. */
function ReferenceDestinations({
  reference,
  value,
}: {
  readonly reference: ReferenceFieldContract | undefined;
  readonly value: unknown;
}) {
  const workbench = useWorkbookWorkbench();
  if (!workbench || !reference || reference.identityKind === "incident_member")
    return null;
  const items: unknown[] =
    reference.kind === "direct"
      ? [value]
      : value &&
          typeof value === "object" &&
          "items" in value &&
          Array.isArray(value.items)
        ? value.items
        : [];
  const ids = [
    ...new Set(
      items.flatMap((item) => {
        const id =
          typeof item === "string"
            ? item
            : item && typeof item === "object" && "linked_record_id" in item
              ? item.linked_record_id
              : null;
        return typeof id === "string" &&
          /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/iu.test(
            id,
          )
          ? [id]
          : [];
      }),
    ),
  ];
  return ids.map((recordId) => (
    <dd
      key={recordId}
      style={{ marginInlineStart: 0 }}
      data-workbook-navigation="true"
    >
      {reference.targetViewSchemaIds.length === 1 ? (
        <Button
          tone="quiet"
          onClick={() => {
            const id = reference.targetViewSchemaIds[0];
            if (id)
              workbench.open({
                sheetRef: { kind: "view_schema", id },
                recordId,
              });
          }}
        >
          Open{" "}
          {requireViewContract(reference.targetViewSchemaIds[0] ?? "").title}{" "}
          record
        </Button>
      ) : (
        <label>
          Open linked record {recordId}
          <select
            style={workbookFormInputStyle}
            value=""
            onChange={(event) => {
              const id = event.currentTarget.value;
              if (reference.targetViewSchemaIds.includes(id))
                workbench.open({
                  sheetRef: { kind: "view_schema", id },
                  recordId,
                });
            }}
          >
            <option value="">Choose its surface…</option>
            {reference.targetViewSchemaIds.map((id) => (
              <option key={id} value={id}>
                {requireViewContract(id).title}
              </option>
            ))}
          </select>
        </label>
      )}
    </dd>
  ));
}
