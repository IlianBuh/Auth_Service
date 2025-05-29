package sqlite

import (
	"Service/internal/domain/models"
	"Service/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	e "Service/internal/lib/errors"

	"github.com/mattn/go-sqlite3"
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

		CREATE INDEX IF NOT EXISTS idx_uuid ON users(uuid);

		CREATE TABLE IF NOT EXISTS followings (
			follower INTEGER NOT NULL,
			followee INTEGER NOT NULL,
			PRIMARY KEY(follower, followee),
			FOREIGN KEY (follower) REFERENCES users(uuid) ON DELETE CASCADE,
			FOREIGN KEY (followee) REFERENCES users(uuid) ON DELETE CASCADE
		);

		DROP TABLE IF EXISTS tokens;

		CREATE TABLE IF NOT EXISTS tokens (
			id integer PRIMARY KEY,
			refresh_token TEXT NOT NULL,
			access_token TEXT NOT NULL
		)
		`,
	)
	if err != nil {
		panic("initDB: failed to prepare query - " + err.Error())
	}
}

func (s *Storage) User(ctx context.Context, key interface{}) (models.User, error) {
	switch key.(type) {
	case string:
		return s.UserByLogin(ctx, key.(string))
	case int:
		return s.UserByUUID(ctx, key.(int))
	default:
		return models.User{}, storage.ErrInvalidUserKey
	}
}

func (s *Storage) UserByLogin(ctx context.Context, login string) (models.User, error) {
	const op = "sqlite.UserByLogin"
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

func (s *Storage) UserByUUID(ctx context.Context, uuid int) (models.User, error) {
	const op = "sqlite.UserByUUID"
	var user models.User

	prep, err := s.db.PrepareContext(ctx, "SELECT uuid, login, email, passhash FROM users WHERE uuid=?;")
	if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}
	defer prep.Close()

	row := prep.QueryRowContext(ctx, uuid)

	err = row.Scan(&user.UUID, &user.Login, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}

		return user, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) Users(ctx context.Context, uuids []int) ([]models.User, error) {
	const op = "sqlite.Users"

	prep, err := s.db.PrepareContext(
		ctx,
		`SELECT uuid, login, email, passhash FROM users WHERE uuid IN (`+
			strings.TrimSuffix(strings.Repeat("?,", len(uuids)), ",")+
			`);`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer prep.Close()

	args := make([]interface{}, len(uuids))
	for i, uuid := range uuids {
		args[i] = uuid
	}
	rows, err := prep.QueryContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	users := make([]models.User, 0)
	var user models.User
	for rows.Next() {
		err = rows.Scan(&user.UUID, &user.Login, &user.Email, &user.PassHash)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return users, nil
}

func (s *Storage) UsersByLogin(ctx context.Context, login string) ([]models.User, error) {
	const op = "sqlite.UsersByLogin"
	const slctQuery = `
		SELECT uuid, login, email 
		FROM users
		WHERE login LIKE ?
	`

	rows, err := s.db.QueryContext(ctx, slctQuery, login+"%")
	if err != nil {
		return nil, e.Fail(op, err)
	}
	defer rows.Close()

	users, err := s.scanUsers(rows)
	if err != nil {
		return nil, e.Fail(op, err)
	}

	return users, nil
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

func (s *Storage) Follow(
	ctx context.Context,
	src, target int,
) error {
	const op = "sqlite.Follow"
	const insrtQuery = `
		INSERT INTO followings(follower, followee) VALUES($1, $2);
	`

	_, err := s.db.ExecContext(ctx, insrtQuery, src, target)
	if err != nil {
		return e.Fail(op, err)
	}

	return nil
}

func (s *Storage) Unfollow(
	ctx context.Context,
	src, target int,
) error {
	const op = "sqlite.Unfollow"
	const deleteQuery = `
		DELETE FROM followings WHERE follower=$1 AND followee=$2;
	`

	_, err := s.db.ExecContext(ctx, deleteQuery, src, target)
	if err != nil {
		return e.Fail(op, err)
	}

	return nil

}

func (s *Storage) Followers(
	ctx context.Context,
	uuid int,
) ([]models.User, error) {
	const op = "sqlite.Followers"
	const insrtQuery = `
		SELECT uuid, login, email
		FROM users
		JOIN followings ON followings.follower = users.uuid
		WHERE followee=$1
	`

	rows, err := s.db.QueryContext(ctx, insrtQuery, uuid)
	if err != nil {
		return nil, e.Fail(op, err)
	}
	defer rows.Close()

	users, err := s.scanUsers(rows)
	if err != nil {
		return nil, e.Fail(op, err)
	}

	return users, nil
}

func (s *Storage) Followees(
	ctx context.Context,
	uuid int,
) ([]models.User, error) {
	const op = "sqlite.Followees"
	const insrtQuery = `
		SELECT uuid, login, email
		FROM users
		JOIN followings ON followings.followee = users.uuid
		WHERE follower=$1
	`

	rows, err := s.db.QueryContext(ctx, insrtQuery, uuid)
	if err != nil {
		return nil, e.Fail(op, err)
	}
	defer rows.Close()

	users, err := s.scanUsers(rows)
	if err != nil {
		return nil, e.Fail(op, err)
	}

	return users, nil
}

func (s *Storage) StoreToken(
	ctx context.Context,
	refreshToken, accessToken string,
) error {
	const op = "sqlite.StoreToken"
	const insrtQuery = `
		INSERT INTO tokens(refresh_token, access_token) VALUES(?, ?);
	`
	_, err := s.db.ExecContext(ctx, insrtQuery, refreshToken, accessToken)
	if err != nil {
		return e.Fail(op, err)
	}

	return nil
}

func (s *Storage) Token(
	ctx context.Context,
	refreshToken string,
) (string, error) {
	const op = "sqlite.StoreToken"
	const insrtQuery = `
		SELECT access_token FROM tokens WHERE refresh_token=?;
	`
	row := s.db.QueryRowContext(ctx, insrtQuery, refreshToken)
	var token string
	if err := row.Scan(&token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", e.Fail(op, storage.ErrNotFound)
		}

		return "", e.Fail(op, err)
	}

	return token, nil
}

func (s *Storage) DeleteToken(ctx context.Context, refreshToken string) error {
	const op = "sqlite.DeleteToken"
	const deleteQuery = `
		DELETE FROM tokens WHERE refresh_token = ?;
	`
	_, err := s.db.ExecContext(ctx, deleteQuery, refreshToken)
	if err != nil {
		return e.Fail(op, err)
	}

	return nil
}

func (s *Storage) scanUsers(rows *sql.Rows) ([]models.User, error) {
	const op = "sqlite.scanFollowUsers"

	users := make([]models.User, 0)
	var user models.User
	for rows.Next() {

		if err := rows.Scan(&user.UUID, &user.Login, &user.Email); err != nil {
			return nil, e.Fail(op, err)
		}

		users = append(users, user)
	}

	return users, nil
}
