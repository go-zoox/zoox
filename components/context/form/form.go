package form

import "net/http"

// Form ...
type Form interface {
	Get(key string, defaultValue ...string) string
}

type form struct {
	request *http.Request
	//
	params map[string]string
}

// New creates a form.
func New(request *http.Request) Form {
	return &form{
		request: request,
		params:  make(map[string]string),
	}
}

// Get gets request form with the given name.
func (f *form) Get(key string, defaultValue ...string) string {
	value := f.request.FormValue(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
}
