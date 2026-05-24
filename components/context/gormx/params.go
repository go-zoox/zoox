package gormx

import (
	"github.com/go-zoox/core-utils/safe"
	gormxlib "github.com/go-zoox/gormx"
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/components/context/param"
	"github.com/go-zoox/zoox/components/context/query"
)

// NewParams creates gormx params from zoox context.
func NewParams(ctx *zoox.Context) *gormxlib.Params {
	return gormxlib.NewParams(&contextAdapter{
		ctx: ctx,
	})
}

type contextAdapter struct {
	ctx *zoox.Context
}

func (a *contextAdapter) BindQuery(obj interface{}) error {
	return a.ctx.BindQuery(obj)
}

func (a *contextAdapter) Param() gormxlib.ParamsGetter {
	return &paramGetterAdapter{
		getter: a.ctx.Param(),
	}
}

func (a *contextAdapter) Query() gormxlib.ParamsGetter {
	return &queryGetterAdapter{
		getter: a.ctx.Query(),
	}
}

func (a *contextAdapter) Queries() gormxlib.ParamsQueries {
	return &queriesAdapter{
		queries: a.ctx.Queries(),
	}
}

type valueAdapter struct {
	value interface {
		String() string
		Int64() int64
	}
}

func (a *valueAdapter) String() string {
	return a.value.String()
}

func (a *valueAdapter) Int64() int64 {
	return a.value.Int64()
}

type paramGetterAdapter struct {
	getter param.Param
}

func (a *paramGetterAdapter) Get(key string) gormxlib.ParamsValue {
	return &valueAdapter{
		value: a.getter.Get(key),
	}
}

type queryGetterAdapter struct {
	getter query.Query
}

func (a *queryGetterAdapter) Get(key string) gormxlib.ParamsValue {
	return &valueAdapter{
		value: a.getter.Get(key),
	}
}

type queriesAdapter struct {
	queries *safe.Map[string, any]
}

func (a *queriesAdapter) Del(key string) {
	a.queries.Del(key)
}

func (a *queriesAdapter) Iterator() map[string]any {
	return a.queries.Iterator()
}
