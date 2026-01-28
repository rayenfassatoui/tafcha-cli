package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// PostgresConfig holds database connection configuration.
type PostgresConfig struct {
	URL         string
	MaxConns    int32
	MinConns    int32
	MaxConnLife time.Duration
}

// NewPostgresRepository creates a new PostgreSQL repository.
func NewPostgresRepository(ctx context.Context, cfg PostgresConfig, logger *slog.Logger) (*PostgresRepository, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing database URL: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLife

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	repo := &PostgresRepository{
		pool:   pool,
		logger: logger,
	}

	return repo, nil
}

// Migrate runs database migrations.
func (r *PostgresRepository) Migrate(ctx context.Context) error {
	migrationSQL, err := migrationsFS.ReadFile("migrations/001_create_snippets.sql")
	if err != nil {
		return fmt.Errorf("reading migration file: %w", err)
	}

	_, err = r.pool.Exec(ctx, string(migrationSQL))
	if err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}

	r.logger.Info("database migration completed")
	return nil
}

// Create stores a new snippet.
func (r *PostgresRepository) Create(id string, content []byte, expiresAt time.Time) (*Snippet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO snippets (id, content, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING created_at
	`

	var createdAt time.Time
	err := r.pool.QueryRow(ctx, query, id, content, expiresAt).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("inserting snippet: %w", err)
	}

	return &Snippet{
		ID:        id,
		Content:   content,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}, nil
}

// Get retrieves a snippet by ID. Returns nil if not found or expired.
func (r *PostgresRepository) Get(id string) (*Snippet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, content, expires_at, created_at
		FROM snippets
		WHERE id = $1 AND expires_at > NOW()
	`

	var s Snippet
	err := r.pool.QueryRow(ctx, query, id).Scan(&s.ID, &s.Content, &s.ExpiresAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying snippet: %w", err)
	}

	return &s, nil
}

// Delete removes a snippet by ID.
func (r *PostgresRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.pool.Exec(ctx, "DELETE FROM snippets WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deleting snippet: %w", err)
	}
	return nil
}

// DeleteExpired removes all expired snippets.
func (r *PostgresRepository) DeleteExpired() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := r.pool.Exec(ctx, "DELETE FROM snippets WHERE expires_at <= NOW()")
	if err != nil {
		return 0, fmt.Errorf("deleting expired snippets: %w", err)
	}

	count := result.RowsAffected()
	if count > 0 {
		r.logger.Info("deleted expired snippets", "count", count)
	}

	return count, nil
}

// Close releases database connections.
func (r *PostgresRepository) Close() {
	r.pool.Close()
}

// Ping checks database connectivity.
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
