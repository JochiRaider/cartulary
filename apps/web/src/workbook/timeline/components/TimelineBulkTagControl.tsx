import {
  type CSSProperties,
  useId,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { useWorkbookCommand } from "../../commands/WorkbookCommands";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import type { useTimelineBulkTagController } from "../bulk/useTimelineBulkTagController";

type Binding = ReturnType<typeof useTimelineBulkTagController>["controls"];

/** Leaf observation of retained authoring; typing never enters root composition. */
export function TimelineBulkTagControl({
  binding,
}: {
  readonly binding: Binding;
}) {
  const draft = useSyncExternalStore(
    binding.authoring.subscribe,
    binding.authoring.getSnapshot,
  );
  const [focused, setFocused] = useState(false);
  const [feedback, setFeedback] = useState<{
    message: string;
    revision: number;
    generation: number;
    selection: ReadonlySet<string>;
  } | null>(null);
  const [submitted, setSubmitted] = useState<{
    id: string;
    revision: number;
    generation: number;
    selection: ReadonlySet<string>;
  } | null>(null);
  const input = useRef<HTMLInputElement>(null);
  const messageId = useId();
  const blocked = useSyncExternalStore(
    binding.subscribeReadiness,
    binding.getBlockingReason,
  );
  const operations = useSyncExternalStore(
    binding.operations.subscribe,
    binding.operations.getSnapshot,
  );
  const matches = (
    value: {
      revision: number;
      generation: number;
      selection: ReadonlySet<string>;
    } | null,
  ) =>
    value !== null &&
    value.revision === draft?.revision &&
    value.generation === draft?.generation &&
    value.selection === binding.selectedRecordIds;
  const entry = matches(submitted)
    ? operations.entries.find((entry) => entry.id === submitted?.id)
    : undefined;
  const localError = matches(feedback) ? feedback?.message : null;
  const operationMessage = entry?.receipt
    ? "Assignment accepted. Batch results and any recovery are available in Recovery."
    : entry?.phase === "waiting"
      ? "Assignment waiting for earlier work. Its selected records are captured."
      : entry?.phase === "uncertain"
        ? "Assignment outcome is uncertain. Review the original action in Recovery."
        : entry?.phase === "rejected"
          ? (entry.failure?.message ??
            "Assignment rejected. Review the original action in Recovery.")
          : entry
            ? "Assigning tag to the captured records."
            : null;
  const message = localError ?? blocked ?? operationMessage;
  const selectedCount = binding.selectedRecordIds.size;
  useWorkbookCommand({
    id: "relate.assign_tag_selection",
    family: "Relate",
    label: "Assign tag to selected records",
    terms: ["bulk", "bounded", "tags"],
    targetKind: "selection",
    availability: (target) =>
      target.kind !== "selection" ||
      target.viewSchemaId !== timelineViewSchemaId ||
      target.recordIds.length !== binding.selectedRecordIds.size ||
      !target.recordIds.every((id) => binding.selectedRecordIds.has(id))
        ? "The selection changed."
        : !binding.canAssign || !draft?.canEdit
          ? "Tag assignment is currently unavailable."
          : blocked,
    invoke: (target) => {
      if (
        target.kind !== "selection" ||
        !binding.canAssign ||
        !draft?.canEdit ||
        blocked ||
        !input.current
      )
        return false;
      input.current.focus();
      return true;
    },
  });

  if (draft === null) return null;
  if (!selectedCount && !draft.raw && !focused && !localError && !entry)
    return null;
  const pending =
    entry && ["waiting", "preparing", "submitting"].includes(entry.phase);
  return (
    <form
      data-grid-editor-external-action="true"
      aria-label="Timeline bulk record actions"
      aria-busy={pending || undefined}
      style={regionStyle}
      onFocusCapture={() => setFocused(true)}
      onBlurCapture={(event) => {
        if (
          !(event.relatedTarget instanceof Node) ||
          !event.currentTarget.contains(event.relatedTarget)
        )
          setFocused(false);
      }}
      onSubmit={(event) => {
        event.preventDefault();
        const result = binding.assignTag(
          draft.raw,
          event.nativeEvent,
          draft.generation,
        );
        if (result.kind === "rejected")
          setFeedback({
            message: result.message,
            revision: draft.revision,
            generation: draft.generation,
            selection: binding.selectedRecordIds,
          });
        else {
          setFeedback(null);
          setSubmitted({
            id: result.operationId,
            revision: draft.revision,
            generation: draft.generation,
            selection: binding.selectedRecordIds,
          });
        }
      }}
    >
      <span style={{ whiteSpace: "nowrap" }}>{selectedCount} selected</span>
      <label style={labelStyle}>
        <span style={visuallyHiddenStyle}>
          Tag for selected Timeline records
        </span>
        <input
          ref={input}
          aria-describedby={message ? messageId : undefined}
          aria-invalid={localError ? true : undefined}
          placeholder="Tag selected"
          type="text"
          value={draft.raw}
          readOnly={!binding.canAssign || !draft.canEdit}
          style={inputStyle}
          onChange={(event) => {
            if (binding.authoring.update(event.target.value, draft.generation))
              setFeedback(null);
          }}
        />
      </label>
      <button
        type="submit"
        disabled={
          !binding.canAssign ||
          !draft.canEdit ||
          selectedCount === 0 ||
          !draft.raw.trim() ||
          blocked !== null
        }
        style={buttonStyle}
      >
        Assign tag
      </button>
      <button
        type="button"
        style={buttonStyle}
        onClick={() => {
          if (!binding.authoring.clear(draft.generation)) return;
          setFeedback(null);
          setSubmitted(null);
          input.current?.focus({ preventScroll: true });
        }}
      >
        Clear tag draft
      </button>
      {message ? (
        <span
          id={messageId}
          role={localError || entry?.phase === "rejected" ? "alert" : "status"}
          style={messageStyle}
        >
          {message}
        </span>
      ) : null}
    </form>
  );
}

const controlStyle = {
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
  font: "inherit",
  boxSizing: "border-box",
} satisfies CSSProperties;
const inputStyle = {
  ...controlStyle,
  inlineSize: "100%",
  minInlineSize: 0,
} satisfies CSSProperties;
const buttonStyle = {
  ...controlStyle,
  cursor: "pointer",
  flex: "0 0 auto",
} satisfies CSSProperties;
const labelStyle = {
  flex: "1 1 12ch",
  minInlineSize: 0,
  maxInlineSize: "var(--ct-layout-viewBarSavedViewMaxInlineSize)",
} satisfies CSSProperties;
const regionStyle = {
  display: "flex",
  flexShrink: 0,
  flexWrap: "wrap",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
  minInlineSize: 0,
  margin: 0,
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;
const messageStyle = {
  flex: "1 1 100%",
  minInlineSize: 0,
  overflowWrap: "anywhere",
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;
