import { requireViewContract } from "@cartulary/view-contracts";
import { type RefCallback, useState } from "react";
import { workbookQueryStateFromSavedViewQueryJson } from "../models/workbookQuery";
import { canMutateSavedView } from "../models/workbookSavedViews";
import { projectWorkbookQueryEntries } from "../models/workbookViewBarWorkingSet";
import {
  type SavedViewSnapshot,
  savedViewOutcome,
} from "../savedviews/savedViewOperationModel";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";

export const savedViewRecoveryKeys = [
  "observe",
  "open",
  "submitted_summary",
  "observed_summary",
  "recovery_name",
  "confirm_create",
  "apply_review",
  "create_review",
  "keep_review",
  "dismiss",
] as const;
export type SavedViewRecoveryKey = (typeof savedViewRecoveryKeys)[number];

export function SavedViewRecovery({
  controller,
  snapshot,
  registerItem,
}: {
  readonly controller: WorkbookSavedViewController;
  readonly snapshot: SavedViewSnapshot;
  readonly registerItem: (
    key: SavedViewRecoveryKey,
  ) => RefCallback<HTMLElement>;
}) {
  const operation = snapshot.operation;
  const attempt = operation.kind === "idle" ? null : operation.attempt;
  const [confirmedAttempt, setConfirmedAttempt] = useState<number | null>(null);
  const recovery =
    operation.kind === "conflict" ||
    operation.kind === "uncertain" ||
    operation.kind === "rejected";
  const observation = snapshot.observation;
  const canReview =
    attempt !== null && controller.canReview(attempt.id, observation);
  const observed =
    attempt?.base && (attempt.kind === "update" || attempt.kind === "delete")
      ? snapshot.observations.get(attempt.base.saved_view_id)?.resource
      : null;
  const canApply =
    observed &&
    canMutateSavedView(
      observed,
      snapshot.authority?.actorId ?? null,
      snapshot.authority?.role ?? null,
    );
  const uncertainCreate =
    operation.kind === "uncertain" &&
    (attempt?.kind === "create" || attempt?.kind === "duplicate");
  const name = attempt
    ? (snapshot.reviewName ?? attempt.definition.displayName)
    : "";
  const nameError =
    recovery &&
    operation.problem.field === "display_name" &&
    name === attempt?.definition.displayName
      ? operation.problem.message
      : null;
  const nameErrorId = `saved-view-review-name-${attempt?.id ?? "none"}`;
  const showObservation = recovery || snapshot.resourceProblem !== null;
  if (operation.kind === "idle" && !showObservation) return null;
  return (
    <section
      aria-label="Saved-view operation"
      data-problem-kind={recovery ? operation.problem.kind : undefined}
      data-public-error-code={
        recovery ? operation.problem.publicCode : undefined
      }
      data-validation-field={recovery ? operation.problem.field : undefined}
      data-validation-reason={recovery ? operation.problem.reason : undefined}
      style={sectionStyle}
    >
      <p role={recovery ? "alert" : "status"} style={textStyle}>
        {savedViewOutcome(operation)}
      </p>
      {attempt && operation.kind === "pending" ? (
        <p style={textStyle}>
          Submitted: {attempt.definition.displayName} (
          {attempt.definition.scope}). You can keep editing while this request
          settles.
        </p>
      ) : null}
      {recovery ? (
        <p style={textStyle}>
          {problemLabel(operation.problem.kind)}: {operation.problem.message}
        </p>
      ) : null}
      {snapshot.resourceProblem ? (
        <p role="status" style={textStyle}>
          {operation.kind === "confirmed" ? "The write is confirmed. " : ""}The
          saved resource could not be refreshed:{" "}
          {snapshot.resourceProblem.message}
        </p>
      ) : null}
      {showObservation ? (
        <button
          ref={registerItem("observe")}
          type="button"
          style={buttonStyle}
          disabled={snapshot.refreshing || snapshot.access === "checking"}
          onClick={() => {
            if (snapshot.access !== "ready") void controller.recheckAccess();
            else void controller.refresh();
          }}
        >
          {snapshot.access !== "ready"
            ? "Check access and refresh saved resource"
            : snapshot.refreshing
              ? "Refreshing saved resource…"
              : "Refresh saved resource for review"}
        </button>
      ) : null}
      {operation.kind === "confirmed" && operation.resource ? (
        <button
          ref={registerItem("open")}
          type="button"
          style={buttonStyle}
          onClick={controller.openConfirmed}
        >
          Open confirmed saved view
        </button>
      ) : null}
      {recovery && attempt ? (
        <>
          <ConfigurationSummary
            summaryRef={registerItem("submitted_summary")}
            label="Submitted configuration"
            name={attempt.definition.displayName}
            scope={attempt.definition.scope}
            schema={attempt.definition.viewSchemaId}
            query={attempt.definition.queryJson}
            layout={attempt.definition.layoutJson}
          />
          {operation.problem.kind === "conflict" &&
          operation.problem.conflict ? (
            <p style={textStyle}>
              Submitted version {operation.problem.conflict.baseVersion}; the
              server reported version{" "}
              {operation.problem.conflict.currentVersion}.
            </p>
          ) : null}
          {observation === null ? (
            <p style={textStyle}>
              Refresh after the request settles to review current saved
              configuration.
            </p>
          ) : observed ? (
            <ConfigurationSummary
              summaryRef={registerItem("observed_summary")}
              label={`Observed saved configuration (version ${observed.saved_view_version})`}
              name={observed.display_name}
              scope={observed.scope}
              schema={observed.view_schema_id}
              query={observed.query_json}
              layout={observed.layout_json}
            />
          ) : (
            <p style={textStyle}>
              {attempt.kind === "update" || attempt.kind === "delete"
                ? "The addressed resource is unavailable. This does not establish the earlier write outcome."
                : "Creation cannot be resolved by matching a name or configuration."}
            </p>
          )}
          {operation.kind === "uncertain" ? (
            <p style={textStyle}>
              Current resource observations do not prove whether the earlier
              request committed. A matching name or configuration is not a
              receipt.
            </p>
          ) : null}
          {snapshot.transportPending ? (
            <p role="status" style={textStyle}>
              The original request is still settling. Another write is blocked.
            </p>
          ) : null}
          {attempt.kind !== "delete" ? (
            <label style={textStyle}>
              Name for reviewed request
              <input
                ref={registerItem("recovery_name")}
                aria-label="Name for reviewed saved-view request"
                aria-invalid={nameError ? true : undefined}
                aria-describedby={nameError ? nameErrorId : undefined}
                value={name}
                onChange={(event) =>
                  controller.setReviewName(
                    attempt.id,
                    event.currentTarget.value,
                  )
                }
                style={inputStyle}
              />
            </label>
          ) : null}
          {nameError ? (
            <p id={nameErrorId} style={textStyle}>
              {nameError}
            </p>
          ) : null}
          {canApply ? (
            <button
              ref={registerItem("apply_review")}
              type="button"
              style={buttonStyle}
              disabled={!controller.canApplyReview(attempt.id, observation)}
              onClick={() =>
                controller.review(
                  attempt.id,
                  observation,
                  "apply",
                  attempt.kind === "delete" || snapshot.reviewName === null
                    ? undefined
                    : (snapshot.reviewName ?? undefined),
                )
              }
            >
              {attempt.kind === "delete"
                ? "Delete the reviewed saved configuration"
                : "Apply submitted changes to reviewed version"}
            </button>
          ) : null}
          {attempt.kind !== "delete" ? (
            <>
              {uncertainCreate ? (
                <label style={textStyle}>
                  <input
                    ref={registerItem("confirm_create")}
                    type="checkbox"
                    checked={confirmedAttempt === attempt.id}
                    onChange={(event) =>
                      setConfirmedAttempt(
                        event.currentTarget.checked ? attempt.id : null,
                      )
                    }
                  />
                  I understand a new create attempt may create a duplicate.
                </label>
              ) : null}
              <button
                ref={registerItem("create_review")}
                type="button"
                style={buttonStyle}
                disabled={
                  !canReview ||
                  (uncertainCreate && confirmedAttempt !== attempt.id)
                }
                onClick={() =>
                  controller.review(attempt.id, observation, "create", name)
                }
              >
                {uncertainCreate
                  ? "Make a new create attempt"
                  : "Save submitted configuration as a new view"}
              </button>
            </>
          ) : null}
          <button
            ref={registerItem("keep_review")}
            type="button"
            style={buttonStyle}
            disabled={!canReview}
            onClick={() => controller.review(attempt.id, observation, "keep")}
          >
            End recovery without another write
          </button>
        </>
      ) : null}
      {operation.kind === "confirmed" || operation.kind === "reviewed" ? (
        <button
          ref={registerItem("dismiss")}
          type="button"
          style={buttonStyle}
          disabled={snapshot.transportPending}
          onClick={controller.dismiss}
        >
          Dismiss confirmation
        </button>
      ) : null}
    </section>
  );
}

function ConfigurationSummary({
  summaryRef,
  label,
  name,
  scope,
  schema,
  query,
  layout,
}: {
  summaryRef: RefCallback<HTMLElement>;
  label: string;
  name: string;
  scope: string;
  schema: string;
  query: unknown;
  layout: unknown;
}) {
  const contract = requireViewContract(schema);
  const entries = projectWorkbookQueryEntries(
    contract,
    workbookQueryStateFromSavedViewQueryJson(contract, query),
  );
  const columns = layout as {
    column_order: string[];
    hidden_field_keys: string[];
    column_widths: { field_key: string; width_px: number }[];
  };
  return (
    <details>
      <summary ref={summaryRef}>{label}</summary>
      <dl style={textStyle}>
        <dt>Name</dt>
        <dd>{name}</dd>
        <dt>Scope</dt>
        <dd>{scope}</dd>
        <dt>Query</dt>
        <dd>
          {entries.length
            ? entries.map((entry) => entry.accessibleName).join("; ")
            : "Default sort; no filters or grouping"}
        </dd>
        <dt>Columns in order</dt>
        <dd>
          {columns.column_order
            .map(
              (key) =>
                `${contract.fieldMap[key]?.label ?? key}${columns.hidden_field_keys.includes(key) ? " (hidden)" : ""}${columns.column_widths.find((w) => w.field_key === key) === undefined ? "" : ` (${columns.column_widths.find((w) => w.field_key === key)?.width_px} px)`}`,
            )
            .join(", ")}
        </dd>
      </dl>
    </details>
  );
}
const textStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
  margin: 0,
  overflowWrap: "anywhere" as const,
};
const sectionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  borderBlockStart: "var(--ct-border-hairline)",
  paddingBlockStart: "var(--ct-spacing-sm)",
};
const buttonStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-xs)",
  background: "transparent",
  color: "var(--ct-colors-ink)",
  cursor: "pointer",
  font: "inherit",
  padding: "var(--ct-component-button-secondary-padding)",
  textAlign: "left" as const,
};
const inputStyle = {
  display: "block",
  boxSizing: "border-box" as const,
  width: "100%",
  border: "var(--ct-component-text-input-border)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
  padding: "var(--ct-component-text-input-padding)",
  font: "inherit",
};

function problemLabel(
  kind: import("../ports/WorkbookSavedViewPort").SavedViewProblem["kind"],
) {
  return {
    validation: "Validation",
    conflict: "Version conflict",
    authentication_required: "Sign-in required",
    authorization_denied: "Permission denied",
    unavailable_target: "Saved view unavailable",
    transport: "Connection interrupted",
    invalid_contract: "Invalid server response",
    terminal: "Request not confirmed",
  }[kind];
}
