-- +goose Up
-- O'yin sessiyasi (xona) — tugagach tarix uchun saqlanadi
CREATE TABLE game_sessions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code           VARCHAR NOT NULL UNIQUE,           -- join kodi
    host_user_id   UUID NOT NULL REFERENCES users(id),
    subject_id     UUID NOT NULL REFERENCES subjects(id),
    category_id    UUID REFERENCES categories(id),    -- NULL = aralash
    mode           VARCHAR NOT NULL,                  -- classic|survival|...
    opponent       VARCHAR NOT NULL DEFAULT 'human',  -- human|bot|mixed
    question_count INT NOT NULL,
    time_per_q     INT NOT NULL,                      -- soniya
    status         VARCHAR NOT NULL DEFAULT 'lobby',  -- lobby|running|finished
    started_at     TIMESTAMPTZ,
    finished_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Natijalar (o'yinchi-sessiya)
CREATE TABLE game_results (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    score       DOUBLE PRECISION NOT NULL DEFAULT 0,
    correct_cnt INT NOT NULL DEFAULT 0,
    rank        INT,
    UNIQUE (session_id, user_id)
);

-- Javoblar logi (analitika / anti-cheat audit)
CREATE TABLE answers_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    question_id UUID NOT NULL REFERENCES questions(id),
    given       JSONB,
    is_correct  BOOLEAN NOT NULL,
    time_ms     INT NOT NULL,            -- server o'lchagan reaksiya vaqti
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_results_session ON game_results (session_id);
CREATE INDEX idx_answers_session ON answers_log (session_id);

-- +goose Down
DROP TABLE answers_log;
DROP TABLE game_results;
DROP TABLE game_sessions;
