package storage

import (
	"Service/internal/domain/models"
	"context"
	"fmt"
)

type Plug struct{}

func (p *Plug) User(ctx context.Context, login string) (models.User, error) {
	fmt.Println(login)
	return models.User{}, nil
}

func (p *Plug) Save(ctx context.Context, login, email string, passHash []byte) (uint64, error) {
	fmt.Println(login, email, string(passHash))
	return 0, nil
}
