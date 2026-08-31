package profile

import (
	"context"
	"strings"

	"ipw/internal/catalog"
	"ipw/internal/httpx"

	"github.com/google/uuid"
)

type Service struct {
	store   Store
	catalog catalog.Store
}

func NewService(store Store, cat catalog.Store) *Service {
	return &Service{store: store, catalog: cat}
}

type FreelancerInput struct {
	Headline          string
	Bio               string
	HourlyRateCents   int64
	Currency          string
	Availability      string
	PrimaryCategoryID *uuid.UUID
	Languages         []string
	Location          string
}

func (s *Service) GetFreelancer(ctx context.Context, userID uuid.UUID) (*Freelancer, error) {
	return s.store.GetFreelancer(ctx, userID)
}

func (s *Service) ListFreelancers(ctx context.Context, f FreelancerFilter) ([]Freelancer, error) {
	return s.store.ListFreelancers(ctx, f)
}

func (s *Service) SaveFreelancer(ctx context.Context, userID uuid.UUID, in FreelancerInput) (*Freelancer, error) {
	availability := Availability(strings.TrimSpace(in.Availability))
	if availability == "" {
		availability = AvailabilityUnknown
	}
	if !availability.valid() {
		return nil, httpx.ErrBadRequest("invalid availability")
	}
	if in.HourlyRateCents < 0 {
		return nil, httpx.ErrBadRequest("hourly rate cannot be negative")
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "RUB"
	}
	if in.PrimaryCategoryID != nil {
		ok, err := s.catalog.CategoryExists(ctx, *in.PrimaryCategoryID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpx.ErrBadRequest("unknown category")
		}
	}

	f := &Freelancer{
		UserID:            userID,
		Headline:          strings.TrimSpace(in.Headline),
		Bio:               strings.TrimSpace(in.Bio),
		HourlyRateCents:   in.HourlyRateCents,
		Currency:          currency,
		Availability:      availability,
		PrimaryCategoryID: in.PrimaryCategoryID,
		Languages:         normalizeStrings(in.Languages),
		Location:          strings.TrimSpace(in.Location),
	}
	if err := s.store.UpsertFreelancer(ctx, f); err != nil {
		return nil, err
	}
	return s.store.GetFreelancer(ctx, userID)
}

func (s *Service) SetFreelancerSkills(ctx context.Context, userID uuid.UUID, skillIDs []uuid.UUID) (*Freelancer, error) {
	skillIDs = dedupeUUIDs(skillIDs)
	if len(skillIDs) > 30 {
		return nil, httpx.ErrBadRequest("at most 30 skills")
	}
	ok, err := s.catalog.SkillIDsExist(ctx, skillIDs)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, httpx.ErrBadRequest("one or more skills are unknown")
	}
	if _, err := s.store.GetFreelancer(ctx, userID); err != nil {
		return nil, err // profile must exist first
	}
	if err := s.store.SetFreelancerSkills(ctx, userID, skillIDs); err != nil {
		return nil, err
	}
	return s.store.GetFreelancer(ctx, userID)
}

type ClientInput struct {
	CompanyName string
	About       string
	Website     string
	Location    string
}

func (s *Service) GetClient(ctx context.Context, userID uuid.UUID) (*Client, error) {
	return s.store.GetClient(ctx, userID)
}

func (s *Service) SaveClient(ctx context.Context, userID uuid.UUID, in ClientInput) (*Client, error) {
	c := &Client{
		UserID:      userID,
		CompanyName: strings.TrimSpace(in.CompanyName),
		About:       strings.TrimSpace(in.About),
		Website:     strings.TrimSpace(in.Website),
		Location:    strings.TrimSpace(in.Location),
	}
	if err := s.store.UpsertClient(ctx, c); err != nil {
		return nil, err
	}
	return s.store.GetClient(ctx, userID)
}

type PortfolioInput struct {
	Title       string
	Description string
	URL         string
	ImageKey    string
}

func (s *Service) AddPortfolioItem(ctx context.Context, userID uuid.UUID, in PortfolioInput) (*PortfolioItem, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, httpx.ErrBadRequest("title is required")
	}
	existing, err := s.store.ListPortfolio(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(existing) >= 20 {
		return nil, httpx.ErrBadRequest("at most 20 portfolio items")
	}
	item := &PortfolioItem{
		UserID:      userID,
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		URL:         strings.TrimSpace(in.URL),
		ImageKey:    in.ImageKey,
		Position:    len(existing),
	}
	if err := s.store.AddPortfolioItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) ListPortfolio(ctx context.Context, userID uuid.UUID) ([]PortfolioItem, error) {
	return s.store.ListPortfolio(ctx, userID)
}

func (s *Service) DeletePortfolioItem(ctx context.Context, userID, itemID uuid.UUID) error {
	return s.store.DeletePortfolioItem(ctx, userID, itemID)
}

func normalizeStrings(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for _, s := range in {
		t := strings.TrimSpace(s)
		if t == "" || seen[strings.ToLower(t)] {
			continue
		}
		seen[strings.ToLower(t)] = true
		out = append(out, t)
	}
	return out
}

func dedupeUUIDs(in []uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(in))
	seen := map[uuid.UUID]bool{}
	for _, id := range in {
		if id == uuid.Nil || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}
