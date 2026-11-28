package pgtest

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"golang.org/x/sys/unix"

	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

// A seed image is a content-addressed build artifact. It holds one PostgreSQL
// data directory that the full cluster provisioning and migration chain
// produced for a TemplateDigest, cleanly shut down, and no run, suite,
// borrower or credential state. Physical generations start from a private
// volume copy of it and seal their own copy of its template.
const (
	// seedFormat enters ProvisioningDigest; change it with the seed procedure.
	seedFormat        = "cartulary.postgres_seed.v1"
	seedRepository    = "cartulary-postgres-seed"
	seedVolumeRoot    = "/pgseed"
	seedDataDirectory = seedVolumeRoot + "/data"
	// A memory-backed generation copies the seed here when it starts
	// (TH-HARNESS-REQ-804).
	memoryDataRoot      = "/pgmem"
	memoryDataDirectory = memoryDataRoot + "/data"
	seedTemplateName    = "ct_seed_template"
	seedDigestLabel     = "org.cartulary.postgres-seed.template-digest"
	// Builds keep the newest seeds so concurrent checkouts with different
	// migrations do not rebuild each other's seed on every generation.
	seedRetained        = 3
	seedResolveAttempts = 3
	seedLockPoll        = 50 * time.Millisecond
	seedStopTimeout     = 60 * time.Second
	seedCleanupTimeout  = 30 * time.Second
)

// clusterSource names the image a physical generation is created from.
type clusterSource struct {
	image  string
	name   string // fixed container name; only seed builders set one
	seeded bool   // the image carries the sealed seed data directory
}

type seedDocker interface {
	ImageInspect(context.Context, string, ...client.ImageInspectOption) (client.ImageInspectResult, error)
	ImageList(context.Context, client.ImageListOptions) (client.ImageListResult, error)
	ImageRemove(context.Context, string, client.ImageRemoveOptions) (client.ImageRemoveResult, error)
	ContainerList(context.Context, client.ContainerListOptions) (client.ContainerListResult, error)
	ContainerRemove(context.Context, string, client.ContainerRemoveOptions) (client.ContainerRemoveResult, error)
	ContainerInspect(context.Context, string, client.ContainerInspectOptions) (client.ContainerInspectResult, error)
	ContainerCommit(context.Context, string, client.ContainerCommitOptions) (client.ContainerCommitResult, error)
}

// seedLock is the host-wide seed authority. Holders of the shared lock pin
// every seed until their container exists; the exclusive holder alone builds,
// prunes and reclaims builders, so any builder container it finds is orphaned.
type seedLock struct{ file *os.File }

func openSeedLock() (*seedLock, error) {
	path := filepath.Join(os.TempDir(), "cartulary-postgres-seed-"+strconv.Itoa(os.Getuid())+".lock")
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open postgres seed lock: %w", err)
	}
	return &seedLock{file: file}, nil
}

// acquire takes or converts the lock without blocking past the context.
func (l *seedLock) acquire(ctx context.Context, mode int) error {
	for {
		err := unix.Flock(int(l.file.Fd()), mode|unix.LOCK_NB)
		if err == nil {
			return nil
		}
		if !errors.Is(err, unix.EWOULDBLOCK) && !errors.Is(err, unix.EINTR) {
			return fmt.Errorf("acquire postgres seed lock: %w", err)
		}
		timer := time.NewTimer(seedLockPoll)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("wait for postgres seed lock: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

func (l *seedLock) unlock() error { return unix.Flock(int(l.file.Fd()), unix.LOCK_UN) }

// Closing the descriptor releases the lock.
func (l *seedLock) close() { _ = l.file.Close() }

func seedReference(digest string) string { return seedRepository + ":" + digest }

func seedBuilderName(digest string) string {
	return seedBuilderPrefix() + digest[:16]
}

func seedBuilderPrefix() string {
	return "cartulary-postgres-seed-build-" + strconv.Itoa(os.Getuid()) + "-"
}

// resolveSeed returns the verified seed for this binary's template digest,
// building it once when absent. The returned release keeps the seed pinned
// until the caller's container exists, which then pins it in Docker.
func resolveSeed(ctx context.Context) (clusterSource, func(), error) {
	lock, err := openSeedLock()
	if err != nil {
		return clusterSource{}, nil, err
	}
	docker, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		lock.close()
		return clusterSource{}, nil, fmt.Errorf("connect to docker for postgres seed: %w", err)
	}
	defer docker.Close()
	image, err := resolveSeedWith(ctx, docker, lock, TemplateDigest(pgschema.MustHash()), buildSeed)
	if err != nil {
		lock.close()
		return clusterSource{}, nil, err
	}
	return clusterSource{image: image, seeded: true}, lock.close, nil
}

type seedBuilder func(context.Context, seedDocker, string) (string, error)

// resolveSeedWith returns holding the shared lock. A shared holder never
// upgrades in place: it releases first, so two waiters cannot deadlock.
func resolveSeedWith(ctx context.Context, docker seedDocker, lock *seedLock, digest string, build seedBuilder) (string, error) {
	for range seedResolveAttempts {
		if err := lock.acquire(ctx, unix.LOCK_SH); err != nil {
			return "", err
		}
		image, err := inspectSeed(ctx, docker, digest)
		if err != nil || image != "" {
			return image, err
		}
		if err := lock.unlock(); err != nil {
			return "", fmt.Errorf("release postgres seed lock: %w", err)
		}
		if err := lock.acquire(ctx, unix.LOCK_EX); err != nil {
			return "", err
		}
		image, err = inspectSeed(ctx, docker, digest)
		if err != nil {
			return "", err
		}
		if image == "" {
			built, err := build(ctx, docker, digest)
			if err != nil {
				return "", fmt.Errorf("build postgres seed image: %w", err)
			}
			pruneSeeds(ctx, docker, built, digest)
		}
		// The next shared acquisition converts the exclusive lock and verifies
		// the seed again; conversion is not atomic.
	}
	return "", errors.New("postgres seed image was removed during startup")
}

// inspectSeed returns the content ID of a seed whose label proves the digest.
// A missing or mislabeled seed is absent and is rebuilt under the tag.
func inspectSeed(ctx context.Context, docker seedDocker, digest string) (string, error) {
	result, err := docker.ImageInspect(ctx, seedReference(digest))
	if errdefs.IsNotFound(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("inspect postgres seed image: %w", err)
	}
	if result.Config == nil || result.Config.Labels[seedDigestLabel] != digest || !strings.HasPrefix(result.ID, "sha256:") {
		return "", nil
	}
	return result.ID, nil
}

// buildSeed provisions one cluster through the ordinary fresh path, then
// commits its stopped data directory. The caller holds the exclusive lock.
func buildSeed(ctx context.Context, docker seedDocker, digest string) (_ string, retErr error) {
	if err := reclaimSeedBuilders(ctx, docker); err != nil {
		return "", err
	}
	builder, err := startOwned(ctx, StartOptions{Events: func(suiteservices.Event) error { return nil }}, seedBuilderOperations(seedBuilderName(digest)))
	if builder != nil {
		defer func() {
			cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), seedCleanupTimeout)
			defer cancel()
			retErr = errors.Join(retErr, builder.Close(cleanup))
		}()
	}
	if err != nil {
		return "", fmt.Errorf("start seed builder: %w", err)
	}
	if err := builder.prepareSeedTemplate(ctx); err != nil {
		return "", err
	}
	// Credentials and the one-time init script are per-generation inputs;
	// they never enter the image.
	code, _, err := builder.Container.Exec(ctx, []string{"rm", "-rf", "--", postgresTLSContainerRoot, postgresInitScriptPath})
	if err != nil || code != 0 {
		return "", errors.Join(fmt.Errorf("remove seed builder credentials: exit %d", code), err)
	}
	timeout := seedStopTimeout
	if err := builder.Container.Stop(ctx, &timeout); err != nil {
		return "", fmt.Errorf("stop seed builder: %w", err)
	}
	inspected, err := docker.ContainerInspect(ctx, builder.ContainerID(), client.ContainerInspectOptions{})
	if err != nil {
		return "", fmt.Errorf("inspect stopped seed builder: %w", err)
	}
	state := inspected.Container.State
	if state == nil || state.Running || state.ExitCode != 0 {
		return "", errors.New("seed builder did not shut down cleanly")
	}
	committed, err := docker.ContainerCommit(ctx, builder.ContainerID(), client.ContainerCommitOptions{
		Reference: seedReference(digest),
		Changes:   seedImageChanges(digest, inspected.Container.Config.Labels),
	})
	if err != nil {
		return "", fmt.Errorf("commit postgres seed image: %w", err)
	}
	return committed.ID, nil
}

// seedImageChanges labels the image with its digest and blanks inherited
// Testcontainers labels, so a builder session's reaper cannot claim the image.
func seedImageChanges(digest string, inherited map[string]string) []string {
	keys := []string{}
	for key := range inherited {
		if key == "org.testcontainers" || strings.HasPrefix(key, "org.testcontainers.") {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	changes := []string{"LABEL " + seedDigestLabel + "=" + digest}
	for _, key := range keys {
		changes = append(changes, "LABEL "+key+"=")
	}
	return changes
}

// reclaimSeedBuilders removes builder containers left by dead builders. Only
// the exclusive holder calls it, so no live builder owns any of them.
func reclaimSeedBuilders(ctx context.Context, docker seedDocker) error {
	prefix := seedBuilderPrefix()
	listed, err := docker.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: client.Filters{}.Add("name", prefix)})
	if err != nil {
		return fmt.Errorf("list orphaned seed builders: %w", err)
	}
	for _, item := range listed.Items {
		if !slices.ContainsFunc(item.Names, func(name string) bool { return strings.HasPrefix(strings.TrimPrefix(name, "/"), prefix) }) {
			continue
		}
		if _, err := docker.ContainerRemove(ctx, item.ID, client.ContainerRemoveOptions{Force: true, RemoveVolumes: true}); err != nil && !errdefs.IsNotFound(err) {
			return fmt.Errorf("remove orphaned seed builder: %w", err)
		}
	}
	return nil
}

// pruneSeeds keeps the current seed and the newest others up to the bound.
// Docker refuses to remove a seed a container references, and shared lock
// holders pin seeds they are about to use, so pruning never removes a seed in
// use. Failure leaves an extra image and does not fail startup.
func pruneSeeds(ctx context.Context, docker seedDocker, current, digest string) {
	listed, err := docker.ImageList(ctx, client.ImageListOptions{Filters: client.Filters{}.Add("label", seedDigestLabel)})
	if err != nil {
		return
	}
	for _, id := range seedsToPrune(listed.Items, current, digest) {
		_, _ = docker.ImageRemove(ctx, id, client.ImageRemoveOptions{})
	}
}

// seedsToPrune keeps the current seed and the newest seeds of other digests
// up to seedRetained in all. Superseded copies of the current digest go.
func seedsToPrune(items []image.Summary, current, digest string) []string {
	ordered := slices.Clone(items)
	slices.SortFunc(ordered, func(a, b image.Summary) int {
		if order := cmp.Compare(b.Created, a.Created); order != 0 {
			return order
		}
		return strings.Compare(a.ID, b.ID)
	})
	kept, prune := 1, []string{}
	for _, item := range ordered {
		if item.ID == current {
			continue
		}
		if item.Labels[seedDigestLabel] != digest && kept < seedRetained {
			kept++
			continue
		}
		prune = append(prune, item.ID)
	}
	return prune
}
