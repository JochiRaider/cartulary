package wstest

import (
	"context"
	"testing"

	"github.com/coder/websocket"

	platformws "example.com/todo/cartulary/internal/platform/ws"
	"example.com/todo/cartulary/internal/testutil/httptestx"
	"example.com/todo/cartulary/internal/testutil/pgtest"
	"example.com/todo/cartulary/internal/testutil/s3test"
)

func TestHarnessOpensAndClosesSocketAgainstBootstrapBoundary(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "wstest")
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	defer func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	}()

	s3Harness := s3test.Start(t)
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "wstest")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env})
	client := Connect(t, server.HTTP.URL, "/ws/v1/bootstrap-harness")

	ack, err := client.Handshake(context.Background())
	if err != nil {
		t.Fatalf("websocket handshake: %v", err)
	}
	RequireMessageType(t, ack, "handshake_ack")

	if err := client.Send(context.Background(), platformws.Message{Type: "trigger_session_revoked"}); err != nil {
		t.Fatalf("send revoke trigger: %v", err)
	}
	message, err := client.Receive(context.Background())
	if err != nil {
		t.Fatalf("receive session_revoked: %v", err)
	}
	RequireSessionRevoked(t, message)

	_, err = client.Receive(context.Background())
	if err == nil {
		t.Fatal("expected websocket close after session_revoked")
	}
	RequireClose(t, err, websocket.StatusPolicyViolation, "session_revoked")
}
