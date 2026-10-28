package incidentwstest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

func TestIncidentClientExpectations(t *testing.T) {
	incident, first, second := uuid.New(), uuid.New(), uuid.New()
	record := func(id uuid.UUID, seq int64) platformws.Message {
		return collabtestprotocol.SequencedMessage(t, "record_changed", incident, uuid.NewString(), seq, time.Now().UTC(), map[string]any{
			"record_id": id.String(), "row_version": int64(1), "client_txn_id": "expected", "change_set_id": uuid.NewString(), "actor_user_id": uuid.NewString(),
			"changed_field_keys": []string{}, "affected_views": []map[string]any{{"view_schema_id": "cartulary.view.notes.v1", "change_kind": "invalidate"}},
		})
	}
	queued := func(messages ...platformws.Message) *Client {
		client := &Client{incidentID: incident.String(), events: make(chan event, len(messages))}
		for _, message := range messages {
			client.events <- event{message: &message}
		}
		return client
	}
	valid := record(first, 1)
	for _, tc := range []struct {
		name   string
		mutate func(*platformws.Message)
	}{
		{"wrong record", func(m *platformws.Message) { *m = record(second, 1) }},
		{"wrong version", func(m *platformws.Message) {
			m.Payload = json.RawMessage(strings.Replace(string(m.Payload), `"row_version":1`, `"row_version":2`, 1))
		}},
		{"wrong incident", func(m *platformws.Message) { m.IncidentID = uuid.NewString() }},
		{"malformed payload", func(m *platformws.Message) { m.Payload = json.RawMessage(`{`) }},
		{"missing envelope", func(m *platformws.Message) { m.EventID = "" }},
		{"unexpected family", func(m *platformws.Message) { m.Type = "session_revoked" }},
		{"missing keys", func(m *platformws.Message) {
			m.Payload = json.RawMessage(strings.Replace(string(m.Payload), `"changed_field_keys":[],`, ``, 1))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := valid
			tc.mutate(&m)
			if _, err := queued(m).awaitRecordChangedEvents(map[uuid.UUID]int64{first: 1}); err == nil {
				t.Fatal("unexpected observation accepted")
			}
		})
	}
	for _, duplicate := range []platformws.Message{valid, record(first, 2)} {
		if _, err := queued(valid, duplicate).awaitRecordChangedEvents(map[uuid.UUID]int64{first: 1, second: 1}); err == nil {
			t.Fatal("duplicate event accepted")
		}
	}
	if _, err := queued(record(first, 2), record(second, 1)).awaitRecordChangedEvents(map[uuid.UUID]int64{first: 1, second: 1}); err == nil {
		t.Fatal("out-of-order stream accepted")
	}
	got, err := queued(platformws.Message{Type: "presence_delta"}, valid, platformws.Message{Type: "ping"}, record(second, 2)).awaitRecordChangedEvents(map[uuid.UUID]int64{first: 1, second: 1})
	if err != nil || len(got) != 2 || got[first].StreamSeq != 1 || got[second].StreamSeq != 2 {
		t.Fatalf("complete observed set: %#v %v", got, err)
	}
	malformed := valid
	malformed.Payload = json.RawMessage(`{`)
	if _, err := queued(malformed).awaitReplayable(time.Second); err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("malformed observation treated as absence: %v", err)
	}
}

func TestIncidentClientReaderLifecycle(t *testing.T) {
	t.Run("timeout leaves socket usable and peer close follows queued messages", func(t *testing.T) {
		send := make(chan struct{})
		client := localClient(t, func(ctx context.Context, conn *websocket.Conn) {
			select {
			case <-send:
			case <-ctx.Done():
				return
			}
			for _, kind := range []string{"first", "second"} {
				if err := wsjson.Write(ctx, conn, platformws.Message{Type: kind}); err != nil {
					return
				}
			}
		})
		if _, err := client.AwaitNextMessage(time.Millisecond); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("observation timeout: %v", err)
		}
		close(send)
		for _, want := range []string{"first", "second"} {
			got, err := client.AwaitNextMessage(time.Second)
			if err != nil || got.Type != want {
				t.Fatalf("FIFO got %q/%v want %q", got.Type, err, want)
			}
		}
		if err := client.AwaitClose(time.Second); err == nil {
			t.Fatal("missing peer error after queued data")
		}
	})
	t.Run("full queue permits shutdown and repeated close", func(t *testing.T) {
		client := localClient(t, func(ctx context.Context, conn *websocket.Conn) { _, _, _ = conn.Read(ctx) })
		for i := 0; i < cap(client.events); i++ {
			client.events <- event{message: &platformws.Message{Type: "queued"}}
		}
		enqueueDone := make(chan bool, 1)
		go func() { enqueueDone <- client.enqueue(event{message: &platformws.Message{Type: "blocked"}}) }()
		closed := make(chan struct{})
		go func() {
			client.Close(wstest.StatusNormalClosure, "test_complete")
			client.Close(wstest.StatusNormalClosure, "again")
			close(closed)
		}()
		select {
		case <-closed:
		case <-time.After(2 * time.Second):
			t.Fatal("reader cleanup blocked on full queue")
		}
		select {
		case sent := <-enqueueDone:
			if sent {
				t.Fatal("full queue unexpectedly accepted an event")
			}
		case <-time.After(time.Second):
			t.Fatal("blocked enqueue leaked")
		}
		select {
		case <-client.done:
		default:
			t.Fatal("close returned before reader joined")
		}
	})
}

func localClient(t *testing.T, serve func(context.Context, *websocket.Conn)) *Client {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		serve(ctx, conn)
	}))
	t.Cleanup(func() { cancel(); server.Close() })
	return newClient(t, wstest.Connect(t, server.URL, "/"), uuid.NewString())
}
