package state

// State is the state for request context.
type State interface {
	Get(key string) interface{}
	Set(key string, value interface{})
}

type state struct {
	data map[string]interface{}
}

// New creates a state.
func New() State {
	return &state{
		data: make(map[string]interface{}),
	}
}

// Get gets the value from context state with the given key.
func (s *state) Get(key string) interface{} {
	return s.data[key]
}

// Set sets the value to context state with the given key.
func (s *state) Set(key string, value interface{}) {
	s.data[key] = value
}
