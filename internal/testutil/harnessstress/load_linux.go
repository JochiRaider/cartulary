//go:build linux

package harnessstress

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"golang.org/x/sys/unix"
)

// loadProfile is one declared synthetic host load (TH-HARNESS-REQ-871).
// CPU load is busy threads per logical CPU; disk load is writers that each
// append a block and wait for it to reach stable storage, the flush pattern
// PostgreSQL commits produce.
type loadProfile struct {
	name             string
	threadsPerCPU    int
	flushWriters     int
	flushBlockBytes  int
	flushFileMaxSize int64
}

var profiles = map[string]loadProfile{
	"none":  {name: "none"},
	"cpu":   {name: "cpu", threadsPerCPU: 2},
	"io":    {name: "io", flushWriters: 4, flushBlockBytes: 16 << 10, flushFileMaxSize: 64 << 20},
	"mixed": {name: "mixed", threadsPerCPU: 2, flushWriters: 4, flushBlockBytes: 16 << 10, flushFileMaxSize: 64 << 20},
}

func (p loadProfile) document() map[string]any {
	return map[string]any{"cpu_threads": p.threadsPerCPU * runtime.NumCPU(), "flush_writers": p.flushWriters, "flush_block_bytes": p.flushBlockBytes}
}

// LoadMain runs a load generator until it is killed. It is started only by
// Main, in its own process group.
func LoadMain(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("stress-load", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("profile", "", "load profile")
	directory := flags.String("directory", "", "private directory for flush files")
	if flags.Parse(args) != nil || flags.NArg() != 0 {
		return 2
	}
	profile, ok := profiles[*name]
	if !ok || profile.flushWriters > 0 && *directory == "" {
		return 2
	}
	threads := profile.threadsPerCPU * runtime.NumCPU()
	runtime.GOMAXPROCS(max(1, threads+profile.flushWriters))
	var wait sync.WaitGroup
	for range threads {
		wait.Add(1)
		go func() {
			defer wait.Done()
			spin()
		}()
	}
	for writer := range profile.flushWriters {
		wait.Add(1)
		go func() {
			defer wait.Done()
			flush(filepath.Join(*directory, "flush-"+strconv.Itoa(writer)), profile.flushBlockBytes, profile.flushFileMaxSize)
		}()
	}
	wait.Wait()
	return 0
}

//go:noinline
func spin() {
	x := uint64(1)
	for {
		x = x*6364136223846793005 + 1442695040888963407
		if x == 0 {
			runtime.Gosched()
		}
	}
}

// flush appends a block and calls fdatasync, restarting the file at its
// size cap so disk use stays bounded. It returns only on a write error.
func flush(path string, block int, limit int64) {
	data := make([]byte, block)
	for i := range data {
		data[i] = byte(i)
	}
	for {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
		if err != nil {
			return
		}
		for written := int64(0); written < limit; written += int64(block) {
			if _, err := file.Write(data); err != nil {
				file.Close()
				return
			}
			if err := unix.Fdatasync(int(file.Fd())); err != nil {
				file.Close()
				return
			}
		}
		file.Close()
	}
}
