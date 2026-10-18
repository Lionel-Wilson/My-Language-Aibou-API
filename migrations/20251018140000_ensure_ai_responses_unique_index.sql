-- +goose Up
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'ai_responses'::regclass
          AND conname = 'uq_ai_responses_lookup'
    ) THEN
        ALTER TABLE ai_responses
            ADD CONSTRAINT uq_ai_responses_lookup
                UNIQUE (kind, input, native_language, prompt_version);
    END IF;
END $$;

-- +goose Down
ALTER TABLE ai_responses DROP CONSTRAINT IF EXISTS uq_ai_responses_lookup;
