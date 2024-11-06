package order

import (
	"context"

	gRPCOrderService "github.com/WM1rr0rB8/contractsTest/gen/go/order_service/v1"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/sfqb"

	domainOrder "software_test/internal/domain/order/model"
	policyOrder "software_test/internal/policy/order"
)

type policy interface {
	SearchOrder(context.Context, sfqb.SFQB) ([]domainOrder.Order, error)
	CreateOrder(context.Context, policyOrder.CreateOrderRequest) (policyOrder.CreateOrderResponse, error)
	SwitchStatus(context.Context, policyOrder.SwitchStatusRequest) error
}

// Controller are used to implement order-service.
type Controller struct {
	gRPCOrderService.UnimplementedOrderServiceServer
	policy policy
}

func NewController(policy policy) *Controller {
	return &Controller{
		policy: policy,
	}
}
