-- Service catalog (categories + skills) and the two profile types.
-- Monetary amounts are integer minor units (kopecks); never floats.

CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id  UUID REFERENCES categories(id) ON DELETE SET NULL,
    slug       TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL,
    position   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_categories_parent_id ON categories(parent_id);

CREATE TABLE skills (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    slug        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_skills_category_id ON skills(category_id);

CREATE TABLE freelancer_profiles (
    user_id             UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    headline            TEXT NOT NULL DEFAULT '',
    bio                 TEXT NOT NULL DEFAULT '',
    hourly_rate_cents   BIGINT NOT NULL DEFAULT 0 CHECK (hourly_rate_cents >= 0),
    currency            TEXT NOT NULL DEFAULT 'RUB',
    availability        TEXT NOT NULL DEFAULT 'unknown'
                        CHECK (availability IN ('available', 'limited', 'unavailable', 'unknown')),
    primary_category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    languages           JSONB NOT NULL DEFAULT '[]',
    location            TEXT NOT NULL DEFAULT '',
    rating_avg          NUMERIC(3,2) NOT NULL DEFAULT 0,
    rating_count        INT NOT NULL DEFAULT 0,
    jobs_completed      INT NOT NULL DEFAULT 0,
    total_earned_cents  BIGINT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE freelancer_skills (
    user_id  UUID NOT NULL REFERENCES freelancer_profiles(user_id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, skill_id)
);
CREATE INDEX idx_freelancer_skills_skill_id ON freelancer_skills(skill_id);

CREATE TABLE portfolio_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES freelancer_profiles(user_id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    url         TEXT NOT NULL DEFAULT '',
    image_key   TEXT NOT NULL DEFAULT '',
    position    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_portfolio_items_user_id ON portfolio_items(user_id);

CREATE TABLE client_profiles (
    user_id          UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    company_name     TEXT NOT NULL DEFAULT '',
    about            TEXT NOT NULL DEFAULT '',
    website          TEXT NOT NULL DEFAULT '',
    location         TEXT NOT NULL DEFAULT '',
    payment_verified BOOLEAN NOT NULL DEFAULT FALSE,
    rating_avg       NUMERIC(3,2) NOT NULL DEFAULT 0,
    rating_count     INT NOT NULL DEFAULT 0,
    hires_count      INT NOT NULL DEFAULT 0,
    total_spent_cents BIGINT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
