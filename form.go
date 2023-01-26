package zoox

// Form ...
type Form interface {
	Get(key string, defaultValue ...string) string
}

type form struct {
	ctx *Context
	//
	params map[string]string
}

func newForm(ctx *Context) Form {
	return &form{
		ctx:    ctx,
		params: make(map[string]string),
	}
}

// Get gets request form with the given name.
func (f *form) Get(key string, defaultValue ...string) string {
	value := f.ctx.Request.FormValue(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
}
