export function startFrontendPublication(measurement: boolean): {
  finished: Promise<string>;
  stop: () => Promise<void>;
};
