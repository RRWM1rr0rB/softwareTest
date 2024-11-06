package model

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/WM1rr0rB8/librariesTest/backend/golang/logging"
)

type Order struct {
	ID          string          `json:"id"`
	UserID      uint64          `json:"user_id"`
	NumberOrder uint64          `json:"number_order"`
	Status      string          `json:"status"`
	TypeProduct string          `json:"type_product"`
	Price       decimal.Decimal `json:"price"`
	Item        uint32          `json:"package"`
	Pack        []Pack          `json:"pack"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (c Order) LogValue() logging.Value {
	return logging.GroupValue(
		logging.StringAttr("id", c.ID),
		logging.Uint64Attr("user_id", c.UserID),
		logging.Uint64Attr("number_order", c.NumberOrder),
		logging.StringAttr("status", c.Status),
		logging.StringAttr("type_product", c.TypeProduct),
		logging.StringAttr("price", c.Price.String()),
		logging.UInt32Attr("package", c.Item),
		logging.TimeAttr("created_at", c.CreatedAt),
		logging.TimeAttr("updated_at", c.UpdatedAt),
	)
}

type Pack struct {
	Size  int `json:"size"`
	Count int `json:"count"`
}

type CreateOrder struct {
	ID          string          `json:"id"`
	UserID      uint64          `json:"user_id"`
	Status      string          `json:"status"`
	TypeProduct string          `json:"type_product"`
	Price       decimal.Decimal `json:"price"`
	Item        uint32          `json:"package"`
	Pack        []Pack          `json:"pack"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (c CreateOrder) LogValue() logging.Value {
	return logging.GroupValue(
		logging.StringAttr("id", c.ID),
		logging.Uint64Attr("user_id", c.UserID),
		logging.StringAttr("status", c.Status),
		logging.StringAttr("type_product", c.TypeProduct),
		logging.StringAttr("price", c.Price.String()),
		logging.UInt32Attr("package", c.Item),
		logging.TimeAttr("created_at", c.CreatedAt),
		logging.TimeAttr("updated_at", c.UpdatedAt),
	)
}

func NewCreateOrder(
	id string,
	userID uint64,
	status, typeProduct string,
	price decimal.Decimal,
	item uint32,
	pack []Pack,
	createdAt, updatedAt time.Time,
) CreateOrder {
	return CreateOrder{
		ID:          id,
		UserID:      userID,
		Status:      status,
		TypeProduct: typeProduct,
		Price:       price,
		Item:        item,
		Pack:        pack,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

type SwitchStatus struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c SwitchStatus) LogValue() logging.Value {
	return logging.GroupValue(
		logging.StringAttr("id", c.ID),
		logging.StringAttr("status", c.Status),
		logging.TimeAttr("updated_at", c.UpdatedAt),
	)
}

func NewSwitchStatus(
	id string,
	status string,
	updatedAt time.Time,
) SwitchStatus {
	return SwitchStatus{
		ID:        id,
		Status:    status,
		UpdatedAt: updatedAt,
	}
}
