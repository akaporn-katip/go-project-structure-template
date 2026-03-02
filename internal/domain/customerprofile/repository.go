package customerprofile

import (
	"context"
)

type Repository interface {
	Create(context context.Context, user *CustomerProfile) error
	Update(context context.Context, user *CustomerProfile) error
	Delete(context context.Context, id CustomerID) error
	FindByID(context context.Context, id CustomerID) (*CustomerProfile, error)
	FindByEmail(context context.Context, email Email) (*CustomerProfile, error)
}
