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
	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		logger.Log.Info("Couldn't begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback() //nolint:all

	// Checking if the provided correlation ID is unique
	stmtCheckCorID, err := tx.PrepareContext(r.ctx, "SELECT full_uri FROM urls WHERE id = $1")
	if err != nil {
		logger.Log.Info("Couldn't prepare SELECT context", zap.Error(err))
		return err
	}
	defer stmtCheckCorID.Close()

	err = stmtCheckCorID.QueryRowContext(r.ctx, url.CorrelationID()).Scan(nil)
	if !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Info("The specified correlation ID is not unique", zap.String("correlation_id", url.CorrelationID()), zap.Error(err))
		return ErrCorIDNotUnique
	}

	// Checking if the provided full_uri is unique
	stmtCheckFullURI, err := tx.PrepareContext(r.ctx, "SELECT key FROM urls WHERE full_uri = $1")
	if err != nil {
		logger.Log.Info("Couldn't prepare SELECT context", zap.Error(err))
		return err
	}
	defer stmtCheckFullURI.Close()

	var oldKey string
	err = stmtCheckFullURI.QueryRowContext(r.ctx, url.LongURL()).Scan(&oldKey)
	if !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Info("The specified full URL is not unique", zap.String("full_uri", url.LongURL()), zap.Error(err))
		return newErrURINotUnique(oldKey)
	}

	// Adding the newly shortened URI to the database
	stmtAddURL, err := tx.PrepareContext(r.ctx, `
		INSERT INTO urls (id, key, full_uri, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE
		SET full_uri = EXCLUDED.full_uri, updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		logger.Log.Info("Couldn't prepare INSERT context", zap.Error(err))
		return err
	}
	defer stmtAddURL.Close()

	_, err = stmtAddURL.ExecContext(r.ctx, url.CorrelationID(), url.ID(), url.LongURL())
	if err != nil {
		logger.Log.Info("Couldn't execute INSERT context", zap.Error(err))
		return err
	}

	return tx.Commit()
}

// Check if there is a URL stored in the Repo with the specified ID
func (r *PostgresRepo) GetURL(key *string) (string, error) {
	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint:all

	stmtGetURL, err := tx.PrepareContext(r.ctx, "SELECT full_uri FROM urls WHERE key = $1")
	if err != nil {
		logger.Log.Info("Couldn't get full_uri with the specified key", zap.Error(err))
		return "", err
	}
	defer stmtGetURL.Close()

	var fullURL string
	err = stmtGetURL.QueryRowContext(r.ctx, key).Scan(&fullURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFoundURL
		}

		logger.Log.Info("Couldn't retrieve shortened URL", zap.Error(err))
		return "", err
	}

	return fullURL, tx.Commit()
}

func (r *PostgresRepo) Ping(ctx context.Context) error {
	if err := r.db.PingContext(r.ctx); err != nil {
		logger.Log.Info("Failed to ping the database", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostgresRepo) initializeDB(config conf.Configuration) error {
	ps := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DB.Host, config.DB.Port, config.DB.User, config.DB.Password, config.DB.DBname, config.DB.SSLmode)
	var err error
	r.db, err = sql.Open("pgx", ps)
	if err != nil {
		logger.Log.Info("Couldn't open the database", zap.Error(err))
		return err
	}
	// defer r.db.Close()

	r.ctx = context.Background()
	if err = r.Ping(r.ctx); err != nil {
		return err
	}

	logger.Log.Info("PostgreSQL initiated succesfully")
	return nil
}

func (r *PostgresRepo) loadTable() error {
	urlsTable := `
		CREATE TABLE IF NOT EXISTS urls (
			id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
			key VARCHAR(100) UNIQUE,
			full_uri VARCHAR(1500) UNIQUE,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)
	`
	_, err := r.db.ExecContext(r.ctx, urlsTable)
	return err
}
