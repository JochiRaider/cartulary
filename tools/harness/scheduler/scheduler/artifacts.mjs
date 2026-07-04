import { writeFile } from "node:fs/promises";

import { prettyJSONString } from "../../contract/index.mjs";

export async function writeSchedulerSummaryArtifacts({
  pressureSummary,
  pressureSummaryPath,
  schedulerSummary,
  summaryPath,
}) {
  await writeFile(summaryPath, prettyJSONString(schedulerSummary));
  await writeFile(pressureSummaryPath, prettyJSONString(pressureSummary));
}
