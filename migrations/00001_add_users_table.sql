-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "sessions" (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "email" VARCHAR(100) UNIQUE NOT NULL,
    "provider" VARCHAR(50) DEFAULT 'local',
    "provider_id" VARCHAR(255),
    "email_confirmed" BOOLEAN DEFAULT FALSE,
    "password" TEXT,
    "first_name" VARCHAR(50),
    "last_name" VARCHAR(50),
    "username" VARCHAR(50) UNIQUE,
    "avatar_url" TEXT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "users";
-- +goose StatementEnd
