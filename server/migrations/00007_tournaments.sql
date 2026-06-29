-- +goose Up
-- Turnirlar (asinxron musobaqa): vaqt oynasida ishtirokchilar bir xil savollarni yechadi.
CREATE TABLE tournaments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title          VARCHAR NOT NULL,
    subject_id     UUID NOT NULL REFERENCES subjects(id),
    question_count INT NOT NULL DEFAULT 10,
    starts_at      TIMESTAMPTZ NOT NULL,
    ends_at        TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Ishtirokchi natijasi — eng yaxshi ball saqlanadi.
CREATE TABLE tournament_entries (
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id),
    score         INT NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tournament_id, user_id)
);

-- +goose Down
DROP TABLE tournament_entries;
DROP TABLE tournaments;
