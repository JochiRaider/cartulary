import { describe, expect, it } from "vitest";
import {
  beginLatestQuery,
  type LatestQueryRuntime,
} from "./workbookLatestRequest";

describe("workbookLatestRequest", () => {
  it("keeps requests exclusive and aborts superseded controllers", () => {
    const runtime: { current: LatestQueryRuntime } = {
      current: { controller: null, sequence: 0 },
    };

    const first = beginLatestQuery(runtime);
    const second = beginLatestQuery(runtime);

    expect(first.signal.aborted).toBe(true);
    expect(first.isCurrent()).toBe(false);
    expect(second.signal.aborted).toBe(false);
    expect(second.isCurrent()).toBe(true);
  });
});
