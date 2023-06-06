package database

import (
	"context"
	"database/sql"
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
