import {
  incidentLandingTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  publicErrorCodeTestId,
  publicErrorSummaryTestIds,
} from "@cartulary/ui-contracts";
import { ChevronRight, RefreshCw, Search } from "lucide-react";
import { useRef } from "react";
import { publicErrorView } from "../services/browserApi";
import { IncidentCreationForm } from "./IncidentCreationForm";
import {
  formatNullableDateTime,
  MutedUnset,
  PublicErrorSummary,
  StatusBadge,
} from "./LandingAdminDisplay";
import {
  countPillStyle,
  dataTableStyle,
  emptyStateStyle,
  errorTextStyle,
  headerActionRowStyle,
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
  surfaceHeaderStyle,
  surfacePanelStyle,
  tableActionCellStyle,
  tableCellStyle,
  tableHeaderActionCellStyle,
  tableHeaderCellStyle,
  tableLinkButtonStyle,
  tableRowStyle,
  tableShellStyle,
  toolbarGridStyle,
} from "./landingAdminStyles";
import type {
  IncidentLandingProps,
  IncidentStatusFilter,
} from "./landingAdminTypes";

export function IncidentLanding({
  bootstrapState,
  creation,
  error,
  hasMoreIncidents,
  incidents,
  incidentSearch,
  incidentStatusFilter,
  isRefreshing,
  onLoadMore,
  onOpenIncident,
  onRefresh,
  onSearchChange,
  onSearchSubmit,
  onStatusFilterChange,
  statusText,
}: IncidentLandingProps) {
  const createTriggerRef = useRef<HTMLButtonElement>(null);
  const headingRef = useRef<HTMLHeadingElement>(null);
  const hasIncidents = incidents.length > 0;
  const hasActiveQuery =
    incidentSearch.trim() !== "" || incidentStatusFilter !== "all";

  return (
    <section
      data-bootstrap-state={bootstrapState}
      data-testid={incidentLandingTestId("shell")}
      style={surfacePanelStyle}
    >
      <header style={surfaceHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Incidents</p>
          <h2 ref={headingRef} tabIndex={-1} style={sectionTitleStyle}>
            Visible incidents
          </h2>
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
          {!creation.state.open ? (
            <button
              ref={createTriggerRef}
              data-testid={incidentLandingTestId("create-open-button")}
              style={primaryButtonStyle}
              type="button"
              onClick={() => {
                creation.controller.open();
              }}
            >
              {creation.state.operation.kind === "editing"
                ? "New incident"
                : "Resume incident creation"}
            </button>
          ) : null}
        </div>
      </header>

      <IncidentCreationForm
        creation={creation}
        triggerRef={createTriggerRef}
        headingRef={headingRef}
      />

      <div style={incidentWorkspaceStyle}>
        <section aria-busy={isRefreshing} style={incidentDirectoryStyle}>
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
