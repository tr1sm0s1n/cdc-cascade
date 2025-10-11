package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConn struct {
	Postgres *gorm.DB
	Redis    *redis.Client
}

func NewDBConn(ctx context.Context) (*DBConn, error) {
	db := &DBConn{}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := db.connectPostgres(ctx); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if err := db.connectRedis(ctx); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return db, nil
}

func (db *DBConn) connectPostgres(ctx context.Context) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	conn, err := connectWithRetry(ctx, func() (*gorm.DB, error) {
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}, "PostgreSQL", 15*time.Second)

	if err != nil {
		return err
	}

	db.Postgres = conn
	return nil
}

func (db *DBConn) connectRedis(ctx context.Context) error {
	url := fmt.Sprintf("redis://default:%s@%s:%s/0",
		os.Getenv("REDIS_PASS"),
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
	)

	conn, err := connectWithRetry(ctx, func() (*redis.Client, error) {
		opt, err := redis.ParseURL(url)
		if err != nil {
			return nil, err
		}

		client := redis.NewClient(opt)
		if err := client.Ping(ctx).Err(); err != nil {
			return nil, err
		}

		return client, nil
	}, "Redis", 15*time.Second)

	if err != nil {
		return err
	}

	db.Redis = conn
	return nil
}

func connectWithRetry[T any](
	ctx context.Context,
	fn func() (T, error),
	name string,
	timeout time.Duration,
) (T, error) {
	var zero T
	start := time.Now()
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
			conn, err := fn()
			if err == nil {
				return conn, nil
			}

			if time.Since(start) > timeout {
				return zero, fmt.Errorf("failed to connect to %s after %v: %w", name, timeout, err)
			}

			sleep := time.Duration(math.Min(float64(time.Second)*math.Pow(2, float64(attempt)), float64(5*time.Second)))
			attempt++

			log.Printf("\033[33m[WRN]\033[0m Waiting for %s to be ready (retry in %v)...", name, sleep)
			time.Sleep(sleep)
		}
	}
}
