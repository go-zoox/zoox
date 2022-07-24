package zoox

// User is the user for request context.
type User struct {
	u interface{}
}

func newUser() *User {
	return &User{}
}

// Get gets user from request context
func (s *User) Get() interface{} {
	return s.u
}

// Set sets user to request context
func (s *User) Set(user interface{}) {
	s.u = user
}
