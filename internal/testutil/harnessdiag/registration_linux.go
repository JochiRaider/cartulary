//go:build linux

package harnessdiag

import (
	"os"
	"strings"
)

func clockCompatible(owner object) bool {
	boot, e := os.ReadFile("/proc/sys/kernel/random/boot_id")
	ns, f := os.Readlink("/proc/self/ns/time")
	return e == nil && f == nil && strings.TrimSpace(string(boot)) == owner["boot_id"] && ns == owner["namespace"]
}
