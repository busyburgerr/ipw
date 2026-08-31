// Package billing orchestrates the money-touching steps of a contract: funding a
// milestone through the payment provider, releasing escrow on approval, refunds,
// and freelancer payouts. It is the only package that combines the contract,
// payment, wallet and ledger features.
package billing

import (
	"context"
	"errors"
	"strings"
	"time"

	"ipw/internal/config"
	"ipw/internal/contract"
	"ipw/internal/httpx"
	"ipw/internal/payment"
	"ipw/internal/user"
	"ipw/internal/wallet"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	contracts contract.Store
	payments  *payment.Service
	wallet    *wallet.Service
	users     user.Store
	payouts   *payoutStore
	cfg       config.BillingConfig
}

func NewService(db *gorm.DB, contracts contract.Store, payments *payment.Service, w *wallet.Service, users user.Store, cfg config.BillingConfig) *Service {
	return &Service{
		contracts: contracts,
		payments:  payments,
		wallet:    w,
		users:     users,
		payouts:   newPayoutStore(db),
		cfg:       cfg,
	}
}

// ---- milestone funding & release ------------------------------------

// FundMilestone creates a payment invoice for a pending milestone. The milestone
// only moves to "funded" once the provider confirms payment (webhook).
func (s *Service) FundMilestone(ctx context.Context, clientID, milestoneID uuid.UUID) (*payment.Payment, error) {
	m, c, err := s.milestoneForClient(ctx, clientID, milestoneID)
	if err != nil {
		return nil, err
	}
	if c.Type != contract.TypeFixed {
		return nil, httpx.ErrConflict("only fixed-price milestones are funded")
	}
	if c.Status != contract.StatusActive {
		return nil, httpx.ErrConflict("contract is not active")
	}
	if m.Status != contract.MilestonePending {
		return nil, httpx.ErrConflict("milestone is not pending")
	}

	client, err := s.users.GetByID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	return s.payments.CreateInvoice(ctx, payment.NewInvoiceParams{
		MilestoneID: m.ID,
		ContractID:  c.ID,
		PayerID:     clientID,
		PayerEmail:  client.Email,
		AmountCents: m.AmountCents,
		Currency:    c.Currency,
		Description: "Milestone: " + m.Title,
	})
}

// ConfirmPaidPayment is called after payment.Service.HandleWebhook reports a
// paid payment. It funds escrow and moves the milestone to "funded". Idempotent.
func (s *Service) ConfirmPaidPayment(ctx context.Context, p *payment.Payment) error {
	m, err := s.contracts.GetMilestone(ctx, p.MilestoneID)
	if err != nil {
		return err
	}
	if m.Status != contract.MilestonePending && m.Status != contract.MilestoneFunded {
		return nil // milestone moved on; nothing to do
	}
	if err := s.wallet.FundEscrow(ctx, p.ContractID, p.MilestoneID, p.AmountCents, p.ID.String()); err != nil {
		return err
	}
	if m.Status == contract.MilestonePending {
		return s.contracts.UpdateMilestoneStatus(ctx, m.ID, contract.MilestoneFunded,
			map[string]time.Time{"funded_at": time.Now()}, "")
	}
	return nil
}

// ApproveMilestone accepts submitted work: it moves the milestone to approved,
// releases escrow (freelancer balance + platform commission), and marks it
// released.
func (s *Service) ApproveMilestone(ctx context.Context, clientID, milestoneID uuid.UUID) (*contract.Milestone, error) {
	m, c, err := s.milestoneForClient(ctx, clientID, milestoneID)
	if err != nil {
		return nil, err
	}
	if m.Status != contract.MilestoneSubmitted {
		return nil, httpx.ErrConflict("milestone is not awaiting review")
	}

	if err := s.contracts.UpdateMilestoneStatus(ctx, m.ID, contract.MilestoneApproved,
		map[string]time.Time{"approved_at": time.Now()}, ""); err != nil {
		return nil, err
	}
	if _, _, err := s.wallet.ReleaseEscrow(ctx, c.ID, m.ID, c.FreelancerID, m.AmountCents); err != nil {
		return nil, err
	}
	if err := s.contracts.UpdateMilestoneStatus(ctx, m.ID, contract.MilestoneReleased,
		map[string]time.Time{"released_at": time.Now()}, ""); err != nil {
		return nil, err
	}
	return s.contracts.GetMilestone(ctx, m.ID)
}

// RefundMilestone returns a funded (but not yet submitted) milestone's escrow.
// The actual refund to the client happens out of band via the payment provider.
func (s *Service) RefundMilestone(ctx context.Context, clientID, milestoneID uuid.UUID) (*contract.Milestone, error) {
	m, c, err := s.milestoneForClient(ctx, clientID, milestoneID)
	if err != nil {
		return nil, err
	}
	if m.Status != contract.MilestoneFunded {
		return nil, httpx.ErrConflict("only a funded milestone that has no submitted work can be refunded")
	}
	if err := s.wallet.RefundEscrow(ctx, c.ID, m.ID, m.AmountCents); err != nil {
		return nil, err
	}
	if err := s.contracts.UpdateMilestoneStatus(ctx, m.ID, contract.MilestoneCancelled, nil, ""); err != nil {
		return nil, err
	}
	return s.contracts.GetMilestone(ctx, m.ID)
}

// ---- payouts --------------------------------------------------------

type PayoutRequest struct {
	AmountCents int64
	Method      string
	Destination string
}

func (s *Service) RequestPayout(ctx context.Context, freelancerID uuid.UUID, in PayoutRequest) (*Payout, error) {
	if in.AmountCents < s.cfg.PayoutMinCents {
		return nil, httpx.ErrBadRequest("amount is below the minimum payout")
	}
	if !validMethod(in.Method) {
		return nil, httpx.ErrBadRequest("method must be one of: sbp, card, manual")
	}
	available, err := s.wallet.AvailableBalance(ctx, freelancerID)
	if err != nil {
		return nil, err
	}
	if in.AmountCents > available {
		return nil, httpx.ErrBadRequest("amount exceeds available balance")
	}

	p := &Payout{
		ID:           uuid.New(),
		FreelancerID: freelancerID,
		AmountCents:  in.AmountCents,
		Currency:     "RUB",
		Method:       in.Method,
		Destination:  maskDestination(in.Destination),
		Status:       PayoutRequested,
	}
	if err := s.payouts.create(ctx, p); err != nil {
		return nil, err
	}
	if err := s.wallet.HoldForPayout(ctx, freelancerID, p.ID, in.AmountCents); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListMyPayouts(ctx context.Context, freelancerID uuid.UUID) ([]Payout, error) {
	return s.payouts.listForFreelancer(ctx, freelancerID)
}

// ProcessPayout is an admin action: "paid" settles the hold, "rejected" returns
// the money to the freelancer's balance.
func (s *Service) ProcessPayout(ctx context.Context, payoutID uuid.UUID, decision PayoutStatus, note string) (*Payout, error) {
	p, err := s.payouts.get(ctx, payoutID)
	if err != nil {
		return nil, httpx.ErrNotFound("payout not found")
	}
	if p.Status != PayoutRequested && p.Status != PayoutProcessing {
		return nil, httpx.ErrConflict("payout is already resolved")
	}
	switch decision {
	case PayoutPaid:
		if err := s.wallet.SettlePayout(ctx, p.FreelancerID, p.ID, p.AmountCents); err != nil {
			return nil, err
		}
	case PayoutRejected:
		if err := s.wallet.ReversePayout(ctx, p.FreelancerID, p.ID, p.AmountCents); err != nil {
			return nil, err
		}
	default:
		return nil, httpx.ErrBadRequest("decision must be paid or rejected")
	}
	if err := s.payouts.setStatus(ctx, p.ID, decision, note, true); err != nil {
		return nil, err
	}
	return s.payouts.get(ctx, p.ID)
}

// ---- wallet summary ------------------------------------------------

type WalletSummary struct {
	AvailableCents     int64
	PendingPayoutCents int64
	Currency           string
}

func (s *Service) WalletSummary(ctx context.Context, userID uuid.UUID) (WalletSummary, error) {
	avail, err := s.wallet.AvailableBalance(ctx, userID)
	if err != nil {
		return WalletSummary{}, err
	}
	pending, err := s.wallet.PendingPayout(ctx, userID)
	if err != nil {
		return WalletSummary{}, err
	}
	return WalletSummary{AvailableCents: avail, PendingPayoutCents: pending, Currency: "RUB"}, nil
}

// ---- helpers ------------------------------------------------------

func (s *Service) milestoneForClient(ctx context.Context, clientID, milestoneID uuid.UUID) (*contract.Milestone, *contract.Contract, error) {
	m, err := s.contracts.GetMilestone(ctx, milestoneID)
	if err != nil {
		if errors.Is(err, contract.ErrNotFound) {
			return nil, nil, httpx.ErrNotFound("milestone not found")
		}
		return nil, nil, err
	}
	c, err := s.contracts.GetContract(ctx, m.ContractID)
	if err != nil {
		return nil, nil, err
	}
	if c.ClientID != clientID {
		return nil, nil, httpx.ErrForbidden("not your contract")
	}
	return m, c, nil
}

func validMethod(m string) bool {
	switch m {
	case "sbp", "card", "manual":
		return true
	}
	return false
}

func maskDestination(d string) string {
	d = strings.TrimSpace(d)
	if len(d) <= 4 {
		return d
	}
	return "****" + d[len(d)-4:]
}
