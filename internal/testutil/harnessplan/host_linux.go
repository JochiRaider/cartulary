//go:build linux

package harnessplan

import (
	"bufio"
	"bytes"
	"os"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// DetectHost reads the host facts capacity capture depends on.
func DetectHost(root string) Host {
	host := Host{CPUs: int64(max(1, runtime.NumCPU())), MemoryBytes: memTotal(), CgroupMemory: positiveFile("/sys/fs/cgroup/memory.max"), WritableRoot: unix.Access(root, unix.W_OK) == nil}
	cgroup, user := positiveFile("/sys/fs/cgroup/pids.max"), userProcessLimit()
	switch {
	case cgroup > 0 && user > 0:
		host.ProcessLimit = min(cgroup, user)
	case cgroup > 0:
		host.ProcessLimit = cgroup
	default:
		host.ProcessLimit = user
	}
	return host
}

func positiveFile(path string) int64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	value, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil || value < 1 {
		return 0
	}
	return value
}

func memTotal() int64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 1
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[0] == "MemTotal:" {
			if kib, err := strconv.ParseInt(fields[1], 10, 64); err == nil && kib > 0 {
				return kib * 1024
			}
		}
	}
	return 1
}

func userProcessLimit() int64 {
	data, err := os.ReadFile("/proc/self/limits")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "Max processes") {
			continue
		}
		fields := strings.Fields(strings.TrimPrefix(line, "Max processes"))
		if len(fields) == 0 || fields[0] == "unlimited" {
			return 0
		}
		value, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil || value < 1 {
			return 0
		}
		return value
	}
	return 0
}

// DetectHostProfile identifies the host a plan is prepared on: operating
// system, architecture, kernel release and the distinct CPU models.
func DetectHostProfile() HostProfile {
	profile := HostProfile{Platform: runtime.GOOS, Architecture: runtime.GOARCH, KernelRelease: "unknown"}
	var name unix.Utsname
	if unix.Uname(&name) == nil {
		if release := unix.ByteSliceToString(name.Release[:]); release != "" {
			profile.KernelRelease = release
		}
	}
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			key, model, ok := strings.Cut(scanner.Text(), ":")
			if ok && strings.TrimSpace(key) == "model name" && strings.TrimSpace(model) != "" {
				profile.CPUModels = append(profile.CPUModels, strings.TrimSpace(model))
			}
		}
	}
	profile.CPUModels = dedupe(profile.CPUModels)
	if len(profile.CPUModels) == 0 {
		profile.CPUModels = []string{"unknown"}
	}
	return profile
}
