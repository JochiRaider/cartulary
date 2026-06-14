import {
  dataTestIdSelector,
  pasteConflictItemTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties, KeyboardEvent as ReactKeyboardEvent } from "react";
import type {
  LocalConflictState,
  TimelineConflictResolution,
} from "./workbookTimelineModel";

export function TimelineConflictResolver({
  activeConflict,
  activeConflictKey,
  activePasteConflictIndex,
  activePasteConflictKeys,
  conflictQueue,
  getFieldLabel,
  onClose,
  onMergedDraftChange,
  onSelectConflictKey,
  onSubmit,
  showPasteConflictNavigator,
}: {
  readonly activeConflict: LocalConflictState;
  readonly activeConflictKey: string | null;
  readonly activePasteConflictIndex: number;
  readonly activePasteConflictKeys: readonly string[];
  readonly conflictQueue: Record<string, LocalConflictState>;
  readonly getFieldLabel: (fieldKey: string) => string;
  readonly onClose: (conflict: LocalConflictState) => void;
  readonly onMergedDraftChange: (conflictKey: string, value: string) => void;
  readonly onSelectConflictKey: (conflictKey: string) => void;
  readonly onSubmit: (
    conflict: LocalConflictState,
    resolution: TimelineConflictResolution,
  ) => void;
  readonly showPasteConflictNavigator: boolean;
}) {
  const handleKeyDown = (event: ReactKeyboardEvent<HTMLElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      event.stopPropagation();
      onClose(activeConflict);
      return;
    }
    const summary = event.currentTarget.querySelector(
      dataTestIdSelector("conflict-resolver-summary"),
    );
    if (event.key === "Enter" && event.target === summary) {
      event.preventDefault();
      event.stopPropagation();
    }
  };

  return (
    <section
      data-testid="conflict-resolver"
      data-conflict-record-id={activeConflict.anchor.record_id}
      data-conflict-field-key={activeConflict.anchor.field_key}
      data-conflict-base-row-version={activeConflict.anchor.base_row_version}
      data-conflict-current-row-version={
        activeConflict.anchor.current_row_version
      }
      style={conflictResolverStyle}
      aria-label="Same-field conflict resolver"
      onKeyDown={handleKeyDown}
    >
      <div data-testid="conflict-resolver-summary" tabIndex={-1}>
        <p style={eyebrowStyle}>Conflict</p>
        <h2 style={sectionTitleStyle}>
          {getFieldLabel(activeConflict.conflict.field_key)}
        </h2>
        <p style={bodyStyle}>
          This field changed before your edit was saved. Review the saved value
          and your unsaved value.
        </p>
      </div>
      {showPasteConflictNavigator ? (
        <nav
          aria-label="Paste conflict navigator"
          data-testid="paste-conflict-navigator"
          style={noticeCardStyle}
        >
          <p data-testid="paste-conflict-position" style={bodyStyle}>
            {activePasteConflictIndex + 1} of {activePasteConflictKeys.length}
          </p>
          <div style={inlineButtonRowStyle}>
            <button
              data-testid="paste-conflict-previous"
              disabled={activePasteConflictIndex <= 0}
              onClick={() => {
                const previousKey =
                  activePasteConflictKeys[activePasteConflictIndex - 1];
                if (previousKey) {
                  onSelectConflictKey(previousKey);
                }
              }}
              style={secondaryActionButtonStyle}
              type="button"
            >
              Previous
            </button>
            <button
              data-testid="paste-conflict-next"
              disabled={
                activePasteConflictIndex >= activePasteConflictKeys.length - 1
              }
              onClick={() => {
                const nextKey =
                  activePasteConflictKeys[activePasteConflictIndex + 1];
                if (nextKey) {
                  onSelectConflictKey(nextKey);
                }
              }}
              style={secondaryActionButtonStyle}
              type="button"
            >
              Next
            </button>
          </div>
          <div style={inlineButtonRowStyle}>
            {activePasteConflictKeys.map((key, index) => {
              const queued = conflictQueue[key];
              if (!queued) {
                return null;
              }
              return (
                <button
                  aria-current={key === activeConflictKey ? "true" : undefined}
                  data-testid={pasteConflictItemTestId(key)}
                  key={key}
                  onClick={() => onSelectConflictKey(key)}
                  style={
                    key === activeConflictKey
                      ? actionButtonStyle
                      : secondaryActionButtonStyle
                  }
                  type="button"
                >
                  {index + 1}. {getFieldLabel(queued.conflict.field_key)}
                </button>
              );
            })}
          </div>
        </nav>
      ) : null}
      <div style={conflictResolverGridStyle}>
        <label style={labelStyle}>
          Field key
          <input
            readOnly
            style={inputStyle}
            data-testid="conflict-field-key"
            value={activeConflict.conflict.field_key}
          />
        </label>
        <label style={labelStyle}>
          Saved by
          <input
            readOnly
            style={inputStyle}
            data-testid="conflict-server-actor"
            value={activeConflict.conflict.server_updated_by ?? ""}
          />
        </label>
        <label style={labelStyle}>
          Saved at
          <input
            readOnly
            style={inputStyle}
            data-testid="conflict-server-updated-at"
            value={activeConflict.conflict.server_updated_at ?? ""}
          />
        </label>
      </div>
      <div style={conflictResolverGridStyle}>
        <label style={labelStyle}>
          Saved value
          <textarea
            readOnly
            style={textareaStyle}
            data-testid="conflict-server-value"
            value={String(activeConflict.conflict.server_value ?? "")}
          />
        </label>
        <label style={labelStyle}>
          Your unsaved value
          <textarea
            readOnly
            style={textareaStyle}
            data-testid="conflict-local-value"
            value={String(activeConflict.localValue ?? "")}
          />
        </label>
        {activeConflict.conflict.conflict_resolution_class ===
        "text_compare_merge" ? (
          <label style={labelStyle}>
            Merged value
            <textarea
              style={textareaStyle}
              data-testid="conflict-merged-value"
              value={activeConflict.mergedDraft}
              onChange={(event) => {
                onMergedDraftChange(
                  activeConflict.key,
                  event.currentTarget.value,
                );
              }}
            />
          </label>
        ) : null}
      </div>
      <div style={inlineButtonRowStyle}>
        <button
          type="button"
          style={secondaryActionButtonStyle}
          data-testid="conflict-close"
          onClick={() => {
            onClose(activeConflict);
          }}
        >
          Close
        </button>
        <button
          type="button"
          style={secondaryActionButtonStyle}
          data-testid="conflict-keep-saved"
          onClick={() => onSubmit(activeConflict, "keep_saved")}
        >
          Keep saved value
        </button>
        <button
          type="button"
          style={actionButtonStyle}
          data-testid="conflict-use-unsaved"
          onClick={() => onSubmit(activeConflict, "use_unsaved")}
        >
          Use my unsaved value
        </button>
        {activeConflict.conflict.conflict_resolution_class ===
        "text_compare_merge" ? (
          <button
            type="button"
            style={actionButtonStyle}
            data-testid="conflict-use-merged"
            onClick={() => onSubmit(activeConflict, "merged_value")}
          >
            Use merged value
          </button>
        ) : null}
      </div>
    </section>
  );
}

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase",
  color: "var(--ct-colors-accent)",
} satisfies CSSProperties;

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
} satisfies CSSProperties;

const inputStyle = {
  boxSizing: "border-box",
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
} satisfies CSSProperties;

const textareaStyle = {
  ...inputStyle,
  resize: "vertical",
} satisfies CSSProperties;

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
} satisfies CSSProperties;

const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
} satisfies CSSProperties;

const conflictResolverStyle = {
  display: "grid",
  gap: "0.75rem",
  padding: "1rem",
  borderTop: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
} satisfies CSSProperties;

const conflictResolverGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.75rem",
} satisfies CSSProperties;

const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap",
} satisfies CSSProperties;

const noticeCardStyle = {
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.85rem 1rem",
  display: "grid",
  gap: "0.5rem",
} satisfies CSSProperties;
