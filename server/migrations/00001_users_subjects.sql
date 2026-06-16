-- +goose Up
-- Foydalanuvchilar (gibrid: mehmon + akkaunt + telegram)
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR UNIQUE,                 -- akkaunt uchun; mehmon uchun NULL
    email         VARCHAR UNIQUE,
    password_hash VARCHAR,                        -- mehmon/telegram uchun NULL
    telegram_id   BIGINT UNIQUE,                  -- Telegram orqali (Bosqich 2)
    is_guest      BOOLEAN NOT NULL DEFAULT false,
    role          VARCHAR NOT NULL DEFAULT 'user',-- 'user' | 'admin'
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Soha (ingliz tili, matematika, IT, umumiy bilim)
CREATE TABLE subjects (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR NOT NULL UNIQUE,                 -- 'english', 'math', ...
    name VARCHAR NOT NULL,
    icon VARCHAR
);

-- Kategoriya (soha ichida)
CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    slug       VARCHAR NOT NULL,
    name       VARCHAR NOT NULL,
    UNIQUE (subject_id, slug)
);

-- +goose Down
DROP TABLE categories;
DROP TABLE subjects;
DROP TABLE users;
