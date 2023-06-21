package routes_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dackroyd/todo-list/backend/todo"
	"github.com/dackroyd/todo-list/backend/todo/routes"
)

func TestListsAPI_Items(t *testing.T) {
	t.Parallel()

	type args struct {
		ListID string
	}

	type fields struct {
		MockExpectations func(ctx context.Context, l *listRepo)
	}

	type want struct {
		Body    string
		Code    int
		Headers http.Header
	}

	testTable := map[string]struct {
		Args   args
		Fields fields
		Want   want
	}{
		"Empty List ID Path Param": {
			Args:   args{ListID: "%20"},
			Fields: fields{MockExpectations: func(context.Context, *listRepo) {}},
			Want:   want{Body: `{"error": "\"list_id\" path param must not be blank"}`, Code: http.StatusBadRequest},
		},
		"Query failure": {
			Args: args{ListID: "1"},
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					l.OnItems(ctx, "1").Return(nil, errors.New("query failure"))
				},
			},
			Want: want{Body: `{"error": "Internal Server Error"}`, Code: http.StatusInternalServerError},
		},
		"No Items": {
			Args: args{ListID: "2"},
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					l.OnItems(ctx, "2").Return(nil, nil)
				},
			},
			Want: want{Body: `{"items": []}`, Code: http.StatusOK},
		},
		"Items": {
			Args: args{ListID: "3"},
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					goSyd := time.Date(2023, time.June, 29, 8, 0, 0, 0, time.UTC)
					items := []todo.Item{
						{ID: "1", Description: "Relax"},
						{ID: "2", Description: "Golang-Syd Meetup June 2023", Due: &goSyd},
					}
					l.OnItems(ctx, "3").Return(items, nil)
				},
			},
			Want: want{
				Body: `{
					"items": [
						{"id": "1", "description": "Relax", "due": null, "completed": null},
						{"id": "2", "description": "Golang-Syd Meetup June 2023", "due": "2023-06-29T08:00:00Z", "completed": null}
					]
				}`,
				Code: http.StatusOK,
			},
		},
	}

	for name, tt := range testTable {
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			defer failOnPanic(t)

			testLogger := NewTestLogger(t)
			ctx := withTestContext(context.Background(), t)

			var repo listRepo
			listsAPI := routes.NewListAPI(&repo)

			tt.Fields.MockExpectations(ctx, &repo)
			defer mock.AssertExpectationsForObjects(t, &repo)

			h := routes.Handler(listsAPI, testLogger)

			route := fmt.Sprintf("/api/v1/lists/%s/items", tt.Args.ListID)
			req := httptest.NewRequest(http.MethodGet, route, http.NoBody).WithContext(ctx)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			res := rec.Result()

			assert.Equal(t, tt.Want.Code, res.StatusCode, "HTTP Status Code")

			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err, "Body Read Error")
			assert.JSONEq(t, tt.Want.Body, string(body), "HTTP Response Body")
		})
	}
}

func TestListsAPI_List(t *testing.T) {
	t.Parallel()

	type args struct {
		ListID string
	}

	type fields struct {
		MockExpectations func(ctx context.Context, l *listRepo)
	}

	type want struct {
		Body    string
		Code    int
		Headers http.Header
	}

	testTable := map[string]struct {
		Args   args
		Fields fields
		Want   want
	}{
		"Empty List ID Path Param": {
			Args:   args{ListID: "%20"},
			Fields: fields{MockExpectations: func(context.Context, *listRepo) {}},
			Want:   want{Body: `{"error": "\"list_id\" path param must not be blank"}`, Code: http.StatusBadRequest},
		},
		"Query failure": {
			Args: args{ListID: "1"},
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					l.OnList(ctx, "1").Return(nil, errors.New("query failure"))
				},
			},
			Want: want{Body: `{"error": "Internal Server Error"}`, Code: http.StatusInternalServerError},
		},
		"Not Found": {
			Args: args{ListID: "2"},
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					l.OnList(ctx, "2").Return(nil, todo.NotFoundError("list not found"))
				},
			},
			Want: want{Body: `{"error": "list not found"}`, Code: http.StatusNotFound},
		},
		"Exists": {
			Args: args{ListID: "1"},
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					list := &todo.DueList{List: todo.List{ID: "1", Description: "Golang-Syd Meetup June 2023"}}
					l.OnList(ctx, "1").Return(list, nil)
				},
			},
			Want: want{
				Body: `{
					"list": {"id": "1", "description": "Golang-Syd Meetup June 2023"},
					"dueItems": null
				}`,
				Code: http.StatusOK,
			},
		},
		"Has Due Items": {
			Args: args{ListID: "1"},
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					list := &todo.DueList{
						List: todo.List{ID: "1", Description: "Golang-Syd Meetup June 2023"},
						DueItems: []todo.Item{
							{ID: "1", Description: "Washing", Due: ptr(time.Date(2023, time.June, 20, 8, 0, 0, 0, time.UTC))},
							{ID: "2", Description: "Mop Floors", Due: ptr(time.Date(2023, time.June, 21, 10, 0, 0, 0, time.UTC))},
							{ID: "3", Description: "Groceries", Due: ptr(time.Date(2023, time.June, 22, 2, 0, 0, 0, time.UTC))},
						},
					}
					l.OnList(ctx, "1").Return(list, nil)
				},
			},
			Want: want{
				Body: `{
					"list": {"id": "1", "description": "Golang-Syd Meetup June 2023"},
					"dueItems": [
						{"id": "1", "description": "Washing", "due": "2023-06-20T08:00:00Z", "completed": null},
						{"id": "2", "description": "Mop Floors", "due": "2023-06-21T10:00:00Z", "completed": null},
						{"id": "3", "description": "Groceries", "due": "2023-06-22T02:00:00Z", "completed": null}
					]
				}`,
				Code: http.StatusOK,
			},
		},
	}

	for name, tt := range testTable {
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			defer failOnPanic(t)

			testLogger := NewTestLogger(t)
			ctx := withTestContext(context.Background(), t)

			var repo listRepo
			listsAPI := routes.NewListAPI(&repo)

			tt.Fields.MockExpectations(ctx, &repo)
			defer mock.AssertExpectationsForObjects(t, &repo)

			h := routes.Handler(listsAPI, testLogger)

			route := fmt.Sprintf("/api/v1/lists/%s", tt.Args.ListID)
			req := httptest.NewRequest(http.MethodGet, route, http.NoBody).WithContext(ctx)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			res := rec.Result()

			assert.Equal(t, tt.Want.Code, res.StatusCode, "HTTP Status Code")

			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err, "Body Read Error")
			assert.JSONEq(t, tt.Want.Body, string(body), "HTTP Response Body")
		})
	}
}

func TestListsAPI_Lists(t *testing.T) {
	t.Parallel()

	type fields struct {
		MockExpectations func(ctx context.Context, l *listRepo)
	}

	type want struct {
		Body    string
		Code    int
		Headers http.Header
	}

	testTable := map[string]struct {
		Fields fields
		Want   want
	}{
		"Query failure": {
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					l.OnLists(ctx).Return(nil, errors.New("query failure"))
				},
			},
			Want: want{Body: `{"error": "Internal Server Error"}`, Code: http.StatusInternalServerError},
		},
		"No Lists": {
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					l.OnLists(ctx).Return(nil, nil)
				},
			},
			Want: want{Body: `{"lists": []}`, Code: http.StatusOK},
		},
		"Lists": {
			Fields: fields{
				MockExpectations: func(ctx context.Context, l *listRepo) {
					lists := []todo.DueList{
						{
							List: todo.List{ID: "1", Description: "Chores"},
							DueItems: []todo.Item{
								{ID: "1", Description: "Washing", Due: ptr(time.Date(2023, time.June, 20, 8, 0, 0, 0, time.UTC))},
								{ID: "2", Description: "Mop Floors", Due: ptr(time.Date(2023, time.June, 21, 10, 0, 0, 0, time.UTC))},
								{ID: "3", Description: "Groceries", Due: ptr(time.Date(2023, time.June, 22, 2, 0, 0, 0, time.UTC))},
							},
						},
						{
							List: todo.List{ID: "2", Description: "Golang-Syd Meetup June 2023"},
							DueItems: []todo.Item{
								{ID: "4", Description: "Prepare Presentation", Due: ptr(time.Date(2023, time.June, 20, 8, 0, 0, 0, time.UTC))},
								{ID: "5", Description: "Practice", Due: ptr(time.Date(2023, time.June, 26, 0, 0, 0, 0, time.UTC))},
								{ID: "6", Description: "Attend & Present", Due: ptr(time.Date(2023, time.June, 29, 8, 0, 0, 0, time.UTC))},
							},
						},
					}
					l.OnLists(ctx).Return(lists, nil)
				},
			},
			Want: want{
				Body: `{
					"lists": [
						{
							"list": {"id": "1", "description": "Chores"},
							"dueItems": [
								{"id": "1", "description": "Washing", "due": "2023-06-20T08:00:00Z", "completed": null},
								{"id": "2", "description": "Mop Floors", "due": "2023-06-21T10:00:00Z", "completed": null},
								{"id": "3", "description": "Groceries", "due": "2023-06-22T02:00:00Z", "completed": null}
							]
						},
						{
							"list": {"id": "2", "description": "Golang-Syd Meetup June 2023"},
							"dueItems": [
								{"id": "4", "description": "Prepare Presentation", "due": "2023-06-20T08:00:00Z", "completed": null},
								{"id": "5", "description": "Practice", "due": "2023-06-26T00:00:00Z", "completed": null},
								{"id": "6", "description": "Attend & Present", "due": "2023-06-29T08:00:00Z", "completed": null}
							]
						}
					]
				}`,
				Code: http.StatusOK,
			},
		},
	}

	for name, tt := range testTable {
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			defer failOnPanic(t)

			testLogger := NewTestLogger(t)
			ctx := withTestContext(context.Background(), t)

			var repo listRepo
			listsAPI := routes.NewListAPI(&repo)

			tt.Fields.MockExpectations(ctx, &repo)
			defer mock.AssertExpectationsForObjects(t, &repo)

			h := routes.Handler(listsAPI, testLogger)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/lists", http.NoBody).WithContext(ctx)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			res := rec.Result()

			assert.Equal(t, tt.Want.Code, res.StatusCode, "HTTP Status Code")

			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err, "Body Read Error")
			assert.JSONEq(t, tt.Want.Body, string(body), "HTTP Response Body")
		})
	}
}

type listRepo struct {
	mock.Mock
}

func (l *listRepo) Items(ctx context.Context, listID string) ([]todo.Item, error) {
	args := l.Called(testContext(ctx), listID)
	return args.Get(0).([]todo.Item), args.Error(1)
}

// OnItems provides a type-safe mock setup function, used instead of using 'On("Items, ...)'
func (l *listRepo) OnItems(ctx context.Context, listID string) *call2[[]todo.Item, error] {
	m := l.On("Items", testContext(ctx), listID)
	return &call2[[]todo.Item, error]{m: m}
}

func (l *listRepo) List(ctx context.Context, listID string) (*todo.DueList, error) {
	args := l.Called(testContext(ctx), listID)
	return args.Get(0).(*todo.DueList), args.Error(1)
}

// OnList provides a type-safe mock setup function, used instead of using 'On("List, ...)'
func (l *listRepo) OnList(ctx context.Context, listID string) *call2[*todo.DueList, error] {
	m := l.On("List", testContext(ctx), listID)
	return &call2[*todo.DueList, error]{m: m}
}

func (l *listRepo) Lists(ctx context.Context) ([]todo.DueList, error) {
	args := l.Called(testContext(ctx))
	return args.Get(0).([]todo.DueList), args.Error(1)
}

// OnLists provides a type-safe mock setup function, used instead of using 'On("Lists, ...)'
func (l *listRepo) OnLists(ctx context.Context) *call2[[]todo.DueList, error] {
	m := l.On("Lists", testContext(ctx))
	return &call2[[]todo.DueList, error]{m: m}
}

func ptr[T any](t T) *T {
	return &t
}
