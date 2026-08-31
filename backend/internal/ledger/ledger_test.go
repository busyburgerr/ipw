package ledger

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestPostRejectsUnbalancedPostings(t *testing.T) {
	l := New(nil) // DB is never touched: validation happens first
	mid := uuid.New()

	_, err := l.Post(context.Background(), TxnInput{
		Kind:           TxnEscrowFund,
		IdempotencyKey: "x",
		Postings: []Posting{
			{AccountKind: AccountExternal, AmountCents: -100},
			{AccountKind: AccountEscrow, OwnerID: &mid, AmountCents: 90},
		},
	})
	if !errors.Is(err, ErrUnbalanced) {
		t.Fatalf("want ErrUnbalanced, got %v", err)
	}
}

func TestPostRejectsSingleLeg(t *testing.T) {
	l := New(nil)
	_, err := l.Post(context.Background(), TxnInput{
		Kind:           TxnRefund,
		IdempotencyKey: "y",
		Postings:       []Posting{{AccountKind: AccountExternal, AmountCents: 0}},
	})
	if !errors.Is(err, ErrTooFewLegs) {
		t.Fatalf("want ErrTooFewLegs, got %v", err)
	}
}
