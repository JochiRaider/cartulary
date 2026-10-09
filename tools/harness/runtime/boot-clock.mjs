import { createHash } from "node:crypto";
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
