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
import {
  WorkbookHistoryEvent,
  WorkbookHistoryList,
} from "./presentation/WorkbookHistoryPresentation";
import {
  WorkbookInspectorConfirmation,
  WorkbookInspectorPublicError,
  WorkbookInspectorTechnicalDetails,
} from "./presentation/WorkbookInspectorPresentation";
import type { WorkbookHistoryEventPresentation } from "./presentation/workbookInspectorPresentationModel";

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
      <WorkbookInspectorTechnicalDetails
        fields={[
          { label: "Record ID", value: subject.recordId },
          { label: "Row version", value: String(subject.rowVersion) },
        ]}
      />
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
        <WorkbookInspectorPublicError
          message={message}
          testId={rowHistoryMessageTestId()}
        />
      ) : null}
      {data === null ? null : (
        <WorkbookHistoryList>
          {data.items.map((item) => {
            const event = recordHistoryEventPresentation(item);
            return (
              <WorkbookHistoryEvent
                actions={
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
                          const target =
                            buildRecordRollbackTargetFromHistoryAction(
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
                        {rollbackLabel(action)}
                      </button>
                    ))}
                  </div>
                }
                event={event}
                key={event.key}
                testId={rowHistoryItemTestId({
                  historyItemRef: item.history_item_ref,
                })}
              />
            );
          })}
        </WorkbookHistoryList>
      )}
      {pending === null ? null : (
        <WorkbookInspectorConfirmation
          cancelTestId={rowHistoryRollbackCancelButtonTestId({
            action: pending.action,
            historyItemRef: pending.historyItemRef,
          })}
          confirmLabel="Confirm rollback"
          confirmTestId={rowHistoryRollbackConfirmButtonTestId({
            action: pending.action,
            historyItemRef: pending.historyItemRef,
          })}
          operation={`${rollbackLabel(pending.action)} rollback`}
          subject="this history state"
          technicalFields={[
            { label: "Record ID", value: pending.recordId },
            { label: "Row version", value: String(pending.rowVersion) },
            { label: "History item", value: pending.historyItemRef },
          ]}
          testId={rowHistoryRollbackPreviewTestId({
            action: pending.action,
            historyItemRef: pending.historyItemRef,
          })}
          onCancel={() => setPending(null)}
          onConfirm={() => void confirmRollback()}
        />
      )}
    </section>
  );
}

function rollbackLabel(action: RecordHistoryRollbackAction): string {
  if (action === "history_entry") return "Rollback entry";
  if (action === "change_set") return "Rollback change set";
  return "Restore row fields";
}

function recordHistoryEventPresentation(
  item: RecordHistoryItem,
): WorkbookHistoryEventPresentation {
  return {
    actorLabel: item.actor_user_id,
    committedAt: item.committed_at,
    key: item.history_item_ref,
    operation: item.operation,
    summary: item.diff_summary.summary,
    technicalFields: [
      { label: "Actor ID", value: item.actor_user_id },
      { label: "History reference", value: item.history_item_ref },
      { label: "Change set ID", value: item.change_set_id },
      ...(item.history_entry_ref === undefined
        ? []
        : [{ label: "History entry", value: item.history_entry_ref }]),
      ...(item.revision_no === undefined
        ? []
        : [{ label: "Revision", value: String(item.revision_no) }]),
    ],
  };
}

const panelStyle = {
  display: "grid",
  gap: "0.6rem",
  paddingBlock: "0.5rem",
} satisfies CSSProperties;
const metadataStyle = {
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "0.4rem",
} satisfies CSSProperties;
