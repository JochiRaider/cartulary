package ws

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
	"github.com/google/uuid"
)

func TestHubSessionRevocationSubscribers(t *testing.T) {
	t.Run("revoke notifies each registered listener exactly once", func(t *testing.T) {
		hub := NewHub()
		sessionID := uuid.New()

		first, unregisterFirst := hub.RegisterSession(sessionID)
		defer unregisterFirst()
		second, unregisterSecond := hub.RegisterSession(sessionID)
		defer unregisterSecond()

		hub.RevokeSession(sessionID, "session_revoked")

		requireRevocationReason(t, first, "session_revoked")
		requireRevocationReason(t, second, "session_revoked")
		requireNoRevocationReason(t, first)
		requireNoRevocationReason(t, second)

		hub.RevokeSession(sessionID, "ignored")
		requireNoRevocationReason(t, first)
		requireNoRevocationReason(t, second)
	})

	t.Run("unregister prevents later delivery", func(t *testing.T) {
		hub := NewHub()
		sessionID := uuid.New()

		revocations, unregister := hub.RegisterSession(sessionID)
		unregister()

		other, unregisterOther := hub.RegisterSession(sessionID)
		defer unregisterOther()

		hub.RevokeSession(sessionID, "session_revoked")

		requireNoRevocationReason(t, revocations)
		requireRevocationReason(t, other, "session_revoked")
	})
}

func TestHubIncidentTerminalSubscribers(t *testing.T) {
	hub := NewHub()
	incidentID := uuid.New()

	first, unregisterFirst := hub.RegisterIncidentTerminal(incidentID)
	defer unregisterFirst()
	second, unregisterSecond := hub.RegisterIncidentTerminal(incidentID)
	defer unregisterSecond()

	hub.TerminateIncident(incidentID, IncidentTerminalClosed)

	requireRevocationReason(t, first, IncidentTerminalClosed)
	requireRevocationReason(t, second, IncidentTerminalClosed)
	requireNoRevocationReason(t, first)
	requireNoRevocationReason(t, second)

	hub.TerminateIncident(incidentID, "ignored")
	requireNoRevocationReason(t, first)
	requireNoRevocationReason(t, second)
}

func TestWSContractJobProgressPayloadShape(t *testing.T) {
	artifact, ok := gencontracts.WSArtifactsIndex["contracts/ws/index.schema.json"]
	if !ok {
		t.Fatal("missing generated websocket contract artifact")
	}

	var document map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode generated websocket contract artifact: %v", err)
	}

	message := findContractMessage(t, document, "job_progress")
	if message["direction"] != "server_to_client" {
		t.Fatalf("job_progress direction = %v, want server_to_client", message["direction"])
	}
	if message["replayable"] != true {
		t.Fatalf("job_progress replayable = %v, want true", message["replayable"])
	}

	payloadSchema := requireObject(t, message, "payload_schema")
	required := requireStringArray(t, payloadSchema, "required")
	wantRequired := []string{"job_id", "scope", "status", "progress", "updated_at"}
	if !reflect.DeepEqual(required, wantRequired) {
		t.Fatalf("job_progress required fields = %#v, want %#v", required, wantRequired)
	}
	for _, field := range required {
		if field == "state" {
			t.Fatal("job_progress must not require stale state field")
		}
	}

	properties := requireObject(t, payloadSchema, "properties")
	if _, ok := properties["state"]; ok {
		t.Fatal("job_progress must not define stale state field")
	}
	scope := requireObject(t, properties, "scope")
	scopeProperties := requireObject(t, scope, "properties")
	kind := requireObject(t, scopeProperties, "kind")
	if got := requireStringArray(t, kind, "enum"); !reflect.DeepEqual(got, []string{"incident", "deployment"}) {
		t.Fatalf("job_progress scope.kind enum = %#v", got)
	}
	status := requireObject(t, properties, "status")
	if got := requireStringArray(t, status, "enum"); !reflect.DeepEqual(got, []string{
		JobStatusQueued,
		JobStatusRunning,
		JobStatusCancelRequested,
		JobStatusSucceeded,
		JobStatusFailed,
		JobStatusCanceled,
	}) {
		t.Fatalf("job_progress status enum = %#v", got)
	}
	progress := requireObject(t, properties, "progress")
	if got := requireStringArray(t, progress, "required"); !reflect.DeepEqual(got, []string{"completed", "total"}) {
		t.Fatalf("job_progress progress required fields = %#v", got)
	}
}

func TestHubPublishJobProgress(t *testing.T) {
	t.Run("emits typed incident scoped payload", func(t *testing.T) {
		hub := NewHub()
		incidentID := uuid.New()
		messages, unsubscribe := hub.SubscribeIncident(incidentID, 1)
		defer unsubscribe()

		total := int64(4)
		now := time.Date(2026, 4, 24, 12, 30, 0, 123, time.UTC)
		cancelable := true
		payload := NewIncidentJobProgressPayload("job-1", incidentID, JobStatusRunning, JobProgress{
			Completed: 1,
			Total:     &total,
		}, now)
		payload.Cancelable = &cancelable
		payload.Message = "Importing rows"

		if err := hub.PublishJobProgress(incidentID, payload); err != nil {
			t.Fatalf("publish job_progress: %v", err)
		}

		message := requireIncidentMessage(t, messages)
		if message.Type != "job_progress" {
			t.Fatalf("message type = %q, want job_progress", message.Type)
		}
		if message.IncidentID != incidentID.String() {
			t.Fatalf("message incident_id = %q, want %q", message.IncidentID, incidentID)
		}
		if message.StreamSeq == nil || *message.StreamSeq != 1 {
			t.Fatalf("message stream_seq = %v, want 1", message.StreamSeq)
		}

		var got map[string]any
		if err := json.Unmarshal(message.Payload, &got); err != nil {
			t.Fatalf("decode job_progress payload: %v", err)
		}
		if got["job_id"] != "job-1" || got["status"] != JobStatusRunning || got["updated_at"] != now.Format(time.RFC3339Nano) {
			t.Fatalf("unexpected job_progress payload: %#v", got)
		}
		scope := got["scope"].(map[string]any)
		if scope["kind"] != JobScopeKindIncident || scope["incident_id"] != incidentID.String() {
			t.Fatalf("unexpected job_progress scope: %#v", scope)
		}
		progress := got["progress"].(map[string]any)
		if progress["completed"] != float64(1) || progress["total"] != float64(4) {
			t.Fatalf("unexpected job_progress progress: %#v", progress)
		}
		if got["state"] != nil {
			t.Fatalf("job_progress payload must not include stale state field: %#v", got)
		}
	})

	t.Run("rejects deployment scoped payload on incident stream", func(t *testing.T) {
		hub := NewHub()
		incidentID := uuid.New()
		messages, unsubscribe := hub.SubscribeIncident(incidentID, 1)
		defer unsubscribe()

		err := hub.PublishJobProgress(incidentID, JobProgressPayload{
			JobID: "job-deployment",
			Scope: JobScope{
				Kind: JobScopeKindDeployment,
			},
			Status:    JobStatusQueued,
			Progress:  JobProgress{Completed: 0},
			UpdatedAt: time.Now().UTC(),
		})
		if err == nil {
			t.Fatal("expected deployment scoped job_progress to be rejected")
		}
		if got := hub.HighWater(incidentID); got != 0 {
			t.Fatalf("deployment scoped job_progress advanced high water to %d", got)
		}
		requireNoIncidentMessage(t, messages)
	})
}

func requireRevocationReason(t testing.TB, revocations <-chan string, want string) {
	t.Helper()

	select {
	case got := <-revocations:
		if got != want {
			t.Fatalf("unexpected revocation reason: got %q want %q", got, want)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for revocation reason %q", want)
	}
}

func requireNoRevocationReason(t testing.TB, revocations <-chan string) {
	t.Helper()

	select {
	case got := <-revocations:
		t.Fatalf("unexpected revocation reason: got %q", got)
	default:
	}
}

func requireIncidentMessage(t testing.TB, messages <-chan Message) Message {
	t.Helper()

	select {
	case got := <-messages:
		return got
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for websocket incident message")
		return Message{}
	}
}

func requireNoIncidentMessage(t testing.TB, messages <-chan Message) {
	t.Helper()

	select {
	case got := <-messages:
		t.Fatalf("unexpected websocket incident message: %#v", got)
	default:
	}
}

func findContractMessage(t testing.TB, document map[string]any, messageType string) map[string]any {
	t.Helper()

	properties := requireObject(t, document, "properties")
	messages := requireObject(t, properties, "messages")
	defaultMessages, ok := messages["default"].([]any)
	if !ok {
		t.Fatalf("contract messages.default was %T, want array", messages["default"])
	}
	for _, raw := range defaultMessages {
		message, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("contract message was %T, want object", raw)
		}
		if message["type"] == messageType {
			return message
		}
	}
	t.Fatalf("missing contract message %q", messageType)
	return nil
}

func requireObject(t testing.TB, parent map[string]any, key string) map[string]any {
	t.Helper()

	value, ok := parent[key].(map[string]any)
	if !ok {
		t.Fatalf("%s was %T, want object", key, parent[key])
	}
	return value
}

func requireStringArray(t testing.TB, parent map[string]any, key string) []string {
	t.Helper()

	rawValues, ok := parent[key].([]any)
	if !ok {
		t.Fatalf("%s was %T, want array", key, parent[key])
	}
	values := make([]string, 0, len(rawValues))
	for _, raw := range rawValues {
		value, ok := raw.(string)
		if !ok {
			t.Fatalf("%s item was %T, want string", key, raw)
		}
		values = append(values, value)
	}
	return values
}
