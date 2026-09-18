import { useMemo, useState } from "react";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorTechnicalDetails } from "../inspector/presentation/WorkbookInspectorFeedback";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookBatchReceipt } from "../runtime/workbookBatchOperation";
import { inputStyle } from "./workbookGridControlStyles";

export type BatchReviewRecord = {
  readonly recordId: string;
  readonly label: string;
};
const pageSize = 20;

/** Disposable choices over the owner's receipt, never another retained inventory. */
export function WorkbookBatchRecordChoices({
  receipt,
  onReview,
}: {
  readonly receipt: WorkbookBatchReceipt;
  readonly onReview: (record: BatchReviewRecord) => void;
}) {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(0);
  const records = useMemo(() => {
    const seen = new Set<string>();
    return receipt.rows.flatMap((row, index) => {
      if (seen.has(row.record_id)) return [];
      seen.add(row.record_id);
      const synopsis = row.cells["timeline.activity_synopsis_text"]?.value;
      const text =
        typeof synopsis === "string"
          ? synopsis.trim().replace(/\s+/gu, " ")
          : "";
      const label = text
        ? `${text.slice(0, 120)}${text.length > 120 ? "…" : ""}`
        : `Timeline record ${index + 1}`;
      return [{ recordId: row.record_id, label }];
    });
  }, [receipt]);
  const filtered = records.filter((record) =>
    record.label
      .toLocaleLowerCase()
      .includes(search.trim().toLocaleLowerCase()),
  );
  const last = Math.max(0, Math.ceil(filtered.length / pageSize) - 1);
  const current = Math.min(page, last);
  const choices = filtered.slice(current * pageSize, (current + 1) * pageSize);
  if (
    receipt.viewSchemaId !== timelineViewSchemaId ||
    !receipt.changeSetId ||
    !records.length
  )
    return null;
  return (
    <section
      aria-label="Returned Timeline records"
      style={{ display: "grid", gap: "var(--ct-spacing-sm)", minInlineSize: 0 }}
    >
      <h3>Review an affected record</h3>
      <p>
        These labels describe records returned when the batch completed. History
        shows current state. The change may also affect other records.
      </p>
      <label>
        Find a returned record
        <input
          value={search}
          onChange={(event) => {
            setSearch(event.target.value);
            setPage(0);
          }}
          style={{
            ...inputStyle,
            inlineSize: "100%",
            minInlineSize: 0,
            boxSizing: "border-box",
          }}
        />
      </label>
      <p role="status">
        {filtered.length
          ? `Showing ${current * pageSize + 1}–${current * pageSize + choices.length} of ${filtered.length} returned records`
          : "No returned records match"}
      </p>
      <ul
        style={{
          margin: 0,
          paddingInlineStart: "var(--ct-spacing-lg)",
          overflowWrap: "anywhere",
        }}
      >
        {choices.map((record) => (
          <li
            key={record.recordId}
            style={{ marginBlockEnd: "var(--ct-spacing-sm)" }}
          >
            <span>{record.label}</span>{" "}
            <WorkbookInspectorActionButton
              aria-label={`Review this change: ${record.label}`}
              onClick={() => onReview(record)}
            >
              Review this change
            </WorkbookInspectorActionButton>
            <WorkbookInspectorTechnicalDetails
              fields={[{ label: "Record ID", value: record.recordId }]}
            />
          </li>
        ))}
      </ul>
      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          gap: "var(--ct-spacing-sm)",
        }}
      >
        <WorkbookInspectorActionButton
          disabled={current === 0}
          onClick={() => setPage(current - 1)}
        >
          Previous records
        </WorkbookInspectorActionButton>
        <WorkbookInspectorActionButton
          disabled={current >= last}
          onClick={() => setPage(current + 1)}
        >
          Next records
        </WorkbookInspectorActionButton>
      </div>
    </section>
  );
}
