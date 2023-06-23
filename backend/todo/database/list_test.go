package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

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

func TestList(t *testing.T) {
	t.Parallel()

	type args struct {
		ListID string
	}

	type fields struct {
		MockExpectations func(sqlmock.Sqlmock)
	}

	type want struct {
		Error error
		List  *todo.DueList
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
					mockListQuery(mock, "1").WillReturnError(queryErr)
				},
			},
			Want: want{Error: queryErr},
		},
		"No Result": {
			Args: args{ListID: "1"},
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockListQuery(mock, "1").WillReturnRows(mockListRows())
				},
			},
			Want: want{Error: todo.NotFoundError(`list with id "1" does not exist`)},
		},
		"Exists - No Items Due": {
			Args: args{ListID: "2"},
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockListQuery(mock, "2").WillReturnRows(mockListRows(todo.List{ID: "2", Description: "Golang-Syd Meetup June 2023"}))
					mockItemsQueryDue(mock, "2").WillReturnRows(mockItemRows())
				},
			},
			Want: want{List: &todo.DueList{List: todo.List{ID: "2", Description: "Golang-Syd Meetup June 2023"}}},
		},
		"Exists - With Due Items": {
			Args: args{ListID: "2"},
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockListQuery(mock, "2").WillReturnRows(mockListRows(todo.List{ID: "2", Description: "Golang-Syd Meetup June 2023"}))
					mockItemsQueryDue(mock, "2").WillReturnRows(mockItemRows(
						todo.Item{ID: "1", Description: "Prepare Presentation", Due: ptr(time.Date(2023, time.June, 20, 8, 0, 0, 0, time.UTC))},
						todo.Item{ID: "2", Description: "Practice", Due: ptr(time.Date(2023, time.June, 26, 0, 0, 0, 0, time.UTC))},
						todo.Item{ID: "3", Description: "Attend & Present", Due: ptr(time.Date(2023, time.June, 29, 8, 0, 0, 0, time.UTC))},
					))
				},
			},
			Want: want{
				List: &todo.DueList{
					List: todo.List{ID: "2", Description: "Golang-Syd Meetup June 2023"},
					DueItems: []todo.Item{
						{ID: "1", Description: "Prepare Presentation", Due: ptr(time.Date(2023, time.June, 20, 8, 0, 0, 0, time.UTC))},
						{ID: "2", Description: "Practice", Due: ptr(time.Date(2023, time.June, 26, 0, 0, 0, 0, time.UTC))},
						{ID: "3", Description: "Attend & Present", Due: ptr(time.Date(2023, time.June, 29, 8, 0, 0, 0, time.UTC))},
					},
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

			list, err := repo.List(context.Background(), tt.Args.ListID)

			assert.NoError(t, mock.ExpectationsWereMet(), "DB Expectations")

			if tt.Want.Error != nil {
				assert.ErrorIs(t, err, tt.Want.Error, "Retrieval error")
				return
			}

			require.NoError(t, err, "Retrieval error")
			assert.Equal(t, tt.Want.List, list, "List")
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
		Lists []todo.DueList
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
		"Non-empty with no due items": {
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockListsQuery(mock).WillReturnRows(mockListRows(
						todo.List{ID: "1", Description: "Chores"},
						todo.List{ID: "2", Description: "Golang-Syd June 2023"},
						todo.List{ID: "3", Description: "Holiday"},
					))
					mockItemsQueryDue(mock, "1").WillReturnRows(mockItemRows())
					mockItemsQueryDue(mock, "2").WillReturnRows(mockItemRows())
					mockItemsQueryDue(mock, "3").WillReturnRows(mockItemRows())
				},
			},
			Want: want{
				Lists: []todo.DueList{
					{List: todo.List{ID: "1", Description: "Chores"}},
					{List: todo.List{ID: "2", Description: "Golang-Syd June 2023"}},
					{List: todo.List{ID: "3", Description: "Holiday"}},
				},
			},
		},
		"Lists with due items": {
			Fields: fields{
				MockExpectations: func(mock sqlmock.Sqlmock) {
					mockListsQuery(mock).WillReturnRows(mockListRows(
						todo.List{ID: "1", Description: "Chores"},
						todo.List{ID: "2", Description: "Golang-Syd June 2023"},
						todo.List{ID: "3", Description: "Holiday"},
					))
					mockItemsQueryDue(mock, "1").WillReturnRows(mockItemRows(
						todo.Item{ID: "1", Description: "Washing", Due: ptr(time.Date(2023, time.June, 20, 8, 0, 0, 0, time.UTC))},
						todo.Item{ID: "2", Description: "Mop Floors", Due: ptr(time.Date(2023, time.June, 21, 10, 0, 0, 0, time.UTC))},
						todo.Item{ID: "3", Description: "Groceries", Due: ptr(time.Date(2023, time.June, 22, 2, 0, 0, 0, time.UTC))},
					))
					mockItemsQueryDue(mock, "2").WillReturnRows(mockItemRows(
						todo.Item{ID: "4", Description: "Prepare Presentation", Due: ptr(time.Date(2023, time.June, 20, 8, 0, 0, 0, time.UTC))},
						todo.Item{ID: "5", Description: "Practice", Due: ptr(time.Date(2023, time.June, 26, 0, 0, 0, 0, time.UTC))},
						todo.Item{ID: "6", Description: "Attend & Present", Due: ptr(time.Date(2023, time.June, 29, 8, 0, 0, 0, time.UTC))},
					))
					mockItemsQueryDue(mock, "3").WillReturnRows(mockItemRows())
				},
			},
			Want: want{
				Lists: []todo.DueList{
					{
						List: todo.List{ID: "1", Description: "Chores"},
						DueItems: []todo.Item{
							{ID: "1", Description: "Washing", Due: ptr(time.Date(2023, time.June, 20, 8, 0, 0, 0, time.UTC))},
							{ID: "2", Description: "Mop Floors", Due: ptr(time.Date(2023, time.June, 21, 10, 0, 0, 0, time.UTC))},
							{ID: "3", Description: "Groceries", Due: ptr(time.Date(2023, time.June, 22, 2, 0, 0, 0, time.UTC))},
						},
					},
					{
						List: todo.List{ID: "2", Description: "Golang-Syd June 2023"},
						DueItems: []todo.Item{
							{ID: "4", Description: "Prepare Presentation", Due: ptr(time.Date(2023, time.June, 20, 8, 0, 0, 0, time.UTC))},
							{ID: "5", Description: "Practice", Due: ptr(time.Date(2023, time.June, 26, 0, 0, 0, 0, time.UTC))},
							{ID: "6", Description: "Attend & Present", Due: ptr(time.Date(2023, time.June, 29, 8, 0, 0, 0, time.UTC))},
						},
					},
					{List: todo.List{ID: "3", Description: "Holiday"}},
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
		-- Name: TODO List Items
		SELECT id,
		       description,
		       due,
		       completed
		  FROM items
		 WHERE list_id = $1
	`

	return mock.ExpectQuery(q).WithArgs(listID)
}

func mockItemsQueryDue(mock sqlmock.Sqlmock, listID string) *sqlmock.ExpectedQuery {
	q := `
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

	return mock.ExpectQuery(q).WithArgs(listID)
}

func mockItemRows(items ...todo.Item) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "description", "due", "completed"})

	for _, item := range items {
		rows.AddRow(item.ID, item.Description, item.Due, item.Completed)
	}

	return rows
}

func mockListQuery(mock sqlmock.Sqlmock, listID string) *sqlmock.ExpectedQuery {
	q := `
		-- Name: TODO List
		SELECT id,
		       description
		  FROM lists
		 WHERE id = $1
	`

	return mock.ExpectQuery(q).WithArgs(listID)
}

func mockListsQuery(mock sqlmock.Sqlmock) *sqlmock.ExpectedQuery {
	q := `
		-- Name: TODO Lists
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

func ptr[T any](t T) *T {
	return &t
}
