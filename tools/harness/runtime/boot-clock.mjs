import { createHash, randomUUID } from "node:crypto";
import { readFileSync, readlinkSync } from "node:fs";

// Linux exposes boot-relative time across short-lived helper processes. This
// clock includes suspend and is deliberately not the scheduler's clock.
export function readBootClock({ read = readFileSync, link = readlinkSync } = {}) {
  try {
    const boot = read("/proc/sys/kernel/random/boot_id", "utf8").trim();
    const namespace = link("/proc/self/ns/time");
    const uptime = read("/proc/uptime", "utf8").trim().split(/\s+/u)[0];
    if (!/^\d+\.\d{2}$/u.test(uptime) || !boot || !namespace) throw new Error("unsupported boot clock");
    const [seconds, hundredths] = uptime.split(".");
    const milliseconds = Number(seconds) * 1000 + Number(hundredths) * 10;
    if (!Number.isSafeInteger(milliseconds)) throw new Error("boot clock overflow");
    return { clock: "linux_boottime_10ms", clock_identity: `sha256:${createHash("sha256").update(`${boot}:${namespace}`).digest("hex")}`, monotonic_ms: milliseconds };
  } catch {
    return { clock: "unavailable", clock_identity: null, monotonic_ms: null };
  }
}

// Process-local monotonic coordinates are retained only for placement. Canonical
// scheduler time and independently measured activity duration keep their owners.
const localClockIdentity = `clock:${randomUUID()}`;
export const readLocalClock = () => ({ clock: "node_monotonic", clock_identity: localClockIdentity,
  resolution_ms: 0.000001, monotonic_ms: Number(process.hrtime.bigint()) / 1e6 });

export function captureClockAnchor(readLocal = readLocalClock, readBoot = readBootClock) {
  try {
    const before = readLocal(), boot = readBoot(), after = readLocal();
    if (!before.clock_identity || before.clock_identity !== after.clock_identity || !boot.clock_identity ||
        ![before.monotonic_ms, after.monotonic_ms, boot.monotonic_ms].every(Number.isFinite) ||
        after.monotonic_ms < before.monotonic_ms) return null;
    return { clock_identity: before.clock_identity, boot_identity: boot.clock_identity,
      local_before_ms: before.monotonic_ms, local_after_ms: after.monotonic_ms,
      boot_ms: boot.monotonic_ms, boot_resolution_ms: 10 };
  } catch { return null; }
}
