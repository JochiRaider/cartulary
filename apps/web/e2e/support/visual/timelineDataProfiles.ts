import { readFileSync } from "node:fs";

export type TimelineDataProfile = {
  capture_id: string;
  kind: "rich" | "empty" | "sparse";
  reason: string;
};
export const timelineDataProfiles = JSON.parse(
  readFileSync(
    new URL(
      "../../../../../tools/frontend_visual_timeline_profiles.json",
      import.meta.url,
    ),
    "utf8",
  ),
) as { schema_id: string; captures: Record<string, TimelineDataProfile> };

export function requireTimelineDataProfile(name: string): TimelineDataProfile {
  if (
    timelineDataProfiles.schema_id !==
    "cartulary.frontend_visual_timeline_profiles.v1"
  )
    throw new Error("Unsupported Timeline capture profile version");
  const profile = timelineDataProfiles.captures[name];
  if (!profile?.reason.trim())
    throw new Error(`Undeclared Timeline capture data profile: ${name}`);
  return profile;
}
