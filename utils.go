package zoox

type Query struct {
	ctx *Context
}

func newQuery(ctx *Context) *Query {
	return &Query{
		ctx: ctx,
	}
}

func (q *Query) Get(key string, defaultValue ...string) string {
	value := q.ctx.Request.URL.Query().Get(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
}

type Param struct {
	ctx *Context
	//
	params map[string]string
}

func newParams(ctx *Context, value map[string]string) *Param {
	return &Param{
		ctx:    ctx,
		params: value,
	}
}

func (q *Param) Get(key string, defaultValue ...string) string {
	value, ok := q.params[key]
	if ok {
		return value
	}

	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
}

func (q *Param) Iterator() map[string]string {
	return q.params
}

type Form struct {
	ctx *Context
	//
	params map[string]string
}

func newForm(ctx *Context) *Form {
	return &Form{
		ctx:    ctx,
		params: make(map[string]string),
	}
}

func (f *Form) Get(key string, defaultValue ...string) string {
	value := f.ctx.Request.FormValue(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
}
