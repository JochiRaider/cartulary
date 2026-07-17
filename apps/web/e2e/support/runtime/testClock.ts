import type { Page } from "@playwright/test";
import { createEnvironmentTestControlClient } from "../transport/testControlEnvironment";
import { apiBase } from "./configuration";

export class TestClock {
  constructor(private readonly page: Page) {}

  async setOffset(offsetSeconds: number) {
    const response = await createEnvironmentTestControlClient(
      this.page.request,
      { endpointOrigin: apiBase },
    ).request({
      body: { offset_seconds: offsetSeconds },
      method: "POST",
      path: "/api/v1/test/clock/set",
    });
    if (!response.ok) {
      throw new Error(
        `test-clock offset update failed with ${response.status}`,
      );
    }
  }

  async setFixed(fixedNow: string | Date) {
    const response = await createEnvironmentTestControlClient(
      this.page.request,
      { endpointOrigin: apiBase },
    ).request({
      body: {
        fixed_now: fixedNow instanceof Date ? fixedNow.toISOString() : fixedNow,
      },
      method: "POST",
      path: "/api/v1/test/clock/set",
    });
    if (!response.ok) {
      throw new Error(`test-clock fixed update failed with ${response.status}`);
    }
  }

  async setAfter(timestamp: string, deltaMs = 2000) {
    const baseline = new Date(timestamp);
    if (Number.isNaN(baseline.getTime())) {
      throw new Error("cannot set test clock after invalid timestamp");
    }
    await this.setFixed(new Date(baseline.getTime() + deltaMs));
  }

  async reset() {
    await this.setOffset(0);
  }
}
