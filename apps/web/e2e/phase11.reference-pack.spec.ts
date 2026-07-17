import {
  phase1LandingTestId,
  referencePackAdminPanelTestId,
  referencePackCancelButtonTestId,
  referencePackJobStatusTestId,
  referencePackRefreshAllButtonTestId,
} from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import { DeploymentAdministration } from "./pages/deploymentAdministration";
import { IncidentDirectory } from "./pages/incidentDirectory";

const jobID = "phase11-reference-pack-browser-job";

test("E-11-01 shows Reference Pack progress and cancel controls without blocking landing interaction", async ({
  page,
}) => {
  let jobReads = 0;
  let cancelRequests = 0;

  await page.route("**/api/v1/reference-packs/refresh", async (route) => {
    expect(route.request().method()).toBe("POST");
    await route.fulfill({
      status: 202,
      contentType: "application/json",
      body: JSON.stringify({
        data: jobResource(jobID, "queued", true, 0, 3),
      }),
    });
  });

  await page.route(new RegExp(`/api/v1/jobs/${jobID}$`), async (route) => {
    jobReads += 1;
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: jobResource(jobID, "running", true, 1, 3),
      }),
    });
  });

  await page.route(
    new RegExp(`/api/v1/jobs/${jobID}/cancel$`),
    async (route) => {
      cancelRequests += 1;
      expect(route.request().method()).toBe("POST");
      await route.fulfill({
        status: 202,
        contentType: "application/json",
        body: JSON.stringify({
          data: jobResource(jobID, "cancel_requested", false, 1, 3),
        }),
      });
    },
  );

  await page.goto("/");
  await new DeploymentAdministration(page).selectPanel("reference-packs");
  await expect(page.getByTestId(referencePackAdminPanelTestId())).toBeVisible();

  await page.getByTestId(referencePackRefreshAllButtonTestId()).click();
  await expect(page.getByTestId(referencePackJobStatusTestId())).toContainText(
    "running",
    { timeout: 1000 },
  );
  await expect(page.getByTestId(referencePackCancelButtonTestId())).toBeVisible(
    {
      timeout: 1000,
    },
  );

  await new IncidentDirectory(page).open();
  await page.getByTestId(phase1LandingTestId("create-open-button")).click();
  await expect(
    page.getByTestId(phase1LandingTestId("incident-key")),
  ).toBeVisible();
  await page
    .getByTestId(phase1LandingTestId("incident-key"))
    .fill("IR-E-11-01");
  await expect(
    page.getByTestId(phase1LandingTestId("incident-key")),
  ).toHaveValue("IR-E-11-01");
  await page.getByRole("button", { name: "Close new incident" }).click();

  await new DeploymentAdministration(page).selectPanel("reference-packs");
  await page.getByTestId(referencePackCancelButtonTestId()).click();
  await expect(page.getByTestId(referencePackJobStatusTestId())).toContainText(
    "cancel_requested",
  );
  expect(jobReads).toBeGreaterThanOrEqual(1);
  expect(cancelRequests).toBe(1);
});

function jobResource(
  id: string,
  status: string,
  cancelable: boolean,
  completed: number,
  total: number,
) {
  return {
    job_id: id,
    status,
    cancelable,
    progress: {
      completed,
      total,
    },
    result_summary: null,
    error_summary: null,
    status_route: `/api/v1/jobs/${id}`,
  };
}
