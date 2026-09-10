import { Buffer } from "node:buffer";
import { networkFlowDecoders } from "@cartulary/protocol-ts/network-flow";
import {
  networkAnalysisEdgeTestId,
  networkAnalysisTestId,
  networkAnalysisVertexTestId,
} from "@cartulary/ui-contracts";
import type { Page, Response } from "@playwright/test";
import { expect, test } from "./fixtures";
import {
  importNetworkFlowCSV,
  openClaimedNetworkAnalysis,
} from "./support/extensions/network_flow_activity/workspace";

const graphRoute = /\/network-flow\/graphs\/query$/;
const contributorRoute = /\/network-flow\/graphs\/contributors\/query$/;
type RenderObservation = {
  samples: number;
  maxVertices: number;
  maxEdges: number;
  wrongBucket: string[];
};
type ObservationWindow = Window & {
  explorationObservation?: {
    observer: MutationObserver;
    result: RenderObservation;
  };
};
async function temporalGraph(page: Page, large = false) {
  await openClaimedNetworkAnalysis(page, large ? "GNBOUNDS" : "GNTEMPORAL");
  const rows = [
    "Start Time,End Time,Source IP,Destination IP,Source Port,Destination Port,Protocol,Bytes,Packets",
  ];
  const vertices = large ? 501 : 3,
    edges = large ? 1001 : 2;
  const ip = (n: number) => `192.0.${Math.floor(n / 256)}.${n % 256}`;
  for (const hour of [0, 2])
    for (let i = 0; i < edges; i++)
      rows.push(
        `2026-07-10T0${hour}:00:00Z,2026-07-10T0${hour}:01:00Z,${ip(i % vertices)},${ip((i + 1) % vertices)},443,${i + 1},6,${i + 1},1`,
      );
  await importNetworkFlowCSV(page, {
    displayName: "exploration-source",
    file: {
      name: "navigation.csv",
      mimeType: "text/csv",
      buffer: Buffer.from(`${rows.join("\n")}\n`),
    },
  });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  await expect(
    page.getByRole("button", { name: /^Select vertex/u }).first(),
  ).toBeVisible();
  await page.getByLabel("Time buckets", { exact: true }).check();
  await page.getByLabel("Flow starts at or after").fill("2026-07-10T00:00:00Z");
  await page.getByLabel("Flow starts before").fill("2026-07-10T03:00:00Z");
  const response = page.waitForResponse((r) => graphRoute.test(r.url()));
  await page.getByTestId(networkAnalysisTestId("accepted-query-apply")).click();
  const result = await response;
  expect(result.status()).toBe(200);
  const decoded = networkFlowDecoders.graphQueryResult.decode(
    (await result.json()).data,
  );
  if (!decoded.ok) throw new Error("Invalid real graph result");
  await expect(page.getByText("Bucket 1 of 3", { exact: true })).toBeVisible();
  return decoded.value;
}
async function contributors(response: Response) {
  expect(response.status()).toBe(200);
  const decoded = networkFlowDecoders.graphContributorQueryResult.decode(
    (await response.json()).data,
  );
  if (!decoded.ok) throw new Error("Invalid real contributors");
  return decoded.value;
}
async function nextBucket(page: Page) {
  const next = page.getByRole("button", { name: "Next bucket" });
  await next.focus();
  await next.press("Enter");
}

test("Network Analysis navigates exact temporal contributors and rejects superseded exposure", async ({
  page,
}) => {
  const graph = await temporalGraph(page);
  let graphQueries = 0;
  page.on("request", (request) => {
    if (graphRoute.test(request.url())) graphQueries++;
  });
  const first = graph.edge_annotations.find(
    (entry) =>
      entry.selector.kind === "time_bucket_edge" &&
      Date.parse(entry.selector.bucket_start_utc) ===
        Date.parse("2026-07-10T00:00:00Z"),
  );
  if (!first) throw new Error("Missing server temporal selector");
  const edgeControl = page
    .getByTestId(networkAnalysisEdgeTestId(first.selector.source_edge_id))
    .getByRole("button", { name: /^Select edge/u });
  let read = page.waitForResponse((r) => contributorRoute.test(r.url()));
  await edgeControl.click();
  const exact = await contributors(await read);
  expect(exact.selector).toEqual(first.selector);
  expect(exact.contributors).toHaveLength(1);
  expect(
    Date.parse(
      String(exact.contributors[0]?.row["network_flow.flow_start_utc"]),
    ),
  ).toBe(Date.parse("2026-07-10T00:00:00Z"));
  const drawer = page.getByTestId(networkAnalysisTestId("contributor-drawer"));
  await expect(drawer).toContainText("Contributors in bucket");
  await expect(drawer.getByRole("button", { name: "Link source" })).toHaveCount(
    0,
  );
  await nextBucket(page);
  await expect(drawer).toHaveCount(0);
  await expect(
    page.getByText("No flows start in this bucket.", { exact: true }),
  ).toBeVisible();
  await expect(page.getByTestId(/^network-flow-edge-/u)).toHaveCount(0);
  await expect(
    page.getByTestId(networkAnalysisTestId("graph-live-region")),
  ).toContainText("Selection cleared.");
  await nextBucket(page);
  const vertex = graph.vertex_selectors.find(
    (entry) => entry.selector.endpoint_value === "192.0.0.1",
  );
  if (!vertex) throw new Error("Missing query-wide vertex");
  const vertexControl = page
    .getByTestId(networkAnalysisVertexTestId(vertex.selector.source_vertex_id))
    .getByRole("button", { name: /^Select vertex/u });
  read = page.waitForResponse((r) => contributorRoute.test(r.url()));
  await vertexControl.click();
  const queryWide = await contributors(await read);
  expect(queryWide.selector).toEqual(vertex.selector);
  expect(queryWide.contributors).toHaveLength(4);
  expect(
    new Set(
      queryWide.contributors.map((item) =>
        Date.parse(String(item.row["network_flow.flow_start_utc"])),
      ),
    ),
  ).toEqual(
    new Set([
      Date.parse("2026-07-10T00:00:00Z"),
      Date.parse("2026-07-10T02:00:00Z"),
    ]),
  );
  await expect(drawer).toContainText(
    "Contributors across the full query range.",
  );
  await drawer
    .getByRole("button", { name: "Close graph contributors" })
    .click();
  await expect(vertexControl).toBeFocused();

  let release = () => {};
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  let fetched = () => {};
  const ready = new Promise<void>((resolve) => {
    fetched = resolve;
  });
  await page.route(
    contributorRoute,
    async (route) => {
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      fetched();
      await gate;
      await route.fulfill({ response }).catch(() => {});
    },
    { times: 1 },
  );
  await vertexControl.click();
  await ready;
  const previous = page.getByRole("button", { name: "Previous bucket" });
  await previous.focus();
  await previous.press("Enter");
  release();
  await expect(drawer).toHaveCount(0);
  await expect(
    page.getByText("No flows start in this bucket.", { exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Saved graphs", exact: true }).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("status-strip")),
  ).not.toContainText("Graph available");
  await page.getByRole("button", { name: "Unsaved exploration" }).click();
  await expect(page.getByText("Bucket 2 of 3", { exact: true })).toBeVisible();
  expect(graphQueries).toBe(0);
});

test("Network Analysis measures bounded exploration and reveals selection across subviews", async ({
  page,
}, testInfo) => {
  const graph = await temporalGraph(page, true);
  expect(graph.graph_projection_result.vertices).toHaveLength(501);
  expect(graph.graph_projection_result.edges).toHaveLength(2002);
  if (graph.result_variant.kind !== "time_bucket_v1")
    throw new Error("Missing temporal result");
  const expectedBuckets: string[][] = graph.result_variant.time_buckets.map(
    (bucket) =>
      graph.edge_annotations
        .filter(
          (entry) =>
            entry.selector.kind === "time_bucket_edge" &&
            entry.selector.bucket_start_utc === bucket.start_utc,
        )
        .map((entry) =>
          networkAnalysisEdgeTestId(entry.selector.source_edge_id),
        ),
  );
  await page.evaluate((expected) => {
    const result: RenderObservation = {
      samples: 0,
      maxVertices: 0,
      maxEdges: 0,
      wrongBucket: [],
    };
    const record = () => {
      const edges = [
        ...document.querySelectorAll<HTMLElement>(
          '[data-testid^="network-flow-edge-"]',
        ),
      ];
      const vertices = document.querySelectorAll(
        '[data-testid^="network-flow-vertex-"]',
      );
      const label =
        document.querySelector(
          'nav[aria-label="Time bucket navigation"] strong',
        )?.textContent ?? "";
      const bucket = Number(/Bucket (\d+)/u.exec(label)?.[1]) - 1;
      result.samples++;
      result.maxVertices = Math.max(result.maxVertices, vertices.length);
      result.maxEdges = Math.max(result.maxEdges, edges.length);
      for (const edge of edges)
        if (!expected[bucket]?.includes(edge.dataset.testid ?? ""))
          result.wrongBucket.push(edge.dataset.testid ?? "missing");
    };
    const observer = new MutationObserver(record);
    observer.observe(document.body, { childList: true, subtree: true });
    (window as ObservationWindow).explorationObservation = { observer, result };
    record();
  }, expectedBuckets);

  const mountedVertices = page.getByTestId(/^network-flow-vertex-/u),
    mountedEdges = page.getByTestId(/^network-flow-edge-/u);
  await expect(mountedVertices).toHaveCount(500);
  await expect(mountedEdges).toHaveCount(1000);
  let graphQueries = 0,
    contributorQueries = 0;
  page.on("request", (request) => {
    if (graphRoute.test(request.url())) graphQueries++;
    if (contributorRoute.test(request.url())) contributorQueries++;
  });
  await nextBucket(page);
  await nextBucket(page);
  const selected = graph.vertex_selectors.find(
    (v) => v.selector.endpoint_value === "192.0.0.0",
  );
  if (!selected) throw new Error("Missing semantic vertex");
  const selectedControl = page
    .getByTestId(
      networkAnalysisVertexTestId(selected.selector.source_vertex_id),
    )
    .getByRole("button", { name: /^Select vertex/u });
  const response = page.waitForResponse((r) => contributorRoute.test(r.url()));
  await selectedControl.click();
  const result = await contributors(await response);
  expect(result.selector).toEqual(selected.selector);
  expect(
    result.contributors.every(
      (item) =>
        item.row["network_flow.src_ip"] === "192.0.0.0" ||
        item.row["network_flow.dst_ip"] === "192.0.0.0",
    ),
  ).toBe(true);
  expect(result.contributors.length).toBeGreaterThan(1);
  const verticesNav = page.getByRole("navigation", {
    name: "vertices navigation",
  });
  const edgesNav = page.getByRole("navigation", { name: "edges navigation" });
  await verticesNav.getByRole("button", { name: "Next" }).focus();
  await verticesNav.getByRole("button", { name: "Next" }).press("Enter");
  await edgesNav.getByRole("button", { name: "Next" }).focus();
  await edgesNav.getByRole("button", { name: "Next" }).press("Enter");
  await expect(mountedVertices).toHaveCount(1);
  await expect(mountedEdges).toHaveCount(1);
  await page.getByRole("button", { name: "Saved graphs", exact: true }).click();
  await page.getByRole("button", { name: "Unsaved exploration" }).click();
  await expect(page.getByText("Bucket 3 of 3", { exact: true })).toBeVisible();
  await expect(mountedVertices).toHaveCount(1);
  await expect(mountedEdges).toHaveCount(1);
  await expect(
    page.getByRole("button", { name: "Unsaved exploration" }),
  ).toBeFocused();
  await page.getByRole("button", { name: "Reveal selected object" }).click();
  await expect(selectedControl).toBeFocused();
  await testInfo.attach("revealed-selection.png", {
    contentType: "image/png",
    body: await page.screenshot(),
  });
  expect(
    await selectedControl.evaluate((node) => {
      const bounds = node.getBoundingClientRect();
      const hit = document.elementFromPoint(
        bounds.x + bounds.width / 2,
        bounds.y + bounds.height / 2,
      );
      return {
        unobscured: node.contains(hit),
        bounds: bounds.toJSON(),
        hit: hit?.outerHTML.slice(0, 300),
      };
    }),
  ).toMatchObject({ unobscured: true });

  await expect(mountedVertices).toHaveCount(500);
  await expect(mountedEdges).toHaveCount(1);
  await page.getByRole("button", { name: "Close graph contributors" }).focus();
  await page.keyboard.press("Escape");
  await expect(selectedControl).toBeFocused();
  expect(graphQueries).toBe(0);
  expect(contributorQueries).toBe(1);
  const observed = await page.evaluate(() => {
    const observation = (window as ObservationWindow).explorationObservation;
    if (!observation) throw new Error("Missing rendering observation");
    observation.observer.disconnect();
    return observation.result;
  });
  expect(observed.samples).toBeGreaterThan(5);
  expect(observed.maxVertices).toBe(500);
  expect(observed.maxEdges).toBe(1000);
  expect(observed.wrongBucket).toEqual([]);
  await testInfo.attach("exploration-rendering.json", {
    contentType: "application/json",
    body: Buffer.from(
      JSON.stringify({
        projection_result_id:
          graph.graph_projection_result.projection_result_id,
        graph_query_digest: graph.graph_query_digest,
        selected: selected.selector,
        vertex_limit: 500,
        edge_limit: 1000,
        logical_vertices: 501,
        logical_edges: 2002,
        bucket_edges: 1001,
        last_vertex_page: 1,
        last_edge_page: 1,
        contributors: result.contributors.map((item) => item.row_ref),
        local_graph_queries: graphQueries,
      }),
    ),
  });
});

test("a11y Network Analysis temporal navigation keeps focus and announces selection clearing", async ({
  page,
}) => {
  const graph = await temporalGraph(page);
  const first = graph.edge_annotations.find(
    (e) =>
      e.selector.kind === "time_bucket_edge" &&
      Date.parse(e.selector.bucket_start_utc) ===
        Date.parse("2026-07-10T00:00:00Z"),
  );
  if (!first) throw new Error("Missing temporal edge");
  const control = page
    .getByTestId(networkAnalysisEdgeTestId(first.selector.source_edge_id))
    .getByRole("button", { name: /^Select edge/u });
  await control.focus();
  await control.press("Enter");
  const close = page.getByRole("button", { name: "Close graph contributors" });
  await close.focus();
  await close.press("Escape");
  await expect(control).toBeFocused();
  await control.press("Enter");
  await nextBucket(page);
  await expect(page.getByRole("button", { name: "Next bucket" })).toBeFocused();
  await expect(
    page.getByTestId(networkAnalysisTestId("graph-live-region")),
  ).toContainText("Bucket 2 of 3. Selection cleared.");
  await nextBucket(page);
  const next = page.getByRole("button", { name: "Next bucket" });
  await expect(next).toHaveAttribute("aria-disabled", "true");
  await next.press("Enter");
  await expect(next).toBeFocused();
  await expect(page.getByText("Bucket 3 of 3", { exact: true })).toBeVisible();
  const focusStyle = await next.evaluate((node) => {
    const style = getComputedStyle(node);
    return { outline: style.outlineStyle, width: style.outlineWidth };
  });
  expect(focusStyle.outline).not.toBe("none");
  expect(focusStyle.width).not.toBe("0px");
  await test.info().attach("exploration-accessibility.txt", {
    contentType: "text/plain",
    body: await page
      .getByTestId(networkAnalysisTestId("graph-panel"))
      .ariaSnapshot(),
  });
});
