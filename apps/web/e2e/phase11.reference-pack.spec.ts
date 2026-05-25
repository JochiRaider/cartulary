import { expect, test } from "./fixtures";

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
  await expect(page.getByTestId("reference-pack-admin-panel")).toBeVisible();
  await expect(page.getByTestId("landing-incident-key")).toBeVisible();

  await page.getByTestId("reference-pack-refresh-all").click();
  await expect(page.getByTestId("reference-pack-job-status")).toContainText(
    "running",
    { timeout: 1000 },
  );
  await expect(page.getByTestId("reference-pack-cancel")).toBeVisible({
    timeout: 1000,
  });

  await page.getByTestId("landing-incident-key").fill("IR-E-11-01");
  await expect(page.getByTestId("landing-incident-key")).toHaveValue(
    "IR-E-11-01",
  );

  await page.getByTestId("reference-pack-cancel").click();
  await expect(page.getByTestId("reference-pack-job-status")).toContainText(
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
