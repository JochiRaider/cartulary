const guideRowTableLinePattern =
  /^\|\s*(FE-(?:U|I|B|E|V|A11Y|S)-P(?:0|[1-9][0-9]*)-[0-9]{2})\s*\|/;
const guideMakeTargetPattern = /`make\s+([^`\s]+)(?:\s+[^`]*)?`/g;

export function collectFrontendGuideTargetRestatementErrors(
  guideText,
  rowTargetNames,
  guidePath = "frontend guide",
) {
  const errors = [];
  for (const [lineIndex, line] of guideText.split("\n").entries()) {
    const rowMatch = guideRowTableLinePattern.exec(line);
    if (!rowMatch) {
      continue;
    }
    const rowID = rowMatch[1];
    const liveTargets = rowTargetNames.get(rowID) ?? [];
    const liveTargetSet =
      liveTargets instanceof Set ? liveTargets : new Set(liveTargets);
    const liveTargetDisplay =
      liveTargetSet.size > 0
        ? [...liveTargetSet].sort().join(", ")
        : "none";
    guideMakeTargetPattern.lastIndex = 0;
    for (const targetMatch of line.matchAll(guideMakeTargetPattern)) {
      const targetName = targetMatch[1];
      if (!liveTargetSet.has(targetName)) {
        errors.push(
          `${guidePath}:${lineIndex + 1} row ${rowID} restates make ${targetName}, but live frontend phase-map targets are ${liveTargetDisplay}`,
        );
      }
    }
  }
  return errors;
}

