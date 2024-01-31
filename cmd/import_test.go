package main

import (
	"context"
	"github.com/yonisaka/similarity/internal/di"
	"log"
	"testing"
)

func TestImport(t *testing.T) {
	importUsecase := di.GetImportUsecase()

	ctx := context.Background()
	err := importUsecase.Import(ctx, "sample_lelang.csv")
	if err != nil {
		log.Println(err)
	}
}

func TestMigrateToQdrant(t *testing.T) {
	importUsecase := di.GetImportUsecase()

	ctx := context.Background()
	err := importUsecase.MigrateToQdrant(ctx)
	if err != nil {
		log.Println(err)
	}
}
