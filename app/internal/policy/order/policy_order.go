package order

import (
	"context"
	"sort"

	"github.com/WM1rr0rB8/librariesTest/backend/golang/errors"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/logging"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/sfqb"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/tracing"

	domainOrder "software_test/internal/domain/order"
	"software_test/internal/domain/order/model"
)

func (p *Policy) SearchOrder(ctx context.Context, filters sfqb.SFQB) ([]model.Order, error) {
	ctx, span := tracing.Continue(ctx, "orderPolicy.SearchOrder")
	defer span.End()

	tracing.TraceAny(ctx, "filters", filters)

	logging.L(ctx).Debug("SearchOrder")

	res, err := p.orderService.All(ctx, filters)
	if err != nil {
		return nil, errors.Wrap(err, "orderService.All")
	}

	return res, nil
}

func (p *Policy) CreateOrder(ctx context.Context, input CreateOrderRequest) (CreateOrderResponse, error) {
	ctx, span := tracing.Continue(ctx, "orderPolicy.CreateOrder")
	defer span.End()

	tracing.TraceAny(ctx, "req", input)

	logging.L(ctx).Debug("CreateOrder", "input", input)

	packs, err := p.calculate(int(input.Item))
	if err != nil {
		return CreateOrderResponse{}, err
	}

	create := model.NewCreateOrder(
		p.BasePolicy.GenerateID(),
		input.UserID,
		input.Status,
		input.TypeProduct,
		input.Price,
		input.Item,
		packs,
		p.Now(),
		p.Now(),
	)

	err = p.orderService.CreateOrder(ctx, create)
	if err != nil {
		if errors.Is(err, domainOrder.ErrOrderAlreadyExist) {
			return CreateOrderResponse{}, ErrOrderAlreadyExists
		}
		return CreateOrderResponse{}, errors.Wrap(err, "orderService.CreateOrder")
	}

	response := CreateOrderResponse{
		Packs: packs,
	}

	return response, nil
}

func (p *Policy) calculate(items int) ([]model.Pack, error) {
	if items <= 0 {
		return nil, errors.New("invalid number of items ordered")
	}

	var packs []model.Pack
	remaining := items

	sort.Sort(sort.Reverse(sort.IntSlice(p.cfg.PacksSize.PackSize)))

	for _, size := range p.cfg.PacksSize.PackSize {
		if remaining <= 0 {
			break
		}
		count := remaining / size
		if count > 0 {
			packs = append(packs, model.Pack{Size: size, Count: count})
			remaining -= count * size
		}
	}

	if remaining > 0 {
		return nil, errors.New("unable to fulfill the order with available pack sizes")
	}

	return packs, nil
}

func (p *Policy) SwitchStatus(ctx context.Context, input SwitchStatusRequest) error {
	logging.L(ctx).Debug("SwitchStatus")

	switchStatus := model.NewSwitchStatus(
		input.ID,
		input.Status,
		p.Now(),
	)

	err := p.orderService.SwitchStatus(ctx, switchStatus)
	if err != nil {
		return errors.Wrap(err, "orderService.SwitchStatus")
	}

	return nil
}
