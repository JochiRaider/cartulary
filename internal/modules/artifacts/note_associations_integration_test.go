package artifacts_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type associationEvidenceRows interface {
	LoadEvidenceTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	RefreshEvidenceTx(context.Context, pgx.Tx, uuid.UUID) error
}

type associationTestRows struct {
	associationEvidenceRows
}

func (associationTestRows) LabelTx(_ context.Context, _ pgx.Tx, r records.Envelope) (string, error) {
	return r.RecordType + " " + r.RecordID.String(), nil
}

// Receipt reads must use the admitted transaction. A second pool checkout can
// deadlock when concurrent callers hold all connections awaiting the same key.
type associationConnectionObserver struct {
	postgres.DB
	active  atomic.Int32
	escaped atomic.Int32
}

func (db *associationConnectionObserver) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if db.active.Load() > 0 {
		db.escaped.Add(1)
	}
	return db.DB.QueryRow(ctx, query, args...)
}
func (db *associationConnectionObserver) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	db.active.Add(1)
	return &associationObservedTx{Tx: tx, owner: db}, nil
}

type associationObservedTx struct {
	pgx.Tx
	owner   *associationConnectionObserver
	release sync.Once
}

func (tx *associationObservedTx) Commit(ctx context.Context) error {
	defer tx.release.Do(func() { tx.owner.active.Add(-1) })
	return tx.Tx.Commit(ctx)
}
func (tx *associationObservedTx) Rollback(ctx context.Context) error {
	defer tx.release.Do(func() { tx.owner.active.Add(-1) })
	return tx.Tx.Rollback(ctx)
}

func TestNoteAssociationResource(t *testing.T) {
	ctx := context.Background()
	h := appsupport.StartStore(t, "note-association-resource")
	actor := authstoretest.SeedLocalUserRecord(t, h.DB, "note-association@example.test", "Notes owner", "NoteAssociationOwner1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "association-incident", "IR-NOTE-ASSOCIATION", "Associations")
	observedDB := &associationConnectionObserver{DB: h.DB}
	t.Cleanup(func() {
		if escaped := observedDB.escaped.Load(); escaped != 0 {
			t.Errorf("%d pool reads escaped an open mutation transaction", escaped)
		}
	})
	f, err := workbookassembly.NewArtifactMutationContribution(observedDB, conflicttest.NewCodec("associations"), revisionsupport.MustAppender(t), revisionsupport.MustConflictFieldResolver(t), appsupport.ArtifactProjectionRows(h.DB), collaborationsupport.NewRecordChangedAppender())
	if err != nil {
		t.Fatal(err)
	}
	o, err := artifacts.NewNoteAssociations(f, admission.NewChecker(h.DB), links.NewStore(), associationTestRows{appsupport.EvidenceProjectionRows(h.DB)})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	source := seedLinkedNoteSource(t, h, incident.ID, actor.ID, "host", now)
	note, err := f.CreateContextualNote(ctx, artifacts.ContextualNoteCreateCommand{ActorUserID: actor.ID, SourceRecordID: source, Admission: mustArtifactContextualNoteAdmission(t, "association-note", map[string]any{"note.title": "Associated note"}, nil), RequestID: "create-note", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	list := func(kind artifacts.NoteAssociationKind, noteID uuid.UUID, limit int, before *time.Time, id uuid.UUID) artifacts.NoteAssociationPage {
		t.Helper()
		p, err := o.List(ctx, artifacts.NoteAssociationRead{ActorUserID: actor.ID, IncidentID: incident.ID, NoteID: noteID, Kind: kind, Limit: limit, BeforeTime: before, BeforeID: id})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	command := func(kind artifacts.NoteAssociationKind, version int64, txn string, actions ...map[string]any) artifacts.NoteAssociationCommand {
		t.Helper()
		data, _ := json.Marshal(map[string]any{"kind": kind, "base_row_version": version, "client_txn_id": txn, "actions": actions})
		a, err := artifacts.AdmitNoteAssociations(strings.NewReader(string(data)))
		if err != nil {
			t.Fatal(err)
		}
		return artifacts.NoteAssociationCommand{ActorUserID: actor.ID, IncidentID: incident.ID, NoteID: note.RecordID, Admission: a, RequestID: txn, Now: now.Add(time.Minute)}
	}
	add := func(id uuid.UUID) map[string]any {
		return map[string]any{"op": "add", "counterpart_record_id": id.String()}
	}
	remove := func(ref string) map[string]any { return map[string]any{"op": "remove", "item_ref": ref} }
	page := list(artifacts.NoteAssociationSource, note.RecordID, 100, nil, uuid.Nil)
	if len(page.Items) != 1 || page.Items[0].CounterpartID != source || page.Items[0].Direction != "incoming" {
		t.Fatalf("existing contextual source missing: %#v", page)
	}
	originalRef := page.Items[0].ItemRef
	noop := command(artifacts.NoteAssociationSource, 1, "duplicate-source", add(source))
	result, err := o.Apply(ctx, noop)
	if err != nil {
		t.Fatal(err)
	}
	if result.ChangeSetID != nil || result.RowVersion != 1 {
		t.Fatalf("duplicate add changed record: %#v", result)
	}
	for i := 0; i < 2; i++ {
		sourceID := seedLinkedNoteSource(t, h, incident.ID, actor.ID, "identity", now)
		result, err = o.Apply(ctx, command(artifacts.NoteAssociationSource, result.RowVersion, fmt.Sprintf("add-source-%d", i), add(sourceID)))
		if err != nil {
			t.Fatal(err)
		}
	}
	replay, err := o.Apply(ctx, noop)
	if err != nil || replay.Outcome != artifacts.MutationOutcomeReplayed || replay.RowVersion != 1 {
		t.Fatalf("no-op replay lost original receipt: %#v %v", replay, err)
	}
	page = list(artifacts.NoteAssociationSource, note.RecordID, 1, nil, uuid.Nil)
	seen := map[string]bool{}
	for {
		for _, item := range page.Items {
			if seen[item.ItemRef] {
				t.Fatal("duplicate pagination item")
			}
			seen[item.ItemRef] = true
		}
		if !page.HasMore {
			break
		}
		stamp := page.LastTime
		page = list(artifacts.NoteAssociationSource, note.RecordID, 1, &stamp, page.LastID)
	}
	if len(seen) != 3 {
		t.Fatalf("paged %d associations", len(seen))
	}
	before := result.RowVersion
	if _, err = o.Apply(ctx, command(artifacts.NoteAssociationSource, before, "atomic-invalid", remove(originalRef), remove(originalRef))); err == nil {
		t.Fatal("duplicate removal accepted")
	}
	page = list(artifacts.NoteAssociationSource, note.RecordID, 100, nil, uuid.Nil)
	if page.RowVersion != before || len(page.Items) != 3 {
		t.Fatal("rejected request partially committed")
	}
	if _, err = o.Apply(ctx, command(artifacts.NoteAssociationSource, before, "self-link", add(note.RecordID))); err == nil {
		t.Fatal("self link accepted")
	}
	evidence := seedLinkedNoteSource(t, h, incident.ID, actor.ID, "evidence", now)
	if _, err := h.DB.Exec(ctx, `INSERT INTO evidence(record_id,incident_id,title) VALUES($1,$2,'Associated evidence')`, evidence, incident.ID); err != nil {
		t.Fatal(err)
	}
	evidenceTx, err := h.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := appsupport.EvidenceProjectionRows(h.DB).RefreshEvidenceTx(ctx, evidenceTx, evidence); err != nil {
		t.Fatal(err)
	}
	if err := evidenceTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	result, err = o.Apply(ctx, command(artifacts.NoteAssociationEvidence, before, "attach-evidence", add(evidence)))
	if err != nil {
		t.Fatal(err)
	}
	page = list(artifacts.NoteAssociationEvidence, note.RecordID, 100, nil, uuid.Nil)
	if len(page.Items) != 1 || page.Items[0].Direction != "outgoing" {
		t.Fatal("missing outgoing evidence")
	}
	second, err := f.CreateContextualNote(ctx, artifacts.ContextualNoteCreateCommand{ActorUserID: actor.ID, SourceRecordID: source, Admission: mustArtifactContextualNoteAdmission(t, "association-second-note", map[string]any{"note.title": "Second note"}, nil), RequestID: "create-second", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	result, err = o.Apply(ctx, command(artifacts.NoteAssociationRelatedNote, result.RowVersion, "related-note", add(second.RecordID)))
	if err != nil {
		t.Fatal(err)
	}
	incoming := list(artifacts.NoteAssociationRelatedNote, second.RecordID, 100, nil, uuid.Nil)
	if len(incoming.Items) != 1 || incoming.Items[0].Direction != "incoming" || incoming.Items[0].CounterpartID != note.RecordID {
		t.Fatalf("incoming traversal: %#v", incoming)
	}
	c := command(artifacts.NoteAssociationRelatedNote, incoming.RowVersion, "incoming-remove", remove(incoming.Items[0].ItemRef))
	c.NoteID = second.RecordID
	if _, err = o.Apply(ctx, c); err == nil {
		t.Fatal("incoming reference mutated from target Note")
	}
	if _, err = o.Apply(ctx, command(artifacts.NoteAssociationSource, 1, "stale-edit", add(source))); err == nil {
		t.Fatal("stale version accepted")
	}

	foreignIncident := appsupport.CreateIncidentInStore(t, h.DB, actor, "association-foreign", "IR-NOTE-FOREIGN", "Foreign")
	foreign := seedLinkedNoteSource(t, h, foreignIncident.ID, actor.ID, "host", now)
	deleted := seedLinkedNoteSource(t, h, incident.ID, actor.ID, "host", now)
	if _, err := h.DB.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, deleted, now, actor.ID); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []struct {
		kind artifacts.NoteAssociationKind
		id   uuid.UUID
	}{{artifacts.NoteAssociationSource, foreign}, {artifacts.NoteAssociationSource, deleted}, {artifacts.NoteAssociationEvidence, source}, {artifacts.NoteAssociationRelatedNote, source}, {artifacts.NoteAssociationSource, uuid.New()}} {
		if _, err := o.Apply(ctx, command(invalid.kind, result.RowVersion, "invalid-"+invalid.id.String(), add(invalid.id))); err == nil {
			t.Fatalf("invalid counterpart accepted: %#v", invalid)
		}
	}
	if _, err := o.Apply(ctx, command(artifacts.NoteAssociationSource, result.RowVersion, "duplicate-source", add(source))); !errors.Is(err, artifacts.ErrClientTxnConflict) {
		t.Fatalf("divergent replay: %v", err)
	}
	appsupport.RequireLinksPortableRoundTrip(t, h.DB, incident.ID, actor.ID)
	projectionTx, err := h.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := appsupport.ArtifactProjectionRows(h.DB).RefreshArtifactTx(ctx, projectionTx, note.RecordID); err != nil {
		t.Fatal(err)
	}
	rebuilt, err := appsupport.ArtifactProjectionRows(h.DB).LoadArtifactTx(ctx, projectionTx, artifacts.NotesViewSchemaID, note.RecordID)
	if err != nil {
		t.Fatal(err)
	}
	expectedJSON, _ := json.Marshal(result.Row)
	rebuiltJSON, _ := json.Marshal(rebuilt)
	if !bytes.Equal(expectedJSON, rebuiltJSON) {
		t.Fatalf("projection rebuild changed Note: %s -> %s", expectedJSON, rebuiltJSON)
	}
	if err := projectionTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	concurrent := command(artifacts.NoteAssociationSource, result.RowVersion, "concurrent-replay", add(source))
	type outcome struct {
		result artifacts.MutationResult
		err    error
	}
	outcomes := make(chan outcome, 2)
	for i := 0; i < 2; i++ {
		go func() { r, e := o.Apply(ctx, concurrent); outcomes <- outcome{r, e} }()
	}
	first, secondResult := <-outcomes, <-outcomes
	if first.err != nil || secondResult.err != nil || first.result.RowVersion != secondResult.result.RowVersion || first.result.Outcome == secondResult.result.Outcome {
		t.Fatalf("concurrent exact replay: %#v %#v", first, secondResult)
	}
	if _, err := h.DB.Exec(ctx, `UPDATE incident_memberships SET role='viewer' WHERE incident_id=$1 AND user_id=$2`, incident.ID, actor.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := o.Apply(ctx, command(artifacts.NoteAssociationSource, result.RowVersion, "viewer-edit", add(source))); !admission.IsDenied(err, admission.DenialInsufficientRole) {
		t.Fatalf("viewer mutation: %v", err)
	}
	if len(list(artifacts.NoteAssociationSource, note.RecordID, 100, nil, uuid.Nil).Items) != 3 {
		t.Fatal("read-only role lost associations")
	}
	if _, err := h.DB.Exec(ctx, `UPDATE incident_memberships SET role='admin' WHERE incident_id=$1 AND user_id=$2`, incident.ID, actor.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = h.DB.Exec(ctx, `UPDATE incidents SET status='closed',closed_at=now() WHERE id=$1`, incident.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = o.Apply(ctx, command(artifacts.NoteAssociationSource, result.RowVersion, "closed-edit", add(source))); !admission.IsDenied(err, admission.DenialIncidentClosed) {
		t.Fatalf("closed incident mutation: %v", err)
	}
	if len(list(artifacts.NoteAssociationSource, note.RecordID, 100, nil, uuid.Nil).Items) != 3 {
		t.Fatal("closed incident lost readable associations")
	}
	if replay, err := o.Apply(ctx, noop); err != nil || replay.Outcome != artifacts.MutationOutcomeReplayed || replay.RowVersion != 1 {
		t.Fatalf("closed incident exact receipt replay: %#v %v", replay, err)
	}
	if _, err := h.DB.Exec(ctx, `DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, incident.ID, actor.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := o.List(ctx, artifacts.NoteAssociationRead{ActorUserID: actor.ID, IncidentID: incident.ID, NoteID: note.RecordID, Kind: artifacts.NoteAssociationSource, Limit: 1}); !admission.IsDenied(err, admission.DenialNotVisible) {
		t.Fatalf("revoked page read: %v", err)
	}
}
