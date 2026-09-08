import {
  administrativeAuditDetailTestId,
  administrativeAuditEventTestId,
} from "@cartulary/ui-contracts";
import { ChevronRight, RefreshCw } from "lucide-react";
import { Fragment, useLayoutEffect, useRef, useSyncExternalStore } from "react";
import { formatAuditJSON } from "../shared/auditReadValues";
import type { AdministrativeAuditController } from "./administrativeAuditController";
import {
  type AuditEvent,
  type AuditField,
  auditActionCodes,
  auditFieldLabels,
  auditFilterFields,
  auditHistoryLimit,
  auditPageSummary,
  auditTargetKinds,
  sameAuditQuery,
  validateAuditInputs,
} from "./administrativeAuditModel";
import {
  auditActionsStyle,
  auditCellStyle,
  auditDetailsStyle,
  auditFieldErrorStyle,
  auditFormGridStyle,
  auditHeaderCellStyle,
  auditInputStyle,
  auditMetadataGridStyle,
  auditPanelStyle,
  auditResponsiveCss,
  auditTableStyle,
  auditTextStyle,
  auditValueStyle,
  definitionLabelStyle,
  emptyStateStyle,
  errorTextStyle,
  labelBlockStyle,
  metadataTextStyle,
  redactedBadgeStyle,
  secondaryButtonStyle,
  sectionEyebrowStyle,
  sectionTitleStyle,
  statusTextStyle,
  tableLinkButtonStyle,
  tableRowStyle,
  tableShellStyle,
} from "./landingAdminStyles";

export function AdministrativeAuditPanel({
  controller,
  active,
}: {
  controller: AdministrativeAuditController;
  active: boolean;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const root = useRef<HTMLElement>(null);
  const refresh = useRef<HTMLButtonElement>(null);
  const tablePort = useRef<HTMLElement>(null);
  const focused = useRef<HTMLElement | null>(null);
  const normalized = validateAuditInputs(state.inputs);
  const unapplied =
    Object.keys(normalized.errors).length > 0 ||
    !sameAuditQuery(normalized.query, state.applied);
  const fieldErrors = Object.keys(state.fieldErrors).length > 0;
  const page = state.page;
  const busy = state.activity !== null;
  const stale = page !== null && state.problem !== null;

  useLayoutEffect(() => {
    if (!active || !state.active || document.visibilityState === "hidden") {
      focused.current = null;
      return;
    }
    const origin = focused.current;
    if (origin && !origin.isConnected) {
      if (
        document.activeElement === document.body ||
        document.activeElement === null
      )
        refresh.current?.focus({ preventScroll: true });
      focused.current = null;
    }
  }, [active, state]);
  useLayoutEffect(() => {
    if (!active || !root.current) return;
    const scrollPort = () => {
      for (
        let element: HTMLElement | null = root.current;
        element;
        element = element.parentElement
      ) {
        if (
          /(auto|scroll)/u.test(getComputedStyle(element).overflowY) &&
          element.scrollHeight > element.clientHeight
        )
          return element;
      }
      return document.scrollingElement;
    };
    const saved = controller.getSnapshot();
    const ancestor = scrollPort();
    if (ancestor) ancestor.scrollTop = saved.scrollTop;
    if (tablePort.current) tablePort.current.scrollLeft = saved.tableScrollLeft;
    const remember = (event: Event) => {
      const port = scrollPort();
      if (
        event.target === port ||
        event.target === document ||
        event.target === tablePort.current
      )
        controller.rememberScroll(
          port?.scrollTop ?? 0,
          tablePort.current?.scrollLeft ?? 0,
        );
    };
    document.addEventListener("scroll", remember, true);
    return () => document.removeEventListener("scroll", remember, true);
  }, [active, controller]);

  function renderFilter(field: AuditField) {
    const id = `administrative-audit-filter-${field}`;
    const error = state.fieldErrors[field];
    const temporal = field === "occurred_at_gte" || field === "occurred_at_lt";
    const descriptions =
      [
        temporal ? "administrative-audit-time-help" : null,
        error ? `${id}-error` : null,
      ]
        .filter(Boolean)
        .join(" ") || undefined;
    const props = {
      id,
      "aria-invalid": error ? true : undefined,
      "aria-describedby": descriptions,
      style: auditInputStyle,
      value: state.inputs[field],
      onChange: (event: { target: { value: string } }) =>
        controller.edit(field, event.target.value),
    };
    return (
      <div key={field} style={auditTextStyle}>
        <label htmlFor={id} style={labelBlockStyle}>
          {auditFieldLabels[field]}
        </label>
        {field === "action_code" || field === "target_kind" ? (
          <select {...props}>
            <option value="">
              {field === "action_code" ? "Any action code" : "Any target kind"}
            </option>
            {(field === "action_code"
              ? auditActionCodes
              : auditTargetKinds
            ).map((value) => (
              <option key={value} value={value}>
                {value}
              </option>
            ))}
          </select>
        ) : (
          <input {...props} type="text" spellCheck={false} autoComplete="off" />
        )}
        {error ? (
          <p id={`${id}-error`} style={auditFieldErrorStyle}>
            {error}
          </p>
        ) : null}
      </div>
    );
  }
  const canPrevious =
    !busy && state.continuationValid && state.history.length > 0;
  const canNext =
    !busy && state.continuationValid && page?.paging.has_more === true;
  const showFirst = page !== null && page.descriptor.number > 1;

  return (
    <section
      ref={root}
      data-audit-browser=""
      aria-label="Deployment audit browser"
      style={auditPanelStyle}
      onFocusCapture={(event) => {
        focused.current = event.target as HTMLElement;
      }}
    >
      <style>{auditResponsiveCss}</style>
      <div style={auditActionsStyle}>
        <div style={auditTextStyle}>
          <p style={sectionEyebrowStyle}>Administrative audit</p>
          <h2 style={sectionTitleStyle}>Deployment events</h2>
        </div>
        <button
          ref={refresh}
          style={secondaryButtonStyle}
          type="button"
          aria-disabled={state.activity === "refresh" || !state.authority}
          onClick={controller.refresh}
        >
          <RefreshCw size={15} aria-hidden="true" />
          Refresh
        </button>
      </div>
      {state.authority ? (
        <>
          <form
            aria-label="Audit filters"
            noValidate
            onSubmit={(event) => {
              event.preventDefault();
              controller.apply();
              const first = auditFilterFields.find(
                (field) => controller.getSnapshot().fieldErrors[field],
              );
              if (first)
                root.current
                  ?.querySelector<HTMLElement>(
                    `#administrative-audit-filter-${first}`,
                  )
                  ?.focus();
            }}
            style={auditDetailsStyle}
          >
            <div style={auditFormGridStyle}>
              {auditFilterFields.map(renderFilter)}
            </div>
            <p id="administrative-audit-time-help" style={metadataTextStyle}>
              Use RFC 3339 with an explicit timezone, for example
              2026-05-24T12:00:00Z or 2026-05-24T08:00:00-04:00. The lower bound
              is inclusive; the upper bound is exclusive. Applied times are
              shown in UTC.
            </p>
            <div style={auditActionsStyle}>
              <button type="submit" style={secondaryButtonStyle}>
                Apply filters
              </button>
              <button
                type="button"
                style={secondaryButtonStyle}
                onClick={controller.clearFilters}
              >
                Clear filters
              </button>
              {unapplied ? (
                <span style={metadataTextStyle}>
                  Unapplied filter edits. Displayed results use the applied
                  filters below.
                </span>
              ) : null}
            </div>
          </form>
          <section aria-label="Applied audit filters" style={auditDetailsStyle}>
            <strong>Applied filters</strong>
            {auditFilterFields.some((field) => state.applied[field] !== "") ? (
              <dl style={auditMetadataGridStyle}>
                {auditFilterFields
                  .filter((field) => state.applied[field] !== "")
                  .map((field) => (
                    <div key={field} style={auditTextStyle}>
                      <dt style={definitionLabelStyle}>
                        {auditFieldLabels[field]}
                      </dt>
                      <dd style={auditValueStyle}>{state.applied[field]}</dd>
                    </div>
                  ))}
              </dl>
            ) : (
              <p style={metadataTextStyle}>
                All deployment events. No filters applied.
              </p>
            )}
          </section>
          <div style={auditDetailsStyle}>
            <p
              role={state.announcementRole}
              aria-atomic="true"
              style={
                state.problem || fieldErrors ? errorTextStyle : statusTextStyle
              }
            >
              <span key={state.announcementSequence}>{state.announcement}</span>
            </p>
            {state.problem ? (
              <div style={auditActionsStyle}>
                {stale ? (
                  <span style={metadataTextStyle}>
                    Displayed events may be stale; the latest read did not
                    replace them.
                  </span>
                ) : null}
                {state.problem.kind === "cursor" ? (
                  <button
                    type="button"
                    style={secondaryButtonStyle}
                    onClick={controller.refresh}
                  >
                    Reload first page
                  </button>
                ) : state.problem.kind === "query" ? (
                  <span style={metadataTextStyle}>
                    Review the applied filters and use Apply filters.
                  </span>
                ) : (
                  <button
                    type="button"
                    style={secondaryButtonStyle}
                    onClick={controller.retry}
                  >
                    Retry{" "}
                    {state.problem.kind === "page"
                      ? "page"
                      : state.problem.kind === "refresh"
                        ? "refresh"
                        : "read"}
                  </button>
                )}
              </div>
            ) : null}
          </div>
          {page ? (
            <>
              <div style={auditActionsStyle}>
                {state.announcement !== auditPageSummary(page) ? (
                  <p style={metadataTextStyle}>{auditPageSummary(page)}</p>
                ) : null}
                <span style={metadataTextStyle}>
                  Pages reflect live authorized results. Refresh starts at the
                  newest events.
                </span>
              </div>
              {page.rows.length === 0 ? (
                <p style={emptyStateStyle}>
                  {page.descriptor.number > 1
                    ? "No events were returned on this page. Use First page for a fresh read."
                    : auditFilterFields.some(
                          (field) => state.applied[field] !== "",
                        )
                      ? "No deployment events match the applied filters. Clear or change the filters to read again."
                      : "No deployment audit events are available."}
                </p>
              ) : (
                <section
                  ref={tablePort}
                  aria-label="Audit events table, horizontally scrollable"
                  // biome-ignore lint/a11y/noNoninteractiveTabindex: Keyboard users must be able to scroll the table region.
                  tabIndex={0}
                  style={tableShellStyle}
                >
                  <table
                    className="aa-events"
                    aria-label="Deployment audit events"
                    aria-busy={busy}
                    style={auditTableStyle}
                  >
                    <thead>
                      <tr>
                        {[
                          "Occurred (UTC)",
                          "Action",
                          "Actor",
                          "Target",
                          "Reason",
                          "Changes",
                          "Details",
                        ].map((label) => (
                          <th
                            scope="col"
                            key={label}
                            style={auditHeaderCellStyle}
                          >
                            {label}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {page.rows.map((event) => {
                        const expanded =
                          state.expandedId === event.audit_event_id;
                        const detailsId = administrativeAuditDetailTestId(
                          event.audit_event_id,
                        );
                        const label = auditActionLabel(event.action_code);
                        return (
                          <Fragment key={event.audit_event_id}>
                            <tr
                              data-testid={administrativeAuditEventTestId(
                                event.audit_event_id,
                              )}
                              style={tableRowStyle}
                            >
                              <td style={auditCellStyle}>
                                {auditCellLabel("Occurred (UTC)")}
                                <time dateTime={event.occurred_at}>
                                  {event.occurred_at}
                                </time>
                                <p style={metadataTextStyle}>{event.source}</p>
                              </td>
                              <td style={auditCellStyle}>
                                {auditCellLabel("Action")}
                                {label}
                              </td>
                              <td style={auditCellStyle}>
                                {auditCellLabel("Actor")}
                                {event.actor_user_id ?? event.actor_kind}
                              </td>
                              <td style={auditCellStyle}>
                                {auditCellLabel("Target")}
                                {auditTarget(event)}
                              </td>
                              <td style={auditCellStyle}>
                                {auditCellLabel("Reason")}
                                {event.reason_code ?? "No reason"}
                              </td>
                              <td style={auditCellStyle}>
                                {auditCellLabel("Changes")}
                                {event.changes.length}
                              </td>
                              <td style={auditCellStyle}>
                                {auditCellLabel("Details")}
                                <button
                                  aria-expanded={expanded}
                                  aria-controls={detailsId}
                                  aria-label={`${expanded ? "Hide" : "Inspect"} ${label} event at ${event.occurred_at}`}
                                  style={tableLinkButtonStyle}
                                  type="button"
                                  onClick={() =>
                                    controller.toggleExpanded(
                                      event.audit_event_id,
                                    )
                                  }
                                >
                                  {expanded ? "Hide" : "Inspect"}
                                  <ChevronRight size={15} aria-hidden="true" />
                                </button>
                              </td>
                            </tr>
                            {expanded ? (
                              <tr className="aa-detail-row">
                                <td colSpan={7} style={auditCellStyle}>
                                  <section
                                    id={detailsId}
                                    data-testid={detailsId}
                                    aria-label={`Details for ${label} at ${event.occurred_at}`}
                                    style={auditDetailsStyle}
                                  >
                                    <dl style={auditMetadataGridStyle}>
                                      {[
                                        [
                                          "Audit event ID",
                                          event.audit_event_id,
                                        ],
                                        ["Raw action code", event.action_code],
                                        ["Target kind", event.target_kind],
                                        [
                                          "Target ID",
                                          event.target_id ?? "No target ID",
                                        ],
                                      ].map(([name, value]) => (
                                        <div key={name} style={auditTextStyle}>
                                          <dt style={definitionLabelStyle}>
                                            {name}
                                          </dt>
                                          <dd style={auditValueStyle}>
                                            {value}
                                          </dd>
                                        </div>
                                      ))}
                                    </dl>
                                    <table
                                      className="aa-changes"
                                      aria-label="Published field changes"
                                      style={auditTableStyle}
                                    >
                                      <thead>
                                        <tr>
                                          {["Field", "Before", "After"].map(
                                            (name) => (
                                              <th
                                                key={name}
                                                scope="col"
                                                style={auditHeaderCellStyle}
                                              >
                                                {name}
                                              </th>
                                            ),
                                          )}
                                        </tr>
                                      </thead>
                                      <tbody>
                                        {event.changes.map((change) => (
                                          <tr key={change.field_path}>
                                            <th
                                              scope="row"
                                              style={auditCellStyle}
                                            >
                                              {auditCellLabel("Field")}
                                              <span style={auditValueStyle}>
                                                {change.field_path}
                                              </span>
                                            </th>
                                            <td style={auditCellStyle}>
                                              {auditCellLabel("Before")}
                                              {renderAuditValue(
                                                change.value_state,
                                                change.before,
                                              )}
                                            </td>
                                            <td style={auditCellStyle}>
                                              {auditCellLabel("After")}
                                              {renderAuditValue(
                                                change.value_state,
                                                change.after,
                                              )}
                                            </td>
                                          </tr>
                                        ))}
                                      </tbody>
                                    </table>
                                  </section>
                                </td>
                              </tr>
                            ) : null}
                          </Fragment>
                        );
                      })}
                    </tbody>
                  </table>
                </section>
              )}
              <nav aria-label="Audit pagination" style={auditActionsStyle}>
                <button
                  type="button"
                  style={secondaryButtonStyle}
                  aria-disabled={!canPrevious}
                  onClick={() => {
                    if (canPrevious) controller.previous();
                  }}
                >
                  Previous page
                </button>
                <button
                  type="button"
                  style={secondaryButtonStyle}
                  aria-disabled={!canNext}
                  onClick={() => {
                    if (canNext) controller.next();
                  }}
                >
                  Next page
                </button>
                {showFirst ? (
                  <button
                    type="button"
                    style={secondaryButtonStyle}
                    onClick={controller.refresh}
                  >
                    First page
                  </button>
                ) : null}
                {page.descriptor.number > 1 &&
                state.history.length === 0 &&
                state.continuationValid ? (
                  <span style={metadataTextStyle}>
                    Previous navigation retains at most {auditHistoryLimit}{" "}
                    pages. Use First page to restart.
                  </span>
                ) : null}
              </nav>
            </>
          ) : null}
        </>
      ) : (
        <p style={emptyStateStyle}>
          Deployment administrator access is required to read the audit.
        </p>
      )}
    </section>
  );
}

function renderAuditValue(state: "visible" | "redacted", value: unknown) {
  if (state === "redacted")
    return <span style={redactedBadgeStyle}>Redacted</span>;
  return <span style={auditValueStyle}>{formatAuditJSON(value)}</span>;
}
function auditActionLabel(code: string) {
  return auditActionLabels.get(code) ?? code;
}
function auditTarget(event: AuditEvent) {
  const label = auditTargetLabels.get(event.target_kind) ?? event.target_kind;
  return event.target_id === null ? label : `${label} ${event.target_id}`;
}
const auditTargetLabels = new Map<string, string>([
  ["user", "User"],
  ["account_preferences", "Account preferences for"],
  ["auth_binding", "Authentication binding"],
  ["backup_set", "Backup set"],
  ["restore_operation", "Restore operation"],
]);

const auditActionLabels = new Map<string, string>(
  Object.entries({
    account_preferences_updated: "Account preferences updated",
    auth_binding_created: "Authentication binding created",
    auth_binding_retired: "Authentication binding retired",
    auth_binding_rotated: "Authentication binding rotated",
    backup_created: "Backup created",
    bootstrap_admin_created: "Bootstrap admin created",
    deployment_admin_granted: "Deployment admin granted",
    deployment_admin_revoked: "Deployment admin revoked",
    password_changed: "Password changed",
    password_reset: "Password reset",
    restore_completed: "Restore completed",
    restore_failed: "Restore failed",
    restore_started: "Restore started",
    restore_verification_completed: "Restore verification completed",
    sessions_revoked: "Sessions revoked",
    totp_enrollment_begun: "TOTP enrollment begun",
    totp_enrollment_completed: "TOTP enrollment completed",
    totp_reset: "TOTP reset",
    user_created: "User created",
    user_profile_updated: "User profile updated",
    user_status_changed: "User status changed",
  }),
);

function auditCellLabel(label: string) {
  return (
    <span className="aa-cell-label" aria-hidden="true">
      {label}
    </span>
  );
}
