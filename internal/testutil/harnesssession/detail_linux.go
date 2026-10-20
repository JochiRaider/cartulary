//go:build linux

package harnesssession

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

const detailSchema = "cartulary.harness_detail_authority.v2"
const detailMarker = "CARTULARY_HARNESS_DETAIL_TOKEN"

// Detail authority is independent of the external-resource journal. Corruption
// of that journal cannot erase admitted producer identities or hold safe detail.
type detailAuthority struct {
	Schema    string                 `json:"schema_id"`
	RunID     string                 `json:"run_id"`
	LeaseID   string                 `json:"lease_id"`
	ID        string                 `json:"id"`
	Owner     harnessadmission.Proof `json:"owner"`
	Scope     diagnosticScope        `json:"scope"`
	State     string                 `json:"state"`
	Producers []detailProducer       `json:"producers"`
}

// Launch binds registration to the runtime owner's exact direct child while
// ancestry is observable. Nested producers may create other groups/sessions;
// their inherited responsibility remains under the detail scope after retirement.
type detailProducer struct {
	Proof  ProcessProof  `json:"proof"`
	Launch *ProcessProof `json:"launch"`
	State  string        `json:"state"`
}

func (r *Runtime) detailFile(id string) string {
	if id == "ordinary" {
		return filepath.Join(r.Root, "detail-authority.json")
	}
	return filepath.Join(r.Root, "recovery-attempts", id+".json")
}
func (r *Runtime) detailRoot(id string) string {
	if id == "ordinary" {
		return filepath.Join(r.Root, "detail")
	}
	return filepath.Join(r.Root, "recovery-detail", id)
}
func (r *Runtime) readDetail(id string) (detailAuthority, error) {
	var authority detailAuthority
	if id != "ordinary" && !uuidPattern.MatchString(id) {
		return authority, harnesslocal.ErrUnsafe
	}
	data, err := harnesslocal.ReadFile(r.detailFile(id), MaximumOwner, false)
	if err != nil {
		return authority, err
	}
	v, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate(detailSchema, v) != nil || json.Unmarshal(data, &authority) != nil {
		return authority, harnesslocal.ErrUnsafe
	}
	if authority.RunID != r.RunID || authority.LeaseID != r.LeaseID || authority.ID != id || authority.Scope.UID != os.Getuid() {
		return authority, harnesslocal.ErrUnsafe
	}
	return authority, nil
}
func (r *Runtime) writeDetail(authority detailAuthority) error {
	dir, err := harnesslocal.PrivateDirectory(filepath.Dir(r.detailFile(authority.ID)))
	if err != nil {
		return err
	}
	defer dir.Close()
	data, err := validateJSON(detailSchema, authority)
	if err != nil {
		return err
	}
	return harnesslocal.ReplaceFileAt(dir, filepath.Base(r.detailFile(authority.ID)), append(data, '\n'))
}
func (r *Runtime) createDetail(id string, owner harnessadmission.Proof) (detailAuthority, error) {
	scope, err := newProcessScope()
	if err != nil {
		return detailAuthority{}, err
	}
	v := detailAuthority{Schema: detailSchema, RunID: r.RunID, LeaseID: r.LeaseID, ID: id, Owner: owner, Scope: scope, State: "open", Producers: []detailProducer{}}
	return v, r.writeDetail(v)
}
func (r *Runtime) actorProof() (harnessadmission.Proof, error) {
	if r.actor.PID != 0 {
		return r.actor, nil
	}
	return harnessadmission.Identity(os.Getpid())
}

// Native admission grants a recovery capability only to the owning recovery
// process. Its fixed physical workers may inherit it; ordinary borrowing cannot
// manufacture it or substitute a PID in the private protocol.
func attemptCaller(authority detailAuthority, actor harnessadmission.Proof) error {
	if authority.State != "open" {
		return harnesslocal.ErrUnsafe
	}
	alive, err := harnessadmission.Alive(authority.Owner)
	if err != nil || !alive {
		return ErrUnresolved
	}
	current, err := harnessadmission.Identity(actor.PID)
	if err != nil || current != actor || actor.Boot != authority.Owner.Boot {
		return harnesslocal.ErrUnsafe
	}
	pid := actor.PID
	for range 128 {
		process, err := readProcess(pid)
		if err != nil {
			return err
		}
		if pid == authority.Owner.PID {
			if process.Start == authority.Owner.Start {
				return nil
			}
			return harnesslocal.ErrUnsafe
		}
		if process.parent < 2 || process.parent == pid {
			break
		}
		pid = process.parent
	}
	return harnesslocal.ErrUnsafe
}
func (r *Runtime) bindRecoveryAttempt(id string) error {
	unlock, err := r.lock()
	if err != nil {
		return err
	}
	defer unlock()
	if id == "ordinary" {
		return harnesslocal.ErrUnsafe
	}
	v, err := r.readDetail(id)
	if err != nil {
		return err
	}
	actor, err := r.actorProof()
	if err != nil {
		return err
	}
	if err := attemptCaller(v, actor); err != nil {
		return err
	}
	r.attempt = id
	return nil
}
func (r *Runtime) productionMutable() error {
	if r.attempt == "" {
		if err := r.mutable(); err != nil {
			return err
		}
		v, err := r.readDetail("ordinary")
		if err != nil || v.State != "open" {
			return errors.Join(err, harnesslocal.ErrUnsafe)
		}
		return nil
	}
	v, err := r.readDetail(r.attempt)
	if err != nil {
		return err
	}
	actor, err := r.actorProof()
	if err != nil {
		return err
	}
	return attemptCaller(v, actor)
}

func (r *Runtime) detailPath(id string, parts ...string) (string, error) {
	path, err := r.RecoveryPath(parts...)
	if err != nil {
		return "", err
	}
	unlock, err := r.lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	if id == "ordinary" {
		if err := r.mutable(); err != nil {
			return "", err
		}
	}
	v, err := r.readDetail(id)
	if err != nil || v.State != "open" {
		return "", errors.Join(err, harnesslocal.ErrUnsafe)
	}
	actor, err := r.actorProof()
	if err != nil {
		return "", err
	}
	if id != "ordinary" {
		if err := attemptCaller(v, actor); err != nil {
			return "", err
		}
	}
	proof, err := CaptureProcess(actor.PID)
	if err != nil || proof.Boot != actor.Boot || proof.Start != actor.Start {
		return "", errors.Join(err, harnesslocal.ErrUnsafe)
	}
	found := false
	for _, producer := range v.Producers {
		if producer.Proof == proof {
			if producer.State != "active" {
				return "", harnesslocal.ErrUnsafe
			}
			found = true
		}
	}
	if !found {
		if len(v.Producers) >= 4096 {
			return "", harnesslocal.ErrUnsafe
		}
		launch, err := enclosingDetailLaunch(proof, v.Owner, v.Scope)
		if err != nil {
			return "", err
		}
		v.Producers = append(v.Producers, detailProducer{Proof: proof, Launch: launch, State: "active"})
		if err := r.writeDetail(v); err != nil {
			return "", err
		}
	}
	dir, err := harnesslocal.PrivateDirectory(r.detailRoot(id))
	if err != nil {
		return "", err
	}
	dir.Close()
	relative, _ := filepath.Rel(r.Root, path)
	return filepath.Join(r.detailRoot(id), relative), nil
}

// An unanchored registration remains active for explicit settlement. Neither
// reparenting nor a group ID alone can manufacture a supervisor relationship.
func enclosingDetailLaunch(producer ProcessProof, owner harnessadmission.Proof, scope diagnosticScope) (*ProcessProof, error) {
	if producer.PID == owner.PID {
		return nil, nil
	}
	chain := []processState{}
	pid := producer.PID
	for range 128 {
		member, err := readProcess(pid)
		if processGone(err) {
			return nil, nil
		}
		if err != nil {
			return nil, atIdentityBoundary("process_read", err)
		}
		if len(chain) == 0 && !sameProcess(member.ProcessProof, producer) {
			return nil, atIdentityBoundary("producer_identity", ErrUnresolved)
		}
		chain = append(chain, member)
		if member.parent == owner.PID {
			alive, err := harnessadmission.Alive(owner)
			if err != nil || !alive || owner.Boot != producer.Boot {
				return nil, atIdentityBoundary("group_anchor", errors.Join(err, ErrUnresolved))
			}
			for _, captured := range chain {
				current, err := readProcess(captured.PID)
				if err != nil || !sameProcess(current.ProcessProof, captured.ProcessProof) || current.parent != captured.parent {
					return nil, atIdentityBoundary("group_anchor", errors.Join(err, ErrUnresolved))
				}
			}
			if !member.Group {
				return nil, nil
			}
			// Only inherited scope authority can retain responsibility for a
			// producer outside the group this supervisor physically drains.
			// Unmarked producers keep their individual outstanding proof.
			if producer.PGID != member.PGID || producer.SID != member.SID {
				current, marked, _, err := scopedMember(scope, producer.PID, detailMarker)
				if err != nil {
					return nil, atIdentityBoundary("descendant_membership", err)
				}
				if !marked || !sameProcess(current.ProcessProof, producer) {
					return nil, nil
				}
			}
			launch := member.ProcessProof
			launch.Boot = producer.Boot
			return &launch, nil
		}
		if member.parent <= 1 || member.parent == pid {
			return nil, nil
		}
		pid = member.parent
	}
	return nil, atIdentityBoundary("group_anchor", ErrUnresolved)
}

// RetireDetailProducers is called by the native supervisor before reaping its
// child. It does not signal anything. A live, foreign, reused or missing group
// anchor cannot authorize retirement, even if a previous settlement succeeded.
func (r *Runtime) RetireDetailProducers(launch ProcessProof) error {
	return r.retireDetailProducers(launch, r.writeDetail)
}

func (r *Runtime) retireDetailProducers(launch ProcessProof, publish func(detailAuthority) error) error {
	unlock, err := r.lock()
	if err != nil {
		return err
	}
	defer unlock()
	current, err := CaptureProcess(launch.PID)
	if err != nil || current != launch || !launch.Group {
		return atIdentityBoundary("producer_identity", errors.Join(err, ErrUnresolved))
	}
	leader, err := readProcess(launch.PID)
	if err != nil || leader.parent != os.Getpid() || (leader.state != "Z" && leader.state != "X") {
		return atIdentityBoundary("group_anchor", errors.Join(err, ErrUnresolved))
	}
	members, err := processMembers(launch)
	if err != nil {
		return err
	}
	for _, member := range members {
		if member.state != "Z" && member.state != "X" {
			return atIdentityBoundary("group_anchor", ErrUnresolved)
		}
	}
	id := "ordinary"
	if r.attempt != "" {
		id = r.attempt
	}
	v, err := r.readDetail(id)
	if err != nil {
		return err
	}
	changed := false
	for i := range v.Producers {
		producer := &v.Producers[i]
		if producer.State == "active" && producer.Launch != nil && *producer.Launch == launch {
			producer.State = "retired"
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return AtCleanupStage("detail_authority", publish(v))
}

func (r *Runtime) fenceDetails() error {
	if !r.owned {
		return harnesslocal.ErrUnsafe
	}
	if err := harnesslocal.ReplaceFileAt(r.Directory, "finalizing", []byte("sealed\n")); err != nil {
		return err
	}
	v, err := r.readDetail("ordinary")
	if err != nil {
		return err
	}
	if v.State == "open" {
		v.State = "fenced"
		return r.writeDetail(v)
	}
	return nil
}

// purgeDetail runs under the shared directory lock. Allocation is already
// fenced; retained paths/FDs are safe only after their admitted producers and
// inherited marked descendants have physically settled.
func (r *Runtime) purgeDetail(id string) error {
	v, err := r.readDetail(id)
	if err != nil {
		return err
	}
	if v.State == "open" {
		v.State = "fenced"
		if err := r.writeDetail(v); err != nil {
			return err
		}
	}
	if v.State != "purged" {
		actor, err := r.actorProof()
		if err != nil {
			return err
		}
		excluded := map[int]ProcessProof{}
		for _, pid := range []int{os.Getpid(), actor.PID} {
			proof, err := CaptureProcess(pid)
			if err != nil {
				return err
			}
			excluded[pid] = proof
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cleanupDeadlinesMS["helper_process"])*time.Millisecond)
		defer cancel()
		var failures []error
		retired := false
		for ordinal, producer := range v.Producers {
			if producer.State == "retired" {
				continue
			}
			proof := producer.Proof
			if exempt, ok := excluded[proof.PID]; ok && exempt.Boot == proof.Boot && sameProcess(exempt, proof) {
				continue
			}
			if err := StopProcess(ctx, proof, 200*time.Millisecond); err != nil {
				failures = append(failures, AtCleanupStage("detail_producer", atDetailProducer(ordinal+1, err)))
				continue
			}
			v.Producers[ordinal].State = "retired"
			retired = true
		}
		if retired {
			failures = append(failures, AtCleanupStage("detail_authority", r.writeDetail(v)))
		}
		if err := settleScope(ctx, v.Scope, detailMarker, excluded); err != nil {
			failures = append(failures, AtCleanupStage("detail_descendants", atIdentityBoundary("descendant_membership", err)))
		}
		if err := errors.Join(failures...); err != nil {
			return err
		}
	}
	root := r.detailRoot(id)
	dir, err := harnesslocal.OpenDirectory(root, true)
	if err == nil {
		err = RemoveOwnedTree(dir, root)
		dir.Close()
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return AtCleanupStage("detail_remove", err)
	}
	v.State = "purged"
	return AtCleanupStage("detail_authority", r.writeDetail(v))
}

func (r *Runtime) BeginRecoveryAttempt() (string, error) {
	unlock, err := r.lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	if !r.owned || r.attempt != "" {
		return "", harnesslocal.ErrUnsafe
	}
	if err := r.fenceDetails(); err != nil {
		return "", err
	}
	actor, err := r.actorProof()
	if err != nil {
		return "", err
	}
	dir, err := harnesslocal.PrivateDirectory(filepath.Join(r.Root, "recovery-attempts"))
	if err != nil {
		return "", err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(129)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if len(names) >= 128 {
		return "", harnesslocal.ErrUnsafe
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			return "", harnesslocal.ErrUnsafe
		}
		id := strings.TrimSuffix(name, ".json")
		v, err := r.readDetail(id)
		if err != nil {
			return "", err
		}
		if v.State == "purged" {
			continue
		}
		alive, err := harnessadmission.Alive(v.Owner)
		if err != nil || alive {
			return "", ErrUnresolved
		}
		if err := r.purgeDetail(id); err != nil {
			return "", err
		}
	}
	id := identifier()
	if _, err := r.createDetail(id, actor); err != nil {
		return "", err
	}
	r.attempt = id
	r.cleanupMu.Lock()
	r.cleanupAttempts = nil
	r.cleanupMu.Unlock()
	return id, nil
}
func (r *Runtime) InRecoveryAttempt() bool { return r.attempt != "" }
func (r *Runtime) RecoveryDetailPath(parts ...string) (string, error) {
	if r.attempt == "" {
		return "", harnesslocal.ErrUnsafe
	}
	return r.detailPath(r.attempt, parts...)
}
func (r *Runtime) FinishRecoveryAttempt() error {
	unlock, err := r.lock()
	if err != nil {
		return err
	}
	defer unlock()
	if !r.owned || r.attempt == "" {
		return harnesslocal.ErrUnsafe
	}
	v, err := r.readDetail(r.attempt)
	if err != nil {
		return err
	}
	actor, err := r.actorProof()
	if err != nil || v.Owner != actor {
		return harnesslocal.ErrUnsafe
	}
	err = r.purgeDetail(r.attempt)
	r.attempt = ""
	return err
}

func (r *Runtime) purgeRecoveryDetails() error {
	return r.settleRecoveryDetails(true)
}

// Ordinary closure may settle its own recovery attempt. Opportunistic stale
// cleanup cannot claim that authority, even when running in the same process.
func (r *Runtime) settleRecoveryDetails(allowActor bool) error {
	dir, err := harnesslocal.OpenDirectory(filepath.Join(r.Root, "recovery-attempts"), true)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(129)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if len(names) > 128 {
		return harnesslocal.ErrUnsafe
	}
	actor, err := r.actorProof()
	if err != nil {
		return err
	}
	var failures []error
	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			failures = append(failures, harnesslocal.ErrUnsafe)
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		v, err := r.readDetail(id)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if v.State == "purged" {
			continue
		}
		if !allowActor || v.Owner != actor {
			alive, err := harnessadmission.Alive(v.Owner)
			if err != nil || alive {
				failures = append(failures, ErrUnresolved)
				continue
			}
		}
		failures = append(failures, r.purgeDetail(id))
	}
	r.attempt = ""
	return errors.Join(failures...)
}
