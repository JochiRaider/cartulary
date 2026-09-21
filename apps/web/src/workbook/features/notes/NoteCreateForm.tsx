import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
} from "@cartulary/ui-contracts";
import { useId, useLayoutEffect, useRef, useSyncExternalStore } from "react";
import {
  workbookFormInputStyle as fieldStyle,
  workbookFormActionsStyle,
  workbookFormFieldStackStyle,
  workbookFormFieldsStyle,
  workbookFormGroupStyle,
  workbookFormHeadingStyle,
  workbookFormMessageStyle,
} from "../../components/workbookFormStyles";
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
      style={workbookFormFieldsStyle}
    >
      <p style={workbookFormMessageStyle}>
        Closing keeps unfinished work in this session. Only the explicit create
        action saves it.
      </p>
      <fieldset
        disabled={owner.busy || !owner.canSubmit()}
        style={workbookFormGroupStyle}
      >
        <legend style={workbookFormHeadingStyle}>Create Note</legend>
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
            style={workbookFormFieldStackStyle}
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
            targetKey={`note:${state.draft.id}`}
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
      <div style={workbookFormActionsStyle}>
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
