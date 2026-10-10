import { createHash } from "node:crypto";
import { readFileSync, readlinkSync } from "node:fs";

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
// This proof is private. Only its opaque digest enters retained evidence.
export function processProof(pid, { read = (file) => readFileSync(file, "utf8"), link = readlinkSync, proc = "/proc", boot = read(`${proc}/sys/kernel/random/boot_id`).trim() } = {}) {
  if (!Number.isSafeInteger(pid) || pid < 1) throw new Error("invalid process identity");
  const stat = parseProcessStat(read(`${proc}/${pid}/stat`));
  const namespace = link(`${proc}/${pid}/ns/pid`);
  return { pid, start: stat.start, identity: `sha256:${createHash("sha256").update(`${boot}:${namespace}:${pid}:${stat.start}`).digest("hex")}`, stat };
}
