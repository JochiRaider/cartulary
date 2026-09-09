import type { GetCurrentAccountPreferencesResponse } from "@cartulary/protocol-ts/http";
import { type Page, type Route, test } from "@playwright/test";

type Preferences = GetCurrentAccountPreferencesResponse["data"];
export type VisualDensity = Preferences["density_mode"];
type WriteHandler = (route: Route) => Promise<void>;

export class VisualPreferences {
  private value: Preferences;
  private writeHandler: WriteHandler | undefined;
  readonly unexpectedWrites: string[] = [];

  constructor(userId: string) {
    this.value = {
      user_id: userId,
      density_mode: null,
      preferences_version: 1,
      created_at: "2026-09-05T12:00:00Z",
      updated_at: "2026-09-05T12:00:00Z",
    };
  }

  read(): Preferences {
    return { ...this.value };
  }

  select(density: VisualDensity) {
    this.value = {
      ...this.value,
      density_mode: density,
      preferences_version: this.value.preferences_version + 1,
    };
  }

  handleWrites(handler: WriteHandler) {
    if (this.writeHandler)
      throw new Error("visual preferences already have a write owner");
    this.writeHandler = handler;
  }

  async route(route: Route) {
    if (route.request().method() === "GET") {
      await route.fulfill({
        json: {
          data: this.read(),
          meta: { request_id: "visual-preferences" },
        } satisfies GetCurrentAccountPreferencesResponse,
      });
    } else if (this.writeHandler) {
      await this.writeHandler(route);
    } else {
      this.unexpectedWrites.push(route.request().method());
      await route.abort("blockedbyclient");
      const error = new Error("Unexpected visual preference write");
      error.name = "CartularyVisualCaptureError";
      throw error;
    }
  }
}

const resources = new WeakMap<Page, VisualPreferences>();

export async function installVisualPreferences(page: Page, userId: string) {
  const existing = resources.get(page);
  if (existing) return existing;
  const resource = new VisualPreferences(userId);
  await page.route("**/api/v1/account/preferences", (route) =>
    resource.route(route),
  );
  resources.set(page, resource);
  return resource;
}

export function visualPreferences(page: Page) {
  const resource = resources.get(page);
  if (!resource)
    throw new Error("visual preferences must be installed before navigation");
  return resource;
}

export async function selectVisualDensity(page: Page, density: VisualDensity) {
  await test.step(`select visual density: ${density ?? "inherited"}`, async () => {
    visualPreferences(page).select(density);
  });
}
