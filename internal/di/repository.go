package di

import (
	"github.com/yonisaka/similarity/internal/entities/repository"
	"github.com/yonisaka/similarity/internal/infrastructure/datastore"
)

// GetBaseRepo returns BaseRepo instance.
func GetBaseRepo() *datastore.BaseRepo {
	return datastore.NewBaseRepo(datastore.GetDatabaseMaster(), datastore.GetDatabaseSlave())
}

// GetEmbeddingRepo returns EmbeddingRepo instance.
func GetEmbeddingRepo() repository.EmbeddingRepo {
	return datastore.NewEmbeddingRepo(GetBaseRepo())
}
