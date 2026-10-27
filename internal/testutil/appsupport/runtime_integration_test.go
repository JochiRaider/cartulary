package appsupport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore/s3transport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestApplicationFixturesAreIndependent_Integration(t *testing.T) {
	runtime := StartRuntime(t)
	first := runtime.StartServer(t, ServerOptions{Prefix: "fixture-first", TestRouteMode: httptestx.TestRouteModeDisabled})
	if _, err := first.DB.Exec("CREATE TABLE fixture_probe (value text); INSERT INTO fixture_probe VALUES ('first')"); err != nil {
		t.Fatal(err)
	}
	putFixtureObject(t, first.ObjectStore, "first")
	var secondDatabase, secondBucket string
	if !t.Run("independent live application", func(t *testing.T) {
		second := runtime.StartServer(t, ServerOptions{
			Prefix: "fixture-second", TestRouteMode: httptestx.TestRouteModeDisabled,
			ConfigureRuntime: func(options *server.Options) { secondBucket = options.Env[objectstore.BucketEnv] },
		})
		secondDatabase = second.Database.Name
		if secondDatabase == first.Database.Name || secondBucket == "" {
			t.Fatal("independent fixture identity missing")
		}
		var exists bool
		if err := second.DB.QueryRow("SELECT to_regclass('public.fixture_probe') IS NOT NULL").Scan(&exists); err != nil || exists {
			t.Fatalf("sibling database state leaked: exists=%t err=%v", exists, err)
		}
		if _, err := second.ObjectStore.StatObject(t.Context(), "fixture/probe"); !objectstore.IsObjectNotFound(err) {
			t.Fatalf("sibling object absence not proved: %v", err)
		}
		putFixtureObject(t, second.ObjectStore, "second")
		requireFixtureObject(t, second.ObjectStore, "second")
		requireFixtureObject(t, first.ObjectStore, "first")
	}) {
		return
	}
	var remains bool
	if err := first.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname=$1)", secondDatabase).Scan(&remains); err != nil || remains {
		t.Fatalf("child database was not retired: exists=%t err=%v", remains, err)
	}
	client, err := s3test.Start(t).Client(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	exists, err := s3transport.BucketExists(t.Context(), client, secondBucket)
	if err != nil || exists {
		t.Fatalf("child namespace was not retired: exists=%t err=%v", exists, err)
	}
	var value string
	if err := first.DB.QueryRow("SELECT value FROM fixture_probe").Scan(&value); err != nil || value != "first" {
		t.Fatalf("sibling cleanup damaged database: value=%q err=%v", value, err)
	}
	requireFixtureObject(t, first.ObjectStore, "first")
	putFixtureObject(t, first.ObjectStore, "after sibling cleanup")
	requireFixtureObject(t, first.ObjectStore, "after sibling cleanup")
}

func TestApplicationBorrowedFixturesSurviveFailureAndRestart_Integration(t *testing.T) {
	runtime := StartRuntime(t)
	database := runtime.PrepareIsolatedDatabase(t, "fixture-restart")
	filesystem, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store := &closeObservedStore{FilesystemStore: filesystem}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	options := ServerOptions{Prefix: "fixture-restart", Database: database, ObjectStore: store, TestRouteMode: httptestx.TestRouteModeDisabled}
	startupErr := errors.New("fixture startup rejected")
	var failedPool *pgxpool.Pool
	failedScope := &startupFailureScope{TB: t}
	t.Cleanup(failedScope.cleanup)
	func() {
		defer func() {
			failure, ok := recover().(expectedStartupFailure)
			if !ok || !errors.Is(failure.cause, startupErr) {
				t.Fatalf("expected failure after fixture acquisition, got %#v", failure)
			}
		}()
		invalid := options
		invalid.ConfigureRuntime = func(options *server.Options) { failedPool = options.Postgres.Pool() }
		invalid.AdditionalRoutes = []httpapi.RouteRegistrar{
			func(*http.ServeMux, httpapi.DependencySet) error { return startupErr },
		}
		runtime.StartServer(failedScope, invalid)
		t.Fatal("rejected startup succeeded")
	}()
	if failedPool == nil || failedPool.Ping(t.Context()) != nil {
		t.Fatal("failed server did not preserve its borrowed runtime pool")
	}
	failedScope.cleanup()
	if err := failedPool.Ping(t.Context()); err == nil {
		t.Fatal("failed startup left its fixture-owned runtime pool open")
	}
	if store.closes != 0 {
		t.Fatal("startup failure closed borrowed store")
	}
	var first *ServerHarness
	var packPath string
	if !t.Run("first server", func(t *testing.T) {
		first = runtime.StartServer(t, options)
		putFixtureObject(t, store, "durable")
		packPath = filepath.Join(first.Server.Config.Roots.ReferencePackStorage.Path, "fixture-bytes")
		if err := os.WriteFile(packPath, []byte("durable pack bytes"), 0o600); err != nil {
			t.Fatal(err)
		}
		first.Server.Close()
		requireFixtureObject(t, store, "durable")
	}) {
		return
	}
	if err := first.Pool.Ping(context.Background()); err == nil {
		t.Fatal("child runtime pool stayed open")
	}
	if err := first.DB.Ping(); err == nil {
		t.Fatal("child SQL handle stayed open")
	}
	t.Run("replacement server", func(t *testing.T) {
		second := runtime.StartServer(t, options)
		if second.Database != database {
			t.Fatal("borrowed database replaced")
		}
		if filepath.Dir(packPath) != second.Server.Config.Roots.ReferencePackStorage.Path {
			t.Fatal("restart changed deployment byte root")
		}
		content, err := os.ReadFile(packPath)
		if err != nil || string(content) != "durable pack bytes" {
			t.Fatalf("restart lost pack bytes: %q %v", content, err)
		}
		requireFixtureObject(t, second.ObjectStore, "durable")
	})
	if store.closes != 0 {
		t.Fatal("server cleanup closed borrowed store")
	}
	requireFixtureObject(t, store, "durable")
}

func putFixtureObject(t testing.TB, store objectstore.Store, content string) {
	t.Helper()
	if err := store.PutObject(context.Background(), "fixture/probe", bytes.NewBufferString(content), int64(len(content)), "text/plain"); err != nil {
		t.Fatal(err)
	}
}

func requireFixtureObject(t testing.TB, store objectstore.Store, want string) {
	t.Helper()
	reader, _, err := store.ReadObject(context.Background(), "fixture/probe", objectstore.ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(reader)
	closeErr := reader.Close()
	if err != nil || closeErr != nil || string(got) != want {
		t.Fatalf("fixture bytes = %q, want %q; read=%v close=%v", got, want, err, closeErr)
	}
}

type closeObservedStore struct {
	*objectstore.FilesystemStore
	closes int
}

func (s *closeObservedStore) Close() error { s.closes++; return s.FilesystemStore.Close() }

// Capture the helper's fatal startup result without turning an expected negative
// scenario into a failed Go subtest. Cleanup still runs against real fixtures.
type startupFailureScope struct {
	testing.TB
	cleanups []func()
	failed   bool
}
type expectedStartupFailure struct {
	message string
	cause   error
}

func (s *startupFailureScope) Fatal(args ...any) {
	s.failed = true
	panic(expectedStartupFailure{message: fmt.Sprint(args...)})
}
func (s *startupFailureScope) Fatalf(format string, args ...any) {
	s.failed = true
	failure := expectedStartupFailure{message: fmt.Sprintf(format, args...)}
	for _, arg := range args {
		if err, ok := arg.(error); ok {
			failure.cause = err
			break
		}
	}
	panic(failure)
}
func (s *startupFailureScope) Failed() bool     { return s.failed || s.TB.Failed() }
func (s *startupFailureScope) Cleanup(f func()) { s.cleanups = append(s.cleanups, f) }
func (s *startupFailureScope) cleanup() {
	for len(s.cleanups) > 0 {
		last := len(s.cleanups) - 1
		f := s.cleanups[last]
		s.cleanups = s.cleanups[:last]
		f()
	}
}
