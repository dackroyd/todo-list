package database

import (
	"context"
	"database/sql"
	"fmt"
)

type rowQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type rowsQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func queryRow[T any](ctx context.Context, db rowQuerier, columns func(*T) []any, query string, args ...any) (*T, error) {
	row := db.QueryRowContext(ctx, query, args...)
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("unable to query for lists: %w", err)
	}

	var item T
	if err := row.Scan(columns(&item)...); err != nil {
		return nil, fmt.Errorf("failed to scan row into type %T: %w", &item, err)
	}

	return &item, nil
}

func queryRows[T any](ctx context.Context, db rowsQuerier, columns func(*T) []any, query string, args ...any) ([]T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	var result []T
	for rows.Next() {
		var elem T

		if err := rows.Scan(columns(&elem)...); err != nil {
			return nil, fmt.Errorf("failed to scan row into type %T: %w", &elem, err)
		}

		result = append(result, elem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failure while iterating over rows: %w", err)
	}

	return result, nil
}
