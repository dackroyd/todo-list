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
		-- Name: TODO List Items
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

func (r *ListRepository) List(ctx context.Context, listID string) (*todo.DueList, error) {
	query := `
		-- Name: TODO List
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

	due, err := r.dueItems(ctx, listID)
	if err != nil {
		return nil, err
	}

	return &todo.DueList{DueItems: due, List: *list}, nil
}

func (r *ListRepository) Lists(ctx context.Context) ([]todo.DueList, error) {
	query := `
		-- Name: TODO Lists
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

	if len(lists) == 0 {
		return nil, nil
	}

	dueList := make([]todo.DueList, len(lists))
	for i, l := range lists {
		due, err := r.dueItems(ctx, l.ID)
		if err != nil {
			return nil, err
		}

		dueList[i] = todo.DueList{DueItems: due, List: l}
	}

	return dueList, nil
}

func (r *ListRepository) dueItems(ctx context.Context, listID string) ([]todo.Item, error) {
	query := `
		-- Name: TODO Due List Items
		SELECT id,
		       description,
		       due,
		       completed
		  FROM items
		 WHERE list_id = $1
		   AND due <= now() + INTERVAL '1 day'
		   AND completed IS NULL
		 ORDER BY due
	 `

	itemCols := func(i *todo.Item) []any {
		return []any{&i.ID, &i.Description, &i.Due, &i.Completed}
	}

	items, err := queryRows(ctx, r.db, itemCols, query, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to query due items for todo list %q: %w", listID, err)
	}

	return items, nil
}
