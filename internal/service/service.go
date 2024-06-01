package service

import (
	"context"
	"strings"

	"github.com/eugene-static/wishlist_bot/internal/entity"
)

type User interface {
	GetUserByID(ctx context.Context, ID int64) (*entity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	AddUser(ctx context.Context, user *entity.User) error
	UpdateUserPassword(ctx context.Context, id int64, new []byte) error
	UpdateUsername(ctx context.Context, id int64, username string) error
}

type List interface {
	CreateWish(ctx context.Context, wish *entity.Wish) error
	GetWishes(ctx context.Context, id int64) ([]*entity.Wish, error)
	DeleteWishes(ctx context.Context, id string) error
}

type Storage interface {
	User
	List
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) GetUser(ctx context.Context, id int64) (*entity.User, error) {
	return s.storage.GetUserByID(ctx, id)
}

func (s *Service) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	return s.storage.GetUserByUsername(ctx, username)
}

func (s *Service) AddUser(ctx context.Context, user *entity.User) error {
	return s.storage.AddUser(ctx, user)
}

func (s *Service) UpdateUser(ctx context.Context, id int64, username string, new []byte) error {
	if new == nil {
		return s.storage.UpdateUsername(ctx, id, username)
	}
	return s.storage.UpdateUserPassword(ctx, id, new)
}

func (s *Service) AddWish(ctx context.Context, wish *entity.Wish) error {
	return s.storage.CreateWish(ctx, wish)
}

func (s *Service) GetWishlistByID(ctx context.Context, id int64) ([]*entity.Wish, error) {
	return s.storage.GetWishes(ctx, id)
}

func (s *Service) DeleteWishes(ctx context.Context, ids []string) error {
	if ids == nil {
		return nil
	}
	idq := strings.Join(ids, ",")
	return s.storage.DeleteWishes(ctx, idq)
}
