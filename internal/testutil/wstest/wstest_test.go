package wstest

import (
	"context"
	"testing"

	"github.com/coder/websocket"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestHarnessOpensAndClosesSocketAgainstBootstrapBoundary(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "wstest")

	s3Harness := s3test.Start(t)
	bucket := s3Harness.BootstrapBucketT(t, "wstest")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env, TestRouteMode: httptestx.TestRouteModeDisabled})
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
