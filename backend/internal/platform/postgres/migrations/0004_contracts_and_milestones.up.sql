-- Contracts (an accepted proposal) and milestones (stages of a fixed-price
-- contract). The funded/released transitions move money and are wired to the
-- ledger in a later migration; here they are structural.

CREATE TABLE contracts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    proposal_id         UUID NOT NULL UNIQUE REFERENCES proposals(id) ON DELETE RESTRICT,
    client_id           UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    freelancer_id       UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    type                TEXT NOT NULL CHECK (type IN ('fixed', 'hourly')),
    agreed_amount_cents BIGINT NOT NULL CHECK (agreed_amount_cents >= 0),
    currency            TEXT NOT NULL DEFAULT 'RUB',
    status              TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'paused', 'completed', 'cancelled', 'disputed')),
    started_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at            TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_contracts_client_id ON contracts(client_id);
CREATE INDEX idx_contracts_freelancer_id ON contracts(freelancer_id);

CREATE TABLE milestones (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id      UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    sequence         INT NOT NULL,
    title            TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    amount_cents     BIGINT NOT NULL CHECK (amount_cents > 0),
    status           TEXT NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'funded', 'submitted', 'approved', 'released', 'cancelled')),
    deliverable_note TEXT NOT NULL DEFAULT '',
    due_date         DATE,
    funded_at        TIMESTAMPTZ,
    submitted_at     TIMESTAMPTZ,
    approved_at      TIMESTAMPTZ,
    released_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (contract_id, sequence)
);
CREATE INDEX idx_milestones_contract_id ON milestones(contract_id);
