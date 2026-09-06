import { accountTestId } from "@cartulary/ui-contracts";
import { useEffect, useMemo } from "react";
import {
  type AccountEditState,
  type AccountResourceKind,
  type AccountSettingsController,
  accountDraftDirty,
  accountReviewRequired,
  accountSavedValue,
  canSaveAccountEdit,
} from "./accountSettingsModel";
import {
  accountDensityChoiceStyle,
  accountDensityGroupStyle,
  definitionLabelStyle,
  definitionValueStyle,
  errorTextStyle,
  formGridStyle,
  inputStyle,
  labelBlockStyle,
  primaryButtonStyle,
  secondaryButtonStyle,
  sectionEyebrowStyle,
  sectionTitleStyle,
  statusTextStyle,
  surfacePanelStyle,
} from "./landingAdminStyles";

type Props = {
  controller: AccountSettingsController;
  state: AccountEditState;
  lifetime: string | null;
};
const densityChoices = [
  { value: null, label: "Use surface default" },
  { value: "compact", label: "Compact" },
  { value: "default", label: "Default" },
  { value: "comfortable", label: "Comfortable" },
] as const;
const densityLabel = (value: string | null) =>
  densityChoices.find((choice) => choice.value === value)?.label ??
  "Use surface default";

export function AccountProfilePanel({ controller, state, lifetime }: Props) {
  const commands = useMemo(
    () => controller.bind("profile", lifetime),
    [controller, lifetime],
  );
  useEffect(() => commands.open(), [commands]);
  return (
    <section className="cartulary-account-edit" style={surfacePanelStyle}>
      <AccountEditFocus />
      <p style={sectionEyebrowStyle}>Profile</p>
      <h2 style={sectionTitleStyle}>Account profile</h2>
      <form
        style={formGridStyle}
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          commands.submit();
        }}
      >
        <div>
          <span style={definitionLabelStyle}>Email</span>
          <div
            data-testid={accountTestId("profile-email")}
            id="account-profile-email"
            style={{ ...definitionValueStyle, overflowWrap: "anywhere" }}
          >
            {state.saved !== null && "email" in state.saved
              ? state.saved.email
              : ""}
          </div>
          <p style={statusTextStyle}>
            Email is managed by your deployment. You can edit your display name
            here.
          </p>
        </div>
        <label htmlFor="account-profile-display-name" style={labelBlockStyle}>
          Display name
          <input
            data-testid={accountTestId("profile-display-name")}
            id="account-profile-display-name"
            style={{
              ...inputStyle,
              minInlineSize: 0,
              inlineSize: "100%",
              boxSizing: "border-box",
            }}
            value={state.draft ?? ""}
            disabled={state.saved === null}
            aria-invalid={state.fieldError !== null}
            aria-describedby={
              state.fieldError === null ? undefined : "account-profile-feedback"
            }
            onChange={(event) => commands.change(event.target.value)}
          />
        </label>
        <AccountSavedValue kind="profile" state={state} />
        <AccountEditActions kind="profile" commands={commands} state={state} />
        <AccountEditFeedback kind="profile" state={state} />
      </form>
    </section>
  );
}

export function AccountAppearancePanel({ controller, state, lifetime }: Props) {
  const commands = useMemo(
    () => controller.bind("appearance", lifetime),
    [controller, lifetime],
  );
  useEffect(() => commands.open(), [commands]);
  return (
    <section className="cartulary-account-edit" style={surfacePanelStyle}>
      <AccountEditFocus />
      <p style={sectionEyebrowStyle}>Appearance</p>
      <h2 style={sectionTitleStyle}>Density</h2>
      <form
        style={formGridStyle}
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          commands.submit();
        }}
      >
        <div
          role="radiogroup"
          aria-labelledby="account-density-label"
          aria-invalid={state.fieldError !== null}
          aria-describedby={
            state.fieldError === null
              ? undefined
              : "account-appearance-feedback"
          }
          data-testid={accountTestId("appearance-density-mode")}
        >
          <p id="account-density-label" style={definitionLabelStyle}>
            Density
          </p>
          <div style={accountDensityGroupStyle}>
            {densityChoices.map((choice) => (
              <label
                key={choice.value ?? "surface-default"}
                style={{
                  ...accountDensityChoiceStyle,
                  background:
                    state.draft === choice.value
                      ? "var(--ct-colors-surface-3)"
                      : "var(--ct-colors-surface-1)",
                  fontWeight: state.draft === choice.value ? 700 : 400,
                }}
              >
                <input
                  type="radio"
                  name="account-density"
                  value={choice.value ?? "surface-default"}
                  checked={state.draft === choice.value}
                  disabled={state.saved === null}
                  style={{
                    margin: 0,
                    flexShrink: 0,
                    accentColor: "var(--ct-colors-ink)",
                  }}
                  onChange={() => commands.change(choice.value)}
                />
                <span>{choice.label}</span>
              </label>
            ))}
          </div>
        </div>
        <p style={statusTextStyle}>
          Use surface default keeps Timeline compact and other workbook surfaces
          at Default. Changes apply after saving.
        </p>
        <AccountSavedValue kind="appearance" state={state} />
        <AccountEditActions
          kind="appearance"
          commands={commands}
          state={state}
        />
        <AccountEditFeedback kind="appearance" state={state} />
      </form>
    </section>
  );
}

function AccountSavedValue({
  kind,
  state,
}: {
  kind: AccountResourceKind;
  state: AccountEditState;
}) {
  if (
    state.saved === null ||
    (!accountDraftDirty(state) &&
      state.operation.kind !== "conflict" &&
      state.operation.kind !== "uncertain")
  )
    return null;
  const saved = accountSavedValue(state.saved);
  return (
    <div style={{ minInlineSize: 0, overflowWrap: "anywhere" }}>
      <span style={definitionLabelStyle}>
        Saved {kind === "profile" ? "display name" : "density"}
      </span>
      <p style={definitionValueStyle}>
        {kind === "profile" ? saved : densityLabel(saved)}
      </p>
      {accountDraftDirty(state) ? (
        <p style={statusTextStyle}>
          The {kind === "profile" ? "field" : "selection"} above contains your
          unsaved edit.
        </p>
      ) : null}
    </div>
  );
}

function AccountEditActions({
  kind,
  commands,
  state,
}: {
  commands: ReturnType<AccountSettingsController["bind"]>;
  state: AccountEditState;
  kind: AccountResourceKind;
}) {
  const operation = state.operation;
  const recovery =
    accountReviewRequired(state) ||
    operation.kind === "conflict" ||
    (operation.kind === "rejected" &&
      (operation.reason === "transaction" ||
        operation.reason === "authorization"));
  const unresolved =
    operation.kind === "saving" || operation.kind === "uncertain";
  const loading = state.read === "loading" || state.read === "refreshing";
  const canSave = canSaveAccountEdit(state);
  return (
    <div
      style={{ display: "flex", flexWrap: "wrap", gap: "var(--ct-spacing-sm)" }}
    >
      <button
        data-testid={accountTestId(
          kind === "profile" ? "profile-save" : "appearance-save",
        )}
        type="submit"
        disabled={!canSave}
        aria-busy={operation.kind === "saving"}
        style={
          canSave
            ? primaryButtonStyle
            : {
                ...secondaryButtonStyle,
                cursor: "default",
                color: "var(--ct-colors-ink-muted)",
              }
        }
      >
        Save {kind}
      </button>
      {operation.kind === "uncertain" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={() => commands.replay()}
        >
          Retry save
        </button>
      ) : null}
      <button
        type="button"
        disabled={loading}
        style={secondaryButtonStyle}
        onClick={() => {
          void commands.refresh();
        }}
      >
        {state.saved === null && state.read === "failed"
          ? "Retry load"
          : `Refresh ${kind}`}
      </button>
      {recovery && !unresolved ? (
        <>
          <button
            type="button"
            disabled={state.read !== "ready"}
            style={secondaryButtonStyle}
            onClick={() => commands.review()}
          >
            Review my edit
          </button>
          <button
            type="button"
            disabled={state.read !== "ready"}
            style={secondaryButtonStyle}
            onClick={() => commands.discard()}
          >
            Use saved value
          </button>
        </>
      ) : accountDraftDirty(state) ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={() => commands.discard()}
        >
          Discard edits
        </button>
      ) : null}
      {operation.kind === "confirmed" && operation.publication === "failed" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={() => commands.retryPublication()}
        >
          Retry refresh
        </button>
      ) : null}
    </div>
  );
}

function AccountEditFeedback({
  kind,
  state,
}: {
  kind: AccountResourceKind;
  state: AccountEditState;
}) {
  const operation = state.operation;
  const title = kind === "profile" ? "Profile" : "Appearance";
  let message = "";
  let alert = false;
  if (state.fieldError !== null) {
    message = state.fieldError;
    alert = true;
  } else if (operation.kind === "saving")
    message = operation.replay
      ? "Checking the previous save. You can keep editing."
      : `Saving ${kind}. You can keep editing.`;
  else if (operation.kind === "uncertain")
    message = `We could not confirm whether your ${kind} was saved. Retry save to check the same request.${operation.reason === "authorization" ? " Access is currently denied." : ""} Discard edits does not cancel that request.`;
  else if (operation.kind === "conflict") {
    message = `${kind === "profile" ? "Display name" : "Density"} changed elsewhere. Your edit is retained. Review the saved value before saving again.`;
    alert = true;
  } else if (operation.kind === "rejected") {
    message =
      operation.reason === "transaction"
        ? "This save request could not be used. Review your edit before starting another save."
        : operation.reason === "authorization"
          ? "Saving is currently denied. Refresh this resource, then review your edit before trying again."
          : operation.reason === "preparation"
            ? "The save could not be prepared. Try again."
            : "The edit was not accepted. Check the field and try again.";
    alert = operation.reason !== "transaction";
  } else if (operation.kind === "confirmed") {
    message = `${title} saved.${accountDraftDirty(state) ? " You have unsaved edits." : ""}`;
    if (operation.publication === "pending")
      message += " Updating account labels.";
    if (operation.publication === "failed")
      message +=
        " Account labels could not refresh. Retry refresh; your save is already confirmed.";
  } else if (accountDraftDirty(state)) message = "Unsaved changes.";
  if (accountReviewRequired(state) && operation.kind !== "conflict")
    message +=
      " The saved value changed. Review your edit before saving again.";
  if (state.read === "loading" || state.read === "unresolved")
    message = `Loading account ${kind}.`;
  else if (state.read === "refreshing") message += ` Refreshing ${kind}.`;
  else if (state.read === "failed" && state.readFailure === "authorization")
    message += ` Access to account ${kind} is currently denied. Refresh ${kind} to check access.`;
  else if (state.read === "failed")
    message +=
      state.saved === null
        ? ` Account ${kind} is unavailable. Retry load.`
        : ` ${title} could not refresh. Showing the last saved value. Retry refresh using Refresh ${kind}.`;
  return (
    <p
      id={`account-${kind}-feedback`}
      role={alert ? "alert" : "status"}
      style={alert ? errorTextStyle : statusTextStyle}
    >
      {message.trim()}
    </p>
  );
}

function AccountEditFocus() {
  return (
    <style>{`.cartulary-account-edit :where(input, button):focus-visible { outline: var(--ct-component-focus-ring-border); outline-offset: var(--ct-component-focus-ring-offset); }`}</style>
  );
}
