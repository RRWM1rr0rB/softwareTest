package order

import (
	gRPCOrderService "github.com/WM1rr0rB8/contractsTest/gen/go/order_service/v1"
	"github.com/shopspring/decimal"

	domainOrder "software_test/internal/domain/order/model"
	policyOrder "software_test/internal/policy/order"
)

const (
	validationErrCode = iota + 100
	minSearchErrCode
)

func decodeCreateOrderRequest(
	data *gRPCOrderService.CreateOrderRequest,
) policyOrder.CreateOrderRequest {
	price, _ := decimal.NewFromString(data.GetPrice())

	return policyOrder.CreateOrderRequest{
		UserID:      data.GetUserId(),
		Status:      data.GetStatus(),
		TypeProduct: data.GetTypeProduct(),
		Price:       price,
		Item:        data.GetItem(),
	}
}

func convertPack(pack domainOrder.Pack) *gRPCOrderService.Pack {
	return &gRPCOrderService.Pack{
		Size:  int32(pack.Size),
		Count: int32(pack.Count),
	}
}

func newSearchOrderResponse(
	data []domainOrder.Order,
) *gRPCOrderService.SearchOrderResponse {
	response := make([]*gRPCOrderService.Order, len(data))

	for i := 0; i < len(data); i++ {
		b := data[i]

		packs := make([]*gRPCOrderService.Pack, len(b.Pack))
		for j := 0; j < len(b.Pack); j++ {
			packs[j] = convertPack(b.Pack[j])
		}

		order := &gRPCOrderService.Order{
			Id:          b.ID,
			UserId:      b.UserID,
			NumberOrder: b.NumberOrder,
			Status:      b.Status,
			TypeProduct: b.TypeProduct,
			Price:       b.Price.String(),
			Item:        b.Item,
			Packs:       packs,
			CreatedAt:   b.CreatedAt.UnixMilli(),
			UpdatedAt:   b.UpdatedAt.UnixMilli(),
		}

		response[i] = order
	}

	resp := &gRPCOrderService.SearchOrderResponse{
		Orders: response,
	}

	return resp
}
