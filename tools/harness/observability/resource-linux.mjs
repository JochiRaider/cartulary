import { createHash } from "node:crypto";
import { readFileSync, readlinkSync, opendirSync } from "node:fs";
import path from "node:path";

export const unavailable = (availability) => ({ value: null, availability });
export const observed = (value) => Number.isFinite(value) && value >= 0
  ? { value, availability: "available" } : unavailable("not_observed");
export function availabilityFor(error) {
  return ["EACCES", "EPERM"].includes(error?.code) ? "permission_denied"
    : ["ENOENT", "ESRCH"].includes(error?.code) ? "process_gone" : "unsupported";
}
export function parseProcessStat(text) {
  const end = text.lastIndexOf(") ");
  if (end < 0) throw new Error("invalid process stat");
  const f = text.slice(end + 2).trim().split(/\s+/u);
  const number = (index) => {
    if (!/^\d+$/u.test(f[index] ?? "")) throw new Error("invalid stat counter");
    const result = Number(f[index]);
    if (!Number.isSafeInteger(result)) throw new Error("stat counter overflow");
    return result;
  };
  number(19);
  return { state: f[0], parent: number(1), start: f[19], user_ticks: number(11),
    system_ticks: number(12), threads: number(17), rss_pages: number(21) };
}
export function parseCounters(text) {
  const result = {};
  for (const line of text.trim().split("\n")) {
    const [key, value] = line.replace(":", "").trim().split(/\s+/u);
    if (/^\d+$/u.test(value ?? "") && Number.isSafeInteger(Number(value))) result[key] = Number(value);
  }
  return result;
}
export function parsePSI(text) {
  const result = {};
  for (const line of text.trim().split("\n")) {
    const match = /^(some|full) .*\btotal=(\d+)$/u.exec(line);
    if (match && Number.isSafeInteger(Number(match[2]))) result[`${match[1]}_us`] = Number(match[2]);
  }
  return result;
}
export function linuxUnits(auxv) {
  // Supported profile is Linux x86_64. AT_PAGESZ and AT_CLKTCK avoid a subprocess.
  let pageBytes = null, ticksPerSecond = null;
  for (let offset = 0; offset + 16 <= auxv.length; offset += 16) {
    const key = auxv.readBigUInt64LE(offset), value = Number(auxv.readBigUInt64LE(offset + 8));
    if (key === 0n) break;
    if (key === 6n) pageBytes = value;
    if (key === 17n) ticksPerSecond = value;
  }
  if (!(pageBytes > 0 && ticksPerSecond > 0)) throw new Error("unsupported Linux counter units");
  return { pageBytes, ticksPerSecond };
}
export function createLinuxResourceAdapter({ read = readFileSync, link = readlinkSync, proc = "/proc", cgroups = "/sys/fs/cgroup" } = {}) {
  let bytesRead = 0;
  const text = (file) => { const value = read(file, "utf8"); bytesRead += Buffer.byteLength(value); return value; };
  const boot = text(`${proc}/sys/kernel/random/boot_id`).trim();
  const units = linuxUnits(read(`${proc}/self/auxv`));
  const digest = (value) => `sha256:${createHash("sha256").update(value).digest("hex")}`;
  const proof = (pid) => {
    const stat = parseProcessStat(text(`${proc}/${pid}/stat`));
    const namespace = link(`${proc}/${pid}/ns/pid`);
    return { pid, start: stat.start, identity: digest(`${boot}:${namespace}:${pid}:${stat.start}`), stat };
  };
  function processSample(identity) {
    const before = proof(identity.pid);
    if (before.identity !== identity.identity || ["Z", "X"].includes(before.stat.state)) return { availability: "process_gone", metrics: {} };
    const stat = before.stat;
    const metrics = {
      cpu_user_us: observed(stat.user_ticks * 1e6 / units.ticksPerSecond),
      cpu_system_us: observed(stat.system_ticks * 1e6 / units.ticksPerSecond),
      rss_bytes: observed(stat.rss_pages * units.pageBytes), threads: observed(stat.threads),
    };
    try {
      const counters = parseCounters(text(`${proc}/${identity.pid}/io`));
      for (const [field, counter] of Object.entries({ io_read_bytes: "read_bytes", io_write_bytes: "write_bytes", io_read_chars: "rchar", io_write_chars: "wchar" })) metrics[field] = observed(counters[counter]);
    } catch (error) {
      for (const field of ["io_read_bytes", "io_write_bytes", "io_read_chars", "io_write_chars"]) metrics[field] = unavailable(availabilityFor(error));
    }
    try { metrics.rss_high_water_bytes = observed(parseCounters(text(`${proc}/${identity.pid}/status`)).VmHWM * 1024); }
    catch (error) { metrics.rss_high_water_bytes = unavailable(availabilityFor(error)); }
    if (proof(identity.pid).identity !== identity.identity) return { availability: "process_gone", metrics: {} };
    return { availability: "available", metrics };
  }
  function context() {
    const contextAvailability = (error) => availabilityFor(error) === "process_gone" ? "unsupported" : availabilityFor(error);
    const records = [];
    for (const resource of ["cpu", "memory", "io"]) {
      let metrics;
      try { metrics = Object.fromEntries(Object.entries(parsePSI(text(`${proc}/pressure/${resource}`))).map(([name, value]) => [`psi_${name}`, observed(value)])); }
      catch (error) { metrics = { psi_some_us: unavailable(contextAvailability(error)), psi_full_us: unavailable(contextAvailability(error)) }; }
      records.push({ scope_ref: `guest:${resource}`, scope: "guest_context", identity_digest: null, metrics });
    }
    const member = text(`${proc}/self/cgroup`).split("\n").find((line) => line.startsWith("0::"))?.slice(3);
    if (!member || member.split("/").includes("..")) return records;
    let directory = path.join(cgroups, member);
    for (let depth = 0; depth < 16; depth += 1) {
      const metrics = {};
      const mappings = {
        "cpu.stat": { usage_usec: "cgroup_cpu_us", throttled_usec: "cgroup_throttled_us", nr_throttled: "cgroup_throttled_periods" },
        "memory.events": { high: "cgroup_memory_high_events", max: "cgroup_memory_max_events", oom: "cgroup_oom_events", oom_kill: "cgroup_oom_kill_events" },
      };
      for (const [file, mapping] of Object.entries(mappings)) {
        let counters, error;
        try { counters = parseCounters(text(path.join(directory, file))); } catch (failure) { error = failure; }
        for (const [source, target] of Object.entries(mapping)) metrics[target] = error ? unavailable(contextAvailability(error)) : observed(counters[source]);
      }
      for (const [file, name] of [["memory.current", "cgroup_memory_bytes"], ["memory.max", "cgroup_memory_limit_bytes"], ["pids.current", "cgroup_pids"], ["pids.max", "cgroup_pids_limit"]]) {
        try { const value = text(path.join(directory, file)).trim(); metrics[name] = value === "max" ? unavailable("not_observed") : observed(Number(value)); }
        catch (error) { metrics[name] = unavailable(contextAvailability(error)); }
      }
      try {
        const [quota, period] = text(path.join(directory, "cpu.max")).trim().split(/\s+/u);
        metrics.cgroup_cpu_quota_cores = quota === "max" ? unavailable("not_observed") : observed(Number(quota) / Number(period));
      } catch (error) { metrics.cgroup_cpu_quota_cores = unavailable(contextAvailability(error)); }
      const identity = digest(`${boot}:${directory}`);
      records.push({ scope_ref: `cgroup:${depth}`, scope: "cgroup_context", identity_digest: identity, metrics });
      if (directory === cgroups) break;
      directory = path.dirname(directory);
    }
    return records;
  }
  function discover(maximum, deadline) {
    const found = [], directory = opendirSync(proc);
    let entries = 0, truncated = false;
    try {
      for (let entry; (entry = directory.readSync()) !== null;) {
        if (++entries > maximum || performance.now() >= deadline) { truncated = true; break; }
        if (!/^\d+$/u.test(entry.name)) continue;
        try { found.push(proof(Number(entry.name))); } catch { /* Vanishing processes are expected. */ }
      }
    } finally { directory.closeSync(); }
    return { found, truncated };
  }
  return { proof, processSample, context, discover, get bytesRead() { return bytesRead; } };
}
