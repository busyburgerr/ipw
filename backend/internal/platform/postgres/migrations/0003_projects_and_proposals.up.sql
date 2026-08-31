-- Projects (a client's job posting) and proposals (a freelancer's bid).

CREATE TABLE projects (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title                 TEXT NOT NULL,
    description           TEXT NOT NULL DEFAULT '',
    category_id           UUID REFERENCES categories(id) ON DELETE SET NULL,
    budget_type           TEXT NOT NULL CHECK (budget_type IN ('fixed', 'hourly')),
    fixed_amount_cents    BIGINT CHECK (fixed_amount_cents IS NULL OR fixed_amount_cents >= 0),
    hourly_rate_min_cents BIGINT CHECK (hourly_rate_min_cents IS NULL OR hourly_rate_min_cents >= 0),
    hourly_rate_max_cents BIGINT CHECK (hourly_rate_max_cents IS NULL OR hourly_rate_max_cents >= 0),
    currency              TEXT NOT NULL DEFAULT 'RUB',
    experience_level      TEXT NOT NULL DEFAULT 'any'
                          CHECK (experience_level IN ('any', 'entry', 'intermediate', 'expert')),
    status                TEXT NOT NULL DEFAULT 'draft'
                          CHECK (status IN ('draft', 'open', 'in_progress', 'completed', 'cancelled')),
    proposals_count       INT NOT NULL DEFAULT 0,
    published_at          TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_projects_client_id ON projects(client_id);
CREATE INDEX idx_projects_status_published ON projects(status, published_at DESC);
CREATE INDEX idx_projects_category_id ON projects(category_id);

CREATE TABLE project_skills (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    skill_id   UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, skill_id)
);
CREATE INDEX idx_project_skills_skill_id ON project_skills(skill_id);

CREATE TABLE proposals (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id     UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    freelancer_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cover_letter   TEXT NOT NULL DEFAULT '',
    bid_amount_cents BIGINT NOT NULL CHECK (bid_amount_cents >= 0),
    estimated_days INT CHECK (estimated_days IS NULL OR estimated_days > 0),
    status         TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'shortlisted', 'accepted', 'declined', 'withdrawn')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, freelancer_id)
);
CREATE INDEX idx_proposals_freelancer_id ON proposals(freelancer_id);
CREATE INDEX idx_proposals_project_status ON proposals(project_id, status);
