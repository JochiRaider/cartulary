//go:build linux

package harnessruntime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
)

func nativeAdmission(capacities map[string]int64, parent *string, root string) func(context.Context, unit) (*hostLease, error) {
	if root == "" {
		root = filepath.Join(os.TempDir(), "cartulary-host-admission-"+strconv.Itoa(os.Getuid()))
	}
	return func(ctx context.Context, u unit) (*hostLease, error) {
		callerContext := ctx
		ctx, cancel := context.WithTimeout(ctx, time.Duration(u.TimeoutMS)*time.Millisecond)
		defer cancel()
		proof, err := harnessadmission.Identity(os.Getpid())
		if err != nil {
			return nil, failed("infra", "resource_conflict")
		}
		var nonce [16]byte
		if _, err := rand.Read(nonce[:]); err != nil {
			return nil, err
		}
		token := hex.EncodeToString(nonce[:])
		claims, available := harnessadmission.Quantities{}, harnessadmission.Quantities{}
		for _, key := range []string{"browser_stack", "cpu", "io", "memory_mb", "process"} {
			claims[key] = float64(u.Claims[key])
			available[key] = float64(capacities[key])
		}
		mode := "shared"
		if slices.Contains(u.ExclusiveLocks, "host_activity") {
			mode = "exclusive"
		}
		r := harnessadmission.Request{Root: root, Operation: "admit", Token: token, Owner: proof, Mode: mode, Claims: claims, Capacities: available, Parent: parent}
		var mu sync.Mutex
		released := false
		release := func(ctx context.Context) error {
			mu.Lock()
			defer mu.Unlock()
			if released {
				return nil
			}
			_, err := harnessadmission.Transact(ctx, harnessadmission.Request{Root: root, Operation: "release", Token: token})
			if err == nil {
				released = true
			}
			return err
		}
		for {
			admitted, err := harnessadmission.Transact(ctx, r)
			if err == nil && admitted && ctx.Err() == nil {
				return &hostLease{Token: token, Release: release, Bind: func(pid int) error {
					guard, err := harnessadmission.Identity(pid)
					if err != nil {
						return err
					}
					// The executor retains a terminal leader until group drainage. A
					// terminal child needs no live lease guard and cannot be recycled.
					alive, err := harnessadmission.Alive(guard)
					if err != nil || !alive {
						return err
					}
					bindCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
					defer done()
					_, err = harnessadmission.Transact(bindCtx, harnessadmission.Request{Root: root, Operation: "bind", Token: token, Guard: guard})
					return err
				}}, nil
			}
			if err != nil || ctx.Err() != nil {
				cleanupCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
				cleanupErr := release(cleanupCtx)
				done()
				if cleanupErr != nil {
					return nil, failed("harness", "cleanup_error")
				}
				if callerContext.Err() != nil {
					return nil, callerContext.Err()
				}
				return nil, failed("infra", "resource_conflict")
			}
			timer := time.NewTimer(50 * time.Millisecond)
			select {
			case <-ctx.Done():
			case <-timer.C:
			}
			timer.Stop()
		}
	}
}
