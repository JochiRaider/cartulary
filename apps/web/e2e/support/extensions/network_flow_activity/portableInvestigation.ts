import type { Buffer } from "node:buffer";
import { expect, type Page } from "@playwright/test";
import { csrfHeaders } from "../../auth/browserSession";
import { authHeadersForStorageState } from "../../auth/storageState";
import { ownedJobsFor } from "../../jobs/ownership";
import { apiBase } from "../../runtime/configuration";
import { uniqueTxn } from "../../runtime/fixtureIdentity";
import { importNetworkFlowCSV, openClaimedNetworkAnalysis } from "./workspace";

type Job = {
  job_id: string;
  status: string;
  result_summary?: { resource_refs: { kind: string; id: string }[] };
};
type SelectedGraph = {
  graph_view: {
    graph_view_id: string;
    selected_result_binding: Record<string, unknown>;
    selected_source_binding: unknown;
  };
  result: {
    graph_projection_result: {
      projection_result_id: string;
      edges: {
        edge_id: string;
        src_vertex_id: string;
        dst_vertex_id: string;
      }[];
    };
  };
};
export type PortableInvestigation = {
  incidentId: string;
  graph: SelectedGraph;
  reportSHA256: string;
  bundle: Buffer;
};

// A fixed small investigation for independently initialized destination journeys.
// Corpus import and mapping have their own full-volume browser journey.
export async function createPortableInvestigation(
  page: Page,
  file: { buffer: Buffer; mimeType: string; name: string },
): Promise<PortableInvestigation> {
  const incidentId = await openClaimedNetworkAnalysis(page, "IRPORTABLE");
  await importNetworkFlowCSV(page, {
    displayName: "Historical browser source",
    file,
  });
  const headers = {
    ...authHeadersForStorageState(await page.context().storageState()),
    ...(await csrfHeaders(page)),
  };
  const get = async <T>(path: string): Promise<T> => {
    const response = await page.request.get(apiBase + path, { headers });
    expect(response.ok(), await response.text()).toBeTruthy();
    return ((await response.json()) as { data: T }).data;
  };
  const post = async <T>(path: string, body: unknown): Promise<T> => {
    const response = await ownedJobsFor(page).admit(
      () => page.request.post(apiBase + path, { headers, data: body }),
      async (response) => {
        if (response.status() !== 202) return null;
        const data = ((await response.json()) as { data: Job }).data;
        return data.job_id ? { jobId: data.job_id, headers } : null;
      },
    );
    expect(response.ok(), await response.text()).toBeTruthy();
    return ((await response.json()) as { data: T }).data;
  };
  const wait = async (jobId: string): Promise<Job> => {
    let terminal: Job | undefined;
    await expect
      .poll(
        async () => {
          terminal = await get<Job>(`/api/v1/jobs/${jobId}`);
          if (["failed", "canceled"].includes(terminal.status))
            throw new Error(`Portable investigation job ${terminal.status}`);
          return terminal.status;
        },
        { timeout: 30_000 },
      )
      .toBe("succeeded");
    if (!terminal) throw new Error("Missing completed job");
    return terminal;
  };
  const resource = async (path: string, body: unknown, kind: string) => {
    const job = await wait((await post<Job>(path, body)).job_id);
    const id = job.result_summary?.resource_refs.find(
      (ref) => ref.kind === kind,
    )?.id;
    if (!id) throw new Error(`Missing ${kind} result`);
    return id;
  };
  const root = `/api/v1/incidents/${incidentId}/network-flow`;
  const { tables } = await get<{ tables: { network_flow_table_id: string }[] }>(
    `${root}/tables`,
  );
  expect(tables).toHaveLength(1);
  const created = await post<{
    job: Job;
    graph_view: { graph_view_id: string };
  }>(`${root}/graph-views`, {
    schema_id: "cartulary.network_flow.graph_view_create_request.v5",
    refresh_scope: "all_active_tables",
    client_txn_id: uniqueTxn("portable-graph"),
    display_name: "Retained browser analysis",
    semantic_query: {
      schema_id: "cartulary.network_flow.graph_semantic_query.v3",
      selected_table_ids: tables.map((table) => table.network_flow_table_id),
      filters: [],
      time_range: { start_utc: null, end_utc: null },
      aggregation: {
        mode: "default_flow_edge_v1",
        include_example_row_refs: true,
      },
    },
  });
  await wait(created.job.job_id);
  const graph = await get<SelectedGraph>(
    `${root}/graph-views/${created.graph_view.graph_view_id}/result`,
  );
  expect(graph.result.graph_projection_result.edges).toHaveLength(1);
  const edge = graph.result.graph_projection_result.edges[0];
  if (!edge) throw new Error("Missing independent minimal edge");
  const snapshot = await resource(
    "/api/v1/snapshots",
    { client_txn_id: uniqueTxn("portable-snapshot"), incident_id: incidentId },
    "snapshot",
  );
  const compositions = `/api/v1/incidents/${incidentId}/report-compositions`;
  const composition = await post<{
    composition_id: string;
    draft_version: number;
  }>(compositions, {
    client_txn_id: uniqueTxn("portable-composition"),
    template_id: "cartulary.report.default",
    template_version: "4",
    deck_ops: [],
    authored_texts: [],
    diagram_decls: [
      {
        decl_id: "browser-retained-flow",
        replaces_decl_id: null,
        diagram_kind: "flowchart",
        diagram_source_kind: "graph",
        source_graph_view_id: graph.graph_view.graph_view_id,
        layout_mode: "auto",
        layout: null,
        label_overrides: [],
        selection_rule: {
          schema_id: "diagram_selection_rule.v1",
          rule_id: "explicit_refs",
          vertex_refs: [edge.src_vertex_id, edge.dst_vertex_id],
          edge_refs: [edge.edge_id],
          seed_vertex_refs: [],
          depth: null,
          edge_kind_filter: [],
          timeline_record_ids: [],
          overflow_policy: "fail",
        },
      },
    ],
  });
  const version = await post<{
    composition_version: number;
    composition_sha256: string;
  }>(`${compositions}/${composition.composition_id}/versions`, {
    client_txn_id: uniqueTxn("portable-freeze"),
    base_draft_version: composition.draft_version,
  });
  const releaseId = await resource(
    "/api/v1/releases",
    {
      client_txn_id: uniqueTxn("portable-release"),
      snapshot_id: snapshot,
      template_id: "cartulary.report.default",
      template_version: "4",
      output_kind: "mermaid",
      output_options: { source_only: true },
      release_scope: "internal_draft",
      redaction_profile_id: "cartulary.redaction.internal",
      redaction_profile_version: "5",
      composition_id: composition.composition_id,
      composition_version: version.composition_version,
      composition_sha256: version.composition_sha256,
      graph_projection_refs: [
        {
          source_owner_id: "network_flow_activity",
          graph_view_id: graph.graph_view.graph_view_id,
          ...graph.graph_view.selected_result_binding,
        },
      ],
    },
    "release",
  );
  const release = await get<{ release_state: string; output_sha256: string }>(
    `/api/v1/releases/${releaseId}`,
  );
  expect(release.release_state).toBe("approved");
  // A later source must not change the selected graph or frozen report.
  await importNetworkFlowCSV(page, {
    displayName: "Later browser source",
    file,
  });
  expect(
    (
      await get<SelectedGraph>(
        `${root}/graph-views/${graph.graph_view.graph_view_id}/result`,
      )
    ).result,
  ).toEqual(graph.result);
  const exported = await resource(
    "/api/v1/incident-bundles/export",
    {
      incident_id: incidentId,
      client_txn_id: uniqueTxn("portable-export"),
      optional_sections: ["snapshots"],
    },
    "incident_bundle",
  );
  const response = await page.request.get(
    `${apiBase}/api/v1/incident-bundles/${exported}/content`,
    { headers },
  );
  expect(response.ok(), await response.text()).toBeTruthy();
  return {
    incidentId,
    graph,
    reportSHA256: release.output_sha256,
    bundle: await response.body(),
  };
}
