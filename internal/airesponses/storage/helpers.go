package storage

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const persistTimeout = 10 * time.Second

func GetCached(
	ctx context.Context,
	logger *zap.Logger,
	repo Repository,
	kind Kind,
	input, nativeLanguage string,
	promptVersion int,
) (string, bool) {
	response, found, err := repo.Get(ctx, kind, input, nativeLanguage, promptVersion)
	if err != nil {
		logger.Error("ai response get failed",
			zap.String("kind", string(kind)),
			zap.String("input", input),
			zap.String("nativeLanguage", nativeLanguage),
			zap.Error(err),
		)
		return "", false
	}
	return response, found
}

func StoreCached(
	_ context.Context,
	logger *zap.Logger,
	repo Repository,
	kind Kind,
	input, nativeLanguage string,
	promptVersion int,
	response string,
) {
	// Never use the request context for writes: lookups can take several seconds of
	// OpenAI time, and proxies/clients often cancel the request before we persist.
	persistCtx, cancel := context.WithTimeout(context.Background(), persistTimeout)
	defer cancel()

	if err := repo.Upsert(persistCtx, kind, input, nativeLanguage, promptVersion, response); err != nil {
		logger.Error("ai response upsert failed",
			zap.String("kind", string(kind)),
			zap.String("input", input),
			zap.String("nativeLanguage", nativeLanguage),
			zap.Error(err),
		)
		return
	}

	logger.Info("ai response stored",
		zap.String("kind", string(kind)),
		zap.String("input", input),
		zap.String("nativeLanguage", nativeLanguage),
	)
}
