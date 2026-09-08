import {
  incidentControlsSurfaceTestId,
  incidentMembershipAdminNoteTestId,
  incidentMembershipCreateButtonTestId,
  incidentMembershipDeleteButtonTestId,
  incidentMembershipEmailInputTestId,
  incidentMembershipListTestId,
  incidentMembershipPatchButtonTestId,
  incidentMembershipRoleDisplayTestId,
  incidentMembershipRoleInputTestId,
  incidentMembershipRoleSelectTestId,
  incidentMembershipRowTestId,
  incidentMembershipVersionTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type RefObject,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import type {
  WorkbookDensityMode,
  WorkbookIncidentControlsRendererProps,
} from "../shared/workbookShellContracts";
import { AccountDialog } from "./AccountDialog";
import type { IncidentMembershipManagementController } from "./incidentMembershipManagementController";
import {
  type MembershipOperation,
  membershipConfirmedText,
  membershipProblemText,
  membershipRoles,
} from "./incidentMembershipManagementModel";
import { primaryButtonStyle, secondaryButtonStyle } from "./landingAdminStyles";

export function IncidentMembershipManagementFeature({
  controller,
  bindSurface,
  ...surface
}: WorkbookIncidentControlsRendererProps & {
  readonly controller: IncidentMembershipManagementController;
  readonly bindSurface: (
    surface: WorkbookIncidentControlsRendererProps | null,
  ) => void;
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
    <IncidentMembershipManagementPanel
      controller={controller}
      density={surface.density}
    />
  );
}
const css = `
[data-membership-management] :is(input,select,button):focus-visible { outline: var(--ct-border-focus); outline-offset: var(--ct-spacing-xs); }
[data-membership-management] :is(input,select,button) { font: inherit; max-inline-size: 100%; }
[data-membership-management] button[aria-disabled="true"] { cursor: not-allowed; color: var(--ct-colors-ink-muted) !important; background: var(--ct-colors-surface-2) !important; border: var(--ct-border-hairline) !important; }
[data-membership-management] :is(input,select) { box-sizing: border-box; min-inline-size: 0; inline-size: 100%; }
`;
const stack: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-md)",
  minWidth: 0,
};
const actions: CSSProperties = {
  display: "flex",
  gap: "var(--ct-spacing-sm)",
  flexWrap: "wrap",
  alignItems: "center",
  minWidth: 0,
};
const card: CSSProperties = {
  ...stack,
  padding: "var(--ct-spacing-md)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
  overflowWrap: "anywhere",
};
const text: CSSProperties = { margin: 0, overflowWrap: "anywhere" };
const muted: CSSProperties = { ...text, color: "var(--ct-colors-ink-muted)" };
const inputStyle: CSSProperties = {
  padding: "var(--ct-component-text-input-padding)",
  border: "var(--ct-component-text-input-border)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
};
const danger: CSSProperties = {
  ...secondaryButtonStyle,
  background: "var(--ct-component-button-danger-backgroundColor)",
  color: "var(--ct-component-button-danger-textColor)",
  border: "var(--ct-border-hairline)",
};

export function IncidentMembershipManagementPanel({
  controller,
  density = "default",
}: {
  readonly controller: IncidentMembershipManagementController;
  readonly density?: WorkbookDensityMode | undefined;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const root = useRef<HTMLElement>(null);
  const heading = useRef<HTMLHeadingElement>(null);
  const focused = useRef<HTMLElement | null>(null);
  const previousEditor = useRef(state.editor?.id);
  const { editor, page, read, operation } = state;
  const busy = read.kind === "pending";
  const allowed = state.authority !== null && state.access === "ready";
  const admin = allowed && state.authority?.role === "admin";
  useLayoutEffect(() => {
    if (!state.active || document.visibilityState === "hidden") {
      focused.current = null;
      return;
    }
    if (editor?.id !== previousEditor.current) {
      if (editor)
        (
          root.current?.querySelector<HTMLElement>(
            "[data-membership-editor] input, [data-membership-editor] select",
          ) ??
          root.current?.querySelector<HTMLElement>(
            "[data-membership-review-title]",
          )
        )?.focus({ preventScroll: true });
      previousEditor.current = editor?.id;
    }
    if (
      focused.current &&
      !focused.current.isConnected &&
      (document.activeElement === document.body ||
        document.activeElement === null)
    )
      heading.current?.focus({ preventScroll: true });
  }, [state, editor]);
  const readText =
    state.access === "unavailable"
      ? "Current membership access could not be checked. Your draft is retained; check access to continue."
      : read.kind === "cursor_rejected"
        ? "This membership continuation is no longer usable. Reload the first page."
        : read.kind === "failed"
          ? read.reason === "access"
            ? "Current membership access could not be checked. Your draft is retained; check access to continue."
            : page
              ? "Membership refresh failed. Previously loaded members may be stale."
              : "Memberships are unavailable. Try the read again."
          : busy
            ? page
              ? "Refreshing memberships. Displayed members are previous results."
              : "Loading memberships…"
            : page
              ? page.rows.length === 0
                ? "No memberships on this page."
                : `Page ${page.position.number}: ${page.rows.length} ${page.rows.length === 1 ? "member" : "members"}. ${page.paging.has_more ? "More members are available." : "End of the current results."}`
              : "Checking membership access…";
  const editBlocked =
    !admin ||
    operation.kind === "pending" ||
    operation.kind === "uncertain" ||
    operation.kind === "conflicted" ||
    state.transportPending;
  const label = editor?.base
    ? `${editor.base.display_name} (${editor.base.user_id})`
    : "existing account";
  const nav = (
    <nav aria-label="Membership pages" style={actions}>
      <button
        type="button"
        style={secondaryButtonStyle}
        aria-disabled={!allowed || busy}
        onClick={() => {
          if (allowed && !busy) controller.refresh();
        }}
      >
        First page
      </button>
      <button
        type="button"
        style={secondaryButtonStyle}
        aria-disabled={!controller.canPage() || state.history.length === 0}
        onClick={controller.previous}
      >
        Previous page
      </button>
      <button
        type="button"
        style={secondaryButtonStyle}
        aria-disabled={!controller.canPage() || !page?.paging.has_more}
        onClick={controller.next}
      >
        Next page
      </button>
    </nav>
  );
  return (
    <section
      ref={root}
      aria-label="Incident memberships"
      data-membership-management=""
      data-testid={incidentControlsSurfaceTestId()}
      data-incident-controls-section="memberships"
      data-membership-operation={operation.kind}
      style={{
        ...stack,
        fontSize: `var(--ct-density-${density}-fontSize)`,
        lineHeight: `var(--ct-density-${density}-lineHeight)`,
        ...{ "--mm-cell-padding": `var(--ct-density-${density}-cellPadding)` },
      }}
      onFocusCapture={(event) => {
        if (event.target instanceof HTMLElement) focused.current = event.target;
      }}
    >
      <style>{css}</style>
      <header style={actions}>
        <h3 ref={heading} tabIndex={-1} style={text}>
          Incident memberships
        </h3>
        <button
          type="button"
          style={secondaryButtonStyle}
          aria-disabled={!state.authority || busy}
          onClick={() => {
            if (state.authority && !busy) controller.refresh();
          }}
        >
          Refresh
        </button>
      </header>
      <p style={muted}>
        Browse incident access in join order. Add an existing account by email,
        change its role, or remove its access to this incident.
      </p>
      <p
        role={
          read.kind === "failed" || read.kind === "cursor_rejected"
            ? "alert"
            : "status"
        }
        style={text}
      >
        {readText}
      </p>
      {read.kind === "failed" ? (
        <div style={actions}>
          <button
            type="button"
            style={secondaryButtonStyle}
            onClick={controller.retryRead}
          >
            {read.reason === "access"
              ? "Check access and reload"
              : "Retry membership read"}
          </button>
        </div>
      ) : null}
      {read.kind === "cursor_rejected" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={controller.refresh}
        >
          Reload first page
        </button>
      ) : null}
      {state.access === "unavailable" && read.kind !== "failed" ? (
        <div style={card}>
          <p role="alert" style={text}>
            Current access is unavailable. Membership editing is suspended; your
            draft is retained.
          </p>
          <button
            type="button"
            style={secondaryButtonStyle}
            onClick={() => {
              void controller.recoverAccess();
            }}
          >
            Check access
          </button>
        </div>
      ) : null}
      {state.transportPending &&
      operation.kind !== "pending" &&
      operation.kind !== "uncertain" ? (
        <p role="status" style={text}>
          A previously admitted membership request is still settling. New
          membership writes are unavailable until its transport settles.
        </p>
      ) : null}
      <MembershipOperationFeedback
        protectedVisible={allowed}
        controller={controller}
        operation={operation}
        canRecover={admin && !state.transportPending}
      />
      {allowed ? (
        <>
          {!admin ? (
            <p data-testid={incidentMembershipAdminNoteTestId()} style={muted}>
              Only incident admins can add, change, or remove memberships.
            </p>
          ) : (
            <div style={actions}>
              <button
                type="button"
                style={primaryButtonStyle}
                aria-disabled={editBlocked}
                onClick={() => {
                  if (!editBlocked) controller.openAdd();
                }}
              >
                Add existing account
              </button>
            </div>
          )}
          {admin && editor ? (
            <form
              noValidate
              data-membership-editor=""
              aria-label={
                editor.kind === "add"
                  ? "Add existing incident member"
                  : editor.kind === "role"
                    ? `Edit role for ${label}`
                    : `Review removal of ${label}`
              }
              style={card}
              onSubmit={(event) => {
                event.preventDefault();
                controller.submit();
              }}
            >
              <h4 style={text} tabIndex={-1} data-membership-review-title="">
                {editor.kind === "add"
                  ? "Add existing account"
                  : editor.kind === "role"
                    ? `Change role for ${label}`
                    : `Remove incident access for ${label}`}
              </h4>
              {editor.kind === "add" ? (
                <label style={stack}>
                  User email
                  <input
                    type="text"
                    inputMode="email"
                    autoComplete="off"
                    spellCheck={false}
                    data-testid={incidentMembershipEmailInputTestId()}
                    style={inputStyle}
                    value={editor.email}
                    aria-invalid={state.fieldError ? true : undefined}
                    aria-describedby={
                      state.fieldError
                        ? "membership-field-error"
                        : "membership-add-help"
                    }
                    onChange={(event) =>
                      controller.editEmail(event.target.value)
                    }
                  />
                </label>
              ) : (
                <p style={muted}>
                  Saved role: {editor.base?.role}. Reviewed version:{" "}
                  {editor.base?.membership_version}.
                </p>
              )}
              {editor.kind === "add" ? (
                <p id="membership-add-help" style={muted}>
                  The account must already exist and be active. This does not
                  create an account or send an invitation.
                </p>
              ) : null}
              {editor.kind !== "remove" ? (
                <label style={stack}>
                  {editor.kind === "add"
                    ? "Role for added account"
                    : `Role for ${label}`}
                  <select
                    data-testid={
                      editor.kind === "add"
                        ? incidentMembershipRoleSelectTestId()
                        : incidentMembershipRoleInputTestId(
                            editor.base?.user_id ?? "",
                          )
                    }
                    style={inputStyle}
                    value={editor.role}
                    onChange={(event) => {
                      const role = membershipRoles.find(
                        (role) => role === event.target.value,
                      );
                      if (role) controller.editRole(role);
                    }}
                  >
                    {membershipRoles.map((role) => (
                      <option key={role} value={role}>
                        {role}
                      </option>
                    ))}
                  </select>
                </label>
              ) : (
                <p style={text}>
                  This removes only this incident membership. The account and
                  its access to other incidents are not deleted.
                </p>
              )}
              {editor.base?.user_id === state.authority?.actorId &&
              (editor.kind === "remove" || editor.role !== "admin") ? (
                <p style={text}>
                  {editor.kind === "remove"
                    ? "You will lose access to this incident and return to Incidents. Your account session remains valid."
                    : "You will lose membership administration for this incident. Your new role still determines ordinary incident access."}{" "}
                  Another current incident administrator must remain.
                </p>
              ) : null}
              {state.fieldError ? (
                <p
                  id="membership-field-error"
                  role="alert"
                  style={{
                    ...text,
                    color: "var(--ct-colors-semantic-conflict)",
                  }}
                >
                  {state.fieldError}
                </p>
              ) : null}
              {editor.reviewRequired ? (
                <div style={stack}>
                  <p role="alert" style={text}>
                    The saved membership changed. Your draft is retained and
                    needs review against current membership.
                  </p>
                  <button
                    type="button"
                    style={secondaryButtonStyle}
                    onClick={controller.refresh}
                  >
                    Observe current membership
                  </button>
                  <button
                    type="button"
                    style={secondaryButtonStyle}
                    aria-disabled={
                      !page?.rows.some(
                        (r) => r.user_id === editor.base?.user_id,
                      )
                    }
                    onClick={controller.reviewObserved}
                  >
                    Review observed membership
                  </button>
                </div>
              ) : null}
              <div style={actions}>
                <button
                  type="submit"
                  style={editor.kind === "remove" ? danger : primaryButtonStyle}
                  aria-disabled={!controller.canSubmit()}
                  data-testid={
                    editor.kind === "add"
                      ? incidentMembershipCreateButtonTestId()
                      : editor.kind === "role"
                        ? incidentMembershipPatchButtonTestId(
                            editor.base?.user_id ?? "",
                          )
                        : undefined
                  }
                >
                  {editor.kind === "add"
                    ? "Add membership"
                    : editor.kind === "role"
                      ? "Save role"
                      : "Confirm removal"}
                </button>
                <button
                  type="button"
                  style={secondaryButtonStyle}
                  aria-disabled={
                    operation.kind === "pending" ||
                    operation.kind === "uncertain" ||
                    operation.kind === "conflicted"
                  }
                  onClick={controller.cancelEditor}
                >
                  Cancel
                </button>
              </div>
            </form>
          ) : null}
          {nav}
          {page && state.history.length === 20 ? (
            <p style={muted}>
              Previous navigation retains up to 20 pages. First page always
              starts a new live read.
            </p>
          ) : null}
          <div
            aria-busy={busy}
            data-testid={incidentMembershipListTestId()}
            style={stack}
          >
            {page?.rows.map((member) => (
              <article
                className="mm-member"
                key={`${member.incident_id}:${member.user_id}`}
                data-testid={incidentMembershipRowTestId(member.user_id)}
                aria-label={`Member ${member.display_name} (${member.user_id})`}
                style={{ ...card, padding: "var(--mm-cell-padding)" }}
              >
                <strong style={text}>{member.display_name}</strong>
                <span style={muted}>{member.user_id}</span>
                <span
                  data-testid={incidentMembershipRoleDisplayTestId(
                    member.user_id,
                  )}
                  style={text}
                >
                  {member.role}
                </span>
                <span
                  data-testid={incidentMembershipVersionTestId(member.user_id)}
                  style={muted}
                >
                  Version {member.membership_version}
                </span>
                {admin ? (
                  <div style={actions}>
                    <button
                      type="button"
                      style={secondaryButtonStyle}
                      aria-label={`Change role for ${member.display_name} (${member.user_id})`}
                      aria-disabled={editBlocked}
                      onClick={() => {
                        if (!editBlocked) controller.openMember(member, "role");
                      }}
                    >
                      Change role
                    </button>
                    <button
                      type="button"
                      style={secondaryButtonStyle}
                      data-testid={incidentMembershipDeleteButtonTestId(
                        member.user_id,
                      )}
                      aria-label={`Remove incident access for ${member.display_name} (${member.user_id})`}
                      aria-disabled={editBlocked}
                      onClick={() => {
                        if (!editBlocked)
                          controller.openMember(member, "remove");
                      }}
                    >
                      Remove incident access
                    </button>
                  </div>
                ) : null}
              </article>
            ))}
          </div>
        </>
      ) : null}
      <IncidentMembershipDepartureDialog
        controller={controller}
        kind="editor"
        fallbackFocusRef={heading}
      />
    </section>
  );
}

function MembershipOperationFeedback({
  controller,
  operation: op,
  canRecover,
  protectedVisible,
}: {
  readonly controller: IncidentMembershipManagementController;
  readonly operation: MembershipOperation;
  readonly canRecover: boolean;
  readonly protectedVisible: boolean;
}) {
  if (op.kind === "idle") return null;
  if (op.kind === "retired")
    return (
      <p role="status" style={text}>
        {op.message}
      </p>
    );
  if (!protectedVisible)
    return (
      <section aria-label="Membership operation" style={card}>
        <p
          role={
            op.kind === "uncertain" ||
            op.kind === "conflicted" ||
            op.kind === "rejected"
              ? "alert"
              : "status"
          }
          style={text}
        >
          {op.kind === "confirmed"
            ? "The membership write is confirmed. Current access must be checked before showing its details; the write will not be resent."
            : op.kind === "pending"
              ? "A membership action is pending. Leaving cannot cancel server execution."
              : op.kind === "uncertain"
                ? "The previous membership action remains unconfirmed. Check current access to recover."
                : "The membership request was rejected. Check current access to review its details."}
        </p>
      </section>
    );
  const title =
    op.attempt.input.kind === "add"
      ? `Add ${op.attempt.label}`
      : op.attempt.input.kind === "role"
        ? `Change role for ${op.attempt.label}`
        : `Remove incident access for ${op.attempt.label}`;
  return (
    <section aria-label="Membership operation" style={card}>
      <strong style={text}>{title}</strong>
      {op.kind === "pending" ? (
        <p role="status" style={text}>
          {op.stage === "authorization"
            ? "Checking administrator access…"
            : "Saving the reviewed membership action…"}{" "}
          Closing this panel cannot cancel a server operation.
        </p>
      ) : null}
      {op.kind === "confirmed" ? (
        <>
          <p role="status" style={text}>
            {membershipConfirmedText(op)}
          </p>
          {op.listRefresh === "pending" || op.accessRefresh === "pending" ? (
            <p style={muted}>
              The write is confirmed. Refreshing current members and access.
            </p>
          ) : null}
          {op.listRefresh === "failed" || op.accessRefresh === "failed" ? (
            <div style={stack}>
              <p role="alert" style={text}>
                The write is confirmed, but follow-up information could not
                refresh. Recover the read or access check; the write will not be
                resent.
              </p>
              <div style={actions}>
                {op.listRefresh === "failed" ? (
                  <button
                    type="button"
                    style={secondaryButtonStyle}
                    onClick={controller.refresh}
                  >
                    Retry member refresh
                  </button>
                ) : null}
                {op.accessRefresh === "failed" ? (
                  <button
                    type="button"
                    style={secondaryButtonStyle}
                    onClick={() => {
                      void controller.recoverAccess();
                    }}
                  >
                    Retry access check
                  </button>
                ) : null}
              </div>
            </div>
          ) : null}
        </>
      ) : null}
      {op.kind === "rejected" || op.kind === "conflicted" ? (
        <p role="alert" style={text}>
          {membershipProblemText(op.problem)}
        </p>
      ) : null}
      {op.kind === "uncertain" ? (
        <p role="alert" style={text}>
          The membership action has an uncertain outcome. A timeout, matching
          current state, or missing row does not confirm whether this attempt
          succeeded.
        </p>
      ) : null}
      {op.kind === "uncertain" || op.kind === "conflicted" ? (
        <>
          <p style={muted}>
            {op.attempt.input.kind === "add"
              ? `Reviewed email: ${op.attempt.input.payload.email}; role: ${op.attempt.input.payload.role}.`
              : `Reviewed member: ${op.attempt.input.userId}; version: ${op.attempt.input.payload.base_membership_version}.`}
          </p>
          <div style={actions}>
            {op.kind === "uncertain" && op.attempt.input.kind === "add" ? (
              <button
                type="button"
                style={secondaryButtonStyle}
                aria-disabled={!canRecover}
                onClick={() => {
                  if (canRecover) controller.replay();
                }}
              >
                Replay exact add request
              </button>
            ) : (
              <>
                <button
                  type="button"
                  style={secondaryButtonStyle}
                  onClick={controller.observeCurrent}
                >
                  Observe current membership
                </button>
                <button
                  type="button"
                  style={secondaryButtonStyle}
                  aria-disabled={!canRecover || !op.observed}
                  onClick={controller.reviewObserved}
                >
                  Review observed membership
                </button>
              </>
            )}
            <button
              type="button"
              style={secondaryButtonStyle}
              aria-disabled={!canRecover}
              onClick={() => {
                if (canRecover) controller.forgetRecovery();
              }}
            >
              Discard draft and dismiss recovery
            </button>
          </div>
          {op.attempt.input.kind !== "add" ? (
            <p style={muted}>
              {op.observed
                ? `Observed role: ${op.observed.role}; version: ${op.observed.membership_version}. Review before a new attempt.`
                : "Observe the current membership pages and use Next page to locate the member. Absence is not a receipt; no write will be retried automatically."}
            </p>
          ) : null}
          {!canRecover ? (
            <p style={muted}>
              Recovery writes require current administrator access and a settled
              transport.
            </p>
          ) : null}
        </>
      ) : null}
    </section>
  );
}

export function IncidentMembershipDepartureDialog({
  controller,
  kind,
  fallbackFocusRef,
}: {
  readonly controller: IncidentMembershipManagementController;
  readonly kind: "editor" | "route";
  readonly fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  if (state.departure?.kind !== kind) return null;
  return (
    <AccountDialog
      labelledBy="membership-departure-title"
      describedBy="membership-departure-description"
      fallbackFocusRef={fallbackFocusRef}
      onClose={() => controller.resolveDeparture("stay")}
      style={card}
    >
      {(dismiss) => (
        <>
          <h2 id="membership-departure-title" style={text}>
            {kind === "route"
              ? "Leave membership work?"
              : "Discard this membership draft?"}
          </h2>
          <p id="membership-departure-description" style={text}>
            {kind === "route"
              ? "Leaving discards the draft and forgets local recovery for this incident. It cannot cancel a pending server operation or establish whether an uncertain operation succeeded."
              : "Stay to continue editing, or discard this draft before changing the active editor. Saving is an explicit action in the form."}
          </p>
          <div style={actions}>
            <button
              type="button"
              style={secondaryButtonStyle}
              onClick={dismiss}
            >
              Stay
            </button>
            <button
              type="button"
              style={danger}
              onClick={() => controller.resolveDeparture("discard")}
            >
              {kind === "route" ? "Leave and forget recovery" : "Discard draft"}
            </button>
          </div>
        </>
      )}
    </AccountDialog>
  );
}
