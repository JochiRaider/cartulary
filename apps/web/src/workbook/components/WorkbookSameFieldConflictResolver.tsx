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
import { useWorkbookRecoveryNavigation } from "../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryNavigation } from "../../shared/workbookRecoveryNavigation";
import { useWorkbookBrowsingRegistry } from "../query/WorkbookQueryBrowsingContext";
import type {
  WorkbookConflictResolutionSettlement,
  WorkbookMutationRuntime,
  WorkbookMutationSnapshot,
} from "../runtime/WorkbookMutationRuntime";

type ResolutionIntent = {
  readonly id: number;
  readonly key: string;
  readonly token: string;
  readonly selectionGeneration: number;
  readonly activation: number;
  readonly attachment: string | null;
  readonly authorityEpoch: number;
  readonly controller: AbortController;
  readonly panel: HTMLElement | null;
  readonly scrollElement: HTMLElement | null;
  readonly viewport: { top: number; left: number } | null;
  readonly view: string;
  readonly recordId: string;
  readonly fieldKey: string;
  eligible: boolean;
  focusing: boolean;
  releaseRetirement: (() => void) | null;
};

type ResolutionAttempt = {
  readonly id: number;
  readonly token: string;
  readonly intent: ResolutionIntent;
};

function attachmentStillCurrent(
  navigation: WorkbookRecoveryNavigation | null,
  intent: ResolutionIntent,
  allowRemovedSelection: boolean,
): boolean {
  const current = navigation?.getSnapshot();
  return (
    current?.open === true &&
    current.activation === intent.activation &&
    (current.selected === intent.attachment ||
      (allowRemovedSelection && current.selected === null))
  );
}

function retireIntent(intent: ResolutionIntent): void {
  intent.eligible = false;
  intent.controller.abort();
}

function bindIntentRetirement(
  intent: ResolutionIntent,
  navigation: WorkbookRecoveryNavigation | null,
  mutationRuntime: WorkbookMutationRuntime,
): () => void {
  const onFocus = (event: FocusEvent) => {
    if (
      intent.focusing &&
      event.target instanceof Node &&
      intent.scrollElement?.contains(event.target)
    )
      return;
    if (!intent.panel?.contains(event.target as Node)) retireIntent(intent);
  };
  const onKey = (event: KeyboardEvent) => {
    if (!["Shift", "Control", "Alt"].includes(event.key)) retireIntent(intent);
  };
  const onPointer = () => retireIntent(intent);
  const onScroll = () => {
    if (!intent.focusing) retireIntent(intent);
  };
  const onNavigation = () => {
    const removed = mutationRuntime
      .getSnapshot()
      .conflicts.every(
        (entry) =>
          entry.key !== intent.key ||
          entry.conflict.conflict_token !== intent.token,
      );
    if (!attachmentStillCurrent(navigation, intent, removed))
      retireIntent(intent);
  };
  const onRuntime = () => {
    if (mutationRuntime.authorizationEpoch !== intent.authorityEpoch)
      retireIntent(intent);
  };
  document.addEventListener("keydown", onKey, true);
  document.addEventListener("focusin", onFocus, true);
  document.addEventListener("pointerdown", onPointer, true);
  document.addEventListener("wheel", onScroll, true);
  const unsubscribeNavigation = navigation?.subscribe(onNavigation);
  const unsubscribeRuntime = mutationRuntime.subscribe(onRuntime);
  return () => {
    document.removeEventListener("keydown", onKey, true);
    document.removeEventListener("focusin", onFocus, true);
    document.removeEventListener("pointerdown", onPointer, true);
    document.removeEventListener("wheel", onScroll, true);
    unsubscribeNavigation?.();
    unsubscribeRuntime();
  };
}

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
      {value === null ? (
        <p>Cleared (null)</p>
      ) : value === "" ? (
        <p>Empty text</p>
      ) : null}
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
  activation,
  apiBase,
  onClose,
  mutationRuntime,
  onActivateOrigin,
  snapshot,
  summaryRef,
}: {
  readonly activation?: {
    readonly conflictKey: string;
    readonly sequence: number;
  } | null;
  readonly apiBase?: string | undefined;
  readonly onClose: () => void;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onActivateOrigin: (viewSchemaId: string) => void;
  readonly snapshot: Pick<WorkbookMutationSnapshot, "conflicts">;
  readonly summaryRef: RefObject<HTMLDivElement | null>;
}) {
  const navigation = useWorkbookRecoveryNavigation();
  const browsing = useWorkbookBrowsingRegistry();
  const [activeKey, setActiveKey] = useState<string | null>(
    activation?.conflictKey ?? snapshot.conflicts[0]?.key ?? null,
  );
  const selectedKey = useRef(activeKey);
  selectedKey.current = activeKey;
  const selectionGeneration = useRef(0);
  const nextSubmissionId = useRef(0);
  const attempts = useRef(new Map<string, ResolutionAttempt>());
  const feedback = useRef(
    new Map<string, { token: string; message: string }>(),
  );
  const [, renderPresentation] = useState(0);
  const publishPresentation = () => renderPresentation((value) => value + 1);
  const selectConflict = (key: string) => {
    if (selectedKey.current === key) return;
    selectedKey.current = key;
    selectionGeneration.current += 1;
    for (const attempt of attempts.current.values())
      retireIntent(attempt.intent);
    setActiveKey(key);
  };
  useEffect(() => {
    if (activation === undefined || activation === null) return;
    if (selectedKey.current === activation.conflictKey) return;
    selectedKey.current = activation.conflictKey;
    selectionGeneration.current += 1;
    for (const attempt of attempts.current.values())
      retireIntent(attempt.intent);
    setActiveKey(activation.conflictKey);
  }, [activation]);
  const conflict =
    snapshot.conflicts.find((entry) => entry.key === activeKey) ?? null;
  useEffect(() => {
    if (activeKey === null || conflict !== null) return;
    const state = navigation?.getSnapshot();
    if (!state?.open || state.selected === null) return;
    if (attempts.current.get(activeKey)?.intent.eligible) return;
    navigation?.openList();
  }, [activeKey, conflict, navigation]);
  const resolverRef = useRef<HTMLElement | null>(null);
  useEffect(() => {
    return () => {
      // The selected entry may be removed by this admitted resolution. Its
      // continuation may still return focus while the source grid remains.
      for (const attempt of attempts.current.values()) {
        const intent = attempt.intent;
        const removedBySettlement =
          navigation?.getSnapshot().open &&
          navigation.getSnapshot().selected === null &&
          mutationRuntime
            .getSnapshot()
            .conflicts.every(
              (entry) =>
                entry.key !== intent.key ||
                entry.conflict.conflict_token !== intent.token,
            );
        if (!removedBySettlement) retireIntent(intent);
      }
    };
  }, [navigation, mutationRuntime]);

  const currentAttempt = conflict && attempts.current.get(conflict.key);
  const submitting =
    conflict !== null &&
    currentAttempt?.token === conflict.conflict.conflict_token;
  const currentFeedback = conflict
    ? feedback.current.get(conflict.key)
    : undefined;
  const message =
    conflict && currentFeedback?.token === conflict.conflict.conflict_token
      ? currentFeedback.message
      : null;

  if (conflict === null)
    return (
      <p>
        This conflict is no longer available. Use All recovery to review
        remaining work.
      </p>
    );

  const submit = (
    resolutionKind: "keep_saved" | "merged_value" | "use_unsaved",
  ) => {
    const previous = attempts.current.get(conflict.key);
    if (previous?.token === conflict.conflict.conflict_token) return;
    if (previous) {
      retireIntent(previous.intent);
      previous.intent.releaseRetirement?.();
      previous.intent.releaseRetirement = null;
      attempts.current.delete(conflict.key);
    }
    const current = navigation?.getSnapshot();
    const panel = resolverRef.current?.closest(
      '[aria-label="Recovery navigation"]',
    ) as HTMLElement | null;
    const view = conflict.origin.viewSchemaId;
    const grid = browsing.grid(view);
    const scrollElement = grid?.getScrollElement() ?? null;
    const viewport = scrollElement
      ? { top: scrollElement.scrollTop, left: scrollElement.scrollLeft }
      : null;
    const start = mutationRuntime.beginConflictResolution({
      apiBase,
      key: conflict.key,
      resolutionKind,
    });
    if (start.kind === "rejected") {
      feedback.current.set(conflict.key, {
        token: conflict.conflict.conflict_token,
        message: start.message,
      });
      publishPresentation();
      return;
    }
    const id = ++nextSubmissionId.current;
    const intent: ResolutionIntent = {
      id,
      key: conflict.key,
      token: start.conflictToken,
      selectionGeneration: selectionGeneration.current,
      activation: current?.activation ?? -1,
      attachment: current?.selected ?? null,
      authorityEpoch: mutationRuntime.authorizationEpoch,
      controller: new AbortController(),
      panel,
      scrollElement,
      viewport,
      view,
      recordId: conflict.conflict.record_id,
      fieldKey: conflict.conflict.field_key,
      eligible: true,
      focusing: false,
      releaseRetirement: null,
    };
    intent.releaseRetirement = bindIntentRetirement(
      intent,
      navigation,
      mutationRuntime,
    );
    attempts.current.set(conflict.key, {
      id,
      token: start.conflictToken,
      intent,
    });
    feedback.current.delete(conflict.key);
    publishPresentation();
    const eligible = (allowRemovedSelection: boolean) =>
      intent.eligible &&
      selectedKey.current === intent.key &&
      selectionGeneration.current === intent.selectionGeneration &&
      mutationRuntime.authorizationEpoch === intent.authorityEpoch &&
      attachmentStillCurrent(
        navigation,
        intent,
        allowRemovedSelection &&
          mutationRuntime
            .getSnapshot()
            .conflicts.every(
              (entry) =>
                entry.key !== intent.key ||
                entry.conflict.conflict_token !== intent.token,
            ),
      ) &&
      intent.panel?.isConnected === true;
    void start.completion
      .then(async (outcome: WorkbookConflictResolutionSettlement) => {
        if (attempts.current.get(intent.key)?.id !== id) return;
        const finish = () => {
          intent.releaseRetirement?.();
          intent.releaseRetirement = null;
          if (attempts.current.get(intent.key)?.id === id)
            attempts.current.delete(intent.key);
          publishPresentation();
        };
        if (outcome.kind === "failed" || outcome.kind === "refreshed") {
          const currentEntry = mutationRuntime
            .getSnapshot()
            .conflicts.find((entry) => entry.key === intent.key);
          if (currentEntry?.conflict.conflict_token === outcome.conflictToken)
            feedback.current.set(intent.key, {
              token: outcome.conflictToken,
              message: outcome.message,
            });
        }
        try {
          if (outcome.kind !== "resolved" || !eligible(true)) return;
          if (
            !intent.panel?.contains(document.activeElement) &&
            document.activeElement !== document.body
          )
            return;
          const currentGrid = browsing.grid(intent.view);
          if (
            !currentGrid ||
            !intent.scrollElement?.isConnected ||
            currentGrid.getScrollElement() !== intent.scrollElement
          )
            return;
          intent.focusing = true;
          const result = await currentGrid.requestFocus(
            {
              kind: "cell",
              anchor: {
                surface: { kind: "view_schema", viewSchemaId: intent.view },
                rowIdentity: { kind: "core_record", recordId: intent.recordId },
                fieldKey: intent.fieldKey,
              },
            },
            { signal: intent.controller.signal },
          );
          intent.focusing = false;
          if (!eligible(true) || result !== "focused") return;
          if (
            intent.viewport &&
            browsing.grid(intent.view)?.getScrollElement() ===
              intent.scrollElement
          ) {
            intent.scrollElement.scrollTop = intent.viewport.top;
            intent.scrollElement.scrollLeft = intent.viewport.left;
          }
          navigation?.close();
        } finally {
          finish();
        }
      })
      .catch(() => {
        if (attempts.current.get(intent.key)?.id !== id) return;
        intent.releaseRetirement?.();
        intent.releaseRetirement = null;
        attempts.current.delete(intent.key);
        if (
          mutationRuntime
            .getSnapshot()
            .conflicts.some(
              (entry) =>
                entry.key === intent.key &&
                entry.conflict.conflict_token === intent.token,
            )
        )
          feedback.current.set(intent.key, {
            token: intent.token,
            message:
              "The resolution could not complete. Review the conflict and retry.",
          });
        publishPresentation();
      });
  };
  const isText = conflict.resolutionClass === "text_compare_merge";
  const isCollection = conflict.resolutionClass === "collection_review";
  const suggestion = conflict.conflict.suggested_merged_value;
  const groupedConflicts = snapshot.conflicts.filter(
    (entry) => entry.batchOperationId === conflict.batchOperationId,
  );
  const activeConflictIndex = groupedConflicts.findIndex(
    (entry) => entry.key === conflict.key,
  );
  const dismiss = onClose;

  return (
    <section
      aria-label="Workbook conflict recovery"
      data-grid-editor-external-action="true"
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
        onClose();
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

      {groupedConflicts.length > 1 ? (
        <nav
          aria-label="Workbook conflict navigator"
          data-testid={workbookConflictControlTestId("paste-navigator")}
          style={navigatorStyle}
        >
          <p
            data-testid={workbookConflictControlTestId("paste-position")}
            style={bodyStyle}
          >
            {activeConflictIndex + 1} of {groupedConflicts.length}
          </p>
          <div style={buttonRowStyle}>
            <button
              data-testid={workbookConflictControlTestId("paste-previous")}
              disabled={activeConflictIndex <= 0}
              onClick={() => {
                const previous = groupedConflicts[activeConflictIndex - 1];
                if (previous !== undefined) selectConflict(previous.key);
              }}
              style={secondaryButtonStyle}
              type="button"
            >
              Previous
            </button>
            <button
              data-testid={workbookConflictControlTestId("paste-next")}
              disabled={activeConflictIndex >= groupedConflicts.length - 1}
              onClick={() => {
                const next = groupedConflicts[activeConflictIndex + 1];
                if (next !== undefined) selectConflict(next.key);
              }}
              style={secondaryButtonStyle}
              type="button"
            >
              Next
            </button>
          </div>
          <div style={buttonRowStyle}>
            {groupedConflicts.map((entry, index) => (
              <button
                aria-current={entry.key === conflict.key ? "true" : undefined}
                data-testid={pasteConflictItemTestId(entry.key)}
                key={entry.key}
                onClick={() => selectConflict(entry.key)}
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
              readOnly={submitting}
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
              aria-disabled={submitting || !!conflict.compoundOperationId}
              onClick={() => {
                if (submitting || conflict.compoundOperationId) return;
                mutationRuntime.updateConflictDraft(
                  conflict.key,
                  displayConflictValue(suggestion),
                );
              }}
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

      {submitting ? (
        <p role="status" aria-live="polite">
          Resolving this conflict…
        </p>
      ) : null}
      {message ? (
        <p aria-live="polite" role="status" style={errorStyle}>
          {message}
        </p>
      ) : null}
      {conflict.compoundOperationId ? (
        <p role="status">
          This field belongs to a complete action. Keep saved to clear the
          conflict without a revision, then review and submit the retained
          action together.
        </p>
      ) : null}
      <div style={buttonRowStyle}>
        <button
          data-testid={workbookConflictControlTestId("keep-saved")}
          aria-busy={submitting}
          aria-disabled={submitting}
          onClick={() => submit("keep_saved")}
          style={destructiveButtonStyle}
          type="button"
        >
          {conflict.compoundOperationId ? "Keep saved" : "Discard local draft"}
        </button>
        {isCollection ? (
          <button
            data-testid={workbookConflictControlTestId("apply-collection")}
            aria-busy={submitting}
            aria-disabled={submitting || !!conflict.compoundOperationId}
            onClick={() => {
              if (!conflict.compoundOperationId) submit("merged_value");
            }}
            style={secondaryButtonStyle}
            type="button"
          >
            Apply reviewed collection
          </button>
        ) : isText ? (
          <>
            <button
              data-testid={workbookConflictControlTestId("use-unsaved")}
              aria-busy={submitting}
              aria-disabled={submitting || !!conflict.compoundOperationId}
              onClick={() => {
                if (!conflict.compoundOperationId) submit("use_unsaved");
              }}
              style={secondaryButtonStyle}
              type="button"
            >
              Use my unsaved value
            </button>
            <button
              data-testid={workbookConflictControlTestId("use-merged")}
              aria-busy={submitting}
              aria-disabled={submitting || !!conflict.compoundOperationId}
              onClick={() => {
                if (!conflict.compoundOperationId) submit("merged_value");
              }}
              style={secondaryButtonStyle}
              type="button"
            >
              Use merged value
            </button>
          </>
        ) : (
          <button
            data-testid={workbookConflictControlTestId("use-unsaved")}
            aria-busy={submitting}
            aria-disabled={submitting || !!conflict.compoundOperationId}
            onClick={() => {
              if (!conflict.compoundOperationId) submit("use_unsaved");
            }}
            style={secondaryButtonStyle}
            type="button"
          >
            Use my unsaved value
          </button>
        )}
      </div>
    </section>
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
