import {
  incidentLandingTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  publicErrorCodeTestId,
  publicErrorSummaryTestIds,
} from "@cartulary/ui-contracts";
import { ChevronRight, RefreshCw, Search, X } from "lucide-react";
import { useId, useState } from "react";

import { publicErrorView } from "../services/browserApi";
import {
  formatNullableDateTime,
  MutedUnset,
  PublicErrorSummary,
  StatusBadge,
} from "./LandingAdminDisplay";
import {
  countPillStyle,
  createDialogStyle,
  dataTableStyle,
  detailsStyle,
  detailsSummaryStyle,
  dialogBackdropStyle,
  dialogHeaderStyle,
  emptyStateStyle,
  errorTextStyle,
  formGridStyle,
  headerActionRowStyle,
  iconButtonStyle,
  incidentDirectoryStyle,
  incidentWorkspaceStyle,
  inlineStatusStyle,
  inputStyle,
  labelBlockStyle,
  metadataTextStyle,
  monoMutedStyle,
  primaryButtonStyle,
  primaryCellStyle,
  searchInputShellStyle,
  searchInputStyle,
  secondaryButtonStyle,
  sectionEyebrowStyle,
  sectionTitleStyle,
  statusTextStyle,
  strongTextStyle,
  subsectionTitleStyle,
  surfaceHeaderStyle,
  surfacePanelStyle,
  tableActionCellStyle,
  tableCellStyle,
  tableHeaderActionCellStyle,
  tableHeaderCellStyle,
  tableLinkButtonStyle,
  tableRowStyle,
  tableShellStyle,
  textAreaStyle,
  toolbarGridStyle,
} from "./landingAdminStyles";
import type {
  IncidentLandingProps,
  IncidentStatusFilter,
} from "./landingAdminTypes";

export function IncidentLanding({
  bootstrapState,
  createIncidentCurrentPhase,
  createIncidentDescription,
  createIncidentExternalCase,
  createIncidentKey,
  createIncidentSeverity,
  createIncidentTitle,
  createIncidentTLP,
  error,
  hasMoreIncidents,
  incidents,
  incidentSearch,
  incidentStatusFilter,
  isRefreshing,
  onCreate,
  onCreateIncidentCurrentPhaseChange,
  onCreateIncidentDescriptionChange,
  onCreateIncidentExternalCaseChange,
  onCreateIncidentKeyChange,
  onCreateIncidentSeverityChange,
  onCreateIncidentTitleChange,
  onCreateIncidentTLPChange,
  onLoadMore,
  onOpenIncident,
  onRefresh,
  onSearchChange,
  onSearchSubmit,
  onStatusFilterChange,
  statusText,
}: IncidentLandingProps) {
  const incidentKeyFieldId = useId();
  const incidentTitleFieldId = useId();
  const [createOpen, setCreateOpen] = useState(false);
  const hasIncidents = incidents.length > 0;
  const hasActiveQuery =
    incidentSearch.trim() !== "" || incidentStatusFilter !== "all";

  async function submitCreate() {
    await onCreate();
  }

  return (
    <section
      aria-busy={isRefreshing}
      data-bootstrap-state={bootstrapState}
      data-testid={incidentLandingTestId("shell")}
      style={surfacePanelStyle}
    >
      <header style={surfaceHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Incidents</p>
          <h2 style={sectionTitleStyle}>Visible incidents</h2>
        </div>
        <div style={headerActionRowStyle}>
          <p
            data-testid={incidentLandingTestId("incidents-count")}
            style={countPillStyle}
          >
            {incidents.length} loaded{hasMoreIncidents ? " +" : ""}
          </p>
          <button
            data-testid={incidentLandingTestId("refresh")}
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void onRefresh();
            }}
          >
            <RefreshCw size={15} />
            Refresh
          </button>
          {!createOpen ? (
            <button
              data-testid={incidentLandingTestId("create-open-button")}
              style={primaryButtonStyle}
              type="button"
              onClick={() => {
                setCreateOpen(true);
              }}
            >
              New incident
            </button>
          ) : null}
        </div>
      </header>

      <div style={incidentWorkspaceStyle}>
        <section style={incidentDirectoryStyle}>
          <div style={toolbarGridStyle}>
            <label htmlFor="incident-filter" style={labelBlockStyle}>
              Search visible incidents
              <span style={searchInputShellStyle}>
                <Search size={16} />
                <input
                  id="incident-filter"
                  data-testid={incidentLandingTestId("search")}
                  style={searchInputStyle}
                  value={incidentSearch}
                  onChange={(event) => {
                    onSearchChange(event.target.value);
                  }}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      void onSearchSubmit();
                    }
                  }}
                  placeholder="Key, title, severity, TLP, phase, external case"
                />
              </span>
            </label>
            <label htmlFor="incident-status-filter" style={labelBlockStyle}>
              Status
              <select
                id="incident-status-filter"
                data-testid={incidentLandingTestId("status-filter")}
                style={inputStyle}
                value={incidentStatusFilter}
                onChange={(event) => {
                  onStatusFilterChange(
                    event.target.value as IncidentStatusFilter,
                  );
                }}
              >
                <option value="all">All</option>
                <option value="active">Active</option>
                <option value="closed">Closed</option>
              </select>
            </label>
          </div>

          {isRefreshing ? (
            <p
              aria-live="polite"
              data-testid={incidentLandingTestId("loading")}
              role="status"
              style={inlineStatusStyle}
            >
              Searching visible incidents…
            </p>
          ) : null}

          {!isRefreshing && !hasIncidents ? (
            <p
              data-testid={incidentLandingTestId("empty-state")}
              style={emptyStateStyle}
            >
              {hasActiveQuery
                ? "No visible incidents match this query."
                : "No incidents are visible for this session yet."}
            </p>
          ) : null}

          {hasIncidents ? (
            <div
              data-testid={incidentLandingTestId("incident-list")}
              style={tableShellStyle}
            >
              <table style={dataTableStyle}>
                <thead>
                  <tr>
                    <th style={tableHeaderCellStyle}>Incident</th>
                    <th style={tableHeaderCellStyle}>Status</th>
                    <th style={tableHeaderCellStyle}>Phase</th>
                    <th style={tableHeaderCellStyle}>Severity</th>
                    <th style={tableHeaderCellStyle}>TLP</th>
                    <th style={tableHeaderCellStyle}>External case</th>
                    <th style={tableHeaderCellStyle}>Updated</th>
                    <th style={tableHeaderActionCellStyle}>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {incidents.map((incident) => (
                    <tr
                      key={incident.incident_id}
                      data-testid={landingIncidentCardTestId(
                        incident.incident_id,
                      )}
                      style={tableRowStyle}
                    >
                      <td style={primaryCellStyle}>
                        <p style={monoMutedStyle}>{incident.incident_key}</p>
                        <p style={strongTextStyle}>{incident.title}</p>
                        <p style={metadataTextStyle}>
                          v{incident.incident_version}
                        </p>
                      </td>
                      <td style={tableCellStyle}>
                        <StatusBadge value={incident.status ?? "active"} />
                      </td>
                      <td style={tableCellStyle}>
                        <MutedUnset value={incident.current_phase} />
                      </td>
                      <td style={tableCellStyle}>
                        <MutedUnset value={incident.severity} />
                      </td>
                      <td style={tableCellStyle}>
                        <MutedUnset value={incident.tlp} />
                      </td>
                      <td style={tableCellStyle}>
                        <MutedUnset
                          value={incident.primary_external_case_ref}
                        />
                      </td>
                      <td
                        style={tableCellStyle}
                        title={incident.updated_at ?? undefined}
                      >
                        {formatNullableDateTime(incident.updated_at)}
                      </td>
                      <td style={tableActionCellStyle}>
                        <button
                          data-testid={landingIncidentOpenButtonTestId(
                            incident.incident_id,
                          )}
                          style={tableLinkButtonStyle}
                          type="button"
                          onClick={() => {
                            onOpenIncident(incident.incident_id);
                          }}
                        >
                          Open
                          <ChevronRight size={15} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : null}

          {hasMoreIncidents ? (
            <button
              style={secondaryButtonStyle}
              type="button"
              onClick={() => {
                void onLoadMore();
              }}
            >
              Load more incidents
            </button>
          ) : null}
        </section>
      </div>

      {createOpen ? (
        <div style={dialogBackdropStyle}>
          <section
            aria-label="New incident"
            aria-modal="true"
            role="dialog"
            style={createDialogStyle}
          >
            <header style={dialogHeaderStyle}>
              <div>
                <p style={sectionEyebrowStyle}>Create</p>
                <h3 style={subsectionTitleStyle}>New incident</h3>
              </div>
              <button
                aria-label="Close new incident"
                style={iconButtonStyle}
                type="button"
                onClick={() => {
                  setCreateOpen(false);
                }}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </header>
            <div style={formGridStyle}>
              <label htmlFor={incidentKeyFieldId} style={labelBlockStyle}>
                Incident key
                <input
                  data-testid={incidentLandingTestId("incident-key")}
                  id={incidentKeyFieldId}
                  style={inputStyle}
                  value={createIncidentKey}
                  onChange={(event) => {
                    onCreateIncidentKeyChange(event.target.value);
                  }}
                  placeholder="IR-2026-001"
                />
              </label>
              <label htmlFor={incidentTitleFieldId} style={labelBlockStyle}>
                Title
                <input
                  data-testid={incidentLandingTestId("incident-title")}
                  id={incidentTitleFieldId}
                  style={inputStyle}
                  value={createIncidentTitle}
                  onChange={(event) => {
                    onCreateIncidentTitleChange(event.target.value);
                  }}
                  placeholder="Credential theft investigation"
                />
              </label>
            </div>
            <details style={detailsStyle}>
              <summary style={detailsSummaryStyle}>More details</summary>
              <div style={formGridStyle}>
                <label
                  htmlFor="incident-create-description"
                  style={labelBlockStyle}
                >
                  Description
                  <textarea
                    data-testid={incidentLandingTestId("create-description")}
                    id="incident-create-description"
                    style={textAreaStyle}
                    value={createIncidentDescription}
                    onChange={(event) => {
                      onCreateIncidentDescriptionChange(event.target.value);
                    }}
                  />
                </label>
                <label
                  htmlFor="incident-create-severity"
                  style={labelBlockStyle}
                >
                  Severity
                  <input
                    data-testid={incidentLandingTestId("create-severity")}
                    id="incident-create-severity"
                    style={inputStyle}
                    value={createIncidentSeverity}
                    onChange={(event) => {
                      onCreateIncidentSeverityChange(event.target.value);
                    }}
                  />
                </label>
                <label htmlFor="incident-create-tlp" style={labelBlockStyle}>
                  TLP
                  <select
                    data-testid={incidentLandingTestId("create-tlp")}
                    id="incident-create-tlp"
                    style={inputStyle}
                    value={createIncidentTLP}
                    onChange={(event) => {
                      onCreateIncidentTLPChange(event.target.value);
                    }}
                  >
                    <option value="">Unset</option>
                    <option value="TLP:CLEAR">Clear</option>
                    <option value="TLP:GREEN">Green</option>
                    <option value="TLP:AMBER">Amber</option>
                    <option value="TLP:AMBER+STRICT">Amber strict</option>
                    <option value="TLP:RED">Red</option>
                  </select>
                </label>
                <label
                  htmlFor="incident-create-current-phase"
                  style={labelBlockStyle}
                >
                  Current phase
                  <input
                    data-testid={incidentLandingTestId("create-current-phase")}
                    id="incident-create-current-phase"
                    style={inputStyle}
                    value={createIncidentCurrentPhase}
                    onChange={(event) => {
                      onCreateIncidentCurrentPhaseChange(event.target.value);
                    }}
                  />
                </label>
                <label
                  htmlFor="incident-create-external-case"
                  style={labelBlockStyle}
                >
                  External case
                  <input
                    data-testid={incidentLandingTestId("create-external-case")}
                    id="incident-create-external-case"
                    style={inputStyle}
                    value={createIncidentExternalCase}
                    onChange={(event) => {
                      onCreateIncidentExternalCaseChange(event.target.value);
                    }}
                  />
                </label>
              </div>
            </details>
            <button
              data-testid={incidentLandingTestId("create-submit-button")}
              style={primaryButtonStyle}
              type="button"
              onClick={() => {
                void submitCreate();
              }}
            >
              Create and open
            </button>
          </section>
        </div>
      ) : null}

      <p
        aria-live="polite"
        data-testid={incidentLandingTestId("status")}
        role="status"
        style={statusTextStyle}
      >
        {statusText}
      </p>
      <p
        aria-live="assertive"
        data-testid={publicErrorCodeTestId("landing")}
        role={error === null ? undefined : "alert"}
        style={errorTextStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
      <PublicErrorSummary
        error={error}
        testIds={publicErrorSummaryTestIds("landing")}
      />
    </section>
  );
}
