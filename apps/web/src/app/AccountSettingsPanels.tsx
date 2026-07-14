import { accountTestId } from "@cartulary/ui-contracts";
import { useCallback, useEffect, useState } from "react";

import {
  type APIError,
  extractError,
  publicErrorView,
} from "../services/browserApi";
import {
  type AccountPreferencesResource,
  type AccountProfileResource,
  type DensityMode,
  loadAccountPreferences,
  loadAccountProfile,
  patchAccountProfile,
  putAccountPreferences,
} from "./api/appShellClient";
import {
  definitionLabelStyle,
  definitionPanelStyle,
  definitionValueStyle,
  errorTextStyle,
  inputStyle,
  labelBlockStyle,
  primaryButtonStyle,
  sectionEyebrowStyle,
  sectionTitleStyle,
  segmentedFormStyle,
  statusTextStyle,
  surfacePanelStyle,
} from "./landingAdminStyles";

export function AccountProfilePanel({
  onRefreshShell,
}: {
  onRefreshShell: () => Promise<void> | void;
}) {
  const [profile, setProfile] = useState<AccountProfileResource | null>(null);
  const [displayName, setDisplayName] = useState("");
  const [status, setStatus] = useState("Loading account profile.");
  const [error, setError] = useState<APIError | null>(null);

  const loadProfile = useCallback(async () => {
    const result = await loadAccountProfile();
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Account profile unavailable.");
      return;
    }
    const nextProfile = (result.payload as { data: AccountProfileResource })
      .data;
    setProfile(nextProfile);
    setDisplayName(nextProfile.display_name);
    setStatus("Account profile loaded.");
  }, []);

  useEffect(() => {
    void loadProfile();
  }, [loadProfile]);

  async function saveProfile() {
    if (profile === null) {
      return;
    }
    setStatus("Saving account profile.");
    const result = await patchAccountProfile({
      baseUserVersion: profile.user_version,
      displayName,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Account profile save failed.");
      return;
    }
    const nextProfile = (result.payload as { data: AccountProfileResource })
      .data;
    setProfile(nextProfile);
    setDisplayName(nextProfile.display_name);
    setStatus("Account profile saved.");
    await onRefreshShell();
  }

  return (
    <section style={surfacePanelStyle}>
      <p style={sectionEyebrowStyle}>Profile</p>
      <h2 style={sectionTitleStyle}>Account profile</h2>
      <div style={definitionPanelStyle}>
        <div>
          <span style={definitionLabelStyle}>Email</span>
          <div
            data-testid={accountTestId("profile-email")}
            id="account-profile-email"
            style={definitionValueStyle}
          >
            {profile?.email ?? ""}
          </div>
        </div>
        <label htmlFor="account-profile-display-name" style={labelBlockStyle}>
          Display name
          <input
            data-testid={accountTestId("profile-display-name")}
            id="account-profile-display-name"
            style={inputStyle}
            value={displayName}
            onChange={(event) => {
              setDisplayName(event.target.value);
            }}
          />
        </label>
        <button
          data-testid={accountTestId("profile-save")}
          disabled={profile === null}
          style={primaryButtonStyle}
          type="button"
          onClick={() => {
            void saveProfile();
          }}
        >
          Save profile
        </button>
      </div>
      <p aria-live="polite" role="status" style={statusTextStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={errorTextStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}

type AccountAppearancePanelProps = {
  readonly preferences?: AccountPreferencesResource | null | undefined;
  readonly onPreferencesChange?:
    | ((preferences: AccountPreferencesResource) => void)
    | undefined;
};

export function AccountAppearancePanel({
  preferences: controlledPreferences,
  onPreferencesChange,
}: AccountAppearancePanelProps = {}) {
  const isControlled = controlledPreferences !== undefined;
  const [localPreferences, setLocalPreferences] =
    useState<AccountPreferencesResource | null>(null);
  const [densityMode, setDensityMode] = useState<DensityMode | "">("");
  const [status, setStatus] = useState("Loading account appearance.");
  const [error, setError] = useState<APIError | null>(null);
  const preferences = isControlled ? controlledPreferences : localPreferences;

  const loadPreferences = useCallback(async () => {
    const result = await loadAccountPreferences();
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Account appearance unavailable.");
      return;
    }
    const nextPreferences = (
      result.payload as { data: AccountPreferencesResource }
    ).data;
    setLocalPreferences(nextPreferences);
    onPreferencesChange?.(nextPreferences);
    setDensityMode(nextPreferences.density_mode ?? "");
    setStatus("Account appearance loaded.");
  }, [onPreferencesChange]);

  useEffect(() => {
    if (!isControlled) {
      void loadPreferences();
    }
  }, [isControlled, loadPreferences]);

  useEffect(() => {
    if (!isControlled) {
      return;
    }
    if (controlledPreferences === null) {
      setDensityMode("");
      setStatus((current) =>
        current === "Loading account appearance."
          ? current
          : "Account appearance unavailable.",
      );
      return;
    }
    setDensityMode(controlledPreferences.density_mode ?? "");
    setStatus((current) =>
      current === "Loading account appearance." ||
      current === "Account appearance unavailable."
        ? "Account appearance loaded."
        : current,
    );
  }, [controlledPreferences, isControlled]);

  async function savePreferences() {
    if (preferences === null) {
      return;
    }
    setStatus("Saving account appearance.");
    const result = await putAccountPreferences({
      basePreferencesVersion: preferences.preferences_version,
      densityMode: densityMode === "" ? null : densityMode,
    });
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Account appearance save failed.");
      if (result.status === 409) {
        void loadPreferences();
      }
      return;
    }
    const nextPreferences = (
      result.payload as { data: AccountPreferencesResource }
    ).data;
    setLocalPreferences(nextPreferences);
    onPreferencesChange?.(nextPreferences);
    setDensityMode(nextPreferences.density_mode ?? "");
    setStatus("Account appearance saved.");
  }

  return (
    <section style={surfacePanelStyle}>
      <p style={sectionEyebrowStyle}>Appearance</p>
      <h2 style={sectionTitleStyle}>Density</h2>
      <div style={segmentedFormStyle}>
        <label htmlFor="account-density-mode" style={labelBlockStyle}>
          Density
          <select
            data-testid={accountTestId("appearance-density-mode")}
            id="account-density-mode"
            style={inputStyle}
            value={densityMode}
            onChange={(event) => {
              setDensityMode(event.target.value as DensityMode | "");
            }}
          >
            <option value="">Use surface default</option>
            <option value="compact">Compact</option>
            <option value="default">Default</option>
            <option value="comfortable">Comfortable</option>
          </select>
        </label>
        <button
          data-testid={accountTestId("appearance-save")}
          disabled={preferences === null}
          style={primaryButtonStyle}
          type="button"
          onClick={() => {
            void savePreferences();
          }}
        >
          Save appearance
        </button>
      </div>
      <p aria-live="polite" role="status" style={statusTextStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={errorTextStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}
