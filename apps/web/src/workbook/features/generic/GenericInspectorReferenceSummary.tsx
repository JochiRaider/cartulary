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
      <dl>
        {fields.map((field) => (
          <div key={field.fieldKey}>
            <dt>{field.label}</dt>
            <dd style={{ marginInlineStart: 0, overflowWrap: "anywhere" }}>
              {Object.hasOwn(row.cells, field.fieldKey)
                ? genericCellLabel(row.cells[field.fieldKey]?.value) || "None"
                : "Not available in this row."}
            </dd>
            {field.patchWritable ? (
              <Button
                tone="secondary"
                disabled={!canEdit}
                onClick={() => onEdit(field.fieldKey)}
              >
                Edit {field.label.toLowerCase()}
              </Button>
            ) : null}
          </div>
        ))}
      </dl>
    </>
  );
}
