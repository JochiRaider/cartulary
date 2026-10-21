//go:build linux

package harnesslocal

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestSecureReadReplacement(t *testing.T) {
	old, newBytes := bytes.Repeat([]byte("a"), 4096), bytes.Repeat([]byte("b"), 4096)
	// Freeze the ctime-only linked-inode transition from the contention
	// rejection, independent of filesystem timestamp resolution or scheduling.
	t.Run("linked_snapshot_ctime_transition", func(t *testing.T) {
		before := syscall.Stat_t{Ino: 1, Dev: 1, Mode: syscall.S_IFREG | 0600,
			Uid: uint32(os.Getuid()), Nlink: 1, Size: 4096,
			Mtim: syscall.Timespec{Sec: 1}, Ctim: syscall.Timespec{Sec: 1}}
		after := before
		after.Ctim.Sec++
		if err := validateReadTransition(before, after, 4096, true); err != nil {
			var rejection *ReadRejection
			if errors.As(err, &rejection) {
				t.Fatalf("snapshot rejected stage=%s reason=%s", rejection.Stage(), rejection.Reason())
			}
			t.Fatal(err)
		}
		if !errors.Is(validateReadTransition(before, after, 4096, false), ErrUnsafe) {
			t.Fatal("immutable evidence admitted ctime transition")
		}
		for name, mutate := range map[string]func(*syscall.Stat_t){
			"owner":     func(s *syscall.Stat_t) { s.Uid++ },
			"mode":      func(s *syscall.Stat_t) { s.Mode = syscall.S_IFREG | 0644 },
			"inode":     func(s *syscall.Stat_t) { s.Ino++ },
			"device":    func(s *syscall.Stat_t) { s.Dev++ },
			"size":      func(s *syscall.Stat_t) { s.Size++ },
			"mtime":     func(s *syscall.Stat_t) { s.Mtim.Sec++ },
			"hard_link": func(s *syscall.Stat_t) { s.Nlink++ },
		} {
			t.Run(name, func(t *testing.T) {
				changed := after
				mutate(&changed)
				for _, snapshot := range []bool{false, true} {
					if !errors.Is(validateReadTransition(before, changed, 4096, snapshot), ErrUnsafe) {
						t.Fatal("unsafe metadata transition admitted")
					}
				}
			})
		}
		for _, links := range [][2]uint64{{1, 1}, {1, 0}, {0, 0}, {0, 1}} {
			before.Nlink, after.Nlink = links[0], links[1]
			err := validateReadTransition(before, after, 4096, true)
			if (err == nil) != (links != [2]uint64{0, 1}) {
				t.Fatal("incorrect snapshot link transition", links)
			}
		}
	})
	t.Run("concurrent_publication", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Chmod(root, 0700); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(root, "snapshot")
		if err := PublishNew(file, old); err != nil {
			t.Fatal(err)
		}
		start, stop := make(chan struct{}), make(chan struct{})
		var workers sync.WaitGroup
		var failed sync.Once
		reject := func(role string, err error) {
			failed.Do(func() {
				close(stop)
				var rejection *ReadRejection
				if errors.As(err, &rejection) {
					t.Errorf("%s rejected stage=%s reason=%s", role, rejection.Stage(), rejection.Reason())
				} else {
					t.Errorf("%s failed outside secure-read boundary", role)
				}
			})
		}
		for worker := range 5 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				limit := 4096
				if worker == 0 {
					limit = 512
				}
				for range limit {
					select {
					case <-stop:
						return
					default:
					}
					if worker == 0 {
						if err := ReplaceFile(file, newBytes); err != nil {
							reject("publisher", err)
							return
						}
						continue
					}
					data, err := ReadFile(file, 8192, true)
					if err != nil || (!bytes.Equal(data, old) && !bytes.Equal(data, newBytes)) {
						reject("reader", err)
						return
					}
				}
			}()
		}
		close(start)
		workers.Wait()
	})
	for _, stage := range []readCheckpoint{readBeforeOpen, readAfterOpen, readAfterMetadata, readDuringBytes, readBeforeFinalMetadata} {
		for _, mutable := range []bool{false, true} {
			t.Run(string(rune('0'+stage))+map[bool]string{false: "-immutable", true: "-snapshot"}[mutable], func(t *testing.T) {
				root := t.TempDir()
				if err := os.Chmod(root, 0700); err != nil {
					t.Fatal(err)
				}
				file := filepath.Join(root, "snapshot")
				if err := PublishNew(file, old); err != nil {
					t.Fatal(err)
				}
				dir, err := OpenDirectory(root, true)
				if err != nil {
					t.Fatal(err)
				}
				defer dir.Close()
				fired := false
				data, err := readFileAt(dir, "snapshot", 8192, mutable, func(at readCheckpoint) {
					if at != stage {
						return
					}
					fired = true
					for range 3 {
						if err := ReplaceFile(file, newBytes); err != nil {
							t.Fatal(err)
						}
					}
				})
				if !fired {
					t.Fatal("barrier was not reached")
				}
				if !mutable && stage != readBeforeOpen {
					if !errors.Is(err, ErrUnsafe) {
						t.Fatal("immutable evidence admitted replacement", err)
					}
					return
				}
				want := old
				if stage == readBeforeOpen {
					want = newBytes
				}
				if err != nil || !bytes.Equal(data, want) {
					t.Fatal("descriptor mixed or rejected complete snapshots", err)
				}
			})
		}
	}
	for _, mutation := range []string{"overwrite", "truncate", "mode", "link"} {
		t.Run(mutation, func(t *testing.T) {
			root := t.TempDir()
			file := filepath.Join(root, "snapshot")
			if err := os.WriteFile(file, old, 0600); err != nil {
				t.Fatal(err)
			}
			// A fixed mtime makes mutation observable without depending on clock resolution.
			if err := os.Chtimes(file, time.Unix(1, 0), time.Unix(1, 0)); err != nil {
				t.Fatal(err)
			}
			dir, err := OpenDirectory(root, false)
			if err != nil {
				t.Fatal(err)
			}
			defer dir.Close()
			_, err = readFileAt(dir, "snapshot", 8192, true, func(at readCheckpoint) {
				if at != readDuringBytes {
					return
				}
				var err error
				switch mutation {
				case "overwrite":
					err = os.WriteFile(file, newBytes, 0600)
				case "truncate":
					err = os.Truncate(file, 2)
				case "mode":
					err = os.Chmod(file, 0644)
				case "link":
					err = os.Link(file, filepath.Join(root, "alias"))
				}
				if err != nil {
					t.Fatal(err)
				}
			})
			var rejection *ReadRejection
			if !errors.Is(err, ErrUnsafe) || !errors.As(err, &rejection) || rejection.Stage() != "final_metadata" {
				t.Fatal("mutation admitted or lost rejection stage", err)
			}
		})
	}
}
