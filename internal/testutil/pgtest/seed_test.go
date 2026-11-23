package pgtest

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"golang.org/x/sys/unix"
)

const fakeSeedDigest = "d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1d1"

type fakeSeedDocker struct {
	images     map[string]image.InspectResponse
	inspects   int
	removed    []string
	containers []dockercontainer.Summary
	listed     []image.Summary
}

func (d *fakeSeedDocker) ImageInspect(_ context.Context, ref string, _ ...client.ImageInspectOption) (client.ImageInspectResult, error) {
	d.inspects++
	found, ok := d.images[ref]
	if !ok {
		return client.ImageInspectResult{}, errdefs.ErrNotFound
	}
	return client.ImageInspectResult{InspectResponse: found}, nil
}

func (d *fakeSeedDocker) ImageList(context.Context, client.ImageListOptions) (client.ImageListResult, error) {
	return client.ImageListResult{Items: d.listed}, nil
}

func (d *fakeSeedDocker) ImageRemove(_ context.Context, id string, _ client.ImageRemoveOptions) (client.ImageRemoveResult, error) {
	d.removed = append(d.removed, id)
	return client.ImageRemoveResult{}, nil
}

func (d *fakeSeedDocker) ContainerList(context.Context, client.ContainerListOptions) (client.ContainerListResult, error) {
	return client.ContainerListResult{Items: d.containers}, nil
}

func (d *fakeSeedDocker) ContainerRemove(_ context.Context, id string, options client.ContainerRemoveOptions) (client.ContainerRemoveResult, error) {
	if !options.Force || !options.RemoveVolumes {
		return client.ContainerRemoveResult{}, errors.New("builder reclaim must remove the container and its volumes")
	}
	d.removed = append(d.removed, id)
	return client.ContainerRemoveResult{}, nil
}

func (d *fakeSeedDocker) ContainerInspect(context.Context, string, client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	return client.ContainerInspectResult{}, errors.New("not used")
}

func (d *fakeSeedDocker) ContainerCommit(context.Context, string, client.ContainerCommitOptions) (client.ContainerCommitResult, error) {
	return client.ContainerCommitResult{}, errors.New("not used")
}

// seedImage decodes the daemon's wire form, as the client does.
func seedImage(id, digest string) image.InspectResponse {
	var response image.InspectResponse
	payload, err := json.Marshal(map[string]any{"Id": id, "Config": map[string]any{"Labels": map[string]string{seedDigestLabel: digest}}})
	if err != nil || json.Unmarshal(payload, &response) != nil {
		panic("invalid seed image fixture")
	}
	return response
}

func lockedBy(t *testing.T, mode int) bool {
	t.Helper()
	other, err := openSeedLock()
	if err != nil {
		t.Fatal(err)
	}
	defer other.close()
	return unix.Flock(int(other.file.Fd()), mode|unix.LOCK_NB) != nil
}

func TestPostgresSeedResolution(t *testing.T) {
	t.Run("generation_identity_keeps_the_published_format", func(t *testing.T) {
		pattern := regexp.MustCompile(`^[1-9][0-9]{0,19}$`)
		seen := map[string]bool{}
		for range 256 {
			identity, err := newGenerationIdentity()
			if err != nil || !pattern.MatchString(identity) || seen[identity] {
				t.Fatalf("identity %q: err=%v repeated=%v", identity, err, seen[identity])
			}
			if value, err := strconv.ParseUint(identity, 10, 64); err != nil || value > 1<<63-1 {
				t.Fatalf("identity %q exceeds 63 bits", identity)
			}
			seen[identity] = true
		}
	})
	t.Run("builder_runs_initdb_outside_declared_volumes", func(t *testing.T) {
		source, release, err := seedBuilderOperations("cartulary-postgres-seed-build-test").source(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		release()
		req := postgresContainerRequest(source)
		if source.seeded || req.Image != postgresImage || req.Name != "cartulary-postgres-seed-build-test" || req.ConfigModifier != nil {
			t.Fatalf("builder request = %#v", req)
		}
		if req.Env["PGDATA"] != seedDataDirectory || req.Env["POSTGRES_INITDB_ARGS"] != postgresInitdbArgs || req.Env["POSTGRES_USER"] != "cartulary" {
			t.Fatalf("builder initdb environment = %#v", req.Env)
		}
		withTLS, directory, err := preparePostgresTLS(req, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(directory) })
		if !slices.ContainsFunc(withTLS.Files, func(file testcontainers.ContainerFile) bool { return file.ContainerFilePath == postgresInitScriptPath }) {
			t.Fatal("builder lost the initdb script")
		}
		if !strings.HasPrefix(seedDataDirectory, seedVolumeRoot+"/") || strings.HasPrefix(seedDataDirectory, "/var/lib/postgresql") {
			t.Fatal("seed data directory must sit under the seed volume and outside the base image volume")
		}
	})
	t.Run("present_seed_is_returned_pinned_without_a_build", func(t *testing.T) {
		t.Setenv("TMPDIR", t.TempDir())
		lock, err := openSeedLock()
		if err != nil {
			t.Fatal(err)
		}
		defer lock.close()
		docker := &fakeSeedDocker{images: map[string]image.InspectResponse{seedReference(fakeSeedDigest): seedImage("sha256:aaaa", fakeSeedDigest)}}
		got, err := resolveSeedWith(t.Context(), docker, lock, fakeSeedDigest, func(context.Context, seedDocker, string) (string, error) {
			t.Fatal("present seed was rebuilt")
			return "", nil
		})
		if err != nil || got != "sha256:aaaa" {
			t.Fatal(got, err)
		}
		if lockedBy(t, unix.LOCK_SH) || !lockedBy(t, unix.LOCK_EX) {
			t.Fatal("returned seed is not held under the shared lock")
		}
	})
	for _, name := range []string{"absent", "mislabeled"} {
		t.Run(name+"_seed_is_built_once_then_verified", func(t *testing.T) {
			t.Setenv("TMPDIR", t.TempDir())
			lock, err := openSeedLock()
			if err != nil {
				t.Fatal(err)
			}
			defer lock.close()
			docker := &fakeSeedDocker{images: map[string]image.InspectResponse{}, listed: []image.Summary{
				{ID: "sha256:new", Created: 30, Labels: map[string]string{seedDigestLabel: fakeSeedDigest}},
				{ID: "sha256:old", Created: 10, Labels: map[string]string{seedDigestLabel: "other"}},
			}}
			if name == "mislabeled" {
				docker.images[seedReference(fakeSeedDigest)] = seedImage("sha256:wrong", "other")
			}
			builds := 0
			got, err := resolveSeedWith(t.Context(), docker, lock, fakeSeedDigest, func(ctx context.Context, d seedDocker, digest string) (string, error) {
				builds++
				if !lockedBy(t, unix.LOCK_SH) {
					t.Fatal("seed built without the exclusive lock")
				}
				docker.images[seedReference(digest)] = seedImage("sha256:new", digest)
				return "sha256:new", nil
			})
			if err != nil || got != "sha256:new" || builds != 1 {
				t.Fatal(got, err, builds)
			}
			if lockedBy(t, unix.LOCK_SH) || !lockedBy(t, unix.LOCK_EX) {
				t.Fatal("built seed was not downgraded to the shared lock")
			}
			if len(docker.removed) != 0 {
				t.Fatal("pruning removed a seed within the retention bound", docker.removed)
			}
		})
	}
	t.Run("failed_or_vanishing_builds_fail_startup", func(t *testing.T) {
		t.Setenv("TMPDIR", t.TempDir())
		lock, err := openSeedLock()
		if err != nil {
			t.Fatal(err)
		}
		defer lock.close()
		injected := errors.New("migration fault")
		docker := &fakeSeedDocker{images: map[string]image.InspectResponse{}}
		if _, err := resolveSeedWith(t.Context(), docker, lock, fakeSeedDigest, func(context.Context, seedDocker, string) (string, error) { return "", injected }); !errors.Is(err, injected) {
			t.Fatal("build failure lost its cause", err)
		}
		builds := 0
		if _, err := resolveSeedWith(t.Context(), docker, lock, fakeSeedDigest, func(context.Context, seedDocker, string) (string, error) { builds++; return "sha256:gone", nil }); err == nil || builds != seedResolveAttempts {
			t.Fatal("vanishing seed was accepted or retried without bound", err, builds)
		}
	})
	t.Run("waiting_for_a_builder_is_bounded_by_the_caller", func(t *testing.T) {
		t.Setenv("TMPDIR", t.TempDir())
		holder, err := openSeedLock()
		if err != nil {
			t.Fatal(err)
		}
		defer holder.close()
		if err := holder.acquire(t.Context(), unix.LOCK_EX); err != nil {
			t.Fatal(err)
		}
		lock, err := openSeedLock()
		if err != nil {
			t.Fatal(err)
		}
		defer lock.close()
		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		defer cancel()
		docker := &fakeSeedDocker{images: map[string]image.InspectResponse{}}
		if _, err := resolveSeedWith(ctx, docker, lock, fakeSeedDigest, nil); !errors.Is(err, context.DeadlineExceeded) || docker.inspects != 0 {
			t.Fatal("seed resolution bypassed a live builder", err, docker.inspects)
		}
	})
	t.Run("pruning_keeps_current_and_newest_other_digests", func(t *testing.T) {
		items := []image.Summary{
			{ID: "sha256:a-old", Created: 1, Labels: map[string]string{seedDigestLabel: "a"}},
			{ID: "sha256:current", Created: 5, Labels: map[string]string{seedDigestLabel: fakeSeedDigest}},
			{ID: "sha256:stale-copy", Created: 4, Labels: map[string]string{seedDigestLabel: fakeSeedDigest}},
			{ID: "sha256:b", Created: 3, Labels: map[string]string{seedDigestLabel: "b"}},
			{ID: "sha256:c", Created: 2, Labels: map[string]string{seedDigestLabel: "c"}},
		}
		got := seedsToPrune(items, "sha256:current", fakeSeedDigest)
		if want := []string{"sha256:stale-copy", "sha256:a-old"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("pruned %v, want %v", got, want)
		}
		if items[0].ID != "sha256:a-old" {
			t.Fatal("pruning reordered its input")
		}
	})
	t.Run("image_labels_drop_inherited_reaper_ownership", func(t *testing.T) {
		got := seedImageChanges(fakeSeedDigest, map[string]string{"org.testcontainers.sessionId": "s", "org.testcontainers": "true", "org.cartulary.other": "kept"})
		want := []string{"LABEL " + seedDigestLabel + "=" + fakeSeedDigest, "LABEL org.testcontainers=", "LABEL org.testcontainers.sessionId="}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("changes = %q, want %q", got, want)
		}
	})
	t.Run("reclaim_removes_only_this_users_builders", func(t *testing.T) {
		prefix := seedBuilderPrefix()
		docker := &fakeSeedDocker{containers: []dockercontainer.Summary{
			{ID: "mine", Names: []string{"/" + prefix + "0123456789abcdef"}},
			{ID: "other-user", Names: []string{"/cartulary-postgres-seed-build-99999-0123456789abcdef"}},
			{ID: "generation", Names: []string{"/suite-postgres-" + prefix}},
		}}
		if err := reclaimSeedBuilders(t.Context(), docker); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(docker.removed, []string{"mine"}) {
			t.Fatalf("reclaimed %v", docker.removed)
		}
		if seedBuilderName(fakeSeedDigest) != prefix+fakeSeedDigest[:16] {
			t.Fatal("builder name is not digest scoped")
		}
	})
}
