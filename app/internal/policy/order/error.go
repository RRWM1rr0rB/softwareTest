package order

import (
	"github.com/WM1rr0rB8/librariesTest/backend/golang/apperror"

	"software_test/internal/domain"
)

const (
	orderNotFoundCode = iota + 100
	orderAlreadyExistsCode
)

var (
	ErrOrderNotFound = apperror.NewNotFoundError(
		domain.SystemCode,
		apperror.WithMessage("order not found"),
		apperror.WithCode(orderNotFoundCode),
		apperror.WithDomain(domain.Order),
	)

	ErrOrderAlreadyExists = apperror.NewInternalError(
		domain.SystemCode,
		apperror.WithMessage("order already exist"),
		apperror.WithCode(orderAlreadyExistsCode),
		apperror.WithDomain(domain.Order),
	)
)
