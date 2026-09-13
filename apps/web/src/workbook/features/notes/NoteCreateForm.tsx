import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
} from "@cartulary/ui-contracts";
import { useId, useLayoutEffect, useRef, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { NoteSourceControl } from "./NoteSourceControl";
import { noteCreateView } from "./noteCreateModel";
import type { WorkbookNoteCreateOwner } from "./WorkbookNoteCreateOwner";

export function NoteCreateForm({
  owner,
  attachment,
  onSubmit,
}: {
  readonly owner: WorkbookNoteCreateOwner;
  readonly attachment: symbol;
  readonly onSubmit: () => void;
}) {
  const state = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const id = useId();
  const title = useRef<HTMLInputElement>(null);
  const root = useRef<HTMLElement>(null);
  const focused = useRef(false);
  const trigger = useRef(document.activeElement);
  useLayoutEffect(() => {
    title.current?.focus({ preventScroll: true });
    const container = root.current;
    return () => {
      const element = trigger.current;
      if (
        (container?.contains(document.activeElement) ||
          (focused.current && document.activeElement === document.body)) &&
        element instanceof HTMLElement &&
        element.isConnected &&
        owner.getSnapshot().authority
      )
        element.focus({ preventScroll: true });
    };
  }, [owner]);
  const returnFocus = () => {
    const element = trigger.current;
    if (
      element instanceof HTMLElement &&
      element.isConnected &&
      owner.getSnapshot().authority
    )
      element.focus({ preventScroll: true });
  };
  if (!state.draft || state.attachment !== attachment) return null;
  const reader = owner.getReader();
  return (
    <section
      ref={root}
      onFocusCapture={() => {
        focused.current = true;
      }}
      onBlurCapture={(event) => {
        if (
          event.relatedTarget &&
          !event.currentTarget.contains(event.relatedTarget)
        )
          focused.current = false;
      }}
      aria-label="Create Note"
      style={{ display: "grid", gap: "0.75rem", minWidth: 0 }}
    >
      <fieldset
        disabled={owner.busy || !owner.canSubmit()}
        style={{
          border: 0,
          padding: 0,
          margin: 0,
          display: "grid",
          gap: "0.75rem",
          minWidth: 0,
        }}
      >
        <legend>Create Note</legend>
        {(
          [
            ["note.title", "Title"],
            ["note.body", "Body"],
            ["note.tags", "Tags (one per line)"],
          ] as const
        ).map(([field, label]) => (
          <label
            key={field}
            htmlFor={`${id}-input-${field}`}
            style={{ display: "grid", gap: "0.25rem", minWidth: 0 }}
          >
            {label}
            {field === "note.title" ? (
              <input
                ref={title}
                id={`${id}-input-${field}`}
                aria-label={label}
                data-testid={genericCreateFieldTestId(field)}
                value={state.draft?.values[field] ?? ""}
                aria-invalid={!!state.errors[field]}
                aria-describedby={
                  state.errors[field] ? `${id}-${field}` : undefined
                }
                onChange={(event) =>
                  owner.update(field, event.currentTarget.value)
                }
                style={fieldStyle}
              />
            ) : (
              <textarea
                id={`${id}-input-${field}`}
                aria-label={label}
                data-testid={genericCreateFieldTestId(field)}
                rows={field === "note.body" ? 5 : 2}
                value={state.draft?.values[field] ?? ""}
                aria-invalid={!!state.errors[field]}
                aria-describedby={
                  state.errors[field] ? `${id}-${field}` : undefined
                }
                onChange={(event) =>
                  owner.update(field, event.currentTarget.value)
                }
                style={fieldStyle}
              />
            )}
            {state.errors[field] ? (
              <span id={`${id}-${field}`} role="alert">
                {state.errors[field]}
              </span>
            ) : null}
          </label>
        ))}
        {reader ? (
          <NoteSourceControl
            source={state.draft.source}
            reader={reader}
            revision={state.candidateRevision}
            disabled={owner.busy || !owner.canSubmit()}
            onChange={(source) => owner.changeSource(source)}
          />
        ) : null}
      </fieldset>
      {state.message ? <p role="status">{state.message}</p> : null}
      {state.needsReview ? (
        <Button
          tone="secondary"
          type="button"
          disabled={owner.busy || !owner.canSubmit()}
          onClick={() => void owner.review()}
        >
          Review source and access
        </Button>
      ) : null}
      <div style={{ display: "flex", flexWrap: "wrap", gap: "0.5rem" }}>
        <Button
          tone="primary"
          type="button"
          data-testid={genericCreateSubmitTestId(noteCreateView)}
          disabled={owner.busy || !owner.canSubmit() || state.needsReview}
          onClick={onSubmit}
        >
          Create Note
        </Button>
        <Button
          tone="secondary"
          type="button"
          onClick={() => {
            owner.detach(attachment);
            returnFocus();
          }}
        >
          Close draft
        </Button>
        <Button
          tone="secondary"
          type="button"
          disabled={owner.busy}
          onClick={() => {
            owner.discard();
            returnFocus();
          }}
        >
          Discard draft
        </Button>
      </div>
    </section>
  );
}
const fieldStyle = {
  width: "100%",
  minWidth: 0,
  boxSizing: "border-box",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
  background: "var(--ct-component-text-input-backgroundColor)",
  border: "var(--ct-component-text-input-border)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  padding: "0.5rem",
} as const;
