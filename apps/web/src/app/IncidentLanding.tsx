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
  directoryCanLoadMore,
  directoryIsLoading,
} from "./incidentDirectoryModel";
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
  error: bootstrapError,
  directory: { controller, state },
  onOpenIncident,
  statusText,
}: IncidentLandingProps) {
  const createTriggerRef = useRef<HTMLButtonElement>(null);
  const headingRef = useRef<HTMLHeadingElement>(null);
  const incidents = state.incidents;
  const incidentSearch = state.query.search;
  const incidentStatusFilter = state.query.statusFilter;
  const isRefreshing =
    directoryIsLoading(state) || state.phase === "debouncing";
  const hasMoreIncidents = state.paging?.has_more ?? false;
  const error = state.failure?.error ?? bootstrapError;
  const previousResults =
    incidents.length > 0 &&
    (state.phase === "refreshing" ||
      state.phase === "debouncing" ||
      (state.phase === "failed" && state.failure?.scope === "replace"));
  const hasIncidents = incidents.length > 0;
  const hasActiveQuery =
    incidentSearch.trim() !== "" || incidentStatusFilter !== "all";

  return (
    <section
      data-bootstrap-state={
        state.phase === "forbidden" ? "forbidden" : bootstrapState
      }
      data-directory-state={state.phase}
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
              controller.refresh();
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
                    controller.changeSearch(event.target.value);
                  }}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      event.preventDefault();
                      controller.submit();
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
                  controller.changeStatus(
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

          {previousResults ? (
            <p role="status" style={inlineStatusStyle}>
              Showing previous results while the current query is unresolved.
            </p>
          ) : null}
          {state.failure !== null ? (
            <button
              style={secondaryButtonStyle}
              type="button"
              onClick={() => controller.retry()}
            >
              {state.failure.restart
                ? "Refresh first page"
                : "Retry loading incidents"}
            </button>
          ) : null}

          {state.phase === "ready" && !hasIncidents ? (
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
                        <StatusBadge value={incident.status} />
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
                      <td style={tableCellStyle} title={incident.updated_at}>
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
              disabled={!directoryCanLoadMore(state)}
              style={secondaryButtonStyle}
              type="button"
              onClick={() => {
                controller.loadMore();
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
