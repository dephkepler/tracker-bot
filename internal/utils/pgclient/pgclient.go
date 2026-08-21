package pgclient

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Client struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Client, error) {
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("error get database driver %w", err)
	}

	if err = db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error connection to database %w", err)
	}

	log.Info().Msg("database connection established")

	r := &Client{
		db: db,
	}
	return r, nil
}
func (c *Client) Pool() *pgxpool.Pool {
	return c.db
}

func (c *Client) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
