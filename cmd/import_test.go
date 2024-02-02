package main

import (
	"context"
	"github.com/webws/go-moda/logger"
	"github.com/yonisaka/similarity/internal/di"
	"testing"
	"time"
)

func TestImport(t *testing.T) {
	importUsecase := di.GetImportUsecase()

	ctx := context.Background()
	count := 0
	for {
		if count > 50 {
			break
		}

		err := importUsecase.Import(ctx, "sample_lelang.csv")
		if err != nil {
			logger.Errorw("error importing", "err", err)
		}

		time.Sleep(60 * time.Second)

		count += 3
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
