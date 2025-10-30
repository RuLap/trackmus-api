-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "medias" (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "task_id" UUID REFERENCES tasks(id) ON DELETE CASCADE,
    "type" VARCHAR(50),
    "filename" TEXT,
    "size" INT,
    "duration" INT,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "medias";
-- +goose StatementEnd
