import {
  incidentAdministrationTestId,
  incidentControlsSurfaceTestId,
  incidentMembershipAuditDetailTestId,
  incidentMembershipAuditRowTestId,
} from "@cartulary/ui-contracts";
import { useLayoutEffect, useRef, useSyncExternalStore } from "react";
import { formatAuditJSON } from "../shared/auditReadValues";
import type {
  WorkbookDensityMode,
  WorkbookIncidentControlsRendererProps,
} from "../shared/workbookShellContracts";
import type { IncidentMembershipAuditController } from "./incidentMembershipAuditController";
import {
  type AuditField,
  auditFilterFields,
  membershipAuditActions,
  membershipAuditHistoryLimit,
  membershipAuditTargets,
  sameMembershipAuditQuery,
  validateMembershipAuditInputs,
} from "./incidentMembershipAuditModel";
import {
  auditActionsStyle,
  auditDetailsStyle,
  auditFieldErrorStyle,
  auditFormGridStyle,
  auditInputStyle,
  auditMetadataGridStyle,
  auditTextStyle,
  auditValueStyle,
  labelBlockStyle,
  metadataTextStyle,
  redactedBadgeStyle,
  secondaryButtonStyle,
} from "./landingAdminStyles";

const labels: Record<AuditField, string> = {
  actor_user_id: "Actor user ID",
  action_code: "Action code",
  target_kind: "Target kind",
  target_id: "Target ID",
  occurred_at_gte: "Occurred at or after",
  occurred_at_lt: "Occurred before",
};
const actionLabels = new Map([
  ["membership_created", "Membership created"],
  ["membership_deleted", "Membership deleted"],
  ["membership_role_changed", "Membership role changed"],
]);
const auditCss = `
[data-membership-audit] :is(input,select,button):focus-visible { outline: var(--ct-border-focus); outline-offset: var(--ct-spacing-xs); }
[data-membership-audit] button[aria-disabled="true"] { color: var(--ct-colors-ink-subtle) !important; background: var(--ct-colors-surface-3) !important; cursor: not-allowed !important; }
[data-membership-audit] input, [data-membership-audit] select, [data-membership-audit] button { font: inherit; }
[data-membership-audit] .ma-event { padding: var(--ma-cell-padding); }
`;

/** Mount/visibility binding is separate from presentation and never constructs reads. */
export function IncidentMembershipAuditFeature({
  controller,
  bindSurface,
  ...surface
}: WorkbookIncidentControlsRendererProps & {
  controller: IncidentMembershipAuditController;
  bindSurface: (surface: WorkbookIncidentControlsRendererProps | null) => void;
}) {
  const latest = useRef({ bindSurface, surface });
  latest.current = { bindSurface, surface };
  useLayoutEffect(() => {
    bindSurface(surface);
  });
  useLayoutEffect(() => {
    const visibility = () => latest.current.bindSurface(latest.current.surface);
    document.addEventListener("visibilitychange", visibility);
    return () => {
      document.removeEventListener("visibilitychange", visibility);
      latest.current.bindSurface(null);
    };
  }, []);
  return (
    <IncidentMembershipAuditPanel
      controller={controller}
      density={surface.density}
    />
  );
}

export function IncidentMembershipAuditPanel({
  controller,
  density = "default",
}: {
  controller: IncidentMembershipAuditController;
  density?: WorkbookDensityMode | undefined;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const root = useRef<HTMLElement>(null);
  const refresh = useRef<HTMLButtonElement>(null);
  const focused = useRef<HTMLElement | null>(null);
  const { page, read } = state;
  const busy = read.kind === "pending";
  const allowed = state.authority !== null && read.kind !== "denied";
  const normalized = validateMembershipAuditInputs(state.inputs);
  const unapplied =
    Object.keys(normalized.errors).length > 0 ||
    !sameMembershipAuditQuery(normalized.query, state.applied);
  const previous =
    page !== null &&
    (busy ||
      read.kind === "failed" ||
      read.kind === "cursor_rejected" ||
      !sameMembershipAuditQuery(page.query, state.applied));
  useLayoutEffect(() => {
    if (!state.active || document.visibilityState === "hidden") {
      focused.current = null;
      return;
    }
    if (
      focused.current &&
      !focused.current.isConnected &&
      (document.activeElement === document.body ||
        document.activeElement === null)
    )
      refresh.current?.focus({ preventScroll: true });
  }, [state]);
  const filterContext = (query: typeof state.applied) =>
    auditFilterFields
      .filter((field) => query[field] !== "")
      .map((field) => `${labels[field]}: ${query[field]}`)
      .join("; ") || "All membership events";
  const navigation = (
    <nav aria-label="Membership audit pages" style={auditActionsStyle}>
      <button
        type="button"
        style={secondaryButtonStyle}
        aria-disabled={!controller.canContinue() || state.history.length === 0}
        onClick={controller.previous}
      >
        Previous page
      </button>
      <button
        type="button"
        style={secondaryButtonStyle}
        aria-disabled={!controller.canContinue() || !page?.paging.has_more}
        onClick={controller.next}
      >
        Next page
      </button>
      <button
        type="button"
        style={secondaryButtonStyle}
        aria-disabled={!allowed || busy}
        onClick={() => {
          if (allowed && !busy) controller.refresh();
        }}
      >
        {read.kind === "cursor_rejected" ? "Reload first page" : "First page"}
      </button>
    </nav>
  );
  return (
    <section
      ref={root}
      aria-label="Incident membership audit browser"
      data-membership-audit=""
      data-testid={incidentControlsSurfaceTestId()}
      data-incident-controls-section="membership-audit"
      data-incident-controls-load-state={
        busy ? "loading" : page ? "synced" : "partial"
      }
      style={{
        ...auditDetailsStyle,
        fontSize: `var(--ct-density-${density}-fontSize)`,
        lineHeight: `var(--ct-density-${density}-lineHeight)`,
        ...{ "--ma-cell-padding": `var(--ct-density-${density}-cellPadding)` },
      }}
      onFocusCapture={(event) => {
        focused.current = event.target as HTMLElement;
      }}
    >
      <style>{auditCss}</style>
      <header style={auditActionsStyle}>
        <h3 style={{ margin: 0 }}>Incident membership audit</h3>
        <button
          ref={refresh}
          type="button"
          style={secondaryButtonStyle}
          aria-disabled={
            !allowed || (busy && read.intent.position.cursor === null)
          }
          onClick={controller.refresh}
        >
          Refresh
        </button>
      </header>
      {allowed ? (
        <>
          <form
            aria-label="Membership audit filters"
            noValidate
            style={auditDetailsStyle}
            onSubmit={(event) => {
              event.preventDefault();
              controller.apply();
              const invalid = auditFilterFields.find(
                (field) => controller.getSnapshot().errors[field],
              );
              if (invalid)
                root.current
                  ?.querySelector<HTMLElement>(`#membership-audit-${invalid}`)
                  ?.focus();
            }}
          >
            <div style={auditFormGridStyle}>
              {auditFilterFields.map((field) => {
                const id = `membership-audit-${field}`;
                const error = state.errors[field];
                const temporal = field.startsWith("occurred_at_");
                const props = {
                  id,
                  value: state.inputs[field],
                  style: auditInputStyle,
                  "aria-invalid": error ? true : undefined,
                  "aria-describedby":
                    [
                      temporal ? "membership-audit-time-help" : "",
                      error ? `${id}-error` : "",
                    ]
                      .filter(Boolean)
                      .join(" ") || undefined,
                  onChange: (event: { target: { value: string } }) =>
                    controller.edit(field, event.target.value),
                };
                return (
                  <div key={field} style={auditTextStyle}>
                    <label htmlFor={id} style={labelBlockStyle}>
                      {labels[field]}
                    </label>
                    {field === "action_code" || field === "target_kind" ? (
                      <select {...props}>
                        <option value="">
                          {field === "action_code"
                            ? "Any action"
                            : "Any target kind"}
                        </option>
                        {(field === "action_code"
                          ? membershipAuditActions
                          : membershipAuditTargets
                        ).map((value) => (
                          <option key={value} value={value}>
                            {value}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <input
                        {...props}
                        type="text"
                        autoComplete="off"
                        spellCheck={false}
                      />
                    )}
                    {error ? (
                      <p id={`${id}-error`} style={auditFieldErrorStyle}>
                        {error}
                      </p>
                    ) : null}
                  </div>
                );
              })}
            </div>
            <p id="membership-audit-time-help" style={metadataTextStyle}>
              Use RFC 3339 with Z or a numeric timezone offset, such as
              2026-05-24T12:00:00Z. The lower bound is inclusive; the upper
              bound is exclusive. Applied times are shown in UTC.
            </p>
            <div style={auditActionsStyle}>
              <button
                type="submit"
                style={secondaryButtonStyle}
                data-testid={incidentAdministrationTestId(
                  "membership-audit-apply-filters",
                )}
              >
                Apply filters
              </button>
              <button
                type="button"
                style={secondaryButtonStyle}
                onClick={controller.clearFilters}
              >
                Clear filters
              </button>
              {unapplied ? <span>Unapplied filter edits</span> : null}
            </div>
          </form>
          <section
            aria-label="Applied membership audit filters"
            style={auditTextStyle}
          >
            Applied filters: {filterContext(state.applied)}
          </section>
        </>
      ) : (
        <p data-testid={incidentAdministrationTestId("membership-audit-note")}>
          Only incident admins can review incident membership audit.
        </p>
      )}
      <p
        role={state.announcementRole}
        aria-live={state.announcementRole === "alert" ? "assertive" : "polite"}
        aria-atomic="true"
        data-testid={incidentAdministrationTestId("membership-audit-status")}
        style={auditTextStyle}
      >
        {state.announcement}
      </p>
      {read.kind === "failed" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={controller.retry}
        >
          Try the read again
        </button>
      ) : null}
      {read.kind === "denied" && state.authority ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={controller.retry}
        >
          Check audit access
        </button>
      ) : null}
      {previous && page ? (
        <p style={auditTextStyle}>
          Previous results
          {read.kind === "failed" || read.kind === "cursor_rejected"
            ? "; displayed events may be stale"
            : ""}
          . Displayed filters: {filterContext(page.query)}
        </p>
      ) : null}
      {allowed ? navigation : null}
      {page && allowed ? (
        <div aria-busy={busy}>
          {page.rows.length === 0 ? (
            <p
              data-testid={incidentAdministrationTestId(
                "membership-audit-empty",
              )}
            >
              {page.position.number > 1
                ? "No further events on this continuation."
                : auditFilterFields.some((field) => page.query[field] !== "")
                  ? "No membership audit events match the applied filters."
                  : "No membership audit events yet."}
            </p>
          ) : (
            <ul
              aria-label="Incident membership audit events"
              data-testid={incidentAdministrationTestId(
                "membership-audit-list",
              )}
              style={{ margin: 0, padding: 0, listStyle: "none" }}
            >
              {page.rows.map((event) => {
                const expanded = state.expandedId === event.audit_event_id;
                const detailId = incidentMembershipAuditDetailTestId(
                  event.audit_event_id,
                );
                const label =
                  actionLabels.get(event.action_code) ?? event.action_code;
                return (
                  <li
                    key={event.audit_event_id}
                    data-testid={incidentMembershipAuditRowTestId(
                      event.audit_event_id,
                    )}
                    className="ma-event"
                    style={{
                      ...auditDetailsStyle,
                      borderBlockEnd: "var(--ct-border-hairline)",
                    }}
                  >
                    <div style={auditActionsStyle}>
                      <strong style={auditTextStyle}>{label}</strong>
                      <time dateTime={event.occurred_at} style={auditTextStyle}>
                        {event.occurred_at}
                      </time>
                      <button
                        type="button"
                        style={secondaryButtonStyle}
                        aria-expanded={expanded}
                        aria-controls={detailId}
                        aria-label={`${expanded ? "Hide" : "Inspect"} ${label} event at ${event.occurred_at}`}
                        onClick={() =>
                          controller.toggleExpanded(event.audit_event_id)
                        }
                      >
                        {expanded ? "Hide" : "Inspect"}
                      </button>
                    </div>
                    <div style={auditMetadataGridStyle}>
                      <span style={auditTextStyle}>
                        Actor: {event.actor_user_id ?? event.actor_kind}
                      </span>
                      <span style={auditTextStyle}>
                        Target: {event.target_id ?? "No target ID"}
                      </span>
                    </div>
                    {expanded ? (
                      <section
                        id={detailId}
                        data-testid={detailId}
                        aria-label={`Details for ${label} at ${event.occurred_at}`}
                        style={auditDetailsStyle}
                      >
                        <dl style={auditMetadataGridStyle}>
                          {[
                            ["Audit event ID", event.audit_event_id],
                            ["Incident ID", event.scope_id],
                            ["Scope", event.scope_kind],
                            ["Raw action code", event.action_code],
                            ["Actor kind", event.actor_kind],
                            [
                              "Actor user ID",
                              event.actor_user_id ?? "No user actor",
                            ],
                            ["Source", event.source],
                            ["Target kind", event.target_kind],
                            ["Target ID", event.target_id ?? "No target ID"],
                            ["Reason", event.reason_code ?? "No reason"],
                          ].map(([name, value]) => (
                            <div key={name} style={auditTextStyle}>
                              <dt>{name}</dt>
                              <dd style={auditValueStyle}>{value}</dd>
                            </div>
                          ))}
                        </dl>
                        <section
                          aria-label="Published field changes"
                          style={auditDetailsStyle}
                        >
                          {event.changes.map((change) => (
                            <section
                              key={change.field_path}
                              aria-label={`Change to ${change.field_path}`}
                              style={auditDetailsStyle}
                            >
                              <h4
                                style={{
                                  ...auditValueStyle,
                                  fontSize: "inherit",
                                }}
                              >
                                {change.field_path}
                              </h4>
                              <dl style={auditMetadataGridStyle}>
                                {(["before", "after"] as const).map((side) => (
                                  <div key={side} style={auditTextStyle}>
                                    <dt>
                                      {side === "before" ? "Before" : "After"}
                                    </dt>
                                    <dd style={auditValueStyle}>
                                      {change.value_state === "redacted" ? (
                                        <span style={redactedBadgeStyle}>
                                          Redacted
                                        </span>
                                      ) : (
                                        formatAuditJSON(change[side])
                                      )}
                                    </dd>
                                  </div>
                                ))}
                              </dl>
                            </section>
                          ))}
                        </section>
                      </section>
                    ) : null}
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      ) : null}
      {allowed ? (
        <p style={metadataTextStyle}>
          One page of up to 100 events is displayed. Previous page retains at
          most {membershipAuditHistoryLimit} prior page links; First page begins
          a fresh traversal.{" "}
          {page && page.position.number > 1 && state.history.length === 0
            ? "Earlier page links are outside the retained navigation window."
            : ""}
        </p>
      ) : null}
    </section>
  );
}
