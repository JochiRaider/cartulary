package postgres

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
)

// This exercises the dependency's pre-dispatch boundary. TLS certificate
// verification is a separate requirement: RequireAuth=none only forbids wire
// authentication mechanisms after the TLS handshake.
func TestPostgresDriverCertificateAuthenticationPolicy(t *testing.T) {
	tests := []struct {
		name      string
		challenge pgproto3.BackendMessage
		method    string
	}{
		{"certificate protocol", &pgproto3.AuthenticationOk{}, ""},
		{"cleartext password", &pgproto3.AuthenticationCleartextPassword{}, "password"},
		{"MD5 password", &pgproto3.AuthenticationMD5Password{Salt: [4]byte{1, 2, 3, 4}}, "md5"},
		{"SCRAM", &pgproto3.AuthenticationSASL{AuthMechanisms: []string{"SCRAM-SHA-256"}}, "scram-sha-256"},
		{"SCRAM channel binding", &pgproto3.AuthenticationSASL{AuthMechanisms: []string{"SCRAM-SHA-256-PLUS", "SCRAM-SHA-256"}}, "scram-sha-256"},
		{"GSS", &pgproto3.AuthenticationGSS{}, "gss"},
		{"OAuth", &pgproto3.AuthenticationSASL{AuthMechanisms: []string{"OAUTHBEARER"}}, "oauth"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			deadline, _ := ctx.Deadline()
			if err := server.SetDeadline(deadline); err != nil {
				t.Fatal(err)
			}
			serverResult := make(chan error, 1)
			go func() {
				backend := pgproto3.NewBackend(server, server)
				if _, err := backend.ReceiveStartupMessage(); err != nil {
					serverResult <- err
					return
				}
				backend.Send(test.challenge)
				if test.method == "" {
					backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
				}
				if err := backend.Flush(); err != nil {
					serverResult <- err
					return
				}
				if test.method == "" {
					message, err := backend.Receive()
					if err == nil {
						if _, ok := message.(*pgproto3.Terminate); !ok {
							err = fmt.Errorf("unexpected client message %T", message)
						}
					}
					serverResult <- err
					return
				}
				var response [1]byte
				n, err := server.Read(response[:])
				if n != 0 || !errors.Is(err, io.EOF) {
					serverResult <- fmt.Errorf("forbidden mechanism response: bytes=%d, error=%v", n, err)
					return
				}
				serverResult <- nil
			}()

			// Deliberately provide a password: an empty password is not proof that
			// a driver rejects a forbidden authentication mechanism.
			config, err := pgconn.ParseConfig("host=127.0.0.1 user=fixture database=fixture password=fixture-password sslmode=disable require_auth=none")
			if err != nil {
				t.Fatal(err)
			}
			config.Fallbacks = nil
			config.ValidateConnect = nil
			config.AfterConnect = nil
			config.DialFunc = func(context.Context, string, string) (net.Conn, error) { return client, nil }
			providerCalled := false
			config.OAuthTokenProvider = func(context.Context) (string, error) {
				providerCalled = true
				return "fixture-token", nil
			}
			connection, err := pgconn.ConnectConfig(ctx, config)
			if connection != nil {
				if closeErr := connection.Close(ctx); closeErr != nil {
					t.Errorf("close connection: %v", closeErr)
				}
			}
			if test.method == "" {
				if err != nil {
					t.Errorf("certificate protocol admission: %v", err)
				}
			} else {
				want := fmt.Sprintf("authentication method requirement %q failed: server requested %s authentication", "none", test.method)
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Errorf("expected pre-dispatch policy rejection %q, got %v", want, err)
				}
			}
			if providerCalled {
				t.Error("OAuth provider invoked before authentication admission")
			}
			if err := <-serverResult; err != nil {
				t.Errorf("server observation: %v", err)
			}
		})
	}
}
