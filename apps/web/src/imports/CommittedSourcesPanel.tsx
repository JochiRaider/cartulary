import { useEffect, useRef, useState } from "react";
import { boundedRead } from "../services/asyncObservation";
import {
  type ImportClient,
  importFailureMessage,
} from "../services/importClient";
import type {
  CommittedImportSource,
  CommittedImportSourcesPage,
} from "../services/importContractAdapter";
import {
  workbookFormButtonStyle,
  workbookFormFieldsStyle,
  workbookFormHeadingStyle,
  workbookFormMessageStyle,
  workbookReferenceIdentityStyle,
  workbookTypography,
} from "../workbook/components/workbookFormStyles";

type State =
  | { readonly kind: "loading" }
  | { readonly kind: "failed"; readonly message: string }
  | { readonly kind: "list"; readonly page: CommittedImportSourcesPage }
  | { readonly kind: "detail"; readonly source: CommittedImportSource };

/** Read-only presentation: closing it aborts observations and drops protected data. */
export function CommittedSourcesPanel({
  client,
  initialSourceId,
  recordId,
}: {
  readonly client: ImportClient;
  readonly initialSourceId: string | null;
  readonly recordId?: string | null;
}) {
  const [selection, setSelection] = useState<{
    sourceId: string | null;
    cursor?: string;
    revision: number;
  }>({ sourceId: initialSourceId, revision: 0 });
  const [state, setState] = useState<State>({ kind: "loading" });
  const heading = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    const abort = new AbortController();
    setState({ kind: "loading" });
    void boundedRead(async (signal) => {
      if (selection.sourceId !== null) {
        const result = await client.readCommittedSource(
          selection.sourceId,
          signal,
        );
        return result.kind === "received"
          ? ({ kind: "detail", source: result.value } as const)
          : ({
              kind: "failed",
              message: importFailureMessage(result.failure),
            } as const);
      }
      const result = await client.listCommittedSources(
        selection.cursor,
        signal,
        recordId ?? undefined,
      );
      return result.kind === "received"
        ? ({ kind: "list", page: result.value } as const)
        : ({
            kind: "failed",
            message: importFailureMessage(result.failure),
          } as const);
    }, abort.signal).then(
      (value) => {
        if (!abort.signal.aborted) setState(value);
      },
      () => {
        if (!abort.signal.aborted)
          setState({
            kind: "failed",
            message: "Source details could not be loaded. Try again.",
          });
      },
    );
    heading.current?.focus({ preventScroll: true });
    return () => abort.abort();
  }, [client, selection, recordId]);
  const choose = (sourceId: string | null, cursor?: string) => {
    setState({ kind: "loading" });
    setSelection((previous) => ({
      sourceId,
      ...(cursor === undefined ? {} : { cursor }),
      revision: previous.revision + 1,
    }));
  };
  return (
    <section
      aria-labelledby="committed-sources-heading"
      style={workbookFormFieldsStyle}
    >
      <h2
        id="committed-sources-heading"
        ref={heading}
        tabIndex={-1}
        style={workbookFormHeadingStyle}
      >
        Source history
      </h2>
      <p style={workbookFormMessageStyle}>
        Retained source files and the mappings used when they were applied.
        Current records may have been edited since import.
      </p>
      {recordId ? <p>Showing sources that produced this record.</p> : null}
      {selection.sourceId !== null || selection.cursor !== undefined ? (
        <button
          type="button"
          style={{ ...workbookFormButtonStyle, justifySelf: "start" }}
          onClick={() => choose(null)}
        >
          Back to sources
        </button>
      ) : null}
      {state.kind === "loading" ? <p role="status">Loading sources…</p> : null}
      {state.kind === "failed" ? (
        <div>
          <p role="alert">{state.message}</p>
          <button
            type="button"
            style={workbookFormButtonStyle}
            onClick={() => choose(selection.sourceId, selection.cursor)}
          >
            Retry source read
          </button>
        </div>
      ) : null}
      {state.kind === "list" ? (
        <>
          {state.page.data.committed_sources.length === 0 ? (
            <p>No committed import sources.</p>
          ) : (
            <ul
              style={{
                display: "grid",
                gap: "var(--ct-spacing-sm)",
                listStyle: "none",
                padding: 0,
                margin: 0,
              }}
            >
              {state.page.data.committed_sources.map((source) => (
                <li key={source.committed_source_id}>
                  <button
                    type="button"
                    style={workbookFormButtonStyle}
                    onClick={() => choose(source.committed_source_id)}
                  >
                    {source.original_filename}
                  </button>
                  <span style={workbookReferenceIdentityStyle}>
                    Applied {source.committed_at} · Source{" "}
                    {source.committed_source_id}
                  </span>
                </li>
              ))}
            </ul>
          )}
          {state.page.meta.paging?.next_cursor ? (
            <button
              type="button"
              style={{ ...workbookFormButtonStyle, justifySelf: "start" }}
              onClick={() =>
                choose(null, state.page.meta.paging?.next_cursor ?? undefined)
              }
            >
              Next sources
            </button>
          ) : null}
        </>
      ) : null}
      {state.kind === "detail" ? (
        <>
          <h3 style={workbookFormHeadingStyle}>
            {state.source.original_filename}
          </h3>
          <dl style={{ ...workbookFormFieldsStyle, margin: 0 }}>
            {(
              [
                ["Applied at", state.source.committed_at],
                ["Original actor ID", state.source.created_by_user_id],
                ["Source ID", state.source.committed_source_id],
                ["Size in bytes", String(state.source.byte_size)],
                ["SHA-256", state.source.content_sha256],
                ["Source parser", state.source.parser_profile_id],
                ["Source parser version", state.source.parser_version],
                [
                  "Applied target",
                  state.source.target_view_schema_id ||
                    state.source.extension_profile_id,
                ],
                ["Source locator", state.source.locator],
                ["Source rectangle", state.source.source_rect_a1],
                [
                  "Original import session",
                  state.source.origin_import_session_id,
                ],
                ["Original import unit", state.source.origin_import_unit_id],
                ["Mapping fingerprint", state.source.mapping_fingerprint],
              ] as const
            ).map(([label, value]) => (
              <div key={label}>
                <dt
                  style={{
                    ...workbookTypography("metadata"),
                    color: "var(--ct-colors-ink-muted)",
                  }}
                >
                  {label}
                </dt>
                <dd style={{ margin: 0, whiteSpace: "pre-wrap" }}>
                  {value || "Not applicable"}
                </dd>
              </div>
            ))}
          </dl>
          <details>
            <summary
              style={{ ...workbookTypography("button"), cursor: "pointer" }}
            >
              Approved mapping
            </summary>
            <pre
              style={{ ...workbookTypography("mono"), whiteSpace: "pre-wrap" }}
            >
              {JSON.stringify(state.source.approved_mapping, null, 2)}
            </pre>
          </details>
          <a
            style={{
              ...workbookFormButtonStyle,
              justifySelf: "start",
              textDecoration: "underline",
            }}
            href={client.committedSourceDownloadURL(
              state.source.committed_source_id,
            )}
            download
          >
            Download original source
          </a>
        </>
      ) : null}
    </section>
  );
}
