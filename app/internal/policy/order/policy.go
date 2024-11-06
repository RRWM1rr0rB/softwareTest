package order

import (
	"context"
	"software_test/internal/config"

	"github.com/WM1rr0rB8/librariesTest/backend/golang/sfqb"

	"software_test/internal/domain/order/model"
	"software_test/internal/policy"
)

type Service interface {
	All(context.Context, sfqb.SFQB) ([]model.Order, error)
	CreateOrder(context.Context, model.CreateOrder) error
	SwitchStatus(context.Context, model.SwitchStatus) error
}

type Policy struct {
	*policy.BasePolicy
	orderService Service

	cfg *config.Config
}

func NewPolicy(
	basePolicy *policy.BasePolicy,
	orderService Service,
	cfg *config.Config,
) *Policy {
	return &Policy{
		BasePolicy:   basePolicy,
		orderService: orderService,
		cfg:          cfg,
	}
}
