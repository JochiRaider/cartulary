//go:build linux

package harnessplan

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// GraphCacheDirectory is the checkout's graph cache, machine state that the
// planner keeps its memos in (P6-5). A memo only saves work: every entry is
// checked against the file it describes before it is used.
const GraphCacheDirectory = ".cache/cartulary/graph-v2"

const (
	digestMemoName     = "source-digests"
	digestMemoSchemaID = "cartulary.harness_source_digest_memo.v1"
)

// settleMargin is how much older than a capture a file's change time must be
// before its digest is kept: a later change then always moves the change
// time, which a file's owner cannot set back. It covers the kernel's coarse
// timestamp clock with room to spare.
var settleMargin = 2 * time.Second

// fileIdentity is what a stat says about a regular file's content: the same
// identity after a settled capture means the same bytes.
type fileIdentity struct {
	dev, ino     uint64
	mode         uint32
	size         int64
	mtime, ctime int64
}

func identityOf(stat unix.Stat_t) fileIdentity {
	return fileIdentity{dev: stat.Dev, ino: stat.Ino, mode: stat.Mode, size: stat.Size, mtime: stat.Mtim.Nano(), ctime: stat.Ctim.Nano()}
}

type digestEntry struct {
	identity fileIdentity
	digest   string
}

// digestMemo keeps each settled regular file's content digest by path, the
// way Git's index keeps stat data: a file whose stat identity is unchanged
// since a capture that hashed it after it settled has the digest recorded.
type digestMemo struct {
	path    string
	started int64
	mu      sync.Mutex
	known   map[string]digestEntry
	next    map[string]digestEntry
	changed bool
}

// openDigestMemo returns the checkout's memo, or nil when Git does not ignore
// the graph cache directory: a memo there would become a source input itself.
func openDigestMemo(root string) *digestMemo {
	file := filepath.Join(GraphCacheDirectory, digestMemoName)
	if _, err := git(root, "check-ignore", "-q", "--no-index", "--", file); err != nil {
		return nil
	}
	memo := &digestMemo{path: filepath.Join(root, file), started: time.Now().UnixNano(), known: map[string]digestEntry{}, next: map[string]digestEntry{}}
	data, err := os.ReadFile(memo.path)
	if err != nil || !bytes.HasPrefix(data, []byte(digestMemoSchemaID+"\x00")) {
		return memo
	}
	records := bytes.Split(data[len(digestMemoSchemaID)+1:], []byte{0})
	for i := 0; i+1 < len(records); i += 2 {
		fields := strings.Fields(string(records[i+1]))
		if len(fields) != 7 {
			return &digestMemo{path: memo.path, started: memo.started, known: map[string]digestEntry{}, next: map[string]digestEntry{}}
		}
		var numbers [6]int64
		for j := range numbers {
			if numbers[j], err = strconv.ParseInt(fields[j], 10, 64); err != nil {
				return &digestMemo{path: memo.path, started: memo.started, known: map[string]digestEntry{}, next: map[string]digestEntry{}}
			}
		}
		memo.known[string(records[i])] = digestEntry{identity: fileIdentity{dev: uint64(numbers[0]), ino: uint64(numbers[1]), mode: uint32(numbers[2]), size: numbers[3], mtime: numbers[4], ctime: numbers[5]}, digest: fields[6]}
	}
	return memo
}

// lookup returns the recorded digest of a file with this identity.
func (m *digestMemo) lookup(relative string, identity fileIdentity) (string, bool) {
	if m == nil {
		return "", false
	}
	entry, ok := m.known[relative]
	if !ok || entry.identity != identity {
		return "", false
	}
	m.mu.Lock()
	m.next[relative] = entry
	m.mu.Unlock()
	return entry.digest, true
}

// remember keeps a digest that was just hashed when its file had settled
// before this capture began.
func (m *digestMemo) remember(relative string, identity fileIdentity, digest string) {
	if m == nil || identity.ctime >= m.started-settleMargin.Nanoseconds() || identity.mtime >= m.started-settleMargin.Nanoseconds() {
		return
	}
	m.mu.Lock()
	m.next[relative], m.changed = digestEntry{identity: identity, digest: digest}, true
	m.mu.Unlock()
}

// save writes the entries this capture used or added, when they differ from
// what it read. A memo that cannot be written is left as it was.
func (m *digestMemo) save() {
	if m == nil || (!m.changed && len(m.next) == len(m.known)) {
		return
	}
	var out bytes.Buffer
	out.WriteString(digestMemoSchemaID + "\x00")
	for relative, entry := range m.next {
		id := entry.identity
		fmt.Fprintf(&out, "%s\x00%d %d %d %d %d %d %s\x00", relative, id.dev, id.ino, id.mode, id.size, id.mtime, id.ctime, entry.digest)
	}
	if os.MkdirAll(filepath.Dir(m.path), 0o755) == nil {
		_ = writeAtomic(m.path, out.Bytes())
	}
}
