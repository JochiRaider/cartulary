//go:build linux

// Package harnesscollect owns bounded optional resource observation. The
// scheduler owns registrations, lifetime and control; adapters own acquisition.
package harnesscollect

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type object = map[string]any
type proof struct {
	PID      int    `json:"pid"`
	Start    string `json:"start"`
	Identity string `json:"identity"`
	Stat     object `json:"stat"`
}
type linuxAdapter struct {
	boot        string
	page, ticks float64
	bytes       int64
}

func digest(data []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(data)) }
func monotonic() int64 {
	var t unix.Timespec
	if unix.ClockGettime(unix.CLOCK_MONOTONIC, &t) != nil {
		panic("monotonic clock unavailable")
	}
	return t.Nano()
}
func (a *linuxAdapter) read(name string) (string, error) {
	b, e := os.ReadFile(name)
	a.bytes += int64(len(b))
	return string(b), e
}
func newAdapter() (*linuxAdapter, error) {
	a := &linuxAdapter{}
	b, e := a.read("/proc/sys/kernel/random/boot_id")
	if e != nil {
		return nil, e
	}
	a.boot = strings.TrimSpace(b)
	data, e := os.ReadFile("/proc/self/auxv")
	if e != nil {
		return nil, e
	}
	// The adopted local observation profile is Linux x86_64.
	for i := 0; i+16 <= len(data); i += 16 {
		key, value := binary.LittleEndian.Uint64(data[i:]), binary.LittleEndian.Uint64(data[i+8:])
		if key == 0 {
			break
		}
		if key == 6 {
			a.page = float64(value)
		}
		if key == 17 {
			a.ticks = float64(value)
		}
	}
	if a.page <= 0 || a.ticks <= 0 {
		return nil, errors.New("unsupported counter units")
	}
	return a, nil
}
func parseStat(text string) (object, error) {
	end := strings.LastIndex(text, ") ")
	if end < 0 {
		return nil, errors.New("invalid stat")
	}
	f := strings.Fields(text[end+2:])
	if len(f) < 22 {
		return nil, errors.New("short stat")
	}
	result := object{"state": f[0], "start": f[19]}
	for index, key := range map[int]string{1: "parent", 11: "user_ticks", 12: "system_ticks", 17: "threads", 19: "start_value", 21: "rss_pages"} {
		n, e := strconv.ParseUint(f[index], 10, 64)
		if e != nil || n > 9007199254740991 {
			return nil, errors.New("invalid counter")
		}
		result[key] = float64(n)
	}
	delete(result, "start_value")
	return result, nil
}
func (a *linuxAdapter) proof(pid int) (proof, error) {
	text, e := a.read(fmt.Sprintf("/proc/%d/stat", pid))
	if e != nil {
		return proof{}, e
	}
	stat, e := parseStat(text)
	if e != nil {
		return proof{}, e
	}
	namespace, e := os.Readlink(fmt.Sprintf("/proc/%d/ns/pid", pid))
	if e != nil {
		return proof{}, e
	}
	start := stat["start"].(string)
	return proof{pid, start, digest([]byte(fmt.Sprintf("%s:%s:%d:%s", a.boot, namespace, pid, start))), stat}, nil
}
func availability(err error) string {
	if errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM) {
		return "permission_denied"
	}
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ESRCH) {
		return "process_gone"
	}
	return "unsupported"
}
func missing(why string) object { return object{"value": nil, "availability": why} }
func observed(n float64) object {
	if n < 0 || math.IsNaN(n) || math.IsInf(n, 0) {
		return missing("not_observed")
	}
	return object{"value": n, "availability": "available"}
}
func counters(text string) map[string]float64 {
	result := map[string]float64{}
	for _, line := range strings.Split(text, "\n") {
		f := strings.Fields(strings.Replace(line, ":", "", 1))
		if len(f) < 2 {
			continue
		}
		n, e := strconv.ParseUint(f[1], 10, 64)
		if e == nil && n <= 9007199254740991 {
			result[f[0]] = float64(n)
		}
	}
	return result
}
func counter(values map[string]float64, key string) float64 {
	if n, ok := values[key]; ok {
		return n
	}
	return math.NaN()
}
func (a *linuxAdapter) sample(p proof) (string, object) {
	before, e := a.proof(p.PID)
	if e != nil {
		return availability(e), object{}
	}
	if before.Identity != p.Identity || before.Stat["state"] == "Z" || before.Stat["state"] == "X" {
		return "process_gone", object{}
	}
	s := before.Stat
	metrics := object{"cpu_user_us": observed(s["user_ticks"].(float64) * 1e6 / a.ticks), "cpu_system_us": observed(s["system_ticks"].(float64) * 1e6 / a.ticks), "rss_bytes": observed(s["rss_pages"].(float64) * a.page), "threads": observed(s["threads"].(float64))}
	text, e := a.read(fmt.Sprintf("/proc/%d/io", p.PID))
	values := counters(text)
	for name, key := range map[string]string{"io_read_bytes": "read_bytes", "io_write_bytes": "write_bytes", "io_read_chars": "rchar", "io_write_chars": "wchar"} {
		if e != nil {
			metrics[name] = missing(availability(e))
		} else {
			metrics[name] = observed(counter(values, key))
		}
	}
	text, e = a.read(fmt.Sprintf("/proc/%d/status", p.PID))
	if e != nil {
		metrics["rss_high_water_bytes"] = missing(availability(e))
	} else {
		metrics["rss_high_water_bytes"] = observed(counter(counters(text), "VmHWM") * 1024)
	}
	after, e := a.proof(p.PID)
	if e != nil {
		return availability(e), object{}
	}
	if after.Identity != p.Identity {
		return "process_gone", object{}
	}
	return "available", metrics
}
func (a *linuxAdapter) discover(max int, deadline time.Time) ([]proof, bool, error) {
	dir, e := os.Open("/proc")
	if e != nil {
		return nil, false, e
	}
	defer dir.Close()
	found := []proof{}
	entries := 0
	for {
		names, e := dir.Readdirnames(1)
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, false, e
		}
		entries++
		if entries > max || !time.Now().Before(deadline) {
			return found, true, nil
		}
		pid, e := strconv.Atoi(names[0])
		if e != nil {
			continue
		}
		p, e := a.proof(pid)
		if e == nil {
			found = append(found, p)
		}
	}
	return found, false, nil
}
func contextAvailability(err error) string {
	why := availability(err)
	if why == "process_gone" {
		return "unsupported"
	}
	return why
}
func (a *linuxAdapter) context() ([]object, error) {
	records := []object{}
	for _, resource := range []string{"cpu", "memory", "io"} {
		text, e := a.read("/proc/pressure/" + resource)
		metrics := object{}
		if e != nil {
			metrics["psi_some_us"] = missing(contextAvailability(e))
			metrics["psi_full_us"] = missing(contextAvailability(e))
		} else {
			for _, line := range strings.Split(text, "\n") {
				f := strings.Fields(line)
				if len(f) < 2 || (f[0] != "some" && f[0] != "full") {
					continue
				}
				for _, part := range f[1:] {
					if strings.HasPrefix(part, "total=") {
						n, e := strconv.ParseUint(strings.TrimPrefix(part, "total="), 10, 64)
						if e == nil && n <= 9007199254740991 {
							metrics["psi_"+f[0]+"_us"] = observed(float64(n))
						}
					}
				}
			}
		}
		records = append(records, object{"scope": "guest_context", "scope_ref": "guest:" + resource, "identity_digest": nil, "metrics": metrics})
	}
	text, e := a.read("/proc/self/cgroup")
	if e != nil {
		return nil, e
	}
	member := ""
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "0::") {
			member = strings.TrimPrefix(line, "0::")
			break
		}
	}
	if member == "" {
		return records, nil
	}
	for _, part := range strings.Split(member, "/") {
		if part == ".." {
			return records, nil
		}
	}
	directory := filepath.Join("/sys/fs/cgroup", member)
	for depth := 0; depth < 16; depth++ {
		metrics := object{}
		for file, mapping := range map[string]map[string]string{"cpu.stat": {"usage_usec": "cgroup_cpu_us", "throttled_usec": "cgroup_throttled_us", "nr_throttled": "cgroup_throttled_periods"}, "memory.events": {"high": "cgroup_memory_high_events", "max": "cgroup_memory_max_events", "oom": "cgroup_oom_events", "oom_kill": "cgroup_oom_kill_events"}} {
			text, e := a.read(filepath.Join(directory, file))
			values := counters(text)
			for source, target := range mapping {
				if e != nil {
					metrics[target] = missing(contextAvailability(e))
				} else {
					metrics[target] = observed(counter(values, source))
				}
			}
		}
		for file, name := range map[string]string{"memory.current": "cgroup_memory_bytes", "memory.max": "cgroup_memory_limit_bytes", "pids.current": "cgroup_pids", "pids.max": "cgroup_pids_limit"} {
			text, e := a.read(filepath.Join(directory, file))
			if e != nil {
				metrics[name] = missing(contextAvailability(e))
			} else {
				n, e := strconv.ParseFloat(strings.TrimSpace(text), 64)
				if e != nil {
					metrics[name] = missing("not_observed")
				} else {
					metrics[name] = observed(n)
				}
			}
		}
		text, e := a.read(filepath.Join(directory, "cpu.max"))
		metrics["cgroup_cpu_quota_cores"] = missing("not_observed")
		if e != nil {
			metrics["cgroup_cpu_quota_cores"] = missing(contextAvailability(e))
		} else {
			f := strings.Fields(text)
			if len(f) == 2 {
				quota, qe := strconv.ParseFloat(f[0], 64)
				period, pe := strconv.ParseFloat(f[1], 64)
				if qe == nil && pe == nil {
					metrics["cgroup_cpu_quota_cores"] = observed(quota / period)
				}
			}
		}
		records = append(records, object{"scope": "cgroup_context", "scope_ref": fmt.Sprintf("cgroup:%d", depth), "identity_digest": digest([]byte(a.boot + ":" + directory)), "metrics": metrics})
		if directory == "/sys/fs/cgroup" {
			break
		}
		directory = filepath.Dir(directory)
	}
	return records, nil
}
