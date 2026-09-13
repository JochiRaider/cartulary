import { getViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { boundedRead } from "../../services/asyncObservation";
import { useWorkbookCandidates } from "../hooks/useWorkbookCandidates";
import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import type {
  WorkbookAuthoringCandidate,
  WorkbookAuthoringReadPort,
} from "../ports/WorkbookAuthoringReadPort";
import { WorkbookRecordCandidatePicker } from "./WorkbookRecordCandidatePicker";

type Props = Readonly<{
  label: string;
  errorId?: string | undefined;
  required?: boolean | undefined;
  testId: string;
  views: readonly string[];
  multiple: boolean;
  selected: readonly WorkbookAuthoringCandidate[];
  reader: Pick<WorkbookAuthoringReadPort, "page" | "availableViews">;
  revision: number;
  disabled: boolean;
  onApply: (selected: readonly WorkbookAuthoringCandidate[]) => void;
}>;
/** Neutral staged identity picker. Page membership never owns retained selection. */
export function WorkbookAuthoringReferenceControl(props: Props) {
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const close = () => {
    setOpen(false);
    trigger.current?.focus({ preventScroll: true });
  };
  return (
    <div style={groupStyle}>
      <span>{props.label}</span>
      {props.selected.length ? (
        <ul style={{ margin: 0, paddingInlineStart: "var(--ct-spacing-lg)" }}>
          {props.selected.map((item) => (
            <li key={item.recordId} style={{ overflowWrap: "anywhere" }}>
              {item.displayText || "Selected reference"}{" "}
              <Button
                tone="secondary"
                type="button"
                disabled={props.disabled}
                aria-label={`Remove ${props.label} ${item.displayText || "reference"}`}
                onClick={() =>
                  props.onApply(
                    props.selected.filter((i) => i.recordId !== item.recordId),
                  )
                }
              >
                Remove
              </Button>
            </li>
          ))}
        </ul>
      ) : (
        <span>No {props.label.toLowerCase()} selected.</span>
      )}
      <Button
        ref={trigger}
        data-create-required={props.required}
        aria-invalid={!!props.errorId}
        aria-describedby={props.errorId}
        tone="secondary"
        type="button"
        disabled={props.disabled}
        onClick={() => setOpen(true)}
      >
        Choose {props.label.toLowerCase()}
      </Button>
      {open && !props.disabled ? (
        <Picker
          {...props}
          onCancel={close}
          onApply={(items) => {
            props.onApply(items);
            close();
          }}
        />
      ) : null}
    </div>
  );
}
function Picker({
  onCancel,
  ...props
}: Props & { readonly onCancel: () => void }) {
  const [view, setView] = useState(
    props.selected[0]?.viewSchemaId ?? props.views[0] ?? "",
  );
  const [selected, setSelected] = useState(props.selected);
  const [inventory, setInventory] = useState<{
    phase: "loading" | "ready" | "failed";
    views: readonly string[];
  }>({ phase: "loading", views: [] });
  const request = useRef<AbortController | null>(null);
  const root = useRef<HTMLElement>(null);
  const allowed = JSON.stringify(props.views);
  const loadViews = useCallback(() => {
    request.current?.abort();
    const controller = new AbortController();
    request.current = controller;
    setInventory({ phase: "loading", views: [] });
    void boundedRead(
      (signal) => props.reader.availableViews(signal),
      controller.signal,
    )
      .then((views) => {
        if (!controller.signal.aborted)
          setInventory({
            phase: "ready",
            views: (JSON.parse(allowed) as string[]).filter(
              (v) => v === "incident_members" || views.includes(v),
            ),
          });
      })
      .catch(() => {
        if (!controller.signal.aborted)
          setInventory({ phase: "failed", views: [] });
      });
  }, [props.reader, allowed]);
  useEffect(() => {
    loadViews();
    return () => request.current?.abort();
  }, [loadViews]);
  useEffect(() => {
    root.current
      ?.querySelector<HTMLElement>("select,button")
      ?.focus({ preventScroll: true });
  }, []);
  const query = useMemo(emptyWorkbookQueryState, []);
  const read = useCallback(
    (input: {
      cursor: string | null;
      signal: AbortSignal;
      queryState: typeof query;
    }) => props.reader.page({ ...input, viewSchemaId: view }),
    [props.reader, view],
  );
  const page = useWorkbookCandidates(
    read,
    query,
    `${view}:${props.revision}`,
    inventory.phase === "ready" && inventory.views.includes(view),
  );
  const ready =
    inventory.phase === "ready" &&
    inventory.views.includes(view) &&
    page.phase === "ready" &&
    !page.stale;
  const candidates = [
    ...new Map(
      [...selected, ...(ready ? page.candidates : [])].map((item) => [
        item.recordId,
        item,
      ]),
    ).values(),
  ];
  return (
    <section
      ref={root}
      aria-label={`Choose ${props.label.toLowerCase()}`}
      style={{
        ...groupStyle,
        border: "var(--ct-border-hairline)",
        padding: "var(--ct-spacing-sm)",
      }}
      onKeyDown={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          event.stopPropagation();
          onCancel();
        }
      }}
    >
      {props.views.length > 1 ? (
        <label>
          Reference surface
          <select
            aria-label="Reference surface"
            style={fieldStyle}
            value={view}
            onChange={(event) => setView(event.currentTarget.value)}
          >
            {[...new Set([view, ...inventory.views])].map((id) => (
              <option key={id} value={id}>
                {getViewContract(id)?.title ?? "Members"}
              </option>
            ))}
          </select>
        </label>
      ) : null}
      {inventory.phase === "loading" ? (
        <p role="status">Loading reference surfaces…</p>
      ) : null}
      {inventory.phase === "failed" ? (
        <p role="alert">
          Reference surfaces could not be loaded.{" "}
          <Button tone="secondary" type="button" onClick={loadViews}>
            Retry surfaces
          </Button>
        </p>
      ) : null}
      {inventory.phase === "ready" && !inventory.views.includes(view) ? (
        <p role="status">
          This reference surface is unavailable. Your selection is retained.
        </p>
      ) : null}
      {inventory.views.includes(view) && page.phase === "loading" ? (
        <p role="status">Loading references…</p>
      ) : null}
      {page.stale ? (
        <p role="status">
          References need refresh. Selected identities are retained.
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
            Retry references
          </Button>
        </p>
      ) : null}
      {ready && !page.candidates.length ? (
        <p>No available references.</p>
      ) : null}
      <WorkbookRecordCandidatePicker
        candidates={candidates}
        disabled={!ready}
        label={props.label}
        testId={props.testId}
        selection={props.multiple ? "multiple" : "single"}
        selectedRecordIds={selected.map((item) => item.recordId)}
        onSelectedRecordIdsChange={(ids) =>
          setSelected(
            ids.flatMap((id) => {
              const candidate = candidates.find((item) => item.recordId === id);
              return candidate ? [candidate] : [];
            }),
          )
        }
      />
      {page.hasMore ? (
        <Button
          tone="secondary"
          type="button"
          disabled={!ready}
          onClick={() => void page.loadMore()}
        >
          Load more references
        </Button>
      ) : null}
      <Button
        tone="secondary"
        type="button"
        disabled={!ready}
        onClick={() => props.onApply(selected)}
      >
        Apply references
      </Button>
      <Button tone="secondary" type="button" onClick={onCancel}>
        Cancel references
      </Button>
    </section>
  );
}
const groupStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
} as const;
const fieldStyle = {
  width: "100%",
  minWidth: 0,
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
  background: "var(--ct-component-text-input-backgroundColor)",
  border: "var(--ct-component-text-input-border)",
} as const;
