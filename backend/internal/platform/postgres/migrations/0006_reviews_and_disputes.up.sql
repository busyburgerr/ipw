-- Two-sided reviews (double-blind publication) and contract disputes.

CREATE TABLE reviews (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id  UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    reviewer_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reviewee_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    direction    TEXT NOT NULL
                 CHECK (direction IN ('client_to_freelancer', 'freelancer_to_client')),
    rating       INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment      TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (contract_id, reviewer_id)
);
CREATE INDEX idx_reviews_reviewee_published ON reviews(reviewee_id, published_at);

CREATE TABLE disputes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id     UUID NOT NULL REFERENCES contracts(id) ON DELETE RESTRICT,
    milestone_id    UUID REFERENCES milestones(id) ON DELETE RESTRICT,
    raised_by       UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reason          TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'open'
                    CHECK (status IN ('open', 'under_review', 'resolved_client',
                                      'resolved_freelancer', 'resolved_split', 'withdrawn')),
    resolution_note TEXT NOT NULL DEFAULT '',
    arbiter_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ
);
CREATE INDEX idx_disputes_status ON disputes(status);
CREATE INDEX idx_disputes_contract_id ON disputes(contract_id);
-- One live dispute per contract.
CREATE UNIQUE INDEX idx_disputes_one_live_per_contract
    ON disputes(contract_id) WHERE status IN ('open', 'under_review');
