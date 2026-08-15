import { validateAc043MeasurementObservation } from "./ac043-large-grid.mjs";

const observationValidators = new Map([
  ["ac043_large_grid_snapshot_v1", validateAc043MeasurementObservation],
]);

export function validateProfileMeasurementObservation(root, profile, observation) {
  const validate = observationValidators.get(profile.fixture_profile_id);
  if (validate === undefined) {
    throw new Error(
      `performance fixture profile lacks a measurement adapter: ${profile.fixture_profile_id}`,
    );
  }
  validate(root, profile, observation);
}
