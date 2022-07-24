package zoox

// HTTPError is a custom error type for HTTP errors.
type HTTPError interface {
	Status() int
	Code() int
	Message() string
	Error() string
	Raw() error
}
