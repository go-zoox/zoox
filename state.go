package zoox

// State is the state for request context.
type State struct {
	data map[string]interface{}
}

func newState() *State {
	return &State{
		data: make(map[string]interface{}),
	}
}

// Get gets the value from context state with the given key.
func (s *State) Get(key string) interface{} {
	return s.data[key]
}

// Set sets the value to context state with the given key.
func (s *State) Set(key string, value interface{}) {
	s.data[key] = value
}
