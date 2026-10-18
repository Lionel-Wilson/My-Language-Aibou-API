-- +goose Up
CREATE TABLE ai_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind VARCHAR(50) NOT NULL,
    input TEXT NOT NULL,
    native_language VARCHAR(100) NOT NULL,
    prompt_version INTEGER NOT NULL,
    response TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT uq_ai_responses_lookup UNIQUE (kind, input, native_language, prompt_version)
);

-- +goose Down
DROP TABLE IF EXISTS ai_responses;
