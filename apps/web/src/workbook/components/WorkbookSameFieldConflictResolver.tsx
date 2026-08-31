import {
  pasteConflictItemTestId,
  workbookConflictControlTestId,
  workbookConflictLocalValueTestId,
  workbookConflictResolverTestId,
  workbookConflictSavedValueTestId,
  workbookConflictSummaryTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type RefObject,
  useEffect,
  useRef,
  useState,
} from "react";
import type {
  WorkbookMutationRuntime,
  WorkbookMutationSnapshot,
} from "../runtime/WorkbookMutationRuntime";
import { RecoverySurface } from "./RecoverySurface";

function displayConflictValue(value: unknown): string {
  if (typeof value === "string") return value;
  if (value === undefined) return "";
  return JSON.stringify(value, null, 2);
}

function LineComparison({
  label,
  testId,
  value,
}: {
  readonly label: string;
  readonly testId?: string | undefined;
  readonly value: unknown;
}) {
  const lines = displayConflictValue(value).split("\n");
  const occurrences = new Map<string, number>();
  const keyedLines = lines.map((line) => {
    const occurrence = (occurrences.get(line) ?? 0) + 1;
    occurrences.set(line, occurrence);
    return { key: `${line}\u0000${occurrence}`, line };
  });
  return (
    <section aria-label={label} style={comparisonStyle}>
      <h3 style={comparisonTitleStyle}>{label}</h3>
      {testId ? (
        <textarea
          aria-hidden="true"
          data-testid={testId}
          readOnly
          style={machineReadableValueStyle}
          tabIndex={-1}
          value={displayConflictValue(value)}
        />
      ) : null}
      <ol style={lineListStyle}>
        {keyedLines.map(({ key, line }) => (
          <li key={key} style={lineStyle}>
            <code>{line === "" ? " " : line}</code>
          </li>
        ))}
      </ol>
    </section>
  );
}

export function WorkbookSameFieldConflictResolver({
  apiBase,
  focusSummary,
  mutationRuntime,
  onActivateOrigin,
  snapshot,
  summaryRef,
}: {
  readonly apiBase?: string | undefined;
  readonly focusSummary: () => void;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onActivateOrigin: (viewSchemaId: string) => void;
  readonly snapshot: Pick<
    WorkbookMutationSnapshot,
    "conflictPanelOpen" | "conflicts"
  >;
  readonly summaryRef: RefObject<HTMLDivElement | null>;
}) {
  const [activeKey, setActiveKey] = useState<string | null>(null);
  const conflict =
    snapshot.conflicts.find((entry) => entry.key === activeKey) ??
    snapshot.conflicts[0] ??
    null;
  const resolverRef = useRef<HTMLElement | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    setActiveKey((current) =>
      current !== null &&
      snapshot.conflicts.some((entry) => entry.key === current)
        ? current
        : (snapshot.conflicts[0]?.key ?? null),
    );
  }, [snapshot.conflicts]);

  useEffect(() => {
    setMessage(null);
    if (snapshot.conflictPanelOpen && conflict !== null) {
      if (resolverRef.current?.contains(document.activeElement)) return;
      focusSummary();
    }
  }, [conflict, focusSummary, snapshot.conflictPanelOpen]);

  useEffect(() => {
    if (!snapshot.conflictPanelOpen || conflict === null) return;
    const dismissFromUnhandledEscape = (event: globalThis.KeyboardEvent) => {
      if (event.key !== "Escape" || event.defaultPrevented) return;
      event.preventDefault();
      event.stopPropagation();
      mutationRuntime.dismissConflict(conflict.key);
    };
    document.addEventListener("keydown", dismissFromUnhandledEscape);
    return () =>
      document.removeEventListener("keydown", dismissFromUnhandledEscape);
  }, [conflict, mutationRuntime, snapshot.conflictPanelOpen]);

  if (conflict === null) return null;
  if (!snapshot.conflictPanelOpen) return null;

  const submit = async (
    resolutionKind: "keep_saved" | "merged_value" | "use_unsaved",
  ) => {
    setSubmitting(true);
    setMessage(null);
    try {
      const nextMessage = await mutationRuntime.resolveConflict({
        apiBase,
        key: conflict.key,
        resolutionKind,
      });
      setMessage(nextMessage);
    } finally {
      setSubmitting(false);
    }
  };
  const isText = conflict.resolutionClass === "text_compare_merge";
  const isCollection = conflict.resolutionClass === "collection_review";
  const suggestion = conflict.conflict.suggested_merged_value;
  const activeConflictIndex = snapshot.conflicts.findIndex(
    (entry) => entry.key === conflict.key,
  );
  const dismiss = () => mutationRuntime.dismissConflict(conflict.key);

  return (
    <RecoverySurface
      aria-label="Workbook conflict recovery"
      ref={resolverRef}
      data-conflict-base-row-version={String(
        conflict.conflict.base_row_version,
      )}
      data-conflict-current-row-version={String(
        conflict.conflict.current_row_version,
      )}
      data-conflict-field-key={conflict.conflict.field_key}
      data-conflict-record-id={conflict.conflict.record_id}
      data-conflict-resolution-class={conflict.resolutionClass}
      data-testid={workbookConflictResolverTestId()}
      onKeyDown={(event) => {
        if (event.key !== "Escape") return;
        event.preventDefault();
        event.stopPropagation();
        mutationRuntime.dismissConflict(conflict.key);
      }}
    >
      <div
        ref={summaryRef}
        data-testid={workbookConflictSummaryTestId()}
        tabIndex={-1}
      >
        <p style={eyebrowStyle}>Conflict requires review</p>
        <h2 style={titleStyle}>
          {conflict.origin.surfaceLabel}: {conflict.origin.rowLabel}
        </h2>
        <p style={bodyStyle}>
          The saved value changed before your edit reached the server. Nothing
          will be chosen automatically.
        </p>
      </div>

      <button
        data-testid={workbookConflictControlTestId("activate-origin")}
        onClick={() => onActivateOrigin(conflict.origin.viewSchemaId)}
        style={secondaryButtonStyle}
        type="button"
      >
        Return to affected surface
      </button>
      <button
        aria-label="Close conflict recovery"
        data-testid={workbookConflictControlTestId("close")}
        onClick={dismiss}
        style={secondaryButtonStyle}
        type="button"
      >
        Close
      </button>

      {snapshot.conflicts.length > 1 ? (
        <nav
          aria-label="Workbook conflict navigator"
          data-testid={workbookConflictControlTestId("paste-navigator")}
          style={navigatorStyle}
        >
          <p
            data-testid={workbookConflictControlTestId("paste-position")}
            style={bodyStyle}
          >
            {activeConflictIndex + 1} of {snapshot.conflicts.length}
          </p>
          <div style={buttonRowStyle}>
            <button
              data-testid={workbookConflictControlTestId("paste-previous")}
              disabled={activeConflictIndex <= 0}
              onClick={() => {
                const previous = snapshot.conflicts[activeConflictIndex - 1];
                if (previous !== undefined) setActiveKey(previous.key);
              }}
              style={secondaryButtonStyle}
              type="button"
            >
              Previous
            </button>
            <button
              data-testid={workbookConflictControlTestId("paste-next")}
              disabled={activeConflictIndex >= snapshot.conflicts.length - 1}
              onClick={() => {
                const next = snapshot.conflicts[activeConflictIndex + 1];
                if (next !== undefined) setActiveKey(next.key);
              }}
              style={secondaryButtonStyle}
              type="button"
            >
              Next
            </button>
          </div>
          <div style={buttonRowStyle}>
            {snapshot.conflicts.map((entry, index) => (
              <button
                aria-current={entry.key === conflict.key ? "true" : undefined}
                data-testid={pasteConflictItemTestId(entry.key)}
                key={entry.key}
                onClick={() => setActiveKey(entry.key)}
                style={
                  entry.key === conflict.key
                    ? selectedConflictButtonStyle
                    : secondaryButtonStyle
                }
                type="button"
              >
                {index + 1}. {entry.conflict.field_key}
              </button>
            ))}
          </div>
        </nav>
      ) : null}

      {isText ? (
        <>
          <div style={comparisonGridStyle}>
            <LineComparison
              label="Base value"
              value={conflict.conflict.base_value}
            />
            <LineComparison
              label="Saved value"
              testId={workbookConflictSavedValueTestId()}
              value={conflict.conflict.server_value}
            />
            <LineComparison
              label="Your unsaved value"
              testId={workbookConflictLocalValueTestId()}
              value={conflict.localValue}
            />
          </div>
          <label style={labelStyle}>
            Merged value
            <textarea
              data-testid={workbookConflictControlTestId("merged-value")}
              onChange={(event) =>
                mutationRuntime.updateConflictDraft(
                  conflict.key,
                  event.currentTarget.value,
                )
              }
              style={textareaStyle}
              value={conflict.mergedDraft}
            />
          </label>
          {suggestion !== undefined ? (
            <button
              data-testid={workbookConflictControlTestId(
                "use-server-suggestion",
              )}
              disabled={submitting}
              onClick={() =>
                mutationRuntime.updateConflictDraft(
                  conflict.key,
                  displayConflictValue(suggestion),
                )
              }
              style={secondaryButtonStyle}
              type="button"
            >
              Copy server suggestion into editor
            </button>
          ) : null}
        </>
      ) : (
        <div style={comparisonGridStyle}>
          {isCollection ? (
            <LineComparison
              label="Base collection"
              value={conflict.conflict.base_value}
            />
          ) : null}
          <LineComparison
            label="Saved value"
            testId={workbookConflictSavedValueTestId()}
            value={conflict.conflict.server_value}
          />
          <LineComparison
            label="Your unsaved value"
            testId={workbookConflictLocalValueTestId()}
            value={conflict.localValue}
          />
          {isCollection ? (
            <LineComparison label="Final preview" value={conflict.localValue} />
          ) : null}
        </div>
      )}

      {message ? (
        <p aria-live="polite" role="status" style={errorStyle}>
          {message}
        </p>
      ) : null}
      <div style={buttonRowStyle}>
        <button
          data-testid={workbookConflictControlTestId("keep-saved")}
          disabled={submitting}
          onClick={() => void submit("keep_saved")}
          style={destructiveButtonStyle}
          type="button"
        >
          Discard local draft
        </button>
        {isCollection ? (
          <button
            data-testid={workbookConflictControlTestId("apply-collection")}
            disabled={submitting}
            onClick={() => void submit("merged_value")}
            style={secondaryButtonStyle}
            type="button"
          >
            Apply reviewed collection
          </button>
        ) : isText ? (
          <>
            <button
              data-testid={workbookConflictControlTestId("use-unsaved")}
              disabled={submitting}
              onClick={() => void submit("use_unsaved")}
              style={secondaryButtonStyle}
              type="button"
            >
              Use my unsaved value
            </button>
            <button
              data-testid={workbookConflictControlTestId("use-merged")}
              disabled={submitting}
              onClick={() => void submit("merged_value")}
              style={secondaryButtonStyle}
              type="button"
            >
              Use merged value
            </button>
          </>
        ) : (
          <button
            data-testid={workbookConflictControlTestId("use-unsaved")}
            disabled={submitting}
            onClick={() => void submit("use_unsaved")}
            style={secondaryButtonStyle}
            type="button"
          >
            Use my unsaved value
          </button>
        )}
      </div>
    </RecoverySurface>
  );
}

const eyebrowStyle = {
  margin: 0,
  color: "var(--ct-colors-semantic-conflict)",
  fontSize: "0.78rem",
  fontWeight: 800,
  letterSpacing: "0.08em",
  textTransform: "uppercase",
} satisfies CSSProperties;
const titleStyle = {
  margin: "0.2rem 0",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const bodyStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const comparisonGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(13rem, 1fr))",
  gap: "0.75rem",
} satisfies CSSProperties;
const comparisonStyle = {
  minWidth: 0,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  overflow: "hidden",
} satisfies CSSProperties;
const machineReadableValueStyle = {
  position: "absolute",
  inlineSize: "1px",
  blockSize: "1px",
  overflow: "hidden",
  clipPath: "inset(50%)",
} satisfies CSSProperties;
const comparisonTitleStyle = {
  margin: 0,
  padding: "0.45rem 0.6rem",
  fontSize: "0.85rem",
  background: "var(--ct-colors-surface-2)",
} satisfies CSSProperties;
const lineListStyle = {
  margin: 0,
  padding: "0.5rem 0.5rem 0.5rem 2.5rem",
  maxHeight: "clamp(6rem, 18vh, 12rem)",
  overflow: "auto",
  whiteSpace: "pre-wrap",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const lineStyle = {
  paddingInlineStart: "0.3rem",
  borderInlineStart: "var(--ct-border-hairline)",
} satisfies CSSProperties;
const labelStyle = {
  display: "grid",
  gap: "0.35rem",
  fontWeight: 700,
} satisfies CSSProperties;
const textareaStyle = {
  minHeight: "clamp(5rem, 14vh, 9rem)",
  boxSizing: "border-box",
  width: "100%",
  resize: "vertical",
  border: "var(--ct-component-text-input-border)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
  font: "inherit",
  padding: "0.65rem",
} satisfies CSSProperties;
const buttonRowStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "0.5rem",
} satisfies CSSProperties;
const baseButtonStyle = {
  maxInlineSize: "100%",
  minInlineSize: 0,
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  padding: "var(--ct-component-button-secondary-padding)",
  font: "inherit",
  fontWeight: 700,
  cursor: "pointer",
  overflowWrap: "anywhere",
  whiteSpace: "normal",
} satisfies CSSProperties;
const secondaryButtonStyle = {
  ...baseButtonStyle,
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  justifySelf: "start",
} satisfies CSSProperties;
const selectedConflictButtonStyle = {
  ...secondaryButtonStyle,
  border: "1px solid var(--ct-colors-semantic-conflict)",
  background: "var(--ct-colors-surface-3)",
} satisfies CSSProperties;
const destructiveButtonStyle = {
  ...baseButtonStyle,
  border: "1px solid var(--ct-colors-semantic-destructive)",
  background: "var(--ct-component-button-danger-backgroundColor)",
  color: "var(--ct-component-button-danger-textColor)",
} satisfies CSSProperties;
const errorStyle = {
  margin: 0,
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
} satisfies CSSProperties;
const navigatorStyle = {
  display: "grid",
  gap: "0.5rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  padding: "0.65rem",
} satisfies CSSProperties;
