package urls

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	conf "github.com/nomardt/urlshortener-x/cmd/config"
	urlsDomain "github.com/nomardt/urlshortener-x/internal/domain/urls"
	"github.com/nomardt/urlshortener-x/internal/infra/logger"
	"go.uber.org/zap"
)

type PostgresRepo struct {
	db  *sql.DB
	ctx context.Context
}

// Initializes the database with the values provided in config
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

// Executes migrations
func (r *PostgresRepo) loadMigrations(dbName string) error {
	driver, err := postgres.WithInstance(r.db, &postgres.Config{})
	if err != nil {
		logger.Log.Info("Failed to create migrate driver", zap.Error(err))
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		dbName, driver)
	if err != nil {
		logger.Log.Info("Failed to create a new Migrate instance", zap.Error(err))
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Log.Info("Failed to migrate", zap.Error(err))
		return err
	}

	return nil
}

// Initializes the database, creates table urls if not present and returns
// postgresRepo object with fields db and ctx
func NewPostgresRepo(config conf.Configuration) *PostgresRepo {
	postgresRepo := &PostgresRepo{}

	if err := postgresRepo.initializeDB(config); err != nil {
		logger.Log.Info("Error when initializing DB", zap.Error(err))
		log.Fatal(err)
	}

	if err := postgresRepo.loadMigrations(config.DB.DBname); err != nil {
		logger.Log.Info("Error when loading migrations", zap.Error(err))
		log.Fatal(err)
	}

	return postgresRepo
}

// Add the specified URL to the Repo
func (r *PostgresRepo) SaveURL(url urlsDomain.URL, userID string) error {
	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		logger.Log.Info("Couldn't begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback() //nolint:all

	// Checking if the provided correlation ID is unique
	stmtCheckCorID, err := tx.PrepareContext(r.ctx, "SELECT full_uri FROM urls WHERE id = $1")
	if err != nil {
		logger.Log.Info("Couldn't prepare SELECT context for table urls", zap.Error(err))
		return err
	}
	defer stmtCheckCorID.Close()

	err = stmtCheckCorID.QueryRowContext(r.ctx, url.CorrelationID()).Scan(nil)
	if !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Info("The specified correlation ID is not unique", zap.String("correlation_id", url.CorrelationID()), zap.Error(err))
		return ErrCorIDNotUnique
	}

	// Checking if the provided full_uri is unique
	stmtCheckFullURI, err := tx.PrepareContext(r.ctx, "SELECT key FROM urls WHERE full_uri = $1 AND is_deleted = FALSE")
	if err != nil {
		logger.Log.Info("Couldn't prepare SELECT statement for table urls", zap.Error(err))
		return err
	}
	defer stmtCheckFullURI.Close()

	var oldKey string
	err = stmtCheckFullURI.QueryRowContext(r.ctx, url.LongURL()).Scan(&oldKey)
	if !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Info("The specified full URL is not unique", zap.String("full_uri", url.LongURL()), zap.Error(err))

		var userNotAttached bool
		stmtCheckUserAttachement, err := tx.PrepareContext(r.ctx, `
			SELECT EXISTS (
				SELECT 1 FROM urls WHERE full_uri = $1 AND NOT $2 = ANY(users)
			)
		`)
		if err != nil {
			logger.Log.Info("Couldn't prepare SELECT EXISTS statement for user attachement check", zap.Error(err))
			return err
		}
		defer stmtCheckUserAttachement.Close()

		err = stmtCheckUserAttachement.QueryRowContext(r.ctx, url.LongURL(), userID).Scan(&userNotAttached)
		if err != nil {
			logger.Log.Info("Couldn't check if a user is already attached to the URL user wants to shorten", zap.Error(err))
			return err
		}

		if userNotAttached {
			stmtAttachUser, err := tx.PrepareContext(r.ctx, "UPDATE urls SET users = array_append(users, $1) WHERE full_uri = $2")
			if err != nil {
				logger.Log.Info("Couldn't prepare statement to attach user to shortened URL", zap.Error(err))
				return err
			}
			defer stmtAttachUser.Close()

			_, err = stmtAttachUser.ExecContext(r.ctx, userID, url.LongURL())
			if err != nil {
				logger.Log.Info("Couldn't attach user to full_uri", zap.Error(err))
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return newErrURINotUnique(oldKey)
	}

	// Adding the newly shortened URI to the table urls
	stmtAddURL, err := tx.PrepareContext(r.ctx, `
		INSERT INTO urls (id, key, full_uri, users, created_at, updated_at)
		VALUES ($1, $2, $3, ARRAY[$4], CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		logger.Log.Info("Couldn't prepare INSERT statement for table urls", zap.Error(err))
		return err
	}
	defer stmtAddURL.Close()

	_, err = stmtAddURL.ExecContext(r.ctx, url.CorrelationID(), url.ID(), url.LongURL(), userID)
	if err != nil {
		logger.Log.Info("Couldn't execute INSERT context for table urls", zap.Error(err))
		return err
	}

	return tx.Commit()
}

// Sets the flag is_deleted to true if the shortened URL belongs to user
func (r *PostgresRepo) DeleteURL(key string, userID string) error {
	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		logger.Log.Info("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback() //nolint:all

	// Check if the URL with the specified key belongs to user
	stmtURLOwner, err := tx.PrepareContext(r.ctx, "SELECT users[1] AS owner FROM urls WHERE key = $1")
	if err != nil {
		logger.Log.Info("Failed to prepare statement to lookup the URL owner", zap.Error(err))
		return err
	}
	defer stmtURLOwner.Close()

	var owner string
	err = stmtURLOwner.QueryRowContext(r.ctx, key).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Info(ErrNotFoundURL.Error(), zap.String("ID", key))
		return ErrNotFoundURL
	}

	if owner != userID {
		logger.Log.Info(ErrNotOwner.Error(), zap.String("user", userID), zap.String("ID", key))
		return ErrNotOwner
	}

	// Update the is_deleted to true
	stmtDeleteURL, err := tx.PrepareContext(r.ctx, "UPDATE urls SET is_deleted = TRUE WHERE key = $1")
	if err != nil {
		logger.Log.Info("Couldn't prepare statement for is_deleted update", zap.Error(err))
		return err
	}
	defer stmtDeleteURL.Close()

	_, err = stmtDeleteURL.ExecContext(r.ctx, key)
	if err != nil {
		logger.Log.Info("Failed to update is_deleted to true", zap.Error(err), zap.String("key", key))
		return err
	}

	return tx.Commit()
}

// Check if there is a URL stored in the Repo with the specified ID
func (r *PostgresRepo) GetURL(key string) (string, error) {
	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		logger.Log.Info("Failed to begin transaction", zap.Error(err))
		return "", err
	}
	defer tx.Rollback() //nolint:all

	stmtGetURL, err := tx.PrepareContext(r.ctx, "SELECT full_uri, is_deleted FROM urls WHERE key = $1")
	if err != nil {
		logger.Log.Info("Couldn't prepare statement to get the specified URL", zap.Error(err))
		return "", err
	}
	defer stmtGetURL.Close()

	var fullURL string
	var isDeleted bool
	err = stmtGetURL.QueryRowContext(r.ctx, key).Scan(&fullURL, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Info(ErrNotFoundURL.Error())
			return "", ErrNotFoundURL
		}

		logger.Log.Info("Couldn't retrieve shortened URL", zap.Error(err))
		return "", err
	}
	err = tx.Commit()

	if isDeleted {
		logger.Log.Info(ErrURLGone.Error(), zap.String("ID", key))
		return "", ErrURLGone
	}

	return fullURL, err
}

// Returns all URLs connected to the user specified. Output is urls[key] = originalURL
func (r *PostgresRepo) GetAllUserURLs(userID string) (map[string]string, error) {
	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		logger.Log.Info("Couldn't begin transaction", zap.Error(err))
		return nil, err
	}
	defer tx.Rollback() //nolint:all

	stmt, err := tx.PrepareContext(r.ctx, `
		SELECT key, full_uri FROM urls
		WHERE $1 = ANY(users) AND is_deleted = FALSE
	`)
	if err != nil {
		logger.Log.Info("Couldn't prepare a statement to get keys and full_uris of the specified user ID", zap.Error(err))
		return nil, err
	}

	rows, err := stmt.QueryContext(r.ctx, userID)
	if err != nil {
		logger.Log.Info("Couldn't query keys and full_uris for the specified user ID", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	urls := make(map[string]string, 0)
	for rows.Next() {
		var key, fullURI string
		err := rows.Scan(&key, &fullURI)
		if err != nil {
			logger.Log.Info("Couldn't retrieve a URL", zap.Error(err))
		}
		urls[key] = fullURI
	}

	if err := rows.Err(); err != nil {
		logger.Log.Info("Error occured during rows iteration", zap.Error(err))
		return nil, err
	}

	return urls, tx.Commit()
}

func (r *PostgresRepo) Ping(ctx context.Context) error {
	if err := r.db.PingContext(r.ctx); err != nil {
		logger.Log.Info("Failed to ping the database", zap.Error(err))
		return err
	}
	return nil
}
