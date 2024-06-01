package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/eugene-static/wishlist_bot/internal/entity"
	"github.com/eugene-static/wishlist_bot/lib/config"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(ctx context.Context, cfg *config.Storage) (*Storage, error) {
	err := os.MkdirAll(path.Dir(cfg.Path), 0750)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(cfg.Driver, cfg.Path)
	if err != nil {
		return nil, err
	}
	query := `CREATE TABLE IF NOT EXISTS users(
    			id INT PRIMARY KEY UNIQUE NOT NULL,
    			username TEXT,
    			password BLOB
			 );	
			  CREATE TABLE IF NOT EXISTS wishes(
				id VARCHAR(16) PRIMARY KEY UNIQUE,
				content TEXT,
				user_id INT,
			    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			 )`
	if _, err = db.ExecContext(ctx, query); err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
func (s *Storage) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	query := `SELECT username, password FROM users WHERE id = ?`
	user := &entity.User{ID: id}
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&user.Name, &user.Password); err != nil {
		return nil, err
	}
	return user, nil
}
func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `SELECT id, password FROM users WHERE username = ?`
	user := &entity.User{Name: username}
	if err := s.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Password); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Storage) AddUser(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users(id, username, password) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, user.ID, user.Name, user.Password)
	return err
}

func (s *Storage) UpdateUserPassword(ctx context.Context, id int64, new []byte) error {
	query := `UPDATE users SET password = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, new, id)
	return err
}

func (s *Storage) UpdateUsername(ctx context.Context, id int64, username string) error {
	query := `UPDATE users SET username = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, username, id)
	return err
}

func (s *Storage) CreateWish(ctx context.Context, wish *entity.Wish) error {
	query := `INSERT INTO wishes(id, content, user_id) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, wish.ID, wish.Content, wish.UserID)
	if errors.Is(err, sqlite3.ErrConstraintUnique) {
		return nil
	}
	return err
}

func (s *Storage) GetWishes(ctx context.Context, id int64) ([]*entity.Wish, error) {
	query := `SELECT id, content FROM wishes WHERE user_id = ?`
	var list []*entity.Wish
	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		wish := &entity.Wish{}
		if err = rows.Scan(&wish.ID, &wish.Content); err != nil {
			return nil, err
		}
		list = append(list, wish)
	}
	return list, rows.Err()
}

func (s *Storage) DeleteWishes(ctx context.Context, ids string) error {
	query := fmt.Sprintf(`DELETE FROM wishes WHERE id IN (%s)`, ids)
	_, err := s.db.ExecContext(ctx, query)
	return err
}
