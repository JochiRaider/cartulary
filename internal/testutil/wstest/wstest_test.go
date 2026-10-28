package wstest

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestTransportHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transport" || r.Header.Get("X-Test-Transport") != "accepted" {
			http.Error(w, "rejected", http.StatusForbidden)
			return
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		var value map[string]any
		if err := wsjson.Read(ctx, conn, &value); err != nil {
			return
		}
		if err := wsjson.Write(ctx, conn, value); err != nil {
			return
		}
		for range 2 {
			kind, payload, err := conn.Read(ctx)
			if err != nil {
				return
			}
			if err := conn.Write(ctx, kind, payload); err != nil {
				return
			}
		}
		_ = conn.Close(websocket.StatusPolicyViolation, "transport_complete")
	}))
	t.Cleanup(server.Close)
	_, response, err := TryConnect(server.URL, "/transport", nil)
	if err == nil || response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("upgrade rejection got response=%v err=%v", response, err)
	}
	client := ConnectWithHeaders(t, server.URL, "/transport", http.Header{"X-Test-Transport": []string{"accepted"}})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.WriteJSON(ctx, map[string]any{"value": "unchanged", "null": nil}); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := client.ReadJSON(ctx, &got); err != nil {
		t.Fatal(err)
	}
	if value, present := got["null"]; got["value"] != "unchanged" || !present || value != nil {
		t.Fatalf("JSON changed: %#v", got)
	}
	for _, frame := range []struct {
		kind    MessageType
		payload []byte
	}{{MessageText, []byte("arbitrary text\n")}, {MessageBinary, []byte{0, 255, 1}}} {
		if err := client.Write(ctx, frame.kind, frame.payload); err != nil {
			t.Fatal(err)
		}
		kind, payload, err := client.Read(ctx)
		if err != nil || kind != frame.kind || !bytes.Equal(payload, frame.payload) {
			t.Fatalf("raw frame changed: %v %q %v", kind, payload, err)
		}
	}
	_, _, err = client.Read(ctx)
	RequireClose(t, err, StatusPolicyViolation, "transport_complete")
	client.Close(StatusNormalClosure, "again")
	abrupt := ConnectWithHeaders(t, server.URL, "/transport", http.Header{"X-Test-Transport": []string{"accepted"}})
	if err := abrupt.CloseNow(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := abrupt.Read(ctx); err == nil {
		t.Fatal("abruptly closed socket remained readable")
	}
}
