package collaboration_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestReplayListenerHoldsNoRuntimePoolConnection_Integration covers the
// Collaboration replay listener's database session. Like the leases of Core 04
// REQ-04-145, the process-lifetime LISTEN runs on its own session outside the
// runtime pool, opened with the pool's credentials, so it never reduces the
// capacity available to request or job work (REQ-04-079).
func TestReplayListenerHoldsNoRuntimePoolConnection_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:        "collaboration-replay-listener-session",
		Dependencies:  httpapi.DependencySet{},
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	listener := awaitReplayListenerSession(ctx, t, harness)

	// The runtime pool serves its complete capacity while the listener runs.
	capacity := int(harness.Pool.Config().MaxConns)
	acquireCtx, cancelAcquire := context.WithTimeout(ctx, 10*time.Second)
	defer cancelAcquire()
	connections := make([]*pgxpool.Conn, 0, capacity)
	defer func() {
		for _, connection := range connections {
			connection.Release()
		}
	}()
	pooled := make(map[uint32]bool, capacity)
	for index := range capacity {
		connection, err := harness.Pool.Acquire(acquireCtx)
		if err != nil {
			t.Fatalf("runtime pool connection %d of %d unavailable while the replay listener runs: %v", index+1, capacity, err)
		}
		connections = append(connections, connection)
		pooled[connection.Conn().PgConn().PID()] = true
	}
	if pooled[listener.pid] {
		t.Fatalf("replay listener session %d is a runtime pool connection", listener.pid)
	}
	var sessionUser string
	if err := connections[0].QueryRow(ctx, `SELECT session_user::text`).Scan(&sessionUser); err != nil {
		t.Fatalf("read runtime pool session user: %v", err)
	}
	if listener.user != sessionUser {
		t.Fatalf("replay listener session user %q differs from the runtime pool's %q", listener.user, sessionUser)
	}
}

type replayListenerSession struct {
	pid  uint32
	user string
}

// awaitReplayListenerSession returns the one backend of the server's database
// that is listening for replay notifications.
func awaitReplayListenerSession(ctx context.Context, t *testing.T, harness *appsupport.ServerHarness) replayListenerSession {
	t.Helper()
	for {
		rows, err := harness.DB.QueryContext(ctx, `
SELECT pid, usename::text
  FROM pg_stat_activity
 WHERE datname = current_database()
   AND query = 'LISTEN cartulary_collaboration_replay'
`)
		if err != nil {
			t.Fatalf("observe replay listener sessions: %v", err)
		}
		var sessions []replayListenerSession
		for rows.Next() {
			var pid int32
			var user string
			if err := rows.Scan(&pid, &user); err != nil {
				_ = rows.Close()
				t.Fatalf("scan replay listener session: %v", err)
			}
			sessions = append(sessions, replayListenerSession{pid: uint32(pid), user: user})
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			t.Fatalf("iterate replay listener sessions: %v", err)
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("close replay listener sessions: %v", err)
		}
		switch len(sessions) {
		case 0:
		case 1:
			return sessions[0]
		default:
			t.Fatalf("server runs %d replay listener sessions, want one: %+v", len(sessions), sessions)
		}
		select {
		case <-ctx.Done():
			t.Fatalf("no replay listener session appeared: %v", ctx.Err())
		case <-time.After(50 * time.Millisecond):
		}
	}
}
