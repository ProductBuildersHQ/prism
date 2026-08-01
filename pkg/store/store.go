// Package store defines the storage interface for PRISM Control.
// The in-memory fake in this package enables unit testing of all domain
// logic without a running Dolt instance; the doltstore subpackage provides
// the Ent-backed production implementation.
package store

import (
	"context"
	"time"
)

// Store is the persistence interface shared by all service-layer operations.
// Every mutating method runs inside a UnitOfWork; reads may be called directly.
type Store interface {
	InitiativeStore
	ProgramStore
	PhaseStore
	RMIStore
	AssignmentStore
	EvidenceStore
	RepositoryStore
}

// UnitOfWork groups a SQL transaction with a subsequent Dolt commit.
// The production implementation wraps an Ent transaction and issues
// CALL DOLT_COMMIT on success; the in-memory fake is a no-op.
type UnitOfWork interface {
	// Execute runs fn inside a transaction. If fn returns nil the
	// transaction and Dolt commit are applied; otherwise both roll back.
	Execute(ctx context.Context, fn func(ctx context.Context, s Store) error) error
}

// Initiative represents a cross-repository initiative.
type Initiative struct {
	ID                 string
	Organization       string
	Title              string
	Description        string
	Status             string
	InitType           string // feature, maintenance, migration, compliance, refactor
	WorkflowID         string // SpecWorkflow override; empty means use the type's default
	Priority           string
	HomeRepo           string
	Workspace          string
	ProgramID          string
	Specs              map[string]string
	CreatedAt          time.Time
	PlannedAt          *time.Time
	ExecutingAt        *time.Time
	DeliveryCompleteAt *time.Time
	ReleasedAt         *time.Time
	ClosedAt           *time.Time
	UpdatedAt          time.Time
}

// Phase is a themed grouping of RMIs within an initiative.
// Phase status is always derived from member RMIs — never stored.
type Phase struct {
	ID             string
	InitiativeID   string
	SequenceNumber int
	Title          string
	Theme          string
}

// RoadmapItem (RMI) is a deliverable within a single repository.
type RoadmapItem struct {
	ID                 string
	RepositoryID       string
	InitiativeID       string
	PhaseID            string
	Title              string
	Description        string
	ItemType           string
	Status             string
	Priority           string
	Required           bool
	SequenceNumber     int
	AcceptanceCriteria []string
	ContextSpec        *ContextSpec
	CreatedAt          time.Time
	CompletedAt        *time.Time
	UpdatedAt          time.Time
}

// ContextSpec contains explicit overrides for context assembly.
// When present, these settings augment or filter the derived context.
type ContextSpec struct {
	// ExtraRepos lists additional repository IDs to include in the context.
	// These are added to the derived repo set with role "explicit".
	ExtraRepos []string `json:"extra_repos,omitempty"`

	// IncludeSpecs lists spec file paths to explicitly include.
	// Paths are relative to the RMI's repository root.
	IncludeSpecs []string `json:"include_specs,omitempty"`

	// ExcludeSpecs lists spec file paths to exclude from the context.
	// Takes precedence over initiative-level specs and IncludeSpecs.
	ExcludeSpecs []string `json:"exclude_specs,omitempty"`
}

// RMIDependency is a directed edge between two RMIs.
type RMIDependency struct {
	SourceRMIID  string
	TargetRMIID  string
	Relationship string // "requires" or "relates"
}

// InitiativeDependency is a directed edge between two initiatives.
type InitiativeDependency struct {
	SourceInitiativeID string
	TargetInitiativeID string
	Relationship       string // "requires" or "relates"
}

// Assignment is a lease-based work claim by an agent session.
type Assignment struct {
	ID             string
	RMIID          string
	Worker         string // session ID; matches omnidevx claudecode collector
	Status         string
	LeaseExpiresAt time.Time
	Workspace      string
	Handoff        *Handoff
	CreatedAt      time.Time
	CompletedAt    *time.Time
	UpdatedAt      time.Time
}

// Handoff carries compact state for session continuity.
type Handoff struct {
	Completed  []string `json:"completed"`
	Remaining  []string `json:"remaining"`
	Decisions  []string `json:"decisions"`
	NextAction string   `json:"next_action"`
}

// DeliveryEvidence links a commit, PR, release, or changelog entry to an RMI.
type DeliveryEvidence struct {
	ID           string
	RMIID        string
	EvidenceType string // commit|pr|release|changelog|test
	Reference    string
	CommitType   string // conventional commit type (commits only)
	CommitScope  string
	OccurredAt   *time.Time
	CreatedAt    time.Time
}

// Repository is a catalog entry for a participating repository.
type Repository struct {
	ID              string
	Organization    string
	RepositoryName  string
	DefaultBranch   string
	LocalPath       string // absolute path on disk (for ingest/scan)
	GoModule        string // go.mod module path (from gitscan)
	Domain          string
	Status          string
	IngestHighWater string // last scanned commit SHA
}

// Program is an organizational grouping of related initiatives.
type Program struct {
	ID           string
	Name         string
	Organization string
	Description  string
	Hidden       bool // when true, omitted from the dashboard homepage by default
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RepositoryDependency is a directed edge between two repositories
// derived from go.mod dependency analysis.
type RepositoryDependency struct {
	SourceRepositoryID string
	TargetRepositoryID string
	DependencyType     string // "go_module"
}

// ProgramStore defines persistence for programs.
type ProgramStore interface {
	CreateProgram(ctx context.Context, prog *Program) error
	GetProgram(ctx context.Context, id string) (*Program, error)
	ListPrograms(ctx context.Context) ([]*Program, error)
	UpdateProgram(ctx context.Context, prog *Program) error
}

// InitiativeStore defines persistence for initiatives.
type InitiativeStore interface {
	CreateInitiative(ctx context.Context, init *Initiative) error
	GetInitiative(ctx context.Context, id string) (*Initiative, error)
	ListInitiatives(ctx context.Context) ([]*Initiative, error)
	UpdateInitiative(ctx context.Context, init *Initiative) error
	CreateInitiativeDependency(ctx context.Context, dep *InitiativeDependency) error
	ListInitiativeDependencies(ctx context.Context, initiativeID string) ([]*InitiativeDependency, error)
	ListAllInitiativeDependencies(ctx context.Context) ([]*InitiativeDependency, error)
}

// PhaseStore defines persistence for phases.
type PhaseStore interface {
	CreatePhase(ctx context.Context, phase *Phase) error
	ListPhases(ctx context.Context, initiativeID string) ([]*Phase, error)
	DeletePhase(ctx context.Context, id string) error
}

// RMIStore defines persistence for roadmap items and dependencies.
type RMIStore interface {
	CreateRMI(ctx context.Context, rmi *RoadmapItem) error
	GetRMI(ctx context.Context, id string) (*RoadmapItem, error)
	ListRMIs(ctx context.Context, initiativeID string) ([]*RoadmapItem, error)
	ListAllRMIs(ctx context.Context) ([]*RoadmapItem, error)
	ListRMIsByRepo(ctx context.Context, repoID string) ([]*RoadmapItem, error)
	ListRMIsByStatus(ctx context.Context, status string) ([]*RoadmapItem, error)
	UpdateRMI(ctx context.Context, rmi *RoadmapItem) error
	CreateDependency(ctx context.Context, dep *RMIDependency) error
	ListDependencies(ctx context.Context, rmiID string) ([]*RMIDependency, error)
	ListAllDependencies(ctx context.Context) ([]*RMIDependency, error)
}

// AssignmentStore defines persistence for lease-based work claims.
type AssignmentStore interface {
	CreateAssignment(ctx context.Context, a *Assignment) error
	GetAssignment(ctx context.Context, id string) (*Assignment, error)
	GetActiveAssignment(ctx context.Context, rmiID string) (*Assignment, error)
	ListActiveAssignments(ctx context.Context) ([]*Assignment, error)
	ListAllAssignments(ctx context.Context) ([]*Assignment, error)
	UpdateAssignment(ctx context.Context, a *Assignment) error
}

// EvidenceStore defines persistence for delivery evidence.
type EvidenceStore interface {
	CreateEvidence(ctx context.Context, ev *DeliveryEvidence) error
	ListEvidenceByRMI(ctx context.Context, rmiID string) ([]*DeliveryEvidence, error)
	ListEvidenceByInitiative(ctx context.Context, initiativeID string) ([]*DeliveryEvidence, error)
	ListAllEvidence(ctx context.Context) ([]*DeliveryEvidence, error)
}

// RepositoryStore defines persistence for the repository catalog.
type RepositoryStore interface {
	CreateRepository(ctx context.Context, repo *Repository) error
	GetRepository(ctx context.Context, id string) (*Repository, error)
	ListRepositories(ctx context.Context) ([]*Repository, error)
	ListRepositoriesByOrg(ctx context.Context, org string) ([]*Repository, error)
	UpdateRepository(ctx context.Context, repo *Repository) error
	CreateRepoDependency(ctx context.Context, dep *RepositoryDependency) error
	ListRepoDependencies(ctx context.Context, repoID string) ([]*RepositoryDependency, error)
	ListAllRepoDependencies(ctx context.Context) ([]*RepositoryDependency, error)
}
