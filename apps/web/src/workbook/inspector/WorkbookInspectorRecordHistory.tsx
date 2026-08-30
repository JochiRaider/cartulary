import {
  rowHistoryActionTestId,
  rowHistoryItemTestId,
  rowHistoryLoadingTestId,
  rowHistoryMessageTestId,
  rowHistoryPanelTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import type { RecordRouteCommandPort } from "../mutations/workbookMutationCommandPorts";
import {
  buildRecordRollbackTargetFromHistoryAction,
  type RecordHistoryData,
  type RecordHistoryItem,
  type RecordHistoryRollbackAction,
} from "../timeline/models/timelineHistoryModel";

type HistorySubject = {
  readonly recordId: string;
  readonly rowVersion: number;
};

type PendingRollback = {
  readonly action: RecordHistoryRollbackAction;
  readonly historyItemRef: string;
  readonly recordId: string;
  readonly rowVersion: number;
  readonly target: Record<string, unknown>;
};

export function WorkbookInspectorRecordHistory({
  canMutate,
  commands,
  onMessage,
  onRefresh,
  subject,
}: {
  readonly canMutate: boolean;
  readonly commands: RecordRouteCommandPort;
  readonly onMessage: (message: string) => void;
  readonly onRefresh: () => Promise<void> | void;
  readonly subject: HistorySubject | null;
}) {
  const [data, setData] = useState<RecordHistoryData | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [pending, setPending] = useState<PendingRollback | null>(null);
  const [status, setStatus] = useState<"idle" | "loading" | "ready">("idle");
  const generationRef = useRef(0);
  const subjectRecordId = subject?.recordId ?? null;
  const subjectRowVersion = subject?.rowVersion ?? null;

  const loadHistory = useCallback(
    async (activeSubject: HistorySubject, generation: number) => {
      const outcome = await commands.loadHistory({
        recordId: activeSubject.recordId,
      });
      if (generationRef.current !== generation) return;
      if (outcome.kind === "rejected") {
        setData(null);
        setMessage(outcome.failure.message);
        setStatus("ready");
        return;
      }
      setData(outcome.value);
      setMessage(null);
      setStatus("ready");
    },
    [commands],
  );

  useEffect(() => {
    const generation = generationRef.current + 1;
    generationRef.current = generation;
    setPending(null);
    setMessage(null);
    setData(null);
    if (subjectRecordId === null || subjectRowVersion === null) {
      setStatus("idle");
      return;
    }
    setStatus("idle");
  }, [subjectRecordId, subjectRowVersion]);

  const openHistory = useCallback(() => {
    if (subjectRecordId === null || subjectRowVersion === null) return;
    const generation = generationRef.current;
    setStatus("loading");
    void loadHistory(
      { recordId: subjectRecordId, rowVersion: subjectRowVersion },
      generation,
    );
  }, [loadHistory, subjectRecordId, subjectRowVersion]);

  const confirmRollback = useCallback(async () => {
    const active = pending;
    const activeSubject = subject;
    if (
      active === null ||
      activeSubject === null ||
      active.recordId !== activeSubject.recordId ||
      !canMutate
    ) {
      setPending(null);
      return;
    }
    const generation = generationRef.current;
    setMessage(null);
    const outcome = await commands.rollback({
      baseRowVersion: active.rowVersion,
      reason: `Rollback ${active.action} from the workbook inspector`,
      recordId: active.recordId,
      target: active.target,
    });
    if (generationRef.current !== generation) return;
    if (outcome.kind === "rejected") {
      setPending(null);
      setMessage(outcome.failure.message);
      return;
    }
    setPending(null);
    onMessage(`Rolled back record ${active.recordId}.`);
    await onRefresh();
    if (generationRef.current !== generation) return;
    await loadHistory(
      { recordId: active.recordId, rowVersion: outcome.value.rowVersion },
      generation,
    );
  }, [
    canMutate,
    commands,
    loadHistory,
    onMessage,
    onRefresh,
    pending,
    subject,
  ]);

  if (subject === null) return null;
  return (
    <section data-testid={rowHistoryPanelTestId()} style={panelStyle}>
      <h4 style={titleStyle}>Row history</h4>
      <p style={metadataStyle}>
        Record <code>{subject.recordId}</code>
      </p>
      {status === "idle" ? (
        <button type="button" onClick={openHistory}>
          Open history
        </button>
      ) : null}
      {status === "loading" ? (
        <p data-testid={rowHistoryLoadingTestId()} style={metadataStyle}>
          Loading history...
        </p>
      ) : null}
      {message !== null ? (
        <p
          aria-live="assertive"
          data-testid={rowHistoryMessageTestId()}
          role="alert"
          style={errorStyle}
        >
          {message}
        </p>
      ) : null}
      {data === null ? null : (
        <ol style={listStyle}>
          {data.items.map((item) => (
            <li
              data-testid={rowHistoryItemTestId({
                historyItemRef: item.history_item_ref,
              })}
              key={item.history_item_ref}
              style={itemStyle}
            >
              <strong>{item.operation}</strong>
              <time dateTime={item.committed_at}>{item.committed_at}</time>
              <span>{item.diff_summary.summary}</span>
              <div style={actionsStyle}>
                {item.available_rollback_actions.map((action) => (
                  <button
                    data-testid={rowHistoryActionTestId({
                      action,
                      historyItemRef: item.history_item_ref,
                    })}
                    disabled={!canMutate}
                    key={action}
                    type="button"
                    onClick={() => {
                      const target = buildRecordRollbackTargetFromHistoryAction(
                        item,
                        action,
                      );
                      if (target === null) return;
                      setPending({
                        action,
                        historyItemRef: item.history_item_ref,
                        recordId: data.record_id,
                        rowVersion: data.row_version,
                        target,
                      });
                    }}
                  >
                    {rollbackLabel(item, action)}
                  </button>
                ))}
              </div>
            </li>
          ))}
        </ol>
      )}
      {pending === null ? null : (
        <div
          aria-label="Rollback preview"
          aria-modal="true"
          data-testid={rowHistoryRollbackPreviewTestId({
            action: pending.action,
            historyItemRef: pending.historyItemRef,
          })}
          role="alertdialog"
          style={confirmationStyle}
        >
          <p style={metadataStyle}>
            Confirm {pending.action} rollback for record{" "}
            <code>{pending.recordId}</code> at row version {pending.rowVersion}.
          </p>
          <div style={actionsStyle}>
            <button
              data-testid={rowHistoryRollbackConfirmButtonTestId({
                action: pending.action,
                historyItemRef: pending.historyItemRef,
              })}
              type="button"
              onClick={() => void confirmRollback()}
            >
              Confirm rollback
            </button>
            <button
              data-testid={rowHistoryRollbackCancelButtonTestId({
                action: pending.action,
                historyItemRef: pending.historyItemRef,
              })}
              type="button"
              onClick={() => setPending(null)}
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </section>
  );
}

function rollbackLabel(
  _item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): string {
  if (action === "history_entry") return "Rollback entry";
  if (action === "change_set") return "Rollback change set";
  return "Restore row fields";
}

const panelStyle = {
  display: "grid",
  gap: "0.6rem",
  paddingBlock: "0.5rem",
} satisfies CSSProperties;
const titleStyle = { margin: 0 } satisfies CSSProperties;
const metadataStyle = {
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const errorStyle = {
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
  margin: 0,
} satisfies CSSProperties;
const listStyle = {
  display: "grid",
  gap: "0.5rem",
  listStyle: "none",
  margin: 0,
  padding: 0,
} satisfies CSSProperties;
const itemStyle = {
  border: "var(--ct-border-hairline)",
  display: "grid",
  gap: "0.35rem",
  padding: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "0.4rem",
} satisfies CSSProperties;
const confirmationStyle = {
  border: "var(--ct-border-focus)",
  display: "grid",
  gap: "0.5rem",
  padding: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
