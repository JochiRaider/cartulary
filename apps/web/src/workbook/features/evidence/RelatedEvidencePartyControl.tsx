import {
  partiesViewSchemaId,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { WorkbookRecordCandidatePicker } from "../../components/WorkbookRecordCandidatePicker";
import {
  useWorkbookCandidates,
  type WorkbookCandidateQuery,
} from "../../hooks/useWorkbookCandidates";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";

export function RelatedEvidencePartyControl({
  field,
  value,
  labels,
  reader,
  revision,
  errorId,
  onChange,
}: {
  readonly field: ViewFieldContract;
  readonly value: string;
  readonly labels: Readonly<Record<string, string>>;
  readonly reader: WorkbookAuthoringReadPort;
  readonly revision: number;
  readonly errorId?: string | undefined;
  readonly onChange: (
    value: string,
    labels: Readonly<Record<string, string>>,
  ) => void;
}) {
  const [open, setOpen] = useState(false),
    trigger = useRef<HTMLButtonElement>(null);
  const close = () => {
    setOpen(false);
    trigger.current?.focus({ preventScroll: true });
  };
  return (
    <div style={{ display: "grid", gap: "var(--ct-spacing-xs)" }}>
      <span>
        {field.label}:{" "}
        {value
          ? (labels[value] ?? "Selected Party (availability needs review)")
          : "None selected"}
      </span>
      {value ? (
        <Button type="button" tone="secondary" onClick={() => onChange("", {})}>
          Remove {field.label}
        </Button>
      ) : null}
      <Button
        ref={trigger}
        aria-describedby={errorId}
        type="button"
        tone="secondary"
        onClick={() => setOpen(true)}
      >
        Choose {field.label}
      </Button>
      {open ? (
        <PartyPicker
          label={field.label}
          fieldKey={field.fieldKey}
          value={value}
          labels={labels}
          reader={reader}
          revision={revision}
          onCancel={close}
          onApply={(id, names) => {
            onChange(id, names);
            close();
          }}
        />
      ) : null}
    </div>
  );
}
function PartyPicker({
  label,
  fieldKey,
  value,
  labels,
  reader,
  revision,
  onApply,
  onCancel,
}: {
  readonly label: string;
  readonly fieldKey: string;
  readonly value: string;
  readonly labels: Readonly<Record<string, string>>;
  readonly reader: WorkbookAuthoringReadPort;
  readonly revision: number;
  readonly onApply: (
    id: string,
    labels: Readonly<Record<string, string>>,
  ) => void;
  readonly onCancel: () => void;
}) {
  const [selected, setSelected] = useState<readonly string[]>(
    value ? [value] : [],
  );
  const query = useMemo(emptyWorkbookQueryState, []);
  const read = useCallback(
    (input: WorkbookCandidateQuery) =>
      reader.page({ ...input, viewSchemaId: partiesViewSchemaId }),
    [reader],
  );
  const page = useWorkbookCandidates(read, query, revision);
  const picker = useRef<HTMLElement>(null);
  const names = useRef({ ...labels }),
    priorRevision = useRef(revision);
  if (priorRevision.current !== revision) {
    names.current = {};
    priorRevision.current = revision;
  }
  useEffect(() => {
    picker.current?.querySelector("select")?.focus({ preventScroll: true });
  }, []);
  const candidates = [
    ...new Map(
      [
        ...selected.map((recordId) => ({
          recordId,
          displayText:
            names.current[recordId] ?? "Selected Party (verify availability)",
        })),
        ...(page.phase === "ready" && !page.stale ? page.candidates : []),
      ].map((item) => [item.recordId, item]),
    ).values(),
  ];
  return (
    <section
      ref={picker}
      aria-label={`Choose ${label}`}
      onKeyDown={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          event.stopPropagation();
          onCancel();
        }
      }}
      style={{
        border: "var(--ct-border-hairline)",
        padding: "var(--ct-spacing-sm)",
        minWidth: 0,
      }}
    >
      {page.phase === "loading" ? <p role="status">Loading Parties…</p> : null}
      {page.stale ? (
        <p role="status">
          Party availability changed. Selected references are retained.
        </p>
      ) : null}
      {page.error ? (
        <p role="alert">
          {page.error}{" "}
          <Button
            type="button"
            tone="secondary"
            onClick={() => void page.retry()}
          >
            Retry Parties
          </Button>
        </p>
      ) : null}
      {page.phase === "ready" && !page.candidates.length ? (
        <p>No available Parties.</p>
      ) : null}
      <WorkbookRecordCandidatePicker
        candidates={candidates}
        disabled={page.phase !== "ready" || page.stale}
        label={label}
        selection="single"
        selectedRecordIds={selected}
        testId={`related-evidence-reference-${fieldKey}`}
        onSelectedRecordIdsChange={(ids) => {
          for (const item of page.candidates)
            names.current[item.recordId] = item.displayText;
          setSelected(ids);
        }}
      />
      {page.hasMore ? (
        <Button
          type="button"
          tone="secondary"
          disabled={page.phase === "loading"}
          onClick={() => void page.loadMore()}
        >
          Load more Parties
        </Button>
      ) : null}
      <Button
        type="button"
        tone="secondary"
        disabled={page.phase !== "ready" || page.stale}
        onClick={() => onApply(selected[0] ?? "", names.current)}
      >
        Apply Party
      </Button>
      <Button type="button" tone="secondary" onClick={onCancel}>
        Cancel Party selection
      </Button>
    </section>
  );
}
