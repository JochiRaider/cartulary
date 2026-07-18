import type { Buffer } from "node:buffer";
import { fileURLToPath } from "node:url";
import {
  networkAnalysisTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { expect, type Page } from "@playwright/test";

import { createIncident } from "../../incidents/fixtures";
import { uniqueIncidentKey } from "../../runtime/fixtureIdentity";

export const networkFlowMinimalCSV = fileURLToPath(
  new URL(
    "../../../../../../fixtures/network-flow/NF-FIX-001-cisco-sna-minimal/source/cisco-sna-minimal.csv",
    import.meta.url,
  ),
);

export async function openNetworkFlowIncident(
  page: Page,
  prefix: string,
): Promise<string> {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey(prefix),
    `Phase 12 ${prefix} real application evidence`,
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  return incidentId;
}

export async function openClaimedNetworkAnalysis(
  page: Page,
  prefix: string,
): Promise<string> {
  expectNetworkFlowRuntimeProfile("network_flow_claimed");
  const incidentId = await openNetworkFlowIncident(page, prefix);
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
    file?: string | { buffer: Buffer; mimeType: string; name: string };
  },
) {
  await page
    .getByTestId(networkAnalysisTestId("import-input"))
    .setInputFiles(options.file ?? networkFlowMinimalCSV);
  const dialog = page.getByTestId(networkAnalysisTestId("mapping-dialog"));
  await expect(dialog).toBeVisible();
  await page
    .getByTestId(networkAnalysisTestId("mapping-display-name"))
    .fill(options.displayName);
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
