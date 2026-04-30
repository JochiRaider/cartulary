export class SchedulerClock {
  constructor({ wallOriginMs = Date.now(), monotonicOriginNs = process.hrtime.bigint() } = {}) {
    this.wallOriginMs = wallOriginMs;
    this.monotonicOriginNs = monotonicOriginNs;
    this.lastRawWallMs = wallOriginMs;
  }

  monotonicMs() {
    const elapsedNs = process.hrtime.bigint() - this.monotonicOriginNs;
    return Math.max(0, Number(elapsedNs / 1_000_000n));
  }

  wallTimestamp(monotonicMs = this.monotonicMs()) {
    return new Date(this.wallOriginMs + monotonicMs).toISOString();
  }

  observeRawWallClock() {
    const current = Date.now();
    const previous = this.lastRawWallMs;
    this.lastRawWallMs = current;
    if (current < previous) {
      return {
        previous_raw_wall_ms: previous,
        current_raw_wall_ms: current,
        regression_ms: previous - current,
      };
    }
    return null;
  }
}

