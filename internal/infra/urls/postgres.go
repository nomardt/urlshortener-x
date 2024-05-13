package urls

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	conf "github.com/nomardt/urlshortener-x/cmd/config"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	"go.uber.org/zap"
)

type PostgresRepo struct {
	db  *sql.DB
	ctx context.Context
}

func NewPostgresRepo(config conf.Configuration) *PostgresRepo {
	postgresRepo := &PostgresRepo{}

	if err := postgresRepo.initializeDB(config); err != nil {
		logger.Log.Info("Error when initializing DB", zap.Error(err))
		log.Fatal(err)
	}

	if err := postgresRepo.loadTable(); err != nil {
		logger.Log.Info("Error when loading table urls", zap.Error(err))
		log.Fatal(err)
	}

	return postgresRepo
}

// Add the specified URL to the Repo
func (r *PostgresRepo) SaveURL(url *urlsDomain.URL) error {
	addURL := `
		INSERT INTO urls (key, full_uri, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE
		SET full_uri = EXCLUDED.full_uri, updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(r.ctx, addURL, url.ID(), url.LongURL())
	return err
}

// Check if there is a URL stored in the Repo with the specified ID
func (r *PostgresRepo) GetURL(id *string) (string, error) {
	getURL := `
		SELECT full_uri FROM urls WHERE key = $1
	`
	var full_uri string
	err := r.db.QueryRowContext(r.ctx, getURL, id).Scan(&full_uri)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFoundURL
		}

		logger.Log.Info("Couldn't retrieve shortened URL", zap.Error(err))
		return "", err
	}

	return full_uri, nil
}

func (r *PostgresRepo) initializeDB(config conf.Configuration) error {
	ps := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DB.Host, config.DB.Port, config.DB.User, config.DB.Password, config.DB.DBname, config.DB.SSLmode)
	var err error
	r.db, err = sql.Open("pgx", ps)
	if err != nil {
		return err
	}
	// defer r.db.Close()

	r.ctx = context.Background()
	if err = r.db.PingContext(r.ctx); err != nil {
		return err
	}

	logger.Log.Info("PostgreSQL initiated succesfully")
	return nil
}

func (r *PostgresRepo) loadTable() error {
	urlsTable := `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			key VARCHAR(100) UNIQUE,
			full_uri VARCHAR(1500),
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)
	`
	_, err := r.db.ExecContext(r.ctx, urlsTable)
	return err
}
