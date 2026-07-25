package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ProductBuildersHQ/prism-control/pkg/initiative"
	"github.com/ProductBuildersHQ/prism-control/pkg/store"
)

// CreateInitiative creates a new initiative in "proposed" status.
func (s *Service) CreateInitiative(ctx context.Context, id, org, title, description, priority string) (*store.Initiative, error) {
	now := time.Now()
	init := &store.Initiative{
		ID:           id,
		Organization: org,
		Title:        title,
		Description:  description,
		Status:       initiative.StatusProposed,
		Priority:     priority,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.Store.CreateInitiative(ctx, init); err != nil {
		return nil, err
	}
	return init, nil
}

// GetInitiative returns an initiative by ID.
func (s *Service) GetInitiative(ctx context.Context, id string) (*store.Initiative, error) {
	return s.Store.GetInitiative(ctx, id)
}

// UpdateInitiative persists changes to an initiative.
func (s *Service) UpdateInitiative(ctx context.Context, init *store.Initiative) error {
	return s.Store.UpdateInitiative(ctx, init)
}

// ListInitiatives returns all initiatives.
func (s *Service) ListInitiatives(ctx context.Context) ([]*store.Initiative, error) {
	return s.Store.ListInitiatives(ctx)
}

// TransitionInitiative changes an initiative's lifecycle status,
// validating the transition and stamping the appropriate timestamp.
func (s *Service) TransitionInitiative(ctx context.Context, id, toStatus string) (*store.Initiative, error) {
	init, err := s.Store.GetInitiative(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := initiative.Transition(init, toStatus, now); err != nil {
		return nil, err
	}

	if err := s.Store.UpdateInitiative(ctx, init); err != nil {
		return nil, fmt.Errorf("save transition: %w", err)
	}
	return init, nil
}

// InitiativeDetail is an initiative with its phases and derived status.
type InitiativeDetail struct {
	Initiative *store.Initiative
	Phases     []PhaseDetail
}

// PhaseDetail is a phase with its derived status from member RMIs.
type PhaseDetail struct {
	Phase  *store.Phase
	Status string
	RMIs   []*store.RoadmapItem
}

// GetInitiativeDetail returns an initiative with phases and derived status.
func (s *Service) GetInitiativeDetail(ctx context.Context, id string) (*InitiativeDetail, error) {
	init, err := s.Store.GetInitiative(ctx, id)
	if err != nil {
		return nil, err
	}

	phases, err := s.Store.ListPhases(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}

	allRMIs, err := s.Store.ListRMIs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list RMIs: %w", err)
	}

	rmisByPhase := make(map[string][]*store.RoadmapItem)
	for _, r := range allRMIs {
		rmisByPhase[r.PhaseID] = append(rmisByPhase[r.PhaseID], r)
	}

	detail := &InitiativeDetail{Initiative: init}
	for _, p := range phases {
		phaseRMIs := rmisByPhase[p.ID]
		detail.Phases = append(detail.Phases, PhaseDetail{
			Phase:  p,
			Status: initiative.DerivePhaseStatus(phaseRMIs),
			RMIs:   phaseRMIs,
		})
	}
	return detail, nil
}

// CreatePhase adds a phase to an initiative.
func (s *Service) CreatePhase(ctx context.Context, id, initiativeID string, seq int, title, theme string) (*store.Phase, error) {
	if _, err := s.Store.GetInitiative(ctx, initiativeID); err != nil {
		return nil, fmt.Errorf("initiative %s: %w", initiativeID, err)
	}

	p := &store.Phase{
		ID:             id,
		InitiativeID:   initiativeID,
		SequenceNumber: seq,
		Title:          title,
		Theme:          theme,
	}
	if err := s.Store.CreatePhase(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ListPhases returns phases for an initiative.
func (s *Service) ListPhases(ctx context.Context, initiativeID string) ([]*store.Phase, error) {
	return s.Store.ListPhases(ctx, initiativeID)
}
