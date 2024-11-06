package order

import (
	"github.com/shopspring/decimal"

	"software_test/internal/domain/order/model"
)

type CreateOrderRequest struct {
	UserID      uint64          `json:"user_id"`
	Status      string          `json:"status"`
	TypeProduct string          `json:"type_product"`
	Price       decimal.Decimal `json:"price"`
	Item        uint32          `json:"package"`
}

type CreateOrderResponse struct {
	Packs []model.Pack `json:"packs"`
}

type SwitchStatusRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func NewSwitchStatusRequest(
	id, status string,
) SwitchStatusRequest {
	return SwitchStatusRequest{
		ID:     id,
		Status: status,
	}
}
