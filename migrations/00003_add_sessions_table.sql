-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "sessions" (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "task_id" UUID REFERENCES tasks(id) ON DELETE CASCADE,
    "bpm" VARCHAR(50),
    "note" TEXT,
    "confidence" INT,
    "start_time" TIMESTAMP WITH TIME ZONE,
    "end_time" TIMESTAMP WITH TIME ZONE,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "sessions";
-- +goose StatementEnd
