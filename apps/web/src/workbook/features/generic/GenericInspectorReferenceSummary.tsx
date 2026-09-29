import {
  getReferenceFieldContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { genericCellLabel } from "../../models/genericWorkbookModel";
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
  readonly onEdit: (fieldKey: string) => void;
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
              {field.patchWritable ? (
                <dd style={{ marginInlineStart: 0, minInlineSize: 0 }}>
                  <Button
                    tone="quiet"
                    disabled={!canEdit || !cell}
                    style={{ maxInlineSize: "100%", overflowWrap: "anywhere" }}
                    onClick={() => onEdit(field.fieldKey)}
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
