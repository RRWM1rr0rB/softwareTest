package order

import (
	"slices"
	"time"

	gRPCOrderService "github.com/WM1rr0rB8/contractsTest/gen/go/order_service/v1"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/apperror"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/errors"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/queryify"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/sfqb"

	"software_test/internal/config"
	"software_test/internal/dal/postgres"
	"software_test/internal/domain"
)

const (
	domainName           = "order"
	operatorNotSupported = "operator not supported"
)

const (
	maxLimit     = 1000
	defaultLimit = 100
)

const (
	fieldNameID          = "order.id"
	fieldNameUserID      = "order.user_id"
	fieldNameNumberOrder = "order.number_order"
	fieldNameStatus      = "order.status"
	fieldNameTypeProduct = "order.type_product"
	fieldNamePrice       = "order.price"
	fieldNameItem        = "order.item"
	fieldNameCreatedAt   = "order.created_at"
	fieldNameUpdatedAt   = "order.updated_at"
)

var AllOrderFields = []string{
	fieldNameID,
	fieldNameUserID,
	fieldNameNumberOrder,
	fieldNameStatus,
	fieldNameTypeProduct,
	fieldNamePrice,
	fieldNameItem,
	fieldNameCreatedAt,
	fieldNameUpdatedAt,
}

//nolint:funlen,gocognit,gocyclo,ineffassign
func buildValidationOrderFilters(req *gRPCOrderService.SearchOrderRequest) (sfqb.SFQB, error) {
	searchFields := []string{
		fieldNameID,
		fieldNameUserID,
		fieldNameNumberOrder,
		fieldNameStatus,
		fieldNameTypeProduct,
		fieldNamePrice,
		fieldNameItem,
		fieldNameCreatedAt,
		fieldNameUpdatedAt,
	}

	errFields := apperror.ErrorFields{}

	filters, err := queryify.NewFilters(
		queryify.WithSearchFields(searchFields),
		queryify.WithPaginator(req, defaultLimit, maxLimit),
		queryify.WithSorter(req),
		queryify.WithSearcher(req),
		queryify.WithMinSearchSymbols(postgres.SearchMinSymbols),
	)
	if err != nil {
		if errors.Is(err, queryify.ErrSearchMinSymbols) {
			return nil, apperror.NewValidationError(
				domain.SystemCode,
				apperror.WithDomain(domainName),
				apperror.WithMessage("search minimum symbols error"),
				apperror.WithCode(minSearchErrCode),
			)
		}

		return nil, errors.Wrap(err, "queryify.NewFilters")
	}

	var (
		operator sfqb.Method
		parseErr error
	)

	if idVal := req.GetId(); idVal != nil {
		operator, parseErr = queryify.MapOperator(idVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameID] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameID,
				Method: operator,
				Value:  idVal.GetVal(),
			})
		}
	}

	if userIDVal := req.GetUserId(); userIDVal != nil {
		operator, parseErr = queryify.MapOperator(userIDVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameUserID] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameUserID,
				Method: operator,
				Value:  userIDVal.GetVal(),
			})
		}
	}

	if numberOrderVal := req.GetNumberOrder(); numberOrderVal != nil {
		operator, parseErr = queryify.MapOperator(numberOrderVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameNumberOrder] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameNumberOrder,
				Method: operator,
				Value:  numberOrderVal.GetVal(),
			})
		}
	}

	if statusVal := req.GetStatus(); statusVal != nil {
		operator, parseErr = queryify.MapOperator(statusVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameStatus] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameStatus,
				Method: operator,
				Value:  statusVal.GetVal(),
			})
		}
	}

	if typeProductVal := req.GetTypeProduct(); typeProductVal != nil {
		operator, parseErr = queryify.MapOperator(typeProductVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameTypeProduct] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameTypeProduct,
				Method: operator,
				Value:  typeProductVal.GetVal(),
			})
		}
	}

	if priceVal := req.GetPrice(); priceVal != nil {
		operator, parseErr = queryify.MapOperator(priceVal.GetOp())
		if parseErr != nil {
			errFields[fieldNamePrice] = parseErr.Error()
		}

		if !slices.Contains([]sfqb.Method{sfqb.NE, sfqb.EQ}, operator) {
			errFields[fieldNamePrice] = operatorNotSupported
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNamePrice,
				Method: operator,
				Value:  priceVal.GetVal(),
			})
		}
	}

	if itemVal := req.GetItem(); itemVal != nil {
		operator, parseErr = queryify.MapOperator(itemVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameItem] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameItem,
				Method: operator,
				Value:  itemVal.GetVal(),
			})
		}
	}

	if createdAtFilter := req.GetCreatedAtToFilter(); createdAtFilter != nil {
		switch createdAtFilter.(type) {
		case *gRPCOrderService.SearchOrderRequest_CreatedAtVal:
			operator, parseErr = queryify.MapOperator(req.GetCreatedAtVal().GetOp())
			if parseErr != nil {
				errFields[fieldNameCreatedAt] = parseErr.Error()
			} else {
				createdVal := req.GetCreatedAtVal().GetVal()

				valTimestamp := time.UnixMilli(createdVal)

				intVal := valTimestamp.UTC().Format(config.TimeFormat)

				filters.AddFilter(sfqb.NewFilterField(fieldNameCreatedAt, operator, intVal))
			}
		case *gRPCOrderService.SearchOrderRequest_CreatedAtRange:
			operator, parseErr = queryify.MapOperator(req.GetCreatedAtRange().GetOp())
			if parseErr != nil {
				errFields[fieldNameCreatedAt] = parseErr.Error()
			} else {
				rangeFrom := req.GetCreatedAtRange().GetFrom()
				rangeTo := req.GetCreatedAtRange().GetTo()

				fromTimestamp := time.UnixMilli(rangeFrom)
				toTimestamp := time.UnixMilli(rangeTo)

				from := fromTimestamp.UTC().Format(config.TimeFormat)
				to := toTimestamp.UTC().Format(config.TimeFormat)

				filters.AddFilter(sfqb.NewFilterField(fieldNameCreatedAt, sfqb.GTE, from))
				filters.AddFilter(sfqb.NewFilterField(fieldNameCreatedAt, sfqb.LTE, to))
			}
		}
	}

	if updatedAtFilter := req.GetUpdatedAtToFilter(); updatedAtFilter != nil {
		switch updatedAtFilter.(type) {
		case *gRPCOrderService.SearchOrderRequest_UpdatedAtVal:
			operator, parseErr = queryify.MapOperator(req.GetUpdatedAtVal().GetOp())
			if parseErr != nil {
				errFields[fieldNameUpdatedAt] = parseErr.Error()
			} else {
				updatedVal := req.GetUpdatedAtVal().GetVal()

				valTimestamp := time.UnixMilli(updatedVal)

				intVal := valTimestamp.UTC().Format(config.TimeFormat)

				filters.AddFilter(sfqb.NewFilterField(fieldNameUpdatedAt, operator, intVal))
			}
		case *gRPCOrderService.SearchOrderRequest_UpdatedAtRange:
			operator, parseErr = queryify.MapOperator(req.GetUpdatedAtRange().GetOp())
			if parseErr != nil {
				errFields[fieldNameUpdatedAt] = parseErr.Error()
			} else {
				rangeFrom := req.GetUpdatedAtRange().GetFrom()
				rangeTo := req.GetUpdatedAtRange().GetTo()

				fromTimestamp := time.UnixMilli(rangeFrom)
				toTimestamp := time.UnixMilli(rangeTo)

				from := fromTimestamp.UTC().Format(config.TimeFormat)
				to := toTimestamp.UTC().Format(config.TimeFormat)

				filters.AddFilter(sfqb.NewFilterField(fieldNameUpdatedAt, sfqb.GTE, from))
				filters.AddFilter(sfqb.NewFilterField(fieldNameUpdatedAt, sfqb.LTE, to))
			}
		}
	}

	if len(errFields) > 0 {
		return nil, apperror.NewValidationError(
			domain.SystemCode,
			apperror.WithDomain(domainName),
			apperror.WithMessage("validation error"),
			apperror.WithFields(errFields),
			apperror.WithCode(validationErrCode),
		)
	}

	return filters, nil
}
