import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const timelineDirectory = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const workbookSource = readFileSync(
  path.join(timelineDirectory, "components/TimelineWorkbook.tsx"),
  "utf8",
);

describe("Timeline controller boundary policy", () => {
  it("composes every capability family through an explicit controller", () => {
    for (const controller of [
      "useTimelineMutationCommands",
      "useTimelinePendingReplayController",
      "useTimelineLiveUpdateController",
      "useTimelineMentionActions",
      "useTimelineEvidenceAttach",
      "useTimelineHistoryActions",
      "useTimelineViewportContinuityController",
    ]) {
      expect(workbookSource).toContain(`${controller}({`);
    }
  });

  it("prevents capability controllers from importing sibling controllers", () => {
    const hooksDirectory = path.join(timelineDirectory, "hooks");
    for (const entry of readdirSync(hooksDirectory, { withFileTypes: true })) {
      if (!entry.isFile() || !entry.name.endsWith(".ts")) {
        continue;
      }
      const source = readFileSync(
        path.join(hooksDirectory, entry.name),
        "utf8",
      );
      expect(source, entry.name).not.toMatch(/from\s+["']\.\/useTimeline/u);
    }
  });

  it("keeps socket lifecycle and private resume state outside Timeline", () => {
    expect(workbookSource).not.toContain("new WebSocket");
    expect(workbookSource).not.toContain("resumeToken");
    expect(workbookSource).not.toContain("heartbeat");
    expect(workbookSource).not.toContain("lastSequence");
  });
});
