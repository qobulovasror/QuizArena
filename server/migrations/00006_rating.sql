-- +goose Up
-- 🏆 Matchmaking reytingi (1v1 duel, subject bo'yicha ELO)
CREATE TABLE user_rating (
    user_id     UUID NOT NULL REFERENCES users(id),
    subject_id  UUID NOT NULL REFERENCES subjects(id),
    rating      INT NOT NULL DEFAULT 1000,   -- ELO
    games       INT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, subject_id)
);

-- +goose Down
DROP TABLE user_rating;
