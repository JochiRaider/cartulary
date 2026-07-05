export const frontendPhaseIDPattern = /^FE-P(?:0|[1-9][0-9]*)$/;
export const frontendRowIDPattern =
  /^FE-(?:U|I|B|E|V|A11Y|S)-P(?:0|[1-9][0-9]*)-[0-9]{2}$/;
export const frontendVisualFixtureIDPattern =
  /^FE-VFIX-(?:0[1-9]|[1-9][0-9]+)$/;

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

export function frontendPhaseNumber(phaseID) {
  const match = /^FE-P(0|[1-9][0-9]*)$/.exec(String(phaseID));
  return match ? Number.parseInt(match[1], 10) : Number.NaN;
}

export function frontendPhaseToBasePhase(phaseID) {
  const phaseNumber = frontendPhaseNumber(phaseID);
  return Number.isFinite(phaseNumber) ? `phase${phaseNumber}` : "";
}

export function compareFrontendPhaseIDs(left, right) {
  const leftNumber = frontendPhaseNumber(left);
  const rightNumber = frontendPhaseNumber(right);
  if (Number.isFinite(leftNumber) && Number.isFinite(rightNumber)) {
    return leftNumber - rightNumber;
  }
  return compareStrings(left, right);
}

export function frontendPhaseRangeLabel(registryOrPhases) {
  const phases = Array.isArray(registryOrPhases)
    ? registryOrPhases
    : registryOrPhases?.phases ?? [];
  const ids = phases
    .map((entry) => entry?.phase_id)
    .filter((phaseID) => frontendPhaseIDPattern.test(String(phaseID)))
    .sort(compareFrontendPhaseIDs);
  if (ids.length === 0) {
    return "declared frontend phases";
  }
  return `${ids[0]} through ${ids[ids.length - 1]}`;
}
