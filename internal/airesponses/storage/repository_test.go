package storage_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	airesponses "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/airesponses/storage"
)

func TestRepositoryUpsertAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := airesponses.NewRepository(sqlxDB)

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO ai_responses (kind, input, native_language, prompt_version, response)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (kind, input, native_language, prompt_version)
		DO UPDATE SET response = EXCLUDED.response, updated_at = now()
	`)).
		WithArgs("definition", "回収", "English", 1, "cached definition").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Upsert(context.Background(), airesponses.KindDefinition, "回収", "English", 1, "cached definition")
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"response"}).AddRow("cached definition")
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT response FROM ai_responses
		WHERE kind = $1 AND input = $2 AND native_language = $3 AND prompt_version = $4
	`)).
		WithArgs("definition", "回収", "English", 1).
		WillReturnRows(rows)

	got, found, err := repo.Get(context.Background(), airesponses.KindDefinition, "回収", "English", 1)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "cached definition", got)

	require.NoError(t, mock.ExpectationsWereMet())
}
