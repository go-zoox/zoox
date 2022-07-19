package zoox

type User struct {
	u interface{}
}

func newUser() *User {
	return &User{}
}

func (s *User) Get() interface{} {
	return s.u
}

func (s *User) Set(user interface{}) {
	s.u = user
}
