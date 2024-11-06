package order

import (
	"context"

	policyOrder "software_test/internal/policy/order"
)

type policy interface {
	CreateOrder(context.Context, policyOrder.CreateOrderRequest) (policyOrder.CreateOrderResponse, error)
}

type Controller struct {
	orderPolicy policy
}

func NewController(
	orderPolicy policy,
) *Controller {
	return &Controller{
		orderPolicy: orderPolicy,
	}
}
