import {
  type ChangeEvent,
  type FormEvent,
  type KeyboardEvent,
  startTransition,
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";

import { Phase1Harness } from "./Phase1Harness";
import { Phase2Harness } from "./Phase2Harness";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const csrfCookieName = "cartulary_csrf";
const csrfHeaderName = "X-CSRF-Token";

type SaveState = "Syncing" | "Saved" | "Conflict";
type EditableField =
  | "timeline.occurred_at"
  | "timeline.summary"
  | "timeline.details"
  | "timeline.source_text";

type RowValues = {
  occurredAt: string;
  summary: string;
  details: string;
  sourceText: string;
};

type WorkbookRow = {
  key: string;
  recordId: string | null;
  rowVersion: number | null;
  captureState: string;
  values: RowValues;
  committedValues: RowValues;
  pendingSignature: string | null;
};

type TimelineWorkbookProps = {
  incidentId: string;
  apiBase?: string;
};

type TimelineQueryEnvelope = {
  data: {
    rows: TimelineApiRow[];
  };
};

type TimelineMutationEnvelope = {
  data: {
    row: TimelineApiRow;
  };
};

type TimelineActionEnvelope = {
  data: {
    record_id: string;
    incident_id: string;
    row_version: number;
    capture_state: string;
    change_set_id: string;
    reason: string;
    replacement_record_id: string | null;
  };
};

type TimelineApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
};

type CollaborationMessage = {
  type: string;
};

type LoadRowsOptions = {
  showLoading: boolean;
};

const fieldOrder: Array<{
  fieldKey: EditableField;
  key: keyof RowValues;
  label: string;
  multiline?: boolean;
}> = [
  {
    fieldKey: "timeline.occurred_at",
    key: "occurredAt",
    label: "Time (RFC3339)",
  },
  {
    fieldKey: "timeline.summary",
    key: "summary",
    label: "Summary",
  },
  {
    fieldKey: "timeline.details",
    key: "details",
    label: "Details",
    multiline: true,
  },
  {
    fieldKey: "timeline.source_text",
    key: "sourceText",
    label: "Source Text",
    multiline: true,
  },
];

function emptyValues(): RowValues {
  return {
    occurredAt: "",
    summary: "",
    details: "",
    sourceText: "",
  };
}

export function createDraftRow(index: number): WorkbookRow {
  return {
    key: `draft-${index}`,
    recordId: null,
    rowVersion: null,
    captureState: "rough",
    values: emptyValues(),
    committedValues: emptyValues(),
    pendingSignature: null,
  };
}

function normalizeValue(value: string): string {
  return value.trim();
}

function valueFromCell(
  row: TimelineApiRow,
  fieldKey: EditableField | "timeline.capture_state",
): string {
  const raw = row.cells[fieldKey]?.value;
  return typeof raw === "string" ? raw : "";
}

function rowFromApi(row: TimelineApiRow): WorkbookRow {
  const values: RowValues = {
    occurredAt: valueFromCell(row, "timeline.occurred_at"),
    summary: valueFromCell(row, "timeline.summary"),
    details: valueFromCell(row, "timeline.details"),
    sourceText: valueFromCell(row, "timeline.source_text"),
  };

  return {
    key: row.record_id,
    recordId: row.record_id,
    rowVersion: row.row_version,
    captureState: valueFromCell(row, "timeline.capture_state"),
    values,
    committedValues: values,
    pendingSignature: null,
  };
}

export function ensureDraftRow(
  rows: WorkbookRow[],
  nextDraftIndex: number,
): WorkbookRow[] {
  if (rows.some((row) => row.recordId === null)) {
    return rows;
  }
  return [...rows, createDraftRow(nextDraftIndex)];
}

function ensureDraftRowWithFreshIndex(
  rows: WorkbookRow[],
  nextDraftIndex: () => number,
): {
  rows: WorkbookRow[];
  draftSummaryKey: string | null;
} {
  if (rows.some((row) => row.recordId === null)) {
    return {
      rows,
      draftSummaryKey: null,
    };
  }

  const draftIndex = nextDraftIndex();
  return {
    rows: [...rows, createDraftRow(draftIndex)],
    draftSummaryKey: `draft-${draftIndex}:summary`,
  };
}

export function buildCreatePayload(row: WorkbookRow, clientTxnId: string) {
  const payload: Record<string, string> = {
    client_txn_id: clientTxnId,
  };

  for (const field of fieldOrder) {
    const normalized = normalizeValue(row.values[field.key]);
    if (normalized !== "") {
      payload[field.fieldKey] = normalized;
    }
  }

  const populatedFieldCount = Object.keys(payload).length - 1;
  if (populatedFieldCount < 1) {
    return null;
  }
  return payload;
}

function buildPatchPayload(row: WorkbookRow, clientTxnId: string) {
  const changes = fieldOrder
    .map((field) => {
      const current = normalizeValue(row.values[field.key]);
      const committed = normalizeValue(row.committedValues[field.key]);
      if (current === committed) {
        return null;
      }
      return {
        field_key: field.fieldKey,
        value: current === "" ? null : current,
      };
    })
    .filter(
      (change): change is { field_key: EditableField; value: string | null } =>
        change !== null,
    )
    .sort((left, right) => left.field_key.localeCompare(right.field_key));

  if (changes.length < 1 || row.rowVersion === null) {
    return null;
  }

  return {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.rowVersion,
    client_txn_id: clientTxnId,
    changes,
  };
}

function buildMutationSignature(payload: unknown): string {
  return JSON.stringify(payload);
}

function readEnvelope<T>(payload: unknown): T {
  return payload as T;
}

function readCookie(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }

  const prefix = `${name}=`;
  for (const segment of document.cookie.split(";")) {
    const trimmed = segment.trim();
    if (trimmed.startsWith(prefix)) {
      return decodeURIComponent(trimmed.slice(prefix.length));
    }
  }
  return null;
}

async function fetchJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<{
  ok: boolean;
  status: number;
  payload: T | { error?: { code?: string } };
}> {
  const method = (init?.method ?? "GET").toUpperCase();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init?.headers as Record<string, string> | undefined),
  };
  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    const csrfToken = readCookie(csrfCookieName);
    if (csrfToken !== null && csrfToken !== "") {
      headers[csrfHeaderName] = csrfToken;
    }
  }

  const response = await fetch(input, {
    credentials: "include",
    ...init,
    headers,
  });
  const payload = (await response.json()) as T | { error?: { code?: string } };
  return { ok: response.ok, status: response.status, payload };
}

function apiPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    return path;
  }
  return `${trimmedBase.replace(/\/$/, "")}${path}`;
}

function websocketPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${window.location.host}${path}`;
  }

  const target = new URL(trimmedBase);
  target.protocol = target.protocol === "https:" ? "wss:" : "ws:";
  target.pathname = path;
  target.search = "";
  target.hash = "";
  return target.toString();
}

export function TimelineWorkbook({
  incidentId,
  apiBase,
}: TimelineWorkbookProps) {
  const [rows, setRows] = useState<WorkbookRow[]>(() => [createDraftRow(1)]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [saveState, setSaveState] = useState<SaveState>("Saved");
  const [pendingFocusKey, setPendingFocusKey] = useState<string | null>(null);
  const [replacementDrafts, setReplacementDrafts] = useState<
    Record<string, string>
  >({});
  const draftCounterRef = useRef(2);
  const clientTxnRef = useRef(1);
  const pendingOpsRef = useRef(0);
  const pendingSignaturesRef = useRef(new Map<string, string>());
  const suppressedSocketMessagesRef = useRef(0);
  const saveQueueRef = useRef(Promise.resolve());
  const rowsRef = useRef(rows);
  const loadSequenceRef = useRef(0);
  const loadRowsRef = useRef<(options: LoadRowsOptions) => Promise<void>>(
    async () => undefined,
  );
  const rowInputRefs = useRef(
    new Map<string, HTMLInputElement | HTMLTextAreaElement>(),
  );

  const queryPath = useMemo(
    () =>
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
      ),
    [apiBase, incidentId],
  );
  const changeSocketURL = useMemo(
    () =>
      websocketPath(
        apiBase,
        `/ws/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/changes`,
      ),
    [apiBase, incidentId],
  );
  const nextDraftIndex = useCallback(() => {
    const value = draftCounterRef.current;
    draftCounterRef.current += 1;
    return value;
  }, []);

  const loadRows = useCallback(
    async (options: LoadRowsOptions) => {
      const requestSequence = loadSequenceRef.current + 1;
      loadSequenceRef.current = requestSequence;

      if (options.showLoading) {
        setIsLoading(true);
      }
      setLoadError(null);

      const result = await fetchJSON<TimelineQueryEnvelope>(queryPath, {
        method: "POST",
        body: JSON.stringify({}),
      });

      if (requestSequence !== loadSequenceRef.current) {
        return;
      }

      if (!result.ok) {
        setLoadError("Timeline projection load failed.");
        setIsLoading(false);
        return;
      }

      const envelope = readEnvelope<TimelineQueryEnvelope>(result.payload);
      const projectedRows = envelope.data.rows.map(rowFromApi);
      const hydratedRows = ensureDraftRowWithFreshIndex(
        projectedRows,
        nextDraftIndex,
      ).rows;
      startTransition(() => {
        rowsRef.current = hydratedRows;
        setRows(hydratedRows);
      });
      setSaveState("Saved");
      setIsLoading(false);
    },
    [queryPath, nextDraftIndex],
  );

  loadRowsRef.current = loadRows;

  useEffect(() => {
    void loadRows({ showLoading: true });
  }, [loadRows]);

  useEffect(() => {
    if (pendingFocusKey === null) {
      return;
    }

    let cancelled = false;
    const focusTarget = (attempt: number) => {
      if (cancelled) {
        return;
      }

      const element = rowInputRefs.current.get(pendingFocusKey);
      if (element) {
        element.focus();
        if (document.activeElement === element) {
          setPendingFocusKey((current) =>
            current === pendingFocusKey ? null : current,
          );
          return;
        }
      }

      if (attempt < 60) {
        window.setTimeout(() => focusTarget(attempt + 1), 50);
      }
    };

    focusTarget(0);
    return () => {
      cancelled = true;
    };
  }, [pendingFocusKey]);

  useEffect(() => {
    if (incidentId.trim() === "") {
      return;
    }

    const socket = new WebSocket(changeSocketURL);
    socket.onmessage = (event) => {
      const message = JSON.parse(event.data) as CollaborationMessage;
      if (message.type === "record_changed") {
        if (suppressedSocketMessagesRef.current > 0) {
          suppressedSocketMessagesRef.current -= 1;
          return;
        }
        void loadRowsRef.current({ showLoading: false });
      }
    };
    return () => {
      socket.close();
    };
  }, [changeSocketURL, incidentId]);

  function nextClientTxnId() {
    const value = clientTxnRef.current;
    clientTxnRef.current += 1;
    return `timeline-client-${value}`;
  }

  function beginSave() {
    pendingOpsRef.current += 1;
    setSaveState("Syncing");
  }

  function finishSave(nextState: SaveState) {
    pendingOpsRef.current = Math.max(0, pendingOpsRef.current - 1);
    if (pendingOpsRef.current > 0 && nextState !== "Conflict") {
      setSaveState("Syncing");
      return;
    }
    setSaveState(nextState);
  }

  function setRowValue(rowKey: string, field: keyof RowValues, value: string) {
    setRows((current) => {
      const nextRows = current.map((row) =>
        row.key === rowKey
          ? {
              ...row,
              values: {
                ...row.values,
                [field]: value,
              },
            }
          : row,
      );
      rowsRef.current = nextRows;
      return nextRows;
    });
  }

  function registerInput(
    rowKey: string,
    field: keyof RowValues,
    element: HTMLInputElement | HTMLTextAreaElement | null,
  ) {
    const key = `${rowKey}:${field}`;
    if (element === null) {
      rowInputRefs.current.delete(key);
      return;
    }
    rowInputRefs.current.set(key, element);
  }

  function queueRowSave(rowKey: string, focusContinuation: boolean) {
    const snapshot = rowsRef.current.find(
      (candidate) => candidate.key === rowKey,
    );
    if (!snapshot) {
      return;
    }

    const clientTxnId = nextClientTxnId();
    const payload =
      snapshot.recordId === null
        ? buildCreatePayload(snapshot, clientTxnId)
        : buildPatchPayload(snapshot, clientTxnId);
    if (payload === null) {
      return;
    }

    const mutationSignature = buildMutationSignature(payload);
    if (pendingSignaturesRef.current.get(rowKey) === mutationSignature) {
      return;
    }
    pendingSignaturesRef.current.set(rowKey, mutationSignature);

    setRows((current) => {
      const nextRows = current.map((row) =>
        row.key === rowKey
          ? { ...row, pendingSignature: mutationSignature }
          : row,
      );
      rowsRef.current = nextRows;
      return nextRows;
    });
    beginSave();

    saveQueueRef.current = saveQueueRef.current
      .catch(() => undefined)
      .then(async () => {
        const targetPath =
          snapshot.recordId === null
            ? apiPath(
                apiBase,
                `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
              )
            : apiPath(apiBase, `/api/v1/records/${snapshot.recordId}`);
        const method = snapshot.recordId === null ? "POST" : "PATCH";

        const result = await fetchJSON<TimelineMutationEnvelope>(targetPath, {
          method,
          body: JSON.stringify(payload),
        });

        if (!result.ok) {
          pendingSignaturesRef.current.delete(rowKey);
          setRows((current) => {
            const nextRows = current.map((row) =>
              row.key === rowKey ? { ...row, pendingSignature: null } : row,
            );
            rowsRef.current = nextRows;
            return nextRows;
          });
          finishSave("Conflict");
          return;
        }

        const envelope = readEnvelope<TimelineMutationEnvelope>(result.payload);
        const committed = rowFromApi(envelope.data.row);
        suppressedSocketMessagesRef.current += 1;
        pendingSignaturesRef.current.delete(rowKey);
        const nextRows = rowsRef.current.map((row) =>
          row.key === rowKey ? committed : row,
        );
        const hydrated = ensureDraftRowWithFreshIndex(nextRows, nextDraftIndex);
        const hydratedRows = hydrated.rows;
        const continuationDraftSummaryKey =
          focusContinuation && snapshot.recordId === null
            ? hydrated.draftSummaryKey
            : null;

        startTransition(() => {
          rowsRef.current = hydratedRows;
          setRows(hydratedRows);
        });
        finishSave("Saved");
        if (continuationDraftSummaryKey !== null) {
          setPendingFocusKey(continuationDraftSummaryKey);
        }
      });
  }

  function queueAction(rowKey: string, action: "mark-reviewed" | "supersede") {
    const snapshot = rowsRef.current.find(
      (candidate) => candidate.key === rowKey,
    );
    if (
      !snapshot ||
      snapshot.recordId === null ||
      snapshot.rowVersion === null ||
      (action === "supersede" &&
        normalizeValue(replacementDrafts[rowKey] ?? "") === "")
    ) {
      return;
    }

    beginSave();
    saveQueueRef.current = saveQueueRef.current
      .catch(() => undefined)
      .then(async () => {
        const body =
          action === "mark-reviewed"
            ? {
                base_row_version: snapshot.rowVersion,
                client_txn_id: nextClientTxnId(),
                reason: "Reviewed from workbook",
              }
            : {
                base_row_version: snapshot.rowVersion,
                client_txn_id: nextClientTxnId(),
                reason: "Superseded from workbook",
                replacement_record_id: normalizeValue(
                  replacementDrafts[rowKey] ?? "",
                ),
              };
        const result = await fetchJSON<TimelineActionEnvelope>(
          apiPath(apiBase, `/api/v1/records/${snapshot.recordId}/${action}`),
          {
            method: "POST",
            body: JSON.stringify(body),
          },
        );
        if (!result.ok) {
          finishSave("Conflict");
          return;
        }

        suppressedSocketMessagesRef.current += 1;
        await loadRowsRef.current({ showLoading: false });
        finishSave("Saved");
      });
  }

  function handleBlur(rowKey: string) {
    queueRowSave(rowKey, false);
  }

  function handleKeyDown(
    event: KeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
    rowKey: string,
  ) {
    if (event.key === "Enter" || event.key === "Tab") {
      event.preventDefault();
      queueRowSave(rowKey, true);
    }
  }

  function handlePaste(rowKey: string) {
    window.setTimeout(() => {
      queueRowSave(rowKey, false);
    }, 0);
  }

  if (isLoading) {
    return (
      <section style={panelStyle}>
        <p style={eyebrowStyle}>Timeline</p>
        <h1 style={headlineStyle}>Loading projection-backed rows.</h1>
      </section>
    );
  }

  if (loadError !== null) {
    return (
      <section style={panelStyle}>
        <p style={eyebrowStyle}>Timeline</p>
        <h1 style={headlineStyle}>Timeline load failed.</h1>
        <p style={bodyStyle}>{loadError}</p>
      </section>
    );
  }

  return (
    <section style={workbookStyle}>
      <header style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Phase 3 Workbook</p>
          <h1 style={headlineStyle}>Timeline mutation substrate</h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
        <div style={statusClusterStyle}>
          <span style={statusLabelStyle}>Save State</span>
          <strong
            aria-live="polite"
            data-testid="save-state"
            style={{
              ...statusValueStyle,
              color:
                saveState === "Conflict"
                  ? "rgb(145 30 30)"
                  : saveState === "Syncing"
                    ? "rgb(146 64 14)"
                    : "rgb(21 128 61)",
            }}
          >
            {saveState}
          </strong>
        </div>
      </header>

      <div style={gridShellStyle}>
        <table style={tableStyle}>
          <thead>
            <tr>
              <th style={headCellStyle}>State</th>
              <th style={headCellStyle}>Version</th>
              {fieldOrder.map((field) => (
                <th key={field.fieldKey} style={headCellStyle}>
                  {field.label}
                </th>
              ))}
              <th style={headCellStyle}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr
                key={row.key}
                style={row.recordId === null ? draftRowStyle : undefined}
              >
                <td
                  data-testid={
                    row.recordId === null
                      ? "draft-row-capture-state"
                      : `row-${row.recordId}-capture-state`
                  }
                  style={metaCellStyle}
                >
                  {row.captureState}
                </td>
                <td
                  data-testid={
                    row.recordId === null
                      ? "draft-row-version"
                      : `row-${row.recordId}-row-version`
                  }
                  style={metaCellStyle}
                >
                  {row.rowVersion ?? "new"}
                </td>
                {fieldOrder.map((field) => (
                  <td key={field.fieldKey} style={bodyCellStyle}>
                    {field.multiline ? (
                      <textarea
                        aria-label={`${field.label} ${row.recordId ?? "draft row"}`}
                        data-testid={
                          row.recordId === null
                            ? `draft-row-${field.key}`
                            : `row-${row.recordId}-${field.key}`
                        }
                        ref={(element) => {
                          registerInput(row.key, field.key, element);
                        }}
                        rows={3}
                        style={textareaStyle}
                        value={row.values[field.key]}
                        onBlur={() => {
                          handleBlur(row.key);
                        }}
                        onChange={(event: ChangeEvent<HTMLTextAreaElement>) => {
                          setRowValue(row.key, field.key, event.target.value);
                        }}
                        onInput={(event: FormEvent<HTMLTextAreaElement>) => {
                          setRowValue(
                            row.key,
                            field.key,
                            event.currentTarget.value,
                          );
                        }}
                        onKeyDown={(event) => {
                          handleKeyDown(event, row.key);
                        }}
                        onPaste={() => {
                          handlePaste(row.key);
                        }}
                      />
                    ) : (
                      <input
                        {...(row.recordId === null && field.key === "summary"
                          ? { autoFocus: true }
                          : {})}
                        aria-label={`${field.label} ${row.recordId ?? "draft row"}`}
                        data-testid={
                          row.recordId === null
                            ? `draft-row-${field.key}`
                            : `row-${row.recordId}-${field.key}`
                        }
                        ref={(element) => {
                          registerInput(row.key, field.key, element);
                        }}
                        style={inputStyle}
                        type="text"
                        value={row.values[field.key]}
                        onBlur={() => {
                          handleBlur(row.key);
                        }}
                        onChange={(event: ChangeEvent<HTMLInputElement>) => {
                          setRowValue(row.key, field.key, event.target.value);
                        }}
                        onInput={(event: FormEvent<HTMLInputElement>) => {
                          setRowValue(
                            row.key,
                            field.key,
                            event.currentTarget.value,
                          );
                        }}
                        onKeyDown={(event) => {
                          handleKeyDown(event, row.key);
                        }}
                        onPaste={() => {
                          handlePaste(row.key);
                        }}
                      />
                    )}
                  </td>
                ))}
                <td style={actionCellStyle}>
                  {row.recordId === null ? (
                    <span style={bodyStyle}>Draft row</span>
                  ) : (
                    <div style={actionStackStyle}>
                      <button
                        data-testid={`row-${row.recordId}-mark-reviewed`}
                        disabled={
                          row.captureState === "reviewed" ||
                          row.captureState === "superseded"
                        }
                        style={actionButtonStyle}
                        type="button"
                        onClick={() => {
                          queueAction(row.key, "mark-reviewed");
                        }}
                      >
                        Mark reviewed
                      </button>
                      <input
                        data-testid={`row-${row.recordId}-replacement-id`}
                        placeholder="Replacement record id"
                        style={replacementInputStyle}
                        type="text"
                        value={replacementDrafts[row.key] ?? ""}
                        onChange={(event) => {
                          const value = event.target.value;
                          setReplacementDrafts((current) => ({
                            ...current,
                            [row.key]: value,
                          }));
                        }}
                      />
                      <button
                        data-testid={`row-${row.recordId}-supersede`}
                        disabled={
                          row.captureState === "superseded" ||
                          normalizeValue(replacementDrafts[row.key] ?? "") ===
                            ""
                        }
                        style={actionButtonStyle}
                        type="button"
                        onClick={() => {
                          queueAction(row.key, "supersede");
                        }}
                      >
                        Supersede
                      </button>
                    </div>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

export function App() {
  const incidentFieldId = useId();
  const params = new URLSearchParams(window.location.search);
  const initialIncidentId = params.get("incident_id") ?? "";
  const [incidentDraft, setIncidentDraft] = useState(initialIncidentId);
  const [activeIncidentId, setActiveIncidentId] = useState(initialIncidentId);

  if (activeIncidentId === "") {
    return (
      <main style={pageStyle}>
        <section style={panelStyle}>
          <div style={heroStyle}>
            <p style={eyebrowStyle}>Cartulary</p>
            <h1 style={headlineStyle}>Phase 1 and Phase 2 harness shell</h1>
            <p style={bodyStyle}>
              Authentication, session lifecycle, deployment-local user
              administration, incident control-envelope flows, and
              extension-dispatch probes are available below. Add an
              `incident_id` query parameter to open the Timeline workbook
              directly.
            </p>
          </div>

          <Phase1Harness />
          <Phase2Harness />
        </section>
      </main>
    );
  }

  return (
    <main style={pageStyle}>
      <section style={panelStyle}>
        <div style={heroStyle}>
          <p style={eyebrowStyle}>Cartulary</p>
          <h1 style={headlineStyle}>Timeline workbook shell</h1>
          <p style={bodyStyle}>
            Phase 3 is wired to the Timeline projection, record mutation routes,
            reviewer actions, and collaboration updates. Load an incident UUID
            and work directly in the workbook grid.
          </p>
        </div>

        <label htmlFor={incidentFieldId} style={labelStyle}>
          Incident ID
        </label>
        <div style={launcherStyle}>
          <input
            id={incidentFieldId}
            style={launcherInputStyle}
            value={incidentDraft}
            onChange={(event) => {
              setIncidentDraft(event.target.value);
            }}
            placeholder="00000000-0000-0000-0000-000000000000"
          />
          <button
            style={launcherButtonStyle}
            type="button"
            onClick={() => {
              setActiveIncidentId(incidentDraft.trim());
            }}
          >
            Open Timeline
          </button>
        </div>

        <TimelineWorkbook incidentId={activeIncidentId} />
      </section>
    </main>
  );
}

const pageStyle = {
  minHeight: "100vh",
  margin: 0,
  padding: "2rem",
  background:
    "radial-gradient(circle at top left, rgb(244 229 213), rgb(245 246 238) 40%, rgb(223 232 226) 100%)",
  color: "rgb(23 37 34)",
  fontFamily: '"IBM Plex Sans", "Segoe UI", sans-serif',
};

const panelStyle = {
  width: "min(82rem, 100%)",
  margin: "0 auto",
  padding: "2rem",
  borderRadius: "1.5rem",
  background: "rgb(255 252 247 / 0.92)",
  boxShadow: "0 24px 80px rgb(29 78 70 / 0.12)",
  border: "1px solid rgb(185 204 196 / 0.8)",
};

const heroStyle = {
  marginBottom: "1.5rem",
};

const workbookStyle = {
  marginTop: "1.5rem",
};

const headerStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "flex-start",
  marginBottom: "1rem",
};

const statusClusterStyle = {
  display: "grid",
  gap: "0.25rem",
  minWidth: "8rem",
  textAlign: "right" as const,
};

const statusLabelStyle = {
  fontSize: "0.75rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  color: "rgb(81 110 103)",
};

const statusValueStyle = {
  fontSize: "1rem",
};

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  color: "rgb(94 125 113)",
};

const headlineStyle = {
  margin: "0.4rem 0",
  fontSize: "clamp(1.8rem, 4vw, 3rem)",
  lineHeight: 1.05,
};

const bodyStyle = {
  margin: 0,
  lineHeight: 1.55,
  color: "rgb(54 74 68)",
};

const labelStyle = {
  display: "block",
  marginBottom: "0.5rem",
  fontSize: "0.9rem",
  fontWeight: 600,
};

const launcherStyle = {
  display: "flex",
  gap: "0.75rem",
  marginBottom: "1.5rem",
  flexWrap: "wrap" as const,
};

const launcherInputStyle = {
  flex: "1 1 24rem",
  minWidth: "18rem",
  padding: "0.85rem 1rem",
  borderRadius: "999px",
  border: "1px solid rgb(168 188 181)",
  background: "rgb(255 255 255)",
  fontSize: "0.95rem",
};

const launcherButtonStyle = {
  padding: "0.85rem 1.25rem",
  borderRadius: "999px",
  border: "none",
  background: "rgb(29 78 70)",
  color: "rgb(255 252 247)",
  fontWeight: 700,
  cursor: "pointer",
};

const gridShellStyle = {
  overflowX: "auto" as const,
  borderRadius: "1rem",
  border: "1px solid rgb(190 209 201)",
  background: "rgb(255 255 255 / 0.85)",
};

const tableStyle = {
  width: "100%",
  borderCollapse: "collapse" as const,
  minWidth: "72rem",
};

const headCellStyle = {
  padding: "0.85rem 0.9rem",
  textAlign: "left" as const,
  fontSize: "0.78rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  color: "rgb(78 102 96)",
  background: "rgb(240 245 240)",
  borderBottom: "1px solid rgb(190 209 201)",
};

const bodyCellStyle = {
  padding: "0.75rem 0.9rem",
  borderBottom: "1px solid rgb(226 234 230)",
  verticalAlign: "top" as const,
};

const metaCellStyle = {
  ...bodyCellStyle,
  whiteSpace: "nowrap" as const,
  color: "rgb(81 110 103)",
  fontSize: "0.9rem",
};

const actionCellStyle = {
  ...bodyCellStyle,
  width: "16rem",
};

const rowStyle = {
  background: "rgb(255 255 255 / 0.8)",
};

const draftRowStyle = {
  ...rowStyle,
  background: "rgb(250 244 232 / 0.75)",
};

const inputStyle = {
  width: "100%",
  border: "1px solid rgb(200 214 208)",
  borderRadius: "0.8rem",
  padding: "0.7rem 0.8rem",
  background: "rgb(255 255 255)",
  font: "inherit",
};

const textareaStyle = {
  ...inputStyle,
  resize: "vertical" as const,
  minHeight: "5.5rem",
};

const actionStackStyle = {
  display: "grid",
  gap: "0.5rem",
};

const actionButtonStyle = {
  border: "none",
  borderRadius: "999px",
  padding: "0.7rem 0.9rem",
  background: "rgb(29 78 70)",
  color: "rgb(255 252 247)",
  fontWeight: 700,
  cursor: "pointer",
};

const replacementInputStyle = {
  ...inputStyle,
  minWidth: "14rem",
};
