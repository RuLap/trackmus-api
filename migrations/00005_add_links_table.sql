-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "links" (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "task_id" UUID REFERENCES tasks(id) ON DELETE CASCADE,
    "url" TEXT,
    "title" VARCHAR(50),
    "type" VARCHAR(50),
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "links";
-- +goose StatementEnd
