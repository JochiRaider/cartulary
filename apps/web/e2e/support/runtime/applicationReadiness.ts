import { expect, type Page, type Request } from "@playwright/test";

const monitors = new WeakMap<Page, string[]>();

export function installApplicationAssetMonitor(page: Page) {
  const failures: string[] = [];
  monitors.set(page, failures);
  let navigation = 0;
  const requests = new WeakMap<Request, number>();
  page.on("request", (request) => {
    if (request.isNavigationRequest() && request.frame() === page.mainFrame()) {
      navigation++;
      failures.splice(0);
    }
    requests.set(request, navigation);
  });
  page.on("response", (response) => {
    if (requests.get(response.request()) !== navigation) return;
    const url = new URL(response.url());
    if (url.pathname.startsWith("/assets/") && response.status() >= 400) {
      failures.push(`${response.status()} ${url.pathname}`);
    }
  });
  page.on("requestfailed", (request) => {
    if (requests.get(request) !== navigation) return;
    const url = new URL(request.url());
    if (url.pathname.startsWith("/assets/"))
      failures.push(`failed ${url.pathname}`);
  });
  return failures;
}

export async function expectApplicationReady(page: Page) {
  const failures = monitors.get(page);
  if (!failures) return;
  await expect
    .poll(
      async () => {
        if (failures.length) return "asset_error";
        return (await page.getByRole("main").isVisible()) ? "ready" : "loading";
      },
      { message: "application assets and shell must be ready" },
    )
    .not.toBe("loading");
  assertApplicationAssetsReady(page);
}

export function assertApplicationAssetsReady(page: Page) {
  const failures = monitors.get(page);
  if (failures?.length) {
    const error = new Error(
      `Required frontend assets failed: ${failures.join(", ")}`,
    );
    error.name = "CartularyFrontendArtifactError";
    throw error;
  }
}

export async function reloadVisualApplication(page: Page) {
  await page.reload();
  await expectApplicationReady(page);
}

export async function navigateVisualApplication(
  page: Page,
  url: string,
  options?: Parameters<Page["goto"]>[1],
) {
  await page.goto(url, options);
  await expectApplicationReady(page);
}
