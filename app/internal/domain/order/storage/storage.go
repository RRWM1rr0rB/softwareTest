package storage

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
	psql "github.com/WM1rr0rB8/librariesTest/backend/golang/postgresql"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/queryify"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/sfqb"
	"github.com/WM1rr0rB8/librariesTest/backend/golang/tracing"

	"software_test/internal/dal"
	"software_test/internal/dal/postgres"
	"software_test/internal/domain"
	domainOrder "software_test/internal/domain/order"
	"software_test/internal/domain/order/model"
)

type Storage struct {
	qb     squirrel.StatementBuilderType
	client *psql.Client
}

func NewStorage(client *psql.Client) *Storage {
	qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &Storage{client: client, qb: qb}
}

func (repo *Storage) All(ctx context.Context, filters sfqb.SFQB) ([]model.Order, error) {
	return repo.findBy(ctx, filters)
}

func (repo *Storage) findBy(ctx context.Context, filters sfqb.SFQB) ([]model.Order, error) {
	queryify.ApplySearchFilters(filters, domain.TextFormat, domain.Percent)

	queryify.ReplaceFilterLike(filters, domain.ILikeFormat)

	queryify.ReplaceTableToAlias(
		filters,
		postgres.OrderTable,
	)

	statement := repo.qb.
		Select(
			"o.id",
			"o.user_id",
			"o.number_order",
			"o.status",
			"o.type_product",
			"o.price",
			"o.item",
			"o.packs",
			"o.created_at",
			"o.updated_at",
		).
		From(postgres.OrderTable.From()).
		Where(filters.Where(), filters.Args()...)

	limit := filters.Limit()
	if limit > 0 {
		statement = statement.Limit(uint64(limit))
	}

	offset := filters.Offset()
	if offset > 0 {
		statement = statement.Offset(uint64(offset))
	}

	order := filters.Order()
	if order != "" {
		statement = statement.OrderBy(strings.Split(order, ",")...)
	}

	query, args, err := statement.ToSql()
	if err != nil {
		err = psql.ErrCreateQuery(err)
		tracing.Error(ctx, err)

		return nil, err
	}

	tracing.SpanEvent(ctx, "select Order query")
	tracing.TraceValue(ctx, "sql", query)

	for i, arg := range args {
		tracing.TraceValue(ctx, strconv.Itoa(i), arg)
	}

	rows, queryErr := repo.client.Query(ctx, query, args...)
	if queryErr != nil {
		queryErr = psql.ErrDoQuery(queryErr)
		tracing.Error(ctx, queryErr)

		return nil, queryErr
	}

	defer rows.Close()

	orders := make([]model.Order, 0, rows.CommandTag().RowsAffected())

	for rows.Next() {
		var ord model.Order
		var packsJSON []byte

		if orderErr := rows.Scan(
			&ord.ID,
			&ord.UserID,
			&ord.NumberOrder,
			&ord.Status,
			&ord.TypeProduct,
			&ord.Price,
			&ord.Item,
			&packsJSON,
			&ord.CreatedAt,
			&ord.UpdatedAt,
		); orderErr != nil {
			orderErr = psql.ErrScan(psql.ParsePgError(orderErr))
			tracing.Error(ctx, orderErr)
			return nil, orderErr
		}

		if len(packsJSON) > 0 {
			if packErr := json.Unmarshal(packsJSON, &ord.Pack); packErr != nil {
				tracing.Error(ctx, packErr)
			}
		}

		orders = append(orders, ord)
	}

	return orders, nil
}

func (repo *Storage) CreateOrder(ctx context.Context, order model.CreateOrder) error {
	packsJSON, err := json.Marshal(order.Pack)
	if err != nil {
		log.Fatalf("Error convert to JSON: %v", err)
	}

	query, args, err := repo.qb.
		Insert(postgres.OrderTable.String()).
		Columns(
			"id",
			"user_id",
			"status",
			"type_product",
			"price",
			"item",
			"packs",
			"created_at",
			"updated_at",
		).
		Values(
			order.ID,
			order.UserID,
			order.Status,
			order.TypeProduct,
			order.Price,
			order.Item,
			packsJSON,
			order.CreatedAt,
			order.UpdatedAt,
		).
		ToSql()
	if err != nil {
		err = psql.ErrCreateQuery(err)
		tracing.Error(ctx, err)

		return err
	}

	tracing.SpanEvent(ctx, "create order query")
	tracing.TraceValue(ctx, "sql", query)

	for i, arg := range args {
		tracing.TraceValue(ctx, strconv.Itoa(i), arg)
	}

	_, execErr := repo.client.Exec(ctx, query, args...)
	if execErr != nil {
		if pgErr, ok := psql.IsErrUniqueViolation(execErr); ok {
			switch pgErr.ConstraintName {
			case domainOrder.OrderIDPkConstraint:
				return errors.New("violates constraint order id pk")
			}
		}

		execErr = psql.ErrDoQuery(psql.ParsePgError(execErr))

		return execErr
	}

	return nil
}

func (repo *Storage) SwitchStatus(ctx context.Context, order model.SwitchStatus) error {
	query, args, err := repo.qb.
		Update(postgres.OrderTable.String()).
		Set("status", order.Status).
		Set("updated_at", order.UpdatedAt).
		Where(squirrel.Eq{"id": order.ID}).
		ToSql()
	if err != nil {
		err = psql.ErrCreateQuery(err)
		tracing.Error(ctx, err)

		return err
	}

	tracing.SpanEvent(ctx, "update order query")
	tracing.TraceValue(ctx, "sql", query)

	for i, arg := range args {
		tracing.TraceValue(ctx, strconv.Itoa(i), arg)
	}

	cmd, execErr := repo.client.Exec(ctx, query, args...)
	if execErr != nil {
		if pgErr, ok := psql.IsErrUniqueViolation(execErr); ok {
			switch pgErr.ConstraintName {
			case domainOrder.OrderIDPkConstraint:
				return errors.New("violates constraint order id pk")
			}
		}

		execErr = psql.ErrDoQuery(psql.ParsePgError(execErr))

		return execErr
	}

	if cmd.RowsAffected() == 0 {
		return dal.ErrNotFound
	}

	return nil
}
