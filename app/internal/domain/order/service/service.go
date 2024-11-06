package service

import (
	"context"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/errors"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/logging"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/sfqb"

	"software_test/internal/dal"
	domainOrder "software_test/internal/domain/order"
	"software_test/internal/domain/order/model"
)

type storage interface {
	All(context.Context, sfqb.SFQB) ([]model.Order, error)
	CreateOrder(context.Context, model.CreateOrder) error
	SwitchStatus(context.Context, model.SwitchStatus) error
}

type Service struct {
	orderStorage storage
}

func NewService(orderStorage storage) *Service {
	return &Service{
		orderStorage: orderStorage,
	}
}

func (s *Service) All(ctx context.Context, filters sfqb.SFQB) (orders []model.Order, err error) {
	logging.L(ctx).Debug("All")

	orders, err = s.orderStorage.All(ctx, filters)
	if err != nil {
		return nil, errors.Wrap(err, "orderStorage.All")
	}

	return orders, nil
}

func (s *Service) CreateOrder(ctx context.Context, order model.CreateOrder) error {
	logging.L(ctx).Debug("CreateOrder")

	err := s.orderStorage.CreateOrder(ctx, order)
	if err != nil {
		switch {
		case errors.Is(err, domainOrder.ErrViolatesConstraintOrderIdPK):
			return domainOrder.ErrOrderAlreadyExist
		}

		return errors.Wrap(err, "orderStorage.CreateOrder")
	}

	return nil
}

func (s *Service) SwitchStatus(ctx context.Context, order model.SwitchStatus) error {
	err := s.orderStorage.SwitchStatus(ctx, order)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			return domainOrder.ErrOrderNotFound
		}

		return errors.Wrap(err, "orderStorage.SwitchStatus")
	}

	return nil
}
