package gormx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-zoox/zoox"
)

func performRequest(app *zoox.Application, target string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	app.ServeHTTP(recorder, request)
	return recorder
}

func TestNewParamsParsesListParams(t *testing.T) {
	var gotErr error
	var gotPage uint
	var gotPageSize uint
	var gotOrderBy string
	var gotWhereSQL string

	app := zoox.New()
	app.Get("/users/:id", func(ctx *zoox.Context) {
		params := NewParams(ctx)
		list, err := params.GetList()
		if err != nil {
			gotErr = err
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		gotPage = list.Page
		gotPageSize = list.PageSize
		gotOrderBy = list.OrderBy.Build()
		gotWhereSQL, _, _ = list.Where.Build()
		ctx.String(http.StatusOK, "ok")
	})

	response := performRequest(app, "/users/42?page=2&pageSize=20&status=active&orderBy=created_at:desc,id:asc")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	if gotErr != nil {
		t.Fatalf("expected no error, got %v", gotErr)
	}

	if gotPage != 2 {
		t.Fatalf("expected page=2, got %d", gotPage)
	}

	if gotPageSize != 20 {
		t.Fatalf("expected pageSize=20, got %d", gotPageSize)
	}

	if gotOrderBy != "created_at DESC,id ASC" {
		t.Fatalf("expected orderBy=created_at DESC,id ASC, got %s", gotOrderBy)
	}

	if gotWhereSQL != "status = ?" {
		t.Fatalf("expected where SQL 'status = ?', got %s", gotWhereSQL)
	}
}

func TestNewParamsParsesIDFromRouteParam(t *testing.T) {
	var gotID uint
	var gotErr error

	app := zoox.New()
	app.Get("/users/:id", func(ctx *zoox.Context) {
		params := NewParams(ctx)
		gotID, gotErr = params.ID()
		if gotErr != nil {
			ctx.Error(http.StatusBadRequest, gotErr.Error())
			return
		}

		ctx.String(http.StatusOK, "ok")
	})

	response := performRequest(app, "/users/99")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	if gotErr != nil {
		t.Fatalf("expected no error, got %v", gotErr)
	}

	if gotID != 99 {
		t.Fatalf("expected id=99, got %d", gotID)
	}
}

func TestNewParamsParsesRangeWhereMode(t *testing.T) {
	var gotErr error
	var gotWhereSQL string
	var gotWhereArgs []interface{}

	app := zoox.New()
	app.Get("/users/:id", func(ctx *zoox.Context) {
		params := NewParams(ctx)
		list, err := params.GetList()
		if err != nil {
			gotErr = err
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		gotWhereSQL, gotWhereArgs, gotErr = list.Where.Build()
		if gotErr != nil {
			ctx.Error(http.StatusInternalServerError, gotErr.Error())
			return
		}

		ctx.String(http.StatusOK, "ok")
	})

	response := performRequest(app, "/users/1?age=18,65:range[)")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	if gotErr != nil {
		t.Fatalf("expected no error, got %v", gotErr)
	}

	if gotWhereSQL != "(age >= ? AND age < ?)" {
		t.Fatalf("expected range SQL '(age >= ? AND age < ?)', got %s", gotWhereSQL)
	}

	if len(gotWhereArgs) != 2 || gotWhereArgs[0] != "18" || gotWhereArgs[1] != "65" {
		t.Fatalf("expected range args [18 65], got %#v", gotWhereArgs)
	}
}
