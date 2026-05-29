package database

import (
	"context"
	"fmt"

	_ "github.com/lib/pq"

	"Eve/internal/database/ent"
	"Eve/internal/logger"
)

var Default *Client

type Client struct {
	ent *ent.Client
}

func Init(dsn string) error {
	c, err := Open(dsn)
	if err != nil {
		return err
	}
	Default = c
	return nil
}

func Open(dsn string) (*Client, error) {
	client, err := ent.Open("postgres", dsn, ent.Log(func(args ...any) {
		logger.DB(fmt.Sprint(args...))
	}))
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	client = client.Debug()
	return &Client{ent: client}, nil
}

func (c *Client) Close() error {
	return c.ent.Close()
}

func (c *Client) Ent() *ent.Client {
	return c.ent
}

func (c *Client) Migrate(ctx context.Context) error {
	if err := c.ent.Schema.Create(ctx); err != nil {
		return fmt.Errorf("running schema migration: %w", err)
	}
	return nil
}
