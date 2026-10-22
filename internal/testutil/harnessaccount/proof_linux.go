//go:build linux

package harnessaccount

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

var scopeName = regexp.MustCompile(`^cartulary-account-[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}\.scope$`)

func accountingRoot(scope string) string {
	return filepath.Join("/run/user", strconv.Itoa(os.Getuid()), "cartulary-invocation-accounting", scope)
}

func createPrivateRoot(scope string) (string, error) {
	if !scopeName.MatchString(scope) {
		return "", errAccounting
	}
	root := accountingRoot(scope)
	parent, err := harnesslocal.PrivateDirectory(filepath.Dir(root))
	if err != nil {
		return "", errAccounting
	}
	defer parent.Close()
	if err = os.Mkdir(root, 0700); err != nil {
		return "", errAccounting
	}
	return root, nil
}

func writeBinding(root string, g *group) error {
	id, err := harnessdiag.InvocationIdentity(filepath.Join(root, "invocation"))
	if err != nil {
		return errAccounting
	}
	stat, err := os.Stat(filepath.Join("/sys/fs/cgroup", g.path))
	if err != nil {
		return errAccounting
	}
	identity, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return errAccounting
	}
	value := map[string]string{"invocation_digest": id, "parent": g.path, "device": strconv.FormatUint(uint64(identity.Dev), 10), "inode": strconv.FormatUint(identity.Ino, 10)}
	data, err := json.Marshal(value)
	if err != nil {
		return errAccounting
	}
	return harnesslocal.PublishNew(filepath.Join(root, "binding.json"), data)
}

// ContainerParent returns only the current invocation's live fenced parent.
// The kernel membership selects one exact private record, without environment
// inheritance or historical discovery. The caller must already be a descendant.
func ContainerParent() (string, error) {
	member, err := ownMembership()
	if err != nil {
		return "", err
	}
	return containerParentForMember(member)
}

func containerParentForMember(member string) (string, error) {
	parts := strings.Split(member, "/")
	position := -1
	for i, part := range parts {
		if strings.HasPrefix(part, "cartulary-account-") {
			if position != -1 || !scopeName.MatchString(part) {
				return "", errAccounting
			}
			position = i
		}
	}
	if position == -1 {
		return "", nil
	}
	if len(parts) <= position+2 || parts[position+1] != "measured" || parts[position+2] != "commands" {
		return "", errAccounting
	}
	root := accountingRoot(parts[position])
	data, err := harnesslocal.ReadFile(filepath.Join(root, "binding.json"), 4096, false)
	if err != nil {
		return "", errAccounting
	}
	value, err := harnessdiag.Decode(data)
	if err != nil {
		return "", errAccounting
	}
	v, ok := value.(map[string]any)
	if !ok || len(v) != 4 {
		return "", errAccounting
	}
	parent, ok := v["parent"].(string)
	if !ok || parent != strings.Join(parts[:position+2], "/") {
		return "", errAccounting
	}
	id, err := harnessdiag.InvocationIdentity(filepath.Join(root, "invocation"))
	if err != nil || v["invocation_digest"] != id {
		return "", errAccounting
	}
	stat, err := os.Stat(filepath.Join("/sys/fs/cgroup", parent))
	if err != nil {
		return "", errAccounting
	}
	s, ok := stat.Sys().(*syscall.Stat_t)
	if !ok || v["device"] != strconv.FormatUint(uint64(s.Dev), 10) || v["inode"] != strconv.FormatUint(s.Ino, 10) {
		return "", errAccounting
	}
	return parent, nil
}

func InvocationRoot() (string, error) {
	parent, err := ContainerParent()
	if err != nil || parent == "" {
		return "", err
	}
	return filepath.Join(accountingRoot(filepath.Base(filepath.Dir(parent))), "invocation"), nil
}

func sameExecutable(left, right string) bool {
	read := func(file string) ([]byte, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		h := sha256.New()
		if _, err = io.Copy(h, f); err != nil {
			return nil, err
		}
		return h.Sum(nil), nil
	}
	a, e := read(left)
	b, f := read(right)
	return e == nil && f == nil && bytes.Equal(a, b)
}

// A Make prerequisite may rebuild the native runtime after the outer
// controller started. Do not bind measurements from stale executable code to
// the new graph's source identity. Other service/test binaries are unaffected.
func VerifyRuntimeExecutable() error {
	root, err := InvocationRoot()
	if err != nil || root == "" {
		return errAccounting
	}
	id, err := harnessdiag.InvocationIdentity(root)
	if err != nil {
		return errAccounting
	}
	data, err := harnesslocal.ReadFile(filepath.Join(root, "owner.json"), 65536, false)
	if err != nil {
		return errAccounting
	}
	owner, err := harnessdiag.Decode(data)
	if err != nil {
		return errAccounting
	}
	value, ok := owner.(map[string]any)
	if !ok {
		return errAccounting
	}
	pid, ok := value["pid"].(float64)
	if !ok || !sameExecutable("/proc/self/exe", fmt.Sprintf("/proc/%.0f/exe", pid)) {
		return errAccounting
	}
	current, err := harnessdiag.InvocationIdentity(root)
	if err != nil || current != id {
		return errAccounting
	}
	return nil
}

func VerifyContainer(parent, id string) error {
	if parent == "" {
		return nil
	}
	if len(id) != 64 || strings.IndexFunc(id, func(r rune) bool { return !(r >= 'a' && r <= 'f' || r >= '0' && r <= '9') }) >= 0 {
		return errAccounting
	}
	current, err := ContainerParent()
	if err != nil || current != parent {
		return errAccounting
	}
	root, err := os.OpenRoot(filepath.Join("/sys/fs/cgroup", parent, id))
	if err != nil {
		return errAccounting
	}
	defer root.Close()
	populated, err := readCounter(root, "cgroup.events", "populated")
	if err != nil || populated != 1 {
		return errAccounting
	}
	return nil
}
