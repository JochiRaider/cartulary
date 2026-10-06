import os from "node:os";
import {readFileSync} from "node:fs";

// Read-only qualification of the selected Windows/WSL/Docker Desktop boundary.
// Formal CMVP applicability is a separate disposition, never inferred here.
export function referencePlatform(command) {
 const docker = (...args) => command("docker", args);
  const info = JSON.parse(docker("info", "--format", "{{json .}}"));
  if (info.NCPU < 2 || !/microsoft.*wsl2/i.test(os.release())) throw new Error("assessment requires the Windows 11 / WSL2 target and two logical CPUs");
  const windows = JSON.parse(command("powershell.exe", ["-NoProfile", "-NonInteractive", "-Command", "$o=Get-CimInstance Win32_OperatingSystem; $c=Get-CimInstance Win32_Processor; $d=Get-ItemProperty 'HKLM:\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion'; [pscustomobject]@{caption=$o.Caption;build=$o.BuildNumber;ubr=$d.UBR;cpu=$c.Name;load_percent=$c.LoadPercentage}|ConvertTo-Json -Compress"]));
  if (!/Windows 11/.test(windows.caption)) throw new Error("assessment requires Windows 11");
  const wsl = command("wsl.exe", ["--version"], { encoding: "utf16le" });
  const desktop = JSON.parse(command("powershell.exe", ["-NoProfile", "-NonInteractive", "-Command", "$v=(Get-Item \"$env:ProgramFiles\\Docker\\Docker\\Docker Desktop.exe\").VersionInfo.ProductVersion; $s=Get-Content \"$env:APPDATA\\Docker\\settings-store.json\" -Raw|ConvertFrom-Json; [pscustomobject]@{application_version=$v;wsl_engine_enabled=$s.wslEngineEnabled}|ConvertTo-Json -Compress"]));
  desktop.wsl_distribution = command("wsl.exe", ["--list", "--verbose"], { encoding: "utf16le" }).split(/\r?\n/).find((line) => /docker-desktop\s/.test(line))?.trim();
  if (info.Name !== "docker-desktop" || info.OperatingSystem !== "Docker Desktop" || desktop.wsl_engine_enabled === false || !/docker-desktop\s+Running\s+2$/.test(desktop.wsl_distribution || "") || !/microsoft/i.test(info.KernelVersion)) throw new Error("assessment requires Docker Desktop's running WSL2 backend");
  const filesystem = JSON.parse(command("findmnt", ["--json", "--target", process.cwd(), "--output", "FSTYPE,SOURCE,TARGET,OPTIONS"]));
  if (!filesystem.filesystems?.length || filesystem.filesystems.some(({fstype}) => fstype !== "ext4")) throw new Error("qualification requires the Ubuntu guest filesystem");
  const guest = readFileSync("/etc/os-release", "utf8");
  if (!/^ID=ubuntu$/m.test(guest)) throw new Error("qualification requires Ubuntu");
  const networking_mode = command("wslinfo", ["--networking-mode"]);
  if (!["nat", "mirrored", "virtioproxy"].includes(networking_mode.toLowerCase())) throw new Error("unrecognized WSL network mode");
  const guest_memory_bytes = os.totalmem();
  const cpu_features = readFileSync("/proc/cpuinfo", "utf8").match(/^flags\s*:\s*(.+)$/m)?.[1]?.split(/\s+/);
  const docker_context = docker("context", "show");
  const engine_client = JSON.parse(docker("version", "--format", "{{json .Client}}"));
  return { filesystem, networking_mode, guest_memory_bytes, cpu_features, docker_context, engine_client, windows, wsl, guest: readFileSync("/etc/os-release", "utf8"), kernel: os.release(), cpus: os.cpus().map(({ model }) => model), load_before: os.loadavg(), docker: { id: info.ID, name: info.Name, operating_system: info.OperatingSystem, architecture: info.Architecture, cpu_count: info.NCPU, memory_bytes: info.MemTotal, server_version: info.ServerVersion, kernel: info.KernelVersion }, desktop, compose: docker("compose", "version") };

}
