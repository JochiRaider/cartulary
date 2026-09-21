import {
  genericCreateFieldTestId,
  mentionCreateEntityButtonTestId,
  mentionDismissButtonTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
} from "@cartulary/ui-contracts";
import { useRef, useState } from "react";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import { WorkbookRecordCandidatePicker } from "../../components/WorkbookRecordCandidatePicker";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import {
  type MentionCreationOperation,
  mentionCreateRequest,
  mentionEntityContract,
} from "../actions/timelineMentionCreationModel";
import {
  type MentionOperation,
  mentionTransitionAllowed,
} from "../actions/timelineMentionOperationModel";
import type { useTimelineMentionActions } from "../hooks/useTimelineMentionActions";
import { inputStyle, labelStyle } from "./TimelineWorkbookStyles";

export type TimelineMentionActions = ReturnType<
  typeof useTimelineMentionActions
>;
export function TimelineMentionActionControls({
  actions,
}: {
  readonly actions: TimelineMentionActions;
}) {
  const { owner, subject, candidates, snapshot, createReview } = actions;
  const [filter, setFilter] = useState("");
  const [correcting, setCorrecting] = useState(true);
  const correction = useRef<HTMLDetailsElement>(null);
  if (!subject)
    return (
      <p role="status">
        Mention identity is unavailable. Refresh before acting.
      </p>
    );
  const blocked = owner.blocksMention(subject.mentionId);
  const operation = [...snapshot.entries]
    .reverse()
    .find(
      (entry) => entry.attempt.review.subject.mentionId === subject.mentionId,
    );
  const creation = owner.creationForMention(subject.mentionId);
  const matches = candidates.candidates.filter((candidate) =>
    candidate.displayText
      .toLocaleLowerCase()
      .includes(filter.toLocaleLowerCase()),
  );
  const selected = candidates.candidates.find(
    (candidate) => candidate.recordId === actions.selectedTargetId,
  );
  const options =
    selected && !matches.includes(selected) ? [selected, ...matches] : matches;
  const allowed = (action: Parameters<typeof owner.canSubmit>[0]) =>
    owner.canSubmit(action) && !blocked;
  return (
    <div style={{ display: "grid", gap: "var(--ct-spacing-sm)", minWidth: 0 }}>
      {!owner.canSubmit("resolve_item") ? (
        <p role="status">
          Mention actions require current editor access and an open incident.
        </p>
      ) : null}
      <details
        ref={correction}
        open={correcting}
        onToggle={(event) => setCorrecting(event.currentTarget.open)}
        onKeyDown={(event) => {
          if (
            event.key === "Escape" &&
            correcting &&
            !event.defaultPrevented &&
            !event.nativeEvent.isComposing
          ) {
            event.preventDefault();
            event.stopPropagation();
            setCorrecting(false);
            correction.current
              ?.querySelector("summary")
              ?.focus({ preventScroll: true });
          }
        }}
      >
        <summary>Correction and resolution</summary>
        <div
          style={{
            display: "grid",
            gap: "var(--ct-spacing-sm)",
            paddingBlockStart: "var(--ct-spacing-xs)",
          }}
        >
          {subject.state !== "dismissed" ? (
            <>
              <label style={labelStyle}>
                Filter loaded targets
                <input
                  style={inputStyle}
                  value={filter}
                  onChange={(event) => setFilter(event.currentTarget.value)}
                />
              </label>
              <WorkbookRecordCandidatePicker
                selection="single"
                label={
                  subject.state === "resolved"
                    ? "Correct target"
                    : "Resolve to existing"
                }
                testId={mentionResolveTargetSelectTestId()}
                disabled={!allowed("resolve_item")}
                candidates={options}
                selectedRecordIds={
                  actions.selectedTargetId ? [actions.selectedTargetId] : []
                }
                onSelectedRecordIdsChange={(ids) =>
                  actions.changeTarget(ids[0] ?? "")
                }
              />
              <p role="status" style={{ margin: 0 }}>
                {candidates.phase === "loading" || candidates.phase === "idle"
                  ? "Loading targets…"
                  : candidates.phase === "failed"
                    ? candidates.error
                    : candidates.candidates.length === 0
                      ? candidates.hasMore
                        ? "No targets on the loaded pages. More targets are available."
                        : "No eligible targets found in this search."
                      : matches.length === 0
                        ? "No loaded targets match this filter."
                        : `${candidates.candidates.length} targets loaded.${candidates.hasMore ? " More targets are available." : " All current pages loaded."}`}
              </p>
              <div style={actionsStyle}>
                {candidates.phase === "failed" ? (
                  <WorkbookInspectorActionButton
                    tone="secondary"
                    onClick={() => void candidates.retry()}
                  >
                    Retry target read
                  </WorkbookInspectorActionButton>
                ) : null}
                {candidates.hasMore ? (
                  <WorkbookInspectorActionButton
                    tone="secondary"
                    disabled={candidates.phase === "loading"}
                    onClick={() => void candidates.loadMore()}
                  >
                    Load more targets
                  </WorkbookInspectorActionButton>
                ) : null}
                <WorkbookInspectorActionButton
                  tone="secondary"
                  data-testid={mentionResolveExistingButtonTestId()}
                  disabled={!allowed("resolve_item") || !selected}
                  onClick={() =>
                    actions.act({
                      action: "resolve_item",
                      resolvedRecordId: actions.selectedTargetId,
                    })
                  }
                >
                  {subject.state === "resolved"
                    ? "Correct target"
                    : "Resolve to existing"}
                </WorkbookInspectorActionButton>
                <WorkbookInspectorActionButton
                  tone="secondary"
                  data-testid={mentionDismissButtonTestId()}
                  disabled={!allowed("dismiss_item")}
                  onClick={() => actions.act({ action: "dismiss_item" })}
                >
                  Dismiss
                </WorkbookInspectorActionButton>
              </div>
            </>
          ) : (
            <p>
              Dismissed mentions stay out of active relationships. Restore
              clears resolution and does not relink a previous target.
            </p>
          )}
          {mentionTransitionAllowed(subject.state, "revert_to_unresolved") ? (
            <WorkbookInspectorActionButton
              tone="secondary"
              data-testid={mentionRestoreUnresolvedButtonTestId()}
              disabled={!allowed("revert_to_unresolved")}
              onClick={() => actions.act({ action: "revert_to_unresolved" })}
            >
              {subject.state === "dismissed"
                ? "Restore to unresolved"
                : "Revert to unresolved"}
            </WorkbookInspectorActionButton>
          ) : null}
          {subject.state === "unresolved" &&
          !creation?.receipt &&
          !createReview ? (
            <WorkbookInspectorActionButton
              tone="secondary"
              data-testid={mentionCreateEntityButtonTestId(subject.entityType)}
              disabled={!owner.canCreate(subject.entityType) || blocked}
              onClick={actions.startCreate}
            >
              Create {subject.entityType}
            </WorkbookInspectorActionButton>
          ) : null}
          {createReview && !creation?.receipt ? (
            <MentionCreateEditor
              actions={actions}
              disabled={blocked || !owner.canCreate(subject.entityType)}
            />
          ) : null}
        </div>
      </details>
      {!correcting && createReview && !creation?.receipt ? (
        <p role="status">
          Unfinished entity creation is retained. Open Correction and resolution
          to continue.
        </p>
      ) : null}
      {creation?.receipt ? (
        <section
          aria-label="Entity created from mention"
          style={{ overflowWrap: "anywhere" }}
        >
          <p>
            {subject.entityType === "host" ? "Host" : "Identity"} saved:{" "}
            {String(
              creation.receipt.data.row.cells[
                `${subject.entityType}.display_name`
              ]?.value ?? creation.receipt.data.row.record_id,
            )}
            .
          </p>
          <MentionCreationRefreshFeedback
            creation={creation}
            refresh={() => void owner.refreshCreation(creation.key)}
          />
          {creation.linkKey === null ||
          (operation &&
            ["rejected", "preparation_failed"].includes(operation.phase)) ? (
            <>
              <p>
                The entity is saved. Review this mention to finish resolving it.
              </p>
              <WorkbookInspectorActionButton
                tone="primary"
                disabled={
                  !owner.canSubmit("resolve_item") ||
                  operation?.phase === "uncertain"
                }
                onClick={() => actions.linkCreated(creation.key)}
              >
                Resolve to created {subject.entityType}
              </WorkbookInspectorActionButton>
            </>
          ) : null}
        </section>
      ) : creation ? (
        <p role="status">
          {creation.phase === "uncertain"
            ? "Creation outcome is uncertain. Use Mention operations to replay the original request."
            : creation.phase === "rejected" ||
                creation.phase === "preparation_failed"
              ? creation.failure?.message
              : "Creating entity…"}
        </p>
      ) : null}
      {operation ? (
        <MentionOperationFeedback
          operation={operation}
          replay={() => void owner.replay(operation.key)}
          refresh={() => void owner.refresh(operation.key)}
          canReplay={owner.canSubmit(operation.attempt.review.intent.action)}
        />
      ) : null}
    </div>
  );
}
function MentionCreateEditor({
  actions,
  disabled,
}: {
  readonly actions: TimelineMentionActions;
  readonly disabled: boolean;
}) {
  const review = actions.createReview;
  if (!review) return null;
  const contract = mentionEntityContract(review.subject.entityType);
  const identifiers = new Set(
    review.subject.entityType === "host"
      ? ["host.hostname", "host.fqdn", "host.aad_device_id"]
      : [
          "identity.upn",
          "identity.email",
          "identity.sam_account_name",
          "identity.sid",
          "identity.aad_object_id",
        ],
  );
  const fields = contract.fields.filter((field) => field.createWritable);
  const primary = fields.filter(
    (field) =>
      field.fieldKey === `${review.subject.entityType}.display_name` ||
      identifiers.has(field.fieldKey),
  );
  const additional = fields.filter((field) => !primary.includes(field));
  const renderField = (field: (typeof fields)[number]) => (
    <label
      key={field.fieldKey}
      htmlFor={`mention-create-${field.fieldKey}`}
      style={labelStyle}
    >
      {field.label}
      <GenericMutationControl
        field={field}
        collectionMode="add"
        id={`mention-create-${field.fieldKey}`}
        testId={genericCreateFieldTestId(field.fieldKey)}
        value={review.draft[field.fieldKey] ?? ""}
        onChange={(value) => actions.updateCreateDraft(field.fieldKey, value)}
      />
    </label>
  );
  return (
    <form
      aria-label={`Create ${review.subject.entityType} from mention`}
      onSubmit={(event) => {
        event.preventDefault();
        actions.submitCreate();
      }}
    >
      <p>
        Review the entity details. Identifier fields are optional and establish
        matching identity only when you author them. Creation may reuse an
        existing exact match.
      </p>
      <fieldset
        disabled={disabled}
        style={{
          border: 0,
          padding: 0,
          margin: 0,
          minWidth: 0,
          display: "grid",
          gap: "var(--ct-spacing-sm)",
        }}
      >
        <legend>
          Create {review.subject.entityType} and resolve this mention
        </legend>
        {primary.map(renderField)}
        {additional.length ? (
          <details>
            <summary>More fields</summary>
            {additional.map(renderField)}
          </details>
        ) : null}
        <WorkbookInspectorActionButton
          tone="primary"
          type="submit"
          disabled={!mentionCreateRequest(review, "review-validation")}
        >
          Create {review.subject.entityType} and resolve
        </WorkbookInspectorActionButton>
      </fieldset>
      <WorkbookInspectorActionButton
        tone="secondary"
        disabled={disabled}
        onClick={actions.cancelCreate}
      >
        Cancel creation
      </WorkbookInspectorActionButton>
    </form>
  );
}
export function MentionCreationRefreshFeedback({
  creation,
  refresh,
}: {
  readonly creation: MentionCreationOperation;
  readonly refresh: () => void;
}) {
  if (!creation.receipt || creation.refresh === "complete") return null;
  return (
    <section aria-label="Created entity refresh">
      <p role="status">Entity saved; its sheet still needs refreshing.</p>
      <WorkbookInspectorActionButton
        tone="secondary"
        disabled={creation.refresh === "refreshing"}
        onClick={refresh}
      >
        Refresh created entity
      </WorkbookInspectorActionButton>
    </section>
  );
}
export function MentionOperationFeedback({
  operation,
  replay,
  refresh,
  canReplay,
}: {
  readonly operation: MentionOperation;
  readonly replay: () => void;
  readonly refresh: () => void;
  readonly canReplay: boolean;
}) {
  const message =
    operation.phase === "accepted"
      ? operation.refresh === "complete"
        ? "Mention action completed."
        : "Mention action completed; refresh is still required."
      : operation.phase === "uncertain"
        ? "The outcome is uncertain. The action may have committed. Replay uses the original request."
        : operation.phase === "preparing"
          ? "Waiting for earlier edits…"
          : operation.phase === "submitting"
            ? "Submitting mention action…"
            : (operation.failure?.message ??
              "The action was rejected. Review the mention before acting again.");
  return (
    <section aria-label="Mention operation status">
      <p role="status">{message}</p>
      {operation.phase === "accepted" &&
      operation.refresh === "required" &&
      operation.failure ? (
        <p>{operation.failure.message}</p>
      ) : null}
      {operation.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          disabled={!canReplay}
          onClick={replay}
        >
          Replay mention action
        </WorkbookInspectorActionButton>
      ) : null}
      {operation.receipt && operation.refresh !== "complete" ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          disabled={operation.refresh === "refreshing"}
          onClick={refresh}
        >
          Refresh mention result
        </WorkbookInspectorActionButton>
      ) : null}
    </section>
  );
}
const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
} as const;
