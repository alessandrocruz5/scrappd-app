package services

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Ping(ctx context.Context) error
	Close() error
	Client() *redis.Client // Expose for future use (caching, sessions, etc.)
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg *config.RedisConfig) (RedisClient, error) {
	var client *redis.Client

	if cfg.URL != "" {
		// Parse URL - handles rediss:// (TLS) automatically
		opt, err := redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid redis url: %w", err)
		}
		client = redis.NewClient(opt)
	} else {
		// Fallback to individual fields (local dev)
		if cfg.Host == "" {
			return nil, fmt.Errorf("redis host is required")
		}
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}

	return &redisClient{client: client}, nil
}

func (c *redisClient) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

func (c *redisClient) Close() error {
	return c.client.Close()
}

func (c *redisClient) Client() *redis.Client {
	return c.client
}

func writeCommand(conn net.Conn, args ...string) error {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(args)))
	for _, arg := range args {
		builder.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}
	_, err := conn.Write([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("redis write failed: %w", err)
	}
	return nil
}

func readOK(reader *bufio.Reader) error {
	line, err := readLine(reader)
	if err != nil {
		return err
	}
	if line != "+OK" {
		return fmt.Errorf("unexpected redis response: %s", line)
	}
	return nil
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("redis read failed: %w", err)
	}
	line = strings.TrimRight(line, "\r\n")
	return line, nil
}
