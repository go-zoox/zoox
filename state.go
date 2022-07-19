package zoox

type State struct {
	data map[string]interface{}
}

func newState() *State {
	return &State{
		data: make(map[string]interface{}),
	}
}

func (s *State) Get(key string) interface{} {
	return s.data[key]
}

func (s *State) Set(key string, value interface{}) {
	s.data[key] = value
}
