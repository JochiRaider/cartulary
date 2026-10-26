package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"os"

	testcontainers "github.com/testcontainers/testcontainers-go"
	"reflect"
	"sync"
	"testing"
	"time"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func testTemplatePreparationContract(t *testing.T) {
	for _, stage := range []string{"success", "create_rejected", "create_committed_error", "created_event", "open", "initialize", "cancelled", "migrated_event", "close", "seal", "initialize_and_cleanup"} {
		t.Run(stage, func(t *testing.T) {
			primary, closeFailure, dropFailure := errors.New("preparation fault"), errors.New("handle close fault"), errors.New("drop fault")
			var sequence []string
			dbOps := defaultDatabaseOperations()
			dbOps.create = func(_ context.Context, _, _, _ string, confirmed func() error) (bool, error) {
				sequence = append(sequence, "create")
				if stage == "create_rejected" {
					return false, primary
				}
				if err := confirmed(); err != nil {
					return true, err
				}
				if stage == "create_committed_error" {
					return true, primary
				}
				return true, nil
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			dbOps.drop = func(cleanup context.Context, _, name string, _ func(postgrescleanup.DropAttempt)) error {
				sequence = append(sequence, "drop")
				deadline, ok := cleanup.Deadline()
				if name != "ct_tpl_contract" || cleanup.Err() != nil || !ok || time.Until(deadline) > 35*time.Second {
					t.Fatal("lost bounded exact cleanup authority")
				}
				if stage == "initialize_and_cleanup" {
					return dropFailure
				}
				return nil
			}
			ops := defaultTemplateOperations()
			ops.open = func(_ context.Context, _ string, purpose postgres.Purpose) (*sql.DB, error) {
				sequence = append(sequence, "open")
				if purpose != postgres.PurposeMigration {
					t.Fatal("wrong template purpose")
				}
				if stage == "open" {
					return nil, primary
				}
				return sql.Open("pgx", "postgres://fixture@localhost/postgres")
			}
			ops.initialize = func(_ context.Context, _ *sql.DB, _ string, source *database_migrations.Source) error {
				sequence = append(sequence, "initialize")
				if source == nil {
					t.Fatal("missing canonical source")
				}
				if stage == "cancelled" {
					cancel()
					return context.Canceled
				}
				if stage == "initialize" || stage == "initialize_and_cleanup" {
					return primary
				}
				return nil
			}
			ops.close = func(db *sql.DB) error {
				sequence = append(sequence, "close")
				if err := db.Close(); err != nil {
					t.Fatal(err)
				}
				if stage == "close" {
					return primary
				}
				if stage == "initialize_and_cleanup" {
					return closeFailure
				}
				return nil
			}
			ops.seal = func(context.Context, string, string) error {
				sequence = append(sequence, "seal")
				if stage == "seal" {
					return primary
				}
				return nil
			}
			h := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps}
			owner := &OwnedCluster{Container: fakePostgresContainer{}, fixture: h, templateOps: &ops}
			owner.ready.Store(true)
			h.events = func(event suiteservices.Event) error {
				switch event.Type {
				case suiteservices.EventPostgresDBCreated:
					sequence = append(sequence, "created")
					if !h.ownsDatabase(event.Name) {
						t.Fatal("creation published before ownership")
					}
					if stage == "created_event" {
						return primary
					}
				case suiteservices.EventPostgresDBMigrated:
					sequence = append(sequence, "migrated")
					if stage == "migrated_event" {
						return primary
					}
				case suiteservices.EventPostgresDBDropped:
					sequence = append(sequence, "dropped")
				}
				return nil
			}

			if stage == "success" {
				var preparations sync.WaitGroup
				for range 8 {
					preparations.Go(func() {
						database, err := owner.PrepareTemplate(ctx, "ct_tpl_contract")
						if err != nil || database == nil || database.Name != "ct_tpl_contract" {
							t.Error("template not published", err)
						}
					})
				}
				preparations.Wait()
				want := []string{"create", "created", "open", "initialize", "migrated", "close", "seal"}
				if !reflect.DeepEqual(sequence, want) || h.templateDB != "ct_tpl_contract" {
					t.Fatal("concurrent preparation replayed or published out of order", sequence)
				}
				return
			}
			database, err := owner.PrepareTemplate(ctx, "ct_tpl_contract")

			wantErr := primary
			if stage == "cancelled" {
				wantErr = context.Canceled
			}
			if !errors.Is(err, wantErr) || database != nil || h.templateDB != "" {
				t.Fatalf("failed preparation published or lost its cause: %v, %v", database, err)
			}
			if stage == "create_rejected" {
				if h.ownsDatabase("ct_tpl_contract") || !reflect.DeepEqual(sequence, []string{"create"}) {
					t.Fatal("failed CREATE acquired deletion authority")
				}
			} else {
				if !h.ownsDatabase("ct_tpl_contract") {
					t.Fatal("lost confirmed creation")
				}
				drops := 0
				closed := stage == "open" || stage == "created_event" || stage == "create_committed_error"
				for _, operation := range sequence {
					if operation == "close" {
						closed = true
					}
					if operation == "drop" {
						drops++
						if !closed {
							t.Fatal("drop preceded handle closure")
						}
					}
				}
				if drops != 1 {
					t.Fatalf("disposal count = %d", drops)
				}
			}
			if stage == "initialize_and_cleanup" && (!errors.Is(err, closeFailure) || !errors.Is(err, dropFailure) || !h.quarantined.Load()) {
				t.Fatal("lost secondary failures or quarantine", err)
			}
			settledOperations := len(sequence)
			if database, repeated := owner.PrepareTemplate(t.Context(), "ct_tpl_contract"); database != nil || !errors.Is(repeated, wantErr) || len(sequence) != settledOperations {
				t.Error("failed preparation did not retain its terminal result", repeated, sequence)
			}
		})
	}
	t.Run("cancelled_waiter_leaves_owner_running", testTemplateCancelledWaiter)
	t.Run("connection_admission_cancellation", testTemplateAdmissionCancellation)
	t.Run("invalid_request_does_not_retire", testInvalidTemplateRequest)
	t.Run("real_transport_open_cancellation", testPurposeOpenCancellation)
	t.Run("ordinary_fixture_requires_template", func(t *testing.T) {
		h := &Harness{}
		if database, err := h.PrepareDatabase(t.Context(), "missing"); err == nil || database != nil {
			t.Fatal("ordinary preparation implicitly provisioned a template")
		}

	})
}

func testTemplateCancelledWaiter(t *testing.T) {
	dbOps := defaultDatabaseOperations()
	dbOps.create = func(_ context.Context, _, _, _ string, confirmed func() error) (bool, error) {
		return true, confirmed()
	}
	dbOps.drop = func(context.Context, string, string, func(postgrescleanup.DropAttempt)) error { return nil }
	ops := defaultTemplateOperations()
	ops.open = func(context.Context, string, postgres.Purpose) (*sql.DB, error) {
		return sql.Open("pgx", "postgres://unused/postgres")
	}
	entered, release := make(chan struct{}), make(chan struct{})
	ops.initialize = func(context.Context, *sql.DB, string, *database_migrations.Source) error {
		close(entered)
		<-release
		return nil
	}
	ops.seal = func(context.Context, string, string) error { return nil }
	h := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps, events: func(suiteservices.Event) error { return nil }}
	owner := &OwnedCluster{Container: fakePostgresContainer{}, fixture: h, templateOps: &ops}
	owner.ready.Store(true)
	finished := make(chan error, 1)
	go func() { _, err := owner.PrepareTemplate(t.Context(), "ct_tpl_waiter"); finished <- err }()
	<-entered
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	waiter := make(chan error, 1)
	go func() { _, err := owner.PrepareTemplate(ctx, "ct_tpl_waiter"); waiter <- err }()
	select {
	case err := <-waiter:
		if !errors.Is(err, context.Canceled) {
			t.Error("waiter lost cancellation", err)
		}
	case <-time.After(time.Second):
		t.Error("cancelled waiter remained blocked behind preparation")
	}
	deadline, expire := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer expire()
	deadlineWaiter := make(chan error, 1)
	go func() { _, err := owner.PrepareTemplate(deadline, "ct_tpl_waiter"); deadlineWaiter <- err }()
	select {
	case err := <-deadlineWaiter:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Error("waiting deadline lost", err)
		}
	case <-time.After(time.Second):
		t.Error("waiting caller ignored its deadline")
	}
	close(release)
	if err := <-finished; err != nil || h.templateDB == "" || h.quarantined.Load() {
		t.Fatal("waiter cancelled the owning preparation", err)
	}
}

func testTemplateAdmissionCancellation(t *testing.T) {
	for _, stage := range []string{"work", "guard"} {
		for _, termination := range []string{"cancel", "deadline"} {
			t.Run(stage+"_"+termination, func(t *testing.T) {
				dbOps := defaultDatabaseOperations()
				dbOps.create = func(_ context.Context, _, _, _ string, confirmed func() error) (bool, error) {
					return true, confirmed()
				}
				settling, release := make(chan struct{}), make(chan struct{})
				dbOps.drop = func(ctx context.Context, _, _ string, _ func(postgrescleanup.DropAttempt)) error {
					if ctx.Err() != nil {
						t.Error("cleanup inherited cancellation")
					}
					close(settling)
					<-release
					return nil
				}
				ops := defaultTemplateOperations()
				entered := make(chan struct{})
				stalled := func(ctx context.Context, _ string, _ postgres.Purpose) (*sql.DB, error) {
					close(entered)
					<-ctx.Done()
					return nil, ctx.Err()
				}
				ops.open = stalled
				if stage == "guard" {
					ops.open = func(context.Context, string, postgres.Purpose) (*sql.DB, error) {
						return sql.Open("pgx", "postgres://unused/postgres")
					}
					ops.initialize = func(ctx context.Context, db *sql.DB, dsn string, source *database_migrations.Source) error {
						return initializeFreshDatabase(ctx, db, dsn, source, stalled)
					}
				}
				ops.seal = func(context.Context, string, string) error { t.Error("failed admission reached sealing"); return nil }
				h := &Harness{binding: bindFixture(processFixtureInputs()), operations: &dbOps, events: func(suiteservices.Event) error { return nil }}
				owner := &OwnedCluster{Container: fakePostgresContainer{}, fixture: h, templateOps: &ops}
				owner.ready.Store(true)
				ctx, cancel := context.WithCancel(t.Context())
				expected := context.Canceled
				if termination == "deadline" {
					cancel()
					ctx, cancel = context.WithTimeout(t.Context(), 100*time.Millisecond)
					expected = context.DeadlineExceeded
				}
				defer cancel()
				finished := make(chan error, 1)
				go func() { _, err := owner.PrepareTemplate(ctx, "ct_tpl_stalled"); finished <- err }()
				<-entered
				if termination == "cancel" {
					cancel()
				}
				select {
				case <-settling:
				case <-time.After(time.Second):
					t.Fatal("admission did not respect cancellation")
				}
				select {
				case err := <-finished:
					t.Error("terminal failure published before settlement", err)
				default:
				}
				close(release)
				if err := <-finished; !errors.Is(err, expected) || !h.quarantined.Load() || h.templateDB != "" {
					t.Fatal("lost cancellation, retirement, or publication boundary", err)
				}
				if _, err := owner.PrepareTemplate(t.Context(), "ct_tpl_stalled"); !errors.Is(err, expected) {
					t.Fatal("failure did not remain terminal", err)
				}
			})
		}
	}
}

func testInvalidTemplateRequest(t *testing.T) {
	h := &Harness{}
	owner := &OwnedCluster{Container: fakePostgresContainer{}, fixture: h}
	owner.ready.Store(true)
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	for _, ctx := range []context.Context{nil, cancelled, t.Context()} {
		if _, err := owner.PrepareTemplate(ctx, "invalid/name"); err == nil || owner.preparation != nil || h.quarantined.Load() {
			t.Fatal("caller error acquired or retired an attempt", err)
		}
	}
	//lint:ignore SA1012 Verify the helper rejects a nil context before I/O.
	if _, err := OpenPurposeDatabase(nil, "invalid", postgres.PurposeMigration); err == nil {
		t.Fatal("nil context accepted")
	}
	if _, err := OpenPurposeDatabase(cancelled, "invalid", postgres.PurposeMigration); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled request reached DSN admission", err)
	}
}

func testPurposeOpenCancellation(t *testing.T) {
	_, directory, err := preparePostgresTLS(testcontainers.ContainerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(directory); err != nil {
			t.Error(err)
		}
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	dsn := certificateDSN(host, port, "postgres", "cartulary", directory)
	for _, deadline := range []bool{false, true} {
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		expected := context.Canceled
		if deadline {
			expected = context.DeadlineExceeded
		}
		finished := make(chan error, 1)
		go func() {
			db, err := OpenPurposeDatabase(ctx, dsn, postgres.PurposeMigration)
			if db != nil {
				_ = db.Close()
			}
			finished <- err
		}()
		_ = listener.(*net.TCPListener).SetDeadline(time.Now().Add(2 * time.Second))
		connection, err := listener.Accept()
		if err != nil {
			cancel()
			t.Fatal("purpose open did not reach transport", err)
		}
		if !deadline {
			cancel()
		}
		select {
		case err := <-finished:
			if !errors.Is(err, expected) {
				t.Error("transport admission lost context cause", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("purpose open ignored cancellation")
		}
		cancel()
		_ = connection.Close()
	}
}
