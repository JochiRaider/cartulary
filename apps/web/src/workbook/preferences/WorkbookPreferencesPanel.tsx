import {
  incidentAdministrationTestId,
  workbookPreferenceTestId,
} from "@cartulary/ui-contracts";
import { getViewContract } from "@cartulary/view-contracts";
import {
  type CSSProperties,
  useId,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookDensityMode } from "../../shared/workbookShellContracts";
import { visuallyHiddenStyle } from "../utils/workbookStyles";
import type { WorkbookPreferenceController } from "./WorkbookPreferenceController";
import {
  type PreferenceKind,
  type PreferenceSlot,
  type PreferenceSurface,
  preferencePointer,
  preferencePointersEqual,
  preferenceProblemText,
} from "./workbookPreferenceModel";

const emptySubscribe = () => () => {};
const emptySnapshot = () => null;
export function useWorkbookPreferencesSnapshot(
  controller?: WorkbookPreferenceController,
) {
  return useSyncExternalStore(
    controller?.subscribe ?? emptySubscribe,
    controller?.getSnapshot ?? emptySnapshot,
  );
}
export function formatPreferencePointer(
  pointer: SheetRef | null,
  surface?: PreferenceSurface | null,
): string {
  if (pointer === null) return "Unset";
  const known =
    pointer.kind === "saved_view"
      ? (surface?.savedViewLabels?.find((label) => label.id === pointer.id)
          ?.label ??
        (surface && preferencePointersEqual(surface.sheetRef, pointer)
          ? surface.label
          : null))
      : surface && preferencePointersEqual(surface.sheetRef, pointer)
        ? surface.label
        : null;
  if (pointer.kind === "view_schema")
    return `View schema: ${getViewContract(pointer.id)?.title ?? known ?? pointer.id} (${pointer.id})`;
  if (pointer.kind === "saved_view")
    return `Saved view: ${known ? `${known} (` : ""}${pointer.id}${known ? ")" : ""}`;
  return `Extension workspace: ${known ? `${known} (` : ""}${pointer.extension_profile_id}/${pointer.workspace_key}${known ? ")" : ""}`;
}
export function preferenceOutcome(
  kind: PreferenceKind,
  slot: PreferenceSlot,
): string {
  const name = kind === "home" ? "Home" : "Incident default";
  const op = slot.operation;
  switch (op.kind) {
    case "idle":
      return "";
    case "reviewed":
      return "Keeping the observed value. No new write was sent. The earlier result remains unknown.";
    case "pending":
      return op.stage === "authorization"
        ? `Checking current access for ${name.toLowerCase()}…`
        : `${name} request sent. Waiting for acknowledgement…`;
    case "confirmed":
      return `${name} ${op.attempt.target === null ? "clear" : "update"} confirmed.`;
    case "uncertain":
      return `${name} update is uncertain. A current value cannot establish whether this attempt succeeded.`;
    case "rejected":
      return preferenceProblemText(op.problem);
  }
}
export function WorkbookPreferenceAnnouncements({
  controller,
}: {
  controller: WorkbookPreferenceController;
}) {
  const state = useWorkbookPreferencesSnapshot(controller);
  return (
    <p
      style={visuallyHiddenStyle}
      role="status"
      aria-label="Workbook preference updates"
      aria-live="polite"
      aria-atomic="true"
    >
      <span key={state?.announcement?.id}>
        {state?.announcement?.text ?? ""}
      </span>
    </p>
  );
}
export function WorkbookPreferencesPanel({
  controller,
  density = "default",
}: {
  controller: WorkbookPreferenceController;
  density?: WorkbookDensityMode | undefined;
}) {
  const state = useWorkbookPreferencesSnapshot(controller);
  const prefix = useId();
  useLayoutEffect(() => {
    controller.setInspectionActive(true);
    return () => controller.setInspectionActive(false);
  }, [controller]);
  if (!state) return null;
  return (
    <section
      aria-labelledby={`${prefix}-heading`}
      data-workbook-preferences=""
      data-density={density}
      style={{
        ...card,
        padding: `var(--ct-density-${density}-cellPadding)`,
        fontSize: `var(--ct-density-${density}-fontSize)`,
      }}
    >
      <style>{`[data-workbook-preferences] button:focus-visible { outline: var(--ct-border-focus); outline-offset: var(--ct-spacing-xs); } [data-workbook-preferences] button[aria-disabled="true"] { color: var(--ct-colors-ink-muted); cursor: not-allowed; }`}</style>
      <h3 id={`${prefix}-heading`} style={text}>
        Workbook startup preferences
      </h3>
      <p style={muted}>
        These are stored pointers for a future workbook open. They need not
        match the active surface or the effective next startup selection.
      </p>
      <p style={muted}>
        Startup order: explicit launch → valid personal home → valid incident
        default → Timeline.
      </p>
      <p style={muted}>
        Setting a pointer does not save unsaved query or layout changes.
        Preferences remain available on closed incidents under ordinary
        permissions.
      </p>
      {(["home", "default"] as const).map((kind) => (
        <PreferenceRow key={kind} kind={kind} controller={controller} />
      ))}
    </section>
  );
}
function PreferenceRow({
  kind,
  controller,
}: {
  kind: PreferenceKind;
  controller: WorkbookPreferenceController;
}) {
  const state = useWorkbookPreferencesSnapshot(controller);
  const prefix = useId();
  const refreshButton = useRef<HTMLButtonElement>(null);
  if (!state) return null;
  const slot = state[kind];
  const op = slot.operation;
  const name = kind === "home" ? "My home surface" : "Incident default surface";
  const value = slot.resource
    ? formatPreferencePointer(preferencePointer(slot.resource), state.surface)
    : slot.read === "failed"
      ? "Unavailable"
      : "Loading…";
  const readText =
    slot.read === "refreshing"
      ? "Refreshing stored value…"
      : slot.read === "failed"
        ? slot.resource
          ? "Displayed stored value is retained and may be stale. Refresh this resource."
          : "The stored value is unavailable. Refresh this resource."
        : slot.read === "loading" || slot.read === "idle"
          ? "Loading stored value…"
          : "";
  const blocked = !state.authority
    ? "Current incident access is unresolved."
    : kind === "default" && state.authority.role !== "admin"
      ? "Only current incident admins can set or clear the incident default."
      : slot.transportPending
        ? "The earlier transport has not settled. Timeout or abort does not establish server cancellation."
        : op.kind === "uncertain"
          ? "Observe the current value and explicitly resolve this uncertainty before another write."
          : op.kind === "pending"
            ? "An operation for this preference is already pending."
            : !state.surface?.available
              ? "Set requires a current authorized surface. Clear remains independent of surface availability."
              : "";
  const canResolve =
    op.kind === "uncertain" &&
    controller.canResolve(kind, op.attempt.id, slot.observation);
  return (
    <section
      style={row}
      aria-labelledby={`${prefix}-heading`}
      data-testid={workbookPreferenceTestId(kind, "row")}
    >
      <h4 id={`${prefix}-heading`} style={text}>
        {name}
      </h4>
      <p
        style={valueStyle}
        data-testid={incidentAdministrationTestId(
          kind === "home" ? "pref-home-sheet-ref" : "pref-default-sheet-ref",
        )}
      >
        {value}
      </p>
      <p style={muted} data-testid={workbookPreferenceTestId(kind, "read")}>
        {readText}
      </p>
      <div style={actions}>
        <button
          type="button"
          style={button}
          aria-label={`Set current surface as ${kind === "home" ? "my home" : "incident default"}`}
          aria-disabled={!controller.canSetCurrent(kind)}
          aria-describedby={`${prefix}-help`}
          data-testid={workbookPreferenceTestId(kind, "set")}
          onClick={() => controller.setCurrent(kind)}
        >
          Set current surface
        </button>
        <button
          type="button"
          style={button}
          aria-disabled={!controller.canWrite(kind)}
          aria-describedby={`${prefix}-help`}
          data-testid={workbookPreferenceTestId(kind, "clear")}
          onClick={() => controller.clear(kind)}
        >
          {kind === "home" ? "Clear my home" : "Clear incident default"}
        </button>
        <button
          ref={refreshButton}
          type="button"
          style={button}
          aria-disabled={
            slot.read === "loading" ||
            slot.read === "refreshing" ||
            op.kind === "pending"
          }
          data-testid={workbookPreferenceTestId(kind, "refresh")}
          onClick={() => {
            if (
              slot.read !== "loading" &&
              slot.read !== "refreshing" &&
              op.kind !== "pending"
            )
              controller.refresh(kind);
          }}
        >
          Refresh {kind === "home" ? "my home" : "incident default"}
        </button>
      </div>
      <p id={`${prefix}-help`} style={muted}>
        {blocked ||
          (state.surface
            ? `Current surface: ${formatPreferencePointer(state.surface.sheetRef, state.surface)}`
            : "")}
      </p>
      <p style={text} data-testid={workbookPreferenceTestId(kind, "outcome")}>
        {preferenceOutcome(kind, slot)}
      </p>
      {op.kind === "confirmed" && slot.read === "failed" ? (
        <p style={text}>
          The update is confirmed, but its current value could not be refreshed.
          Retry the read; do not resubmit the acknowledged write.
        </p>
      ) : null}
      {op.kind === "confirmed" &&
      slot.resource &&
      !preferencePointersEqual(
        preferencePointer(slot.resource),
        op.attempt.target,
      ) ? (
        <p style={muted}>
          The current stored value differs from the acknowledged update. Another
          change may have occurred.
        </p>
      ) : null}
      {op.kind === "uncertain" ? (
        <div
          style={row}
          data-testid={workbookPreferenceTestId(kind, "recovery")}
        >
          <p style={text}>
            Captured target: {formatPreferencePointer(op.attempt.target)}
          </p>
          <p style={text}>
            Current observation:{" "}
            {slot.read === "ready" && slot.observation !== null && slot.resource
              ? formatPreferencePointer(
                  preferencePointer(slot.resource),
                  state.surface,
                )
              : "Refresh after the earlier transport settles."}
          </p>
          <p style={muted}>
            {preferenceProblemText(op.problem)} A matching GET is not a receipt
            for this attempt. Writing again sends a new request and may
            overwrite an intervening preference change. Keeping the observed
            value ends only local recovery; it cannot cancel or undo the earlier
            request.
          </p>
          <div style={actions}>
            <button
              type="button"
              style={button}
              aria-disabled={
                !canResolve ||
                (kind === "default" && state.authority?.role !== "admin")
              }
              data-testid={workbookPreferenceTestId(kind, "write-captured")}
              onClick={() => {
                if (
                  canResolve &&
                  (kind === "home" || state.authority?.role === "admin")
                ) {
                  refreshButton.current?.focus();
                  controller.resolve(
                    kind,
                    op.attempt.id,
                    slot.observation,
                    "write",
                  );
                }
              }}
            >
              Write captured target again
            </button>
            <button
              type="button"
              style={button}
              aria-disabled={!canResolve}
              data-testid={workbookPreferenceTestId(kind, "keep-observed")}
              onClick={() => {
                if (canResolve) {
                  refreshButton.current?.focus();
                  controller.resolve(
                    kind,
                    op.attempt.id,
                    slot.observation,
                    "keep",
                  );
                }
              }}
            >
              Keep observed value
            </button>
          </div>
        </div>
      ) : null}
    </section>
  );
}
const text: CSSProperties = { margin: 0, overflowWrap: "anywhere" };
const muted: CSSProperties = { ...text, color: "var(--ct-colors-ink-muted)" };
const valueStyle: CSSProperties = { ...text, whiteSpace: "pre-wrap" };
const actions: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
};
const row: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
  borderBlockStart: "var(--ct-border-hairline)",
  paddingBlockStart: "var(--ct-spacing-md)",
};
const card: CSSProperties = {
  gridColumn: "1 / -1",
  display: "grid",
  gap: "var(--ct-spacing-md)",
  minWidth: 0,
  background: "var(--ct-colors-surface-1)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
};
const button: CSSProperties = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-xs)",
  color: "var(--ct-colors-ink)",
  background: "var(--ct-colors-surface-2)",
  padding: "var(--ct-component-button-secondary-padding)",
  font: "inherit",
  maxInlineSize: "100%",
  overflowWrap: "anywhere",
  cursor: "pointer",
};
