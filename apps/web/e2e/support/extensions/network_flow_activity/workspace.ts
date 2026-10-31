import type { Buffer } from "node:buffer";
import {
  networkAnalysisTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { expect, type Fixtures, type Page } from "@playwright/test";
import { runNetworkFlowFixture } from "../../../../src/testing/networkFlowFixtureTestSupport";

import { createIncident } from "../../incidents/fixtures";
import { uniqueIncidentKey } from "../../runtime/fixtureIdentity";

export type NetworkFlowFixtureFields = {
  networkFlowMinimalCSV: { buffer: Buffer; mimeType: string; name: string };
  networkFlowCiscoSearchCSV: { buffer: Buffer; mimeType: string; name: string };
};

// Per-test admission wraps the actual browser assertions. Listing tests never
// reads the corpus or creates an execution summary.
export const networkFlowFixtureSetup: Fixtures<NetworkFlowFixtureFields> = {
  // biome-ignore lint/correctness/noEmptyPattern: Playwright requires a destructured dependency list; admission has no fixture dependencies.
  networkFlowMinimalCSV: async ({}, use, info) => {
    await runNetworkFlowFixture(
      { fixtureID: "NF-FIX-001-cisco-sna-minimal", revision: 8 },
      async (fixture) => {
        await use({
          buffer: fixture.read("source/cisco-sna-minimal.csv"),
          mimeType: "text/csv",
          name: "cisco-sna-minimal.csv",
        });
        fixture.record(
          "browser-comparison.json",
          JSON.stringify(
            {
              test_id: info.testId,
              title: info.title,
              status: info.status,
              expected_status: info.expectedStatus,
            },
            null,
            2,
          ) + "\n",
        );
        if (info.status !== "passed" || info.expectedStatus !== "passed") {
          throw new Error("Browser fixture comparison did not pass");
        }
      },
    );
  },
  // biome-ignore lint/correctness/noEmptyPattern: Playwright requires a destructured dependency list; admission has no fixture dependencies.
  networkFlowCiscoSearchCSV: async ({}, use, info) => {
    await runNetworkFlowFixture(
      { fixtureID: "NF-FIX-029-cisco-sna-flow-search", revision: 5 },
      async (fixture) => {
        await use({
          buffer: fixture.read(
            "source/cisco_sna_netflow_example_1200_rows.csv",
          ),
          mimeType: "text/csv",
          name: "cisco_sna_netflow_example_1200_rows.csv",
        });
        fixture.record(
          "browser-comparison.json",
          JSON.stringify(
            {
              test_id: info.testId,
              title: info.title,
              status: info.status,
              expected_status: info.expectedStatus,
            },
            null,
            2,
          ) + "\n",
        );
        if (info.status !== "passed" || info.expectedStatus !== "passed") {
          throw new Error("Browser fixture comparison did not pass");
        }
      },
    );
  },
};

export async function openNetworkFlowIncident(
  page: Page,
  prefix: string,
  options: {
    readonly onIncidentCreated?: (incidentId: string) => void;
  } = {},
): Promise<string> {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey(prefix),
    `Network Flow ${prefix} real application evidence`,
  );
  options.onIncidentCreated?.(incidentId);
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  return incidentId;
}

export async function openClaimedNetworkAnalysis(
  page: Page,
  prefix: string,
  options: {
    readonly onIncidentCreated?: (incidentId: string) => void;
  } = {},
): Promise<string> {
  expectNetworkFlowRuntimeProfile("network_flow_claimed");
  const incidentId = await openNetworkFlowIncident(page, prefix, options);
  await page.getByRole("button", { name: "More views", exact: true }).click();
  const tab = page.getByTestId(networkAnalysisTestId("tab"));
  await expect(tab).toBeVisible();
  await tab.click();
  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")),
  ).toBeVisible();
  return incidentId;
}

export async function importNetworkFlowCSV(
  page: Page,
  options: {
    displayName: string;
    sourceProfileId?: string;
    file: string | { buffer: Buffer; mimeType: string; name: string };
  },
) {
  await page
    .getByTestId(networkAnalysisTestId("import-input"))
    .setInputFiles(options.file);
  const dialog = page.getByTestId(networkAnalysisTestId("mapping-dialog"));
  await expect(dialog).toBeVisible({ timeout: 30_000 });
  await page
    .getByTestId(networkAnalysisTestId("mapping-display-name"))
    .fill(options.displayName);
  if (options.sourceProfileId) {
    await page
      .getByTestId(networkAnalysisTestId("mapping-profile"))
      .selectOption(options.sourceProfileId);
  }
  await page.getByTestId(networkAnalysisTestId("mapping-preview")).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("mapping-preview-summary")),
  ).toBeVisible();
  await expect(
    page.getByTestId(networkAnalysisTestId("mapping-apply")),
  ).toBeEnabled();
  await page.getByTestId(networkAnalysisTestId("mapping-apply")).click();
  await expect(dialog).toHaveCount(0);
  await expect(
    page.getByRole("tab", { name: new RegExp(options.displayName) }),
  ).toBeVisible();
}

export function expectNetworkFlowRuntimeProfile(
  expected: "default" | "network_flow_claimed",
) {
  expect(process.env.CARTULARY_BROWSER_RUNTIME_PROFILE_ID ?? "default").toBe(
    expected,
  );
}
