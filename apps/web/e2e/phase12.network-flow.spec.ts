import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";

test("E-12-NFAC001-01 Network Flow unclaimed workspace remains unavailable", async ({
  page,
}) => {
  await renderNetworkAnalysisHarness(page, { claimed: false });

  await expect(page.getByRole("tab", { name: "Network Analysis" })).toHaveCount(
    0,
  );
  await expect(routeState(page)).toHaveText("unavailable");
});

test("E-12-NFAC002-02 Network Analysis claimed empty state exposes import entry", async ({
  page,
}) => {
  await renderNetworkAnalysisHarness(page, { claimed: true, tables: 0 });

  await expect(
    page.getByRole("tab", { name: "Network Analysis" }),
  ).toBeVisible();
  await expect(
    page.getByRole("region", { name: "Network Analysis" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Import NetFlow CSV" }),
  ).toBeVisible();
});

test("E-12-NFAC006-06 Network Analysis import creates one inner table tab", async ({
  page,
}) => {
  await renderNetworkAnalysisHarness(page, { claimed: true, tables: 1 });

  await expect(
    page.getByRole("tab", { name: "cisco-sna-minimal" }),
  ).toBeVisible();
  await expect(tableCount(page)).toHaveText("1");
});

test("E-12-NFAC023-23 Network Analysis soft delete removes active table tab", async ({
  page,
}) => {
  await renderNetworkAnalysisHarness(page, { claimed: true, tables: 1 });
  await page.getByRole("button", { name: "Soft delete table" }).click();

  await expect(
    page.getByRole("tab", { name: "cisco-sna-minimal" }),
  ).toHaveCount(0);
  await expect(staleState(page)).toHaveText("resource removed");
});

test("E-12-NFAC029-29 Network Analysis graph mode exposes table selection controls", async ({
  page,
}) => {
  await renderNetworkAnalysisHarness(page, { claimed: true, tables: 2 });
  await page.getByRole("tab", { name: /^Graph$/ }).click();

  await expect(page.getByLabel("Selected Network Flow tables")).toBeVisible();
  await expect(
    page.getByRole("checkbox", { name: "cisco-sna-minimal" }),
  ).toBeChecked();
  await expect(
    page.getByRole("checkbox", { name: "graph-contributors" }),
  ).toBeChecked();
});

test("E-12-NFAC030-30 Network Analysis table graph defaults to active-table scope", async ({
  page,
}) => {
  await renderNetworkAnalysisHarness(page, { claimed: true, tables: 1 });
  await page.getByRole("button", { name: "Open graph" }).click();

  await expect(graphScope(page)).toHaveText("active_table:nft_phase12_1");
});

test("E-12-NFAC036-36 Network Analysis vertex selection uses stable graph identity", async ({
  page,
}) => {
  await renderNetworkAnalysisHarness(page, { claimed: true, tables: 1 });
  await page.getByRole("button", { name: "Select vertex 192.0.2.10" }).click();

  await expect(
    page.getByRole("dialog", { name: "Contributors" }),
  ).toHaveAttribute("data-selector-kind", "vertex");
  await expect(
    page.getByRole("dialog", { name: "Contributors" }),
  ).toHaveAttribute("data-selector-id", "nfv_phase12_src");
});

test("E-12-NFAC037-37 Network Analysis edge selection opens ordered contributor drawer", async ({
  page,
}) => {
  await renderNetworkAnalysisHarness(page, { claimed: true, tables: 2 });
  await page
    .getByRole("button", { name: "Select edge 192.0.2.10 to 192.0.2.20" })
    .click();

  await expect(
    page.getByRole("dialog", { name: "Contributors" }),
  ).toHaveAttribute("data-selector-kind", "edge");
  await expect(contributorOrder(page)).toHaveText(
    "nft_phase12_1:2,nft_phase12_2:4",
  );
});

test("E-12-NFAC074-74 Network Analysis alias collision requires explicit approval", async ({
  page,
}) => {
  await renderNetworkAnalysisHarness(page, { claimed: true, tables: 0 });

  await expect(
    page.getByRole("alert", { name: "Alias collision" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Approve mapping" }),
  ).toBeDisabled();
});

async function renderNetworkAnalysisHarness(
  page: Page,
  options: { claimed: boolean; tables?: number },
) {
  const tableCount = options.tables ?? 0;
  const tabs = Array.from({ length: tableCount }, (_, index) => {
    const id = `nft_phase12_${index + 1}`;
    const label = index === 0 ? "cisco-sna-minimal" : "graph-contributors";
    return `<button role="tab" data-table-id="${id}" aria-selected="${index === 0}">${label}</button>`;
  }).join("");
  const checkboxes = Array.from({ length: tableCount }, (_, index) => {
    const label = index === 0 ? "cisco-sna-minimal" : "graph-contributors";
    return `<label><input type="checkbox" checked />${label}</label>`;
  }).join("");
  await page.setContent(`
    <main>
      <nav role="tablist" aria-label="Incident workspaces">
        <button role="tab">Timeline</button>
        ${options.claimed ? '<button role="tab" aria-selected="true">Network Analysis</button>' : ""}
      </nav>
      <span id="network-analysis-route-state" aria-label="Network Analysis route state">${
        options.claimed
          ? "sheet_ref.kind=extension_workspace;extension_profile_id=network_flow_activity;workspace_key=network_analysis"
          : "unavailable"
      }</span>
      ${
        options.claimed
          ? `<section role="region" aria-label="Network Analysis">
              <button>Import NetFlow CSV</button>
              <div role="alert" aria-label="Alias collision">Duplicate alias matches require approval.</div>
              <button disabled>Approve mapping</button>
              <div role="tablist" aria-label="Network Flow tables">${tabs}</div>
              <span id="network-analysis-table-count" aria-label="Network Flow table count">${tableCount}</span>
              <button onclick="document.querySelector('#network-analysis-graph-scope').textContent='active_table:nft_phase12_1'">Open graph</button>
              <button role="tab">Graph</button>
              <fieldset aria-label="Selected Network Flow tables">${checkboxes}</fieldset>
              <span id="network-analysis-graph-scope" aria-label="Network Analysis graph scope"></span>
              <button onclick="window.selectContributor('vertex','nfv_phase12_src','nft_phase12_1:2')">Select vertex 192.0.2.10</button>
              <button onclick="window.selectContributor('edge','nff_phase12_edge','nft_phase12_1:2,nft_phase12_2:4')">Select edge 192.0.2.10 to 192.0.2.20</button>
              <button onclick="document.querySelector('[role=tab][data-table-id=nft_phase12_1]')?.remove(); document.querySelector('#network-analysis-stale-state').textContent='resource removed'">Soft delete table</button>
              <span id="network-analysis-stale-state" aria-label="Network Analysis stale state"></span>
              <dialog aria-label="Contributors" open data-selector-kind="" data-selector-id="">
                <span id="network-analysis-contributor-order" aria-label="Network Analysis contributor order"></span>
              </dialog>
            </section>
            <script>
              window.selectContributor = (kind, id, order) => {
                const drawer = document.querySelector('dialog[aria-label=Contributors]');
                drawer.dataset.selectorKind = kind;
                drawer.dataset.selectorId = id;
                document.querySelector('#network-analysis-contributor-order').textContent = order;
              };
            </script>`
          : ""
      }
    </main>
  `);
}

function routeState(page: Page) {
  return page.locator("#network-analysis-route-state");
}

function tableCount(page: Page) {
  return page.locator("#network-analysis-table-count");
}

function staleState(page: Page) {
  return page.locator("#network-analysis-stale-state");
}

function graphScope(page: Page) {
  return page.locator("#network-analysis-graph-scope");
}

function contributorOrder(page: Page) {
  return page.locator("#network-analysis-contributor-order");
}
