-- Double-entry ledger, external payment tracking, and freelancer payouts.
--
-- Ledger invariant: every ledger_transaction has >= 2 entries whose signed
-- amount_cents sum to exactly 0. An account's balance is SUM(amount_cents) of
-- its entries. Positive amount = credit (balance up), negative = debit.

CREATE TABLE ledger_accounts (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind      TEXT NOT NULL
              CHECK (kind IN ('escrow', 'user_balance', 'payout_clearing', 'platform_revenue', 'external')),
    owner_id  UUID,            -- user id / milestone id; NULL for singletons
    currency  TEXT NOT NULL DEFAULT 'RUB',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, owner_id, currency)
);
-- Enforce uniqueness for singleton accounts (owner_id IS NULL).
CREATE UNIQUE INDEX idx_ledger_accounts_singleton
    ON ledger_accounts (kind, currency) WHERE owner_id IS NULL;

-- Pre-create the singleton accounts so the engine never races to insert them.
INSERT INTO ledger_accounts (kind, currency) VALUES
    ('external', 'RUB'),
    ('platform_revenue', 'RUB');

CREATE TABLE ledger_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind            TEXT NOT NULL
                    CHECK (kind IN ('escrow_fund', 'escrow_release', 'refund', 'payout', 'payout_reversal')),
    idempotency_key TEXT NOT NULL UNIQUE,
    reference       TEXT NOT NULL DEFAULT '',
    contract_id     UUID,
    milestone_id    UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ledger_entries (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES ledger_transactions(id) ON DELETE RESTRICT,
    account_id     UUID NOT NULL REFERENCES ledger_accounts(id) ON DELETE RESTRICT,
    amount_cents   BIGINT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_ledger_entries_account_id ON ledger_entries(account_id);
CREATE INDEX idx_ledger_entries_transaction_id ON ledger_entries(transaction_id);

CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    milestone_id        UUID NOT NULL REFERENCES milestones(id) ON DELETE RESTRICT,
    contract_id         UUID NOT NULL REFERENCES contracts(id) ON DELETE RESTRICT,
    payer_id            UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount_cents        BIGINT NOT NULL CHECK (amount_cents > 0),
    currency            TEXT NOT NULL DEFAULT 'RUB',
    provider            TEXT NOT NULL,            -- 'lava' | 'stub'
    provider_invoice_id TEXT NOT NULL DEFAULT '',
    payment_url         TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'paid', 'failed', 'expired')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at             TIMESTAMPTZ
);
CREATE INDEX idx_payments_milestone_id ON payments(milestone_id);
CREATE UNIQUE INDEX idx_payments_provider_invoice
    ON payments(provider, provider_invoice_id) WHERE provider_invoice_id <> '';
-- At most one live (pending/paid) payment per milestone.
CREATE UNIQUE INDEX idx_payments_one_live_per_milestone
    ON payments(milestone_id) WHERE status IN ('pending', 'paid');

CREATE TABLE payouts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    freelancer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount_cents  BIGINT NOT NULL CHECK (amount_cents > 0),
    currency      TEXT NOT NULL DEFAULT 'RUB',
    method        TEXT NOT NULL CHECK (method IN ('sbp', 'card', 'manual')),
    destination   TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'requested'
                  CHECK (status IN ('requested', 'processing', 'paid', 'rejected')),
    note          TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at  TIMESTAMPTZ
);
CREATE INDEX idx_payouts_freelancer_id ON payouts(freelancer_id);
