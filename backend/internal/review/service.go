package review

import (
	"context"
	"errors"
	"strings"
	"time"

	"ipw/internal/contract"
	"ipw/internal/httpx"

	"github.com/google/uuid"
)

// ReviewWindow is how long after the first review the second party has before
// the pending review is published unilaterally.
const ReviewWindow = 14 * 24 * time.Hour

type Service struct {
	store     Store
	contracts contract.Store
	profiles  ProfilePort
}

func NewService(store Store, contracts contract.Store, profiles ProfilePort) *Service {
	return &Service{store: store, contracts: contracts, profiles: profiles}
}

type Input struct {
	Rating  int
	Comment string
}

func (s *Service) Submit(ctx context.Context, reviewerID, contractID uuid.UUID, in Input) (*Review, error) {
	if in.Rating < 1 || in.Rating > 5 {
		return nil, httpx.ErrBadRequest("rating must be between 1 and 5")
	}
	c, err := s.contracts.GetContract(ctx, contractID)
	if err != nil {
		if errors.Is(err, contract.ErrNotFound) {
			return nil, httpx.ErrNotFound("contract not found")
		}
		return nil, err
	}
	if c.Status != contract.StatusCompleted {
		return nil, httpx.ErrConflict("reviews open once the contract is completed")
	}

	var revieweeID uuid.UUID
	var dir Direction
	switch reviewerID {
	case c.ClientID:
		revieweeID, dir = c.FreelancerID, ClientToFreelancer
	case c.FreelancerID:
		revieweeID, dir = c.ClientID, FreelancerToClient
	default:
		return nil, httpx.ErrForbidden("not a party to this contract")
	}

	if _, err := s.store.ByContractAndReviewer(ctx, contractID, reviewerID); err == nil {
		return nil, httpx.ErrConflict("you have already reviewed this contract")
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	r := &Review{
		ID: uuid.New(), ContractID: contractID, ReviewerID: reviewerID, RevieweeID: revieweeID,
		Direction: dir, Rating: in.Rating, Comment: strings.TrimSpace(in.Comment),
	}
	if err := s.store.Create(ctx, r); err != nil {
		return nil, err
	}

	// If both parties have now reviewed, publish the pair.
	existing, err := s.store.ListByContract(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if len(existing) >= 2 {
		if err := s.publishContract(ctx, contractID); err != nil {
			return nil, err
		}
	}
	return s.store.ByContractAndReviewer(ctx, contractID, reviewerID)
}

// VisibleForContract returns the reviews on a contract that the viewer is
// allowed to see: published ones for everyone, plus the viewer's own.
func (s *Service) VisibleForContract(ctx context.Context, contractID, viewerID uuid.UUID) ([]Review, error) {
	all, err := s.store.ListByContract(ctx, contractID)
	if err != nil {
		return nil, err
	}
	out := make([]Review, 0, len(all))
	for _, r := range all {
		if r.PublishedAt != nil || r.ReviewerID == viewerID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *Service) PublishedForUser(ctx context.Context, userID uuid.UUID) ([]Review, error) {
	return s.store.ListPublishedForReviewee(ctx, userID)
}

// PublishExpired publishes single-sided reviews whose window has elapsed. Meant
// to be run periodically by a scheduler (TODO: wire to the job runner).
func (s *Service) PublishExpired(ctx context.Context) (int, error) {
	contractIDs, err := s.store.UnpublishedOlderThan(ctx, time.Now().Add(-ReviewWindow))
	if err != nil {
		return 0, err
	}
	for _, cid := range contractIDs {
		if err := s.publishContract(ctx, cid); err != nil {
			return 0, err
		}
	}
	return len(contractIDs), nil
}

func (s *Service) publishContract(ctx context.Context, contractID uuid.UUID) error {
	now := time.Now()
	if err := s.store.PublishForContract(ctx, contractID, now); err != nil {
		return err
	}
	reviews, err := s.store.ListByContract(ctx, contractID)
	if err != nil {
		return err
	}
	for _, r := range reviews {
		avg, count, err := s.store.RevieweeRatingSummary(ctx, r.RevieweeID)
		if err != nil {
			return err
		}
		switch r.Direction {
		case ClientToFreelancer:
			if err := s.profiles.SetFreelancerRating(ctx, r.RevieweeID, round2(avg), count); err != nil {
				return err
			}
		case FreelancerToClient:
			if err := s.profiles.SetClientRating(ctx, r.RevieweeID, round2(avg), count); err != nil {
				return err
			}
		}
	}
	return nil
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
