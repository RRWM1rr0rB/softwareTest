package order

import (
	"context"

	gRPCOrderService "github.com/WM1rr0rB8/contractsTest/gen/go/order_service/v1"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/errors"

	policyOrder "software_test/internal/policy/order"
)

// CreateOrder order.
func (c *Controller) CreateOrder(
	ctx context.Context,
	data *gRPCOrderService.CreateOrderRequest,
) (*gRPCOrderService.CreateOrderResponse, error) {
	req := decodeCreateOrderRequest(data)

	packs, err := c.policy.CreateOrder(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "policy.CreateOrder")
	}

	var grpcPacks []*gRPCOrderService.Pack
	for _, pack := range packs.Packs {
		grpcPack := &gRPCOrderService.Pack{
			Size:  int32(pack.Size),
			Count: int32(pack.Count),
		}
		grpcPacks = append(grpcPacks, grpcPack)
	}

	response := &gRPCOrderService.CreateOrderResponse{
		Packs: grpcPacks,
	}

	return response, nil
}

// SearchOrder implements order-service , search for all field .
func (c *Controller) SearchOrder(
	ctx context.Context,
	data *gRPCOrderService.SearchOrderRequest,
) (*gRPCOrderService.SearchOrderResponse, error) {
	filters, bvfErr := buildValidationOrderFilters(data)
	if bvfErr != nil {
		return nil, errors.Wrap(bvfErr, "buildValidationOrderFilters")
	}

	output, err := c.policy.SearchOrder(ctx, filters)
	if err != nil {
		return nil, errors.Wrap(err, "policy.SearchOrder")
	}

	resp := newSearchOrderResponse(output)

	return resp, nil
}

// SwitchStatusOrder switch  status order(create, accepted, sent, delivered)..
func (c *Controller) SwitchStatusOrder(
	ctx context.Context, data *gRPCOrderService.SwitchStatusOrderRequest,
) (*gRPCOrderService.SwitchStatusOrderResponse, error) {
	if switchErr := c.policy.SwitchStatus(
		ctx,
		policyOrder.NewSwitchStatusRequest(data.Id, data.Status)); switchErr != nil {
		return nil, errors.Wrap(switchErr, "policy.SwitchStatus")
	}

	return &gRPCOrderService.SwitchStatusOrderResponse{}, nil
}
