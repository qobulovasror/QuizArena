-- +goose Up
-- 📊 Baholash: soha/kategoriya bo'yicha mastery (vaqt bo'yicha o'zgaradi)
CREATE TABLE user_mastery (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    mastery     DOUBLE PRECISION NOT NULL DEFAULT 50,  -- 0..100
    attempts    INT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, category_id)
);

-- +goose Down
DROP TABLE user_mastery;
