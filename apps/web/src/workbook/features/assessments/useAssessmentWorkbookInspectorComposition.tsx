import type { GridInteractionMode } from "@cartulary/grid-adapter";
import { assessmentCreateControlTestId } from "@cartulary/ui-contracts";
import type {
  InspectorDisabledCondition,
  ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useMemo, useState } from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { WorkbookRecordCandidatePicker } from "../../components/WorkbookRecordCandidatePicker";
import { useAssessmentSupportCandidates } from "../../hooks/useAssessmentSupportCandidates";
import { inspectorRecordHistoryActions } from "../../inspector/inspectorCapabilityResolver";
import { useInspectorCreateRelatedWorkflow } from "../../inspector/useInspectorCreateRelatedWorkflow";
import { useWorkbookInspectorCoordinator } from "../../inspector/useWorkbookInspectorCoordinator";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import {
  buildWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
} from "../../inspector/workbookInspectorSubject";
import { isAssessmentConfidenceBand } from "../../models/assessmentWorkbookModel";
import type { EntityRow } from "../../models/entityWorkbookModel";
import {
  enumValuesFor,
  genericCellLabel,
  genericInspectorRowLabel,
} from "../../models/genericWorkbookModel";
import { workbookInspectorStateIsOpen } from "../../models/workbookInspectorModel";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type {
  AssessmentMutationCommandPort,
  RecordRouteCommandPort,
  TimelineRelatedRecordPort,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { AssessmentWorkbookInspector } from "./AssessmentWorkbookInspector";
import { useAssessmentCreationController } from "./useAssessmentCreationController";

export function useAssessmentWorkbookInspectorComposition({
  canCreate,
  contract,
  currentIncidentRole,
  currentUserId,
  hostRows,
  identityRows,
  incidentClosed,
  inspectorResetKey,
  interactionMode,
  mutationCommands,
  mutationRuntime,
  onCaptureFocus,
  onClearSelectedAssessment,
  onClearSurfaceSelection,
  onRefreshAssessmentRows,
  onRestoreFocus,
  onSelectAssessment,
  recordMutationCommands,
  relatedMutationCommands,
  roleCanCreate,
  selectedAssessment,
  viewQuery,
}: {
  readonly canCreate: boolean;
  readonly contract: ViewContract;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly currentUserId: string | null;
  readonly hostRows: readonly EntityRow[];
  readonly identityRows: readonly EntityRow[];
  readonly incidentClosed: boolean;
  readonly inspectorResetKey: string;
  readonly interactionMode: GridInteractionMode;
  readonly mutationCommands: AssessmentMutationCommandPort;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onCaptureFocus: () => void;
  readonly onClearSelectedAssessment: () => void;
  readonly onClearSurfaceSelection: () => void;
  readonly onRefreshAssessmentRows: () => Promise<void>;
  readonly onRestoreFocus: () => void;
  readonly onSelectAssessment: (recordId: string) => void;
  readonly recordMutationCommands: RecordRouteCommandPort;
  readonly relatedMutationCommands: TimelineRelatedRecordPort;
  readonly roleCanCreate: boolean;
  readonly selectedAssessment: WorkbookQueryRow | null;
  readonly viewQuery: WorkbookViewQueryPort;
}) {
  const inspectorConfig = contract.inspectorConfig;
  const [deletedHistorySubject, setDeletedHistorySubject] =
    useState<WorkbookInspectorSubject | null>(null);
  const [relatedFeedback, setRelatedFeedback] =
    useState<WorkbookInspectorFeedback | null>(null);
  const supportCandidates = useAssessmentSupportCandidates({ viewQuery });
  const subjectRecordIds = useMemo(
    () => ({
      host: hostRows.map((row) => row.recordId),
      identity: identityRows.map((row) => row.recordId),
    }),
    [hostRows, identityRows],
  );
  const beginMutation = useCallback(
    () => mutationRuntime.beginExplicitMutation(),
    [mutationRuntime],
  );
  const creation = useAssessmentCreationController({
    beginMutation,
    lifecycleResetKey: inspectorResetKey,
    mutationCommands,
    onRefreshAssessmentRows,
    subjectRecordIds,
  });
  const { draft, draftMode, feedback, isSubmitting } = creation.snapshot;
  const subject: WorkbookInspectorSubject | null =
    selectedAssessment === null
      ? deletedHistorySubject
      : buildWorkbookInspectorSubject({
          config: inspectorConfig,
          kind: "live",
          label: genericInspectorRowLabel(contract, selectedAssessment),
          recordId: selectedAssessment.record_id,
          rowVersion: selectedAssessment.row_version,
          stateLabel: `Follow-on subject: ${draft.subjectRecordId || "not selected"}`,
          surfaceLabel: contract.title,
        });
  const relatedReferenceOptions = useMemo(
    () => emptyGenericReferenceOptions(),
    [],
  );
  const related = useInspectorCreateRelatedWorkflow({
    beginMutation,
    currentUserId,
    mutationCommands: relatedMutationCommands,
    onCreated: onRefreshAssessmentRows,
    onFeedback: setRelatedFeedback,
    selectedSubject:
      selectedAssessment === null || subject?.kind !== "live"
        ? null
        : { cells: selectedAssessment.cells, subject },
  });
  const inspector = useWorkbookInspectorCoordinator({
    actionPorts: {
      resetOwnerState: ({ cause, scope }) => {
        creation.commands.reset();
        setRelatedFeedback(null);
        if (cause === "close" || scope === "surface") {
          setDeletedHistorySubject(null);
        }
        if (scope === "surface") onClearSurfaceSelection();
      },
      restoreFocus: onRestoreFocus,
    },
    config: inspectorConfig,
    lifecycleKey: inspectorResetKey,
    subject,
  });
  const isOpen = workbookInspectorStateIsOpen(inspector.snapshot);
  const recordHistoryActions = useMemo(
    () => inspectorRecordHistoryActions(inspectorConfig),
    [inspectorConfig],
  );
  const disabledTokens = useMemo(() => {
    const tokens = new Set<InspectorDisabledCondition>();
    if (selectedAssessment === null) tokens.add("no_row_selected");
    else tokens.add("record_not_deleted");
    tokens.add("rollback_target_unavailable");
    if (!roleCanCreate) tokens.add("authorization_lost");
    else if (incidentClosed) tokens.add("incident_closed");
    return tokens;
  }, [incidentClosed, roleCanCreate, selectedAssessment]);
  const stateOptions = enumValuesFor(contract, "assessment.assessment_state", [
    "unknown",
    "suspected",
    "confirmed",
    "disproven",
    "cleared",
  ]);
  const confidenceBandOptions = enumValuesFor(
    contract,
    "assessment.confidence_band",
    ["unset", "low", "medium", "high"],
  ).filter(isAssessmentConfidenceBand);
  const subjectRows = draft.subjectType === "host" ? hostRows : identityRows;
  const close = () => {
    creation.commands.cancel();
    inspector.commands.close({ restoreFocus: true });
  };
  const node = isOpen ? (
    <AssessmentWorkbookInspector
      config={inspectorConfig}
      currentIncidentRole={currentIncidentRole}
      disabledTokens={disabledTokens}
      draftMode={draftMode}
      feedback={feedback}
      feedbackTestId={assessmentCreateControlTestId("message")}
      followOn={{
        canCreate,
        open: () => creation.commands.openFollowOn(selectedAssessment),
        opened: () => {
          if (!isOpen) onCaptureFocus();
          inspector.commands.open();
        },
        reject: creation.commands.rejectStart,
      }}
      history={{
        beginMutation,
        actions: recordHistoryActions,
        canMutate:
          canCreate &&
          interactionMode.kind === "editable" &&
          currentIncidentRole !== null &&
          currentIncidentRole !== "viewer",
        commands: recordMutationCommands,
        effects: {
          deleteAccepted: async (accepted) => {
            related.commands.cancel();
            creation.commands.reset();
            onClearSelectedAssessment();
            setDeletedHistorySubject(
              buildWorkbookInspectorSubject({
                config: inspectorConfig,
                kind: "deleted",
                label: "Deleted assessment",
                recordId: accepted.recordId,
                rowVersion: accepted.rowVersion,
                stateLabel: "Deleted",
                surfaceLabel: contract.title,
              }),
            );
            await onRefreshAssessmentRows();
          },
          restoreAccepted: async (accepted) => {
            await onRefreshAssessmentRows();
            setDeletedHistorySubject(null);
            onSelectAssessment(accepted.recordId);
          },
          rollbackAccepted: onRefreshAssessmentRows,
        },
      }}
      relationshipsContent={
        selectedAssessment === null ? null : (
          <p style={bodyStyle}>
            Supporting records:{" "}
            {genericCellLabel(
              selectedAssessment.cells["assessment.support_refs"]?.value,
            )}
          </p>
        )
      }
      related={{
        begin: related.commands.begin,
        cancel: related.commands.cancel,
        referenceOptions: relatedReferenceOptions,
        state: related.snapshot.workflow,
        submit: related.commands.submit,
        updateDraft: related.commands.updateDraft,
      }}
      relatedFeedback={relatedFeedback}
      subject={subject}
      onClose={close}
      workflowContent={
        <>
          <label style={labelStyle}>
            Subject type
            <select
              data-testid={assessmentCreateControlTestId("subject-type")}
              disabled={isSubmitting}
              style={selectStyle}
              value={draft.subjectType}
              onChange={(event) => {
                const subjectType =
                  event.target.value === "identity" ? "identity" : "host";
                const nextRows =
                  subjectType === "host" ? hostRows : identityRows;
                creation.commands.updateDraft((current) => ({
                  ...current,
                  subjectType,
                  subjectRecordId: nextRows[0]?.recordId ?? "",
                }));
              }}
            >
              {enumValuesFor(contract, "assessment.subject_type", [
                "host",
                "identity",
              ]).map((value) => (
                <option key={value} value={value}>
                  {value}
                </option>
              ))}
            </select>
          </label>
          <label style={labelStyle}>
            Subject
            <select
              data-testid={assessmentCreateControlTestId("subject")}
              disabled={isSubmitting}
              style={selectStyle}
              value={draft.subjectRecordId}
              onChange={(event) =>
                creation.commands.updateDraft((current) => ({
                  ...current,
                  subjectRecordId: event.target.value,
                }))
              }
            >
              <option value="">Select subject</option>
              {subjectRows.map((row) => (
                <option key={row.recordId} value={row.recordId}>
                  {row.label}
                </option>
              ))}
            </select>
          </label>
          <label style={labelStyle}>
            State
            <select
              data-testid={assessmentCreateControlTestId("state")}
              disabled={isSubmitting}
              style={selectStyle}
              value={draft.assessmentState}
              onChange={(event) =>
                creation.commands.updateDraft((current) => ({
                  ...current,
                  assessmentState: event.target.value,
                }))
              }
            >
              {stateOptions.map((value) => (
                <option key={value} value={value}>
                  {value}
                </option>
              ))}
            </select>
          </label>
          <label style={labelStyle}>
            Confidence
            <select
              data-testid={assessmentCreateControlTestId("confidence-band")}
              disabled={isSubmitting}
              style={selectStyle}
              value={draft.confidenceBand}
              onChange={(event) => {
                const confidenceBand = isAssessmentConfidenceBand(
                  event.target.value,
                )
                  ? event.target.value
                  : "unset";
                creation.commands.updateDraft((current) => ({
                  ...current,
                  confidenceBand,
                }));
              }}
            >
              {confidenceBandOptions.map((value) => (
                <option key={value} value={value}>
                  {value}
                </option>
              ))}
            </select>
          </label>
          <label style={labelStyle}>
            Rationale
            <textarea
              data-testid={assessmentCreateControlTestId("rationale")}
              disabled={isSubmitting}
              rows={4}
              style={textareaStyle}
              value={draft.rationale}
              onChange={(event) =>
                creation.commands.updateDraft((current) => ({
                  ...current,
                  rationale: event.target.value,
                }))
              }
            />
          </label>
          <label style={labelStyle}>
            Assessed
            <input
              data-testid={assessmentCreateControlTestId("assessed-at")}
              disabled={isSubmitting}
              placeholder="RFC3339 timestamp"
              style={inputStyle}
              type="text"
              value={draft.assessedAt}
              onChange={(event) =>
                creation.commands.updateDraft((current) => ({
                  ...current,
                  assessedAt: event.target.value,
                }))
              }
            />
          </label>
          <WorkbookRecordCandidatePicker
            candidates={supportCandidates}
            disabled={isSubmitting}
            label="Support refs"
            selectedRecordIds={draft.supportRecordIds}
            testId={assessmentCreateControlTestId("support-refs")}
            onSelectedRecordIdsChange={(supportRecordIds) =>
              creation.commands.updateDraft((current) => ({
                ...current,
                supportRecordIds,
              }))
            }
          />
          <button
            data-testid={assessmentCreateControlTestId("submit")}
            disabled={!canCreate || isSubmitting}
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => void creation.commands.submit(canCreate)}
          >
            Create assessment
          </button>
        </>
      }
    />
  ) : undefined;

  return {
    close,
    node,
    open: inspector.commands.open,
    openStandalone: (defaultHostRecordId: string) => {
      creation.commands.openStandalone(defaultHostRecordId);
      inspector.commands.open();
    },
  };
}

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
};
const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
};
const textareaStyle = { ...inputStyle, resize: "vertical" as const };
const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};
const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};
const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
};
const selectStyle = { ...inputStyle, appearance: "auto" as const };
