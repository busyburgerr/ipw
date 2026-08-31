package project

import (
	"context"
	"testing"

	"ipw/internal/catalog"

	"github.com/google/uuid"
)

// stubCatalog accepts everything.
type stubCatalog struct{}

func (stubCatalog) ListCategories(context.Context) ([]catalog.Category, error)  { return nil, nil }
func (stubCatalog) ListSkills(context.Context, string) ([]catalog.Skill, error) { return nil, nil }
func (stubCatalog) SkillsByIDs(context.Context, []uuid.UUID) ([]catalog.Skill, error) {
	return nil, nil
}
func (stubCatalog) CategoryExists(context.Context, uuid.UUID) (bool, error)  { return true, nil }
func (stubCatalog) SkillIDsExist(context.Context, []uuid.UUID) (bool, error) { return true, nil }

func newTestService() *Service { return NewService(nil, stubCatalog{}) }

func ptrI64(v int64) *int64 { return &v }

func TestValidateProjectInputBudget(t *testing.T) {
	s := newTestService()
	ctx := context.Background()

	cases := []struct {
		name    string
		in      ProjectInput
		wantErr bool
	}{
		{"fixed ok", ProjectInput{Title: "Build an API", BudgetType: "fixed", FixedAmountCents: ptrI64(100000)}, false},
		{"fixed missing amount", ProjectInput{Title: "Build an API", BudgetType: "fixed"}, true},
		{"fixed zero amount", ProjectInput{Title: "Build an API", BudgetType: "fixed", FixedAmountCents: ptrI64(0)}, true},
		{"hourly ok", ProjectInput{Title: "Build an API", BudgetType: "hourly", HourlyRateMinCents: ptrI64(1000), HourlyRateMaxCents: ptrI64(2000)}, false},
		{"hourly min>max", ProjectInput{Title: "Build an API", BudgetType: "hourly", HourlyRateMinCents: ptrI64(2000), HourlyRateMaxCents: ptrI64(1000)}, true},
		{"unknown budget", ProjectInput{Title: "Build an API", BudgetType: "milestone"}, true},
		{"short title", ProjectInput{Title: "ab", BudgetType: "fixed", FixedAmountCents: ptrI64(1)}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.validateProjectInput(ctx, tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateProjectInputDefaults(t *testing.T) {
	p, err := newTestService().validateProjectInput(context.Background(), ProjectInput{
		Title: "Design a landing page", BudgetType: "fixed", FixedAmountCents: ptrI64(50000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Currency != "RUB" {
		t.Errorf("currency default = %q, want RUB", p.Currency)
	}
	if p.ExperienceLevel != ExperienceAny {
		t.Errorf("experience default = %q, want any", p.ExperienceLevel)
	}
}
