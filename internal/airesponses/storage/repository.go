package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Kind identifies which AI prompt produced a stored response. Bump the matching
// prompt version constant in the service when that prompt changes.
type Kind string

const (
	KindDefinition   Kind = "definition"
	KindSynonyms     Kind = "synonyms"
	KindHistory      Kind = "history"
	KindContext      Kind = "context"
	KindExplanation  Kind = "explanation"
	KindTranslation  Kind = "translation"
	KindCorrection   Kind = "correction"
)

//go:generate mockgen -source=repository.go -destination=mock/repository.go
type Repository interface {
	Get(ctx context.Context, kind Kind, input, nativeLanguage string, promptVersion int) (string, bool, error)
	Upsert(ctx context.Context, kind Kind, input, nativeLanguage string, promptVersion int, response string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Get(ctx context.Context, kind Kind, input, nativeLanguage string, promptVersion int) (string, bool, error) {
	const q = `
		SELECT response FROM ai_responses
		WHERE kind = $1 AND input = $2 AND native_language = $3 AND prompt_version = $4
	`
	var response string
	err := r.db.GetContext(ctx, &response, q, string(kind), input, nativeLanguage, promptVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("get ai response: %w", err)
	}
	return response, true, nil
}

func (r *repository) Upsert(ctx context.Context, kind Kind, input, nativeLanguage string, promptVersion int, response string) error {
	const q = `
		INSERT INTO ai_responses (kind, input, native_language, prompt_version, response)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (kind, input, native_language, prompt_version)
		DO UPDATE SET response = EXCLUDED.response, updated_at = now()
	`
	_, err := r.db.ExecContext(ctx, q, string(kind), input, nativeLanguage, promptVersion, response)
	if err != nil {
		return fmt.Errorf("upsert ai response: %w", err)
	}
	return nil
}
