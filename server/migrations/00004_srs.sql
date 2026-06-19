-- +goose Up
-- 📚 O'rganish: Spaced Repetition (har user-savol uchun SM-2 holati)
CREATE TABLE srs_cards (
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id   UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    ease          DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    interval_days INT NOT NULL DEFAULT 0,
    repetitions   INT NOT NULL DEFAULT 0,
    due_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_reviewed TIMESTAMPTZ,
    PRIMARY KEY (user_id, question_id)
);

CREATE INDEX idx_srs_due ON srs_cards (user_id, due_at);

-- +goose Down
DROP TABLE srs_cards;
