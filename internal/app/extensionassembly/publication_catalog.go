package extensionassembly

import (
	"fmt"
	"sort"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
)

type PublicationCatalog struct {
	contributions map[string]extensions.ContributionPublication
	workers       map[string]extensions.WorkerPublication
	jobs          map[string]extensions.JobKindContract
	participants  map[string]extensions.ParticipantContract
	bindings      map[string]extensions.ImplementationBindingPublication
}

func NewPublicationCatalog(plan extensions.PublicationPlan, participantContracts []extensions.ParticipantContract) (PublicationCatalog, error) {
	catalog := PublicationCatalog{
		contributions: map[string]extensions.ContributionPublication{},
		workers:       map[string]extensions.WorkerPublication{},
		jobs:          map[string]extensions.JobKindContract{},
		participants:  map[string]extensions.ParticipantContract{},
		bindings:      map[string]extensions.ImplementationBindingPublication{},
	}
	participantSources := make(map[string]extensions.ParticipantContract, len(participantContracts))
	for _, participant := range participantContracts {
		if participant.ParticipantID == "" || participant.OwnerProfileID == "" {
			return PublicationCatalog{}, fmt.Errorf("extension publication participant identity is incomplete")
		}
		if _, duplicate := participantSources[participant.ParticipantID]; duplicate {
			return PublicationCatalog{}, fmt.Errorf("duplicate extension publication participant %q", participant.ParticipantID)
		}
		participantSources[participant.ParticipantID] = cloneParticipantContract(participant)
	}

	usedParticipants := map[string]struct{}{}
	for _, contribution := range plan.Contributions() {
		if contribution.ContributionID == "" || contribution.ProfileID == "" ||
			contribution.Kind == "" || contribution.ContributionSHA256 == "" ||
			contribution.ImplementationBindingSHA256 == "" {
			return PublicationCatalog{}, fmt.Errorf("extension publication contribution identity is incomplete")
		}
		if _, duplicate := catalog.contributions[contribution.ContributionID]; duplicate {
			return PublicationCatalog{}, fmt.Errorf("duplicate extension publication contribution %q", contribution.ContributionID)
		}
		catalog.contributions[contribution.ContributionID] = contribution
		if contribution.ParticipantID == "" {
			continue
		}
		participant, present := participantSources[contribution.ParticipantID]
		if !present || participant.OwnerProfileID != contribution.ProfileID {
			return PublicationCatalog{}, fmt.Errorf("extension contribution %q has no exact participant", contribution.ContributionID)
		}
		if _, duplicate := catalog.participants[contribution.ParticipantID]; duplicate {
			return PublicationCatalog{}, fmt.Errorf("duplicate extension publication participant %q", contribution.ParticipantID)
		}
		catalog.participants[contribution.ParticipantID] = participant
		usedParticipants[contribution.ParticipantID] = struct{}{}
	}
	for _, participant := range participantContracts {
		if plan.ResolvedClaims().Contains(participant.OwnerProfileID) {
			if _, used := usedParticipants[participant.ParticipantID]; !used {
				return PublicationCatalog{}, fmt.Errorf("claimed extension participant %q has no contribution", participant.ParticipantID)
			}
		}
	}
	for _, worker := range plan.Workers() {
		if worker.ProfileID == "" || worker.WorkerKind == "" || len(worker.JobKinds) == 0 ||
			worker.MaxActiveAttemptsPerProcess < 1 || worker.MaxActiveAttemptsPerProcess > 8 {
			return PublicationCatalog{}, fmt.Errorf("extension publication worker identity is incomplete")
		}
		for index, jobKind := range worker.JobKinds {
			if jobKind == "" || index > 0 && worker.JobKinds[index-1] >= jobKind {
				return PublicationCatalog{}, fmt.Errorf("extension publication worker %q job assignments are invalid", worker.WorkerKind)
			}
		}
		if _, duplicate := catalog.workers[worker.WorkerKind]; duplicate {
			return PublicationCatalog{}, fmt.Errorf("duplicate extension publication worker %q", worker.WorkerKind)
		}
		worker.JobKinds = append([]string(nil), worker.JobKinds...)
		catalog.workers[worker.WorkerKind] = worker
	}
	for _, binding := range plan.ImplementationBindings() {
		if binding.ProfileID == "" || binding.BindingSHA256 == "" {
			return PublicationCatalog{}, fmt.Errorf("extension publication implementation binding identity is incomplete")
		}
		if _, duplicate := catalog.bindings[binding.ProfileID]; duplicate {
			return PublicationCatalog{}, fmt.Errorf("duplicate extension publication implementation binding %q", binding.ProfileID)
		}
		catalog.bindings[binding.ProfileID] = binding
	}
	for contributionID, contribution := range catalog.contributions {
		binding, present := catalog.bindings[contribution.ProfileID]
		if !present || binding.BindingSHA256 != contribution.ImplementationBindingSHA256 {
			return PublicationCatalog{}, fmt.Errorf("extension contribution %q has no exact implementation binding", contributionID)
		}
	}
	for workerKind, worker := range catalog.workers {
		if _, present := catalog.bindings[worker.ProfileID]; !present {
			return PublicationCatalog{}, fmt.Errorf("extension worker %q has no implementation binding", workerKind)
		}
	}
	for _, job := range plan.JobKindContracts() {
		if job.ProfileID == "" || job.JobKind == "" {
			return PublicationCatalog{}, fmt.Errorf("extension publication job identity is incomplete")
		}
		if _, duplicate := catalog.jobs[job.JobKind]; duplicate {
			return PublicationCatalog{}, fmt.Errorf("duplicate extension publication job %q", job.JobKind)
		}
		catalog.jobs[job.JobKind] = cloneJobKindContract(job)
	}
	for jobKind, job := range catalog.jobs {
		if _, present := catalog.bindings[job.ProfileID]; !present {
			return PublicationCatalog{}, fmt.Errorf("extension job %q has no implementation binding", jobKind)
		}
	}
	assignedJobs := make(map[string]string, len(catalog.jobs))
	for workerKind, worker := range catalog.workers {
		for _, jobKind := range worker.JobKinds {
			job, present := catalog.jobs[jobKind]
			if !present || job.ProfileID != worker.ProfileID {
				return PublicationCatalog{}, fmt.Errorf("extension worker %q has an unknown or cross-profile job %q", workerKind, jobKind)
			}
			if _, duplicate := assignedJobs[jobKind]; duplicate {
				return PublicationCatalog{}, fmt.Errorf("extension job %q has duplicate worker assignments", jobKind)
			}
			assignedJobs[jobKind] = workerKind
		}
	}
	if len(assignedJobs) != len(catalog.jobs) {
		return PublicationCatalog{}, fmt.Errorf("extension worker assignments are incomplete")
	}
	return catalog, nil
}

func (catalog PublicationCatalog) Contribution(contributionID string) (extensions.ContributionPublication, bool) {
	contribution, present := catalog.contributions[contributionID]
	return contribution, present
}

func (catalog PublicationCatalog) Worker(workerKind string) (extensions.WorkerPublication, bool) {
	worker, present := catalog.workers[workerKind]
	worker.JobKinds = append([]string(nil), worker.JobKinds...)
	return worker, present
}

func (catalog PublicationCatalog) Job(jobKind string) (extensions.JobKindContract, bool) {
	job, present := catalog.jobs[jobKind]
	return cloneJobKindContract(job), present
}

func (catalog PublicationCatalog) Participant(participantID string) (extensions.ParticipantContract, bool) {
	participant, present := catalog.participants[participantID]
	return cloneParticipantContract(participant), present
}

func (catalog PublicationCatalog) ImplementationBinding(profileID string) (extensions.ImplementationBindingPublication, bool) {
	binding, present := catalog.bindings[profileID]
	return binding, present
}

// ExactProfileContributionSet returns whether one profile is admitted and
// verifies that its installed contributions of the requested kind are exactly
// the application composition set. An admitted profile cannot be partially
// wired, and an unadmitted profile cannot contribute executable behavior.
func (catalog PublicationCatalog) ExactProfileContributionSet(profileID string, kind string, expectedIDs []string) (bool, error) {
	if profileID == "" || kind == "" || len(expectedIDs) == 0 {
		return false, fmt.Errorf("exact extension contribution projection is incomplete")
	}
	expected := append([]string(nil), expectedIDs...)
	sort.Strings(expected)
	for index, contributionID := range expected {
		if contributionID == "" {
			return false, fmt.Errorf("exact extension contribution projection contains an empty identity")
		}
		if index > 0 && contributionID == expected[index-1] {
			return false, fmt.Errorf("exact extension contribution projection contains duplicate %q", contributionID)
		}
	}

	actual := make([]string, 0, len(expected))
	for contributionID, contribution := range catalog.contributions {
		if contribution.ProfileID == profileID && contribution.Kind == kind {
			actual = append(actual, contributionID)
		}
	}
	sort.Strings(actual)
	_, admitted := catalog.bindings[profileID]
	if !admitted {
		if len(actual) != 0 {
			return false, fmt.Errorf("unadmitted extension profile %q has executable contributions %v", profileID, actual)
		}
		return false, nil
	}
	if len(actual) != len(expected) {
		return false, fmt.Errorf("extension profile %q %s contributions got %v want %v", profileID, kind, actual, expected)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			return false, fmt.Errorf("extension profile %q %s contributions got %v want %v", profileID, kind, actual, expected)
		}
	}
	return true, nil
}

func (catalog PublicationCatalog) ContributionIDs(kind string) []string {
	ids := []string{}
	for contributionID, contribution := range catalog.contributions {
		if kind == "" || contribution.Kind == kind {
			ids = append(ids, contributionID)
		}
	}
	sort.Strings(ids)
	return ids
}

func (catalog PublicationCatalog) WorkerKinds() []string {
	kinds := make([]string, 0, len(catalog.workers))
	for workerKind := range catalog.workers {
		kinds = append(kinds, workerKind)
	}
	sort.Strings(kinds)
	return kinds
}

func (catalog PublicationCatalog) JobKinds() []string {
	kinds := make([]string, 0, len(catalog.jobs))
	for jobKind := range catalog.jobs {
		kinds = append(kinds, jobKind)
	}
	sort.Strings(kinds)
	return kinds
}

func (catalog PublicationCatalog) ParticipantIDs() []string {
	ids := make([]string, 0, len(catalog.participants))
	for participantID := range catalog.participants {
		ids = append(ids, participantID)
	}
	sort.Strings(ids)
	return ids
}

func (catalog PublicationCatalog) BindingProfileIDs() []string {
	profileIDs := make([]string, 0, len(catalog.bindings))
	for profileID := range catalog.bindings {
		profileIDs = append(profileIDs, profileID)
	}
	sort.Strings(profileIDs)
	return profileIDs
}

func cloneJobKindContract(contract extensions.JobKindContract) extensions.JobKindContract {
	contract.ResourceRefContracts = append([]extensions.JobResourceRefContract(nil), contract.ResourceRefContracts...)
	return contract
}

func cloneParticipantContract(contract extensions.ParticipantContract) extensions.ParticipantContract {
	contract.AlgorithmIDs = append([]string(nil), contract.AlgorithmIDs...)
	contract.SerializationKeyKinds = append([]string(nil), contract.SerializationKeyKinds...)
	contract.OwnedStateFamilyIDs = append([]string(nil), contract.OwnedStateFamilyIDs...)
	contract.Operations = append([]extensions.ParticipantOperation(nil), contract.Operations...)
	for index := range contract.Operations {
		contract.Operations[index].StateFamilyIDs = append([]string(nil), contract.Operations[index].StateFamilyIDs...)
	}
	return contract
}
