package postgres

import (
	"github.com/WM1rr0rB8/librariesTest/backend/golang/queryify"
)

var (
	OrderTable = queryify.NewTable("public", "order", "o", "id")
)
