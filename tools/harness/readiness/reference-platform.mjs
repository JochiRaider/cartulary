import os from "node:os";
import {readFileSync} from "node:fs";

// Read-only record of the Linux/amd64 Docker host that runs package
// qualification. It queries only the guest and its Docker engine; no host
// operating system outside the guest is required or consulted. Formal CMVP
// applicability is a separate disposition, never inferred here.
export function referencePlatform(command) {
  const docker = (...args) => command("docker", args);
  const info = JSON.parse(docker("info", "--format", "{{json .}}"));
  if (info.NCPU < 2 || info.OSType !== "linux" || info.Architecture !== "x86_64") throw new Error("assessment requires a Linux x86_64 Docker engine with two logical CPUs");
  const filesystem = JSON.parse(command("findmnt", ["--json", "--target", process.cwd(), "--output", "FSTYPE,SOURCE,TARGET,OPTIONS"]));
  if (!filesystem.filesystems?.length || filesystem.filesystems.some(({fstype}) => fstype !== "ext4")) throw new Error("qualification requires the Ubuntu guest filesystem");
  const guest = readFileSync("/etc/os-release", "utf8");
  if (!/^ID=ubuntu$/m.test(guest)) throw new Error("qualification requires Ubuntu");
  const guest_memory_bytes = os.totalmem();
  const cpu_features = readFileSync("/proc/cpuinfo", "utf8").match(/^flags\s*:\s*(.+)$/m)?.[1]?.split(/\s+/);
  const docker_context = docker("context", "show");
  const engine_client = JSON.parse(docker("version", "--format", "{{json .Client}}"));
  const security_options = info.SecurityOptions;
  return { filesystem, guest_memory_bytes, cpu_features, docker_context, engine_client, guest, kernel: os.release(), cpus: os.cpus().map(({ model }) => model), load_before: os.loadavg(), docker: { id: info.ID, name: info.Name, operating_system: info.OperatingSystem, os_type: info.OSType, architecture: info.Architecture, cpu_count: info.NCPU, memory_bytes: info.MemTotal, server_version: info.ServerVersion, kernel: info.KernelVersion, security_options }, compose: docker("compose", "version") };
}
