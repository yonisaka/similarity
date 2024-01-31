package main

import (
	"context"
	"github.com/webws/go-moda/logger"
	"github.com/yonisaka/similarity/internal/di"
	"testing"
)

func TestImport(t *testing.T) {
	importUsecase := di.GetImportUsecase()

	ctx := context.Background()
	err := importUsecase.Import(ctx, "sample_lelang.csv")
	if err != nil {
		logger.Errorw("error importing", "err", err)
	}
}

func TestMigrateToQdrant(t *testing.T) {
	importUsecase := di.GetImportUsecase()

	ctx := context.Background()
	err := importUsecase.MigrateToQdrant(ctx)
	if err != nil {
		logger.Errorw("error migrating", "err", err)
	}
}
