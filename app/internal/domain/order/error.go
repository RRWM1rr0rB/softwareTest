package order

import (
	"github.com/WM1rr0rB8/librariesTest/backend/golang/errors"
)

// -------------------------------------- Errors and constants from service  --------------------------------------

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderAlreadyExist = errors.New("collection already exist")
)

// -------------------------------------- Errors and constants from storage  --------------------------------------

const (
	OrderIDPkConstraint = "order_id_pk"
)

var (
	ErrViolatesConstraintOrderIdPK = errors.New("violates constraint order id pk")
)
