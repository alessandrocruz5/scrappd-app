package services

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
)

type RedisClient interface {
	Ping(ctx context.Context) error
	Close() error
}

type redisClient struct {
	addr     string
	password string
	db       int
	timeout  time.Duration
}

func NewRedisClient(cfg *config.RedisConfig) (RedisClient, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("redis host is required")
	}
	if cfg.Port == "" {
		return nil, fmt.Errorf("redis port is required")
	}

	return &redisClient{
		addr:     net.JoinHostPort(cfg.Host, cfg.Port),
		password: cfg.Password,
		db:       cfg.DB,
		timeout:  2 * time.Second,
	}, nil
}

func (c *redisClient) Ping(ctx context.Context) error {
	dialer := &net.Dialer{Timeout: c.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("redis dial failed: %w", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	if c.password != "" {
		if err := writeCommand(conn, "AUTH", c.password); err != nil {
			return err
		}
		if err := readOK(reader); err != nil {
			return fmt.Errorf("redis auth failed: %w", err)
		}
	}

	if c.db != 0 {
		if err := writeCommand(conn, "SELECT", strconv.Itoa(c.db)); err != nil {
			return err
		}
		if err := readOK(reader); err != nil {
			return fmt.Errorf("redis select failed: %w", err)
		}
	}

	if err := writeCommand(conn, "PING"); err != nil {
		return err
	}

	line, err := readLine(reader)
	if err != nil {
		return err
	}
	if line != "+PONG" {
		return fmt.Errorf("unexpected redis ping response: %s", line)
	}
	return nil
}

func (c *redisClient) Close() error {
	return nil
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
