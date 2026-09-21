import type { GridInteractionMode } from "@cartulary/grid-adapter";
import { assessmentCreateControlTestId } from "@cartulary/ui-contracts";
import type {
  InspectorDisabledCondition,
  ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useMemo, useState } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import {
  workbookFormMessageStyle as bodyStyle,
  workbookFormInputStyle as inputStyle,
  workbookFormFieldStackStyle as labelStyle,
  workbookFormActionsStyle,
} from "../../components/workbookFormStyles";
import { useWorkbookHistorySurfaceRefresh } from "../../history/WorkbookHistoryContext";
import { inspectorRecordHistoryActions } from "../../inspector/inspectorCapabilityResolver";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { useInspectorCreateRelatedWorkflow } from "../../inspector/useInspectorCreateRelatedWorkflow";
import { useWorkbookInspectorCoordinator } from "../../inspector/useWorkbookInspectorCoordinator";
import { WorkbookInspectorReadOnlyDetails } from "../../inspector/WorkbookInspectorDetails";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import {
  buildWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
} from "../../inspector/workbookInspectorSubject";
import { isAssessmentConfidenceBand } from "../../models/assessmentWorkbookModel";
import {
  enumValuesFor,
  genericCellLabel,
  genericInspectorRowLabel,
} from "../../models/genericWorkbookModel";
import { workbookInspectorStateIsOpen } from "../../models/workbookInspectorModel";
import type {
  RecordRouteCommandPort,
  TimelineRelatedRecordPort,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import {
  AssessmentSubjectPicker,
  AssessmentSupportPicker,
} from "./AssessmentDiscovery";
import { AssessmentWorkbookInspector } from "./AssessmentWorkbookInspector";
import type { AssessmentCandidateReadPort } from "./assessmentCandidatePort";
import { useAssessmentCreationController } from "./useAssessmentCreationController";

export function useAssessmentWorkbookInspectorComposition({
  sheetRef,
  canCreate,
  contract,
  currentIncidentRole,
  currentUserId,
  candidateReader,
  incidentClosed,
  inspectorResetKey,
  interactionMode,
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
}: {
  readonly sheetRef: SheetRef;
  readonly canCreate: boolean;
  readonly contract: ViewContract;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly currentUserId: string | null;
  readonly candidateReader: AssessmentCandidateReadPort;
  readonly incidentClosed: boolean;
  readonly inspectorResetKey: string;
  readonly interactionMode: GridInteractionMode;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onCaptureFocus: () => void;
  readonly onClearSelectedAssessment: () => void;
  readonly onClearSurfaceSelection: () => void;
  readonly onRefreshAssessmentRows: (options?: {
    readonly requireAcceptance?: boolean;
  }) => Promise<void>;
  readonly onRestoreFocus: () => void;
  readonly onSelectAssessment: (recordId: string) => void;
  readonly recordMutationCommands: RecordRouteCommandPort;
  readonly relatedMutationCommands: TimelineRelatedRecordPort;
  readonly roleCanCreate: boolean;
  readonly selectedAssessment: WorkbookQueryRow | null;
}) {
  const inspectorConfig = contract.inspectorConfig;
  useWorkbookHistorySurfaceRefresh(inspectorConfig.viewSchemaId, () =>
    onRefreshAssessmentRows({ requireAcceptance: true }),
  );
  const [deletedHistorySubject, setDeletedHistorySubject] =
    useState<WorkbookInspectorSubject | null>(null);
  const [relatedFeedback, setRelatedFeedback] =
    useState<WorkbookInspectorFeedback | null>(null);
  const beginMutation = useCallback(
    () => mutationRuntime.beginExplicitMutation(),
    [mutationRuntime],
  );
  const creation = useAssessmentCreationController({
    owner: mutationRuntime.assessmentAuthoring,
    sheetRef,
    lifecycleResetKey: `${inspectorResetKey}:${selectedAssessment?.record_id ?? "none"}:${selectedAssessment?.row_version ?? 0}`,
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
          interactionMode.kind === "editable" &&
          currentIncidentRole !== null &&
          currentIncidentRole !== "viewer",
        commands: recordMutationCommands,
        effects: {
          deleteAccepted: (accepted) => {
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
          },
          restoreAccepted: (accepted) => {
            setDeletedHistorySubject(null);
            onSelectAssessment(accepted.recordId);
          },
          rollbackAccepted: () => {},
          refresh: () => onRefreshAssessmentRows({ requireAcceptance: true }),
        },
      }}
      detailsContent={
        selectedAssessment ? (
          <WorkbookInspectorReadOnlyDetails
            contract={contract}
            row={selectedAssessment}
          />
        ) : null
      }
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
        state: related.snapshot.workflow,
        submit: related.commands.submit,
        updateDraft: related.commands.updateDraft,
      }}
      relatedFeedback={relatedFeedback}
      subject={subject}
      onClose={close}
      workflowContent={
        !creation.snapshot.hasDraft ? (
          <Button
            tone="secondary"
            type="button"
            disabled={!canCreate}
            onClick={creation.commands.openStandalone}
          >
            Start another assessment
          </Button>
        ) : !creation.snapshot.attached ? (
          <>
            <p>
              Your unfinished assessment is retained. Review it before
              appending.
            </p>
            <Button
              tone="secondary"
              type="button"
              onClick={creation.commands.resume}
            >
              Resume assessment draft
            </Button>
            <Button
              tone="secondary"
              type="button"
              disabled={isSubmitting || !canCreate}
              onClick={creation.commands.discard}
            >
              Discard assessment draft
            </Button>
          </>
        ) : (
          <>
            <p>
              Closing retains this draft. After dispatch, closing does not
              cancel the append.
            </p>
            <label style={labelStyle}>
              Subject type
              <select
                data-testid={assessmentCreateControlTestId("subject-type")}
                disabled={isSubmitting || !canCreate}
                style={selectStyle}
                value={draft.subjectType}
                onChange={(event) => {
                  const subjectType =
                    event.target.value === "identity" ? "identity" : "host";
                  creation.commands.updateDraft((current) => ({
                    ...current,
                    subjectType,
                    subjectRecordId: "",
                    subjectDisplayText: "",
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
            <AssessmentSubjectPicker
              key={draft.subjectType}
              reader={candidateReader}
              draft={draft}
              disabled={isSubmitting || !canCreate}
              revision={creation.snapshot.candidateRevision}
              update={creation.commands.updateDraft}
            />
            <label style={labelStyle}>
              State
              <select
                data-testid={assessmentCreateControlTestId("state")}
                disabled={isSubmitting || !canCreate}
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
                disabled={isSubmitting || !canCreate}
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
                aria-invalid={!!creation.snapshot.errors.rationale}
                aria-describedby={
                  creation.snapshot.errors.rationale
                    ? "assessment-error-rationale"
                    : undefined
                }
                data-testid={assessmentCreateControlTestId("rationale")}
                disabled={isSubmitting || !canCreate}
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
              {creation.snapshot.errors.rationale ? (
                <span id="assessment-error-rationale">
                  {creation.snapshot.errors.rationale}
                </span>
              ) : null}
            </label>
            <label style={labelStyle}>
              Assessed
              <input
                aria-invalid={!!creation.snapshot.errors.assessedAt}
                aria-describedby={
                  creation.snapshot.errors.assessedAt
                    ? "assessment-error-assessed-at"
                    : undefined
                }
                data-testid={assessmentCreateControlTestId("assessed-at")}
                disabled={isSubmitting || !canCreate}
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
              {creation.snapshot.errors.assessedAt ? (
                <span id="assessment-error-assessed-at">
                  {creation.snapshot.errors.assessedAt}
                </span>
              ) : null}
            </label>
            <AssessmentSupportPicker
              reader={candidateReader}
              draft={draft}
              disabled={isSubmitting || !canCreate}
              revision={creation.snapshot.candidateRevision}
              update={creation.commands.updateDraft}
            />
            <div style={workbookFormActionsStyle}>
              <Button
                tone="primary"
                data-testid={assessmentCreateControlTestId("submit")}
                disabled={!canCreate || isSubmitting}
                type="button"
                onClick={() => void creation.commands.submit(canCreate)}
              >
                Create assessment
              </Button>
              <Button tone="secondary" onClick={creation.commands.cancel}>
                Close draft
              </Button>
              <Button
                tone="secondary"
                type="button"
                disabled={isSubmitting}
                onClick={creation.commands.discard}
              >
                Discard assessment draft
              </Button>
            </div>
          </>
        )
      }
    />
  ) : undefined;

  return {
    close,
    node,
    open: inspector.commands.open,
    openStandalone: () => {
      creation.commands.openStandalone();
      inspector.commands.open();
    },
  };
}

const textareaStyle = { ...inputStyle, resize: "vertical" as const };
const selectStyle = { ...inputStyle, appearance: "auto" as const };
