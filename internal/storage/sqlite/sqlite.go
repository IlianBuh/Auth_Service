package sqlite

import (
	"Service/internal/domain/models"
	"Service/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"time"
)

type Storage struct {
	db *sql.DB
}

// New creates instance of storage using sqlite
func New(storagePath string) *Storage {

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		panic("failed to open database: " + err.Error())
	}

	initDB(db)

	return &Storage{
		db: db,
	}
}

// initDB creates 'users' table if the table does not exist
func initDB(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS users(
					uuid INTEGER PRIMARY KEY,
					login TEXT NOT NULL UNIQUE,
					email TEXT NOT NULL UNIQUE,
					passhash BLOB NOT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_uuid ON users(uuid)`,
	)
	if err != nil {
		panic("initDB: failed to prepare query - " + err.Error())
	}
}

func (s *Storage) User(ctx context.Context, login string) (models.User, error) {
	const op = "sqlite.User"
	var user models.User

	prep, err := s.db.PrepareContext(ctx, "SELECT uuid, login, email, passhash FROM users WHERE login=?;")
	if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}
	defer prep.Close()

	row := prep.QueryRowContext(ctx, login)

	err = row.Scan(&user.UUID, &user.Login, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}

		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) Save(ctx context.Context, login, email string, passHash []byte) (uint64, error) {
	const op = "sqlite.Save"
	uuid := uint64(0)

	prep, err := s.db.PrepareContext(ctx, "INSERT INTO users(login, email, passhash) VALUES(?, ?, ?);")
	if err != nil {
		return uuid, fmt.Errorf("%s: %w", op, err)
	}
	defer prep.Close()

	res, err := prep.ExecContext(ctx, login, email, passHash)
	if err != nil {
		var sqlerr sqlite3.Error
		if errors.As(err, &sqlerr) {
			switch {
			case errors.Is(sqlerr.ExtendedCode, sqlite3.ErrConstraintNotNull):
				return uuid, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
			case errors.Is(sqlerr.ExtendedCode, sqlite3.ErrConstraintUnique):
				return uuid, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
			}
		}

		return uuid, fmt.Errorf("%s: %w", op, err)
	}

	temp, err := res.LastInsertId()
	if err != nil {
		return uuid, fmt.Errorf("%s: %w", op, err)
	}

	uuid = uint64(temp)
	return uuid, nil
}
