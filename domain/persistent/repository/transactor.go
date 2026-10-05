package repository

import (
	"context"

	"gorm.io/gorm"
)

type Transactor interface {
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}
