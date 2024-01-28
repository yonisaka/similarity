package datastore

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/yonisaka/similarity/pkg/di"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	poolMasterOnce sync.Once
	poolSlaveOnce  sync.Once
	poolMaster     *pgxpool.Pool
	poolSlave      *pgxpool.Pool
)

type wrapPool struct {
	pool *pgxpool.Pool
}

func (w *wrapPool) Close() error {
	w.pool.Close()
	return nil
}

// NewBaseRepo returns a base repository.
func NewBaseRepo(dbMaster, dbSlave *pgxpool.Pool) *BaseRepo {
	return &BaseRepo{dbMaster: dbMaster, dbSlave: dbSlave}
}

// BaseRepo is a base repository.
type BaseRepo struct {
	dbMaster *pgxpool.Pool
	dbSlave  *pgxpool.Pool
}

func getConnString(connType string) string {
	envType := "_SLAVE"
	portType := "5433"

	if connType == "master" {
		envType = "_MASTER"
		portType = "5432"
	}

	if os.Getenv("APP_ENV") == "test" || os.Getenv("APP_ENV") == "" {
		return fmt.Sprintf("postgres://test:test@localhost:%s/test?sslmode=disable", portType)
	}

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"+envType),
		os.Getenv("POSTGRES_PASSWORD"+envType),
		os.Getenv("POSTGRES_HOST"+envType)+":"+os.Getenv("POSTGRES_PORT"+envType), // for lint purpose
		os.Getenv("POSTGRES_DB"+envType),
	)
}

// GetDatabaseMaster returns postgresql Pool for Master.
func GetDatabaseMaster() *pgxpool.Pool {
	poolMasterOnce.Do(func() {
		ctx := context.Background()

		var err error

		connString := getConnString("master")

		// Use default config.
		poolMaster, err = pgxpool.New(ctx, connString)
		if err != nil {
			log.Fatalf("failed to connect to timescaleDB pool: %v", err)
		}

		err = poolMaster.Ping(ctx)
		if err != nil {
			log.Fatalf("failed to ping database: %v", err)
		}

		var c io.Closer = &wrapPool{
			pool: poolMaster,
		}

		di.RegisterCloser("TimescaleDB Master Connection", c)
	})

	return poolMaster
}

// GetDatabaseSlave returns postgresql Pool for Slave.
func GetDatabaseSlave() *pgxpool.Pool {
	poolSlaveOnce.Do(func() {
		ctx := context.Background()

		var err error

		isReplica, err := strconv.ParseBool(os.Getenv("IS_REPLICA"))
		if err != nil {
			log.Fatalf("failed to parse is replica: %v", err)
		}

		connString := getConnString("master")

		if isReplica {
			connString = getConnString("slave")
		}

		// Use default config.
		poolSlave, err = pgxpool.New(ctx, connString)
		if err != nil {
			log.Fatalf("failed to connect to timescaleDB pool: %v", err)
		}

		err = poolSlave.Ping(ctx)
		if err != nil {
			log.Fatalf("failed to ping database: %v", err)
		}

		var c io.Closer = &wrapPool{
			pool: poolSlave,
		}

		di.RegisterCloser("TimescaleDB Slave Connection", c)
	})

	return poolSlave
}
