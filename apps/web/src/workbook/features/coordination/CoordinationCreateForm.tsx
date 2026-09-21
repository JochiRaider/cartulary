import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
} from "@cartulary/ui-contracts";
import { useId, useLayoutEffect, useRef, useSyncExternalStore } from "react";
import { WorkbookAuthoringReferenceControl } from "../../components/WorkbookAuthoringReferenceControl";
import {
  workbookFormInputStyle as fieldStyle,
  workbookFormFieldsStyle as groupStyle,
  workbookFormActionsStyle,
  workbookFormGroupStyle,
  workbookFormHeadingStyle,
  workbookFormMessageStyle,
} from "../../components/workbookFormStyles";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import {
  coordinationIds,
  coordinationReferenceView,
  coordinationSourceInput,
  coordinationSourceViews,
} from "./coordinationCreateModel";
import type { WorkbookCoordinationCreateOwner } from "./WorkbookCoordinationCreateOwner";

export function CoordinationCreateForm({
  owner,
  attachment,
  onSubmit,
}: {
  readonly owner: WorkbookCoordinationCreateOwner;
  readonly attachment: symbol;
  readonly onSubmit: () => void;
}) {
  const state = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const id = useId();
  const root = useRef<HTMLElement>(null);
  const trigger = useRef(document.activeElement);
  useLayoutEffect(() => {
    (
      root.current?.querySelector<HTMLElement>(
        '[data-create-required="true"]',
      ) ?? root.current?.querySelector<HTMLElement>("input,textarea,select")
    )?.focus({ preventScroll: true });
    const container = root.current;
    return () => {
      if (
        (container?.contains(document.activeElement) ||
          document.activeElement === document.body) &&
        trigger.current instanceof HTMLElement &&
        trigger.current.isConnected &&
        owner.getSnapshot().authority
      )
        trigger.current.focus({ preventScroll: true });
    };
  }, [owner]);
  useLayoutEffect(() => {
    if (Object.keys(state.errors).length)
      root.current
        ?.querySelector<HTMLElement>('[aria-invalid="true"]')
        ?.focus({ preventScroll: true });
  }, [state.errors]);
  const draft = state.draft,
    reader = owner.getReader();
  if (!draft || state.attachment !== attachment) return null;
  const disabled = owner.busy || !owner.canSubmit();
  const actions = owner.captureDraftActions(attachment);
  const returnFocus = () => {
    if (
      trigger.current instanceof HTMLElement &&
      trigger.current.isConnected &&
      owner.getSnapshot().authority
    )
      trigger.current.focus({ preventScroll: true });
  };
  return (
    <section
      ref={root}
      aria-label={`Create ${draft.target.title}`}
      style={groupStyle}
    >
      <p style={workbookFormMessageStyle}>
        Closing keeps unfinished work in this session. Only the explicit create
        action saves it.
      </p>
      <fieldset disabled={disabled} style={workbookFormGroupStyle}>
        <legend style={workbookFormHeadingStyle}>
          Create {draft.target.title}
        </legend>
        <p style={{ margin: 0 }}>
          {draft.source
            ? "Source context will be saved as a related artifact link."
            : "Choose a source to save a related artifact link."}{" "}
          Target references are separate.
        </p>
        {draft.target.fields
          .filter((field) => field.createWritable)
          .map((field) => {
            const key = field.fieldKey,
              view = coordinationReferenceView(field);
            const value = draft.values[key] ?? "";
            const fieldId = `${id}-${key}`;
            const error = state.errors[key];
            const required =
              draft.target.minimumCreateFieldSets[0]?.includes(key);
            const defaultActor = [
              "handoff.outgoing_owner_user_id",
              "status_review.review_owner_user_id",
              "lesson.owner_user_id",
            ].includes(key);
            return (
              <div key={key} style={groupStyle}>
                {view && reader ? (
                  <WorkbookAuthoringReferenceControl
                    targetKey={`coordination:${draft.id}:${key}`}
                    maximum={field.readKind === "collection" ? 64 : 1}
                    required={required}
                    label={field.label}
                    errorId={error ? `${fieldId}-error` : undefined}
                    testId={genericCreateFieldTestId(key)}
                    views={[view]}
                    multiple={field.readKind === "collection"}
                    selected={coordinationIds(value).map((recordId) => ({
                      recordId,
                      displayText: draft.labels[recordId] ?? recordId,
                      viewSchemaId: view,
                    }))}
                    reader={reader}
                    revision={state.candidateRevision}
                    disabled={disabled}
                    onApply={(items) => {
                      if (
                        !items.length &&
                        field.readKind !== "collection" &&
                        !required
                      )
                        actions.omit(key);
                      else
                        actions.update(
                          key,
                          items.map((i) => i.recordId).join("\n"),
                          Object.fromEntries(
                            items.map((i) => [i.recordId, i.displayText]),
                          ),
                        );
                    }}
                  />
                ) : (
                  <label htmlFor={fieldId} style={groupStyle}>
                    {field.label}
                    {required ? " (required)" : ""}
                    {field.enumValues ? (
                      <select
                        data-create-required={required}
                        aria-required={required}
                        id={fieldId}
                        data-testid={genericCreateFieldTestId(key)}
                        style={fieldStyle}
                        value={value}
                        aria-invalid={!!error}
                        aria-describedby={
                          error ? `${fieldId}-error` : undefined
                        }
                        onChange={(event) => {
                          if (!event.currentTarget.value && !required)
                            actions.omit(key);
                          else actions.update(key, event.currentTarget.value);
                        }}
                      >
                        <option value="">
                          {key === "lesson.closure_state"
                            ? "Open (default)"
                            : "Choose a value"}
                        </option>
                        {field.enumValues.map((v) => (
                          <option key={v} value={v}>
                            {v}
                          </option>
                        ))}
                      </select>
                    ) : field.stringContractId === "multiline_body_v1" ||
                      key === "handoff.open_risk_refs" ? (
                      <textarea
                        data-create-required={required}
                        aria-required={required}
                        id={fieldId}
                        data-testid={genericCreateFieldTestId(key)}
                        rows={3}
                        style={fieldStyle}
                        value={value}
                        aria-invalid={!!error}
                        aria-describedby={
                          error ? `${fieldId}-error` : undefined
                        }
                        onChange={(event) =>
                          actions.update(key, event.currentTarget.value)
                        }
                      />
                    ) : (
                      <input
                        data-create-required={required}
                        aria-required={required}
                        id={fieldId}
                        data-testid={genericCreateFieldTestId(key)}
                        style={fieldStyle}
                        value={value}
                        aria-invalid={!!error}
                        aria-describedby={
                          error ? `${fieldId}-error` : undefined
                        }
                        placeholder={
                          field.directScalarContractId
                            ? "YYYY-MM-DDTHH:mm:ssZ"
                            : undefined
                        }
                        onChange={(event) =>
                          actions.update(key, event.currentTarget.value)
                        }
                      />
                    )}
                  </label>
                )}
                {required && view ? (
                  <span>Required {field.label.toLowerCase()}.</span>
                ) : null}
                {defaultActor && !Object.hasOwn(draft.values, key) ? (
                  <span>Uses the current actor when omitted.</span>
                ) : null}
                {key.endsWith(".timestamp_utc") &&
                !Object.hasOwn(draft.values, key) ? (
                  <span>Uses the commit time when omitted.</span>
                ) : null}
                {key === "handoff.open_risk_refs" ? (
                  <span>
                    One risk per line. Risks are text references within this
                    Handoff.
                  </span>
                ) : null}
                {key === "comm_log.audience" ? (
                  <span>
                    Audience text is required even when audience Parties are
                    selected.
                  </span>
                ) : null}
                {field.clearable && field.readKind !== "collection" ? (
                  <Button
                    type="button"
                    tone="secondary"
                    onClick={() => actions.update(key, null)}
                  >
                    Clear {field.label.toLowerCase()}
                  </Button>
                ) : null}
                {!required && Object.hasOwn(draft.values, key) ? (
                  <Button
                    type="button"
                    tone="secondary"
                    onClick={() => actions.omit(key)}
                  >
                    Use default for {field.label.toLowerCase()}
                  </Button>
                ) : null}
                {draft.values[key] === null ? (
                  <span>Will be cleared.</span>
                ) : null}
                {error ? (
                  <span id={`${fieldId}-error`} role="alert">
                    {error}
                  </span>
                ) : null}
              </div>
            );
          })}
        {reader ? (
          <WorkbookAuthoringReferenceControl
            targetKey={`coordination:${draft.id}:source`}
            maximum={1}
            captureRowVersion
            label="Source"
            errorId={
              state.errors[coordinationSourceInput]
                ? `${id}-source-error`
                : undefined
            }
            testId="coordination-source-record"
            views={coordinationSourceViews(draft.variant)}
            multiple={false}
            selected={
              draft.source
                ? [
                    {
                      recordId: draft.source.recordId,
                      viewSchemaId: draft.source.viewSchemaId,
                      rowVersion: draft.source.rowVersion,
                      displayText: draft.source.label || draft.source.recordId,
                    },
                  ]
                : []
            }
            reader={reader}
            revision={state.candidateRevision}
            disabled={disabled}
            onApply={(items) => {
              const item = items[0];
              if (!item) actions.changeSource(null);
              else if (item.rowVersion !== undefined)
                actions.changeSource({
                  recordId: item.recordId,
                  viewSchemaId: item.viewSchemaId,
                  rowVersion: item.rowVersion,
                  label: item.displayText,
                });
            }}
          />
        ) : null}
        {draft.source ? (
          <Button
            type="button"
            tone="secondary"
            onClick={() => actions.changeSource(null)}
          >
            Clear source
          </Button>
        ) : (
          <span>No source link will be saved.</span>
        )}
        {state.errors[coordinationSourceInput] ? (
          <span id={`${id}-source-error`} role="alert">
            {state.errors[coordinationSourceInput]}
          </span>
        ) : null}
      </fieldset>
      {state.message ? <p role="status">{state.message}</p> : null}
      {state.needsReview ? (
        <Button
          tone="secondary"
          type="button"
          disabled={disabled}
          onClick={() => void actions.review()}
        >
          Review source and access
        </Button>
      ) : null}
      <div style={workbookFormActionsStyle}>
        <Button
          tone="primary"
          type="button"
          data-testid={genericCreateSubmitTestId(draft.target.viewSchemaId)}
          disabled={disabled || state.needsReview}
          onClick={() => {
            if (actions.validate()) onSubmit();
            else
              root.current
                ?.querySelector<HTMLElement>('[aria-invalid="true"]')
                ?.focus({ preventScroll: true });
          }}
        >
          Create {draft.target.title}
        </Button>
        <Button
          tone="secondary"
          type="button"
          onClick={() => {
            actions.detach();
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
            actions.discard();
            returnFocus();
          }}
        >
          Discard draft
        </Button>
      </div>
    </section>
  );
}
