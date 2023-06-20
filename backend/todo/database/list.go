package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dackroyd/todo-list/backend/todo"
)

type ListRepository struct {
	db *sql.DB
}

func NewListRepository(db *sql.DB) *ListRepository {
	return &ListRepository{db: db}
}

func (r *ListRepository) Items(ctx context.Context, listID string) ([]todo.Item, error) {
	query := `
		SELECT id,
		       description,
		       due,
		       completed
		  FROM items
		 WHERE list_id = $1
	 `

	cols := func(i *todo.Item) []any {
		return []any{&i.ID, &i.Description, &i.Due, &i.Completed}
	}

	items, err := queryRows(ctx, r.db, cols, query, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to query for list items: %w", err)
	}

	return items, nil
}

func (r *ListRepository) List(ctx context.Context, listID string) (*todo.List, error) {
	query := `
		SELECT id,
		       description
		  FROM lists
		  WHERE id = $1
	`

	cols := func(l *todo.List) []any {
		return []any{&l.ID, &l.Description}
	}

	list, err := queryRow(ctx, r.db, cols, query, listID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, todo.NotFoundError(fmt.Sprintf("list with id %q does not exist", listID))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query todo list: %w", err)
	}

	return list, nil
}

func (r *ListRepository) Lists(ctx context.Context) ([]todo.List, error) {
	query := `
		SELECT id,
		       description
		  FROM lists
	`

	cols := func(l *todo.List) []any {
		return []any{&l.ID, &l.Description}
	}

	lists, err := queryRows(ctx, r.db, cols, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query todo lists: %w", err)
	}

	return lists, nil
}
