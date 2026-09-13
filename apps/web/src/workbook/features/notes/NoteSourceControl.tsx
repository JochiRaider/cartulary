import { getViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { boundedRead } from "../../../services/asyncObservation";
import { WorkbookRecordCandidatePicker } from "../../components/WorkbookRecordCandidatePicker";
import { useWorkbookCandidates } from "../../hooks/useWorkbookCandidates";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import {
  type NoteCreateReader,
  type NoteSource,
  noteSourceViews,
} from "./noteCreateModel";

export function NoteSourceControl({
  source,
  reader,
  revision,
  disabled,
  onChange,
}: {
  readonly source: NoteSource | null;
  readonly reader: NoteCreateReader;
  readonly revision: number;
  readonly disabled: boolean;
  readonly onChange: (source: NoteSource | null) => void;
}) {
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const close = () => {
    setOpen(false);
    trigger.current?.focus({ preventScroll: true });
  };
  return (
    <div style={{ display: "grid", gap: "0.5rem", minWidth: 0 }}>
      <span style={{ overflowWrap: "anywhere" }}>
        Source:{" "}
        {source ? source.label || source.recordId : "None (unlinked Note)"}
      </span>
      <Button
        type="button"
        tone="secondary"
        ref={trigger}
        disabled={disabled}
        onClick={() => setOpen(true)}
      >
        Choose source
      </Button>
      {source ? (
        <Button
          type="button"
          tone="secondary"
          disabled={disabled}
          onClick={() => onChange(null)}
        >
          Clear source
        </Button>
      ) : null}
      {open && !disabled ? (
        <SourcePicker
          source={source}
          reader={reader}
          revision={revision}
          onCancel={close}
          onApply={(value) => {
            onChange(value);
            close();
          }}
        />
      ) : null}
    </div>
  );
}
function SourcePicker({
  source,
  reader,
  revision,
  onCancel,
  onApply,
}: {
  readonly source: NoteSource | null;
  readonly reader: NoteCreateReader;
  readonly revision: number;
  readonly onCancel: () => void;
  readonly onApply: (source: NoteSource | null) => void;
}) {
  const [view, setView] = useState(source?.viewSchemaId ?? noteSourceViews[0]);
  const [selected, setSelected] = useState(source);
  const priorRevision = useRef(revision);
  useEffect(() => {
    if (priorRevision.current === revision) return;
    priorRevision.current = revision;
    // Refresh discovery without replacing the analyst's staged identity or sheet.
    setSelected((current) => (current ? { ...current, label: "" } : null));
  }, [revision]);
  const [inventory, setInventory] = useState<{
    views: readonly string[];
    phase: "loading" | "ready" | "failed";
  }>({ views: [], phase: "loading" });
  const inventoryRequest = useRef<AbortController | null>(null);
  const root = useRef<HTMLElement>(null);
  useEffect(() => {
    root.current?.querySelector("select")?.focus({ preventScroll: true });
  }, []);
  const loadViews = useCallback(() => {
    inventoryRequest.current?.abort();
    const controller = new AbortController();
    inventoryRequest.current = controller;
    setInventory({ views: [], phase: "loading" });
    void boundedRead(
      (signal) => reader.availableViews(signal),
      controller.signal,
    )
      .then((views) => {
        if (!controller.signal.aborted)
          setInventory({
            views: views.filter((id) => noteSourceViews.includes(id)),
            phase: "ready",
          });
      })
      .catch(() => {
        if (!controller.signal.aborted)
          setInventory({ views: [], phase: "failed" });
      });
  }, [reader]);
  useEffect(() => {
    loadViews();
    return () => inventoryRequest.current?.abort();
  }, [loadViews]);
  const query = useMemo(emptyWorkbookQueryState, []);
  const read = useCallback(
    (input: {
      cursor: string | null;
      signal: AbortSignal;
      queryState: typeof query;
    }) => reader.page({ ...input, viewSchemaId: view }),
    [reader, view],
  );
  const page = useWorkbookCandidates(
    read,
    query,
    `${view}:${revision}`,
    inventory.phase === "ready" && inventory.views.includes(view),
  );
  const ready =
    inventory.phase === "ready" &&
    inventory.views.includes(view) &&
    page.phase === "ready" &&
    !page.stale;
  const candidates = [
    ...new Map(
      [
        ...(selected
          ? [
              {
                recordId: selected.recordId,
                displayText: selected.label || selected.recordId,
              },
            ]
          : []),
        ...(ready ? page.candidates : []),
      ].map((item) => [item.recordId, item]),
    ).values(),
  ];
  return (
    <section
      ref={root}
      aria-label="Choose Note source"
      style={{
        display: "grid",
        gap: "0.5rem",
        border: "var(--ct-border-hairline)",
        padding: "0.5rem",
        minWidth: 0,
      }}
      onKeyDown={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          event.stopPropagation();
          onCancel();
        }
      }}
    >
      <label>
        Source sheet
        <select
          aria-label="Source sheet"
          value={view}
          onChange={(event) => setView(event.currentTarget.value)}
          style={{ width: "100%", minWidth: 0 }}
        >
          {[...new Set([view, ...inventory.views])].map((id) => (
            <option key={id} value={id}>
              {getViewContract(id)?.title ?? id}
            </option>
          ))}
        </select>
      </label>
      {inventory.phase === "loading" ? (
        <p role="status">Loading source sheets…</p>
      ) : null}
      {inventory.phase === "failed" ? (
        <p role="alert">
          Source sheets could not be loaded.{" "}
          <Button tone="secondary" type="button" onClick={loadViews}>
            Retry sheets
          </Button>
        </p>
      ) : null}
      {inventory.phase === "ready" && !inventory.views.length ? (
        <p>No available source sheets.</p>
      ) : null}
      {inventory.views.includes(view) && page.phase === "loading" ? (
        <p role="status">Loading sources…</p>
      ) : null}
      {page.stale ? (
        <p role="status">
          Sources need refresh. Your selected source is retained.
        </p>
      ) : null}
      {page.error ? (
        <p role="alert">
          {page.error}{" "}
          <Button
            tone="secondary"
            type="button"
            onClick={() => void page.retry()}
          >
            Retry sources
          </Button>
        </p>
      ) : null}
      {ready && !page.candidates.length ? (
        <p>No available sources on this sheet.</p>
      ) : null}
      <WorkbookRecordCandidatePicker
        candidates={candidates}
        disabled={!ready}
        label="Note source"
        selection="single"
        selectedRecordIds={selected ? [selected.recordId] : []}
        testId="note-source-record"
        onSelectedRecordIdsChange={(ids) => {
          if (!ids.length) {
            setSelected(null);
            return;
          }
          const candidate = page.candidates.find(
            (item) => item.recordId === ids[0],
          );
          if (candidate?.row)
            setSelected({
              recordId: candidate.recordId,
              viewSchemaId: candidate.viewSchemaId,
              rowVersion: candidate.row.row_version,
              label: candidate.displayText,
            });
        }}
      />
      {page.hasMore ? (
        <Button
          tone="secondary"
          type="button"
          disabled={!ready}
          onClick={() => void page.loadMore()}
        >
          Load more sources
        </Button>
      ) : null}
      <Button
        tone="secondary"
        type="button"
        disabled={!ready}
        onClick={() => onApply(selected)}
      >
        Apply source
      </Button>
      <Button tone="secondary" type="button" onClick={onCancel}>
        Cancel source
      </Button>
    </section>
  );
}
