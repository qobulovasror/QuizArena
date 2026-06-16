-- +goose Up
-- Savol banki (statik savollar) — to'g'ri javob (correct) client'ga hech qachon yuborilmaydi.
CREATE TABLE questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    type        VARCHAR NOT NULL,        -- §5 katalog: mcq|true_false|...
    prompt      TEXT NOT NULL,
    options     JSONB,                   -- variantlar (mcq/match/ordering...)
    correct     JSONB NOT NULL,          -- TO'G'RI JAVOB — faqat serverda
    accept      JSONB,                   -- qabul-ro'yxati/tolerance (type_answer, numeric)
    media_url   TEXT,                    -- rasm/audio turlari uchun
    explanation TEXT,
    difficulty  SMALLINT NOT NULL DEFAULT 1,  -- 1..5
    meta        JSONB,                   -- tarjima, manba, taglar
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_questions_category ON questions (category_id);

-- +goose Down
DROP TABLE questions;
