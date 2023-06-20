package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dackroyd/todo-list/backend/todo"
	"github.com/dackroyd/todo-list/backend/todo/database"
)

func TestItems(t *testing.T) {
	t.Parallel()

	type args struct {
		ListID string
	}

	type fields struct {
		MockExpectations func(sqlmock.Sqlmock)
	}

	type want struct {
		Error error
		Items []todo.Item
	}

	queryErr := errors.New("failed to execute query")

	testTable := map[string]struct {
		Args   args
		Fields fields
		Want   want
	}{
		"Query failure": {
			Args: args{ListID: "1"},
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockItemsQuery(mock, "1").WillReturnError(queryErr)
				},
			},
			Want: want{Error: queryErr},
		},
		"Empty": {
			Args: args{ListID: "1"},
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockItemsQuery(mock, "1").WillReturnRows(mockItemRows())
				},
			},
		},
		"Non-empty": {
			Args: args{ListID: "2"},
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockItemsQuery(mock, "2").WillReturnRows(mockItemRows(
						todo.Item{ID: "1", Description: "Bananas"},
						todo.Item{ID: "2", Description: "Apples"},
						todo.Item{ID: "3", Description: "Strawberries"},
					))
				},
			},
			Want: want{
				Items: []todo.Item{
					{ID: "1", Description: "Bananas"},
					{ID: "2", Description: "Apples"},
					{ID: "3", Description: "Strawberries"},
				},
			},
		},
	}

	for name, tt := range testTable {
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, mock := mockDB(t)
			repo := database.NewListRepository(db)

			tt.Fields.MockExpectations(mock)

			items, err := repo.Items(context.Background(), tt.Args.ListID)

			assert.NoError(t, mock.ExpectationsWereMet(), "DB Expectations")

			if tt.Want.Error != nil {
				assert.ErrorIs(t, err, tt.Want.Error, "Retrieval error")
				return
			}

			require.NoError(t, err, "Retrieval error")
			assert.Equal(t, tt.Want.Items, items, "Items")
		})
	}
}

func TestLists(t *testing.T) {
	t.Parallel()

	type fields struct {
		MockExpectations func(sqlmock.Sqlmock)
	}

	type want struct {
		Error error
		Lists []todo.List
	}

	queryErr := errors.New("failed to execute query")

	testTable := map[string]struct {
		Fields fields
		Want   want
	}{
		"Query failure": {
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockListsQuery(mock).WillReturnError(queryErr)
				},
			},
			Want: want{Error: queryErr},
		},
		"No Lists": {
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockListsQuery(mock).WillReturnRows(mockListRows())
				},
			},
		},
		"Non-empty": {
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockListsQuery(mock).WillReturnRows(mockListRows(
						todo.List{ID: "1", Description: "Chores"},
						todo.List{ID: "2", Description: "Golang-Syd June 2023"},
						todo.List{ID: "3", Description: "Holiday"},
					))
				},
			},
			Want: want{
				Lists: []todo.List{
					{ID: "1", Description: "Chores"},
					{ID: "2", Description: "Golang-Syd June 2023"},
					{ID: "3", Description: "Holiday"},
				},
			},
		},
	}

	for name, tt := range testTable {
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, mock := mockDB(t)
			repo := database.NewListRepository(db)

			tt.Fields.MockExpectations(mock)

			lists, err := repo.Lists(context.Background())

			assert.NoError(t, mock.ExpectationsWereMet(), "DB Expectations")

			if tt.Want.Error != nil {
				assert.ErrorIs(t, err, tt.Want.Error, "Retrieval error")
				return
			}

			require.NoError(t, err, "Retrieval error")
			assert.Equal(t, tt.Want.Lists, lists, "Lists")
		})
	}
}

func mockItemsQuery(mock sqlmock.Sqlmock, listID string) *sqlmock.ExpectedQuery {
	q := `
		SELECT id,
		       description,
		       due,
		       completed
		  FROM items
		 WHERE list_id = $1
	`

	return mock.ExpectQuery(q).WithArgs(listID)
}

func mockItemRows(items ...todo.Item) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "description", "due", "completed"})

	for _, item := range items {
		rows.AddRow(item.ID, item.Description, item.Due, item.Completed)
	}

	return rows
}

func mockListsQuery(mock sqlmock.Sqlmock) *sqlmock.ExpectedQuery {
	q := `
		SELECT id,
		       description
		  FROM lists
	`

	return mock.ExpectQuery(q)
}

func mockListRows(lists ...todo.List) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "description"})

	for _, list := range lists {
		rows.AddRow(list.ID, list.Description)
	}

	return rows
}

func mockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err, "Opening a stub database connection encountered an error")

	t.Cleanup(func() {
		mock.ExpectClose()
		assert.NoError(t, db.Close(), "Closing mock DB")
	})

	return db, mock
}
