package user

// User is the user for request context.
type User interface {
	// Get gets user from request context
	Get() interface{}
	// Set sets user to request context
	Set(user interface{})
}

type user struct {
	u interface{}
}

func New() User {
	return &user{}
}

// Get gets user from request context
func (s *user) Get() interface{} {
	return s.u
}

// Set sets user to request context
func (s *user) Set(user interface{}) {
	s.u = user
}
