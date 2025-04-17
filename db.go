package main

import (
	"context"
	"sync"

	queries "github.com/adisuper94/orcidparser/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

var q *queries.Queries
var pool *pgxpool.Pool
var mutex = &sync.Mutex{}

func getDBConn() *pgxpool.Pool {
	if pool == nil {
		mutex.Lock()
		if pool == nil {
			pool = createDBConnection(32)
		}
		mutex.Unlock()
	}

	return pool
}

func GetQueries() *queries.Queries {
	pool := getDBConn()
	if q == nil {
		mutex.Lock()
		if q == nil {
			q = queries.New(pool)
		}
		mutex.Unlock()
	}
	return q
}

func createDBConnection(connectionCount int32) *pgxpool.Pool {
	pgxConfig, err := pgxpool.ParseConfig("postgres://adisuper:ilovepostgres@localhost:5432/orcid?sslmode=disable")
	if err != nil {
		panic(err)
	}
	pgxConfig.MaxConns = connectionCount

	conn, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		panic(err)
	}
	return conn
}
