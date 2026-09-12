import {
  decisionsViewSchemaId,
  getViewContract,
  partiesViewSchemaId,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { boundedRead } from "../../../services/asyncObservation";
import { WorkbookRecordCandidatePicker } from "../../components/WorkbookRecordCandidatePicker";
import { useWorkbookCandidates } from "../../hooks/useWorkbookCandidates";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import {
  type ContextualCreateDraft,
  contextualReferenceIds,
  contextualReferenceKind,
} from "./contextualCreateModel";
import type { ContextualCreateReader } from "./contextualCreateOperation";

export function ContextualReferenceControl({
  draft,
  field,
  reader,
  revision,
  onChange,
}: {
  readonly draft: ContextualCreateDraft;
  readonly field: ViewFieldContract;
  readonly reader: ContextualCreateReader;
  readonly revision: number;
  readonly onChange: (
    value: string,
    labels: Readonly<Record<string, string>>,
  ) => void;
}) {
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const ids = contextualReferenceIds(draft.values[field.fieldKey] ?? "");
  const close = () => {
    setOpen(false);
    trigger.current?.focus({ preventScroll: true });
  };
  return (
    <div style={{ display: "grid", gap: "0.5rem", minWidth: 0 }}>
      <span>{field.label}</span>
      {ids.length ? (
        <ul>
          {ids.map((id) => (
            <li key={id} style={{ overflowWrap: "anywhere" }}>
              {draft.labels[id] ??
                (contextualReferenceKind(field) === "members" &&
                id === draft.actorId
                  ? "Current actor"
                  : id)}
              {draft.seeds[field.fieldKey]?.split("\n").includes(id)
                ? " (source context)"
                : ""}{" "}
              <WorkbookInspectorActionButton
                tone="secondary"
                type="button"
                aria-label={`Remove ${field.label} ${draft.labels[id] ?? id}`}
                onClick={() =>
                  onChange(ids.filter((item) => item !== id).join("\n"), {})
                }
              >
                Remove
              </WorkbookInspectorActionButton>
              <details>
                <summary>Reference ID</summary>
                <span>{id}</span>
              </details>
            </li>
          ))}
        </ul>
      ) : (
        <span>
          {contextualReferenceKind(field) === "members"
            ? "Uses the current actor as owner when unset."
            : "No references selected."}
        </span>
      )}
      <WorkbookInspectorActionButton
        tone="secondary"
        ref={trigger}
        type="button"
        onClick={() => setOpen(true)}
      >
        Choose {field.label}
      </WorkbookInspectorActionButton>
      {open ? (
        <ReferencePicker
          key={field.fieldKey}
          draft={draft}
          field={field}
          reader={reader}
          revision={revision}
          onCancel={close}
          onApply={(value, labels) => {
            onChange(value, labels);
            close();
          }}
        />
      ) : null}
    </div>
  );
}
function ReferencePicker({
  draft,
  field,
  reader,
  revision,
  onApply,
  onCancel,
}: {
  readonly draft: ContextualCreateDraft;
  readonly field: ViewFieldContract;
  readonly reader: ContextualCreateReader;
  readonly revision: number;
  readonly onApply: (
    value: string,
    labels: Readonly<Record<string, string>>,
  ) => void;
  readonly onCancel: () => void;
}) {
  const kind = contextualReferenceKind(field);
  const initial =
    kind === "members"
      ? "incident_members"
      : kind === "parties"
        ? partiesViewSchemaId
        : kind === "decisions"
          ? decisionsViewSchemaId
          : draft.source.viewSchemaId;
  const [view, setView] = useState(initial);
  const [selected, setSelected] = useState(() =>
    contextualReferenceIds(draft.values[field.fieldKey] ?? ""),
  );
  const [views, setViews] = useState<readonly string[]>([]);
  const [viewError, setViewError] = useState(false);
  const inventoryRequest = useRef<AbortController | null>(null);
  const labels = useRef({ ...draft.labels });
  const labelRevision = useRef(revision);
  if (labelRevision.current !== revision) {
    labels.current = {};
    labelRevision.current = revision;
  }
  const picker = useRef<HTMLElement>(null);
  useEffect(() => {
    picker.current?.querySelector("select")?.focus({ preventScroll: true });
  }, []);
  const loadViews = useCallback(() => {
    if (kind !== "records") return;
    inventoryRequest.current?.abort();
    const controller = new AbortController();
    inventoryRequest.current = controller;
    setViewError(false);
    void boundedRead(
      (signal) => reader.availableViews(signal),
      controller.signal,
    )
      .then((value) => {
        if (!controller.signal.aborted) setViews(value);
      })
      .catch(() => {
        if (!controller.signal.aborted) setViewError(true);
      });
  }, [kind, reader]);
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
  const page = useWorkbookCandidates(read, query, `${view}:${revision}`);
  const candidates = [
    ...new Map(
      [
        ...selected.map((id) => ({
          recordId: id,
          displayText: labels.current[id] ?? id,
        })),
        ...(page.phase === "ready" && !page.stale ? page.candidates : []),
      ].map((candidate) => [candidate.recordId, candidate]),
    ).values(),
  ];
  return (
    <section
      ref={picker}
      tabIndex={-1}
      aria-label={`Choose ${field.label}`}
      onKeyDown={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          event.stopPropagation();
          onCancel();
        }
      }}
      style={{
        border: "var(--ct-border-hairline)",
        padding: "0.5rem",
        display: "grid",
        gap: "0.5rem",
      }}
    >
      {kind === "records" ? (
        <label>
          Reference surface
          <select
            aria-label="Reference surface"
            value={view}
            onChange={(event) => setView(event.currentTarget.value)}
            style={{
              width: "100%",
              minWidth: 0,
              boxSizing: "border-box",
              borderRadius: "var(--ct-component-text-input-rounded)",
              border: "var(--ct-component-text-input-border)",
              background: "var(--ct-component-text-input-backgroundColor)",
              color: "var(--ct-component-text-input-textColor)",
              font: "inherit",
              padding: "0.65rem 0.75rem",
            }}
          >
            {[...new Set([initial, ...views])].map((id) => (
              <option key={id} value={id}>
                {getViewContract(id)?.title ?? id}
              </option>
            ))}
          </select>
        </label>
      ) : null}
      {viewError ? (
        <p role="alert">
          Reference surfaces could not be loaded.{" "}
          <WorkbookInspectorActionButton
            tone="secondary"
            type="button"
            onClick={loadViews}
          >
            Retry surfaces
          </WorkbookInspectorActionButton>
        </p>
      ) : null}
      {kind === "records" && !views.length && !viewError ? (
        <p role="status">Loading reference surfaces…</p>
      ) : null}
      {page.phase === "loading" ? (
        <p role="status">Loading references…</p>
      ) : null}
      {page.stale ? (
        <p role="status">
          Loaded references need refresh. Selected IDs are retained.
        </p>
      ) : null}
      {page.error ? (
        <p role="alert">
          {page.error}{" "}
          <WorkbookInspectorActionButton
            tone="secondary"
            type="button"
            onClick={() => void page.retry()}
          >
            Retry references
          </WorkbookInspectorActionButton>
        </p>
      ) : null}
      {page.phase === "ready" && !page.candidates.length ? (
        <p>No available references on this surface.</p>
      ) : null}
      <WorkbookRecordCandidatePicker
        candidates={candidates}
        disabled={page.phase !== "ready" || page.stale}
        label={field.label}
        selection={field.readKind === "collection" ? "multiple" : "single"}
        selectedRecordIds={selected}
        testId={`contextual-reference-${field.fieldKey}`}
        onSelectedRecordIdsChange={(ids) => {
          for (const candidate of page.candidates)
            labels.current[candidate.recordId] = candidate.displayText;
          setSelected(ids);
        }}
      />
      {page.hasMore ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          type="button"
          disabled={page.phase === "loading"}
          onClick={() => void page.loadMore()}
        >
          Load more references
        </WorkbookInspectorActionButton>
      ) : null}
      <WorkbookInspectorActionButton
        tone="secondary"
        type="button"
        disabled={
          page.phase !== "ready" ||
          page.stale ||
          (kind === "records" && (!views.includes(view) || viewError))
        }
        onClick={() => onApply(selected.join("\n"), labels.current)}
      >
        Apply references
      </WorkbookInspectorActionButton>
      <WorkbookInspectorActionButton
        tone="secondary"
        type="button"
        onClick={onCancel}
      >
        Cancel references
      </WorkbookInspectorActionButton>
    </section>
  );
}
